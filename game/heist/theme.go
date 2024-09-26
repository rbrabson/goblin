package heist

var (
	themes = make(map[string]*Theme)
)

// Theme is a heist theme.
type Theme struct {
	ID       string        `json:"_id" bson:"_id"`
	Good     []GoodMessage `json:"good"`
	Bad      []BadMessage  `json:"bad"`
	Jail     string        `json:"jail" bson:"jail"`
	OOB      string        `json:"oob" bson:"oob"`
	Police   string        `json:"police" bson:"police"`
	Bail     string        `json:"bail" bson:"bail"`
	Crew     string        `json:"crew" bson:"crew"`
	Sentence string        `json:"sentence" bson:"sentence"`
	Heist    string        `json:"heist" bson:"heist"`
	Vault    string        `json:"vault" bson:"vault"`
}

type GoodMessage struct {
	Message string `json:"message" bson:"message"`
	Amount  int    `json:"amount" bson:"amount"`
}

type BadMessage struct {
	Message string `json:"message" bson:"message"`
	Result  string `json:"result" bson:"result"`
}

// GetThemeNames returns a list of available themes.
func GetThemeNames(map[string]*Theme) ([]string, error) {
	var fileNames []string
	for _, theme := range themes {
		fileNames = append(fileNames, theme.ID)
	}

	return fileNames, nil
}
