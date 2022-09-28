// MIT License
//
// Copyright (c) 2021 buildtool
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

	"github.com/apex/log"
	"github.com/docker/docker/api/types"

	"github.com/buildtool/build-tools/pkg/docker"
)

type Registry interface {
	Configured() bool
	Name() string
	Login(client docker.Client) error
	GetAuthConfig() types.AuthConfig
	GetAuthInfo() string
	RegistryUrl() string
	Create(repository string) error
	PushImage(client docker.Client, auth, image string) error
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

type dockerRegistry struct{}

func (dockerRegistry) PushImage(client docker.Client, auth, image string) error {
	if out, err := client.ImagePush(context.Background(), image, types.ImagePushOptions{All: true, RegistryAuth: auth}); err != nil {
		return err
	} else {
		scanner := bufio.NewScanner(out)
		for scanner.Scan() {
			r := &responsetype{}
			response := scanner.Bytes()
			if err := json.Unmarshal(response, &r); err != nil {
				log.Errorf("Unable to parse response: %s, Error: %v\n", string(response), err)
				return err
			} else {
				if r.ErrorDetail != nil {
					return errors.New(r.ErrorDetail.Message)
				}
			}
		}

		return nil
	}
}
