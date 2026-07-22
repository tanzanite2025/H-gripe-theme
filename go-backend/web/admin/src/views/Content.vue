<template>
  <div class="space-y-4">
    <AdminPageHeader title="内容管理" description="管理文章、发布状态、多语言版本和搜索信息">
      <template #actions>
        <Button v-if="hasPermission('content:create')" @click="showCreateDialog">
          <Plus class="size-4" />
          添加文章
        </Button>
      </template>
    </AdminPageHeader>

    <AdminStatsGrid :items="statItems" />

    <AdminFilterPanel>
      <form class="grid grid-cols-1 gap-3 md:grid-cols-2 xl:grid-cols-[minmax(240px,1.5fr)_repeat(2,minmax(140px,0.7fr))_auto]" @submit.prevent="applyFilters">
        <label class="space-y-1 block">
          <span class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/70 block">SEARCH / 搜索</span>
          <div class="relative">
            <Search class="pointer-events-none absolute left-3 top-1/2 size-3.5 -translate-y-1/2 text-muted-foreground/60" />
            <Input v-model="filters.search" class="h-9 pl-9" placeholder="标题、Slug 或正文" />
          </div>
        </label>

        <FilterSelect v-model="filters.status" label="状态" :options="statusFilterOptions" />
        <FilterSelect v-model="filters.locale" label="语言" :options="localeFilterOptions" />

        <div class="flex items-end gap-2">
          <Button type="submit" class="h-9 rounded-full px-4 font-black text-xs uppercase tracking-wider">
            <Search class="size-3.5" />
            搜索
          </Button>
          <Button type="button" variant="outline" class="h-9 rounded-full px-3 font-black text-xs uppercase tracking-wider" @click="resetFilters">
            <RotateCcw class="size-3.5" />
            重置
          </Button>
        </div>
      </form>
    </AdminFilterPanel>

    <AdminTablePanel :loading="loading" :batch-visible="selectedPosts.length > 0">
      <template #batch>
        <div class="flex flex-wrap items-center justify-between gap-2">
          <span class="text-xs font-medium">已选择 {{ selectedPosts.length }} 篇文章</span>
          <div class="flex flex-wrap gap-2">
            <Button v-if="hasPermission('content:edit')" size="sm" @click="requestBatchStatus('published')">
              <Send class="size-3.5" />
              批量发布
            </Button>
            <Button v-if="hasPermission('content:edit')" variant="outline" size="sm" @click="requestBatchStatus('draft')">
              <FilePenLine class="size-3.5" />
              转为草稿
            </Button>
            <Button v-if="hasPermission('content:delete')" variant="destructive" size="sm" @click="requestBatchDelete">
              <Trash2 class="size-3.5" />
              批量删除
            </Button>
          </div>
        </div>
      </template>

      <Table class="min-w-[1020px]">
        <TableHeader>
          <TableRow>
            <TableHead class="w-11">
              <Checkbox
                :model-value="selectionState"
                aria-label="选择当前页文章"
                @update:model-value="toggleAllPosts"
              />
            </TableHead>
            <TableHead class="w-16">ID</TableHead>
            <TableHead>标题</TableHead>
            <TableHead class="w-56">Slug</TableHead>
            <TableHead class="w-24">状态</TableHead>
            <TableHead class="w-20">语言</TableHead>
            <TableHead class="w-24 text-right">浏览量</TableHead>
            <TableHead class="w-44">创建时间</TableHead>
            <TableHead class="w-16 text-right">操作</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          <TableEmpty v-if="posts.length === 0" :colspan="9">
            <div class="flex flex-col items-center text-muted-foreground">
              <Newspaper class="mb-2 size-7 opacity-55" />
              <span class="text-xs">暂无文章</span>
            </div>
          </TableEmpty>

          <TableRow v-for="post in posts" :key="post.id">
            <TableCell>
              <Checkbox
                :model-value="isSelected(post.id)"
                :aria-label="`选择文章 ${post.title}`"
                @update:model-value="togglePost(post, $event)"
              />
            </TableCell>
            <TableCell class="font-mono text-xs text-muted-foreground">{{ post.id }}</TableCell>
            <TableCell class="max-w-80 truncate font-bold text-xs">{{ post.title }}</TableCell>
            <TableCell class="max-w-56 truncate font-mono text-xs text-muted-foreground">{{ post.slug }}</TableCell>
            <TableCell>
              <AdminStatusBadge :tone="statusTone(post.status)">{{ getStatusName(post.status) }}</AdminStatusBadge>
            </TableCell>
            <TableCell>{{ localeName(post.locale) }}</TableCell>
            <TableCell class="text-right tabular-nums">{{ Number(post.view_count || 0).toLocaleString('zh-CN') }}</TableCell>
            <TableCell class="text-xs text-muted-foreground">{{ formatDate(post.created_at) }}</TableCell>
            <TableCell class="text-right">
              <DropdownMenu>
                <DropdownMenuTrigger as-child>
                  <Button variant="ghost" size="icon" :aria-label="`管理文章 ${post.title}`">
                    <MoreHorizontal class="size-4" />
                  </Button>
                </DropdownMenuTrigger>
                <DropdownMenuContent align="end" class="w-40">
                  <DropdownMenuItem v-if="hasPermission('content:edit')" @select="showEditDialog(post)">
                    <Pencil class="size-4" />
                    编辑
                  </DropdownMenuItem>
                  <DropdownMenuItem v-if="hasPermission('content:edit')" @select="showTranslationsDialog(post)">
                    <Languages class="size-4" />
                    翻译版本
                  </DropdownMenuItem>
                  <DropdownMenuItem v-if="hasPermission('content:edit')" @select="requestToggleStatus(post)">
                    <Send v-if="post.status !== 'published'" class="size-4" />
                    <FilePenLine v-else class="size-4" />
                    {{ post.status === 'published' ? '转为草稿' : '发布' }}
                  </DropdownMenuItem>
                  <DropdownMenuSeparator v-if="hasPermission('content:delete')" />
                  <DropdownMenuItem
                    v-if="hasPermission('content:delete')"
                    class="text-destructive focus:text-destructive"
                    @select="requestDelete(post)"
                  >
                    <Trash2 class="size-4" />
                    删除
                  </DropdownMenuItem>
                </DropdownMenuContent>
              </DropdownMenu>
            </TableCell>
          </TableRow>
        </TableBody>
      </Table>

      <template #footer>
        <AdminPagination
          :page="pagination.page"
          :page-size="pagination.pageSize"
          :total="pagination.total"
          @update:page="updatePage"
          @update:page-size="updatePageSize"
        />
      </template>
    </AdminTablePanel>

    <Dialog v-model:open="dialogVisible">
      <DialogContent size="xl" class="max-h-[92dvh] overflow-y-auto p-0" @open-auto-focus.prevent>
        <form @submit.prevent="submitForm">
          <DialogHeader class="border-b px-5 py-4 pr-12">
            <DialogTitle>{{ dialogMode === 'create' ? '添加文章' : '编辑文章' }}</DialogTitle>
            <DialogDescription>正文支持 Markdown，发布状态和 SEO 信息可独立维护。</DialogDescription>
          </DialogHeader>

          <div class="space-y-7 px-5 py-5">
            <FormSection title="正文内容" description="文章标题、摘要与 Markdown 正文。">
              <div class="grid gap-4 md:grid-cols-[minmax(0,2fr)_minmax(150px,0.7fr)]">
                <FormField label="标题" required :error="formErrors.title">
                  <Input v-model="postForm.title" placeholder="请输入文章标题" @input="clearFieldError('title')" />
                </FormField>
                <FormField label="语言" required>
                  <Select v-model="postForm.locale">
                    <SelectTrigger class="w-full"><SelectValue /></SelectTrigger>
                    <SelectContent>
                      <SelectItem value="zh">中文</SelectItem>
                      <SelectItem value="en">English</SelectItem>
                    </SelectContent>
                  </Select>
                </FormField>
                <FormField label="Slug" required :error="formErrors.slug" class="md:col-span-2">
                  <Input v-model="postForm.slug" placeholder="例如 crystal-care-guide" @input="clearFieldError('slug')" />
                </FormField>
                <FormField label="摘要" class="md:col-span-2">
                  <Textarea v-model="postForm.excerpt" class="min-h-24" placeholder="请输入文章摘要" />
                </FormField>
                <FormField label="内容（Markdown）" class="md:col-span-2">
                  <Textarea
                    v-model="postForm.content"
                    class="min-h-80 resize-y font-mono text-[13px] leading-6"
                    placeholder="请输入文章内容"
                  />
                </FormField>
              </div>
            </FormSection>

            <FormSection title="发布信息" description="控制文章状态、封面图和内容标签。">
              <div class="grid gap-4 md:grid-cols-2">
                <FormField label="状态" required>
                  <Select v-model="postForm.status">
                    <SelectTrigger class="w-full"><SelectValue /></SelectTrigger>
                    <SelectContent>
                      <SelectItem value="draft">草稿</SelectItem>
                      <SelectItem value="published">已发布</SelectItem>
                      <SelectItem value="archived">已归档</SelectItem>
                    </SelectContent>
                  </Select>
                </FormField>
                <FormField label="特色图片">
                  <div class="relative">
                    <ImageIcon class="pointer-events-none absolute left-2.5 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" />
                    <Input v-model="postForm.featured_image" class="pl-9" placeholder="图片 URL" />
                  </div>
                </FormField>
                <FormField label="标签" class="md:col-span-2">
                  <Input v-model="postForm.tags" placeholder="多个标签用逗号分隔" />
                </FormField>
              </div>
            </FormSection>

            <FormSection title="SEO" description="可选的搜索结果信息与规范链接。">
              <div class="grid gap-4 md:grid-cols-2">
                <FormField label="SEO 标题">
                  <Input v-model="postForm.meta_title" placeholder="SEO 标题" />
                </FormField>
                <FormField label="SEO 关键词">
                  <Input v-model="postForm.meta_keywords" placeholder="多个关键词用逗号分隔" />
                </FormField>
                <FormField label="SEO 描述" class="md:col-span-2">
                  <Textarea v-model="postForm.meta_description" class="min-h-20" placeholder="SEO 描述" />
                </FormField>
                <FormField label="规范 URL" class="md:col-span-2">
                  <Input v-model="postForm.canonical_url" type="url" placeholder="https://example.com/article" />
                </FormField>
              </div>
            </FormSection>
          </div>

          <DialogFooter class="sticky bottom-0 mx-0 mb-0 rounded-b-lg border-t bg-background px-5 py-4">
            <Button type="button" variant="outline" @click="dialogVisible = false">取消</Button>
            <Button type="submit" :disabled="submitting">
              <LoaderCircle v-if="submitting" class="size-4 animate-spin" />
              {{ submitting ? '保存中' : '保存文章' }}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>

    <Dialog v-model:open="translationsDialogVisible">
      <DialogContent size="md" class="max-h-[85dvh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle>翻译管理</DialogTitle>
          <DialogDescription v-if="currentPost">
            {{ currentPost.title }} · {{ localeName(currentPost.locale) }}
          </DialogDescription>
        </DialogHeader>

        <div class="relative min-h-32 overflow-x-auto rounded-lg border">
          <div v-if="translationsLoading" class="absolute inset-0 z-10 flex items-center justify-center bg-background/80">
            <LoaderCircle class="size-5 animate-spin text-primary" aria-label="正在加载翻译版本" />
          </div>
          <Table class="min-w-[520px]">
            <TableHeader>
              <TableRow>
                <TableHead class="w-24">语言</TableHead>
                <TableHead>标题</TableHead>
                <TableHead class="w-24">状态</TableHead>
                <TableHead class="w-16 text-right">操作</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              <TableEmpty v-if="!translationsLoading && translations.length === 0" :colspan="4">暂无翻译版本</TableEmpty>
              <TableRow v-for="translation in translations" :key="translation.id">
                <TableCell>{{ localeName(translation.locale) }}</TableCell>
                <TableCell class="font-medium">{{ translation.title }}</TableCell>
                <TableCell>
                  <AdminStatusBadge :tone="statusTone(translation.status)">{{ getStatusName(translation.status) }}</AdminStatusBadge>
                </TableCell>
                <TableCell class="text-right">
                  <Button variant="ghost" size="icon" :aria-label="`编辑翻译 ${translation.title}`" @click="editTranslation(translation)">
                    <Pencil class="size-4" />
                  </Button>
                </TableCell>
              </TableRow>
            </TableBody>
          </Table>
        </div>
      </DialogContent>
    </Dialog>

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

<script setup>
import { computed, defineComponent, h, onMounted, reactive, ref } from 'vue'
import { toast } from 'vue-sonner'
import {
  Eye,
  FilePenLine,
  FileText,
  Image as ImageIcon,
  Languages,
  LoaderCircle,
  MoreHorizontal,
  Newspaper,
  Pencil,
  Plus,
  RotateCcw,
  Search,
  Send,
  Trash2
} from '@lucide/vue'
import AdminConfirmDialog from '@/components/admin/AdminConfirmDialog.vue'
import AdminFilterPanel from '@/components/admin/AdminFilterPanel.vue'
import AdminPageHeader from '@/components/admin/AdminPageHeader.vue'
import AdminPagination from '@/components/admin/AdminPagination.vue'
import AdminStatsGrid from '@/components/admin/AdminStatsGrid.vue'
import AdminStatusBadge from '@/components/admin/AdminStatusBadge.vue'
import AdminTablePanel from '@/components/admin/AdminTablePanel.vue'
import { Button } from '@/components/ui/button'
import { Checkbox } from '@/components/ui/checkbox'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle
} from '@/components/ui/dialog'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger
} from '@/components/ui/dropdown-menu'
import { Input } from '@/components/ui/input'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Table, TableBody, TableCell, TableEmpty, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Textarea } from '@/components/ui/textarea'
import { useAuthStore } from '@/stores/auth'
import axios from '@/utils/axios'

const FilterSelect = defineComponent({
  props: {
    modelValue: { type: String, required: true },
    label: { type: String, required: true },
    options: { type: Array, required: true }
  },
  emits: ['update:modelValue'],
  setup(props, { emit }) {
    return () => h('label', { class: 'space-y-1 block' }, [
      h('span', { class: 'text-[10px] font-black uppercase tracking-widest text-muted-foreground/70 block' }, props.label),
      h(Select, {
        modelValue: props.modelValue,
        'onUpdate:modelValue': (value) => emit('update:modelValue', value)
      }, {
        default: () => [
          h(SelectTrigger, { class: 'h-9 w-full' }, { default: () => h(SelectValue) }),
          h(SelectContent, {}, {
            default: () => props.options.map((option) => h(SelectItem, { value: option.value }, { default: () => option.label }))
          })
        ]
      })
    ])
  }
})

const FormSection = defineComponent({
  props: {
    title: { type: String, required: true },
    description: { type: String, default: '' }
  },
  setup(props, { slots }) {
    return () => h('section', { class: 'grid gap-4 border-t border-dashed pt-6 first:border-t-0 first:pt-0 lg:grid-cols-[170px_minmax(0,1fr)]' }, [
      h('div', {}, [
        h('h3', { class: 'text-sm font-black tracking-tighter italic uppercase text-foreground' }, props.title),
        props.description ? h('p', { class: 'mt-1 text-[9px] font-black uppercase tracking-widest text-muted-foreground/60' }, props.description) : null
      ]),
      h('div', { class: 'min-w-0' }, slots.default?.())
    ])
  }
})

const FormField = defineComponent({
  props: {
    label: { type: String, required: true },
    required: { type: Boolean, default: false },
    error: { type: String, default: '' }
  },
  setup(props, { slots, attrs }) {
    return () => h('label', { ...attrs, class: ['block space-y-1', attrs.class] }, [
      h('span', { class: 'text-[10px] font-black uppercase tracking-widest text-muted-foreground/70 block' }, [
        props.label,
        props.required ? h('span', { class: 'ml-0.5 text-destructive', 'aria-hidden': 'true' }, '*') : null
      ]),
      slots.default?.(),
      props.error ? h('span', { class: 'block text-xs font-medium text-destructive' }, props.error) : null
    ])
  }
})

const authStore = useAuthStore()
const loading = ref(false)
const posts = ref([])
const selectedPosts = ref([])
const dialogVisible = ref(false)
const translationsDialogVisible = ref(false)
const translationsLoading = ref(false)
const dialogMode = ref('create')
const submitting = ref(false)
const currentPost = ref(null)
const translations = ref([])
const stats = ref({})
const formErrors = reactive({})

const filters = reactive({ search: '', status: 'all', locale: 'all' })
const pagination = reactive({ page: 1, pageSize: 20, total: 0 })
const postForm = reactive({
  id: null,
  title: '',
  slug: '',
  content: '',
  excerpt: '',
  status: 'draft',
  locale: 'zh',
  featured_image: '',
  tags: '',
  meta_title: '',
  meta_description: '',
  meta_keywords: '',
  canonical_url: '',
  translation_group_id: null
})
const confirmation = reactive({
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
const localeFilterOptions = [
  { label: '全部语言', value: 'all' },
  { label: '中文', value: 'zh' },
  { label: 'English', value: 'en' }
]

const statItems = computed(() => [
  { key: 'total', label: '总文章数', value: stats.value.total || 0, icon: FileText, tone: 'gray' },
  { key: 'published', label: '已发布', value: stats.value.published || 0, icon: Send, tone: 'green' },
  { key: 'draft', label: '草稿', value: stats.value.draft || 0, icon: FilePenLine, tone: 'amber' },
  { key: 'views', label: '总浏览量', value: Number(stats.value.total_views || 0).toLocaleString('zh-CN'), icon: Eye, tone: 'blue' }
])
const selectionState = computed(() => {
  if (posts.value.length === 0 || selectedPosts.value.length === 0) return false
  return selectedPosts.value.length === posts.value.length ? true : 'indeterminate'
})

const hasPermission = (permission) => authStore.hasPermission(permission)
const getStatusName = (status) => ({ draft: '草稿', published: '已发布', archived: '已归档' })[status] || status
const statusTone = (status) => ({ draft: 'gray', published: 'green', archived: 'amber' })[status] || 'gray'
const localeName = (locale) => ({ zh: '中文', en: 'English' })[locale] || locale || '-'
const formatDate = (dateString) => dateString ? new Date(dateString).toLocaleString('zh-CN') : '-'

const clearFormErrors = () => Object.keys(formErrors).forEach((key) => delete formErrors[key])
const clearFieldError = (field) => { delete formErrors[field] }
const buildPostPayload = () => ({
  title: postForm.title.trim(),
  slug: postForm.slug.trim(),
  content: postForm.content,
  excerpt: postForm.excerpt,
  status: postForm.status,
  locale: postForm.locale,
  featured_image: postForm.featured_image.trim(),
  tags: postForm.tags,
  meta_title: postForm.meta_title,
  meta_description: postForm.meta_description,
  meta_keywords: postForm.meta_keywords,
  canonical_url: postForm.canonical_url.trim(),
  translation_group_id: postForm.translation_group_id
})
const validateForm = (payload) => {
  clearFormErrors()
  if (!payload.title) formErrors.title = '请输入文章标题'
  if (!payload.slug) formErrors.slug = '请输入 URL slug'
  if (Object.keys(formErrors).length > 0) {
    toast.error('请检查文章表单中的必填项')
    return false
  }
  return true
}
const resetForm = () => {
  Object.assign(postForm, {
    id: null,
    title: '',
    slug: '',
    content: '',
    excerpt: '',
    status: 'draft',
    locale: 'zh',
    featured_image: '',
    tags: '',
    meta_title: '',
    meta_description: '',
    meta_keywords: '',
    canonical_url: '',
    translation_group_id: null
  })
  clearFormErrors()
}

const buildFilterParams = () => ({
  ...(filters.search.trim() ? { search: filters.search.trim() } : {}),
  ...(filters.status !== 'all' ? { status: filters.status } : {}),
  ...(filters.locale !== 'all' ? { locale: filters.locale } : {})
})
const fetchStats = async () => {
  try {
    const response = await axios.get('/api/admin/content/posts/stats')
    stats.value = response.data || {}
  } catch (error) {
    console.error('Failed to fetch content stats:', error)
  }
}
const fetchPosts = async () => {
  loading.value = true
  try {
    const response = await axios.get('/api/admin/content/posts', {
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
const refreshContent = () => Promise.all([fetchPosts(), fetchStats()])
const applyFilters = () => { pagination.page = 1; fetchPosts() }
const resetFilters = () => {
  Object.assign(filters, { search: '', status: 'all', locale: 'all' })
  pagination.page = 1
  fetchPosts()
}
const updatePage = (page) => { pagination.page = page; fetchPosts() }
const updatePageSize = (pageSize) => { pagination.pageSize = pageSize; pagination.page = 1; fetchPosts() }

const showCreateDialog = () => {
  dialogMode.value = 'create'
  resetForm()
  dialogVisible.value = true
}
const showEditDialog = (post) => {
  dialogMode.value = 'edit'
  Object.assign(postForm, {
    id: post.id,
    title: post.title || '',
    slug: post.slug || '',
    content: post.content || '',
    excerpt: post.excerpt || '',
    status: post.status || 'draft',
    locale: post.locale || 'zh',
    featured_image: post.featured_image || '',
    tags: post.tags || '',
    meta_title: post.meta_title || '',
    meta_description: post.meta_description || post.meta_desc || '',
    meta_keywords: post.meta_keywords || '',
    canonical_url: post.canonical_url || '',
    translation_group_id: post.translation_group_id || null
  })
  clearFormErrors()
  dialogVisible.value = true
}
const submitForm = async () => {
  const payload = buildPostPayload()
  if (!validateForm(payload)) return
  submitting.value = true
  try {
    if (dialogMode.value === 'create') {
      await axios.post('/api/admin/content/posts', payload)
      toast.success('文章创建成功')
    } else {
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

const showTranslationsDialog = async (post) => {
  currentPost.value = post
  translations.value = []
  translationsDialogVisible.value = true
  translationsLoading.value = true
  try {
    const response = await axios.get(`/api/admin/content/posts/${post.id}/translations`)
    if (!Array.isArray(response.data?.translations)) throw new Error('Missing translations array in response')
    translations.value = response.data.translations
  } catch (error) {
    console.error('Failed to fetch translations:', error)
    translationsDialogVisible.value = false
  } finally {
    translationsLoading.value = false
  }
}
const editTranslation = (translation) => {
  translationsDialogVisible.value = false
  showEditDialog(translation)
}

const isSelected = (postId) => selectedPosts.value.some((post) => post.id === postId)
const toggleAllPosts = (checked) => { selectedPosts.value = checked === true ? [...posts.value] : [] }
const togglePost = (post, checked) => {
  if (checked === true && !isSelected(post.id)) selectedPosts.value = [...selectedPosts.value, post]
  else if (checked !== true) selectedPosts.value = selectedPosts.value.filter((selected) => selected.id !== post.id)
}
const setConfirmation = (values) => Object.assign(confirmation, {
  open: true,
  type: '',
  target: null,
  status: '',
  confirmLabel: '确定',
  destructive: false,
  ...values
})
const requestToggleStatus = (post) => {
  const status = post.status === 'published' ? 'draft' : 'published'
  const action = status === 'published' ? '发布' : '转为草稿'
  setConfirmation({
    type: 'status', target: post, status, title: `${action}文章？`,
    description: `文章“${post.title}”将被${action}。`, confirmLabel: action
  })
}
const requestDelete = (post) => setConfirmation({
  type: 'delete', target: post, title: '删除文章？',
  description: `文章“${post.title}”将被永久删除，此操作不可恢复。`, confirmLabel: '删除', destructive: true
})
const requestBatchStatus = (status) => {
  const action = status === 'published' ? '发布' : '转为草稿'
  setConfirmation({
    type: 'batch-status', target: [...selectedPosts.value], status, title: `批量${action}文章？`,
    description: `将 ${selectedPosts.value.length} 篇文章批量${action}。`, confirmLabel: `批量${action}`
  })
}
const requestBatchDelete = () => setConfirmation({
  type: 'batch-delete', target: [...selectedPosts.value], title: '批量删除文章？',
  description: `${selectedPosts.value.length} 篇文章将被永久删除，此操作不可恢复。`, confirmLabel: '批量删除', destructive: true
})
const executeConfirmedAction = async () => {
  const { type, target, status } = confirmation
  confirmation.open = false
  try {
    if (type === 'status') {
      await axios.patch(`/api/admin/content/posts/${target.id}/status`, { status })
      toast.success(status === 'published' ? '文章已发布' : '文章已转为草稿')
    } else if (type === 'delete') {
      await axios.delete(`/api/admin/content/posts/${target.id}`)
      toast.success('文章已删除')
    } else if (type === 'batch-status') {
      await axios.post('/api/admin/content/posts/batch-status', { post_ids: target.map((post) => post.id), status })
      toast.success(status === 'published' ? '文章已批量发布' : '文章已批量转为草稿')
    } else if (type === 'batch-delete') {
      await axios.post('/api/admin/content/posts/batch-delete', { post_ids: target.map((post) => post.id) })
      toast.success('文章已批量删除')
    }
    await refreshContent()
  } catch (error) {
    console.error('Failed to update posts:', error)
  }
}

onMounted(() => Promise.all([fetchStats(), fetchPosts()]))
</script>
