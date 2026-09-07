import { createServer } from 'node:net'

const isUsablePort = (port: number): boolean => (
  Number.isInteger(port) && port > 0 && port <= 65535
)

const canListenOnPort = (port: number): Promise<boolean> => new Promise((resolve) => {
  const server = createServer()

  server.once('error', () => {
    resolve(false)
  })
  server.once('listening', () => {
    server.close(() => resolve(true))
  })
  server.unref()
  server.listen(port, '127.0.0.1')
})

const randomPreviewPort = (): number => 43000 + Math.floor(Math.random() * 20000)

export async function resolvePreviewPort(
  envPort: string | undefined,
  preferredPort: number,
): Promise<number> {
  if (envPort) {
    const parsedEnvPort = Number.parseInt(envPort, 10)
    if (!isUsablePort(parsedEnvPort)) {
      throw new Error(`Invalid preview port: ${envPort}`)
    }
    return parsedEnvPort
  }

  if (await canListenOnPort(preferredPort)) {
    return preferredPort
  }

  for (let attempt = 0; attempt < 20; attempt += 1) {
    const candidate = randomPreviewPort()
    if (await canListenOnPort(candidate)) return candidate
  }

  throw new Error('Unable to find an available localhost preview port')
}
