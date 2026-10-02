// MIT License
//
// Copyright (c) 2018 buildtool
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package registry

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"regexp"
	"strings"
	"time"

	"github.com/apex/log"
	mobyclient "github.com/moby/moby/client"

	"github.com/moby/moby/api/types/registry"

	"github.com/buildtool/build-tools/pkg/docker"
)

var digestRegexp = regexp.MustCompile(`sha256:[a-fA-F0-9]{64}`)

type Registry interface {
	Configured() bool
	Name() string
	Login(client docker.Client) error
	GetAuthConfig() registry.AuthConfig
	GetAuthInfo() string
	RegistryUrl() string
	Create(repository string) error
	PushImage(client docker.Client, auth, image string) (string, error)
}

type responsetype struct {
	Status      string `json:"status"`
	ErrorDetail *struct {
		Message string `json:"message"`
	} `json:"errorDetail"`
	Error          string `json:"error"`
	ProgressDetail *struct {
		Current int64 `json:"current"`
		Total   int64 `json:"total"`
	} `json:"progressDetail"`
	Progress string `json:"progress"`
	Id       string `json:"id"`
	Aux      *struct {
		Tag    string `json:"Tag"`
		Digest string `json:"Digest"`
		Size   int64  `json:"Size"`
	} `json:"aux"`
}

// toLoginOptions converts a registry.AuthConfig to mobyclient.RegistryLoginOptions.
func toLoginOptions(auth registry.AuthConfig) mobyclient.RegistryLoginOptions {
	return mobyclient.RegistryLoginOptions{
		Username:      auth.Username,
		Password:      auth.Password,
		ServerAddress: auth.ServerAddress,
	}
}

type dockerRegistry struct{}

// A registry answers a burst of pushes with HTTP 429, which the daemon
// reports as an error line in the push stream ("toomanyrequests: Rate
// exceeded" on ECR). The burst is over within seconds, so a push that was
// throttled is retried a few times with a doubling wait rather than failing
// the build.
const (
	pushRetries = 3
	pushBackoff = 2 * time.Second
)

// pushSleep is time.Sleep, replaceable so tests do not wait.
var pushSleep = time.Sleep

func (dockerRegistry) PushImage(client docker.Client, auth, image string) (string, error) {
	backoff := pushBackoff
	for attempt := 0; ; attempt++ {
		digest, err := pushOnce(client, auth, image)
		if err == nil || !isThrottled(err) || attempt == pushRetries {
			return digest, err
		}
		log.Warnf("registry throttled the push of %s, retrying in %s", image, backoff)
		pushSleep(backoff)
		backoff *= 2
	}
}

func isThrottled(err error) bool {
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "toomanyrequests") || strings.Contains(msg, "too many requests")
}

func pushOnce(client docker.Client, auth, image string) (string, error) {
	// One tag per call. With All set the client drops the tag from the
	// request and the daemon pushes every local tag of the repository, so a
	// loop over N tags made N*N pushes - which is what tripped the registry's
	// rate limit in the first place.
	out, err := client.ImagePush(context.Background(), image, mobyclient.ImagePushOptions{RegistryAuth: auth})
	if err != nil {
		return "", err
	}
	var digestResult string
	scanner := bufio.NewScanner(out)
	for scanner.Scan() {
		r := &responsetype{}
		response := scanner.Bytes()
		if err := json.Unmarshal(response, &r); err != nil {
			log.Errorf("Unable to parse response: %s, Error: %v\n", string(response), err)
			return "", err
		}
		if r.ErrorDetail != nil {
			return "", errors.New(r.ErrorDetail.Message)
		}
		if r.Aux != nil && r.Aux.Digest != "" {
			digestResult = r.Aux.Digest
		} else if digestResult == "" && r.Status != "" {
			if match := digestRegexp.FindString(r.Status); match != "" {
				digestResult = match
			}
		}
	}
	return digestResult, nil
}
