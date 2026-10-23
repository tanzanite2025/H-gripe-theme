//go:build !windows

package localbackup

import (
	"fmt"
	"golang.org/x/sys/unix"
)

func customerServiceCandidateCapacityFilesystemIdentity(directory string) (string, error) {
	var status unix.Stat_t
	if err := unix.Stat(directory, &status); err != nil {
		return "", err
	}
	return fmt.Sprintf("unix-device:%d", status.Dev), nil
}
