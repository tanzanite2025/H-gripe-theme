<template>
  <MediaAssetPickerDialog
    :open="open"
    :selected-urls="selectedUrls"
    @update:open="emit('update:open', $event)"
    @select="emit('select', $event)"
  />
</template>

<script setup lang="ts">
import MediaAssetPickerDialog from '@/components/admin/media/MediaAssetPickerDialog.vue'

interface MediaAssetSelection {
  url: string
  image?: unknown
  asset?: unknown
}

const props = withDefaults(defineProps<{
  open?: boolean
  selectedUrls?: string[]
}>(), {
  open: false,
  selectedUrls: () => [],
})

const emit = defineEmits<{
  (event: 'update:open', value: boolean): void
  (event: 'select', value: MediaAssetSelection): void
}>()
</script>
