<template>
  <Card class="overflow-hidden">
    <CardHeader class="border-b">
      <div class="flex flex-wrap items-start justify-between gap-3">
        <div>
          <CardTitle>清关资料模板</CardTitle>
          <CardDescription>系统已内置常用自行车部件模板，并显示官方资料核验截至日期；距复核超过 30 天为健康，30 天内为临期，过期或缺日期需重新核验。</CardDescription>
        </div>
        <Button v-if="canCreate" variant="outline" size="sm" @click="emit('create')">
          <Plus class="size-3.5" />
          添加自定义模板
        </Button>
      </div>
    </CardHeader>
    <CardContent class="p-0">
      <div v-if="loading" class="p-6 text-center text-sm text-muted-foreground">加载中...</div>
      <div v-else-if="templates.length" class="divide-y">
        <div v-for="template in templates" :key="template.id" class="flex items-start justify-between gap-3 p-4">
          <div class="min-w-0">
            <div class="flex flex-wrap items-center gap-2">
              <p class="font-semibold">{{ template.name }}</p>
              <span v-if="template.source === 'built_in'" class="rounded-full border border-sky-500/20 bg-sky-500/10 px-2 py-0.5 text-[11px] text-sky-700 dark:text-sky-200">
                内置
              </span>
              <span class="rounded-full bg-muted px-2 py-0.5 text-[11px] text-muted-foreground">
                {{ template.status === 'paused' ? '暂停' : template.status === 'draft' ? '草稿' : '启用' }}
              </span>
              <span
                class="rounded-full border px-2 py-0.5 text-[11px] font-medium"
                :class="getCustomsReviewLifecycleStatusBadgeClass(getCustomsReviewLifecycleStatus(template.verified_at, template.review_due_at))"
              >
                {{ getCustomsReviewLifecycleStatusLabel(getCustomsReviewLifecycleStatus(template.verified_at, template.review_due_at)) }}
              </span>
            </div>
            <p class="mt-1 font-mono text-xs text-muted-foreground">
              {{ template.component_kind || '未分组件' }} · {{ template.material || '未分材质' }}
            </p>
            <p class="mt-2 font-mono text-xs text-foreground">
              HS {{ template.hs_code }}<span v-if="template.cn_code"> · CN {{ template.cn_code }}</span>
            </p>
            <p v-if="template.customs_description" class="mt-1 line-clamp-2 text-xs text-muted-foreground">
              {{ template.customs_description }}
            </p>
            <div
              v-if="hasTradeRemedyRisk(template)"
              class="mt-3 rounded-md border px-3 py-2 text-xs"
              :class="tradeRemedyRiskClass(template)"
            >
              <div class="flex flex-wrap items-center gap-1.5 font-semibold">
                <span>贸易救济 {{ tradeRemedyRiskHeadline(template) }}</span>
              </div>
              <p v-if="template.trade_remedy_declaration_advice" class="mt-1 whitespace-pre-line leading-5">
                {{ template.trade_remedy_declaration_advice }}
              </p>
            </div>
            <p class="mt-2 text-xs text-muted-foreground">
              <template v-if="template.verified_at">
                数据核验截至 {{ formatCustomsVerificationDate(template.verified_at) }}
              </template>
              <span v-else>暂无核验日期</span>
              <span v-if="template.review_due_at"> · 建议复核 {{ formatCustomsVerificationDate(template.review_due_at) }}</span>
            </p>
            <div
              v-if="template.source_url_us || template.source_url_eu || template.source_url_uk || template.source_url"
              class="mt-2 flex flex-wrap items-center gap-x-3 gap-y-1 text-xs"
            >
              <span class="text-muted-foreground">目的国官方来源：</span>
              <a
                v-if="template.source_url_us || template.source_url"
                :href="template.source_url_us || template.source_url"
                target="_blank"
                rel="noopener noreferrer"
                class="text-primary underline-offset-2 hover:underline"
              >
                🇺🇸 US HTS
              </a>
              <a
                v-if="template.source_url_eu"
                :href="template.source_url_eu"
                target="_blank"
                rel="noopener noreferrer"
                class="text-primary underline-offset-2 hover:underline"
              >
                🇪🇺 EU TARIC
              </a>
              <a
                v-if="template.source_url_uk"
                :href="template.source_url_uk"
                target="_blank"
                rel="noopener noreferrer"
                class="text-primary underline-offset-2 hover:underline"
              >
                🇬🇧 UK Trade Tariff
              </a>
            </div>
          </div>
          <div class="flex shrink-0 items-center gap-1">
            <Button
              v-if="canEdit"
              variant="ghost"
              size="icon"
              aria-label="编辑清关模板"
              @click="emit('edit', template)"
            >
              <Pencil class="size-4" />
            </Button>
            <Button
              v-if="canDelete && template.source !== 'built_in'"
              variant="ghost"
              size="icon"
              aria-label="删除清关模板"
              @click="emit('delete', template)"
            >
              <Trash2 class="size-4 text-destructive" />
            </Button>
          </div>
        </div>
      </div>
      <p v-else class="p-6 text-center text-sm text-muted-foreground">暂无清关模板。</p>
    </CardContent>
  </Card>
</template>

<script setup lang="ts">
import { Pencil, Plus, Trash2 } from '@lucide/vue'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import {
  customsTradeRemedyRiskLevelLabels,
  customsTradeRemedyRiskTagLabels,
  type CustomsClassificationRecord,
  type CustomsTradeRemedyRiskLevel,
  type CustomsTradeRemedyRiskTag,
} from '@/modules/customs/customsClassificationTypes'
import {
  getCustomsReviewLifecycleStatus,
  getCustomsReviewLifecycleStatusBadgeClass,
  getCustomsReviewLifecycleStatusLabel,
} from '@/modules/customs/customsReviewLifecycle'

withDefaults(defineProps<{
  templates?: CustomsClassificationRecord[]
  loading?: boolean
  canCreate?: boolean
  canEdit?: boolean
  canDelete?: boolean
}>(), {
  templates: () => [],
  loading: false,
  canCreate: false,
  canEdit: false,
  canDelete: false,
})

const emit = defineEmits<{
  (event: 'create'): void
  (event: 'edit', template: CustomsClassificationRecord): void
  (event: 'delete', template: CustomsClassificationRecord): void
}>()

const formatCustomsVerificationDate = (value: string | null | undefined) => (
  value ? value.slice(0, 10) : '未记录'
)

const riskLevelLabel = (template: CustomsClassificationRecord): string => (
  customsTradeRemedyRiskLevelLabels[template.trade_remedy_risk_level as CustomsTradeRemedyRiskLevel] || '无特别标记'
)

const riskTagLabel = (tag: string): string => (
  customsTradeRemedyRiskTagLabels[tag as CustomsTradeRemedyRiskTag] || tag
)

const tradeRemedyRiskHeadline = (template: CustomsClassificationRecord): string => {
  const tags = (template.trade_remedy_risk_tags || []).map(riskTagLabel)
  return `[${riskLevelLabel(template)}${tags.length ? `·${tags.join(' / ')}` : ''}]`
}

const hasTradeRemedyRisk = (template: CustomsClassificationRecord): boolean => (
  (template.trade_remedy_risk_level || 'none') !== 'none'
  || Boolean(template.trade_remedy_risk_tags?.length)
  || Boolean(template.trade_remedy_declaration_advice?.trim())
)

const tradeRemedyRiskClass = (template: CustomsClassificationRecord): string => {
  const level = template.trade_remedy_risk_level || 'none'
  return level === 'high' || level === 'critical'
    ? 'border-red-500/40 bg-red-500/10 text-red-800 dark:text-red-200'
    : 'border-amber-500/30 bg-amber-500/10 text-amber-800 dark:text-amber-200'
}
</script>
