package firebasetools

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"firebase.google.com/go/auth"
	"github.com/savannahghi/serverutils"
)

// authCheckFn is a function type for authorization and authentication checks
// there can be several e.g an authentication check runs first then an authorization
// check runs next if the authentication passes etc
type authCheckFn = func(
	r *http.Request,
	firebaseApp IFirebaseApp,
) (bool, map[string]string, *auth.Token)

// AuthenticationMiddleware decodes the share session cookie and packs the session into context
func AuthenticationMiddleware(firebaseApp IFirebaseApp) func(http.Handler) http.Handler {
	// multiple checks will be run in sequence (order matters)
	// the first check to succeed will call `c.Next()` and `return`
	// this means that more permissive checks (e.g exceptions) should come first
	checkFuncs := []authCheckFn{HasValidFirebaseBearerToken}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				errs := []map[string]string{}
				// in case authorization does not succeed, accumulated errors
				// are returned to the client
				for _, checkFunc := range checkFuncs {
					shouldContinue, errMap, authToken := checkFunc(r, firebaseApp)
					if shouldContinue {
						// put the auth token in the context
						ctx := context.WithValue(r.Context(), AuthTokenContextKey, authToken)

						// and call the next with our new context
						r = r.WithContext(ctx)
						next.ServeHTTP(w, r)
						return
					}
					errs = append(errs, errMap)
				}

				// if we got here, it is because we have errors.
				// write an error response)
				serverutils.WriteJSONResponse(w, errs, http.StatusUnauthorized)
			},
		)
	}
}

// HasValidFirebaseBearerToken returns true with no errors if the request has a valid bearer token in the authorization header.
// Otherwise, it returns false and the error in a map with the key "error"
func HasValidFirebaseBearerToken(r *http.Request, firebaseApp IFirebaseApp) (bool, map[string]string, *auth.Token) {
	bearerToken, err := ExtractBearerToken(r)
	if err != nil {
		// this error here will only be returned to the user if all the verification functions in the chain fail
		return false, serverutils.ErrorMap(err), nil
	}

	validToken, err := ValidateBearerToken(r.Context(), bearerToken)
	if err != nil {
		return false, serverutils.ErrorMap(err), nil
	}

	return true, nil, validToken
}

// ExtractBearerToken gets a bearer token from an Authorization header.
//
// This is expected to contain a Firebase idToken prefixed with "Bearer "
func ExtractBearerToken(r *http.Request) (string, error) {
	return ExtractToken(r, "Authorization", "Bearer")
}

// ExtractToken extracts a token with the specified prefix from the specified header
func ExtractToken(r *http.Request, header string, prefix string) (string, error) {
	if r == nil {
		return "", fmt.Errorf("nil request")
	}
	if r.Header == nil {
		return "", fmt.Errorf("no headers, can't extract bearer token")
	}
	authHeader := r.Header.Get(header)
	if authHeader == "" {
		return "", fmt.Errorf("expected an `%s` request header", header)
	}
	if !strings.HasPrefix(authHeader, prefix) {
		return "", fmt.Errorf("the `Authorization` header contents should start with `Bearer`")
	}
	tokenOnly := strings.TrimSpace(strings.TrimPrefix(authHeader, prefix))
	return tokenOnly, nil
}
