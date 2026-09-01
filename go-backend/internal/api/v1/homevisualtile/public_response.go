package homevisualtile

import (
	"commerce-platform/internal/api/v1/publicmedia"
	domain "commerce-platform/internal/domain/homevisualtile"
)

type publicItemResponse struct {
	ID              uint   `json:"id"`
	TileSetKey      string `json:"showcase_key"`
	Locale          string `json:"locale"`
	ImageURL        string `json:"image_url"`
	ThumbnailURL    string `json:"thumbnail_url"`
	Title           string `json:"title"`
	Caption         string `json:"caption"`
	AltText         string `json:"alt_text"`
	DesktopOrder    int    `json:"desktop_order"`
	MobilePairIndex int    `json:"mobile_pair_index"`
	TargetURL       string `json:"target_url,omitempty"`
	TargetLabel     string `json:"target_label,omitempty"`
	LayoutVariant   string `json:"layout_variant"`
	Width           int    `json:"width"`
	Height          int    `json:"height"`
}

func publicItemFromDomain(item domain.Tile, resolver publicmedia.Resolver) publicItemResponse {
	return publicItemResponse{
		ID:              item.ID,
		TileSetKey:      item.TileSetKey,
		Locale:          item.Locale,
		ImageURL:        publicmedia.URL(resolver, item.ImageURL),
		ThumbnailURL:    publicmedia.URL(resolver, item.ThumbnailURL),
		Title:           item.Title,
		Caption:         item.Caption,
		AltText:         item.AltText,
		DesktopOrder:    item.DesktopOrder,
		MobilePairIndex: item.MobilePairIndex,
		TargetURL:       item.TargetURL,
		TargetLabel:     item.TargetLabel,
		LayoutVariant:   item.LayoutVariant,
		Width:           item.Width,
		Height:          item.Height,
	}
}

func publicItemsFromDomain(items []domain.Tile, resolver publicmedia.Resolver) []publicItemResponse {
	result := make([]publicItemResponse, 0, len(items))
	for _, item := range items {
		result = append(result, publicItemFromDomain(item, resolver))
	}
	return result
}
