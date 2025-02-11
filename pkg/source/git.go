package source

import (
	"context"
	"log/slog"
	"os"
	"os/exec"
	"sync"
	"time"
)

// Git loads the website data from a Git repository.
type Git struct {
	dir           string
	repository    string
	updateSeconds int
	lastUpdate    time.Time
	m             sync.RWMutex
}

// NewGit creates a new Provider for Git.
func NewGit(dir, repository string, updateSeconds int) *Git {
	if updateSeconds == 0 {
		updateSeconds = 15 * 60
	}

	provider := &Git{
		dir:           dir,
		repository:    repository,
		updateSeconds: updateSeconds,
	}
	provider.init()
	return provider
}

// Watch implements the Provider interface.
func (provider *Git) Watch(ctx context.Context, update func()) {
	if provider.repository != "" {
		go func() {
			timerDuration := time.Second * time.Duration(provider.updateSeconds)
			timer := time.NewTimer(timerDuration)
			defer timer.Stop()

			for {
				timer.Reset(timerDuration)

				select {
				case <-timer.C:
					provider.Update(update)
				case <-ctx.Done():
					return
				}
			}
		}()
	}

	update()
}

// Update implements the Provider interface.
func (provider *Git) Update(update func()) {
	provider.pullGitRepo()
	provider.m.Lock()
	update()
	provider.lastUpdate = time.Now().UTC()
	provider.m.Unlock()
}

// LastUpdate implements the Provider interface.
func (provider *Git) LastUpdate() time.Time {
	provider.m.RLock()
	defer provider.m.RUnlock()
	return provider.lastUpdate
}

func (provider *Git) init() {
	if provider.repository != "" {
		if _, err := os.Stat(provider.dir); os.IsNotExist(err) {
			slog.Info("Cloning website git repository...")

			if err := os.MkdirAll(provider.dir, 0744); err != nil {
				slog.Error("Error creating repository directory", "error", err)
				return
			}

			cmd := exec.Command("git", "clone", provider.repository, provider.dir)
			out, err := cmd.CombinedOutput()

			if err != nil {
				slog.Error("Error cloning git repository", "repo", provider.repository, "error", err, "output", string(out))
				return
			}

			cmd = exec.Command("git", "submodule", "update", "--init", "--recursive")
			cmd.Dir = provider.dir
			out, err = cmd.CombinedOutput()

			if err != nil {
				slog.Error("Error loading git submodules", "repo", provider.repository, "error", err, "output", string(out))
				return
			}

			slog.Info("Git repository cloned")
		} else {
			slog.Info("Git repository already exists, no initialization required")
		}

		provider.pullGitRepo()
	}
}

func (provider *Git) pullGitRepo() bool {
	slog.Info("Checking for website updates...")
	cmd := exec.Command("git", "remote", "update")
	cmd.Dir = provider.dir
	out, err := cmd.CombinedOutput()

	if err != nil {
		slog.Error("Error updating remote git repository", "error", err, "output", string(out))
		return false
	}

	cmd = exec.Command("git", "rev-parse", "@")
	cmd.Dir = provider.dir
	localRev, err := cmd.CombinedOutput()

	if err != nil {
		slog.Error("Error parsing local git revision hash", "error", err, "rev", string(localRev))
		return false
	}

	cmd = exec.Command("git", "rev-parse", "@{u}")
	cmd.Dir = provider.dir
	remoteRev, err := cmd.CombinedOutput()

	if err != nil {
		slog.Error("Error parsing remote git revision hash", "error", err, "rev", string(remoteRev))
		return false
	}

	if string(localRev) != string(remoteRev) {
		slog.Info("Updating website git repository...")
		cmd = exec.Command("git", "pull", "--rebase")
		cmd.Dir = provider.dir
		out, err = cmd.CombinedOutput()

		if err != nil {
			slog.Error("Error pulling from remote git repository", "error", err, "output", string(out))
			return false
		}

		cmd = exec.Command("git", "submodule", "update", "--recursive", "--remote")
		cmd.Dir = provider.dir
		out, err = cmd.CombinedOutput()

		if err != nil {
			slog.Error("Error updating git submodules", "error", err, "output", string(out))
			// continue in this case...
		}

		slog.Info("Git repository updated", "local_rev", string(localRev), "remote_rev", string(remoteRev))
		return true
	}

	slog.Info("Done checking for website updates")
	return false
}
