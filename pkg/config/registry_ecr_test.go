package config

import (
	"bytes"
	"fmt"
	awsecr "github.com/aws/aws-sdk-go/service/ecr"
	"github.com/aws/aws-sdk-go/service/ecr/ecriface"
	"github.com/stretchr/testify/assert"
	"gitlab.com/sparetimecoders/build-tools/pkg/docker"
	"os"
	"testing"
)

func TestEcr_Identify(t *testing.T) {
	os.Clearenv()
	_ = os.Setenv("ECR_URL", "url")
	_ = os.Setenv("ECR_REGION", "region")

	out := &bytes.Buffer{}
	cfg, err := Load(name, out)
	assert.NoError(t, err)
	registry, err := cfg.CurrentRegistry()
	assert.NoError(t, err)
	assert.NotNil(t, registry)
	assert.Equal(t, "url", registry.RegistryUrl())
	assert.Equal(t, "", out.String())
}

func TestEcr_Name(t *testing.T) {
	os.Clearenv()
	_ = os.Setenv("ECR_URL", "url")
	_ = os.Setenv("ECR_REGION", "region")

	out := &bytes.Buffer{}
	cfg, err := Load(name, out)
	assert.NoError(t, err)
	registry, err := cfg.CurrentRegistry()
	assert.NoError(t, err)
	assert.Equal(t, "ECR", registry.Name())
}

func TestEcr_Identify_BrokenConfig(t *testing.T) {
	os.Clearenv()
	_ = os.Setenv("ECR_URL", "url")
	_ = os.Setenv("ECR_REGION", "region")
	_ = os.Setenv("AWS_CA_BUNDLE", "/missing/bundle")

	out := &bytes.Buffer{}
	cfg, err := Load(name, out)
	assert.NoError(t, err)
	registry, err := cfg.CurrentRegistry()
	assert.EqualError(t, err, "no Docker registry found")
	assert.Nil(t, registry)
	assert.Equal(t, "", out.String())
}

func TestEcr_LoginAuthRequestFailed(t *testing.T) {
	client := &docker.MockDocker{}
	registry := &ECRRegistry{Url: "ecr-url", Region: "eu-west-1", svc: &MockECR{loginError: fmt.Errorf("auth failure")}}
	out := &bytes.Buffer{}
	err := registry.Login(client, out)
	assert.EqualError(t, err, "auth failure")
	assert.Equal(t, "", out.String())
}

func TestEcr_LoginInvalidAuthData(t *testing.T) {
	client := &docker.MockDocker{}
	registry := &ECRRegistry{Url: "ecr-url", Region: "eu-west-1", svc: &MockECR{authData: "aaabbb"}}
	out := &bytes.Buffer{}
	err := registry.Login(client, out)
	assert.EqualError(t, err, "illegal base64 data at input byte 4")
	assert.Equal(t, "", out.String())
}

func TestEcr_LoginFailed(t *testing.T) {
	client := &docker.MockDocker{LoginError: fmt.Errorf("invalid username/password")}
	registry := &ECRRegistry{Url: "ecr-url", Region: "eu-west-1", svc: &MockECR{authData: "QVdTOmFiYzEyMw=="}}
	out := &bytes.Buffer{}
	err := registry.Login(client, out)
	assert.EqualError(t, err, "invalid username/password")
	assert.Equal(t, "", out.String())
}

func TestEcr_LoginSuccess(t *testing.T) {
	client := &docker.MockDocker{}
	registry := &ECRRegistry{Url: "ecr-url", Region: "eu-west-1", svc: &MockECR{authData: "QVdTOmFiYzEyMw=="}}
	out := &bytes.Buffer{}
	err := registry.Login(client, out)
	assert.Nil(t, err)
	assert.Equal(t, "AWS", client.Username)
	assert.Equal(t, "abc123", client.Password)
	assert.Equal(t, "ecr-url", client.ServerAddress)
	assert.Equal(t, "Logged in\n", out.String())
}

func TestEcr_GetAuthInfo(t *testing.T) {
	registry := &ECRRegistry{Url: "ecr-url", Region: "eu-west-1", username: "AWS", password: "abc123"}
	auth := registry.GetAuthInfo()
	assert.Equal(t, "eyJ1c2VybmFtZSI6IkFXUyIsInBhc3N3b3JkIjoiYWJjMTIzIn0=", auth)
}

func TestEcr_ExistingRepository(t *testing.T) {
	mock := &MockECR{repoExists: true}
	registry := &ECRRegistry{svc: mock}
	repo := "repo"
	err := registry.Create(repo)
	assert.Nil(t, err)
	assert.Equal(t, []*string{&repo}, mock.describeRepositoriesInput.RepositoryNames)
}

func TestEcr_NewRepositoryCreateError(t *testing.T) {
	registry := &ECRRegistry{svc: &MockECR{createError: fmt.Errorf("create error")}}
	err := registry.Create("repo")
	assert.EqualError(t, err, "create error")
}

func TestEcr_NewRepositoryPutError(t *testing.T) {
	registry := &ECRRegistry{svc: &MockECR{putError: fmt.Errorf("put error")}}
	err := registry.Create("repo")
	assert.EqualError(t, err, "put error")
}

func TestEcr_NewRepository(t *testing.T) {
	mock := &MockECR{}
	registry := &ECRRegistry{svc: mock}
	repo := "repo"
	err := registry.Create(repo)
	assert.Nil(t, err)
	assert.Equal(t, &repo, mock.createRepositoryInput.RepositoryName)
	policyText := `{"rules":[{"rulePriority":10,"description":"Only keep 20 images","selection":{"tagStatus":"untagged","countType":"imageCountMoreThan","countNumber":20},"action":{"type":"expire"}}]}`
	assert.Equal(t, &policyText, mock.putLifecyclePolicyInput.LifecyclePolicyText)
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
