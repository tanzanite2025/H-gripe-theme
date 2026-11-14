package localbackup

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"time"

	"github.com/lib/pq"

	"commerce-platform/internal/domain/currency"
)

func configurationModuleMissingCouponWritableTableNames() []string {
	return []string{"coupons"}
}

// Read numeric fields directly from JSON tokens; never round money through
// float64, call model hooks, renew expiry or reset the original used count.
func validateConfigurationModuleMissingCouponDomainRows(missing map[string][]json.RawMessage) error {
	for _, row := range missing["coupons"] {
		var object map[string]json.RawMessage
		if json.Unmarshal(row, &object) != nil || object == nil {
			return errors.New("优惠券原完整行无效")
		}
		readText := func(column string) string {
			value, _ := readBusinessModuleSnapshotRowColumnText(row, column)
			return value
		}
		for _, column := range []string{"id", "value_minor", "min_amount_minor", "max_discount_minor", "usage_limit", "usage_limit_per_user", "used_count"} {
			value, err := strconv.ParseInt(string(object[column]), 10, 64)
			if err != nil || value < 0 || (column == "id" && value == 0) {
				return fmt.Errorf("优惠券原整数字段无效：%s", column)
			}
		}
		if strings.TrimSpace(readText("code")) == "" || (readText("type") != "fixed" && readText("type") != "percentage") || !currency.IsCatalogCode(readText("currency")) || currency.NormalizeCode(readText("currency")) != readText("currency") {
			return errors.New("优惠券原代码、类型或币种无效")
		}
		if readText("type") == "fixed" && readText("value_minor") == "0" {
			return errors.New("固定优惠券原金额必须大于零")
		}
		rate, valid := new(big.Rat).SetString(readText("value_rate_decimal"))
		rawRate := object["value_rate_decimal"]
		if len(rawRate) == 0 || rawRate[0] == '"' || !valid || rate.Sign() < 0 || rate.Cmp(big.NewRat(100, 1)) > 0 || (readText("type") == "percentage" && rate.Sign() == 0) {
			return errors.New("优惠券原百分比无效")
		}
		limit, _ := strconv.ParseInt(readText("usage_limit"), 10, 64)
		used, _ := strconv.ParseInt(readText("used_count"), 10, 64)
		if limit > 0 && used > limit {
			return errors.New("优惠券原使用次数超过总限制")
		}
		start, startErr := time.Parse(time.RFC3339Nano, readText("start_date"))
		end, endErr := time.Parse(time.RFC3339Nano, readText("end_date"))
		if startErr != nil || endErr != nil || !end.After(start) {
			return errors.New("优惠券原有效期无效")
		}
		if string(object["enabled"]) != "true" && string(object["enabled"]) != "false" {
			return errors.New("优惠券原启停状态缺失")
		}
		if recipient := readText("referral_recipient_user_id"); recipient != "" && !canonicalModuleRecoveryDurableEventPositiveID(recipient) {
			return errors.New("优惠券原私人接收人无效")
		}
		if _, err := readConfigurationCouponProductAndCategoryScopeReferences(row); err != nil {
			return fmt.Errorf("优惠券原商品或商品分类范围无效：%w", err)
		}
	}
	return nil
}

func validateConfigurationModuleRestoredCouponsAgainstOriginalScopeAndRetainedUsage(ctx context.Context, tx *sql.Tx, snapshot selectedBusinessModuleCompleteSnapshot, missing map[string][]json.RawMessage) error {
	if len(missing["coupons"]) == 0 {
		return nil
	}
	if err := validateConfigurationModuleMissingCouponDomainRows(missing); err != nil {
		return err
	}
	sets, err := readConfigurationCouponRetainedUsageReceiptSets(ctx, tx, missing["coupons"])
	if err != nil {
		return err
	}
	expected := map[string]string{}
	for _, reference := range snapshot.ExternalDependencies {
		if reference.Table == "coupon_usage" && len(reference.Values) == 1 {
			expected[reference.Values["coupon_id"]] = reference.RowSHA256
		}
	}
	for _, row := range missing["coupons"] {
		text := func(column string) string {
			value, _ := readBusinessModuleSnapshotRowColumnText(row, column)
			return value
		}
		receipts := sets[text("id")]
		digest, err := businessModuleRecoveryCanonicalJSONSHA256(receipts)
		if err != nil || expected[text("id")] != digest {
			return errors.New("优惠券缺少原使用记录集合摘要或当前使用记录已变化，请重新导出或使用合格整站包")
		}
		total, userCounts, emailCounts := int64(0), map[string]int64{}, map[string]int64{}
		for _, receipt := range receipts {
			readReceiptText := func(column string) string {
				value, _ := readBusinessModuleSnapshotRowColumnText(receipt, column)
				return value
			}
			userID, userErr := strconv.ParseInt(readReceiptText("user_id"), 10, 64)
			discount, discountErr := strconv.ParseInt(readReceiptText("discount_minor"), 10, 64)
			if !canonicalModuleRecoveryDurableEventPositiveID(readReceiptText("order_id")) || userErr != nil || userID < 0 || discountErr != nil || discount < 0 || !currency.IsCatalogCode(readReceiptText("currency")) {
				return errors.New("优惠券原使用记录的订单、用户或精确金额无效")
			}
			status, _ := readBusinessModuleSnapshotRowColumnText(receipt, "status")
			if status == "reversed" {
				continue
			}
			if status != "applied" {
				return errors.New("优惠券原使用记录状态无效")
			}
			total++
			user, _ := readBusinessModuleSnapshotRowColumnText(receipt, "user_id")
			email, _ := readBusinessModuleSnapshotRowColumnText(receipt, "email")
			email = strings.ToLower(strings.TrimSpace(email))
			userCounts[user]++
			emailCounts[email]++
			if recipient := text("referral_recipient_user_id"); recipient != "" && recipient != user {
				return errors.New("优惠券原使用记录与私人接收人不符")
			}
		}
		used, _ := strconv.ParseInt(text("used_count"), 10, 64)
		if total != used {
			return errors.New("优惠券原使用次数与仍生效的使用记录不一致；不会重算或补造记录")
		}
		perUser, _ := strconv.ParseInt(text("usage_limit_per_user"), 10, 64)
		if perUser > 0 {
			for _, receipt := range receipts {
				status, _ := readBusinessModuleSnapshotRowColumnText(receipt, "status")
				if status != "applied" {
					continue
				}
				user, _ := readBusinessModuleSnapshotRowColumnText(receipt, "user_id")
				email, _ := readBusinessModuleSnapshotRowColumnText(receipt, "email")
				email = strings.ToLower(strings.TrimSpace(email))
				if (user != "0" && userCounts[user] > perUser) || (user == "0" && (email == "" || emailCounts[email] > perUser)) {
					return errors.New("优惠券原用户或游客使用记录超出每人限制或缺少身份")
				}
			}
		}
		references, _ := readConfigurationCouponProductAndCategoryScopeReferences(row)
		for _, reference := range references {
			predicate := pq.QuoteIdentifier(reference.Column) + "::text=$1"
			if reference.Column == "slug" {
				predicate = "lower(btrim(slug))=$1"
			}
			rows, err := readProductCatalogRows(ctx, tx, reference.Table, predicate, reference.Value)
			if err != nil {
				return err
			}
			if len(rows) != 1 {
				return errors.New("优惠券原商品或商品分类缺失或身份不唯一")
			}
			if reference.Table == "products" {
				deleted, _ := readBusinessModuleSnapshotRowColumnText(rows[0], "deleted_at")
				if deleted != "" {
					return errors.New("优惠券原范围商品已删除")
				}
			} else {
				enabled, _ := readBusinessModuleSnapshotRowColumnText(rows[0], "is_enabled")
				if enabled != "true" {
					return errors.New("优惠券原商品分类已停用")
				}
			}
		}
	}
	return nil
}
