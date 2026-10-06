import axios from '@/utils/axios'
import {
  requireApiArrayField,
  requireApiObject,
  unwrapApiPayload,
} from '@/utils/apiResponse'

export interface FpxPublishedCollectionReference {
  id: number
  service_code: string
  display_name: string
  countries: string
  enabled: boolean
}

export interface YanwenPublishedCollectionReference {
  id: number
  product_code: string
  display_name: string
  countries: string
  enabled: boolean
}

const readShippingServiceCollectionPayload = (response: unknown, endpoint: string) => (
  unwrapApiPayload(response, endpoint)
)

const readShippingServiceCollectionArray = (response: unknown, endpoint: string) => (
  requireApiArrayField(
    requireApiObject(readShippingServiceCollectionPayload(response, endpoint), endpoint),
    'data',
    endpoint,
  )
)

export const shippingServiceCollectionReferenceApi = {
  async listFpxPublishedCollection(): Promise<FpxPublishedCollectionReference[]> {
    const endpoint = '/api/admin/logistics/4px/collection'
    return readShippingServiceCollectionArray(await axios.get(endpoint), endpoint) as FpxPublishedCollectionReference[]
  },

  async listYanwenPublishedCollection(): Promise<YanwenPublishedCollectionReference[]> {
    const endpoint = '/api/admin/logistics/yanwen/collection'
    return readShippingServiceCollectionArray(await axios.get(endpoint), endpoint) as YanwenPublishedCollectionReference[]
  },
}

export default shippingServiceCollectionReferenceApi
