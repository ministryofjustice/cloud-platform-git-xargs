package main

import (
	"context"
	"flag"
	"fmt"
	"strings"

	"github.com/google/go-github/v35/github"
	"github.com/ministryofjustice/cloud-platform-git-xargs/cmd"
)

func main() {
	var token string
	flag.StringVar(&token, "token", "", "GitHub token")
	flag.Parse()

	client := cmd.GitHubClient(token)

	opt := &github.RepositoryListByOrgOptions{
		ListOptions: github.ListOptions{PerPage: 10},
		Type:        "public",
	}

	for {
		repo, resp, err := client.Repositories.ListByOrg(context.Background(), "ministryofjustice", opt)
		if err != nil {
			panic(err)
		}

		for _, r := range repo {
			if *r.Archived == false && *r.Fork == false && strings.Contains(*r.Name, "cloud-platform-terraform") {
				fmt.Println(*r.Name)
			}
		}

		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

}
