package firebasetools_test

import (
	"context"
	"testing"

	fb "github.com/savannahghi/firebasetools"
	"github.com/savannahghi/serverutils"
	"github.com/stretchr/testify/assert"
)

func TestShortenLink(t *testing.T) {
	dynamicLinkDomain, err := serverutils.GetEnvVar(fb.FDLDomainEnvironmentVariableName)
	assert.Nil(t, err)
	faultyDynamicLinkDomain, err := serverutils.GetEnvVar("")
	assert.NotNil(t, err)
	assert.Equal(t, faultyDynamicLinkDomain, "")

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
				longLink: "europe-west1-sghi-307909.cloudfunctions.net",
			},
			wantErr: false,
		},
		{
			name: "missing longLink",
			args: args{
				ctx: context.Background(),
				// TODO: MOVE this to an env var
				longLink: "",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := fb.ShortenLink(tt.args.ctx, tt.args.longLink)
			if (err != nil) != tt.wantErr {
				t.Errorf("ShortenLink() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				assert.NotNil(t, err)
				assert.Equal(t, "", got)
				return
			}
			assert.Contains(t, got, dynamicLinkDomain)
		})
	}
}
