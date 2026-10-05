<template>
  <Dialog :open="open" @update:open="emit('update:open', $event)">
    <DialogContent class="max-w-3xl">
      <DialogHeader>
        <DialogTitle>{{ form.id ? '编辑清关资料模板' : '新建清关资料模板' }}</DialogTitle>
        <DialogDescription>模板保存后，可在商品编辑器中一键填入四项清关资料。</DialogDescription>
      </DialogHeader>
      <div v-if="form.id" class="rounded-md border bg-muted/30 px-3 py-2 text-xs text-muted-foreground">
        <p>数据核验截至：{{ formatCustomsVerificationDate(form.verified_at) }}</p>
        <p class="mt-1">
          建议复核：{{ formatCustomsVerificationDate(form.review_due_at) }}
          <span
            class="ml-1 inline-flex rounded-full border px-2 py-0.5 font-medium"
            :class="getCustomsReviewLifecycleStatusBadgeClass(getCustomsReviewLifecycleStatus(form.verified_at, form.review_due_at))"
          >
            {{ getCustomsReviewLifecycleStatusLabel(getCustomsReviewLifecycleStatus(form.verified_at, form.review_due_at)) }}
          </span>
        </p>
        <p class="mt-1">修改编码、材质、原产国、官方来源或贸易救济风险字段后，系统会清空日期并要求重新核验。</p>
      </div>
      <form class="space-y-4" @submit.prevent="emit('save')">
        <div class="grid gap-4 sm:grid-cols-2">
          <AdminFormField label="模板名称" required>
            <Input v-model="form.name" placeholder="例如 Carbon bicycle rim" />
          </AdminFormField>
          <AdminFormField label="Slug" required>
            <Input v-model="form.slug" class="font-mono" placeholder="例如 carbon-bicycle-rim" />
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
            <Input v-model="form.source" class="font-mono" placeholder="us_hts" :disabled="form.source === 'built_in'" />
          </AdminFormField>
          <AdminFormField label="来源编码">
            <Input v-model="form.source_code" class="font-mono" />
          </AdminFormField>
          <AdminFormField label="🇺🇸 US HTS 官方链接" description="美国 HTS；保存时同步到兼容字段 source_url">
            <Input v-model="form.source_url_us" type="url" placeholder="https://hts.usitc.gov/..." />
          </AdminFormField>
          <AdminFormField label="🇪🇺 EU TARIC 官方链接">
            <Input v-model="form.source_url_eu" type="url" placeholder="https://ec.europa.eu/taxation_customs/dds2/taric" />
          </AdminFormField>
          <AdminFormField label="🇬🇧 UK Trade Tariff 官方链接">
            <Input v-model="form.source_url_uk" type="url" placeholder="https://www.gov.uk/trade-tariff/..." />
          </AdminFormField>
          <AdminFormField label="备注" class="sm:col-span-2">
            <Textarea v-model="form.notes" class="min-h-20" placeholder="记录确认依据或适用范围。" />
          </AdminFormField>
          <div class="sm:col-span-2 rounded-lg border border-amber-500/30 bg-amber-500/5 p-3">
            <div class="flex flex-wrap items-center justify-between gap-2">
              <div>
                <p class="text-sm font-semibold">贸易救济敏感度</p>
                <p class="mt-1 text-xs text-muted-foreground">标签只记录复核重点，不替代目的国当日的 TARIC / HTS 判定。</p>
              </div>
              <Select
                :model-value="form.trade_remedy_risk_level"
                @update:model-value="setTradeRemedyRiskLevel(form, $event)"
              >
                <SelectTrigger class="w-36"><SelectValue /></SelectTrigger>
                <SelectContent>
                  <SelectItem v-for="(label, value) in customsTradeRemedyRiskLevelLabels" :key="value" :value="value">
                    {{ label }}
                  </SelectItem>
                </SelectContent>
              </Select>
            </div>
            <div class="mt-3 grid gap-2 sm:grid-cols-3">
              <label
                v-for="option in customsTradeRemedyRiskTagOptions"
                :key="option.value"
                class="flex items-start gap-2 rounded-md border bg-background/70 px-2.5 py-2 text-xs"
              >
                <input
                  type="checkbox"
                  class="mt-0.5 size-3.5 accent-amber-600"
                  :checked="form.trade_remedy_risk_tags.includes(option.value)"
                  @change="toggleTradeRemedyRiskTag(form, option.value, $event)"
                />
                <span>{{ option.label }}</span>
              </label>
            </div>
            <AdminFormField label="申报建议" class="mt-3">
              <Textarea
                v-model="form.trade_remedy_declaration_advice"
                class="min-h-28"
                maxlength="4000"
                placeholder="写明目的国措施、原产地、货物完整性和必须复核的条件；禁止把拆单作为规避措施。"
              />
            </AdminFormField>
          </div>
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
import { LoaderCircle } from '@lucide/vue'
import AdminFormField from '@/components/admin/AdminFormField.vue'
import { Button } from '@/components/ui/button'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Textarea } from '@/components/ui/textarea'
import {
  customsTradeRemedyRiskLevelLabels,
  customsTradeRemedyRiskTagOptions,
  type CustomsClassificationForm,
  type CustomsTradeRemedyRiskTag,
} from '@/modules/customs/customsClassificationTypes'
import {
  getCustomsReviewLifecycleStatus,
  getCustomsReviewLifecycleStatusBadgeClass,
  getCustomsReviewLifecycleStatusLabel,
} from '@/modules/customs/customsReviewLifecycle'

withDefaults(defineProps<{
  open: boolean
  form: CustomsClassificationForm
  saving?: boolean
}>(), {
  saving: false,
})

const emit = defineEmits<{
  (event: 'update:open', value: boolean): void
  (event: 'save'): void
}>()

const formatCustomsVerificationDate = (value: string | null | undefined) => (
  value ? value.slice(0, 10) : '暂无记录'
)

const toggleTradeRemedyRiskTag = (form: CustomsClassificationForm, tag: CustomsTradeRemedyRiskTag, event: Event): void => {
  const checked = (event.target as HTMLInputElement | null)?.checked === true
  const tags = new Set(form.trade_remedy_risk_tags)
  if (checked) tags.add(tag)
  else tags.delete(tag)
  form.trade_remedy_risk_tags = Array.from(tags) as CustomsTradeRemedyRiskTag[]
}

const setTradeRemedyRiskLevel = (form: CustomsClassificationForm, value: unknown): void => {
  form.trade_remedy_risk_level = String(value) as CustomsClassificationForm['trade_remedy_risk_level']
}
</script>
