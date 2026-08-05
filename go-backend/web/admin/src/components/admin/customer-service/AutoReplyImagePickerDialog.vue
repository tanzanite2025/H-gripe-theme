<template>
  <MediaAssetPickerDialog
    :open="open"
    :selected-urls="selectedUrls"
    @update:open="emit('update:open', $event)"
    @select="emit('select', $event)"
  />
</template>

<script setup>
import MediaAssetPickerDialog from '@/components/admin/media/MediaAssetPickerDialog.vue'

defineProps({
  open: { type: Boolean, default: false },
  selectedUrls: { type: Array, default: () => [] },
})

const emit = defineEmits(['update:open', 'select'])
</script>
