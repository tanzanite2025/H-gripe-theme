import { useApiRequest } from '~/composables/useApiRequest'
import type { WorkbenchFeedListResult } from './types'

interface ApiEnvelope {
  data?: WorkbenchFeedListResult
}

const normalizeResult = (payload: unknown): WorkbenchFeedListResult => {
  const envelope = payload as ApiEnvelope | undefined
  const result = envelope?.data
  if (!result || !Array.isArray(result.entries) || !result.pagination) {
    throw new Error('Invalid Workbench Feed response')
  }
  return result
}

export const useWorkbenchFeedApi = () => {
  const { request } = useApiRequest()

  const list = async (options: {
    locale?: string
    page?: number
    pageSize?: number
    tag?: string
  } = {}): Promise<WorkbenchFeedListResult> => {
    const payload = await request<unknown>('/workbench-feed/entries', {
      params: {
        locale: options.locale || undefined,
        page: options.page || 1,
        page_size: options.pageSize || 10,
        tag: options.tag || undefined,
      },
    }, 'Failed to load Workbench Feed')
    return normalizeResult(payload)
  }

  return { list }
}
