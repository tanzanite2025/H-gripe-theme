type AdminStatusTone = 'blue' | 'green' | 'amber' | 'coral' | 'gray'

export const sourceLabel = (value: string): string => ({
  static: '静态',
  product: '产品',
  blog: 'Blog',
  alias: '兼容',
}[value] || value || '未知')

const sourceToneMap: Record<string, AdminStatusTone> = {
  static: 'blue',
  product: 'green',
  blog: 'amber',
  alias: 'gray',
}

export const sourceTone = (value: string): AdminStatusTone => sourceToneMap[value] || 'gray'

export const entryLabel = (value: string): string => ({
  active: '有效',
  alias: '兼容',
  duplicate: '重复',
  stale: '失效',
}[value] || value || '未同步')

const entryToneMap: Record<string, AdminStatusTone> = {
  active: 'green',
  alias: 'gray',
  duplicate: 'coral',
  stale: 'amber',
}

export const entryTone = (value: string): AdminStatusTone => entryToneMap[value] || 'gray'

export const checkLabel = (status?: string | null, httpStatus?: number | null): string => {
  if (!status) return '未检查'
  if (status === 'ok') return httpStatus ? `正常 ${httpStatus}` : '正常'
  if (status === 'redirect') return httpStatus ? `跳转 ${httpStatus}` : '跳转'
  if (status === 'redirect_chain') return '重定向链'
  if (status === 'redirect_target_mismatch') return '跳转目标不一致'
  if (status === 'not_found') return '404'
  if (status === 'server_error') return httpStatus ? `5xx ${httpStatus}` : '5xx'
  if (status === 'canonical_mismatch') return 'Canonical'
  return '失败'
}

const checkToneMap: Record<string, AdminStatusTone> = {
  ok: 'green',
  redirect: 'amber',
  redirect_chain: 'coral',
  redirect_target_mismatch: 'coral',
  not_found: 'coral',
  server_error: 'coral',
  canonical_mismatch: 'amber',
  error: 'coral',
}

export const checkTone = (status?: string | null): AdminStatusTone => checkToneMap[status || ''] || 'gray'

export const formatRouteCatalogDate = (value?: string | null): string => {
  if (!value) return '未同步'
  const date = new Date(value)
  return Number.isNaN(date.getTime()) ? value : date.toLocaleString('zh-CN')
}
