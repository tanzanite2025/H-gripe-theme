import crypto from 'node:crypto'
import http from 'node:http'

import {
  cleanupOrphanLighthouseBrowsers,
  LighthouseExecutionError,
  runLighthouseInWorker,
} from './lighthouse.mjs'
import { RunnerInputError, loadRunnerConfig, normalizeRunInput } from './validation.mjs'

const config = loadRunnerConfig()
const port = Number.parseInt(process.env.PORT || '8080', 10)
let activeRuns = 0

const server = http.createServer(async (request, response) => {
  if (request.method === 'GET' && request.url === '/healthz') {
    writeJSON(response, 200, { status: 'ok', active_runs: activeRuns })
    return
  }

  if (request.method !== 'POST' || request.url !== '/v1/lighthouse/run') {
    writeJSON(response, 404, { error: { code: 'not_found', message: 'not found' } })
    return
  }
  if (!validToken(request.headers.authorization, config.token)) {
    writeJSON(response, 401, { error: { code: 'unauthorized', message: 'runner authentication failed' } })
    return
  }
  if (activeRuns >= config.maxConcurrency) {
    writeJSON(response, 429, { error: { code: 'runner_busy', message: 'runner capacity is currently exhausted' } })
    return
  }

  activeRuns++
  let startedRun = false
  try {
    const input = await readJSON(request)
    const normalized = await normalizeRunInput(input, config)
    startedRun = true

    if (config.maxConcurrency === 1) {
      const cleanedCount = await cleanupOrphanLighthouseBrowsers()
      if (cleanedCount > 0) {
        console.warn(`Cleaned ${cleanedCount} stale Chromium process(es) before running Lighthouse`)
      }
    }

    const report = await runLighthouseInWorker(normalized, config)
    writeJSON(response, 200, report)
  } catch (error) {
    const statusCode = error instanceof RunnerInputError
      ? 400
      : error instanceof LighthouseExecutionError
        ? error.statusCode
        : 500
    writeJSON(response, statusCode, {
      error: {
        code: error instanceof RunnerInputError ? 'invalid_input' : 'lighthouse_failed',
        message: error instanceof Error ? error.message : 'site quality runner failed',
      },
    })
  } finally {
    if (startedRun) {
      activeRuns--
    } else if (activeRuns > 0) {
      activeRuns--
    }
  }
})

server.requestTimeout = config.timeoutMilliseconds + 10_000
server.headersTimeout = 15_000
server.keepAliveTimeout = 5_000
server.listen(port, '0.0.0.0', () => {
  console.log(`Site Quality runner listening on port ${port}`)
})

function validToken(header, expected) {
  const actual = String(header || '').replace(/^Bearer\s+/i, '').trim()
  if (!actual || actual.length !== expected.length) {
    return false
  }
  return crypto.timingSafeEqual(Buffer.from(actual), Buffer.from(expected))
}

function readJSON(request) {
  return new Promise((resolve, reject) => {
    const chunks = []
    let size = 0
    request.on('data', (chunk) => {
      size += chunk.length
      if (size > 64 * 1024) {
        reject(new RunnerInputError('request body is too large'))
        request.destroy()
        return
      }
      chunks.push(chunk)
    })
    request.on('end', () => {
      try {
        resolve(JSON.parse(Buffer.concat(chunks).toString('utf8')))
      } catch {
        reject(new RunnerInputError('request body must be valid JSON'))
      }
    })
    request.on('error', reject)
  })
}

function writeJSON(response, statusCode, body) {
  if (response.writableEnded) {
    return
  }
  response.writeHead(statusCode, { 'Content-Type': 'application/json; charset=utf-8', 'Cache-Control': 'no-store' })
  response.end(JSON.stringify(body))
}
