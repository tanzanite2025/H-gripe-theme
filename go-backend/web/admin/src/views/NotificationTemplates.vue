<template>
  <div class="flex h-full min-h-0 flex-col gap-4 overflow-y-auto lg:overflow-hidden">
    <AdminPageHeader class="shrink-0" title="邮件模板" description="按前台 20 种语言独立维护订单与售后邮件，保存会生成版本快照。" />

    <div class="shrink-0 space-y-2">
      <p class="px-1 text-[10px] font-bold text-muted-foreground">邮件内容语言 · 与后台界面语言独立</p>
      <AdminStorefrontLanguageDisplayCard
        :model-value="notificationTemplateContentLocale"
        :language-options="notificationTemplateContentLanguageOptions"
        :disabled="loading || saving || rollingBack"
        aria-label="邮件模板内容语言"
        @update:model-value="selectNotificationTemplateContentLocale"
      />
    </div>

    <div class="grid gap-4 lg:min-h-0 lg:flex-1 lg:grid-cols-[18rem_minmax(0,1fr)]">
      <Card class="min-h-0">
        <CardHeader class="shrink-0 border-b border-dashed">
          <CardTitle><h2>模板列表</h2></CardTitle>
          <CardAction>
            <Button size="icon-sm" variant="ghost" :disabled="loading" title="刷新" aria-label="刷新模板列表" @click="loadTemplates">
              <RefreshCw class="size-4" :class="loading ? 'animate-spin' : ''" />
            </Button>
          </CardAction>
        </CardHeader>
        <CardContent class="min-h-0 flex-1 overflow-y-auto">
          <div v-if="loading" class="py-8 text-center text-xs text-muted-foreground">正在加载...</div>
          <div v-else class="space-y-3">
            <section v-for="group in notificationTemplateGroups" :key="group.key" class="border-b border-dashed border-border/60 pb-3 last:border-0 last:pb-0">
              <button
                type="button"
                class="flex w-full items-center justify-between gap-2 rounded-xl px-2 py-2 text-left text-xs font-black transition-colors hover:bg-muted/70 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
                :aria-expanded="isNotificationTemplateGroupExpanded(group.key)"
                @click="toggleNotificationTemplateGroup(group.key)"
              >
                <span>{{ group.label }}</span>
                <span class="flex items-center gap-2 text-xs text-muted-foreground">
                  <span>{{ group.templates.length }}</span>
                  <ChevronDown class="size-4 transition-transform" :class="isNotificationTemplateGroupExpanded(group.key) ? 'rotate-180' : ''" />
                </span>
              </button>
              <div v-if="isNotificationTemplateGroupExpanded(group.key)" class="mt-1 space-y-1">
                <button
                  v-for="template in group.templates"
                :key="template.code"
                  type="button"
                  class="w-full rounded-xl border border-dashed px-3 py-2 text-left text-xs transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
                  :class="selected?.code === template.code ? 'border-admin-selected-border bg-admin-selected-soft' : 'border-transparent hover:bg-muted/70'"
                  :aria-pressed="selected?.code === template.code"
                  :disabled="saving || rollingBack"
                  @click="selectNotificationTemplateFromList(template)"
                >
                  <div class="flex items-center justify-between gap-2">
                    <span class="truncate font-bold">{{ getNotificationTemplateDisplayName(template) }}</span>
                    <span class="shrink-0 font-mono text-[10px] text-muted-foreground">{{ template.locale }}</span>
                  </div>
                  <div class="mt-1.5 flex items-center justify-between gap-2 text-[10px] text-muted-foreground">
                    <span class="min-w-0 truncate font-mono" :title="template.code">{{ template.code }}</span>
                    <span class="flex shrink-0 items-center gap-1.5">
                      <span v-if="template.id">v{{ template.version }}</span>
                      <AdminStatusBadge :tone="template.id && template.is_enabled ? 'green' : 'gray'">{{ template.id ? (template.is_enabled ? '启用' : '停用') : '未配置' }}</AdminStatusBadge>
                    </span>
                  </div>
                </button>
              </div>
            </section>
          </div>
        </CardContent>
      </Card>

      <Card v-if="selected" class="min-h-0 min-w-0 overflow-y-auto">
        <CardHeader class="shrink-0 border-b border-dashed">
          <div class="flex flex-wrap items-start justify-between gap-3">
            <div class="min-w-0 space-y-1.5">
              <div class="flex flex-wrap items-center gap-2">
                <CardTitle><h2>{{ getNotificationTemplateDisplayName(selected) }}</h2></CardTitle>
                <AdminStatusBadge :tone="selected.id && selected.is_enabled ? 'green' : 'gray'">{{ selected.id ? (selected.is_enabled ? '启用' : '停用') : '未配置' }}</AdminStatusBadge>
              </div>
              <p class="break-all font-mono text-[10px] text-muted-foreground">
                {{ selected.code }} / {{ notificationTemplateContentLanguageLabel }} · {{ selected.id ? `当前版本 v${selected.version}` : '尚未创建此语言版本' }}
              </p>
            </div>
            <div class="flex gap-2">
              <Button :disabled="loading || saving || rollingBack || !canEditNotificationTemplates" @click="saveTemplate">
                <LoaderCircle v-if="saving" class="mr-2 size-4 animate-spin" />
                <Save v-else class="mr-2 size-4" />保存
              </Button>
            </div>
          </div>
        </CardHeader>

        <CardContent class="space-y-5">
          <div v-if="!selected.id" class="flex flex-wrap items-center justify-between gap-3 rounded-2xl border border-dashed border-border/80 bg-muted/25 p-3">
            <p class="min-w-0 flex-1 text-xs leading-5 text-muted-foreground">
              {{ notificationTemplateContentLanguageLabel }}尚未配置此模板。填写并保存后创建独立版本；未配置或停用时，发信会回退到已启用的英文模板。
            </p>
            <Button v-if="englishNotificationTemplateForSelectedCode" variant="outline" size="sm" :disabled="!canEditNotificationTemplates" @click="copyEnglishNotificationTemplateIntoCurrentLocaleDraft">从英文复制</Button>
          </div>
          <div class="grid items-end gap-4 sm:grid-cols-[minmax(0,1fr)_auto]">
            <AdminFormField label="名称" for-id="notification-template-name">
              <Input id="notification-template-name" v-model="draft.name" :disabled="!canEditNotificationTemplates" />
            </AdminFormField>
            <div class="flex h-9 items-center gap-3 px-1">
              <label for="notification-template-enabled" class="text-xs font-bold">启用模板</label>
              <Switch id="notification-template-enabled" v-model="draft.is_enabled" size="sm" :disabled="!canEditNotificationTemplates" />
            </div>
            <AdminFormField label="邮件标题" for-id="notification-template-subject" class="sm:col-span-2">
              <NotificationTemplateSubjectEditor
                id="notification-template-subject"
                aria-label="邮件标题"
                v-model="draft.subject_template"
                :lang="notificationTemplateContentHTMLLanguage"
                :dir="notificationTemplateContentTextDirection"
                :disabled="!canEditNotificationTemplates"
                :template-variable-display-values="notificationTemplateVariableDisplayValues"
              />
            </AdminFormField>
          </div>
          <div class="grid items-stretch gap-4 xl:grid-cols-2">
            <AdminFormField label="邮件正文">
              <RichTextEditor
                class="notification-template-body-editor h-[28rem] min-w-0"
                v-model="draft.body_html"
                :lang="notificationTemplateContentHTMLLanguage"
                :dir="notificationTemplateContentTextDirection"
                :disabled="!canEditNotificationTemplates"
                :template-variable-display-values="notificationTemplateVariableDisplayValues"
              />
            </AdminFormField>
            <AdminFormField label="实时预览">
              <div class="flex h-[28rem] min-h-0 flex-col overflow-hidden rounded-2xl border border-dashed border-border/70 bg-muted/10 p-3">
                <div class="mb-3 border-b border-dashed border-border/60 px-1 pb-3">
                  <div class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/70">邮件标题</div>
                  <div class="mt-1 break-words text-xs font-bold">
                    {{ liveNotificationTemplatePreviewSubject }}
                  </div>
                </div>
                <iframe
                  :srcdoc="liveNotificationTemplatePreviewDocument"
                  title="邮件正文实时预览"
                  class="min-h-0 w-full flex-1 rounded-xl border-0 bg-white"
                  sandbox="allow-same-origin"
                />
              </div>
            </AdminFormField>
          </div>
          <AdminFormField label="变更原因" for-id="notification-template-change-reason">
            <Input id="notification-template-change-reason" v-model="draft.change_reason" :disabled="!canEditNotificationTemplates" placeholder="例如：更新售后寄回说明" />
          </AdminFormField>

          <div v-if="versions.length" class="border-t border-dashed pt-4">
            <button
              type="button"
              class="flex w-full items-center justify-between gap-2 rounded-xl text-left text-xs font-black focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
              :aria-expanded="versionsHistoryExpanded"
              @click="versionsHistoryExpanded = !versionsHistoryExpanded"
            >
              <span>版本历史</span>
              <span class="flex items-center gap-2 text-xs text-muted-foreground">
                <span>{{ versions.length }}</span>
                <ChevronDown class="size-4 transition-transform" :class="versionsHistoryExpanded ? 'rotate-180' : ''" />
              </span>
            </button>
            <div v-if="versionsHistoryExpanded" class="mt-3 divide-y divide-dashed text-xs">
              <div v-for="version in versions" :key="version.id" class="flex flex-wrap items-center justify-between gap-x-3 gap-y-1 py-2">
                <span class="min-w-0 truncate">v{{ version.version }} {{ version.change_reason || '无说明' }}</span>
                <span class="flex shrink-0 items-center gap-2 text-muted-foreground">
                  <span>{{ formatDate(version.created_at) }}</span>
                  <Button
                    v-if="version.version < selected.version"
                    size="icon"
                    variant="ghost"
                    class="size-7"
                    :disabled="saving || rollingBack || !canEditNotificationTemplates"
                    title="回滚为新版本"
                    @click.stop="rollbackVersion(version.version)"
                  >
                    <RotateCcw class="size-3.5" :class="rollingBack ? 'animate-spin' : ''" />
                  </Button>
                </span>
              </div>
            </div>
          </div>
        </CardContent>
      </Card>
      <Card v-else class="min-h-0 min-w-0 items-center justify-center">
        <CardContent class="py-10 text-center text-xs text-muted-foreground">请选择一个模板</CardContent>
      </Card>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { toast } from 'vue-sonner'
import { ChevronDown, LoaderCircle, RefreshCw, RotateCcw, Save } from '@lucide/vue'
import AdminPageHeader from '@/components/admin/AdminPageHeader.vue'
import AdminFormField from '@/components/admin/AdminFormField.vue'
import AdminStatusBadge from '@/components/admin/AdminStatusBadge.vue'
import AdminStorefrontLanguageDisplayCard from '@/components/admin/AdminStorefrontLanguageDisplayCard.vue'
import NotificationTemplateSubjectEditor from '@/components/admin/settings/NotificationTemplateSubjectEditor.vue'
import RichTextEditor from '@/components/admin/settings/RichTextEditor.vue'
import { Button } from '@/components/ui/button'
import { Card, CardAction, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Switch } from '@/components/ui/switch'
import { useAdminI18n } from '@/i18n'
import { buildLanguageOptions, normalizeLocaleCode, STOREFRONT_SUPPORTED_LANGUAGES } from '@/lib/languages'
import { useAuthStore } from '@/stores/auth'
import notificationTemplatesApi, { type NotificationTemplate, type NotificationTemplateVersion } from '@/api/notificationTemplates'

const notificationTemplatePreviewFontURL = new URL(
  '../assets/fonts/maple-ui/MapleUI-CJK.f8ce6d72e8cb.woff2',
  import.meta.url,
).href

const templates = ref<NotificationTemplate[]>([])
const selected = ref<NotificationTemplate | null>(null)
const versions = ref<NotificationTemplateVersion[]>([])
const loading = ref(false)
const saving = ref(false)
const rollingBack = ref(false)
const versionsHistoryExpanded = ref(false)
const draft = reactive({ name: '', subject_template: '', body_html: '', is_enabled: true, change_reason: '' })
const { locale } = useAdminI18n()
const notificationTemplateContentLanguageOptions = buildLanguageOptions(STOREFRONT_SUPPORTED_LANGUAGES)
const notificationTemplateContentLocale = ref('en')
const notificationTemplateContentHTMLLanguage = computed(() => notificationTemplateContentLocale.value.replace('_', '-'))
const notificationTemplateContentTextDirection = computed(() => notificationTemplateContentLocale.value === 'ar' ? 'rtl' : 'ltr')
const notificationTemplateContentLanguageLabel = computed(() => notificationTemplateContentLanguageOptions.find(
  (language) => language.value === notificationTemplateContentLocale.value,
)?.label || notificationTemplateContentLocale.value)
const notificationTemplateAuthStore = useAuthStore()
const canEditNotificationTemplates = computed(() => notificationTemplateAuthStore.hasPermission('settings:edit'))
let notificationTemplateSelectionRequestSequence = 0

const englishNotificationTemplateForSelectedCode = computed(() => templates.value.find(
  (template) => template.code === selected.value?.code && normalizeLocaleCode(template.locale) === 'en',
))

const hasUnsavedNotificationTemplateChanges = computed(() => Boolean(selected.value && (
  draft.name !== selected.value.name || draft.subject_template !== selected.value.subject_template
  || draft.body_html !== selected.value.body_html || draft.is_enabled !== selected.value.is_enabled || draft.change_reason.trim()
)))

const getNotificationTemplateForContentLocale = (source: NotificationTemplate, contentLocale: string): NotificationTemplate => {
  const existing = templates.value.find((template) => template.code === source.code && normalizeLocaleCode(template.locale) === contentLocale)
  if (existing) return existing
  return {
    ...source, id: 0, locale: contentLocale, version: 0, subject_template: '', body_html: '', body_text: '', is_enabled: false, updated_at: '',
  }
}

const selectNotificationTemplateContentLocale = async (contentLocale: string): Promise<void> => {
  if (loading.value || saving.value || rollingBack.value || contentLocale === notificationTemplateContentLocale.value) return
  if (!notificationTemplateContentLanguageOptions.some((language) => language.value === contentLocale)) return
  if (hasUnsavedNotificationTemplateChanges.value && !window.confirm('当前模板有未保存的修改，确定切换邮件内容语言并放弃修改吗？')) return
  notificationTemplateContentLocale.value = contentLocale
  if (selected.value) await selectTemplate(getNotificationTemplateForContentLocale(selected.value, contentLocale))
}

const selectNotificationTemplateFromList = async (template: NotificationTemplate): Promise<void> => {
  if (saving.value || rollingBack.value || selected.value?.code === template.code) return
  await selectTemplate(template)
}

const copyEnglishNotificationTemplateIntoCurrentLocaleDraft = (): void => {
  const source = englishNotificationTemplateForSelectedCode.value
  if (!source || !selected.value || selected.value.id || !canEditNotificationTemplates.value) return
  if (hasUnsavedNotificationTemplateChanges.value && !window.confirm('确定用英文内容替换当前未保存的修改吗？')) return
  draft.name = source.name
  draft.subject_template = source.subject_template
  draft.body_html = source.body_html
}

const notificationTemplateChineseNameByCode: Record<string, string> = {
  order_confirmation: '订单确认',
  order_payment_expired: '付款已超时',
  order_cancelled: '订单已取消',
  order_shipping_notification: '订单发货通知',
  order_delivered: '订单已送达',
  order_completed: '订单已完成',
  order_refunded: '订单退款完成',
  after_sales_requested: '已收到售后申请',
  after_sales_approved: '售后申请已批准',
  after_sales_awaiting_return: '售后寄回说明',
  after_sales_return_in_transit: '售后退货运输中',
  after_sales_received: '售后退货已收到',
  after_sales_resolving: '售后处理中',
  after_sales_completed: '售后已完成',
  after_sales_rejected: '售后申请已拒绝',
}

const getNotificationTemplateDisplayName = (template: NotificationTemplate): string => {
  if (locale.value === 'zh-CN') {
    return notificationTemplateChineseNameByCode[template.code] || template.name
  }
  return template.name
}

type NotificationTemplateGroupKey = 'order' | 'after_sales' | 'shipping'

interface NotificationTemplateGroup {
  key: NotificationTemplateGroupKey
  label: string
  templates: NotificationTemplate[]
}

const expandedNotificationTemplateGroupKey = ref<NotificationTemplateGroupKey | null>('order')

const getNotificationTemplateGroupKey = (template: NotificationTemplate): NotificationTemplateGroupKey => {
  if (template.code === 'order_shipping_notification') return 'shipping'
  if (template.category === 'after_sales') return 'after_sales'
  return 'order'
}

const notificationTemplateGroups = computed<NotificationTemplateGroup[]>(() => {
  const groups: NotificationTemplateGroup[] = [
    { key: 'order', label: '订单', templates: [] },
    { key: 'after_sales', label: '售后', templates: [] },
    { key: 'shipping', label: '发货', templates: [] },
  ]
  const groupsByKey = new Map(groups.map((group) => [group.key, group]))

  const templateSourcesByCode = new Map<string, NotificationTemplate>()
  templates.value.forEach((template) => {
    if (!templateSourcesByCode.has(template.code) || normalizeLocaleCode(template.locale) === 'en') templateSourcesByCode.set(template.code, template)
  })
  templateSourcesByCode.forEach((template) => {
    groupsByKey.get(getNotificationTemplateGroupKey(template))?.templates.push(getNotificationTemplateForContentLocale(template, notificationTemplateContentLocale.value))
  })

  return groups.filter((group) => group.templates.length > 0)
})

const isNotificationTemplateGroupExpanded = (groupKey: NotificationTemplateGroupKey): boolean => (
  expandedNotificationTemplateGroupKey.value === groupKey
)

const toggleNotificationTemplateGroup = (groupKey: NotificationTemplateGroupKey): void => {
  expandedNotificationTemplateGroupKey.value = expandedNotificationTemplateGroupKey.value === groupKey ? null : groupKey
}

const notificationTemplateVariableDisplayValueByName: Record<string, string> = {
  order_number: 'SO-20250101-001',
  order_amount: '¥99.00',
  paid_at: '2025-01-01 12:00',
  expired_at: '2025-01-01 12:30',
  cancelled_at: '2025-01-01 12:30',
  cancellation_reason: '客户申请取消',
  carrier_name: '顺丰速运',
  tracking_number: 'SF123456789',
  tracking_url: '查看物流详情',
  shipped_at: '2025-01-02 09:00',
  delivered_at: '2025-01-03 16:00',
  completed_at: '2025-01-04 10:00',
  refund_amount: '¥99.00',
  currency: 'CNY',
  refunded_at: '2025-01-05 11:00',
  refund_reason: '商品退回完成',
  after_sales_case_number: 'AS-20250101-001',
  after_sales_type: '退货退款',
  after_sales_status: '处理中',
  requested_at: '2025-01-01 13:00',
  approved_at: '2025-01-01 14:00',
  next_step: '请等待后续处理',
  return_warehouse_name: '上海退货仓',
  return_warehouse_address: '上海市浦东新区示例路 1 号',
  return_deadline: '2025-01-10',
  received_at: '2025-01-06 15:00',
  resolution_note: '审核完成',
  rejection_reason: '不符合售后条件',
  rejected_at: '2025-01-02 09:00',
  customer_name: '张三',
  items: '商品明细',
  resolution: '退款已完成',
}

const notificationTemplateVariableDisplayValues = computed<Record<string, string>>(() => (
  Object.fromEntries((selected.value?.allowed_variables || []).map((variableName) => [
    variableName,
    notificationTemplateVariableDisplayValueByName[variableName] || '示例内容',
  ]))
))

const escapeNotificationTemplatePreviewHTML = (value: string): string => value
  .replace(/&/g, '&amp;')
  .replace(/</g, '&lt;')
  .replace(/>/g, '&gt;')
  .replace(/"/g, '&quot;')
  .replace(/'/g, '&#39;')

const convertNotificationTemplateHTMLToPlainText = (html: string): string => {
  if (!html.trim() || typeof DOMParser === 'undefined') return html.replace(/<[^>]+>/g, '').trim()

  const parsedDocument = new DOMParser().parseFromString(html, 'text/html')
  parsedDocument.body.querySelectorAll('br').forEach((lineBreak) => lineBreak.replaceWith('\n'))
  parsedDocument.body.querySelectorAll('p, div, li, h1, h2, h3, h4, blockquote').forEach((block) => {
    block.insertAdjacentText('afterend', '\n')
  })

  return (parsedDocument.body.textContent || '')
    .replace(/\u00a0/g, ' ')
    .replace(/[ \t]+\n/g, '\n')
    .replace(/\n{3,}/g, '\n\n')
    .trim()
}

const renderNotificationTemplatePreviewSource = (source: string, variables: Record<string, string>): string => (
  source.replace(/\{\{\s*([a-z][a-z0-9_]*)\s*\}\}/gi, (_, variableName: string) => (
    escapeNotificationTemplatePreviewHTML(variables[variableName] || '示例内容')
  ))
)

const renderNotificationTemplatePreviewText = (source: string, variables: Record<string, string>): string => (
  source.replace(/\{\{\s*([a-z][a-z0-9_]*)\s*\}\}/gi, (_, variableName: string) => (
    variables[variableName] || '示例内容'
  ))
)

const liveNotificationTemplatePreviewVariables = notificationTemplateVariableDisplayValues

const liveNotificationTemplatePreviewSubject = computed(() => renderNotificationTemplatePreviewText(
  draft.subject_template,
  liveNotificationTemplatePreviewVariables.value,
))

const liveNotificationTemplatePreviewDocument = computed(() => {
  const previewBody = renderNotificationTemplatePreviewSource(
    draft.body_html,
    liveNotificationTemplatePreviewVariables.value,
  )
  return `<!doctype html><html lang="${notificationTemplateContentHTMLLanguage.value}" dir="${notificationTemplateContentTextDirection.value}"><head><meta charset="utf-8"><style>@font-face{font-family:MapleUICJK;src:url('${notificationTemplatePreviewFontURL}') format('woff2');font-weight:100 900;font-style:normal;font-display:block}body{margin:0;padding:18px;color:#172033;font:14px/1.65 MapleUICJK}p{margin:0 0 12px}h1,h2,h3{margin:0 0 12px}a{color:#2563eb}blockquote{border-inline-start:3px solid #94a3b8;margin:0 0 12px;padding-inline-start:12px;color:#64748b}</style></head><body>${previewBody || '<p style="color:#94a3b8">暂无正文</p>'}</body></html>`
})

const applyDraft = (template: NotificationTemplate) => {
  draft.name = template.name; draft.subject_template = template.subject_template; draft.body_html = template.body_html; draft.is_enabled = template.is_enabled; draft.change_reason = ''
}

const loadTemplates = async () => {
  if (saving.value || rollingBack.value) return
  if (hasUnsavedNotificationTemplateChanges.value && !window.confirm('当前模板有未保存的修改，确定刷新并放弃修改吗？')) return
  loading.value = true
  try {
    templates.value = await notificationTemplatesApi.list()
    const source = templates.value.find((template) => template.code === selected.value?.code) || templates.value[0]
    if (source) await selectTemplate(getNotificationTemplateForContentLocale(source, notificationTemplateContentLocale.value))
    else { selected.value = null; versions.value = []; notificationTemplateSelectionRequestSequence += 1 }
  } catch (error) { toast.error(error instanceof Error ? error.message : '模板加载失败') } finally { loading.value = false }
}

const selectTemplate = async (template: NotificationTemplate) => {
  if (selected.value?.code !== template.code && hasUnsavedNotificationTemplateChanges.value && !window.confirm('当前模板有未保存的修改，确定切换模板并放弃修改吗？')) return
  const selectionRequestSequence = ++notificationTemplateSelectionRequestSequence
  selected.value = template
  versions.value = []
  expandedNotificationTemplateGroupKey.value = getNotificationTemplateGroupKey(template)
  versionsHistoryExpanded.value = false
  applyDraft(template)
  if (!template.id) return
  try {
    const loadedVersions = await notificationTemplatesApi.versions(template.id)
    if (selectionRequestSequence === notificationTemplateSelectionRequestSequence) versions.value = loadedVersions
  } catch (error) { if (selectionRequestSequence === notificationTemplateSelectionRequestSequence) toast.error(error instanceof Error ? error.message : '版本历史加载失败') }
}

const saveTemplate = async () => {
  if (!selected.value || saving.value || !canEditNotificationTemplates.value) return
  if (!draft.name.trim() || !draft.subject_template.trim() || !convertNotificationTemplateHTMLToPlainText(draft.body_html)) {
    toast.error('请填写模板名称、邮件标题和邮件正文')
    return
  }
  saving.value = true
  try {
    const updated = await notificationTemplatesApi.save({
      id: selected.value.id,
      code: selected.value.code,
      locale: selected.value.locale,
      category: selected.value.category,
      version: selected.value.version + 1,
      allowed_variables: selected.value.allowed_variables,
      required_variables: selected.value.required_variables,
      ...draft,
      body_text: convertNotificationTemplateHTMLToPlainText(draft.body_html),
    })
    selected.value = updated
    const existingTemplateIndex = templates.value.findIndex((item) => item.id === updated.id)
    if (existingTemplateIndex >= 0) templates.value[existingTemplateIndex] = updated
    else templates.value.push(updated)
    await selectTemplate(updated)
    toast.success('模板已保存')
  } catch (error) {
    toast.error(error instanceof Error ? error.message : '模板保存失败')
  } finally {
    saving.value = false
  }
}

const rollbackVersion = async (version: number) => {
  if (!selected.value?.id || saving.value || rollingBack.value || !canEditNotificationTemplates.value || !window.confirm(`确定将模板回滚到 v${version} 吗？这会生成一个新的当前版本。`)) return
  rollingBack.value = true
  try {
    const updated = await notificationTemplatesApi.rollback(selected.value.id, version, `rollback to version ${version}`)
    selected.value = updated
    templates.value = templates.value.map((item) => item.id === updated.id ? updated : item)
    await selectTemplate(updated)
    toast.success('模板已回滚并生成新版本')
  } catch (error) {
    toast.error(error instanceof Error ? error.message : '模板回滚失败')
  } finally {
    rollingBack.value = false
  }
}

const formatDate = (value: string) => value ? new Date(value).toLocaleString() : ''
onMounted(loadTemplates)
</script>

<style scoped>
.notification-template-body-editor {
  min-height: 0;
  border-radius: 1rem;
  border-style: dashed;
  border-color: color-mix(in oklab, var(--border) 70%, transparent);
  background: color-mix(in oklab, var(--muted) 10%, transparent);
  padding: 0.75rem;
}

.notification-template-body-editor :deep(.border-b) {
  border-bottom-style: dashed;
}

.notification-template-body-editor :deep(.rich-text-editor__canvas) {
  min-height: 0;
  border-radius: 0.75rem;
}
</style>
