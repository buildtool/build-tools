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
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/apex/log"
	"github.com/buildtool/build-tools/pkg/docker"
	"github.com/docker/docker/api/types/registry"
)

type Credential interface {
	GetToken(ctx context.Context, opts policy.TokenRequestOptions) (azcore.AccessToken, error)
}

type ACR struct {
	dockerRegistry `yaml:"-"`
	TenantId       string `yaml:"tenantId" env:"ACR_TENANT_ID"`
	Url            string `yaml:"url" env:"ACR_URL"`
	token          string
	credential     Credential
}

func (r *ACR) Configured() bool {
	if len(r.Url) > 0 {
		credential, err := azidentity.NewDefaultAzureCredential(nil)
		if err != nil {
			log.Errorf("Error while getting default azure credential: %v\n", err)
			return false
		}
		r.credential = credential
		return true
	}
	return false
}

func (r *ACR) Name() string {
	return "ACR"
}

func (r *ACR) Login(client docker.Client) error {
	token, err := r.credential.GetToken(context.Background(), policy.TokenRequestOptions{
		Scopes: []string{"https://management.azure.com/.default"},
	})
	if err != nil {
		return err
	}
	formData := url.Values{
		"grant_type":   {"access_token"},
		"service":      {r.Url},
		"tenant":       {r.TenantId},
		"access_token": {token.Token},
	}
	jsonResponse, err := http.PostForm(fmt.Sprintf("https://%s/oauth2/exchange", r.Url), formData)
	if err != nil {
		return err
	}
	var response map[string]interface{}
	err = json.NewDecoder(jsonResponse.Body).Decode(&response)
	if err != nil {
		return err
	}
	r.token = response["refresh_token"].(string)
	ok, err := client.RegistryLogin(context.Background(), registry.AuthConfig{
		Username:      "00000000-0000-0000-0000-000000000000",
		Password:      r.token,
		ServerAddress: r.Url,
	})
	if err != nil {
		return err
	}
	log.Debugf("Status: %s\n", ok.Status)
	return nil
}

func (r *ACR) GetAuthConfig() registry.AuthConfig {
	return registry.AuthConfig{
		Username:      "00000000-0000-0000-0000-000000000000",
		Password:      r.token,
		ServerAddress: r.Url,
	}
}

func (r *ACR) GetAuthInfo() string {
	authBytes, _ := json.Marshal(r.GetAuthConfig())
	return base64.URLEncoding.EncodeToString(authBytes)
}

func (r *ACR) RegistryUrl() string {
	return r.Url
}

func (r *ACR) Create(_ string) error {
	return nil
}

var _ Registry = &ACR{}
