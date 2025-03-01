package mongo

import "errors"

var (
	ErrDocumentNotFound        = errors.New("document not found")
	ErrInvalidDocument         = errors.New("unable to decode document")
	ErrDbInaccessable          = errors.New("unable to create or access the database")
	ErrCollectionNotAccessable = errors.New("unable to create or access the collection")
)
