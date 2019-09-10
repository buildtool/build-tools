package config

import (
	"context"
	"fmt"
	"github.com/golang/mock/gomock"
	"github.com/google/go-github/v28/github"
	"github.com/stretchr/testify/assert"
	"gitlab.com/sparetimecoders/build-tools/pkg/config/mocks"
	"net/http"
	"testing"
)

func TestGithubVCS_Scaffold(t *testing.T) {
	repoName := "reponame"
	repoCloneUrl := "cloneurl"
	orgName := "org"

	git := GithubVCS{
		Organisation: orgName,
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	m := mocks.NewMockRepositoriesService(ctrl)

	repository := github.Repository{
		Name:     wrapString(repoName),
		AutoInit: wrapBool(true),
		Private:  wrapBool(false),
	}

	repositoryResponse := repository
	repositoryResponse.CloneURL = wrapString(repoCloneUrl)

	m.EXPECT().
		Create(context.Background(), orgName, &repository).Return(
		&repositoryResponse,
		&github.Response{
			Response: &http.Response{
				StatusCode: http.StatusCreated,
			},
		}, nil).
		Times(1)

	m.EXPECT().
		UpdateBranchProtection(context.Background(), orgName, repoName, "master", &github.ProtectionRequest{
			EnforceAdmins: true,
			RequiredPullRequestReviews: &github.PullRequestReviewsEnforcementRequest{
				DismissStaleReviews:          true,
				RequiredApprovingReviewCount: 1,
			},
		}).Return(nil,
		&github.Response{
			Response: &http.Response{
				StatusCode: http.StatusOK,
			},
		}, nil).
		Times(1)

	res, err := git.scaffold(m, repoName)
	assert.NoError(t, err)
	assert.Equal(t, repoCloneUrl, res)
}

func TestGithubVCS_ScaffoldWithoutOrganisation(t *testing.T) {
	repoName := "reponame"
	repoCloneUrl := "cloneurl"

	git := GithubVCS{}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	m := mocks.NewMockRepositoriesService(ctrl)

	repository := github.Repository{
		Name:     wrapString(repoName),
		AutoInit: wrapBool(true),
		Private:  wrapBool(false),
	}

	repositoryResponse := repository
	repositoryResponse.CloneURL = wrapString(repoCloneUrl)
	repositoryResponse.Owner = &github.User{
		Login: wrapString("user-login"),
	}

	m.EXPECT().
		Create(context.Background(), "", &repository).Return(
		&repositoryResponse,
		&github.Response{
			Response: &http.Response{
				StatusCode: http.StatusCreated,
			},
		}, nil).
		Times(1)

	m.EXPECT().
		UpdateBranchProtection(context.Background(), "user-login", repoName, "master", &github.ProtectionRequest{
			EnforceAdmins: true,
			RequiredPullRequestReviews: &github.PullRequestReviewsEnforcementRequest{
				DismissStaleReviews:          true,
				RequiredApprovingReviewCount: 1,
			},
		}).Return(nil,
		&github.Response{
			Response: &http.Response{
				StatusCode: http.StatusOK,
			},
		}, nil).
		Times(1)

	res, err := git.scaffold(m, repoName)
	assert.NoError(t, err)
	assert.Equal(t, repoCloneUrl, res)

}

func TestGithubVCS_Scaffold_RepositoryAlreadyExist(t *testing.T) {
	// TODO In reality we get this error for any response that is not http.StatusCreated
	repoName := "ALREADY_EXISTS"
	git := GithubVCS{}
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	m := mocks.NewMockRepositoriesService(ctrl)

	repository := &github.Repository{
		Name:     wrapString(repoName),
		AutoInit: wrapBool(true),
		Private:  wrapBool(false),
	}

	m.EXPECT().
		Create(context.Background(), "", repository).Return(repository,
		&github.Response{
			Response: &http.Response{
				StatusCode: http.StatusUnprocessableEntity,
				Status:     "already exists",
			},
		}, nil).
		Times(1)

	_, err := git.scaffold(m, repoName)
	assert.EqualError(t, err, "failed to create repository ALREADY_EXISTS, already exists")
}

func TestGithubVCS_Scaffold_CreateError(t *testing.T) {
	repoName := "ALREADY_EXISTS"
	git := GithubVCS{
		Organisation: "org",
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	m := mocks.NewMockRepositoriesService(ctrl)

	repository := &github.Repository{
		Name:     wrapString(repoName),
		AutoInit: wrapBool(true),
		Private:  wrapBool(false),
	}

	m.EXPECT().
		Create(context.Background(), "org", repository).Return(
		repository, nil, fmt.Errorf("failed to create repo")).
		Times(1)

	_, err := git.scaffold(m, repoName)
	assert.EqualError(t, err, "failed to create repo")
}

func TestGithubVCS_ScaffoldProtectBranchError(t *testing.T) {
	repoName := "reponame"
	repoCloneUrl := "cloneurl"

	git := GithubVCS{}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	m := mocks.NewMockRepositoriesService(ctrl)

	repository := github.Repository{
		Name:     wrapString(repoName),
		AutoInit: wrapBool(true),
		Private:  wrapBool(false),
	}

	repositoryResponse := repository
	repositoryResponse.CloneURL = wrapString(repoCloneUrl)
	repositoryResponse.Owner = &github.User{
		Login: wrapString("user-login"),
	}

	m.EXPECT().
		Create(context.Background(), "", &repository).Return(
		&repositoryResponse,
		&github.Response{
			Response: &http.Response{
				StatusCode: http.StatusCreated,
			},
		}, nil).
		Times(1)

	m.EXPECT().
		UpdateBranchProtection(context.Background(), "user-login", repoName, "master", &github.ProtectionRequest{
			EnforceAdmins: true,
			RequiredPullRequestReviews: &github.PullRequestReviewsEnforcementRequest{
				DismissStaleReviews:          true,
				RequiredApprovingReviewCount: 1,
			},
		}).Return(
		nil,
		&github.Response{
			Response: &http.Response{
				StatusCode: http.StatusBadRequest,
				Status:     "something went wrong",
			},
		}, nil).
		Times(1)

	_, err := git.scaffold(m, repoName)
	assert.EqualError(t, err, "failed to set repository branch protection something went wrong")

}
