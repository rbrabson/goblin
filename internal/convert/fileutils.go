package convert

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"time"
)

// readFile gets the contents of a file
func readFile(fileName string) []byte {
	out, err := os.ReadFile(fileName)
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
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
func asMap(elements interface{}) map[string]interface{} {
	bytes, err := json.Marshal(elements)
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
	var out map[string]interface{}
	json.Unmarshal(bytes, &out)
	return out
}

// asString converts an interface to a string
func asString(element interface{}) string {
	return fmt.Sprintf("%v", element)
}

func asInteger(element interface{}) int {
	return int(element.(float64))
}

func asTime(element interface{}) time.Time {
	mapElement := asMap(element)
	bytes, err := json.Marshal(mapElement["$date"])
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
	var out time.Time
	json.Unmarshal(bytes, &out)
	return out
}
