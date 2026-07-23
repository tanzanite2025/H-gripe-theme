import { spawn } from 'node:child_process'
import { existsSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'

const scriptDir = dirname(fileURLToPath(import.meta.url))
const projectRoot = resolve(scriptDir, '../..')
const serverEntry = resolve(projectRoot, '.output/server/index.mjs')

const port = Number.parseInt(process.env.HTML_CACHE_SMOKE_PORT || '4020', 10)
const token = process.env.NUXT_HTML_CACHE_PURGE_TOKEN || 'codex-html-cache-smoke-token'
const targetPath = process.env.HTML_CACHE_SMOKE_PATH || '/support/shipping'
const origin = `http://127.0.0.1:${port}`

const timeout = (ms) => new Promise(resolveTimeout => setTimeout(resolveTimeout, ms))

const fail = (message) => {
  console.error(`[html-cache-smoke] FAILED: ${message}`)
  process.exitCode = 1
}

const waitForReady = async () => {
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

const requestCachedPage = async () => {
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

const purge = async () => {
  const response = await fetch(`${origin}/_internal/html-cache/purge`, {
    method: 'POST',
    headers: {
      'content-type': 'application/json',
      'x-html-cache-purge-token': token,
    },
    body: JSON.stringify({ reason: 'html-cache-smoke' }),
  })
  const payload = await response.json().catch(() => null)

  if (response.status !== 200) {
    throw new Error(`purge returned HTTP ${response.status}`)
  }
  if (!payload?.ok) {
    throw new Error(`purge response did not include ok=true: ${JSON.stringify(payload)}`)
  }
  if (payload.storageBase !== '/cache/html') {
    throw new Error(`purge storageBase mismatch: ${payload.storageBase}`)
  }
  if (!Number.isInteger(payload.purgedKeys) || payload.purgedKeys < 1) {
    throw new Error(`purge did not remove a cached HTML key: ${JSON.stringify(payload)}`)
  }

  return payload
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
      NUXT_HTML_CACHE_DRIVER: 'memory',
      NUXT_HTML_CACHE_PURGE_TOKEN: token,
    },
    stdio: ['ignore', 'pipe', 'pipe'],
    windowsHide: true,
  })

  const logs = []
  const capture = (chunk) => {
    logs.push(String(chunk))
    if (logs.length > 20) logs.shift()
  }
  child.stdout.on('data', capture)
  child.stderr.on('data', capture)

  try {
    await waitForReady()
    const cacheControl = await requestCachedPage()
    const payload = await purge()
    console.log(`[html-cache-smoke] OK: ${targetPath} cache-control="${cacheControl}", purgedKeys=${payload.purgedKeys}`)
  } catch (error) {
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
