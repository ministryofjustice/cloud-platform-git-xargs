package execute

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/go-git/go-git/v5"
	"github.com/google/go-github/v35/github"
	local "github.com/ministryofjustice/cloud-platform-git-xargs/internal/git"
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

// createMock simply executes the mockRepo function call making it much
// easier to read.
func createMock() (repoDir string, tree *git.Worktree) {
	repo := mockRepo()
	client := github.NewClient(nil)

	repoDir, localRepo, _ := local.Clone(repo, client)

	tree, _ = localRepo.Worktree()

	return
}

// TestCommand mocks a repository and clones it locally. It then performs a series of steps
// that determine if the function Command is working as expected.
func TestCommand(t *testing.T) {
	// Setup test by creating mock repo and cloning locally.
	t.Parallel()

	repoDir, tree := createMock()

	// Should fail if a command with len == 0 is used in argument.
	err := Command(repoDir, "", tree, false)
	if err == nil {
		t.Error("When provided with an empty string; want fail, got continue")
	}

	// Will pass if a file called file.md exists in each directory.
	err = Command(repoDir, "touch file.md", tree, true)
	if err != nil {
		t.Error("Unable to execute command when loop == true")
	}

	_ = filepath.Walk(repoDir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			t.Error("Unable to walk the tree. Fail.")
		}
		if info.IsDir() {
			_, err = os.Stat(path + "/file.md")
			if os.IsNotExist(err) {
				t.Error("Loop has NOT created file.md in directory. Fail.", path)
			}
		}
		return nil
	})
	// test: if loop == false it creates only one file
	repoDir, tree = createMock()
	filePath := repoDir + "/cmd/file.md"

	err = Command(repoDir, "touch file.md", tree, true)
	if err != nil {
		t.Error("Unable to run command when loop == false")
	}

	fmt.Println(filePath)
	_, err = os.Stat(filePath)
	if os.IsExist(err) {
		t.Error("File exists where it shouldn't; want no file.md, got file.md")
	}

	// test: if I pass a bad command, it fails.
}
