<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import { yanwenLogisticsTabs } from '@/lib/logisticsDomainRegistry'
import YanwenOverviewTab from '@/components/admin/logistics/yanwen/OverviewTab.vue'
import YanwenWaybillsTab from '@/components/admin/logistics/yanwen/WaybillsTab.vue'
import YanwenCollectionTab from '@/components/admin/logistics/yanwen/CollectionTab.vue'
import YanwenTrackingTab from '@/components/admin/logistics/yanwen/TrackingTab.vue'
import YanwenCustomsTab from '@/components/admin/logistics/yanwen/CustomsTab.vue'
import YanwenConfigTab from '@/components/admin/logistics/yanwen/ConfigTab.vue'

const route = useRoute()

const activeTab = computed(() => yanwenLogisticsTabs.find((tab) => tab.routeName === route.name)?.key ?? 'overview')

const tabComponents = {
  overview: YanwenOverviewTab,
  waybills: YanwenWaybillsTab,
  collection: YanwenCollectionTab,
  tracking: YanwenTrackingTab,
  customs: YanwenCustomsTab,
  config: YanwenConfigTab,
} as const
const activeComponent = computed(() => tabComponents[activeTab.value as keyof typeof tabComponents] ?? tabComponents.overview)

</script>

<template>
  <div class="space-y-5">
    <component :is="activeComponent" />
  </div>
</template>
