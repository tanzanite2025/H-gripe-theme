<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import { Globe2, PackageCheck, Settings2 } from '@lucide/vue'
import { fpxLogisticsTabs } from '@/lib/logisticsDomainRegistry'
import LogisticsDomainHeader from '@/components/admin/logistics/LogisticsDomainHeader.vue'
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
const activeDefinition = computed(() => tabs.find((tab) => tab.key === activeTab.value) ?? tabs[0])

const tabComponents = {
  overview: FpxOverviewTab,
  collection: FpxCollectionTab,
  config: FpxConfigTab,
} as const
const activeComponent = computed(() => tabComponents[activeTab.value as keyof typeof tabComponents] ?? tabComponents.overview)

</script>

<template>
  <div class="space-y-5">
    <LogisticsDomainHeader theme="orange" badge="4PX DOMAIN" domain-label="递四方独立物流域" title="4PX 服务目录中台" description="读取 4PX 官方服务目录，治理已启用服务集合并提供网关配置，不与主运费域或燕文域混用。" :active-label="activeDefinition.label" />
    <component :is="activeComponent" />
  </div>
</template>
