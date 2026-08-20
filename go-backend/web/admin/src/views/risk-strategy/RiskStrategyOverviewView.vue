<template>
  <div class="flex h-full min-h-0 flex-col gap-3 overflow-hidden">
    <AdminPageHeader
      class="shrink-0"
      title="风控策略总览"
      description="查看 Stripe / PayPal 风险指标、策略口径、阈值和 3DS 决策统计。"
    >
      <template #actions>
        <Button variant="outline" size="sm" class="rounded-full font-black uppercase tracking-wider" :disabled="loading" @click="fetchSummary">
          <RefreshCw :class="['size-3.5', { 'animate-spin': loading }]" />
          刷新
        </Button>
        <Button
          v-if="canRecompute"
          variant="outline"
          size="sm"
          class="rounded-full font-black uppercase tracking-wider"
          :disabled="loading"
          @click="recomputeSummary(canRecompute)"
        >
          <RefreshCw :class="['size-3.5', { 'animate-spin': loading }]" />
          立即重算
        </Button>
      </template>
    </AdminPageHeader>

    <div class="min-h-0 flex-1 overflow-auto">
      <RiskStrategyPolicyPanel
        :configuration="summary.configuration"
        :gateway-runtime="summary.gatewayRuntime"
        :gateway-health="summary.gatewayHealth"
      />
      <RiskStrategySummaryPanel
        :reports="summary.reports"
        :policy="summary.policy"
        :enabled="summary.enabled"
        :loading="loading"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted } from 'vue'
import { RefreshCw } from '@lucide/vue'
import AdminPageHeader from '@/components/admin/AdminPageHeader.vue'
import RiskStrategyPolicyPanel from '@/components/admin/payment/RiskStrategyPolicyPanel.vue'
import RiskStrategySummaryPanel from '@/components/admin/payment/RiskStrategySummaryPanel.vue'
import { Button } from '@/components/ui/button'
import { useRiskStrategySummary } from '@/composables/useRiskStrategySummary'
import { useAuthStore } from '@/stores/auth'

const { loading, summary, fetchSummary, recomputeSummary } = useRiskStrategySummary()
const canRecompute = useAuthStore().hasPermission('order:edit')

onMounted(fetchSummary)
</script>

