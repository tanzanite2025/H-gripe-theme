<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import { Globe2, PackageCheck, Settings2 } from '@lucide/vue'
import { fpxLogisticsTabs } from '@/lib/logisticsDomainRegistry'
import FpxOverviewTab from '@/components/admin/logistics/fpx/OverviewTab.vue'
import FpxCollectionTab from '@/components/admin/logistics/fpx/CollectionTab.vue'
import FpxConfigTab from '@/components/admin/logistics/fpx/ConfigTab.vue'

const route = useRoute()

const fpxTabIcons = {
  overview: Globe2,
  collection: PackageCheck,
  config: Settings2,
} as const
const tabs = fpxLogisticsTabs.map((tab) => ({
  ...tab,
  icon: fpxTabIcons[tab.key as keyof typeof fpxTabIcons],
}))

const activeTab = computed(() => tabs.find((tab) => tab.routeName === route.name)?.key ?? 'overview')

const tabComponents = {
  overview: FpxOverviewTab,
  collection: FpxCollectionTab,
  config: FpxConfigTab,
} as const
const activeComponent = computed(() => tabComponents[activeTab.value as keyof typeof tabComponents] ?? tabComponents.overview)

</script>

<template>
  <div class="space-y-5">
    <component :is="activeComponent" />
  </div>
</template>
