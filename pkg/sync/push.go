package sync

import "github.com/emvi/shifu/pkg/cfg"

// Push pushes changed static and content files to a remote Shifu server.
func Push(dir string) error {
	if err := cfg.Load(dir, nil); err != nil {
		return err
	}

	if err := hasRemoteConfig(); err != nil {
		return err
	}

	// TODO

	return nil
}
