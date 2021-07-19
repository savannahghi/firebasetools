package firebasetools

const (

	// FirebaseWebAPIKeyEnvVarName is the name of the env var that holds a Firebase web API key
	// for this project
	FirebaseWebAPIKeyEnvVarName = "FIREBASE_WEB_API_KEY"

	// FirebaseCustomTokenSigninURL is the Google Identity Toolkit API for signing in over REST
	FirebaseCustomTokenSigninURL = "https://identitytoolkit.googleapis.com/v1/accounts:signInWithCustomToken?key="

	// FirebaseRefreshTokenURL is used to request Firebase refresh tokens from Google APIs
	FirebaseRefreshTokenURL = "https://securetoken.googleapis.com/v1/token?key="

	// GoogleApplicationCredentialsEnvVarName is used to obtain service account details from the
	// local server when necessary e.g when running tests on CI or a local developer setup
	GoogleApplicationCredentialsEnvVarName = "GOOGLE_APPLICATION_CREDENTIALS"

	// GoogleProjectNumberEnvVarName is a numeric project number that
	GoogleProjectNumberEnvVarName = "GOOGLE_PROJECT_NUMBER"

	// AuthTokenContextKey is used to add/retrieve the Firebase UID on the context
	AuthTokenContextKey = ContextKey("UID")

	// HTTPClientTimeoutSecs is used to set HTTP client Timeout setting for a request
	HTTPClientTimeoutSecs = 10

	// TestUserEmail is used by integration tests
	TestUserEmail = "test@bewell.co.ke"

	// FDLDomainEnvironmentVariableName is firebase dynamic link domain/URL
	// e.g https://example-one.page.link or https://example-two.page.link
	FDLDomainEnvironmentVariableName = "FIREBASE_DYNAMIC_LINKS_DOMAIN"

	// DefaultPageSize is used to paginate records (e.g those fetched from Firebase)
	// if there is no user specified page size
	DefaultPageSize = 100

	// Sep is a separator, used to create "opaque" IDs
	Sep = "|"

	// DefaultRESTAPIPageSize is the page size to use when calling Slade REST API services if the
	// client does not specify a page size
	DefaultRESTAPIPageSize = 100
)
