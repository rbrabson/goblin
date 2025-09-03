package mongo

import "errors"

var (
	ErrDocumentNotFound        = errors.New("document not found")
	ErrInvalidDocument         = errors.New("unable to decode document")
	ErrCollectionNotAccessible = errors.New("unable to create or access the collection")
)
