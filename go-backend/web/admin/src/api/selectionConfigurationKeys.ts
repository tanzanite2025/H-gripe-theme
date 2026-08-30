import axios from '@/utils/axios'
import {
  requireApiArray,
  requireApiObject,
  requireApiNumberField,
  requireApiStringField,
  unwrapApiPayload,
} from '@/utils/apiResponse'
import type { SelectionConfigurationKeyKind } from '@/modules/selection-configuration/selectionConfigurationKeys'

export type { SelectionConfigurationKeyKind } from '@/modules/selection-configuration/selectionConfigurationKeys'

export interface SelectionConfigurationKeyRecord {
  id: number
  kind: SelectionConfigurationKeyKind
  code: string
  display_label: string
  description: string
  is_enabled: boolean
  sort_order: number
  created_at?: string
  updated_at?: string
}

export interface SelectionConfigurationKeyOption {
  id: number
  code: string
  display_label: string
}

export interface SelectionConfigurationKeyPayload {
  kind: SelectionConfigurationKeyKind
  code: string
  display_label: string
  description: string
  is_enabled: boolean
  sort_order: number
}

const baseEndpoint = '/api/admin/selection-configuration/keys'

const readKeyList = (response: unknown, endpoint: string): SelectionConfigurationKeyRecord[] => {
  return requireApiArray(unwrapApiPayload(response, endpoint), endpoint, 'data') as SelectionConfigurationKeyRecord[]
}

const readKeyOptions = (response: unknown, endpoint: string): SelectionConfigurationKeyOption[] => {
  return requireApiArray(unwrapApiPayload(response, endpoint), endpoint, 'data') as SelectionConfigurationKeyOption[]
}

const readKeyRecord = (response: unknown, endpoint: string): SelectionConfigurationKeyRecord => {
  const key = requireApiObject(unwrapApiPayload(response, endpoint), endpoint, 'data')
  requireApiNumberField(key, 'id', endpoint)
  requireApiStringField(key, 'kind', endpoint)
  requireApiStringField(key, 'code', endpoint)
  requireApiStringField(key, 'display_label', endpoint)
  requireApiStringField(key, 'description', endpoint)
  return key as unknown as SelectionConfigurationKeyRecord
}

export const selectionConfigurationKeyApi = {
  async listKeys(kind: SelectionConfigurationKeyKind, includeDisabled = true): Promise<SelectionConfigurationKeyRecord[]> {
    const endpoint = `${baseEndpoint}?kind=${encodeURIComponent(kind)}&include_disabled=${includeDisabled ? 'true' : 'false'}`
    return readKeyList(await axios.get(endpoint), endpoint)
  },

  async listEnabledKeyOptions(kind: SelectionConfigurationKeyKind): Promise<SelectionConfigurationKeyOption[]> {
    const endpoint = `${baseEndpoint}/options?kind=${encodeURIComponent(kind)}`
    return readKeyOptions(await axios.get(endpoint), endpoint)
  },

  async createKey(payload: SelectionConfigurationKeyPayload): Promise<SelectionConfigurationKeyRecord> {
    const endpoint = baseEndpoint
    return readKeyRecord(await axios.post(endpoint, payload), endpoint)
  },

  async updateKey(id: number, payload: SelectionConfigurationKeyPayload): Promise<SelectionConfigurationKeyRecord> {
    const endpoint = `${baseEndpoint}/${id}`
    return readKeyRecord(await axios.put(endpoint, payload), endpoint)
  },
}

export default selectionConfigurationKeyApi
