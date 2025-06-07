package convert

var (
	GUILD_ID = "236523452230533121"
	OUT_DIR  string
)

// Initialize sets the guild ID and output directory
func Initialize(guildID string, outputDir string) {
	GUILD_ID = guildID
	OUT_DIR = outputDir
}
