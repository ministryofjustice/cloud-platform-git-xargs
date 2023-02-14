package terraform

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
)

func BumpTfVersion(repoDir, tfVersion string, loop bool) error {
	// Parse versions.tf file
	// if the loop switch is set to true, the chosen command will execute in every directory.
	if loop {
		err := filepath.Walk(repoDir, func(path string, info fs.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.Name() == "versions.tf" {
				err := updateVersions(path, tfVersion)
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
		file := repoDir + "/" + "versions.tf"
		err := updateVersions(file, tfVersion)
		if err != nil {
			return err
		}
	}
	return nil

}

func updateVersions(path, tfVersion string) error {
	var blocks []*hclwrite.Block

	data, err := os.ReadFile(path)

	if err != nil {
		return fmt.Errorf("error reading file %s", err)
	}

	f, diags := hclwrite.ParseConfig(data, path, hcl.Pos{
		Line:   0,
		Column: 0,
	})

	if diags.HasErrors() {
		return fmt.Errorf("error getting TF resource: %s", diags)
	}
	// Grab slice of blocks in HCL file.
	blocks = f.Body().Blocks()
	for _, block := range blocks {
		fmt.Println(block.Labels())

	}
	return nil
}
