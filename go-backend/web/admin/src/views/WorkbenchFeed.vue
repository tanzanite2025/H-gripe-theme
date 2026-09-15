<template>
  <div class="space-y-4">
    <AdminPageHeader
      title="车间手记"
      description="用简短现场记录展示装配、维修和测量过程；提到的商品在前台直接打开商品详情抽屉。"
    >
      <template #actions>
        <Button v-if="canCreate" @click="openCreate">
          <Plus class="size-4" />
          新建手记
        </Button>
      </template>
    </AdminPageHeader>

    <AdminFilterPanel>
      <form class="grid gap-3 md:grid-cols-[minmax(220px,1.3fr)_minmax(140px,.55fr)_minmax(130px,.45fr)_auto]" @submit.prevent="applyFilters">
        <label class="block space-y-1">
          <span class="block text-[10px] font-black uppercase tracking-widest text-muted-foreground/70">SEARCH / 搜索</span>
          <div class="relative">
            <Search class="pointer-events-none absolute left-3 top-1/2 size-3.5 -translate-y-1/2 text-muted-foreground/60" />
            <Input v-model="filters.search" class="h-9 pl-9" placeholder="LOG 编号或正文内容" />
          </div>
        </label>

        <label class="block space-y-1">
          <span class="block text-[10px] font-black uppercase tracking-widest text-muted-foreground/70">STATUS / 状态</span>
          <Select v-model="filters.status">
            <SelectTrigger class="h-9"><SelectValue /></SelectTrigger>
            <SelectContent>
              <SelectItem value="all">全部状态</SelectItem>
              <SelectItem value="draft">草稿</SelectItem>
              <SelectItem value="published">已发布</SelectItem>
              <SelectItem value="archived">已归档</SelectItem>
            </SelectContent>
          </Select>
        </label>

        <AdminFilterSelect v-model="filters.locale" label="LOCALE / 语言" :options="localeFilterOptions" />

        <div class="flex items-end gap-2">
          <Button type="submit" class="h-9 px-4">
            <Search class="size-3.5" />
            搜索
          </Button>
          <Button type="button" variant="outline" class="h-9" aria-label="重置筛选" title="重置筛选" @click="resetFilters">
            <RotateCcw class="size-3.5" />
          </Button>
        </div>
      </form>
    </AdminFilterPanel>

    <AdminTablePanel :loading="loading">
      <Table class="min-w-[1040px]">
        <TableHeader>
          <TableRow>
            <TableHead class="w-32">记录</TableHead>
            <TableHead class="w-[31rem]">现场内容</TableHead>
            <TableHead class="w-28">媒体</TableHead>
            <TableHead class="w-28">商品</TableHead>
            <TableHead class="w-28">状态</TableHead>
            <TableHead class="w-36">发布时间</TableHead>
            <TableHead class="w-16 text-right">操作</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          <TableEmpty v-if="entries.length === 0" :colspan="7">
            <div class="flex flex-col items-center text-muted-foreground">
              <ClipboardList class="mb-2 size-7 opacity-55" />
              <span class="text-xs">{{ filters.search ? '没有匹配的车间手记' : '还没有车间手记' }}</span>
            </div>
          </TableEmpty>

          <TableRow v-for="entry in entries" :key="entry.id">
            <TableCell>
              <div>
                <p class="font-mono text-xs font-black">{{ entry.entry_number }}</p>
                <p class="mt-1 text-[10px] text-muted-foreground">{{ entry.locale }}</p>
              </div>
            </TableCell>
            <TableCell>
              <p class="line-clamp-2 max-w-[31rem] whitespace-pre-wrap text-xs leading-5">{{ entry.content }}</p>
              <div v-if="entry.tags.length" class="mt-1.5 flex flex-wrap gap-1">
                <span v-for="tag in entry.tags.slice(0, 4)" :key="tag" class="rounded-full bg-muted px-2 py-0.5 text-[9px] font-bold text-muted-foreground">
                  {{ tag }}
                </span>
              </div>
            </TableCell>
            <TableCell class="font-mono text-xs font-bold">{{ entry.media.length }}</TableCell>
            <TableCell class="font-mono text-xs font-bold">{{ entry.tagged_products.length }}</TableCell>
            <TableCell>
              <span class="inline-flex rounded-full px-2 py-1 text-[10px] font-black" :class="statusClass(entry.status)">
                {{ statusLabel(entry.status) }}
              </span>
            </TableCell>
            <TableCell class="font-mono text-[10px] text-muted-foreground">{{ formatDate(entry.published_at) }}</TableCell>
            <TableCell class="text-right">
              <DropdownMenu>
                <DropdownMenuTrigger as-child>
                  <Button variant="ghost" size="icon" :aria-label="`管理 ${entry.entry_number}`">
                    <MoreHorizontal class="size-4" />
                  </Button>
                </DropdownMenuTrigger>
                <DropdownMenuContent align="end" class="w-36">
                  <DropdownMenuItem v-if="canEdit" @select="openEdit(entry)">
                    <Pencil class="size-4" />
                    编辑
                  </DropdownMenuItem>
                  <DropdownMenuItem v-if="canEdit && entry.status !== 'published'" @select="setStatus(entry, 'published')">
                    <Send class="size-4" />
                    发布
                  </DropdownMenuItem>
                  <DropdownMenuItem v-if="canEdit && entry.status !== 'draft'" @select="setStatus(entry, 'draft')">
                    <FilePenLine class="size-4" />
                    转为草稿
                  </DropdownMenuItem>
                  <DropdownMenuItem v-if="canEdit && entry.status !== 'archived'" @select="setStatus(entry, 'archived')">
                    <Archive class="size-4" />
                    归档
                  </DropdownMenuItem>
                  <DropdownMenuSeparator v-if="canDelete" />
                  <DropdownMenuItem v-if="canDelete" class="text-destructive focus:text-destructive" @select="removeEntry(entry)">
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
          :page-size="pagination.page_size"
          :total="pagination.total"
          @update:page="updatePage"
          @update:page-size="updatePageSize"
        />
      </template>
    </AdminTablePanel>

    <Dialog v-model:open="editorOpen">
      <DialogContent size="xl">
        <DialogHeader>
          <DialogTitle>{{ form.id ? `编辑 ${form.entry_number}` : '新建车间手记' }}</DialogTitle>
          <DialogDescription>
            现场记录最多 300 个字符。关联商品在前台展示实时可见的商品信息，点击后直接打开全局商品详情抽屉。
          </DialogDescription>
        </DialogHeader>

        <form class="space-y-5" @submit.prevent="save(form.status)">
          <div class="grid gap-3 md:grid-cols-[minmax(0,1fr)_9rem_10rem]">
            <AdminFormField label="现场记录" required>
              <Textarea
                v-model="form.content"
                :disabled="saving"
                class="min-h-32"
                maxlength="300"
                placeholder="例如：今天换了一条断辐，顺手复核这颗花鼓的轴承预压..."
              />
              <p class="mt-1 text-right font-mono text-[10px] text-muted-foreground">{{ form.content.length }}/300</p>
            </AdminFormField>
            <AdminFormField label="语言">
              <StorefrontLocaleSelect
                v-model="form.locale"
                :language-options="languageOptions"
                :disabled="saving"
                :loading="supportedLanguages.loading.value"
              />
            </AdminFormField>
            <AdminFormField label="保存状态">
              <Select v-model="form.status" :disabled="saving">
                <SelectTrigger><SelectValue /></SelectTrigger>
                <SelectContent>
                  <SelectItem value="draft">草稿</SelectItem>
                  <SelectItem value="published">已发布</SelectItem>
                  <SelectItem value="archived">已归档</SelectItem>
                </SelectContent>
              </Select>
            </AdminFormField>
          </div>

          <AdminFormField label="话题标签" description="用逗号或换行分隔，保存后自动补上 #。">
            <Textarea
              v-model="form.tagsText"
              :disabled="saving"
              class="min-h-20"
              maxlength="400"
              placeholder="Gravel, WheelBuild, TireFitment"
            />
          </AdminFormField>

          <section class="space-y-3">
            <div class="flex flex-wrap items-end justify-between gap-3">
              <div>
                <h2 class="text-sm font-black tracking-tight">车间照片</h2>
                <p class="mt-1 text-[10px] text-muted-foreground">最多 4 张。文件选择或拖放后立即上传；每张图片可添加简短实测注解。</p>
              </div>
              <label
                class="inline-flex h-8 cursor-pointer items-center justify-center gap-1.5 rounded-full border border-dashed border-border px-3 text-xs font-black uppercase tracking-tight transition-colors hover:border-primary hover:bg-primary/5"
                :class="{ 'pointer-events-none opacity-50': saving || uploadingMedia || form.media.length >= 4 }"
              >
                <Upload class="size-3.5" />
                <span>{{ uploadingMedia ? '上传中...' : '选择图片' }}</span>
                <input
                  class="sr-only"
                  type="file"
                  accept="image/jpeg,image/png,image/webp,image/gif"
                  multiple
                  :disabled="saving || uploadingMedia || form.media.length >= 4"
                  @change="onMediaFileChange"
                >
              </label>
            </div>

            <div
              class="grid min-h-32 place-items-center rounded-lg border border-dashed px-4 py-6 text-center transition-colors"
              :class="dragOver ? 'border-primary bg-primary/5' : 'border-border/80 bg-muted/15'"
              @dragenter.prevent="dragOver = true"
              @dragover.prevent="dragOver = true"
              @dragleave.prevent="dragOver = false"
              @drop.prevent="onMediaDrop"
            >
              <div v-if="form.media.length === 0" class="grid justify-items-center gap-2 text-muted-foreground">
                <Images class="size-6" />
                <span class="text-xs">把车间照片拖放到这里</span>
              </div>

              <div v-else class="grid w-full gap-3 sm:grid-cols-2">
                <article v-for="(media, index) in form.media" :key="`${media.file_path}-${index}`" class="grid gap-2 rounded-lg border bg-background p-2 text-left">
                  <div class="relative aspect-[4/3] overflow-hidden rounded-md bg-muted">
                    <img :src="media.file_path" :alt="media.caption || `车间照片 ${index + 1}`" class="size-full object-cover">
                    <Button
                      type="button"
                      variant="destructive"
                      size="icon-xs"
                      class="absolute right-1.5 top-1.5"
                      :disabled="saving"
                      :aria-label="`移除照片 ${index + 1}`"
                      title="移除照片"
                      @click="removeMedia(index)"
                    >
                      <X class="size-3" />
                    </Button>
                  </div>
                  <Input
                    v-model="media.caption"
                    :disabled="saving"
                    maxlength="160"
                    placeholder="实测 65psi 下充气宽度 29.8mm"
                  />
                  <p class="font-mono text-[9px] text-muted-foreground">{{ media.width }} x {{ media.height }} · {{ media.file_size_kb }}KB</p>
                </article>
              </div>
            </div>
          </section>

          <section class="space-y-3">
            <div class="flex flex-wrap items-end justify-between gap-3">
              <div>
                <h2 class="text-sm font-black tracking-tight">提到 / 使用的商品</h2>
                <p class="mt-1 text-[10px] text-muted-foreground">只读搜索商品目录。前台标签显示商品信息，点击打开商品详情抽屉，不触发 QuickBuy。</p>
              </div>
              <Button v-if="canEdit || canCreate" type="button" variant="outline" :disabled="saving || form.products.length >= 12" @click="openProductPicker">
                <Search class="size-3.5" />
                关联商品
              </Button>
            </div>

            <div v-if="form.products.length" class="grid gap-2 sm:grid-cols-2">
              <article v-for="(product, index) in form.products" :key="productKey(product)" class="flex min-w-0 items-center gap-3 rounded-lg border bg-muted/15 px-3 py-2.5">
                <Package class="size-4 shrink-0 text-emerald-600" />
                <div class="min-w-0 flex-1">
                  <p class="truncate text-xs font-black">{{ product.display_title }}</p>
                  <p class="mt-1 truncate font-mono text-[10px] text-muted-foreground">
                    {{ product.product_slug }} · {{ formatMoney(product.price, product.currency) }}
                  </p>
                </div>
                <Button
                  type="button"
                  variant="ghost"
                  size="icon-xs"
                  :disabled="saving"
                  :aria-label="`移除商品 ${product.display_title}`"
                  title="移除商品"
                  @click="form.products.splice(index, 1)"
                >
                  <X class="size-3.5" />
                </Button>
              </article>
            </div>
            <div v-else class="rounded-lg border border-dashed px-3 py-4 text-xs text-muted-foreground">
              暂未关联商品。可不选；如果选中，前台会在手记底部显示商品信息标签。
            </div>
          </section>

          <DialogFooter class="gap-2 sm:gap-2">
            <Button type="button" variant="outline" :disabled="saving" @click="editorOpen = false">取消</Button>
            <Button
              v-if="canCreate || canEdit"
              type="button"
              variant="outline"
              :disabled="saving"
              @click="save('draft')"
            >
              <LoaderCircle v-if="saving && saveIntent === 'draft'" class="size-4 animate-spin" />
              保存草稿
            </Button>
            <Button
              v-if="canCreate || canEdit"
              type="submit"
              :disabled="saving"
              @click="form.status = 'published'"
            >
              <LoaderCircle v-if="saving && saveIntent === 'published'" class="size-4 animate-spin" />
              发布手记
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>

    <Dialog v-model:open="productPickerOpen">
      <DialogContent size="lg">
        <DialogHeader>
          <DialogTitle>关联商品</DialogTitle>
          <DialogDescription>从当前可用的商品目录中搜索商品或指定变体。</DialogDescription>
        </DialogHeader>

        <div class="space-y-3">
          <div class="relative">
            <Search class="pointer-events-none absolute left-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground/60" />
            <Input
              v-model="productSearch"
              autofocus
              class="h-10 pl-9"
              placeholder="搜索商品名、slug、SKU 或变体名称"
              @update:model-value="scheduleProductSearch"
            />
          </div>
          <div class="max-h-[min(52vh,30rem)] overflow-y-auto rounded-lg border bg-background">
            <div v-if="productOptionsLoading" class="flex items-center gap-2 px-3 py-5 text-xs text-muted-foreground">
              <LoaderCircle class="size-4 animate-spin" />
              正在读取商品...
            </div>
            <div v-else-if="productOptions.length === 0" class="px-3 py-5 text-xs text-muted-foreground">
              {{ productSearch.trim() ? '没有找到可用商品' : '输入关键词开始搜索商品' }}
            </div>
            <div v-else class="divide-y">
              <button
                v-for="option in productOptions"
                :key="productOptionKey(option)"
                type="button"
                class="block w-full px-3 py-3 text-left transition-colors hover:bg-muted/60 disabled:opacity-45"
                :disabled="isProductSelected(option) || saving"
                @click="selectProduct(option)"
              >
                <span class="flex items-start justify-between gap-3">
                  <span class="min-w-0">
                    <span class="block truncate text-xs font-black">{{ option.product_name }}</span>
                    <span v-if="option.variant_title" class="mt-1 block truncate text-[10px] text-muted-foreground">{{ option.variant_title }}</span>
                    <span class="mt-1 block truncate font-mono text-[10px] text-muted-foreground">{{ option.product_slug }}</span>
                  </span>
                  <span class="shrink-0 text-right">
                    <span class="block font-mono text-xs font-black">{{ formatMoney(option.price, option.currency) }}</span>
                    <span v-if="isProductSelected(option)" class="mt-1 block text-[10px] font-bold text-emerald-600">已关联</span>
                  </span>
                </span>
              </button>
            </div>
          </div>
          <Button
            v-if="productOptionPagination.page < productOptionPagination.total_pages && productOptions.length"
            type="button"
            variant="outline"
            class="w-full"
            :disabled="productOptionsLoading"
            @click="loadMoreProductOptions"
          >
            <LoaderCircle v-if="productOptionsLoading" class="size-4 animate-spin" />
            继续加载商品
          </Button>
        </div>

        <DialogFooter>
          <Button type="button" variant="outline" @click="productPickerOpen = false">完成</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  </div>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, reactive, ref } from 'vue'
import { toast } from 'vue-sonner'
import {
  Archive,
  ClipboardList,
  FilePenLine,
  Images,
  LoaderCircle,
  MoreHorizontal,
  Package,
  Pencil,
  Plus,
  RotateCcw,
  Search,
  Send,
  Trash2,
  Upload,
  X,
} from '@lucide/vue'
import workbenchFeedApi, {
  type WorkbenchFeedEntry,
  type WorkbenchFeedEntryPayload,
  type WorkbenchFeedMedia,
  type WorkbenchFeedProductOption,
  type WorkbenchFeedStatus,
  type WorkbenchFeedTaggedProduct,
} from '@/api/workbenchFeed'
import AdminFilterPanel from '@/components/admin/AdminFilterPanel.vue'
import AdminFilterSelect from '@/components/admin/AdminFilterSelect.vue'
import AdminFormField from '@/components/admin/AdminFormField.vue'
import AdminPageHeader from '@/components/admin/AdminPageHeader.vue'
import AdminPagination from '@/components/admin/AdminPagination.vue'
import AdminTablePanel from '@/components/admin/AdminTablePanel.vue'
import StorefrontLocaleSelect from '@/components/admin/StorefrontLocaleSelect.vue'
import { Button } from '@/components/ui/button'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { Input } from '@/components/ui/input'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Table, TableBody, TableCell, TableEmpty, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Textarea } from '@/components/ui/textarea'
import { useSupportedLanguages } from '@/composables/useSupportedLanguages'
import { useAuthStore } from '@/stores/auth'

type StatusFilter = WorkbenchFeedStatus | 'all'

interface WorkbenchFeedForm {
  id?: number
  entry_number: string
  locale: string
  content: string
  tagsText: string
  status: WorkbenchFeedStatus
  media: WorkbenchFeedMedia[]
  products: WorkbenchFeedTaggedProduct[]
}

const authStore = useAuthStore()
const canCreate = computed(() => authStore.hasPermission('workbench_feed:create'))
const canEdit = computed(() => authStore.hasPermission('workbench_feed:edit'))
const canDelete = computed(() => authStore.hasPermission('workbench_feed:delete'))
const supportedLanguages = useSupportedLanguages()
const languageOptions = supportedLanguages.languageOptions
const localeFilterOptions = supportedLanguages.localeFilterOptions

const entries = ref<WorkbenchFeedEntry[]>([])
const loading = ref(false)
const editorOpen = ref(false)
const productPickerOpen = ref(false)
const saving = ref(false)
const saveIntent = ref<WorkbenchFeedStatus | null>(null)
const uploadingMedia = ref(false)
const dragOver = ref(false)
const filters = reactive<{ search: string; status: StatusFilter; locale: string }>({
  search: '',
  status: 'all',
  locale: 'all',
})
const pagination = reactive({ page: 1, page_size: 20, total: 0, total_pages: 0 })
const form = reactive<WorkbenchFeedForm>(emptyForm())
const productSearch = ref('')
const productOptions = ref<WorkbenchFeedProductOption[]>([])
const productOptionsLoading = ref(false)
const productOptionPagination = reactive({ page: 1, page_size: 20, total: 0, total_pages: 0 })
const productRequestVersion = ref(0)
let productSearchTimer: ReturnType<typeof setTimeout> | undefined

function emptyForm(): WorkbenchFeedForm {
  return {
    entry_number: '',
    locale: 'en',
    content: '',
    tagsText: '',
    status: 'draft',
    media: [],
    products: [],
  }
}

const splitTags = (value: string): string[] => value
  .split(/[\n,]/)
  .map(tag => tag.trim())
  .filter(Boolean)

const assignForm = (entry?: WorkbenchFeedEntry) => {
  Object.assign(form, emptyForm())
  if (!entry) return

  Object.assign(form, {
    id: entry.id,
    entry_number: entry.entry_number,
    locale: entry.locale,
    content: entry.content,
    tagsText: entry.tags.join(', '),
    status: entry.status,
    media: entry.media.map(media => ({ ...media })),
    products: entry.tagged_products.map(product => ({ ...product })),
  })
}

const listParams = () => ({
  page: pagination.page,
  page_size: pagination.page_size,
  search: filters.search.trim() || undefined,
  locale: filters.locale === 'all' ? undefined : filters.locale,
  status: filters.status === 'all' ? undefined : filters.status,
})

const loadEntries = async () => {
  loading.value = true
  try {
    const result = await workbenchFeedApi.list(listParams())
    entries.value = result.entries
    Object.assign(pagination, result.pagination)
  } catch (error) {
    toast.error(error instanceof Error ? error.message : '车间手记加载失败')
  } finally {
    loading.value = false
  }
}

const applyFilters = () => {
  pagination.page = 1
  void loadEntries()
}

const resetFilters = () => {
  filters.search = ''
  filters.status = 'all'
  filters.locale = 'all'
  applyFilters()
}

const updatePage = (page: number) => {
  pagination.page = page
  void loadEntries()
}

const updatePageSize = (pageSize: number) => {
  pagination.page_size = pageSize
  pagination.page = 1
  void loadEntries()
}

const openCreate = () => {
  assignForm()
  editorOpen.value = true
}

const openEdit = async (entry: WorkbenchFeedEntry) => {
  try {
    assignForm(await workbenchFeedApi.get(entry.id))
    editorOpen.value = true
  } catch (error) {
    toast.error(error instanceof Error ? error.message : '车间手记详情加载失败')
  }
}

const uploadFiles = async (files: File[]) => {
  const remaining = 4 - form.media.length
  const selected = files
    .filter(file => file.type.startsWith('image/'))
    .slice(0, Math.max(0, remaining))

  if (!selected.length) {
    if (files.length && remaining <= 0) toast.error('每条手记最多上传 4 张图片')
    else if (files.length) toast.error('请选择图片文件')
    return
  }
  if (selected.length < files.length) toast.message(`只上传前 ${selected.length} 张图片`)

  uploadingMedia.value = true
  try {
    for (const file of selected) {
      const media = await workbenchFeedApi.uploadMedia(file)
      form.media.push({ ...media, sort_order: form.media.length })
    }
    toast.success(`${selected.length} 张图片已上传`)
  } catch (error) {
    toast.error(error instanceof Error ? error.message : '图片上传失败')
  } finally {
    uploadingMedia.value = false
  }
}

const onMediaFileChange = (event: Event) => {
  const target = event.target as HTMLInputElement | null
  const files = target?.files ? Array.from(target.files) : []
  if (files.length) void uploadFiles(files)
  if (target) target.value = ''
}

const onMediaDrop = (event: DragEvent) => {
  dragOver.value = false
  const files = event.dataTransfer?.files ? Array.from(event.dataTransfer.files) : []
  if (files.length) void uploadFiles(files)
}

const removeMedia = (index: number) => {
  form.media.splice(index, 1)
  form.media.forEach((media, mediaIndex) => {
    media.sort_order = mediaIndex
  })
}

const productOptionKey = (option: WorkbenchFeedProductOption) => `${option.product_id}:${option.variant_id || 0}`
const productKey = (product: Pick<WorkbenchFeedTaggedProduct, 'product_id' | 'variant_id'>) => `${product.product_id}:${product.variant_id || 0}`
const isProductSelected = (option: WorkbenchFeedProductOption) => form.products.some(product => productKey(product) === productOptionKey(option))

const loadProductOptions = async (reset = true) => {
  if (reset) {
    productOptionPagination.page = 1
    productOptions.value = []
  }
  const requestVersion = productRequestVersion.value + 1
  productRequestVersion.value = requestVersion
  productOptionsLoading.value = true
  try {
    const result = await workbenchFeedApi.listProductOptions({
      page: productOptionPagination.page,
      page_size: productOptionPagination.page_size,
      search: productSearch.value.trim() || undefined,
    })
    if (requestVersion !== productRequestVersion.value) return
    productOptions.value = reset ? result.options : [...productOptions.value, ...result.options]
    Object.assign(productOptionPagination, result.pagination)
  } catch (error) {
    if (requestVersion !== productRequestVersion.value) return
    toast.error(error instanceof Error ? error.message : '商品目录读取失败')
  } finally {
    if (requestVersion === productRequestVersion.value) productOptionsLoading.value = false
  }
}

const openProductPicker = () => {
  if (productSearchTimer) {
    clearTimeout(productSearchTimer)
    productSearchTimer = undefined
  }
  productPickerOpen.value = true
  productSearch.value = ''
  productOptions.value = []
  productOptionPagination.page = 1
  productOptionPagination.total = 0
  productOptionPagination.total_pages = 0
  void loadProductOptions(true)
}

const scheduleProductSearch = () => {
  if (productSearchTimer) clearTimeout(productSearchTimer)
  productSearchTimer = setTimeout(() => {
    void loadProductOptions(true)
  }, 250)
}

const loadMoreProductOptions = async () => {
  if (productOptionPagination.page >= productOptionPagination.total_pages) return
  productOptionPagination.page += 1
  await loadProductOptions(false)
}

const selectProduct = (option: WorkbenchFeedProductOption) => {
  if (isProductSelected(option) || form.products.length >= 12) return
  form.products.push({
    id: 0,
    product_id: option.product_id,
    variant_id: option.variant_id || null,
    product_slug: option.product_slug,
    display_title: option.variant_title && option.variant_title !== option.product_name
      ? `${option.product_name} · ${option.variant_title}`
      : option.product_name,
    price: option.price,
    currency: option.currency,
    direct_action: 'detail_drawer',
    available: option.available,
    sort_order: form.products.length,
  })
}

const toPayload = (status: WorkbenchFeedStatus): WorkbenchFeedEntryPayload => ({
  locale: form.locale.trim().toLowerCase() || 'en',
  content: form.content.trim(),
  tags: splitTags(form.tagsText),
  status,
  media: form.media.map((media, index) => ({
    file_path: media.file_path,
    width: media.width,
    height: media.height,
    file_size_kb: media.file_size_kb,
    caption: media.caption.trim(),
    sort_order: index,
  })),
  tagged_products: form.products.map((product, index) => ({
    product_id: product.product_id,
    variant_id: product.variant_id || null,
    direct_action: 'detail_drawer',
    sort_order: index,
  })),
})

const save = async (status: WorkbenchFeedStatus) => {
  if (!canEdit.value && !(!form.id && canCreate.value)) return
  const payload = toPayload(status)
  if (!payload.content) {
    toast.error('请填写车间现场记录')
    return
  }
  if (payload.content.length > 300) {
    toast.error('现场记录不能超过 300 个字符')
    return
  }
  if (payload.tags.length > 12) {
    toast.error('最多添加 12 个话题标签')
    return
  }

  saving.value = true
  saveIntent.value = status
  try {
    if (form.id) {
      await workbenchFeedApi.update(form.id, payload)
      toast.success(status === 'published' ? '车间手记已发布' : '车间手记已保存')
    } else {
      await workbenchFeedApi.create(payload)
      toast.success(status === 'published' ? '车间手记已发布' : '车间手记草稿已保存')
    }
    editorOpen.value = false
    await loadEntries()
  } catch (error) {
    toast.error(error instanceof Error ? error.message : '车间手记保存失败')
  } finally {
    saving.value = false
    saveIntent.value = null
  }
}

const setStatus = async (entry: WorkbenchFeedEntry, status: WorkbenchFeedStatus) => {
  try {
    await workbenchFeedApi.updateStatus(entry.id, status)
    toast.success(`${entry.entry_number} 已${statusLabel(status)}`)
    await loadEntries()
  } catch (error) {
    toast.error(error instanceof Error ? error.message : '状态更新失败')
  }
}

const removeEntry = async (entry: WorkbenchFeedEntry) => {
  if (!window.confirm(`确定删除 ${entry.entry_number} 吗？这不会删除已上传的媒体资产。`)) return
  try {
    await workbenchFeedApi.remove(entry.id)
    toast.success('车间手记已删除')
    if (entries.value.length === 1 && pagination.page > 1) pagination.page -= 1
    await loadEntries()
  } catch (error) {
    toast.error(error instanceof Error ? error.message : '车间手记删除失败')
  }
}

const statusLabel = (status: WorkbenchFeedStatus) => ({
  draft: '草稿',
  published: '已发布',
  archived: '已归档',
}[status])

const statusClass = (status: WorkbenchFeedStatus) => ({
  draft: 'bg-amber-500/10 text-amber-700 dark:text-amber-300',
  published: 'bg-emerald-500/10 text-emerald-700 dark:text-emerald-300',
  archived: 'bg-muted text-muted-foreground',
}[status])

const formatDate = (value?: string) => {
  if (!value) return '-'
  const date = new Date(value)
  return Number.isNaN(date.getTime()) ? '-' : date.toLocaleString('zh-CN', { dateStyle: 'medium', timeStyle: 'short' })
}

const formatMoney = (amount: number, currency: string) => {
  try {
    return new Intl.NumberFormat('zh-CN', { style: 'currency', currency }).format(Number(amount || 0))
  } catch {
    return `${currency || 'USD'} ${Number(amount || 0).toFixed(2)}`
  }
}

onMounted(() => {
  void Promise.all([supportedLanguages.fetchLanguages(), loadEntries()])
})

onBeforeUnmount(() => {
  if (productSearchTimer) clearTimeout(productSearchTimer)
})
</script>
