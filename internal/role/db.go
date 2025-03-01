package role

import (
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
)

// readServer gets the server from the database and returns the value, if it exists, or returns nil if the
func readServer(guildID string) *Server {
	log.Trace("--> server.readServer")
	defer log.Trace("<-- server.readServer")

	filter := bson.M{"guild_id": guildID}
	var server Server
	err := db.FindOne(SERVER_COLLECTION, filter, &server)
	if err != nil {
		log.WithFields(log.Fields{"guild": guildID}).Debug("server not found in the database")
		return nil
	}
	log.WithFields(log.Fields{"server": server}).Debug("read server from the database")
	return &server
}

// writeServer creates or updates the server data in the database being used by the Discord bot.
func writeServer(server *Server) error {
	log.Trace("--> server.writeServer")
	defer log.Trace("<-- server.writeServer")

	filter := bson.M{"guild_id": server.GuildID}
	err := db.UpdateOrInsert(SERVER_COLLECTION, filter, server)
	if err != nil {
		log.WithFields(log.Fields{"guild": server.GuildID}).Error("unable to save server to the database")
		return err
	}
	log.WithFields(log.Fields{"guild": server.GuildID}).Info("save server to the database")

	return nil
}
