package service

import (
	"bytes"
	"context"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"commerce-platform/internal/testfixtures"
)

func createSiteLogoAndFaviconLifecycleMultipartHeaderFromBackendOwnedWebPFixture(t *testing.T, filename string) *multipart.FileHeader {
	t.Helper()
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := part.Write(testfixtures.SiteLogoAndFaviconUploadLifecycle512By512LosslessWebPFixtureBytes()); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	request := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/", &body)
	request.Header.Set("Content-Type", writer.FormDataContentType())
	if err := request.ParseMultipartForm(int64(body.Len() + 1024)); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = request.MultipartForm.RemoveAll() })
	files := request.MultipartForm.File["file"]
	if len(files) != 1 {
		t.Fatalf("expected one fixture upload, got %d", len(files))
	}
	return files[0]
}
