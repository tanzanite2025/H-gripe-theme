<template>
  <Dialog :open="open" @update:open="emit('update:open', $event)">
    <DialogContent class="max-w-3xl">
      <DialogHeader>
        <DialogTitle>{{ form.id ? '编辑清关资料模板' : '新建清关资料模板' }}</DialogTitle>
        <DialogDescription>模板保存后，可在商品编辑器中一键填入四项清关资料。</DialogDescription>
      </DialogHeader>
      <form class="space-y-4" @submit.prevent="emit('save')">
        <div class="grid gap-4 sm:grid-cols-2">
          <AdminFormField label="模板名称" required>
            <Input v-model="form.name" placeholder="例如 Carbon bicycle rim" />
          </AdminFormField>
          <AdminFormField label="Slug" required>
            <Input v-model="form.slug" class="font-mono" placeholder="例如 carbon-bicycle-rim" />
          </AdminFormField>
          <AdminFormField label="商品规格模板">
            <Select :model-value="templateProductSpecTemplateValue" @update:model-value="setProductSpecTemplate(String($event))">
              <SelectTrigger><SelectValue placeholder="全部商品规格模板" /></SelectTrigger>
              <SelectContent>
                <SelectItem value="__none__">全部商品规格模板</SelectItem>
                <SelectItem v-for="type in productSpecTemplates" :key="type.id" :value="String(type.id)">
                  {{ type.name }}
                </SelectItem>
              </SelectContent>
            </Select>
          </AdminFormField>
          <AdminFormField label="状态">
            <Select v-model="form.status">
              <SelectTrigger><SelectValue /></SelectTrigger>
              <SelectContent>
                <SelectItem value="active">启用</SelectItem>
                <SelectItem value="draft">草稿</SelectItem>
                <SelectItem value="paused">暂停</SelectItem>
              </SelectContent>
            </Select>
          </AdminFormField>
          <AdminFormField label="清关部件/产品族" description="例如 rim、hub、spoke、wheelset">
            <Input v-model="form.component_kind" class="font-mono" placeholder="rim" />
          </AdminFormField>
          <AdminFormField label="材质" description="例如 carbon_fiber、aluminum">
            <Input v-model="form.material" class="font-mono" placeholder="carbon_fiber" />
          </AdminFormField>
          <AdminFormField label="HS Code" required description="6 位数字">
            <Input v-model="form.hs_code" inputmode="numeric" maxlength="6" class="font-mono" />
          </AdminFormField>
          <AdminFormField label="CN Code" description="欧盟 8 位编码，可选">
            <Input v-model="form.cn_code" inputmode="numeric" maxlength="8" class="font-mono" />
          </AdminFormField>
          <AdminFormField label="原产国代码" description="可作为默认值，商品级允许覆盖">
            <Input v-model="form.country_of_origin" maxlength="2" class="font-mono uppercase" />
          </AdminFormField>
          <AdminFormField label="英文报关品名" class="sm:col-span-2">
            <Input v-model="form.customs_description" maxlength="255" placeholder="Bicycle carbon rim" />
          </AdminFormField>
          <AdminFormField label="来源">
            <Input v-model="form.source" class="font-mono" placeholder="us_hts" />
          </AdminFormField>
          <AdminFormField label="来源编码">
            <Input v-model="form.source_code" class="font-mono" />
          </AdminFormField>
          <AdminFormField label="来源链接" class="sm:col-span-2">
            <Input v-model="form.source_url" type="url" placeholder="https://..." />
          </AdminFormField>
          <AdminFormField label="备注" class="sm:col-span-2">
            <Textarea v-model="form.notes" class="min-h-20" placeholder="记录确认依据或适用范围。" />
          </AdminFormField>
        </div>
        <DialogFooter>
          <Button type="button" variant="outline" @click="emit('update:open', false)">取消</Button>
          <Button type="submit" :disabled="saving">
            <LoaderCircle v-if="saving" class="size-4 animate-spin" />
            保存模板
          </Button>
        </DialogFooter>
      </form>
    </DialogContent>
  </Dialog>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { LoaderCircle } from '@lucide/vue'
import AdminFormField from '@/components/admin/AdminFormField.vue'
import { Button } from '@/components/ui/button'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Textarea } from '@/components/ui/textarea'
import type { CustomsClassificationForm } from './customsClassificationTypes'

const props = withDefaults(defineProps<{
  open: boolean
  form: CustomsClassificationForm
  productSpecTemplates?: any[]
  saving?: boolean
}>(), {
  productSpecTemplates: () => [],
  saving: false,
})

const emit = defineEmits<{
  (event: 'update:open', value: boolean): void
  (event: 'save'): void
  (event: 'update:product-spec-template', value: string): void
}>()

const templateProductSpecTemplateValue = computed(() => (
  props.form.product_specification_template_id ? String(props.form.product_specification_template_id) : '__none__'
))

const setProductSpecTemplate = (value: string) => {
  emit('update:product-spec-template', value)
}
</script>
