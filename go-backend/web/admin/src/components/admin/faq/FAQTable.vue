<template>
  <AdminTablePanel :loading="loading" :batch-visible="selectedFaqs.length > 0">
    <template #batch>
      <div class="flex flex-wrap items-center justify-between gap-2">
        <span class="text-xs font-medium">已选择 {{ selectedFaqs.length }} 个 FAQ</span>
        <Button v-if="hasPermission('faq:delete')" variant="destructive" size="sm" @click="$emit('batch-delete')">
          <Trash2 class="size-3.5" />
          批量删除
        </Button>
      </div>
    </template>

    <Table class="min-w-[1180px]">
      <TableHeader>
        <TableRow>
          <TableHead class="w-11">
            <Checkbox
              :model-value="selectionState"
              aria-label="选择当前页 FAQ"
              @update:model-value="$emit('toggle-all', $event)"
            />
          </TableHead>
          <TableHead class="w-16">ID</TableHead>
          <TableHead>问题</TableHead>
          <TableHead>答案</TableHead>
          <TableHead class="w-48">页面</TableHead>
          <TableHead class="w-32">分类</TableHead>
          <TableHead class="w-24">状态</TableHead>
          <TableHead class="w-20">语言</TableHead>
          <TableHead class="w-20 text-right">排序</TableHead>
          <TableHead class="w-44">创建时间</TableHead>
          <TableHead class="w-16 text-right">操作</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        <TableEmpty v-if="faqs.length === 0" :colspan="11">
          <div class="flex flex-col items-center text-muted-foreground">
            <CircleHelp class="mb-2 size-7 opacity-55" />
            <span class="text-xs">暂无 FAQ</span>
          </div>
        </TableEmpty>

        <TableRow v-for="faq in faqs" :key="faq.id">
          <TableCell>
            <Checkbox
              :model-value="isSelected(faq.id)"
              :aria-label="`选择 FAQ ${faq.question}`"
              @update:model-value="$emit('toggle-faq', faq, $event)"
            />
          </TableCell>
          <TableCell class="font-mono text-xs text-muted-foreground">{{ faq.id }}</TableCell>
          <TableCell class="max-w-72">
            <p class="line-clamp-2 font-bold text-xs leading-5">{{ faq.question }}</p>
          </TableCell>
          <TableCell class="max-w-80">
            <p class="line-clamp-2 text-muted-foreground">{{ plainText(faq.answer) }}</p>
            <p v-if="faq.answer_image_url" class="mt-1 text-[11px] font-bold text-sky-600">含 FAQ 图</p>
          </TableCell>
          <TableCell class="max-w-48">
            <p class="truncate text-xs font-bold">{{ pageTitle(faq) }}</p>
            <p class="truncate font-mono text-[11px] text-muted-foreground">{{ faq.page_id || '-' }}</p>
          </TableCell>
          <TableCell>
            <AdminStatusBadge tone="blue">{{ categoryLabel(faq) }}</AdminStatusBadge>
          </TableCell>
          <TableCell>
            <AdminStatusBadge :tone="statusTone(faq.status)">{{ statusName(faq.status) }}</AdminStatusBadge>
          </TableCell>
          <TableCell>{{ localeName(faq.locale) }}</TableCell>
          <TableCell class="text-right tabular-nums">{{ faq.order ?? faq.sort_order ?? 0 }}</TableCell>
          <TableCell class="text-xs text-muted-foreground">{{ formatDate(faq.created_at) }}</TableCell>
          <TableCell class="text-right">
            <DropdownMenu>
              <DropdownMenuTrigger as-child>
                <Button variant="ghost" size="icon" :aria-label="`管理 FAQ ${faq.question}`">
                  <MoreHorizontal class="size-4" />
                </Button>
              </DropdownMenuTrigger>
              <DropdownMenuContent align="end" class="w-36">
                <DropdownMenuItem v-if="hasPermission('faq:edit')" @select="$emit('edit', faq)">
                  <Pencil class="size-4" />
                  编辑
                </DropdownMenuItem>
                <DropdownMenuSeparator v-if="hasPermission('faq:delete')" />
                <DropdownMenuItem
                  v-if="hasPermission('faq:delete')"
                  class="text-destructive focus:text-destructive"
                  @select="$emit('delete', faq)"
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
        @update:page="$emit('update-page', $event)"
        @update:page-size="$emit('update-page-size', $event)"
      />
    </template>
  </AdminTablePanel>
</template>

<script setup>
import { CircleHelp, MoreHorizontal, Pencil, Trash2 } from '@lucide/vue'
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
  DropdownMenuTrigger
} from '@/components/ui/dropdown-menu'
import { Table, TableBody, TableCell, TableEmpty, TableHead, TableHeader, TableRow } from '@/components/ui/table'

defineProps({
  loading: { type: Boolean, default: false },
  faqs: { type: Array, required: true },
  selectedFaqs: { type: Array, required: true },
  selectionState: { type: [Boolean, String], default: false },
  pagination: { type: Object, required: true },
  hasPermission: { type: Function, required: true },
  isSelected: { type: Function, required: true },
  plainText: { type: Function, required: true },
  pageTitle: { type: Function, required: true },
  categoryLabel: { type: Function, required: true },
  statusTone: { type: Function, required: true },
  statusName: { type: Function, required: true },
  localeName: { type: Function, required: true },
  formatDate: { type: Function, required: true }
})

defineEmits([
  'toggle-all',
  'toggle-faq',
  'edit',
  'delete',
  'batch-delete',
  'update-page',
  'update-page-size'
])
</script>
