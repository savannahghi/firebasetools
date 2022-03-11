package firebasetools_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/savannahghi/firebasetools"
	"github.com/stretchr/testify/assert"
)

func TestValidateLoginCreds(t *testing.T) {
	goodCreds := &firebasetools.LoginCredentials{
		Username: "yusa",
		Password: "pass",
	}
	goodCredsJSONBytes, err := json.Marshal(goodCreds)
	if err != nil {
		t.Errorf("failed to marshall payload")
		return
	}

	validRequest := httptest.NewRequest(http.MethodGet, "/", nil)
	validRequest.Body = ioutil.NopCloser(bytes.NewReader(goodCredsJSONBytes))

	emptyCredsRequest := httptest.NewRequest(http.MethodGet, "/", nil)
	emptyCredsRequest.Body = ioutil.NopCloser(bytes.NewReader([]byte{}))

	type args struct {
		w http.ResponseWriter
		r *http.Request
	}
	tests := []struct {
		name    string
		args    args
		want    *firebasetools.LoginCredentials
		wantErr bool
	}{
		{
			name: "valid creds",
			args: args{
				w: httptest.NewRecorder(),
				r: validRequest,
			},
			want: &firebasetools.LoginCredentials{
				Username: "yusa",
				Password: "pass",
			},
			wantErr: false,
		},
		{
			name: "invalid creds",
			args: args{
				w: httptest.NewRecorder(),
				r: emptyCredsRequest,
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := firebasetools.ValidateLoginCreds(tt.args.w, tt.args.r)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateLoginCreds() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ValidateLoginCreds() = %v, want %v", got, tt.want)
			}
		})
	}

}

func TestGetFirebaseUser(t *testing.T) {
	type args struct {
		ctx   context.Context
		creds *firebasetools.LoginCredentials
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Sad Case - nil creds",
			args: args{
				ctx:   context.Background(),
				creds: nil,
			},
			wantErr: true,
		},
		{
			name: "Sad Case - Bad creds",
			args: args{
				ctx: context.Background(),
				creds: &firebasetools.LoginCredentials{
					Username: "non",
					Password: "existent",
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := firebasetools.GetFirebaseUser(tt.args.ctx, tt.args.creds)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetFirebaseUser() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if got == nil {
					t.Errorf("expected a response but got %v", got)
				}
			}
		})
	}
}

func TestGetLoginFunc(t *testing.T) {
	ctx := context.Background()
	fc := &firebasetools.FirebaseClient{}
	loginFunc := firebasetools.GetLoginFunc(ctx, fc)

	incorrectLoginCredsJSONBytes, err := json.Marshal(&firebasetools.LoginCredentials{
		Username: "not a real username",
		Password: "not a real password",
	})
	if err != nil {
		t.Errorf("failed to marshall payload")
		return
	}
	incorrectLoginCredsReq := httptest.NewRequest(http.MethodGet, "/", nil)
	incorrectLoginCredsReq.Body = ioutil.NopCloser(bytes.NewReader(incorrectLoginCredsJSONBytes))

	wrongFormatLoginCredsJSONBytes, err := json.Marshal(&firebasetools.AuditLog{})
	if err != nil {
		t.Errorf("failed to marshall payload")
		return
	}
	wrongFormatLoginCredsReq := httptest.NewRequest(http.MethodGet, "/", nil)
	wrongFormatLoginCredsReq.Body = ioutil.NopCloser(bytes.NewReader(wrongFormatLoginCredsJSONBytes))

	type args struct {
		w http.ResponseWriter
		r *http.Request
	}
	tests := []struct {
		name           string
		args           args
		wantStatusCode int
	}{
		{
			name: "Sad Case - invalid login credentials - format",
			args: args{
				w: httptest.NewRecorder(),
				r: wrongFormatLoginCredsReq,
			},
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name: "Sad Case - incorrect login credentials - good format but won't login",
			args: args{
				w: httptest.NewRecorder(),
				r: incorrectLoginCredsReq,
			},
			wantStatusCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loginFunc(tt.args.w, tt.args.r)
			rec, ok := tt.args.w.(*httptest.ResponseRecorder)
			assert.True(t, ok)
			assert.NotNil(t, rec)

			assert.Equal(t, rec.Code, tt.wantStatusCode)

		})
	}
}
