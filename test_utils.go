package firebasetools

import (
	"context"
	"fmt"
	"testing"

	"cloud.google.com/go/firestore"
	"firebase.google.com/go/auth"
	"github.com/stretchr/testify/assert"
)

// GetAuthenticatedContext returns a logged in context, useful for test purposes
func GetAuthenticatedContext(t *testing.T) context.Context {
	ctx := context.Background()
	authToken := GetAuthToken(ctx, t)
	authenticatedContext := context.WithValue(ctx, AuthTokenContextKey, authToken)
	return authenticatedContext
}

// GetFirestoreClientTestUtil ...
func GetFirestoreClientTestUtil(t *testing.T) *firestore.Client {
	fc := &FirebaseClient{}
	firebaseApp, err := fc.InitFirebase()
	assert.Nil(t, err)

	ctx := GetAuthenticatedContext(t)
	firestoreClient, err := firebaseApp.Firestore(ctx)
	assert.Nil(t, err)
	assert.NotNil(t, firestoreClient)
	return firestoreClient
}

// GetAuthenticatedContextAndToken returns a logged in context and ID token.
// It is useful for test purposes
func GetAuthenticatedContextAndToken(t *testing.T) (context.Context, *auth.Token) {
	ctx := context.Background()
	authToken := GetAuthToken(ctx, t)
	authenticatedContext := context.WithValue(ctx, AuthTokenContextKey, authToken)
	return authenticatedContext, authToken
}

// GetAuthToken ...
func GetAuthToken(ctx context.Context, t *testing.T) *auth.Token {
	authToken, _ := getAuthTokenAndBearerToken(ctx, t)
	return authToken
}

func getAuthTokenAndBearerToken(ctx context.Context, t *testing.T) (*auth.Token, string) {
	user, userErr := GetOrCreateFirebaseUser(ctx, TestUserEmail)
	assert.Nil(t, userErr)
	assert.NotNil(t, user)

	customToken, tokenErr := CreateFirebaseCustomToken(ctx, user.UID)
	assert.Nil(t, tokenErr)
	assert.NotNil(t, customToken)

	idTokens, idErr := AuthenticateCustomFirebaseToken(customToken)
	assert.Nil(t, idErr)
	assert.NotNil(t, idTokens)

	bearerToken := idTokens.IDToken
	authToken, err := ValidateBearerToken(ctx, bearerToken)
	assert.Nil(t, err)
	assert.NotNil(t, authToken)

	return authToken, bearerToken
}

// GetOrCreateAnonymousUser creates an anonymous user
// For documentation and test purposes only
func GetOrCreateAnonymousUser(ctx context.Context) (*auth.UserRecord, error) {
	authClient, err := GetFirebaseAuthClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to get or create Firebase client: %w", err)
	}
	anonymousUserUID := "AgkGYKUsRifO2O9fTLDuVCMr2hb2" // This is an anonymous user

	existingUser, userErr := authClient.GetUser(ctx, anonymousUserUID)

	if userErr == nil {
		return existingUser, nil
	}

	params := (&auth.UserToCreate{})
	newUser, createErr := authClient.CreateUser(ctx, params)
	if createErr != nil {
		return nil, createErr
	}
	return newUser, nil
}

// GetAnonymousContext returns an anonymous logged in context, useful for test purposes
func GetAnonymousContext(t *testing.T) context.Context {
	ctx := context.Background()
	authToken := getAnonymousAuthToken(ctx, t)
	authenticatedContext := context.WithValue(ctx, AuthTokenContextKey, authToken)
	return authenticatedContext
}

func getAnonymousAuthToken(ctx context.Context, t *testing.T) *auth.Token {
	user, userErr := GetOrCreateAnonymousUser(ctx)
	assert.Nil(t, userErr)
	assert.NotNil(t, user)

	customToken, tokenErr := CreateFirebaseCustomToken(ctx, user.UID)
	assert.Nil(t, tokenErr)
	assert.NotNil(t, customToken)

	idTokens, idErr := AuthenticateCustomFirebaseToken(customToken)
	assert.Nil(t, idErr)
	assert.NotNil(t, idTokens)

	bearerToken := idTokens.IDToken
	authToken, err := ValidateBearerToken(ctx, bearerToken)
	assert.Nil(t, err)
	assert.NotNil(t, authToken)

	return authToken
}

// GetAuthenticatedContextFromUID creates an auth.Token given a valid uid
func GetAuthenticatedContextFromUID(ctx context.Context, uid string) (*auth.Token, error) {
	customToken, err := CreateFirebaseCustomToken(ctx, uid)
	if err != nil {
		return nil, fmt.Errorf("failed to create an authenticated token: %w", err)
	}

	idTokens, err := AuthenticateCustomFirebaseToken(customToken)
	if err != nil {
		return nil, fmt.Errorf("failed to authenticated custom token: %w", err)
	}

	authToken, err := ValidateBearerToken(ctx, idTokens.IDToken)
	if err != nil {
		return nil, fmt.Errorf("failed to validate bearer token: %w", err)
	}

	return authToken, nil
}
