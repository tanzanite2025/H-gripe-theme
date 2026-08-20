<script setup lang="ts">
import { computed, defineAsyncComponent } from 'vue'
import { useRouteTab } from '@/composables/useRouteTab'

const props = withDefaults(defineProps<{
  defaultDisputeProvider?: string
}>(), {
  defaultDisputeProvider: '',
})

const riskTabValues = ['overview', 'three_ds', 'reviews', 'refunds', 'disputes', 'controls'] as const
type RiskTabValue = typeof riskTabValues[number]

const activeTab = useRouteTab<RiskTabValue>({
  defaultValue: 'overview',
  values: [...riskTabValues],
  routes: {
    overview: 'RiskStrategyOverview',
    three_ds: ['RiskStrategyThreeDS', 'PaymentStripeThreeDS'],
    reviews: 'RiskStrategyReviews',
    refunds: 'RiskStrategyRefundRecommendations',
    disputes: ['RiskStrategyDisputes', 'PaymentStripeDisputes', 'PaymentPayPalDisputes'],
    controls: 'RiskStrategyControls',
  },
})

const riskStrategyViews = {
  overview: defineAsyncComponent(() => import('./risk-strategy/RiskStrategyOverviewView.vue')),
  three_ds: defineAsyncComponent(() => import('./risk-strategy/RiskStrategyThreeDSView.vue')),
  reviews: defineAsyncComponent(() => import('./risk-strategy/RiskStrategyReviewsView.vue')),
  refunds: defineAsyncComponent(() => import('./risk-strategy/RiskStrategyRefundRecommendationsView.vue')),
  disputes: defineAsyncComponent(() => import('./risk-strategy/RiskStrategyDisputesView.vue')),
  controls: defineAsyncComponent(() => import('./risk-strategy/RiskStrategyControlsView.vue')),
}

const activeView = computed(() => riskStrategyViews[activeTab.value])

const activeViewProps = computed(() => (
  activeTab.value === 'disputes'
    ? { defaultDisputeProvider: props.defaultDisputeProvider }
    : {}
))
</script>

<template>
  <component :is="activeView" v-bind="activeViewProps" />
</template>
