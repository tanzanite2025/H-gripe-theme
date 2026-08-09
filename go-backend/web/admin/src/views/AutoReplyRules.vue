<template>
  <div class="flex h-full min-h-0 flex-col gap-4 overflow-hidden">
    <AdminPageHeader
      class="shrink-0"
      title="自动回复"
      description="管理欢迎语、关键词规则和结构化客服回复"
    >
      <template #actions>
        <Button variant="outline" size="sm" :disabled="loading" @click="loadRules">
          <RefreshCw :class="['size-3.5', { 'animate-spin': loading }]" />
          刷新
        </Button>
        <Button v-if="canEdit" size="sm" @click="startCreate">
          <Plus class="size-3.5" />
          新建规则
        </Button>
      </template>
    </AdminPageHeader>

    <div class="grid min-h-0 flex-1 gap-4 overflow-hidden xl:grid-cols-[minmax(0,2fr)_minmax(0,3fr)]">
      <Card class="min-h-0 overflow-hidden">
        <CardHeader class="shrink-0 border-b px-4 py-3">
          <div class="flex flex-wrap items-center justify-between gap-3">
            <div>
              <CardTitle class="text-sm font-black">规则列表</CardTitle>
              <CardDescription class="mt-1 text-xs">
                后端负责匹配、冷却和消息落库，前台只负责展示。
              </CardDescription>
            </div>
            <select
              v-model="typeFilter"
              class="h-8 rounded-xl border border-border bg-muted/50 px-2 text-xs font-bold outline-none focus:ring-2 focus:ring-ring/50"
              aria-label="筛选规则类型"
            >
              <option value="all">全部类型</option>
              <option value="welcome">欢迎语</option>
              <option value="keyword">关键词</option>
            </select>
          </div>
        </CardHeader>

        <CardContent class="min-h-0 overflow-y-auto p-0">
          <div v-if="loading" class="flex min-h-48 items-center justify-center text-xs text-muted-foreground">
            正在加载规则
          </div>
          <div v-else-if="filteredRules.length === 0" class="flex min-h-48 flex-col items-center justify-center gap-2 text-muted-foreground">
            <Bot class="size-7 opacity-40" />
            <span class="text-xs">暂无自动回复规则</span>
          </div>
          <div v-else class="divide-y divide-border/70">
            <article
              v-for="rule in filteredRules"
              :key="rule.id"
              class="flex flex-col gap-3 px-4 py-4 transition-colors hover:bg-muted/30 lg:flex-row lg:items-start lg:justify-between"
            >
              <button
                type="button"
                class="min-w-0 flex-1 text-left"
                :class="{ 'opacity-55': !rule.is_active }"
                @click="editRule(rule)"
              >
                <div class="flex flex-wrap items-center gap-2">
                  <span class="font-mono text-[10px] font-black text-muted-foreground">#{{ rule.id }}</span>
                  <Badge variant="outline">{{ rule.type === 'welcome' ? '欢迎语' : '关键词' }}</Badge>
                  <Badge variant="outline">{{ rule.message_type || 'text' }}</Badge>
                  <Badge :variant="rule.is_active ? 'default' : 'secondary'">
                    {{ rule.is_active ? '已启用' : '已停用' }}
                  </Badge>
                </div>
                <h2 class="mt-2 truncate text-sm font-black text-foreground">
                  {{ rule.type === 'keyword' ? rule.trigger_keyword : '进入客服会话' }}
                </h2>
                <p class="mt-1 line-clamp-2 whitespace-pre-wrap break-words text-xs leading-5 text-muted-foreground">
                  {{ rule.reply_message }}
                </p>
                <div class="mt-2 flex flex-wrap gap-x-3 gap-y-1 text-[10px] font-bold text-muted-foreground">
                  <span>语言：{{ localeLabel(rule.locale) }}</span>
                  <span>优先级：{{ rule.priority || 0 }}</span>
                  <span>冷却：{{ formatCooldown(rule.cooldown_seconds) }}</span>
                  <span v-if="rule.agent_id">客服：{{ rule.agent_id }}</span>
                  <span v-if="rule.group_id">客服组：{{ groupName(rule.group_id) }}</span>
                </div>
              </button>

              <div v-if="canEdit || canDelete" class="flex shrink-0 items-center gap-2">
                <Button v-if="canEdit" variant="outline" size="xs" @click="editRule(rule)">
                  <Pencil class="size-3" />
                  编辑
                </Button>
                <Button v-if="canDelete" variant="destructive" size="xs" @click="removeRule(rule)">
                  <Trash2 class="size-3" />
                  删除
                </Button>
              </div>
            </article>
          </div>
        </CardContent>
      </Card>

      <Card class="min-h-0 overflow-y-auto">
        <CardHeader class="border-b px-4 py-3">
          <CardTitle class="text-sm font-black">
            {{ editingId ? '编辑自动回复' : '新建自动回复' }}
          </CardTitle>
          <CardDescription class="mt-1 text-xs">
            回复内容先使用安全结构化字段，避免把链接和图片编码进文本。
          </CardDescription>
        </CardHeader>

        <CardContent class="space-y-4 p-4">
          <div class="grid grid-cols-2 gap-3">
            <label class="space-y-1.5">
              <span class="text-[10px] font-black uppercase tracking-wider text-muted-foreground">触发类型</span>
              <select v-model="form.type" class="admin-auto-reply-input">
                <option value="welcome">欢迎语</option>
                <option value="keyword">关键词</option>
              </select>
            </label>
            <div class="space-y-1.5">
              <span class="text-[10px] font-black uppercase tracking-wider text-muted-foreground">保存类型</span>
              <div class="admin-auto-reply-type-preview">
                <component :is="detectedMessageTypeIcon" class="size-3.5" />
                {{ detectedMessageTypeLabel }}
              </div>
            </div>
          </div>

          <label v-if="form.type === 'keyword'" class="block space-y-1.5">
            <span class="text-[10px] font-black uppercase tracking-wider text-muted-foreground">触发关键词</span>
            <Input v-model="form.trigger_keyword" placeholder="例如：shipping" />
          </label>

          <div class="grid grid-cols-2 gap-3">
            <div class="space-y-1.5">
              <span class="text-[10px] font-black uppercase tracking-wider text-muted-foreground">语言</span>
              <StorefrontLocaleSelect
                v-model="form.locale"
                :language-options="languageOptions"
                :loading="languageLoading"
                placeholder="请选择一种语言"
              />
            </div>
            <label class="space-y-1.5">
              <span class="text-[10px] font-black uppercase tracking-wider text-muted-foreground">客服用户 ID</span>
              <select v-model="form.agent_id" class="admin-auto-reply-input">
                <option value="">全部客服</option>
                <option v-for="agent in agents" :key="agent.id" :value="String(agent.id)">
                  {{ agent.name || agent.email || `客服 #${agent.id}` }}
                </option>
              </select>
            </label>
            <label class="space-y-1.5">
              <span class="text-[10px] font-black uppercase tracking-wider text-muted-foreground">客服组</span>
              <select v-model="form.group_id" class="admin-auto-reply-input">
                <option value="">全部客服组</option>
                <option v-for="group in groups" :key="group.id" :value="String(group.id)">
                  {{ group.name }}
                </option>
              </select>
            </label>
          </div>

          <div v-if="form.type === 'keyword'" class="grid grid-cols-2 gap-3">
            <label class="space-y-1.5">
              <span class="text-[10px] font-black uppercase tracking-wider text-muted-foreground">匹配方式</span>
              <select v-model="form.match_type" class="admin-auto-reply-input">
                <option value="exact">完全匹配</option>
                <option value="contains">包含匹配</option>
              </select>
            </label>
            <label class="space-y-1.5">
              <span class="text-[10px] font-black uppercase tracking-wider text-muted-foreground">优先级</span>
              <Input v-model.number="form.priority" type="number" min="0" />
            </label>
          </div>

          <label class="block space-y-1.5">
            <span class="text-[10px] font-black uppercase tracking-wider text-muted-foreground">回复文本</span>
            <Textarea v-model="form.reply_message" class="min-h-24 resize-y" placeholder="客户最终看到的安全文本内容" />
          </label>

          <div class="space-y-2">
            <div class="flex flex-wrap items-center justify-between gap-2">
              <div>
                <span class="text-[10px] font-black uppercase tracking-wider text-muted-foreground">引用 FAQ</span>
                <p class="mt-1 text-[11px] font-medium text-muted-foreground">
                  前台会按 FAQ 卡片展示，并可直接打开对应答案。
                </p>
              </div>
              <Button variant="outline" size="xs" type="button" :disabled="!faqPickerLocale" @click="faqPickerOpen = true">
                <HelpCircle class="size-3.5" />
                {{ faqMetadata.faq_id ? '更换 FAQ' : '选择 FAQ' }}
              </Button>
              <Button v-if="faqMetadata.faq_id" variant="ghost" size="xs" type="button" @click="clearFAQ">
                <X class="size-3.5" />
                移除 FAQ
              </Button>
            </div>

            <div
              v-if="faqMetadata.faq_id"
              class="rounded-2xl border border-[var(--admin-selected)]/35 bg-[var(--admin-selected)]/5 p-3"
            >
              <div class="flex items-start gap-3">
                <div class="flex size-9 shrink-0 items-center justify-center rounded-full bg-[var(--admin-selected)]/10 text-[var(--admin-selected)]">
                  <HelpCircle class="size-4" />
                </div>
                <div class="min-w-0 flex-1">
                  <p class="text-[10px] font-black uppercase tracking-wider text-[var(--admin-selected)]">
                    {{ faqMetadata.page_title || faqMetadata.page_id || 'FAQ' }}
                  </p>
                  <h3 class="mt-1 break-words text-sm font-black text-foreground">
                    {{ faqMetadata.question }}
                  </h3>
                  <p v-if="faqMetadata.category_label || faqMetadata.category" class="mt-1 text-[11px] font-bold text-muted-foreground">
                    {{ faqMetadata.category_label || faqMetadata.category }}
                  </p>
                  <p v-if="faqMetadata.answer_excerpt" class="mt-2 line-clamp-3 text-xs leading-5 text-muted-foreground">
                    {{ faqMetadata.answer_excerpt }}
                  </p>
                  <p class="mt-2 break-all font-mono text-[10px] text-muted-foreground/80">
                    {{ faqMetadata.url }}
                  </p>
                </div>
              </div>
            </div>

            <div v-else class="rounded-2xl border border-dashed border-border bg-muted/20 px-3 py-4 text-center text-xs font-bold text-muted-foreground">
              尚未选择 FAQ
            </div>
          </div>

          <label class="block space-y-1.5">
            <span class="text-[10px] font-black uppercase tracking-wider text-muted-foreground">结构化 metadata JSON</span>
            <Textarea
              v-model="form.metadata"
              class="min-h-24 resize-y font-mono text-[11px]"
              :placeholder="metadataPlaceholder"
            />
          </label>

          <div class="space-y-2">
            <div class="flex flex-wrap items-center justify-between gap-2">
              <span class="text-[10px] font-black uppercase tracking-wider text-muted-foreground">图片附件</span>
              <Button variant="outline" size="xs" type="button" @click="imagePickerOpen = true">
                <ImagePlus class="size-3" />
                选择内部图片
              </Button>
            </div>

            <div v-if="attachmentUrls.length === 0" class="rounded-xl border border-dashed border-border bg-muted/30 px-3 py-4 text-center text-xs font-bold text-muted-foreground">
              尚未选择图片附件
            </div>
            <div v-else class="space-y-2">
              <div
                v-for="url in attachmentUrls"
                :key="url"
                class="flex min-w-0 items-center gap-3 rounded-xl border bg-background p-2"
              >
                <img :src="url" alt="" class="size-14 shrink-0 rounded-lg border object-cover" />
                <span class="min-w-0 flex-1 truncate font-mono text-[11px] text-muted-foreground">{{ url }}</span>
                <Button variant="ghost" size="icon-sm" type="button" aria-label="移除图片附件" @click="removeAttachmentUrl(url)">
                  <X class="size-3.5" />
                </Button>
              </div>
            </div>
          </div>

          <div class="grid grid-cols-2 gap-3">
            <label class="space-y-1.5">
              <span class="text-[10px] font-black uppercase tracking-wider text-muted-foreground">冷却秒数</span>
              <Input v-model.number="form.cooldown_seconds" type="number" min="0" />
            </label>
            <label class="flex items-end gap-2 pb-2">
              <input v-model="form.is_active" type="checkbox" class="size-4 accent-[var(--admin-selected)]" />
              <span class="text-xs font-black">启用规则</span>
            </label>
          </div>

          <p v-if="errorMessage" class="rounded-xl border border-rose-500/20 bg-rose-500/5 px-3 py-2 text-xs font-bold text-rose-600">
            {{ errorMessage }}
          </p>

          <div v-if="canEdit" class="flex items-center justify-end gap-2 border-t pt-4">
            <Button variant="outline" size="sm" :disabled="saving" @click="startCreate">清空</Button>
            <Button size="sm" :disabled="saving" @click="saveRule">
              <LoaderCircle v-if="saving" class="size-3.5 animate-spin" />
              <Save v-else class="size-3.5" />
              {{ editingId ? '保存修改' : '创建规则' }}
            </Button>
          </div>
        </CardContent>
      </Card>
    </div>
  </div>

  <AutoReplyImagePickerDialog
    v-model:open="imagePickerOpen"
    :selected-urls="attachmentUrls"
    @select="addAttachmentImage"
  />

  <AutoReplyFaqPickerDialog
    v-model:open="faqPickerOpen"
    :locale="faqPickerLocale"
    :selected-faq-id="faqMetadata.faq_id"
    @select="selectFAQ"
  />
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { toast } from 'vue-sonner'
import { Bot, FileText, HelpCircle, Image as ImageIcon, ImagePlus, Link2, LoaderCircle, Package, Pencil, Plus, ReceiptText, RefreshCw, Save, Trash2, X } from '@lucide/vue'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import AdminPageHeader from '@/components/admin/AdminPageHeader.vue'
import StorefrontLocaleSelect from '@/components/admin/StorefrontLocaleSelect.vue'
import AutoReplyFaqPickerDialog from '@/components/admin/customer-service/AutoReplyFaqPickerDialog.vue'
import AutoReplyImagePickerDialog from '@/components/admin/customer-service/AutoReplyImagePickerDialog.vue'
import { useAuthStore } from '@/stores/auth'
import { useSupportedLanguages } from '@/composables/useSupportedLanguages'
import customerServiceApi from '@/api/customerService'

const authStore = useAuthStore()
const supportedLanguages = useSupportedLanguages()
const languageOptions = supportedLanguages.languageOptions
const languageLoading = supportedLanguages.loading
const fetchLanguages = supportedLanguages.fetchLanguages
const defaultReplyLocale = computed(() => (
  languageOptions.value.some((language) => language.value === 'en')
    ? 'en'
    : supportedLanguages.defaultLocale.value || languageOptions.value[0]?.value || 'en'
))
const loading = ref(false)
const saving = ref(false)
const errorMessage = ref('')
const rules = ref([])
const groups = ref([])
const agents = ref([])
const typeFilter = ref('all')
const editingId = ref(null)
const imagePickerOpen = ref(false)
const faqPickerOpen = ref(false)

const createForm = () => ({
  type: 'welcome',
  trigger_keyword: '',
  reply_message: '',
  agent_id: '',
  group_id: '',
  locale: defaultReplyLocale.value,
  message_type: 'text',
  metadata: '',
  attachments: '',
  is_active: true,
  priority: 0,
  match_type: 'exact',
  cooldown_seconds: 86400,
})

const form = reactive(createForm())
const canEdit = computed(() => authStore.hasPermission('ticket:edit'))
const canDelete = computed(() => authStore.hasPermission('ticket:delete'))
const filteredRules = computed(() => typeFilter.value === 'all'
  ? rules.value
  : rules.value.filter((rule) => rule.type === typeFilter.value))
const metadataPlaceholder = '{"url":"https://example.com","title":"Open guide"}'
const faqPickerLocale = computed(() => {
  const locale = String(form.locale || '').trim().toLowerCase().replace(/-/g, '_')
  return locale && locale !== '*' ? locale : ''
})
const parseMetadataObject = (value) => {
  if (value && typeof value === 'object' && !Array.isArray(value)) return value
  const raw = String(value || '').trim()
  if (!raw) return {}
  try {
    const parsed = JSON.parse(raw)
    return parsed && typeof parsed === 'object' && !Array.isArray(parsed) ? parsed : {}
  } catch {
    return null
  }
}
const parsedMetadata = computed(() => parseMetadataObject(form.metadata))
const faqMetadata = computed(() => {
  const metadata = parsedMetadata.value
  return metadata?.faq_id ? metadata : {}
})
const attachmentUrls = computed(() => parseAttachmentUrls(form.attachments))
const detectedMessageType = computed(() => {
  if (faqMetadata.value.faq_id) return 'faq'
  if (attachmentUrls.value.length > 0) return 'image'

  const metadata = parsedMetadata.value
  if (metadata && Object.keys(metadata).length > 0) {
    if (metadata.order_id || metadata.orderId || metadata.order_number || metadata.orderNumber) return 'order'
    if (metadata.product_id || metadata.productId || metadata.sku || metadata.thumbnail || metadata.price) return 'product'
    if (metadata.url) return 'link'
  }

  return 'text'
})
const detectedMessageTypeLabel = computed(() => ({
  text: '文本',
  link: '链接',
  image: '图片',
  product: '商品',
  order: '订单',
  faq: 'FAQ',
}[detectedMessageType.value] || '文本'))
const detectedMessageTypeIcon = computed(() => ({
  text: FileText,
  link: Link2,
  image: ImageIcon,
  product: Package,
  order: ReceiptText,
  faq: HelpCircle,
}[detectedMessageType.value] || Bot))

const parseAttachmentUrls = (value) => {
  const raw = typeof value === 'string' ? value.trim() : ''
  if (!raw) return []
  try {
    const parsed = JSON.parse(raw)
    if (!Array.isArray(parsed)) return []
    return parsed
      .map((item) => String(item || '').trim())
      .filter(Boolean)
  } catch {
    return []
  }
}

const setAttachmentUrls = (urls) => {
  const normalized = [...new Set(urls.map((url) => String(url || '').trim()).filter(Boolean))]
  form.attachments = normalized.length > 0 ? JSON.stringify(normalized) : ''
}

const addAttachmentImage = ({ url }) => {
  setAttachmentUrls([...attachmentUrls.value, url])
  if (form.message_type === 'text') {
    form.message_type = 'image'
  }
}

const removeAttachmentUrl = (url) => {
  setAttachmentUrls(attachmentUrls.value.filter((item) => item !== url))
}

const plainText = (value) => String(value || '')
  .replace(/<[^>]+>/g, ' ')
  .replace(/&nbsp;/g, ' ')
  .replace(/\s+/g, ' ')
  .trim()

const buildFAQMetadata = (page, category, faq) => {
  const answerExcerpt = plainText(faq?.answer).slice(0, 320)
  const faqID = String(faq?.id || '').trim()
  const pageID = String(page?.page_id || '').trim()
  return {
    faq_id: faqID,
    page_id: pageID,
    page_title: String(page?.title || pageID).trim(),
    route_path: String(page?.route_path || '').trim(),
    category: String(category?.category_key || '').trim(),
    category_label: String(category?.name || category?.category_key || '').trim(),
    locale: faqPickerLocale.value,
    question: String(faq?.question || '').trim(),
    answer_excerpt: answerExcerpt,
    url: `/support/faqs?page=${encodeURIComponent(pageID)}&faq=${encodeURIComponent(faqID)}`,
    answer_image_url: String(faq?.answer_image_url || '').trim(),
    answer_image_alt: String(faq?.answer_image_alt || '').trim(),
    answer_image_width: Number(faq?.answer_image_width || 0) || 0,
    answer_image_height: Number(faq?.answer_image_height || 0) || 0,
  }
}

const selectFAQ = ({ page, category, faq }) => {
  const metadata = buildFAQMetadata(page, category, faq)
  form.message_type = 'faq'
  form.metadata = JSON.stringify(metadata)
  form.reply_message = metadata.question
  errorMessage.value = ''
}

const clearFAQ = () => {
  if (!faqMetadata.value.faq_id) return
  form.metadata = ''
  form.message_type = attachmentUrls.value.length > 0 ? 'image' : 'text'
}

const loadRules = async () => {
  loading.value = true
  errorMessage.value = ''
  try {
    rules.value = await customerServiceApi.listAutoReplyRules()
  } catch (error) {
    errorMessage.value = error?.response?.data?.message || '自动回复规则加载失败'
  } finally {
    loading.value = false
  }
}

const loadGroups = async () => {
  try {
    groups.value = await customerServiceApi.listGroups()
  } catch (error) {
    console.error('客服组加载失败:', error)
    groups.value = []
  }
}

const loadAgents = async () => {
  try {
    agents.value = await customerServiceApi.listAgents()
  } catch (error) {
    console.error('客服列表加载失败:', error)
    agents.value = []
  }
}

const startCreate = () => {
  editingId.value = null
  Object.assign(form, createForm())
  errorMessage.value = ''
}

const editRule = (rule) => {
  editingId.value = rule.id
  Object.assign(form, {
    ...createForm(),
    ...rule,
    group_id: rule.group_id ? String(rule.group_id) : '',
    metadata: typeof rule.metadata === 'string' ? rule.metadata : (rule.metadata ? JSON.stringify(rule.metadata) : ''),
    attachments: typeof rule.attachments === 'string' ? rule.attachments : (rule.attachments ? JSON.stringify(rule.attachments) : ''),
  })
  errorMessage.value = ''
}

const saveRule = async () => {
  if (!canEdit.value) return
  const metadataText = String(form.metadata || '').trim()
  if (metadataText && !parsedMetadata.value) {
    errorMessage.value = 'metadata JSON 格式不正确。'
    return
  }
  saving.value = true
  errorMessage.value = ''
  try {
    const payload = {
      ...form,
      message_type: detectedMessageType.value,
      metadata: metadataText && parsedMetadata.value ? JSON.stringify(parsedMetadata.value) : '',
      group_id: form.group_id ? Number(form.group_id) : null,
      priority: Number(form.priority) || 0,
      cooldown_seconds: Number(form.cooldown_seconds) || 0,
    }
    if (editingId.value) {
      await customerServiceApi.updateAutoReplyRule(editingId.value, payload)
      toast.success('自动回复规则已更新')
    } else {
      await customerServiceApi.createAutoReplyRule(payload)
      toast.success('自动回复规则已创建')
    }
    await loadRules()
    startCreate()
  } catch (error) {
    errorMessage.value = error?.response?.data?.message || '自动回复规则保存失败'
  } finally {
    saving.value = false
  }
}

const removeRule = async (rule) => {
  if (!canDelete.value || !window.confirm(`确定删除自动回复规则 #${rule.id} 吗？`)) return
  try {
    await customerServiceApi.deleteAutoReplyRule(rule.id)
    toast.success('自动回复规则已删除')
    if (editingId.value === rule.id) startCreate()
    await loadRules()
  } catch (error) {
    errorMessage.value = error?.response?.data?.message || '自动回复规则删除失败'
  }
}

const formatCooldown = (seconds) => {
  const value = Number(seconds || 0)
  if (value >= 86400) return `${Math.floor(value / 86400)}天`
  if (value >= 3600) return `${Math.floor(value / 3600)}小时`
  if (value >= 60) return `${Math.floor(value / 60)}分钟`
  return `${value}秒`
}

const groupName = (groupID) => {
  return groups.value.find((group) => Number(group.id) === Number(groupID))?.name || `组 #${groupID}`
}

const localeLabel = (locale) => supportedLanguages.localeName(locale)

onMounted(() => {
  loadRules()
  loadGroups()
  loadAgents()
  fetchLanguages()
})
</script>

<style scoped>
.admin-auto-reply-input {
  width: 100%;
  height: 2.25rem;
  border: 0;
  border-radius: 0.75rem;
  background: hsl(var(--muted) / 0.5);
  padding-inline: 0.75rem;
  font-size: 0.75rem;
  font-weight: 700;
  outline: none;
}

.admin-auto-reply-input:focus {
  box-shadow: 0 0 0 2px hsl(var(--ring) / 0.5);
}

.admin-auto-reply-type-preview {
  display: inline-flex;
  width: 100%;
  height: 2.25rem;
  align-items: center;
  gap: 0.5rem;
  border-radius: 0.75rem;
  background: hsl(var(--muted) / 0.5);
  padding-inline: 0.75rem;
  color: hsl(var(--foreground));
  font-size: 0.75rem;
  font-weight: 800;
}
</style>
