package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
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
	// to_dir, err = os.Open(args[1])
	// if err != nil {
	// 	log.Fatal(err)
	// }

	files, err := listFiles(from_dir)
	if err != nil {
		log.Fatal(err)
	}
	fileName := from_dir.Name() + "/" + files[0]
	b := readFile(fileName)
	out := asArray(b)
	x := out[0]["characters"]
	raceModel := asMap(x)
	convertRaceModel(raceModel)
}

// listFiles lists the files in a directory
func listFiles(dir *os.File) ([]string, error) {
	files, err := dir.Readdirnames(0)
	if err != nil {
		return nil, err
	}
	return files, nil
}

// readFile gets the contents of a file
func readFile(fileName string) []byte {
	out, err := os.ReadFile(fileName)
	if err != nil {
		log.Fatal(err)
	}

	return out
}

// asArray converts a byte slice to an array of interfaces
func asArray(bytes []byte) []map[string]interface{} {
	var out []map[string]interface{}
	json.Unmarshal(bytes, &out)
	return out
}

// asMap converts an interface slice to a slice of maps
func asMap(elements interface{}) []map[string]interface{} {
	bytes, err := json.Marshal(elements)
	if err != nil {
		log.Fatal(err)
	}
	var out []map[string]interface{}
	json.Unmarshal(bytes, &out)
	return out
}

func convertRaceModel(raceModel []map[string]interface{}) {
	for _, race := range raceModel {
		for k, v := range race {
			fmt.Println(k, v)
		}
	}
}
