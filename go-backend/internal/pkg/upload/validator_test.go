package upload

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestValidateFileRejectsScriptRenamedAsVideo(t *testing.T) {
	file := testFileHeader(t, "payload.mp4", []byte("<?php echo 'owned'; ?>"))

	err := ValidateFile(file, WarrantyVideoRule)
	if err == nil {
		t.Fatal("expected renamed script upload to be rejected")
	}
	if ErrorCode(err) != CodeInvalidType {
		t.Fatalf("expected %q, got %q", CodeInvalidType, ErrorCode(err))
	}
}

func TestValidateFileAcceptsMP4Magic(t *testing.T) {
	file := testFileHeader(t, "clip.mp4", append([]byte{
		0x00, 0x00, 0x00, 0x18,
		'f', 't', 'y', 'p',
		'i', 's', 'o', 'm',
	}, bytes.Repeat([]byte{0x00}, 32)...))

	if err := ValidateFile(file, WarrantyVideoRule); err != nil {
		t.Fatalf("expected MP4 upload to be accepted, got %v", err)
	}
}

func TestValidateFileRejectsUnsupportedExtension(t *testing.T) {
	file := testFileHeader(t, "payload.php", []byte("<?php echo 'owned'; ?>"))

	err := ValidateFile(file, SuggestionImageRule)
	if err == nil {
		t.Fatal("expected unsupported extension to be rejected")
	}
	if ErrorCode(err) != CodeInvalidType {
		t.Fatalf("expected %q, got %q", CodeInvalidType, ErrorCode(err))
	}
}

func testFileHeader(t *testing.T, filename string, contents []byte) *multipart.FileHeader {
	t.Helper()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		t.Fatalf("create form file: %v", err)
	}
	if _, err := part.Write(contents); err != nil {
		t.Fatalf("write form file: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close multipart writer: %v", err)
	}

	request := httptest.NewRequest(http.MethodPost, "/", body)
	request.Header.Set("Content-Type", writer.FormDataContentType())
	if err := request.ParseMultipartForm(int64(body.Len() + 1024)); err != nil {
		t.Fatalf("parse multipart form: %v", err)
	}

	files := request.MultipartForm.File["file"]
	if len(files) != 1 {
		t.Fatalf("expected 1 uploaded file, got %d", len(files))
	}
	return files[0]
}
