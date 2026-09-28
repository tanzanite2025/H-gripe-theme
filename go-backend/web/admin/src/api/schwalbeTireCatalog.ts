import axios from '@/utils/axios'
import { requireApiArray, requireApiBooleanField, requireApiObject, requireApiStringField, unwrapApiPayload } from '@/utils/apiResponse'

export interface SchwalbeTireCatalogItem {
  article_no: string
  ean?: string
  model_name: string
  etrto: string
  inch_designation?: string
  weight_g?: number
  version_label?: string
  compound?: string
  color?: string
  bead?: string
  e_bike_rating?: string
  epi?: number
  load_kg?: number
  seal?: string
  tread?: string
  min_pressure_bar?: number
  max_pressure_bar?: number
  min_pressure_psi?: number
  max_pressure_psi?: number
  source_url: string
  source_checked_at: string
  product_exists: boolean
}

const readCatalogItem = (value: unknown, endpoint: string): SchwalbeTireCatalogItem => {
  const item = requireApiObject(value, endpoint, 'catalog item')
  requireApiStringField(item, 'article_no', endpoint)
  requireApiStringField(item, 'model_name', endpoint)
  requireApiStringField(item, 'etrto', endpoint)
  requireApiStringField(item, 'source_url', endpoint)
  requireApiStringField(item, 'source_checked_at', endpoint)
  requireApiBooleanField(item, 'product_exists', endpoint)
  return item as SchwalbeTireCatalogItem
}

const schwalbeTireCatalogApi = {
  async list(search = ''): Promise<SchwalbeTireCatalogItem[]> {
    const endpoint = '/api/v1/products/schwalbe-tire-catalog'
    const payload = requireApiArray(
      unwrapApiPayload(await axios.get(endpoint, { params: search.trim() ? { search } : {} }), endpoint),
      endpoint,
    )
    return payload.map((item) => readCatalogItem(item, endpoint))
  },
}

export default schwalbeTireCatalogApi
