package firebase_tools_test

import (
	"context"
	"fmt"
	"reflect"
	"testing"
	"time"

	"cloud.google.com/go/firestore"
	fb "github.com/savannahghi/firebase_tools"
	base "github.com/savannahghi/go_utils"
	"github.com/stretchr/testify/assert"
)

// Dummy is a test node
type Dummy struct{}

func (d Dummy) IsNode() {}

func (d Dummy) GetID() base.ID {
	return base.IDValue("dummy id")
}

func (d Dummy) SetID(string) {}

func TestSuffixCollection_staging(t *testing.T) {
	col := "otp"
	expect := fmt.Sprintf("%v_bewell_%v", col, "staging")
	s := fb.SuffixCollection(col)
	assert.Equal(t, expect, s)
}

func TestServiceSuffixCollection_testing(t *testing.T) {
	col := "otp"
	expect := fmt.Sprintf("%v_bewell_%v", col, "staging")
	s := fb.SuffixCollection(col)
	assert.Equal(t, expect, s)
}

func Test_getCollectionName(t *testing.T) {
	n1 := &Dummy{}
	assert.Equal(t, "dummy_bewell_staging", base.GetCollectionName(n1))
}

func Test_validatePaginationParameters(t *testing.T) {
	first := 10
	after := "30"
	last := 10
	before := "20"

	tests := map[string]struct {
		pagination           *base.PaginationInput
		expectError          bool
		expectedErrorMessage string
	}{
		"first_last_specified": {
			pagination: &base.PaginationInput{
				First: first,
				Last:  last,
			},
			expectError:          true,
			expectedErrorMessage: "if `first` is specified for pagination, `last` cannot be specified",
		},
		"first_only": {
			pagination: &base.PaginationInput{
				First: first,
			},
			expectError: false,
		},
		"last_only": {
			pagination: &base.PaginationInput{
				Last: last,
			},
			expectError: false,
		},
		"first_and_after": {
			pagination: &base.PaginationInput{
				First: first,
				After: after,
			},
			expectError: false,
		},
		"last_and_before": {
			pagination: &base.PaginationInput{
				Last:   last,
				Before: before,
			},
			expectError: false,
		},
		"nil_pagination": {
			pagination:  nil,
			expectError: false,
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			err := base.ValidatePaginationParameters(tc.pagination)
			if tc.expectError {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), tc.expectedErrorMessage)
			}
			if !tc.expectError {
				assert.Nil(t, err)
			}
		})
	}
}

func Test_opstring(t *testing.T) {
	tests := map[string]struct {
		op                   base.Operation
		expectedOutput       string
		expectError          bool
		expectedErrorMessage string
	}{
		"invalid_operation": {
			op:                   base.Operation("invalid unknown operation"),
			expectedOutput:       "",
			expectError:          true,
			expectedErrorMessage: "unknown operation; did you forget to update this function after adding new operations in the schema?",
		},
		"less than": {
			op:             base.OperationLessThan,
			expectedOutput: "<",
			expectError:    false,
		},
		"less than_or_equal_to": {
			op:             base.OperationLessThanOrEqualTo,
			expectedOutput: "<=",
			expectError:    false,
		},
		"equal_to": {
			op:             base.OperationEqual,
			expectedOutput: "==",
			expectError:    false,
		},
		"greater_than": {
			op:             base.OperationGreaterThan,
			expectedOutput: ">",
			expectError:    false,
		},
		"greater_than_or_equal_to": {
			op:             base.OperationGreaterThanOrEqualTo,
			expectedOutput: ">=",
			expectError:    false,
		},
		"in": {
			op:             base.OperationIn,
			expectedOutput: "in",
			expectError:    false,
		},
		"contains": {
			op:             base.OperationContains,
			expectedOutput: "array-contains",
			expectError:    false,
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			opString, err := base.OpString(tc.op)
			assert.Equal(t, tc.expectedOutput, opString)
			if tc.expectError {
				assert.Equal(t, tc.expectedErrorMessage, err.Error())
			}
		})
	}
}

func TestGetFirestoreClient(t *testing.T) {
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "good case",
			args: args{
				ctx: context.Background(),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := base.GetFirestoreClient(tt.args.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetFirestoreClient() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.NotNil(t, got)
		})
	}
}

func TestComposeUnpaginatedQuery(t *testing.T) {
	ctx := context.Background()
	node := &base.Model{}
	sortAsc := base.SortInput{
		SortBy: []*base.SortParam{
			{
				FieldName: "name",
				SortOrder: base.SortOrderAsc,
			},
		},
	}
	sortDesc := base.SortInput{
		SortBy: []*base.SortParam{
			{
				FieldName: "name",
				SortOrder: base.SortOrderDesc,
			},
		},
	}
	invalidFilter := base.FilterInput{
		FilterBy: []*base.FilterParam{
			{
				FieldName:           "name",
				FieldType:           base.FieldTypeString,
				ComparisonOperation: base.Operation("not a valid operation"),
				FieldValue:          "val",
			},
		},
	}
	booleanFilter := base.FilterInput{
		FilterBy: []*base.FilterParam{
			{
				FieldName:           "deleted",
				FieldType:           base.FieldTypeBoolean,
				ComparisonOperation: base.OperationEqual,
				FieldValue:          "false",
			},
		},
	}
	invalidBoolFilterWrongType := base.FilterInput{
		FilterBy: []*base.FilterParam{
			{
				FieldName:           "deleted",
				FieldType:           base.FieldTypeBoolean,
				ComparisonOperation: base.OperationEqual,
				FieldValue:          false,
			},
		},
	}
	invalidBoolFilterUnparseableString := base.FilterInput{
		FilterBy: []*base.FilterParam{
			{
				FieldName:           "deleted",
				FieldType:           base.FieldTypeBoolean,
				ComparisonOperation: base.OperationEqual,
				FieldValue:          "bad format",
			},
		},
	}
	intFilter := base.FilterInput{
		FilterBy: []*base.FilterParam{
			{
				FieldName:           "count",
				FieldType:           base.FieldTypeInteger,
				ComparisonOperation: base.OperationGreaterThan,
				FieldValue:          0,
			},
		},
	}
	invalidIntFilter := base.FilterInput{
		FilterBy: []*base.FilterParam{
			{
				FieldName:           "count",
				FieldType:           base.FieldTypeInteger,
				ComparisonOperation: base.OperationGreaterThan,
				FieldValue:          "not a valid int",
			},
		},
	}
	timestampFilter := base.FilterInput{
		FilterBy: []*base.FilterParam{
			{
				FieldName:           "updated",
				FieldType:           base.FieldTypeTimestamp,
				ComparisonOperation: base.OperationGreaterThan,
				FieldValue:          time.Now(),
			},
		},
	}
	numberFilter := base.FilterInput{
		FilterBy: []*base.FilterParam{
			{
				FieldName:           "numfield",
				FieldType:           base.FieldTypeNumber,
				ComparisonOperation: base.OperationLessThan,
				FieldValue:          1.0,
			},
		},
	}
	stringFilter := base.FilterInput{
		FilterBy: []*base.FilterParam{
			{
				FieldName:           "name",
				FieldType:           base.FieldTypeString,
				ComparisonOperation: base.OperationEqual,
				FieldValue:          "a string",
			},
		},
	}

	unknownFieldType := base.FilterInput{
		FilterBy: []*base.FilterParam{
			{
				FieldName:           "name",
				FieldType:           base.FieldType("this is a strange field type"),
				ComparisonOperation: base.OperationEqual,
				FieldValue:          "a string",
			},
		},
	}

	type args struct {
		ctx    context.Context
		filter *base.FilterInput
		sort   *base.SortInput
		node   base.Node
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "nil sort and filter",
			args: args{
				ctx:    ctx,
				filter: nil,
				sort:   nil,
				node:   node,
			},
			wantErr: false,
		},
		{
			name: "ascending sort",
			args: args{
				ctx:    ctx,
				filter: nil,
				sort:   &sortAsc,
				node:   node,
			},
			wantErr: false,
		},
		{
			name: "descending sort",
			args: args{
				ctx:    ctx,
				filter: nil,
				sort:   &sortDesc,
				node:   node,
			},
			wantErr: false,
		},
		{
			name: "invalid filter",
			args: args{
				ctx:    ctx,
				filter: &invalidFilter,
				sort:   &sortDesc,
				node:   node,
			},
			wantErr: true,
		},
		{
			name: "valid boolean filter",
			args: args{
				ctx:    ctx,
				filter: &booleanFilter,
				sort:   &sortAsc,
				node:   node,
			},
			wantErr: false,
		},
		{
			name: "invalid boolean filter - wrong type",
			args: args{
				ctx:    ctx,
				filter: &invalidBoolFilterWrongType,
				sort:   &sortAsc,
				node:   node,
			},
			wantErr: true,
		},
		{
			name: "invalid boolean filter - unparseable string",
			args: args{
				ctx:    ctx,
				filter: &invalidBoolFilterUnparseableString,
				sort:   &sortAsc,
				node:   node,
			},
			wantErr: true,
		},
		{
			name: "valid integer filter",
			args: args{
				ctx:    ctx,
				filter: &intFilter,
				sort:   &sortAsc,
				node:   node,
			},
			wantErr: false,
		},
		{
			name: "invalid integer filter",
			args: args{
				ctx:    ctx,
				filter: &invalidIntFilter,
				sort:   &sortAsc,
				node:   node,
			},
			wantErr: true,
		},
		{
			name: "valid timestamp filter",
			args: args{
				ctx:    ctx,
				filter: &timestampFilter,
				sort:   &sortAsc,
				node:   node,
			},
			wantErr: false,
		},
		{
			name: "valid number filter",
			args: args{
				ctx:    ctx,
				filter: &numberFilter,
				sort:   &sortAsc,
				node:   node,
			},
			wantErr: false,
		},
		{
			name: "valid string filter",
			args: args{
				ctx:    ctx,
				filter: &stringFilter,
				sort:   &sortAsc,
				node:   node,
			},
			wantErr: false,
		},
		{
			name: "unknown field type",
			args: args{
				ctx:    ctx,
				filter: &unknownFieldType,
				sort:   &sortAsc,
				node:   node,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := base.ComposeUnpaginatedQuery(tt.args.ctx, tt.args.filter, tt.args.sort, tt.args.node)
			if (err != nil) != tt.wantErr {
				t.Errorf("ComposeUnpaginatedQuery() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				assert.NotNil(t, got)
			}
		})
	}
}

func TestCreateNode(t *testing.T) {
	type args struct {
		ctx  context.Context
		node base.Node
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "good case",
			args: args{
				ctx:  context.Background(),
				node: &base.Model{},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, timestamp, err := base.CreateNode(tt.args.ctx, tt.args.node)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateNode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.NotZero(t, id)
			assert.NotZero(t, timestamp)
		})
	}
}

func TestUpdateNode(t *testing.T) {
	ctx := context.Background()
	node := &base.Model{}
	id, _, err := base.CreateNode(ctx, node)
	assert.Nil(t, err)
	assert.NotZero(t, id)

	node.Name = "updated" // the update that we are testing

	type args struct {
		ctx  context.Context
		id   string
		node base.Node
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "good case",
			args: args{
				ctx:  ctx,
				id:   id,
				node: node,
			},
			wantErr: false,
		},
		{
			name: "node that does not exist",
			args: args{
				ctx:  ctx,
				id:   "this is a bogus ID that should not exist",
				node: node,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := base.UpdateNode(tt.args.ctx, tt.args.id, tt.args.node)
			if (err != nil) != tt.wantErr {
				t.Errorf("UpdateNode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.NotZero(t, got)
		})
	}
}

func TestRetrieveNode(t *testing.T) {
	ctx := context.Background()
	node := &base.Model{}
	id, _, err := base.CreateNode(ctx, node)
	assert.Nil(t, err)
	assert.NotZero(t, id)

	type args struct {
		ctx  context.Context
		id   string
		node base.Node
	}
	tests := []struct {
		name    string
		args    args
		want    base.Node
		wantErr bool
	}{
		{
			name: "good case",
			args: args{
				ctx:  ctx,
				id:   id,
				node: node,
			},
			want:    node,
			wantErr: false,
		},
		{
			name: "non existent node",
			args: args{
				ctx:  ctx,
				id:   "fake ID - should not exist",
				node: node,
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := base.RetrieveNode(tt.args.ctx, tt.args.id, tt.args.node)
			if (err != nil) != tt.wantErr {
				t.Errorf("RetrieveNode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("RetrieveNode() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestQueryNodes(t *testing.T) {
	ctx := context.Background()
	node := &base.Model{
		Name:        "test model instance",
		Description: "this is a test description",
		Deleted:     false,
	}
	id, _, err := base.CreateNode(ctx, node)
	assert.Nil(t, err)
	assert.NotZero(t, id)

	sortAsc := base.SortInput{
		SortBy: []*base.SortParam{
			{
				FieldName: "name",
				SortOrder: base.SortOrderAsc,
			},
		},
	}

	type args struct {
		ctx        context.Context
		pagination *base.PaginationInput
		filter     *base.FilterInput
		sort       *base.SortInput
		node       base.Node
	}
	tests := []struct {
		name    string
		args    args
		want    []*firestore.DocumentSnapshot
		wantErr bool
	}{
		{
			name: "no pagination, filter or sort",
			args: args{
				ctx:        ctx,
				pagination: nil,
				filter:     nil,
				sort:       nil,
				node:       &base.Model{},
			},
			wantErr: false,
		},
		{
			name: "with pagination, first",
			args: args{
				ctx: ctx,
				pagination: &base.PaginationInput{
					First: 10,
					After: id,
				},
				filter: nil,
				sort:   &sortAsc,
				node:   &base.Model{},
			},
			wantErr: false,
		},
		{
			name: "with pagination, last",
			args: args{
				ctx: ctx,
				pagination: &base.PaginationInput{
					Last:   1,
					Before: id,
				},
				filter: nil,
				sort:   &sortAsc,
				node:   &base.Model{},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			snapshots, pageInfo, err := base.QueryNodes(tt.args.ctx, tt.args.pagination, tt.args.filter, tt.args.sort, tt.args.node)
			if (err != nil) != tt.wantErr {
				t.Errorf("QueryNodes() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				assert.NotNil(t, pageInfo)
				if pageInfo.StartCursor != nil {
					assert.NotNil(t, snapshots)
				}
			}
		})
	}
}

func TestDeleteNode(t *testing.T) {
	ctx := context.Background()
	node := &base.Model{
		Name:        "test model instance",
		Description: "this is a test description",
		Deleted:     false,
	}
	id, _, err := base.CreateNode(ctx, node)
	assert.Nil(t, err)
	assert.NotZero(t, id)

	type args struct {
		ctx  context.Context
		id   string
		node base.Node
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{
			name: "existing node",
			args: args{
				ctx:  ctx,
				id:   node.GetID().String(),
				node: &base.Model{},
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "non existent node",
			args: args{
				ctx:  ctx,
				id:   "this should not exist",
				node: &base.Model{},
			},
			want:    true,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := base.DeleteNode(tt.args.ctx, tt.args.id, tt.args.node)
			if (err != nil) != tt.wantErr {
				t.Errorf("DeleteNode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("DeleteNode() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDeleteCollection(t *testing.T) {
	ctx := base.GetAuthenticatedContext(t)
	firestoreClient := fb.GetFirestoreClient(t)
	collection := "test_collection_deletion"
	data := map[string]string{
		"a_key_for_testing": "random-test-key-value",
	}
	id, err := base.SaveDataToFirestore(firestoreClient, collection, data)
	if err != nil {
		t.Errorf("unable to save data to firestore: %v", err)
		return
	}
	if id == "" {
		t.Errorf("id got is empty")
		return
	}

	ref := firestoreClient.Collection(collection)

	type args struct {
		ctx       context.Context
		client    *firestore.Client
		ref       *firestore.CollectionRef
		batchSize int
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "good case - successfully deleted collection",
			args: args{
				ctx:       ctx,
				client:    firestoreClient,
				ref:       ref,
				batchSize: 10,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := base.DeleteCollection(tt.args.ctx, tt.args.client, tt.args.ref, tt.args.batchSize); (err != nil) != tt.wantErr {
				t.Errorf("DeleteCollection() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
