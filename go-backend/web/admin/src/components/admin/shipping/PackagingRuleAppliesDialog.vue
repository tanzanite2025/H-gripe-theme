<template>
  <Dialog :open="open" @update:open="emit('update:open', $event)">
    <DialogContent size="lg" class="max-h-[90dvh] overflow-y-auto" @open-auto-focus.prevent>
      <DialogHeader>
        <DialogTitle>维护包装规则适用商品</DialogTitle>
        <DialogDescription>
          {{ rule?.rule_name || '包装规则' }} · 当前为产品级唯一绑定，一个商品只能命中一个包装规则。
        </DialogDescription>
      </DialogHeader>

      <section v-if="rule" class="grid gap-3 md:grid-cols-3">
        <div class="rounded-lg border bg-muted/25 p-3">
          <p class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/70">BOX WEIGHT</p>
          <p class="mt-1 text-xs font-bold">{{ formatWeight(rule.box_weight) }}</p>
        </div>
        <div class="rounded-lg border bg-muted/25 p-3">
          <p class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/70">DIMENSIONS</p>
          <p class="mt-1 font-mono text-xs">{{ formatDimensions(rule) }}</p>
        </div>
        <div class="rounded-lg border bg-muted/25 p-3">
          <p class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/70">APPLIES</p>
          <p class="mt-1 text-2xl font-black tabular-nums">{{ localApplies.length }}</p>
        </div>
      </section>

      <form class="grid gap-3 md:grid-cols-[1fr_auto]" @submit.prevent="addApply">
        <AdminFormField label="商品 ID" required description="当前事实源是 shipping_packaging_rule_applies.product_id；后续若要 SKU 级包装规则，需要升级为 scope 结构。">
          <Input
            v-model.number="productIDInput"
            type="number"
            min="1"
            step="1"
            placeholder="例如 1001"
          />
        </AdminFormField>
        <div class="flex items-end">
          <Button type="submit" class="w-full md:w-auto" :disabled="applySubmitting || !rule?.id">
            <LoaderCircle v-if="applySubmitting" class="size-4 animate-spin" />
            添加适用商品
          </Button>
        </div>
      </form>

      <div class="rounded-lg border bg-muted/35 p-3 text-xs text-muted-foreground">
        稳定口径：包装规则负责箱规与包装重量，SKU 负责商品实重，线路服务负责计费口径；这里不要录入 SKU 重量，也不和图库/商品图片混在一起。
        如果商品需要换包装规则，请先移除旧绑定再添加新绑定。
      </div>

      <div class="rounded-lg border">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead class="w-24">绑定 ID</TableHead>
              <TableHead>商品 ID</TableHead>
              <TableHead class="w-44">创建时间</TableHead>
              <TableHead class="w-24 text-right">操作</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            <TableEmpty v-if="localApplies.length === 0" :colspan="4">
              <div class="py-4 text-xs text-muted-foreground">暂无适用商品，报价时不会通过商品匹配到此包装规则。</div>
            </TableEmpty>
            <TableRow v-for="apply in localApplies" :key="apply.id || apply.product_id">
              <TableCell class="font-mono text-xs">#{{ apply.id || '-' }}</TableCell>
              <TableCell class="font-mono text-xs font-bold">product_id={{ apply.product_id || '-' }}</TableCell>
              <TableCell class="font-mono text-[10px] text-muted-foreground">{{ formatDate(apply.created_at) }}</TableCell>
              <TableCell class="text-right">
                <Button
                  variant="ghost"
                  size="icon-sm"
                  class="text-destructive hover:text-destructive"
                  :disabled="isDeletingApply(apply)"
                  :aria-label="`移除商品 ${apply.product_id} 的包装规则绑定`"
                  @click="deleteApply(apply)"
                >
                  <LoaderCircle v-if="isDeletingApply(apply)" class="size-4 animate-spin" />
                  <Trash2 v-else class="size-4" />
                </Button>
              </TableCell>
            </TableRow>
          </TableBody>
        </Table>
      </div>
    </DialogContent>
  </Dialog>
</template>

<script setup>
import { computed, ref, watch } from 'vue'
import { toast } from 'vue-sonner'
import { LoaderCircle, Trash2 } from '@lucide/vue'
import shippingApi from '@/api/shipping'
import AdminFormField from '@/components/admin/AdminFormField.vue'
import { Button } from '@/components/ui/button'
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'
import { Table, TableBody, TableCell, TableEmpty, TableHead, TableHeader, TableRow } from '@/components/ui/table'

const props = defineProps({
  open: { type: Boolean, default: false },
  rule: { type: Object, default: null },
})

const emit = defineEmits(['update:open', 'updated'])

const productIDInput = ref('')
const localApplies = ref([])
const applySubmitting = ref(false)
const deletingApplyIds = ref(new Set())

const existingProductIDs = computed(() => new Set(localApplies.value.map((apply) => Number(apply.product_id || 0)).filter(Boolean)))

const syncLocalApplies = () => {
  localApplies.value = Array.isArray(props.rule?.applies)
    ? props.rule.applies.map((apply) => ({ ...apply }))
    : []
}

watch(() => props.open, (open) => {
  if (open) {
    productIDInput.value = ''
    syncLocalApplies()
  }
})

watch(() => props.rule, () => {
  if (props.open) syncLocalApplies()
}, { deep: true })

const addApply = async () => {
  const ruleID = Number(props.rule?.id || 0)
  const productID = Number(productIDInput.value || 0)
  if (!ruleID) {
    toast.error('请先选择包装规则')
    return
  }
  if (!Number.isInteger(productID) || productID <= 0) {
    toast.error('请输入有效的商品 ID')
    return
  }
  if (existingProductIDs.value.has(productID)) {
    toast.error('这个商品已绑定当前包装规则')
    return
  }

  applySubmitting.value = true
  try {
    const apply = await shippingApi.createPackagingRuleApply({ rule_id: ruleID, product_id: productID })
    localApplies.value = [...localApplies.value, apply]
    productIDInput.value = ''
    emit('updated')
    toast.success('包装规则适用商品已添加')
  } catch (error) {
    console.error('Failed to create packaging rule apply:', error)
    toast.error(error.response?.data?.error || '添加适用商品失败')
  } finally {
    applySubmitting.value = false
  }
}

const isDeletingApply = (apply) => deletingApplyIds.value.has(Number(apply?.id || 0))

const deleteApply = async (apply) => {
  const applyID = Number(apply?.id || 0)
  if (!applyID || isDeletingApply(apply)) return

  deletingApplyIds.value = new Set(deletingApplyIds.value).add(applyID)
  try {
    await shippingApi.deletePackagingRuleApply(applyID)
    localApplies.value = localApplies.value.filter((item) => Number(item.id || 0) !== applyID)
    emit('updated')
    toast.success('包装规则适用商品已移除')
  } catch (error) {
    console.error('Failed to delete packaging rule apply:', error)
    toast.error(error.response?.data?.error || '移除适用商品失败')
  } finally {
    const next = new Set(deletingApplyIds.value)
    next.delete(applyID)
    deletingApplyIds.value = next
  }
}

const formatDate = (value) => value ? new Date(value).toLocaleString('zh-CN') : '-'
const formatWeight = (value) => `${Number(value || 0).toFixed(3)} kg`
const formatDimensions = (rule) => {
  const length = Number(rule?.box_length || 0).toFixed(2)
  const width = Number(rule?.box_width || 0).toFixed(2)
  const height = Number(rule?.box_height || 0).toFixed(2)
  return `${length} × ${width} × ${height} cm`
}
</script>
