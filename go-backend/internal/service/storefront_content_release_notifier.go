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

const storefrontContentReleaseTokenHeader = "X-Storefront-Content-Release-Token"

type StorefrontContentReleaseNotifier struct {
	webhookURL string
	token      string
	client     *http.Client
	timeout    time.Duration
	debounce   time.Duration

	mu            sync.Mutex
	pendingReason string
	debounceTimer *time.Timer
}

func NewStorefrontContentReleaseNotifierFromEnv() *StorefrontContentReleaseNotifier {
	webhookURL := strings.TrimSpace(os.Getenv("STOREFRONT_CONTENT_RELEASE_WEBHOOK_URL"))
	token := strings.TrimSpace(os.Getenv("STOREFRONT_CONTENT_RELEASE_WEBHOOK_TOKEN"))
	if webhookURL == "" || token == "" {
		return nil
	}

	return &StorefrontContentReleaseNotifier{
		webhookURL: webhookURL,
		token:      token,
		client: &http.Client{
			Timeout: 3 * time.Second,
		},
		timeout:  3 * time.Second,
		debounce: durationFromMillisecondsEnv("STOREFRONT_CONTENT_RELEASE_DEBOUNCE_MS", time.Second),
	}
}

func (n *StorefrontContentReleaseNotifier) Enabled() bool {
	return n != nil &&
		strings.TrimSpace(n.webhookURL) != "" &&
		strings.TrimSpace(n.token) != ""
}

func (n *StorefrontContentReleaseNotifier) TriggerAsync(reason string) {
	if !n.Enabled() {
		return
	}

	reason = strings.TrimSpace(reason)
	if n.debounce > 0 {
		n.scheduleDebouncedTrigger(reason)
		return
	}

	go n.triggerWithTimeout(reason)
}

func (n *StorefrontContentReleaseNotifier) scheduleDebouncedTrigger(reason string) {
	n.mu.Lock()
	defer n.mu.Unlock()

	if reason != "" {
		n.pendingReason = reason
	}
	if n.debounceTimer != nil {
		return
	}

	n.debounceTimer = time.AfterFunc(n.debounce, func() {
		n.mu.Lock()
		pendingReason := n.pendingReason
		n.pendingReason = ""
		n.debounceTimer = nil
		n.mu.Unlock()

		n.triggerWithTimeout(pendingReason)
	})
}

func (n *StorefrontContentReleaseNotifier) triggerWithTimeout(reason string) {
	ctx, cancel := context.WithTimeout(context.Background(), n.timeout)
	defer cancel()

	if err := n.Trigger(ctx, reason); err != nil {
		log.Printf("storefront content release webhook failed: %v", err)
	}
}

func (n *StorefrontContentReleaseNotifier) Trigger(ctx context.Context, reason string) error {
	if !n.Enabled() {
		return nil
	}

	payload := map[string]string{
		"event":  "storefront_content_published",
		"reason": strings.TrimSpace(reason),
		"source": "commerce_platform-go-backend",
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, n.webhookURL, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(storefrontContentReleaseTokenHeader, n.token)

	resp, err := n.client.Do(req)
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
