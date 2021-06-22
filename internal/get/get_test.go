package get

import (
	"fmt"
	"testing"

	"github.com/google/go-github/v35/github"
)

// TestGetRepository tests the FetchRepositories function by passing a static
// GitHub client, an organisation and a pattern of a repository name. It should
// pass if a single repository is returned and matches the value of blob.
func TestFetchRepositories(t *testing.T) {
	client := github.NewClient(nil)
	org := "ministryofjustice"
	blob := "cloud-platform-cli"

	repos, err := FetchRepositories(client, org, blob)
	if err != nil {
		t.Error("Unable to fetch repositories:", err)
	}

	for _, repo := range repos {
		fmt.Println(repo.Name)
		if repo.Name != &blob {
			t.Errorf("Repository name not found; got %p, wanted cloud-platform-cli", repo.Name)
		}
	}
}
