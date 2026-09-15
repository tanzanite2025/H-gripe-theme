/**
 * The spoke catalog is no longer synchronized into source code. CAD geometry
 * must remain in the Go service. This command is retained as a compatibility
 * check for CI and only verifies that the public projection is reachable.
 */
const apiBase = (process.env.GO_API_URL || process.env.NUXT_PUBLIC_API_BASE || 'http://localhost:8080/api/v1').replace(/\/$/, '')
const apiUrl = process.env.SPOKE_API_URL || `${apiBase}/spoke/catalog/export`

try {
  const response = await fetch(apiUrl)
  if (!response.ok) throw new Error(`API responded with ${response.status}: ${response.statusText}`)
  const payload = await response.json() as { rims?: unknown[]; hubs?: unknown[] }
  console.log(`Public spoke catalog reachable (${payload.rims?.length || 0} rim brands, ${payload.hubs?.length || 0} hub brands). No source files were modified.`)
} catch (error: unknown) {
  console.error('Spoke catalog check failed:', error instanceof Error ? error.message : error)
  process.exit(1)
}
