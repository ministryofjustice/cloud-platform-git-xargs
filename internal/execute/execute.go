package execute

import (
	"errors"
	"io/fs"
	"os/exec"
	"path/filepath"

	"github.com/go-git/go-git/v5"
)

// executeCommand takes a directory path, a command to execute and a git
// worktree. If a loop is specified, it'll execute the command argument on
// every directory. Otherwise it'll just execute once on the root of the
// repository. It outputs an error if found.
func executeCommand(dir, command string, tree *git.Worktree, loop bool) error {
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
