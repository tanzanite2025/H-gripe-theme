package localbackup

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"commerce-platform/internal/domain/notification"
	"commerce-platform/internal/pkg/secretbox"
)

func configurationModuleMissingEmailSenderWritableTableNames() []string {
	return []string{"email_provider_configs"}
}

// Recovery copies original rows through the normal-constraint writer. It must
// not call Create/Update/SetDefault/Test: those methods change other defaults,
// normalize fields, re-encrypt credentials or send a real message.
func validateConfigurationModuleMissingEmailSenderDomainRows(missing map[string][]json.RawMessage) error {
	for _, row := range missing["email_provider_configs"] {
		var record notification.EmailProviderConfig
		var fields map[string]json.RawMessage
		if json.Unmarshal(row, &record) != nil || json.Unmarshal(row, &fields) != nil || fields == nil {
			return errors.New("邮件发送方需要原完整配置行")
		}
		id, err := strconv.ParseUint(string(fields["id"]), 10, 64)
		if err != nil || id == 0 {
			return errors.New("邮件发送方原身份无效")
		}
		for _, column := range []string{"is_active", "is_default"} {
			if string(fields[column]) != "true" && string(fields[column]) != "false" {
				return errors.New("邮件发送方缺少原启停或默认标记")
			}
		}
		// PasswordEncrypted has json:"-" on the business model. Validate the
		// actual database column separately instead of silently discarding it.
		var encrypted string
		if string(fields["password_encrypted"]) == "null" || json.Unmarshal(fields["password_encrypted"], &encrypted) != nil {
			return errors.New("邮件发送方缺少原密码密文字段")
		}
		if record.IsDefault && !record.IsActive {
			return errors.New("原默认邮件发送方必须处于启用状态；恢复不会自动启用")
		}
		if err := record.Validate(); err != nil {
			// Business validation errors are fixed field descriptions, never row
			// values. Keep passwords, SMTP hosts and historical test errors out.
			return fmt.Errorf("邮件发送方原SMTP字段无效：%w", err)
		}
	}
	return nil
}

// Match EmailProviderMasterKey's first-nonempty precedence using only keys
// authenticated by the package and application boundary. Never try lower keys
// after a decryption failure, or consult the maintenance process environment.
func (cfg Configuration) resolveConfigurationEmailSenderOriginalBoundBusinessKey() string {
	for _, name := range []string{"EMAIL_PROVIDER_MASTER_KEY", "SECURITY_MASTER_KEY", "PAYMENT_CONFIG_MASTER_KEY"} {
		if value := strings.TrimSpace(cfg.RecoverySecrets[name]); value != "" {
			return value
		}
	}
	return ""
}

func (cfg Configuration) validateConfigurationModuleMissingEmailSenderCredentialsWithOriginalBoundKeyPrecedence(missing map[string][]json.RawMessage) error {
	if err := validateConfigurationModuleMissingEmailSenderDomainRows(missing); err != nil {
		return err
	}
	for _, row := range missing["email_provider_configs"] {
		encrypted, _ := readBusinessModuleSnapshotRowColumnText(row, "password_encrypted")
		if encrypted == "" {
			// SMTP without authentication is supported by the existing business
			// service, for example a local relay. Do not invent a new password.
			continue
		}
		plaintext, err := secretbox.DecryptString(encrypted, cfg.resolveConfigurationEmailSenderOriginalBoundBusinessKey())
		if err != nil || strings.TrimSpace(plaintext) == "" {
			return errors.New("邮件发送方原凭据无法用已绑定业务密钥验证；请按EMAIL_PROVIDER_MASTER_KEY、SECURITY_MASTER_KEY、PAYMENT_CONFIG_MASTER_KEY的原优先级核对后重新演练")
		}
	}
	return nil
}
