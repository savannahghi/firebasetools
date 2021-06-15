package firebasetools_test

import (
	"bytes"
	"strconv"
	"testing"

	fb "github.com/savannahghi/firebasetools"
)

func TestFieldType_IsValid(t *testing.T) {
	tests := []struct {
		name string
		e    fb.FieldType
		want bool
	}{
		{
			name: "valid string field type",
			e:    fb.FieldTypeString,
			want: true,
		},
		{
			name: "invalid field type",
			e:    fb.FieldType("this is not a real field type"),
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.e.IsValid(); got != tt.want {
				t.Errorf("FieldType.IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFieldType_String(t *testing.T) {
	tests := []struct {
		name string
		e    fb.FieldType
		want string
	}{
		{
			name: "valid boolean field type as string",
			e:    fb.FieldTypeBoolean,
			want: "BOOLEAN",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.e.String(); got != tt.want {
				t.Errorf("FieldType.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFieldType_UnmarshalGQL(t *testing.T) {
	intEnum := fb.FieldType("")
	invalid := fb.FieldType("")
	type args struct {
		v interface{}
	}
	tests := []struct {
		name    string
		e       *fb.FieldType
		args    args
		wantErr bool
	}{
		{
			name: "valid integer enum",
			e:    &intEnum,
			args: args{
				v: "INTEGER",
			},
			wantErr: false,
		},
		{
			name: "invalid enum",
			e:    &invalid,
			args: args{
				v: "NOT A VALID ENUM",
			},
			wantErr: true,
		},
		{
			name: "wrong type -int",
			e:    &invalid,
			args: args{
				v: 1,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.e.UnmarshalGQL(tt.args.v); (err != nil) != tt.wantErr {
				t.Errorf("FieldType.UnmarshalGQL() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestFieldType_MarshalGQL(t *testing.T) {
	tests := []struct {
		name  string
		e     fb.FieldType
		wantW string
	}{
		{
			name:  "number field type",
			e:     fb.FieldTypeNumber,
			wantW: strconv.Quote("NUMBER"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &bytes.Buffer{}
			tt.e.MarshalGQL(w)
			if gotW := w.String(); gotW != tt.wantW {
				t.Errorf("FieldType.MarshalGQL() = %v, want %v", gotW, tt.wantW)
			}
		})
	}
}

func TestOperation_IsValid(t *testing.T) {
	tests := []struct {
		name string
		e    fb.Operation
		want bool
	}{
		{
			name: "valid operation",
			e:    fb.OperationEqual,
			want: true,
		},
		{
			name: "invalid operation",
			e:    fb.Operation("hii sio valid"),
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.e.IsValid(); got != tt.want {
				t.Errorf("Operation.IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestOperation_String(t *testing.T) {
	tests := []struct {
		name string
		e    fb.Operation
		want string
	}{
		{
			name: "valid case - contains",
			e:    fb.OperationContains,
			want: "CONTAINS",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.e.String(); got != tt.want {
				t.Errorf("Operation.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestOperation_UnmarshalGQL(t *testing.T) {
	valid := fb.Operation("")
	invalid := fb.Operation("")
	type args struct {
		v interface{}
	}
	tests := []struct {
		name    string
		e       *fb.Operation
		args    args
		wantErr bool
	}{
		{
			name: "valid case",
			e:    &valid,
			args: args{
				v: "CONTAINS",
			},
			wantErr: false,
		},
		{
			name: "invalid string value",
			e:    &invalid,
			args: args{
				v: "NOT A REAL OPERATION",
			},
			wantErr: true,
		},
		{
			name: "invalid non string value",
			e:    &invalid,
			args: args{
				v: 1,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.e.UnmarshalGQL(tt.args.v); (err != nil) != tt.wantErr {
				t.Errorf("Operation.UnmarshalGQL() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestOperation_MarshalGQL(t *testing.T) {
	tests := []struct {
		name  string
		e     fb.Operation
		wantW string
	}{
		{
			name:  "good case",
			e:     fb.OperationContains,
			wantW: strconv.Quote("CONTAINS"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &bytes.Buffer{}
			tt.e.MarshalGQL(w)
			if gotW := w.String(); gotW != tt.wantW {
				t.Errorf("Operation.MarshalGQL() = %v, want %v", gotW, tt.wantW)
			}
		})
	}
}

func TestSortOrder_String(t *testing.T) {
	tests := []struct {
		name string
		e    fb.SortOrder
		want string
	}{
		{
			name: "good case",
			e:    fb.SortOrderAsc,
			want: "ASC",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.e.String(); got != tt.want {
				t.Errorf("SortOrder.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSortOrder_UnmarshalGQL(t *testing.T) {
	so := fb.SortOrder("")
	type args struct {
		v interface{}
	}
	tests := []struct {
		name    string
		e       *fb.SortOrder
		args    args
		wantErr bool
	}{
		{
			name: "valid sort order",
			e:    &so,
			args: args{
				v: "ASC",
			},
			wantErr: false,
		},
		{
			name: "invalid sort order string",
			e:    &so,
			args: args{
				v: "not a valid sort order",
			},
			wantErr: true,
		},
		{
			name: "invalid sort order - non string",
			e:    &so,
			args: args{
				v: 1,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.e.UnmarshalGQL(tt.args.v); (err != nil) != tt.wantErr {
				t.Errorf("SortOrder.UnmarshalGQL() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSortOrder_MarshalGQL(t *testing.T) {
	tests := []struct {
		name  string
		e     fb.SortOrder
		wantW string
	}{
		{
			name:  "good case",
			e:     fb.SortOrderDesc,
			wantW: strconv.Quote("DESC"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &bytes.Buffer{}
			tt.e.MarshalGQL(w)
			if gotW := w.String(); gotW != tt.wantW {
				t.Errorf("SortOrder.MarshalGQL() = %v, want %v", gotW, tt.wantW)
			}
		})
	}
}
