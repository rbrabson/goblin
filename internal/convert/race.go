package convert

import "fmt"

func ConvertRaces(fileName string) {
	fmt.Printf("convert race, file=%s\n", fileName)

	bytes := readFile(fileName)
	fileContents := asArray(bytes)
	for _, fileContent := range fileContents {
		guildID := asString(fileContent["_id"])
		if guildID == GUILD_ID {
			convertRaceModel(fileContent)
		}
	}
}

func convertRaceModel(raceModel map[string]interface{}) {
	fmt.Println("TODO: convert once we have a new model")
}
