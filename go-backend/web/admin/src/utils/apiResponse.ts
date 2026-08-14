export interface ApiPagination {
  page: number
  page_size: number
  total: number
  total_pages: number
}

export type ApiRecord = Record<string, any>

interface ApiResponseLike {
  data?: unknown
  status?: number
}

const hasOwn = (value: ApiRecord, key: string): boolean => (
  Object.prototype.hasOwnProperty.call(value, key)
)

const criticalResponseError = (endpoint: string, detail: string): never => {
  throw new Error(`[CRITICAL] Invalid API response for ${endpoint}: ${detail}`)
}

const describeValue = (value: unknown): string => {
  if (value === null) return 'null'
  if (Array.isArray(value)) return 'array'
  return typeof value
}

const isRecord = (value: unknown): value is ApiRecord => (
  typeof value === 'object' && value !== null && !Array.isArray(value)
)

export const readApiBody = (response: unknown, endpoint: string): unknown => {
  if (!isRecord(response) || !hasOwn(response, 'data')) {
    return criticalResponseError(endpoint, 'response body is missing')
  }

  const body = (response as ApiResponseLike).data
  if (body === undefined || body === null || body === '') {
    if ((response as ApiResponseLike).status === 204) return undefined
    return criticalResponseError(endpoint, `response body is ${describeValue(body)}`)
  }

  if (isRecord(body) && hasOwn(body, 'code') && body.code !== 0 && body.code !== '0') {
    return criticalResponseError(endpoint, `API returned non-success code ${String(body.code)}`)
  }

  return body
}

export const unwrapApiPayload = (response: unknown, endpoint: string): unknown => {
  const body = readApiBody(response, endpoint)
  if (isRecord(body) && hasOwn(body, 'data')) return body.data
  return body
}

export const requireApiObject = (value: unknown, endpoint: string, label = 'payload'): ApiRecord => {
  if (!isRecord(value)) {
    return criticalResponseError(endpoint, `${label} must be an object, received ${describeValue(value)}`)
  }
  return value
}

export const requireApiArray = <T = any>(value: unknown, endpoint: string, label = 'payload'): T[] => {
  if (!Array.isArray(value)) {
    return criticalResponseError(endpoint, `${label} must be an array, received ${describeValue(value)}`)
  }
  return value as T[]
}

export const requireApiField = <T = any>(value: unknown, field: string, endpoint: string): T => {
  const object = requireApiObject(value, endpoint)
  if (!hasOwn(object, field) || object[field] === undefined) {
    return criticalResponseError(endpoint, `required field "${field}" is missing`)
  }
  return object[field] as T
}

export const requireApiObjectField = <T = any>(value: unknown, field: string, endpoint: string): T => (
  requireApiObject(requireApiField(value, field, endpoint), endpoint, `field "${field}"`) as T
)

export const requireApiArrayField = <T = any>(value: unknown, field: string, endpoint: string): T[] => (
  requireApiArray<T>(requireApiField(value, field, endpoint), endpoint, `field "${field}"`)
)

export const requireApiBooleanField = (value: unknown, field: string, endpoint: string): boolean => {
  const fieldValue = requireApiField(value, field, endpoint)
  if (typeof fieldValue !== 'boolean') {
    return criticalResponseError(endpoint, `field "${field}" must be a boolean, received ${describeValue(fieldValue)}`)
  }
  return fieldValue
}

export const requireApiNumberField = (value: unknown, field: string, endpoint: string): number => {
  const fieldValue = requireApiField(value, field, endpoint)
  if (typeof fieldValue !== 'number' || !Number.isFinite(fieldValue)) {
    return criticalResponseError(endpoint, `field "${field}" must be a finite number, received ${describeValue(fieldValue)}`)
  }
  return fieldValue
}

export const requireApiStringField = (value: unknown, field: string, endpoint: string): string => {
  const fieldValue = requireApiField(value, field, endpoint)
  if (typeof fieldValue !== 'string') {
    return criticalResponseError(endpoint, `field "${field}" must be a string, received ${describeValue(fieldValue)}`)
  }
  return fieldValue
}

export const requireApiAcknowledgement = (response: unknown, endpoint: string): ApiRecord => {
  const body = requireApiObject(readApiBody(response, endpoint), endpoint, 'response body')
  if (Object.keys(body).length === 0) {
    return criticalResponseError(endpoint, 'response body is empty')
  }
  return body
}

export const requireApiSuccess = (response: unknown, endpoint: string): ApiRecord | undefined => {
  const body = readApiBody(response, endpoint)
  if (body === undefined) return undefined
  return requireApiAcknowledgement({ data: body }, endpoint)
}

export const requireApiPagination = (
  responseBody: unknown,
  payload: unknown,
  endpoint: string,
): ApiPagination => {
  const body = requireApiObject(responseBody, endpoint, 'response body')
  const payloadObject = isRecord(payload) ? payload : null
  const pagination = hasOwn(body, 'pagination')
    ? requireApiObjectField(body, 'pagination', endpoint)
    : payloadObject
      ? requireApiObjectField(payloadObject, 'pagination', endpoint)
      : criticalResponseError(endpoint, 'required field "pagination" is missing')

  return {
    page: requireApiNumberField(pagination, 'page', endpoint),
    page_size: requireApiNumberField(pagination, 'page_size', endpoint),
    total: requireApiNumberField(pagination, 'total', endpoint),
    total_pages: requireApiNumberField(pagination, 'total_pages', endpoint),
  }
}
