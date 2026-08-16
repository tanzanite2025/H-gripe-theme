<template>
  <AdminTablePanel :loading="loading">
    <Table class="min-w-[920px]">
      <TableHeader>
        <TableRow>
          <TableHead class="w-16">ID</TableHead>
          <TableHead>名称</TableHead>
          <TableHead>标识</TableHead>
          <TableHead class="w-28 text-right">字段模板</TableHead>
          <TableHead class="w-28 text-right">变体字段</TableHead>
          <TableHead class="w-24">状态</TableHead>
          <TableHead class="w-20 text-right">排序</TableHead>
          <TableHead class="w-44">更新时间</TableHead>
          <TableHead class="w-16 text-right">操作</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        <TableEmpty v-if="templates.length === 0" :colspan="9">
          <div class="flex flex-col items-center text-muted-foreground">
            <Tags class="mb-2 size-7 opacity-55" />
            <span class="text-xs">暂无商品规格模板</span>
          </div>
        </TableEmpty>

        <TableRow v-for="template in templates" :key="template.id">
          <TableCell class="font-mono text-[10px] font-bold text-muted-foreground">{{ template.id }}</TableCell>
          <TableCell>
            <div class="max-w-80">
              <p class="truncate font-bold text-xs">{{ template.name }}</p>
              <p v-if="template.description" class="mt-0.5 truncate text-[10px] text-muted-foreground/70">{{ template.description }}</p>
            </div>
          </TableCell>
          <TableCell class="font-mono text-[11px] font-bold text-muted-foreground/80">{{ template.slug }}</TableCell>
          <TableCell class="text-right font-mono text-xs font-bold tabular-nums">{{ template.spec_definitions?.length || 0 }}</TableCell>
          <TableCell class="text-right font-mono text-xs font-bold tabular-nums">{{ variantSpecCount(template) }}</TableCell>
          <TableCell>
            <AdminStatusBadge :tone="template.is_enabled ? 'green' : 'gray'">
              {{ template.is_enabled ? '已启用' : '已停用' }}
            </AdminStatusBadge>
          </TableCell>
          <TableCell class="text-right font-mono text-xs font-bold tabular-nums">{{ template.sort_order || 0 }}</TableCell>
          <TableCell class="font-mono text-[10px] text-muted-foreground/80">{{ formatDate(template.updated_at) }}</TableCell>
          <TableCell class="text-right">
            <DropdownMenu>
              <DropdownMenuTrigger as-child>
                <Button variant="ghost" size="icon" :aria-label="`管理商品规格模板 ${template.name}`">
                  <MoreHorizontal class="size-4" />
                </Button>
              </DropdownMenuTrigger>
              <DropdownMenuContent align="end" class="w-40">
                <DropdownMenuItem v-if="canEdit" @select="emit('edit', template)">
                  <Pencil class="size-4" />
                  编辑
                </DropdownMenuItem>
                <DropdownMenuItem v-if="canEdit" @select="emit('toggle', template)">
                  <CircleCheck v-if="!template.is_enabled" class="size-4" />
                  <CircleOff v-else class="size-4" />
                  {{ template.is_enabled ? '停用' : '启用' }}
                </DropdownMenuItem>
                <DropdownMenuSeparator v-if="canDelete" />
                <DropdownMenuItem
                  v-if="canDelete"
                  class="text-destructive focus:text-destructive"
                  @select="emit('delete', template)"
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
import { CircleCheck, CircleOff, MoreHorizontal, Pencil, Tags, Trash2 } from '@lucide/vue'
import AdminStatusBadge from '@/components/admin/AdminStatusBadge.vue'
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
  ProductSpecTemplateDateFormatter,
  ProductSpecTemplateRecord,
  ProductSpecTemplateVariantSpecCounter
} from './productSpecificationTemplateTypes'

withDefaults(defineProps<{
  loading?: boolean
  templates?: ProductSpecTemplateRecord[]
  canEdit?: boolean
  canDelete?: boolean
  variantSpecCount: ProductSpecTemplateVariantSpecCounter
  formatDate: ProductSpecTemplateDateFormatter
}>(), {
  loading: false,
  templates: () => [],
  canEdit: false,
  canDelete: false
})

const emit = defineEmits<{
  (event: 'edit', template: ProductSpecTemplateRecord): void
  (event: 'toggle', template: ProductSpecTemplateRecord): void
  (event: 'delete', template: ProductSpecTemplateRecord): void
}>()
</script>
