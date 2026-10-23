import axios from '@/utils/axios'

export interface PortableSettingsConfigurationEntry {
  key: string
  value: string
  type: string
  group: string
  locale: string
  is_public: boolean
  description: string
}

export interface PortableSettingsConfigurationFile {
  format: string
  version: number
  exported_at: string
  settings: PortableSettingsConfigurationEntry[]
  excluded_setting_count: number
}

export interface SettingsConfigurationBackupOverview {
  configuration_export_available: boolean
  configuration_count: number
  excluded_setting_count: number
  scopes: { group: string; label: string; count: number }[]
  max_configuration_file_bytes: number
  whole_storefront_maintenance_configured: boolean
}

export interface SettingsConfigurationImportPreview {
  backup_sha256: string
  target_sha256: string
  create_count: number
  update_count: number
  unchanged_count: number
  changes: {
    key: string; group: string; locale: string; action: string
    current_value: string; imported_value: string
    current_type: string; imported_type: string
    current_is_public: boolean; imported_is_public: boolean
    current_description: string; imported_description: string
  }[]
}

const configurationBackupEndpoint = '/api/admin/settings/backup-recovery'

export const settingsConfigurationBackupAndRecoveryApi = {
  async getOverview(): Promise<SettingsConfigurationBackupOverview> {
    return (await axios.get<SettingsConfigurationBackupOverview>(`${configurationBackupEndpoint}/overview`)).data
  },
  async exportConfiguration(): Promise<PortableSettingsConfigurationFile> {
    return (await axios.get<PortableSettingsConfigurationFile>(`${configurationBackupEndpoint}/configuration/export`)).data
  },
  async previewConfiguration(backup: PortableSettingsConfigurationFile): Promise<SettingsConfigurationImportPreview> {
    return (await axios.post<SettingsConfigurationImportPreview>(`${configurationBackupEndpoint}/configuration/preview`, { backup })).data
  },
  async importConfiguration(backup: PortableSettingsConfigurationFile, expectedTargetSHA256: string): Promise<{ count: number }> {
    return (await axios.post<{ count: number }>(`${configurationBackupEndpoint}/configuration/import`, {
      backup, expected_target_sha256: expectedTargetSHA256, confirmation: 'IMPORT CONFIGURATION',
    })).data
  },
}
