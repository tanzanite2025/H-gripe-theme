<template>
  <AdminFormSection title="发布设置" :description="description">
    <div class="grid gap-4 md:grid-cols-2">
      <div v-if="notice" class="md:col-span-2 rounded-lg border bg-muted/30 px-3 py-2.5 text-xs text-muted-foreground">
        {{ notice }}
      </div>
      <AdminFormField label="状态" required>
        <Select v-model="form.status">
          <SelectTrigger class="w-full"><SelectValue /></SelectTrigger>
          <SelectContent>
            <SelectItem value="active">在售</SelectItem>
            <SelectItem value="inactive">下架</SelectItem>
            <SelectItem value="out_of_stock">缺货</SelectItem>
          </SelectContent>
        </Select>
      </AdminFormField>
      <AdminFormField label="运费模板">
        <Select :model-value="shippingTemplateSelectValue" @update:model-value="emit('product-shipping-template-select', $event)">
          <SelectTrigger class="w-full"><SelectValue placeholder="未设置" /></SelectTrigger>
          <SelectContent>
            <SelectItem value="__none__">未设置</SelectItem>
            <SelectItem v-for="template in shippingTemplates" :key="template.id" :value="String(template.id)">
              {{ template.name }}{{ template.enabled === false ? '（停用）' : '' }}
            </SelectItem>
          </SelectContent>
        </Select>
      </AdminFormField>
      <AdminFormField label="After-sales 模板">
        <Select :model-value="afterSalesTemplateSelectValue" @update:model-value="emit('product-information-template-select', 'after_sales_template_id', $event)">
          <SelectTrigger class="w-full"><SelectValue placeholder="未设置" /></SelectTrigger>
          <SelectContent>
            <SelectItem value="__none__">未设置</SelectItem>
            <SelectItem v-for="template in afterSalesTemplates" :key="template.id" :value="String(template.id)">
              {{ template.name }}{{ template.locale ? `（${template.locale}）` : '' }}{{ template.is_enabled === false ? '（停用）' : '' }}
            </SelectItem>
          </SelectContent>
        </Select>
      </AdminFormField>
      <AdminFormField label="Packaging 模板">
        <Select :model-value="packagingTemplateSelectValue" @update:model-value="emit('product-information-template-select', 'packaging_template_id', $event)">
          <SelectTrigger class="w-full"><SelectValue placeholder="未设置" /></SelectTrigger>
          <SelectContent>
            <SelectItem value="__none__">未设置</SelectItem>
            <SelectItem v-for="template in packagingTemplates" :key="template.id" :value="String(template.id)">
              {{ template.name }}{{ template.locale ? `（${template.locale}）` : '' }}{{ template.is_enabled === false ? '（停用）' : '' }}
            </SelectItem>
          </SelectContent>
        </Select>
      </AdminFormField>
      <div class="flex items-center justify-between gap-4 rounded-lg border px-3 py-2.5 md:col-span-2">
        <div>
          <Label :for="`${idPrefix}-featured`">精选商品</Label>
          <p class="mt-0.5 text-xs text-muted-foreground">在前台精选区域优先展示该商品。</p>
        </div>
        <Switch :id="`${idPrefix}-featured`" v-model="form.featured" />
      </div>
    </div>
  </AdminFormSection>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import AdminFormField from '@/components/admin/AdminFormField.vue'
import AdminFormSection from '@/components/admin/AdminFormSection.vue'
import { Label } from '@/components/ui/label'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Switch } from '@/components/ui/switch'

interface TemplateRecord {
  id: number | string
  name: string
  locale?: string
  enabled?: boolean
  is_enabled?: boolean
}

const props = withDefaults(defineProps<{
  form: Record<string, any>
  shippingTemplates?: TemplateRecord[]
  afterSalesTemplates?: TemplateRecord[]
  packagingTemplates?: TemplateRecord[]
  idPrefix?: string
  description?: string
  notice?: string
}>(), {
  shippingTemplates: () => [],
  afterSalesTemplates: () => [],
  packagingTemplates: () => [],
  idPrefix: 'product',
  description: '控制商品的公开状态和前台可见性。',
  notice: '',
})

const emit = defineEmits<{
  (event: 'product-shipping-template-select', value: unknown): void
  (event: 'product-information-template-select', field: 'after_sales_template_id' | 'packaging_template_id', value: unknown): void
}>()

const shippingTemplateSelectValue = computed(() => props.form.shipping_template_id ? String(props.form.shipping_template_id) : '__none__')
const afterSalesTemplateSelectValue = computed(() => props.form.after_sales_template_id ? String(props.form.after_sales_template_id) : '__none__')
const packagingTemplateSelectValue = computed(() => props.form.packaging_template_id ? String(props.form.packaging_template_id) : '__none__')
</script>
