import axios from '@/utils/axios'
import {
  requireApiAcknowledgement,
  requireApiArrayField,
  requireApiBooleanField,
  requireApiNumberField,
  requireApiObject,
  requireApiStringField,
  readApiBody,
} from '@/utils/apiResponse'
import {
  socialProviderList,
  type SocialFieldKey,
  type SocialProvider,
} from '@/utils/socialPlatforms'

export interface SocialSettingRecord {
  key: string
  value: string
}

export interface SocialPublicLinkUpdate {
  key: SocialFieldKey
  value: string
  type: 'string'
  group: 'social'
  locale: string
  is_public: true
  description: string
}

export interface SocialProviderPage {
  id: string
  name?: string
  url?: string
}

export interface SocialInstagramAccount {
  id: string
  username?: string
  name?: string
  url?: string
}

export interface SocialProviderResources {
  pages?: SocialProviderPage[]
  instagram_accounts?: SocialInstagramAccount[]
}

export interface SocialConnection {
  provider: SocialProvider
  label: string
  configured: boolean
  connected: boolean
  status: string
  provider_account_id?: string
  provider_account_name?: string
  provider_account_url?: string
  provider_account_email?: string
  granted_scopes?: string[]
  provider_resources?: SocialProviderResources
  last_error?: string
}

const readOptionalString = (record: Record<string, unknown>, key: string, endpoint: string): string | undefined => {
  if (record[key] === undefined || record[key] === null) return undefined
  return requireApiStringField(record, key, endpoint)
}

const readProviderResources = (value: unknown): SocialProviderResources | undefined => {
  if (!value || typeof value !== 'object' || Array.isArray(value)) return undefined
  const resources = value as Record<string, unknown>
  const pages = Array.isArray(resources.pages)
    ? resources.pages
      .filter((item): item is Record<string, unknown> => Boolean(item) && typeof item === 'object' && !Array.isArray(item))
      .map((item) => ({
        id: String(item.id ?? ''),
        ...(item.name ? { name: String(item.name) } : {}),
        ...(item.url ? { url: String(item.url) } : {}),
      }))
      .filter((item) => item.id)
    : undefined
  const instagramAccounts = Array.isArray(resources.instagram_accounts)
    ? resources.instagram_accounts
      .filter((item): item is Record<string, unknown> => Boolean(item) && typeof item === 'object' && !Array.isArray(item))
      .map((item) => ({
        id: String(item.id ?? ''),
        ...(item.username ? { username: String(item.username) } : {}),
        ...(item.name ? { name: String(item.name) } : {}),
        ...(item.url ? { url: String(item.url) } : {}),
      }))
      .filter((item) => item.id)
    : undefined

  return {
    ...(pages ? { pages } : {}),
    ...(instagramAccounts ? { instagram_accounts: instagramAccounts } : {}),
  }
}

const readConnection = (value: unknown, endpoint: string): SocialConnection => {
  const record = requireApiObject(value, endpoint, 'connection')
  const provider = requireApiStringField(record, 'provider', endpoint)
  if (!socialProviderList.includes(provider as SocialProvider)) {
    throw new Error(`[CRITICAL] Invalid API response for ${endpoint}: unsupported provider "${provider}"`)
  }

  const connection: SocialConnection = {
    provider: provider as SocialProvider,
    label: requireApiStringField(record, 'label', endpoint),
    configured: requireApiBooleanField(record, 'configured', endpoint),
    connected: requireApiBooleanField(record, 'connected', endpoint),
    status: requireApiStringField(record, 'status', endpoint),
  }

  const optionalFields: Array<keyof Pick<
    SocialConnection,
    'provider_account_id' | 'provider_account_name' | 'provider_account_url' | 'provider_account_email' | 'last_error'
  >> = [
    'provider_account_id',
    'provider_account_name',
    'provider_account_url',
    'provider_account_email',
    'last_error',
  ]
  optionalFields.forEach((field) => {
    const value = readOptionalString(record, field, endpoint)
    if (value !== undefined) connection[field] = value
  })

  if (record.granted_scopes !== undefined && record.granted_scopes !== null) {
    if (!Array.isArray(record.granted_scopes)) {
      throw new Error(`[CRITICAL] Invalid API response for ${endpoint}: granted_scopes must be an array`)
    }
    connection.granted_scopes = record.granted_scopes.map((scope) => String(scope))
  }
  if (record.provider_resources !== undefined && record.provider_resources !== null) {
    connection.provider_resources = readProviderResources(record.provider_resources)
  }

  return connection
}

const readSetting = (value: unknown, endpoint: string): SocialSettingRecord => {
  const record = requireApiObject(value, endpoint, 'setting')
  return {
    key: requireApiStringField(record, 'key', endpoint),
    value: requireApiStringField(record, 'value', endpoint),
  }
}

export const socialApi = {
  async listPublicLinks(locale = 'en'): Promise<SocialSettingRecord[]> {
    const endpoint = '/api/admin/settings/social'
    const body = requireApiObject(
      readApiBody(await axios.get(endpoint, { params: { locale } }), endpoint),
      endpoint,
      'response body',
    )
    return requireApiArrayField(body, 'settings', endpoint)
      .map((setting) => readSetting(setting, endpoint))
  },

  async savePublicLinks(settings: SocialPublicLinkUpdate[]): Promise<number> {
    const endpoint = '/api/admin/settings/batch'
    const body = requireApiObject(
      readApiBody(await axios.post(endpoint, { settings }), endpoint),
      endpoint,
      'response body',
    )
    return requireApiNumberField(body, 'count', endpoint)
  },

  async listOAuthConnections(): Promise<SocialConnection[]> {
    const endpoint = '/api/admin/social/oauth'
    const body = requireApiObject(
      readApiBody(await axios.get(endpoint), endpoint),
      endpoint,
      'response body',
    )
    return requireApiArrayField(body, 'connections', endpoint)
      .map((connection) => readConnection(connection, endpoint))
  },

  async startOAuth(provider: SocialProvider, returnPath: string): Promise<string> {
    const endpoint = `/api/admin/social/oauth/${provider}/start`
    const body = requireApiObject(
      readApiBody(await axios.post(endpoint, { return_path: returnPath }), endpoint),
      endpoint,
      'response body',
    )
    return requireApiStringField(body, 'authorization_url', endpoint)
  },

  async disconnect(provider: SocialProvider): Promise<void> {
    const endpoint = `/api/admin/social/oauth/${provider}`
    requireApiAcknowledgement(await axios.delete(endpoint), endpoint)
  },
}

export default socialApi
