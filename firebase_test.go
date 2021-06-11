package firebase_tools_test

import (
	"context"
	"reflect"
	"testing"

	"cloud.google.com/go/firestore"
	"firebase.google.com/go/auth"
	"github.com/google/uuid"
	fb "github.com/savannahghi/firebase_tools"
	"github.com/savannahghi/go_utils"
	"github.com/stretchr/testify/assert"
)

func TestInitFirebase(t *testing.T) {
	fc := fb.FirebaseClient{}
	fb, err := fc.InitFirebase()
	assert.Nil(t, err)
	assert.NotNil(t, fb)
}

func GetIDToken(t *testing.T) string {
	ctx := context.Background()
	user, err := fb.GetOrCreateFirebaseUser(ctx, fb.TestUserEmail)
	if err != nil {
		t.Errorf("unable to create Firebase user for email %v, error %v", fb.TestUserEmail, err)
	}

	// test custom token generation
	customToken, err := fb.CreateFirebaseCustomToken(ctx, user.UID)
	if err != nil {
		t.Errorf("unable to get custom token for %#v", user)
	}

	// test authentication of custom Firebase tokens
	idTokens, err := fb.AuthenticateCustomFirebaseToken(customToken)
	if err != nil {
		t.Errorf("unable to exchange custom token for ID tokens, error %s", err)
	}
	if idTokens.IDToken == "" {
		t.Errorf("got blank ID token")
	}
	return idTokens.IDToken
}

func TestGetOrCreateFirebaseUser(t *testing.T) {
	ctx := context.Background()
	tests := []struct {
		email string
	}{
		{email: fb.TestUserEmail},
	}
	for _, tc := range tests {
		user, err := fb.GetOrCreateFirebaseUser(ctx, tc.email)
		if err != nil {
			t.Errorf("unable to create Firebase user for email %v, error %v", tc.email, err)
		}

		// sanity check
		if user.Email != tc.email {
			t.Errorf("expected to get back a user with email %s, got %s", tc.email, user.Email)
		}

		// test custom token generation
		customToken, err := fb.CreateFirebaseCustomToken(ctx, user.UID)
		if err != nil {
			t.Errorf("unable to get custom token for %#v", user)
		}

		// test authentication of custom Firebase tokens
		idTokens, err := fb.AuthenticateCustomFirebaseToken(customToken)
		if err != nil {
			t.Errorf("unable to exchange custom token for ID tokens, error %s", err)
		}
		if idTokens.IDToken == "" {
			t.Errorf("got blank ID token")
		}
	}
}

func TestAuthenticateCustomFirebaseToken_Invalid_Token(t *testing.T) {
	invalidToken := uuid.New().String()
	returnToken, err := fb.AuthenticateCustomFirebaseToken(invalidToken)
	assert.Errorf(t, err, "expected invalid token to fail authentication with message %s")
	var nilToken *fb.FirebaseUserTokens
	assert.Equal(t, nilToken, returnToken)
}

func TestAuthenticateCustomFirebaseToken_Valid_Token(t *testing.T) {
	ctx := context.Background()
	user, err := fb.GetOrCreateFirebaseUser(ctx, fb.TestUserEmail)
	assert.Nilf(t, err, "unexpected user retrieval error '%s'")
	validToken, tokenErr := fb.CreateFirebaseCustomToken(ctx, user.UID)
	assert.Nilf(t, tokenErr, "unexpected custom token creation error '%s'")
	idTokens, validateErr := fb.AuthenticateCustomFirebaseToken(validToken)
	assert.Nilf(t, validateErr, "unexpected custom token validation/exchange error '%s'")
	assert.NotNilf(t, idTokens.IDToken, "expected ID token to be non nil")
}

func TestGenerateSafeIdentifier(t *testing.T) {
	id := fb.GenerateSafeIdentifier()
	assert.NotZero(t, id)
}

func TestUpdateRecordOnFirestore(t *testing.T) {
	firestoreClient := fb.GetFirestoreClient(t)
	collection := "integration_test_collection"
	data := map[string]string{
		"a_key_for_testing": uuid.New().String(),
	}
	id, err := fb.SaveDataToFirestore(firestoreClient, collection, data)
	assert.Nil(t, err)
	assert.NotZero(t, id)

	type args struct {
		firestoreClient *firestore.Client
		collection      string
		id              string
		data            interface{}
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "good case",
			args: args{
				firestoreClient: firestoreClient,
				collection:      collection,
				id:              id,
				data: map[string]string{
					"a_key_for_testing": uuid.New().String(),
				},
			},
			wantErr: false,
		},
		{
			name: "invalid id",
			args: args{
				firestoreClient: firestoreClient,
				collection:      collection,
				id:              "this is a fake id that should not be found",
				data: map[string]string{
					"a_key_for_testing": uuid.New().String(),
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := fb.UpdateRecordOnFirestore(tt.args.firestoreClient, tt.args.collection, tt.args.id, tt.args.data); (err != nil) != tt.wantErr {
				t.Errorf("UpdateRecordOnFirestore() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGetUserTokenFromContext(t *testing.T) {
	authenticatedContext, authToken := go_utils.GetAuthenticatedContextAndToken(t)
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		args    args
		want    *auth.Token
		wantErr bool
	}{
		{
			name: "good case - authenticated context",
			args: args{
				ctx: authenticatedContext,
			},
			want:    authToken,
			wantErr: false,
		},
		{
			name: "unauthenticated context",
			args: args{
				ctx: context.Background(),
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "context with bad value",
			args: args{
				ctx: context.WithValue(
					context.Background(),
					fb.AuthTokenContextKey,
					"this is definitely not an auth token",
				),
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := fb.GetUserTokenFromContext(tt.args.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetUserTokenFromContext() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetUserTokenFromContext() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCheckIsAnonymousUser(t *testing.T) {
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{
			name: "Anonymous user",
			args: args{
				ctx: go_utils.GetAnonymousContext(t),
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "Known user",
			args: args{
				ctx: go_utils.GetAuthenticatedContext(t),
			},
			want:    false,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := fb.CheckIsAnonymousUser(tt.args.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("CheckIsAnonymousUser() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("CheckIsAnonymousUser() = %v, want %v", got, tt.want)
			}
		})
	}
}
