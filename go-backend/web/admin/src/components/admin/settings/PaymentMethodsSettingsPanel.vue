<template>
  <div class="flex h-full min-h-0 flex-col gap-4 overflow-hidden">
    <AdminPageHeader
      class="shrink-0"
      title="收款方式"
      description="管理前台可用的收款方式、费用规则和展示顺序。"
    />

    <div class="min-h-0 flex-1 overflow-auto">
      <section class="max-w-none space-y-3">
        <div class="flex flex-wrap items-center justify-between gap-2">
          <div class="text-xs text-muted-foreground">
            {{ t('payment.paymentMethodsHelp') }}
          </div>
          <div class="flex items-center gap-2">
            <Button type="button" variant="outline" size="sm" :disabled="loading" @click="fetchPaymentMethods">
              <RefreshCw :class="['size-3.5', { 'animate-spin': loading }]" />
              {{ t('common.refresh') }}
            </Button>
            <Button v-if="canEdit" type="button" size="sm" @click="openCreateDialog">
              <Plus class="size-3.5" />
              {{ t('payment.addPaymentMethod') }}
            </Button>
          </div>
        </div>

        <div class="flex gap-2 border-l-2 border-primary/40 px-3 py-2 text-xs text-muted-foreground">
          <Info class="mt-0.5 size-4 shrink-0 text-primary" />
          <p>
            {{ t('payment.paymentMethodCurrencyHelp') }}
          </p>
        </div>

        <div class="overflow-hidden rounded-lg border">
          <div v-if="loading" class="flex h-32 items-center justify-center text-xs text-muted-foreground">
            <LoaderCircle class="mr-2 size-4 animate-spin" />
            {{ t('payment.loadingPaymentMethods') }}
          </div>

          <Table v-else>
            <TableHeader>
              <TableRow>
                <TableHead class="w-[90px]">{{ t('payment.status') }}</TableHead>
                <TableHead>{{ t('settings.paymentMethods') }}</TableHead>
                <TableHead class="hidden w-[90px] text-right md:table-cell">{{ t('payment.sortOrder') }}</TableHead>
                <TableHead v-if="canEdit" class="w-[120px] text-right">{{ t('payment.actions') }}</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              <TableRow v-if="paymentMethods.length === 0">
                <TableCell :colspan="canEdit ? 4 : 3" class="h-28 text-center text-xs text-muted-foreground">
                  {{ t('payment.noPaymentMethods') }}
                </TableCell>
              </TableRow>
              <TableRow v-for="method in paymentMethods" :key="method.id">
                <TableCell>
                  <Badge :variant="method.enabled ? 'default' : 'secondary'">
                    {{ method.enabled ? t('common.enabled') : t('common.disabled') }}
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
                    <Button type="button" variant="ghost" size="icon" :aria-label="t('payment.editPaymentMethod')" @click="openEditDialog(method)">
                      <Pencil class="size-3.5" />
                    </Button>
                    <Button
                      type="button"
                      variant="ghost"
                      size="icon"
                      :aria-label="t('payment.deletePaymentMethod')"
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
      </section>
    </div>

    <Dialog v-model:open="dialogOpen">
      <DialogContent size="lg" class="max-h-[92dvh] overflow-y-auto" @open-auto-focus.prevent>
        <form class="space-y-5" @submit.prevent="savePaymentMethod">
          <DialogHeader>
            <DialogTitle>{{ dialogMode === 'create' ? t('payment.addPaymentMethod') : t('payment.editPaymentMethod') }}</DialogTitle>
            <DialogDescription>{{ t('payment.paymentMethodDialogDescription') }}</DialogDescription>
          </DialogHeader>

          <div class="grid gap-4 md:grid-cols-2">
            <AdminFormField :label="t('payment.name')" required>
              <Input v-model.trim="form.name" />
            </AdminFormField>

            <AdminFormField :label="t('payment.code')" required>
              <Input v-model.trim="form.code" class="font-mono lowercase" />
            </AdminFormField>

            <AdminFormField :label="t('payment.feeType')">
              <Select v-model="form.fee_type">
                <SelectTrigger class="w-full"><SelectValue /></SelectTrigger>
                <SelectContent>
                  <SelectItem value="fixed">{{ t('payment.fixedAmount') }}</SelectItem>
                  <SelectItem value="percentage">{{ t('payment.percentage') }}</SelectItem>
                </SelectContent>
              </Select>
            </AdminFormField>

            <AdminFormField :label="t('payment.feeValue')">
              <Input v-model.number="form.fee_value" type="number" min="0" step="0.01" />
            </AdminFormField>

            <AdminFormField :label="t('payment.minimumAmount')">
              <Input v-model.number="form.min_amount" type="number" min="0" step="0.01" />
            </AdminFormField>

            <AdminFormField :label="t('payment.maximumAmount')">
              <Input v-model.number="form.max_amount" type="number" min="0" step="0.01" />
            </AdminFormField>

            <AdminFormField :label="t('payment.sortOrder')">
              <Input v-model.number="form.sort_order" type="number" step="1" />
            </AdminFormField>

            <div class="flex items-center justify-between gap-3 rounded-lg border px-3 py-2.5">
              <div>
                <span class="text-xs font-medium">{{ t('common.enabled') }}</span>
                <p class="mt-0.5 text-xs text-muted-foreground">{{ t('payment.disabledMethodHelp') }}</p>
              </div>
              <Switch v-model="form.enabled" :aria-label="t('payment.paymentMethodEnabled')" />
            </div>

            <AdminFormField :label="t('payment.icon')" class="md:col-span-2">
              <Input v-model.trim="form.icon" />
            </AdminFormField>

            <AdminFormField :label="t('payment.description')" class="md:col-span-2">
              <Textarea v-model="form.description" class="min-h-20" />
            </AdminFormField>

            <AdminFormField :label="t('payment.advancedSettings')" class="md:col-span-2">
              <Textarea v-model="form.settings" class="min-h-20 font-mono" />
            </AdminFormField>
          </div>

          <DialogFooter>
            <Button type="button" variant="outline" @click="dialogOpen = false">{{ t('common.cancel') }}</Button>
            <Button type="submit" :disabled="submitting">
              <LoaderCircle v-if="submitting" class="size-4 animate-spin" />
              {{ submitting ? t('common.saving') : t('payment.savePaymentMethod') }}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { toast } from 'vue-sonner'
import { Info, LoaderCircle, Pencil, Plus, RefreshCw, Trash2 } from '@lucide/vue'
import AdminFormField from '@/components/admin/AdminFormField.vue'
import AdminPageHeader from '@/components/admin/AdminPageHeader.vue'
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
import { useAdminI18n } from '@/i18n'
import { useAuthStore } from '@/stores/auth'
import type { PaymentMethodForm, PaymentMethodRecord } from '@/modules/settings/types'

const props = defineProps<{
  canEdit?: boolean
}>()
const authStore = useAuthStore()
const canEdit = computed(() => props.canEdit ?? authStore.hasPermission('settings:edit'))
const { t } = useAdminI18n()

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
    toast.error(t('payment.paymentMethodRequired'))
    return
  }

  submitting.value = true
  try {
    if (dialogMode.value === 'edit' && form.id) {
      await axios.put(`/api/admin/settings/payment-methods/${form.id}`, payload)
      toast.success(t('payment.paymentMethodUpdated'))
    } else {
      await axios.post('/api/admin/settings/payment-methods', payload)
      toast.success(t('payment.paymentMethodAdded'))
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
  if (!window.confirm(t('payment.confirmDeletePaymentMethod', { name: method.name }))) return

  deletingID.value = method.id
  try {
    await axios.delete(`/api/admin/settings/payment-methods/${method.id}`)
    toast.success(t('payment.paymentMethodDeleted'))
    await fetchPaymentMethods()
  } catch (error) {
    console.error('Failed to delete payment method:', error)
  } finally {
    deletingID.value = null
  }
}

onMounted(fetchPaymentMethods)
</script>
