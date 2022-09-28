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

package docker

import (
	"context"
	"crypto/ed25519"
	"crypto/hmac"
	"crypto/sha256"
	"fmt"
	"net/http"
	"time"

	authutil "github.com/containerd/containerd/remotes/docker/auth"

	"github.com/docker/docker/api/types"
	"github.com/moby/buildkit/session"
	"github.com/moby/buildkit/session/auth"
	"golang.org/x/crypto/nacl/sign"
	"google.golang.org/grpc"
)

type authenticator struct {
	authConfig   types.AuthConfig
	registryHost string
}

func NewAuthenticator(registryHost string, authConfig types.AuthConfig) Authenticator {
	return &authenticator{
		authConfig:   authConfig,
		registryHost: registryHost,
	}
}

func (a *authenticator) Register(server *grpc.Server) {
	auth.RegisterAuthServer(server, a)
}

func (a authenticator) Credentials(ctx context.Context, req *auth.CredentialsRequest) (*auth.CredentialsResponse, error) {
	if req.Host != a.registryHost {
		return &auth.CredentialsResponse{}, nil
	}
	return &auth.CredentialsResponse{Username: a.authConfig.Username, Secret: a.authConfig.Password}, nil
}

func (a authenticator) FetchToken(ctx context.Context, req *auth.FetchTokenRequest) (*auth.FetchTokenResponse, error) {
	to := authutil.TokenOptions{
		Realm:    req.Realm,
		Service:  req.Service,
		Scopes:   req.Scopes,
		Username: "",
		Secret:   "",
	}
	// do request anonymously
	resp, err := authutil.FetchToken(ctx, http.DefaultClient, nil, to)
	if err != nil {
		// try with auth
		to.Username = a.authConfig.Username
		to.Secret = a.authConfig.Password
		resp, err = authutil.FetchToken(ctx, http.DefaultClient, nil, to)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch anonymous and authenticated token, %w", err)
		}
	}
	return toTokenResponse(resp.Token, resp.IssuedAt, resp.ExpiresIn), nil
}

func (ap *authenticator) GetTokenAuthority(ctx context.Context, req *auth.GetTokenAuthorityRequest) (*auth.GetTokenAuthorityResponse, error) {
	key := ap.getAuthorityKey(req.Host, req.Salt)

	return &auth.GetTokenAuthorityResponse{PublicKey: key[32:]}, nil
}

func (ap *authenticator) VerifyTokenAuthority(ctx context.Context, req *auth.VerifyTokenAuthorityRequest) (*auth.VerifyTokenAuthorityResponse, error) {
	key := ap.getAuthorityKey(req.Host, req.Salt)

	priv := new([64]byte)
	copy((*priv)[:], key)

	return &auth.VerifyTokenAuthorityResponse{Signed: sign.Sign(nil, req.Payload, priv)}, nil
}

var _ Authenticator = &authenticator{}

type Authenticator interface {
	auth.AuthServer
	session.Attachable
}

func toTokenResponse(token string, issuedAt time.Time, expires int) *auth.FetchTokenResponse {
	resp := &auth.FetchTokenResponse{
		Token:     token,
		ExpiresIn: int64(expires),
	}
	if !issuedAt.IsZero() {
		resp.IssuedAt = issuedAt.Unix()
	}
	return resp
}

func (ap *authenticator) getAuthorityKey(host string, salt []byte) ed25519.PrivateKey {
	mac := hmac.New(sha256.New, salt)
	sum := mac.Sum(nil)

	return ed25519.NewKeyFromSeed(sum[:ed25519.SeedSize])
}
