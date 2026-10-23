package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"commerce-platform/internal/domain/setting"
)

const SettingsConfigurationBackupFormat = "tanzanite-site-configuration"
const SettingsConfigurationBackupVersion = 1
const MaximumSettingsConfigurationFileBytes = 8 << 20
const maximumSettingsConfigurationEntries = 10000

var ErrSettingsConfigurationPreviewChanged = errors.New("配置已在预览后变化，请重新预览后再导入")
var configurationBackupLocalePattern = regexp.MustCompile(`^(global|[A-Za-z]{2,3}(?:-[A-Za-z0-9]{2,8})*)$`)

type SettingsConfigurationBackupFile struct {
	Format               string                         `json:"format"`
	Version              int                            `json:"version"`
	ExportedAt           time.Time                      `json:"exported_at"`
	Settings             []setting.UpdateSettingRequest `json:"settings"`
	ExcludedSettingCount int                            `json:"excluded_setting_count"`
}

type SettingsConfigurationBackupScope struct {
	Group string `json:"group"`
	Label string `json:"label"`
	Count int    `json:"count"`
}

type SettingsConfigurationImportChange struct {
	Key                 string `json:"key"`
	Group               string `json:"group"`
	Locale              string `json:"locale"`
	Action              string `json:"action"`
	CurrentValue        string `json:"current_value"`
	ImportedValue       string `json:"imported_value"`
	CurrentType         string `json:"current_type"`
	ImportedType        string `json:"imported_type"`
	CurrentIsPublic     bool   `json:"current_is_public"`
	ImportedIsPublic    bool   `json:"imported_is_public"`
	CurrentDescription  string `json:"current_description"`
	ImportedDescription string `json:"imported_description"`
}

type SettingsConfigurationImportPreview struct {
	BackupSHA256   string                              `json:"backup_sha256"`
	TargetSHA256   string                              `json:"target_sha256"`
	CreateCount    int                                 `json:"create_count"`
	UpdateCount    int                                 `json:"update_count"`
	UnchangedCount int                                 `json:"unchanged_count"`
	Changes        []SettingsConfigurationImportChange `json:"changes"`
}

// An explicit portable-key list prevents newly added secrets or domain-owned
// configuration from silently becoming downloadable or writable by this route.
func portableSettingsConfigurationKeyTypes() map[string]map[string]string {
	groups := map[string]map[string]string{}
	registerStrings := func(group, keys string) {
		groups[group] = map[string]string{}
		for _, key := range strings.Fields(keys) {
			groups[group][key] = "string"
		}
	}
	registerStrings("site", "site_name site_description site_logo contact_email contact_phone admin_brand_name admin_brand_initial admin_panel_label admin_login_title admin_footer_text admin_html_title")
	registerStrings("social", "facebook instagram x youtube reddit")
	registerStrings("email", "smtp_host smtp_username from_email from_name")
	groups["email"]["smtp_port"] = "number"
	registerStrings("api", "time_api_provider time_api_endpoint time_api_query_template time_api_default_timezone customs_lookup_us_hts_endpoint customs_lookup_us_hts_api_key_header customs_lookup_uk_trade_tariff_endpoint customs_lookup_uk_trade_tariff_api_key_header")
	for _, key := range []string{"time_api_enabled", "customs_lookup_us_hts_enabled", "customs_lookup_uk_trade_tariff_enabled"} {
		groups["api"][key] = "boolean"
	}
	groups["api"]["time_api_refresh_minutes"] = "number"
	registerStrings(setting.WebsiteProfileGroup, "eyebrow title lead scope avatar_url avatar_label avatar_mark profile_label profile_role profile_context statement_eyebrow statement_title statement_body statement_paragraph_1 statement_paragraph_2 factory_image_url factory_image_alt factory_image_caption factory_eyebrow factory_title factory_body factory_cta factory_link")
	return groups
}

func (s *AdminSettingsService) ExportPortableSettingsConfiguration(ctx context.Context) (*SettingsConfigurationBackupFile, []SettingsConfigurationBackupScope, error) {
	rows, err := s.settings.settingRepo.ListAllLocalesForSettingsConfigurationBackup(ctx, maximumSettingsConfigurationEntries+1)
	if err != nil {
		return nil, nil, err
	}
	if len(rows) > maximumSettingsConfigurationEntries {
		return nil, nil, fmt.Errorf("%w: 配置数量超过导出上限", ErrSettingInvalid)
	}
	file := &SettingsConfigurationBackupFile{Format: SettingsConfigurationBackupFormat, Version: SettingsConfigurationBackupVersion, ExportedAt: time.Now().UTC(), Settings: []setting.UpdateSettingRequest{}}
	allowed := portableSettingsConfigurationKeyTypes()
	counts := map[string]int{}
	for _, row := range rows {
		expectedType, ok := allowed[row.Group][row.Key]
		if !ok || IsDomainManagedSettingGroup(row.Group) || IsDomainManagedSettingKey(row.Key) || isSensitiveSettingKey(row.Key) {
			file.ExcludedSettingCount++
			continue
		}
		entry := settingsConfigurationTransferEntryFromStoredSetting(row)
		entry.Type = expectedType
		if entry.Group == "api" || entry.Group == "email" {
			entry.IsPublic = false
		}
		// Legacy invalid values remain in the database; exclude them instead of
		// producing a downloaded file that cannot pass the import contract.
		if _, err := validateAndSortPortableSettingsConfigurationFile(SettingsConfigurationBackupFile{Format: SettingsConfigurationBackupFormat, Version: SettingsConfigurationBackupVersion, Settings: []setting.UpdateSettingRequest{entry}}); err != nil {
			file.ExcludedSettingCount++
			continue
		}
		file.Settings = append(file.Settings, entry)
		counts[row.Group]++
	}
	encoded, err := json.Marshal(file)
	if err != nil {
		return nil, nil, err
	}
	if len(encoded) > MaximumSettingsConfigurationFileBytes {
		return nil, nil, fmt.Errorf("%w: 配置文件超过 8 MiB 上限", ErrSettingInvalid)
	}
	scopes := []SettingsConfigurationBackupScope{}
	for _, scope := range []SettingsConfigurationBackupScope{{Group: "site", Label: "站点资料与后台品牌"}, {Group: "social", Label: "社交链接"}, {Group: "website_profile", Label: "网站介绍"}, {Group: "api", Label: "API 管理普通配置"}, {Group: "email", Label: "旧版邮件普通配置"}} {
		scope.Count = counts[scope.Group]
		scopes = append(scopes, scope)
	}
	return file, scopes, nil
}

func settingsConfigurationTransferEntryFromStoredSetting(row setting.Setting) setting.UpdateSettingRequest {
	return setting.UpdateSettingRequest{Key: row.Key, Value: row.Value, Type: row.Type, Group: row.Group, Locale: row.Locale, IsPublic: row.IsPublic, Description: row.Description}
}

func validateAndSortPortableSettingsConfigurationFile(file SettingsConfigurationBackupFile) ([]setting.UpdateSettingRequest, error) {
	invalid := func(message string) ([]setting.UpdateSettingRequest, error) {
		return nil, fmt.Errorf("%w: %s", ErrSettingInvalid, message)
	}
	if file.Format != SettingsConfigurationBackupFormat || file.Version != SettingsConfigurationBackupVersion {
		return invalid("配置文件格式或版本不兼容")
	}
	if len(file.Settings) == 0 || len(file.Settings) > maximumSettingsConfigurationEntries {
		return invalid("配置文件必须包含 1–10000 项配置")
	}
	encoded, err := json.Marshal(file)
	if err != nil || len(encoded) > MaximumSettingsConfigurationFileBytes {
		return invalid("配置文件超过大小上限")
	}
	allowed := portableSettingsConfigurationKeyTypes()
	seen := map[string]bool{}
	entries := append([]setting.UpdateSettingRequest(nil), file.Settings...)
	for _, entry := range entries {
		expectedType, ok := allowed[entry.Group][entry.Key]
		if !ok || IsDomainManagedSettingGroup(entry.Group) || IsDomainManagedSettingKey(entry.Key) || isSensitiveSettingKey(entry.Key) {
			return invalid("文件包含不支持导入的配置项")
		}
		if !configurationBackupLocalePattern.MatchString(entry.Locale) || entry.Type != expectedType || len(entry.Value) > 256<<10 || len(entry.Description) > 4096 {
			return invalid("配置项的语言、类型或长度无效")
		}
		if (entry.Group == "api" || entry.Group == "email") && entry.IsPublic {
			return invalid("API 与邮件配置必须保持非公开")
		}
		if entry.Type == "boolean" {
			if _, err := strconv.ParseBool(entry.Value); err != nil {
				return invalid("布尔配置值无效")
			}
		}
		if entry.Type == "number" {
			var number float64
			if strings.TrimSpace(entry.Value) == "null" || json.Unmarshal([]byte(entry.Value), &number) != nil {
				return invalid("数字配置值无效")
			}
		}
		identity := entry.Key + "\x00" + entry.Locale
		if seen[identity] {
			return invalid("文件包含重复的配置项与语言")
		}
		seen[identity] = true
		if err := validateSettingRequest(entry); err != nil {
			return nil, err
		}
	}
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].Key == entries[j].Key {
			return entries[i].Locale < entries[j].Locale
		}
		return entries[i].Key < entries[j].Key
	})
	return entries, nil
}

func validateSettingsConfigurationTargetGroups(entries []setting.UpdateSettingRequest, current []setting.Setting) error {
	requestedGroups := map[string]string{}
	for _, entry := range entries {
		requestedGroups[entry.Key+"\x00"+entry.Locale] = entry.Group
	}
	for _, row := range current {
		if group, requested := requestedGroups[row.Key+"\x00"+row.Locale]; requested && row.Group != group {
			return fmt.Errorf("%w: 当前配置分组与文件不一致，请通过对应设置页面处理", ErrSettingInvalid)
		}
	}
	return nil
}

func buildSettingsConfigurationImportPreview(entries []setting.UpdateSettingRequest, current []setting.Setting) SettingsConfigurationImportPreview {
	byIdentity := map[string]setting.UpdateSettingRequest{}
	for _, row := range current {
		byIdentity[row.Key+"\x00"+row.Locale] = settingsConfigurationTransferEntryFromStoredSetting(row)
	}
	target := make([]*setting.UpdateSettingRequest, 0, len(entries))
	preview := SettingsConfigurationImportPreview{Changes: []SettingsConfigurationImportChange{}}
	for _, entry := range entries {
		previous, exists := byIdentity[entry.Key+"\x00"+entry.Locale]
		change := SettingsConfigurationImportChange{Key: entry.Key, Group: entry.Group, Locale: entry.Locale, ImportedValue: entry.Value, CurrentValue: previous.Value}
		change.CurrentType, change.ImportedType = previous.Type, entry.Type
		change.CurrentIsPublic, change.ImportedIsPublic = previous.IsPublic, entry.IsPublic
		change.CurrentDescription, change.ImportedDescription = previous.Description, entry.Description
		if exists {
			copy := previous
			target = append(target, &copy)
			if previous == entry {
				change.Action = "unchanged"
				preview.UnchangedCount++
			} else {
				change.Action = "update"
				preview.UpdateCount++
			}
		} else {
			target = append(target, nil)
			change.Action = "create"
			preview.CreateCount++
		}
		preview.Changes = append(preview.Changes, change)
	}
	hashJSON := func(value any) string {
		data, _ := json.Marshal(value)
		hash := sha256.Sum256(data)
		return hex.EncodeToString(hash[:])
	}
	preview.BackupSHA256 = hashJSON(entries)
	preview.TargetSHA256 = hashJSON(struct {
		Entries []setting.UpdateSettingRequest
		Current []*setting.UpdateSettingRequest
	}{entries, target})
	return preview
}

func (s *AdminSettingsService) PreviewPortableSettingsConfigurationImport(ctx context.Context, file SettingsConfigurationBackupFile) (*SettingsConfigurationImportPreview, error) {
	entries, err := validateAndSortPortableSettingsConfigurationFile(file)
	if err != nil {
		return nil, err
	}
	current, err := s.settings.settingRepo.ListAllLocalesForSettingsConfigurationBackup(ctx, maximumSettingsConfigurationEntries+1)
	if err != nil {
		return nil, err
	}
	if len(current) > maximumSettingsConfigurationEntries {
		return nil, fmt.Errorf("%w: 当前配置数量超过预览上限", ErrSettingInvalid)
	}
	if err := validateSettingsConfigurationTargetGroups(entries, current); err != nil {
		return nil, err
	}
	preview := buildSettingsConfigurationImportPreview(entries, current)
	return &preview, nil
}

func (s *AdminSettingsService) ImportPortableSettingsConfigurationWithPreview(ctx context.Context, file SettingsConfigurationBackupFile, expectedTargetSHA256, confirmation string) (int, error) {
	if confirmation != "IMPORT CONFIGURATION" || len(expectedTargetSHA256) != 64 {
		return 0, fmt.Errorf("%w: 请先预览并确认导入配置", ErrSettingInvalid)
	}
	entries, err := validateAndSortPortableSettingsConfigurationFile(file)
	if err != nil {
		return 0, err
	}
	values := make([]setting.Setting, 0, len(entries))
	for _, entry := range entries {
		value, err := s.normalizeRequest(entry)
		if err != nil {
			return 0, err
		}
		values = append(values, value)
	}
	err = s.settings.settingRepo.ImportSettingsConfigurationAfterLockedSnapshotValidation(ctx, values, func(current []setting.Setting) error {
		if err := validateSettingsConfigurationTargetGroups(entries, current); err != nil {
			return err
		}
		if buildSettingsConfigurationImportPreview(entries, current).TargetSHA256 != expectedTargetSHA256 {
			return ErrSettingsConfigurationPreviewChanged
		}
		return nil
	})
	if err != nil {
		return 0, mapSettingRepositoryMediaErrorToInvalidSettingValue(err)
	}
	for _, value := range values {
		s.settings.invalidateSettingCaches(value.Key, value.Group, value.Locale)
	}
	return len(values), nil
}
