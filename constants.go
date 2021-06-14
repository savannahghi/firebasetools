package firebase_tools

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

	// HTTP client settings
	HTTPClientTimeoutSecs = 10
	// TestUserEmail is used by integration tests
	TestUserEmail = "be.well@bewell.co.ke"

	// e.g https://healthcloud.page.link or https://bwl.page.link
	FDLDomainEnvironmentVariableName = "FIREBASE_DYNAMIC_LINKS_DOMAIN"

	// DefaultPageSize is used to paginate records (e.g those fetched from Firebase)
	// if there is no user specified page size
	DefaultPageSize = 100

	// Sep is a separator, used to create "opaque" IDs
	Sep = "|"
)
