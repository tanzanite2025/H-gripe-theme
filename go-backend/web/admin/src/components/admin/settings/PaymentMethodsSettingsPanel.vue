<template>
  <section class="grid max-w-5xl gap-5 lg:grid-cols-[190px_minmax(0,1fr)]">
    <div>
      <h2 class="text-sm font-black tracking-tighter italic uppercase text-foreground">支付方式</h2>
      <p class="mt-1 text-[9px] font-black uppercase tracking-widest text-muted-foreground/60">
        网关绑定后即可参与前台支付
      </p>
    </div>

    <div class="min-w-0 space-y-3">
      <div class="flex flex-wrap items-center justify-between gap-2">
        <div class="text-xs text-muted-foreground">
          这里只维护前台支付按钮、手续费和启停状态；启用 PayPal，前台就展示 PayPal。
        </div>
        <div class="flex items-center gap-2">
          <Button type="button" variant="outline" size="sm" :disabled="loading" @click="fetchPaymentMethods">
            <RefreshCw :class="['size-3.5', { 'animate-spin': loading }]" />
            刷新
          </Button>
          <Button v-if="canEdit" type="button" size="sm" @click="openCreateDialog">
            <Plus class="size-3.5" />
            添加支付方式
          </Button>
        </div>
      </div>

      <div class="flex gap-2 border-l-2 border-primary/40 px-3 py-2 text-xs text-muted-foreground">
        <Info class="mt-0.5 size-4 shrink-0 text-primary" />
        <p>
          商品价格和订单币种由商品/规格价格决定；汇率 API 只影响前台展示换算。绑定 Stripe、PayPal、微信等收款方式后，可收交易币种以该网关及其商户账户能力为准，客户侧付款币种由网关负责换汇和扣款。
        </p>
      </div>

      <div class="overflow-hidden rounded-lg border">
        <div v-if="loading" class="flex h-32 items-center justify-center text-xs text-muted-foreground">
          <LoaderCircle class="mr-2 size-4 animate-spin" />
          正在加载支付方式
        </div>

        <Table v-else>
          <TableHeader>
            <TableRow>
              <TableHead class="w-[90px]">状态</TableHead>
              <TableHead>支付方式</TableHead>
              <TableHead class="hidden w-[90px] text-right md:table-cell">排序</TableHead>
              <TableHead v-if="canEdit" class="w-[120px] text-right">操作</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            <TableRow v-if="paymentMethods.length === 0">
              <TableCell :colspan="canEdit ? 4 : 3" class="h-28 text-center text-xs text-muted-foreground">
                暂无支付方式
              </TableCell>
            </TableRow>
            <TableRow v-for="method in paymentMethods" :key="method.id">
              <TableCell>
                <Badge :variant="method.enabled ? 'default' : 'secondary'">
                  {{ method.enabled ? '启用' : '停用' }}
                </Badge>
              </TableCell>
              <TableCell>
                <div class="min-w-0">
                  <div class="truncate text-sm font-bold text-foreground">{{ method.name }}</div>
                  <div class="mt-0.5 font-mono text-[11px] text-muted-foreground">{{ method.code }}</div>
                </div>
              </TableCell>
              <TableCell class="hidden text-right font-mono text-xs text-muted-foreground md:table-cell">
                {{ method.sort_order ?? 0 }}
              </TableCell>
              <TableCell v-if="canEdit" class="text-right">
                <div class="flex justify-end gap-1">
                  <Button type="button" variant="ghost" size="icon" aria-label="编辑支付方式" @click="openEditDialog(method)">
                    <Pencil class="size-3.5" />
                  </Button>
                  <Button
                    type="button"
                    variant="ghost"
                    size="icon"
                    aria-label="删除支付方式"
                    :disabled="deletingID === method.id"
                    @click="deletePaymentMethod(method)"
                  >
                    <LoaderCircle v-if="deletingID === method.id" class="size-3.5 animate-spin" />
                    <Trash2 v-else class="size-3.5" />
                  </Button>
                </div>
              </TableCell>
            </TableRow>
          </TableBody>
        </Table>
      </div>
    </div>

    <Dialog v-model:open="dialogOpen">
      <DialogContent size="lg" class="max-h-[92dvh] overflow-y-auto" @open-auto-focus.prevent>
        <form class="space-y-5" @submit.prevent="savePaymentMethod">
          <DialogHeader>
            <DialogTitle>{{ dialogMode === 'create' ? '添加支付方式' : '编辑支付方式' }}</DialogTitle>
            <DialogDescription>维护前台支付按钮、手续费和展示状态。启用后 Nuxt 前台会展示该支付方式；客户侧付款币种由网关处理。</DialogDescription>
          </DialogHeader>

          <div class="grid gap-4 md:grid-cols-2">
            <AdminFormField label="名称" required>
              <Input v-model.trim="form.name" />
            </AdminFormField>

            <AdminFormField label="代码" required>
              <Input v-model.trim="form.code" class="font-mono lowercase" />
            </AdminFormField>

            <AdminFormField label="手续费类型">
              <Select v-model="form.fee_type">
                <SelectTrigger class="w-full"><SelectValue /></SelectTrigger>
                <SelectContent>
                  <SelectItem value="fixed">固定金额</SelectItem>
                  <SelectItem value="percentage">百分比</SelectItem>
                </SelectContent>
              </Select>
            </AdminFormField>

            <AdminFormField label="手续费值">
              <Input v-model.number="form.fee_value" type="number" min="0" step="0.01" />
            </AdminFormField>

            <AdminFormField label="最小金额">
              <Input v-model.number="form.min_amount" type="number" min="0" step="0.01" />
            </AdminFormField>

            <AdminFormField label="最大金额">
              <Input v-model.number="form.max_amount" type="number" min="0" step="0.01" />
            </AdminFormField>

            <AdminFormField label="排序">
              <Input v-model.number="form.sort_order" type="number" step="1" />
            </AdminFormField>

            <div class="flex items-center justify-between gap-3 rounded-lg border px-3 py-2.5">
              <div>
                <span class="text-xs font-medium">启用</span>
                <p class="mt-0.5 text-xs text-muted-foreground">停用后前台不会展示该支付方式。</p>
              </div>
              <Switch v-model="form.enabled" aria-label="支付方式启用状态" />
            </div>

            <AdminFormField label="图标" class="md:col-span-2">
              <Input v-model.trim="form.icon" />
            </AdminFormField>

            <AdminFormField label="描述" class="md:col-span-2">
              <Textarea v-model="form.description" class="min-h-20" />
            </AdminFormField>

            <AdminFormField label="高级设置 JSON" class="md:col-span-2">
              <Textarea v-model="form.settings" class="min-h-20 font-mono" />
            </AdminFormField>
          </div>

          <DialogFooter>
            <Button type="button" variant="outline" @click="dialogOpen = false">取消</Button>
            <Button type="submit" :disabled="submitting">
              <LoaderCircle v-if="submitting" class="size-4 animate-spin" />
              {{ submitting ? '保存中' : '保存支付方式' }}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  </section>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { toast } from 'vue-sonner'
import { Info, LoaderCircle, Pencil, Plus, RefreshCw, Trash2 } from '@lucide/vue'
import AdminFormField from '@/components/admin/AdminFormField.vue'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Switch } from '@/components/ui/switch'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Textarea } from '@/components/ui/textarea'
import axios from '@/utils/axios'
import type { PaymentMethodForm, PaymentMethodRecord } from './settingsTypes'

withDefaults(defineProps<{
  canEdit?: boolean
}>(), {
  canEdit: false,
})

const emptyForm = (): PaymentMethodForm => ({
  id: null,
  name: '',
  code: '',
  icon: '',
  description: '',
  fee_type: 'fixed',
  fee_value: 0,
  min_amount: 0,
  max_amount: 0,
  enabled: true,
  sort_order: 0,
  settings: '',
})

const paymentMethods = ref<PaymentMethodRecord[]>([])
const loading = ref(false)
const submitting = ref(false)
const deletingID = ref<number | string | null>(null)
const dialogOpen = ref(false)
const dialogMode = ref<'create' | 'edit'>('create')
const form = reactive<PaymentMethodForm>(emptyForm())

const unwrapPaymentMethods = (payload: Record<string, any>): PaymentMethodRecord[] => {
  const root = payload?.data
  if (Array.isArray(root)) return root
  if (Array.isArray(root?.data)) return root.data
  if (Array.isArray(root?.data?.data)) return root.data.data
  return []
}

const assignForm = (method: Partial<PaymentMethodRecord | PaymentMethodForm> = emptyForm()): void => {
  Object.assign(form, emptyForm(), {
    ...method,
    fee_type: method.fee_type || 'fixed',
    fee_value: Number(method.fee_value || 0),
    min_amount: Number(method.min_amount || 0),
    max_amount: Number(method.max_amount || 0),
    enabled: method.enabled !== false,
    sort_order: Number(method.sort_order || 0),
    settings: method.settings || '',
  })
}

const fetchPaymentMethods = async () => {
  loading.value = true
  try {
    const response = await axios.get('/api/admin/settings/payment-methods')
    paymentMethods.value = unwrapPaymentMethods(response.data)
  } catch (error) {
    console.error('Failed to fetch payment methods:', error)
    paymentMethods.value = []
  } finally {
    loading.value = false
  }
}

const openCreateDialog = () => {
  dialogMode.value = 'create'
  assignForm()
  dialogOpen.value = true
}

const openEditDialog = (method: PaymentMethodRecord): void => {
  dialogMode.value = 'edit'
  assignForm(method)
  dialogOpen.value = true
}

const buildPayload = (): Omit<PaymentMethodForm, 'id'> => {
  return {
    name: form.name.trim(),
    code: form.code.trim().toLowerCase(),
    icon: form.icon.trim(),
    description: form.description.trim(),
    fee_type: form.fee_type || 'fixed',
    fee_value: Number(form.fee_value || 0),
    min_amount: Number(form.min_amount || 0),
    max_amount: Number(form.max_amount || 0),
    enabled: form.enabled === true,
    sort_order: Number(form.sort_order || 0),
    settings: String(form.settings || '').trim(),
  }
}

const savePaymentMethod = async () => {
  const payload = buildPayload()
  if (!payload) return
  if (!payload.name || !payload.code) {
    toast.error('请填写支付方式名称和代码')
    return
  }

  submitting.value = true
  try {
    if (dialogMode.value === 'edit' && form.id) {
      await axios.put(`/api/admin/settings/payment-methods/${form.id}`, payload)
      toast.success('已更新支付方式')
    } else {
      await axios.post('/api/admin/settings/payment-methods', payload)
      toast.success('已添加支付方式')
    }
    dialogOpen.value = false
    await fetchPaymentMethods()
  } catch (error) {
    console.error('Failed to save payment method:', error)
  } finally {
    submitting.value = false
  }
}

const deletePaymentMethod = async (method: PaymentMethodRecord): Promise<void> => {
  if (!window.confirm(`删除支付方式「${method.name}」？`)) return

  deletingID.value = method.id
  try {
    await axios.delete(`/api/admin/settings/payment-methods/${method.id}`)
    toast.success('已删除支付方式')
    await fetchPaymentMethods()
  } catch (error) {
    console.error('Failed to delete payment method:', error)
  } finally {
    deletingID.value = null
  }
}

onMounted(fetchPaymentMethods)
</script>
