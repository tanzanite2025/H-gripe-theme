//go:build !windows

package localbackup

import "golang.org/x/sys/unix"

func availableDiskBytes(directory string) (uint64, error) {
	var status unix.Statfs_t
	err := unix.Statfs(directory, &status)
	return status.Bavail * uint64(status.Bsize), err
}
