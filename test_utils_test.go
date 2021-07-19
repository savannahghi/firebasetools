package firebasetools_test

import (
	"context"
	"os"
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

func TestGetAuthenticatedContextFromUID(t *testing.T) {
	ctx := context.Background()

	// create a valid uid

	type args struct {
		uid string
	}
	tests := []struct {
		name      string
		args      args
		changeEnv bool
		wantErr   bool
	}{
		{
			name: "valid case",
			args: args{
				uid: "some invalid uid",
			},
			changeEnv: false,
			wantErr:   false,
		},
		{
			name: "invalid: wrong uid",
			args: args{
				uid: "some invalid uid",
			},
			changeEnv: true,
			wantErr:   true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			initialKey := os.Getenv("FIREBASE_WEB_API_KEY")

			if tt.changeEnv {
				os.Setenv("FIREBASE_WEB_API_KEY", "invalidkey")
			}

			got, err := firebasetools.GetAuthenticatedContextFromUID(ctx, tt.args.uid)
			if got == nil && !tt.wantErr {
				t.Errorf("invalid auth token")
			}
			if (err != nil) != tt.wantErr {
				t.Errorf("GetAuthenticatedContextFromUID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			os.Setenv("FIREBASE_WEB_API_KEY", initialKey)

		})
	}
}
