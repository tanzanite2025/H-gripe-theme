const API_SETTINGS_GROUP = 'api'
const API_SETTINGS_LOCALE = 'en'
const CSRF_COOKIE_NAME = 'csrf_token'
const CSRF_HEADER_NAME = 'X-CSRF-Token'
const DEFAULT_SAVE_TIMEOUT_MS = 12000

export type ApiSettingValue = string | number | boolean | null | undefined

export interface ApiSettingPayload {
  key: string
  value: string
  type: 'string' | 'number' | 'boolean' | 'json'
  group: typeof API_SETTINGS_GROUP
  locale: typeof API_SETTINGS_LOCALE
  is_public: false
  description: string
}

interface SaveAPISettingsOptions {
  label?: string
  timeoutMs?: number
}

interface ErrorPayload {
  message?: string
  error?: string
}

export const apiSettingPayload = (
  key: string,
  value: ApiSettingValue,
  type: ApiSettingPayload['type'],
  description: string,
): ApiSettingPayload => ({
  key,
  value: String(value ?? ''),
  type,
  group: API_SETTINGS_GROUP,
  locale: API_SETTINGS_LOCALE,
  is_public: false,
  description,
})

const readCookieValue = (name: string): string => {
  if (typeof document === 'undefined') return ''
  const prefix = `${encodeURIComponent(name)}=`
  const match = document.cookie
    .split(';')
    .map((item) => item.trim())
    .find((item) => item.startsWith(prefix))
  return match ? decodeURIComponent(match.slice(prefix.length)) : ''
}

const apiSettingsSaveHeaders = (): Record<string, string> => {
  const csrfToken = readCookieValue(CSRF_COOKIE_NAME)
  return {
    'Content-Type': 'application/json',
    Accept: 'application/json',
    ...(csrfToken ? { [CSRF_HEADER_NAME]: csrfToken } : {}),
  }
}

const responseErrorMessage = (status: number, payload: ErrorPayload | null, label: string): string => {
  const serverMessage = payload?.message || payload?.error
  if (serverMessage) return serverMessage
  if (status === 401) return '登录已过期，请重新登录后再保存'
  if (status === 403) return `当前账号没有 settings:edit 权限，无法保存${label}`
  if (status === 404) return '保存接口不存在，请确认 DEV 后端 9200 已启动并加载 admin settings 路由'
  if (status >= 500) return `后端保存${label}失败`
  return `${label}保存失败（HTTP ${status}）`
}

export const postApiSettingsBatch = async (
  settings: ApiSettingPayload[],
  options: SaveAPISettingsOptions = {},
): Promise<unknown> => {
  const label = options.label || 'API 设置'
  const controller = new AbortController()
  const timeoutID = window.setTimeout(
    () => controller.abort(),
    options.timeoutMs || DEFAULT_SAVE_TIMEOUT_MS,
  )
  try {
    const response = await fetch('/api/admin/settings/batch', {
      method: 'POST',
      credentials: 'include',
      headers: apiSettingsSaveHeaders(),
      body: JSON.stringify({ settings }),
      signal: controller.signal,
    })
    const payload = await response.json().catch(() => null) as ErrorPayload | null
    if (!response.ok) {
      throw new Error(responseErrorMessage(response.status, payload, label))
    }
    return payload
  } catch (error) {
    if (error instanceof DOMException && error.name === 'AbortError') {
      throw new Error(`${label}保存请求超时，请确认 DEV 后端 9200 正在运行且后台登录未过期`)
    }
    throw error
  } finally {
    window.clearTimeout(timeoutID)
  }
}
