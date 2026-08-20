package admin

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"commerce-platform/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

const (
	fontPreflightManifestPath = "/_internal/font-preflight.json"
	fontPreflightMaxBodyBytes = 1 << 20
)

var requiredFontPreflightCheckKeys = []string{
	"no-external-system-fonts",
	"font-face-contract",
	"multilingual-split",
	"layout-parity",
	"subset-completeness",
}

var requiredFontPreflightFamilies = []string{
	"MapleUILatin",
	"MapleUICJK",
	"MapleUICoverageNotoSansDevanagari",
	"MapleUICoverageNotoSansLatinAccents",
	"MapleUICoverageNotoSansArabic",
	"MapleUICoverageNotoSansThai",
}

type FontPreflightHandler struct {
	storefrontBaseURL string
	httpClient        *http.Client
}

type FontPreflightReport struct {
	SchemaVersion int                   `json:"schema_version"`
	Project       string                `json:"project"`
	GeneratedAt   string                `json:"generated_at"`
	OverallStatus string                `json:"overall_status"`
	Baseline      FontPreflightBaseline `json:"baseline"`
	Checks        []FontPreflightCheck  `json:"checks"`
	Strategy      FontPreflightStrategy `json:"strategy"`
	Faces         []FontPreflightFace   `json:"faces"`
	Coverage      FontPreflightCoverage `json:"coverage"`
}

type FontPreflightBaseline struct {
	ID          string   `json:"id"`
	Label       string   `json:"label"`
	FontDisplay string   `json:"font_display"`
	Rules       []string `json:"rules"`
}

type FontPreflightCheck struct {
	Key     string   `json:"key"`
	Label   string   `json:"label"`
	Status  string   `json:"status"`
	Message string   `json:"message"`
	Details []string `json:"details"`
}

type FontPreflightStrategy struct {
	Status               string   `json:"status"`
	Label                string   `json:"label"`
	DefaultStack         []string `json:"default_stack"`
	LatinBytes           int64    `json:"latin_bytes"`
	LatinBudgetBytes     int64    `json:"latin_budget_bytes"`
	MapleUICJKFamily     string   `json:"maple_ui_cjk_family"`
	CoverageSourceFaces  []string `json:"coverage_source_faces"`
	CJKUnicodeRange      string   `json:"cjk_unicode_range"`
	LayoutParityVerified bool     `json:"layout_parity_verified"`
	Rationale            string   `json:"rationale"`
}

type FontPreflightFace struct {
	Family       string `json:"family"`
	Role         string `json:"role"`
	Script       string `json:"script"`
	SourceFace   string `json:"source_face"`
	Filename     string `json:"filename"`
	Bytes        int64  `json:"bytes"`
	FontDisplay  string `json:"font_display"`
	UnicodeRange string `json:"unicode_range"`
	SelfHosted   bool   `json:"self_hosted"`
}

type FontPreflightCoverage struct {
	LocaleCount       int                           `json:"locale_count"`
	SourceFileCount   int                           `json:"source_file_count"`
	CheckedCharacters int                           `json:"checked_characters"`
	MissingCharacters int                           `json:"missing_characters"`
	Locales           []FontPreflightLocaleCoverage `json:"locales"`
}

type FontPreflightLocaleCoverage struct {
	Locale            string   `json:"locale"`
	SourceFiles       int      `json:"source_files"`
	CheckedCharacters int      `json:"checked_characters"`
	MissingCharacters int      `json:"missing_characters"`
	MissingSample     []string `json:"missing_sample"`
	FontStack         []string `json:"font_stack"`
	Status            string   `json:"status"`
}

func NewFontPreflightHandler(storefrontBaseURL string, httpClient *http.Client) *FontPreflightHandler {
	if httpClient == nil {
		httpClient = &http.Client{
			Timeout: 5 * time.Second,
		}
	}
	client := *httpClient
	client.CheckRedirect = func(_ *http.Request, _ []*http.Request) error {
		return http.ErrUseLastResponse
	}

	return &FontPreflightHandler{
		storefrontBaseURL: storefrontBaseURL,
		httpClient:        &client,
	}
}

func (h *FontPreflightHandler) Get(c *gin.Context) {
	c.Header("Cache-Control", "no-store, max-age=0")

	report, err := h.Fetch(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "deployed storefront font preflight manifest is unavailable",
		})
		return
	}

	response.Success(c, report)
}

func (h *FontPreflightHandler) Fetch(ctx context.Context) (*FontPreflightReport, error) {
	if h == nil || h.httpClient == nil {
		return nil, errors.New("font preflight client is not configured")
	}

	manifestURL, err := fontPreflightURL(h.storefrontBaseURL)
	if err != nil {
		return nil, err
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, manifestURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create font preflight request: %w", err)
	}
	request.Header.Set("Accept", "application/json")

	httpResponse, err := h.httpClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("fetch deployed font preflight manifest: %w", err)
	}
	defer httpResponse.Body.Close()

	if httpResponse.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("deployed font preflight manifest returned HTTP %d", httpResponse.StatusCode)
	}

	if httpResponse.ContentLength > fontPreflightMaxBodyBytes {
		return nil, errors.New("deployed font preflight manifest exceeds the response size limit")
	}

	body, err := io.ReadAll(io.LimitReader(httpResponse.Body, fontPreflightMaxBodyBytes+1))
	if err != nil {
		return nil, fmt.Errorf("read deployed font preflight manifest: %w", err)
	}
	if int64(len(body)) > fontPreflightMaxBodyBytes {
		return nil, errors.New("deployed font preflight manifest exceeds the response size limit")
	}

	var report FontPreflightReport
	if err := json.Unmarshal(body, &report); err != nil {
		return nil, fmt.Errorf("decode deployed font preflight manifest: %w", err)
	}
	if err := validateFontPreflightReport(&report); err != nil {
		return nil, err
	}

	return &report, nil
}

func fontPreflightURL(storefrontBaseURL string) (string, error) {
	baseURL, err := url.Parse(strings.TrimSpace(storefrontBaseURL))
	if err != nil || baseURL.Scheme == "" || baseURL.Host == "" {
		return "", errors.New("storefront base URL is not configured")
	}
	if baseURL.Scheme != "http" && baseURL.Scheme != "https" {
		return "", errors.New("storefront base URL must use HTTP or HTTPS")
	}
	if baseURL.User != nil {
		return "", errors.New("storefront base URL must not include credentials")
	}

	return baseURL.ResolveReference(&url.URL{Path: fontPreflightManifestPath}).String(), nil
}

func validateFontPreflightReport(report *FontPreflightReport) error {
	if report == nil {
		return errors.New("font preflight manifest is empty")
	}
	if report.SchemaVersion != 1 {
		return fmt.Errorf("unsupported font preflight manifest schema version: %d", report.SchemaVersion)
	}
	if strings.TrimSpace(report.Project) == "" || strings.TrimSpace(report.GeneratedAt) == "" {
		return errors.New("font preflight manifest is missing identity metadata")
	}
	if !isFontPreflightStatus(report.OverallStatus) {
		return fmt.Errorf("invalid font preflight overall status: %q", report.OverallStatus)
	}
	if report.Baseline.ID != "storefront-built-in-font-shards-v1" || report.Baseline.FontDisplay != "block" {
		return errors.New("font preflight manifest does not declare the required built-in block-loading font baseline")
	}
	if len(report.Checks) == 0 || len(report.Faces) == 0 || len(report.Coverage.Locales) == 0 {
		return errors.New("font preflight manifest is incomplete")
	}
	if report.OverallStatus == "pass" && !report.Strategy.LayoutParityVerified {
		return errors.New("passing font preflight manifest must verify layout parity")
	}
	if report.Strategy.Status != "" && !isFontPreflightStatus(report.Strategy.Status) {
		return fmt.Errorf("invalid font preflight strategy status: %q", report.Strategy.Status)
	}
	if len(report.Strategy.DefaultStack) != 2 || report.Strategy.DefaultStack[0] != "MapleUILatin" || report.Strategy.DefaultStack[1] != "MapleUICJK" {
		return errors.New("font preflight manifest does not declare the required Maple UI default font stack")
	}
	if report.Strategy.MapleUICJKFamily != "MapleUICJK" {
		return errors.New("font preflight manifest does not declare the required Maple UI CJK family")
	}
	if report.Strategy.LatinBudgetBytes <= 0 || report.Strategy.LatinBytes < 0 || report.Strategy.LatinBytes > report.Strategy.LatinBudgetBytes {
		return errors.New("font preflight manifest contains an invalid Latin subset budget")
	}

	checksByKey := make(map[string]FontPreflightCheck, len(report.Checks))
	hasBlockingCheck := false
	for _, check := range report.Checks {
		if strings.TrimSpace(check.Key) == "" || strings.TrimSpace(check.Label) == "" || !isFontPreflightStatus(check.Status) {
			return errors.New("font preflight manifest contains an invalid check")
		}
		if _, exists := checksByKey[check.Key]; exists {
			return fmt.Errorf("font preflight manifest contains duplicate check: %s", check.Key)
		}
		checksByKey[check.Key] = check
		if check.Status == "block" {
			hasBlockingCheck = true
		}
	}
	for _, key := range requiredFontPreflightCheckKeys {
		if _, exists := checksByKey[key]; !exists {
			return fmt.Errorf("font preflight manifest is missing required check: %s", key)
		}
	}
	if hasBlockingCheck != (report.OverallStatus == "block") {
		return errors.New("font preflight manifest overall status does not match blocking checks")
	}

	facesByFamily := make(map[string]FontPreflightFace, len(report.Faces))
	for _, face := range report.Faces {
		if strings.TrimSpace(face.Family) == "" || strings.TrimSpace(face.Filename) == "" || strings.TrimSpace(face.SourceFace) == "" || face.FontDisplay != "block" || !face.SelfHosted {
			return errors.New("font preflight manifest contains an invalid built-in font face")
		}
		if _, exists := facesByFamily[face.Family]; exists {
			return fmt.Errorf("font preflight manifest contains duplicate font face: %s", face.Family)
		}
		facesByFamily[face.Family] = face
	}
	for _, family := range requiredFontPreflightFamilies {
		if _, exists := facesByFamily[family]; !exists {
			return fmt.Errorf("font preflight manifest is missing required font face: %s", family)
		}
	}
	if report.Coverage.LocaleCount != len(report.Coverage.Locales) || report.Coverage.MissingCharacters < 0 {
		return errors.New("font preflight manifest contains invalid locale coverage totals")
	}
	for _, locale := range report.Coverage.Locales {
		if strings.TrimSpace(locale.Locale) == "" || !isFontPreflightStatus(locale.Status) || locale.MissingCharacters < 0 || len(locale.FontStack) == 0 {
			return errors.New("font preflight manifest contains invalid locale coverage")
		}
	}

	return nil
}

func isFontPreflightStatus(value string) bool {
	return value == "pass" || value == "warning" || value == "block"
}
