import axios from '@/utils/axios'

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

const unwrapPayload = (response: any) => response.data?.data ?? response.data ?? {}

export const spokeCatalogApi = {
  async get(): Promise<SpokeCatalog> {
    return unwrapPayload(await axios.get('/api/admin/spoke-catalog'))
  },

  async save(payload: SpokeCatalog): Promise<SpokeCatalog> {
    return unwrapPayload(await axios.put('/api/admin/spoke-catalog', payload))
  },

  async importFile(file: File): Promise<SpokeCatalog> {
    const formData = new FormData()
    formData.append('file', file)
    return unwrapPayload(await axios.post('/api/admin/spoke-catalog/import', formData))
  },

  async downloadPresetTemplate(): Promise<Blob> {
    const response = await axios.get('/api/admin/spoke-catalog/preset-template', {
      responseType: 'blob',
    })
    return response.data
  },

  async importPresetTemplate(file: File): Promise<SpokeCatalog> {
    const formData = new FormData()
    formData.append('file', file)
    return unwrapPayload(await axios.post('/api/admin/spoke-catalog/preset-template/import', formData))
  }
}

export default spokeCatalogApi
