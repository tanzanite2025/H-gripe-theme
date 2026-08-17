package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	stdhtml "html"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"
	"unicode"

	preflightdomain "commerce-platform/internal/domain/preflight"
	seodomain "commerce-platform/internal/domain/seo"
	"commerce-platform/internal/repository"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
	"gorm.io/gorm"
)

const (
	contentLinkTargetSelectionPageSize = 200
	defaultContentLinkMaxHTMLBytes     = 5 * 1024 * 1024
)

var (
	ErrContentLinkPreflightUnavailable = errors.New("content link preflight service is unavailable")
	ErrContentLinkIssueNotFixable      = errors.New("content link issue is not directly fixable")
	ErrContentLinkSourceStale          = errors.New("content link source no longer matches the latest evidence")
)

type PreflightContentLinkConfig struct {
	BaseURL      string
	HTTPClient   *http.Client
	MaxHTMLBytes int64
}

type ContentLinkTargetOption struct {
	URL        string `json:"url"`
	Path       string `json:"path"`
	Title      string `json:"title"`
	Locale     string `json:"locale"`
	SourceType string `json:"source_type"`
	IsHome     bool   `json:"is_home"`
}

type ContentLinkTargetOptions struct {
	DefaultURL string                    `json:"default_url"`
	Items      []ContentLinkTargetOption `json:"items"`
}

type ContentLinkRunInput struct {
	TargetURL   string
	ActorUserID uint
}

type ContentLinkRunResult struct {
	Run    preflightdomain.ContentLinkRun        `json:"run"`
	Issues []preflightdomain.ContentLinkIssue    `json:"issues"`
	Stats  preflightdomain.ContentLinkIssueStats `json:"stats"`
}

type PreflightContentLinkService struct {
	repository   *repository.PreflightContentLinkRepository
	routeCatalog *repository.StorefrontRouteCatalogRepository
	postService  *PostService
	baseURL      string
	httpClient   *http.Client
	maxHTMLBytes int64
}

type contentLinkAnchor struct {
	Href     string
	Text     string
	Selector string
	Snippet  string
	Heading  string
}

type contentLinkTextMatch struct {
	Normalized string
	Language   string
	Kind       string
}

type contentLinkSourceLocation struct {
	SourceType  string
	SourceID    *uint
	SourceKey   string
	SourceField string
	FixStatus   string
	SourceHint  string
}

type contentLinkEvidence struct {
	Rule           string `json:"rule"`
	Href           string `json:"href"`
	LinkURL        string `json:"link_url"`
	LinkText       string `json:"link_text"`
	NormalizedText string `json:"normalized_text"`
	Selector       string `json:"selector"`
	Snippet        string `json:"snippet"`
	ContextHeading string `json:"context_heading,omitempty"`
	SuggestedText  string `json:"suggested_text"`
	SourceType     string `json:"source_type"`
	SourceKey      string `json:"source_key"`
	SourceField    string `json:"source_field,omitempty"`
	SourceHint     string `json:"source_hint,omitempty"`
}

func NewPreflightContentLinkService(
	repo *repository.PreflightContentLinkRepository,
	routeCatalog *repository.StorefrontRouteCatalogRepository,
	postService *PostService,
	cfg PreflightContentLinkConfig,
) *PreflightContentLinkService {
	client := cfg.HTTPClient
	if client == nil {
		client = &http.Client{Timeout: 20 * time.Second}
	}
	maxBytes := cfg.MaxHTMLBytes
	if maxBytes <= 0 {
		maxBytes = defaultContentLinkMaxHTMLBytes
	}
	return &PreflightContentLinkService{
		repository:   repo,
		routeCatalog: routeCatalog,
		postService:  postService,
		baseURL:      strings.TrimRight(strings.TrimSpace(cfg.BaseURL), "/"),
		httpClient:   client,
		maxHTMLBytes: maxBytes,
	}
}

func (s *PreflightContentLinkService) ListTargetOptions() (*ContentLinkTargetOptions, error) {
	if s == nil {
		return nil, ErrContentLinkPreflightUnavailable
	}
	fallbackURL, err := s.defaultTargetURL()
	if err != nil {
		return nil, err
	}
	if s.routeCatalog == nil {
		return &ContentLinkTargetOptions{
			DefaultURL: fallbackURL,
			Items: []ContentLinkTargetOption{{
				URL:    fallbackURL,
				Path:   "/",
				IsHome: true,
			}},
		}, nil
	}

	entries := make([]seodomain.StorefrontRouteCatalogEntry, 0, contentLinkTargetSelectionPageSize)
	for page := 1; ; page++ {
		batch, total, err := s.routeCatalog.List(repository.StorefrontRouteCatalogListFilter{
			Page:         page,
			PageSize:     contentLinkTargetSelectionPageSize,
			EntryStatus:  seodomain.RouteEntryStatusActive,
			ExcludeAlias: true,
		})
		if err != nil {
			return nil, err
		}
		for _, entry := range batch {
			if entry.IsAlias || !entry.IsCheckable || entry.EntryStatus != seodomain.RouteEntryStatusActive {
				continue
			}
			entries = append(entries, entry)
		}
		if len(entries) >= int(total) || len(batch) == 0 {
			break
		}
	}

	options := make([]ContentLinkTargetOption, 0, len(entries)+1)
	seen := make(map[string]struct{}, len(entries)+1)
	defaultURL := fallbackURL
	homeFound := false
	for _, entry := range entries {
		targetURL, err := s.urlForRoute(entry)
		if err != nil {
			continue
		}
		if _, ok := seen[targetURL]; ok {
			continue
		}
		seen[targetURL] = struct{}{}
		option := ContentLinkTargetOption{
			URL:        targetURL,
			Path:       strings.TrimSpace(entry.Path),
			Title:      strings.TrimSpace(entry.Title),
			Locale:     strings.TrimSpace(entry.Locale),
			SourceType: strings.TrimSpace(entry.SourceType),
			IsHome:     routeEntryIsHome(entry),
		}
		if option.IsHome && !homeFound {
			defaultURL = targetURL
			homeFound = true
		}
		options = append(options, option)
	}
	if !homeFound {
		fallback := ContentLinkTargetOption{URL: fallbackURL, Path: "/", IsHome: true}
		if _, ok := seen[fallbackURL]; !ok {
			options = append([]ContentLinkTargetOption{fallback}, options...)
		}
		defaultURL = fallbackURL
	}
	sort.SliceStable(options, func(i, j int) bool {
		if options[i].IsHome != options[j].IsHome {
			return options[i].IsHome
		}
		if options[i].Path != options[j].Path {
			return options[i].Path < options[j].Path
		}
		return options[i].Title < options[j].Title
	})
	return &ContentLinkTargetOptions{DefaultURL: defaultURL, Items: options}, nil
}

func (s *PreflightContentLinkService) RunCheck(
	ctx context.Context,
	input ContentLinkRunInput,
) (*ContentLinkRunResult, error) {
	if s == nil || s.repository == nil {
		return nil, ErrContentLinkPreflightUnavailable
	}
	targetURL, err := s.normalizeTargetURL(input.TargetURL)
	if err != nil {
		return nil, err
	}
	routeEntry := s.routeEntryForURL(targetURL)
	routeEntryID := routeEntryIDFor(routeEntry)
	checkedAt := time.Now().UTC()
	run := &preflightdomain.ContentLinkRun{
		TargetURL:    targetURL,
		RouteEntryID: routeEntryID,
		Status:       preflightdomain.ContentLinkRunStatusSuccess,
		CheckedAt:    checkedAt,
	}

	htmlBody, finalURL, fetchErr := s.fetchHTML(ctx, targetURL)
	if fetchErr != nil {
		run.Status = preflightdomain.ContentLinkRunStatusFailed
		run.ErrorMessage = fetchErr.Error()
		if err := s.repository.CreateRun(run); err != nil {
			return nil, err
		}
		stats, _ := s.repository.Stats()
		return &ContentLinkRunResult{Run: *run, Stats: stats}, nil
	}
	if finalURL == "" {
		finalURL = targetURL
	}

	detections, err := s.detectContentLinks(htmlBody, targetURL, finalURL, routeEntry)
	if err != nil {
		run.Status = preflightdomain.ContentLinkRunStatusFailed
		run.ErrorMessage = err.Error()
		if err := s.repository.CreateRun(run); err != nil {
			return nil, err
		}
		stats, _ := s.repository.Stats()
		return &ContentLinkRunResult{Run: *run, Stats: stats}, nil
	}
	run.IssueCount = len(detections)
	for _, detection := range detections {
		if detection.FixStatus == preflightdomain.ContentLinkFixStatusPending {
			run.FixableCount++
		}
	}
	if err := s.repository.CreateRun(run); err != nil {
		return nil, err
	}
	for index := range detections {
		detections[index].RunID = run.ID
	}
	if err := s.repository.RecordDetections(run, detections); err != nil {
		return nil, err
	}
	issues, _, err := s.repository.ListIssues(repository.PreflightContentLinkIssueListFilter{
		Page:      1,
		PageSize:  200,
		State:     "active",
		TargetURL: targetURL,
	})
	if err != nil {
		return nil, err
	}
	stats, err := s.repository.Stats()
	if err != nil {
		return nil, err
	}
	return &ContentLinkRunResult{Run: *run, Issues: issues, Stats: stats}, nil
}

func (s *PreflightContentLinkService) ListIssues(
	filter repository.PreflightContentLinkIssueListFilter,
) ([]preflightdomain.ContentLinkIssue, int64, error) {
	if s == nil || s.repository == nil {
		return nil, 0, ErrContentLinkPreflightUnavailable
	}
	return s.repository.ListIssues(filter)
}

func (s *PreflightContentLinkService) Stats() (preflightdomain.ContentLinkIssueStats, error) {
	if s == nil || s.repository == nil {
		return preflightdomain.ContentLinkIssueStats{}, ErrContentLinkPreflightUnavailable
	}
	return s.repository.Stats()
}

func (s *PreflightContentLinkService) GetIssue(
	id uint,
) (*preflightdomain.ContentLinkIssue, error) {
	if s == nil || s.repository == nil {
		return nil, ErrContentLinkPreflightUnavailable
	}
	return s.repository.FindIssueByID(id)
}

func (s *PreflightContentLinkService) ListIssueEvents(
	issueID uint,
	page int,
	pageSize int,
) ([]preflightdomain.ContentLinkIssueEvent, int64, error) {
	if s == nil || s.repository == nil {
		return nil, 0, ErrContentLinkPreflightUnavailable
	}
	return s.repository.ListIssueEvents(issueID, page, pageSize)
}

func (s *PreflightContentLinkService) ApplySuggestion(
	ctx context.Context,
	issueID uint,
	actorUserID uint,
) (*preflightdomain.ContentLinkIssue, error) {
	if s == nil || s.repository == nil || s.postService == nil {
		return nil, ErrContentLinkPreflightUnavailable
	}
	issue, err := s.repository.FindIssueByID(issueID)
	if err != nil {
		return nil, err
	}
	if issue.SourceType != "blog_post" || issue.SourceID == nil || issue.SourceField != "content" {
		return nil, ErrContentLinkIssueNotFixable
	}
	if strings.TrimSpace(issue.SuggestedText) == "" {
		return nil, errors.New("content link suggestion is empty")
	}

	foundPost, err := s.postService.GetAdminPost(*issue.SourceID)
	if err != nil {
		return nil, err
	}
	nextContent, changed, err := replaceFirstContentLinkSuggestion(foundPost.Content, issue)
	if err != nil || !changed {
		markErr := ErrContentLinkSourceStale
		if err != nil {
			markErr = err
		}
		_, _ = s.repository.UpdateIssueWithEvent(issue.ID, map[string]interface{}{
			"fix_status": preflightdomain.ContentLinkFixStatusFailed,
			"fix_error":  markErr.Error(),
		}, preflightdomain.ContentLinkEventFixFailed, actorUserID, markErr.Error(), nil)
		return nil, markErr
	}

	if _, err := s.postService.UpdateAdminPost(foundPost.ID, PostUpdateInput{Content: &nextContent}); err != nil {
		_, _ = s.repository.UpdateIssueWithEvent(issue.ID, map[string]interface{}{
			"fix_status": preflightdomain.ContentLinkFixStatusFailed,
			"fix_error":  err.Error(),
		}, preflightdomain.ContentLinkEventFixFailed, actorUserID, err.Error(), map[string]interface{}{
			"post_id": foundPost.ID,
		})
		return nil, err
	}

	now := time.Now().UTC()
	return s.repository.UpdateIssueWithEvent(issue.ID, map[string]interface{}{
		"state":       preflightdomain.ContentLinkIssueStateResolved,
		"fix_status":  preflightdomain.ContentLinkFixStatusApplied,
		"fix_error":   "",
		"resolved_at": now,
		"fixed_at":    now,
	}, preflightdomain.ContentLinkEventFixApplied, actorUserID, "applied suggested link text", map[string]interface{}{
		"post_id":        foundPost.ID,
		"suggested_text": issue.SuggestedText,
	})
}

func (s *PreflightContentLinkService) ResolveIssue(
	issueID uint,
	actorUserID uint,
	note string,
) (*preflightdomain.ContentLinkIssue, error) {
	if s == nil || s.repository == nil {
		return nil, ErrContentLinkPreflightUnavailable
	}
	now := time.Now().UTC()
	return s.repository.UpdateIssueWithEvent(issueID, map[string]interface{}{
		"state":       preflightdomain.ContentLinkIssueStateResolved,
		"resolved_at": now,
	}, preflightdomain.ContentLinkEventResolutionRecorded, actorUserID, note, nil)
}

func (s *PreflightContentLinkService) RecheckIssue(
	ctx context.Context,
	issueID uint,
	actorUserID uint,
) (*ContentLinkRunResult, error) {
	if s == nil || s.repository == nil {
		return nil, ErrContentLinkPreflightUnavailable
	}
	issue, err := s.repository.FindIssueByID(issueID)
	if err != nil {
		return nil, err
	}
	_, _ = s.repository.UpdateIssueWithEvent(issue.ID, nil, preflightdomain.ContentLinkEventManualRecheck, actorUserID, "", nil)
	return s.RunCheck(ctx, ContentLinkRunInput{TargetURL: issue.TargetURL, ActorUserID: actorUserID})
}

func (s *PreflightContentLinkService) detectContentLinks(
	htmlBody string,
	targetURL string,
	finalURL string,
	routeEntry *seodomain.StorefrontRouteCatalogEntry,
) ([]preflightdomain.ContentLinkDetection, error) {
	doc, err := html.Parse(strings.NewReader(htmlBody))
	if err != nil {
		return nil, fmt.Errorf("parse storefront HTML: %w", err)
	}
	anchors := collectContentLinkAnchors(doc)
	detections := make([]preflightdomain.ContentLinkDetection, 0)
	for _, anchor := range anchors {
		if !contentLinkHrefInspectable(anchor.Href) {
			continue
		}
		match, ok := matchGenericContentLinkText(anchor.Text)
		if !ok {
			continue
		}
		linkURL := resolvedLinkURL(anchor.Href, finalURL, targetURL)
		if linkURL == "" {
			linkURL = strings.TrimSpace(anchor.Href)
		}
		suggestedText := s.suggestContentLinkText(match, anchor, linkURL)
		source := s.locateContentLinkSource(routeEntry, anchor, linkURL, targetURL)
		evidence := contentLinkEvidence{
			Rule:           "descriptive_link_text",
			Href:           strings.TrimSpace(anchor.Href),
			LinkURL:        linkURL,
			LinkText:       strings.TrimSpace(anchor.Text),
			NormalizedText: match.Normalized,
			Selector:       anchor.Selector,
			Snippet:        anchor.Snippet,
			ContextHeading: strings.TrimSpace(anchor.Heading),
			SuggestedText:  suggestedText,
			SourceType:     source.SourceType,
			SourceKey:      source.SourceKey,
			SourceField:    source.SourceField,
			SourceHint:     source.SourceHint,
		}
		detection := preflightdomain.ContentLinkDetection{
			RouteEntryID:    routeEntryIDFor(routeEntry),
			TargetURL:       targetURL,
			FinalURL:        finalURL,
			LinkURL:         linkURL,
			LinkText:        strings.TrimSpace(anchor.Text),
			Selector:        anchor.Selector,
			Snippet:         anchor.Snippet,
			SourceType:      source.SourceType,
			SourceID:        source.SourceID,
			SourceKey:       source.SourceKey,
			SourceField:     source.SourceField,
			IssueKey:        contentLinkIssueKey(targetURL, linkURL, match.Normalized, anchor.Selector),
			Severity:        "medium",
			SuggestedText:   suggestedText,
			FixStatus:       source.FixStatus,
			LatestEvidence:  encodeContentLinkEvidence(evidence),
			FirstDetectedAt: time.Now().UTC(),
			LastDetectedAt:  time.Now().UTC(),
		}
		detections = append(detections, detection)
	}
	return detections, nil
}

func (s *PreflightContentLinkService) fetchHTML(ctx context.Context, targetURL string) (string, string, error) {
	if s == nil || s.httpClient == nil {
		return "", "", ErrContentLinkPreflightUnavailable
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, targetURL, nil)
	if err != nil {
		return "", "", err
	}
	req.Header.Set("Accept", "text/html,application/xhtml+xml")
	req.Header.Set("User-Agent", "Tanzanite-Preflight-ContentLinks/1.0")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	finalURL := targetURL
	if resp.Request != nil && resp.Request.URL != nil {
		finalURL = resp.Request.URL.String()
	}
	contentType := strings.ToLower(resp.Header.Get("Content-Type"))
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", finalURL, fmt.Errorf("storefront returned HTTP %d", resp.StatusCode)
	}
	if contentType != "" && !strings.Contains(contentType, "text/html") && !strings.Contains(contentType, "application/xhtml") {
		return "", finalURL, fmt.Errorf("storefront response is not HTML: %s", contentType)
	}

	limited := io.LimitReader(resp.Body, s.maxHTMLBytes+1)
	body, err := io.ReadAll(limited)
	if err != nil {
		return "", finalURL, err
	}
	if int64(len(body)) > s.maxHTMLBytes {
		return "", finalURL, fmt.Errorf("storefront HTML exceeds %d bytes", s.maxHTMLBytes)
	}
	return string(body), finalURL, nil
}

func (s *PreflightContentLinkService) normalizeTargetURL(value string) (string, error) {
	defaultURL, err := s.defaultTargetURL()
	if err != nil {
		return "", err
	}
	base, err := url.Parse(defaultURL)
	if err != nil || base.Scheme == "" || base.Host == "" {
		return "", errors.New("content link preflight base URL is invalid")
	}
	raw := strings.TrimSpace(value)
	if raw == "" {
		raw = defaultURL
	}
	parsed, err := url.Parse(raw)
	if err != nil {
		return "", fmt.Errorf("invalid target URL: %w", err)
	}
	if parsed.Scheme == "" && parsed.Host == "" {
		parsed = base.ResolveReference(parsed)
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return "", errors.New("content link target must use http or https")
	}
	if !strings.EqualFold(parsed.Host, base.Host) {
		return "", errors.New("content link target must be same-origin storefront URL")
	}
	parsed.Fragment = ""
	if parsed.Path == "" {
		parsed.Path = "/"
	}
	return parsed.String(), nil
}

func (s *PreflightContentLinkService) defaultTargetURL() (string, error) {
	raw := strings.TrimSpace(s.baseURL)
	if raw == "" {
		return "", errors.New("content link preflight base URL is unavailable")
	}
	parsed, err := url.Parse(raw)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return "", errors.New("content link preflight base URL is invalid")
	}
	if parsed.Path == "" {
		parsed.Path = "/"
	}
	parsed.Fragment = ""
	return parsed.String(), nil
}

func (s *PreflightContentLinkService) routeEntryForURL(rawURL string) *seodomain.StorefrontRouteCatalogEntry {
	if s == nil || s.routeCatalog == nil {
		return nil
	}
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return nil
	}
	path := parsed.EscapedPath()
	if path == "" {
		path = "/"
	}
	entry, err := s.routeCatalog.FindCurrentByPath(path)
	if err == nil {
		return entry
	}
	if errors.Is(err, gorm.ErrRecordNotFound) && path != parsed.Path {
		entry, err = s.routeCatalog.FindCurrentByPath(parsed.Path)
		if err == nil {
			return entry
		}
	}
	return nil
}

func (s *PreflightContentLinkService) urlForRoute(entry seodomain.StorefrontRouteCatalogEntry) (string, error) {
	defaultURL, err := s.defaultTargetURL()
	if err != nil {
		return "", err
	}
	base, err := url.Parse(defaultURL)
	if err != nil {
		return "", err
	}
	path := strings.TrimSpace(entry.CanonicalPath)
	if path == "" {
		path = strings.TrimSpace(entry.Path)
	}
	if path == "" {
		path = "/"
	}
	relative, err := url.Parse(path)
	if err != nil {
		return "", err
	}
	absolute := base.ResolveReference(relative)
	absolute.Fragment = ""
	return absolute.String(), nil
}

func (s *PreflightContentLinkService) suggestContentLinkText(
	match contentLinkTextMatch,
	anchor contentLinkAnchor,
	linkURL string,
) string {
	destination := s.destinationLabel(linkURL)
	if destination == "" {
		destination = strings.TrimSpace(anchor.Heading)
	}
	if destination == "" {
		if match.Language == "zh" {
			return "改为描述目标页面的链接文字"
		}
		return fmt.Sprintf("Describe destination instead of %q", strings.TrimSpace(anchor.Text))
	}
	if match.Language == "zh" {
		return "查看" + destination
	}
	switch match.Kind {
	case "learn":
		return "Learn about " + destination
	case "view":
		return "View " + destination
	default:
		return "Read more about " + destination
	}
}

func (s *PreflightContentLinkService) destinationLabel(linkURL string) string {
	if s == nil {
		return ""
	}
	parsed, err := url.Parse(linkURL)
	if err != nil || parsed.Path == "" {
		return ""
	}
	if defaultURL, err := s.defaultTargetURL(); err == nil {
		if base, baseErr := url.Parse(defaultURL); baseErr == nil &&
			parsed.Host != "" &&
			base.Host != "" &&
			!strings.EqualFold(parsed.Host, base.Host) {
			return parsed.Host
		}
	}
	if s.routeCatalog == nil {
		return ""
	}
	entry, err := s.routeCatalog.FindCurrentByPath(parsed.EscapedPath())
	if err != nil && parsed.EscapedPath() != parsed.Path {
		entry, err = s.routeCatalog.FindCurrentByPath(parsed.Path)
	}
	if err != nil || entry == nil {
		return ""
	}
	if title := strings.TrimSpace(entry.Title); title != "" {
		return title
	}
	if summary := strings.TrimSpace(entry.Summary); summary != "" {
		return summary
	}
	return strings.Trim(strings.TrimSpace(entry.Path), "/")
}

func (s *PreflightContentLinkService) locateContentLinkSource(
	routeEntry *seodomain.StorefrontRouteCatalogEntry,
	anchor contentLinkAnchor,
	linkURL string,
	targetURL string,
) contentLinkSourceLocation {
	sourceKey := targetURL
	sourceType := "storefront_render"
	if routeEntry != nil {
		if strings.TrimSpace(routeEntry.SourceKey) != "" {
			sourceKey = strings.TrimSpace(routeEntry.SourceKey)
		} else if strings.TrimSpace(routeEntry.Path) != "" {
			sourceKey = strings.TrimSpace(routeEntry.Path)
		}
	}
	location := contentLinkSourceLocation{
		SourceType: sourceType,
		SourceKey:  sourceKey,
		FixStatus:  preflightdomain.ContentLinkFixStatusNotFixable,
		SourceHint: "rendered storefront HTML; edit the owning component or i18n copy",
	}
	if routeEntry == nil ||
		routeEntry.SourceID == nil ||
		routeEntry.SourceType != seodomain.RouteSourceBlog ||
		s.postService == nil {
		return location
	}
	foundPost, err := s.postService.GetAdminPost(*routeEntry.SourceID)
	if err != nil || foundPost == nil {
		return location
	}
	if !contentLinkHTMLHasMatchingAnchor(foundPost.Content, anchor.Href, linkURL, anchor.Text, targetURL) {
		return location
	}
	return contentLinkSourceLocation{
		SourceType:  "blog_post",
		SourceID:    routeEntry.SourceID,
		SourceKey:   fmt.Sprintf("posts:%d", *routeEntry.SourceID),
		SourceField: "content",
		FixStatus:   preflightdomain.ContentLinkFixStatusPending,
		SourceHint:  "matched posts.content anchor; backend can update this source safely",
	}
}

func collectContentLinkAnchors(root *html.Node) []contentLinkAnchor {
	anchors := make([]contentLinkAnchor, 0)
	currentHeading := ""
	var walk func(node *html.Node)
	walk = func(node *html.Node) {
		if node == nil {
			return
		}
		if node.Type == html.ElementNode && isHeadingTag(node.Data) {
			headingText := collapseWhitespace(textContent(node))
			if headingText != "" {
				currentHeading = headingText
			}
		}
		if node.Type == html.ElementNode && strings.EqualFold(node.Data, "a") {
			href := strings.TrimSpace(htmlNodeAttr(node, "href"))
			if href != "" {
				anchors = append(anchors, contentLinkAnchor{
					Href:     href,
					Text:     collapseWhitespace(textContent(node)),
					Selector: htmlNodeSelector(node),
					Snippet:  truncateContentLinkText(renderHTMLNode(node), 260),
					Heading:  currentHeading,
				})
			}
		}
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			walk(child)
		}
	}
	walk(root)
	return anchors
}

func matchGenericContentLinkText(value string) (contentLinkTextMatch, bool) {
	normalized := normalizeContentLinkText(value)
	if normalized == "" {
		return contentLinkTextMatch{}, false
	}
	generic := map[string]contentLinkTextMatch{
		"read more":    {Normalized: normalized, Language: "en", Kind: "read"},
		"learn more":   {Normalized: normalized, Language: "en", Kind: "learn"},
		"click here":   {Normalized: normalized, Language: "en", Kind: "view"},
		"here":         {Normalized: normalized, Language: "en", Kind: "view"},
		"more":         {Normalized: normalized, Language: "en", Kind: "read"},
		"view more":    {Normalized: normalized, Language: "en", Kind: "view"},
		"see more":     {Normalized: normalized, Language: "en", Kind: "view"},
		"details":      {Normalized: normalized, Language: "en", Kind: "view"},
		"open article": {Normalized: normalized, Language: "en", Kind: "view"},
		"了解更多":         {Normalized: normalized, Language: "zh", Kind: "learn"},
		"查看更多":         {Normalized: normalized, Language: "zh", Kind: "view"},
		"点击这里":         {Normalized: normalized, Language: "zh", Kind: "view"},
		"更多":           {Normalized: normalized, Language: "zh", Kind: "read"},
		"阅读全文":         {Normalized: normalized, Language: "zh", Kind: "read"},
	}
	match, ok := generic[normalized]
	if !ok {
		return contentLinkTextMatch{}, false
	}
	return match, true
}

func contentLinkHrefInspectable(value string) bool {
	value = strings.TrimSpace(value)
	if value == "" || strings.HasPrefix(value, "#") {
		return false
	}
	parsed, err := url.Parse(value)
	if err != nil {
		return false
	}
	switch strings.ToLower(parsed.Scheme) {
	case "", "http", "https":
		return true
	default:
		return false
	}
}

func normalizeContentLinkText(value string) string {
	value = collapseWhitespace(stdhtml.UnescapeString(value))
	value = strings.TrimFunc(value, func(r rune) bool {
		return unicode.IsSpace(r) || strings.ContainsRune(".,;:!?-–—_()[]{}<>/\\|·•»›→←…", r)
	})
	return strings.ToLower(value)
}

func replaceFirstContentLinkSuggestion(
	content string,
	issue *preflightdomain.ContentLinkIssue,
) (string, bool, error) {
	if issue == nil {
		return "", false, errors.New("content link issue is required")
	}
	suggestedText := strings.TrimSpace(issue.SuggestedText)
	if suggestedText == "" {
		return "", false, errors.New("content link suggestion is empty")
	}
	context := &html.Node{Type: html.ElementNode, DataAtom: atom.Div, Data: "div"}
	nodes, err := html.ParseFragment(strings.NewReader(content), context)
	if err != nil {
		return "", false, fmt.Errorf("parse source content: %w", err)
	}
	replaced := false
	var walk func(node *html.Node)
	walk = func(node *html.Node) {
		if node == nil || replaced {
			return
		}
		if node.Type == html.ElementNode && strings.EqualFold(node.Data, "a") {
			href := strings.TrimSpace(htmlNodeAttr(node, "href"))
			text := collapseWhitespace(textContent(node))
			if contentLinkMatchesIssue(href, issue.LinkURL, text, issue.LinkText, issue.TargetURL) {
				node.FirstChild = nil
				node.LastChild = nil
				node.AppendChild(&html.Node{Type: html.TextNode, Data: suggestedText})
				replaced = true
				return
			}
		}
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			walk(child)
		}
	}
	for _, node := range nodes {
		walk(node)
		if replaced {
			break
		}
	}
	if !replaced {
		return "", false, ErrContentLinkSourceStale
	}
	var builder strings.Builder
	for _, node := range nodes {
		if err := html.Render(&builder, node); err != nil {
			return "", false, err
		}
	}
	return strings.TrimSpace(builder.String()), true, nil
}

func contentLinkHTMLHasMatchingAnchor(
	content string,
	sourceHref string,
	linkURL string,
	linkText string,
	baseURL string,
) bool {
	context := &html.Node{Type: html.ElementNode, DataAtom: atom.Div, Data: "div"}
	nodes, err := html.ParseFragment(strings.NewReader(content), context)
	if err != nil {
		return false
	}
	found := false
	var walk func(node *html.Node)
	walk = func(node *html.Node) {
		if node == nil || found {
			return
		}
		if node.Type == html.ElementNode && strings.EqualFold(node.Data, "a") {
			href := strings.TrimSpace(htmlNodeAttr(node, "href"))
			text := collapseWhitespace(textContent(node))
			if contentLinkMatchesIssue(href, linkURL, text, linkText, baseURL) ||
				contentLinkMatchesIssue(sourceHref, linkURL, text, linkText, baseURL) {
				found = true
				return
			}
		}
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			walk(child)
		}
	}
	for _, node := range nodes {
		walk(node)
		if found {
			return true
		}
	}
	return false
}

func contentLinkMatchesIssue(href string, issueLinkURL string, text string, issueText string, baseURL string) bool {
	if normalizeContentLinkText(text) != normalizeContentLinkText(issueText) {
		return false
	}
	return sameContentLinkDestination(href, issueLinkURL, baseURL)
}

func sameContentLinkDestination(href string, issueLinkURL string, baseURL string) bool {
	left := normalizedContentLinkDestination(href, baseURL)
	right := normalizedContentLinkDestination(issueLinkURL, baseURL)
	return left != "" && right != "" && left == right
}

func normalizedContentLinkDestination(raw string, baseURL string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	parsed, err := url.Parse(raw)
	if err != nil {
		return raw
	}
	base, baseErr := url.Parse(baseURL)
	if baseErr == nil && parsed.Scheme == "" && parsed.Host == "" {
		parsed = base.ResolveReference(parsed)
	}
	parsed.Fragment = ""
	if parsed.Path == "" {
		parsed.Path = "/"
	}
	return parsed.String()
}

func resolvedLinkURL(href string, finalURL string, targetURL string) string {
	baseRaw := strings.TrimSpace(finalURL)
	if baseRaw == "" {
		baseRaw = strings.TrimSpace(targetURL)
	}
	base, err := url.Parse(baseRaw)
	if err != nil {
		return ""
	}
	parsed, err := url.Parse(strings.TrimSpace(href))
	if err != nil {
		return ""
	}
	resolved := base.ResolveReference(parsed)
	resolved.Fragment = ""
	if resolved.Path == "" {
		resolved.Path = "/"
	}
	return resolved.String()
}

func contentLinkIssueKey(targetURL string, linkURL string, normalizedText string, selector string) string {
	hash := sha256.Sum256([]byte(strings.Join([]string{
		strings.TrimSpace(targetURL),
		strings.TrimSpace(linkURL),
		strings.TrimSpace(normalizedText),
		strings.TrimSpace(selector),
	}, "\x00")))
	return "content-link:" + hex.EncodeToString(hash[:])[:32]
}

func encodeContentLinkEvidence(evidence contentLinkEvidence) string {
	encoded, err := json.Marshal(evidence)
	if err != nil {
		return "{}"
	}
	return string(encoded)
}

func htmlNodeAttr(node *html.Node, key string) string {
	for _, attr := range node.Attr {
		if strings.EqualFold(attr.Key, key) {
			return strings.TrimSpace(attr.Val)
		}
	}
	return ""
}

func textContent(node *html.Node) string {
	if node == nil {
		return ""
	}
	if node.Type == html.TextNode {
		return node.Data
	}
	if node.Type == html.ElementNode {
		switch strings.ToLower(node.Data) {
		case "script", "style", "noscript", "svg":
			return ""
		}
	}
	var parts []string
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		if value := textContent(child); value != "" {
			parts = append(parts, value)
		}
	}
	return strings.Join(parts, " ")
}

func collapseWhitespace(value string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(value)), " ")
}

func isHeadingTag(tag string) bool {
	switch strings.ToLower(tag) {
	case "h1", "h2", "h3", "h4", "h5", "h6":
		return true
	default:
		return false
	}
}

func renderHTMLNode(node *html.Node) string {
	var builder strings.Builder
	if err := html.Render(&builder, node); err != nil {
		return ""
	}
	return builder.String()
}

func truncateContentLinkText(value string, limit int) string {
	value = collapseWhitespace(value)
	if limit <= 0 || len(value) <= limit {
		return value
	}
	return value[:limit] + "..."
}

func htmlNodeSelector(node *html.Node) string {
	if node == nil {
		return ""
	}
	parts := make([]string, 0, 8)
	for current := node; current != nil; current = current.Parent {
		if current.Type != html.ElementNode {
			continue
		}
		tag := strings.ToLower(current.Data)
		if tag == "" {
			continue
		}
		index := 1
		for sibling := current.PrevSibling; sibling != nil; sibling = sibling.PrevSibling {
			if sibling.Type == html.ElementNode && strings.EqualFold(sibling.Data, current.Data) {
				index++
			}
		}
		parts = append([]string{fmt.Sprintf("%s:nth-of-type(%d)", tag, index)}, parts...)
		if tag == "html" {
			break
		}
	}
	if len(parts) > 6 {
		parts = parts[len(parts)-6:]
	}
	return strings.Join(parts, " > ")
}

func routeEntryIDFor(entry *seodomain.StorefrontRouteCatalogEntry) *uint {
	if entry == nil || entry.ID == 0 {
		return nil
	}
	id := entry.ID
	return &id
}

func routeEntryIsHome(entry seodomain.StorefrontRouteCatalogEntry) bool {
	path := strings.TrimSpace(entry.CanonicalPath)
	if path == "" {
		path = strings.TrimSpace(entry.Path)
	}
	return path == "/"
}
