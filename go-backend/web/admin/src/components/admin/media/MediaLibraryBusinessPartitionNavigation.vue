<template>
  <nav aria-label="媒体业务分区" class="flex shrink-0 flex-wrap gap-1.5 rounded-xl border bg-background p-2">
    <Button v-for="partition in mediaLibraryBusinessPartitions" :key="partition.value" size="sm"
      :variant="selectedPartition === partition.value ? 'default' : 'ghost'"
      :aria-current="selectedPartition === partition.value ? 'page' : undefined"
      @click="emit('select', partition.value)">
      {{ partition.label }}
      <span v-if="counts[partition.value] !== undefined" class="tabular-nums opacity-70">{{ counts[partition.value] }}</span>
    </Button>
  </nav>
</template>

<script setup lang="ts">
import { Button } from '@/components/ui/button'
import { mediaLibraryBusinessPartitions, type MediaLibraryPartitionView } from '@/modules/media/mediaLibraryBusinessPartitionUploadOriginAndCustomerServicePurposePresentation'
defineProps<{ selectedPartition: MediaLibraryPartitionView; counts: Record<string, number> }>()
const emit = defineEmits<{ (event: 'select', partition: MediaLibraryPartitionView): void }>()
</script>
