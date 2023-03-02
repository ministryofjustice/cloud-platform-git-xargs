package main

import (
	_ "embed"
	"fmt"
	"strings"
)

//go:embed repos.txt
var repos string

func main() {

	for _, repo := range strings.Split(strings.TrimSpace(repos), "\n") {
		url := fmt.Sprintf("https://github.com/ministryofjustice/%s/pulls", repo)
		fmt.Println(url)
	}

}
