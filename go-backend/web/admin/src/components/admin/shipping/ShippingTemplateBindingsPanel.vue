<template>
  <TabsContent value="bindings" class="space-y-3">
    <div class="flex flex-wrap items-end justify-between gap-3">
      <div>
        <h2 class="text-sm font-black tracking-tighter italic uppercase">产品/SKU 绑定</h2>
        <p class="mt-1 text-xs text-muted-foreground">设置默认、产品类型、产品或 SKU 应使用的运费模板。</p>
      </div>
      <Button v-if="canCreate" size="sm" :disabled="templates.length === 0" @click="emit('create')">
        <Plus class="size-3.5" />
        新增绑定
      </Button>
    </div>

    <AdminTablePanel :loading="loading">
      <Table class="min-w-[980px]">
        <TableHeader>
          <TableRow>
            <TableHead class="w-32">范围</TableHead>
            <TableHead>目标</TableHead>
            <TableHead>运费模板</TableHead>
            <TableHead class="w-24 text-right">优先级</TableHead>
            <TableHead class="w-24">状态</TableHead>
            <TableHead class="w-32 text-right">操作</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          <TableEmpty v-if="templateBindings.length === 0" :colspan="6">
            <div class="flex flex-col items-center text-muted-foreground">
              <Link2 class="mb-2 size-7 opacity-55" />
              <span class="text-xs">暂无模板绑定</span>
            </div>
          </TableEmpty>
          <TableRow v-for="binding in templateBindings" :key="binding.id">
            <TableCell>{{ bindingScopeLabel(binding.scope) }}</TableCell>
            <TableCell class="font-mono text-xs text-muted-foreground">{{ bindingTargetLabel(binding) }}</TableCell>
            <TableCell>
              <span class="block font-bold text-xs">{{ bindingTemplateName(binding, templates) }}</span>
              <span class="block text-xs text-muted-foreground">ID: {{ binding.template_id }}</span>
            </TableCell>
            <TableCell class="text-right tabular-nums">{{ binding.priority || 0 }}</TableCell>
            <TableCell>
              <AdminStatusBadge :tone="binding.enabled ? 'green' : 'gray'">
                {{ binding.enabled ? '启用' : '停用' }}
              </AdminStatusBadge>
            </TableCell>
            <TableCell class="text-right">
              <div class="inline-flex items-center gap-1">
                <Button
                  v-if="canEdit"
                  variant="ghost"
                  size="icon-sm"
                  :aria-label="`编辑模板绑定 ${binding.id}`"
                  @click="emit('edit', binding)"
                >
                  <Pencil class="size-4" />
                </Button>
                <Button
                  v-if="canDelete"
                  variant="ghost"
                  size="icon-sm"
                  class="text-destructive hover:text-destructive"
                  :aria-label="`删除模板绑定 ${binding.id}`"
                  @click="emit('delete', binding)"
                >
                  <Trash2 class="size-4" />
                </Button>
              </div>
            </TableCell>
          </TableRow>
        </TableBody>
      </Table>
    </AdminTablePanel>
  </TabsContent>
</template>

<script setup>
import { Link2, Pencil, Plus, Trash2 } from '@lucide/vue'
import AdminStatusBadge from '@/components/admin/AdminStatusBadge.vue'
import AdminTablePanel from '@/components/admin/AdminTablePanel.vue'
import { Button } from '@/components/ui/button'
import { Table, TableBody, TableCell, TableEmpty, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { TabsContent } from '@/components/ui/tabs'
import { bindingScopeLabel, bindingTargetLabel, bindingTemplateName } from '@/lib/shippingPresentation'

defineProps({
  templateBindings: { type: Array, default: () => [] },
  templates: { type: Array, default: () => [] },
  loading: { type: Boolean, default: false },
  canCreate: { type: Boolean, default: false },
  canEdit: { type: Boolean, default: false },
  canDelete: { type: Boolean, default: false },
})

const emit = defineEmits(['create', 'edit', 'delete'])
</script>
