package firebasetools

import (
	"context"
	"fmt"
	"net/http"

	"firebase.google.com/go/auth"
	"github.com/savannahghi/errorcodeutil"
	"github.com/savannahghi/serverutils"
)

// ValidateLoginCreds checks that the credentials supplied in the indicated request are valid
func ValidateLoginCreds(w http.ResponseWriter, r *http.Request) (*LoginCredentials, error) {
	creds := &LoginCredentials{}
	serverutils.DecodeJSONToTargetStruct(w, r, creds)
	if creds.Username == "" || creds.Password == "" {
		err := fmt.Errorf("invalid credentials, expected a username AND password")
		serverutils.WriteJSONResponse(w, errorcodeutil.CustomError{
			Err:     err,
			Message: err.Error(),
		}, http.StatusBadRequest)
		return nil, err
	}
	return creds, nil
}

// GetFirebaseUser logs in the user with the supplied credentials and returns their
// Firebase auth user record
func GetFirebaseUser(ctx context.Context, creds *LoginCredentials) (*auth.UserRecord, error) {
	if creds == nil {
		return nil, fmt.Errorf("nil creds, can't get firebase user")
	}
	user, err := GetOrCreateFirebaseUser(ctx, creds.Username)
	if err != nil {
		return nil, err
	}
	return user, nil
}

// GetLoginFunc returns a function that can authenticate against Firebase
func GetLoginFunc(ctx context.Context, fc IFirebaseClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		creds, err := ValidateLoginCreds(w, r)
		if err != nil {
			serverutils.WriteJSONResponse(w, errorcodeutil.CustomError{
				Err:     err,
				Message: err.Error(),
			}, http.StatusBadRequest)
			return
		}

		firebaseUser, err := GetFirebaseUser(ctx, creds)
		if err != nil {
			serverutils.WriteJSONResponse(w, errorcodeutil.CustomError{
				Err:     err,
				Message: err.Error(),
			}, http.StatusInternalServerError)
			return
		}

		customToken, err := CreateFirebaseCustomToken(ctx, firebaseUser.UID)
		if err != nil {
			serverutils.WriteJSONResponse(w, errorcodeutil.CustomError{
				Err:     err,
				Message: err.Error(),
			}, http.StatusInternalServerError)
			return
		}

		userTokens, err := AuthenticateCustomFirebaseToken(customToken)
		if err != nil {
			serverutils.WriteJSONResponse(w, errorcodeutil.CustomError{
				Err:     err,
				Message: err.Error(),
			}, http.StatusInternalServerError)
			return
		}

		loginResp := LoginResponse{
			CustomToken:   customToken,
			ExpiresIn:     serverutils.ConvertStringToInt(w, userTokens.ExpiresIn),
			IDToken:       userTokens.IDToken,
			RefreshToken:  userTokens.RefreshToken,
			UID:           firebaseUser.UID,
			Email:         firebaseUser.Email,
			DisplayName:   firebaseUser.DisplayName,
			EmailVerified: firebaseUser.EmailVerified,
			PhoneNumber:   firebaseUser.PhoneNumber,
			PhotoURL:      firebaseUser.PhotoURL,
			Disabled:      firebaseUser.Disabled,
			TenantID:      firebaseUser.TenantID,
			ProviderID:    firebaseUser.ProviderID,
		}
		serverutils.WriteJSONResponse(w, loginResp, http.StatusOK)
	}
}
