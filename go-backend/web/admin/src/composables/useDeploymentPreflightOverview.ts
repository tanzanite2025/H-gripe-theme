import { computed, type Ref } from 'vue'
import type {
  OpsDeploymentPreflight,
  OpsDeploymentPreflightGroup,
  OpsDeploymentPreflightOverview,
  OpsDeploymentPreflightSummary,
} from '@/api/ops'
import {
  aggregatePreflightOverviewCategories,
  environmentLabel,
  formatDeploymentPreflightDate,
  preflightSummaryRiskScore,
} from '@/lib/deploymentPreflightPresentation'

export type DeploymentPreflightOverviewStatus = 'all' | 'blocked' | 'review' | 'ready'
export type DeploymentPreflightOverviewSort = 'risk' | 'name' | 'generated'

export interface DeploymentPreflightProjectOption {
  id: number
  name: string
  environment: string
}

export interface DeploymentPreflightOverviewStat {
  key: string
  label: string
  value: number
  detail: string
}

export interface UseDeploymentPreflightOverviewOptions {
  projects: Ref<Array<{
    id: number
    name: string
    environment: string
  }>>
  overview: Ref<OpsDeploymentPreflightOverview | null>
  statusFilter: Ref<DeploymentPreflightOverviewStatus>
  environmentFilter: Ref<string>
  sort: Ref<DeploymentPreflightOverviewSort>
}

export const useDeploymentPreflightOverview = ({
  projects,
  overview,
  statusFilter,
  environmentFilter,
  sort,
}: UseDeploymentPreflightOverviewOptions) => {
  const overviewProjects = computed(() => overview.value?.projects || [])
  const overviewCategories = computed(() => overview.value?.categories || [])
  const overviewRiskCategories = computed(() => overviewCategories.value
    .filter((group) => group.blocking_count > 0 || group.warning_count > 0)
    .slice(0, 4))
  const filteredOverviewProjects = computed(() => {
    const filtered = overviewProjects.value.filter((summary) => {
      if (environmentFilter.value && summary.environment !== environmentFilter.value) {
        return false
      }
      if (statusFilter.value === 'blocked') return summary.blocking_count > 0
      if (statusFilter.value === 'review') return summary.status_level === 'review'
      if (statusFilter.value === 'ready') return summary.status_level === 'ready'
      return true
    })

    return [...filtered].sort((a, b) => {
      if (sort.value === 'name') {
        return a.project.localeCompare(b.project, 'zh-CN')
      }
      if (sort.value === 'generated') {
        return new Date(b.generated_at).getTime() - new Date(a.generated_at).getTime()
      }
      return preflightSummaryRiskScore(b) - preflightSummaryRiskScore(a) || a.project.localeCompare(b.project, 'zh-CN')
    })
  })
  const projectOptions = computed<DeploymentPreflightProjectOption[]>(() => {
    if (projects.value.length > 0) {
      return projects.value.map((project) => ({
        id: project.id,
        name: project.name,
        environment: project.environment,
      }))
    }
    return overviewProjects.value.map((summary) => ({
      id: summary.project_id,
      name: summary.project,
      environment: summary.environment,
    }))
  })
  const overviewStats = computed<DeploymentPreflightOverviewStat[]>(() => {
    const data = overview.value
    return [
      {
        key: 'projects',
        label: '项目数',
        value: data?.project_count ?? projects.value.length,
        detail: data ? `${environmentLabel(data.environment)}范围 · ${formatDeploymentPreflightDate(data.generated_at)}` : '等待总览生成',
      },
      {
        key: 'ready',
        label: 'READY',
        value: data?.ready_count ?? 0,
        detail: '无警告且无阻断项',
      },
      {
        key: 'review',
        label: 'REVIEW',
        value: data?.review_count ?? 0,
        detail: '无阻断但需要人工确认',
      },
      {
        key: 'blocked',
        label: 'BLOCKED',
        value: data?.blocked_count ?? 0,
        detail: '存在发布前阻断项',
      },
    ]
  })

  const mergeReportSummary = (nextReport: OpsDeploymentPreflight): void => {
    if (!overview.value) return
    const nextSummary: OpsDeploymentPreflightSummary = {
      project_id: nextReport.project_id,
      project: nextReport.project,
      environment: nextReport.environment,
      ready: nextReport.ready,
      status_level: nextReport.status_level,
      blocking_count: nextReport.blocking_count,
      warning_count: nextReport.warning_count,
      pass_count: nextReport.pass_count,
      info_count: nextReport.info_count,
      summary: nextReport.summary,
      generated_at: nextReport.generated_at,
      next_actions: nextReport.next_actions || [],
      categories: nextReport.categories || [],
      block_reasons: nextReport.checks
        .filter((check) => check.status === 'block')
        .map((check) => `${check.label}：${check.message}`),
      warn_reasons: nextReport.checks
        .filter((check) => check.status === 'warning')
        .map((check) => `${check.label}：${check.message}`),
    }
    const found = overview.value.projects.some((summary) => summary.project_id === nextSummary.project_id)
    const nextProjects = found
      ? overview.value.projects.map((summary) => (summary.project_id === nextSummary.project_id ? nextSummary : summary))
      : [...overview.value.projects, nextSummary]
    overview.value = {
      ...overview.value,
      projects: nextProjects,
      project_count: nextProjects.length,
      categories: aggregatePreflightOverviewCategories(nextProjects),
      ready_count: nextProjects.filter((summary) => summary.status_level === 'ready').length,
      review_count: nextProjects.filter((summary) => summary.status_level === 'review').length,
      blocked_count: nextProjects.filter((summary) => summary.status_level === 'blocked').length,
      warning_count: nextProjects.filter((summary) => summary.warning_count > 0).length,
    }
  }

  return {
    overviewProjects,
    overviewCategories,
    overviewRiskCategories,
    filteredOverviewProjects,
    projectOptions,
    overviewStats,
    mergeReportSummary,
  }
}
