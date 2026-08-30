<template>
  <Dialog :open="open" @update:open="emit('update:open', $event)">
    <DialogContent size="xl" class="gap-0 p-0" @open-auto-focus.prevent>
      <DialogHeader class="border-b px-5 py-4 pr-12">
        <DialogTitle>选择媒体库{{ dialogMediaTypeLabel }}</DialogTitle>
        <DialogDescription>从通用媒体库选择{{ dialogMediaTypeLabel }}资源，图库不再作为附件仓库。</DialogDescription>
      </DialogHeader>

      <div class="flex min-h-[30rem] max-h-[74dvh] min-w-0 flex-col overflow-hidden">
        <div class="flex shrink-0 flex-wrap items-center gap-3 border-b px-4 py-3">
          <div class="relative min-w-[16rem] flex-1">
            <Search class="pointer-events-none absolute left-3 top-1/2 size-3.5 -translate-y-1/2 text-muted-foreground" />
            <Input v-model="filters.search" class="pl-8" placeholder="搜索文件名、ALT、说明或 URL" @keyup.enter="reload" />
          </div>
          <Select v-model="filters.status">
            <SelectTrigger class="w-32">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="active">启用</SelectItem>
              <SelectItem value="archived">归档</SelectItem>
              <SelectItem value="all">全部状态</SelectItem>
            </SelectContent>
          </Select>
          <Button variant="outline" size="sm" :disabled="loading" @click="reload">
            <RefreshCw :class="['size-3.5', { 'animate-spin': loading }]" />
            刷新
          </Button>
        </div>

        <div v-if="loading" class="flex min-h-0 flex-1 items-center justify-center text-xs font-bold text-muted-foreground">
          正在加载媒体库
        </div>
        <div v-else-if="assets.length === 0" class="flex min-h-0 flex-1 flex-col items-center justify-center gap-2 text-muted-foreground">
          <Video v-if="mediaType === 'video'" class="size-8 opacity-50" />
          <ImageOff v-else class="size-8 opacity-50" />
          <span class="text-xs font-bold">暂无可选{{ dialogMediaTypeLabel }}</span>
        </div>
        <div v-else class="grid min-h-0 flex-1 auto-rows-min grid-cols-2 gap-3 overflow-y-auto p-4 md:grid-cols-3 xl:grid-cols-5">
          <button
            v-for="asset in assets"
            :key="asset.id"
            type="button"
            class="group min-w-0 overflow-hidden rounded-2xl border bg-background text-left shadow-sm transition hover:-translate-y-0.5 hover:shadow-md focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
            :class="isAssetSelected(asset) ? 'border-[var(--admin-selected)]' : 'border-border'"
            @click="selectAsset(asset)"
          >
            <span class="block aspect-[4/3] overflow-hidden bg-muted">
              <img
                v-if="asset.media_type !== 'video' && assetAccessURL(asset)"
                :src="assetAccessURL(asset)"
                :alt="asset.alt || asset.original_filename || asset.filename || ''"
                class="size-full object-cover transition duration-200 group-hover:scale-[1.03]"
                @load="handleMediaLoad($event, asset)"
              />
              <video
                v-else-if="asset.media_type === 'video' && assetAccessURL(asset)"
                :src="assetAccessURL(asset)"
                class="size-full object-cover"
                muted
                playsinline
                preload="metadata"
                @loadedmetadata="handleMediaLoad($event, asset)"
              />
              <span v-else class="flex size-full items-center justify-center text-muted-foreground">
                <Video v-if="asset.media_type === 'video'" class="size-6 opacity-50" />
                <Images v-else class="size-6 opacity-50" />
              </span>
            </span>
            <span class="block min-w-0 space-y-1 px-3 py-2">
              <span class="block truncate text-xs font-black">{{ assetTitle(asset) }}</span>
              <span class="flex items-center justify-between gap-2 text-[10px] font-bold text-muted-foreground">
                <span>{{ mediaTypeLabel(asset.media_type) }} · {{ assetDimensionLabel(asset) }}</span>
                <span>{{ formatMediaSize(asset.size) }}</span>
              </span>
              <span class="flex justify-end text-[10px] font-bold text-muted-foreground">
                <span :class="isAssetSelected(asset) ? 'text-[var(--admin-selected)]' : ''">
                  {{ isAssetSelected(asset) ? '已加入' : '选择' }}
                </span>
              </span>
            </span>
          </button>
        </div>

        <div class="shrink-0 border-t px-4 py-3">
          <AdminPagination
            :page="pagination.page"
            :page-size="pagination.pageSize"
            :total="pagination.total"
            :page-sizes="[20, 40, 80]"
            @update:page="updatePage"
            @update:page-size="updatePageSize"
          />
        </div>
      </div>
    </DialogContent>
  </Dialog>
</template>

<script setup lang="ts">
import { computed, reactive, toRef } from 'vue'
import { ImageOff, Images, RefreshCw, Search, Video } from '@lucide/vue'
import AdminPagination from '@/components/admin/AdminPagination.vue'
import { Button } from '@/components/ui/button'
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { useMediaAssetPicker } from '@/composables/media/useMediaAssetPicker'
import { assetAccessURL, assetTitle, formatMediaDimensions, formatMediaSize, mediaTypeLabel } from '@/lib/mediaPresentation'
import type { MediaAsset } from '@/api/media'

interface MediaAssetSelection {
  url: string
  image: MediaAsset
  asset: MediaAsset
}

const props = withDefaults(defineProps<{
  open?: boolean
  selectedUrls?: string[]
  mediaType?: 'image' | 'video' | 'all'
}>(), {
  open: false,
  selectedUrls: () => [],
  mediaType: 'image',
})

const emit = defineEmits<{
  (event: 'update:open', value: boolean): void
  (event: 'select', selection: MediaAssetSelection): void
}>()

const mediaType = toRef(props, 'mediaType')
const dialogMediaTypeLabel = computed(() => mediaTypeLabel(mediaType.value))
const {
  loading,
  assets,
  filters,
  pagination,
  reload,
  updatePage,
  updatePageSize,
} = useMediaAssetPicker(toRef(props, 'open'), mediaType)

const loadedDimensions = reactive<Record<string, { width: number; height: number }>>({})

const assetURL = (asset: MediaAsset): string => String(asset.url || asset.access_url || '').trim()
const isAssetSelected = (asset: MediaAsset): boolean => {
  const urls = [String(asset.url || '').trim(), String(asset.access_url || '').trim(), String(assetAccessURL(asset) || '').trim()]
    .filter(Boolean)
  return urls.some((url) => props.selectedUrls.includes(url))
}

const assetDimensionLabel = (asset: MediaAsset): string => {
  const storedWidth = Number(asset.width || 0)
  const storedHeight = Number(asset.height || 0)
  if (storedWidth > 0 && storedHeight > 0) {
    return formatMediaDimensions(storedWidth, storedHeight)
  }

  const loaded = loadedDimensions[String(asset.id)]
  return loaded
    ? formatMediaDimensions(loaded.width, loaded.height)
    : '读取尺寸中'
}

const handleMediaLoad = (event: Event, asset: MediaAsset): void => {
  const media = event.currentTarget as HTMLImageElement | HTMLVideoElement | null
  if (!media) return

  const width = 'naturalWidth' in media ? media.naturalWidth : media.videoWidth
  const height = 'naturalHeight' in media ? media.naturalHeight : media.videoHeight
  if (width <= 0 || height <= 0) return

  loadedDimensions[String(asset.id)] = {
    width,
    height,
  }
}

const selectAsset = (asset: MediaAsset): void => {
  const url = assetURL(asset)
  if (!url) return
  emit('select', { url, image: asset, asset })
}

</script>
