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

    <section class="rounded-lg border bg-card p-4">
      <div class="flex flex-wrap items-start justify-between gap-3">
        <div>
          <h2 class="text-sm font-semibold">BLOG 分类管理</h2>
          <p class="mt-1 text-xs text-muted-foreground">分类由后台维护，前台仅读取并用于 Blog 页面筛选。</p>
        </div>
        <div class="flex flex-wrap items-end gap-2">
          <Input v-model="categoryDraft.name" class="w-44" placeholder="分类名称" />
          <Input v-model="categoryDraft.slug" class="w-44" placeholder="slug，例如 guides" />
          <StorefrontLocaleSelect
            v-model="categoryDraft.locale"
            :language-options="languageOptions"
            class="w-44"
          />
          <Input v-model.number="categoryDraft.sort_order" class="w-24" type="number" placeholder="排序" />
          <Button :disabled="!categoryDraft.name.trim() || !categoryDraft.slug.trim()" @click="saveCategory">
            {{ editingCategoryId ? '保存分类' : '新增分类' }}
          </Button>
          <Button v-if="editingCategoryId" variant="outline" @click="resetCategoryDraft">取消</Button>
        </div>
      </div>
      <div v-if="categories.length" class="mt-4 grid gap-2 sm:grid-cols-2 lg:grid-cols-3">
        <div v-for="category in categories" :key="category.id" class="flex items-center gap-2 rounded-md border px-3 py-2">
          <span class="min-w-0 flex-1 truncate text-sm font-medium">{{ category.name }}</span>
          <span class="text-xs text-muted-foreground">{{ category.locale }} / {{ category.slug }}</span>
          <Button variant="ghost" size="sm" @click="editCategory(category)">编辑</Button>
          <Button variant="ghost" size="sm" class="text-destructive" @click="deleteCategory(category)">删除</Button>
        </div>
      </div>
      <p v-else class="mt-4 text-sm text-muted-foreground">暂无分类。</p>
    </section>

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
      :categories="categoryOptions"
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
  BlogCategory,
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
import { Input } from '@/components/ui/input'
import StorefrontLocaleSelect from '@/components/admin/StorefrontLocaleSelect.vue'
import { useSupportedLanguages } from '@/composables/useSupportedLanguages'
import { useAuthStore } from '@/stores/auth'
import axios from '@/utils/axios'

const authStore = useAuthStore()
const loading = ref(false)
const posts = ref<ContentPost[]>([])
const categories = ref<BlogCategory[]>([])
const editingCategoryId = ref<number | null>(null)
const categoryDraft = reactive({
  name: '',
  slug: '',
  description: '',
  locale: 'en',
  sort_order: 0
})
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
  category_ids: [],
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
const categoryOptions = computed(() => categories.value.filter((category) => category.locale === postForm.locale))

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
  category_ids: postForm.category_ids,
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
    category_ids: [],
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
const fetchCategories = async (): Promise<void> => {
  try {
    const response = await axios.get<{ categories?: BlogCategory[] }>('/api/admin/content/categories')
    categories.value = response.data.categories || []
  } catch (error) {
    console.error('Failed to fetch blog categories:', error)
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
  await Promise.all([fetchPosts(), fetchStats(), fetchCategories()])
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
    category_ids: (post.categories || []).map((category) => category.id),
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

const resetCategoryDraft = (): void => {
  editingCategoryId.value = null
  Object.assign(categoryDraft, {
    name: '',
    slug: '',
    description: '',
    locale: resolveDefaultLocale() || 'en',
    sort_order: 0
  })
}

const editCategory = (category: BlogCategory): void => {
  editingCategoryId.value = category.id
  Object.assign(categoryDraft, {
    name: category.name,
    slug: category.slug,
    description: category.description || '',
    locale: category.locale || resolveDefaultLocale() || 'en',
    sort_order: category.sort_order || 0
  })
}

const saveCategory = async (): Promise<void> => {
  const payload = {
    ...categoryDraft,
    name: categoryDraft.name.trim(),
    slug: categoryDraft.slug.trim()
  }
  try {
    if (editingCategoryId.value) {
      await axios.put(`/api/admin/content/categories/${editingCategoryId.value}`, payload)
      toast.success('分类已更新')
    } else {
      await axios.post('/api/admin/content/categories', payload)
      toast.success('分类已创建')
    }
    resetCategoryDraft()
    await fetchCategories()
  } catch (error) {
    console.error('Failed to save blog category:', error)
  }
}

const deleteCategory = async (category: BlogCategory): Promise<void> => {
  if (!window.confirm(`确定删除分类“${category.name}”？`)) return
  try {
    await axios.delete(`/api/admin/content/categories/${category.id}`)
    toast.success('分类已删除')
    await fetchCategories()
  } catch (error) {
    console.error('Failed to delete blog category:', error)
  }
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
  void Promise.all([supportedLanguages.fetchLanguages(), fetchStats(), fetchPosts(), fetchCategories()])
})
</script>

