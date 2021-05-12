/*
Copyright © 2021 Minisitry of Justice

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"sync"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/google/go-github/v35/github"
	"github.com/spf13/cobra"
)

var (
	command    []string
	repos, org string
	commit     bool
)

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Executes a cli command on a collection of repositories.",
	Long: `Given a GitHub organisation and a blob of repository names
pull the repository down locally, execute the command and then PR back
to main.

An example of this would be:

cloud-platform-git-xargs run --command "touch blankfile" \
															--organisation "github" \
															--repository "github"`,
	RunE: func(cmd *cobra.Command, args []string) error {
		token := os.Getenv("GITHUB_OAUTH_TOKEN")
		if os.Getenv("GITHUB_OAUTH_TOKEN") == "" {
			return errors.New("You must have the GITHUB_OAUTH_TOKEN env var")
		}

		// Create GH client using your personal access token
		client := GitHubClient(token)

		// Get repository names
		repos, err := fetchRepositories(client, org, repos)
		if err != nil {
			return err
		}

		// Clone repositories to disk
		var wg sync.WaitGroup
		for _, repo := range repos {
			wg.Add(1)
			go func(repo *github.Repository, client *github.Client) error {
				defer wg.Done()

				repoDir, localRepo, err = clone(repo, client)
				if err != nil {
					return err
				}
				return nil
			}(repo, client)
		}
		wg.Wait()

		// err = execute(command)
		// if err != nil {
		// 	return
		// }

		// if commit {
		// 	err = commit(repo)
		// 	if err != nil {
		// 		return
		// 	}
		// }
		return nil
	},
}

func init() {
	rootCmd.AddCommand(runCmd)

	runCmd.Flags().StringSliceVarP(&command, "command", "c", []string{}, "the command you'd like to execute i.e. touch file")
	runCmd.Flags().StringVarP(&repos, "repository", "r", "cloud-platform-environments", "a blob of the repository name i.e. cloud-platform-terraform")
	runCmd.Flags().StringVarP(&org, "organisation", "o", "ministryofjustice", "organisation of the repository i.e. ministryofjustice")
	runCmd.Flags().BoolVarP(&commit, "skip-commit", "s", false, "whether or not you want to create a commit and PR.")
}

func fetchRepositories(client *github.Client, org, blob string) ([]*github.Repository, error) {
	ctx := context.Background()
	opt := &github.RepositoryListByOrgOptions{
		Sort:        "full_name",
		Type:        "public",
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

		allRepos = append(allRepos, repos...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	var list []*github.Repository
	for _, repo := range allRepos {
		if strings.Contains(*repo.FullName, blob) {
			list = append(list, repo)
		}
	}

	return list, nil
}

// func binarySearch(a []*github.Repository, b string) []*github.Repository {
// 	start := 0
// 	end := len(a)-1
// 	mid := len(a) / 2

// 	for start <= end {
// 		value := a[mid]

// 		if strings.Contains(b, )
// 	}

// }

func clone(repo *github.Repository, token *github.Client) (string, *git.Repository, error) {
	// Create temporary directory on disk
	repoDir, err := ioutil.TempDir("./tmp", fmt.Sprintf(repo.GetName()))
	if err != nil {
		return "", nil, err
	}

	localRepo, err := git.PlainClone(repoDir, false, &git.CloneOptions{
		URL: repo.GetCloneURL(),
		Auth: &http.BasicAuth{
			Username: repo.GetOwner().GetLogin(),
			Password: os.Getenv("GITHUB_OAUTH_TOKEN"),
		},
	})
	if err != nil {
		return repoDir, nil, err
	}

	return repoDir, localRepo, nil
}
