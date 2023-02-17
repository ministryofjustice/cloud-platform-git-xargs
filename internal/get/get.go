package get

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/google/go-github/v35/github"
)

// FetchRepositories takes a GitHub client, an org and a pattern/blob of a repository. It will query
// the GitHub API for all public occurrences of the pattern. It will return a list of GitHub repositories.
func FetchRepositories(client *github.Client, org, blob, filePath string) (allRepos []*github.Repository, err error) {
	ctx := context.Background()
	opt := &github.RepositoryListByOrgOptions{
		ListOptions: github.ListOptions{PerPage: 10},
	}

	if filePath != "" {
		if blob != "" {
			return nil, fmt.Errorf("cannot use both blob and file path flags")
		}
		fmt.Println("Fetching repositories from file...")
		repos, err := getReposFromFile(filePath)
		if err != nil {
			return nil, err
		}
		fmt.Println("Repositories fetched.", repos)

		allRepos, err = FetchRepositoriesFromList(client, repos, org)
		if err != nil {
			return nil, err
		}
	} else {
		allRepos, err = getReposFromOrg(client, ctx, org, blob, opt)
		if err != nil {
			return nil, err
		}
	}

	return allRepos, nil
}

func getReposFromFile(filePath string) ([]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}

	scanner := bufio.NewScanner(file)

	scanner.Split(bufio.ScanLines)
	var text []string

	for scanner.Scan() {
		text = append(text, scanner.Text())
	}
	fmt.Println("Repositories fetched from file: ", text)

	file.Close()

	return text, nil
}

func FetchRepositoriesFromList(client *github.Client, repos []string, org string) ([]*github.Repository, error) {
	var allRepos []*github.Repository
	for _, repo := range repos {
		fmt.Println("Fetching repository: ", repo)
		repo, err := getRepo(client, repo, org)
		if err != nil {
			return nil, err
		}
		allRepos = append(allRepos, repo)
	}

	return allRepos, nil
}

func getRepo(client *github.Client, repo, org string) (*github.Repository, error) {
	ctx := context.Background()
	r, _, err := client.Repositories.Get(ctx, org, repo)
	if err != nil {
		fmt.Println("Error fetching repository: ", repo)
		return nil, err
	}

	return r, nil
}

func getReposFromOrg(client *github.Client, ctx context.Context, org, blob string, opt *github.RepositoryListByOrgOptions) ([]*github.Repository, error) {
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

	return nil, nil
}
