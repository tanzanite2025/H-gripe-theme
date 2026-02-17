import fs from 'node:fs'
import path from 'node:path'
import process from 'node:process'

const wpApiBase = process.env.WP_API_BASE || 'https://tanzanite.site/wp-json'
const wpLocalesUrl = process.env.WP_LOCALES_URL
const manifestPath = path.resolve('i18n/locales.manifest.js')

function loadManifestLocales() {
  if (!fs.existsSync(manifestPath)) {
    throw new Error(`Manifest not found: ${manifestPath}`)
  }

  const moduleUrl = `file://${manifestPath}`
  return import(moduleUrl).then((mod) => mod.default || mod.locales || [])
}

function resolveEndpoint() {
  if (wpLocalesUrl) return wpLocalesUrl
  const base = wpApiBase.replace(/\/$/, '')
  if (base.endsWith('/wp-json')) return `${base}/tanzanite/v1/languages`
  return `${base}/wp-json/tanzanite/v1/languages`
}

async function fetchWpLocales() {
  const url = resolveEndpoint()
  const res = await fetch(url)
  if (!res.ok) {
    if (res.status === 404) {
      throw new Error(`Fetch WP locales failed: 404 Not Found. Check that tanzanite-blog-i18n plugin is active and endpoint exists at ${url}. Override with WP_API_BASE or WP_LOCALES_URL if needed.`)
    }
    throw new Error(`Fetch WP locales failed: ${res.status} ${res.statusText} (url: ${url})`)
  }
  const data = await res.json()
  if (!data || !Array.isArray(data.locales)) {
    throw new Error('Unexpected WP locales response shape')
  }
  return data.locales
}

function toMap(list) {
  const map = new Map()
  for (const item of list) {
    if (!item || !item.code) continue
    map.set(String(item.code), item)
  }
  return map
}

function diffLocales(manifestList, wpList) {
  const manifestMap = toMap(manifestList)
  const wpMap = toMap(wpList)

  const missingInManifest = []
  const extraInManifest = []
  const nameMismatch = []

  for (const code of wpMap.keys()) {
    if (!manifestMap.has(code)) {
      missingInManifest.push(code)
    } else {
      const m = manifestMap.get(code)
      const w = wpMap.get(code)
      if (m?.name && w?.name && m.name !== w.name) {
        nameMismatch.push({ code, manifest: m.name, wp: w.name })
      }
    }
  }

  for (const code of manifestMap.keys()) {
    if (!wpMap.has(code)) {
      extraInManifest.push(code)
    }
  }

  return { missingInManifest, extraInManifest, nameMismatch }
}

async function main() {
  try {
    const [manifestLocales, wpLocales] = await Promise.all([
      loadManifestLocales(),
      fetchWpLocales(),
    ])

    const { missingInManifest, extraInManifest, nameMismatch } = diffLocales(manifestLocales, wpLocales)

    if (
      missingInManifest.length === 0 &&
      extraInManifest.length === 0 &&
      nameMismatch.length === 0
    ) {
      console.log('Locales are aligned with WordPress source.')
      process.exit(0)
    }

    if (missingInManifest.length) {
      console.error('Missing in manifest (present in WP):', missingInManifest.join(', '))
    }

    if (extraInManifest.length) {
      console.error('Extra in manifest (absent in WP):', extraInManifest.join(', '))
    }

    if (nameMismatch.length) {
      console.error('Name mismatch:')
      for (const item of nameMismatch) {
        console.error(`  ${item.code}: manifest="${item.manifest}" wp="${item.wp}"`)
      }
    }

    process.exit(1)
  } catch (err) {
    console.error('check-locales failed:', err.message || err)
    process.exit(1)
  }
}

main()
