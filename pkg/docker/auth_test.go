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

package docker

import (
	"context"
	"testing"

	"github.com/docker/docker/api/types/registry"
	auth2 "github.com/moby/buildkit/session/auth"
	"github.com/stretchr/testify/require"
)

func Test_Credentials(t *testing.T) {
	auth := NewAuthenticator("use-auth.com", registry.AuthConfig{
		Username: "user",
		Password: "password",
	})
	anonymousCreds, err := auth.Credentials(context.TODO(), &auth2.CredentialsRequest{Host: "docker.io"})
	require.NoError(t, err)
	require.Equal(t, "", anonymousCreds.Username)
	require.Equal(t, "", anonymousCreds.Secret)

	creds, err := auth.Credentials(context.TODO(), &auth2.CredentialsRequest{Host: "use-auth.com"})
	require.NoError(t, err)
	require.Equal(t, "user", creds.Username)
	require.Equal(t, "password", creds.Secret)

}
