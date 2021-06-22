package execute

import (
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/ministryofjustice/cloud-platform-git-xargs/internal/helper"
)

// TestCommandLoop will check to see if the loop bool is working when set
// to true and false.
func TestCommandLoop(t *testing.T) {
	defer helper.CleanUpRepo()
	t.Parallel()

	repoDir, tree := helper.CreateMock()
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
	defer helper.CleanUpRepo()

	// Create mock repository for first tests.
	repoDir, tree := helper.CreateMock()

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
