package firebasetools_test

import (
	"context"
	"testing"

	"github.com/savannahghi/firebasetools"
	"github.com/stretchr/testify/assert"
)

const (
	// OnboardingRootDomain represents onboarding ISC URL
	OnboardingRootDomain = "https://profile-staging.healthcloud.co.ke"

	// OnboardingName represents the onboarding service ISC name
	OnboardingName = "onboarding"
)

func TestGetOrCreateAnonymousUser(t *testing.T) {
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Anonymous user happy case",
			args: args{
				ctx: context.Background(),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := firebasetools.GetOrCreateAnonymousUser(tt.args.ctx)
			assert.NotNil(t, got)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateAnonymousUser() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestGetFirestoreClientTestUtil(t *testing.T) {
	tests := []struct {
		name string

		wantErr bool
	}{
		{
			name: "Get Firestore Client Test Utils",

			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := firebasetools.GetFirestoreClientTestUtil(t)
			assert.NotNil(t, got)
		})
	}

}
