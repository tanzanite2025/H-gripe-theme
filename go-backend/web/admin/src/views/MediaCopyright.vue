<template>
  <div class="space-y-4">
    <AdminPageHeader
      :title="t('settings.imageCopyrightTitle')"
      :description="t('settings.imageCopyrightDescription')"
    >
      <template #actions>
        <Button
          v-if="canEdit"
          :disabled="loading || saving"
          @click="saveSettings"
        >
          <LoaderCircle v-if="saving" class="size-4 animate-spin" />
          <Save v-else class="size-4" />
          {{ saving ? t('common.saving') : t('settings.saveSettings') }}
        </Button>
      </template>
    </AdminPageHeader>

    <div class="relative min-h-64">
      <div
        v-if="loading"
        class="absolute inset-0 z-10 flex items-center justify-center bg-background/75"
      >
        <LoaderCircle class="size-5 animate-spin text-primary" :aria-label="t('settings.loadingSettings')" />
      </div>

      <section class="w-full max-w-none space-y-3">
        <div class="max-w-3xl">
          <h2 class="text-sm font-black tracking-tighter uppercase text-foreground">
            {{ t('settings.copyrightEvidence') }}
          </h2>
          <p class="mt-1 text-[9px] font-black uppercase tracking-widest text-muted-foreground/60">
            {{ t('settings.copyrightEvidenceDescription') }}
          </p>
        </div>

        <div class="grid gap-4 md:grid-cols-2">
          <AdminFormField :label="t('settings.rightHolder')">
            <Input
              v-model="copyrightSettings.copyright_holder"
              :disabled="!canEdit || saving"
              placeholder="公司或版权主体名称"
            />
          </AdminFormField>
          <AdminFormField :label="t('settings.copyrightPolicyUrl')">
            <Input
              v-model="copyrightSettings.copyright_url"
              :disabled="!canEdit || saving"
              type="url"
              placeholder="https://example.com/policies/copyright"
            />
          </AdminFormField>
          <AdminFormField :label="t('settings.copyrightNotice')" class="md:col-span-2">
            <Textarea
              v-model="copyrightSettings.copyright_notice"
              :disabled="!canEdit || saving"
              class="min-h-24"
              placeholder="Copyright 2026 Example. All rights reserved."
            />
          </AdminFormField>
        </div>
      </section>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { toast } from 'vue-sonner'
import { LoaderCircle, Save } from '@lucide/vue'
import AdminFormField from '@/components/admin/AdminFormField.vue'
import AdminPageHeader from '@/components/admin/AdminPageHeader.vue'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import { useAdminI18n } from '@/i18n'
import { useAuthStore } from '@/stores/auth'
import axios from '@/utils/axios'

interface CopyrightSettings {
  copyright_holder: string
  copyright_notice: string
  copyright_url: string
}

interface SettingRecord {
  key?: string
  value?: unknown
}

const { t } = useAdminI18n()
const authStore = useAuthStore()
const canEdit = authStore.hasPermission('settings:edit')
const loading = ref(false)
const saving = ref(false)
const copyrightSettings = reactive<CopyrightSettings>({
  copyright_holder: '',
  copyright_notice: '',
  copyright_url: '',
})

const assignSettings = (settings: Record<string, unknown> = {}): void => {
  copyrightSettings.copyright_holder = String(settings.copyright_holder ?? '')
  copyrightSettings.copyright_notice = String(settings.copyright_notice ?? '')
  copyrightSettings.copyright_url = String(settings.copyright_url ?? '')
}

const loadSettings = async (): Promise<void> => {
  loading.value = true
  try {
    const response = await axios.get('/api/admin/settings/site', {
      params: { locale: 'en' },
    })
    const settings = (Array.isArray(response.data?.settings) ? response.data.settings : []) as SettingRecord[]
    const values = settings.reduce((result: Record<string, string>, setting: SettingRecord) => {
      if (['copyright_holder', 'copyright_notice', 'copyright_url'].includes(setting.key)) {
        result[setting.key as string] = String(setting.value ?? '')
      }
      return result
    }, {})
    assignSettings(values)
  } catch (error) {
    console.error('Failed to load image copyright settings:', error)
    toast.error('图片版权设置加载失败')
  } finally {
    loading.value = false
  }
}

const saveSettings = async (): Promise<void> => {
  if (!canEdit || saving.value) return

  saving.value = true
  try {
    await axios.post('/api/admin/settings/batch', {
      settings: [
        {
          key: 'copyright_holder',
          value: copyrightSettings.copyright_holder.trim(),
          type: 'string',
          group: 'site',
          locale: 'en',
          is_public: false,
          description: 'Copyright holder for image evidence',
        },
        {
          key: 'copyright_notice',
          value: copyrightSettings.copyright_notice.trim(),
          type: 'string',
          group: 'site',
          locale: 'en',
          is_public: false,
          description: 'Copyright notice for image evidence',
        },
        {
          key: 'copyright_url',
          value: copyrightSettings.copyright_url.trim(),
          type: 'string',
          group: 'site',
          locale: 'en',
          is_public: false,
          description: 'Copyright policy URL for image evidence',
        },
      ],
    })
    await loadSettings()
    toast.success('图片版权设置已保存')
  } catch (error) {
    console.error('Failed to save image copyright settings:', error)
    toast.error(error?.response?.data?.error || '图片版权设置保存失败')
  } finally {
    saving.value = false
  }
}

onMounted(() => {
  void loadSettings()
})
</script>
