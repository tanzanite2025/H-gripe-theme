<template>
  <div class="space-y-3">
    <AdminPageHeader
      title="我与这个网站"
      description="当前语言设置优先于 English fallback，未配置时才使用页面默认内容；标记为“全局”的资源会影响所有语言。"
    >
      <template #actions>
        <Button :disabled="loading || saving || !canEdit" @click="saveProfile">
          <LoaderCircle v-if="saving" class="size-4 animate-spin" />
          <Save v-else class="size-4" />
          {{ saving ? '保存中' : '保存设置' }}
        </Button>
      </template>
    </AdminPageHeader>

    <div class="relative space-y-4">
      <div
        v-if="loading"
        class="absolute inset-0 z-10 flex items-start justify-center bg-background/75 pt-24"
      >
        <LoaderCircle class="size-5 animate-spin text-primary" aria-label="正在加载设置" />
      </div>

      <section class="border-b border-dashed border-border/80 pb-4">
        <div class="grid gap-3 lg:grid-cols-[minmax(0,1fr)_18rem] lg:items-end">
          <AdminFormField
            label="内容编辑语言"
            description="这里只决定正在编辑哪一套页面文案，不代表市场范围、语言范围或站点全局语言配置。"
          >
            <Select v-model="contentLocale" :disabled="loading || saving">
              <SelectTrigger class="max-w-md">
                <SelectValue placeholder="选择内容语言" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem
                  v-for="option in languageOptions"
                  :key="option.value"
                  :value="option.value"
                >
                  {{ option.label }}
                </SelectItem>
              </SelectContent>
            </Select>
          </AdminFormField>

          <div class="rounded-2xl border border-dashed border-border/80 bg-muted/25 px-3 py-2.5">
            <p class="text-[9px] font-black uppercase tracking-widest text-muted-foreground/60">当前编辑对象</p>
            <p class="mt-0.5 text-sm font-black text-foreground">{{ selectedLanguageLabel }}</p>
            <p class="mt-0.5 text-[10px] leading-relaxed text-muted-foreground">
              头像、工厂图片和工厂入口为全局资源，不会因切换语言而变成不同市场配置。
            </p>
          </div>
        </div>
      </section>

      <Tabs v-model="activeSection" class="space-y-3">
        <TabsList class="h-9 w-full max-w-2xl overflow-x-auto rounded-xl bg-muted/50 p-1">
          <TabsTrigger value="identity" class="min-w-28">页面身份</TabsTrigger>
          <TabsTrigger value="statement" class="min-w-28">页面说明</TabsTrigger>
          <TabsTrigger value="factory" class="min-w-28">工厂关联</TabsTrigger>
        </TabsList>

        <TabsContent value="identity">
          <section class="grid w-full max-w-none gap-4 lg:grid-cols-[160px_minmax(0,1fr)]">
            <div>
              <h2 class="text-sm font-black tracking-tighter uppercase text-foreground">页面身份</h2>
              <p class="mt-1 text-[9px] font-black uppercase tracking-widest text-muted-foreground/60">
                首屏说明与头像信息
              </p>
            </div>
            <div class="min-w-0 space-y-4">
              <div class="grid gap-3 md:grid-cols-2">
                <AdminFormField label="眉题">
                  <Input v-model="form.eyebrow" :disabled="!canEdit" />
                </AdminFormField>
                <AdminFormField label="页面标题">
                  <Input v-model="form.title" :disabled="!canEdit" />
                </AdminFormField>
                <AdminFormField label="页面简介" class="md:col-span-2">
                  <Textarea v-model="form.lead" class="min-h-20" :disabled="!canEdit" />
                </AdminFormField>
                <AdminFormField label="范围说明" class="md:col-span-2">
                  <Input v-model="form.scope" :disabled="!canEdit" />
                </AdminFormField>
              </div>

              <div class="grid gap-3 lg:grid-cols-[minmax(0,1fr)_12rem] lg:items-start">
                <div class="grid gap-3 md:grid-cols-2">
                  <AdminFormField label="头像 URL（全局）">
                    <div class="flex min-w-0 items-center gap-2">
                      <Input v-model="form.avatar_url" :disabled="!canEdit || uploading === 'avatar'" />
                      <Button
                        type="button"
                        variant="outline"
                        size="icon"
                        :disabled="!canEdit || uploading === 'avatar'"
                        title="上传头像"
                        @click="chooseUpload('avatar')"
                      >
                        <LoaderCircle v-if="uploading === 'avatar'" class="size-4 animate-spin" />
                        <ImagePlus v-else class="size-4" />
                      </Button>
                    </div>
                    <input
                      ref="avatarInput"
                      type="file"
                      class="hidden"
                      :accept="uploadSpecAccept('website_profile_avatar')"
                      :disabled="!canEdit || uploading === 'avatar'"
                      @change="handleUpload('avatar', $event)"
                    />
                    <UploadSpecHint code="website_profile_avatar" />
                  </AdminFormField>
                  <AdminFormField label="头像占位文字">
                    <Input v-model="form.avatar_mark" maxlength="8" :disabled="!canEdit" />
                  </AdminFormField>
                  <AdminFormField label="头像替代文字" class="md:col-span-2">
                    <Input v-model="form.avatar_label" :disabled="!canEdit" />
                  </AdminFormField>
                </div>

                <div class="flex min-h-28 items-center justify-center rounded-2xl border border-dashed border-border/80 bg-muted/25 p-3">
                  <div
                    v-if="form.avatar_url"
                    class="size-20 overflow-hidden rounded-full border border-primary/50 bg-muted shadow-lg"
                  >
                    <img :src="form.avatar_url" :alt="form.avatar_label" class="size-full object-cover" />
                  </div>
                  <div v-else class="flex size-20 items-center justify-center rounded-full border border-primary/50 bg-muted text-lg font-black text-primary">
                    {{ form.avatar_mark || 'ME' }}
                  </div>
                </div>
              </div>

              <div class="grid gap-3 md:grid-cols-3">
                <AdminFormField label="身份标签">
                  <Input v-model="form.profile_label" :disabled="!canEdit" />
                </AdminFormField>
                <AdminFormField label="负责内容">
                  <Input v-model="form.profile_role" :disabled="!canEdit" />
                </AdminFormField>
                <AdminFormField label="与工厂的关系">
                  <Input v-model="form.profile_context" :disabled="!canEdit" />
                </AdminFormField>
              </div>
            </div>
          </section>
        </TabsContent>

        <TabsContent value="statement">
          <section class="grid w-full max-w-none gap-4 lg:grid-cols-[160px_minmax(0,1fr)]">
            <div>
              <h2 class="text-sm font-black tracking-tighter uppercase text-foreground">页面说明</h2>
              <p class="mt-1 text-[9px] font-black uppercase tracking-widest text-muted-foreground/60">
                解释页面为什么存在
              </p>
            </div>
            <div class="grid min-w-0 gap-3 md:grid-cols-2">
              <AdminFormField label="说明眉题">
                <Input v-model="form.statement_eyebrow" :disabled="!canEdit" />
              </AdminFormField>
              <AdminFormField label="说明标题">
                <Input v-model="form.statement_title" :disabled="!canEdit" />
              </AdminFormField>
              <AdminFormField label="第一段正文" class="md:col-span-2">
                <Textarea v-model="form.statement_paragraph_1" class="min-h-24" :disabled="!canEdit" />
              </AdminFormField>
              <AdminFormField label="第二段正文" class="md:col-span-2">
                <Textarea v-model="form.statement_paragraph_2" class="min-h-24" :disabled="!canEdit" />
              </AdminFormField>
            </div>
          </section>
        </TabsContent>

        <TabsContent value="factory">
          <section class="grid w-full max-w-none gap-4 lg:grid-cols-[160px_minmax(0,1fr)]">
            <div>
              <h2 class="text-sm font-black tracking-tighter uppercase text-foreground">工厂关联</h2>
              <p class="mt-1 text-[9px] font-black uppercase tracking-widest text-muted-foreground/60">
                保持网站与工厂工作的同一口径
              </p>
            </div>
            <div class="min-w-0 space-y-4">
              <div class="grid gap-3 lg:grid-cols-[minmax(0,1fr)_15rem] lg:items-start">
                <div class="grid gap-3 md:grid-cols-2">
                  <AdminFormField label="工厂图片 URL（全局）" class="md:col-span-2">
                    <div class="flex min-w-0 items-center gap-2">
                      <Input v-model="form.factory_image_url" :disabled="!canEdit || uploading === 'factory'" />
                      <Button
                        type="button"
                        variant="outline"
                        size="icon"
                        :disabled="!canEdit || uploading === 'factory'"
                        title="上传工厂图片"
                        @click="chooseUpload('factory')"
                      >
                        <LoaderCircle v-if="uploading === 'factory'" class="size-4 animate-spin" />
                        <ImagePlus v-else class="size-4" />
                      </Button>
                    </div>
                    <input
                      ref="factoryInput"
                      type="file"
                      class="hidden"
                      :accept="uploadSpecAccept('website_profile_image')"
                      :disabled="!canEdit || uploading === 'factory'"
                      @change="handleUpload('factory', $event)"
                    />
                    <UploadSpecHint code="website_profile_image" />
                  </AdminFormField>
                  <AdminFormField label="图片替代文字">
                    <Input v-model="form.factory_image_alt" :disabled="!canEdit" />
                  </AdminFormField>
                  <AdminFormField label="图片说明">
                    <Input v-model="form.factory_image_caption" :disabled="!canEdit" />
                  </AdminFormField>
                </div>

                <div class="overflow-hidden rounded-2xl border border-dashed border-border/80 bg-muted/25">
                  <div v-if="form.factory_image_url" class="aspect-[16/9] bg-muted">
                    <img :src="form.factory_image_url" :alt="form.factory_image_alt" class="size-full object-cover" />
                  </div>
                  <div v-else class="flex aspect-[16/9] items-center justify-center text-xs font-bold text-muted-foreground">
                    暂无图片预览
                  </div>
                </div>
              </div>

              <div class="grid gap-3 md:grid-cols-2">
                <AdminFormField label="工厂区块眉题">
                  <Input v-model="form.factory_eyebrow" :disabled="!canEdit" />
                </AdminFormField>
                <AdminFormField label="工厂区块标题">
                  <Input v-model="form.factory_title" :disabled="!canEdit" />
                </AdminFormField>
                <AdminFormField label="工厂区块正文" class="md:col-span-2">
                  <Textarea v-model="form.factory_body" class="min-h-20" :disabled="!canEdit" />
                </AdminFormField>
                <AdminFormField label="按钮文案">
                  <Input v-model="form.factory_cta" :disabled="!canEdit" />
                </AdminFormField>
                <AdminFormField
                  label="工厂入口链接（全局）"
                  description="可以填写站内相对路径，例如 /company/about/factory；修改后所有语言共用。"
                >
                  <Input v-model="form.factory_link" :disabled="!canEdit" />
                </AdminFormField>
              </div>
            </div>
          </section>
        </TabsContent>
      </Tabs>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { toast } from 'vue-sonner'
import { ImagePlus, LoaderCircle, Save } from '@lucide/vue'
import AdminFormField from '@/components/admin/AdminFormField.vue'
import AdminPageHeader from '@/components/admin/AdminPageHeader.vue'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Textarea } from '@/components/ui/textarea'
import mediaApi from '@/api/media'
import UploadSpecHint from '@/components/admin/UploadSpecHint.vue'
import { assetAccessURL } from '@/lib/mediaPresentation'
import { buildLanguageOptions, STOREFRONT_SUPPORTED_LANGUAGES } from '@/lib/languages'
import { uploadSpecAccept, validateUploadFile } from '@/lib/uploadSpecs'
import { useAuthStore } from '@/stores/auth'
import axios from '@/utils/axios'

type UploadTarget = 'avatar' | 'factory'

interface WebsiteProfileForm {
  locale: string
  eyebrow: string
  title: string
  lead: string
  scope: string
  avatar_url: string
  avatar_label: string
  avatar_mark: string
  profile_label: string
  profile_role: string
  profile_context: string
  statement_eyebrow: string
  statement_title: string
  statement_paragraph_1: string
  statement_paragraph_2: string
  factory_image_url: string
  factory_image_alt: string
  factory_image_caption: string
  factory_eyebrow: string
  factory_title: string
  factory_body: string
  factory_cta: string
  factory_link: string
}

const languageOptions = buildLanguageOptions(STOREFRONT_SUPPORTED_LANGUAGES)
const authStore = useAuthStore()
const canEdit = computed(() => authStore.hasPermission('settings:edit'))
const contentLocale = ref('en')
const activeSection = ref('identity')
const loading = ref(false)
const saving = ref(false)
const uploading = ref<UploadTarget | ''>('')
const avatarInput = ref<HTMLInputElement | null>(null)
const factoryInput = ref<HTMLInputElement | null>(null)
let requestSequence = 0

const form = reactive<WebsiteProfileForm>({
  locale: 'en',
  eyebrow: '',
  title: '',
  lead: '',
  scope: '',
  avatar_url: '',
  avatar_label: '',
  avatar_mark: '',
  profile_label: '',
  profile_role: '',
  profile_context: '',
  statement_eyebrow: '',
  statement_title: '',
  statement_paragraph_1: '',
  statement_paragraph_2: '',
  factory_image_url: '',
  factory_image_alt: '',
  factory_image_caption: '',
  factory_eyebrow: '',
  factory_title: '',
  factory_body: '',
  factory_cta: '',
  factory_link: '',
})

const selectedLanguageLabel = computed(() => (
  languageOptions.find((option) => option.value === contentLocale.value)?.label || contentLocale.value
))

const assignProfile = (data: Partial<WebsiteProfileForm> = {}): void => {
  Object.keys(form).forEach((key) => {
    const typedKey = key as keyof WebsiteProfileForm
    form[typedKey] = String(data[typedKey] ?? '')
  })
  form.locale = contentLocale.value
}

const loadProfile = async (): Promise<void> => {
  const sequence = ++requestSequence
  loading.value = true
  try {
    const response = await axios.get('/api/admin/settings/website-profile', {
      params: { locale: contentLocale.value },
    })
    if (sequence !== requestSequence) return
    assignProfile(response.data?.data || {})
  } catch (error) {
    if (sequence !== requestSequence) return
    console.error('Failed to load website profile settings:', error)
    toast.error('我与这个网站设置加载失败')
  } finally {
    if (sequence === requestSequence) loading.value = false
  }
}

const chooseUpload = (target: UploadTarget): void => {
  if (!canEdit.value || uploading.value) return
  const input = target === 'avatar' ? avatarInput.value : factoryInput.value
  input?.click()
}

const handleUpload = async (target: UploadTarget, event: Event): Promise<void> => {
  const input = event.target instanceof HTMLInputElement ? event.target : null
  const file = input?.files?.[0] || null
  if (input) input.value = ''
  if (!file) return
  const purpose = target === 'avatar' ? 'website_profile_avatar' : 'website_profile_image'
  const validation = await validateUploadFile(file, purpose)
  if (!validation.ok) {
    toast.error(validation.error || '图片不符合上传规范')
    return
  }
  if (validation.warning) toast.warning(validation.warning)

  uploading.value = target
  try {
    const formData = new FormData()
    formData.append('file', file)
    formData.append('media_type', 'image')
    formData.append('image_purpose', purpose)
    const asset = await mediaApi.uploadAsset(formData)
    const url = String(assetAccessURL(asset) || '').trim()
    if (!url) {
      toast.error('上传成功但没有返回图片地址')
      return
    }
    if (target === 'avatar') form.avatar_url = url
    else form.factory_image_url = url
    toast.success('图片已上传，保存设置后前台生效')
  } catch (error) {
    console.error('Failed to upload website profile image:', error)
    toast.error('图片上传失败，请检查文件类型和大小')
  } finally {
    uploading.value = ''
  }
}

const saveProfile = async (): Promise<void> => {
  if (!canEdit.value || saving.value) return
  saving.value = true
  try {
    const response = await axios.put('/api/admin/settings/website-profile', {
      ...form,
      locale: contentLocale.value,
    })
    assignProfile(response.data?.data || form)
    toast.success('我与这个网站设置已保存')
  } catch (error) {
    console.error('Failed to save website profile settings:', error)
    toast.error(error?.response?.data?.error || '设置保存失败')
  } finally {
    saving.value = false
  }
}

watch(contentLocale, () => {
  void loadProfile()
})

onMounted(() => {
  void loadProfile()
})
</script>
