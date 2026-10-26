import { mkdir, writeFile } from 'node:fs/promises'
import { dirname } from 'node:path'
import { gzipSync } from 'node:zlib'
import type { Plugin } from 'vite'

const deferredPublicClientModulePatterns = [
  /\/components\/account\//,
  /\/components\/CheckoutModal\.vue(?:\?|$)/,
  /\/node_modules\/(?:country-region-data|zod|focus-trap|qrcode)\//,
  /\/node_modules\/@stripe\/stripe-js\//,
]

export function createPublicClientEntryJavascriptDependencyBoundaryPlugin(reportFilename: string): Plugin {
  return {
    name: 'storefront-public-client-entry-javascript-dependency-boundaries',
    apply: (config, environment) => environment.command === 'build' && !config.build?.ssr,
    async writeBundle(_options, bundle) {
      const pendingChunkFilenames = Object.values(bundle)
        .filter(output => output.type === 'chunk' && output.isEntry)
        .map(output => output.fileName)
      const visitedChunkFilenames = new Set<string>()
      const eagerDeferredModuleIds: string[] = []
      const chunks: Array<{ file: string; bytes: number; gzipBytes: number }> = []

      while (pendingChunkFilenames.length > 0) {
        const filename = pendingChunkFilenames.pop()!
        if (visitedChunkFilenames.has(filename)) continue
        visitedChunkFilenames.add(filename)
        const chunk = bundle[filename]
        if (!chunk || chunk.type !== 'chunk') continue

        chunks.push({
          file: filename,
          bytes: Buffer.byteLength(chunk.code),
          gzipBytes: gzipSync(chunk.code).byteLength,
        })
        // Dynamic imports belong to the feature that activates them. Only
        // follow imports that every visitor must fetch with the public entry.
        pendingChunkFilenames.push(...chunk.imports)
        for (const moduleId of Object.keys(chunk.modules)) {
          const normalizedModuleId = moduleId.replaceAll('\\', '/')
          if (deferredPublicClientModulePatterns.some(pattern => pattern.test(normalizedModuleId))) {
            eagerDeferredModuleIds.push(normalizedModuleId)
          }
        }
      }

      if (eagerDeferredModuleIds.length > 0) {
        this.error(`Public client entry eagerly includes deferred account or payment modules:\n${eagerDeferredModuleIds.join('\n')}`)
      }
      if (chunks.length === 0) this.error('Unable to inspect the public client entry dependency graph.')

      await mkdir(dirname(reportFilename), { recursive: true })
      await writeFile(reportFilename, JSON.stringify({
        scope: 'Global client entry and its static JavaScript dependencies; route and locale chunks are additional.',
        bytes: chunks.reduce((total, chunk) => total + chunk.bytes, 0),
        gzipBytes: chunks.reduce((total, chunk) => total + chunk.gzipBytes, 0),
        chunks,
      }, null, 2) + '\n')
    },
  }
}
