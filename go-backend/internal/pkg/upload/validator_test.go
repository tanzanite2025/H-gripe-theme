package upload

import (
	"bytes"
	"encoding/base64"
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
	file := testFileHeader(t, "wheel.webp", validWebPFixture(t))

	rule := FileRule{
		MaxSize:             ProductImageRule.MaxSize,
		AllowedExtensions:   ProductImageRule.AllowedExtensions,
		AllowedContentTypes: ProductImageRule.AllowedContentTypes,
		MaxWidth:            6000,
		MaxHeight:           6000,
		MaxPixels:           16_000_000,
	}
	if err := ValidateFile(file, rule); err != nil {
		t.Fatalf("expected valid WebP dimensions to be accepted, got %v", err)
	}
}

func TestValidateFileRejectsTruncatedWebPWithValidHeader(t *testing.T) {
	data := validWebPFixture(t)
	data = data[:30]
	file := testFileHeader(t, "truncated.webp", data)

	err := ValidateFile(file, FileRule{
		MaxSize:             3 << 20,
		AllowedExtensions:   []string{".webp"},
		AllowedContentTypes: []string{"image/webp"},
		MaxWidth:            6000,
		MaxHeight:           6000,
		MaxPixels:           16_000_000,
	})
	if err == nil {
		t.Fatal("expected truncated WebP to be rejected")
	}
	if ErrorCode(err) != CodeInvalidType {
		t.Fatalf("expected %q, got %q", CodeInvalidType, ErrorCode(err))
	}
}

func TestValidateFileRejectsAnimatedWebPForGeneratedMedia(t *testing.T) {
	data := testWebPVP8X(75, 100)
	data[20] = 0x02
	file := testFileHeader(t, "animated.webp", data)

	err := ValidateFile(file, ProductImageRule)
	if err == nil {
		t.Fatal("expected animated WebP to be rejected")
	}
	if ErrorCode(err) != CodeInvalidType {
		t.Fatalf("expected %q, got %q", CodeInvalidType, ErrorCode(err))
	}
}

func TestValidateFileAcceptsFixedDimensionWebP(t *testing.T) {
	file := testFileHeader(t, "category.webp", validWebPFixture(t))

	rule := FileRule{
		MaxSize:             3 << 20,
		AllowedExtensions:   []string{".webp"},
		AllowedContentTypes: []string{"image/webp"},
		ExactWidth:          75,
		ExactHeight:         100,
	}
	if err := ValidateFile(file, rule); err != nil {
		t.Fatalf("expected fixed-size WebP to be accepted, got %v", err)
	}
}

func TestValidateSVGFileAccepts48By48SiteLogo(t *testing.T) {
	file := testFileHeader(t, "site-logo.svg", []byte(`<svg xmlns="http://www.w3.org/2000/svg" width="48" height="48" viewBox="0 0 48 48"><path d="M0 0h48v48H0z"/></svg>`))

	if err := ValidateSVGFile(file, SiteLogoSVGRule); err != nil {
		t.Fatalf("expected 48x48 SVG to be accepted, got %v", err)
	}
}

func TestValidateSVGFileAccepts48By48ViewBox(t *testing.T) {
	file := testFileHeader(t, "site-logo.svg", []byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 48 48"><path d="M0 0h48v48H0z"/></svg>`))

	if err := ValidateSVGFile(file, SiteLogoSVGRule); err != nil {
		t.Fatalf("expected 48x48 viewBox SVG to be accepted, got %v", err)
	}
}

func TestValidateSVGFileRejectsNon48By48SiteLogo(t *testing.T) {
	file := testFileHeader(t, "site-logo.svg", []byte(`<svg xmlns="http://www.w3.org/2000/svg" width="500" height="500"><path d="M0 0h500v500H0z"/></svg>`))

	err := ValidateSVGFile(file, SiteLogoSVGRule)
	if err == nil {
		t.Fatal("expected non-48x48 SVG to be rejected")
	}
	if ErrorCode(err) != CodeInvalidDimensions {
		t.Fatalf("expected %q, got %q", CodeInvalidDimensions, ErrorCode(err))
	}
}

func TestValidateSVGFileRejectsActiveContent(t *testing.T) {
	file := testFileHeader(t, "site-logo.svg", []byte(`<svg xmlns="http://www.w3.org/2000/svg" width="48" height="48"><script>alert(1)</script></svg>`))

	err := ValidateSVGFile(file, SiteLogoSVGRule)
	if err == nil {
		t.Fatal("expected active SVG content to be rejected")
	}
	if ErrorCode(err) != CodeInvalidType {
		t.Fatalf("expected %q, got %q", CodeInvalidType, ErrorCode(err))
	}
}

func TestReadImageDimensionsReturnsWebPDimensions(t *testing.T) {
	file := testFileHeader(t, "category.webp", validWebPFixture(t))

	width, height, err := ReadImageDimensions(file)
	if err != nil {
		t.Fatalf("expected image dimensions to be readable, got %v", err)
	}
	if width != 75 || height != 100 {
		t.Fatalf("expected 75x100 dimensions, got %dx%d", width, height)
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

func validWebPFixture(t *testing.T) []byte {
	t.Helper()

	const encoded = "UklGRrIBAABXRUJQVlA4TKUBAAAvSsAYAA8w//M///MfeJAkbXvaSG7m8Q3GfYSBJekwQztm/IcZlgwnmWImn2BK7aFmBtnVir6q//8VOkFE/xm4baTIu8c48ArEo6+B3zFKYln3pqClSCKX0begFTAXFOLXHSyF8cCNcZEG4OywuA4KVVfJCiArU7GAgJI8+lJP/OKMT/fBAjevg1cYB7YVkFuWga2lyPi5I0HFy5YTpWIHg0RZpkniRVW9odHAKOwosWuOGdxIyn2OvaCDvhg/we6TwadPBPbqBV58MsLmMJ8yZnOWk8SRz4N+QoyPL+MnamzMvcE1rHNEr91F9GKZPVUcS9w7PhhH36suB9qPeYb/oLk6cuTiJ0wOK3m5h1cKjW6EVZCYMK7dxcKCBdgP9HkKr9gkAO2P8GKZGWVdIAatQa+1IDpt6qyorVwdy01xdW8Jkfk6xjEXmVQQ+HQdFr6OKhIN34dXWq0+0qr6EJSCeeVLH9+gvGTLyqM65PQ44ihzlTXxQKjKbAvshXgir7Lil9w4L2bvMycmjQcqXaMCO6BlY28i+FOLzbfI1vEqxAhotocAAA=="
	data, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		t.Fatalf("decode WebP fixture: %v", err)
	}
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
