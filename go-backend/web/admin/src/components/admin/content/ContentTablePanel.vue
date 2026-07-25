<template>
  <AdminTablePanel :loading="loading" :batch-visible="selectedPosts.length > 0">
    <template #batch>
      <div class="flex flex-wrap items-center justify-between gap-2">
        <span class="text-xs font-medium">已选择 {{ selectedPosts.length }} 篇文章</span>
        <div class="flex flex-wrap gap-2">
          <Button v-if="canEdit" size="sm" @click="emit('batch-status', 'published')">
            <Send class="size-3.5" />
            批量发布
          </Button>
          <Button v-if="canEdit" variant="outline" size="sm" @click="emit('batch-status', 'draft')">
            <FilePenLine class="size-3.5" />
            转为草稿
          </Button>
          <Button v-if="canDelete" variant="destructive" size="sm" @click="emit('batch-delete')">
            <Trash2 class="size-3.5" />
            批量删除
          </Button>
        </div>
      </div>
    </template>

    <Table class="min-w-[1020px]">
      <TableHeader>
        <TableRow>
          <TableHead class="w-11">
            <Checkbox
              :model-value="selectionState"
              aria-label="选择当前页文章"
              @update:model-value="emit('toggle-all-posts', $event)"
            />
          </TableHead>
          <TableHead class="w-16">ID</TableHead>
          <TableHead>标题</TableHead>
          <TableHead class="w-56">Slug</TableHead>
          <TableHead class="w-24">状态</TableHead>
          <TableHead class="w-20">语言</TableHead>
          <TableHead class="w-24 text-right">浏览量</TableHead>
          <TableHead class="w-44">创建时间</TableHead>
          <TableHead class="w-16 text-right">操作</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        <TableEmpty v-if="posts.length === 0" :colspan="9">
          <div class="flex flex-col items-center text-muted-foreground">
            <Newspaper class="mb-2 size-7 opacity-55" />
            <span class="text-xs">暂无文章</span>
          </div>
        </TableEmpty>

        <TableRow v-for="post in posts" :key="post.id">
          <TableCell>
            <Checkbox
              :model-value="isPostSelected(post.id)"
              :aria-label="`选择文章 ${post.title}`"
              @update:model-value="emit('toggle-post', post, $event)"
            />
          </TableCell>
          <TableCell class="font-mono text-xs text-muted-foreground">{{ post.id }}</TableCell>
          <TableCell class="max-w-80 truncate font-bold text-xs">{{ post.title }}</TableCell>
          <TableCell class="max-w-56 truncate font-mono text-xs text-muted-foreground">{{ post.slug }}</TableCell>
          <TableCell>
            <AdminStatusBadge :tone="statusTone(post.status)">{{ getStatusName(post.status) }}</AdminStatusBadge>
          </TableCell>
          <TableCell>{{ localeName(post.locale) }}</TableCell>
          <TableCell class="text-right tabular-nums">{{ Number(post.view_count || 0).toLocaleString('zh-CN') }}</TableCell>
          <TableCell class="text-xs text-muted-foreground">{{ formatDate(post.created_at) }}</TableCell>
          <TableCell class="text-right">
            <DropdownMenu>
              <DropdownMenuTrigger as-child>
                <Button variant="ghost" size="icon" :aria-label="`管理文章 ${post.title}`">
                  <MoreHorizontal class="size-4" />
                </Button>
              </DropdownMenuTrigger>
              <DropdownMenuContent align="end" class="w-40">
                <DropdownMenuItem v-if="canEdit" @select="emit('edit', post)">
                  <Pencil class="size-4" />
                  编辑
                </DropdownMenuItem>
                <DropdownMenuItem v-if="canEdit" @select="emit('translations', post)">
                  <Languages class="size-4" />
                  翻译版本
                </DropdownMenuItem>
                <DropdownMenuItem v-if="canEdit" @select="emit('toggle-status', post)">
                  <Send v-if="post.status !== 'published'" class="size-4" />
                  <FilePenLine v-else class="size-4" />
                  {{ post.status === 'published' ? '转为草稿' : '发布' }}
                </DropdownMenuItem>
                <DropdownMenuSeparator v-if="canDelete" />
                <DropdownMenuItem
                  v-if="canDelete"
                  class="text-destructive focus:text-destructive"
                  @select="emit('delete', post)"
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

<script setup>
import { FilePenLine, Languages, MoreHorizontal, Newspaper, Pencil, Send, Trash2 } from '@lucide/vue'
import AdminPagination from '@/components/admin/AdminPagination.vue'
import AdminStatusBadge from '@/components/admin/AdminStatusBadge.vue'
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

const props = defineProps({
  loading: { type: Boolean, default: false },
  posts: { type: Array, default: () => [] },
  selectedPosts: { type: Array, default: () => [] },
  pagination: { type: Object, required: true },
  selectionState: { type: [Boolean, String], default: false },
  canEdit: { type: Boolean, default: false },
  canDelete: { type: Boolean, default: false },
  getStatusName: { type: Function, required: true },
  statusTone: { type: Function, required: true },
  localeName: { type: Function, required: true },
  formatDate: { type: Function, required: true },
})

const emit = defineEmits([
  'batch-status',
  'batch-delete',
  'toggle-all-posts',
  'toggle-post',
  'edit',
  'translations',
  'toggle-status',
  'delete',
  'update-page',
  'update-page-size',
])

const isPostSelected = (postId) => props.selectedPosts.some((post) => post.id === postId)
</script>
