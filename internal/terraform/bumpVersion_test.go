package terraform

import "testing"

func TestBumpTfVersion(t *testing.T) {
	type args struct {
		repoDir   string
		tfVersion string
		loop      bool
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Successfully bump terraform version",
			args: args{
				repoDir:   "/Users/poornima.krishnasamy/go/src/ministryofjustice/cloud-platform-environments/namespaces/live.cloud-platform.service.justice.gov.uk/abundant-namespace-dev/resources",
				tfVersion: "1.2.9",
				loop:      false,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := BumpTfVersion(tt.args.repoDir, tt.args.tfVersion, tt.args.loop); (err != nil) != tt.wantErr {
				t.Errorf("BumpTfVersion() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
