<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import { Calculator, ClipboardCheck, Globe2, PackageCheck, RadioTower, Send, Settings2 } from '@lucide/vue'
import { yanwenLogisticsTabs } from '@/lib/logisticsDomainRegistry'
import LogisticsDomainHeader from '@/components/admin/logistics/LogisticsDomainHeader.vue'
import YanwenOverviewTab from '@/components/admin/logistics/yanwen/OverviewTab.vue'
import YanwenWaybillsTab from '@/components/admin/logistics/yanwen/WaybillsTab.vue'
import YanwenCollectionTab from '@/components/admin/logistics/yanwen/CollectionTab.vue'
import YanwenCalculatorTab from '@/components/admin/logistics/yanwen/CalculatorTab.vue'
import YanwenTrackingTab from '@/components/admin/logistics/yanwen/TrackingTab.vue'
import YanwenCustomsTab from '@/components/admin/logistics/yanwen/CustomsTab.vue'
import YanwenConfigTab from '@/components/admin/logistics/yanwen/ConfigTab.vue'

const route = useRoute()

const yanwenTabIcons = {
  overview: Globe2,
  waybills: Send,
  collection: PackageCheck,
  calculator: Calculator,
  tracking: RadioTower,
  customs: ClipboardCheck,
  config: Settings2,
} as const
const tabs = yanwenLogisticsTabs.map((tab) => ({
  ...tab,
  icon: yanwenTabIcons[tab.key as keyof typeof yanwenTabIcons],
}))

const activeTab = computed(() => tabs.find((tab) => tab.routeName === route.name)?.key ?? 'overview')
const activeDefinition = computed(() => tabs.find((tab) => tab.key === activeTab.value) ?? tabs[0])

const tabComponents = {
  overview: YanwenOverviewTab,
  waybills: YanwenWaybillsTab,
  collection: YanwenCollectionTab,
  calculator: YanwenCalculatorTab,
  tracking: YanwenTrackingTab,
  customs: YanwenCustomsTab,
  config: YanwenConfigTab,
} as const
const activeComponent = computed(() => tabComponents[activeTab.value as keyof typeof tabComponents] ?? tabComponents.overview)

</script>

<template>
  <div class="space-y-5">
    <LogisticsDomainHeader theme="emerald" badge="YANWEN DOMAIN" domain-label="燕文独立物流域" title="燕文专线物流中台" description="独立承载燕文专线运单、服务集合、轨迹、关务与网关配置，不与 4PX 域或主运费域混用。" " :active-label="activeDefinition.label" />
    <component :is="activeComponent" />
  </div>
</template>