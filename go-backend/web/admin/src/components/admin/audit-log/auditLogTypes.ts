export type AuditLogTone = 'blue' | 'green' | 'amber' | 'coral' | 'gray'
export type AuditLogLabelResolver = (value?: string | null) => string
export type AuditLogToneResolver = (value?: string | null) => AuditLogTone
export type AuditLogDurationClassResolver = (value?: number | string | null) => string
export type AuditLogDateFormatter = (value?: string | null) => string

export interface AuditLogFilters {
  keyword: string
  action: string
  resource: string
  user_id: string
  ip_address: string
  start_date: string
  end_date: string
}

export interface AuditLogPagination {
  page: number
  pageSize: number
  total: number
}

export type AuditLogJsonValue = string | number | boolean | null | Record<string, unknown> | unknown[]

export interface AuditLogRecord {
  id: number | string
  username?: string | null
  user_id?: number | string | null
  action?: string | null
  resource?: string | null
  resource_id?: number | string | null
  method?: string | null
  path?: string | null
  ip_address?: string | null
  status?: string | null
  duration?: number | string | null
  created_at?: string | null
  user_agent?: string | null
  error_message?: string | null
  changes?: AuditLogJsonValue
  old_value?: AuditLogJsonValue
  new_value?: AuditLogJsonValue
}

export interface AuditLogStats {
  total_count?: number
  today_count?: number
  success_count?: number
  failed_count?: number
}

export interface AuditLogsResponse {
  logs?: AuditLogRecord[]
  total?: number
}

export interface AuditLogDetailResponse {
  log?: AuditLogRecord
}
