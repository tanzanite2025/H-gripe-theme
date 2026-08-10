package upload

import (
	"bytes"
	"encoding/binary"
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

func TestValidateFileRejectsWebPWithExcessivePixelCount(t *testing.T) {
	file := testFileHeader(t, "oversized.webp", testWebPVP8X(8000, 8000))

	err := ValidateFile(file, ProductImageRule)
	if err == nil {
		t.Fatal("expected oversized WebP dimensions to be rejected")
	}
	if ErrorCode(err) != CodeInvalidDimensions {
		t.Fatalf("expected %q, got %q", CodeInvalidDimensions, ErrorCode(err))
	}
	if HTTPStatus(err) != http.StatusUnprocessableEntity {
		t.Fatalf("unexpected HTTP status %d", HTTPStatus(err))
	}
}

func TestValidateFileAcceptsWebPWithinProductLimits(t *testing.T) {
	file := testFileHeader(t, "wheel.webp", testWebPVP8X(2400, 1600))

	if err := ValidateFile(file, ProductImageRule); err != nil {
		t.Fatalf("expected valid WebP dimensions to be accepted, got %v", err)
	}
}

func TestValidateFileAcceptsFixedProductTypeWebP(t *testing.T) {
	file := testFileHeader(t, "category.webp", testWebPVP8X(ProductTypeImageWidth, ProductTypeImageHeight))

	if err := ValidateFile(file, ProductTypeImageRule); err != nil {
		t.Fatalf("expected fixed-size product type WebP to be accepted, got %v", err)
	}
}

func TestValidateFileRejectsProductTypeWebPWithWrongDimensions(t *testing.T) {
	file := testFileHeader(t, "category.webp", testWebPVP8X(ProductTypeImageWidth, ProductTypeImageHeight-1))

	err := ValidateFile(file, ProductTypeImageRule)
	if err == nil {
		t.Fatal("expected product type image with wrong dimensions to be rejected")
	}
	if ErrorCode(err) != CodeInvalidDimensions {
		t.Fatalf("expected %q, got %q", CodeInvalidDimensions, ErrorCode(err))
	}
	if HTTPStatus(err) != http.StatusUnprocessableEntity {
		t.Fatalf("unexpected HTTP status %d", HTTPStatus(err))
	}
}

func testWebPVP8X(width, height int) []byte {
	data := make([]byte, 30)
	copy(data[0:4], []byte("RIFF"))
	binary.LittleEndian.PutUint32(data[4:8], uint32(len(data)-8))
	copy(data[8:12], []byte("WEBP"))
	copy(data[12:16], []byte("VP8X"))
	width--
	height--
	data[24] = byte(width)
	data[25] = byte(width >> 8)
	data[26] = byte(width >> 16)
	data[27] = byte(height)
	data[28] = byte(height >> 8)
	data[29] = byte(height >> 16)
	return data
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
