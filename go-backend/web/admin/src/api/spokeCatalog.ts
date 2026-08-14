import axios from '@/utils/axios'
import {
  readApiBody,
  requireApiArrayField,
  requireApiObject,
  requireApiObjectField,
} from '@/utils/apiResponse'

export interface SpokeIntOption {
  value: number
  label: string
}

export interface SpokeStringOption {
  value: string
  label: string
}

export interface SpokeCatalogOptions {
  spokeCounts: SpokeIntOption[]
  crossings: SpokeIntOption[]
  nippleTypes: SpokeStringOption[]
  wheelPositions: SpokeStringOption[]
}

export interface SpokeRimModel {
  id: string
  name: string
  erd: number | null
  weight?: number | null
}

export interface SpokeHubGeometry {
  leftFlange: number | null
  rightFlange: number | null
  leftFlangePcd: number | null
  rightFlangePcd: number | null
  spokeHoleDiameter?: number | null
}

export interface SpokeHubModel {
  id: string
  name: string
  front?: SpokeHubGeometry | null
  rear?: SpokeHubGeometry | null
}

export interface SpokeBuildActualLengths {
  frontLeft: number | null
  frontRight: number | null
  rearLeft: number | null
  rearRight: number | null
  notes?: string
}

export interface SpokeBrand<T> {
  id: string
  name: string
  items: T[]
}

export interface SpokeBuildPreset {
  id: string
  name: string
  description?: string
  keywords: string[]
  rimBrandId: string
  rimModelId: string
  hubBrandId: string
  hubModelId: string
  wheelPosition?: string
  spokeCount: number
  crossing: number
  nippleType: string
  nippleLength: number | null
  actualLengths?: SpokeBuildActualLengths | null
}

export interface SpokeCatalog {
  options: SpokeCatalogOptions
  rims: SpokeBrand<SpokeRimModel>[]
  hubs: SpokeBrand<SpokeHubModel>[]
  presets: SpokeBuildPreset[]
}

const requireApiBlob = (value: unknown, endpoint: string): Blob => {
  if (!(value instanceof Blob) || value.size === 0) {
    throw new Error(`[CRITICAL] Invalid API response for ${endpoint}: response body must be a non-empty Blob`)
  }
  return value
}

const readCatalog = (response: unknown, endpoint: string): SpokeCatalog => {
  const catalog = requireApiObject(readApiBody(response, endpoint), endpoint, 'response body')
  const options = requireApiObjectField<SpokeCatalogOptions>(catalog, 'options', endpoint)

  requireApiArrayField(options, 'spokeCounts', endpoint)
  requireApiArrayField(options, 'crossings', endpoint)
  requireApiArrayField(options, 'nippleTypes', endpoint)
  requireApiArrayField(options, 'wheelPositions', endpoint)
  requireApiArrayField<SpokeBrand<SpokeRimModel>>(catalog, 'rims', endpoint)
  requireApiArrayField<SpokeBrand<SpokeHubModel>>(catalog, 'hubs', endpoint)
  requireApiArrayField<SpokeBuildPreset>(catalog, 'presets', endpoint)

  return catalog as SpokeCatalog
}

export const spokeCatalogApi = {
  async get(): Promise<SpokeCatalog> {
    const endpoint = '/api/admin/spoke-catalog'
    return readCatalog(await axios.get(endpoint), endpoint)
  },

  async save(payload: SpokeCatalog): Promise<SpokeCatalog> {
    const endpoint = '/api/admin/spoke-catalog'
    return readCatalog(await axios.put(endpoint, payload), endpoint)
  },

  async importFile(file: File): Promise<SpokeCatalog> {
    const formData = new FormData()
    formData.append('file', file)
    const endpoint = '/api/admin/spoke-catalog/import'
    return readCatalog(await axios.post(endpoint, formData), endpoint)
  },

  async downloadPresetTemplate(): Promise<Blob> {
    const endpoint = '/api/admin/spoke-catalog/preset-template'
    const response = await axios.get(endpoint, {
      responseType: 'blob',
    })
    return requireApiBlob(readApiBody(response, endpoint), endpoint)
  },

  async importPresetTemplate(file: File): Promise<SpokeCatalog> {
    const formData = new FormData()
    formData.append('file', file)
    const endpoint = '/api/admin/spoke-catalog/preset-template/import'
    return readCatalog(await axios.post(endpoint, formData), endpoint)
  }
}

export default spokeCatalogApi
