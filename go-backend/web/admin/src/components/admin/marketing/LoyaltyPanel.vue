<template>
  <div class="grid gap-3 xl:grid-cols-[minmax(0,1fr)_360px]">
    <section class="space-y-3">
      <div class="flex flex-wrap items-end justify-between gap-3">
        <AdminFormField label="用户 ID" class="w-48">
          <Input v-model.number="filters.user_id" type="number" min="1" step="1" placeholder="输入用户 ID 查询" @keyup.enter="emit('apply-filter')" />
        </AdminFormField>
        <Button size="sm" @click="emit('apply-filter')">
          <Search class="size-3.5" />
          查询流水
        </Button>
      </div>
      <AdminTablePanel :loading="loading">
        <Table class="min-w-[760px]">
          <TableHeader>
            <TableRow>
              <TableHead class="w-28">类型</TableHead>
              <TableHead class="w-28 text-right">积分</TableHead>
              <TableHead class="w-28 text-right">余额</TableHead>
              <TableHead class="w-32">来源</TableHead>
              <TableHead>说明</TableHead>
              <TableHead class="w-44">时间</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            <TableEmpty v-if="transactions.length === 0" :colspan="6">
              <div class="flex flex-col items-center text-muted-foreground">
                <Coins class="mb-2 size-7 opacity-55" />
                <span class="text-xs">{{ filters.user_id ? '暂无积分流水' : '请输入用户 ID 查询积分流水' }}</span>
              </div>
            </TableEmpty>
            <TableRow v-for="transaction in transactions" :key="transaction.id">
              <TableCell>{{ loyaltyTypeName(transaction.type) }}</TableCell>
              <TableCell class="text-right font-bold tabular-nums" :class="Number(transaction.points) >= 0 ? 'text-emerald-600' : 'text-destructive'">
                {{ Number(transaction.points) > 0 ? '+' : '' }}{{ transaction.points }}
              </TableCell>
              <TableCell class="text-right tabular-nums">{{ transaction.balance }}</TableCell>
              <TableCell class="font-mono text-xs">{{ transaction.source || '-' }} #{{ transaction.source_id || 0 }}</TableCell>
              <TableCell class="max-w-80 truncate text-muted-foreground">{{ transaction.description || '-' }}</TableCell>
              <TableCell class="text-xs text-muted-foreground">{{ formatDate(transaction.created_at) }}</TableCell>
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
    </section>

    <section v-if="canAdjust" class="rounded-xl border bg-card p-4 shadow-sm">
      <div class="mb-4 space-y-1">
        <h3 class="text-sm font-black tracking-tighter italic uppercase">手动调整积分</h3>
        <p class="text-xs text-muted-foreground">所有调整都会写入积分流水，负数表示扣减。</p>
      </div>
      <form class="space-y-4" @submit.prevent="emit('submit')">
        <AdminFormField label="用户 ID" required :error="errors.user_id">
          <Input v-model.number="form.user_id" type="number" min="1" step="1" @input="emit('clear-error', 'user_id')" />
        </AdminFormField>
        <AdminFormField label="调整积分" required :error="errors.points">
          <Input v-model.number="form.points" type="number" step="1" placeholder="例如 100 或 -50" @input="emit('clear-error', 'points')" />
        </AdminFormField>
        <AdminFormField label="调整原因" required :error="errors.description">
          <Textarea v-model="form.description" class="min-h-24" placeholder="必须写清楚原因，方便后续审计" @input="emit('clear-error', 'description')" />
        </AdminFormField>
        <Button class="w-full" type="submit" :disabled="submitting">
          <LoaderCircle v-if="submitting" class="size-4 animate-spin" />
          {{ submitting ? '提交中' : '提交调整' }}
        </Button>
      </form>
    </section>
  </div>
</template>

<script setup>
import { Coins, LoaderCircle, Search } from '@lucide/vue'
import AdminFormField from '@/components/admin/AdminFormField.vue'
import AdminPagination from '@/components/admin/AdminPagination.vue'
import AdminTablePanel from '@/components/admin/AdminTablePanel.vue'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Table, TableBody, TableCell, TableEmpty, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Textarea } from '@/components/ui/textarea'

defineProps({
  loading: { type: Boolean, default: false },
  transactions: { type: Array, default: () => [] },
  filters: { type: Object, required: true },
  pagination: { type: Object, required: true },
  form: { type: Object, required: true },
  errors: { type: Object, required: true },
  submitting: { type: Boolean, default: false },
  canAdjust: { type: Boolean, default: false },
  loyaltyTypeName: { type: Function, required: true },
  formatDate: { type: Function, required: true },
})

const emit = defineEmits(['apply-filter', 'update-page', 'update-page-size', 'submit', 'clear-error'])
</script>
