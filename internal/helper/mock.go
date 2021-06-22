package helper

import (
	"log"
	"os"

	"github.com/go-git/go-git/v5"
	"github.com/google/go-github/v35/github"
	local "github.com/ministryofjustice/cloud-platform-git-xargs/internal/git"
)

// mockRepo uses a relatively small repository to test against.
func mockRepo() (r *github.Repository) {
	org := "ministryofjustice"
	repo := "cloud-platform-cli"

	url := "https://github.com/" + org + "/" + repo

	r = &github.Repository{
		Name:     &repo,
		CloneURL: &url,
	}

	return
}

// CreateMock simply executes the mockRepo function call making it much
// easier to read.
func CreateMock() (repoDir string, tree *git.Worktree) {
	repo := mockRepo()
	client := github.NewClient(nil)

	repoDir, localRepo, _ := local.Clone(repo, client)

	tree, _ = localRepo.Worktree()

	return
}

// cleanUp simply removes the directory created by the Clone function.
func CleanUpRepo() {
	err := os.RemoveAll("tmp")
	if err != nil {
		log.Fatalln("Temp directory not cleaned up. You may need to manually remove the ./cmd dir.")
	}
}
