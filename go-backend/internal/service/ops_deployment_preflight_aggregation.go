package service

import (
	"fmt"
	"sort"
	"strings"

	"commerce-platform/internal/domain/ops"
)

func summarizeDeploymentPreflight(report *ops.DeploymentPreflight) ops.DeploymentPreflightSummary {
	if report == nil {
		return ops.DeploymentPreflightSummary{}
	}
	blockReasons := make([]string, 0)
	warnReasons := make([]string, 0)
	for _, check := range report.Checks {
		reason := strings.TrimSpace(fmt.Sprintf("%s：%s", check.Label, check.Message))
		switch check.Status {
		case ops.DeploymentCheckBlock:
			blockReasons = append(blockReasons, reason)
		case ops.DeploymentCheckWarning:
			warnReasons = append(warnReasons, reason)
		}
	}
	return ops.DeploymentPreflightSummary{
		ProjectID:     report.ProjectID,
		Project:       report.Project,
		Environment:   report.Environment,
		Ready:         report.Ready,
		StatusLevel:   report.StatusLevel,
		BlockingCount: report.BlockingCount,
		WarningCount:  report.WarningCount,
		PassCount:     report.PassCount,
		InfoCount:     report.InfoCount,
		Summary:       report.Summary,
		BlockReasons:  blockReasons,
		WarnReasons:   warnReasons,
		NextActions:   report.NextActions,
		Categories:    report.Categories,
		GeneratedAt:   report.GeneratedAt,
	}
}

func summarizePreflightGroups(checks []ops.DeploymentPreflightCheck) []ops.DeploymentPreflightGroup {
	groups := make(map[string]ops.DeploymentPreflightGroup)
	for _, check := range checks {
		key := strings.TrimSpace(check.Category)
		if key == "" {
			key = "other"
		}
		group := groups[key]
		group.Category = key
		group.Label = preflightCategoryLabel(key)
		group.TotalCount++
		switch check.Status {
		case ops.DeploymentCheckBlock:
			group.BlockingCount++
		case ops.DeploymentCheckWarning:
			group.WarningCount++
		case ops.DeploymentCheckPass:
			group.PassCount++
		case ops.DeploymentCheckInfo:
			group.InfoCount++
		}
		groups[key] = group
	}
	return sortedPreflightGroups(groups)
}

func mergePreflightGroups(target map[string]ops.DeploymentPreflightGroup, groups []ops.DeploymentPreflightGroup) {
	for _, group := range groups {
		key := strings.TrimSpace(group.Category)
		if key == "" {
			key = "other"
		}
		current := target[key]
		current.Category = key
		current.Label = preflightCategoryLabel(key)
		current.TotalCount += group.TotalCount
		current.BlockingCount += group.BlockingCount
		current.WarningCount += group.WarningCount
		current.PassCount += group.PassCount
		current.InfoCount += group.InfoCount
		target[key] = current
	}
}

func sortedPreflightGroups(groups map[string]ops.DeploymentPreflightGroup) []ops.DeploymentPreflightGroup {
	result := make([]ops.DeploymentPreflightGroup, 0, len(groups))
	for _, group := range groups {
		result = append(result, group)
	}
	sort.Slice(result, func(i, j int) bool {
		left := result[i]
		right := result[j]
		if left.BlockingCount != right.BlockingCount {
			return left.BlockingCount > right.BlockingCount
		}
		if left.WarningCount != right.WarningCount {
			return left.WarningCount > right.WarningCount
		}
		if left.TotalCount != right.TotalCount {
			return left.TotalCount > right.TotalCount
		}
		return left.Category < right.Category
	})
	return result
}

func preflightCategoryLabel(category string) string {
	switch category {
	case "compose":
		return "Compose"
	case "version":
		return "版本"
	case "boundary":
		return "边界"
	case "infra":
		return "VPS"
	case "connector":
		return "连接器"
	case "domain":
		return "域名"
	case "runtime":
		return "远端健康"
	case "evidence":
		return "证据"
	default:
		return "其他"
	}
}

func preflightStatusLevel(blockingCount, warningCount int) string {
	if blockingCount > 0 {
		return ops.DeploymentStatusBlocked
	}
	if warningCount > 0 {
		return ops.DeploymentStatusReview
	}
	return ops.DeploymentStatusReady
}

func preflightNextActions(checks []ops.DeploymentPreflightCheck) []string {
	actions := make([]string, 0)
	seen := map[string]struct{}{}
	for _, status := range []string{ops.DeploymentCheckBlock, ops.DeploymentCheckWarning} {
		for _, check := range checks {
			if check.Status != status {
				continue
			}
			action := strings.TrimSpace(fmt.Sprintf("%s：%s", check.Label, nextActionMessage(check)))
			if action == "" {
				continue
			}
			if _, ok := seen[action]; ok {
				continue
			}
			seen[action] = struct{}{}
			actions = append(actions, action)
			if len(actions) >= 6 {
				return actions
			}
		}
	}
	if len(actions) == 0 {
		return []string{"保留当前台账和观察证据；发布前重新生成一次 preflight 报告。"}
	}
	return actions
}

func nextActionMessage(check ops.DeploymentPreflightCheck) string {
	detail := strings.TrimSpace(check.Detail)
	switch check.Key {
	case "compose_source":
		return "补齐或确认部署使用的 Compose 来源。"
	case "project_name":
		return "统一项目绑定名和 Compose 项目名。"
	case "image_tag":
		return "登记不可变镜像标签或 digest。"
	case "commit_sha":
		return "登记完整 40 位源码 Commit SHA。"
	case "image_commit":
		return "重新核对镜像标签与源码 Commit 的对应关系。"
	case "services", "networks", "volumes", "gateway_boundary":
		return "对齐 Compose 服务、网络、卷或网关边界台账。"
	case "vps":
		return "修正 VPS 绑定、启用状态或 Hostinger 资源 ID。"
	case "vps_observed":
		return "刷新并核对 Hostinger VPS 只读观察状态。"
	case "vps_connector":
		return "为 VPS 绑定可读取的 Hostinger 连接器。"
	case "hostinger_connector":
		return "补齐 Hostinger 只读连接器、凭据和读取作用域。"
	case "hostinger_project_identity":
		return "补录 Hostinger 项目标识以提升证据质量。"
	case "ghcr_connector":
		return "补齐 GHCR/GitHub 只读镜像连接器或 packages 读取作用域。"
	case "domains":
		return "修正域名目标、启用状态或刷新 Cloudflare 观察结果。"
	case "cloudflare_connector":
		return "补齐 Cloudflare 只读连接器和 zones/dns 读取作用域。"
	case "remote_health":
		return "完成项目远端只读同步，并确认容器健康计数。"
	case "deployment_record":
		return "补录最近一次部署时间。"
	case "backups":
		return "补齐备份策略和恢复演练记录。"
	}
	if detail != "" {
		return detail
	}
	return check.Message
}
