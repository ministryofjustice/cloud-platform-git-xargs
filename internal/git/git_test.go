package git

import (
	"testing"

	"github.com/google/go-github/v35/github"
	"github.com/ministryofjustice/cloud-platform-git-xargs/internal/helper"
)

func TestCheckout(t *testing.T) {
	defer helper.CleanUpRepo()
	_ = github.NewClient(nil)
	_, _ := helper.CreateMock()

}
