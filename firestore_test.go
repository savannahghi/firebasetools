package firebasetools_test

import (
	"context"
	"fmt"
	"reflect"
	"testing"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/savannahghi/enumutils"
	fb "github.com/savannahghi/firebasetools"
	"github.com/stretchr/testify/assert"
)

// Dummy is a test node
type Dummy struct{}

func (d Dummy) IsNode() {}

func (d Dummy) GetID() fb.ID {
	return fb.IDValue("dummy id")
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
	assert.Equal(t, "dummy_bewell_staging", fb.GetCollectionName(n1))
}

func Test_validatePaginationParameters(t *testing.T) {
	first := 10
	after := "30"
	last := 10
	before := "20"

	tests := map[string]struct {
		pagination           *fb.PaginationInput
		expectError          bool
		expectedErrorMessage string
	}{
		"first_last_specified": {
			pagination: &fb.PaginationInput{
				First: first,
				Last:  last,
			},
			expectError:          true,
			expectedErrorMessage: "if `first` is specified for pagination, `last` cannot be specified",
		},
		"first_only": {
			pagination: &fb.PaginationInput{
				First: first,
			},
			expectError: false,
		},
		"last_only": {
			pagination: &fb.PaginationInput{
				Last: last,
			},
			expectError: false,
		},
		"first_and_after": {
			pagination: &fb.PaginationInput{
				First: first,
				After: after,
			},
			expectError: false,
		},
		"last_and_before": {
			pagination: &fb.PaginationInput{
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
			err := fb.ValidatePaginationParameters(tc.pagination)
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
		op                   enumutils.Operation
		expectedOutput       string
		expectError          bool
		expectedErrorMessage string
	}{
		"invalid_operation": {
			op:                   enumutils.Operation("invalid unknown operation"),
			expectedOutput:       "",
			expectError:          true,
			expectedErrorMessage: "unknown operation; did you forget to update this function after adding new operations in the schema?",
		},
		"less than": {
			op:             enumutils.OperationLessThan,
			expectedOutput: "<",
			expectError:    false,
		},
		"less than_or_equal_to": {
			op:             enumutils.OperationLessThanOrEqualTo,
			expectedOutput: "<=",
			expectError:    false,
		},
		"equal_to": {
			op:             enumutils.OperationEqual,
			expectedOutput: "==",
			expectError:    false,
		},
		"greater_than": {
			op:             enumutils.OperationGreaterThan,
			expectedOutput: ">",
			expectError:    false,
		},
		"greater_than_or_equal_to": {
			op:             enumutils.OperationGreaterThanOrEqualTo,
			expectedOutput: ">=",
			expectError:    false,
		},
		"in": {
			op:             enumutils.OperationIn,
			expectedOutput: "in",
			expectError:    false,
		},
		"contains": {
			op:             enumutils.OperationContains,
			expectedOutput: "array-contains",
			expectError:    false,
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			opString, err := fb.OpString(tc.op)
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
			got, err := fb.GetFirestoreClient(tt.args.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetFirestoreClientTestUtil() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.NotNil(t, got)
		})
	}
}

func TestComposeUnpaginatedQuery(t *testing.T) {
	ctx := context.Background()
	node := &fb.Model{}
	sortAsc := fb.SortInput{
		SortBy: []*fb.SortParam{
			{
				FieldName: "name",
				SortOrder: enumutils.SortOrderAsc,
			},
		},
	}
	sortDesc := fb.SortInput{
		SortBy: []*fb.SortParam{
			{
				FieldName: "name",
				SortOrder: enumutils.SortOrderDesc,
			},
		},
	}
	invalidFilter := fb.FilterInput{
		FilterBy: []*fb.FilterParam{
			{
				FieldName:           "name",
				FieldType:           enumutils.FieldTypeString,
				ComparisonOperation: enumutils.Operation("not a valid operation"),
				FieldValue:          "val",
			},
		},
	}
	booleanFilter := fb.FilterInput{
		FilterBy: []*fb.FilterParam{
			{
				FieldName:           "deleted",
				FieldType:           enumutils.FieldTypeBoolean,
				ComparisonOperation: enumutils.OperationEqual,
				FieldValue:          "false",
			},
		},
	}
	invalidBoolFilterWrongType := fb.FilterInput{
		FilterBy: []*fb.FilterParam{
			{
				FieldName:           "deleted",
				FieldType:           enumutils.FieldTypeBoolean,
				ComparisonOperation: enumutils.OperationEqual,
				FieldValue:          false,
			},
		},
	}
	invalidBoolFilterUnparseableString := fb.FilterInput{
		FilterBy: []*fb.FilterParam{
			{
				FieldName:           "deleted",
				FieldType:           enumutils.FieldTypeBoolean,
				ComparisonOperation: enumutils.OperationEqual,
				FieldValue:          "bad format",
			},
		},
	}
	intFilter := fb.FilterInput{
		FilterBy: []*fb.FilterParam{
			{
				FieldName:           "count",
				FieldType:           enumutils.FieldTypeInteger,
				ComparisonOperation: enumutils.OperationGreaterThan,
				FieldValue:          0,
			},
		},
	}
	invalidIntFilter := fb.FilterInput{
		FilterBy: []*fb.FilterParam{
			{
				FieldName:           "count",
				FieldType:           enumutils.FieldTypeInteger,
				ComparisonOperation: enumutils.OperationGreaterThan,
				FieldValue:          "not a valid int",
			},
		},
	}
	timestampFilter := fb.FilterInput{
		FilterBy: []*fb.FilterParam{
			{
				FieldName:           "updated",
				FieldType:           enumutils.FieldTypeTimestamp,
				ComparisonOperation: enumutils.OperationGreaterThan,
				FieldValue:          time.Now(),
			},
		},
	}
	numberFilter := fb.FilterInput{
		FilterBy: []*fb.FilterParam{
			{
				FieldName:           "numfield",
				FieldType:           enumutils.FieldTypeNumber,
				ComparisonOperation: enumutils.OperationLessThan,
				FieldValue:          1.0,
			},
		},
	}
	stringFilter := fb.FilterInput{
		FilterBy: []*fb.FilterParam{
			{
				FieldName:           "name",
				FieldType:           enumutils.FieldTypeString,
				ComparisonOperation: enumutils.OperationEqual,
				FieldValue:          "a string",
			},
		},
	}

	unknownFieldType := fb.FilterInput{
		FilterBy: []*fb.FilterParam{
			{
				FieldName:           "name",
				FieldType:           enumutils.FieldType("this is a strange field type"),
				ComparisonOperation: enumutils.OperationEqual,
				FieldValue:          "a string",
			},
		},
	}

	type args struct {
		ctx    context.Context
		filter *fb.FilterInput
		sort   *fb.SortInput
		node   fb.Node
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
			got, err := fb.ComposeUnpaginatedQuery(tt.args.ctx, tt.args.filter, tt.args.sort, tt.args.node)
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
		node fb.Node
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
				node: &fb.Model{},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, timestamp, err := fb.CreateNode(tt.args.ctx, tt.args.node)
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
	node := &fb.Model{}
	id, _, err := fb.CreateNode(ctx, node)
	assert.Nil(t, err)
	assert.NotZero(t, id)

	node.Name = "updated" // the update that we are testing

	type args struct {
		ctx  context.Context
		id   string
		node fb.Node
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
			got, err := fb.UpdateNode(tt.args.ctx, tt.args.id, tt.args.node)
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
	node := &fb.Model{}
	id, _, err := fb.CreateNode(ctx, node)
	assert.Nil(t, err)
	assert.NotZero(t, id)

	type args struct {
		ctx  context.Context
		id   string
		node fb.Node
	}
	tests := []struct {
		name    string
		args    args
		want    fb.Node
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
			got, err := fb.RetrieveNode(tt.args.ctx, tt.args.id, tt.args.node)
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
	node := &fb.Model{
		Name:        "test model instance",
		Description: "this is a test description",
		Deleted:     false,
	}
	id, _, err := fb.CreateNode(ctx, node)
	assert.Nil(t, err)
	assert.NotZero(t, id)

	sortAsc := fb.SortInput{
		SortBy: []*fb.SortParam{
			{
				FieldName: "name",
				SortOrder: enumutils.SortOrderAsc,
			},
		},
	}

	type args struct {
		ctx        context.Context
		pagination *fb.PaginationInput
		filter     *fb.FilterInput
		sort       *fb.SortInput
		node       fb.Node
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
				node:       &fb.Model{},
			},
			wantErr: false,
		},
		{
			name: "with pagination, first",
			args: args{
				ctx: ctx,
				pagination: &fb.PaginationInput{
					First: 10,
					After: id,
				},
				filter: nil,
				sort:   &sortAsc,
				node:   &fb.Model{},
			},
			wantErr: false,
		},
		{
			name: "with pagination, last",
			args: args{
				ctx: ctx,
				pagination: &fb.PaginationInput{
					Last:   1,
					Before: id,
				},
				filter: nil,
				sort:   &sortAsc,
				node:   &fb.Model{},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			snapshots, pageInfo, err := fb.QueryNodes(tt.args.ctx, tt.args.pagination, tt.args.filter, tt.args.sort, tt.args.node)
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
	node := &fb.Model{
		Name:        "test model instance",
		Description: "this is a test description",
		Deleted:     false,
	}
	id, _, err := fb.CreateNode(ctx, node)
	assert.Nil(t, err)
	assert.NotZero(t, id)

	type args struct {
		ctx  context.Context
		id   string
		node fb.Node
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
				node: &fb.Model{},
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "non existent node",
			args: args{
				ctx:  ctx,
				id:   "this should not exist",
				node: &fb.Model{},
			},
			want:    true,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := fb.DeleteNode(tt.args.ctx, tt.args.id, tt.args.node)
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
	ctx := fb.GetAuthenticatedContext(t)
	firestoreClient := fb.GetFirestoreClientTestUtil(t)
	collection := "test_collection_deletion"
	data := map[string]string{
		"a_key_for_testing": "random-test-key-value",
	}
	id, err := fb.SaveDataToFirestore(firestoreClient, collection, data)
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
			if err := fb.DeleteCollection(tt.args.ctx, tt.args.client, tt.args.ref, tt.args.batchSize); (err != nil) != tt.wantErr {
				t.Errorf("DeleteCollection() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
