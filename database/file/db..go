package file

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"
)

// FileDB is a database used to load and save documents to a file on the local file system.
type FileDB struct {
	dir string
}

// NewDatabase creates a new filesystem database
func NewDatabase() *FileDB {
	log.Trace("--> file.NewDatabase")
	defer log.Trace("<-- file.NewDatabase")

	dir := os.Getenv("DISCORD_FILESYSTEM_DIR")
	db := &FileDB{
		dir: dir,
	}
	return db
}

// ListDocuments returns the list of files in the sub-directory (collection).
func (db *FileDB) ListDocuments(database string, collection string) ([]string, error) {
	log.Trace("--> FileDB.List")
	defer log.Trace("<-- FileDB.List")

	dirName := fmt.Sprintf("%s/%s/%s", db.dir, database, collection)
	files, err := os.ReadDir(dirName)
	if err != nil {
		log.WithFields(log.Fields{"database": database, "collection": collection, "error": err}).Error("failed to get the list of documents")
		return nil, err
	}
	fileNames := make([]string, 0, len(files))
	for _, file := range files {
		split := strings.Split(file.Name(), ".json")
		fileNames = append(fileNames, split[0])
	}
	return fileNames, nil
}

// Read loads a file identified by documentID from the subdirectory (collection) into data.
func (db *FileDB) Read(database string, collection string, documentID string, data interface{}) error {
	log.Trace("--> FileDB.Read")
	defer log.Trace("<-- FileDB.Read")

	filename := fmt.Sprintf("%s/%s/%s/%s.json", db.dir, database, collection, documentID)
	b, err := os.ReadFile(filename)
	if err != nil {
		log.WithFields(log.Fields{"database": database, "collection": collection, "documentID": documentID}).Error("failed to read the document")
		return err
	}

	err = json.Unmarshal(b, data)
	if err != nil {
		log.WithFields(log.Fields{"database": database, "collection": collection, "documentID": documentID, "error": err}).Error("failed to unmarshall the document")
		return err
	}
	return nil
}

// Write stores data into a subdirectory (collection) with the file name documentID.
func (db *FileDB) Write(database string, collection string, documentID string, data interface{}) error {
	log.Trace("--> FileDB.Write")
	defer log.Trace("<-- FileDB.Write")

	b, err := json.Marshal(data)
	if err != nil {
		log.WithFields(log.Fields{"database": database, "collection": collection, "documentID": documentID, "error": err}).Error("failed to marshall the document")
		return err
	}

	filename := fmt.Sprintf("%s/%s/%s/%s.json", db.dir, database, collection, documentID)
	err = os.WriteFile(filename, b, 0644)
	if err != nil {
		log.WithFields(log.Fields{"database": database, "collection": collection, "documentID": documentID, "error": err}).Error("failed to write the document")
		return err
	}
	return nil
}

// Close closes the access to the file system
func (db *FileDB) Close() error {
	// NO-OP
	return nil
}
