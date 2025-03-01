package convert

import (
	log "github.com/sirupsen/logrus"
)

var (
	GUILD_ID = "236523452230533121"
	OUT_DIR  string
)

func Intialize(guildID string, outputDir string) {
	GUILD_ID = guildID
	OUT_DIR = outputDir
	log.SetLevel(log.TraceLevel)
}
