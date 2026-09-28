import fs from 'node:fs'
import path from 'node:path'
import process from 'node:process'
import { gunzipSync } from 'node:zlib'

const DEFAULT_SITEMAP_URL = 'https://www.schwalbe.com/en/sitemap.xml'
const DEFAULT_SOURCE_CHECKED_AT = process.env.SCHWALBE_SOURCE_CHECKED_AT || new Date().toISOString().slice(0, 10)
const tireArticlePattern = /\/en\/[^/]+-(?:(?:111|112|116)\d{3,})(?:\.\d+)?\/?$/
const articlePattern = /-((?:111|112|116)\d{3,})(?:\.(\d+))?\/?$/
const concurrency = 8

const args = process.argv.slice(2)
const readArg = (name, fallback = '') => {
  const index = args.indexOf(name)
  return index >= 0 && args[index + 1] ? args[index + 1] : fallback
}
const outputJson = readArg('--output-json')
const outputSql = readArg('--output-sql')
const sitemapUrl = readArg('--sitemap-url', DEFAULT_SITEMAP_URL)
const sourceCheckedAt = readArg('--source-checked-at', DEFAULT_SOURCE_CHECKED_AT)

const entityNames = {
  amp: '&', apos: "'", gt: '>', lt: '<', nbsp: ' ', quot: '"', ndash: '–', mdash: '—', hellip: '…',
}

function decodeHtml(value) {
  let decoded = String(value || '')
  // Some product values are double escaped (for example, &amp;#039;). Two
  // passes preserve the official punctuation without leaving HTML entities
  // in the catalog text.
  for (let pass = 0; pass < 2; pass += 1) {
    const next = decoded.replace(/&(#x[0-9a-f]+|#\d+|[a-z][a-z0-9]+);/gi, (whole, entity) => {
      if (entity[0] === '#') {
        const hex = entity[1].toLowerCase() === 'x'
        const codePoint = Number.parseInt(entity.slice(hex ? 2 : 1), hex ? 16 : 10)
        return Number.isFinite(codePoint) ? String.fromCodePoint(codePoint) : whole
      }
      return entityNames[entity.toLowerCase()] || whole
    })
    if (next === decoded) break
    decoded = next
  }
  return decoded
}

function cleanText(value) {
  return decodeHtml(String(value || ''))
    .replace(/<[^>]*>/g, ' ')
    .replace(/\s+/g, ' ')
    .trim()
}

function parsePositiveNumber(value, { integer = false } = {}) {
  const raw = cleanText(value)
  if (!raw) return null

  // EPI is a scalar field, while Schwalbe sometimes displays a dual-ply
  // value such as "2x67". Keep the per-ply EPI below instead of parsing the
  // layer count (2) as the measurement.
  const composite = raw.match(/\d+(?:[.,]\d+)?\s*[x×]\s*(\d+(?:[.,]\d+)?)/i)
  const token = composite ? composite[1] : raw.match(/\d[\d.,]*/)?.[0]
  if (!token) return null

  let normalized = token
  if (normalized.includes(',') && normalized.includes('.')) {
    // The final separator is the decimal separator when both forms occur.
    const decimalSeparator = Math.max(normalized.lastIndexOf(','), normalized.lastIndexOf('.'))
    const integerPart = normalized.slice(0, decimalSeparator).replace(/[.,]/g, '')
    const fractionPart = normalized.slice(decimalSeparator + 1).replace(/[.,]/g, '')
    normalized = `${integerPart}.${fractionPart}`
  } else if (normalized.includes(',')) {
    const parts = normalized.split(',')
    // English Schwalbe pages use commas for thousands (for example,
    // "1,005 g"). A one- or two-digit suffix is a decimal comma instead.
    normalized = parts.length > 1 && parts.slice(1).every((part) => part.length === 3)
      ? parts.join('')
      : normalized.replace(',', '.')
  }

  const number = Number(normalized)
  return Number.isFinite(number) && number > 0 && (!integer || Number.isInteger(number)) ? number : null
}

function parseJsonLd(html) {
  const objects = []
  for (const match of html.matchAll(/<script[^>]+type=["']application\/ld\+json["'][^>]*>([\s\S]*?)<\/script>/gi)) {
    try {
      const parsed = JSON.parse(match[1].trim())
      const queue = Array.isArray(parsed) ? [...parsed] : [parsed]
      while (queue.length) {
        const item = queue.shift()
        if (!item || typeof item !== 'object') continue
        if (Array.isArray(item)) queue.push(...item)
        else {
          objects.push(item)
          for (const value of Object.values(item)) if (value && typeof value === 'object') queue.push(value)
        }
      }
    } catch {
      // Ignore unrelated JSON-LD blocks that are not valid product data.
    }
  }
  return objects
}

function parseProperties(html) {
  const properties = {}
  const pattern = /<span[^>]*class=["'][^"']*properties-label[^"']*["'][^>]*>([\s\S]*?)<\/span>\s*<div[^>]*class=["'][^"']*properties-value[^"']*["'][^>]*>([\s\S]*?)<\/div>/gi
  for (const match of html.matchAll(pattern)) {
    const label = cleanText(match[1]).replace(/:\s*$/, '').toLowerCase()
    const value = cleanText(match[2])
    if (label && !(label in properties)) properties[label] = value
  }
  return properties
}

function extractRow(html, item) {
  const properties = parseProperties(html)
  const jsonLd = parseJsonLd(html)
  const product = jsonLd.find((value) => String(value.sku || value.mpn || '') === item.articleNo)
    || jsonLd.find((value) => value['@type'] === 'ProductGroup' || value['@type'] === 'Product')
    || {}
  const ogTitle = html.match(/<meta[^>]+property=["']og:title["'][^>]+content=["']([^"']*)["']/i)?.[1]
  const modelName = cleanText(product.name || ogTitle || '')
  const ean = properties.ean || String(product.gtin13 || product.gtin || '')
  return {
    article_no: item.articleNo,
    ean: ean || null,
    model_name: modelName,
    etrto: properties.etrto || null,
    inch_designation: properties.inch || null,
    weight_g: parsePositiveNumber(properties.weight),
    version_label: properties.version || null,
    compound: properties.compound || null,
    color: properties.color || null,
    bead: properties.bead || null,
    e_bike_rating: properties['e-bike'] || null,
    epi: parsePositiveNumber(properties.epi, { integer: true }),
    load_kg: parsePositiveNumber(properties['load (kg)']),
    seal: properties.seal || null,
    tread: properties.tread || null,
    min_pressure_bar: parsePositiveNumber(properties['min. bar']),
    max_pressure_bar: parsePositiveNumber(properties['max. bar']),
    min_pressure_psi: parsePositiveNumber(properties['min. psi']),
    max_pressure_psi: parsePositiveNumber(properties['max. psi']),
    source_url: item.url,
    source_checked_at: sourceCheckedAt,
  }
}

async function fetchText(url) {
  const response = await fetch(url, {
    headers: {
      'user-agent': 'Tanzanite-Schwalbe-Catalog-Importer/1.0 (+official-source-import)',
      accept: 'text/html,application/xhtml+xml',
    },
    signal: AbortSignal.timeout(60000),
  })
  if (response.status === 404) return { status: 404, text: '' }
  if (!response.ok) throw new Error(`HTTP ${response.status}`)
  return { status: response.status, text: await response.text() }
}

async function fetchSitemapXml(url) {
  const response = await fetch(url, {
    headers: { 'user-agent': 'Tanzanite-Schwalbe-Catalog-Importer/1.0 (+official-source-import)' },
    signal: AbortSignal.timeout(60000),
  })
  if (!response.ok) throw new Error(`Sitemap ${url} returned HTTP ${response.status}`)
  const bytes = Buffer.from(await response.arrayBuffer())
  return url.endsWith('.gz') ? gunzipSync(bytes).toString('utf8') : bytes.toString('utf8')
}

async function fetchProduct(item) {
  let lastError
  for (let attempt = 1; attempt <= 3; attempt += 1) {
    try {
      const response = await fetchText(item.url)
      if (response.status === 404) return { item, missing: true }
      return { item, row: extractRow(response.text, item) }
    } catch (error) {
      lastError = error
      await new Promise((resolve) => setTimeout(resolve, attempt * 500))
    }
  }
  return { item, error: String(lastError?.message || lastError) }
}

function sqlString(value) {
  if (value === null || value === undefined || value === '') return 'NULL'
  return `'${String(value).replaceAll("'", "''")}'`
}

function sqlNumber(value) {
  return value === null || value === undefined ? 'NULL' : String(value)
}

function renderSql(rows, errors) {
  const columns = [
    'article_no', 'ean', 'model_name', 'etrto', 'inch_designation', 'weight_g', 'version_label',
    'compound', 'color', 'bead', 'e_bike_rating', 'epi', 'load_kg', 'seal', 'tread',
    'min_pressure_bar', 'max_pressure_bar', 'min_pressure_psi', 'max_pressure_psi',
    'source_url', 'source_checked_at',
  ]
  const numeric = new Set(['weight_g', 'epi', 'load_kg', 'min_pressure_bar', 'max_pressure_bar', 'min_pressure_psi', 'max_pressure_psi'])
  const valueSql = (row) => columns.map((column) => numeric.has(column) ? sqlNumber(row[column]) : sqlString(row[column])).join(', ')
  const chunks = []
  for (let index = 0; index < rows.length; index += 100) {
    chunks.push(`INSERT INTO schwalbe_tire_specifications (\n    ${columns.join(',\n    ')}\n) VALUES\n${rows.slice(index, index + 100).map((row) => `    (${valueSql(row)})`).join(',\n')}\nON CONFLICT (article_no) DO UPDATE SET\n    ${columns.filter((column) => column !== 'article_no').map((column) => `${column} = EXCLUDED.${column}`).join(',\n    ')},\n    updated_at = NOW();`)
  }
  const missing = errors.filter((error) => error.missing).map((error) => error.item.articleNo).sort((a, b) => a.localeCompare(b))
  return `-- Official Schwalbe English product sitemap snapshot: ${sitemapUrl}\n-- Checked at: ${sourceCheckedAt}\n-- Imported live tire product pages: ${rows.length}\n-- Sitemap entries returning HTTP 404 and intentionally excluded: ${missing.length}${missing.length ? ` (${missing.join(', ')})` : ''}\n-- This migration upserts sourced catalog facts; it does not create sales Products or SKUs.\n\n${chunks.join('\n\n')}\n`
}

async function main() {
  const sitemapIndex = await fetchSitemapXml(sitemapUrl)
  const sitemapLocs = [...sitemapIndex.matchAll(/<loc>([^<]+)<\/loc>/g)].map((match) => match[1])
  const sitemap = /<sitemapindex\b/i.test(sitemapIndex)
    ? (await Promise.all(sitemapLocs.map((url) => fetchSitemapXml(url)))).join('\n')
    : sitemapIndex
  const candidates = [...sitemap.matchAll(/<loc>([^<]+)<\/loc>/g)]
    .map((match) => match[1])
    .filter((url) => tireArticlePattern.test(url))
    .map((url) => {
      const match = url.match(articlePattern)
      return { url, articleNo: `${match[1]}${match[2] ? `.${match[2]}` : ''}` }
    })
  const unique = [...new Map(candidates.map((item) => [item.articleNo, item])).values()]
  const results = []
  const errors = []
  let nextIndex = 0
  async function worker() {
    while (true) {
      const index = nextIndex++
      if (index >= unique.length) return
      const result = await fetchProduct(unique[index])
      if (result.row) results.push(result.row)
      else errors.push(result)
      if ((index + 1) % 25 === 0) console.log(`Fetched ${index + 1}/${unique.length}`)
    }
  }
  await Promise.all(Array.from({ length: concurrency }, () => worker()))
  results.sort((a, b) => a.article_no.localeCompare(b.article_no))
  const invalid = results.filter((row) => {
    const numericFields = ['weight_g', 'epi', 'load_kg', 'min_pressure_bar', 'max_pressure_bar', 'min_pressure_psi', 'max_pressure_psi']
    return !row.article_no || !row.model_name || !row.etrto
      || numericFields.some((field) => row[field] !== null && (!Number.isFinite(row[field]) || row[field] <= 0))
      || (row.epi !== null && !Number.isInteger(row.epi))
      || (row.min_pressure_bar !== null && row.max_pressure_bar !== null && row.min_pressure_bar > row.max_pressure_bar)
      || (row.min_pressure_psi !== null && row.max_pressure_psi !== null && row.min_pressure_psi > row.max_pressure_psi)
  })
  if (invalid.length) throw new Error(`Invalid scraped rows: ${invalid.length}`)
  if (errors.some((error) => !error.missing)) throw new Error(`Fetch failures: ${errors.filter((error) => !error.missing).length}`)
  if (outputJson) {
    fs.mkdirSync(path.dirname(path.resolve(outputJson)), { recursive: true })
    fs.writeFileSync(outputJson, JSON.stringify(results, null, 2) + '\n')
  }
  if (outputSql) {
    fs.mkdirSync(path.dirname(path.resolve(outputSql)), { recursive: true })
    fs.writeFileSync(outputSql, renderSql(results, errors))
  }
  console.log(JSON.stringify({ sitemapUrl, sourceCheckedAt, candidates: unique.length, rows: results.length, missing404: errors.filter((error) => error.missing).length, outputJson, outputSql }, null, 2))
}

await main()
