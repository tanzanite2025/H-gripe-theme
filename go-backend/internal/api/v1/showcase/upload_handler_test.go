package showcase

import (
	"bytes"
	"context"
	"encoding/base64"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	showcasedomain "commerce-platform/internal/domain/showcase"
	"commerce-platform/internal/service"

	"github.com/gin-gonic/gin"
)

func TestShowcaseUploadFailsClosedWithoutUploadProtection(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := NewShowcaseHandler(nil)
	router := gin.New()
	router.POST("/showcase/upload", func(c *gin.Context) {
		c.Set("user_id", uint(42))
		handler.Upload(c)
	})

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/showcase/upload", strings.NewReader(""))
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusServiceUnavailable)
	}
	if !strings.Contains(recorder.Body.String(), "Upload protection is temporarily unavailable") {
		t.Fatalf("body = %q, want upload protection unavailable message", recorder.Body.String())
	}
}

func TestShowcaseUploadBudgetBytesUsesRequestLengthWhenKnown(t *testing.T) {
	const requestBytes = int64(12 << 20)
	if actual := showcaseUploadBudgetBytes(requestBytes); actual != requestBytes {
		t.Fatalf("showcaseUploadBudgetBytes(%d) = %d, want %d", requestBytes, actual, requestBytes)
	}
}

func TestShowcaseUploadBudgetBytesUsesMaximumForUnknownLength(t *testing.T) {
	for _, contentLength := range []int64{-1, 0} {
		if actual := showcaseUploadBudgetBytes(contentLength); actual != showcaseMaxRequestBytes {
			t.Fatalf("showcaseUploadBudgetBytes(%d) = %d, want %d", contentLength, actual, showcaseMaxRequestBytes)
		}
	}
}

func TestReadShowcaseUploadParamsTrimsAndRequiresRegion(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Request = httptest.NewRequest("POST", "/showcase/upload", strings.NewReader("region=%20US%20&notes=%20hello%20"))
	context.Request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	params, code, message := readShowcaseUploadParams(context)
	if code != "" || message != "" {
		t.Fatalf("readShowcaseUploadParams() code=%q message=%q", code, message)
	}
	if params["region"] != "US" || params["notes"] != "hello" {
		t.Fatalf("params = %#v, want trimmed region and notes", params)
	}
}

func TestReadShowcaseUploadParamsRejectsLongFields(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	longNotes := strings.Repeat("a", showcaseMaxNotes+1)
	context.Request = httptest.NewRequest("POST", "/showcase/upload", strings.NewReader("region=US&notes="+longNotes))
	context.Request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	_, code, message := readShowcaseUploadParams(context)
	if code != "tpg_field_too_long" || !strings.Contains(message, "notes") {
		t.Fatalf("readShowcaseUploadParams() code=%q message=%q, want notes field length error", code, message)
	}
}

func TestReadShowcaseUploadOrderIDRequiresPositiveInteger(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name     string
		body     string
		wantID   uint
		wantCode string
	}{
		{name: "missing", body: "region=US", wantCode: "showcase_upload_order_required"},
		{name: "invalid", body: "region=US&order_id=abc", wantCode: "showcase_upload_order_invalid"},
		{name: "zero", body: "region=US&order_id=0", wantCode: "showcase_upload_order_invalid"},
		{name: "valid", body: "region=US&order_id=42", wantID: 42},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			context, _ := gin.CreateTestContext(recorder)
			context.Request = httptest.NewRequest(
				http.MethodPost,
				"/showcase/upload",
				strings.NewReader(test.body),
			)
			context.Request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

			orderID, code, _ := readShowcaseUploadOrderID(context)
			if orderID != test.wantID {
				t.Fatalf("order id = %d, want %d", orderID, test.wantID)
			}
			if code != test.wantCode {
				t.Fatalf("code = %q, want %q", code, test.wantCode)
			}
		})
	}
}

func TestReadShowcaseUploadOrderIDRejectsUintOverflow(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Request = httptest.NewRequest(
		http.MethodPost,
		"/showcase/upload",
		strings.NewReader("region=US&order_id="+strconv.FormatUint(^uint64(0), 10)),
	)
	context.Request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	_, code, _ := readShowcaseUploadOrderID(context)
	if code != "showcase_upload_order_invalid" {
		t.Fatalf("code = %q, want showcase_upload_order_invalid", code)
	}
}

func TestShowcaseUploadInvalidRequestsRecordFailureWithoutEvaluatingQuota(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name          string
		request       func(*testing.T) *http.Request
		uploadService *showcaseUploadServiceSpy
		wantStatus    int
	}{
		{
			name: "request too large",
			request: func(_ *testing.T) *http.Request {
				request := httptest.NewRequest(http.MethodPost, "/showcase/upload", strings.NewReader(""))
				request.ContentLength = showcaseMaxRequestBytes + 1
				return request
			},
			wantStatus: http.StatusRequestEntityTooLarge,
		},
		{
			name: "malformed multipart form",
			request: func(_ *testing.T) *http.Request {
				request := httptest.NewRequest(http.MethodPost, "/showcase/upload", strings.NewReader("not multipart"))
				request.Header.Set("Content-Type", "multipart/form-data; boundary=missing")
				return request
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "missing files",
			request: func(t *testing.T) *http.Request {
				return newShowcaseUploadMultipartRequest(t, nil, "", map[string]string{
					"region":   "US",
					"order_id": "42",
				})
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "invalid file",
			request: func(t *testing.T) *http.Request {
				return newShowcaseUploadMultipartRequest(t, []byte("not a webp"), "photo.webp", map[string]string{
					"region":   "US",
					"order_id": "42",
				})
			},
			wantStatus: http.StatusUnsupportedMediaType,
		},
		{
			name: "missing region",
			request: func(t *testing.T) *http.Request {
				return newShowcaseUploadMultipartRequest(t, showcaseUploadValidWebPFixture(t), "photo.webp", map[string]string{
					"order_id": "42",
				})
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "invalid order id",
			request: func(t *testing.T) *http.Request {
				return newShowcaseUploadMultipartRequest(t, showcaseUploadValidWebPFixture(t), "photo.webp", map[string]string{
					"region":   "US",
					"order_id": "not-a-number",
				})
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "ineligible order",
			request: func(t *testing.T) *http.Request {
				return newShowcaseUploadMultipartRequest(t, showcaseUploadValidWebPFixture(t), "photo.webp", map[string]string{
					"region":   "US",
					"order_id": "42",
				})
			},
			uploadService: &showcaseUploadServiceSpy{
				validateOrderErr: service.ErrShowcaseUploadOrderNotEligible,
			},
			wantStatus: http.StatusForbidden,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			protection := &showcaseUploadProtectionSpy{}
			uploadService := test.uploadService
			if uploadService == nil {
				uploadService = &showcaseUploadServiceSpy{}
			}
			handler := &ShowcaseHandler{
				uploadService:    uploadService,
				uploadProtection: protection,
			}

			recorder := serveShowcaseUploadTestRequest(handler, test.request(t))
			if recorder.Code != test.wantStatus {
				t.Fatalf("status = %d, want %d; body = %s", recorder.Code, test.wantStatus, recorder.Body.String())
			}
			if protection.evaluateCalls != 0 {
				t.Fatalf("Evaluate() calls = %d, want 0", protection.evaluateCalls)
			}
			if protection.failureCalls != 1 {
				t.Fatalf("RecordFailure() calls = %d, want 1", protection.failureCalls)
			}
			if uploadService.countPendingCalls != 0 {
				t.Fatalf("CountPendingSubmissions() calls = %d, want 0", uploadService.countPendingCalls)
			}
		})
	}
}

func TestShowcaseUploadValidRequestEvaluatesQuotaAfterValidation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	events := make([]string, 0, 4)
	uploadService := &showcaseUploadServiceSpy{
		events: &events,
		item: &showcasedomain.Showcase{
			ID:     99,
			Status: showcasedomain.StatusPending,
		},
	}
	protection := &showcaseUploadProtectionSpy{
		events: &events,
	}
	handler := &ShowcaseHandler{
		uploadService:    uploadService,
		uploadProtection: protection,
	}
	request := newShowcaseUploadMultipartRequest(t, showcaseUploadValidWebPFixture(t), "photo.webp", map[string]string{
		"region":   "US",
		"location": "Seattle",
		"order_id": "42",
	})

	recorder := serveShowcaseUploadTestRequest(handler, request)
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body = %s", recorder.Code, http.StatusOK, recorder.Body.String())
	}
	if protection.evaluateCalls != 1 {
		t.Fatalf("Evaluate() calls = %d, want 1", protection.evaluateCalls)
	}
	if protection.failureCalls != 0 {
		t.Fatalf("RecordFailure() calls = %d, want 0", protection.failureCalls)
	}
	if uploadService.countPendingCalls != 1 {
		t.Fatalf("CountPendingSubmissions() calls = %d, want 1", uploadService.countPendingCalls)
	}
	if got, want := strings.Join(events, ","), "validate_order,count_pending,evaluate,upload"; got != want {
		t.Fatalf("call order = %q, want %q", got, want)
	}
	if protection.lastInput.PendingSubmissions != 0 {
		t.Fatalf("Evaluate() pending submissions = %d, want 0", protection.lastInput.PendingSubmissions)
	}
	if protection.lastInput.UploadBytes != request.ContentLength {
		t.Fatalf("Evaluate() upload bytes = %d, want request content length %d", protection.lastInput.UploadBytes, request.ContentLength)
	}
}

type showcaseUploadProtectionSpy struct {
	evaluateCalls int
	failureCalls  int
	lastInput     service.ShowcaseUploadProtectionInput
	events        *[]string
}

func (s *showcaseUploadProtectionSpy) Evaluate(_ context.Context, input service.ShowcaseUploadProtectionInput) (service.ShowcaseUploadProtectionDecision, error) {
	s.evaluateCalls++
	s.lastInput = input
	if s.events != nil {
		*s.events = append(*s.events, "evaluate")
	}
	return service.ShowcaseUploadProtectionDecision{
		Allowed: true,
		Action:  service.ShowcaseUploadProtectionDecisionAllow,
	}, nil
}

func (s *showcaseUploadProtectionSpy) RecordFailure(_ context.Context, _ service.ShowcaseUploadProtectionIdentity) error {
	s.failureCalls++
	return nil
}

type showcaseUploadServiceSpy struct {
	countPendingCalls int
	validateCalls     int
	uploadCalls       int
	validateOrderErr  error
	countPendingErr   error
	uploadErr         error
	item              *showcasedomain.Showcase
	events            *[]string
}

func (s *showcaseUploadServiceSpy) CountPendingSubmissions(_ uint) (int64, error) {
	s.countPendingCalls++
	if s.events != nil {
		*s.events = append(*s.events, "count_pending")
	}
	return 0, s.countPendingErr
}

func (s *showcaseUploadServiceSpy) ValidateUploadOrder(_ context.Context, _, _ uint) error {
	s.validateCalls++
	if s.events != nil {
		*s.events = append(*s.events, "validate_order")
	}
	return s.validateOrderErr
}

func (s *showcaseUploadServiceSpy) UploadPhotos(_ context.Context, _ uint, _ uint, _ []*multipart.FileHeader, _ map[string]string) (*showcasedomain.Showcase, error) {
	s.uploadCalls++
	if s.events != nil {
		*s.events = append(*s.events, "upload")
	}
	if s.uploadErr != nil {
		return nil, s.uploadErr
	}
	return s.item, nil
}

func serveShowcaseUploadTestRequest(handler *ShowcaseHandler, request *http.Request) *httptest.ResponseRecorder {
	router := gin.New()
	router.POST("/showcase/upload", func(c *gin.Context) {
		c.Set("user_id", uint(42))
		handler.Upload(c)
	})
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)
	return recorder
}

func newShowcaseUploadMultipartRequest(t *testing.T, fileContents []byte, filename string, fields map[string]string) *http.Request {
	t.Helper()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	for key, value := range fields {
		if err := writer.WriteField(key, value); err != nil {
			t.Fatalf("write multipart field %q: %v", key, err)
		}
	}
	if fileContents != nil {
		part, err := writer.CreateFormFile("file[]", filename)
		if err != nil {
			t.Fatalf("create multipart file: %v", err)
		}
		if _, err := part.Write(fileContents); err != nil {
			t.Fatalf("write multipart file: %v", err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close multipart writer: %v", err)
	}

	request := httptest.NewRequest(http.MethodPost, "/showcase/upload", body)
	request.Header.Set("Content-Type", writer.FormDataContentType())
	return request
}

func showcaseUploadValidWebPFixture(t *testing.T) []byte {
	t.Helper()

	const encoded = "UklGRrIBAABXRUJQVlA4TKUBAAAvSsAYAA8w//M///MfeJAkbXvaSG7m8Q3GfYSBJekwQztm/IcZlgwnmWImn2BK7aFmBtnVir6q//8VOkFE/xm4baTIu8c48ArEo6+B3zFKYln3pqClSCKX0begFTAXFOLXHSyF8cCNcZEG4OywuA4KVVfJCiArU7GAgJI8+lJP/OKMT/fBAjevg1cYB7YVkFuWga2lyPi5I0HFy5YTpWIHg0RZpkniRVW9odHAKOwosWuOGdxIyn2OvaCDvhg/we6TwadPBPbqBV58MsLmMJ8yZnOWk8SRz4N+QoyPL+MnamzMvcE1rHNEr91F9GKZPVUcS9w7PhhH36suB9qPeYb/oLk6cuTiJ0wOK3m5h1cKjW6EVZCYMK7dxcKCBdgP9HkKr9gkAO2P8GKZGWVdIAatQa+1IDpt6qyorVwdy01xdW8Jkfk6xjEXmVQQ+HQdFr6OKhIN34dXWq0+0qr6EJSCeeVLH9+gvGTLyqM65PQ44ihzlTXxQKjKbAvshXgir7Lil9w4L2bvMycmjQcqXaMCO6BlY28i+FOLzbfI1vEqxAhotocAAA=="
	data, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		t.Fatalf("decode WebP fixture: %v", err)
	}
	return data
}
