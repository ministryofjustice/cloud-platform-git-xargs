package main

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

type GitHubAction struct {
	On struct {
		PullRequest struct {
			Types []string `yaml:"types"`
		} `yaml:"pull_request"`
	} `yaml:"on"`
	Jobs struct {
		GoTests struct {
			Name   string `yaml:"name"`
			RunsOn string `yaml:"runs-on"`
			Steps  []struct {
				Name string `yaml:"name"`
				Uses string `yaml:"uses,omitempty"`
				With struct {
					TerraformVersion string `yaml:"terraform_version"`
					TerraformWrapper bool   `yaml:"terraform_wrapper"`
				} `yaml:"with,omitempty"`
				WorkingDirectory string `yaml:"working-directory,omitempty"`
				Run              string `yaml:"run,omitempty"`
			} `yaml:"steps"`
		} `yaml:"go-tests"`
	} `yaml:"jobs"`
}

func main() {
	g := GitHubAction{}
	wd, _ := os.Getwd()
	// marshal the file into the struct
	f, err := os.ReadFile(".github/workflows/unit.yml")
	if err != nil {
		fmt.Println("no unit.yml found in", strings.Split(wd, "/")[len(strings.Split(wd, "/"))-1])
	}

	err = yaml.Unmarshal(f, &g)
	if err != nil {
		fmt.Println("error:", err)
	}

	for step := 0; step < len(g.Jobs.GoTests.Steps); step++ {
		if g.Jobs.GoTests.Steps[step].Name == "Setup Terraform" {
			g.Jobs.GoTests.Steps[step].With.TerraformVersion = "1.2.5"
		}
	}

	// marshal the struct back into the file
	out, err := yaml.Marshal(g)
	if err != nil {
		fmt.Println("error:", err)
	}

	err = os.WriteFile(".github/workflows/unit.yml", out, 0644)
	if err != nil {
		fmt.Println("error:", err)
	}

	fmt.Println("updated unit.yml in", strings.Split(wd, "/")[len(strings.Split(wd, "/"))-1])
}
