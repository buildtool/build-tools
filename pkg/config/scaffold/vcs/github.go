package vcs

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/go-github/v28/github"
	"github.com/sparetimecoders/build-tools/pkg"
	"golang.org/x/oauth2"
	"net/http"
)

type Github struct {
	Git
	Token        string `yaml:"token" env:"GITHUB_TOKEN"`
	Organisation string `yaml:"organisation" env:"GITHUB_ORG"`
	Public       bool   `yaml:"public"`
	repoOwner    string
	repositories RepositoriesService
}

func (v *Github) Name() string {
	return "Github"
}

func (v *Github) ValidateConfig() error {
	if len(v.Token) == 0 {
		return errors.New("token is required")
	}
	return nil
}

func (v *Github) Scaffold(name string) (*RepositoryInfo, error) {
	repo := &github.Repository{
		Name:     pkg.String(name),
		Private:  pkg.Bool(v.Public),
		AutoInit: pkg.Bool(true),
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
		if err != nil || (response != nil && response.StatusCode != http.StatusOK) {
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

func (v *Github) Webhook(name, url string) error {
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
		Active: pkg.Bool(true),
	}

	_, resp, err := v.repositories.CreateHook(context.Background(), v.repoOwner, name, hook)
	if err != nil || (resp != nil && resp.StatusCode != http.StatusCreated) {
		return fmt.Errorf("failed to create webhook %s", resp.Status)
	}

	return nil
}

func (v *Github) Validate(name string) error {
	// TODO: Check that repository doesn't already exists
	return nil
}

func (v *Github) Configure() {
	client := github.NewClient(oauth2.NewClient(context.Background(), oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: v.Token},
	)))
	v.repositories = client.Repositories
}

var _ VCS = &Github{}

type RepositoriesService interface {
	Create(ctx context.Context, org string, repo *github.Repository) (*github.Repository, *github.Response, error)
	UpdateBranchProtection(ctx context.Context, owner, repo, branch string, preq *github.ProtectionRequest) (*github.Protection, *github.Response, error)
	CreateHook(ctx context.Context, owner, repo string, hook *github.Hook) (*github.Hook, *github.Response, error)
}
