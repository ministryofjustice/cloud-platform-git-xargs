package cmd

import (
	"context"

	"github.com/google/go-github/v35/github"
	"golang.org/x/oauth2"
)

// ent, error) {
// 	// Ensure user provided a GITHUB_OAUTH_TOKEN
// 	GithubOauthToken := os.Getenv("GITHUB_OAUTH_TOKEN")
// 	if GithubOauthToken == "" {
// 		return nil, errors.New("You must have the GITHUB_OAUTH_TOKEN mapped")
// 	}

// 	ts := oauth2.StaticTokenSource(
// 		&oauth2.Token{AccessToken: GithubOauthToken},
// 	)

// 	tc := oauth2.NewClient(context.Background(), ts)

// 	// Wrap the go-github client in a GithubClient struct, which is common between production and test code
// 	client := NewClient(github.NewClient(tc))

// 	return client
// }

func GitHubClient(token string) *github.Client {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)

	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)

	return client
}
