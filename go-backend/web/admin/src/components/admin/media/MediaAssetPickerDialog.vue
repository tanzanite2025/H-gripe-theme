<template>
  <Dialog :open="open" @update:open="emit('update:open', $event)">
    <DialogContent size="xl" class="gap-0 p-0" @open-auto-focus.prevent>
      <DialogHeader class="border-b px-5 py-4 pr-12">
        <DialogTitle>选择媒体库图片</DialogTitle>
        <DialogDescription>从通用媒体库选择图片资源，图库不再作为附件仓库。</DialogDescription>
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
          <ImageOff class="size-8 opacity-50" />
          <span class="text-xs font-bold">暂无可选图片</span>
        </div>
        <div v-else class="grid min-h-0 flex-1 auto-rows-min grid-cols-2 gap-3 overflow-y-auto p-4 md:grid-cols-3 xl:grid-cols-5">
          <button
            v-for="asset in assets"
            :key="asset.id"
            type="button"
            class="group min-w-0 overflow-hidden rounded-2xl border bg-background text-left shadow-sm transition hover:-translate-y-0.5 hover:shadow-md focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
            :class="isSelected(asset.url) ? 'border-[var(--admin-selected)]' : 'border-border'"
            @click="selectAsset(asset)"
          >
            <span class="block aspect-[4/3] overflow-hidden bg-muted">
              <img
                v-if="assetAccessURL(asset)"
                :src="assetAccessURL(asset)"
                :alt="asset.alt || asset.original_filename || asset.filename || ''"
                class="size-full object-cover transition duration-200 group-hover:scale-[1.03]"
              />
              <span v-else class="flex size-full items-center justify-center text-muted-foreground">
                <Images class="size-6 opacity-50" />
              </span>
            </span>
            <span class="block min-w-0 space-y-1 px-3 py-2">
              <span class="block truncate text-xs font-black">{{ assetTitle(asset) }}</span>
              <span class="flex items-center justify-between gap-2 text-[10px] font-bold text-muted-foreground">
                <span>{{ formatMediaSize(asset.size) }}</span>
                <span :class="isSelected(asset.url) ? 'text-[var(--admin-selected)]' : ''">
                  {{ isSelected(asset.url) ? '已加入' : '选择' }}
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
import { toRef } from 'vue'
import { ImageOff, Images, RefreshCw, Search } from '@lucide/vue'
import AdminPagination from '@/components/admin/AdminPagination.vue'
import { Button } from '@/components/ui/button'
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { useMediaAssetPicker } from '@/composables/media/useMediaAssetPicker'
import { assetAccessURL, assetTitle, formatMediaSize } from '@/lib/mediaPresentation'
import type { MediaAsset } from '@/api/media'

interface MediaAssetSelection {
  url: string
  image: MediaAsset
  asset: MediaAsset
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
  (event: 'select', selection: MediaAssetSelection): void
}>()

const isSelected = (url?: string | null): boolean => Boolean(url) && props.selectedUrls.includes(String(url))
const {
  loading,
  assets,
  filters,
  pagination,
  reload,
  updatePage,
  updatePageSize,
} = useMediaAssetPicker(toRef(props, 'open'))

const selectAsset = (asset: MediaAsset): void => {
  if (!asset?.url) return
  emit('select', { url: asset.url, image: asset, asset })
}

</script>
