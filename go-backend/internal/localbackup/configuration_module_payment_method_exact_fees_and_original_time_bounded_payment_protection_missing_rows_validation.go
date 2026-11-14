package localbackup

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"unicode/utf8"

	paymentdomain "commerce-platform/internal/domain/payment"
	paymentgateway "commerce-platform/internal/pkg/payment"
)

func configurationModuleMissingPaymentMethodAndTimeBoundedProtectionWritableTableNames() []string {
	return []string{"payment_methods", "payment_protection_controls"}
}

// Manual protection actors have no database FK. A numeric identity must still
// refer to the same original account; zero is the existing system-actor value.
func configurationModulePaymentProtectionActorReadOnlyDependencyReferences() []ModuleRecoveryDependencyReference {
	return []ModuleRecoveryDependencyReference{
		{SourceTable: "payment_protection_controls", SourceColumns: []string{"created_by"}, TargetTable: "users", TargetColumns: []string{"id"}},
		{SourceTable: "payment_protection_controls", SourceColumns: []string{"updated_by"}, TargetTable: "users", TargetColumns: []string{"id"}},
	}
}

func validateConfigurationModuleMissingPaymentMethodAndTimeBoundedProtectionDomainRows(missing map[string][]json.RawMessage) error {
	for _, table := range configurationModuleMissingPaymentMethodAndTimeBoundedProtectionWritableTableNames() {
		for _, row := range missing[table] {
			if err := validateConfigurationModuleRestoredPaymentMethodOrProtectionCompleteRow(table, row); err != nil {
				return err
			}
		}
	}
	return nil
}

// Validate the stored configuration without calling payment creation, gateway
// readiness, CreateControl or RevokeControl. In particular, recovery must not
// renew an expired override or stamp a new actor, audit or creation time.
func validateConfigurationModuleRestoredPaymentMethodOrProtectionCompleteRow(table string, row json.RawMessage) error {
	var fields map[string]json.RawMessage
	if json.Unmarshal(row, &fields) != nil || fields == nil {
		return errors.New("支付目录与保护规则需要原完整配置行")
	}
	id, err := strconv.ParseInt(string(fields["id"]), 10, 64)
	if err != nil || id <= 0 {
		return errors.New("支付目录或保护规则原身份无效")
	}
	if string(fields["enabled"]) != "true" && string(fields["enabled"]) != "false" {
		return errors.New("支付目录或保护规则缺少原启停标记")
	}
	if table == "payment_methods" {
		var method paymentdomain.PaymentMethod
		for column, destination := range map[string]*string{"name": &method.Name, "code": &method.Code, "fee_type": &method.FeeType} {
			if string(fields[column]) == "null" || json.Unmarshal(fields[column], destination) != nil || strings.TrimSpace(*destination) == "" {
				return errors.New("支付方式原名称、代码或计费类型无效")
			}
		}
		if utf8.RuneCountInString(method.Name) > 255 || utf8.RuneCountInString(method.Code) > 100 || (method.FeeType != "fixed" && method.FeeType != "percentage") {
			return errors.New("支付方式原名称、代码或计费类型无效")
		}
		for column, destination := range map[string]*int64{"fee_value_minor": &method.FeeValueMinor, "min_amount_minor": &method.MinAmountMinor, "max_amount_minor": &method.MaxAmountMinor} {
			value, err := strconv.ParseInt(string(fields[column]), 10, 64)
			if err != nil {
				return errors.New("支付方式原金额必须是完整精确的最小货币单位整数")
			}
			*destination = value
		}
		// PostgreSQL emits NUMERIC as a JSON number, whereas the business model
		// exposes a string. Read the original numeric text without float64 or
		// unmarshalling it into that string field. Only a local copy is checked.
		method.FeeRateDecimal = string(fields["fee_rate_decimal"])
		if err := method.BeforeSave(nil); err != nil {
			return fmt.Errorf("支付方式原费率或金额无效：%w", err)
		}
		return nil
	}
	if table != "payment_protection_controls" {
		return errors.New("支付配置表没有对应的恢复校验契约")
	}
	var control paymentdomain.PaymentProtectionControl
	if json.Unmarshal(row, &control) != nil {
		return errors.New("原支付保护规则字段无效")
	}
	for _, column := range []string{"created_by", "updated_by"} {
		actor, err := strconv.ParseInt(string(fields[column]), 10, 64)
		if err != nil || actor < 0 {
			return errors.New("原支付保护规则操作员身份无效")
		}
	}
	if control.Action != paymentdomain.PaymentProtectionActionForce3DS && control.Action != paymentdomain.PaymentProtectionActionPausePayment {
		return errors.New("原支付保护动作不受支持")
	}
	value := control.ScopeValue
	switch control.ScopeType {
	case paymentdomain.PaymentProtectionScopeGlobal:
		if value != "" {
			return errors.New("全局支付保护规则不应包含局部范围")
		}
	case paymentdomain.PaymentProtectionScopeProvider:
		if _, err := paymentgateway.ParseGatewayType(value); err != nil || value != strings.ToLower(strings.TrimSpace(value)) {
			return errors.New("原支付保护网关范围无效")
		}
	case paymentdomain.PaymentProtectionScopeCountry:
		if len(value) != 2 || value[0] < 'A' || value[0] > 'Z' || value[1] < 'A' || value[1] > 'Z' {
			return errors.New("原支付保护国家范围无效")
		}
	case paymentdomain.PaymentProtectionScopePaymentMethod:
		// This is a gateway method token, not a payment_methods numeric ID.
		if value == "" || len(value) > 128 || value != strings.ToLower(strings.TrimSpace(value)) {
			return errors.New("原支付保护方式范围无效")
		}
	default:
		return errors.New("原支付保护范围类型不受支持")
	}
	if strings.TrimSpace(control.Reason) == "" || len(control.Reason) > 2000 {
		return errors.New("原支付保护原因无效")
	}
	if control.CreatedAt.IsZero() || control.UpdatedAt.IsZero() || control.ExpiresAt.IsZero() || !control.ExpiresAt.After(control.CreatedAt) {
		return errors.New("原支付保护创建时间或到期时间无效；恢复不会延长有效期")
	}
	return nil
}
