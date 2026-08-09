import { spawn } from 'node:child_process'
import net from 'node:net'
import path from 'node:path'

type PortCheckResult =
  | { available: true; error: null }
  | { available: false; error: NodeJS.ErrnoException }

const host = process.env.NUXT_HOST || process.env.HOST || '127.0.0.1'
const port = Number(process.env.NUXT_PORT || process.env.PORT || 9199)
const retryCount = Number(process.env.NUXT_PORT_CHECK_RETRIES || 20)
const retryDelayMs = Number(process.env.NUXT_PORT_CHECK_DELAY_MS || 500)

if (!Number.isInteger(port) || port <= 0 || port > 65535) {
  console.error(`Invalid dev port: ${process.env.NUXT_PORT || process.env.PORT}`)
  process.exit(1)
}

const sleep = (ms: number): Promise<void> => new Promise((resolve) => setTimeout(resolve, ms))

const checkPortAvailable = (): Promise<PortCheckResult> => {
  return new Promise((resolve) => {
    const server = net.createServer()

    server.once('error', (error: NodeJS.ErrnoException) => resolve({ available: false, error }))
    server.once('listening', () => {
      server.close(() => resolve({ available: true, error: null }))
    })
    server.listen(port, host)
  })
}

let result = await checkPortAvailable()
for (let attempt = 1; !result.available && attempt <= retryCount; attempt += 1) {
  await sleep(retryDelayMs)
  result = await checkPortAvailable()
}

if (!result.available) {
  const detail = result.error.code || result.error.message || 'unknown listen error'
  console.error(`Dev port ${host}:${port} is not available (${detail}). Stop that process first, then run npm run dev again.`)
  process.exit(1)
}

const nuxiBin = path.resolve('node_modules/@nuxt/cli/bin/nuxi.mjs')
const child = spawn(process.execPath, [nuxiBin, 'dev', '--host', host, '--port', String(port)], {
  env: process.env,
  stdio: 'inherit',
})

const stopChild = (signal: NodeJS.Signals): void => {
  child.kill(signal)
}

process.on('SIGINT', () => stopChild('SIGINT'))
process.on('SIGTERM', () => stopChild('SIGTERM'))

child.on('exit', (code, signal) => {
  if (signal) {
    process.kill(process.pid, signal)
    return
  }
  process.exit(code ?? 0)
})
