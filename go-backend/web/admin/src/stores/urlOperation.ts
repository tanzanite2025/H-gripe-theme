import { computed, ref } from 'vue'
import { defineStore } from 'pinia'
import { toast } from 'vue-sonner'
import { storefrontRouteCatalogApi } from '@/modules/url-management/routeCatalog'
import type {
  StorefrontRouteCatalogCheckSummary,
  StorefrontRouteCatalogCheckTask,
  StorefrontRouteCatalogCheckTaskStatus,
  StorefrontRouteCatalogListParams,
} from '@/modules/url-management/routeCatalogTypes'

type StartCheckParams = Omit<StorefrontRouteCatalogListParams, 'page' | 'page_size'> & { limit?: number }

const emptySummary = (): StorefrontRouteCatalogCheckSummary => ({
  checked: 0,
  eligible: 0,
  remaining: 0,
  ok: 0,
  redirects: 0,
  not_found: 0,
  server_errors: 0,
  canonical_mismatch: 0,
  errors: 0,
})

const readError = (error: unknown): string => {
  if (error && typeof error === 'object') {
    const responseData = (error as { response?: { data?: unknown } }).response?.data
    if (responseData && typeof responseData === 'object') {
      const record = responseData as Record<string, unknown>
      if (typeof record.error === 'string' && record.error.trim()) return record.error.trim()
      if (typeof record.message === 'string' && record.message.trim()) return record.message.trim()
    }
  }
  return error instanceof Error ? error.message : '未知错误'
}

const isActiveStatus = (status: StorefrontRouteCatalogCheckTaskStatus | null): boolean => (
  status === 'queued' || status === 'running'
)

export const useURLOperationStore = defineStore('url-operation', () => {
  const running = ref(false)
  const taskId = ref('')
  const status = ref<StorefrontRouteCatalogCheckTaskStatus | null>(null)
  const locale = ref('')
  const checked = ref(0)
  const eligible = ref(0)
  const remaining = ref(0)
  const summary = ref<StorefrontRouteCatalogCheckSummary>(emptySummary())
  const error = ref('')
  const startedAt = ref<string | null>(null)
  const updatedAt = ref<string | null>(null)
  const endedAt = ref<string | null>(null)
  let runToken = 0

  const percentage = computed(() => (
    eligible.value > 0
      ? Math.min(100, Math.round((checked.value / eligible.value) * 100))
      : status.value === 'completed' ? 100 : 0
  ))

  const estimatedSeconds = computed(() => {
    if (!startedAt.value || checked.value <= 0 || remaining.value <= 0) return 0
    const elapsedSeconds = Math.max(0, (Date.now() - new Date(startedAt.value).getTime()) / 1000)
    return Math.max(1, Math.ceil((elapsedSeconds / checked.value) * remaining.value))
  })

  const progress = computed(() => ({
    checked: checked.value,
    eligible: eligible.value,
    remaining: remaining.value,
    percentage: percentage.value,
    estimatedSeconds: estimatedSeconds.value,
  }))

  const applyTask = (task: StorefrontRouteCatalogCheckTask): void => {
    taskId.value = task.task_id || taskId.value
    status.value = task.status || status.value
    locale.value = task.locale || locale.value
    checked.value = Number(task.checked || task.summary?.checked || 0)
    eligible.value = Number(task.eligible || task.summary?.eligible || 0)
    remaining.value = Number(task.remaining ?? task.summary?.remaining ?? 0)
    summary.value = { ...emptySummary(), ...(task.summary || {}) }
    error.value = task.error || ''
    startedAt.value = task.started_at || startedAt.value
    updatedAt.value = task.updated_at || updatedAt.value
    endedAt.value = task.ended_at || endedAt.value
    running.value = isActiveStatus(status.value)
  }

  const wait = (milliseconds: number): Promise<void> => new Promise((resolve) => {
    window.setTimeout(resolve, milliseconds)
  })

  const reset = (): void => {
    runToken += 1
    running.value = false
    taskId.value = ''
    status.value = null
    locale.value = ''
    checked.value = 0
    eligible.value = 0
    remaining.value = 0
    summary.value = emptySummary()
    error.value = ''
    startedAt.value = null
    updatedAt.value = null
    endedAt.value = null
  }

  const run = async (params: StartCheckParams, requestedLocale = ''): Promise<boolean> => {
    if (running.value) return false

    const token = ++runToken
    running.value = true
    error.value = ''
    status.value = 'queued'
    locale.value = requestedLocale
    try {
      const task = await storefrontRouteCatalogApi.startCheck(params)
      if (token !== runToken) return false
      applyTask(task)

      while (token === runToken && isActiveStatus(status.value) && taskId.value) {
        await wait(500)
        if (token !== runToken) return false
        applyTask(await storefrontRouteCatalogApi.checkStatus(taskId.value))
      }

      if (token !== runToken) return false
      const finalStatus = status.value as StorefrontRouteCatalogCheckTaskStatus | null
      if (finalStatus === 'completed') {
        toast.success(`URL 检查已完成：${checked.value} 条`)
        return true
      }
      if (finalStatus === 'failed') {
        toast.error(error.value ? `URL 检查失败：${error.value}` : 'URL 检查失败')
      }
      return false
    } catch (caught) {
      if (token !== runToken) return false
      error.value = readError(caught)
      status.value = 'failed'
      running.value = false
      toast.error(`URL 检查失败：${error.value}`)
      return false
    }
  }

  return {
    running,
    taskId,
    status,
    locale,
    checked,
    eligible,
    remaining,
    summary,
    error,
    startedAt,
    updatedAt,
    endedAt,
    percentage,
    estimatedSeconds,
    progress,
    run,
    reset,
  }
})
