package service

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	seodomain "commerce-platform/internal/domain/seo"
	"commerce-platform/internal/repository"

	"golang.org/x/net/html"
)

const routeCheckBodyLimit = 4 * 1024 * 1024
const routeCheckConcurrency = 6

func (s *StorefrontRouteCatalogService) CheckEntry(ctx context.Context, id uint) (seodomain.StorefrontRouteCheckResult, error) {
	if s == nil || s.repository == nil {
		return seodomain.StorefrontRouteCheckResult{}, errors.New("storefront route catalog service is unavailable")
	}
	if id == 0 {
		return seodomain.StorefrontRouteCheckResult{}, errors.New("route entry ID is required")
	}
	if ctx == nil {
		ctx = context.Background()
	}

	entry, err := s.repository.FindByID(id)
	if err != nil {
		return seodomain.StorefrontRouteCheckResult{}, err
	}
	if entry.EntryStatus == seodomain.RouteEntryStatusStale {
		return seodomain.StorefrontRouteCheckResult{}, fmt.Errorf("route %s is stale and must be synced before checking", entry.Path)
	}
	if !entry.IsCheckable {
		return seodomain.StorefrontRouteCheckResult{}, fmt.Errorf("route %s is not checkable", entry.Path)
	}

	result := s.checkEntry(ctx, *entry)
	if err := s.repository.SaveCheck(&result); err != nil {
		return seodomain.StorefrontRouteCheckResult{}, fmt.Errorf("save URL check for %s: %w", entry.Path, err)
	}
	if s.issueReconciler != nil {
		if err := s.issueReconciler.ReconcileEntry(ctx, entry.ID, &result.ID); err != nil {
			return seodomain.StorefrontRouteCheckResult{}, fmt.Errorf("reconcile URL issue for %s: %w", entry.Path, err)
		}
	}
	return result, nil
}

func (s *StorefrontRouteCatalogService) Check(ctx context.Context, filter repository.StorefrontRouteCatalogListFilter, limit int) (StorefrontRouteCatalogCheckSummary, error) {
	return s.checkBatch(ctx, filter, limit, nil)
}

type storefrontRouteCatalogCheckProgress func(StorefrontRouteCatalogCheckSummary)

func (s *StorefrontRouteCatalogService) checkBatch(
	ctx context.Context,
	filter repository.StorefrontRouteCatalogListFilter,
	limit int,
	progress storefrontRouteCatalogCheckProgress,
) (StorefrontRouteCatalogCheckSummary, error) {
	if s == nil || s.repository == nil {
		return StorefrontRouteCatalogCheckSummary{}, errors.New("storefront route catalog service is unavailable")
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if limit < 1 {
		limit = 100
	} else if limit > 200 {
		limit = 200
	}
	filter.CheckableOnly = true
	filter.Page = 1
	filter.PageSize = limit

	entries, _, err := s.repository.List(filter)
	if err != nil {
		return StorefrontRouteCatalogCheckSummary{}, err
	}

	checkableEntries := make([]seodomain.StorefrontRouteCatalogEntry, 0, len(entries))
	for _, entry := range entries {
		if routeEntryCanBeChecked(entry) {
			checkableEntries = append(checkableEntries, entry)
		}
	}

	// The repository's total is the coarse SQL count. Keep task progress tied
	// to the entries that can actually be checked after the same guard used by
	// the worker, otherwise stale rows would make `remaining` never reach zero.
	summary := StorefrontRouteCatalogCheckSummary{
		Eligible:  len(checkableEntries),
		Remaining: len(checkableEntries),
	}
	if progress != nil {
		progress(summary)
	}

	workerCount := routeCheckConcurrency
	if len(checkableEntries) < workerCount {
		workerCount = len(checkableEntries)
	}
	if workerCount == 0 {
		return summary, nil
	}

	type checkResult struct {
		entry  seodomain.StorefrontRouteCatalogEntry
		result seodomain.StorefrontRouteCheckResult
	}
	jobs := make(chan seodomain.StorefrontRouteCatalogEntry)
	results := make(chan checkResult, len(checkableEntries))
	var wg sync.WaitGroup
	for worker := 0; worker < workerCount; worker++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for entry := range jobs {
				results <- checkResult{entry: entry, result: s.checkEntry(ctx, entry)}
			}
		}()
	}
	go func() {
		defer close(jobs)
		for _, entry := range checkableEntries {
			jobs <- entry
		}
	}()
	go func() {
		wg.Wait()
		close(results)
	}()

	for checked := range results {
		if err := s.repository.SaveCheck(&checked.result); err != nil {
			return summary, fmt.Errorf("save URL check for %s: %w", checked.entry.Path, err)
		}
		if s.issueReconciler != nil {
			if err := s.issueReconciler.ReconcileEntry(ctx, checked.entry.ID, &checked.result.ID); err != nil {
				return summary, fmt.Errorf("reconcile URL issue for %s: %w", checked.entry.Path, err)
			}
		}
		summary.Checked++
		summary.Remaining--
		incrementRouteCatalogCheckSummary(&summary, checked.result.Status)
		if progress != nil {
			progress(summary)
		}
	}
	return summary, nil
}

func routeEntryCanBeChecked(entry seodomain.StorefrontRouteCatalogEntry) bool {
	return entry.IsCheckable && entry.EntryStatus != seodomain.RouteEntryStatusStale
}

func incrementRouteCatalogCheckSummary(summary *StorefrontRouteCatalogCheckSummary, status string) {
	switch status {
	case seodomain.RouteCheckStatusOK:
		summary.OK++
	case seodomain.RouteCheckStatusRedirect:
		summary.Redirects++
	case seodomain.RouteCheckStatusNotFound:
		summary.NotFound++
	case seodomain.RouteCheckStatusServerError:
		summary.ServerErrors++
	case seodomain.RouteCheckStatusCanonicalMisfit:
		summary.CanonicalMiss++
	default:
		summary.Errors++
	}
}

func (s *StorefrontRouteCatalogService) checkEntry(ctx context.Context, entry seodomain.StorefrontRouteCatalogEntry) seodomain.StorefrontRouteCheckResult {
	startedAt := time.Now()
	result := seodomain.StorefrontRouteCheckResult{
		RouteEntryID: entry.ID,
		CheckedAt:    time.Now().UTC(),
		Status:       seodomain.RouteCheckStatusError,
	}

	target := s.internalBaseURL + entry.Path
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
	if err != nil {
		result.ErrorMessage = err.Error()
		result.ResponseMS = int(time.Since(startedAt).Milliseconds())
		return result
	}

	redirectCount := 0
	client := *s.httpClient
	client.CheckRedirect = func(_ *http.Request, via []*http.Request) error {
		redirectCount = len(via)
		if len(via) >= 10 {
			return http.ErrUseLastResponse
		}
		return nil
	}

	response, err := client.Do(request)
	if err != nil {
		result.ErrorMessage = err.Error()
		result.RedirectCount = redirectCount
		result.ResponseMS = int(time.Since(startedAt).Milliseconds())
		return result
	}
	defer response.Body.Close()

	body, readErr := io.ReadAll(io.LimitReader(response.Body, routeCheckBodyLimit))
	result.HTTPStatus = response.StatusCode
	result.RedirectCount = redirectCount
	result.ResponseMS = int(time.Since(startedAt).Milliseconds())
	if response.Request != nil && response.Request.URL != nil {
		result.FinalURL = response.Request.URL.String()
	}
	if readErr != nil {
		result.ErrorMessage = readErr.Error()
		return result
	}

	sum := sha256.Sum256(body)
	result.ContentHash = hex.EncodeToString(sum[:])
	if redirectCount == 0 && response.StatusCode == http.StatusOK {
		result.CanonicalURL = extractCanonicalURL(body)
	}

	switch {
	case redirectCount > 0 && entry.IsAlias && canonicalPath(result.FinalURL) != normalizeCatalogRoutePath(entry.CanonicalPath):
		result.Status = seodomain.RouteCheckStatusRedirectTarget
	case redirectCount > 1 && entry.IsAlias:
		result.Status = seodomain.RouteCheckStatusRedirectChain
	case redirectCount > 0:
		result.Status = seodomain.RouteCheckStatusRedirect
	case response.StatusCode == http.StatusNotFound:
		result.Status = seodomain.RouteCheckStatusNotFound
	case response.StatusCode >= http.StatusInternalServerError:
		result.Status = seodomain.RouteCheckStatusServerError
	case response.StatusCode >= http.StatusMultipleChoices:
		result.Status = seodomain.RouteCheckStatusRedirect
	case entry.IsAlias && redirectCount == 0:
		result.Status = seodomain.RouteCheckStatusRedirectTarget
	// A direct 200 is the only response for which the HTML canonical belongs
	// to this route. If the request redirected, the cases above classify that
	// redirect first; otherwise the destination page's canonical could be
	// reported as a false mismatch for the original URL.
	case response.StatusCode == http.StatusOK && result.CanonicalURL != "" && canonicalPath(result.CanonicalURL) != normalizeCatalogRoutePath(entry.CanonicalPath):
		result.Status = seodomain.RouteCheckStatusCanonicalMisfit
	case response.StatusCode >= http.StatusOK && response.StatusCode < http.StatusMultipleChoices:
		result.Status = seodomain.RouteCheckStatusOK
	default:
		result.Status = seodomain.RouteCheckStatusError
	}

	return result
}

func extractCanonicalURL(body []byte) string {
	document, err := html.Parse(bytes.NewReader(body))
	if err != nil {
		return ""
	}
	var walk func(*html.Node) string
	walk = func(node *html.Node) string {
		if node.Type == html.ElementNode && node.Data == "link" {
			rel := ""
			href := ""
			for _, attr := range node.Attr {
				switch attr.Key {
				case "rel":
					rel = strings.ToLower(strings.TrimSpace(attr.Val))
				case "href":
					href = strings.TrimSpace(attr.Val)
				}
			}
			if href != "" && strings.Contains(rel, "canonical") {
				return href
			}
		}
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			if found := walk(child); found != "" {
				return found
			}
		}
		return ""
	}
	return walk(document)
}

func canonicalPath(value string) string {
	parsed, err := url.Parse(strings.TrimSpace(value))
	if err != nil {
		return ""
	}
	return normalizeCatalogRoutePath(parsed.Path)
}
