<template>
  <div class="flex h-full min-h-0 flex-col gap-3 overflow-hidden">
    <AdminPageHeader
      class="shrink-0"
      title="Stripe 风控总览"
      description="只查看 Stripe 域的 30D 争议率、退款率、当前灯号、告警和网关熔断状态。"
    >
      <template #actions>
        <Button variant="outline" size="sm" class="rounded-full font-black uppercase tracking-wider" :disabled="stripeRiskStrategyOverviewSummaryLoading" @click="fetchStripeRiskStrategyOverviewSummary">
          <RefreshCw :class="['size-3.5', { 'animate-spin': stripeRiskStrategyOverviewSummaryLoading }]" />
          刷新
        </Button>
        <Button
          v-if="stripeRiskStrategyOverviewCanRecompute"
          variant="outline"
          size="sm"
          class="rounded-full font-black uppercase tracking-wider"
          :disabled="stripeRiskStrategyOverviewSummaryLoading"
          @click="recomputeStripeRiskStrategyOverviewSummary"
        >
          <RefreshCw :class="['size-3.5', { 'animate-spin': stripeRiskStrategyOverviewSummaryLoading }]" />
          立即重算
        </Button>
      </template>
    </AdminPageHeader>

    <div class="min-h-0 flex-1 overflow-auto">
      <SingleProviderRiskStrategyOverviewPanel
        payment-provider-display-name="Stripe"
        :risk-strategy-provider-report="stripeRiskStrategyOverviewSummary.reports.stripe"
        :risk-strategy-monitoring-policy="stripeRiskStrategyOverviewSummary.policy"
        :gateway-health="stripeRiskStrategyOverviewSummary.gatewayHealth.stripe"
        :monitoring-enabled="stripeRiskStrategyOverviewSummary.enabled"
        :loading="stripeRiskStrategyOverviewSummaryLoading"
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
  loading: stripeRiskStrategyOverviewSummaryLoading,
  summary: stripeRiskStrategyOverviewSummary,
  applySummary: applyStripeRiskStrategyOverviewSummary,
  fetchSummaryForProvider: fetchRiskStrategySummaryForProvider,
} = useRiskStrategySummary()
const stripeRiskStrategyOverviewCanRecompute = useAuthStore().hasPermission('order:edit')

const fetchStripeRiskStrategyOverviewSummary = async (): Promise<void> => {
  await fetchRiskStrategySummaryForProvider('stripe')
}

const recomputeStripeRiskStrategyOverviewSummary = async (): Promise<void> => {
  if (!stripeRiskStrategyOverviewCanRecompute) return
  stripeRiskStrategyOverviewSummaryLoading.value = true
  try {
    applyStripeRiskStrategyOverviewSummary(await riskStrategyApi.recomputeSummary('stripe'))
    toast.success('Stripe 风控策略指标已重算')
  } finally {
    stripeRiskStrategyOverviewSummaryLoading.value = false
  }
}

onMounted(fetchStripeRiskStrategyOverviewSummary)
</script>

