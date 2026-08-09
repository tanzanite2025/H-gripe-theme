import fs from 'node:fs/promises'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

interface HubGeometry {
  leftFlange: number
  rightFlange: number
  leftFlangePcd: number
  rightFlangePcd: number
  spokeHoleDiameter?: number
}

interface RimModel {
  id: string
  name: string
  erd: number
  weight?: number
}

interface HubModel {
  id: string
  name: string
  front?: HubGeometry
  rear?: HubGeometry
}

interface Brand<T> {
  id: string
  name: string
  items: T[]
}

interface SpokeExportPayload {
  rims: Brand<RimModel>[]
  hubs: Brand<HubModel>[]
}

const apiBase = (process.env.GO_API_URL || process.env.NUXT_PUBLIC_API_BASE || 'http://localhost:8080/api/v1').replace(/\/$/, '')
const apiUrl = process.env.SPOKE_API_URL || `${apiBase}/spoke/export`
const targetFile = path.join(path.dirname(fileURLToPath(import.meta.url)), '../app/data/spoke-calculator/database.ts')

function isRecord(value: unknown): value is Record<string, unknown> {
  return value !== null && typeof value === 'object' && !Array.isArray(value)
}

function isNumber(value: unknown): value is number {
  return typeof value === 'number' && Number.isFinite(value)
}

function isHubGeometry(value: unknown): value is HubGeometry {
  return (
    isRecord(value) &&
    isNumber(value.leftFlange) &&
    isNumber(value.rightFlange) &&
    isNumber(value.leftFlangePcd) &&
    isNumber(value.rightFlangePcd) &&
    (value.spokeHoleDiameter === undefined || isNumber(value.spokeHoleDiameter))
  )
}

function isRimModel(value: unknown): value is RimModel {
  return (
    isRecord(value) &&
    typeof value.id === 'string' &&
    typeof value.name === 'string' &&
    isNumber(value.erd) &&
    (value.weight === undefined || isNumber(value.weight))
  )
}

function isHubModel(value: unknown): value is HubModel {
  return (
    isRecord(value) &&
    typeof value.id === 'string' &&
    typeof value.name === 'string' &&
    (value.front === undefined || isHubGeometry(value.front)) &&
    (value.rear === undefined || isHubGeometry(value.rear))
  )
}

function isBrand<T>(value: unknown, isItem: (item: unknown) => item is T): value is Brand<T> {
  return (
    isRecord(value) &&
    typeof value.id === 'string' &&
    typeof value.name === 'string' &&
    Array.isArray(value.items) &&
    value.items.every(isItem)
  )
}

function isSpokeExportPayload(value: unknown): value is SpokeExportPayload {
  return (
    isRecord(value) &&
    Array.isArray(value.rims) &&
    value.rims.every((item) => isBrand(item, isRimModel)) &&
    Array.isArray(value.hubs) &&
    value.hubs.every((item) => isBrand(item, isHubModel))
  )
}

function buildGeneratedHeader(data: SpokeExportPayload): string {
  return `// AUTO-GENERATED FILE. DO NOT EDIT BRANDS MANUALLY.
// Run "npm run sync-data" to update from Go.

export interface HubGeometry {
  leftFlange: number
  rightFlange: number
  leftFlangePcd: number
  rightFlangePcd: number
  spokeHoleDiameter?: number
}

export interface HubModel {
  id: string
  name: string
  front?: HubGeometry
  rear?: HubGeometry
}

export interface RimModel {
  id: string
  name: string
  erd: number
  weight?: number
}

export interface Brand<T> {
  id: string
  name: string
  items: T[]
}

export const RIM_DATABASE: Brand<RimModel>[] = ${JSON.stringify(data.rims, null, 4)}

export const HUB_DATABASE: Brand<HubModel>[] = ${JSON.stringify(data.hubs, null, 4)}
`
}

function getPresetData(existingContent: string): string {
  const presetMarker = 'export interface WheelBuildPreset'
  const splitParts = existingContent.split(presetMarker)

  if (splitParts.length > 1) {
    return presetMarker + splitParts[1]
  }

  console.warn('PRESET_BUILDS section not found in existing file. Appending default empty structure.')
  return `
export interface WheelBuildPreset {
  id: string
  name: string
  keywords: string[]
  description?: string

  rimBrandId: string
  rimModelId: string
  hubBrandId: string
  hubModelId: string
  spokeCount: number
  crossing: number
  nippleType: 'standard' | 'hidden'
  nippleLength: number | null
}

export const PRESET_BUILDS: WheelBuildPreset[] = []
`
}

async function main(): Promise<void> {
  console.log(`Connecting to ${apiUrl}...`)

  try {
    const response = await fetch(apiUrl)
    if (!response.ok) {
      throw new Error(`API responded with ${response.status}: ${response.statusText}`)
    }

    const data = await response.json() as unknown
    if (!isSpokeExportPayload(data)) {
      throw new Error('Unexpected spoke export response shape')
    }

    console.log(`Data received: ${data.rims.length} rim brands, ${data.hubs.length} hub brands.`)

    let existingContent = ''
    try {
      existingContent = await fs.readFile(targetFile, 'utf-8')
    } catch {
      console.warn('Could not read existing file. PRESET_BUILDS might be lost if not handled.')
    }

    const finalContent = `${buildGeneratedHeader(data)}\n\n${getPresetData(existingContent)}`
    await fs.writeFile(targetFile, finalContent, 'utf-8')

    console.log(`Database synced successfully to ${targetFile}`)
  } catch (error: unknown) {
    console.error('Sync failed:', error instanceof Error ? error.message : error)
    process.exit(1)
  }
}

main()
