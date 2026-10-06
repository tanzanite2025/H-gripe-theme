<template>
  <div class="flex h-full min-h-0 flex-col gap-6 overflow-y-auto lg:overflow-hidden">
    <AdminPageHeader class="shrink-0" title="邮件模板" description="维护订单与售后事务邮件模板，保存会生成不可变版本快照。" />

    <div class="grid gap-6 lg:min-h-0 lg:flex-1 lg:grid-cols-[18rem_minmax(0,1fr)]">
      <section class="flex min-h-0 flex-col rounded-lg border bg-card p-4">
        <div class="mb-3 flex shrink-0 items-center justify-between">
          <h2 class="font-semibold">模板列表</h2>
          <Button size="icon" variant="ghost" :disabled="loading" title="刷新" @click="loadTemplates">
            <RefreshCw class="size-4" :class="loading ? 'animate-spin' : ''" />
          </Button>
        </div>
        <div v-if="loading" class="min-h-0 flex-1 py-8 text-center text-sm text-muted-foreground">正在加载...</div>
        <div v-else class="min-h-0 flex-1 space-y-2 overflow-y-auto pr-1">
          <section v-for="group in notificationTemplateGroups" :key="group.key" class="rounded-md border border-border/70">
            <button
              type="button"
              class="flex w-full items-center justify-between gap-2 px-3 py-2 text-left text-sm font-semibold transition-colors hover:bg-muted/70"
              :aria-expanded="isNotificationTemplateGroupExpanded(group.key)"
              @click="toggleNotificationTemplateGroup(group.key)"
            >
              <span>{{ group.label }}</span>
              <span class="flex items-center gap-2 text-xs text-muted-foreground">
                <span>{{ group.templates.length }}</span>
                <ChevronDown class="size-4 transition-transform" :class="isNotificationTemplateGroupExpanded(group.key) ? 'rotate-180' : ''" />
              </span>
            </button>
            <div v-if="isNotificationTemplateGroupExpanded(group.key)" class="space-y-1 border-t border-border/60 p-1">
              <button
                v-for="template in group.templates"
                :key="template.id"
                type="button"
                class="w-full rounded-md border px-3 py-2 text-left text-sm transition-colors hover:bg-muted"
                :class="selected?.id === template.id ? 'border-primary bg-muted' : 'border-transparent'"
                @click="selectTemplate(template)"
              >
                <div class="flex items-center justify-between gap-2">
                  <span class="truncate font-medium">{{ getNotificationTemplateDisplayName(template) }}</span>
                  <span class="text-xs text-muted-foreground">{{ template.locale }}</span>
                </div>
                <div class="mt-1 flex items-center justify-between text-xs text-muted-foreground">
                  <span class="truncate font-mono text-[10px]">{{ template.code }}</span>
                  <span>v{{ template.version }} · {{ template.is_enabled ? '启用' : '停用' }}</span>
                </div>
              </button>
            </div>
          </section>
        </div>
      </section>

      <section v-if="selected" class="min-h-0 overflow-y-auto rounded-lg border bg-card p-6">
        <div class="flex flex-wrap items-start justify-between gap-3">
          <div>
            <h2 class="text-lg font-semibold">{{ getNotificationTemplateDisplayName(selected) }}</h2>
            <p class="text-sm text-muted-foreground">{{ selected.code }} / {{ selected.locale }} · 当前版本 v{{ selected.version }} · {{ selected.category }}</p>
          </div>
          <div class="flex gap-2">
            <Button :disabled="saving" @click="saveTemplate">
              <LoaderCircle v-if="saving" class="mr-2 size-4 animate-spin" />
              <Save v-else class="mr-2 size-4" />保存
            </Button>
          </div>
        </div>

        <div class="grid gap-4 lg:grid-cols-[minmax(0,1fr)_auto_minmax(0,1.5fr)]">
          <label class="space-y-1 text-sm"><span>名称</span><Input v-model="draft.name" /></label>
          <label class="flex items-center gap-2 whitespace-nowrap pt-6 text-sm"><input v-model="draft.is_enabled" type="checkbox" />启用模板</label>
          <label class="block space-y-1 text-sm">
            <span>邮件标题</span>
            <NotificationTemplateSubjectEditor
              v-model="draft.subject_template"
              :template-variable-display-values="notificationTemplateVariableDisplayValues"
            />
          </label>
        </div>
        <div class="grid items-stretch gap-4 xl:grid-cols-2">
          <div class="space-y-2">
            <div>
              <span class="block text-sm font-medium">邮件正文</span>
            </div>
            <RichTextEditor
              class="h-[28rem]"
              v-model="draft.body_html"
              :template-variable-display-values="notificationTemplateVariableDisplayValues"
            />
          </div>
          <div class="space-y-2">
            <div>
              <span class="block text-sm font-medium">实时预览</span>
            </div>
            <div class="flex h-[28rem] min-h-0 flex-col overflow-hidden rounded-lg border bg-muted/15 p-3">
              <div class="mb-3 rounded-md border bg-background px-3 py-2">
                <div class="text-[10px] font-medium uppercase tracking-wider text-muted-foreground">邮件标题</div>
                <div class="mt-1 text-sm font-medium">{{ liveNotificationTemplatePreviewSubject }}</div>
              </div>
              <iframe
                :srcdoc="liveNotificationTemplatePreviewDocument"
                title="邮件正文实时预览"
                class="min-h-0 w-full flex-1 rounded-md border bg-white"
                sandbox="allow-same-origin"
              />
            </div>
          </div>
        </div>
        <label class="block space-y-1 text-sm"><span>变更原因</span><Input v-model="draft.change_reason" placeholder="例如：更新售后寄回说明" /></label>

        <div v-if="versions.length" class="border-t pt-4">
          <button
            type="button"
            class="flex w-full items-center justify-between gap-2 text-left text-sm font-medium"
            :aria-expanded="versionsHistoryExpanded"
            @click="versionsHistoryExpanded = !versionsHistoryExpanded"
          >
            <span>版本历史</span>
            <span class="flex items-center gap-2 text-xs text-muted-foreground">
              <span>{{ versions.length }}</span>
              <ChevronDown class="size-4 transition-transform" :class="versionsHistoryExpanded ? 'rotate-180' : ''" />
            </span>
          </button>
          <div v-if="versionsHistoryExpanded" class="mt-2 space-y-1 text-sm">
            <div v-for="version in versions" :key="version.id" class="flex justify-between rounded bg-muted/50 px-3 py-2">
              <span class="min-w-0 truncate">v{{ version.version }} {{ version.change_reason || '无说明' }}</span>
              <span class="flex shrink-0 items-center gap-2 text-muted-foreground">
                <span>{{ formatDate(version.created_at) }}</span>
                <Button
                  v-if="version.version < selected.version"
                  size="icon"
                  variant="ghost"
                  class="size-7"
                  :disabled="rollingBack"
                  title="回滚为新版本"
                  @click.stop="rollbackVersion(version.version)"
                >
                  <RotateCcw class="size-3.5" :class="rollingBack ? 'animate-spin' : ''" />
                </Button>
              </span>
            </div>
          </div>
        </div>
      </section>
      <section v-else class="rounded-lg border bg-card p-10 text-center text-muted-foreground">请选择一个模板</section>
    </div>

  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { toast } from 'vue-sonner'
import { ChevronDown, LoaderCircle, RefreshCw, RotateCcw, Save } from '@lucide/vue'
import AdminPageHeader from '@/components/admin/AdminPageHeader.vue'
import NotificationTemplateSubjectEditor from '@/components/admin/settings/NotificationTemplateSubjectEditor.vue'
import RichTextEditor from '@/components/admin/settings/RichTextEditor.vue'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { useAdminI18n } from '@/i18n'
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

  templates.value.forEach((template) => {
    groupsByKey.get(getNotificationTemplateGroupKey(template))?.templates.push(template)
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
  return `<!doctype html><html><head><meta charset="utf-8"><style>@font-face{font-family:MapleUICJK;src:url('${notificationTemplatePreviewFontURL}') format('woff2');font-weight:100 900;font-style:normal;font-display:block}body{margin:0;padding:18px;color:#172033;font:14px/1.65 MapleUICJK}p{margin:0 0 12px}h1,h2,h3{margin:0 0 12px}a{color:#2563eb}blockquote{border-left:3px solid #94a3b8;margin:0 0 12px;padding-left:12px;color:#64748b}</style></head><body>${previewBody || '<p style="color:#94a3b8">暂无正文</p>'}</body></html>`
})

const applyDraft = (template: NotificationTemplate) => {
  draft.name = template.name; draft.subject_template = template.subject_template; draft.body_html = template.body_html; draft.is_enabled = template.is_enabled; draft.change_reason = ''
}

const loadTemplates = async () => {
  loading.value = true
  try { templates.value = await notificationTemplatesApi.list(); if (!selected.value && templates.value.length) await selectTemplate(templates.value[0]) } catch (error) { toast.error(error instanceof Error ? error.message : '模板加载失败') } finally { loading.value = false }
}

const selectTemplate = async (template: NotificationTemplate) => {
  selected.value = template
  expandedNotificationTemplateGroupKey.value = getNotificationTemplateGroupKey(template)
  versionsHistoryExpanded.value = false
  applyDraft(template)
  try { versions.value = await notificationTemplatesApi.versions(template.id) } catch (error) { toast.error(error instanceof Error ? error.message : '版本历史加载失败') }
}

const saveTemplate = async () => {
  if (!selected.value) return
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
    templates.value = templates.value.map((item) => item.id === updated.id ? updated : item)
    await selectTemplate(updated)
    toast.success('模板已保存')
  } catch (error) {
    toast.error(error instanceof Error ? error.message : '模板保存失败')
  } finally {
    saving.value = false
  }
}

const rollbackVersion = async (version: number) => {
  if (!selected.value || !window.confirm(`确定将模板回滚到 v${version} 吗？这会生成一个新的当前版本。`)) return
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
