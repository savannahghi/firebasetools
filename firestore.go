package firebasetools

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/google/uuid"
	"github.com/savannahghi/server_utils"
	"google.golang.org/api/iterator"
)

// UnixEpoch is used as our version of "time zero".
// We don't (shouldn't) change it so it's safe to make it a global.
var UnixEpoch = time.Unix(0, 0)

// GetCollectionName calculates the name to give to a node's collection on Firestore
func GetCollectionName(n Node) string {
	fullName := fmt.Sprintf("%T", n) // e.g "*authorization.Store"
	split := strings.Split(fullName, ".")
	lastPart := split[len(split)-1]
	return SuffixCollection(strings.ToLower(lastPart))
}

// GetFirestoreEnvironmentSuffix get the env suffix where the app is running
func GetFirestoreEnvironmentSuffix() string {
	return server_utils.MustGetEnvVar("ROOT_COLLECTION_SUFFIX")
}

// SuffixCollection adds a suffix to the collection name. This will aid in separating
// collections for different environments
func SuffixCollection(c string) string {
	return fmt.Sprintf("%v_bewell_%v", c, GetFirestoreEnvironmentSuffix())
}

// ValidatePaginationParameters ensures that the supplied pagination parameters make sense
func ValidatePaginationParameters(pagination *PaginationInput) error {
	if pagination == nil {
		return nil // not having pagination is not a fatal error
	}

	// if `first` is specified, `last` cannot be specified
	if pagination.First > 0 && pagination.Last > 0 {
		return fmt.Errorf("if `first` is specified for pagination, `last` cannot be specified")
	}

	return nil
}

// OpString translates between an Operation enum value and the appropriate firestore
// query operator
func OpString(op Operation) (string, error) {
	switch op {
	case OperationLessThan:
		return "<", nil
	case OperationLessThanOrEqualTo:
		return "<=", nil
	case OperationEqual:
		return "==", nil
	case OperationGreaterThan:
		return ">", nil
	case OperationGreaterThanOrEqualTo:
		return ">=", nil
	case OperationIn:
		return "in", nil
	case OperationContains:
		return "array-contains", nil
	default:
		return "", fmt.Errorf("unknown operation; did you forget to update this function after adding new operations in the schema?")
	}
}

// GetFirestoreClient initializes a Firestore client
func GetFirestoreClient(ctx context.Context) (*firestore.Client, error) {
	fc := &FirebaseClient{}
	fa, err := fc.InitFirebase()
	if err != nil {
		return nil, fmt.Errorf("unable to initialize Firebase client: %w", err)
	}
	firestore, err := fa.Firestore(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to initialize Firestore client: %w", err)
	}
	return firestore, nil
}

// ComposeUnpaginatedQuery creates a Cloud Firestore query
func ComposeUnpaginatedQuery(
	ctx context.Context,
	filter *FilterInput,
	sort *SortInput,
	node Node,
) (*firestore.Query, error) {
	collectionName := GetCollectionName(node)
	firestoreClient, err := GetFirestoreClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("can't compose unpaginated query: %w", err)
	}

	// apply filters
	query := firestoreClient.Collection(collectionName).Query
	if filter != nil {
		for _, filterParam := range filter.FilterBy {
			op, err := OpString(filterParam.ComparisonOperation)
			if err != nil {
				return nil, err
			}

			switch filterParam.FieldType {
			case FieldTypeBoolean:
				boolFilterVal, ok := filterParam.FieldValue.(string)
				if !ok {
					return nil, fmt.Errorf("a boolean filter value should be the string 'true' or the string 'false'")
				}
				parsed, err := strconv.ParseBool(boolFilterVal)
				if err != nil {
					return nil, err
				}
				query = query.Where(filterParam.FieldName, op, parsed)
			case FieldTypeInteger:
				intFilterValue, ok := filterParam.FieldValue.(int)
				if !ok {
					return nil, fmt.Errorf("expected the filter value to be an int")
				}
				query = query.Where(filterParam.FieldName, op, intFilterValue)
			case FieldTypeTimestamp:
				// a future decision on timestamp formats would affect this
				query = query.Where(filterParam.FieldName, op, filterParam.FieldValue)
			case FieldTypeNumber:
				query = query.Where(filterParam.FieldName, op, filterParam.FieldValue)
			case FieldTypeString:
				query = query.Where(filterParam.FieldName, op, filterParam.FieldValue)
			default:
				return nil, fmt.Errorf("unexpected field type '%s'", filterParam.FieldType.String())
			}
		}
	}

	if sort != nil {
		for _, sortParam := range sort.SortBy {
			switch sortParam.SortOrder {
			case SortOrderAsc:
				query = query.OrderBy(sortParam.FieldName, firestore.Asc)
			case SortOrderDesc:
				query = query.OrderBy(sortParam.FieldName, firestore.Desc)
			}
		}
	}
	return &query, nil
}

// QueryNodes prepares and executes queries against Firebase collections
func QueryNodes(
	ctx context.Context, pagination *PaginationInput,
	filter *FilterInput, sort *SortInput, node Node) ([]*firestore.DocumentSnapshot, *PageInfo, error) {
	queryPtr, err := ComposeUnpaginatedQuery(ctx, filter, sort, node)
	if err != nil {
		return nil, nil, err
	}
	query := *queryPtr

	// pagination
	pageSize := DefaultPageSize
	if pagination != nil {
		if pagination.First > 0 {
			if pagination.After != "" {
				query = query.StartAfter(pagination.After)
			}
			pageSize = pagination.First
		}
		if pagination.Last > 0 {
			if pagination.Before != "" {
				query = query.EndBefore(pagination.Before)
			}
			pageSize = pagination.Last
		}
	}
	query = query.Limit(pageSize)

	// start with a default PageInfo
	docs, err := query.Documents(ctx).GetAll()
	if err != nil {
		return nil, nil, err
	}

	cursors := []string{}
	for _, doc := range docs {
		cursors = append(cursors, doc.Ref.ID)
	}

	// check if there is a next page
	pageInfo := &PageInfo{
		HasPreviousPage: pagination != nil && pagination.After != "",
	}
	if len(docs) > 0 {
		secondQueryPtr, err := ComposeUnpaginatedQuery(ctx, filter, sort, node)
		if err != nil {
			return nil, nil, err
		}
		secondQuery := *secondQueryPtr
		lastSnapshot := docs[len(docs)-1]
		nextDoc, err := secondQuery.StartAfter(lastSnapshot).Limit(1).Documents(ctx).GetAll()
		if err != nil {
			return nil, nil, err
		}
		pageInfo.HasNextPage = len(nextDoc) > 0
	}
	if len(cursors) > 0 {
		startCursor := cursors[0]
		endCursor := cursors[len(cursors)-1]

		pageInfo.StartCursor = &startCursor
		pageInfo.EndCursor = &endCursor
	}
	return docs, pageInfo, nil
}

// RetrieveNode retrieves a node from Firestore
func RetrieveNode(ctx context.Context, id string, node Node) (Node, error) {
	collName := GetCollectionName(node)
	firestoreClient, err := GetFirestoreClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to initialize Firestore: %w", err)
	}
	dsnap, err := firestoreClient.Collection(collName).Doc(id).Get(ctx)
	if err != nil {
		return nil, err
	}
	err = dsnap.DataTo(node)
	if err != nil {
		return nil, err
	}
	return node, nil
}

// DeleteNode retrieves a node from Firestore
func DeleteNode(ctx context.Context, id string, node Node) (bool, error) {
	collName := GetCollectionName(node)
	firestoreClient, err := GetFirestoreClient(ctx)
	if err != nil {
		return false, fmt.Errorf("unable to initialize Firestore: %w", err)
	}
	_, err = firestoreClient.Collection(collName).Doc(id).Delete(ctx)
	if err != nil {
		return false, fmt.Errorf("unable to delete %T with ID %s: %w", node, id, err)
	}
	return true, nil
}

// CreateNode creates a Node on Firebase
func CreateNode(ctx context.Context, node Node) (string, time.Time, error) {
	collectionName := GetCollectionName(node)
	firestoreClient, err := GetFirestoreClient(ctx)
	if err != nil {
		return "", UnixEpoch, fmt.Errorf("unable to update node: %w", err)
	}

	// assign a random ID if one does not already exist
	// but respect the ones that exist i.e don't overwrite
	id := node.GetID().String()
	if id == "" {
		node.SetID(uuid.New().String())
		id = node.GetID().String()
	}

	result, err := firestoreClient.Collection(collectionName).Doc(id).Create(ctx, node)
	if err != nil {
		return "", UnixEpoch, err
	}
	return id, result.UpdateTime, nil
}

// UpdateNode updates an existing node's document on Firestore
func UpdateNode(ctx context.Context, id string, node Node) (time.Time, error) {
	collName := GetCollectionName(node)
	firestoreClient, err := GetFirestoreClient(ctx)
	if err != nil {
		return UnixEpoch, fmt.Errorf("unable to update node: %w", err)
	}
	result, err := firestoreClient.Collection(collName).Doc(id).Set(ctx, node)
	if err != nil {
		return UnixEpoch, err
	}
	return result.UpdateTime, nil
}

// DeleteCollection deletes a firestore collection
func DeleteCollection(
	ctx context.Context,
	client *firestore.Client,
	ref *firestore.CollectionRef,
	batchSize int) error {
	for {
		iter := ref.Limit(batchSize).Documents(ctx)
		numDeleted := 0
		batch := client.Batch()
		for {
			doc, err := iter.Next()
			if err == iterator.Done {
				break
			}
			if err != nil {
				return err
			}

			batch.Delete(doc.Ref)
			numDeleted++
		}

		if numDeleted == 0 {
			return nil
		}

		_, err := batch.Commit(ctx)
		if err != nil {
			return err
		}
	}
}
