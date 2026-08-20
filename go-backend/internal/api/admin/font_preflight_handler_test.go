package admin

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"commerce-platform/internal/pkg/config"

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
	require.Equal(t, "MapleUILatin", body.Data.Strategy.DefaultStack[0])
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

func TestFontPreflightStorefrontOriginPrefersInternalOrigin(t *testing.T) {
	t.Setenv("STOREFRONT_INTERNAL_ORIGIN", "http://frontend:3000/")
	t.Setenv("STOREFRONT_BASE_URL", "http://localhost:9199")

	origin := fontPreflightStorefrontOrigin(&config.Config{
		Server: config.ServerConfig{
			Mode:    "debug",
			BaseURL: "http://localhost:9200",
		},
	})

	require.Equal(t, "http://frontend:3000", origin)
}

func TestFontPreflightStorefrontOriginUsesDevHostFallback(t *testing.T) {
	t.Setenv("STOREFRONT_INTERNAL_ORIGIN", "")
	t.Setenv("STOREFRONT_BASE_URL", "")

	origin := fontPreflightStorefrontOrigin(&config.Config{
		Server: config.ServerConfig{
			Mode:    "debug",
			BaseURL: "http://localhost:9200",
		},
	})

	require.Equal(t, "http://localhost:9199", origin)
}

func TestFontPreflightStorefrontOriginDoesNotUsePublicServerBaseInProduction(t *testing.T) {
	t.Setenv("STOREFRONT_INTERNAL_ORIGIN", "")
	t.Setenv("STOREFRONT_BASE_URL", "https://learn.gripe")

	origin := fontPreflightStorefrontOrigin(&config.Config{
		Server: config.ServerConfig{
			Mode:    "release",
			BaseURL: "https://learn.gripe",
		},
	})

	require.Empty(t, origin)
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
			body: strings.Replace(validFontPreflightManifest, `"font_display": "block"`, `"font_display": "swap"`, 1),
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
    "id": "storefront-built-in-font-shards-v1",
    "label": "built-in font shards only",
    "font_display": "block",
    "rules": ["no OS/system or generic font fallback"]
  },
  "checks": [
    {
      "key": "no-external-system-fonts",
      "label": "no system fonts",
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
    "default_stack": ["MapleUILatin", "MapleUICJK"],
    "latin_bytes": 100,
    "latin_budget_bytes": 163840,
    "maple_ui_cjk_family": "MapleUICJK",
    "coverage_source_faces": ["Noto Sans Devanagari", "Noto Sans", "Noto Sans Arabic", "Noto Sans Thai"],
    "cjk_unicode_range": "U+4E00-9FFF",
    "layout_parity_verified": true,
    "rationale": "same metrics"
  },
  "faces": [
    {
      "family": "MapleUILatin",
      "role": "Latin",
      "script": "Latin",
      "source_face": "Maple UI",
      "filename": "MapleUI-Latin.00af3fec5b34.woff2",
      "bytes": 100,
      "font_display": "block",
      "unicode_range": "",
      "self_hosted": true
    },
    {
      "family": "MapleUICJK",
      "role": "CJK",
      "script": "CJK",
      "source_face": "Maple UI",
      "filename": "MapleUI-CJK.f8ce6d72e8cb.woff2",
      "bytes": 100,
      "font_display": "block",
      "unicode_range": "U+4E00-9FFF",
      "self_hosted": true
    },
    {
      "family": "MapleUICoverageNotoSansDevanagari",
      "role": "Devanagari",
      "script": "Devanagari",
      "source_face": "Noto Sans Devanagari",
      "filename": "MapleUI-Coverage-NotoSans-Devanagari.3b3cae4d2600.woff2",
      "bytes": 100,
      "font_display": "block",
      "unicode_range": "U+0900-097F",
      "self_hosted": true
    },
    {
      "family": "MapleUICoverageNotoSansLatinAccents",
      "role": "Latin accents",
      "script": "Latin accents",
      "source_face": "Noto Sans",
      "filename": "MapleUI-Coverage-NotoSans-Latin-Accents.e645edc952b6.woff2",
      "bytes": 100,
      "font_display": "block",
      "unicode_range": "U+00C0-00C1",
      "self_hosted": true
    },
    {
      "family": "MapleUICoverageNotoSansArabic",
      "role": "Arabic",
      "script": "Arabic",
      "source_face": "Noto Sans Arabic",
      "filename": "MapleUI-Coverage-NotoSans-Arabic.ce85091f0209.woff2",
      "bytes": 100,
      "font_display": "block",
      "unicode_range": "U+0600-06FF",
      "self_hosted": true
    },
    {
      "family": "MapleUICoverageNotoSansThai",
      "role": "Thai",
      "script": "Thai",
      "source_face": "Noto Sans Thai",
      "filename": "MapleUI-Coverage-NotoSans-Thai.1f5a173641bb.woff2",
      "bytes": 100,
      "font_display": "block",
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
      "font_stack": ["MapleUILatin", "MapleUICJK"],
      "status": "pass"
    }]
  }
}`
