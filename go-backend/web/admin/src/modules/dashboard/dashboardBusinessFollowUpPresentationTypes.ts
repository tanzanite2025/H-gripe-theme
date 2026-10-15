import type { DashboardBusinessFollowUpItem, DashboardCommandCenterOutboxFailures } from '@/api/dashboardCommandCenterOutboxFailures'

export interface DashboardBusinessFollowUpPanelState {
  loading: boolean
  loadError: boolean
  outboxFailures: DashboardCommandCenterOutboxFailures | null
  lastSuccessfulTimestamp: string | null
  selectedEventNotice: string
  selectedEventOutsidePreview: DashboardBusinessFollowUpItem | null
  selectedEventId: number | null
  selectedProcessingState: string
  visibleBusinessFollowUps: DashboardBusinessFollowUpItem[]
}
