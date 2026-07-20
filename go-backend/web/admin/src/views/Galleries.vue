<template>
  <div class="space-y-4">
    <AdminPageHeader title="图库管理" description="管理图片集合、封面和图库内素材">
      <template #actions>
        <Button v-if="hasPermission('gallery:create')" @click="showCreateDialog">
          <Plus class="size-4" />
          创建图库
        </Button>
      </template>
    </AdminPageHeader>

    <AdminTablePanel :loading="loading">
      <Table class="min-w-[940px]">
        <TableHeader>
          <TableRow>
            <TableHead class="w-16">ID</TableHead>
            <TableHead class="w-28">封面</TableHead>
            <TableHead>图库</TableHead>
            <TableHead>描述</TableHead>
            <TableHead class="w-24 text-right">图片数</TableHead>
            <TableHead class="w-44">创建时间</TableHead>
            <TableHead class="w-16 text-right">操作</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          <TableEmpty v-if="galleries.length === 0" :colspan="7">
            <div class="flex flex-col items-center text-muted-foreground">
              <Images class="mb-2 size-7 opacity-55" />
              <span class="text-xs">暂无图库</span>
            </div>
          </TableEmpty>

          <TableRow v-for="gallery in galleries" :key="gallery.id">
            <TableCell class="font-mono text-xs text-muted-foreground">{{ gallery.id }}</TableCell>
            <TableCell>
              <button
                v-if="galleryCover(gallery)"
                type="button"
                class="block size-[72px] overflow-hidden rounded-md border bg-muted focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
                :aria-label="`预览图库 ${galleryTitle(gallery)} 的封面`"
                @click="showImagePreview(galleryCover(gallery), galleryTitle(gallery))"
              >
                <img :src="galleryCover(gallery)" :alt="galleryTitle(gallery)" class="size-full object-cover" />
              </button>
              <div v-else class="flex size-[72px] items-center justify-center rounded-md border bg-muted text-muted-foreground">
                <ImageIcon class="size-5 opacity-50" />
              </div>
            </TableCell>
            <TableCell>
              <button type="button" class="block max-w-72 text-left" @click="viewImages(gallery)">
                <span class="block truncate font-medium hover:text-primary">{{ galleryTitle(gallery) }}</span>
                <span class="mt-1 block truncate font-mono text-xs text-muted-foreground">{{ gallery.slug || '-' }}</span>
              </button>
            </TableCell>
            <TableCell class="max-w-80">
              <p class="line-clamp-2 text-muted-foreground">{{ gallery.description || '-' }}</p>
            </TableCell>
            <TableCell class="text-right tabular-nums">{{ galleryImageCount(gallery) }}</TableCell>
            <TableCell class="text-xs text-muted-foreground">{{ formatDate(gallery.created_at) }}</TableCell>
            <TableCell class="text-right">
              <DropdownMenu>
                <DropdownMenuTrigger as-child>
                  <Button variant="ghost" size="icon" :aria-label="`管理图库 ${galleryTitle(gallery)}`">
                    <MoreHorizontal class="size-4" />
                  </Button>
                </DropdownMenuTrigger>
                <DropdownMenuContent align="end" class="w-40">
                  <DropdownMenuItem @select="viewImages(gallery)">
                    <Images class="size-4" />
                    查看图片
                  </DropdownMenuItem>
                  <DropdownMenuItem v-if="hasPermission('gallery:edit')" @select="showEditDialog(gallery)">
                    <Pencil class="size-4" />
                    编辑图库
                  </DropdownMenuItem>
                  <DropdownMenuSeparator v-if="hasPermission('gallery:delete')" />
                  <DropdownMenuItem
                    v-if="hasPermission('gallery:delete')"
                    class="text-destructive focus:text-destructive"
                    @select="requestDeleteGallery(gallery)"
                  >
                    <Trash2 class="size-4" />
                    删除图库
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
      <DialogContent size="md" @open-auto-focus.prevent>
        <form class="space-y-5" @submit.prevent="submitGalleryForm">
          <DialogHeader>
            <DialogTitle>{{ dialogMode === 'create' ? '创建图库' : '编辑图库' }}</DialogTitle>
            <DialogDescription>图库的封面由其中首张排序图片自动提供。</DialogDescription>
          </DialogHeader>

          <FormField label="标题" required :error="galleryErrors.title">
            <Input v-model="galleryForm.title" placeholder="请输入图库标题" @input="clearGalleryError('title')" />
          </FormField>
          <FormField label="Slug" required :error="galleryErrors.slug">
            <Input v-model="galleryForm.slug" placeholder="例如 customer-stories" @input="clearGalleryError('slug')" />
          </FormField>
          <FormField label="描述">
            <Textarea v-model="galleryForm.description" class="min-h-24" placeholder="请输入图库描述" />
          </FormField>

          <DialogFooter>
            <Button type="button" variant="outline" @click="dialogVisible = false">取消</Button>
            <Button type="submit" :disabled="submitting">
              <LoaderCircle v-if="submitting" class="size-4 animate-spin" />
              {{ submitting ? '保存中' : '保存图库' }}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>

    <Dialog v-model:open="imagesDialogVisible">
      <DialogContent size="xl" class="max-h-[92dvh] overflow-y-auto p-0" @open-auto-focus.prevent>
        <DialogHeader class="border-b px-5 py-4 pr-12">
          <DialogTitle>{{ currentGallery ? galleryTitle(currentGallery) : '图库图片' }}</DialogTitle>
          <DialogDescription>查看和维护图库内图片、缩略图、标签及展示顺序。</DialogDescription>
        </DialogHeader>

        <div class="space-y-4 px-5 py-5">
          <div class="flex flex-wrap items-center justify-between gap-2">
            <span class="text-xs text-muted-foreground">共 {{ images.length }} 张图片</span>
            <Button v-if="hasPermission('gallery:create')" size="sm" @click="showAddImageDialog">
              <Plus class="size-3.5" />
              添加图片
            </Button>
          </div>

          <AdminTablePanel :loading="imagesLoading" :batch-visible="selectedImages.length > 0">
            <template #batch>
              <div class="flex flex-wrap items-center justify-between gap-2">
                <span class="text-xs font-medium">已选择 {{ selectedImages.length }} 张图片</span>
                <Button v-if="hasPermission('gallery:delete')" variant="destructive" size="sm" @click="requestBatchDeleteImages">
                  <Trash2 class="size-3.5" />
                  批量删除
                </Button>
              </div>
            </template>

            <Table class="min-w-[920px]">
              <TableHeader>
                <TableRow>
                  <TableHead class="w-11">
                    <Checkbox
                      :model-value="imageSelectionState"
                      aria-label="选择当前图库全部图片"
                      @update:model-value="toggleAllImages"
                    />
                  </TableHead>
                  <TableHead class="w-28">图片</TableHead>
                  <TableHead>标题</TableHead>
                  <TableHead>描述</TableHead>
                  <TableHead class="w-40">标签</TableHead>
                  <TableHead class="w-20 text-right">排序</TableHead>
                  <TableHead class="w-16 text-right">操作</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                <TableEmpty v-if="images.length === 0" :colspan="7">
                  <div class="flex flex-col items-center text-muted-foreground">
                    <ImageOff class="mb-2 size-7 opacity-55" />
                    <span class="text-xs">该图库暂无图片</span>
                  </div>
                </TableEmpty>

                <TableRow v-for="image in images" :key="image.id">
                  <TableCell>
                    <Checkbox
                      :model-value="isImageSelected(image.id)"
                      :aria-label="`选择图片 ${image.title}`"
                      @update:model-value="toggleImage(image, $event)"
                    />
                  </TableCell>
                  <TableCell>
                    <button
                      type="button"
                      class="block h-16 w-20 overflow-hidden rounded-md border bg-muted focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
                      :aria-label="`预览图片 ${image.title}`"
                      @click="showImagePreview(image.url, image.title)"
                    >
                      <img :src="image.thumbnail || image.url" :alt="image.title" class="size-full object-cover" />
                    </button>
                  </TableCell>
                  <TableCell class="max-w-64 font-medium">{{ image.title || '-' }}</TableCell>
                  <TableCell class="max-w-72"><p class="line-clamp-2 text-muted-foreground">{{ image.description || '-' }}</p></TableCell>
                  <TableCell class="max-w-40 truncate text-xs text-muted-foreground">{{ image.tags || '-' }}</TableCell>
                  <TableCell class="text-right tabular-nums">{{ image.order ?? image.sort_order ?? 0 }}</TableCell>
                  <TableCell class="text-right">
                    <DropdownMenu>
                      <DropdownMenuTrigger as-child>
                        <Button variant="ghost" size="icon" :aria-label="`管理图片 ${image.title}`">
                          <MoreHorizontal class="size-4" />
                        </Button>
                      </DropdownMenuTrigger>
                      <DropdownMenuContent align="end" class="w-36">
                        <DropdownMenuItem v-if="hasPermission('gallery:edit')" @select="showEditImageDialog(image)">
                          <Pencil class="size-4" />
                          编辑
                        </DropdownMenuItem>
                        <DropdownMenuSeparator v-if="hasPermission('gallery:delete')" />
                        <DropdownMenuItem
                          v-if="hasPermission('gallery:delete')"
                          class="text-destructive focus:text-destructive"
                          @select="requestDeleteImage(image)"
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
          </AdminTablePanel>
        </div>
      </DialogContent>
    </Dialog>

    <Dialog v-model:open="imageDialogVisible">
      <DialogContent size="md" class="max-h-[90dvh] overflow-y-auto" @open-auto-focus.prevent>
        <form class="space-y-5" @submit.prevent="submitImageForm">
          <DialogHeader>
            <DialogTitle>{{ imageDialogMode === 'create' ? '添加图片' : '编辑图片' }}</DialogTitle>
            <DialogDescription>维护原图、缩略图和用于检索的图片信息。</DialogDescription>
          </DialogHeader>

          <div v-if="imageForm.url" class="flex h-40 items-center justify-center overflow-hidden rounded-lg border bg-muted">
            <img :src="imageForm.thumbnail || imageForm.url" :alt="imageForm.title || '图片预览'" class="size-full object-contain" />
          </div>

          <div class="grid gap-4 sm:grid-cols-2">
            <FormField label="图片 URL" required :error="imageErrors.url" class="sm:col-span-2">
              <Input v-model="imageForm.url" type="url" placeholder="https://example.com/image.jpg" @input="clearImageError('url')" />
            </FormField>
            <FormField label="缩略图 URL" class="sm:col-span-2">
              <Input v-model="imageForm.thumbnail" type="url" placeholder="可选" />
            </FormField>
            <FormField label="标题" required :error="imageErrors.title">
              <Input v-model="imageForm.title" placeholder="请输入图片标题" @input="clearImageError('title')" />
            </FormField>
            <FormField label="排序">
              <Input v-model.number="imageForm.order" type="number" min="0" max="9999" step="1" />
            </FormField>
            <FormField label="标签" class="sm:col-span-2">
              <Input v-model="imageForm.tags" placeholder="多个标签用逗号分隔" />
            </FormField>
            <FormField label="描述" class="sm:col-span-2">
              <Textarea v-model="imageForm.description" class="min-h-24" placeholder="请输入图片描述" />
            </FormField>
          </div>

          <DialogFooter>
            <Button type="button" variant="outline" @click="imageDialogVisible = false">取消</Button>
            <Button type="submit" :disabled="imageSubmitting">
              <LoaderCircle v-if="imageSubmitting" class="size-4 animate-spin" />
              {{ imageSubmitting ? '保存中' : '保存图片' }}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>

    <Dialog v-model:open="previewDialogVisible">
      <DialogContent size="lg" class="overflow-hidden p-0">
        <DialogHeader class="sr-only">
          <DialogTitle>{{ previewImage.title || '图片预览' }}</DialogTitle>
          <DialogDescription>图库图片大图预览</DialogDescription>
        </DialogHeader>
        <div class="flex max-h-[85vh] min-h-64 items-center justify-center bg-black/95 p-3">
          <img :src="previewImage.url" :alt="previewImage.title" class="max-h-[80vh] max-w-full object-contain" />
        </div>
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
import { Image as ImageIcon, ImageOff, Images, LoaderCircle, MoreHorizontal, Pencil, Plus, Trash2 } from '@lucide/vue'
import AdminConfirmDialog from '@/components/admin/AdminConfirmDialog.vue'
import AdminPageHeader from '@/components/admin/AdminPageHeader.vue'
import AdminPagination from '@/components/admin/AdminPagination.vue'
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
import { Table, TableBody, TableCell, TableEmpty, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Textarea } from '@/components/ui/textarea'
import { useAuthStore } from '@/stores/auth'
import axios from '@/utils/axios'

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
const galleries = ref([])
const dialogVisible = ref(false)
const dialogMode = ref('create')
const submitting = ref(false)
const galleryErrors = reactive({})

const imagesDialogVisible = ref(false)
const imagesLoading = ref(false)
const images = ref([])
const currentGallery = ref(null)
const selectedImages = ref([])

const imageDialogVisible = ref(false)
const imageDialogMode = ref('create')
const imageSubmitting = ref(false)
const imageErrors = reactive({})
const previewDialogVisible = ref(false)
const previewImage = reactive({ url: '', title: '' })

const pagination = reactive({ page: 1, pageSize: 20, total: 0 })
const galleryForm = reactive({ id: null, title: '', slug: '', description: '' })
const imageForm = reactive({ id: null, url: '', thumbnail: '', title: '', description: '', tags: '', order: 0 })
const confirmation = reactive({
  open: false,
  type: '',
  target: null,
  title: '',
  description: '',
  confirmLabel: '删除'
})

const imageSelectionState = computed(() => {
  if (images.value.length === 0 || selectedImages.value.length === 0) return false
  return selectedImages.value.length === images.value.length ? true : 'indeterminate'
})

const hasPermission = (permission) => authStore.hasPermission(permission)
const galleryTitle = (gallery) => gallery?.name || gallery?.title || '-'
const galleryCover = (gallery) => gallery?.cover_image || gallery?.images?.[0]?.thumbnail || gallery?.images?.[0]?.url || ''
const galleryImageCount = (gallery) => gallery?.image_count ?? gallery?.images_count ?? '-'
const formatDate = (dateString) => dateString ? new Date(dateString).toLocaleString('zh-CN') : '-'

const clearErrors = (errors) => Object.keys(errors).forEach((key) => delete errors[key])
const clearGalleryError = (field) => { delete galleryErrors[field] }
const clearImageError = (field) => { delete imageErrors[field] }
const resetGalleryForm = () => {
  Object.assign(galleryForm, { id: null, title: '', slug: '', description: '' })
  clearErrors(galleryErrors)
}
const resetImageForm = () => {
  Object.assign(imageForm, { id: null, url: '', thumbnail: '', title: '', description: '', tags: '', order: 0 })
  clearErrors(imageErrors)
}
const buildGalleryPayload = () => ({
  title: galleryForm.title.trim(),
  slug: galleryForm.slug.trim(),
  description: galleryForm.description.trim()
})
const buildImagePayload = () => ({
  url: imageForm.url.trim(),
  thumbnail: imageForm.thumbnail.trim(),
  title: imageForm.title.trim(),
  description: imageForm.description.trim(),
  tags: imageForm.tags.trim(),
  order: Math.max(0, Number(imageForm.order || 0))
})
const validateGallery = (payload) => {
  clearErrors(galleryErrors)
  if (!payload.title) galleryErrors.title = '请输入图库标题'
  if (!payload.slug) galleryErrors.slug = '请输入图库 Slug'
  if (Object.keys(galleryErrors).length) {
    toast.error('请检查图库表单中的必填项')
    return false
  }
  return true
}
const validateImage = (payload) => {
  clearErrors(imageErrors)
  if (!payload.url) imageErrors.url = '请输入图片 URL'
  if (!payload.title) imageErrors.title = '请输入图片标题'
  if (Object.keys(imageErrors).length) {
    toast.error('请检查图片表单中的必填项')
    return false
  }
  return true
}

const fetchGalleries = async () => {
  loading.value = true
  try {
    const response = await axios.get('/api/admin/galleries', {
      params: { page: pagination.page, page_size: pagination.pageSize }
    })
    galleries.value = response.data.galleries || []
    pagination.total = response.data.pagination?.total ?? response.data.total ?? 0
  } catch (error) {
    console.error('Failed to fetch galleries:', error)
  } finally {
    loading.value = false
  }
}
const updatePage = (page) => { pagination.page = page; fetchGalleries() }
const updatePageSize = (pageSize) => { pagination.pageSize = pageSize; pagination.page = 1; fetchGalleries() }

const showCreateDialog = () => {
  dialogMode.value = 'create'
  resetGalleryForm()
  dialogVisible.value = true
}
const showEditDialog = async (gallery) => {
  dialogMode.value = 'edit'
  try {
    const response = await axios.get(`/api/admin/galleries/${gallery.id}`)
    const detail = response.data.gallery || gallery
    Object.assign(galleryForm, {
      id: detail.id,
      title: galleryTitle(detail),
      slug: detail.slug || '',
      description: detail.description || ''
    })
    clearErrors(galleryErrors)
    dialogVisible.value = true
  } catch (error) {
    console.error('Failed to fetch gallery detail:', error)
  }
}
const submitGalleryForm = async () => {
  const payload = buildGalleryPayload()
  if (!validateGallery(payload)) return
  submitting.value = true
  try {
    if (dialogMode.value === 'create') {
      await axios.post('/api/admin/galleries', payload)
      toast.success('图库创建成功')
    } else {
      await axios.put(`/api/admin/galleries/${galleryForm.id}`, payload)
      toast.success('图库更新成功')
    }
    dialogVisible.value = false
    await fetchGalleries()
  } catch (error) {
    console.error('Failed to save gallery:', error)
  } finally {
    submitting.value = false
  }
}

const viewImages = async (gallery) => {
  currentGallery.value = gallery
  images.value = []
  selectedImages.value = []
  imagesDialogVisible.value = true
  await fetchImages(gallery.id)
}
const fetchImages = async (galleryId) => {
  imagesLoading.value = true
  try {
    const response = await axios.get(`/api/admin/galleries/${galleryId}/images`)
    images.value = response.data.images || []
    selectedImages.value = []
  } catch (error) {
    console.error('Failed to fetch gallery images:', error)
  } finally {
    imagesLoading.value = false
  }
}
const showAddImageDialog = () => {
  imageDialogMode.value = 'create'
  resetImageForm()
  imageDialogVisible.value = true
}
const showEditImageDialog = (image) => {
  imageDialogMode.value = 'edit'
  Object.assign(imageForm, {
    id: image.id,
    url: image.url || '',
    thumbnail: image.thumbnail || '',
    title: image.title || '',
    description: image.description || '',
    tags: image.tags || '',
    order: image.order ?? image.sort_order ?? 0
  })
  clearErrors(imageErrors)
  imageDialogVisible.value = true
}
const submitImageForm = async () => {
  const payload = buildImagePayload()
  if (!validateImage(payload)) return
  imageSubmitting.value = true
  try {
    if (imageDialogMode.value === 'create') {
      await axios.post(`/api/admin/galleries/${currentGallery.value.id}/images`, payload)
      toast.success('图片添加成功')
    } else {
      await axios.put(`/api/admin/galleries/${currentGallery.value.id}/images/${imageForm.id}`, payload)
      toast.success('图片更新成功')
    }
    imageDialogVisible.value = false
    await Promise.all([fetchImages(currentGallery.value.id), fetchGalleries()])
  } catch (error) {
    console.error('Failed to save gallery image:', error)
  } finally {
    imageSubmitting.value = false
  }
}

const showImagePreview = (url, title) => {
  Object.assign(previewImage, { url, title: title || '图片预览' })
  previewDialogVisible.value = true
}
const isImageSelected = (imageId) => selectedImages.value.some((image) => image.id === imageId)
const toggleAllImages = (checked) => { selectedImages.value = checked === true ? [...images.value] : [] }
const toggleImage = (image, checked) => {
  if (checked === true && !isImageSelected(image.id)) selectedImages.value = [...selectedImages.value, image]
  else if (checked !== true) selectedImages.value = selectedImages.value.filter((selected) => selected.id !== image.id)
}
const setConfirmation = (values) => Object.assign(confirmation, {
  open: true,
  type: '',
  target: null,
  confirmLabel: '删除',
  ...values
})
const requestDeleteGallery = (gallery) => setConfirmation({
  type: 'delete-gallery', target: gallery, title: '删除图库？',
  description: `图库“${galleryTitle(gallery)}”及其中图片将被永久删除，此操作不可恢复。`
})
const requestDeleteImage = (image) => setConfirmation({
  type: 'delete-image', target: image, title: '删除图片？',
  description: `图片“${image.title || image.id}”将被永久删除，此操作不可恢复。`
})
const requestBatchDeleteImages = () => setConfirmation({
  type: 'batch-delete-images', target: [...selectedImages.value], title: '批量删除图片？',
  description: `${selectedImages.value.length} 张图片将被永久删除，此操作不可恢复。`, confirmLabel: '批量删除'
})
const executeConfirmedAction = async () => {
  const { type, target } = confirmation
  confirmation.open = false
  try {
    if (type === 'delete-gallery') {
      await axios.delete(`/api/admin/galleries/${target.id}`)
      toast.success('图库已删除')
      await fetchGalleries()
    } else if (type === 'delete-image') {
      await axios.delete(`/api/admin/galleries/${currentGallery.value.id}/images/${target.id}`)
      toast.success('图片已删除')
      await Promise.all([fetchImages(currentGallery.value.id), fetchGalleries()])
    } else if (type === 'batch-delete-images') {
      await axios.post(`/api/admin/galleries/${currentGallery.value.id}/images/batch-delete`, {
        image_ids: target.map((image) => image.id)
      })
      toast.success('图片已批量删除')
      await Promise.all([fetchImages(currentGallery.value.id), fetchGalleries()])
    }
  } catch (error) {
    console.error('Failed to delete gallery content:', error)
  }
}

onMounted(fetchGalleries)
</script>
