package localbackup

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"commerce-platform/internal/pkg/secretbox"
)

// Account recovery copies original rows, never Save/Ping/Sync business methods.
// Gateway credentials stay encrypted; the maintenance process verifies them
// locally with its original, application-bound business keys before writing.
func configurationModuleMissingFpxAndYanwenAccountWritableTableNames() []string {
	return []string{"shipping_fpx_api_configs", "shipping_yanwen_api_configs"}
}

func validateConfigurationModuleMissingFpxAndYanwenAccountDomainRows(missing map[string][]json.RawMessage) error {
	for _, table := range configurationModuleMissingFpxAndYanwenAccountWritableTableNames() {
		for _, row := range missing[table] {
			if err := validateConfigurationModuleOriginalFpxOrYanwenAccountRow(table, row); err != nil {
				return fmt.Errorf("物流账号原环境、网关或字段无效（%s）：%w", table, err)
			}
		}
	}
	return nil
}

func validateConfigurationModuleOriginalFpxOrYanwenAccountRow(table string, row json.RawMessage) error {
	var fields map[string]json.RawMessage
	if json.Unmarshal(row, &fields) != nil || fields == nil {
		return errors.New("original complete account row is required")
	}
	id, err := strconv.ParseUint(string(fields["id"]), 10, 64)
	if err != nil || id == 0 {
		return errors.New("original account identity is invalid")
	}
	if string(fields["enabled"]) != "true" && string(fields["enabled"]) != "false" {
		return errors.New("original account enabled state is required")
	}
	readOriginalText := func(column string) string {
		value, _ := readBusinessModuleSnapshotRowColumnText(row, column)
		return value
	}
	environment := readOriginalText("environment")
	host, path := "", ""
	switch table {
	case "shipping_fpx_api_configs":
		host, path = "open.4px.com", "/router/api/service"
		if environment == "test" {
			host = "open-test.4px.com"
		} else if environment != "production" {
			return errors.New("original 4PX environment must be test or production")
		}
		for _, column := range []string{"last_sync_scanned", "last_sync_added", "last_sync_updated", "last_sync_preserved_enabled"} {
			value, err := strconv.ParseInt(string(fields[column]), 10, 64)
			if err != nil || value < 0 {
				return errors.New("original account sync count is invalid")
			}
		}
		if original := readOriginalText("last_synced_at"); original != "" {
			if _, err := time.Parse(time.RFC3339Nano, original); err != nil {
				return errors.New("original account sync timestamp is invalid")
			}
		}
	case "shipping_yanwen_api_configs":
		host, path = "open.yw56.com.cn", "/api/order"
		if environment == "fat" {
			host = "open-fat.yw56.com.cn"
		} else if environment != "production" {
			return errors.New("original Yanwen environment must be fat or production")
		}
	default:
		return errors.New("unregistered logistics account table")
	}
	endpoint, err := url.Parse(readOriginalText("endpoint"))
	if err != nil || endpoint == nil || endpoint.Scheme != "https" || !strings.EqualFold(endpoint.Host, host) || endpoint.Path != path || endpoint.RawPath != "" || endpoint.RawQuery != "" || endpoint.ForceQuery || endpoint.User != nil || endpoint.Fragment != "" {
		return errors.New("original account gateway must match its official environment")
	}
	for _, column := range configurationModuleFpxOrYanwenAccountEncryptedCredentialColumns(table) {
		var encrypted string
		if string(fields[column]) == "null" || json.Unmarshal(fields[column], &encrypted) != nil {
			return errors.New("original encrypted credential field is required")
		}
		// Optional 4PX access_token and unfinished disabled configurations remain
		// optional. Every nonempty ciphertext is still authenticated below.
		if string(fields["enabled"]) == "true" && column != "access_token_encrypted" && strings.TrimSpace(encrypted) == "" {
			return errors.New("enabled account requires its original encrypted credentials")
		}
	}
	return nil
}

func configurationModuleFpxOrYanwenAccountEncryptedCredentialColumns(table string) []string {
	if table == "shipping_fpx_api_configs" {
		return []string{"app_key_encrypted", "app_secret_encrypted", "access_token_encrypted"}
	}
	return []string{"user_id_encrypted", "api_token_encrypted"}
}

func (cfg Configuration) validateConfigurationModuleMissingFpxAndYanwenAccountCredentialsWithOriginalBusinessKeys(missing map[string][]json.RawMessage) error {
	if err := validateConfigurationModuleMissingFpxAndYanwenAccountDomainRows(missing); err != nil {
		return err
	}
	for _, table := range configurationModuleMissingFpxAndYanwenAccountWritableTableNames() {
		keyName := "FPX_API_MASTER_KEY"
		if table == "shipping_yanwen_api_configs" {
			keyName = "YANWEN_API_MASTER_KEY"
		}
		for _, row := range missing[table] {
			for _, column := range configurationModuleFpxOrYanwenAccountEncryptedCredentialColumns(table) {
				encrypted, _ := readBusinessModuleSnapshotRowColumnText(row, column)
				if encrypted == "" {
					continue
				}
				// Do not fall back to process environment: these keys are bound to
				// the application boundary and authenticated package fingerprints.
				plaintext, err := secretbox.DecryptString(encrypted, strings.TrimSpace(cfg.RecoverySecrets[keyName]))
				if err != nil || strings.TrimSpace(plaintext) == "" {
					// Never expose ciphertext, plaintext, endpoint or crypto errors.
					return fmt.Errorf("物流账号原凭据无法用已绑定业务密钥验证（%s）；请核对 %s 后重新演练", table, keyName)
				}
			}
		}
	}
	return nil
}
