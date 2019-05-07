package registry

import (
	"gitlab.com/sparetimecoders/build-tools/pkg/docker"
	"log"
	"os"
)

type ecr struct {
	url    string
	region string
}

var _ Registry = &ecr{}

func (r *ecr) identify() bool {
	if url, exists := os.LookupEnv("ECR_URL"); exists {
		log.Println("Will use AWS ECR as container registry")
		r.url = url
		r.region = os.Getenv("ECR_REGION")
		return true
	}
	return false
}

func (r ecr) Login(client docker.Client) bool {
	// TODO: Use AWS SDK to get auth token etc. https://docs.aws.amazon.com/sdk-for-go/api/service/ecr/#example_ECR_GetAuthorizationToken_shared00
	return false
}

func (r ecr) RegistryUrl() string {
	return r.url
}

func (r ecr) Create() bool {
	// TODO: Use AWS SDK to create registry
	panic("implement me")
}

func (r ecr) Validate() bool {
	panic("implement me")
}
