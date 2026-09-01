<template>
  <AdminTablePanel :loading="loading">
    <Table class="min-w-[940px]">
      <TableHeader>
        <TableRow>
          <TableHead class="w-16">ID</TableHead>
          <TableHead class="w-28">封面</TableHead>
          <TableHead>图库</TableHead>
          <TableHead>描述</TableHead>
          <TableHead class="w-24 text-right">图片数</TableHead>
          <TableHead class="w-44">创建时间</TableHead>
          <TableHead class="w-16 text-right">操作</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        <TableEmpty v-if="galleries.length === 0" :colspan="7">
          <div class="flex flex-col items-center text-muted-foreground">
            <Images class="mb-2 size-7 opacity-55" />
            <span class="text-xs">暂无图库</span>
          </div>
        </TableEmpty>

        <TableRow v-for="gallery in galleries" :key="gallery.id">
          <TableCell class="font-mono text-xs text-muted-foreground">{{ gallery.id }}</TableCell>
          <TableCell>
            <button
              v-if="galleryCover(gallery)"
              type="button"
              class="block size-[72px] overflow-hidden rounded-md border bg-muted focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
              :aria-label="`预览图库 ${galleryTitle(gallery)} 的封面`"
              @click="emit('preview', galleryCover(gallery), galleryTitle(gallery))"
            >
              <img :src="galleryCover(gallery)" :alt="galleryTitle(gallery)" class="size-full object-cover" />
            </button>
            <div v-else class="flex size-[72px] items-center justify-center rounded-md border bg-muted text-muted-foreground">
              <ImageIcon class="size-5 opacity-50" />
            </div>
          </TableCell>
          <TableCell>
            <button type="button" class="block max-w-72 text-left" @click="emit('view-images', gallery)">
              <span class="block truncate font-bold text-xs hover:text-primary">{{ galleryTitle(gallery) }}</span>
 <span class="mt-1 block truncate font-mono text-xs text-muted-foreground">{{ gallery.slug || '-'}}</span>
            </button>
          </TableCell>
          <TableCell class="max-w-80">
 <p class="line-clamp-2 text-muted-foreground">{{ gallery.description || '-'}}</p>
          </TableCell>
          <TableCell class="text-right tabular-nums">{{ galleryImageCount(gallery) }}</TableCell>
          <TableCell class="text-xs text-muted-foreground">{{ formatDate(gallery.created_at) }}</TableCell>
          <TableCell class="text-right">
            <DropdownMenu>
              <DropdownMenuTrigger as-child>
                <Button variant="ghost" size="icon" :aria-label="`管理图库 ${galleryTitle(gallery)}`">
                  <MoreHorizontal class="size-4" />
                </Button>
              </DropdownMenuTrigger>
              <DropdownMenuContent align="end" class="w-40">
                <DropdownMenuItem @select="emit('view-images', gallery)">
                  <Images class="size-4" />
                  查看图片
                </DropdownMenuItem>
                <DropdownMenuItem v-if="canEdit" @select="emit('edit', gallery)">
                  <Pencil class="size-4" />
                  编辑图库
                </DropdownMenuItem>
                <DropdownMenuSeparator v-if="canDelete" />
                <DropdownMenuItem
                  v-if="canDelete"
                  class="text-destructive focus:text-destructive"
                  @select="emit('delete', gallery)"
                >
                  <Trash2 class="size-4" />
                  删除图库
                </DropdownMenuItem>
              </DropdownMenuContent>
            </DropdownMenu>
          </TableCell>
        </TableRow>
      </TableBody>
    </Table>

    <template #footer>
      <AdminPagination
        :page="pagination.page"
        :page-size="pagination.pageSize"
        :total="pagination.total"
        @update:page="emit('update-page', $event)"
        @update:page-size="emit('update-page-size', $event)"
      />
    </template>
  </AdminTablePanel>
</template>

<script setup lang="ts">
import { Image as ImageIcon, Images, MoreHorizontal, Pencil, Trash2 } from '@lucide/vue'
import AdminPagination from '@/components/admin/AdminPagination.vue'
import AdminTablePanel from '@/components/admin/AdminTablePanel.vue'
import { Button } from '@/components/ui/button'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { Table, TableBody, TableCell, TableEmpty, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import type {
  GalleryCoverResolver,
  GalleryDateFormatter,
  GalleryImageCountResolver,
  GalleryPagination,
  GalleryRecord,
  GalleryTitleResolver
} from '@/modules/gallery/galleryTypes'

withDefaults(defineProps<{
  loading?: boolean
  galleries?: GalleryRecord[]
  pagination: GalleryPagination
  canEdit?: boolean
  canDelete?: boolean
  galleryTitle: GalleryTitleResolver
  galleryCover: GalleryCoverResolver
  galleryImageCount: GalleryImageCountResolver
  formatDate: GalleryDateFormatter
}>(), {
  loading: false,
  galleries: () => [],
  canEdit: false,
  canDelete: false
})

const emit = defineEmits<{
  (event: 'preview', url: string, title: string): void
  (event: 'view-images', gallery: GalleryRecord): void
  (event: 'edit', gallery: GalleryRecord): void
  (event: 'delete', gallery: GalleryRecord): void
  (event: 'update-page', page: number): void
  (event: 'update-page-size', pageSize: number): void
}>()
</script>

