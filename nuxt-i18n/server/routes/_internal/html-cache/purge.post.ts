import { timingSafeEqual } from 'node:crypto'
import { createError, defineEventHandler, getHeader, readBody, setHeader } from 'h3'
import { useStorage } from 'nitropack/runtime'
import {
  htmlCachePurgeHeader,
  htmlRouteCacheStorageBase,
} from '../../../../config/storefront/html-cache-runtime'

interface PurgeHtmlCachePayload {
  reason?: string
}

const purgeDeleteBatchSize = 100

const getExpectedToken = () => {
  return String(process.env.NUXT_HTML_CACHE_PURGE_TOKEN || process.env.HTML_CACHE_PURGE_TOKEN || '').trim()
}

const getProvidedToken = (authorizationHeader: string | undefined, purgeHeader: string | undefined) => {
  const headerToken = String(purgeHeader || '').trim()
  if (headerToken) return headerToken

  const authorization = String(authorizationHeader || '').trim()
  const match = authorization.match(/^Bearer\s+(.+)$/i)
  return match?.[1]?.trim() || ''
}

const safeTokenEquals = (provided: string, expected: string) => {
  if (!provided || !expected) return false

  const providedBuffer = Buffer.from(provided)
  const expectedBuffer = Buffer.from(expected)
  if (providedBuffer.length !== expectedBuffer.length) return false

  return timingSafeEqual(providedBuffer, expectedBuffer)
}

export default defineEventHandler(async (event) => {
  setHeader(event, 'cache-control', 'no-store, max-age=0')

  const expectedToken = getExpectedToken()
  if (!expectedToken) {
    throw createError({
      statusCode: 404,
      statusMessage: 'Not Found',
    })
  }

  const providedToken = getProvidedToken(
    getHeader(event, 'authorization'),
    getHeader(event, htmlCachePurgeHeader),
  )

  if (!safeTokenEquals(providedToken, expectedToken)) {
    throw createError({
      statusCode: 401,
      statusMessage: 'Unauthorized',
    })
  }

  const payload = await readBody<PurgeHtmlCachePayload>(event).catch(() => ({}))
  const reason = String(payload?.reason || 'manual').slice(0, 160)

  const storage = useStorage()
  const keys = await storage.getKeys(htmlRouteCacheStorageBase)
  for (let index = 0; index < keys.length; index += purgeDeleteBatchSize) {
    const batch = keys.slice(index, index + purgeDeleteBatchSize)
    await Promise.all(batch.map(key => storage.removeItem(key)))
  }

  return {
    ok: true,
    scope: 'storefront-html',
    storageBase: htmlRouteCacheStorageBase,
    purgedKeys: keys.length,
    reason,
    purgedAt: new Date().toISOString(),
  }
})
