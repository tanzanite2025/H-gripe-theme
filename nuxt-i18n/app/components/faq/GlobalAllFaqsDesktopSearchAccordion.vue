<template>
  <DesktopFaqMasterDetail
    class="global-all-faqs-desktop-search-accordion"
    :items="items"
    :expanded-items="expandedItems"
    id-prefix="global-faq-desktop-answer"
    @toggle-item="emit('toggle-item', $event)"
  />
</template>

<script setup lang="ts">
import DesktopFaqMasterDetail from '~/components/faq/DesktopFaqMasterDetail.vue'
import type { GlobalAllFaqFlatItem } from '~/data/faq'

defineProps<{
  items: GlobalAllFaqFlatItem[]
  expandedItems: ReadonlySet<string>
}>()

const emit = defineEmits<{
  'toggle-item': [itemId: string]
}>()
</script>
