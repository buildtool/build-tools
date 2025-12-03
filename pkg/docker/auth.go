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
	"crypto/ed25519"
	"crypto/hmac"
	"crypto/sha256"
	"fmt"
	"net/http"
	"strings"
	"time"

	authutil "github.com/containerd/containerd/v2/core/remotes/docker/auth"

	"github.com/docker/docker/api/types/registry"
	"github.com/moby/buildkit/session"
	"github.com/moby/buildkit/session/auth"
	"golang.org/x/crypto/nacl/sign"
	"google.golang.org/grpc"
)

// TokenFetcher is a function type for fetching tokens from a registry.
// This allows injecting mock implementations for testing.
type TokenFetcher func(ctx context.Context, client *http.Client, headers http.Header, to authutil.TokenOptions) (*authutil.FetchTokenResponse, error)

// defaultTokenFetcher is the real implementation that calls authutil.FetchToken.
var defaultTokenFetcher TokenFetcher = authutil.FetchToken

type authenticator struct {
	authConfig   registry.AuthConfig
	registryHost string
	tokenFetcher TokenFetcher
}

// NewAuthenticator creates a new authenticator with the default token fetcher.
func NewAuthenticator(registryHost string, authConfig registry.AuthConfig) Authenticator {
	return NewAuthenticatorWithTokenFetcher(registryHost, authConfig, defaultTokenFetcher)
}

// NewAuthenticatorWithTokenFetcher creates a new authenticator with a custom token fetcher.
func NewAuthenticatorWithTokenFetcher(registryHost string, authConfig registry.AuthConfig, tokenFetcher TokenFetcher) Authenticator {
	return &authenticator{
		authConfig:   authConfig,
		registryHost: registryHost,
		tokenFetcher: tokenFetcher,
	}
}

func (a *authenticator) Register(server *grpc.Server) {
	auth.RegisterAuthServer(server, a)
}

func (a authenticator) Credentials(ctx context.Context, req *auth.CredentialsRequest) (*auth.CredentialsResponse, error) {
	// Check if the request host matches the registry host.
	// The registryHost may contain a path (e.g., "registry.gitlab.com/org/repo")
	// while req.Host only contains the hostname (e.g., "registry.gitlab.com").
	// We check both exact match and prefix match to handle both cases.
	if req.Host != a.registryHost && !strings.HasPrefix(a.registryHost, req.Host+"/") && !strings.HasPrefix(a.registryHost, req.Host) {
		return &auth.CredentialsResponse{}, nil
	}
	return &auth.CredentialsResponse{Username: a.authConfig.Username, Secret: a.authConfig.Password}, nil
}

// GetRegistryHost returns the configured registry host for this authenticator.
func (a authenticator) GetRegistryHost() string {
	return a.registryHost
}

func (a authenticator) FetchToken(ctx context.Context, req *auth.FetchTokenRequest) (*auth.FetchTokenResponse, error) {
	to := authutil.TokenOptions{
		Realm:   req.Realm,
		Service: req.Service,
		Scopes:  req.Scopes,
	}

	// Always use credentials when available to ensure proper scopes (especially for push)
	if a.authConfig.Username != "" && a.authConfig.Password != "" {
		to.Username = a.authConfig.Username
		to.Secret = a.authConfig.Password
	}

	// Use the injected token fetcher
	fetcher := a.tokenFetcher
	if fetcher == nil {
		fetcher = defaultTokenFetcher
	}

	resp, err := fetcher(ctx, http.DefaultClient, nil, to)
	if err != nil {
		// If authenticated request failed and we had credentials, try anonymous as fallback
		if to.Username != "" {
			to.Username = ""
			to.Secret = ""
			resp, err = fetcher(ctx, http.DefaultClient, nil, to)
		}
		if err != nil {
			return nil, fmt.Errorf("failed to fetch token: %w", err)
		}
	}
	return toTokenResponse(resp.Token, resp.IssuedAt, resp.ExpiresInSeconds), nil
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
