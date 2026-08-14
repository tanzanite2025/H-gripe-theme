package service

import (
	"errors"
	"strings"
	"time"

	"commerce-platform/internal/domain/ops"
	"commerce-platform/internal/repository"
)

type OpsDomainDiffService struct {
	domainRepo *repository.OpsDomainBindingRepository
}

func NewOpsDomainDiffService(domainRepo *repository.OpsDomainBindingRepository) *OpsDomainDiffService {
	return &OpsDomainDiffService{domainRepo: domainRepo}
}

func (s *OpsDomainDiffService) Diff(id uint) (*ops.DomainDiff, error) {
	if s == nil || s.domainRepo == nil {
		return nil, errors.New("operations domain diff service is not configured")
	}
	domain, err := s.domainRepo.FindByID(id)
	if err != nil {
		return nil, err
	}
	return BuildOpsDomainDiff(*domain), nil
}

func BuildOpsDomainDiff(domain ops.DomainBinding) *ops.DomainDiff {
	generatedAt := time.Now().UTC()
	diff := &ops.DomainDiff{
		DomainID:       domain.ID,
		Domain:         domain.Domain,
		Environment:    domain.Environment,
		GeneratedAt:    generatedAt,
		Status:         normalizeDomainDiffStatus(domain.ObservedStatus),
		ObservedSource: strings.TrimSpace(domain.ObservedSource),
		LastObservedAt: domain.LastObservedAt,
		ObservedError:  strings.TrimSpace(domain.ObservedError),
		Items:          make([]ops.DomainDiffItem, 0, 5),
	}

	if diff.Status == ops.DomainDiffStatusUnknown {
		diff.Summary = "尚未同步实际状态，当前无法判断是否匹配。"
	} else if diff.Status == ops.DomainDiffStatusError {
		diff.Summary = "最近一次实际状态检查失败，请先检查连接器或手动重新同步。"
	} else if diff.Status == ops.DomainDiffStatusDrifted {
		diff.Summary = "实际状态与后台期望状态存在差异。"
	} else {
		diff.Summary = "实际状态与后台期望状态一致。"
	}

	diff.Items = append(diff.Items,
		compareDomainDiffItem(
			"target",
			"目标",
			domain.Target,
			domain.ObservedTarget,
			diff.Status,
			"DNS 目标",
			compareDomainTargetValues,
			comparableDomainTarget,
		),
		compareDomainDiffItem(
			"proxy_mode",
			"代理模式",
			domain.ProxyMode,
			domain.ObservedProxy,
			diff.Status,
			"代理模式",
			compareExactDomainValues,
			comparableDomainMode,
		),
		compareDomainDiffItem(
			"tls_mode",
			"TLS 模式",
			domain.TLSMode,
			domain.ObservedTLS,
			diff.Status,
			"SSL/TLS 模式",
			compareExactDomainValues,
			comparableDomainMode,
		),
		compareDomainDiffItem(
			"zone",
			"DNS Zone",
			domain.Zone,
			domain.Zone,
			ops.DomainDiffStatusMatched,
			"后台登记的 DNS Zone",
			compareExactDomainValues,
			func(value string) bool { return strings.TrimSpace(value) != "" },
		),
	)

	return diff
}

func compareDomainDiffItem(
	key string,
	label string,
	desired string,
	observed string,
	overallStatus string,
	subject string,
	matches func(string, string) bool,
	comparable func(string) bool,
) ops.DomainDiffItem {
	rawDesired := strings.TrimSpace(desired)
	rawObserved := strings.TrimSpace(observed)
	item := ops.DomainDiffItem{
		Key:      key,
		Label:    label,
		Desired:  displayDomainDiffValue(rawDesired),
		Observed: displayDomainDiffValue(rawObserved),
		Status:   ops.DomainDiffStatusMatched,
	}

	if overallStatus == ops.DomainDiffStatusUnknown {
		item.Status = ops.DomainDiffStatusUnknown
		item.Message = "尚未取得实际值。"
		return item
	}
	if overallStatus == ops.DomainDiffStatusError {
		item.Status = ops.DomainDiffStatusError
		item.Message = "实际检查失败，不能确认该项。"
		return item
	}
	if !comparable(rawDesired) {
		item.Status = ops.DomainDiffStatusUnknown
		item.Message = "后台未设置可比较的期望值，跳过比较。"
		return item
	}
	if matches(rawDesired, rawObserved) {
		item.Message = subject + "已匹配。"
		return item
	}
	item.Status = ops.DomainDiffStatusDrifted
	item.Message = subject + "与期望值不一致。"
	return item
}

func compareDomainTargetValues(desired, observed string) bool {
	return comparableDomainTarget(desired) && targetContains(observed, desired)
}

func compareExactDomainValues(desired, observed string) bool {
	return desired == observed
}

func comparableDomainMode(value string) bool {
	return value != "" && value != ops.DomainProxyUnknown && value != ops.DomainTLSUnknown
}

func normalizeDomainDiffStatus(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case ops.DomainDiffStatusMatched:
		return ops.DomainDiffStatusMatched
	case ops.DomainDiffStatusDrifted:
		return ops.DomainDiffStatusDrifted
	case ops.DomainDiffStatusError:
		return ops.DomainDiffStatusError
	default:
		return ops.DomainDiffStatusUnknown
	}
}

func displayDomainDiffValue(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "未设置"
	}
	return value
}
