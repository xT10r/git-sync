//go:build !linux && !darwin && !freebsd

package snapshot

import (
	"errors"
	"os"
)

func lockFile(_ *os.File) error {
	return errors.New("snapshot publication currently requires Linux, macOS or FreeBSD with local filesystem locking and symlinks")
}
