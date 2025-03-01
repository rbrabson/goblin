package mongo

import (
	"context"
	"os"
	"time"

	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	DB_TIMEOUT = 10 * time.Second
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
	log.Trace("--> mongo.NewDatabase")
	defer log.Trace("<-- mongo.NewDatabase")

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
		log.WithField("error", err).Fatal("unable to connect to the MongoDB database")
		return nil
	}

	// Check the connection
	err = m.Client.Ping(ctx, nil)
	if err != nil {
		log.WithField("error", err).Fatal("unable to ping the MongoDB database")
		err = nil
	}

	return m
}

// FindAllIDs returns the ID of each document in a collection in the database.
func (m *MongoDB) FindAllIDs(collectionName string, filter interface{}) ([]string, error) {
	log.Trace("--> mongoDB.FindAllIDs")
	defer log.Trace("<-- mongoDB.FindAllIDs")

	ctx, cancel := context.WithTimeout(context.Background(), DB_TIMEOUT)
	defer cancel()

	collection, err := m.getCollection(ctx, collectionName)
	if err != nil {
		return nil, err
	}

	opts := options.Find().SetProjection(bson.M{"_id": 1})

	cur, err := collection.Find(ctx, filter, opts)
	if err != nil {
		log.WithFields(log.Fields{"collection": collectionName, "error": err}).Error("Failed to read the collection")
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
		log.WithFields(log.Fields{"collection": collectionName, "error": err}).Error("error getting IDs for the collection")
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
	log.Trace("--> mongo.FindMany")
	defer log.Trace("<-- mongoDB.FindMany")

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
		log.WithFields(log.Fields{"database": m.dbname, "collection": collectionName, "filter": filter, "error": err}).Debug("unable to find the document")
		return err
	}
	defer func() {
		cur.Close(ctx)
	}()
	err = cur.All(ctx, data)
	if err != nil {
		log.WithFields(log.Fields{"database": m.dbname, "collection": collectionName, "filter": filter, "error": err}).Error("unable to decode the documents")
		return ErrInvalidDocument
	}

	return nil
}

// FindOne loads a document identified by documentID from the collection into data.
func (m *MongoDB) FindOne(collectionName string, filter interface{}, data interface{}) error {
	log.Trace("--> mongoDB.FindOne")
	defer log.Trace("<-- mongoDB.FindOne")

	ctx, cancel := context.WithTimeout(context.Background(), DB_TIMEOUT)
	defer cancel()

	collection, err := m.getCollection(ctx, collectionName)
	if err != nil {
		return err
	}

	res := collection.FindOne(ctx, filter)
	if res.Err() != nil {
		log.WithFields(log.Fields{"database": m.dbname, "collection": collectionName, "filter": filter, "error": res.Err()}).Debug("unable to find the document")
		return res.Err()
	}
	if res == nil {
		log.WithFields(log.Fields{"database": m.dbname, "collection": collectionName, "filter": filter}).Debug("unable to find the document")
		return ErrDocumentNotFound
	}
	err = res.Decode(data)
	if err != nil {
		log.WithFields(log.Fields{"database": m.dbname, "collection": collectionName, "filter": filter, "error": err}).Error("unable to decode the document")
		return ErrInvalidDocument
	}
	return nil
}

// UpdateOrInsert stores data into a documeent within the specified collection.
func (m *MongoDB) UpdateOrInsert(collectionName string, filter interface{}, data interface{}) error {
	log.Trace("--> mongoDB.Update")
	defer log.Trace("<-- mongoDB.Update")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	collection, err := m.getCollection(ctx, collectionName)
	if err != nil {
		return err
	}

	update := bson.M{"$set": data}
	_, err = collection.UpdateOne(ctx, filter, update, options.Update().SetUpsert(true))
	if err != nil {
		log.WithFields(log.Fields{"collection": collectionName, "filter": filter, "error": err, "data": data}).Error("unable to insert or update the document the collection")
		return err
	}
	log.WithFields(log.Fields{"collection": collectionName, "filter": filter, "data": data}).Trace("inserted or updated document in the collection")

	return nil
}

// Write stores data into multiple documeents within the specified collection.
func (m *MongoDB) UpdateMany(collectionName string, filter interface{}, data interface{}) error {
	log.Trace("--> mongoDB.UpdateMany")
	defer log.Trace("<-- mongoDB.UpdateMany")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	collection, err := m.getCollection(ctx, collectionName)
	if err != nil {
		return err
	}

	update := bson.M{"$set": data}
	_, err = collection.UpdateMany(ctx, filter, update, options.Update().SetUpsert(true))
	if err != nil {
		log.WithFields(log.Fields{"collection": collectionName, "filter": filter, "error": err, "data": data}).Error("unable to insert or update the document the collection")
		return err
	}
	log.WithFields(log.Fields{"collection": collectionName, "filter": filter, "data": data}).Debug("updated document in the collection")

	return nil
}

// Count returns the count of documents that match the filter.
func (m *MongoDB) Count(collectionName string, filter interface{}) (int, error) {
	log.Trace("--> mongoDB.Count")
	defer log.Trace("<-- mongoDB.Count")

	ctx, cancel := context.WithTimeout(context.Background(), DB_TIMEOUT)
	defer cancel()

	collection, err := m.getCollection(ctx, collectionName)
	if err != nil {
		return 0, err
	}

	opts := options.Count()
	count, err := collection.CountDocuments(ctx, filter, opts)
	if err != nil {
		log.WithFields(log.Fields{"collection": collectionName, "error": err}).Error("Failed to read the collection")
		return 0, ErrCollectionNotAccessable
	}
	log.WithFields(log.Fields{"collection": collectionName, "filter": filter, "count": count}).Debug("count")

	return int(count), nil
}

// Delete removes a document from the collection that matches the filter.
func (m *MongoDB) Delete(collectionName string, filter interface{}) error {
	log.Trace("--> mongoDB.Delete")
	defer log.Trace("<-- mongoDB.Delete")

	ctx, cancel := context.WithTimeout(context.Background(), DB_TIMEOUT)
	defer cancel()

	collection, err := m.getCollection(ctx, collectionName)
	if err != nil {
		return err
	}

	res, err := collection.DeleteOne(ctx, filter)
	if err != nil {
		log.WithFields(log.Fields{"collection": collectionName, "filter": filter, "error": err}).Error("unable to delete the document")
		return err
	}
	if res.DeletedCount == 0 {
		log.WithFields(log.Fields{"collection": collectionName, "filter": filter}).Warning("document not found")
	}
	log.WithFields(log.Fields{"collection": collectionName, "filter": filter, "count": res.DeletedCount}).Debug("deleted document")

	return nil
}

// Delete removes all documents from the collection that matche the filter.
func (m *MongoDB) DeleteMany(collectionName string, filter interface{}) error {
	log.Trace("--> mongoDB.DeleteMany")
	defer log.Trace("<-- mongoDB.DeleteMany")

	ctx, cancel := context.WithTimeout(context.Background(), DB_TIMEOUT)
	defer cancel()

	collection, err := m.getCollection(ctx, collectionName)
	if err != nil {
		return err
	}

	res, err := collection.DeleteMany(ctx, filter)
	if err != nil {
		log.WithFields(log.Fields{"collection": collectionName, "filter": filter, "error": err}).Error("unable to delete the document")
		return err
	}
	if res.DeletedCount == 0 {
		log.WithFields(log.Fields{"collection": collectionName, "filter": filter}).Warning("document not found")
	}
	log.WithFields(log.Fields{"collection": collectionName, "filter": filter, "count": res.DeletedCount}).Debug("deleted document")

	return nil
}

// Close closes the mongo database client connection
func (m *MongoDB) Close() error {
	log.Trace("--> mongoDB.Close")
	defer log.Trace("<-- mongoDB.Close")

	ctx, cancel := context.WithTimeout(context.Background(), DB_TIMEOUT)
	defer cancel()
	if err := m.Client.Disconnect(ctx); err != nil {
		log.WithField("error", err).Error("unable to close the mongo database client")
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
			log.WithFields(log.Fields{"error": err}).Error("unable to connect to the MongoDB database")
			return nil, err
		}
	}

	db := m.Client.Database(m.dbname)
	collection := db.Collection(collectionName)
	if collection == nil {
		log.WithField("collection", collectionName).Error("uanble to access the collection")
		return nil, ErrCollectionNotAccessable
	}

	return collection, nil
}

// String returns the name of the database
func (db *MongoDB) String() string {
	return "mongo"
}
