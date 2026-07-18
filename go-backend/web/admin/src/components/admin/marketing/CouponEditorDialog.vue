<template>
  <Dialog :open="open" @update:open="emit('update:open', $event)">
    <DialogContent class="max-h-[92vh] overflow-y-auto sm:max-w-3xl" @open-auto-focus.prevent>
      <form class="space-y-5" @submit.prevent="emit('submit')">
        <DialogHeader>
          <DialogTitle>{{ mode === 'create' ? '创建优惠券' : '编辑优惠券' }}</DialogTitle>
          <DialogDescription>设置折扣规则、适用范围、使用次数和有效时间。</DialogDescription>
        </DialogHeader>

        <div class="grid gap-4 sm:grid-cols-2">
          <AdminFormField label="优惠码" required :error="errors.code">
            <Input v-model="form.code" class="font-mono uppercase" placeholder="例如 SUMMER20" @input="emit('clear-error', 'code')" />
          </AdminFormField>
          <AdminFormField label="折扣类型" required>
            <RadioGroup v-model="form.type" class="grid grid-cols-2 gap-2">
              <label class="flex h-9 cursor-pointer items-center gap-2 rounded-md border px-3 has-data-[state=checked]:border-primary has-data-[state=checked]:bg-accent">
                <RadioGroupItem value="fixed" />
                <span class="text-sm">固定金额</span>
              </label>
              <label class="flex h-9 cursor-pointer items-center gap-2 rounded-md border px-3 has-data-[state=checked]:border-primary has-data-[state=checked]:bg-accent">
                <RadioGroupItem value="percentage" />
                <span class="text-sm">百分比</span>
              </label>
            </RadioGroup>
          </AdminFormField>
          <AdminFormField label="折扣值" required :error="errors.value">
            <div class="relative">
              <Input v-model.number="form.value" type="number" min="0" step="0.01" class="pr-10" @input="emit('clear-error', 'value')" />
              <span class="pointer-events-none absolute right-3 top-1/2 -translate-y-1/2 text-xs text-muted-foreground">
                {{ form.type === 'percentage' ? '%' : '元' }}
              </span>
            </div>
          </AdminFormField>
          <AdminFormField label="最低消费">
            <Input v-model.number="form.min_amount" type="number" min="0" step="0.01" />
          </AdminFormField>
          <AdminFormField label="最大折扣" description="0 表示不限制">
            <Input v-model.number="form.max_discount" type="number" min="0" step="0.01" />
          </AdminFormField>
          <AdminFormField label="总使用次数" description="0 表示不限制">
            <Input v-model.number="form.usage_limit" type="number" min="0" step="1" />
          </AdminFormField>
          <AdminFormField label="单用户使用次数" description="0 表示不限制">
            <Input v-model.number="form.usage_limit_per_user" type="number" min="0" step="1" />
          </AdminFormField>
          <div class="flex items-center justify-between gap-3 rounded-lg border px-3 py-2.5">
            <div>
              <span class="text-xs font-medium">启用优惠券</span>
              <p class="mt-0.5 text-xs text-muted-foreground">停用后无法在结账时使用。</p>
            </div>
            <Switch v-model="form.enabled" aria-label="启用优惠券" />
          </div>
          <AdminFormField label="开始时间" required :error="errors.start_date">
            <Input v-model="form.start_date" type="datetime-local" @input="emit('clear-error', 'start_date')" />
          </AdminFormField>
          <AdminFormField label="结束时间" required :error="errors.end_date">
            <Input v-model="form.end_date" type="datetime-local" @input="emit('clear-error', 'end_date')" />
          </AdminFormField>
          <AdminFormField label="描述" class="sm:col-span-2">
            <Textarea v-model="form.description" class="min-h-20" />
          </AdminFormField>
        </div>

        <div class="space-y-4 border-t pt-5">
          <h3 class="text-sm font-semibold">适用范围</h3>
          <div class="grid gap-4 sm:grid-cols-2">
            <AdminFormField label="适用商品" description="JSON 数组；留空表示全部商品">
              <Textarea v-model="form.applicable_products" class="min-h-20 font-mono text-xs" placeholder='[1, 2, 3]' />
            </AdminFormField>
            <AdminFormField label="排除商品" description="JSON 数组">
              <Textarea v-model="form.excluded_products" class="min-h-20 font-mono text-xs" placeholder='[4, 5]' />
            </AdminFormField>
            <AdminFormField label="适用分类" description="JSON 数组；留空表示全部分类" class="sm:col-span-2">
              <Textarea v-model="form.applicable_categories" class="min-h-20 font-mono text-xs" placeholder='["bracelets", "rings"]' />
            </AdminFormField>
          </div>
        </div>

        <DialogFooter>
          <Button type="button" variant="outline" @click="emit('update:open', false)">取消</Button>
          <Button type="submit" :disabled="submitting">
            <LoaderCircle v-if="submitting" class="size-4 animate-spin" />
            {{ submitting ? '保存中' : '保存优惠券' }}
          </Button>
        </DialogFooter>
      </form>
    </DialogContent>
  </Dialog>
</template>

<script setup>
import { LoaderCircle } from '@lucide/vue'
import AdminFormField from '@/components/admin/AdminFormField.vue'
import { Button } from '@/components/ui/button'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'
import { RadioGroup, RadioGroupItem } from '@/components/ui/radio-group'
import { Switch } from '@/components/ui/switch'
import { Textarea } from '@/components/ui/textarea'

defineProps({
  open: { type: Boolean, default: false },
  mode: { type: String, default: 'create' },
  form: { type: Object, required: true },
  errors: { type: Object, required: true },
  submitting: { type: Boolean, default: false }
})
const emit = defineEmits(['update:open', 'submit', 'clear-error'])
</script>
