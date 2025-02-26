package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/rbrabson/goblin/internal/convert"
	log "github.com/sirupsen/logrus"
)

const (
	GUILD_ID = "236523452230533121"
)

var (
	from_dir *os.File
	// to_dir   *os.File
)

// setLogLevel sets the logging level. If the LOG_LEVEL environment variable isn't set or the value
// isn't recognized, logging defaults to the `debug` level
func setLogLevel() {
	switch strings.ToLower(os.Getenv("LOG_LEVEL")) {
	case "panic":
		log.SetLevel(log.PanicLevel)
	case "fatal":
		log.SetLevel(log.FatalLevel)
	case "error":
		log.SetLevel(log.ErrorLevel)
	case "warn":
		log.SetLevel(log.WarnLevel)
	case "info":
		log.SetLevel(log.InfoLevel)
	case "debug":
		log.SetLevel(log.DebugLevel)
	case "trace":
		log.SetLevel(log.TraceLevel)
	default:
		log.SetLevel(log.DebugLevel)
	}
}

func main() {
	godotenv.Load()
	setLogLevel()

	args := os.Args[1:]
	if len(args) < 2 {
		fmt.Println("Usage: convert <amount> <from_dir> <to_dir>")
		os.Exit(1)
	}
	var err error
	from_dir, err = os.Open(args[0])
	if err != nil {
		log.Fatal(err)
	}

	fileNames, err := listFiles(from_dir)
	if err != nil {
		log.Fatal(err)
	}

	outDir := from_dir.Name() + "/" + "converted"
	convert.Intialize(GUILD_ID, outDir)

	for _, fileName := range fileNames {
		fullFileName := from_dir.Name() + "/" + fileName
		switch fileName {
		case "Heist.economy.json":
			convert.ConvertEconomy(fullFileName)
		case "Heist.heist.json":
			fmt.Println("heist")
		case "Heist.mode.json":
			fmt.Println("mode")
		case "Heist.payday.json":
			fmt.Println("payday")
		case "Heist.race.json":
			convert.ConvertRaces(fullFileName)
		case "Heist.reminder.json":
			fmt.Println("reminder")
		case "Heist.target.json":
			fmt.Println("target")
		case "Heist.theme.json":
			fmt.Println("theme")
		}
	}

	// fileName := from_dir.Name() + "/" + fileNames[0]
	// b := readFile(fileName)
	// out := asArray(b)
	// x := out[0]["characters"]
	// raceModel := asMap(x)
	// convertRaceModel(raceModel)
}

// listFiles lists the files in a directory
func listFiles(dir *os.File) ([]string, error) {
	files, err := dir.Readdirnames(0)
	if err != nil {
		return nil, err
	}
	return files, nil
}
