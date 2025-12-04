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

	"github.com/apex/log"
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

// registryCredentials holds credentials for a specific registry host.
type registryCredentials struct {
	host       string
	authConfig registry.AuthConfig
}

type authenticator struct {
	credentials  []registryCredentials
	tokenFetcher TokenFetcher
}

// NewAuthenticator creates a new authenticator with the default token fetcher.
func NewAuthenticator(registryHost string, authConfig registry.AuthConfig) Authenticator {
	return NewAuthenticatorWithTokenFetcher(registryHost, authConfig, defaultTokenFetcher)
}

// NewAuthenticatorWithTokenFetcher creates a new authenticator with a custom token fetcher.
func NewAuthenticatorWithTokenFetcher(registryHost string, authConfig registry.AuthConfig, tokenFetcher TokenFetcher) Authenticator {
	return &authenticator{
		credentials: []registryCredentials{
			{host: registryHost, authConfig: authConfig},
		},
		tokenFetcher: tokenFetcher,
	}
}

// AddCredentials adds additional registry credentials to the authenticator.
// This is useful when pushing to multiple registries (e.g., image registry and cache registry).
func (a *authenticator) AddCredentials(registryHost string, authConfig registry.AuthConfig) {
	log.Debugf("AddCredentials: adding credentials for host=%s (username=%s)\n", registryHost, authConfig.Username)
	a.credentials = append(a.credentials, registryCredentials{
		host:       registryHost,
		authConfig: authConfig,
	})
}

func (a *authenticator) Register(server *grpc.Server) {
	auth.RegisterAuthServer(server, a)
}

// findCredentials finds the credentials for the given host.
// The registryHost may contain a path (e.g., "registry.gitlab.com/org/repo")
// while reqHost only contains the hostname (e.g., "registry.gitlab.com").
// We check exact match first, then prefix match with path separator.
func (a authenticator) findCredentials(reqHost string) *registryCredentials {
	for i := range a.credentials {
		cred := &a.credentials[i]
		// Exact match
		if reqHost == cred.host {
			return cred
		}
		// reqHost is the base hostname and cred.host has a path
		// e.g., reqHost="registry.gitlab.com" matches cred.host="registry.gitlab.com/org/repo"
		if strings.HasPrefix(cred.host, reqHost+"/") {
			return cred
		}
		// cred.host is the base hostname and reqHost has a path
		// e.g., cred.host="registry.gitlab.com" matches reqHost="registry.gitlab.com/org/repo"
		if strings.HasPrefix(reqHost, cred.host+"/") {
			return cred
		}
	}
	return nil
}

func (a authenticator) Credentials(ctx context.Context, req *auth.CredentialsRequest) (*auth.CredentialsResponse, error) {
	log.Debugf("Credentials request: Host=%s (have %d credential sets)\n", req.Host, len(a.credentials))
	for i, c := range a.credentials {
		log.Debugf("  credential[%d]: host=%s, username=%s, hasPassword=%v\n", i, c.host, c.authConfig.Username, c.authConfig.Password != "")
	}
	cred := a.findCredentials(req.Host)
	if cred == nil {
		log.Debugf("Credentials: no credentials found for host=%s\n", req.Host)
		return &auth.CredentialsResponse{}, nil
	}
	log.Debugf("Credentials: returning username=%s, hasSecret=%v for host=%s (matched %s)\n", cred.authConfig.Username, cred.authConfig.Password != "", req.Host, cred.host)
	return &auth.CredentialsResponse{Username: cred.authConfig.Username, Secret: cred.authConfig.Password}, nil
}

// GetRegistryHost returns the first configured registry host for this authenticator.
func (a authenticator) GetRegistryHost() string {
	if len(a.credentials) > 0 {
		return a.credentials[0].host
	}
	return ""
}

func (a authenticator) FetchToken(ctx context.Context, req *auth.FetchTokenRequest) (*auth.FetchTokenResponse, error) {
	log.Debugf("FetchToken request: Host=%s, Realm=%s, Service=%s, Scopes=%v\n", req.Host, req.Realm, req.Service, req.Scopes)

	to := authutil.TokenOptions{
		Realm:   req.Realm,
		Service: req.Service,
		Scopes:  req.Scopes,
	}

	// Find credentials for this service or host
	// Try service first, then fall back to host (some registries use different service names)
	cred := a.findCredentials(req.Service)
	if cred == nil && req.Host != "" && req.Host != req.Service {
		cred = a.findCredentials(req.Host)
	}
	if cred != nil && cred.authConfig.Username != "" && cred.authConfig.Password != "" {
		log.Debugf("FetchToken: using credentials for host=%s\n", cred.host)
		to.Username = cred.authConfig.Username
		to.Secret = cred.authConfig.Password
	} else {
		log.Debugf("FetchToken: no credentials found for service=%s or host=%s\n", req.Service, req.Host)
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
	// AddCredentials adds additional registry credentials for authentication.
	AddCredentials(registryHost string, authConfig registry.AuthConfig)
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
