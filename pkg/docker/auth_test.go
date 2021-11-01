package docker

import (
	"context"
	"testing"

	"github.com/docker/docker/api/types"
	auth2 "github.com/moby/buildkit/session/auth"
	"github.com/stretchr/testify/require"
)

func Test_Credentials(t *testing.T) {
	auth := NewAuthenticator("use-auth.com", types.AuthConfig{
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
