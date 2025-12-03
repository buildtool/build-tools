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
	"errors"
	"net/http"
	"testing"
	"time"

	authutil "github.com/containerd/containerd/v2/core/remotes/docker/auth"
	"github.com/docker/docker/api/types/registry"
	auth2 "github.com/moby/buildkit/session/auth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
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

func Test_Credentials_WithPath(t *testing.T) {
	// When registryHost contains a path (e.g., GitLab registry),
	// but the request only contains the hostname
	authWithPath := NewAuthenticator("registry.gitlab.com/org/repo", registry.AuthConfig{
		Username: "gitlab-ci-token",
		Password: "token123",
	})

	// Request with just the hostname should still match
	creds, err := authWithPath.Credentials(context.TODO(), &auth2.CredentialsRequest{Host: "registry.gitlab.com"})
	require.NoError(t, err)
	require.Equal(t, "gitlab-ci-token", creds.Username)
	require.Equal(t, "token123", creds.Secret)

	// Exact match should also work
	authExact := NewAuthenticator("registry.gitlab.com", registry.AuthConfig{
		Username: "user",
		Password: "pass",
	})
	creds, err = authExact.Credentials(context.TODO(), &auth2.CredentialsRequest{Host: "registry.gitlab.com"})
	require.NoError(t, err)
	require.Equal(t, "user", creds.Username)
	require.Equal(t, "pass", creds.Secret)

	// Different registry should not match
	creds, err = authWithPath.Credentials(context.TODO(), &auth2.CredentialsRequest{Host: "ghcr.io"})
	require.NoError(t, err)
	require.Equal(t, "", creds.Username)
	require.Equal(t, "", creds.Secret)
}

func Test_GetRegistryHost(t *testing.T) {
	auth := NewAuthenticator("registry.example.com/org/repo", registry.AuthConfig{
		Username: "user",
		Password: "pass",
	})

	// Type assert to access GetRegistryHost
	a, ok := auth.(*authenticator)
	require.True(t, ok)
	assert.Equal(t, "registry.example.com/org/repo", a.GetRegistryHost())
}

func Test_Register(t *testing.T) {
	auth := NewAuthenticator("registry.example.com", registry.AuthConfig{
		Username: "user",
		Password: "pass",
	})

	server := grpc.NewServer()
	// Should not panic
	auth.Register(server)
	server.Stop()
}

func Test_GetTokenAuthority(t *testing.T) {
	auth := NewAuthenticator("registry.example.com", registry.AuthConfig{
		Username: "user",
		Password: "pass",
	})

	salt := []byte("test-salt-value")
	resp, err := auth.GetTokenAuthority(context.TODO(), &auth2.GetTokenAuthorityRequest{
		Host: "registry.example.com",
		Salt: salt,
	})

	require.NoError(t, err)
	require.NotNil(t, resp)
	// Public key should be 32 bytes (ed25519 public key)
	assert.Len(t, resp.PublicKey, 32)
}

func Test_VerifyTokenAuthority(t *testing.T) {
	auth := NewAuthenticator("registry.example.com", registry.AuthConfig{
		Username: "user",
		Password: "pass",
	})

	salt := []byte("test-salt-value")
	payload := []byte("test-payload")

	resp, err := auth.VerifyTokenAuthority(context.TODO(), &auth2.VerifyTokenAuthorityRequest{
		Host:    "registry.example.com",
		Salt:    salt,
		Payload: payload,
	})

	require.NoError(t, err)
	require.NotNil(t, resp)
	// Signed response should be longer than payload (signature + payload)
	assert.Greater(t, len(resp.Signed), len(payload))
}

func Test_GetTokenAuthority_ConsistentKeys(t *testing.T) {
	auth := NewAuthenticator("registry.example.com", registry.AuthConfig{
		Username: "user",
		Password: "pass",
	})

	salt := []byte("test-salt")

	// Same salt should produce same public key
	resp1, err := auth.GetTokenAuthority(context.TODO(), &auth2.GetTokenAuthorityRequest{
		Host: "registry.example.com",
		Salt: salt,
	})
	require.NoError(t, err)

	resp2, err := auth.GetTokenAuthority(context.TODO(), &auth2.GetTokenAuthorityRequest{
		Host: "registry.example.com",
		Salt: salt,
	})
	require.NoError(t, err)

	assert.Equal(t, resp1.PublicKey, resp2.PublicKey)

	// Different salt should produce different public key
	resp3, err := auth.GetTokenAuthority(context.TODO(), &auth2.GetTokenAuthorityRequest{
		Host: "registry.example.com",
		Salt: []byte("different-salt"),
	})
	require.NoError(t, err)

	assert.NotEqual(t, resp1.PublicKey, resp3.PublicKey)
}

func Test_toTokenResponse(t *testing.T) {
	tests := []struct {
		name      string
		token     string
		issuedAt  time.Time
		expires   int
		wantToken string
		wantExp   int64
		wantIssAt int64
	}{
		{
			name:      "with all fields",
			token:     "test-token",
			issuedAt:  time.Unix(1000, 0),
			expires:   3600,
			wantToken: "test-token",
			wantExp:   3600,
			wantIssAt: 1000,
		},
		{
			name:      "with zero issued at",
			token:     "another-token",
			issuedAt:  time.Time{},
			expires:   7200,
			wantToken: "another-token",
			wantExp:   7200,
			wantIssAt: 0,
		},
		{
			name:      "empty token",
			token:     "",
			issuedAt:  time.Unix(500, 0),
			expires:   0,
			wantToken: "",
			wantExp:   0,
			wantIssAt: 500,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := toTokenResponse(tt.token, tt.issuedAt, tt.expires)

			assert.Equal(t, tt.wantToken, resp.Token)
			assert.Equal(t, tt.wantExp, resp.ExpiresIn)
			assert.Equal(t, tt.wantIssAt, resp.IssuedAt)
		})
	}
}

func Test_getAuthorityKey(t *testing.T) {
	auth := &authenticator{
		registryHost: "registry.example.com",
		authConfig: registry.AuthConfig{
			Username: "user",
			Password: "pass",
		},
	}

	salt := []byte("test-salt")
	key := auth.getAuthorityKey("registry.example.com", salt)

	// ed25519 private key should be 64 bytes
	assert.Len(t, key, 64)

	// Same inputs should produce same key
	key2 := auth.getAuthorityKey("registry.example.com", salt)
	assert.Equal(t, key, key2)

	// Different salt should produce different key
	key3 := auth.getAuthorityKey("registry.example.com", []byte("other-salt"))
	assert.NotEqual(t, key, key3)
}

func Test_FetchToken_Success(t *testing.T) {
	mockFetcher := func(_ context.Context, _ *http.Client, _ http.Header, to authutil.TokenOptions) (*authutil.FetchTokenResponse, error) {
		return &authutil.FetchTokenResponse{
			Token:            "mock-token",
			IssuedAt:         time.Unix(1000, 0),
			ExpiresInSeconds: 3600,
		}, nil
	}

	auth := NewAuthenticatorWithTokenFetcher("registry.example.com", registry.AuthConfig{
		Username: "user",
		Password: "pass",
	}, mockFetcher)

	resp, err := auth.FetchToken(context.TODO(), &auth2.FetchTokenRequest{
		Realm:   "https://registry.example.com/token",
		Service: "registry",
		Scopes:  []string{"repository:image:pull,push"},
	})

	require.NoError(t, err)
	assert.Equal(t, "mock-token", resp.Token)
	assert.Equal(t, int64(3600), resp.ExpiresIn)
	assert.Equal(t, int64(1000), resp.IssuedAt)
}

func Test_FetchToken_WithCredentials(t *testing.T) {
	var capturedOptions authutil.TokenOptions

	mockFetcher := func(_ context.Context, _ *http.Client, _ http.Header, to authutil.TokenOptions) (*authutil.FetchTokenResponse, error) {
		capturedOptions = to
		return &authutil.FetchTokenResponse{
			Token: "auth-token",
		}, nil
	}

	auth := NewAuthenticatorWithTokenFetcher("registry.example.com", registry.AuthConfig{
		Username: "myuser",
		Password: "mypass",
	}, mockFetcher)

	_, err := auth.FetchToken(context.TODO(), &auth2.FetchTokenRequest{
		Realm:   "https://registry.example.com/token",
		Service: "registry",
		Scopes:  []string{"repository:image:pull"},
	})

	require.NoError(t, err)
	assert.Equal(t, "myuser", capturedOptions.Username)
	assert.Equal(t, "mypass", capturedOptions.Secret)
}

func Test_FetchToken_AnonymousFallback(t *testing.T) {
	callCount := 0
	var capturedOptionsOnRetry authutil.TokenOptions

	mockFetcher := func(_ context.Context, _ *http.Client, _ http.Header, to authutil.TokenOptions) (*authutil.FetchTokenResponse, error) {
		callCount++
		if callCount == 1 {
			// First call with credentials fails
			return nil, errors.New("unauthorized")
		}
		// Second call (anonymous) succeeds
		capturedOptionsOnRetry = to
		return &authutil.FetchTokenResponse{
			Token: "anonymous-token",
		}, nil
	}

	auth := NewAuthenticatorWithTokenFetcher("registry.example.com", registry.AuthConfig{
		Username: "user",
		Password: "pass",
	}, mockFetcher)

	resp, err := auth.FetchToken(context.TODO(), &auth2.FetchTokenRequest{
		Realm:   "https://registry.example.com/token",
		Service: "registry",
		Scopes:  []string{"repository:image:pull"},
	})

	require.NoError(t, err)
	assert.Equal(t, "anonymous-token", resp.Token)
	assert.Equal(t, 2, callCount)
	// On retry, credentials should be cleared
	assert.Equal(t, "", capturedOptionsOnRetry.Username)
	assert.Equal(t, "", capturedOptionsOnRetry.Secret)
}

func Test_FetchToken_AllFailures(t *testing.T) {
	mockFetcher := func(_ context.Context, _ *http.Client, _ http.Header, to authutil.TokenOptions) (*authutil.FetchTokenResponse, error) {
		return nil, errors.New("network error")
	}

	auth := NewAuthenticatorWithTokenFetcher("registry.example.com", registry.AuthConfig{
		Username: "user",
		Password: "pass",
	}, mockFetcher)

	_, err := auth.FetchToken(context.TODO(), &auth2.FetchTokenRequest{
		Realm:   "https://registry.example.com/token",
		Service: "registry",
		Scopes:  []string{"repository:image:pull"},
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to fetch token")
}

func Test_FetchToken_NoCredentials(t *testing.T) {
	var capturedOptions authutil.TokenOptions

	mockFetcher := func(_ context.Context, _ *http.Client, _ http.Header, to authutil.TokenOptions) (*authutil.FetchTokenResponse, error) {
		capturedOptions = to
		return &authutil.FetchTokenResponse{
			Token: "anon-token",
		}, nil
	}

	// No credentials provided
	auth := NewAuthenticatorWithTokenFetcher("registry.example.com", registry.AuthConfig{}, mockFetcher)

	resp, err := auth.FetchToken(context.TODO(), &auth2.FetchTokenRequest{
		Realm:   "https://registry.example.com/token",
		Service: "registry",
		Scopes:  []string{"repository:image:pull"},
	})

	require.NoError(t, err)
	assert.Equal(t, "anon-token", resp.Token)
	// Should not have credentials
	assert.Equal(t, "", capturedOptions.Username)
	assert.Equal(t, "", capturedOptions.Secret)
}
