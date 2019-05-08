package registry

import (
	"context"
	"docker.io/go-docker/api/types"
	"encoding/base64"
	"encoding/json"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	awsecr "github.com/aws/aws-sdk-go/service/ecr"
	"github.com/aws/aws-sdk-go/service/ecr/ecriface"
	"gitlab.com/sparetimecoders/build-tools/pkg/docker"
	"log"
	"os"
	"strings"
)

type ecr struct {
	url      string
	region   string
	username string
	password string
	svc      ecriface.ECRAPI
}

var _ Registry = &ecr{}

func (r *ecr) identify() bool {
	if url, exists := os.LookupEnv("ECR_URL"); exists {
		sess, err := session.NewSession(&aws.Config{Region: &r.region})
		if err != nil {
			return false
		}
		r.svc = awsecr.New(sess)
		log.Println("Will use AWS ECR as container registry")
		r.url = url
		r.region = os.Getenv("ECR_REGION")
		return true
	}
	return false
}

func (r *ecr) Login(client docker.Client) error {
	input := &awsecr.GetAuthorizationTokenInput{}

	result, err := r.svc.GetAuthorizationToken(input)
	if err != nil {
		return err
	}

	decoded, err := base64.StdEncoding.DecodeString(*result.AuthorizationData[0].AuthorizationToken)
	if err != nil {
		return err
	}
	parts := strings.Split(string(decoded), ":")
	r.username = parts[0]
	r.password = parts[1]

	if ok, err := client.RegistryLogin(context.Background(), types.AuthConfig{Username: r.username, Password: r.password, ServerAddress: r.url}); err == nil {
		log.Println(ok.Status)
		return nil
	} else {
		return err
	}
}

func (r *ecr) GetAuthInfo() string {
	auth := types.AuthConfig{Username: r.username, Password: r.password}
	authBytes, _ := json.Marshal(auth)
	return base64.URLEncoding.EncodeToString(authBytes)
}

func (r ecr) RegistryUrl() string {
	return r.url
}

func (r ecr) Create(repository string) error {
	if _, err := r.svc.DescribeRepositories(&awsecr.DescribeRepositoriesInput{RepositoryNames: []*string{&repository}}); err != nil {
		input := &awsecr.CreateRepositoryInput{
			RepositoryName: aws.String(repository),
		}

		if _, err := r.svc.CreateRepository(input); err != nil {
			return err
		} else {
			policyText := `{"rules":[{"rulePriority":10,"description":"Only keep 20 images","selection":{"tagStatus":"untagged","countType":"imageCountMoreThan","countNumber":20},"action":{"type":"expire"}}]}`
			if _, err := r.svc.PutLifecyclePolicy(&awsecr.PutLifecyclePolicyInput{LifecyclePolicyText: &policyText, RepositoryName: &repository}); err != nil {
				return err
			}
			return nil
		}
	}
	return nil
}
