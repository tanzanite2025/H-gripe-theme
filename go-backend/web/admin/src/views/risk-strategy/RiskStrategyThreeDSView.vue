<template>
  <div class="flex h-full min-h-0 flex-col gap-3 overflow-hidden">
    <AdminPageHeader
      class="shrink-0"
      title="3DS 策略"
      description="解释 Stripe 基础模式、自适应升级、人工保护和 3DS 统计的真实含义。"
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
      <RiskStrategyThreeDSPanel
        :configuration="summary.configuration"
        :gateway-runtime="summary.gatewayRuntime"
        :gateway-health="summary.gatewayHealth"
        :reports="summary.reports"
        :enabled="summary.enabled"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted } from 'vue'
import { RefreshCw } from '@lucide/vue'
import AdminPageHeader from '@/components/admin/AdminPageHeader.vue'
import RiskStrategyThreeDSPanel from '@/components/admin/payment/RiskStrategyThreeDSPanel.vue'
import { Button } from '@/components/ui/button'
import { useRiskStrategySummary } from '@/composables/useRiskStrategySummary'
import { useAuthStore } from '@/stores/auth'

const { loading, summary, fetchSummary, recomputeSummary } = useRiskStrategySummary()
const canRecompute = useAuthStore().hasPermission('order:edit')

onMounted(fetchSummary)
</script>

