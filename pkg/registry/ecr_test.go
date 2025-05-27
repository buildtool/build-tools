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
	"context"
	"fmt"
	"testing"

	"github.com/apex/log"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecr"
	"github.com/aws/aws-sdk-go-v2/service/ecr/types"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/stretchr/testify/assert"
	mocks "gitlab.com/unboundsoftware/apex-mocks"

	"github.com/buildtool/build-tools/pkg/docker"
)

func TestEcr_LoginAuthRequestFailed(t *testing.T) {
	client := &docker.MockDocker{}
	registry := &ECR{Url: "ecr-url", Region: "eu-west-1", ecrSvc: &MockECR{loginError: fmt.Errorf("auth failure")}}
	logMock := mocks.New()
	log.SetHandler(logMock)
	log.SetLevel(log.DebugLevel)
	err := registry.Login(client)
	assert.EqualError(t, err, "auth failure")
	logMock.Check(t, []string{})
}

func TestEcr_LoginInvalidAuthData(t *testing.T) {
	client := &docker.MockDocker{}
	registry := &ECR{Url: "ecr-url", Region: "eu-west-1", ecrSvc: &MockECR{authData: "aaabbb"}}
	logMock := mocks.New()
	log.SetHandler(logMock)
	log.SetLevel(log.DebugLevel)
	err := registry.Login(client)
	assert.EqualError(t, err, "illegal base64 data at input byte 4")
	logMock.Check(t, []string{})
}

func TestEcr_LoginFailed(t *testing.T) {
	client := &docker.MockDocker{LoginError: fmt.Errorf("invalid username/password")}
	registry := &ECR{Url: "ecr-url", Region: "eu-west-1", ecrSvc: &MockECR{authData: "QVdTOmFiYzEyMw=="}}
	logMock := mocks.New()
	log.SetHandler(logMock)
	log.SetLevel(log.DebugLevel)
	err := registry.Login(client)
	assert.EqualError(t, err, "invalid username/password")
	logMock.Check(t, []string{})
}

func TestEcr_LoginSuccess(t *testing.T) {
	client := &docker.MockDocker{}
	registry := &ECR{Url: "ecr-url", Region: "eu-west-1", ecrSvc: &MockECR{authData: "QVdTOmFiYzEyMw=="}}
	logMock := mocks.New()
	log.SetHandler(logMock)
	log.SetLevel(log.DebugLevel)
	err := registry.Login(client)
	assert.Nil(t, err)
	assert.Equal(t, "AWS", client.Username)
	assert.Equal(t, "abc123", client.Password)
	assert.Equal(t, "ecr-url", client.ServerAddress)
	logMock.Check(t, []string{"debug: Logged in\n"})
}

func TestEcr_GetAuthInfo(t *testing.T) {
	registry := &ECR{Url: "ecr-url", Region: "eu-west-1", username: "AWS", password: "abc123"}
	auth := registry.GetAuthInfo()
	assert.Equal(t, "eyJ1c2VybmFtZSI6IkFXUyIsInBhc3N3b3JkIjoiYWJjMTIzIn0=", auth)
}

func TestEcr_RegistryAndClientInDifferentAccounts(t *testing.T) {
	registryId := "repo"
	mockECR := &MockECR{repoAccessNotAllowed: true}
	mockSTS := &MockSTS{userAccountId: aws.String("1234")}
	registry := &ECR{ecrSvc: mockECR, stsSvc: mockSTS}
	registry.registryId = &registryId
	repo := "repo"
	err := registry.Create(repo)
	assert.EqualError(t, err, "account mismatch, logged in at '1234' got 'repo' from repository url ")
}

func TestEcr_RepositoryAccessNotAllowed(t *testing.T) {
	registryId := "1234"
	mockECR := &MockECR{repoExists: true, repoAccessNotAllowed: true}
	mockSTS := &MockSTS{userAccountId: &registryId}
	registry := &ECR{ecrSvc: mockECR, stsSvc: mockSTS, registryId: &registryId}
	repo := "repo"
	err := registry.Create(repo)
	assert.EqualError(t, err, "not allowed")
}

func TestEcr_ExistingRepository(t *testing.T) {
	registryId := "1234"
	mockECR := &MockECR{repoExists: true}
	mockSTS := &MockSTS{userAccountId: &registryId}
	registry := &ECR{ecrSvc: mockECR, stsSvc: mockSTS, registryId: &registryId}
	repo := "repo"
	err := registry.Create(repo)
	assert.Nil(t, err)
	assert.Equal(t, []string{repo}, mockECR.describeRepositoriesInput.RepositoryNames)
	assert.Equal(t, &registryId, mockECR.describeRepositoriesInput.RegistryId)
}

func TestEcr_NewRepositoryCreateError(t *testing.T) {
	registryId := "1234"
	mockSTS := &MockSTS{userAccountId: &registryId}
	mockECR := &MockECR{createError: fmt.Errorf("create error"), repoExists: false}
	registry := &ECR{ecrSvc: mockECR, stsSvc: mockSTS, registryId: &registryId}
	err := registry.Create("repo")
	assert.EqualError(t, err, "create error")
}

func TestEcr_NewRepositoryPutError(t *testing.T) {
	registryId := "1234"
	mockSTS := &MockSTS{userAccountId: &registryId}
	registry := &ECR{ecrSvc: &MockECR{putError: fmt.Errorf("put error")}, stsSvc: mockSTS, registryId: &registryId}
	err := registry.Create("repo")
	assert.EqualError(t, err, "put error")
}

func TestEcr_NewRepository(t *testing.T) {
	registryId := "1234"
	mockSTS := &MockSTS{userAccountId: &registryId}
	mockECR := &MockECR{}
	registry := &ECR{ecrSvc: mockECR, stsSvc: mockSTS, registryId: &registryId}
	repo := "repo"
	err := registry.Create(repo)
	assert.Nil(t, err)
	assert.Equal(t, &repo, mockECR.createRepositoryInput.RepositoryName)
	policyText := `{"rules":[{"rulePriority":10,"description":"Only keep 20 images","selection":{"tagStatus":"untagged","countType":"imageCountMoreThan","countNumber":20},"action":{"type":"expire"}}]}`
	assert.Equal(t, &policyText, mockECR.putLifecyclePolicyInput.LifecyclePolicyText)
}

func TestEcr_ParseECRUrlIfNoRegionIsSet(t *testing.T) {
	ecr := ECR{
		Url: "12345678.dkr.ecr.eu-west-1.amazonaws.com",
	}
	assert.Equal(t, "eu-west-1", *ecr.region())
}

func TestEcr_UseRegionIfSet(t *testing.T) {
	ecr := ECR{
		Url:    "12345678.dkr.ecr.eu-west-1.amazonaws.com",
		Region: "region",
	}
	assert.Equal(t, "region", *ecr.region())
}

func TestEcr_ParseECRUrlRepositoryId(t *testing.T) {
	ecr := ECR{
		Url: "12345678.dkr.ecr.eu-west-1.amazonaws.com",
	}
	registry, err := ecr.registry()
	assert.Nil(t, err)
	assert.Equal(t, "12345678", *registry)
}

func TestEcr_ParseInvalidECRUrlRepositoryId(t *testing.T) {
	ecr := ECR{
		Url: "12345678.ecr.eu-west-1.amazonaws.com",
	}
	_, err := ecr.registry()
	assert.EqualError(t, err, "failed to extract registryid from string 12345678.ecr.eu-west-1.amazonaws.com")
}

type MockSTS struct {
	STSClient
	userAccountId *string
}

type MockECR struct {
	ECRClient
	loginError                error
	authData                  string
	describeRepositoriesInput *ecr.DescribeRepositoriesInput
	repoExists                bool
	createError               error
	createRepositoryInput     *ecr.CreateRepositoryInput
	putLifecyclePolicyInput   *ecr.PutLifecyclePolicyInput
	putError                  error
	repoAccessNotAllowed      bool
}

func (r *MockECR) GetAuthorizationToken(_ context.Context, input *ecr.GetAuthorizationTokenInput, optFns ...func(*ecr.Options)) (*ecr.GetAuthorizationTokenOutput, error) {
	if r.loginError != nil {
		return &ecr.GetAuthorizationTokenOutput{AuthorizationData: []types.AuthorizationData{}}, r.loginError
	}
	return &ecr.GetAuthorizationTokenOutput{AuthorizationData: []types.AuthorizationData{{AuthorizationToken: &r.authData}}}, nil
}

func (r *MockECR) DescribeRepositories(_ context.Context, input *ecr.DescribeRepositoriesInput, optFns ...func(*ecr.Options)) (*ecr.DescribeRepositoriesOutput, error) {
	r.describeRepositoriesInput = input
	if r.repoAccessNotAllowed {
		return &ecr.DescribeRepositoriesOutput{Repositories: []types.Repository{}}, fmt.Errorf("not allowed")
	}
	if r.repoExists {
		return &ecr.DescribeRepositoriesOutput{Repositories: []types.Repository{}}, nil
	}
	return &ecr.DescribeRepositoriesOutput{Repositories: []types.Repository{}},
		&types.RepositoryNotFoundException{}
}

func (r *MockECR) CreateRepository(_ context.Context, input *ecr.CreateRepositoryInput, optFns ...func(*ecr.Options)) (*ecr.CreateRepositoryOutput, error) {
	r.createRepositoryInput = input
	return &ecr.CreateRepositoryOutput{}, r.createError
}

func (r *MockECR) PutLifecyclePolicy(_ context.Context, input *ecr.PutLifecyclePolicyInput, optFns ...func(*ecr.Options)) (*ecr.PutLifecyclePolicyOutput, error) {
	r.putLifecyclePolicyInput = input
	return &ecr.PutLifecyclePolicyOutput{}, r.putError
}

func (r *MockSTS) GetCallerIdentity(context.Context, *sts.GetCallerIdentityInput, ...func(*sts.Options)) (*sts.GetCallerIdentityOutput, error) {
	if r.userAccountId != nil {
		return &sts.GetCallerIdentityOutput{Account: r.userAccountId}, nil
	}
	return nil, fmt.Errorf("failed to get caller identity")
}
