package admin

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestFontPreflightHandlerReadsConfiguredStorefrontManifest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	storefront := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		require.Equal(t, http.MethodGet, request.Method)
		require.Equal(t, fontPreflightManifestPath, request.URL.Path)
		writer.Header().Set("Content-Type", "application/json")
		_, _ = writer.Write([]byte(validFontPreflightManifest))
	}))
	t.Cleanup(storefront.Close)

	handler := NewFontPreflightHandler(storefront.URL, storefront.Client())
	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Request = httptest.NewRequest(http.MethodGet, "/api/admin/preflight/fonts", nil)

	handler.Get(context)

	require.Equal(t, http.StatusOK, recorder.Code)
	var body struct {
		Code int                 `json:"code"`
		Data FontPreflightReport `json:"data"`
	}
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &body))
	require.Equal(t, 0, body.Code)
	require.Equal(t, "pass", body.Data.OverallStatus)
	require.Equal(t, "StorefrontSystemLatin", body.Data.Strategy.DefaultStack[0])
	require.Equal(t, "no-store, max-age=0", recorder.Header().Get("Cache-Control"))
}

func TestFontPreflightHandlerRejectsInvalidManifest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	storefront := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		writer.Header().Set("Content-Type", "application/json")
		_, _ = writer.Write([]byte(`{"schema_version":1}`))
	}))
	t.Cleanup(storefront.Close)

	handler := NewFontPreflightHandler(storefront.URL, storefront.Client())
	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Request = httptest.NewRequest(http.MethodGet, "/api/admin/preflight/fonts", nil)

	handler.Get(context)

	require.Equal(t, http.StatusServiceUnavailable, recorder.Code)
	require.Equal(t, "no-store, max-age=0", recorder.Header().Get("Cache-Control"))
}

func TestFontPreflightURLOnlyUsesConfiguredStorefrontOrigin(t *testing.T) {
	resolved, err := fontPreflightURL("https://storefront.example.test/a/path?ignored=1")

	require.NoError(t, err)
	require.Equal(t, "https://storefront.example.test/_internal/font-preflight.json", resolved)

	_, err = fontPreflightURL("https://token@storefront.example.test")
	require.Error(t, err)
}

func TestFontPreflightHandlerDoesNotFollowStorefrontRedirects(t *testing.T) {
	gin.SetMode(gin.TestMode)
	var redirectedRequests int
	redirectTarget := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		redirectedRequests++
	}))
	t.Cleanup(redirectTarget.Close)
	storefront := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		http.Redirect(writer, request, redirectTarget.URL, http.StatusFound)
	}))
	t.Cleanup(storefront.Close)

	handler := NewFontPreflightHandler(storefront.URL, storefront.Client())
	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Request = httptest.NewRequest(http.MethodGet, "/api/admin/preflight/fonts", nil)

	handler.Get(context)

	require.Equal(t, http.StatusServiceUnavailable, recorder.Code)
	require.Equal(t, 0, redirectedRequests)
}

func TestFontPreflightHandlerRejectsOversizedAndBaselineBreakingManifest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	testCases := []struct {
		name string
		body string
	}{
		{
			name: "oversized",
			body: strings.Repeat(" ", fontPreflightMaxBodyBytes+1),
		},
		{
			name: "wrong_font_display",
			body: strings.Replace(validFontPreflightManifest, `"font_display": "swap"`, `"font_display": "block"`, 1),
		},
		{
			name: "inconsistent_overall_status",
			body: strings.Replace(validFontPreflightManifest, `"status": "pass"`, `"status": "block"`, 1),
		},
		{
			name: "missing_required_gate",
			body: strings.Replace(validFontPreflightManifest, `"key": "subset-completeness"`, `"key": "unknown"`, 1),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			storefront := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
				writer.Header().Set("Content-Type", "application/json")
				_, _ = writer.Write([]byte(testCase.body))
			}))
			t.Cleanup(storefront.Close)

			handler := NewFontPreflightHandler(storefront.URL, storefront.Client())
			recorder := httptest.NewRecorder()
			context, _ := gin.CreateTestContext(recorder)
			context.Request = httptest.NewRequest(http.MethodGet, "/api/admin/preflight/fonts", nil)

			handler.Get(context)

			require.Equal(t, http.StatusServiceUnavailable, recorder.Code)
		})
	}
}

const validFontPreflightManifest = `{
  "schema_version": 1,
  "project": "tanzanite-theme storefront",
  "generated_at": "2026-08-17T00:00:00.000Z",
  "overall_status": "pass",
  "baseline": {
    "id": "storefront-self-hosted-no-fallback-v1",
    "label": "self-hosted only",
    "font_display": "swap",
    "rules": ["no external fallback"]
  },
  "checks": [
    {
      "key": "no-external-fallback",
      "label": "no fallback",
      "status": "pass",
      "message": "passed",
      "details": []
    },
    {
      "key": "font-face-contract",
      "label": "font contract",
      "status": "pass",
      "message": "passed",
      "details": []
    },
    {
      "key": "multilingual-split",
      "label": "language split",
      "status": "pass",
      "message": "passed",
      "details": []
    },
    {
      "key": "layout-parity",
      "label": "layout parity",
      "status": "pass",
      "message": "passed",
      "details": []
    },
    {
      "key": "subset-completeness",
      "label": "subset completeness",
      "status": "pass",
      "message": "passed",
      "details": []
    }
  ],
  "strategy": {
    "status": "pass",
    "label": "Latin subset",
    "default_stack": ["StorefrontSystemLatin", "StorefrontSystem"],
    "latin_bytes": 100,
    "latin_budget_bytes": 163840,
    "complete_maple_ui_family": "StorefrontSystem",
    "cjk_unicode_range": "U+4E00-9FFF",
    "layout_parity_verified": true,
    "rationale": "same metrics"
  },
  "faces": [
    {
      "family": "StorefrontSystemLatin",
      "role": "Latin",
      "script": "Latin",
      "filename": "StorefrontSystem-Latin.00af3fec5b34.woff2",
      "bytes": 100,
      "font_display": "swap",
      "unicode_range": "",
      "self_hosted": true
    },
    {
      "family": "StorefrontSystem",
      "role": "CJK",
      "script": "CJK",
      "filename": "StorefrontSystem-CJK.f8ce6d72e8cb.woff2",
      "bytes": 100,
      "font_display": "swap",
      "unicode_range": "U+4E00-9FFF",
      "self_hosted": true
    },
    {
      "family": "StorefrontSystemDevanagari",
      "role": "Devanagari",
      "script": "Devanagari",
      "filename": "StorefrontSystem-Devanagari.3b3cae4d2600.woff2",
      "bytes": 100,
      "font_display": "swap",
      "unicode_range": "U+0900-097F",
      "self_hosted": true
    },
    {
      "family": "StorefrontSystemLatinAccents",
      "role": "Latin accents",
      "script": "Latin accents",
      "filename": "StorefrontSystem-Latin-Accents.e645edc952b6.woff2",
      "bytes": 100,
      "font_display": "swap",
      "unicode_range": "U+00C0-00C1",
      "self_hosted": true
    },
    {
      "family": "StorefrontSystemArabic",
      "role": "Arabic",
      "script": "Arabic",
      "filename": "StorefrontSystem-Arabic.ce85091f0209.woff2",
      "bytes": 100,
      "font_display": "swap",
      "unicode_range": "U+0600-06FF",
      "self_hosted": true
    },
    {
      "family": "StorefrontSystemThai",
      "role": "Thai",
      "script": "Thai",
      "filename": "StorefrontSystem-Thai.1f5a173641bb.woff2",
      "bytes": 100,
      "font_display": "swap",
      "unicode_range": "U+0E01-0E5B",
      "self_hosted": true
    }
  ],
  "coverage": {
    "locale_count": 1,
    "source_file_count": 1,
    "checked_characters": 10,
    "missing_characters": 0,
    "locales": [{
      "locale": "en",
      "source_files": 1,
      "checked_characters": 10,
      "missing_characters": 0,
      "missing_sample": [],
      "font_stack": ["StorefrontSystemLatin", "StorefrontSystem"],
      "status": "pass"
    }]
  }
}`
