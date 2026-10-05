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

// Storefront route checks render the complete Nuxt HTML document so the
// checker can validate canonical metadata. Keep the worker count low enough
// that several SSR renders do not queue behind one another in development or
// on a small production instance.
const routeCheckConcurrency = 2

// Nuxt SSR can take several seconds while loading page data or compiling a
// route for the first time. Five seconds turns that normal startup cost into
// a false failed check, so allow a full page render up to fifteen seconds.
const storefrontRouteCatalogRequestTimeout = 15 * time.Second

func (s *StorefrontRouteCatalogService) CheckEntry(ctx context.Context, id uint) (seodomain.StorefrontRouteCheckResult, error) {
	return s.checkRouteCatalogEntryByID(ctx, id, false)
}

// CheckStaleRouteEntry is reserved for the issue queue's explicit historical
// path verification. Current route checks must never probe a path that the
// latest manifest has removed.
func (s *StorefrontRouteCatalogService) CheckStaleRouteEntry(ctx context.Context, id uint) (seodomain.StorefrontRouteCheckResult, error) {
	return s.checkRouteCatalogEntryByID(ctx, id, true)
}

func (s *StorefrontRouteCatalogService) checkRouteCatalogEntryByID(
	ctx context.Context,
	id uint,
	includeStale bool,
) (seodomain.StorefrontRouteCheckResult, error) {
	if s == nil || s.repository == nil {
		return seodomain.StorefrontRouteCheckResult{}, errors.New("storefront route catalog service is unavailable")
	}
	if id == 0 {
		return seodomain.StorefrontRouteCheckResult{}, errors.New("route entry ID is required")
	}
	releaseOperation, err := s.beginCatalogOperation()
	if err != nil {
		return seodomain.StorefrontRouteCheckResult{}, err
	}
	defer releaseOperation()
	if ctx == nil {
		ctx = context.Background()
	}

	entry, err := s.repository.FindByID(id)
	if err != nil {
		return seodomain.StorefrontRouteCheckResult{}, err
	}
	if entry.EntryStatus == seodomain.RouteEntryStatusStale && !includeStale {
		return seodomain.StorefrontRouteCheckResult{}, fmt.Errorf("route %s is stale and can only be checked from the historical route issue workflow", entry.Path)
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

func (s *StorefrontRouteCatalogService) Check(ctx context.Context, filter repository.StorefrontRouteCatalogListFilter, batchSize int) (StorefrontRouteCatalogCheckSummary, error) {
	if s == nil || s.repository == nil {
		return StorefrontRouteCatalogCheckSummary{}, errors.New("storefront route catalog service is unavailable")
	}
	releaseOperation, err := s.beginCatalogOperation()
	if err != nil {
		return StorefrontRouteCatalogCheckSummary{}, err
	}
	defer releaseOperation()
	return s.checkBatch(ctx, filter, batchSize, nil)
}

type storefrontRouteCatalogCheckProgress func(StorefrontRouteCatalogCheckSummary)

func (s *StorefrontRouteCatalogService) checkBatch(
	ctx context.Context,
	filter repository.StorefrontRouteCatalogListFilter,
	batchSize int,
	progress storefrontRouteCatalogCheckProgress,
) (StorefrontRouteCatalogCheckSummary, error) {
	if s == nil || s.repository == nil {
		return StorefrontRouteCatalogCheckSummary{}, errors.New("storefront route catalog service is unavailable")
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if batchSize < 1 || batchSize > 200 {
		batchSize = 200
	}
	filter.CheckableOnly = true
	filter.PageSize = batchSize

	// Snapshot every matching route before starting checks. Some filters depend
	// on the current check status, which changes as routes are checked; applying
	// those filters while walking pages would skip later rows.
	checkableEntries := make([]seodomain.StorefrontRouteCatalogEntry, 0)
	for page := 1; ; page++ {
		filter.Page = page
		entries, total, err := s.repository.List(filter)
		if err != nil {
			return StorefrontRouteCatalogCheckSummary{}, err
		}
		for _, entry := range entries {
			if routeEntryCanBeChecked(entry, filter.IncludeStaleWhenChecking) {
				checkableEntries = append(checkableEntries, seodomain.StorefrontRouteCatalogEntry{
					ID:            entry.ID,
					Path:          entry.Path,
					CanonicalPath: entry.CanonicalPath,
					IsAlias:       entry.IsAlias,
					IsCheckable:   entry.IsCheckable,
					EntryStatus:   entry.EntryStatus,
				})
			}
		}
		if len(entries) == 0 || int64(page*batchSize) >= total {
			break
		}
	}

	// The repository's total is the coarse SQL count. Keep task progress tied
	// to the entries that can actually be checked after the same guard used by
	// the worker, otherwise stale rows would make `remaining` never reach zero.
	totalBatches := 0
	if len(checkableEntries) > 0 {
		totalBatches = (len(checkableEntries) + batchSize - 1) / batchSize
	}
	summary := StorefrontRouteCatalogCheckSummary{
		Eligible:     len(checkableEntries),
		Remaining:    len(checkableEntries),
		BatchSize:    batchSize,
		TotalBatches: totalBatches,
	}
	if progress != nil {
		progress(summary)
	}

	for batchStart := 0; batchStart < len(checkableEntries); batchStart += batchSize {
		if err := ctx.Err(); err != nil {
			return summary, err
		}
		batchEnd := batchStart + batchSize
		if batchEnd > len(checkableEntries) {
			batchEnd = len(checkableEntries)
		}
		summary.CurrentBatch = batchStart/batchSize + 1
		if progress != nil {
			progress(summary)
		}

		batchResults := s.checkRouteCatalogBatch(ctx, checkableEntries[batchStart:batchEnd])
		for checked := range batchResults {
			// A task timeout means the result currently being processed is no
			// longer part of this run. Leave it unpersisted so `remaining` really
			// identifies work that can be retried.
			if err := ctx.Err(); err != nil {
				return summary, err
			}
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
	}
	if summary.TotalBatches > 0 {
		summary.CurrentBatch = summary.TotalBatches
	}
	return summary, nil
}

type storefrontRouteCatalogCheckResult struct {
	entry  seodomain.StorefrontRouteCatalogEntry
	result seodomain.StorefrontRouteCheckResult
}

// checkRouteCatalogBatch runs at most routeCheckConcurrency SSR requests at a
// time and returns a bounded result set for persistence before the next batch
// starts. This makes batch_size an actual work boundary instead of only a SQL
// page size.
func (s *StorefrontRouteCatalogService) checkRouteCatalogBatch(
	ctx context.Context,
	entries []seodomain.StorefrontRouteCatalogEntry,
) <-chan storefrontRouteCatalogCheckResult {
	results := make(chan storefrontRouteCatalogCheckResult, len(entries))
	jobs := make(chan seodomain.StorefrontRouteCatalogEntry)
	workerCount := routeCheckConcurrency
	if len(entries) < workerCount {
		workerCount = len(entries)
	}
	var workers sync.WaitGroup
	for worker := 0; worker < workerCount; worker++ {
		workers.Add(1)
		go func() {
			defer workers.Done()
			for entry := range jobs {
				results <- storefrontRouteCatalogCheckResult{entry: entry, result: s.checkEntry(ctx, entry)}
			}
		}()
	}
	go func() {
		defer close(jobs)
		for _, entry := range entries {
			jobs <- entry
		}
	}()
	go func() {
		workers.Wait()
		close(results)
	}()
	return results
}

func routeEntryCanBeChecked(
	entry seodomain.StorefrontRouteCatalogEntry,
	includeStaleWhenChecking bool,
) bool {
	return entry.IsCheckable && (includeStaleWhenChecking || entry.EntryStatus != seodomain.RouteEntryStatusStale)
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
