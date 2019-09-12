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
	repoSSHUrl := "cloneurl"
	repoCloneUrl := "https://github.com/example/repo"
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
	repositoryResponse.SSHURL = wrapString(repoSSHUrl)
	repositoryResponse.CloneURL = wrapString(repoCloneUrl)

	m.EXPECT().
		Create(context.Background(), orgName, &repository).Return(
		&repositoryResponse, githubCreatedResponse, nil).
		Times(1)

	m.EXPECT().
		UpdateBranchProtection(context.Background(), orgName, repoName, "master", &github.ProtectionRequest{
			EnforceAdmins: true,
			RequiredPullRequestReviews: &github.PullRequestReviewsEnforcementRequest{
				DismissStaleReviews:          true,
				RequiredApprovingReviewCount: 1,
			},
		}).Return(nil, githubOkResponse, nil).
		Times(1)

	res, err := git.scaffold(m, repoName)
	assert.NoError(t, err)
	assert.Equal(t, &RepositoryInfo{repoSSHUrl, repoCloneUrl}, res)
}

func TestGithubVCS_ScaffoldWithoutOrganisation(t *testing.T) {
	repoName := "reponame"
	repoSSHUrl := "cloneurl"
	repoCloneUrl := "https://github.com/example/repo"

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
	repositoryResponse.SSHURL = wrapString(repoSSHUrl)
	repositoryResponse.CloneURL = wrapString(repoCloneUrl)
	repositoryResponse.Owner = &github.User{
		Login: wrapString("user-login"),
	}

	m.EXPECT().
		Create(context.Background(), "", &repository).Return(
		&repositoryResponse, githubCreatedResponse, nil).
		Times(1)

	m.EXPECT().
		UpdateBranchProtection(context.Background(), "user-login", repoName, "master", &github.ProtectionRequest{
			EnforceAdmins: true,
			RequiredPullRequestReviews: &github.PullRequestReviewsEnforcementRequest{
				DismissStaleReviews:          true,
				RequiredApprovingReviewCount: 1,
			},
		}).Return(nil, githubOkResponse, nil).
		Times(1)

	res, err := git.scaffold(m, repoName)
	assert.NoError(t, err)
	assert.Equal(t, &RepositoryInfo{repoSSHUrl, repoCloneUrl}, res)

}

func TestGithubVCS_Scaffold_RepositoryAlreadyExist(t *testing.T) {
	// TODO In reality we get this error for any response that is not http.StatusCreated
	repoName := "ALREADY_EXISTS"
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
	repositoryResponse.Owner = &github.User{
		Login: wrapString("user-login"),
	}

	m.EXPECT().
		Create(context.Background(), "", &repository).Return(&repositoryResponse,
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
	repositoryResponse.SSHURL = wrapString(repoCloneUrl)
	repositoryResponse.Owner = &github.User{
		Login: wrapString("user-login"),
	}

	m.EXPECT().
		Create(context.Background(), "", &repository).Return(
		&repositoryResponse, githubCreatedResponse, nil).
		Times(1)

	m.EXPECT().
		UpdateBranchProtection(context.Background(), "user-login", repoName, "master", &github.ProtectionRequest{
			EnforceAdmins: true,
			RequiredPullRequestReviews: &github.PullRequestReviewsEnforcementRequest{
				DismissStaleReviews:          true,
				RequiredApprovingReviewCount: 1,
			},
		}).Return(
		nil, githubBadRequestResponse, nil).
		Times(1)

	_, err := git.scaffold(m, repoName)
	assert.EqualError(t, err, "failed to set repository branch protection something went wrong")
}

func TestGithubVCS_SillyTests(t *testing.T) {
	githubVCS := GithubVCS{}
	assert.EqualErrorf(t, githubVCS.Validate(), "token is required", "")
	githubVCS.Token = ""
	assert.EqualErrorf(t, githubVCS.Validate(), "token is required", "")

	githubVCS.Token = "token"
	assert.NoError(t, githubVCS.Validate())

	assert.Equal(t, githubVCS.Name(), "Github")
}

func TestGithubVCS_Webhook(t *testing.T) {
	githubVCS := GithubVCS{
		repoOwner: "test",
	}
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	m := mocks.NewMockRepositoriesService(ctrl)
	m.EXPECT().CreateHook(context.Background(), "test", "repo", &github.Hook{
		Events: []string{
			"push",
			"pull_request",
			"deployment",
		},
		Config: map[string]interface{}{
			"url":          "https://ab.cd",
			"content_type": "json",
		},
		Active: wrapBool(true),
	}).Return(nil, githubCreatedResponse, nil).
		Times(1)

	err := githubVCS.webhook(m, "repo", "https://ab.cd")
	assert.NoError(t, err)
}

func TestGithubVCS_WebhookError(t *testing.T) {
	githubVCS := GithubVCS{
		repoOwner: "test",
	}
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	m := mocks.NewMockRepositoriesService(ctrl)
	m.EXPECT().CreateHook(context.Background(), "test", "repo", &github.Hook{
		Events: []string{
			"push",
			"pull_request",
			"deployment",
		},
		Config: map[string]interface{}{
			"url":          "https://ab.cd",
			"content_type": "json",
		},
		Active: wrapBool(true),
	}).Return(nil, githubBadRequestResponse, nil).
		Times(1)

	err := githubVCS.webhook(m, "repo", "https://ab.cd")
	assert.EqualError(t, err, "failed to create webhook something went wrong")
}

var githubOkResponse = &github.Response{
	Response: &http.Response{
		StatusCode: http.StatusOK,
	},
}

var githubCreatedResponse = &github.Response{
	Response: &http.Response{
		StatusCode: http.StatusCreated,
	},
}

var githubBadRequestResponse = &github.Response{
	Response: &http.Response{
		StatusCode: http.StatusBadRequest,
		Status:     "something went wrong",
	},
}
