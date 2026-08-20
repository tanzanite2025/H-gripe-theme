import process from 'node:process'
import path from 'node:path'
import { existsSync } from 'node:fs'
import { fileURLToPath, pathToFileURL } from 'node:url'

const scriptDir = path.dirname(fileURLToPath(import.meta.url))
const projectRoot = path.resolve(scriptDir, '../..')
const serverEntry = path.resolve(projectRoot, '.output/server/index.mjs')
const args = process.argv.slice(2)

const flagValue = (...names) => {
  for (const name of names) {
    const inlineValue = args.find(value => value.startsWith(`${name}=`))
    if (inlineValue) return inlineValue.slice(name.length + 1)

    const index = args.indexOf(name)
    if (index >= 0 && args[index + 1]) return args[index + 1]
  }

  return ''
}

const host = flagValue('--host', '--hostname') || process.env.HOST || process.env.NITRO_HOST || '127.0.0.1'
const port = flagValue('--port', '-p') || process.env.PORT || process.env.NITRO_PORT || '3000'

process.env.HOST = host
process.env.NITRO_HOST = host
process.env.PORT = port
process.env.NITRO_PORT = port

if (!existsSync(serverEntry)) {
  throw new Error(`Unable to resolve production server entry: ${serverEntry}`)
}

globalThis._importMeta_ = {
  url: pathToFileURL(serverEntry).href,
  env: process.env,
}

await import(pathToFileURL(serverEntry).href)
