<template>
  <div class="space-y-4">
    <AdminPageHeader title="FAQ 管理" description="维护常见问题、分类、发布状态和展示顺序">
      <template #actions>
        <Button v-if="hasPermission('faq:create')" @click="showCreateDialog">
          <Plus class="size-4" />
          添加 FAQ
        </Button>
      </template>
    </AdminPageHeader>

    <AdminFilterPanel>
      <form class="grid grid-cols-1 gap-3 md:grid-cols-2 xl:grid-cols-[minmax(240px,1.4fr)_repeat(3,minmax(140px,0.7fr))_auto]" @submit.prevent="applyFilters">
        <label class="space-y-1.5">
          <span class="text-xs font-medium text-muted-foreground">搜索</span>
          <div class="relative">
            <Search class="pointer-events-none absolute left-2.5 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" />
            <Input v-model="filters.search" class="h-9 pl-9" placeholder="问题或答案" />
          </div>
        </label>

        <FilterSelect v-model="filters.category" label="分类" :options="categoryFilterOptions" />
        <FilterSelect v-model="filters.status" label="状态" :options="statusFilterOptions" />
        <FilterSelect v-model="filters.locale" label="语言" :options="localeFilterOptions" />

        <div class="flex items-end gap-2">
          <Button type="submit" class="h-9">
            <Search class="size-4" />
            搜索
          </Button>
          <Button type="button" variant="outline" class="h-9" @click="resetFilters">
            <RotateCcw class="size-4" />
            重置
          </Button>
        </div>
      </form>
    </AdminFilterPanel>

    <AdminTablePanel :loading="loading" :batch-visible="selectedFAQs.length > 0">
      <template #batch>
        <div class="flex flex-wrap items-center justify-between gap-2">
          <span class="text-xs font-medium">已选择 {{ selectedFAQs.length }} 个 FAQ</span>
          <Button v-if="hasPermission('faq:delete')" variant="destructive" size="sm" @click="requestBatchDelete">
            <Trash2 class="size-3.5" />
            批量删除
          </Button>
        </div>
      </template>

      <Table class="min-w-[1040px]">
        <TableHeader>
          <TableRow>
            <TableHead class="w-11">
              <Checkbox
                :model-value="selectionState"
                aria-label="选择当前页 FAQ"
                @update:model-value="toggleAllFAQs"
              />
            </TableHead>
            <TableHead class="w-16">ID</TableHead>
            <TableHead>问题</TableHead>
            <TableHead>答案</TableHead>
            <TableHead class="w-32">分类</TableHead>
            <TableHead class="w-24">状态</TableHead>
            <TableHead class="w-20">语言</TableHead>
            <TableHead class="w-20 text-right">排序</TableHead>
            <TableHead class="w-44">创建时间</TableHead>
            <TableHead class="w-16 text-right">操作</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          <TableEmpty v-if="faqs.length === 0" :colspan="10">
            <div class="flex flex-col items-center text-muted-foreground">
              <CircleHelp class="mb-2 size-7 opacity-55" />
              <span class="text-xs">暂无 FAQ</span>
            </div>
          </TableEmpty>

          <TableRow v-for="faq in faqs" :key="faq.id">
            <TableCell>
              <Checkbox
                :model-value="isSelected(faq.id)"
                :aria-label="`选择 FAQ ${faq.question}`"
                @update:model-value="toggleFAQ(faq, $event)"
              />
            </TableCell>
            <TableCell class="font-mono text-xs text-muted-foreground">{{ faq.id }}</TableCell>
            <TableCell class="max-w-72">
              <p class="line-clamp-2 font-medium leading-5">{{ faq.question }}</p>
            </TableCell>
            <TableCell class="max-w-80">
              <p class="line-clamp-2 text-muted-foreground">{{ faq.answer }}</p>
            </TableCell>
            <TableCell>
              <AdminStatusBadge tone="blue">{{ faq.category || '-' }}</AdminStatusBadge>
            </TableCell>
            <TableCell>
              <AdminStatusBadge :tone="statusTone(faq.status)">{{ statusName(faq.status) }}</AdminStatusBadge>
            </TableCell>
            <TableCell>{{ localeName(faq.locale) }}</TableCell>
            <TableCell class="text-right tabular-nums">{{ faq.order ?? faq.sort_order ?? 0 }}</TableCell>
            <TableCell class="text-xs text-muted-foreground">{{ formatDate(faq.created_at) }}</TableCell>
            <TableCell class="text-right">
              <DropdownMenu>
                <DropdownMenuTrigger as-child>
                  <Button variant="ghost" size="icon" :aria-label="`管理 FAQ ${faq.question}`">
                    <MoreHorizontal class="size-4" />
                  </Button>
                </DropdownMenuTrigger>
                <DropdownMenuContent align="end" class="w-36">
                  <DropdownMenuItem v-if="hasPermission('faq:edit')" @select="showEditDialog(faq)">
                    <Pencil class="size-4" />
                    编辑
                  </DropdownMenuItem>
                  <DropdownMenuSeparator v-if="hasPermission('faq:delete')" />
                  <DropdownMenuItem
                    v-if="hasPermission('faq:delete')"
                    class="text-destructive focus:text-destructive"
                    @select="requestDelete(faq)"
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
      <DialogContent class="max-h-[90vh] overflow-y-auto sm:max-w-2xl" @open-auto-focus.prevent>
        <form class="space-y-5" @submit.prevent="submitForm">
          <DialogHeader>
            <DialogTitle>{{ dialogMode === 'create' ? '添加 FAQ' : '编辑 FAQ' }}</DialogTitle>
            <DialogDescription>填写面向访客的问题与答案，并设置其展示位置。</DialogDescription>
          </DialogHeader>

          <FormField label="问题" required :error="formErrors.question">
            <Textarea
              v-model="faqForm.question"
              class="min-h-20"
              placeholder="请输入问题"
              @input="clearFieldError('question')"
            />
          </FormField>

          <FormField label="答案" required :error="formErrors.answer">
            <Textarea
              v-model="faqForm.answer"
              class="min-h-36"
              placeholder="请输入答案"
              @input="clearFieldError('answer')"
            />
          </FormField>

          <div class="grid gap-4 sm:grid-cols-2">
            <FormField label="分类" required :error="formErrors.category">
              <Input
                v-model="faqForm.category"
                list="faq-category-suggestions"
                placeholder="例如 Orders"
                @input="clearFieldError('category')"
              />
              <datalist id="faq-category-suggestions">
                <option v-for="category in categories" :key="category" :value="category" />
              </datalist>
            </FormField>

            <FormField label="页面标识">
              <Input v-model="faqForm.page_id" placeholder="例如 support" />
            </FormField>

            <FormField label="语言" required>
              <Select v-model="faqForm.locale">
                <SelectTrigger class="w-full"><SelectValue /></SelectTrigger>
                <SelectContent>
                  <SelectItem value="zh">中文</SelectItem>
                  <SelectItem value="en">English</SelectItem>
                </SelectContent>
              </Select>
            </FormField>

            <FormField label="状态" required>
              <Select v-model="faqForm.status">
                <SelectTrigger class="w-full"><SelectValue /></SelectTrigger>
                <SelectContent>
                  <SelectItem value="published">已发布</SelectItem>
                  <SelectItem value="draft">草稿</SelectItem>
                </SelectContent>
              </Select>
            </FormField>

            <FormField label="排序">
              <Input v-model.number="faqForm.order" type="number" min="0" max="9999" step="1" />
            </FormField>
          </div>

          <DialogFooter>
            <Button type="button" variant="outline" @click="dialogVisible = false">取消</Button>
            <Button type="submit" :disabled="submitting">
              <LoaderCircle v-if="submitting" class="size-4 animate-spin" />
              {{ submitting ? '保存中' : '保存 FAQ' }}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>

    <AdminConfirmDialog
      v-model:open="confirmation.open"
      :title="confirmation.title"
      :description="confirmation.description"
      :confirm-label="confirmation.confirmLabel"
      destructive
      @confirm="executeConfirmedAction"
    />
  </div>
</template>

<script setup>
import { computed, defineComponent, h, onMounted, reactive, ref } from 'vue'
import { toast } from 'vue-sonner'
import { CircleHelp, LoaderCircle, MoreHorizontal, Pencil, Plus, RotateCcw, Search, Trash2 } from '@lucide/vue'
import AdminConfirmDialog from '@/components/admin/AdminConfirmDialog.vue'
import AdminFilterPanel from '@/components/admin/AdminFilterPanel.vue'
import AdminPageHeader from '@/components/admin/AdminPageHeader.vue'
import AdminPagination from '@/components/admin/AdminPagination.vue'
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
    return () => h('label', { class: 'space-y-1.5' }, [
      h('span', { class: 'text-xs font-medium text-muted-foreground' }, props.label),
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

const FormField = defineComponent({
  props: {
    label: { type: String, required: true },
    required: { type: Boolean, default: false },
    error: { type: String, default: '' }
  },
  setup(props, { slots, attrs }) {
    return () => h('label', { ...attrs, class: ['block space-y-1.5', attrs.class] }, [
      h('span', { class: 'text-xs font-medium' }, [
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
const faqs = ref([])
const categories = ref([])
const selectedFAQs = ref([])
const dialogVisible = ref(false)
const dialogMode = ref('create')
const submitting = ref(false)
const formErrors = reactive({})

const filters = reactive({ search: '', category: 'all', status: 'all', locale: 'all' })
const pagination = reactive({ page: 1, pageSize: 20, total: 0 })
const faqForm = reactive({
  id: null,
  question: '',
  answer: '',
  category: '',
  page_id: '',
  locale: 'zh',
  status: 'published',
  order: 0
})
const confirmation = reactive({
  open: false,
  type: '',
  target: null,
  title: '',
  description: '',
  confirmLabel: '删除'
})

const statusFilterOptions = [
  { label: '全部状态', value: 'all' },
  { label: '已发布', value: 'published' },
  { label: '草稿', value: 'draft' }
]
const localeFilterOptions = [
  { label: '全部语言', value: 'all' },
  { label: '中文', value: 'zh' },
  { label: 'English', value: 'en' }
]
const categoryFilterOptions = computed(() => [
  { label: '全部分类', value: 'all' },
  ...categories.value.map((category) => ({ label: category, value: category }))
])
const selectionState = computed(() => {
  if (faqs.value.length === 0 || selectedFAQs.value.length === 0) return false
  return selectedFAQs.value.length === faqs.value.length ? true : 'indeterminate'
})

const hasPermission = (permission) => authStore.hasPermission(permission)
const localeName = (locale) => ({ zh: '中文', en: 'English' })[locale] || locale || '-'
const statusName = (status) => ({ published: '已发布', draft: '草稿' })[status] || status || '-'
const statusTone = (status) => ({ published: 'green', draft: 'gray' })[status] || 'gray'
const formatDate = (dateString) => dateString ? new Date(dateString).toLocaleString('zh-CN') : '-'

const clearFormErrors = () => Object.keys(formErrors).forEach((key) => delete formErrors[key])
const clearFieldError = (field) => { delete formErrors[field] }
const buildFAQPayload = () => ({
  question: faqForm.question.trim(),
  answer: faqForm.answer.trim(),
  category: faqForm.category.trim(),
  page_id: faqForm.page_id.trim(),
  locale: faqForm.locale,
  status: faqForm.status,
  order: Math.max(0, Number(faqForm.order || 0))
})
const validateForm = (payload) => {
  clearFormErrors()
  if (!payload.question) formErrors.question = '请输入问题'
  if (!payload.answer) formErrors.answer = '请输入答案'
  if (!payload.category) formErrors.category = '请输入分类'
  if (Object.keys(formErrors).length > 0) {
    toast.error('请检查 FAQ 表单中的必填项')
    return false
  }
  return true
}
const resetForm = () => {
  Object.assign(faqForm, {
    id: null,
    question: '',
    answer: '',
    category: '',
    page_id: '',
    locale: 'zh',
    status: 'published',
    order: 0
  })
  clearFormErrors()
}

const buildFilterParams = () => ({
  ...(filters.search.trim() ? { search: filters.search.trim() } : {}),
  ...(filters.category !== 'all' ? { category: filters.category } : {}),
  ...(filters.status !== 'all' ? { status: filters.status } : {}),
  ...(filters.locale !== 'all' ? { locale: filters.locale } : {})
})
const fetchFAQs = async () => {
  loading.value = true
  try {
    const response = await axios.get('/api/admin/faqs', {
      params: { page: pagination.page, page_size: pagination.pageSize, ...buildFilterParams() }
    })
    faqs.value = response.data.faqs || []
    pagination.total = response.data.pagination?.total ?? response.data.total ?? 0
    selectedFAQs.value = []
  } catch (error) {
    console.error('Failed to fetch FAQs:', error)
  } finally {
    loading.value = false
  }
}
const fetchCategories = async () => {
  try {
    const responses = await Promise.all([
      axios.get('/api/admin/faqs/categories', { params: { locale: 'zh' } }),
      axios.get('/api/admin/faqs/categories', { params: { locale: 'en' } })
    ])
    categories.value = [...new Set(responses.flatMap((response) => response.data.categories || []))].sort()
  } catch (error) {
    console.error('Failed to fetch FAQ categories:', error)
  }
}
const refreshFAQs = () => Promise.all([fetchFAQs(), fetchCategories()])
const applyFilters = () => { pagination.page = 1; fetchFAQs() }
const resetFilters = () => {
  Object.assign(filters, { search: '', category: 'all', status: 'all', locale: 'all' })
  pagination.page = 1
  fetchFAQs()
}
const updatePage = (page) => { pagination.page = page; fetchFAQs() }
const updatePageSize = (pageSize) => { pagination.pageSize = pageSize; pagination.page = 1; fetchFAQs() }

const showCreateDialog = () => {
  dialogMode.value = 'create'
  resetForm()
  dialogVisible.value = true
}
const showEditDialog = async (faq) => {
  dialogMode.value = 'edit'
  try {
    const response = await axios.get(`/api/admin/faqs/${faq.id}`)
    const detail = response.data.faq || faq
    Object.assign(faqForm, {
      id: detail.id,
      question: detail.question || '',
      answer: detail.answer || '',
      category: detail.category || '',
      page_id: detail.page_id || '',
      locale: detail.locale || 'zh',
      status: detail.status || 'published',
      order: detail.order ?? detail.sort_order ?? 0
    })
    clearFormErrors()
    dialogVisible.value = true
  } catch (error) {
    console.error('Failed to fetch FAQ detail:', error)
  }
}
const submitForm = async () => {
  const payload = buildFAQPayload()
  if (!validateForm(payload)) return
  submitting.value = true
  try {
    if (dialogMode.value === 'create') {
      await axios.post('/api/admin/faqs', payload)
      toast.success('FAQ 创建成功')
    } else {
      await axios.put(`/api/admin/faqs/${faqForm.id}`, payload)
      toast.success('FAQ 更新成功')
    }
    dialogVisible.value = false
    await refreshFAQs()
  } catch (error) {
    console.error('Failed to save FAQ:', error)
  } finally {
    submitting.value = false
  }
}

const isSelected = (faqId) => selectedFAQs.value.some((faq) => faq.id === faqId)
const toggleAllFAQs = (checked) => { selectedFAQs.value = checked === true ? [...faqs.value] : [] }
const toggleFAQ = (faq, checked) => {
  if (checked === true && !isSelected(faq.id)) selectedFAQs.value = [...selectedFAQs.value, faq]
  else if (checked !== true) selectedFAQs.value = selectedFAQs.value.filter((selected) => selected.id !== faq.id)
}
const requestDelete = (faq) => Object.assign(confirmation, {
  open: true,
  type: 'delete',
  target: faq,
  title: '删除 FAQ？',
  description: `问题“${faq.question}”将被永久删除，此操作不可恢复。`,
  confirmLabel: '删除'
})
const requestBatchDelete = () => Object.assign(confirmation, {
  open: true,
  type: 'batch-delete',
  target: [...selectedFAQs.value],
  title: '批量删除 FAQ？',
  description: `${selectedFAQs.value.length} 个 FAQ 将被永久删除，此操作不可恢复。`,
  confirmLabel: '批量删除'
})
const executeConfirmedAction = async () => {
  const { type, target } = confirmation
  confirmation.open = false
  try {
    if (type === 'delete') {
      await axios.delete(`/api/admin/faqs/${target.id}`)
      toast.success('FAQ 已删除')
    } else if (type === 'batch-delete') {
      const response = await axios.post('/api/admin/faqs/batch-delete', { faq_ids: target.map((faq) => faq.id) })
      toast.success(`已删除 ${response.data.deleted ?? target.length} 个 FAQ`)
    }
    await refreshFAQs()
  } catch (error) {
    console.error('Failed to delete FAQs:', error)
  }
}

onMounted(() => Promise.all([fetchFAQs(), fetchCategories()]))
</script>
