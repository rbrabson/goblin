package discord

import (
	"log/slog"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	SERVER_COLLECTION = "goblin"
)

// ReadServer reads the server from the database
func ReadServer() *Server {
	filter := bson.M{}
	var server Server
	err := db.FindOne(SERVER_COLLECTION, filter, &server)
	if err != nil {
		slog.Debug("server not found in the database",
			slog.Any("error", err),
		)
		return nil
	}
	slog.Debug("read server from the database")
	return &server
}

// WriteServer writes the server to the database
func WriteServer(server *Server) error {
	var filter bson.M
	if server.ID == primitive.NilObjectID {
		filter = bson.M{}
	} else {
		filter = bson.M{"_id": server.ID}
	}
	err := db.UpdateOrInsert(SERVER_COLLECTION, filter, server)
	if err != nil {
		slog.Error("unable to save server to the database",
			slog.Any("filter", filter),
			slog.Any("error", err),
		)
		return err
	}
	slog.Debug("save server to the database",
		slog.Any("filter", filter),
	)

	return nil
}
