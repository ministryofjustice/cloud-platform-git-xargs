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

	"github.com/go-git/go-git/v5/plumbing"
	"github.com/ministryofjustice/cloud-platform-git-xargs/internal/get"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/google/go-github/v35/github"
	"github.com/spf13/cobra"
)

// All passed via flags
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
		// GITHUB_OAUTH_TOKEN must be set
		token := os.Getenv("GITHUB_OAUTH_TOKEN")
		if os.Getenv("GITHUB_OAUTH_TOKEN") == "" {
			return errors.New("you must have the GITHUB_OAUTH_TOKEN env var")
		}

		// Create GH client using your personal access token
		client := GitHubClient(token)

		// Get all repository names containing value to repos variable/flag
		repos, err := get.FetchRepositories(client, org, repos)
		if err != nil {
			return err
		}

		// Loop over all repositories and perform operations
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

			// As long as skipCommit isn't true, stage, push and pr changes
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

// pushChanges takes a GitHub client, a branch, tree and repository. It first adds all changes to the git staging area, then commits,
// pushes and creates a PR, outputting any errors.
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

// executeCommand takes a directory path, a command to execute and a git
// worktree. If a loop is specified, it'll execute the command argument on
// every directory. Otherwise it'll just execute once on the root of the
// repository. It outputs an error if found.
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
	return nil
}

// checkout takes a GitHub client, a git reference and tree, along with local and remote repository.
// It will create a branch with the hardcoded name 'update', and will output a new git reference.
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

// Clone takes a GitHub repository and client. It will look to create a local copy of the
// repository in the `tmp/` directory. It will then output the repository directory, name and
// an error if there is one.
func clone(repo *github.Repository, token *github.Client) (string, *git.Repository, error) {
	tmpDir := "./tmp"
	if _, err := os.Stat(tmpDir); os.IsNotExist(err) {
		file := filepath.Join(".", tmpDir)
		os.MkdirAll(file, os.ModePerm)
	}

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
