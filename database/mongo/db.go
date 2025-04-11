package mongo

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/rbrabson/goblin/internal/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	DB_TIMEOUT = 10 * time.Second
)

var (
	sslog = logger.GetLogger()
)

// MongoDB represents a connection to a mongo database
type MongoDB struct {
	Client     *mongo.Client
	clientOpts *options.ClientOptions
	dbname     string
	uri        string
}

// NewDatabase creates a database to load and save documents in a MongoDB database.
func NewDatabase() *MongoDB {
	uri := os.Getenv("MONGODB_URI")
	dbname := os.Getenv("MONGODB_DATABASE")

	m := &MongoDB{
		uri:    uri,
		dbname: dbname,
	}

	// Wait for MongoDB to become active before proceeding
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var err error
	m.clientOpts = options.Client().ApplyURI(m.uri)
	m.Client, err = mongo.Connect(ctx, m.clientOpts)
	if err != nil {
		sslog.Error("unable to connect to the MongoDB database",
			slog.Any("error", err),
		)
		return nil
	}

	// Check the connection
	err = m.Client.Ping(ctx, nil)
	if err != nil {
		sslog.Error("unable to ping the MongoDB database",
			slog.Any("error", err),
		)
		err = nil
	}

	return m
}

// FindAllIDs returns the ID of each document in a collection in the database.
func (m *MongoDB) FindAllIDs(collectionName string, filter interface{}) ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), DB_TIMEOUT)
	defer cancel()

	collection, err := m.getCollection(ctx, collectionName)
	if err != nil {
		return nil, err
	}

	opts := options.Find().SetProjection(bson.M{"_id": 1})

	cur, err := collection.Find(ctx, filter, opts)
	if err != nil {
		sslog.Error("Failed to read the collection",
			slog.String("collection", collectionName),
			slog.Any("error", err),
		)
		return nil, ErrCollectionNotAccessable
	}
	defer func() {
		cur.Close(ctx)
	}()

	type result struct {
		ID string `bson:"_id"`
	}
	var results []result
	err = cur.All(ctx, &results)
	if err != nil {
		sslog.Error("error getting IDs for the collection",
			slog.String("collection", collectionName),
			slog.Any("error", err),
		)
		return nil, ErrCollectionNotAccessable
	}
	defer func() {
		cur.Close(ctx)
	}()

	idList := make([]string, 0, len(results))
	for _, r := range results {
		idList = append(idList, r.ID)
	}

	return idList, nil
}

// FindMany reads all documents from the database that match the filter
func (m *MongoDB) FindMany(collectionName string, filter interface{}, data interface{}, sortBy interface{}, limit int64) error {
	ctx, cancel := context.WithTimeout(context.Background(), DB_TIMEOUT)
	defer cancel()

	collection, err := m.getCollection(ctx, collectionName)
	if err != nil {
		return err
	}

	// Limit the number of documents to return
	findOptions := options.Find()
	findOptions.Sort = sortBy
	findOptions.SetLimit(limit)

	cur, err := collection.Find(ctx, filter, findOptions)
	if err != nil {
		sslog.Debug("unable to find the document",
			slog.String("database", m.dbname),
			slog.String("collection", collectionName),
			slog.Any("error", err),
		)
		return err
	}
	defer func() {
		cur.Close(ctx)
	}()
	err = cur.All(ctx, data)
	if err != nil {
		sslog.Error("unable to decode the documents",
			slog.String("database", m.dbname),
			slog.String("collection", collectionName),
			slog.Any("error", err),
			"data", data,
		)
		return ErrInvalidDocument
	}

	return nil
}

// FindOne loads a document identified by documentID from the collection into data.
func (m *MongoDB) FindOne(collectionName string, filter interface{}, data interface{}) error {
	ctx, cancel := context.WithTimeout(context.Background(), DB_TIMEOUT)
	defer cancel()

	collection, err := m.getCollection(ctx, collectionName)
	if err != nil {
		return err
	}

	res := collection.FindOne(ctx, filter)
	if res.Err() != nil {
		sslog.Debug("unable to find the document",
			slog.String("database", m.dbname),
			slog.String("collection", collectionName),
			slog.String("error", res.Err().Error()),
			"filter", filter,
		)
		return res.Err()
	}
	if res == nil {
		sslog.Debug("unable to find the document",
			slog.String("database", m.dbname),
			slog.String("collection", collectionName),
		)
		return ErrDocumentNotFound
	}
	err = res.Decode(data)
	if err != nil {
		sslog.Error("unable to decode the document",
			slog.String("database", m.dbname),
			slog.String("collection", collectionName),
			slog.Any("error", err),
			"data", data,
		)
		return ErrInvalidDocument
	}
	return nil
}

// UpdateOrInsert stores data into a documeent within the specified collection.
func (m *MongoDB) UpdateOrInsert(collectionName string, filter interface{}, data interface{}) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	collection, err := m.getCollection(ctx, collectionName)
	if err != nil {
		return err
	}

	update := bson.M{"$set": data}
	_, err = collection.UpdateOne(ctx, filter, update, options.Update().SetUpsert(true))
	if err != nil {
		sslog.Error("unable to insert or update the document the collection",
			slog.String("database", m.dbname),
			slog.String("collection", collectionName),
			"filter", filter,
			"data", data,
		)
		return err
	}

	return nil
}

// Write stores data into multiple documeents within the specified collection.
func (m *MongoDB) UpdateMany(collectionName string, filter interface{}, data interface{}) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	collection, err := m.getCollection(ctx, collectionName)
	if err != nil {
		return err
	}

	update := bson.M{"$set": data}
	_, err = collection.UpdateMany(ctx, filter, update, options.Update().SetUpsert(true))
	if err != nil {
		sslog.Error("unable to insert or update the document the collection",
			slog.String("collection", collectionName),
			slog.Any("error", err),
			"filter", filter,
			"data", data,
		)
		return err
	}
	sslog.Debug("updated document in the collection",
		slog.String("collection", collectionName),
		"filter", filter,
		"data", data,
	)

	return nil
}

// Count returns the count of documents that match the filter.
func (m *MongoDB) Count(collectionName string, filter interface{}) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), DB_TIMEOUT)
	defer cancel()

	collection, err := m.getCollection(ctx, collectionName)
	if err != nil {
		return 0, err
	}

	opts := options.Count()
	count, err := collection.CountDocuments(ctx, filter, opts)
	if err != nil {
		sslog.Error("Failed to read the collection",
			slog.String("collection", collectionName),
			slog.Any("error", err),
			"filter", filter,
		)
		return 0, ErrCollectionNotAccessable
	}
	sslog.Debug("count",
		slog.String("collection", collectionName),
		slog.Int64("count", count),
		"filter", filter,
	)

	return int(count), nil
}

// Delete removes a document from the collection that matches the filter.
func (m *MongoDB) Delete(collectionName string, filter interface{}) error {
	ctx, cancel := context.WithTimeout(context.Background(), DB_TIMEOUT)
	defer cancel()

	collection, err := m.getCollection(ctx, collectionName)
	if err != nil {
		return err
	}

	res, err := collection.DeleteOne(ctx, filter)
	if err != nil {
		sslog.Error("unable to delete the document",
			slog.String("collection", collectionName),
			slog.Any("error", err),
			"filter", filter,
		)
		return err
	}
	if res.DeletedCount == 0 {
		sslog.Warn("document not found",
			slog.String("collection", collectionName),
			"filter", filter,
		)
	}
	sslog.Debug("deleted document",
		slog.String("collection", collectionName),
		slog.Int64("count", res.DeletedCount),
		"filter", filter,
	)

	return nil
}

// Delete removes all documents from the collection that matche the filter.
func (m *MongoDB) DeleteMany(collectionName string, filter interface{}) error {
	ctx, cancel := context.WithTimeout(context.Background(), DB_TIMEOUT)
	defer cancel()

	collection, err := m.getCollection(ctx, collectionName)
	if err != nil {
		return err
	}

	res, err := collection.DeleteMany(ctx, filter)
	if err != nil {
		sslog.Error("unable to delete the document",
			slog.String("collection", collectionName),
			slog.Any("error", err),
			"filter", filter,
		)
		return err
	}
	if res.DeletedCount == 0 {
		sslog.Warn("document not found",
			slog.String("collection", collectionName),
			slog.Int64("count", res.DeletedCount),
			"filter", filter,
		)
	}
	sslog.Debug("deleted document",
		slog.String("collection", collectionName),
		slog.Int64("count", res.DeletedCount),
		"filter", filter,
	)

	return nil
}

// Close closes the mongo database client connection
func (m *MongoDB) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), DB_TIMEOUT)
	defer cancel()
	if err := m.Client.Disconnect(ctx); err != nil {
		sslog.Error("unable to close the mongo database client",
			slog.Any("error", err),
		)
		return err
	}
	return nil
}

// getCollection returns a collection from the database that may be used for database operations.
func (m *MongoDB) getCollection(ctx context.Context, collectionName string) (*mongo.Collection, error) {
	if m.clientOpts == nil {
		var err error
		m.clientOpts = options.Client().ApplyURI(m.uri)
		m.Client, err = mongo.Connect(ctx, m.clientOpts)
		if err != nil {
			sslog.Error("unable to connect to the MongoDB database",
				slog.Any("error", err),
			)
			return nil, err
		}
	}

	db := m.Client.Database(m.dbname)
	collection := db.Collection(collectionName)
	if collection == nil {
		sslog.Error("uanble to access the collection",
			slog.String("collection", collectionName),
		)
		return nil, ErrCollectionNotAccessable
	}

	return collection, nil
}

// String returns the name of the database
func (db *MongoDB) String() string {
	return "mongo"
}
