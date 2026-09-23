package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"commerce-platform/internal/pkg/metrics"
	"commerce-platform/internal/pkg/resilience"
)

const (
	customerServiceSearchIndexDeleteEndpointEnv = "CUSTOMER_SERVICE_SEARCH_INDEX_DELETE_URL"
	customerServiceSearchIndexTokenEnv          = "CUSTOMER_SERVICE_SEARCH_INDEX_TOKEN"
)

// CustomerServiceHTTPSearchIndex is a provider-neutral HTTP adapter for a
// conversation search projection. The endpoint may either contain the
// {ticket_id} placeholder or be a collection URL to which the ticket id is
// appended as one path segment.
type CustomerServiceHTTPSearchIndex struct {
	endpoint string
	token    string
	client   *resilience.HTTPClient
}

// NewCustomerServiceHTTPSearchIndex constructs the optional search adapter.
// A blank endpoint deliberately returns nil so existing deployments remain
// database-only. Once configured, malformed endpoints fail startup rather than
// silently skipping index cleanup.
func NewCustomerServiceHTTPSearchIndex(
	endpoint string,
	token string,
	client *http.Client,
	retry resilience.HTTPRetryPolicy,
	breaker resilience.CircuitController,
) (*CustomerServiceHTTPSearchIndex, error) {
	endpoint = strings.TrimSpace(endpoint)
	if endpoint == "" {
		return nil, nil
	}
	if err := validateCustomerServiceSearchIndexEndpoint(endpoint); err != nil {
		return nil, err
	}
	if client == nil {
		client = &http.Client{Timeout: 8 * time.Second}
	}
	return &CustomerServiceHTTPSearchIndex{
		endpoint: endpoint,
		token:    strings.TrimSpace(token),
		client:   resilience.NewHTTPClient(client, retry, breaker, "customer-service-search-index-delete"),
	}, nil
}

// NewCustomerServiceHTTPSearchIndexFromEnv wires the optional adapter from
// deployment environment variables. It is intentionally independent from the
// retention worker switch: an index projection can be cleaned by an explicit
// admin purge even while the scheduled worker remains disabled.
func NewCustomerServiceHTTPSearchIndexFromEnv(
	retry resilience.HTTPRetryPolicy,
	breaker resilience.CircuitController,
) (*CustomerServiceHTTPSearchIndex, error) {
	return NewCustomerServiceHTTPSearchIndex(
		os.Getenv(customerServiceSearchIndexDeleteEndpointEnv),
		os.Getenv(customerServiceSearchIndexTokenEnv),
		&http.Client{Timeout: 8 * time.Second},
		retry,
		breaker,
	)
}

func (i *CustomerServiceHTTPSearchIndex) DeleteConversation(ctx context.Context, ticketID uint) error {
	if i == nil || i.client == nil {
		return errors.New("customer-service search index is not configured")
	}
	if ticketID == 0 {
		return errors.New("customer-service search index ticket id is required")
	}
	if ctx == nil {
		ctx = context.Background()
	}
	deleteURL, err := i.deleteURL(ticketID)
	if err != nil {
		metrics.CustomerServiceRetentionSearchIndexDeletes.WithLabelValues("error").Inc()
		return err
	}
	idempotencyKey := "customer-service-conversation:" + strconv.FormatUint(uint64(ticketID), 10)
	resp, err := i.client.Do(ctx, func() (*http.Request, error) {
		req, requestErr := http.NewRequestWithContext(ctx, http.MethodDelete, deleteURL, nil)
		if requestErr != nil {
			return nil, requestErr
		}
		req.Header.Set("Accept", "application/json")
		req.Header.Set("Idempotency-Key", idempotencyKey)
		req.Header.Set("X-Customer-Service-Conversation-ID", strconv.FormatUint(uint64(ticketID), 10))
		if i.token != "" {
			req.Header.Set("Authorization", "Bearer "+i.token)
		}
		return req, nil
	})
	if err != nil {
		if resp != nil && resp.Body != nil {
			_ = resp.Body.Close()
		}
		metrics.CustomerServiceRetentionSearchIndexDeletes.WithLabelValues("error").Inc()
		return fmt.Errorf("delete conversation from search index: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound || resp.StatusCode == http.StatusGone {
		_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 1<<20))
		metrics.CustomerServiceRetentionSearchIndexDeletes.WithLabelValues("already_deleted").Inc()
		return nil
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		metrics.CustomerServiceRetentionSearchIndexDeletes.WithLabelValues("error").Inc()
		return fmt.Errorf("search index returned %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 1<<20))
	metrics.CustomerServiceRetentionSearchIndexDeletes.WithLabelValues("deleted").Inc()
	return nil
}

func (i *CustomerServiceHTTPSearchIndex) deleteURL(ticketID uint) (string, error) {
	if i == nil || strings.TrimSpace(i.endpoint) == "" {
		return "", errors.New("customer-service search index endpoint is not configured")
	}
	if strings.Contains(i.endpoint, "{ticket_id}") {
		return strings.ReplaceAll(i.endpoint, "{ticket_id}", url.PathEscape(strconv.FormatUint(uint64(ticketID), 10))), nil
	}
	parsed, err := url.Parse(i.endpoint)
	if err != nil {
		return "", fmt.Errorf("parse customer-service search index endpoint: %w", err)
	}
	parsed.Path = strings.TrimRight(parsed.Path, "/") + "/" + strconv.FormatUint(uint64(ticketID), 10)
	parsed.RawPath = ""
	return parsed.String(), nil
}

func validateCustomerServiceSearchIndexEndpoint(endpoint string) error {
	parsed, err := url.Parse(endpoint)
	if err != nil {
		return fmt.Errorf("invalid customer-service search index endpoint: %w", err)
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return errors.New("customer-service search index endpoint must use HTTP or HTTPS")
	}
	if parsed.Host == "" {
		return errors.New("customer-service search index endpoint must include a host")
	}
	if parsed.User != nil {
		return errors.New("customer-service search index endpoint must not include credentials")
	}
	if parsed.Fragment != "" {
		return errors.New("customer-service search index endpoint must not include a fragment")
	}
	return nil
}
