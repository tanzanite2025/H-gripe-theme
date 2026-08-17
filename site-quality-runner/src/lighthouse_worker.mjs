import { LighthouseExecutionError, runLighthouse } from './lighthouse.mjs'

process.once('message', async (message) => {
  try {
    const result = await runLighthouse(message?.input, message?.config)
    sendResult({ ok: true, result }, 0)
  } catch (error) {
    sendResult({
      ok: false,
      error: {
        message: error instanceof Error ? error.message : String(error),
        statusCode: error instanceof LighthouseExecutionError ? error.statusCode : 502,
      },
    }, 1)
  }
})

function sendResult(payload, exitCode) {
  if (typeof process.send !== 'function') {
    process.exit(exitCode)
    return
  }
  process.send(payload, () => {
    process.exit(exitCode)
  })
}
