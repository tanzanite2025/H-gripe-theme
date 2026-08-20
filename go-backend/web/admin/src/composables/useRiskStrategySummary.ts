import { reactive, ref } from 'vue'
import { toast } from 'vue-sonner'
import { paymentRiskApi } from '@/api/paymentRisk'

export interface RiskStrategySummaryState {
  enabled: boolean
  policy: Record<string, any> | null
  reports: Record<string, any>
  configuration: Record<string, any> | null
  gatewayRuntime: Record<string, any> | null
  gatewayHealth: Record<string, any>
}

export function useRiskStrategySummary() {
  const loading = ref(false)
  const summary = reactive<RiskStrategySummaryState>({
    enabled: false,
    policy: null,
    reports: {},
    configuration: null,
    gatewayRuntime: null,
    gatewayHealth: {},
  })

  const applySummary = (payload: Record<string, any>): void => {
    summary.enabled = Boolean(payload.enabled)
    summary.policy = payload.policy || null
    summary.reports = payload.reports || {}
    summary.configuration = payload.configuration || null
    summary.gatewayRuntime = payload.gateway_runtime || null
    summary.gatewayHealth = payload.gateway_health || {}
  }

  const fetchSummaryForProvider = async (provider = ''): Promise<void> => {
    loading.value = true
    try {
      applySummary(await paymentRiskApi.getSummary(provider))
    } finally {
      loading.value = false
    }
  }

  const fetchSummary = async (): Promise<void> => {
    await fetchSummaryForProvider()
  }

  const recomputeSummaryForProvider = async (provider: string, canRecompute: boolean): Promise<void> => {
    if (!canRecompute) return
    loading.value = true
    try {
      applySummary(await paymentRiskApi.recomputeSummary(provider))
      toast.success('风控策略指标已重算')
    } finally {
      loading.value = false
    }
  }

  const recomputeSummary = async (canRecompute: boolean): Promise<void> => {
    await recomputeSummaryForProvider('', canRecompute)
  }

  return {
    loading,
    summary,
    applySummary,
    fetchSummary,
    fetchSummaryForProvider,
    recomputeSummary,
    recomputeSummaryForProvider,
  }
}
