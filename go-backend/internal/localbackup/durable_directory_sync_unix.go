//go:build !windows

package localbackup

import "os"

func syncDirectoryFile(directory *os.File) error { return directory.Sync() }
