<template>
  <div class="space-y-4">
    <AdminPageHeader title="Analytics" description="前台访问统计与标签容器">
      <template #actions>
        <Button variant="outline" :disabled="loading || saving" @click="load">
          <RefreshCw :class="['size-4', loading ? 'animate-spin' : '']" />
          刷新
        </Button>
        <Button v-if="canEdit" :disabled="loading || saving" @click="save">
          <LoaderCircle v-if="saving" class="size-4 animate-spin" />
          <Save v-else class="size-4" />
          {{ saving ? '保存中' : '保存 Analytics' }}
        </Button>
      </template>
    </AdminPageHeader>

    <section class="rounded-2xl border bg-muted/20 p-4">
      <div class="mb-4">
        <p class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/60">Tracking</p>
        <h2 class="mt-1 text-sm font-black">第三方统计标识</h2>
      </div>

      <div class="grid gap-4 md:grid-cols-2">
        <AdminFormField label="Google Analytics" description="例如 G-XXXXXXXXXX 或 UA-XXXXXXXX-X。">
          <Input v-model="form.google_analytics" maxlength="128" placeholder="GA 测量 ID" :disabled="!canEdit || loading" />
        </AdminFormField>
        <AdminFormField label="Google Tag Manager" description="例如 GTM-XXXXXXX。">
          <Input v-model="form.google_tag_manager" maxlength="128" placeholder="GTM 容器 ID" :disabled="!canEdit || loading" />
        </AdminFormField>
      </div>
    </section>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { toast } from 'vue-sonner'
import { LoaderCircle, RefreshCw, Save } from '@lucide/vue'
import AdminFormField from '@/components/admin/AdminFormField.vue'
import AdminPageHeader from '@/components/admin/AdminPageHeader.vue'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { useAuthStore } from '@/stores/auth'
import { analyticsApi } from '@/modules/analytics/api'
import type { AnalyticsSettings } from '@/modules/analytics/types'

const authStore = useAuthStore()
const canEdit = authStore.hasPermission('analytics:edit')
const loading = ref(false)
const saving = ref(false)
const form = reactive<AnalyticsSettings>({
  google_analytics: '',
  google_tag_manager: '',
})

const applySettings = (settings: AnalyticsSettings) => {
  Object.assign(form, {
    google_analytics: settings.google_analytics || '',
    google_tag_manager: settings.google_tag_manager || '',
  })
}

const load = async () => {
  loading.value = true
  try {
    applySettings(await analyticsApi.get())
  } catch (error) {
    console.error('Failed to load analytics settings:', error)
    toast.error('Analytics 设置加载失败')
  } finally {
    loading.value = false
  }
}

const save = async () => {
  if (!canEdit) return
  saving.value = true
  try {
    applySettings(await analyticsApi.update({ ...form, locale: 'en' }))
    toast.success('Analytics 设置已保存')
  } catch (error) {
    console.error('Failed to save analytics settings:', error)
    toast.error('Analytics 设置保存失败')
  } finally {
    saving.value = false
  }
}

onMounted(load)
</script>
