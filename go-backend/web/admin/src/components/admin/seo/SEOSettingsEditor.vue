<template>
  <div class="space-y-4">
    <AdminPageHeader :title="title" :description="description">
      <template #actions>
        <Button variant="outline" as-child>
          <a :href="publicPageHref" target="_blank" rel="noreferrer">
            <ExternalLink class="size-4" />
            打开页面
          </a>
        </Button>
        <Button variant="outline" :disabled="loading || saving" @click="load">
          <RefreshCw :class="['size-4', loading ? 'animate-spin' : '']" />
          刷新
        </Button>
        <Button v-if="canEdit" :disabled="loading || saving" @click="save">
          <LoaderCircle v-if="saving" class="size-4 animate-spin" />
          <Save v-else class="size-4" />
          {{ saving ? '保存中' : '保存 SEO' }}
        </Button>
      </template>
    </AdminPageHeader>

    <section class="rounded-2xl border bg-muted/20 p-4">
      <div class="mb-4 flex items-start justify-between gap-3">
        <div>
          <p class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/60">SEO Defaults</p>
          <h2 class="mt-1 text-sm font-black">{{ sectionTitle }}</h2>
        </div>
        <StorefrontLocaleSelect
          v-model="selectedLocale"
          :language-options="languageOptions"
          :loading="languagesLoading"
          :disabled="loading || saving"
          class="w-44"
        />
      </div>

      <div class="grid gap-4">
        <AdminFormField label="页面路由" description="首页路由固定为根路径，不能修改。">
          <div class="relative">
            <LockKeyhole class="pointer-events-none absolute left-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" />
            <Input :model-value="pageRoute" class="pl-9 font-mono" disabled />
          </div>
        </AdminFormField>
        <AdminFormField label="Meta 标题" description="页面没有单独设置标题时使用。">
          <Input v-model="form.meta_title" maxlength="160" :disabled="!canEdit || loading" />
        </AdminFormField>
        <AdminFormField label="Meta 描述" description="页面没有单独设置描述时使用。">
          <Textarea v-model="form.meta_description" maxlength="320" class="min-h-28" :disabled="!canEdit || loading" />
        </AdminFormField>
      </div>
    </section>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { toast } from 'vue-sonner'
import { ExternalLink, LoaderCircle, LockKeyhole, RefreshCw, Save } from '@lucide/vue'
import AdminFormField from '@/components/admin/AdminFormField.vue'
import AdminPageHeader from '@/components/admin/AdminPageHeader.vue'
import StorefrontLocaleSelect from '@/components/admin/StorefrontLocaleSelect.vue'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import { useSupportedLanguages } from '@/composables/useSupportedLanguages'
import { storefrontHref } from '@/modules/seo/routes'
import { useAuthStore } from '@/stores/auth'
import type { SEOHomeApi, SEOSettings } from '@/modules/seo/types'

const props = defineProps<{
  title: string
  description: string
  sectionTitle: string
  pageRoute?: string
  api: SEOHomeApi
}>()

const authStore = useAuthStore()
const canEdit = authStore.hasPermission('seo:edit')
const pageRoute = props.pageRoute || '/'
const publicPageHref = computed(() => storefrontHref(pageRoute))
const loading = ref(false)
const saving = ref(false)
const selectedLocale = ref('en')
const { languageOptions, loading: languagesLoading, fetchLanguages } = useSupportedLanguages()
const form = reactive<SEOSettings>({
  meta_title: '',
  meta_description: '',
})

const applySettings = (settings: SEOSettings) => {
  Object.assign(form, {
    meta_title: settings.meta_title || '',
    meta_description: settings.meta_description || '',
  })
}

const load = async () => {
  loading.value = true
  try {
    applySettings(await props.api.get(selectedLocale.value))
  } catch (error) {
    console.error('Failed to load SEO settings:', error)
    toast.error('SEO 设置加载失败')
  } finally {
    loading.value = false
  }
}

const save = async () => {
  if (!canEdit) return
  saving.value = true
  try {
    applySettings(await props.api.update({ ...form, locale: selectedLocale.value }))
    toast.success('SEO 设置已保存')
  } catch (error) {
    console.error('Failed to save SEO settings:', error)
    toast.error('SEO 设置保存失败')
  } finally {
    saving.value = false
  }
}

watch(selectedLocale, load)

onMounted(async () => {
  await fetchLanguages()
  await load()
})
</script>
