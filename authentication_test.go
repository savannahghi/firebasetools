package firebasetools_test

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/savannahghi/firebasetools"
	"github.com/stretchr/testify/assert"
)

func GetIDToken(t *testing.T) string {
	ctx := context.Background()
	user, err := firebasetools.GetOrCreateFirebaseUser(ctx, firebasetools.TestUserEmail)
	if err != nil {
		t.Errorf("unable to create Firebase user for email %v, error %v", firebasetools.TestUserEmail, err)
	}

	// test custom token generation
	customToken, err := firebasetools.CreateFirebaseCustomToken(ctx, user.UID)
	if err != nil {
		t.Errorf("unable to get custom token for %#v", user)
	}

	// test authentication of custom Firebase tokens
	idTokens, err := firebasetools.AuthenticateCustomFirebaseToken(customToken)
	if err != nil {
		t.Errorf("unable to exchange custom token for ID tokens, error %s", err)
	}
	if idTokens.IDToken == "" {
		t.Errorf("got blank ID token")
	}
	return idTokens.IDToken
}

func TestAuthenticationMiddleware(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
	fc := &firebasetools.FirebaseClient{}
	fa, err := fc.InitFirebase()
	assert.Nil(t, err)
	assert.NotNil(t, fa)

	mw := firebasetools.AuthenticationMiddleware(fa)
	h := mw(next)
	rw := httptest.NewRecorder()
	reader := bytes.NewBuffer([]byte("sample"))
	idToken := GetIDToken(t)
	authHeader := fmt.Sprintf("Bearer %s", idToken)
	req := httptest.NewRequest(http.MethodPost, "/", reader)
	req.Header.Add("Authorization", authHeader)
	h.ServeHTTP(rw, req)

	rw1 := httptest.NewRecorder()
	req1 := httptest.NewRequest(http.MethodPost, "/", reader)
	h.ServeHTTP(rw1, req1)
}
