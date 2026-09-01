<template>
  <div class="space-y-4">
    <AdminPageHeader title="博客内容" description="管理前台 Blog 文章、发布状态和多语言版本">
      <template #actions>
        <Button v-if="hasPermission('content:create')" @click="showCreateDialog">
          <Plus class="size-4" />
          添加文章
        </Button>
      </template>
    </AdminPageHeader>

    <AdminStatsGrid :items="statItems" />

    <ContentFilterPanel
      :filters="filters"
      :status-filter-options="statusFilterOptions"
      :locale-filter-options="localeFilterOptions"
      @apply="applyFilters"
      @reset="resetFilters"
    />

    <ContentTablePanel
      :loading="loading"
      :posts="posts"
      :selected-posts="selectedPosts"
      :pagination="pagination"
      :selection-state="selectionState"
      :can-edit="hasPermission('content:edit')"
      :can-delete="hasPermission('content:delete')"
      :get-status-name="getStatusName"
      :status-tone="statusTone"
      :locale-name="localeName"
      :format-date="formatDate"
      @batch-status="requestBatchStatus"
      @batch-delete="requestBatchDelete"
      @toggle-all-posts="toggleAllPosts"
      @toggle-post="togglePost"
      @edit="showEditDialog"
      @translations="showTranslationsDialog"
      @toggle-status="requestToggleStatus"
      @delete="requestDelete"
      @update-page="updatePage"
      @update-page-size="updatePageSize"
    />

    <ContentEditorDialog
      v-model:open="dialogVisible"
      :mode="dialogMode"
      :form="postForm"
      :errors="formErrors"
      :submitting="submitting"
      :language-options="languageOptions"
      @submit="submitForm"
      @clear-error="clearFieldError"
    />

    <ContentTranslationsDialog
      v-model:open="translationsDialogVisible"
      :current-post="currentPost"
      :translations-loading="translationsLoading"
      :translations="translations"
      :locale-name="localeName"
      :status-tone="statusTone"
      :get-status-name="getStatusName"
      @edit="editTranslation"
    />

    <AdminConfirmDialog
      v-model:open="confirmation.open"
      :title="confirmation.title"
      :description="confirmation.description"
      :confirm-label="confirmation.confirmLabel"
      :destructive="confirmation.destructive"
      @confirm="executeConfirmedAction"
    />
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { toast } from 'vue-sonner'
import { Eye, FilePenLine, FileText, Plus, Send } from '@lucide/vue'
import AdminConfirmDialog from '@/components/admin/AdminConfirmDialog.vue'
import ContentEditorDialog from '@/components/admin/content/ContentEditorDialog.vue'
import ContentFilterPanel from '@/components/admin/content/ContentFilterPanel.vue'
import ContentTablePanel from '@/components/admin/content/ContentTablePanel.vue'
import ContentTranslationsDialog from '@/components/admin/content/ContentTranslationsDialog.vue'
import type {
  ContentBadgeTone,
  ContentConfirmation,
  ContentDialogMode,
  ContentFilters,
  ContentFormErrors,
  ContentListResponse,
  ContentPagination,
  ContentPost,
  ContentPostForm,
  ContentPostId,
  ContentPostPayload,
  ContentSelectionState,
  ContentStats,
  ContentStatus,
  ContentTranslationsResponse
} from '@/modules/content/contentTypes'
import AdminPageHeader from '@/components/admin/AdminPageHeader.vue'
import AdminStatsGrid from '@/components/admin/AdminStatsGrid.vue'
import { Button } from '@/components/ui/button'
import { useSupportedLanguages } from '@/composables/useSupportedLanguages'
import { useAuthStore } from '@/stores/auth'
import axios from '@/utils/axios'

const authStore = useAuthStore()
const loading = ref(false)
const posts = ref<ContentPost[]>([])
const selectedPosts = ref<ContentPost[]>([])
const dialogVisible = ref(false)
const translationsDialogVisible = ref(false)
const translationsLoading = ref(false)
const dialogMode = ref<ContentDialogMode>('create')
const submitting = ref(false)
const currentPost = ref<ContentPost | null>(null)
const translations = ref<ContentPost[]>([])
const stats = ref<ContentStats>({})
const formErrors = reactive<ContentFormErrors>({})
const supportedLanguages = useSupportedLanguages()
const languageOptions = supportedLanguages.languageOptions
const resolveDefaultLocale = (): string => supportedLanguages.defaultLocale.value || ''

const filters = reactive<ContentFilters>({ search: '', status: 'all', locale: 'all' })
const pagination = reactive<ContentPagination>({ page: 1, pageSize: 20, total: 0 })
const postForm = reactive<ContentPostForm>({
  id: null,
  title: '',
  slug: '',
  content: '',
  excerpt: '',
  status: 'draft',
  locale: resolveDefaultLocale(),
  featured_image: '',
  tags: '',
  translation_group_id: null
})
const confirmation = reactive<ContentConfirmation>({
  open: false,
  type: '',
  target: null,
  status: '',
  title: '',
  description: '',
  confirmLabel: '确定',
  destructive: false
})

const statusFilterOptions = [
  { label: '全部状态', value: 'all' },
  { label: '草稿', value: 'draft' },
  { label: '已发布', value: 'published' },
  { label: '已归档', value: 'archived' }
]
const localeFilterOptions = supportedLanguages.localeFilterOptions

const statItems = computed(() => [
  { key: 'total', label: '总文章数', value: stats.value.total || 0, icon: FileText, tone: 'gray' },
  { key: 'published', label: '已发布', value: stats.value.published || 0, icon: Send, tone: 'green' },
  { key: 'draft', label: '草稿', value: stats.value.draft || 0, icon: FilePenLine, tone: 'amber' },
  { key: 'views', label: '总浏览量', value: Number(stats.value.total_views || 0).toLocaleString('zh-CN'), icon: Eye, tone: 'blue' }
])
const selectionState = computed<ContentSelectionState>(() => {
  if (posts.value.length === 0 || selectedPosts.value.length === 0) return false
  return selectedPosts.value.length === posts.value.length ? true : 'indeterminate'
})

const statusNames: Record<string, string> = { draft: '草稿', published: '已发布', archived: '已归档' }
const statusTones: Record<string, ContentBadgeTone> = { draft: 'gray', published: 'green', archived: 'amber' }

const hasPermission = (permission: string): boolean => authStore.hasPermission(permission)
const getStatusName = (status?: string | null): string => statusNames[status || ''] || status || '-'
const statusTone = (status?: string | null): ContentBadgeTone => statusTones[status || ''] || 'gray'
const localeName = supportedLanguages.localeName
const formatDate = (dateString?: string | null): string => dateString ? new Date(dateString).toLocaleString('zh-CN') : '-'

const contentStatuses: ContentStatus[] = ['draft', 'published', 'archived']
const normalizeContentStatus = (status?: string | null): ContentStatus => (
  contentStatuses.includes(status as ContentStatus) ? status as ContentStatus : 'draft'
)
const isContentPost = (target: ContentConfirmation['target']): target is ContentPost => (
  Boolean(target) && !Array.isArray(target)
)

const clearFormErrors = (): void => Object.keys(formErrors).forEach((key) => delete formErrors[key])
const clearFieldError = (field: string): void => { delete formErrors[field] }
const buildPostPayload = (): ContentPostPayload => ({
  title: postForm.title.trim(),
  slug: postForm.slug.trim(),
  content: postForm.content,
  excerpt: postForm.excerpt,
  status: postForm.status,
  locale: postForm.locale,
  featured_image: postForm.featured_image.trim(),
  tags: postForm.tags,
  translation_group_id: postForm.translation_group_id
})
const validateForm = (payload: ContentPostPayload): boolean => {
  clearFormErrors()
  if (!payload.title) formErrors.title = '请输入文章标题'
  if (!payload.slug) formErrors.slug = '请输入 URL slug'
  if (!payload.locale) formErrors.locale = '请选择语言'
  if (Object.keys(formErrors).length > 0) {
    toast.error('请检查文章表单中的必填项')
    return false
  }
  return true
}
const resetForm = (): void => {
  Object.assign(postForm, {
    id: null,
    title: '',
    slug: '',
    content: '',
    excerpt: '',
    status: 'draft',
    locale: resolveDefaultLocale(),
    featured_image: '',
    tags: '',
    translation_group_id: null
  })
  clearFormErrors()
}

const buildFilterParams = (): Record<string, string> => ({
  ...(filters.search.trim() ? { search: filters.search.trim() } : {}),
  ...(filters.status !== 'all' ? { status: filters.status } : {}),
  ...(filters.locale !== 'all' ? { locale: filters.locale } : {})
})
const fetchStats = async (): Promise<void> => {
  try {
    const response = await axios.get<ContentStats>('/api/admin/content/posts/stats')
    stats.value = response.data || {}
  } catch (error) {
    console.error('Failed to fetch content stats:', error)
  }
}
const fetchPosts = async (): Promise<void> => {
  loading.value = true
  try {
    const response = await axios.get<ContentListResponse>('/api/admin/content/posts', {
      params: { page: pagination.page, page_size: pagination.pageSize, ...buildFilterParams() }
    })
    posts.value = response.data.posts || []
    pagination.total = response.data.pagination?.total || 0
    selectedPosts.value = []
  } catch (error) {
    console.error('Failed to fetch posts:', error)
  } finally {
    loading.value = false
  }
}
const refreshContent = async (): Promise<void> => {
  await Promise.all([fetchPosts(), fetchStats()])
}
const applyFilters = (): void => {
  pagination.page = 1
  void fetchPosts()
}
const resetFilters = (): void => {
  Object.assign(filters, { search: '', status: 'all', locale: 'all' })
  pagination.page = 1
  void fetchPosts()
}
const updatePage = (page: number): void => {
  pagination.page = page
  void fetchPosts()
}
const updatePageSize = (pageSize: number): void => {
  pagination.pageSize = pageSize
  pagination.page = 1
  void fetchPosts()
}

const showCreateDialog = (): void => {
  dialogMode.value = 'create'
  resetForm()
  dialogVisible.value = true
}
const showEditDialog = (post: ContentPost): void => {
  dialogMode.value = 'edit'
  Object.assign(postForm, {
    id: post.id,
    title: post.title || '',
    slug: post.slug || '',
    content: post.content || '',
    excerpt: post.excerpt || '',
    status: normalizeContentStatus(post.status),
    locale: post.locale || resolveDefaultLocale(),
    featured_image: post.featured_image || '',
    tags: post.tags || '',
    translation_group_id: post.translation_group_id || null
  })
  clearFormErrors()
  dialogVisible.value = true
}
const submitForm = async (): Promise<void> => {
  const payload = buildPostPayload()
  if (!validateForm(payload)) return
  submitting.value = true
  try {
    if (dialogMode.value === 'create') {
      await axios.post('/api/admin/content/posts', payload)
      toast.success('文章创建成功')
    } else if (postForm.id !== null) {
      await axios.put(`/api/admin/content/posts/${postForm.id}`, payload)
      toast.success('文章更新成功')
    }
    dialogVisible.value = false
    await refreshContent()
  } catch (error) {
    console.error('Failed to save post:', error)
  } finally {
    submitting.value = false
  }
}

const showTranslationsDialog = async (post: ContentPost): Promise<void> => {
  currentPost.value = post
  translations.value = []
  translationsDialogVisible.value = true
  translationsLoading.value = true
  try {
    const response = await axios.get<ContentTranslationsResponse>(`/api/admin/content/posts/${post.id}/translations`)
    if (!Array.isArray(response.data?.translations)) throw new Error('Missing translations array in response')
    translations.value = response.data.translations
  } catch (error) {
    console.error('Failed to fetch translations:', error)
    translationsDialogVisible.value = false
  } finally {
    translationsLoading.value = false
  }
}
const editTranslation = (translation: ContentPost): void => {
  translationsDialogVisible.value = false
  showEditDialog(translation)
}

const isSelected = (postId: ContentPostId): boolean => selectedPosts.value.some((post) => post.id === postId)
const toggleAllPosts = (checked: ContentSelectionState): void => { selectedPosts.value = checked === true ? [...posts.value] : [] }
const togglePost = (post: ContentPost, checked: ContentSelectionState): void => {
  if (checked === true && !isSelected(post.id)) selectedPosts.value = [...selectedPosts.value, post]
  else if (checked !== true) selectedPosts.value = selectedPosts.value.filter((selected) => selected.id !== post.id)
}
const setConfirmation = (values: Partial<ContentConfirmation>): void => {
  Object.assign(confirmation, {
    open: true,
    type: '',
    target: null,
    status: '',
    confirmLabel: '确定',
    destructive: false,
    ...values
  })
}
const requestToggleStatus = (post: ContentPost): void => {
  const status = post.status === 'published' ? 'draft' : 'published'
  const action = status === 'published' ? '发布' : '转为草稿'
  setConfirmation({
    type: 'status', target: post, status, title: `${action}文章？`,
    description: `文章“${post.title}”将被${action}。`, confirmLabel: action
  })
}
const requestDelete = (post: ContentPost): void => setConfirmation({
  type: 'delete', target: post, title: '删除文章？',
  description: `文章“${post.title}”将被永久删除，此操作不可恢复。`, confirmLabel: '删除', destructive: true
})
const requestBatchStatus = (status: ContentStatus): void => {
  const action = status === 'published' ? '发布' : '转为草稿'
  setConfirmation({
    type: 'batch-status', target: [...selectedPosts.value], status, title: `批量${action}文章？`,
    description: `将 ${selectedPosts.value.length} 篇文章批量${action}。`, confirmLabel: `批量${action}`
  })
}
const requestBatchDelete = (): void => setConfirmation({
  type: 'batch-delete', target: [...selectedPosts.value], title: '批量删除文章？',
  description: `${selectedPosts.value.length} 篇文章将被永久删除，此操作不可恢复。`, confirmLabel: '批量删除', destructive: true
})
const executeConfirmedAction = async (): Promise<void> => {
  const { type, target, status } = confirmation
  confirmation.open = false
  try {
    if (type === 'status' && isContentPost(target)) {
      await axios.patch(`/api/admin/content/posts/${target.id}/status`, { status })
      toast.success(status === 'published' ? '文章已发布' : '文章已转为草稿')
    } else if (type === 'delete' && isContentPost(target)) {
      await axios.delete(`/api/admin/content/posts/${target.id}`)
      toast.success('文章已删除')
    } else if (type === 'batch-status' && Array.isArray(target)) {
      await axios.post('/api/admin/content/posts/batch-status', { post_ids: target.map((post) => post.id), status })
      toast.success(status === 'published' ? '文章已批量发布' : '文章已批量转为草稿')
    } else if (type === 'batch-delete' && Array.isArray(target)) {
      await axios.post('/api/admin/content/posts/batch-delete', { post_ids: target.map((post) => post.id) })
      toast.success('文章已批量删除')
    }
    await refreshContent()
  } catch (error) {
    console.error('Failed to update posts:', error)
  }
}

onMounted(() => {
  void Promise.all([supportedLanguages.fetchLanguages(), fetchStats(), fetchPosts()])
})
</script>

