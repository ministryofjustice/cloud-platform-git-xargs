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
	"errors"
	"fmt"
	"os"

	"github.com/google/go-github/v35/github"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"

	"github.com/ministryofjustice/cloud-platform-git-xargs/internal/execute"
	"github.com/ministryofjustice/cloud-platform-git-xargs/internal/get"
	"github.com/ministryofjustice/cloud-platform-git-xargs/internal/git"
)

// All passed via flags
var (
	command, message string
	repos, org       string
	skipCommit, loop bool
)

// runCmd represents the run command. This command, with arguments,
// will enable the user to run a command against a collection of repositories.
// Commit that change and then create a pull request.
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
		// You must set a GITHUB_OAUTH_TOKEN environment variable
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

		// Loop over all repositories, run a command, commit the change and
		// create a pull request. Decided to use errgroup instead of waitgroups
		// as it was easier to understand.
		g := new(errgroup.Group)
		for _, repo := range repos {
			g.Go(func() error {
				err = processRepo(repo, client)
				if err != nil {
					return err
				}
				return err
			})
		}
		if err := g.Wait(); err != nil {
			fmt.Println("Error processing repositories", err)
		}
		return nil
	},
}

func processRepo(repo *github.Repository, client *github.Client) error {
	// Clone repository to local disk
	repoDir, localRepo, err := git.Clone(repo, client)
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
	branch, err := git.Checkout(client, ref, tree, repo, localRepo)
	if err != nil {
		return err
	}

	// Execute command
	err = execute.Command(repoDir, command, tree, loop)
	if err != nil {
		return err
	}

	// As long as skipCommit isn't true, stage, push and pr changes
	if !skipCommit {
		err = git.PushChanges(client, branch.String(), tree, repoDir, message, localRepo, repo)
		if err != nil {
			return err
		}
	}
	return nil
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
