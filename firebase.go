package firebase_tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go"
	"firebase.google.com/go/auth"
	"firebase.google.com/go/messaging"
	"github.com/lithammer/shortuuid"
	"github.com/savannahghi/server_utils"
	"google.golang.org/api/option"
)

// FirebaseTokenExchangePayload is marshalled into JSON and sent to the Firebase Auth REST API
// when exchanging a custom token for an ID token that can be used to make API calls
type FirebaseTokenExchangePayload struct {
	Token             string `json:"token"`
	ReturnSecureToken bool   `json:"returnSecureToken"`
}

// FirebaseUserTokens is the unmarshalling target for the JSON response received from the Firebase Auth REST API
// when exchanging a custom token for an ID token that can be used to make API calls
type FirebaseUserTokens struct {
	IDToken      string `json:"idToken"`
	RefreshToken string `json:"refreshToken"`
	ExpiresIn    string `json:"expiresIn"`
}

// IFirebaseApp is an interface that has been extracted in order to support mocking of Firebase auth in tests
type IFirebaseApp interface {
	Auth(ctx context.Context) (*auth.Client, error)
	Firestore(ctx context.Context) (*firestore.Client, error)
	Messaging(ctx context.Context) (*messaging.Client, error)
}

// IFirebaseClient defines the Firebase methods that we depend on
// It has been defined in order to facilitate mocking for tests
type IFirebaseClient interface {
	InitFirebase() (IFirebaseApp, error)
}

// FirebaseClient is an implementation of the FirebaseClient interface
type FirebaseClient struct{}

// InitFirebase ensures that we have a working Firebase configuration
func (fc *FirebaseClient) InitFirebase() (IFirebaseApp, error) {
	appCreds, err := server_utils.GetEnvVar(GoogleApplicationCredentialsEnvVarName)
	if err != nil {
		return firebase.NewApp(
			context.Background(),
			nil,
			option.WithCredentialsFile(appCreds),
		)
	}
	return firebase.NewApp(
		context.Background(),
		nil,
	)
}

// AuthenticateCustomFirebaseToken takes a custom Firebase auth token and tries to fetch an ID token
// If successful, a pointer to the ID token is returned
// Otherwise, an error is returned
func AuthenticateCustomFirebaseToken(customAuthToken string) (*FirebaseUserTokens, error) {
	apiKey, apiKeyErr := server_utils.GetEnvVar(FirebaseWebAPIKeyEnvVarName)
	if apiKeyErr != nil {
		return nil, apiKeyErr
	}

	payload := FirebaseTokenExchangePayload{
		Token:             customAuthToken,
		ReturnSecureToken: true,
	}
	payloadBytes, _ := json.Marshal(payload) // err intentionally ignored, static typing makes it very hard to get this error

	url := FirebaseCustomTokenSigninURL + apiKey
	httpClient := http.DefaultClient
	httpClient.Timeout = time.Second * HTTPClientTimeoutSecs
	resp, err := httpClient.Post(url, "application/json", bytes.NewReader(payloadBytes))
	defer CloseRespBody(resp)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		bs, err := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf(
			"firebase HTTP error, status code %d\nBody: %s\nBody read error: %s", resp.StatusCode, string(bs), err)
	}
	var tokenResp FirebaseUserTokens
	unmarshalErr := json.NewDecoder(resp.Body).Decode(&tokenResp)
	if unmarshalErr != nil {
		return nil, unmarshalErr
	}
	return &tokenResp, nil
}

// CreateFirebaseCustomToken creates a custom auth token for the user with the
// indicated UID
func CreateFirebaseCustomToken(ctx context.Context, uid string) (string, error) {
	fc := &FirebaseClient{}
	firebaseApp, err := fc.InitFirebase()
	if err != nil {
		return "", fmt.Errorf("unable to initialize Firebase app: %w", err)
	}
	authClient, err := firebaseApp.Auth(ctx)
	if err != nil {
		return "", fmt.Errorf("unable to create custom Firebase token: %w", err)
	}
	return authClient.CustomToken(ctx, uid)
}

// GetOrCreateFirebaseUser retrieves the user record of the user with the given email
// or creates a new one if no user has the specified email
func GetOrCreateFirebaseUser(ctx context.Context, email string) (*auth.UserRecord, error) {
	authClient, err := GetFirebaseAuthClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to get or create Firebase client: %w", err)
	}
	existingUser, userErr := authClient.GetUserByEmail(ctx, email)
	if userErr == nil {
		return existingUser, nil
	}

	// try creating, assume the user could not be found
	params := (&auth.UserToCreate{}).
		Email(email).
		EmailVerified(false).
		Disabled(false)
	newUser, createErr := authClient.CreateUser(ctx, params)
	if createErr != nil {
		return nil, createErr
	}
	return newUser, nil
}

// GetFirebaseAuthClient initializes a Firebase Authentication client
func GetFirebaseAuthClient(ctx context.Context) (*auth.Client, error) {
	fc := &FirebaseClient{}
	firebaseApp, err := fc.InitFirebase()
	if err != nil {
		return nil, fmt.Errorf("unable to initialize Firebase app: %w", err)
	}
	client, err := firebaseApp.Auth(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to initialize Firebase auth client: %w", err)
	}
	return client, nil
}

// ValidateBearerToken checks the bearer token for validity against Firebase
func ValidateBearerToken(ctx context.Context, token string) (*auth.Token, error) {
	fc := &FirebaseClient{}
	firebaseApp, err := fc.InitFirebase()
	if err != nil {
		return nil, fmt.Errorf("can't initialize Firebase: %w", err)
	}

	client, clientErr := firebaseApp.Auth(ctx)
	if clientErr != nil {
		return nil, fmt.Errorf("error getting Auth client: " + clientErr.Error())
	}

	verifiedToken, verifyErr := client.VerifyIDToken(ctx, token)
	if verifyErr != nil {
		return nil, fmt.Errorf("invalid auth token: " + verifyErr.Error())
	}

	return verifiedToken, nil
}

// SaveDataToFirestore takes the supplied data (which can be a map of string to
// interface{} or a struct with json/firestore tags), a collection name and an
// intialized firestore client then tries to save the data to that collection.
func SaveDataToFirestore(firestoreClient *firestore.Client, collection string,
	data interface{}) (string, error) {
	ctx := context.Background()
	docRef, _, err := firestoreClient.Collection(collection).Add(ctx, data)
	if err != nil {
		return "", err
	}
	return docRef.ID, err
}

// UpdateRecordOnFirestore takes the supplied data (which can be a map of string to
// interface{} or a struct with json/firestore tags), a collection name and an
// intialized firestore client then tries to update the data in that object
func UpdateRecordOnFirestore(
	firestoreClient *firestore.Client,
	collection string,
	id string,
	data interface{},
) error {
	ctx := context.Background()
	_, err := firestoreClient.Collection(collection).Doc(id).Set(ctx, data)
	if err != nil {
		return err
	}
	return nil
}

// GenerateSafeIdentifier generates a shortened alphanumeric identifier.
func GenerateSafeIdentifier() string {
	return shortuuid.New()
}

// GetUserTokenFromContext retrieves a Firebase *auth.Token from the supplied context
func GetUserTokenFromContext(ctx context.Context) (*auth.Token, error) {
	val := ctx.Value(AuthTokenContextKey)
	if val == nil {
		return nil, fmt.Errorf(
			"unable to get auth token from context with key %#v", AuthTokenContextKey)
	}

	token, ok := val.(*auth.Token)
	if !ok {
		return nil, fmt.Errorf("wrong auth token type, got %#v, expected a Firebase *auth.Token", val)
	}
	return token, nil
}

// CheckIsAnonymousUser determines if the logged in user is an anonymous user
func CheckIsAnonymousUser(ctx context.Context) (bool, error) {
	authToken, err := GetUserTokenFromContext(ctx)
	if err != nil {
		return false, fmt.Errorf("user auth token not found in context: %w", err)
	}

	authClient, err := GetFirebaseAuthClient(ctx)
	if err != nil {
		return false, fmt.Errorf("unable to get or create Firebase client: %w", err)
	}

	user, err := authClient.GetUser(ctx, authToken.UID)
	if err != nil {
		return false, fmt.Errorf("unable to get user: %w", err)
	}

	// The firebase SDK doesn't provide an isAnonymous field in a user account
	// We're making assumptions that the other fields will be null/empty making a user anonymous
	// i.e no email,phoneNumber only a uid
	if user.Email != "" || user.PhoneNumber != "" {
		return false, nil
	}

	return true, nil
}
