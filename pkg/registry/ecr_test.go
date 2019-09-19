package registry

import (
	"bytes"
	"fmt"
	awsecr "github.com/aws/aws-sdk-go/service/ecr"
	"github.com/aws/aws-sdk-go/service/ecr/ecriface"
	"github.com/stretchr/testify/assert"
	"gitlab.com/sparetimecoders/build-tools/pkg/docker"
	"testing"
)

func TestEcr_LoginAuthRequestFailed(t *testing.T) {
	client := &docker.MockDocker{}
	registry := &ECR{Url: "ecr-url", Region: "eu-west-1", svc: &MockECR{loginError: fmt.Errorf("auth failure")}}
	out := &bytes.Buffer{}
	err := registry.Login(client, out)
	assert.EqualError(t, err, "auth failure")
	assert.Equal(t, "", out.String())
}

func TestEcr_LoginInvalidAuthData(t *testing.T) {
	client := &docker.MockDocker{}
	registry := &ECR{Url: "ecr-url", Region: "eu-west-1", svc: &MockECR{authData: "aaabbb"}}
	out := &bytes.Buffer{}
	err := registry.Login(client, out)
	assert.EqualError(t, err, "illegal base64 data at input byte 4")
	assert.Equal(t, "", out.String())
}

func TestEcr_LoginFailed(t *testing.T) {
	client := &docker.MockDocker{LoginError: fmt.Errorf("invalid username/password")}
	registry := &ECR{Url: "ecr-url", Region: "eu-west-1", svc: &MockECR{authData: "QVdTOmFiYzEyMw=="}}
	out := &bytes.Buffer{}
	err := registry.Login(client, out)
	assert.EqualError(t, err, "invalid username/password")
	assert.Equal(t, "", out.String())
}

func TestEcr_LoginSuccess(t *testing.T) {
	client := &docker.MockDocker{}
	registry := &ECR{Url: "ecr-url", Region: "eu-west-1", svc: &MockECR{authData: "QVdTOmFiYzEyMw=="}}
	out := &bytes.Buffer{}
	err := registry.Login(client, out)
	assert.Nil(t, err)
	assert.Equal(t, "AWS", client.Username)
	assert.Equal(t, "abc123", client.Password)
	assert.Equal(t, "ecr-url", client.ServerAddress)
	assert.Equal(t, "Logged in\n", out.String())
}

func TestEcr_GetAuthInfo(t *testing.T) {
	registry := &ECR{Url: "ecr-url", Region: "eu-west-1", username: "AWS", password: "abc123"}
	auth := registry.GetAuthInfo()
	assert.Equal(t, "eyJ1c2VybmFtZSI6IkFXUyIsInBhc3N3b3JkIjoiYWJjMTIzIn0=", auth)
}

func TestEcr_ExistingRepository(t *testing.T) {
	mock := &MockECR{repoExists: true}
	registry := &ECR{svc: mock}
	repo := "repo"
	err := registry.Create(repo)
	assert.Nil(t, err)
	assert.Equal(t, []*string{&repo}, mock.describeRepositoriesInput.RepositoryNames)
}

func TestEcr_NewRepositoryCreateError(t *testing.T) {
	registry := &ECR{svc: &MockECR{createError: fmt.Errorf("create error")}}
	err := registry.Create("repo")
	assert.EqualError(t, err, "create error")
}

func TestEcr_NewRepositoryPutError(t *testing.T) {
	registry := &ECR{svc: &MockECR{putError: fmt.Errorf("put error")}}
	err := registry.Create("repo")
	assert.EqualError(t, err, "put error")
}

func TestEcr_NewRepository(t *testing.T) {
	mock := &MockECR{}
	registry := &ECR{svc: mock}
	repo := "repo"
	err := registry.Create(repo)
	assert.Nil(t, err)
	assert.Equal(t, &repo, mock.createRepositoryInput.RepositoryName)
	policyText := `{"rules":[{"rulePriority":10,"description":"Only keep 20 images","selection":{"tagStatus":"untagged","countType":"imageCountMoreThan","countNumber":20},"action":{"type":"expire"}}]}`
	assert.Equal(t, &policyText, mock.putLifecyclePolicyInput.LifecyclePolicyText)
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

type MockECR struct {
	ecriface.ECRAPI
	loginError                error
	authData                  string
	describeRepositoriesInput *awsecr.DescribeRepositoriesInput
	repoExists                bool
	createError               error
	createRepositoryInput     *awsecr.CreateRepositoryInput
	putLifecyclePolicyInput   *awsecr.PutLifecyclePolicyInput
	putError                  error
}

func (r MockECR) GetAuthorizationToken(input *awsecr.GetAuthorizationTokenInput) (*awsecr.GetAuthorizationTokenOutput, error) {
	if r.loginError != nil {
		return &awsecr.GetAuthorizationTokenOutput{AuthorizationData: []*awsecr.AuthorizationData{}}, r.loginError
	}
	return &awsecr.GetAuthorizationTokenOutput{AuthorizationData: []*awsecr.AuthorizationData{{AuthorizationToken: &r.authData}}}, nil
}

func (r *MockECR) DescribeRepositories(input *awsecr.DescribeRepositoriesInput) (*awsecr.DescribeRepositoriesOutput, error) {
	r.describeRepositoriesInput = input
	if r.repoExists {
		return &awsecr.DescribeRepositoriesOutput{Repositories: []*awsecr.Repository{}}, nil
	}
	return &awsecr.DescribeRepositoriesOutput{Repositories: []*awsecr.Repository{}}, fmt.Errorf("no repository found")
}

func (r *MockECR) CreateRepository(input *awsecr.CreateRepositoryInput) (*awsecr.CreateRepositoryOutput, error) {
	r.createRepositoryInput = input
	return &awsecr.CreateRepositoryOutput{}, r.createError
}

func (r *MockECR) PutLifecyclePolicy(input *awsecr.PutLifecyclePolicyInput) (*awsecr.PutLifecyclePolicyOutput, error) {
	r.putLifecyclePolicyInput = input
	return &awsecr.PutLifecyclePolicyOutput{}, r.putError
}
