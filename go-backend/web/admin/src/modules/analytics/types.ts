export interface AnalyticsSettings {
  google_analytics: string
  google_tag_manager: string
}

export interface AnalyticsUpdateRequest extends Partial<AnalyticsSettings> {
  locale?: string
}
