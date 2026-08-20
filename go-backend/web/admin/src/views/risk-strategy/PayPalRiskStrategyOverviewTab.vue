<template>
  <div class="flex h-full min-h-0 flex-col gap-3 overflow-hidden">
    <AdminPageHeader
      class="shrink-0"
      title="PayPal 风控总览"
      description="只查看 PayPal 域的 30D 争议率、退款率、当前灯号、告警和网关熔断状态。"
    >
      <template #actions>
        <Button variant="outline" size="sm" class="rounded-full font-black uppercase tracking-wider" :disabled="payPalRiskStrategyOverviewSummaryLoading" @click="fetchPayPalRiskStrategyOverviewSummary">
          <RefreshCw :class="['size-3.5', { 'animate-spin': payPalRiskStrategyOverviewSummaryLoading }]" />
          刷新
        </Button>
        <Button
          v-if="payPalRiskStrategyOverviewCanRecompute"
          variant="outline"
          size="sm"
          class="rounded-full font-black uppercase tracking-wider"
          :disabled="payPalRiskStrategyOverviewSummaryLoading"
          @click="recomputePayPalRiskStrategyOverviewSummary"
        >
          <RefreshCw :class="['size-3.5', { 'animate-spin': payPalRiskStrategyOverviewSummaryLoading }]" />
          立即重算
        </Button>
      </template>
    </AdminPageHeader>

    <div class="min-h-0 flex-1 overflow-auto">
      <SingleProviderRiskStrategyOverviewPanel
        payment-provider-display-name="PayPal"
        :risk-strategy-provider-report="payPalRiskStrategyOverviewSummary.reports.paypal"
        :risk-strategy-monitoring-policy="payPalRiskStrategyOverviewSummary.policy"
        :gateway-health="payPalRiskStrategyOverviewSummary.gatewayHealth.paypal"
        :monitoring-enabled="payPalRiskStrategyOverviewSummary.enabled"
        :loading="payPalRiskStrategyOverviewSummaryLoading"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted } from 'vue'
import { RefreshCw } from '@lucide/vue'
import { toast } from 'vue-sonner'
import AdminPageHeader from '@/components/admin/AdminPageHeader.vue'
import SingleProviderRiskStrategyOverviewPanel from '@/components/admin/payment/SingleProviderRiskStrategyOverviewPanel.vue'
import { Button } from '@/components/ui/button'
import { paymentRiskApi as riskStrategyApi } from '@/api/paymentRisk'
import { useRiskStrategySummary } from '@/composables/useRiskStrategySummary'
import { useAuthStore } from '@/stores/auth'

const {
  loading: payPalRiskStrategyOverviewSummaryLoading,
  summary: payPalRiskStrategyOverviewSummary,
  applySummary: applyPayPalRiskStrategyOverviewSummary,
  fetchSummaryForProvider: fetchRiskStrategySummaryForProvider,
} = useRiskStrategySummary()
const payPalRiskStrategyOverviewCanRecompute = useAuthStore().hasPermission('order:edit')

const fetchPayPalRiskStrategyOverviewSummary = async (): Promise<void> => {
  await fetchRiskStrategySummaryForProvider('paypal')
}

const recomputePayPalRiskStrategyOverviewSummary = async (): Promise<void> => {
  if (!payPalRiskStrategyOverviewCanRecompute) return
  payPalRiskStrategyOverviewSummaryLoading.value = true
  try {
    applyPayPalRiskStrategyOverviewSummary(await riskStrategyApi.recomputeSummary('paypal'))
    toast.success('PayPal 风控策略指标已重算')
  } finally {
    payPalRiskStrategyOverviewSummaryLoading.value = false
  }
}

onMounted(fetchPayPalRiskStrategyOverviewSummary)
</script>

