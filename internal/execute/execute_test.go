package execute

import (
	"testing"

	"github.com/google/go-github/v35/github"
	"github.com/ministryofjustice/cloud-platform-git-xargs/internal/git"
)

// mockRepo uses the a relatively small repository to test against.
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

// TestCommand mocks a repository and clones it locally. It then performs a series of steps
// that determine if the function Command is working as expected.
func TestCommand(t *testing.T) {
	// Setup test by creating mock repo and cloning locally.
	t.Parallel()

	repo := mockRepo()
	client := github.NewClient(nil)

	repoDir, localRepo, _ := git.Clone(repo, client)

	tree, _ := localRepo.Worktree()

	// Should fail if a command with len == 0 is used in argument.
	err := Command(repoDir, "", tree, false)
	if err == nil {
		t.Error("When provided with an empty string; want fail, got continue")
	}

	// test: if loop == true it creates more than one file
	// test: if loop == false it creates only one file
	// test: if I pass a bad command, it fails.
}
