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
	"time"

	seodomain "commerce-platform/internal/domain/seo"
	"commerce-platform/internal/repository"

	"golang.org/x/net/html"
)

const routeCheckBodyLimit = 4 * 1024 * 1024

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
	if s == nil || s.repository == nil {
		return StorefrontRouteCatalogCheckSummary{}, errors.New("storefront route catalog service is unavailable")
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if limit < 1 || limit > 200 {
		limit = 100
	}
	filter.Page = 1
	filter.PageSize = limit

	entries, _, err := s.repository.List(filter)
	if err != nil {
		return StorefrontRouteCatalogCheckSummary{}, err
	}

	summary := StorefrontRouteCatalogCheckSummary{}
	for _, entry := range entries {
		if !entry.IsCheckable {
			continue
		}
		result := s.checkEntry(ctx, entry)
		if err := s.repository.SaveCheck(&result); err != nil {
			return summary, fmt.Errorf("save URL check for %s: %w", entry.Path, err)
		}
		if s.issueReconciler != nil {
			if err := s.issueReconciler.ReconcileEntry(ctx, entry.ID, &result.ID); err != nil {
				return summary, fmt.Errorf("reconcile URL issue for %s: %w", entry.Path, err)
			}
		}
		summary.Checked++
		incrementRouteCatalogCheckSummary(&summary, result.Status)
	}
	return summary, nil
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
	result.CanonicalURL = extractCanonicalURL(body)

	switch {
	case response.StatusCode == http.StatusNotFound:
		result.Status = seodomain.RouteCheckStatusNotFound
	case response.StatusCode >= http.StatusInternalServerError:
		result.Status = seodomain.RouteCheckStatusServerError
	case response.StatusCode >= http.StatusMultipleChoices:
		result.Status = seodomain.RouteCheckStatusRedirect
	case entry.IsAlias && redirectCount == 0:
		result.Status = seodomain.RouteCheckStatusRedirectTarget
	case entry.IsAlias && canonicalPath(result.FinalURL) != normalizeCatalogRoutePath(entry.CanonicalPath):
		result.Status = seodomain.RouteCheckStatusRedirectTarget
	case entry.IsAlias && redirectCount > 1:
		result.Status = seodomain.RouteCheckStatusRedirectChain
	case result.CanonicalURL != "" && canonicalPath(result.CanonicalURL) != normalizeCatalogRoutePath(entry.CanonicalPath):
		result.Status = seodomain.RouteCheckStatusCanonicalMisfit
	case redirectCount > 0:
		result.Status = seodomain.RouteCheckStatusRedirect
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
