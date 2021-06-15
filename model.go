package firebasetools

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/url"
	"time"

	"github.com/gofrs/uuid"
)

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

//IsEntity ...
func (p PaginationInput) IsEntity() {}

// SortInput is a generic container for strongly typed sorting parameters
type SortInput struct {
	SortBy []*SortParam `json:"sortBy"`
}

//IsEntity ...
func (s SortInput) IsEntity() {}

// SortParam represents a single field sort parameter
type SortParam struct {
	FieldName string    `json:"fieldName"`
	SortOrder SortOrder `json:"sortOrder"`
}

//IsEntity ...
func (s SortParam) IsEntity() {}

// FilterInput is s generic container for strongly type filter parameters
type FilterInput struct {
	Search   *string        `json:"search"`
	FilterBy []*FilterParam `json:"filterBy"`
}

//IsEntity ...
func (f FilterInput) IsEntity() {}

// FilterParam represents a single field filter parameter
type FilterParam struct {
	FieldName           string      `json:"fieldName"`
	FieldType           FieldType   `json:"fieldType"`
	ComparisonOperation Operation   `json:"comparisonOperation"`
	FieldValue          interface{} `json:"fieldValue"`
}

//IsEntity ...
func (f FilterParam) IsEntity() {}

// FirebaseRefreshResponse is used to (de)serialize the results of a successful Firebase token refresh
type FirebaseRefreshResponse struct {
	ExpiresIn    string `json:"expires_in"`
	TokenType    string `json:"token_type"`
	RefreshToken string `json:"refresh_token"`
	IDToken      string `json:"id_token"`
	UserID       string `json:"user_id"`
	ProjectID    string `json:"project_id"`
}

// ID is fulfilled by all stringifiable types.
// A valid Relay ID must fulfill this interface.
type ID interface {
	fmt.Stringer
}

// Node is a Relay (GraphQL Relay) node.
// Any valid type in this server should be a node.
type Node interface {
	IsNode()
	GetID() ID
	SetID(string)
}

// IDValue represents GraphQL object identifiers
type IDValue string

func (val IDValue) String() string { return string(val) }

// Typeof returns the type name for the supplied value
func Typeof(v interface{}) string {
	return fmt.Sprintf("%T", v)
}

// MarshalID get's a re-fetchable GraphQL Relay ID that combines an objects's ID with it's type
// and encodes it into an "opaque" Base64 string.
func MarshalID(id string, n Node) ID {
	nodeType := Typeof(n)
	combinedID := fmt.Sprintf("%s%s%s", id, Sep, nodeType)
	return IDValue(base64.StdEncoding.EncodeToString([]byte(combinedID)))
}

// PageInfo is used to add pagination information to Relay edges.
type PageInfo struct {
	HasNextPage     bool    `json:"hasNextPage"`
	HasPreviousPage bool    `json:"hasPreviousPage"`
	StartCursor     *string `json:"startCursor"`
	EndCursor       *string `json:"endCursor"`
}

//IsEntity ...
func (p PageInfo) IsEntity() {}

// NewString returns a pointer to the supplied string.
func NewString(s string) *string {
	return &s
}

// AuditLog records changes made to models
type AuditLog struct {
	ID        uuid.UUID
	RecordID  uuid.UUID        // ID of the audited record
	TypeName  string           // type of the audited record
	Operation string           // e.g pre_save, post_save
	When      time.Time        // timestamp of the operation
	UID       string           // UID of the involved user
	JSON      *json.RawMessage // serialized JSON snapshot
}

// Model defines common behavior for our models.
// It is also an ideal place to place hooks that are common to all models
// e.g audit, streaming analytics etc.
// CAUTION: Model should be evolved with cautions, because of migrations.
type Model struct {
	ID string `json:"id" firestore:"id"`

	// All models have a non nullable name field
	// If a derived model does not need this, it should use a placeholder e.g "-"
	Name string `json:"name" firestore:"name,omitempty"`

	// All records have an optional description
	Description string `json:"description" firestore:"description,omitempty"`

	// bug alert! If you add "omitempty" to the firestore struct tag, `false`
	// values will not be saved
	Deleted bool `json:"deleted,omitempty" firestore:"deleted"`

	// This is used for audit tracking but is not saved or serialized
	CreatedByUID string `json:"createdByUID" firestore:"createdByUID,omitempty"`
	UpdatedByUID string `json:"updatedByUID" firestore:"updatedByUID,omitempty"`
}

// IsNode is a "label" that marks this struct (and those that embed it) as
// implementations of the "Base" interface defined in our GraphQL schema.
func (c *Model) IsNode() {}

// GetID returns the struct's ID value
func (c *Model) GetID() ID {
	return IDValue(c.ID)
}

// SetID sets the struct's ID value
func (c *Model) SetID(id string) {
	c.ID = id
}

//IsEntity ...
func (c Model) IsEntity() {}
