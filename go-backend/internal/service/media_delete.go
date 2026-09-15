package service

import (
	"bytes"
	"commerce-platform/internal/domain/media"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)

var (
	ErrMediaAssetInUse                 = errors.New("media asset is currently in use")
	ErrMediaDeleteConfirmationRequired = errors.New("media delete confirmation is invalid")
)

type MediaAssetInUseError struct {
	References []media.AssetReference
}

func (e *MediaAssetInUseError) Error() string {
	return ErrMediaAssetInUse.Error()
}

func (e *MediaAssetInUseError) Unwrap() error {
	return ErrMediaAssetInUse
}

func MediaAssetDeleteConfirmation(id uint) string {
	return fmt.Sprintf("DELETE %d", id)
}

// DeleteAsset permanently removes an unreferenced asset from storage and the
// database. The reference check is repeated here so the client-side dialog
// cannot bypass safety rules.
func (s *MediaService) DeleteAsset(ctx context.Context, id uint, confirmation string) error {
	asset, err := s.GetAsset(id)
	if err != nil {
		return err
	}
	if !isMediaDeleteConfirmationValid(id, confirmation) {
		return ErrMediaDeleteConfirmationRequired
	}

	references, err := s.repo.FindAssetReferences(asset)
	if err != nil {
		return err
	}
	if len(references) > 0 {
		return &MediaAssetInUseError{References: references}
	}
	if s.storage == nil {
		return ErrMediaStorageUnavailable
	}

	derivatives, err := s.repo.FindAssetDerivatives(asset.ID)
	if err != nil {
		return fmt.Errorf("list media derivatives: %w", err)
	}
	for _, derivative := range derivatives {
		if strings.TrimSpace(derivative.URL) == "" {
			continue
		}
		if err := s.storage.Delete(ctx, derivative.URL); err != nil {
			return fmt.Errorf("delete media derivative object: %w", err)
		}
	}
	if err := s.storage.Delete(ctx, asset.URL); err != nil {
		return fmt.Errorf("delete media object: %w", err)
	}
	if err := s.repo.HardDeleteAsset(id); err != nil {
		return fmt.Errorf("delete media asset record: %w", err)
	}
	if s.cdnPurger != nil {
		s.cdnPurger.PurgeAsync(s.CanonicalPublicMediaURL(asset.URL))
		for _, derivative := range derivatives {
			s.cdnPurger.PurgeAsync(s.CanonicalPublicMediaURL(derivative.URL))
		}
	}
	return nil
}

type mediaCDNPurger struct {
	url    string
	token  string
	client *http.Client
}

func newMediaCDNPurgerFromEnv() *mediaCDNPurger {
	url := strings.TrimSpace(os.Getenv("MEDIA_CDN_PURGE_URL"))
	token := strings.TrimSpace(os.Getenv("MEDIA_CDN_PURGE_TOKEN"))
	if url == "" || token == "" {
		return nil
	}
	return &mediaCDNPurger{url: url, token: token, client: &http.Client{Timeout: 3 * time.Second}}
}

func (p *mediaCDNPurger) PurgeAsync(reference string) {
	if p == nil || strings.TrimSpace(reference) == "" {
		return
	}
	client := p.client
	if client == nil {
		client = http.DefaultClient
	}
	go func() {
		payload, err := json.Marshal(map[string]interface{}{"files": []string{reference}})
		if err != nil {
			return
		}
		req, err := http.NewRequest(http.MethodPost, p.url, bytes.NewReader(payload))
		if err != nil {
			return
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+p.token)
		resp, err := client.Do(req)
		if err == nil && resp != nil {
			_ = resp.Body.Close()
		}
	}()
}

func isMediaDeleteConfirmationValid(id uint, confirmation string) bool {
	return strings.TrimSpace(confirmation) == MediaAssetDeleteConfirmation(id)
}
