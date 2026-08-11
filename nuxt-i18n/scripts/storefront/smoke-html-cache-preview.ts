import { spawn } from 'node:child_process'
import { existsSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'

interface PurgePayload {
  ok?: boolean
  storageBase?: string
  purgedKeys?: number
}

interface VerifiedPurgePayload {
  ok: true
  storageBase: '/cache/html'
  purgedKeys: number
}

const scriptDir = dirname(fileURLToPath(import.meta.url))
const projectRoot = resolve(scriptDir, '../..')
const serverEntry = resolve(projectRoot, '.output/server/index.mjs')

const port = Number.parseInt(process.env.HTML_CACHE_SMOKE_PORT || '4020', 10)
const token = process.env.NUXT_HTML_CACHE_PURGE_TOKEN || 'codex-html-cache-smoke-token'
const targetPath = process.env.HTML_CACHE_SMOKE_PATH || '/support/shipping'
const driver = String(process.env.HTML_CACHE_SMOKE_DRIVER || process.env.NUXT_HTML_CACHE_DRIVER || 'memory').toLowerCase()
const redisPrefix = process.env.HTML_CACHE_SMOKE_REDIS_PREFIX || process.env.NUXT_HTML_CACHE_PREFIX || 'commerce-platform:storefront:html-cache:smoke'
const origin = `http://127.0.0.1:${port}`

const timeout = (ms: number): Promise<void> => new Promise((resolveTimeout) => setTimeout(resolveTimeout, ms))

const fail = (message: string): void => {
  console.error(`[html-cache-smoke] FAILED: ${message}`)
  process.exitCode = 1
}

const waitForReady = async (): Promise<void> => {
  for (let attempt = 0; attempt < 60; attempt += 1) {
    try {
      const response = await fetch(`${origin}${targetPath}`)
      if (response.status < 500) return
    } catch {
      // Server is still booting.
    }
    await timeout(250)
  }
  throw new Error(`Nuxt preview did not become ready on ${origin}`)
}

const requestCachedPage = async (): Promise<string> => {
  const response = await fetch(`${origin}${targetPath}`)
  const cacheControl = response.headers.get('cache-control') || ''
  const body = await response.text()

  if (response.status !== 200) {
    throw new Error(`${targetPath} returned HTTP ${response.status}`)
  }
  if (!cacheControl.includes('s-maxage=')) {
    throw new Error(`${targetPath} did not return an SSR route cache header: ${cacheControl || '(empty)'}`)
  }
  if (!body.trim()) {
    throw new Error(`${targetPath} returned an empty body`)
  }

  return cacheControl
}

const purge = async (): Promise<VerifiedPurgePayload> => {
  const response = await fetch(`${origin}/_internal/html-cache/purge`, {
    method: 'POST',
    headers: {
      'content-type': 'application/json',
      'x-html-cache-purge-token': token,
    },
    body: JSON.stringify({ reason: 'html-cache-smoke' }),
  })
  const payload = await response.json().catch(() => null) as PurgePayload | null

  if (response.status !== 200) {
    throw new Error(`purge returned HTTP ${response.status}`)
  }
  if (!payload?.ok) {
    throw new Error(`purge response did not include ok=true: ${JSON.stringify(payload)}`)
  }
  const storageBase = payload.storageBase
  if (storageBase !== '/cache/html') {
    throw new Error(`purge storageBase mismatch: ${storageBase}`)
  }
  const purgedKeys = payload.purgedKeys
  if (typeof purgedKeys !== 'number' || !Number.isInteger(purgedKeys) || purgedKeys < 1) {
    throw new Error(`purge did not remove a cached HTML key: ${JSON.stringify(payload)}`)
  }

  return {
    ok: true,
    storageBase,
    purgedKeys,
  }
}

const getRedisSmokeEnv = (): NodeJS.ProcessEnv => {
  if (driver !== 'redis') return {}

  return {
    NUXT_HTML_CACHE_PREFIX: redisPrefix,
    NUXT_HTML_CACHE_REDIS_HOST: process.env.HTML_CACHE_SMOKE_REDIS_HOST || process.env.NUXT_HTML_CACHE_REDIS_HOST || '127.0.0.1',
    NUXT_HTML_CACHE_REDIS_PORT: process.env.HTML_CACHE_SMOKE_REDIS_PORT || process.env.NUXT_HTML_CACHE_REDIS_PORT || '6379',
    NUXT_HTML_CACHE_REDIS_DB: process.env.HTML_CACHE_SMOKE_REDIS_DB || process.env.NUXT_HTML_CACHE_REDIS_DB || '1',
    NUXT_HTML_CACHE_REDIS_TTL_SECONDS: process.env.HTML_CACHE_SMOKE_REDIS_TTL_SECONDS || process.env.NUXT_HTML_CACHE_REDIS_TTL_SECONDS || '604800',
    NUXT_HTML_CACHE_REDIS_SCAN_COUNT: process.env.HTML_CACHE_SMOKE_REDIS_SCAN_COUNT || process.env.NUXT_HTML_CACHE_REDIS_SCAN_COUNT || '100',
  }
}

const assertRuntimeLogs = (logs: string[]): void => {
  const output = logs.join('')

  if (driver === 'redis' && !output.includes('Nitro HTML route cache mounted on Redis')) {
    throw new Error('Redis smoke did not mount Nitro HTML route cache on Redis')
  }

  const cacheReadErrorPattern = /Cache read error|Stream isn't writeable|Redis cache mount failed/
  if (cacheReadErrorPattern.test(output)) {
    throw new Error('Preview logged an HTML cache startup/read error')
  }
}

if (!existsSync(serverEntry)) {
  fail(`Missing ${serverEntry}. Run npm run build first.`)
} else {
  const child = spawn(process.execPath, [serverEntry], {
    cwd: projectRoot,
    env: {
      ...process.env,
      NODE_ENV: 'production',
      HOST: '127.0.0.1',
      PORT: String(port),
      NITRO_PORT: String(port),
      NUXT_HTML_CACHE_DRIVER: driver,
      NUXT_HTML_CACHE_PURGE_TOKEN: token,
      ...getRedisSmokeEnv(),
    },
    stdio: ['ignore', 'pipe', 'pipe'],
    windowsHide: true,
  })

  const logs: string[] = []
  const capture = (chunk: Buffer): void => {
    logs.push(String(chunk))
    if (logs.length > 20) logs.shift()
  }
  child.stdout.on('data', capture)
  child.stderr.on('data', capture)

  try {
    await waitForReady()
    const cacheControl = await requestCachedPage()
    const payload = await purge()
    assertRuntimeLogs(logs)
    console.log(`[html-cache-smoke] OK: driver=${driver}, ${targetPath} cache-control="${cacheControl}", purgedKeys=${payload.purgedKeys}`)
  } catch (error: unknown) {
    fail(error instanceof Error ? error.message : String(error))
    if (logs.length > 0) {
      console.error('[html-cache-smoke] Recent server output:')
      console.error(logs.join('').trim())
    }
  } finally {
    child.kill('SIGTERM')
    setTimeout(() => {
      if (!child.killed) child.kill('SIGKILL')
    }, 1000).unref()
  }
}
