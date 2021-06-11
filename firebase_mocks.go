package firebase_tools

import (
	"context"
	"net/http"

	"cloud.google.com/go/firestore"
	"firebase.google.com/go/auth"
	"firebase.google.com/go/messaging"
)

// MockFirebaseApp is used to mock the behavior of a Firebase app for testing
type MockFirebaseApp struct {
	MockAuthErr    error
	MockAuthClient *auth.Client
	MockRefreshErr error

	MockFirestore    *firestore.Client
	MockFirestoreErr error

	MockMessaging    *messaging.Client
	MockMessagingErr error
}

// Auth returns a mock client or error, as set on the struct
func (fa *MockFirebaseApp) Auth(_ context.Context) (*auth.Client, error) {
	return fa.MockAuthClient, fa.MockAuthErr
}

// RevokeRefreshTokens returns an error on attempted refresh
func (fa *MockFirebaseApp) RevokeRefreshTokens(ctx context.Context, uid string) error {
	return fa.MockRefreshErr
}

// Firestore returns a mock Firestore or error, as set on the struct
func (fa *MockFirebaseApp) Firestore(ctx context.Context) (*firestore.Client, error) {
	return fa.MockFirestore, fa.MockFirestoreErr
}

// Messaging returns a mock Firebase Cloud Messaging client or error, as set on the struct
func (fa *MockFirebaseApp) Messaging(ctx context.Context) (*messaging.Client, error) {
	return fa.MockMessaging, fa.MockMessagingErr
}

// MockFirebaseClient is used to mock the behavior of a Firebase client for testing
type MockFirebaseClient struct {
	MockApp                IFirebaseApp
	MockAppInitErr         error
	MockFirebaseUserTokens *FirebaseUserTokens
	MockFirebaseAuthError  error
}

// InitFirebase returns a mock Firebase app or error, as set on the struct
func (fc *MockFirebaseClient) InitFirebase() (IFirebaseApp, error) {
	return fc.MockApp, fc.MockAppInitErr
}

// AuthenticateCustomFirebaseToken returns mock user tokens or an error, as set on the struct
func (fc *MockFirebaseClient) AuthenticateCustomFirebaseToken(_ string, _ *http.Client) (*FirebaseUserTokens, error) {
	return fc.MockFirebaseUserTokens, fc.MockFirebaseAuthError
}
