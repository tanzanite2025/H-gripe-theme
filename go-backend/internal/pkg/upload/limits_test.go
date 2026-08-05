package upload

import (
	"mime/multipart"
	"testing"
)

func TestValidateTotalSizeCoversMixedFileClasses(t *testing.T) {
	files := []*multipart.FileHeader{
		{Filename: "photo.jpg", Size: 60},
		{Filename: "claim.mp4", Size: 40},
	}

	if err := ValidateTotalSize(files, 100); err != nil {
		t.Fatalf("ValidateTotalSize rejected an exact-limit upload: %v", err)
	}
	if err := ValidateTotalSize(files, 99); ErrorCode(err) != CodeFileTooLarge {
		t.Fatalf("ValidateTotalSize error code = %q, want %q", ErrorCode(err), CodeFileTooLarge)
	}
}
