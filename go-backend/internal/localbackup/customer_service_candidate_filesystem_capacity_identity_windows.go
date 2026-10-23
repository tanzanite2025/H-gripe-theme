package localbackup

import (
	"fmt"
	"golang.org/x/sys/windows"
)

func customerServiceCandidateCapacityFilesystemIdentity(directory string) (string, error) {
	path, err := windows.UTF16PtrFromString(directory)
	if err != nil {
		return "", err
	}
	handle, err := windows.CreateFile(path, 0, windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE|windows.FILE_SHARE_DELETE, nil, windows.OPEN_EXISTING, windows.FILE_FLAG_BACKUP_SEMANTICS, 0)
	if err != nil {
		return "", err
	}
	defer windows.CloseHandle(handle)
	var information windows.ByHandleFileInformation
	if err := windows.GetFileInformationByHandle(handle, &information); err != nil {
		return "", err
	}
	// A serial collision only combines otherwise separate budgets conservatively.
	return fmt.Sprintf("windows-volume:%08x", information.VolumeSerialNumber), nil
}
