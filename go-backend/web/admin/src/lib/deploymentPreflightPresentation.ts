import type {
  OpsDeploymentPreflight,
  OpsDeploymentPreflightGroup,
  OpsDeploymentPreflightOverview,
  OpsDeploymentPreflightSummary,
} from '@/api/ops'

export interface PreflightOverviewMarkdownInput {
  overview: OpsDeploymentPreflightOverview
  projects: OpsDeploymentPreflightSummary[]
  filterLabel: string
}

export const statusLevelLabel = (value?: string): string => ({
  ready: 'READY',
  review: 'REVIEW',
  blocked: 'BLOCKED',
}[value || 'ready'] || value || '-')

export const environmentLabel = (value: string): string => ({
  all: '全部',
  production: '生产',
  staging: '预发布',
  test: '测试',
  local: '本地',
}[value] || value || '-')

export const checkStatusLabel = (status: string): string => ({
  pass: 'PASS',
  warning: 'WARN',
  block: 'BLOCK',
  info: 'INFO',
}[status] || status || '-')

export const categoryLabel = (value?: string): string => ({
  all: '全部',
  compose: 'Compose',
  version: '版本',
  boundary: '边界',
  infra: 'VPS',
  connector: '连接器',
  domain: '域名',
  runtime: '远端健康',
  evidence: '证据',
  other: '其他',
}[value || 'other'] || value || '其他')

export const formatDeploymentPreflightDate = (value?: string): string => (
  value ? new Date(value).toLocaleString('zh-CN') : '-'
)

export const preflightSummaryRiskScore = (summary: OpsDeploymentPreflightSummary): number => (
  summary.blocking_count * 1000 + summary.warning_count * 10 + (summary.status_level === 'ready' ? 0 : 1)
)

export const summaryReason = (summary: OpsDeploymentPreflightSummary): string => {
  if (summary.block_reasons?.length) return summary.block_reasons[0]
  if (summary.warn_reasons?.length) return summary.warn_reasons[0]
  return summary.summary
}

export const aggregatePreflightOverviewCategories = (
  summaries: OpsDeploymentPreflightSummary[],
): OpsDeploymentPreflightGroup[] => {
  const groups = new Map<string, OpsDeploymentPreflightGroup>()
  summaries.forEach((summary) => {
    ;(summary.categories || []).forEach((group) => {
      const current = groups.get(group.category) || {
        category: group.category,
        label: group.label,
        total_count: 0,
        blocking_count: 0,
        warning_count: 0,
        pass_count: 0,
        info_count: 0,
      }
      current.total_count += group.total_count
      current.blocking_count += group.blocking_count
      current.warning_count += group.warning_count
      current.pass_count += group.pass_count
      current.info_count += group.info_count
      groups.set(group.category, current)
    })
  })
  return Array.from(groups.values()).sort((a, b) => (
    b.blocking_count - a.blocking_count
      || b.warning_count - a.warning_count
      || b.total_count - a.total_count
      || a.category.localeCompare(b.category)
  ))
}

export const buildPreflightReportMarkdown = (value: OpsDeploymentPreflight): string => {
  const sections = [
    `# Deployment Preflight · ${value.project}`,
    '',
    `- 状态：${statusLevelLabel(value.status_level)}`,
    `- 环境：${environmentLabel(value.environment)}`,
    `- 生成时间：${formatDeploymentPreflightDate(value.generated_at)}`,
    `- 统计：阻断 ${value.blocking_count} / 警告 ${value.warning_count} / 通过 ${value.pass_count} / 信息 ${value.info_count}`,
    `- 摘要：${value.summary}`,
  ]
  const blockers = value.checks.filter((check) => check.status === 'block')
  const warnings = value.checks.filter((check) => check.status === 'warning')
  const nextActions = value.next_actions || []
  const categories = value.categories || []
  if (categories.length > 0) {
    sections.push('', '## 类别分布', ...categories.map((group) => `- ${group.label}：阻断 ${group.blocking_count} / 警告 ${group.warning_count} / 总计 ${group.total_count}`))
  }
  if (nextActions.length > 0) {
    sections.push('', '## 下一步建议', ...nextActions.map((action) => `- ${action}`))
  }
  if (blockers.length > 0) {
    sections.push('', '## 阻断项', ...blockers.map((check) => `- ${check.label}：${check.message}${check.detail ? `（${check.detail}）` : ''}`))
  }
  if (warnings.length > 0) {
    sections.push('', '## 警告项', ...warnings.map((check) => `- ${check.label}：${check.message}${check.detail ? `（${check.detail}）` : ''}`))
  }
  sections.push('', '_该报告由部署中心只读 preflight 生成，不执行远端变更。_')
  return sections.join('\n')
}

export const buildPreflightChecklistMarkdown = (value: OpsDeploymentPreflight): string => {
  const sections = [
    `# Deployment Preflight Checklist · ${value.project}`,
    '',
    `- 状态：${statusLevelLabel(value.status_level)}`,
    `- 环境：${environmentLabel(value.environment)}`,
    `- 生成时间：${formatDeploymentPreflightDate(value.generated_at)}`,
    '',
    '## 类别分布',
    '',
    '| 类别 | 阻断 | 警告 | 通过 | 信息 | 总计 |',
    '| --- | --- | --- | --- | --- | --- |',
    ...(value.categories || []).map((group) => `| ${escapeMarkdownCell(group.label)} | ${group.blocking_count} | ${group.warning_count} | ${group.pass_count} | ${group.info_count} | ${group.total_count} |`),
    '',
    '## 检查清单',
    '',
    '| 检查 | 类别 | 状态 | 说明 | 证据 |',
    '| --- | --- | --- | --- | --- |',
    ...value.checks.map((check) => `| ${escapeMarkdownCell(check.label)} | ${escapeMarkdownCell(categoryLabel(check.category))} | ${checkStatusLabel(check.status)} | ${escapeMarkdownCell(check.message)} | ${escapeMarkdownCell(check.detail || '-')} |`),
  ]
  return sections.join('\n')
}

export const buildPreflightOverviewMarkdown = ({
  overview,
  projects,
  filterLabel,
}: PreflightOverviewMarkdownInput): string => {
  const sections = [
    '# Deployment Preflight Overview',
    '',
    `- 范围：${environmentLabel(overview.environment)}`,
    `- 生成时间：${formatDeploymentPreflightDate(overview.generated_at)}`,
    `- 当前筛选：${filterLabel}`,
    `- 项目：${projects.length} / ${overview.project_count}`,
    `- READY：${overview.ready_count} · REVIEW：${overview.review_count} · BLOCKED：${overview.blocked_count} · 有警告：${overview.warning_count}`,
  ]
  if (overview.categories?.length) {
    sections.push(
      '',
      '## 风险类别',
      ...overview.categories.map((group) => `- ${group.label}：阻断 ${group.blocking_count} / 警告 ${group.warning_count} / 总计 ${group.total_count}`),
    )
  }
  if (projects.length) {
    sections.push(
      '',
      '## 项目',
      ...projects.map((summary) => `- [${statusLevelLabel(summary.status_level)}] ${summary.project}（${environmentLabel(summary.environment)}）：${summaryReason(summary)}`),
    )
  }
  sections.push('', '_该总览由部署中心只读 preflight 生成，不执行远端变更。_')
  return sections.join('\n')
}

const escapeMarkdownCell = (value: string): string => value.replace(/\|/g, '\\|').replace(/\n/g, ' ')
