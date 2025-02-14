package database

import (
	"os"

	"github.com/rbrabson/dgame/database/file"
	"github.com/rbrabson/dgame/database/mongo"
	log "github.com/sirupsen/logrus"
)

// Client defines the methods required to query, load, and save documents for a given guild (server).
type Client interface {
	// Returns a list of all documents within the database collection
	ListDocuments(database string, collection string) ([]string, error)
	// Loads the document from the database collection
	Read(database string, collection string, documentID string, data interface{}) error
	// Creates or updates the document in the database collection
	Write(database string, collection string, documentID string, data interface{}) error
	// Closes the connection to the database
	Close() error
}

// NewClient creates a new store to be used to load and save documents used by the bot. By default,
// documents are stored on the local file system.
func NewClient() Client {
	log.Trace("--> database.newDatabase")
	defer log.Trace("<-- database.newDatabase")

	dbaseType := os.Getenv("DISCORD_DATABASE")

	var database Client
	switch dbaseType {
	case "file":
		database = file.NewDatabase()
	case "mongo":
		database = mongo.NewDatabase()
	default:
		database = file.NewDatabase()
	}

	log.WithField("type", dbaseType).Info("using database")
	return database
}
