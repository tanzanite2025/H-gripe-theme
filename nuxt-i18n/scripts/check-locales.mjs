import fs from 'node:fs'
import path from 'node:path'
import process from 'node:process'

const apiBase =
  process.env.NUXT_PUBLIC_API_BASE ||
  process.env.GO_API_BASE ||
  process.env.API_BASE ||
  'https://tanzanite.site/api/v1'
const localesUrl = process.env.GO_LOCALES_URL || process.env.LOCALES_URL
const manifestPath = path.resolve('i18n/locales.manifest.js')

function loadManifestLocales() {
  if (!fs.existsSync(manifestPath)) {
    throw new Error(`Manifest not found: ${manifestPath}`)
  }

  const moduleUrl = `file://${manifestPath}`
  return import(moduleUrl).then((mod) => mod.default || mod.locales || [])
}

function resolveEndpoint() {
  if (localesUrl) return localesUrl
  const base = apiBase.replace(/\/$/, '')
  if (base.endsWith('/api/v1')) return `${base}/i18n/languages`
  return `${base}/api/v1/i18n/languages`
}

async function fetchBackendLocales() {
  const url = resolveEndpoint()
  const res = await fetch(url)
  if (!res.ok) {
    if (res.status === 404) {
      throw new Error(`Fetch backend locales failed: 404 Not Found. Check that the Go backend exposes ${url}. Override with GO_LOCALES_URL or LOCALES_URL if needed.`)
    }
    throw new Error(`Fetch backend locales failed: ${res.status} ${res.statusText} (url: ${url})`)
  }
  const data = await res.json()
  if (!data || !Array.isArray(data.languages)) {
    throw new Error('Unexpected backend locales response shape')
  }
  return data.languages
}

function toMap(list) {
  const map = new Map()
  for (const item of list) {
    if (!item || !item.code) continue
    map.set(String(item.code), item)
  }
  return map
}

function diffLocales(manifestList, backendList) {
  const manifestMap = toMap(manifestList)
  const backendMap = toMap(backendList)

  const missingInManifest = []
  const extraInManifest = []
  const nameMismatch = []

  for (const code of backendMap.keys()) {
    if (!manifestMap.has(code)) {
      missingInManifest.push(code)
    } else {
      const m = manifestMap.get(code)
      const b = backendMap.get(code)
      if (m?.name && b?.name && m.name !== b.name) {
        nameMismatch.push({ code, manifest: m.name, backend: b.name })
      }
    }
  }

  for (const code of manifestMap.keys()) {
    if (!backendMap.has(code)) {
      extraInManifest.push(code)
    }
  }

  return { missingInManifest, extraInManifest, nameMismatch }
}

async function main() {
  try {
    const [manifestLocales, backendLocales] = await Promise.all([
      loadManifestLocales(),
      fetchBackendLocales(),
    ])

    const { missingInManifest, extraInManifest, nameMismatch } = diffLocales(manifestLocales, backendLocales)

    if (
      missingInManifest.length === 0 &&
      extraInManifest.length === 0 &&
      nameMismatch.length === 0
    ) {
      console.log('Locales are aligned with Go backend source.')
      process.exit(0)
    }

    if (missingInManifest.length) {
      console.error('Missing in manifest (present in backend):', missingInManifest.join(', '))
    }

    if (extraInManifest.length) {
      console.error('Extra in manifest (absent in backend):', extraInManifest.join(', '))
    }

    if (nameMismatch.length) {
      console.error('Name mismatch:')
      for (const item of nameMismatch) {
        console.error(`  ${item.code}: manifest="${item.manifest}" backend="${item.backend}"`)
      }
    }

    process.exit(1)
  } catch (err) {
    console.error('check-locales failed:', err.message || err)
    process.exit(1)
  }
}

main()
