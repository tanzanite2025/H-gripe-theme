import { ApiRequestError } from '~/composables/useApiRequest'

const readErrorMessage = (error: unknown) => {
  if (!error || typeof error !== 'object') {
    return ''
  }

  const dataMessage = (error as { data?: { message?: unknown } }).data?.message
  if (typeof dataMessage === 'string' && dataMessage.trim()) {
    return dataMessage.trim()
  }

  const message = (error as { message?: unknown }).message
  return typeof message === 'string' ? message.trim() : ''
}

export const toUserFacingApiError = (error: unknown, fallback: string) => {
  const message = readErrorMessage(error)
  if (
    !message
    || /failed to fetch|no response|network error/i.test(message)
    || /https?:\/\//i.test(message)
  ) {
    return fallback
  }

  return message
}

const toStatus = (error: unknown) => {
  if (error instanceof ApiRequestError) {
    return error.status
  }

  if (!error || typeof error !== 'object') {
    return 0
  }

  const status = (error as { status?: unknown }).status
  return typeof status === 'number' && Number.isFinite(status) ? status : 0
}

export const isApiStatus = (error: unknown, statuses: readonly number[]) => {
  const status = toStatus(error)
  return status > 0 && statuses.includes(status)
}

export const isExpectedAnonymousApiMiss = (error: unknown) =>
  isApiStatus(error, [401, 403])

export const isExpectedOptionalConfigMiss = (error: unknown) =>
  isApiStatus(error, [401, 403, 404, 410])

export const logUnexpectedApiError = (
  message: string,
  error: unknown,
  expected: (error: unknown) => boolean,
) => {
  if (expected(error)) {
    return
  }

  console.error(message, error)
}
