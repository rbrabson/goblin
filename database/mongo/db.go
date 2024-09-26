package mongo

import (
	"context"
	"os"
	"time"

	"github.com/joho/godotenv"
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
	client *mongo.Client
}

// getClient returns a mongo database client
func getClient() *mongo.Client {
	log.Trace("--> mongo.getClient")

	godotenv.Load()
	uri := os.Getenv("MONGODB_URI")
	if uri == "" {
		log.Fatal("You must set your 'MONGODB_URI' environmental variable. See\n\t https://www.mongodb.com/docs/drivers/go/current/usage-examples/#environment-variable")
	}

	// Wait for MongoDB to become active before proceeding
	ctx, cancel := context.WithTimeout(context.Background(), DB_TIMEOUT)
	defer cancel()

	clientOpts := options.Client().ApplyURI(uri)
	client, err := mongo.Connect(ctx, clientOpts)
	if err != nil {
		log.WithField("error", err).Fatal("unable to connect to the mongo database")
	}

	log.Trace("<-- mongo.getClient")
	return client
}

// NewDatabase creates a database to load and save documents in a MongoDB database.
func NewDatabase() *MongoDB {
	log.Trace("--> mongo.NewDatabase")
	defer log.Trace("<-- mongo.NewDatabase")

	mongoDB := &MongoDB{
		client: getClient(),
	}

	return mongoDB
}

// ListDocuments returns the ID of each document in a collection in the database.
func (m *MongoDB) ListDocuments(dbName string, collectionName string) []string {
	log.Trace("--> MongoDB.ListDocuments")
	defer log.Trace("<-- MongoDB.ListDocuments")

	database := m.client.Database(dbName)
	if database == nil {
		log.WithField("database", dbName).Error("unable to create or access the database")
		return nil
	}
	collection := database.Collection(collectionName)
	if collection == nil {
		log.WithFields(log.Fields{"database": dbName, "collection": collectionName}).Error("unable to create or access the collection")
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), DB_TIMEOUT)
	defer cancel()
	opts := options.Find().SetProjection(bson.M{"_id": 1})
	cur, err := collection.Find(ctx, bson.D{}, opts)
	if err != nil {
		log.WithFields(log.Fields{"database": dbName, "collection": collectionName, "error": err}).Error("unable to search the database")
		return nil
	}
	type result struct {
		ID string `bson:"_id"`
	}
	var results []result

	err = cur.All(ctx, &results)
	if err != nil {
		log.WithFields(log.Fields{"database": dbName, "collection": collectionName, "error": err}).Error("unable to retrieve the IDs for the collection")
		log.Errorf("Error getting the IDs for collection %s, error=%s", collectionName, err.Error())
	}

	idList := make([]string, 0, len(results))
	for _, r := range results {
		idList = append(idList, r.ID)
	}

	return idList
}

// Load loads a document identified by documentID from the collection into data.
func (m *MongoDB) Load(dbName string, collectionName string, documentID string, data interface{}) {
	log.Trace("--> MongoDB.Load")
	defer log.Trace("<-- MongoDB.Load")

	db := m.client.Database(dbName)
	if db == nil {
		log.WithFields(log.Fields{"database": dbName, "collection": collectionName}).Error("unable to create or access the database")
		return
	}
	collection := db.Collection(collectionName)
	if collection == nil {
		log.WithFields(log.Fields{"database": dbName, "collection": collectionName}).Error("unable to create or access the collection")
		return
	}
	log.Debug("Collection:", collection.Name())

	ctx, cancel := context.WithTimeout(context.Background(), DB_TIMEOUT)
	defer cancel()
	res := collection.FindOne(ctx, bson.D{{Key: "_id", Value: documentID}})
	if res == nil {
		log.WithFields(log.Fields{"database": dbName, "collection": collectionName, "document": documentID}).Error("unable to find the document")
	}
	err := res.Decode(data)
	if err != nil {
		log.WithFields(log.Fields{"database": dbName, "collection": collectionName, "document": documentID, "error": err}).Error("unable to decode the document")
	}
}

// Save stores data into a documeent within the specified collection.
func (m *MongoDB) Save(dbName string, collectionName string, documentID string, data interface{}) {
	log.Trace("--> MongoDB.Save")
	defer log.Trace("<-- MongoDB.Save")

	findOptions := options.Find()
	// Set the limit of the number of record to find
	findOptions.SetLimit(5)
	defer log.Debug("disconnected from mongo database")

	db := m.client.Database(dbName)
	if db == nil {
		log.WithFields(log.Fields{"database": dbName, "collection": collectionName}).Error("unable to create or access the database")
		return
	}
	collection := db.Collection(collectionName)
	if collection == nil {
		ctx, cancel := context.WithTimeout(context.Background(), DB_TIMEOUT)
		defer cancel()
		if err := db.CreateCollection(ctx, collectionName); err != nil {
			log.WithFields(log.Fields{"collection": collectionName, "error": err}).Error("unable to create the collection")
			return
		}
		collection = db.Collection(collectionName)
	}

	ctx, cancel := context.WithTimeout(context.Background(), DB_TIMEOUT)
	defer cancel()
	_, err := collection.InsertOne(ctx, data)
	if err != nil {
		_, err = collection.ReplaceOne(ctx, bson.D{{Key: "_id", Value: documentID}}, data)
		if err != nil {
			log.WithFields(log.Fields{"collection": collectionName, "document": documentID, "error": err}).Error("unable to insert the document the collection")
		}
	}
}

// Close closes the mongo database client connection
func (m *MongoDB) Close() {
	log.Trace("--> MongoDB.Close")
	defer log.Trace("<-- MongoDB.Close")

	ctx, cancel := context.WithTimeout(context.Background(), DB_TIMEOUT)
	defer cancel()
	if err := m.client.Disconnect(ctx); err != nil {
		log.Error("unable to close the mongo database client")
	}
}
