package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

const storefrontHTMLCachePurgeHeader = "X-HTML-Cache-Purge-Token"

type StorefrontHTMLCacheInvalidator struct {
	purgeURL      string
	token         string
	client        *http.Client
	timeout       time.Duration
	debounce      time.Duration
	mu            sync.Mutex
	pendingReason string
	debounceTimer *time.Timer
}

func NewStorefrontHTMLCacheInvalidatorFromEnv() *StorefrontHTMLCacheInvalidator {
	purgeURL := strings.TrimSpace(os.Getenv("STOREFRONT_HTML_CACHE_PURGE_URL"))
	token := strings.TrimSpace(os.Getenv("STOREFRONT_HTML_CACHE_PURGE_TOKEN"))
	if purgeURL == "" || token == "" {
		return nil
	}

	return &StorefrontHTMLCacheInvalidator{
		purgeURL: purgeURL,
		token:    token,
		client: &http.Client{
			Timeout: 3 * time.Second,
		},
		timeout:  3 * time.Second,
		debounce: durationFromMillisecondsEnv("STOREFRONT_HTML_CACHE_PURGE_DEBOUNCE_MS", 500*time.Millisecond),
	}
}

func durationFromMillisecondsEnv(key string, fallback time.Duration) time.Duration {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback
	}

	parsed, err := time.ParseDuration(raw + "ms")
	if err != nil || parsed < 0 {
		return fallback
	}

	return parsed
}

func (i *StorefrontHTMLCacheInvalidator) Enabled() bool {
	return i != nil && strings.TrimSpace(i.purgeURL) != "" && strings.TrimSpace(i.token) != ""
}

func (i *StorefrontHTMLCacheInvalidator) PurgeAllAsync(reason string) {
	if !i.Enabled() {
		return
	}

	reason = strings.TrimSpace(reason)
	if i.debounce > 0 {
		i.scheduleDebouncedPurge(reason)
		return
	}

	go i.purgeAllWithTimeout(reason)
}

func (i *StorefrontHTMLCacheInvalidator) scheduleDebouncedPurge(reason string) {
	i.mu.Lock()
	defer i.mu.Unlock()

	if reason != "" {
		i.pendingReason = reason
	}
	if i.debounceTimer != nil {
		return
	}

	i.debounceTimer = time.AfterFunc(i.debounce, func() {
		i.mu.Lock()
		pendingReason := i.pendingReason
		i.pendingReason = ""
		i.debounceTimer = nil
		i.mu.Unlock()

		i.purgeAllWithTimeout(pendingReason)
	})
}

func (i *StorefrontHTMLCacheInvalidator) purgeAllWithTimeout(reason string) {
	ctx, cancel := context.WithTimeout(context.Background(), i.timeout)
	defer cancel()

	if err := i.PurgeAll(ctx, reason); err != nil {
		log.Printf("storefront html cache purge failed: %v", err)
	}
}

func (i *StorefrontHTMLCacheInvalidator) PurgeAll(ctx context.Context, reason string) error {
	if !i.Enabled() {
		return nil
	}

	payload := map[string]string{
		"reason": strings.TrimSpace(reason),
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, i.purgeURL, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(storefrontHTMLCachePurgeHeader, i.token)

	resp, err := i.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		responseBody, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return fmt.Errorf("status=%d body=%s", resp.StatusCode, strings.TrimSpace(string(responseBody)))
	}

	return nil
}
