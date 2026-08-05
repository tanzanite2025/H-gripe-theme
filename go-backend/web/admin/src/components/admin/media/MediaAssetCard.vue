<template>
  <article class="group flex min-w-0 flex-col overflow-hidden rounded-2xl border border-dashed bg-background shadow-sm transition hover:-translate-y-0.5 hover:shadow-md">
    <button type="button" class="aspect-[4/3] overflow-hidden bg-muted text-left" @click="emit('preview', asset)">
      <img
        v-if="asset.media_type === 'image' && assetAccessURL(asset)"
        :src="assetAccessURL(asset)"
        :alt="asset.alt || asset.original_filename || asset.filename || ''"
        class="size-full object-cover transition duration-200 group-hover:scale-[1.03]"
      />
      <span v-else class="flex size-full items-center justify-center text-muted-foreground">
        <FileVideo v-if="asset.media_type === 'video'" class="size-8 opacity-60" />
        <Images v-else class="size-8 opacity-60" />
      </span>
    </button>

    <div class="flex min-w-0 flex-1 flex-col gap-3 p-3">
      <div class="min-w-0">
        <div class="flex items-center gap-2">
          <Badge :variant="asset.status === 'active' ? 'default' : 'secondary'">{{ statusLabel(asset.status) }}</Badge>
          <Badge variant="outline">{{ mediaTypeLabel(asset.media_type) }}</Badge>
        </div>
        <h2 class="mt-2 truncate text-sm font-black">{{ assetTitle(asset) }}</h2>
        <p class="mt-1 truncate font-mono text-[10px] text-muted-foreground">{{ asset.storage_key || asset.url }}</p>
      </div>

      <div class="grid grid-cols-2 gap-2 text-[10px] font-bold text-muted-foreground">
        <span>{{ formatMediaSize(asset.size) }}</span>
        <span class="text-right">{{ formatMediaDate(asset.created_at) }}</span>
      </div>

      <div class="mt-auto flex items-center justify-between gap-2 border-t pt-3">
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

<script setup>
import { Copy, FileArchive, FileVideo, Images, Pencil, Trash2 } from '@lucide/vue'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import {
  assetAccessURL,
  assetTitle,
  formatMediaDate,
  formatMediaSize,
  mediaTypeLabel,
  statusLabel,
} from '@/lib/mediaPresentation'

defineProps({
  asset: { type: Object, required: true },
  canEdit: { type: Boolean, default: false },
  canDelete: { type: Boolean, default: false },
})

const emit = defineEmits(['preview', 'copy-url', 'export-evidence', 'edit', 'delete'])
</script>
