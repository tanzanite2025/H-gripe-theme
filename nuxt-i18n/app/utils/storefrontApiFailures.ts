import { ApiRequestError } from '~/composables/useApiRequest'

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
