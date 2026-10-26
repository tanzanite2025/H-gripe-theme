import type { Component } from 'vue'
import type { RouteLocationRaw } from 'vue-router'

export interface AdminGlobalSearchResult {
  id: string
  sourceId: string
  title: string
  description: string
  categoryLabel: string
  linkLabel: string
  routeTarget: RouteLocationRaw
  icon?: Component
}

export interface AdminGlobalSearchRequestContext {
  signal: AbortSignal
}

export interface AdminGlobalSearchSource {
  id: string
  label: string
  recommendationsLabel: string
  queryDebounceMs?: number
  getRecommendations(context: AdminGlobalSearchRequestContext): AdminGlobalSearchResult[] | Promise<AdminGlobalSearchResult[]>
  search(query: string, context: AdminGlobalSearchRequestContext): AdminGlobalSearchResult[] | Promise<AdminGlobalSearchResult[]>
}

export interface AdminGlobalSearchResultGroup {
  sourceId: string
  label: string
  results: AdminGlobalSearchResult[]
  failed: boolean
  loading: boolean
}
