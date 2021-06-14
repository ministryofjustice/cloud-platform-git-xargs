package gitAction

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/google/go-github/v35/github"
)

// PushChanges takes a GitHub client, a branch, tree and repository. It first adds all changes to the git staging area, then commits,
// pushes and creates a PR, outputting any errors.
func PushChanges(client *github.Client, branch string, tree *git.Worktree, repo, message string, localRepo *git.Repository, remoteRepo *github.Repository) error {
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

// Checkout takes a GitHub client, a git reference and tree, along with local and remote repository.
// It will create a branch with the hardcoded name 'update', and will output a new git reference.
func Checkout(client *github.Client, ref *plumbing.Reference, tree *git.Worktree, remote *github.Repository, local *git.Repository) (plumbing.ReferenceName, error) {
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
func Clone(repo *github.Repository, token *github.Client) (string, *git.Repository, error) {
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
