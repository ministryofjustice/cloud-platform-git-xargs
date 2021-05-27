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
	"io/fs"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5/plumbing"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/google/go-github/v35/github"
	"github.com/spf13/cobra"
)

var (
	command, message string
	repos, org       string
	skipCommit, loop bool
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
			return errors.New("you must have the GITHUB_OAUTH_TOKEN env var")
		}

		// Create GH client using your personal access token
		client := GitHubClient(token)

		// Get repository names
		repos, err := fetchRepositories(client, org, repos)
		if err != nil {
			return err
		}

		// Clone repositories to disk
		// var wg sync.WaitGroup
		for _, repo := range repos {
			// 	wg.Add(1)
			// 	go func(repo *github.Repository, client *github.Client) error {
			// defer wg.Done()

			// Clone repository to local disk
			repoDir, localRepo, err := clone(repo, client)
			if err != nil {
				return err
			}

			// Get HEAD ref from repository
			ref, err := localRepo.Head()
			if err != nil {
				return err
			}

			// Get the worktree for the local repository
			tree, err := localRepo.Worktree()
			if err != nil {
				return err
			}

			// Create local branch
			branch, err := checkout(client, ref, tree, repo, localRepo)
			if err != nil {
				return err
			}

			// Execute command
			err = executeCommand(repoDir, command, tree)
			if err != nil {
				return err
			}

			if !skipCommit {
				err = pushChanges(client, branch.String(), tree, repoDir, localRepo, repo)
				if err != nil {
					return err
				}
			}
		}
		// 		return nil
		// 	}(repo, client)
		// }
		// wg.Wait()

		return nil
	},
}

func init() {
	rootCmd.AddCommand(runCmd)

	runCmd.Flags().StringVarP(&command, "command", "c", "", "the command you'd like to execute i.e. touch file")
	runCmd.Flags().StringVarP(&repos, "repository", "r", "cloud-platform-environments", "a blob of the repository name i.e. cloud-platform-terraform")
	runCmd.Flags().StringVarP(&org, "organisation", "o", "ministryofjustice", "organisation of the repository i.e. ministryofjustice")
	runCmd.Flags().StringVarP(&message, "commit", "m", "perform command on repository", "the commit message you'd like to make")
	runCmd.Flags().BoolVarP(&skipCommit, "skip-commit", "s", false, "whether or not you want to create a commit and PR.")
	runCmd.Flags().BoolVarP(&loop, "loop-dir", "l", false, "if you wish to execute the command on every directory in repository.")
}

func pushChanges(client *github.Client, branch string, tree *git.Worktree, repo string, localRepo *git.Repository, remoteRepo *github.Repository) error {
	defaultBranch := remoteRepo.GetDefaultBranch()
	status, err := tree.Status()
	if err != nil {
		return err
	}

	if status.IsClean() {
		return errors.New("warning: no changes to commit")
	}

	for path := range status {
		if status.IsUntracked(path) {
			_, err := tree.Add(path)
			if err != nil {
				return nil
			}
		}
	}

	_, err = tree.Commit(message, &git.CommitOptions{
		All: true,
	})
	if err != nil {
		return err
	}

	err = localRepo.Push(&git.PushOptions{
		RemoteName: "origin",
		Auth: &http.BasicAuth{
			Username: remoteRepo.GetOwner().GetLogin(),
			Password: os.Getenv("GITHUB_OAUTH_TOKEN"),
		},
	})
	if err != nil {
		return err
	}

	createPR := &github.NewPullRequest{
		Title: github.String(message),
		Head:  github.String(branch),
		Base:  github.String(defaultBranch),
	}

	_, _, err = client.PullRequests.Create(context.Background(), *remoteRepo.GetOwner().Login, remoteRepo.GetName(), createPR)
	if err != nil {
		return err
	}

	return nil
}

func executeCommand(dir, command string, tree *git.Worktree) error {
	if len(command) < 1 {
		return errors.New("no command executed")
	}

	// if the loop switch is set to true, the chosen command will execute in every directory.
	if loop {
		err := filepath.Walk(dir, func(path string, info fs.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				cmd := exec.Command("/bin/sh", "-c", command)
				cmd.Dir = filepath.Dir(path) + "/" + info.Name()
				err := cmd.Run()
				if err != nil {
					return err
				}
			}
			return nil
		})
		if err != nil {
			return err
		}
	} else {
		cmd := exec.Command("/bin/sh", "-c", command)
		cmd.Dir = dir
		err := cmd.Run()
		if err != nil {
			return err
		}
	}

	// status, err := tree.Status()
	// if err != nil {
	// 	return err
	// }

	// if !status.IsClean() {
	// 	return errors.New("repository worktree is no longer clean. Stage new files and commit")
	// }

	return nil
}

func checkout(client *github.Client, ref *plumbing.Reference, tree *git.Worktree, remote *github.Repository, local *git.Repository) (plumbing.ReferenceName, error) {
	branchName := plumbing.NewBranchReferenceName("update")

	create := &git.CheckoutOptions{
		Hash:   ref.Hash(),
		Branch: branchName,
		Create: true,
	}

	err := tree.Checkout(create)
	if err != nil {
		return "", err
	}

	return branchName, nil
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
