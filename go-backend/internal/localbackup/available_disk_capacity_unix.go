//go:build !windows

package localbackup

import (
	"errors"

	"golang.org/x/sys/unix"
)

func availableDiskBytes(directory string) (uint64, error) {
	var status unix.Statfs_t
	if err := unix.Statfs(directory, &status); err != nil {
		return 0, err
	}
	if status.Bsize <= 0 {
		return 0, errors.New("filesystem block size must be positive")
	}
	blockSize := uint64(status.Bsize)
	if status.Bavail > ^uint64(0)/blockSize {
		return 0, errors.New("available filesystem capacity exceeds supported range")
	}
	return status.Bavail * blockSize, nil
}
