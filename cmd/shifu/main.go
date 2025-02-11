package main

import (
	shifu "github.com/emvi/shifu/pkg"
	"github.com/emvi/shifu/pkg/sync"
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
		server, err := shifu.NewServer(dir, shifu.ServerOptions{})

		if err != nil {
			slog.Error("Error setting up Shifu server", "error", err)
			return
		}

		if err := server.Start(nil, dir); err != nil {
			slog.Error("Error starting Shifu", "error", err)
		}

		break
	case "init":
		if err := shifu.Init(dir); err != nil {
			slog.Error("Error initializing new Shifu project", "error", err)
		}
		break
	case "version":
		slog.Info("Shifu version", "version", shifu.Version())
		break
	case "pull":
		if err := sync.Pull(dir); err != nil {
			slog.Error("Error pulling changes", "error", err)
		}

		break
	case "push":
		if err := sync.Push(dir); err != nil {
			slog.Error("Error pushing changes", "error", err)
		}

		break
	default:
		slog.Info("Command unknown. Usage: shifu run|init|version|pull|push <path>")
	}
}
