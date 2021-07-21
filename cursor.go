package firebasetools

import (
	"encoding/base64"
	"fmt"
	"log"

	"github.com/vmihailenco/msgpack"
)

// Cursor represents an opaque "position" for a record, for use in pagination
type Cursor struct {
	Offset int `json:"offset"`
}

// NewCursor creates a cursor from an offset and ID
func NewCursor(offset int) *Cursor {
	return &Cursor{Offset: offset}
}

// EncodeCursor converts a cursor to a string
func EncodeCursor(cursor *Cursor) string {
	b, err := msgpack.Marshal(cursor)
	if err != nil {
		msg := fmt.Sprintf("unable to encode cursor: %s", err)
		log.Println(msg)
		return msg
	}
	return base64.StdEncoding.EncodeToString(b)
}

// CreateAndEncodeCursor creates a cursor and immediately encodes it.
// It panics if it cannot encode the cursor.
// These cursors use ZERO BASED indexing.
func CreateAndEncodeCursor(offset int) *string {
	c := NewCursor(offset)
	enc := EncodeCursor(c)
	return &enc
}
