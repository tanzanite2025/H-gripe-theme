import { defineNitroPlugin, useStorage } from 'nitropack/runtime'
import redisDriver, { type RedisOptions as StorageRedisOptions } from 'unstorage/drivers/redis'

interface ReadyRedisClient {
  connect: () => Promise<unknown>
  ping: () => Promise<unknown>
  disconnect: () => void
}

const toInteger = (value: string | undefined, fallback: number) => {
  const parsed = Number.parseInt(String(value || ''), 10)
  return Number.isFinite(parsed) && parsed >= 0 ? parsed : fallback
}

const cleanRedisBase = (value: string | undefined) => {
  return String(value || 'h-gripe:storefront:html-cache').replace(/:+$/, '')
}

export default defineNitroPlugin(async () => {
  const driver = String(process.env.NUXT_HTML_CACHE_DRIVER || 'memory').toLowerCase()
  if (driver !== 'redis') return

  const url = process.env.NUXT_HTML_CACHE_REDIS_URL || process.env.REDIS_URL || ''
  const host = process.env.NUXT_HTML_CACHE_REDIS_HOST || process.env.REDIS_HOST || 'redis'
  const port = toInteger(process.env.NUXT_HTML_CACHE_REDIS_PORT || process.env.REDIS_PORT, 6379)
  const db = toInteger(process.env.NUXT_HTML_CACHE_REDIS_DB || process.env.REDIS_DB, 1)
  const ttl = toInteger(process.env.NUXT_HTML_CACHE_REDIS_TTL_SECONDS, 7 * 24 * 60 * 60)
  const base = cleanRedisBase(process.env.NUXT_HTML_CACHE_PREFIX)
  const connectTimeout = toInteger(process.env.NUXT_HTML_CACHE_REDIS_CONNECT_TIMEOUT_MS, 1000)
  const maxRetriesPerRequest = toInteger(process.env.NUXT_HTML_CACHE_REDIS_MAX_RETRIES, 1)
  const scanCount = toInteger(process.env.NUXT_HTML_CACHE_REDIS_SCAN_COUNT, 100)

  const connectionOptions: Record<string, any> = {
    connectTimeout,
    maxRetriesPerRequest,
    enableOfflineQueue: false,
    lazyConnect: true,
    retryStrategy: () => null,
  }

  const storageOptions: StorageRedisOptions = {
    ...connectionOptions,
    base,
    ttl,
    scanCount,
  }

  if (url) {
    storageOptions.url = url
  } else {
    storageOptions.host = host
    storageOptions.port = port
    storageOptions.db = db
    storageOptions.password = process.env.NUXT_HTML_CACHE_REDIS_PASSWORD || process.env.REDIS_PASSWORD || undefined
  }

  let redisClient: ReadyRedisClient | undefined

  try {
    const cacheDriver = redisDriver(storageOptions)
    redisClient = cacheDriver.getInstance?.() as ReadyRedisClient | undefined
    if (!redisClient) throw new Error('Redis cache driver did not expose a client instance')

    // Warm up the same Redis client Nitro will use before exposing the cache mount to first SSR requests.
    await redisClient.connect()
    await redisClient.ping()

    const storage = useStorage()
    await storage.unmount('cache')
    storage.mount('cache', cacheDriver)

    if (process.env.NUXT_HTML_CACHE_LOG !== 'silent') {
      const target = url ? 'redis-url' : `${host}:${port}/${db}`
      console.info(`[html-cache] Nitro HTML route cache mounted on Redis (${target}, base=${base}, ttl=${ttl}s)`)
    }
  } catch (error) {
    redisClient?.disconnect()
    console.warn('[html-cache] Redis cache mount failed; Nitro will use the default in-memory cache.', error)
  }
})
