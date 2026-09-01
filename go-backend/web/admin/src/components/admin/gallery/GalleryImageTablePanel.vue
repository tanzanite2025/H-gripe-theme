<template>
  <AdminTablePanel :loading="loading" :batch-visible="selectedImages.length > 0">
    <template #batch>
      <div class="flex flex-wrap items-center justify-between gap-2">
        <span class="text-xs font-medium">已选择 {{ selectedImages.length }} 张图片</span>
        <Button v-if="canDelete" variant="destructive" size="sm" @click="emit('batch-delete')">
          <Trash2 class="size-3.5" />
          批量删除
        </Button>
      </div>
    </template>

    <Table class="min-w-[920px]">
      <TableHeader>
        <TableRow>
          <TableHead class="w-11">
            <Checkbox
              :model-value="imageSelectionState"
              aria-label="选择当前图库全部图片"
              @update:model-value="emit('toggle-all', $event)"
            />
          </TableHead>
          <TableHead class="w-28">图片</TableHead>
          <TableHead>标题</TableHead>
          <TableHead>描述</TableHead>
          <TableHead class="w-40">标签</TableHead>
          <TableHead class="w-20 text-right">排序</TableHead>
          <TableHead class="w-16 text-right">操作</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        <TableEmpty v-if="images.length === 0" :colspan="7">
          <div class="flex flex-col items-center text-muted-foreground">
            <ImageOff class="mb-2 size-7 opacity-55" />
            <span class="text-xs">该图库暂无图片</span>
          </div>
        </TableEmpty>

        <TableRow v-for="image in images" :key="image.id">
          <TableCell>
            <Checkbox
              :model-value="isImageSelected(image.id)"
              :aria-label="`选择图片 ${image.title}`"
              @update:model-value="emit('toggle-image', image, $event)"
            />
          </TableCell>
          <TableCell>
            <button
              type="button"
              class="block h-16 w-20 overflow-hidden rounded-md border bg-muted focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
              :aria-label="`预览图片 ${image.title}`"
              @click="emit('preview', image.url, image.title)"
            >
              <img :src="image.thumbnail || image.url" :alt="image.title" class="size-full object-cover" />
            </button>
          </TableCell>
 <TableCell class="max-w-64 font-bold text-xs">{{ image.title || '-'}}</TableCell>
 <TableCell class="max-w-72"><p class="line-clamp-2 text-muted-foreground">{{ image.description || '-'}}</p></TableCell>
 <TableCell class="max-w-40 truncate text-xs text-muted-foreground">{{ image.tags || '-'}}</TableCell>
          <TableCell class="text-right tabular-nums">{{ image.order ?? image.sort_order ?? 0 }}</TableCell>
          <TableCell class="text-right">
            <DropdownMenu>
              <DropdownMenuTrigger as-child>
                <Button variant="ghost" size="icon" :aria-label="`管理图片 ${image.title}`">
                  <MoreHorizontal class="size-4" />
                </Button>
              </DropdownMenuTrigger>
              <DropdownMenuContent align="end" class="w-36">
                <DropdownMenuItem v-if="canEdit" @select="emit('edit', image)">
                  <Pencil class="size-4" />
                  编辑
                </DropdownMenuItem>
                <DropdownMenuSeparator v-if="canDelete" />
                <DropdownMenuItem
                  v-if="canDelete"
                  class="text-destructive focus:text-destructive"
                  @select="emit('delete', image)"
                >
                  <Trash2 class="size-4" />
                  删除
                </DropdownMenuItem>
              </DropdownMenuContent>
            </DropdownMenu>
          </TableCell>
        </TableRow>
      </TableBody>
    </Table>
  </AdminTablePanel>
</template>

<script setup lang="ts">
import { ImageOff, MoreHorizontal, Pencil, Trash2 } from '@lucide/vue'
import AdminTablePanel from '@/components/admin/AdminTablePanel.vue'
import { Button } from '@/components/ui/button'
import { Checkbox } from '@/components/ui/checkbox'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { Table, TableBody, TableCell, TableEmpty, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import type { GalleryId, GalleryImage, GallerySelectionState } from '@/modules/gallery/galleryTypes'

const props = withDefaults(defineProps<{
  loading?: boolean
  images?: GalleryImage[]
  selectedImages?: GalleryImage[]
  imageSelectionState?: GallerySelectionState
  canEdit?: boolean
  canDelete?: boolean
}>(), {
  loading: false,
  images: () => [],
  selectedImages: () => [],
  imageSelectionState: false,
  canEdit: false,
  canDelete: false
})

const emit = defineEmits<{
  (event: 'batch-delete'): void
  (event: 'toggle-all', checked: GallerySelectionState): void
  (event: 'toggle-image', image: GalleryImage, checked: GallerySelectionState): void
  (event: 'preview', url?: string | null, title?: string | null): void
  (event: 'edit', image: GalleryImage): void
  (event: 'delete', image: GalleryImage): void
}>()
const isImageSelected = (imageId: GalleryId): boolean => props.selectedImages.some((image) => image.id === imageId)
</script>

