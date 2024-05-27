package main

import (
	shifu "github.com/emvi/shifu/pkg"
	"log/slog"
	"os"
	"strings"
)

const (
	shifuDirEnv = "SHIFU_DIR"
)

func main() {
	cmd := "run"
	dir := "."

	if os.Getenv(shifuDirEnv) != "" {
		dir = os.Getenv(shifuDirEnv)
	}

	if len(os.Args) > 1 {
		cmd = strings.ToLower(os.Args[1])
	}

	if len(os.Args) > 2 {
		dir = os.Args[2]
	}

	switch cmd {
	case "run":
		if err := shifu.Start(dir, nil); err != nil {
			slog.Error("Error starting Shifu: %s", err)
		}
	case "init":
		if err := shifu.Init(dir); err != nil {
			slog.Error("Error initializing new Shifu project: %s", err)
		}
	case "version":
		slog.Info("Shifu version", "version", shifu.Version())
	default:
		slog.Info("Command unknown. Usage: shifu run|init|version <path>")
	}
}
