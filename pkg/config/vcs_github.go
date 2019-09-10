package config

import (
	"context"
	"fmt"
	"github.com/google/go-github/v28/github"
	"golang.org/x/oauth2"
	"net/http"
)

type GithubVCS struct {
	git
	Token        string `yaml:"Token" env:"GITHUB_TOKEN"`
	Organisation string `yaml:"Organisation" env:"GITHUB_ORG"`
	Public       bool   `yaml:"Public"`
}

func (v GithubVCS) Name() string {
	return "Github"
}

func (v GithubVCS) Scaffold(name string) (string, error) {
	client := github.NewClient(oauth2.NewClient(context.Background(), oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: v.Token},
	)))

	return v.scaffold(client.Repositories, name)
}

func (v GithubVCS) Webhook(name, url string) {
}

var _ VCS = &GithubVCS{}

func (v GithubVCS) scaffold(repositoriesService RepositoriesService, name string) (string, error) {
	repo := &github.Repository{
		Name:     wrapString(name),
		Private:  wrapBool(v.Public),
		AutoInit: wrapBool(true),
	}
	repo, resp, err := repositoriesService.Create(context.Background(), v.Organisation, repo)
	if err != nil {
		return "", err
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
		owner := v.Organisation
		if owner == "" {
			owner = *repo.Owner.Login
		}

		_, response, err := repositoriesService.UpdateBranchProtection(context.Background(), owner, *repo.Name, "master", preq)
		if err != nil || response.StatusCode != http.StatusOK {
			return "", fmt.Errorf("failed to set repository branch protection %s", response.Status)
		}
	default:
		return "", fmt.Errorf("failed to create repository %s, %s", name, resp.Status)
	}
	return *repo.CloneURL, nil

}

type RepositoriesService interface {
	Create(ctx context.Context, org string, repo *github.Repository) (*github.Repository, *github.Response, error)
	UpdateBranchProtection(ctx context.Context, owner, repo, branch string, preq *github.ProtectionRequest) (*github.Protection, *github.Response, error)
}
