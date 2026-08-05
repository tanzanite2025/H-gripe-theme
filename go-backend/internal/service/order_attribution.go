package service

import (
	attributiondomain "tanzanite/internal/domain/attribution"
	attributionpkg "tanzanite/internal/pkg/attribution"
	"tanzanite/internal/repository"
	"time"
)

func persistOrderAttribution(
	repo *repository.OrderAttributionRepository,
	orderID uint,
	context attributionpkg.Context,
) error {
	if repo == nil || orderID == 0 {
		return nil
	}
	normalized, ok := attributionpkg.Normalize(context)
	if !ok {
		return nil
	}
	if normalized.CapturedAt.IsZero() {
		normalized.CapturedAt = time.Now().UTC()
	}
	return repo.Create(&attributiondomain.OrderAttribution{
		OrderID:     orderID,
		Source:      normalized.Source,
		Medium:      normalized.Medium,
		Campaign:    normalized.Campaign,
		Term:        normalized.Term,
		Content:     normalized.Content,
		ClickIDKind: normalized.ClickIDKind,
		ClickID:     normalized.ClickID,
		CapturedAt:  normalized.CapturedAt,
	})
}
