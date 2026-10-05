<template>
  <article class="group flex min-w-0 flex-col overflow-hidden rounded-2xl border border-dashed bg-background shadow-sm transition hover:-translate-y-0.5 hover:shadow-md">
    <button type="button" class="h-32 overflow-hidden bg-muted text-left sm:h-36" @click="emit('preview', asset)">
      <img
        v-if="asset.media_type === 'image' && thumbnailURL && !thumbnailUnavailable"
        :src="thumbnailURL"
        :alt="asset.alt || asset.original_filename || asset.filename || ''"
        class="size-full object-contain transition duration-200 group-hover:scale-[1.03]"
        @error="handleThumbnailError"
      />
      <span v-else class="flex size-full items-center justify-center text-muted-foreground">
        <FileVideo v-if="asset.media_type === 'video'" class="size-8 opacity-60" />
        <Images v-else class="size-8 opacity-60" />
      </span>
    </button>

    <div class="flex min-w-0 flex-1 flex-col gap-2 p-2.5">
      <div class="min-w-0">
        <div class="flex items-center gap-1.5">
          <Badge :variant="asset.status === 'active' ? 'default' : 'secondary'">{{ statusLabel(asset.status) }}</Badge>
          <Badge variant="outline">{{ mediaTypeLabel(asset.media_type) }}</Badge>
        </div>
        <h2 class="mt-1 truncate text-sm font-black">{{ assetTitle(asset) }}</h2>
        <p class="mt-0.5 truncate font-mono text-[10px] text-muted-foreground">{{ asset.storage_key || asset.url }}</p>
      </div>

      <div class="grid grid-cols-3 gap-1 text-[10px] font-bold text-muted-foreground">
        <span>{{ formatMediaDimensions(asset.width, asset.height) }}</span>
        <span class="text-right">{{ formatMediaSize(asset.size) }}</span>
        <span class="text-right">{{ formatMediaDate(typeof asset.created_at === 'string' || typeof asset.created_at === 'number' ? asset.created_at : null) }}</span>
      </div>

      <div class="mt-auto flex items-center justify-between gap-2 border-t pt-2">
        <Button variant="outline" size="xs" @click="emit('copy-url', asset)">
          <Copy class="size-3" />
          URL
        </Button>
        <div class="flex items-center gap-1">
          <Button
            v-if="canEdit"
            variant="outline"
            size="icon-xs"
            title="导出版权证据包"
            aria-label="导出版权证据包"
            @click="emit('export-evidence', asset)"
          >
            <FileArchive class="size-3" />
          </Button>
          <Button v-if="canEdit" variant="outline" size="icon-xs" aria-label="编辑媒体" @click="emit('edit', asset)">
            <Pencil class="size-3" />
          </Button>
          <Button v-if="canDelete" variant="destructive" size="icon-xs" aria-label="检查引用并删除媒体" @click="emit('delete', asset)">
            <Trash2 class="size-3" />
          </Button>
        </div>
      </div>
    </div>
  </article>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { Copy, FileArchive, FileVideo, Images, Pencil, Trash2 } from '@lucide/vue'
import type { MediaAsset } from '@/api/media'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import {
  adminMediaAssetFileURL,
  assetAccessURL,
  assetTitle,
  formatMediaDate,
  formatMediaDimensions,
  formatMediaSize,
  mediaTypeLabel,
  statusLabel,
} from '@/lib/mediaPresentation'

const props = withDefaults(defineProps<{
  asset: MediaAsset
  canEdit?: boolean
  canDelete?: boolean
}>(), {
  canEdit: false,
  canDelete: false
})

const useAdminFileFallback = ref(false)
const thumbnailUnavailable = ref(false)
const thumbnailURL = computed(() => {
  const assetURL = assetAccessURL(props.asset)
  return useAdminFileFallback.value || !assetURL
    ? adminMediaAssetFileURL(props.asset)
    : assetURL
})

watch(() => [props.asset.id, assetAccessURL(props.asset)], () => {
  useAdminFileFallback.value = false
  thumbnailUnavailable.value = false
})

const handleThumbnailError = (): void => {
  const adminFileURL = adminMediaAssetFileURL(props.asset)
  if (!useAdminFileFallback.value && adminFileURL && thumbnailURL.value !== adminFileURL) {
    useAdminFileFallback.value = true
    return
  }
  thumbnailUnavailable.value = true
}

const emit = defineEmits<{
  (event: 'preview', asset: MediaAsset): void
  (event: 'copy-url', asset: MediaAsset): void
  (event: 'export-evidence', asset: MediaAsset): void
  (event: 'edit', asset: MediaAsset): void
  (event: 'delete', asset: MediaAsset): void
}>()
</script>
