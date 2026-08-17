<template>
  <DesktopFaqMasterDetail
    class="global-all-faqs-desktop-grouped-search-results-panel"
    :items="items"
    :expanded-items="expandedItems"
    id-prefix="global-all-faqs-desktop-answer"
    @toggle-item="emit('toggle-item', $event)"
  />
</template>

<script setup lang="ts">
import { computed } from 'vue'
import DesktopFaqMasterDetail from '~/components/faq/DesktopFaqMasterDetail.vue'
import type {
  GlobalAllFaqFlatItem,
  GlobalAllFaqsDisplayGroup,
} from '~/data/faq'

const props = defineProps<{
  groups: GlobalAllFaqsDisplayGroup[]
  expandedItems: ReadonlySet<string>
}>()

const emit = defineEmits<{
  'toggle-item': [itemId: string]
}>()

const items = computed<GlobalAllFaqFlatItem[]>(() => (
  props.groups.flatMap(group => group.items)
))
</script>

<style scoped>
.global-all-faqs-desktop-grouped-search-results-panel {
  display: none !important;
  min-width: 0;
}

@media (min-width: 768px) {
  .global-all-faqs-desktop-grouped-search-results-panel {
    display: grid !important;
  }
}
</style>
