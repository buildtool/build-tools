package registry

import (
	"bufio"
	"context"
	"docker.io/go-docker/api/types"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/sparetimecoders/build-tools/pkg/docker"
	"io"
)

type Registry interface {
	Configured() bool
	Name() string
	Login(client docker.Client, out io.Writer) error
	GetAuthInfo() string
	RegistryUrl() string
	Create(repository string) error
	PushImage(client docker.Client, auth, image string, out, eout io.Writer) error
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

func (dockerRegistry) PushImage(client docker.Client, auth, image string, ow, eout io.Writer) error {
	if out, err := client.ImagePush(context.Background(), image, types.ImagePushOptions{All: true, RegistryAuth: auth}); err != nil {
		return err
	} else {
		scanner := bufio.NewScanner(out)
		for scanner.Scan() {
			r := &responsetype{}
			response := scanner.Bytes()
			if err := json.Unmarshal(response, &r); err != nil {
				_, _ = fmt.Fprintf(eout, "Unable to parse response: %s, Error: %v\n", string(response), err)
				return err
			} else {
				if len(r.Status) != 0 {
					if len(r.Id) == 0 {
						_, _ = fmt.Fprintln(ow, r.Status)
					} else if len(r.Progress) == 0 {
						_, _ = fmt.Fprintf(ow, "%s: %s\n", r.Id, r.Status)
					} else {
						_, _ = fmt.Fprintf(ow, "%s: %s %s\n", r.Id, r.Status, r.Progress)
					}
				} else if r.ErrorDetail != nil {
					return errors.New(r.ErrorDetail.Message)
				}
			}
		}

		return nil
	}
}
