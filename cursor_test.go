package firebasetools_test

import (
	"testing"

	"github.com/savannahghi/firebasetools"
	"github.com/stretchr/testify/assert"
)

func TestCreateAndEncodeCursor(t *testing.T) {
	zeroCur := "gaZPZmZzZXTTAAAAAAAAAAA="
	negCur := "gaZPZmZzZXTT//////////8="
	type args struct {
		offset int
	}
	tests := []struct {
		name string
		args args
		want *string
	}{
		{
			name: "default zero offset cursor",
			args: args{
				offset: 0,
			},
			want: &zeroCur,
		},
		{
			name: "negative cursor",
			args: args{
				offset: -1,
			},
			want: &negCur,
		},
	}
	for _, tt := range tests {
		got := firebasetools.CreateAndEncodeCursor(tt.args.offset)
		assert.NotNil(t, got)
		assert.Equal(t, *got, *tt.want)
	}
}
