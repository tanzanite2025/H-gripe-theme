import type { LocationQuery, LocationQueryRaw } from 'vue-router'
import type { OpsEnvironment } from '@/api/ops'

export const opsEnvironmentOptions = [
  { value: 'production', label: '生产' },
  { value: 'staging', label: '预发布' },
  { value: 'test', label: '测试' },
  { value: 'local', label: '本地' },
]

export const isOpsEnvironment = (value: unknown): value is OpsEnvironment => (
  value === 'production' || value === 'staging' || value === 'test' || value === 'local'
)

export const readOpsEnvironmentQuery = (
  value: unknown,
  fallback: OpsEnvironment | '' = 'production',
): OpsEnvironment | '' => {
  const candidate = Array.isArray(value) ? value[0] : value
  if (candidate === 'all') return ''
  return isOpsEnvironment(candidate) ? candidate : fallback
}

export const withOpsEnvironmentQuery = (
  query: LocationQuery,
  environment: OpsEnvironment | '',
): LocationQueryRaw => {
  const { environment: _environment, ...rest } = query
  return {
    ...rest,
    environment: environment || 'all',
  }
}
