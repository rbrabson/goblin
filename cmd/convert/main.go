package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/joho/godotenv"
	"github.com/rbrabson/goblin/internal/convert"
)

const (
	GUILD_ID = "236523452230533121"
)

var (
	from_dir *os.File
)

func main() {
	godotenv.Load(".env")

	args := os.Args[1:]
	if len(args) < 1 {
		fmt.Println("Usage: convert <from_dir>")
		os.Exit(1)
	}
	var err error
	from_dir, err = os.Open(args[0])
	if err != nil {
		slog.Error("failed to open directory", slog.String("directory", args[0]), slog.Any("error", err))
		os.Exit(1)
	}

	fileNames, err := listFiles(from_dir)
	if err != nil {
		slog.Error("failed to list files", slog.String("directory", from_dir.Name()), slog.Any("error", err))
		os.Exit(1)
	}

	outDir := from_dir.Name() + "/" + "converted"
	convert.Initialize(GUILD_ID, outDir)

	slog.Info("Starting conversion", slog.String("output_directory", outDir))
	for _, fileName := range fileNames {
		fullFileName := from_dir.Name() + "/" + fileName
		switch fileName {
		case "heist.economy.json":
			convert.ConvertEconomy(fullFileName)
		case "heist.heist.json":
			slog.Info("Processing heist", slog.String("file", fullFileName))
		case "heist.mode.json":
			slog.Info("Processing mode", slog.String("file", fullFileName))
		case "heist.payday.json":
			slog.Info("Processing payday", slog.String("file", fullFileName))
		case "heist.race.json":
			convert.ConvertRaces(fullFileName)
		case "heist.reminder.json":
			slog.Info("Processing reminder", slog.String("file", fullFileName))
		case "heist.target.json":
			slog.Info("Processing target", slog.String("file", fullFileName))
		case "heist.theme.json":
			slog.Info("Processing theme", slog.String("file", fullFileName))
		}
	}
}

// listFiles lists the files in a directory
func listFiles(dir *os.File) ([]string, error) {
	files, err := dir.Readdirnames(0)
	if err != nil {
		return nil, err
	}
	return files, nil
}
