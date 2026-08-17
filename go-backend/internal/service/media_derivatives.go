package service

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"image"
	_ "image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"math"
	"mime/multipart"
	"path/filepath"
	"strings"

	mediadomain "commerce-platform/internal/domain/media"
	productdomain "commerce-platform/internal/domain/product"
	"commerce-platform/internal/pkg/storage"
	"commerce-platform/internal/repository"

	xdraw "golang.org/x/image/draw"
	_ "golang.org/x/image/webp"
)

const (
	mediaDerivativePrefix             = "media-derivatives"
	mediaDerivativeCacheControl       = "public, max-age=31536000, immutable"
	mediaDerivativeMaxPixels          = 24_000_000
	mediaDerivativeMaxDimension       = 8000
	mediaDerivativeMaxSourceBytes     = 12 << 20
	mediaDerivativeActivePresetLimit  = 12
	mediaDerivativeGenerationCapacity = 2
)

var mediaDerivativeGenerationSlots = make(chan struct{}, mediaDerivativeGenerationCapacity)

// MediaDerivativePresetDefinition is one persistent image conversion emitted
// for every image asset. All upload, backfill, and preflight checks consume
// this same registry.
type MediaDerivativePresetDefinition struct {
	Name              string `json:"name"`
	Label             string `json:"label"`
	MaxWidth          int    `json:"max_width"`
	GenerationVersion int    `json:"generation_version"`
	SortOrder         int    `json:"sort_order"`
	IsSystem          bool   `json:"is_system"`
}

var mediaDerivativePresets = []MediaDerivativePresetDefinition{
	{Name: "thumbnail", Label: "缩略图", MaxWidth: 320, GenerationVersion: 1, SortOrder: 10, IsSystem: true},
	{Name: "card", Label: "卡片图", MaxWidth: 640, GenerationVersion: 1, SortOrder: 20, IsSystem: true},
	{Name: "large", Label: "大图", MaxWidth: 1600, GenerationVersion: 1, SortOrder: 30, IsSystem: true},
}

type mediaDerivativeSourceOpener func() (io.ReadCloser, error)

// MediaDerivativePresetDefinitions returns a copy so API consumers can expose
// the active conversion contract without mutating the generation registry.
func MediaDerivativePresetDefinitions() []MediaDerivativePresetDefinition {
	definitions := make([]MediaDerivativePresetDefinition, len(mediaDerivativePresets))
	copy(definitions, mediaDerivativePresets)
	return definitions
}

func DefaultMediaDerivativePresetModels() []mediadomain.MediaDerivativePreset {
	defaults := MediaDerivativePresetDefinitions()
	models := make([]mediadomain.MediaDerivativePreset, 0, len(defaults))
	for _, preset := range defaults {
		models = append(models, mediadomain.MediaDerivativePreset{
			Code:              preset.Name,
			Label:             preset.Label,
			MaxWidth:          preset.MaxWidth,
			SortOrder:         preset.SortOrder,
			Enabled:           true,
			GenerationVersion: mediaDerivativePresetVersion(preset),
			IsSystem:          preset.IsSystem,
		})
	}
	return models
}

func SeedDefaultMediaDerivativePresets(repo *repository.MediaDerivativePresetRepository) error {
	if repo == nil {
		return nil
	}
	return repo.EnsureSystemDefaults(DefaultMediaDerivativePresetModels())
}

func (s *MediaService) generateAssetDerivatives(ctx context.Context, asset *mediadomain.MediaAsset, file *multipart.FileHeader) ([]mediadomain.MediaAssetDerivative, error) {
	if s == nil || s.storage == nil || asset == nil || file == nil {
		return nil, ErrMediaStorageUnavailable
	}
	// HTTP upload validation provides dimensions before this point. Keeping the
	// guard here preserves legacy service callers that create non-image test
	// fixtures without entering the upload validator.
	if asset.MediaType != "image" || asset.Width <= 0 || asset.Height <= 0 {
		return nil, nil
	}
	presets, err := s.activeMediaDerivativePresetDefinitions()
	if err != nil {
		return nil, err
	}
	return s.generateAssetDerivativesFromSource(ctx, asset, func() (io.ReadCloser, error) {
		source, err := file.Open()
		if err != nil {
			return nil, err
		}
		return source, nil
	}, presets)
}

func (s *MediaService) generateAssetDerivativesFromSource(
	ctx context.Context,
	asset *mediadomain.MediaAsset,
	openSource mediaDerivativeSourceOpener,
	presets []MediaDerivativePresetDefinition,
) ([]mediadomain.MediaAssetDerivative, error) {
	if s == nil || s.storage == nil || asset == nil || openSource == nil {
		return nil, ErrMediaStorageUnavailable
	}
	if asset.MediaType != "image" || len(presets) == 0 {
		return nil, nil
	}

	release, err := acquireMediaDerivativeGenerationSlot(ctx)
	if err != nil {
		return nil, err
	}
	defer release()

	decoded, err := decodeDerivativeSource(openSource)
	if err != nil {
		return nil, err
	}

	preserveAlpha := imageHasAlpha(decoded)
	derivatives := make([]mediadomain.MediaAssetDerivative, 0, len(presets))
	uploadedURLs := make([]string, 0, len(presets))
	for _, preset := range presets {
		if strings.TrimSpace(preset.Name) == "" || preset.MaxWidth <= 0 {
			continue
		}
		resized := resizeImageToMaxWidth(decoded, preset.MaxWidth)
		payload, filename, mimeType, width, height, err := encodeDerivativeImage(asset, resized, preset, preserveAlpha)
		if err != nil {
			deleteUploadedMediaObjectsBestEffort(ctx, s.storage, uploadedURLs)
			return nil, err
		}

		url, err := uploadDerivativeObject(ctx, s.storage, bytes.NewReader(payload), filename, derivativePrefixForAsset(asset.ID, preset.Name, mediaDerivativePresetVersion(preset)))
		if err != nil {
			deleteUploadedMediaObjectsBestEffort(ctx, s.storage, uploadedURLs)
			return nil, fmt.Errorf("%w: upload %s derivative: %v", ErrMediaDerivativeGenerationFailed, preset.Name, err)
		}
		uploadedURLs = append(uploadedURLs, url)

		derivatives = append(derivatives, mediadomain.MediaAssetDerivative{
			MediaAssetID:  asset.ID,
			Preset:        preset.Name,
			PresetVersion: mediaDerivativePresetVersion(preset),
			URL:           url,
			StorageKey:    storageObjectKey(s.storage, url),
			MimeType:      mimeType,
			Width:         width,
			Height:        height,
			Size:          int64(len(payload)),
		})
	}

	return derivatives, nil
}

func decodeDerivativeSource(openSource mediaDerivativeSourceOpener) (image.Image, error) {
	source, err := openSource()
	if err != nil {
		return nil, fmt.Errorf("%w: open source image: %v", ErrMediaDerivativeGenerationFailed, err)
	}
	payload, readErr := io.ReadAll(io.LimitReader(source, mediaDerivativeMaxSourceBytes+1))
	closeSourceErr := source.Close()
	if readErr != nil {
		return nil, fmt.Errorf("%w: read source image: %v", ErrMediaDerivativeGenerationFailed, readErr)
	}
	if len(payload) > mediaDerivativeMaxSourceBytes {
		return nil, fmt.Errorf(
			"%w: source image exceeds %d bytes",
			ErrMediaDerivativeGenerationFailed,
			mediaDerivativeMaxSourceBytes,
		)
	}
	if closeSourceErr != nil {
		return nil, fmt.Errorf("%w: close source image: %v", ErrMediaDerivativeGenerationFailed, closeSourceErr)
	}

	config, _, decodeConfigErr := image.DecodeConfig(bytes.NewReader(payload))
	if decodeConfigErr != nil {
		return nil, fmt.Errorf("%w: inspect source image: %v", ErrMediaDerivativeGenerationFailed, decodeConfigErr)
	}
	if err := validateMediaDerivativeSourceDimensions(config.Width, config.Height); err != nil {
		return nil, err
	}

	decoded, decodedFormat, decodeErr := image.Decode(bytes.NewReader(payload))
	if decodeErr != nil {
		return nil, fmt.Errorf("%w: decode source image: %v", ErrMediaDerivativeGenerationFailed, decodeErr)
	}
	if decodedFormat == "webp" && isAnimatedWebPData(payload) {
		return nil, fmt.Errorf("%w: animated WebP is not supported for generated media", ErrMediaDerivativeGenerationFailed)
	}
	if decodedFormat == "jpeg" {
		decoded = applyJPEGExifOrientation(decoded, payload)
	}

	bounds := decoded.Bounds()
	if err := validateMediaDerivativeSourceDimensions(bounds.Dx(), bounds.Dy()); err != nil {
		return nil, err
	}
	return decoded, nil
}

func isAnimatedWebPData(data []byte) bool {
	return len(data) >= 21 &&
		string(data[0:4]) == "RIFF" &&
		string(data[8:12]) == "WEBP" &&
		string(data[12:16]) == "VP8X" &&
		data[20]&0x02 != 0
}

func applyJPEGExifOrientation(src image.Image, data []byte) image.Image {
	orientation := jpegExifOrientation(data)
	if src == nil || orientation < 2 || orientation > 8 {
		return src
	}

	bounds := src.Bounds()
	width, height := bounds.Dx(), bounds.Dy()
	if width <= 0 || height <= 0 {
		return src
	}
	dstWidth, dstHeight := width, height
	if orientation >= 5 {
		dstWidth, dstHeight = height, width
	}
	dst := image.NewNRGBA(image.Rect(0, 0, dstWidth, dstHeight))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			dstX, dstY := x, y
			switch orientation {
			case 2:
				dstX = width - 1 - x
			case 3:
				dstX, dstY = width-1-x, height-1-y
			case 4:
				dstY = height - 1 - y
			case 5:
				dstX, dstY = y, x
			case 6:
				dstX, dstY = height-1-y, x
			case 7:
				dstX, dstY = height-1-y, width-1-x
			case 8:
				dstX, dstY = y, width-1-x
			}
			dst.Set(dstX, dstY, src.At(bounds.Min.X+x, bounds.Min.Y+y))
		}
	}
	return dst
}

func jpegExifOrientation(data []byte) int {
	if len(data) < 4 || data[0] != 0xff || data[1] != 0xd8 {
		return 1
	}
	for offset := 2; offset+4 <= len(data); {
		if data[offset] != 0xff {
			offset++
			continue
		}
		for offset < len(data) && data[offset] == 0xff {
			offset++
		}
		if offset >= len(data) {
			break
		}
		marker := data[offset]
		offset++
		if marker == 0xd9 || marker == 0xda {
			break
		}
		if marker == 0x01 || (marker >= 0xd0 && marker <= 0xd7) {
			continue
		}
		if offset+2 > len(data) {
			break
		}
		segmentLength := int(binary.BigEndian.Uint16(data[offset : offset+2]))
		if segmentLength < 2 || offset+segmentLength > len(data) {
			break
		}
		if marker == 0xe1 && segmentLength >= 10 &&
			bytes.Equal(data[offset+2:offset+8], []byte("Exif\x00\x00")) {
			return tiffExifOrientation(data[offset+8 : offset+segmentLength])
		}
		offset += segmentLength
	}
	return 1
}

func tiffExifOrientation(data []byte) int {
	if len(data) < 8 {
		return 1
	}
	littleEndian := bytes.Equal(data[0:2], []byte("II"))
	if !littleEndian && !bytes.Equal(data[0:2], []byte("MM")) {
		return 1
	}
	readUint16 := func(offset int) (uint16, bool) {
		if offset < 0 || offset+2 > len(data) {
			return 0, false
		}
		if littleEndian {
			return binary.LittleEndian.Uint16(data[offset : offset+2]), true
		}
		return binary.BigEndian.Uint16(data[offset : offset+2]), true
	}
	readUint32 := func(offset int) (uint32, bool) {
		if offset < 0 || offset+4 > len(data) {
			return 0, false
		}
		if littleEndian {
			return binary.LittleEndian.Uint32(data[offset : offset+4]), true
		}
		return binary.BigEndian.Uint32(data[offset : offset+4]), true
	}

	magic, ok := readUint16(2)
	if !ok || magic != 42 {
		return 1
	}
	ifdOffset, ok := readUint32(4)
	if !ok || ifdOffset > uint32(len(data)) {
		return 1
	}
	entryCount, ok := readUint16(int(ifdOffset))
	if !ok {
		return 1
	}
	for index := 0; index < int(entryCount); index++ {
		entryOffset := int(ifdOffset) + 2 + index*12
		if entryOffset+12 > len(data) {
			break
		}
		tag, ok := readUint16(entryOffset)
		if !ok || tag != 0x0112 {
			continue
		}
		valueType, ok := readUint16(entryOffset + 2)
		count, countOK := readUint32(entryOffset + 4)
		if !ok || !countOK || valueType != 3 || count != 1 {
			return 1
		}
		value, ok := readUint16(entryOffset + 8)
		if !ok || value < 1 || value > 8 {
			return 1
		}
		return int(value)
	}
	return 1
}

func acquireMediaDerivativeGenerationSlot(ctx context.Context) (func(), error) {
	if ctx == nil {
		ctx = context.Background()
	}
	select {
	case mediaDerivativeGenerationSlots <- struct{}{}:
		return func() {
			<-mediaDerivativeGenerationSlots
		}, nil
	case <-ctx.Done():
		return nil, fmt.Errorf("%w: derivative generation canceled: %v", ErrMediaDerivativeGenerationFailed, ctx.Err())
	}
}

func validateMediaDerivativeSourceDimensions(width, height int) error {
	if width <= 0 || height <= 0 {
		return fmt.Errorf("%w: source image dimensions are invalid", ErrMediaDerivativeGenerationFailed)
	}
	if width > mediaDerivativeMaxDimension || height > mediaDerivativeMaxDimension {
		return fmt.Errorf(
			"%w: source image exceeds %dx%d dimensions",
			ErrMediaDerivativeGenerationFailed,
			mediaDerivativeMaxDimension,
			mediaDerivativeMaxDimension,
		)
	}
	if int64(width)*int64(height) > mediaDerivativeMaxPixels {
		return fmt.Errorf("%w: source image exceeds %d pixels", ErrMediaDerivativeGenerationFailed, mediaDerivativeMaxPixels)
	}
	return nil
}

func resizeImageToMaxWidth(src image.Image, maxWidth int) image.Image {
	bounds := src.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()
	if width <= 0 || height <= 0 || maxWidth <= 0 || width <= maxWidth {
		return src
	}

	nextWidth := maxWidth
	nextHeight := int(math.Round(float64(height) * float64(nextWidth) / float64(width)))
	if nextHeight < 1 {
		nextHeight = 1
	}
	dst := image.NewNRGBA(image.Rect(0, 0, nextWidth, nextHeight))
	xdraw.CatmullRom.Scale(dst, dst.Bounds(), src, bounds, xdraw.Over, nil)
	return dst
}

func encodeDerivativeImage(
	asset *mediadomain.MediaAsset,
	img image.Image,
	preset MediaDerivativePresetDefinition,
	preserveAlpha bool,
) ([]byte, string, string, int, int, error) {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()
	if width <= 0 || height <= 0 {
		return nil, "", "", 0, 0, fmt.Errorf("%w: %s derivative has invalid dimensions", ErrMediaDerivativeGenerationFailed, preset.Name)
	}

	var buffer bytes.Buffer
	ext := ".jpg"
	mimeType := "image/jpeg"
	if preserveAlpha {
		ext = ".png"
		mimeType = "image/png"
		if err := png.Encode(&buffer, img); err != nil {
			return nil, "", "", 0, 0, fmt.Errorf("%w: encode %s PNG derivative: %v", ErrMediaDerivativeGenerationFailed, preset.Name, err)
		}
	} else if err := jpeg.Encode(&buffer, img, &jpeg.Options{Quality: 84}); err != nil {
		return nil, "", "", 0, 0, fmt.Errorf("%w: encode %s JPEG derivative: %v", ErrMediaDerivativeGenerationFailed, preset.Name, err)
	}

	filename := derivativeFilename(asset, preset.Name, width, ext)
	return buffer.Bytes(), filename, mimeType, width, height, nil
}

func imageHasAlpha(img image.Image) bool {
	switch img.(type) {
	case *image.YCbCr, *image.Gray, *image.Gray16, *image.CMYK:
		return false
	}
	if opaque, ok := img.(interface{ Opaque() bool }); ok {
		return !opaque.Opaque()
	}
	bounds := img.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			_, _, _, alpha := img.At(x, y).RGBA()
			if alpha != 0xffff {
				return true
			}
		}
	}
	return false
}

func uploadDerivativeObject(ctx context.Context, storageService storage.StorageService, reader *bytes.Reader, filename string, prefix string) (string, error) {
	if uploader, ok := storageService.(storage.CacheControlledReaderUploader); ok {
		return uploader.UploadFromReaderWithPrefixAndCacheControl(ctx, reader, filename, prefix, mediaDerivativeCacheControl)
	}
	return storageService.UploadFromReaderWithPrefix(ctx, reader, filename, prefix)
}

func storageObjectKey(storageService storage.StorageService, reference string) string {
	if storageService != nil {
		if key, err := storageService.ObjectKey(reference); err == nil && key != "" {
			return key
		}
	}
	return extractStorageKey(reference)
}

func derivativePrefixForAsset(assetID uint, preset string, version int) string {
	return fmt.Sprintf("%s/%d/%s/v%d", mediaDerivativePrefix, assetID, preset, normalizeMediaDerivativeVersion(version))
}

func derivativeFilename(asset *mediadomain.MediaAsset, preset string, width int, ext string) string {
	base := strings.TrimSuffix(filepath.Base(strings.TrimSpace(asset.Filename)), filepath.Ext(asset.Filename))
	if base == "" {
		base = fmt.Sprintf("asset-%d", asset.ID)
	}
	return fmt.Sprintf("%s-%s-%d%s", base, preset, width, ext)
}

func deleteUploadedMediaObjectsBestEffort(ctx context.Context, storageService storage.StorageService, urls []string) {
	if storageService == nil {
		return
	}
	for _, url := range urls {
		if strings.TrimSpace(url) == "" {
			continue
		}
		_ = storageService.Delete(ctx, url)
	}
}

func IsMediaDerivativeStorageKey(key string) bool {
	key = strings.Trim(strings.ReplaceAll(key, "\\", "/"), "/")
	return strings.HasPrefix(key, mediaDerivativePrefix+"/")
}

func (s *MediaService) activeMediaDerivativePresetDefinitions() ([]MediaDerivativePresetDefinition, error) {
	if s == nil || s.derivativePresets == nil {
		return MediaDerivativePresetDefinitions(), nil
	}

	presets, err := s.derivativePresets.ListActive()
	if err != nil {
		return nil, err
	}
	if presets == nil {
		return MediaDerivativePresetDefinitions(), nil
	}
	if len(presets) > mediaDerivativeActivePresetLimit {
		return nil, ErrMediaDerivativePresetLimitReached
	}

	definitions := make([]MediaDerivativePresetDefinition, 0, len(presets))
	for _, preset := range presets {
		definition := MediaDerivativePresetDefinition{
			Name:              strings.TrimSpace(preset.Code),
			Label:             strings.TrimSpace(preset.Label),
			MaxWidth:          preset.MaxWidth,
			GenerationVersion: normalizeMediaDerivativeVersion(preset.GenerationVersion),
			SortOrder:         preset.SortOrder,
			IsSystem:          preset.IsSystem,
		}
		if definition.Name == "" || definition.MaxWidth <= 0 {
			return nil, ErrInvalidMediaDerivativePreset
		}
		if definition.Label == "" {
			definition.Label = definition.Name
		}
		definitions = append(definitions, definition)
	}
	return definitions, nil
}

func missingMediaDerivativePresets(existing []mediadomain.MediaAssetDerivative, presets []MediaDerivativePresetDefinition) []MediaDerivativePresetDefinition {
	existingPresets := make(map[string]map[int]struct{}, len(existing))
	for _, derivative := range existing {
		if strings.TrimSpace(derivative.URL) == "" {
			continue
		}
		preset := strings.TrimSpace(derivative.Preset)
		if preset == "" {
			continue
		}
		if existingPresets[preset] == nil {
			existingPresets[preset] = map[int]struct{}{}
		}
		existingPresets[preset][normalizeMediaDerivativeVersion(derivative.PresetVersion)] = struct{}{}
	}

	missing := make([]MediaDerivativePresetDefinition, 0, len(presets))
	for _, preset := range presets {
		if _, ok := existingPresets[preset.Name][mediaDerivativePresetVersion(preset)]; !ok {
			missing = append(missing, preset)
		}
	}
	return missing
}

func replaceMediaAssetDerivatives(existing []mediadomain.MediaAssetDerivative, replacements []mediadomain.MediaAssetDerivative) []mediadomain.MediaAssetDerivative {
	if len(replacements) == 0 {
		return existing
	}
	replacedPresets := make(map[string]struct{}, len(replacements))
	for _, derivative := range replacements {
		preset := strings.TrimSpace(derivative.Preset)
		if preset != "" {
			replacedPresets[preset] = struct{}{}
		}
	}

	next := make([]mediadomain.MediaAssetDerivative, 0, len(existing)+len(replacements))
	for _, derivative := range existing {
		if _, ok := replacedPresets[strings.TrimSpace(derivative.Preset)]; ok {
			continue
		}
		next = append(next, derivative)
	}
	return append(next, replacements...)
}

func (s *MediaService) ProductMediaImageVariants(assetID uint) (map[string]productdomain.ProductMediaImageVariant, string, error) {
	if s == nil || s.repo == nil || assetID == 0 {
		return nil, "", nil
	}
	presets, err := s.activeMediaDerivativePresetDefinitions()
	if err != nil {
		return nil, "", err
	}
	required := mediaDerivativePresetVersionMap(presets)
	derivatives, err := s.repo.FindAssetDerivatives(assetID)
	if err != nil {
		return nil, "", err
	}

	variants := make(map[string]productdomain.ProductMediaImageVariant, len(derivatives))
	for _, derivative := range derivatives {
		preset := strings.TrimSpace(derivative.Preset)
		if preset == "" || strings.TrimSpace(derivative.URL) == "" {
			continue
		}
		if required[preset] != normalizeMediaDerivativeVersion(derivative.PresetVersion) {
			continue
		}
		variants[preset] = productdomain.ProductMediaImageVariant{
			URL:      strings.TrimSpace(derivative.URL),
			Width:    derivative.Width,
			Height:   derivative.Height,
			MimeType: strings.TrimSpace(derivative.MimeType),
		}
	}

	thumbnail := preferredProductMediaThumbnail(variants)
	return variants, thumbnail, nil
}

func mediaDerivativePresetVersionMap(presets []MediaDerivativePresetDefinition) map[string]int {
	versions := make(map[string]int, len(presets))
	for _, preset := range presets {
		name := strings.TrimSpace(preset.Name)
		if name == "" {
			continue
		}
		versions[name] = mediaDerivativePresetVersion(preset)
	}
	return versions
}

func mediaDerivativePresetRequirements(presets []MediaDerivativePresetDefinition) []repository.MediaDerivativeRequirement {
	requirements := make([]repository.MediaDerivativeRequirement, 0, len(presets))
	for _, preset := range presets {
		name := strings.TrimSpace(preset.Name)
		if name == "" {
			continue
		}
		requirements = append(requirements, repository.MediaDerivativeRequirement{
			Preset:        name,
			PresetVersion: mediaDerivativePresetVersion(preset),
		})
	}
	return requirements
}

func mediaDerivativePresetVersion(preset MediaDerivativePresetDefinition) int {
	return normalizeMediaDerivativeVersion(preset.GenerationVersion)
}

func normalizeMediaDerivativeVersion(version int) int {
	if version <= 0 {
		return 1
	}
	return version
}

func preferredProductMediaThumbnail(variants map[string]productdomain.ProductMediaImageVariant) string {
	for _, preset := range []string{"thumbnail", "card", "large"} {
		if item, ok := variants[preset]; ok && strings.TrimSpace(item.URL) != "" {
			return strings.TrimSpace(item.URL)
		}
	}
	return ""
}
