import axios from '@/utils/axios'

export interface SpokeRepairKitCatalogModel {
  brandSlug: string
  brandName: string
  slug: string
  model: string
  lifecycleStatus: string
}

const spokeRepairKitCatalogApi = {
  async listSpokeRepairKitWheelsetModels(): Promise<SpokeRepairKitCatalogModel[]> {
    const response = await axios.get('/api/v1/wheelset-spoke-specs/selector-models')
    const models = response.data?.data?.models
    if (!Array.isArray(models)) return []
    return models.filter((item: any) => (
      item && typeof item.brandSlug === 'string' && typeof item.slug === 'string' && typeof item.model === 'string'
    )) as SpokeRepairKitCatalogModel[]
  }
}

export default spokeRepairKitCatalogApi
