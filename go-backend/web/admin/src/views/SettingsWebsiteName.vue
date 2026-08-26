<template>
  <div class="space-y-3">
    <AdminPageHeader
      title="为什么叫这个名字"
      description="编辑 Nuxt 页面「为什么叫这个名字」的公开文案。每种语言独立保存，未配置的语言会回退到 English。"
    >
      <template #actions>
        <Button v-if="loadError" variant="outline" :disabled="loading || saving" @click="loadSettings">
          <RefreshCw class="size-4" />
          重新加载
        </Button>
        <Button :disabled="loading || saving || !canEdit || loadError" @click="saveSettings">
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
            description="选择要编辑的前台语言。系统当前支持 20 种语言。"
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
              只影响当前选中的语言，不会覆盖其他语言内容。
            </p>
          </div>
        </div>
      </section>

      <section class="grid w-full max-w-none gap-4 lg:grid-cols-[160px_minmax(0,1fr)]">
        <div>
          <h2 class="text-sm font-black tracking-tighter uppercase text-foreground">页面文案</h2>
          <p class="mt-1 text-[9px] font-black uppercase tracking-widest text-muted-foreground/60">
            控制页面首屏与正文内容
          </p>
        </div>

        <div class="grid min-w-0 gap-3 md:grid-cols-2">
          <AdminFormField label="状态文字">
            <Input v-model="form.status" :disabled="!canEdit || loadError" />
          </AdminFormField>
          <AdminFormField label="眉题">
            <Input v-model="form.eyebrow" :disabled="!canEdit || loadError" />
          </AdminFormField>
          <AdminFormField label="页面标题">
            <Input v-model="form.title" :disabled="!canEdit || loadError" />
          </AdminFormField>
          <AdminFormField label="备注文字">
            <Input v-model="form.note" :disabled="!canEdit || loadError" />
          </AdminFormField>
          <AdminFormField label="页面引导语" class="md:col-span-2">
            <Textarea v-model="form.intro" class="min-h-24" :disabled="!canEdit || loadError" />
          </AdminFormField>
          <AdminFormField label="正文内容" class="md:col-span-2">
            <Textarea v-model="form.body" class="min-h-44" :disabled="!canEdit || loadError" />
          </AdminFormField>
        </div>
      </section>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { toast } from 'vue-sonner'
import { LoaderCircle, RefreshCw, Save } from '@lucide/vue'
import AdminFormField from '@/components/admin/AdminFormField.vue'
import AdminPageHeader from '@/components/admin/AdminPageHeader.vue'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Textarea } from '@/components/ui/textarea'
import { buildLanguageOptions, STOREFRONT_SUPPORTED_LANGUAGES } from '@/lib/languages'
import { useAuthStore } from '@/stores/auth'
import axios from '@/utils/axios'

interface WebsiteNameForm {
  locale: string
  status: string
  intro: string
  eyebrow: string
  title: string
  body: string
  note: string
}

const languageOptions = buildLanguageOptions(STOREFRONT_SUPPORTED_LANGUAGES)
const authStore = useAuthStore()
const canEdit = computed(() => authStore.hasPermission('settings:edit'))
const contentLocale = ref('en')
const loading = ref(false)
const loadError = ref(false)
const saving = ref(false)
let requestSequence = 0

const form = reactive<WebsiteNameForm>({
  locale: 'en',
  status: '',
  intro: '',
  eyebrow: '',
  title: '',
  body: '',
  note: '',
})

const selectedLanguageLabel = computed(() => (
  languageOptions.find((option) => option.value === contentLocale.value)?.label || contentLocale.value
))

const assignSettings = (data: Partial<WebsiteNameForm> = {}): void => {
  Object.keys(form).forEach((key) => {
    const typedKey = key as keyof WebsiteNameForm
    form[typedKey] = String(data[typedKey] ?? '')
  })
  form.locale = contentLocale.value
}

const loadSettings = async (): Promise<void> => {
  const sequence = ++requestSequence
  loading.value = true
  loadError.value = false
  try {
    const response = await axios.get('/api/admin/settings/website-name', {
      params: { locale: contentLocale.value },
    })
    if (sequence !== requestSequence) return
    assignSettings(response.data?.data || {})
  } catch (error) {
    if (sequence !== requestSequence) return
    console.error('Failed to load why this name settings:', error)
    assignSettings({})
    loadError.value = true
    toast.error('为什么叫这个名字设置加载失败')
  } finally {
    if (sequence === requestSequence) loading.value = false
  }
}

const saveSettings = async (): Promise<void> => {
  if (!canEdit.value || saving.value || loadError.value) return
  saving.value = true
  try {
    const response = await axios.put('/api/admin/settings/website-name', {
      ...form,
      locale: contentLocale.value,
    })
    assignSettings(response.data?.data || form)
    loadError.value = false
    toast.success('为什么叫这个名字设置已保存')
  } catch (error) {
    console.error('Failed to save why this name settings:', error)
    toast.error(error?.response?.data?.error || '设置保存失败')
  } finally {
    saving.value = false
  }
}

watch(contentLocale, () => {
  void loadSettings()
})

onMounted(() => {
  void loadSettings()
})
</script>
