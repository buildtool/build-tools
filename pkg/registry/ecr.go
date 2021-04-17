package registry

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	awsecr "github.com/aws/aws-sdk-go/service/ecr"
	"github.com/aws/aws-sdk-go/service/ecr/ecriface"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/aws/aws-sdk-go/service/sts/stsiface"
	"github.com/docker/docker/api/types"

	"github.com/buildtool/build-tools/pkg/docker"
)

type ECR struct {
	dockerRegistry `yaml:"-"`
	Url            string `yaml:"url" env:"ECR_URL"`
	Region         string `yaml:"region,omitempty" env:"ECR_REGION"`
	username       string
	password       string
	ecrSvc         ecriface.ECRAPI
	stsSvc         stsiface.STSAPI
	registryId     *string
}

var _ Registry = &ECR{}

func (r *ECR) Name() string {
	return "ECR"
}

func (r *ECR) Configured() bool {
	if len(r.Url) > 0 {
		sess, err := session.NewSession(&aws.Config{Region: r.region()})
		if err != nil {
			return false
		}
		r.ecrSvc = awsecr.New(sess)
		r.stsSvc = sts.New(sess)
		registryId, err := r.registry()
		if err != nil {
			return false
		}
		r.registryId = registryId
		return true
	}
	return false
}

func (r *ECR) region() *string {
	if r.Region == "" {
		regex := regexp.MustCompile(`.*\.dkr\.ecr.(.*)\.amazonaws\.com`)
		if submatch := regex.FindStringSubmatch(r.Url); len(submatch) == 2 {
			r.Region = submatch[1]
		}
	}
	return &r.Region
}

func (r *ECR) registry() (*string, error) {
	regex := regexp.MustCompile(`(.*)\.dkr\.ecr..*\.amazonaws\.com`)
	if submatch := regex.FindStringSubmatch(r.Url); len(submatch) == 2 {
		return &submatch[1], nil
	}
	return nil, fmt.Errorf("failed to extract registryid from string %s", r.Url)
}

func (r *ECR) Login(client docker.Client, out io.Writer) error {
	input := &awsecr.GetAuthorizationTokenInput{}

	result, err := r.ecrSvc.GetAuthorizationToken(input)
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

	if ok, err := client.RegistryLogin(context.Background(), types.AuthConfig{Username: r.username, Password: r.password, ServerAddress: r.Url}); err == nil {
		_, _ = fmt.Fprintln(out, ok.Status)
		return nil
	} else {
		return err
	}
}

func (r *ECR) GetAuthConfig() types.AuthConfig {
	return types.AuthConfig{Username: r.username, Password: r.password}
}

func (r *ECR) GetAuthInfo() string {
	authBytes, _ := json.Marshal(r.GetAuthConfig())
	return base64.URLEncoding.EncodeToString(authBytes)
}

func (r ECR) RegistryUrl() string {
	return r.Url
}

func (r ECR) Create(repository string) error {
	identity, err := r.stsSvc.GetCallerIdentity(&sts.GetCallerIdentityInput{})
	if err != nil {
		return err
	}
	if *identity.Account != *r.registryId {
		return fmt.Errorf("account mismatch, logged in at '%s' got '%s' from repository url %s", *identity.Account, *r.registryId, r.Url)
	}
	if _, err := r.ecrSvc.DescribeRepositories(&awsecr.DescribeRepositoriesInput{
		RegistryId:      r.registryId,
		RepositoryNames: []*string{&repository},
	}); err != nil {
		switch err.(type) {
		case *awsecr.RepositoryNotFoundException:
			break
		default:
			return err
		}
		input := &awsecr.CreateRepositoryInput{
			RepositoryName: aws.String(repository),
		}

		if _, err := r.ecrSvc.CreateRepository(input); err != nil {
			return err
		} else {
			policyText := `{"rules":[{"rulePriority":10,"description":"Only keep 20 images","selection":{"tagStatus":"untagged","countType":"imageCountMoreThan","countNumber":20},"action":{"type":"expire"}}]}`
			if _, err := r.ecrSvc.PutLifecyclePolicy(&awsecr.PutLifecyclePolicyInput{LifecyclePolicyText: &policyText, RepositoryName: &repository}); err != nil {
				return err
			}
			return nil
		}
	}
	return nil
}
