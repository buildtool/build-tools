package docker

import (
	"context"

	"github.com/docker/docker/api/types"
	"github.com/moby/buildkit/session"
	"github.com/moby/buildkit/session/auth"
	"google.golang.org/grpc"
)

type authenticator struct {
	authConfig types.AuthConfig
}

func NewAuthenticator(authConfig types.AuthConfig) session.Attachable {
	return &authenticator{
		authConfig: authConfig,
	}
}

func (a authenticator) Register(server *grpc.Server) {
	auth.RegisterAuthServer(server, a)
}

func (a authenticator) Credentials(ctx context.Context, request *auth.CredentialsRequest) (*auth.CredentialsResponse, error) {
	return &auth.CredentialsResponse{Username: a.authConfig.Username, Secret: a.authConfig.Password}, nil
}

func (a authenticator) FetchToken(ctx context.Context, request *auth.FetchTokenRequest) (*auth.FetchTokenResponse, error) {
	panic("implement me FetchToken")
}

func (a authenticator) GetTokenAuthority(ctx context.Context, request *auth.GetTokenAuthorityRequest) (*auth.GetTokenAuthorityResponse, error) {
	panic("implement me GetTokenAuthority")
}

func (a authenticator) VerifyTokenAuthority(ctx context.Context, request *auth.VerifyTokenAuthorityRequest) (*auth.VerifyTokenAuthorityResponse, error) {
	panic("implement me VerifyTokenAuthority")
}

var _ auth.AuthServer = &authenticator{}
var _ session.Attachable = &authenticator{}
