import { spawn } from 'node:child_process'
import net from 'node:net'
import path from 'node:path'

const host = process.env.NUXT_HOST || process.env.HOST || '127.0.0.1'
const port = Number(process.env.NUXT_PORT || process.env.PORT || 9100)

if (!Number.isInteger(port) || port <= 0 || port > 65535) {
  console.error(`Invalid dev port: ${process.env.NUXT_PORT || process.env.PORT}`)
  process.exit(1)
}

const checkPortAvailable = () => {
  return new Promise((resolve) => {
    const server = net.createServer()

    server.once('error', () => resolve(false))
    server.once('listening', () => {
      server.close(() => resolve(true))
    })
    server.listen(port, host)
  })
}

const isAvailable = await checkPortAvailable()

if (!isAvailable) {
  console.error(`Dev port ${host}:${port} is already in use. Stop that process first, then run npm run dev again.`)
  process.exit(1)
}

const nuxiBin = path.resolve('node_modules/@nuxt/cli/bin/nuxi.mjs')
const child = spawn(process.execPath, [nuxiBin, 'dev', '--host', host, '--port', String(port)], {
  env: process.env,
  stdio: 'inherit',
})

const stopChild = (signal) => {
  child.kill(signal)
}

process.on('SIGINT', stopChild)
process.on('SIGTERM', stopChild)

child.on('exit', (code, signal) => {
  if (signal) {
    process.kill(process.pid, signal)
    return
  }
  process.exit(code ?? 0)
})
