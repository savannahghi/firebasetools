// Package client provides general purpose facilities for interaction with Slade 360 APIs.
// These APIs include but are not limited to:
//
// - Slade 360 EDI
// - Slade 360 Charge Master
// - Slade 360 Authentication Server
// - Slade 360 ERP
// - Slade 360 Health Passport
// etc (any other server that speaks HTTP and uses our auth server)
//
// It also provides some shared (cross-server) infrastructure and authentication (auth server)
// support.
package client

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/pkg/errors"
	"moul.io/http2curl"
)

// validation constants
const (
	tokenMinLength       = 12
	apiPasswordMinLength = 3
	tokenExpiryRatio     = 0.95 // Refresh access tokens after 95% of the time is spent
	meURLFragment        = "v1/user/me/?format=json"

	// BeWellVirtualPayerSladeCode is the Slade Code for the virtual provider used by the Be.Well app for e.g telemedicine
	BeWellVirtualPayerSladeCode = 2019 // PRO-4683

	// BeWellVirtualProviderSladeCode is the Slade Code for the virtual payer used by the Be.Well app for e.g healthcare lending
	BeWellVirtualProviderSladeCode = 4683 // PAY-2019

	// DefaultRESTAPIPageSize is the page size to use when calling Slade REST API services if the
	// client does not specify a page size
	DefaultRESTAPIPageSize = 100

	// MaxRestAPIPageSize is the largest page size we'll request
	MaxRestAPIPageSize = 250

	// AppName is the name of "this server"
	AppName = "api-gateway"

	// DSNEnvVarName is the Sentry reporting config
	DSNEnvVarName = "SENTRY_DSN"

	// AppVersion is the app version (used for StackDriver error reporting)
	AppVersion = "0.0.1"

	// PortEnvVarName is the name of the environment variable that defines the
	// server port
	PortEnvVarName = "PORT"

	// DefaultPort is the default port at which the server listens if the port
	// environment variable is not set
	DefaultPort = "8080"

	// BearerTokenPrefix is the prefix that comes before the authorization token
	// in the authorization header
	BearerTokenPrefix = "Bearer "

	// GoogleCloudProjectIDEnvVarName is used to determine the ID of the GCP project e.g for setting up StackDriver client
	GoogleCloudProjectIDEnvVarName = "GOOGLE_CLOUD_PROJECT"

	// DebugEnvVarName is used to determine if we should print extended tracing / logging (debugging aids)
	// to the console
	DebugEnvVarName = "DEBUG"

	// TestsEnvVarName is used to determine if we are running in a test environment
	IsRunningTestsEnvVarName = "IS_RUNNING_TESTS"

	// CIEnvVarName is set to "true" in CI enviroments e.g Gitlab CI, Github actions etc.
	// It can be used to opt in to / out of tests in such environments
	CIEnvVarName = "CI"
)

// OAUTHResponse holds OAuth2 tokens and scope, to be referred to when composing Authentication headers
// and when checking permissions
type OAUTHResponse struct {
	Scope        string `json:"scope"`
	ExpiresIn    int    `json:"expires_in"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
}

// Client describes the interface that a Slade 360 EDI client should offer
// It is used extensively in tests (to mock responses)
type Client interface {
	IsInitialized() bool
	Refresh() error
	Authenticate() error
	MakeRequest(method string, url string, body io.Reader) (*http.Response, error)
	APIScheme() string
	APIHost() string
	HTTPClient() *http.Client
	AccessToken() string
	TokenType() string
	RefreshToken() string
	AccessScope() string
	ExpiresIn() int
	RefreshAt() time.Time
	MeURL() (string, error)
	ClientID() string
	ClientSecret() string
	APITokenURL() string
	GrantType() string
	Username() string
	Password() string

	// private setters
	setInitialized(b bool)
	updateAuth(authResp *OAUTHResponse)
}

func boolEnv(envVarName string) bool {
	envVar, err := GetEnvVar(envVarName)
	if err != nil {
		return false
	}
	val, err := strconv.ParseBool(envVar)
	if err != nil {
		return false
	}
	return val
}

// IsDebug returns true if debug has been turned on in the environment
func IsDebug() bool {
	return boolEnv(DebugEnvVarName)
}

// IsRunningTests returns true if debug has been turned on in the environment
func IsRunningTests() bool {
	return boolEnv(IsRunningTestsEnvVarName)
}

// IsCI returns true when running in CI environments that set the CI env var
// e.g Gitlab CI, Github Actions, CircleCI etc
func IsCI() bool {
	return boolEnv(CIEnvVarName)
}

// GetEnvVar retrieves the environment variable with the supplied name and fails
// if it is not able to do so
func GetEnvVar(envVarName string) (string, error) {
	envVar := os.Getenv(envVarName)
	if envVar == "" {
		envErrMsg := fmt.Sprintf("the environment variable '%s' is not set", envVarName)
		return "", errors.New(envErrMsg)
	}
	return envVar, nil
}

// MustGetEnvVar returns the value of the environment variable with the indicated name or panics.
// It is intended to be used in the INTERNALS of the server when we can guarantee (through orderly
// coding) that the environment variable was set at server startup.
func MustGetEnvVar(envVarName string) string {
	val, err := GetEnvVar(envVarName)
	if err != nil {
		msg := fmt.Sprintf("mandatory environment variable %s not found", envVarName)
		log.Panicf(msg)
	}
	return val
}

// NewServerClient initializes a generic OAuth2 + HTTP server client
func NewServerClient(clientID string, clientSecret string, apiTokenURL string, apiHost string, apiScheme string, grantType string, username string, password string, extraHeaders map[string]string) (*ServerClient, error) {
	c := ServerClient{
		clientID:     clientID,
		clientSecret: clientSecret,
		apiTokenURL:  apiTokenURL,
		apiHost:      apiHost,
		apiScheme:    apiScheme,
		grantType:    grantType,
		username:     username,
		password:     password,
	}
	clientErr := c.Initialize()
	if clientErr != nil {
		return nil, clientErr
	}
	// used to set e.g Slade 360 ERP's X-Workstation header
	if extraHeaders != nil {
		c.extraHeaders = extraHeaders
	}
	return &c, nil
}

// ServerClient is a general purpose client for interacting with servers that:
//
//  1. Offer a HTTP API (it need not be RESTful)
//  2. Support OAuth2 authentication with the password grant type
//
// Examples of such servers in the Slade 360 ecosystem are:
//
//  1. Slade 360 EDI
//  2. Slade 360 Auth Server
//  3. Slade 360 ERP
//  4. Slade 360 Clinical / HealthCloud API
//  5. Slade 360 Charge Master
//  6. Slade 360 (Provider) Integration Services
//  7. Slade 360 Payer Integration Services
//  ... any other HTTP server that talks OAuth2 and supports the password grant type
//
// ServerClient MUST be configured by calling the `Initialize` method.
type ServerClient struct {
	// key connec
	clientID         string
	clientSecret     string
	apiTokenURL      string
	authServerDomain string
	apiHost          string
	apiScheme        string
	grantType        string
	username         string
	password         string
	extraHeaders     map[string]string // optional extra headers

	// these fields are set by the constructor upon successful initialization
	httpClient   *http.Client
	accessToken  string
	tokenType    string
	refreshToken string
	accessScope  string
	expiresIn    int
	refreshAt    time.Time

	// sentinel value to simplify later checks
	isInitialized bool
}

// MeURL calculates and returns the EDI user profile URL
func (c *ServerClient) MeURL() (string, error) {
	parsedTokenURL, parseErr := url.Parse(c.apiTokenURL)
	if parseErr != nil {
		return "", parseErr
	}
	meURL := fmt.Sprintf("%s://%s/%s", parsedTokenURL.Scheme, parsedTokenURL.Host, meURLFragment)
	return meURL, nil
}

// Refresh uses the refresh token to obtain a fresh access token
func (c *ServerClient) Refresh() error {
	if !c.IsInitialized() {
		return errors.New("cannot Refresh API tokens on an uninitialized client")
	}
	refreshData := url.Values{}
	refreshData.Set("client_id", c.clientID)
	refreshData.Set("client_secret", c.clientSecret)
	refreshData.Set("grant_type", "refresh_token")
	refreshData.Set("refresh_token", c.refreshToken)
	encodedRefreshData := strings.NewReader(refreshData.Encode())
	resp, err := c.httpClient.Post(c.apiTokenURL, "application/x-www-form-urlencoded", encodedRefreshData)
	if err != nil {
		return err
	}
	if resp != nil && (resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices) {
		msg := fmt.Sprintf("server error status: %d", resp.StatusCode)
		return errors.New(msg)
	}
	authResp, decodeErr := decodeOauthResponseFromJSON(resp)
	if decodeErr != nil {
		return decodeErr
	}
	c.updateAuth(authResp)
	return nil
}

// updateAuth updates the tokens stored on the EDI API client after successful authentication or refreshes
func (c *ServerClient) updateAuth(authResp *OAUTHResponse) {
	c.accessToken = authResp.AccessToken
	c.tokenType = authResp.TokenType
	c.accessScope = authResp.Scope
	c.refreshToken = authResp.RefreshToken
	c.expiresIn = authResp.ExpiresIn

	// wait out most of the token's duration to expiry before attempting to Refresh
	secondsToRefresh := int(float64(c.expiresIn) * tokenExpiryRatio)
	c.refreshAt = time.Now().Add(time.Second * time.Duration(secondsToRefresh))
	c.isInitialized = true
}

// Authenticate uses client credentials stored on the client to log in to a Slade 360 authentication server
// and update stored credentials
func (c *ServerClient) Authenticate() error {
	if err := CheckAPIClientPreconditions(c); err != nil {
		return errors.Wrap(err, "Authenticate_CheckEDIClientPreconditions")
	}
	credsData := url.Values{}
	credsData.Set("client_id", c.clientID)
	credsData.Set("client_secret", c.clientSecret)
	credsData.Set("grant_type", c.grantType)
	credsData.Set("username", c.username)
	credsData.Set("password", c.password)
	encodedCredsData := strings.NewReader(credsData.Encode())

	authResp, authErr := c.httpClient.Post(c.apiTokenURL, "application/x-www-form-urlencoded", encodedCredsData)
	if authErr != nil {
		return authErr
	}
	if authResp != nil && (authResp.StatusCode < http.StatusOK || authResp.StatusCode >= http.StatusMultipleChoices) {
		msg := fmt.Sprintf("server error status: %d", authResp.StatusCode)
		return errors.New(msg)
	}
	decodedAuthResp, decodeErr := decodeOauthResponseFromJSON(authResp)
	if decodeErr != nil {
		return decodeErr
	}
	c.updateAuth(decodedAuthResp)
	return nil // no error
}

// MakeRequest composes an authenticated EDI request that has the correct content type
func (c *ServerClient) MakeRequest(method string, url string, body io.Reader) (*http.Response, error) {
	if time.Now().UnixNano() > c.refreshAt.UnixNano() {
		refreshErr := c.Refresh()
		if refreshErr != nil {
			return nil, refreshErr
		}
	}
	req, reqErr := http.NewRequest(method, url, body)
	if reqErr != nil {
		return nil, reqErr
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.accessToken)

	// set extra headers e.g the Slade 360 ERP X-Workstation header
	if c.extraHeaders != nil {
		for k, v := range c.extraHeaders {
			req.Header.Set(k, v)
		}
	}

	if IsDebug() {
		command, _ := http2curl.GetCurlCommand(req)
		log.Printf("\nCurl command:\n%s\n", command)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	contentType := resp.Header.Get("Content-Type")
	if contentType != "application/json" {
		bs, err := ioutil.ReadAll(resp.Body)
		log.Printf("Mismatched content type error: %s\n", err)
		log.Printf("Mismatched content type body: %s\b", string(bs))
		return nil, errors.New("expected application/json Content-Type, got " + contentType)
	}
	return resp, nil
}

// Initialize MUST be used to set up a working EDI Client
func (c *ServerClient) Initialize() error {
	preErr := CheckAPIClientPreconditions(c)
	if preErr != nil {
		return preErr
	}

	// the timeout is half an hour, to match the timeout of a Cloud Run function
	// and to support somewhat long lived data "crawls"
	c.httpClient = &http.Client{Timeout: time.Second * 60 * 30}

	authErr := c.Authenticate()
	if authErr != nil {
		return authErr
	}

	checkErr := CheckAPIClientPostConditions(c)
	if checkErr != nil {
		return checkErr
	}

	c.setInitialized(true)
	return nil
}

// HTTPClient returns a properly configured HTTP client
func (c *ServerClient) HTTPClient() *http.Client {
	return c.httpClient
}

// AccessToken returns the latest access token
func (c *ServerClient) AccessToken() string {
	return c.accessToken
}

// TokenType returns the latest OAuth access token's token type
func (c *ServerClient) TokenType() string {
	return c.tokenType
}

// RefreshToken returns the latest refresh token
func (c *ServerClient) RefreshToken() string {
	return c.refreshToken
}

// AccessScope returns the latest access scope
func (c *ServerClient) AccessScope() string {
	return c.accessScope
}

// ExpiresIn returns the expiry seconds value returned after the last authentication
func (c *ServerClient) ExpiresIn() int {
	return c.expiresIn
}

// RefreshAt returns the target refresh time
func (c *ServerClient) RefreshAt() time.Time {
	return c.refreshAt
}

// APIScheme exports the configured EDI API scheme
func (c *ServerClient) APIScheme() string {
	return c.apiScheme
}

// APIHost returns the configured EDI API host
func (c *ServerClient) APIHost() string {
	return c.apiHost
}

// ClientID returns the configured client ID
func (c *ServerClient) ClientID() string {
	return c.clientID
}

// ClientSecret returns the configured client secret
func (c *ServerClient) ClientSecret() string {
	return c.clientSecret
}

// APITokenURL returns the configured API token URL on the client
func (c *ServerClient) APITokenURL() string {
	return c.apiTokenURL
}

// AuthServerDomain returns the configured auth server domain on the client
func (c *ServerClient) AuthServerDomain() string {
	return c.authServerDomain
}

// GrantType returns the configured grant type on the client
func (c *ServerClient) GrantType() string {
	return c.grantType
}

// Username returns the configured Username on the client
func (c *ServerClient) Username() string {
	return c.username
}

// Password returns the configured Password on the client
func (c *ServerClient) Password() string {
	return c.password
}

// IsInitialized returns true if the EDI httpClient is correctly initialized
func (c *ServerClient) IsInitialized() bool {
	return c.isInitialized
}

// setInitialized sets the value of the isInitialized bool
func (c *ServerClient) setInitialized(isInitialized bool) {
	c.isInitialized = isInitialized
}

// CheckAPIClientPreconditions ensures that all the parameters passed into `Initialize` make sense
func CheckAPIClientPreconditions(client Client) error {
	clientID := client.ClientID()
	if !govalidator.IsAlphanumeric(clientID) || len(clientID) < tokenMinLength {
		errMsg := fmt.Sprintf("%s is not a valid clientId, expected a non-blank alphanumeric string of at least %d characters", clientID, tokenMinLength)
		return errors.New(errMsg)
	}

	clientSecret := client.ClientSecret()
	if !govalidator.IsAlphanumeric(clientSecret) || len(clientSecret) < tokenMinLength {
		errMsg := fmt.Sprintf("%s is not a valid clientSecret, expected a non-blank alphanumeric string of at least %d characters", clientSecret, tokenMinLength)
		return errors.New(errMsg)
	}

	apiTokenURL := client.APITokenURL()
	if !govalidator.IsRequestURL(apiTokenURL) {
		errMsg := fmt.Sprintf("%s is not a valid apiTokenURL, expected an http(s) URL", apiTokenURL)
		return errors.New(errMsg)
	}

	apiHost := client.APIHost()
	if !govalidator.IsHost(apiHost) {
		errMsg := fmt.Sprintf("%s is not a valid apiHost, expected a valid IP or domain name", apiHost)
		return errors.New(errMsg)
	}

	apiScheme := client.APIScheme()
	if apiScheme != "http" && apiScheme != "https" {
		errMsg := fmt.Sprintf("%s is not a valid apiScheme, expected http or https", apiScheme)
		return errors.New(errMsg)
	}

	grantType := client.GrantType()
	if grantType != "password" {
		return errors.New("the only supported OAuth grant type for now is 'password'")
	}

	username := client.Username()
	if !govalidator.IsEmail(username) {
		return errors.New("the Username should be a valid email address")
	}

	password := client.Password()
	if len(password) < apiPasswordMinLength {
		msg := fmt.Sprintf("the Password should be a string of at least %d characters", apiPasswordMinLength)
		return errors.New(msg)
	}

	return nil
}

// CheckAPIClientPostConditions performs sanity checks on a freshly initialized EDI client
func CheckAPIClientPostConditions(client Client) error {
	accessToken := client.AccessToken()
	if !govalidator.IsAlphanumeric(accessToken) || len(accessToken) < tokenMinLength {
		return errors.New("invalid access token after EDIAPIClient initialization")
	}

	tokenType := client.TokenType()
	if tokenType != "Bearer" {
		return errors.New("invalid token type after EDIAPIClient initialization, expected 'Bearer'")
	}

	refreshToken := client.RefreshToken()
	if !govalidator.IsAlphanumeric(refreshToken) || len(refreshToken) < tokenMinLength {
		return errors.New("invalid Refresh token after EDIAPIClient initialization")
	}

	accessScope := client.AccessScope()
	if !govalidator.IsASCII(accessScope) || len(accessScope) < tokenMinLength {
		return errors.New("invalid access scope text after EDIAPIClient initialization")
	}

	expiresIn := client.ExpiresIn()
	if expiresIn < 1 {
		return errors.New("invalid expiresIn after EDIAPIClient initialization")
	}

	refreshAt := client.RefreshAt()
	if refreshAt.UnixNano() < time.Now().UnixNano() {
		return errors.New("invalid past refreshAt after EDIAPIClient initialization")
	}

	return nil // no errors found
}

// CheckAPIInitialization returns and error if the EDI httpClient was not correctly initialized by calling `.Initialize()`
func CheckAPIInitialization(client Client) error {
	if client == nil || !client.IsInitialized() {
		return errors.New("the EDI httpClient is not correctly initialized. Please use the `.Initialize` constructor")
	}
	return nil
}

// CloseRespBody closes the body of the supplied HTTP response
func CloseRespBody(resp *http.Response) {
	if resp != nil {
		err := resp.Body.Close()
		if err != nil {
			log.Println("unable to close response body for request made to ", resp.Request.RequestURI)
		}
	}
}

// ComposeAPIURL assembles an EDI URL string for the supplied path and query string
func ComposeAPIURL(client Client, path string, query string) string {
	apiURL := url.URL{
		Scheme:   client.APIScheme(),
		Host:     client.APIHost(),
		Path:     path,
		RawQuery: query,
	}
	return apiURL.String()
}

func decodeOauthResponseFromJSON(resp *http.Response) (*OAUTHResponse, error) {
	defer CloseRespBody(resp)
	var decodedAuthResp OAUTHResponse
	respBytes, readErr := ioutil.ReadAll(resp.Body)
	if readErr != nil {
		return nil, readErr
	}
	decodeErr := json.Unmarshal(respBytes, &decodedAuthResp)
	if decodeErr != nil {
		return nil, decodeErr
	}
	return &decodedAuthResp, nil
}

// GetAPIPaginationParams composes pagination parameters for use by a REST API that uses
// offset based pagination e.g Slade 360 APIS
func GetAPIPaginationParams(pagination *PaginationInput) (url.Values, error) {
	if pagination == nil {
		return nil, nil
	}

	// Treat first or last, when set, literally as page sizes
	// We intentionally "demote" `last`; if both `first` and `last` are specified,
	// `first` will supersede `last`
	var err error
	pageSize := DefaultRESTAPIPageSize
	if pagination.Last > 0 {
		pageSize = pagination.Last
	}
	if pagination.First > 0 {
		pageSize = pagination.First
	}

	// For these "pass through APIs", "after" and "before" should be parseable as ints
	// (literal offsets).
	// We intentionally demote `before` i.e if both `before` and `after` are set,
	// `after` will supersede `before`
	offset := 0
	if pagination.Before != "" {
		offset, err = strconv.Atoi(pagination.Before)
		if err != nil {
			return nil, fmt.Errorf("expected `before` to be parseable as an int; got %s", pagination.Before)
		}
	}
	if pagination.After != "" {
		offset, err = strconv.Atoi(pagination.After)
		if err != nil {
			return nil, fmt.Errorf("expected `after` to be parseable as an int; got %s", pagination.After)
		}
	}
	page := int(offset/pageSize) + 1 // page numbers are one based
	values := url.Values{}
	values.Set("page", fmt.Sprintf("%d", page))
	values.Set("page_size", fmt.Sprintf("%d", pageSize))
	return values, nil
}

// MergeURLValues merges > 1 url.Values into one
func MergeURLValues(values ...url.Values) url.Values {
	merged := url.Values{}
	for _, value := range values {
		for k, v := range value {
			merged[k] = v
		}
	}
	return merged
}

// QueryParam is an interface used for filter and sort parameters
type QueryParam interface {
	ToURLValues() (values url.Values)
}

// PaginationInput represents paging parameters
type PaginationInput struct {
	First  int    `json:"first"`
	Last   int    `json:"last"`
	After  string `json:"after"`
	Before string `json:"before"`
}