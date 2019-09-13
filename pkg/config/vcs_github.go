package config

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/go-github/v28/github"
	"golang.org/x/oauth2"
	"net/http"
)

type GithubVCS struct {
	git
	Token        string `yaml:"token" env:"GITHUB_TOKEN"`
	Organisation string `yaml:"organisation" env:"GITHUB_ORG"`
	Public       bool   `yaml:"public"`
	repoOwner    string
	repositories RepositoriesService
}

func (v *GithubVCS) Name() string {
	return "Github"
}

func (v *GithubVCS) Scaffold(name string) (*RepositoryInfo, error) {
	repo := &github.Repository{
		Name:     wrapString(name),
		Private:  wrapBool(v.Public),
		AutoInit: wrapBool(true),
	}
	repo, resp, err := v.repositories.Create(context.Background(), v.Organisation, repo)
	if err != nil {
		return nil, err
	}

	v.repoOwner = v.Organisation
	if v.repoOwner == "" {
		v.repoOwner = *repo.Owner.Login
	}
	switch resp.StatusCode {
	case http.StatusCreated:
		preq := &github.ProtectionRequest{
			RequiredPullRequestReviews: &github.PullRequestReviewsEnforcementRequest{
				DismissStaleReviews:          true,
				RequiredApprovingReviewCount: 1,
			},
			EnforceAdmins: true,
		}

		_, response, err := v.repositories.UpdateBranchProtection(context.Background(), v.repoOwner, *repo.Name, "master", preq)
		if err != nil || response.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("failed to set repository branch protection %s", response.Status)
		}
	default:
		return nil, fmt.Errorf("failed to create repository %s, %s", name, resp.Status)
	}
	return &RepositoryInfo{
		SSHURL:  *repo.SSHURL,
		HTTPURL: *repo.CloneURL,
	}, nil
}

func (v *GithubVCS) Webhook(name, url string) error {
	hook := &github.Hook{
		Events: []string{
			"push",
			"pull_request",
			"deployment",
		},
		Config: map[string]interface{}{
			"url":          url,
			"content_type": "json",
		},
		Active: wrapBool(true),
	}

	_, resp, err := v.repositories.CreateHook(context.Background(), v.repoOwner, name, hook)
	if err != nil || resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("failed to create webhook %s", resp.Status)
	}

	return nil
}

func (v *GithubVCS) Validate(name string) error {
	if len(v.Token) == 0 {
		return errors.New("token is required")
	}
	// TODO: Check that repository doesn't already exists
	return nil
}

func (v *GithubVCS) configure() {
	client := github.NewClient(oauth2.NewClient(context.Background(), oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: v.Token},
	)))
	v.repositories = client.Repositories
}

var _ VCS = &GithubVCS{}

type RepositoriesService interface {
	Create(ctx context.Context, org string, repo *github.Repository) (*github.Repository, *github.Response, error)
	UpdateBranchProtection(ctx context.Context, owner, repo, branch string, preq *github.ProtectionRequest) (*github.Protection, *github.Response, error)
	CreateHook(ctx context.Context, owner, repo string, hook *github.Hook) (*github.Hook, *github.Response, error)
}
