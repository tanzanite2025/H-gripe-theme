import { createError, defineEventHandler, getQuery, setHeader } from 'h3'
import { useImage, useRuntimeConfig } from '#imports'

const healthAssetPath = '/_internal/ipx-health.svg'
const healthTimeoutMs = 8_000

const localServerOrigin = () => {
  const port = Number(process.env.PORT || 3000)
  return `http://127.0.0.1:${Number.isInteger(port) && port > 0 ? port : 3000}`
}

export default defineEventHandler(async (event) => {
  setHeader(event, 'cache-control', 'no-store, max-age=0')

  const sourceMode = String(getQuery(event).source || '') === 'internal' ? 'internal' : 'core'
  const runtimeConfig = useRuntimeConfig()
  const internalImageOrigin = String(runtimeConfig.imageInternalOrigin || '').trim()

  if (sourceMode === 'internal' && !internalImageOrigin) {
    throw createError({
      statusCode: 503,
      statusMessage: 'Image pipeline source is not configured',
    })
  }

  const source = sourceMode === 'internal'
    ? new URL('/favicon.svg', internalImageOrigin).toString()
    : healthAssetPath

  try {
    const image = useImage(event).getImage(source, {
      modifiers: {
        width: 1,
        height: 1,
        format: 'webp',
      },
    })
    const transformedURL = new URL(image.url, localServerOrigin())

    if (!transformedURL.pathname.startsWith('/_ipx/')) {
      throw new Error('IPX did not generate a transform URL for the health source')
    }

    const response = await fetch(transformedURL, {
      headers: {
        accept: 'image/webp',
      },
      signal: AbortSignal.timeout(healthTimeoutMs),
    })
    const bytes = await response.arrayBuffer()
    const contentType = String(response.headers.get('content-type') || '').toLowerCase()

    if (!response.ok || !contentType.startsWith('image/webp') || bytes.byteLength === 0) {
      throw new Error(`Unexpected IPX response: status=${response.status}, content-type=${contentType || 'missing'}, bytes=${bytes.byteLength}`)
    }
  } catch (error) {
    const message = error instanceof Error ? error.message : String(error)
    console.error(`[image-health] ${sourceMode} transform check failed: ${message}`)
    throw createError({
      statusCode: 503,
      statusMessage: 'Image pipeline unavailable',
    })
  }

  return {
    ok: true,
    source: sourceMode,
    checkedAt: new Date().toISOString(),
  }
})
