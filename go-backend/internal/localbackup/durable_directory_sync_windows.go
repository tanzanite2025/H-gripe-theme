package localbackup

import "os"

// Windows does not implement fsync on directory handles through os.File.
func syncDirectoryFile(directory *os.File) error { return nil }
