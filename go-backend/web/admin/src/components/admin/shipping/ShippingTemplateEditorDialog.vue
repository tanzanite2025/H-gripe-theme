<template>
  <Dialog :open="open" @update:open="emit('update:open', $event)">
    <DialogContent size="full" class="max-h-[90dvh] overflow-y-auto" @open-auto-focus.prevent>
      <form class="space-y-6" @submit.prevent="emit('submit')">
        <DialogHeader>
          <DialogTitle>{{ mode === 'create' ? '新增运费模板' : '编辑运费模板' }}</DialogTitle>
          <DialogDescription>
            先按当前后端模型维护区域与重量/数量/金额规则；后续会扩展到承运商线路、SKU 绑定和体积重。
          </DialogDescription>
        </DialogHeader>

        <section class="grid gap-4 lg:grid-cols-4">
          <AdminFormField label="模板名称" required :error="errors.name" class="lg:col-span-2">
            <Input v-model.trim="form.name" placeholder="例如 全球空运标准模板" @input="emit('clear-error', 'name')" />
          </AdminFormField>

          <AdminFormField label="计费类型" required :error="errors.type">
            <Select v-model="form.type" @update:model-value="emit('clear-error', 'type')">
              <SelectTrigger class="w-full"><SelectValue placeholder="请选择计费类型" /></SelectTrigger>
              <SelectContent>
                <SelectItem value="weight">按重量</SelectItem>
                <SelectItem value="quantity">按数量</SelectItem>
                <SelectItem value="price">按订单金额</SelectItem>
              </SelectContent>
            </Select>
          </AdminFormField>

          <div class="flex items-center justify-between gap-3 rounded-lg border px-3 py-2.5">
            <div>
              <span class="text-xs font-bold uppercase tracking-wider">启用模板 / ENABLED</span>
              <p class="mt-0.5 text-xs text-muted-foreground">停用后不会参与运费规则计算。</p>
            </div>
            <Switch v-model="form.enabled" aria-label="启用运费模板" />
          </div>

          <AdminFormField label="默认运费" required :error="errors.default_fee">
            <Input v-model.number="form.default_fee" type="number" min="0" step="0.01" @input="emit('clear-error', 'default_fee')" />
          </AdminFormField>

          <AdminFormField label="免运门槛">
            <Input v-model.number="form.free_threshold" type="number" min="0" step="0.01" />
          </AdminFormField>

          <div class="flex items-end justify-between gap-3 rounded-lg border px-3 py-2.5">
            <div>
              <span class="text-xs font-bold uppercase tracking-wider">开启免运 / FREE SHIPPING</span>
              <p class="mt-0.5 text-xs text-muted-foreground">订单金额达到门槛时返回 0 运费。</p>
            </div>
            <Switch v-model="form.free_shipping" aria-label="开启免运" />
          </div>

          <AdminFormField label="说明" class="lg:col-span-4">
            <Textarea v-model="form.description" class="min-h-20" placeholder="内部说明、适用渠道或注意事项" />
          </AdminFormField>
        </section>

        <section class="space-y-3 border-t border-dashed pt-5">
          <div class="flex flex-wrap items-center justify-between gap-3">
            <div>
              <h3 class="text-sm font-black tracking-tighter italic uppercase text-foreground">规则矩阵</h3>
              <p class="mt-1 text-[9px] font-black uppercase tracking-widest text-muted-foreground/60">Region 建议填国家/区域代码，例如 US、EU、CN；最大值为 0 表示不设上限。</p>
            </div>
            <Button type="button" variant="outline" size="sm" @click="addRule">
              <Plus class="size-3.5" />
              新增规则
            </Button>
          </div>

          <div v-if="!form.rules?.length" class="rounded-lg border border-dashed p-6 text-center text-sm text-muted-foreground">
            暂无规则；没有匹配规则时会使用模板默认运费。
          </div>

          <div v-for="(rule, index) in form.rules" :key="index" class="grid gap-3 rounded-lg border p-3 lg:grid-cols-12">
            <AdminFormField label="Region" class="lg:col-span-2">
              <Input v-model.trim="rule.region" class="font-mono uppercase" placeholder="US" />
            </AdminFormField>
            <AdminFormField label="最小值" class="lg:col-span-2">
              <Input v-model.number="rule.min_value" type="number" min="0" step="0.001" />
            </AdminFormField>
            <AdminFormField label="最大值" class="lg:col-span-2">
              <Input v-model.number="rule.max_value" type="number" min="0" step="0.001" />
            </AdminFormField>
            <AdminFormField label="运费" class="lg:col-span-2">
              <Input v-model.number="rule.fee" type="number" min="0" step="0.01" />
            </AdminFormField>
            <AdminFormField label="续费" class="lg:col-span-2">
              <Input v-model.number="rule.additional" type="number" min="0" step="0.01" />
            </AdminFormField>
            <div class="flex items-end justify-end lg:col-span-2">
              <Button type="button" variant="ghost" size="icon-sm" class="text-destructive hover:text-destructive" @click="removeRule(index)">
                <Trash2 class="size-4" />
                <span class="sr-only">删除规则</span>
              </Button>
            </div>
          </div>
        </section>

        <DialogFooter>
          <Button type="button" variant="outline" @click="emit('update:open', false)">取消</Button>
          <Button type="submit" :disabled="submitting">
            <LoaderCircle v-if="submitting" class="size-4 animate-spin" />
            {{ submitting ? '保存中' : '保存模板' }}
          </Button>
        </DialogFooter>
      </form>
    </DialogContent>
  </Dialog>
</template>

<script setup>
import { LoaderCircle, Plus, Trash2 } from '@lucide/vue'
import AdminFormField from '@/components/admin/AdminFormField.vue'
import { Button } from '@/components/ui/button'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Switch } from '@/components/ui/switch'
import { Textarea } from '@/components/ui/textarea'

const props = defineProps({
  open: { type: Boolean, default: false },
  mode: { type: String, default: 'create' },
  form: { type: Object, required: true },
  errors: { type: Object, required: true },
  submitting: { type: Boolean, default: false },
})

const emit = defineEmits(['update:open', 'submit', 'clear-error'])

const ensureRules = () => {
  if (!Array.isArray(props.form.rules)) {
    props.form.rules = []
  }
}

const addRule = () => {
  ensureRules()
  props.form.rules.push({
    region: '',
    min_value: 0,
    max_value: 0,
    fee: 0,
    additional: 0,
  })
}

const removeRule = (index) => {
  ensureRules()
  props.form.rules.splice(index, 1)
}
</script>
