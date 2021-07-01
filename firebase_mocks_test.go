package firebasetools_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/savannahghi/firebasetools"
	"github.com/stretchr/testify/assert"
)

func TestAuth(t *testing.T) {
	ctx := context.Background()
	authToken := firebasetools.GetAuthToken(ctx, t)
	authenticatedContext := context.WithValue(ctx, firebasetools.AuthTokenContextKey, authToken)
	fb := firebasetools.MockFirebaseApp{}
	_, err := fb.Auth(authenticatedContext)
	assert.Nil(t, err)
}

func TestRevokeRefreshTokens(t *testing.T) {
	ctx := context.Background()
	authToken := firebasetools.GetAuthToken(ctx, t)
	authenticatedContext := context.WithValue(ctx, firebasetools.AuthTokenContextKey, authToken)
	fb := firebasetools.MockFirebaseApp{}
	err := fb.RevokeRefreshTokens(authenticatedContext, "string")
	assert.Nil(t, err)
}

func TestRevokeFirestore(t *testing.T) {
	ctx := context.Background()
	authToken := firebasetools.GetAuthToken(ctx, t)
	authenticatedContext := context.WithValue(ctx, firebasetools.AuthTokenContextKey, authToken)
	fb := firebasetools.MockFirebaseApp{}
	_, err := fb.Firestore(authenticatedContext)
	assert.Nil(t, err)
}

func TestMessaging(t *testing.T) {
	ctx := context.Background()
	authToken := firebasetools.GetAuthToken(ctx, t)
	authenticatedContext := context.WithValue(ctx, firebasetools.AuthTokenContextKey, authToken)
	fb := firebasetools.MockFirebaseApp{}
	_, err := fb.Messaging(authenticatedContext)
	assert.Nil(t, err)
}

func TestMockInitFirebase(t *testing.T) {
	fb := firebasetools.MockFirebaseClient{}
	_, err := fb.InitFirebase()
	assert.Nil(t, err)
}

func TestAuthenticateCustomFirebaseToken(t *testing.T) {
	fb := firebasetools.MockFirebaseClient{}
	_, err := fb.AuthenticateCustomFirebaseToken("token", http.DefaultClient)
	assert.Nil(t, err)
}
