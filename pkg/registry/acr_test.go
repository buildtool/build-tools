// MIT License
//
// Copyright (c) 2025 buildtool
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
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/apex/log"
	"github.com/stretchr/testify/assert"
	mocks "gitlab.com/unboundsoftware/apex-mocks"

	"github.com/buildtool/build-tools/pkg/docker"
)

func TestAcr_LoginTokenRequestFailed(t *testing.T) {
	client := &docker.MockDocker{}
	registry := &ACR{Url: "ecr-url", credential: &MockCredential{
		getToken: func(ctx context.Context, opts policy.TokenRequestOptions) (azcore.AccessToken, error) {
			assert.Equal(t, []string{"https://management.azure.com/.default"}, opts.Scopes)
			return azcore.AccessToken{}, fmt.Errorf("auth failure")
		},
	}}
	logMock := mocks.New()
	log.SetHandler(logMock)
	log.SetLevel(log.DebugLevel)
	err := registry.Login(client)
	assert.EqualError(t, err, "auth failure")
	logMock.Check(t, []string{})
}

func TestAcr_LoginTokenExchangeFailed(t *testing.T) {
	client := &docker.MockDocker{}
	registry := &ACR{Url: "some-non-existing-site.nosuchtld", credential: &MockCredential{
		getToken: func(ctx context.Context, opts policy.TokenRequestOptions) (azcore.AccessToken, error) {
			return azcore.AccessToken{Token: "aaabbb"}, nil
		},
	}}
	logMock := mocks.New()
	log.SetHandler(logMock)
	log.SetLevel(log.DebugLevel)
	err := registry.Login(client)
	assert.ErrorContains(t, err, "no such host")
	logMock.Check(t, []string{})
}

func TestAcr_LoginInvalidExchangeResponse(t *testing.T) {
	srv := httptest.NewTLSServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		_, _ = writer.Write([]byte(`{refresh_token":"aaabbb"`))
	}))
	defer srv.Close()
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{
		RootCAs:            srv.TLS.RootCAs,
		InsecureSkipVerify: true,
	}
	client := &docker.MockDocker{LoginError: fmt.Errorf("invalid username/password")}
	registry := &ACR{Url: strings.TrimPrefix(srv.URL, "https://"), credential: &MockCredential{
		getToken: func(ctx context.Context, opts policy.TokenRequestOptions) (azcore.AccessToken, error) {
			return azcore.AccessToken{Token: "QVdTOmFiYzEyMw=="}, nil
		},
	}}
	logMock := mocks.New()
	log.SetHandler(logMock)
	log.SetLevel(log.DebugLevel)
	err := registry.Login(client)
	assert.EqualError(t, err, "invalid character 'r' looking for beginning of object key string")
	logMock.Check(t, []string{})
}

func TestAcr_LoginFailed(t *testing.T) {
	srv := httptest.NewTLSServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		_, _ = writer.Write([]byte(`{"refresh_token":"aaabbb"}`))
	}))
	defer srv.Close()
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{
		RootCAs:            srv.TLS.RootCAs,
		InsecureSkipVerify: true,
	}
	client := &docker.MockDocker{LoginError: fmt.Errorf("invalid username/password")}
	registry := &ACR{Url: strings.TrimPrefix(srv.URL, "https://"), credential: &MockCredential{
		getToken: func(ctx context.Context, opts policy.TokenRequestOptions) (azcore.AccessToken, error) {
			return azcore.AccessToken{Token: "QVdTOmFiYzEyMw=="}, nil
		},
	}}
	logMock := mocks.New()
	log.SetHandler(logMock)
	log.SetLevel(log.DebugLevel)
	err := registry.Login(client)
	assert.EqualError(t, err, "invalid username/password")
	logMock.Check(t, []string{})
}

func TestAcr_LoginSuccess(t *testing.T) {
	srv := httptest.NewTLSServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		_, _ = writer.Write([]byte(`{"refresh_token":"aaabbb"}`))
	}))
	defer srv.Close()
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{
		RootCAs:            srv.TLS.RootCAs,
		InsecureSkipVerify: true,
	}
	client := &docker.MockDocker{}
	registryUrl := strings.TrimPrefix(srv.URL, "https://")
	registry := &ACR{Url: registryUrl, credential: &MockCredential{
		getToken: func(ctx context.Context, opts policy.TokenRequestOptions) (azcore.AccessToken, error) {
			return azcore.AccessToken{Token: "QVdTOmFiYzEyMw=="}, nil
		},
	}}
	logMock := mocks.New()
	log.SetHandler(logMock)
	log.SetLevel(log.DebugLevel)
	err := registry.Login(client)
	assert.Nil(t, err)
	assert.Equal(t, "00000000-0000-0000-0000-000000000000", client.Username)
	assert.Equal(t, "aaabbb", client.Password)
	assert.Equal(t, registryUrl, client.ServerAddress)
	logMock.Check(t, []string{"debug: Status: Logged in\n"})
}

func TestAcr_GetAuthInfo(t *testing.T) {
	registry := &ACR{Url: "ecr-url", token: "aaabbb"}
	auth := registry.GetAuthInfo()
	assert.Equal(t, "eyJ1c2VybmFtZSI6IjAwMDAwMDAwLTAwMDAtMDAwMC0wMDAwLTAwMDAwMDAwMDAwMCIsInBhc3N3b3JkIjoiYWFhYmJiIiwic2VydmVyYWRkcmVzcyI6ImVjci11cmwifQ==", auth)
}

func TestAcr_ParseECRUrlRepositoryId(t *testing.T) {
	ecr := ECR{
		Url: "12345678.dkr.ecr.eu-west-1.amazonaws.com",
	}
	registry, err := ecr.registry()
	assert.Nil(t, err)
	assert.Equal(t, "12345678", *registry)
}

func TestAcr_ParseInvalidECRUrlRepositoryId(t *testing.T) {
	ecr := ECR{
		Url: "12345678.ecr.eu-west-1.amazonaws.com",
	}
	_, err := ecr.registry()
	assert.EqualError(t, err, "failed to extract registryid from string 12345678.ecr.eu-west-1.amazonaws.com")
}

type MockCredential struct {
	getToken func(ctx context.Context, opts policy.TokenRequestOptions) (azcore.AccessToken, error)
}

func (m MockCredential) GetToken(ctx context.Context, opts policy.TokenRequestOptions) (azcore.AccessToken, error) {
	if m.getToken != nil {
		return m.getToken(ctx, opts)
	}
	return azcore.AccessToken{}, nil
}

var _ Credential = &MockCredential{}
