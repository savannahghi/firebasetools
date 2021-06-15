package firebasetools_test

import (
	"reflect"
	"testing"

	fb "github.com/savannahghi/firebasetools"
	"github.com/stretchr/testify/assert"
)

func TestTypeof(t *testing.T) {
	type args struct {
		v interface{}
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "string",
			args: args{
				v: "this is a string",
			},
			want: "string",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := fb.Typeof(tt.args.v); got != tt.want {
				t.Errorf("Typeof() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIDValue_String(t *testing.T) {
	tests := []struct {
		name string
		val  fb.IDValue
		want string
	}{
		{
			name: "happy case",
			val:  fb.IDValue("mimi ni id"),
			want: "mimi ni id",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.val.String(); got != tt.want {
				t.Errorf("IDValue.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMarshalID(t *testing.T) {
	type args struct {
		id string
		n  fb.Node
	}
	tests := []struct {
		name string
		args args
		want fb.ID
	}{
		{
			name: "good case",
			args: args{
				id: "1",
				n:  &fb.Model{},
			},
			want: fb.IDValue("MXwqZmlyZWJhc2V0b29scy5Nb2RlbA=="),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := fb.MarshalID(tt.args.id, tt.args.n); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MarshalID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewString(t *testing.T) {
	ns := fb.NewString("a string")
	assert.Equal(t, "a string", *ns)
}

func TestModel_GetID(t *testing.T) {
	type fields struct {
		ID           string
		Name         string
		Description  string
		Deleted      bool
		CreatedByUID string
		UpdatedByUID string
	}
	tests := []struct {
		name   string
		fields fields
		want   fb.ID
	}{
		{
			name: "good case",
			fields: fields{
				ID: "an ID",
			},
			want: fb.IDValue("an ID"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &fb.Model{
				ID:           tt.fields.ID,
				Name:         tt.fields.Name,
				Description:  tt.fields.Description,
				Deleted:      tt.fields.Deleted,
				CreatedByUID: tt.fields.CreatedByUID,
				UpdatedByUID: tt.fields.UpdatedByUID,
			}
			if got := c.GetID(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Model.GetID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestModel_SetID(t *testing.T) {
	type fields struct {
		ID           string
		Name         string
		Description  string
		Deleted      bool
		CreatedByUID string
		UpdatedByUID string
	}
	type args struct {
		id string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "good case",
			args: args{
				id: "an ID",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &fb.Model{
				ID:           tt.fields.ID,
				Name:         tt.fields.Name,
				Description:  tt.fields.Description,
				Deleted:      tt.fields.Deleted,
				CreatedByUID: tt.fields.CreatedByUID,
				UpdatedByUID: tt.fields.UpdatedByUID,
			}
			c.SetID(tt.args.id)
			assert.Equal(t, c.GetID(), fb.IDValue(tt.args.id))
		})
	}
}

func TestModel_IsNode(t *testing.T) {
	type fields struct {
		ID           string
		Name         string
		Description  string
		Deleted      bool
		CreatedByUID string
		UpdatedByUID string
	}
	tests := []struct {
		name   string
		fields fields
	}{
		{
			name: "default case",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &fb.Model{
				ID:           tt.fields.ID,
				Name:         tt.fields.Name,
				Description:  tt.fields.Description,
				Deleted:      tt.fields.Deleted,
				CreatedByUID: tt.fields.CreatedByUID,
				UpdatedByUID: tt.fields.UpdatedByUID,
			}
			c.IsNode()
		})
	}
}

func TestModelsIsEntity(t *testing.T) {

	t14 := fb.FilterParam{}
	t14.IsEntity()

	t15 := fb.FilterInput{}
	t15.IsEntity()

	t16 := fb.SortInput{}
	t16.IsEntity()

	t17 := fb.PaginationInput{}
	t17.IsEntity()
}
