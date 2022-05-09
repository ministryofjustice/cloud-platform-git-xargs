package get

import (
	"context"
	"strings"

	"github.com/google/go-github/v35/github"
)

// FetchRepositories takes a GitHub client, an org and a pattern/blob of a repository. It will query
// the GitHub API for all public occurrences of the pattern. It will return a list of GitHub repositories.
func FetchRepositories(client *github.Client, org, blob string) ([]*github.Repository, error) {
	ctx := context.Background()
	opt := &github.RepositoryListByOrgOptions{
		ListOptions: github.ListOptions{PerPage: 10},
	}

	// Becuase of the potential number of org repositories pagination is added.
	// Warning: this can take a while if the org contains a number of repositories.
	var allRepos []*github.Repository
	for {
		repos, resp, err := client.Repositories.ListByOrg(ctx, org, opt)
		if err != nil {
			return nil, err
		}

		for _, repo := range repos {
			if strings.Contains(*repo.Name, blob) {
				allRepos = append(allRepos, repo)
			}
		}
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	return allRepos, nil
}
