package convert

import (
	"fmt"

	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
)

var (
	GUILD_ID = "236523452230533121"
	OUT_DIR  string
)

func init() {
	err := godotenv.Load(".env_test")
	if err != nil {
		fmt.Println("unable to load .env_test file")
	}
	GUILD_ID = "123456789012345678"
	log.SetLevel(log.TraceLevel)
}

func Intialize(guildID string, outputDir string) {
	GUILD_ID = guildID
	OUT_DIR = outputDir
	log.SetLevel(log.TraceLevel)
}
