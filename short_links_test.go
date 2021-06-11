package firebase_tools_test

import (
	"context"
	"testing"

	fb "github.com/savannahghi/firebase_tools"
	"github.com/savannahghi/server_utils"
	"github.com/stretchr/testify/assert"
)

func TestShortenLink(t *testing.T) {
	dynamicLinkDomain, err := server_utils.GetEnvVar(fb.FDLDomainEnvironmentVariableName)
	assert.Nil(t, err)

	type args struct {
		ctx      context.Context
		longLink string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "good case",
			args: args{
				ctx: context.Background(),
				// TODO: MOVE this to an env var
				longLink: "https://console.cloud.google.com/run/detail/europe-west1/api-gateway/revisions?project=bewell-app-testing",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := fb.ShortenLink(tt.args.ctx, tt.args.longLink)
			if (err != nil) != tt.wantErr {
				t.Errorf("ShortenLink() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Contains(t, got, dynamicLinkDomain)
		})
	}
}
