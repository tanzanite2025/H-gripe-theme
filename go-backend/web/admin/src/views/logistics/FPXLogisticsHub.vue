<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import { Calculator, ClipboardCheck, Globe2, PackageCheck, RadioTower, Settings2, Truck } from '@lucide/vue'
import { fpxLogisticsTabs } from '@/lib/logisticsDomainRegistry'
import LogisticsDomainHeader from '@/components/admin/logistics/LogisticsDomainHeader.vue'
import FpxOverviewTab from '@/components/admin/logistics/fpx/OverviewTab.vue'
import FpxDirectTab from '@/components/admin/logistics/fpx/DirectTab.vue'
import FpxCollectionTab from '@/components/admin/logistics/fpx/CollectionTab.vue'
import FpxCalculatorTab from '@/components/admin/logistics/fpx/CalculatorTab.vue'
import FpxTrackingTab from '@/components/admin/logistics/fpx/TrackingTab.vue'
import FpxRmaTab from '@/components/admin/logistics/fpx/RmaTab.vue'
import FpxConfigTab from '@/components/admin/logistics/fpx/ConfigTab.vue'

const route = useRoute()

const fpxTabIcons = {
  overview: Globe2,
  direct: Truck,
  collection: PackageCheck,
  calculator: Calculator,
  tracking: RadioTower,
  rma: ClipboardCheck,
  config: Settings2,
} as const
const tabs = fpxLogisticsTabs.map((tab) => ({
  ...tab,
  icon: fpxTabIcons[tab.key as keyof typeof fpxTabIcons],
}))

const activeTab = computed(() => tabs.find((tab) => tab.routeName === route.name)?.key ?? 'overview')
const activeDefinition = computed(() => tabs.find((tab) => tab.key === activeTab.value) ?? tabs[0])

const tabComponents = {
  overview: FpxOverviewTab,
  direct: FpxDirectTab,
  collection: FpxCollectionTab,
  calculator: FpxCalculatorTab,
  tracking: FpxTrackingTab,
  rma: FpxRmaTab,
  config: FpxConfigTab,
} as const
const activeComponent = computed(() => tabComponents[activeTab.value as keyof typeof tabComponents] ?? tabComponents.overview)

</script>

<template>
  <div class="space-y-5">
    <LogisticsDomainHeader theme="orange" badge="4PX DOMAIN" domain-label="递四方独立物流域" title="4PX 发货中台" description="独立承载大件专线发货、面单、运费试算、轨迹与接口配置，不与主运费域或燕文域混用。" :active-label="activeDefinition.label" />
    <component :is="activeComponent" />
  </div>
</template>
