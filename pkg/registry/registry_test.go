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
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/buildtool/build-tools/pkg/docker"
)

func TestDockerRegistry_PushImage(t *testing.T) {
	registry := &Gitlab{}
	client := &docker.MockDocker{PushError: errors.New("error")}

	digest, err := registry.PushImage(client, "dummy", "unknown")
	assert.EqualError(t, err, "error")
	assert.Empty(t, digest)
}

func TestDockerRegistry_PushImage_ReturnsDigest(t *testing.T) {
	pushOut := `{"status":"Pushing"}
{"progressDetail":{},"aux":{"Tag":"v1","Digest":"sha256:af534ee896ce2ac80f3413318329e45e3b3e74b89eb337b9364b8ac1e83498b7","Size":2828}}`
	registry := &Gitlab{}
	client := &docker.MockDocker{PushOutput: &pushOut}

	digest, err := registry.PushImage(client, "dummy", "image:v1")
	assert.NoError(t, err)
	assert.Equal(t, "sha256:af534ee896ce2ac80f3413318329e45e3b3e74b89eb337b9364b8ac1e83498b7", digest)
}

func TestDockerRegistry_PushImage_NoDigest(t *testing.T) {
	pushOut := `{"status":"Push successful"}`
	registry := &Gitlab{}
	client := &docker.MockDocker{PushOutput: &pushOut}

	digest, err := registry.PushImage(client, "dummy", "image:v1")
	assert.NoError(t, err)
	assert.Empty(t, digest)
}
