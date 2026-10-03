package service

import (
	"errors"
	"testing"
)

func TestHomeHeroVisualShowcaseAllowsUpToNineItems(t *testing.T) {
	tests := []struct {
		name      string
		itemCount int
		wantError error
	}{
		{name: "no items clears the showcase", itemCount: 0},
		{name: "six items", itemCount: 6},
		{name: "exactly nine items", itemCount: HomeHeroVisualShowcaseMaximumItemCount},
		{name: "more than nine items", itemCount: HomeHeroVisualShowcaseMaximumItemCount + 1, wantError: ErrHomeVisualTileItemCountInvalid},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := validateHomeVisualTileItemCount(HomeHeroVisualShowcaseTileSetKey, test.itemCount)
			if test.wantError == nil {
				if err != nil {
					t.Fatalf("expected valid item count, got %v", err)
				}
				return
			}
			if !errors.Is(err, test.wantError) {
				t.Fatalf("error = %v, want %v", err, test.wantError)
			}
		})
	}
}

func TestHomeHeroVisualShowcaseRequires600By600ImageDimensions(t *testing.T) {
	tests := []struct {
		name      string
		width     int
		height    int
		wantValid bool
	}{
		{name: "exact dimensions", width: HomeHeroVisualShowcaseImageDimension, height: HomeHeroVisualShowcaseImageDimension, wantValid: true},
		{name: "smaller square", width: HomeHeroVisualShowcaseImageDimension - 1, height: HomeHeroVisualShowcaseImageDimension - 1},
		{name: "non square", width: HomeHeroVisualShowcaseImageDimension, height: HomeHeroVisualShowcaseImageDimension - 1},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := validateHomeVisualTileDimensionsForTileSet(
				HomeHeroVisualShowcaseTileSetKey,
				test.width,
				test.height,
			)
			if test.wantValid {
				if err != nil {
					t.Fatalf("expected valid dimensions, got %v", err)
				}
				return
			}
			if !errors.Is(err, ErrHomeVisualTileImageDimensionsInvalid) {
				t.Fatalf("error = %v, want %v", err, ErrHomeVisualTileImageDimensionsInvalid)
			}
		})
	}
}

func TestHomeHeroVisualShowcaseAlwaysUsesRequestedLocale(t *testing.T) {
	if !shouldUseRequestedHomeVisualTileLocale(
		HomeHeroVisualShowcaseTileSetKey,
		"zh_cn",
		HomeHeroVisualShowcaseMaximumItemCount,
		HomeHeroVisualShowcaseMaximumItemCount,
	) {
		t.Fatal("expected a complete requested locale configuration to be used")
	}
	if !shouldUseRequestedHomeVisualTileLocale(
		HomeHeroVisualShowcaseTileSetKey,
		"zh_cn",
		6,
		6,
	) {
		t.Fatal("expected a partial requested locale configuration to be used")
	}
	if !shouldUseRequestedHomeVisualTileLocale(
		HomeHeroVisualShowcaseTileSetKey,
		"zh_cn",
		0,
		0,
	) {
		t.Fatal("expected an empty requested locale configuration to remain empty")
	}
}
