package firebase_tools

import (
	"context"
	"fmt"
	"log"
	"net/http"
)

// CloseRespBody closes the body of the supplied HTTP response
func CloseRespBody(resp *http.Response) {
	if resp != nil {
		err := resp.Body.Close()
		if err != nil {
			log.Println("unable to close response body for request made to ", resp.Request.RequestURI)
		}
	}
}

// ContextKey is used as a type for the UID key for the Firebase *auth.Token on context.Context.
// It is a custom type in order to minimize context key collissions on the context
// (.and to shut up golint).
type ContextKey string

// GetLoggedInUser retrieves logged in user information
func GetLoggedInUser(ctx context.Context) (*UserInfo, error) {
	authToken, err := GetUserTokenFromContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("user auth token not found in context: %w", err)
	}

	authClient, err := GetFirebaseAuthClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to get or create Firebase client: %w", err)
	}

	user, err := authClient.GetUser(ctx, authToken.UID)
	if err != nil {

		return nil, fmt.Errorf("unable to get user: %w", err)
	}
	return &UserInfo{
		UID:         user.UID,
		Email:       user.Email,
		PhoneNumber: user.PhoneNumber,
		DisplayName: user.DisplayName,
		ProviderID:  user.ProviderID,
		PhotoURL:    user.PhotoURL,
	}, nil
}

// UserInfo is a collection of standard profile information for a user.
type UserInfo struct {
	DisplayName string `json:"displayName,omitempty"`
	Email       string `json:"email,omitempty"`
	PhoneNumber string `json:"phoneNumber,omitempty"`
	PhotoURL    string `json:"photoUrl,omitempty"`
	// In the ProviderUserInfo[] ProviderID can be a short domain name (e.g. google.com),
	// or the identity of an OpenID identity provider.
	// In UserRecord.UserInfo it will return the constant string "firebase".
	ProviderID string `json:"providerId,omitempty"`
	UID        string `json:"rawId,omitempty"`
}
