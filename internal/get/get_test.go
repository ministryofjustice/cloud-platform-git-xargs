package get

import (
	"reflect"
	"testing"

	"github.com/google/go-github/v35/github"
	"github.com/migueleliasweb/go-github-mock/src/mock"
)

// TestGetRepository tests the getRepository function by mocking the github API
// and responding with a mocked repository collection.
func TestFetchRepositories(t *testing.T) {
	mockedClient := mock.NewMockedHTTPClient(
		mock.WithRequestMatchPages(
			mock.GetOrgsReposByOrg,
			[]github.Repository{
				{
					Name: github.String("repo-A-on-first-page"),
				},
				{
					Name: github.String("repo-B-on-first-page"),
				},
			},
			[]github.Repository{
				{
					Name: github.String("repo-C-on-second-page"),
				},
				{
					Name: github.String("repo-D-on-second-page"),
				},
			},
		),
	)

	type args struct {
		client *github.Client
		org    string
		blob   string
	}
	tests := []struct {
		name    string
		args    args
		want    []*github.Repository
		wantErr bool
	}{
		{
			name: "get correct repositories",
			args: args{
				client: github.NewClient(mockedClient),
				org:    "test",
				blob:   "repo",
			},
			want: []*github.Repository{
				{
					Name: github.String("repo-A-on-first-page"),
				},
				{
					Name: github.String("repo-B-on-first-page"),
				},
				{
					Name: github.String("repo-C-on-second-page"),
				},
				{
					Name: github.String("repo-D-on-second-page"),
				},
			},
			wantErr: false,
		},
		{
			name: "pass incorrect blob",
			args: args{
				client: github.NewClient(mockedClient),
				org:    "test",
				blob:   "obviouslyWrong",
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := FetchRepositories(tt.args.client, tt.args.org, tt.args.blob)
			if (err != nil) != tt.wantErr {
				t.Errorf("FetchRepositories() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FetchRepositories() = %v, want %v", got, tt.want)
			}
		})
	}
}
