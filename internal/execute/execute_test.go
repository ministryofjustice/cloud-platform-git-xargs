package execute

import (
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"testing"

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

// createMock simply executes the mockRepo function call making it much
// easier to read.
func createMock() (repoDir string, tree *git.Worktree) {
	repo := mockRepo()
	client := github.NewClient(nil)

	repoDir, localRepo, _ := local.Clone(repo, client)

	tree, _ = localRepo.Worktree()

	return
}

// cleanUp simply removes the directory created by the Clone function.
func cleanUp() {
	err := os.RemoveAll("tmp")
	if err != nil {
		log.Fatalln("Temp directory not cleaned up. You may need to manually remove the ./cmd dir.")
	}
}

// TestCommandLoop will check to see if the loop bool is working when set
// to true and false.
func TestCommandLoop(t *testing.T) {
	defer cleanUp()
	t.Parallel()

	repoDir, tree := createMock()
	filePath := repoDir + "/cmd/file.md"

	// Set loop to false and ensure command is run once and that the file is only
	// created in a single dir. If it exists in a child dir it'll fail.
	err := Command(repoDir, "touch file.md", tree, false)
	if err != nil {
		t.Error("Unable to run command when loop == false")
	}

	if _, err := os.Stat(filePath); err == nil {
		t.Error("File exists where it shouldn't; want no file.md, got file.md")
	}

	// Set loop to true. Will pass if a file called file.md exists in each directory.
	err = Command(repoDir, "touch file.md", tree, true)
	if err != nil {
		t.Error("Unable to execute command when loop == true")
	}

	// Walk all directories in the repository and look for existance of file in each dir.
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
}

// TestInvalidCommand tests whether sending an invalid command is handled.
func TestInvalidCommand(t *testing.T) {
	defer cleanUp()

	// Create mock repository for first tests.
	repoDir, tree := createMock()

	// Send an empty command and expect a failure.
	err := Command(repoDir, "", tree, false)
	if err == nil {
		t.Error("When provided with an empty string; want fail, got continue")
	}

	// Send a false command and expect a failure.
	err = Command(repoDir, "NOTACOMMAND", tree, false)
	if err == nil {
		t.Error("A command that should of failed, passed; want error, got success.")
	}
}
