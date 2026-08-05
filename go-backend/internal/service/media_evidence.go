package service

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	"tanzanite/internal/domain/media"
	"tanzanite/internal/pkg/storage"
)

var (
	ErrMediaEvidenceUnavailable       = errors.New("media copyright evidence unavailable")
	ErrMediaEvidenceIntegrityMismatch = errors.New("media copyright evidence integrity mismatch")
)

type CopyrightEvidenceManifest struct {
	Format       string                        `json:"format"`
	GeneratedAt  time.Time                     `json:"generated_at"`
	Asset        CopyrightEvidenceAsset        `json:"asset"`
	Copyright    CopyrightEvidenceClaim        `json:"copyright_claim"`
	Verification CopyrightEvidenceVerification `json:"verification"`
}

type CopyrightEvidenceAsset struct {
	ID               uint      `json:"id"`
	OriginalFilename string    `json:"original_filename"`
	MimeType         string    `json:"mime_type"`
	MediaType        string    `json:"media_type"`
	Size             int64     `json:"size"`
	SourceURL        string    `json:"source_url"`
	RecordedAt       time.Time `json:"recorded_at"`
	UploaderID       uint      `json:"uploader_id"`
}

type CopyrightEvidenceVerification struct {
	Algorithm        string `json:"algorithm"`
	OriginalSHA256   string `json:"original_sha256"`
	StoredSHA256     string `json:"stored_sha256"`
	IntegrityChecked bool   `json:"integrity_checked"`
}

type CopyrightEvidenceClaim struct {
	SiteName           string    `json:"site_name"`
	SiteURL            string    `json:"site_url"`
	SiteDomain         string    `json:"site_domain"`
	RightsHolder       string    `json:"rights_holder"`
	CopyrightNotice    string    `json:"copyright_notice"`
	CopyrightPolicyURL string    `json:"copyright_policy_url"`
	OriginalFilename   string    `json:"original_filename"`
	OriginalSHA256     string    `json:"original_sha256"`
	UploaderID         uint      `json:"uploader_id"`
	ServerReceivedAt   time.Time `json:"server_received_at"`
}

func (s *MediaService) buildCopyrightClaim(input MediaUploadInput, contentSHA256 string) (string, error) {
	claim := CopyrightEvidenceClaim{
		OriginalFilename: input.File.Filename,
		OriginalSHA256:   contentSHA256,
		UploaderID:       input.UploaderID,
		ServerReceivedAt: time.Now().UTC(),
	}

	if s.settings != nil {
		claim.SiteName = settingValue(s.settings, "site_name")
		claim.SiteURL = settingValue(s.settings, "site_url")
		claim.RightsHolder = settingValue(s.settings, "copyright_holder")
		claim.CopyrightNotice = settingValue(s.settings, "copyright_notice")
		claim.CopyrightPolicyURL = settingValue(s.settings, "copyright_url")
	}

	claim.SiteDomain = siteDomainFromURL(claim.SiteURL)
	if strings.TrimSpace(claim.RightsHolder) == "" {
		claim.RightsHolder = claim.SiteName
	}
	if strings.TrimSpace(claim.CopyrightNotice) == "" && strings.TrimSpace(claim.RightsHolder) != "" {
		claim.CopyrightNotice = fmt.Sprintf(
			"Copyright %d %s. All rights reserved.",
			claim.ServerReceivedAt.Year(),
			claim.RightsHolder,
		)
	}

	encoded, err := json.Marshal(claim)
	if err != nil {
		return "", fmt.Errorf("encode copyright claim: %w", err)
	}
	return string(encoded), nil
}

func settingValue(settings *SettingService, key string) string {
	if settings == nil {
		return ""
	}
	item, err := settings.Get(key, "en")
	if err != nil || item == nil {
		return ""
	}
	return strings.TrimSpace(item.Value)
}

func siteDomainFromURL(rawURL string) string {
	parsed, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil {
		return ""
	}
	return strings.TrimSpace(parsed.Hostname())
}

func contentSHA256FromMultipartFile(file *multipart.FileHeader) (string, error) {
	if file == nil {
		return "", ErrMediaUploadFileRequired
	}

	reader, err := file.Open()
	if err != nil {
		return "", fmt.Errorf("open upload for hashing: %w", err)
	}
	defer func() { _ = reader.Close() }()

	hash := sha256.New()
	if _, err := io.Copy(hash, reader); err != nil {
		return "", fmt.Errorf("hash upload: %w", err)
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

func (s *MediaService) ExportCopyrightEvidence(ctx context.Context, id uint) ([]byte, string, error) {
	asset, err := s.GetAsset(id)
	if err != nil {
		return nil, "", err
	}

	reader, size, err := s.openEvidenceReader(ctx, asset)
	if err != nil {
		return nil, "", err
	}
	defer func() { _ = reader.Close() }()

	var original bytes.Buffer
	hash := sha256.New()
	if _, err := io.Copy(io.MultiWriter(&original, hash), reader); err != nil {
		return nil, "", fmt.Errorf("%w: read original object: %v", ErrMediaEvidenceUnavailable, err)
	}

	computedSHA256 := hex.EncodeToString(hash.Sum(nil))
	if asset.ContentSHA256 != "" && !strings.EqualFold(asset.ContentSHA256, computedSHA256) {
		return nil, "", ErrMediaEvidenceIntegrityMismatch
	}
	if size >= 0 && int64(original.Len()) != size {
		return nil, "", ErrMediaEvidenceIntegrityMismatch
	}

	originalName := safeEvidenceFilename(asset.OriginalFilename, asset.Filename)
	originalPath := "original/" + originalName
	manifest := CopyrightEvidenceManifest{
		Format:      "copyright-evidence/v1",
		GeneratedAt: time.Now().UTC(),
		Asset: CopyrightEvidenceAsset{
			ID:               asset.ID,
			OriginalFilename: asset.OriginalFilename,
			MimeType:         asset.MimeType,
			MediaType:        asset.MediaType,
			Size:             int64(original.Len()),
			SourceURL:        asset.URL,
			RecordedAt:       asset.CreatedAt.UTC(),
			UploaderID:       asset.UploaderID,
		},
		Copyright: CopyrightEvidenceClaim{
			OriginalFilename: asset.OriginalFilename,
			OriginalSHA256:   computedSHA256,
			UploaderID:       asset.UploaderID,
			ServerReceivedAt: asset.CreatedAt.UTC(),
		},
		Verification: CopyrightEvidenceVerification{
			Algorithm:        "SHA-256",
			OriginalSHA256:   computedSHA256,
			StoredSHA256:     asset.ContentSHA256,
			IntegrityChecked: true,
		},
	}
	if strings.TrimSpace(asset.CopyrightClaimJSON) != "" {
		if err := json.Unmarshal([]byte(asset.CopyrightClaimJSON), &manifest.Copyright); err != nil {
			return nil, "", fmt.Errorf("%w: decode copyright claim: %v", ErrMediaEvidenceUnavailable, err)
		}
		manifest.Copyright.OriginalSHA256 = computedSHA256
	}
	manifestBytes, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return nil, "", fmt.Errorf("%w: encode manifest: %v", ErrMediaEvidenceUnavailable, err)
	}

	manifestSHA256 := sha256.Sum256(manifestBytes)
	var archive bytes.Buffer
	archiveWriter := zip.NewWriter(&archive)

	manifestWriter, err := archiveWriter.Create("manifest.json")
	if err != nil {
		return nil, "", fmt.Errorf("%w: create manifest: %v", ErrMediaEvidenceUnavailable, err)
	}
	if _, err := manifestWriter.Write(manifestBytes); err != nil {
		return nil, "", fmt.Errorf("%w: write manifest: %v", ErrMediaEvidenceUnavailable, err)
	}

	originalWriter, err := archiveWriter.Create(originalPath)
	if err != nil {
		return nil, "", fmt.Errorf("%w: create original: %v", ErrMediaEvidenceUnavailable, err)
	}
	if _, err := io.Copy(originalWriter, &original); err != nil {
		return nil, "", fmt.Errorf("%w: write original: %v", ErrMediaEvidenceUnavailable, err)
	}

	checksumWriter, err := archiveWriter.Create("SHA256SUMS.txt")
	if err != nil {
		return nil, "", fmt.Errorf("%w: create checksums: %v", ErrMediaEvidenceUnavailable, err)
	}
	checksums := fmt.Sprintf(
		"%s  %s\n%s  manifest.json\n",
		computedSHA256,
		originalPath,
		hex.EncodeToString(manifestSHA256[:]),
	)
	if _, err := checksumWriter.Write([]byte(checksums)); err != nil {
		return nil, "", fmt.Errorf("%w: write checksums: %v", ErrMediaEvidenceUnavailable, err)
	}

	readmeWriter, err := archiveWriter.Create("README.txt")
	if err != nil {
		return nil, "", fmt.Errorf("%w: create readme: %v", ErrMediaEvidenceUnavailable, err)
	}
	readme := "This package preserves the uploaded source object and its server-recorded hash. " +
		"It is supporting evidence for a copyright review, not a legal determination.\n"
	if _, err := readmeWriter.Write([]byte(readme)); err != nil {
		return nil, "", fmt.Errorf("%w: write readme: %v", ErrMediaEvidenceUnavailable, err)
	}

	if err := archiveWriter.Close(); err != nil {
		return nil, "", fmt.Errorf("%w: close archive: %v", ErrMediaEvidenceUnavailable, err)
	}
	return archive.Bytes(), fmt.Sprintf("copyright-evidence-%d.zip", asset.ID), nil
}

func (s *MediaService) openEvidenceReader(ctx context.Context, asset *media.MediaAsset) (io.ReadCloser, int64, error) {
	key := assetStorageKey(asset)
	if key == "" {
		return nil, 0, ErrMediaEvidenceUnavailable
	}

	if opener, ok := s.storage.(storage.ObjectOpener); ok {
		object, err := opener.Open(ctx, key)
		if err != nil {
			return nil, 0, fmt.Errorf("%w: open source object: %v", ErrMediaEvidenceUnavailable, err)
		}
		return object.ReadCloser, object.Size, nil
	}

	signer, ok := s.storage.(storage.PresignedURLProvider)
	if !ok {
		return nil, 0, ErrMediaEvidenceUnavailable
	}
	url, err := signer.GetPresignedURL(ctx, key, protectedMediaSignedURLTTL)
	if err != nil {
		return nil, 0, fmt.Errorf("%w: sign source object: %v", ErrMediaEvidenceUnavailable, err)
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, 0, fmt.Errorf("%w: create source request: %v", ErrMediaEvidenceUnavailable, err)
	}
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, 0, fmt.Errorf("%w: fetch source object: %v", ErrMediaEvidenceUnavailable, err)
	}
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		_ = response.Body.Close()
		return nil, 0, fmt.Errorf("%w: source object returned HTTP %d", ErrMediaEvidenceUnavailable, response.StatusCode)
	}
	return response.Body, response.ContentLength, nil
}

func safeEvidenceFilename(values ...string) string {
	for _, value := range values {
		name := path.Base(strings.ReplaceAll(strings.TrimSpace(value), "\\", "/"))
		if name == "" || name == "." || name == ".." {
			continue
		}
		name = strings.Map(func(r rune) rune {
			switch r {
			case 0, '/', '\\', ':':
				return '_'
			default:
				return r
			}
		}, name)
		if name != "" && name != "." && name != ".." {
			return name
		}
	}
	return "original.bin"
}
