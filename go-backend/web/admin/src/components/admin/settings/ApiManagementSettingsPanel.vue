<template>
  <div class="grid items-start gap-4 xl:grid-cols-2">
    <TimeApiSettingsCard
      :api-settings="apiSettings"
      :can-edit="canEdit"
      :saving="savingApiSettings"
    />

    <CustomsLookupApiSettingsCard
      :api-settings="apiSettings"
      :can-edit="canEdit"
      :saving="savingApiSettings"
    />

    <section class="rounded-2xl border bg-muted/30 p-4 xl:col-span-2">
      <div class="grid gap-4 md:grid-cols-3">
        <div class="flex items-start gap-3">
          <ShieldCheck class="mt-0.5 size-4 flex-none text-admin-selected" />
          <div>
            <p class="text-xs font-black text-foreground">后端代理</p>
            <p class="mt-1 text-xs leading-relaxed text-muted-foreground">第三方接口由后端请求并缓存，前台不直接拿接口地址和密钥。</p>
          </div>
        </div>
        <div class="flex items-start gap-3">
          <KeyRound class="mt-0.5 size-4 flex-none text-admin-selected" />
          <div>
            <p class="text-xs font-black text-foreground">私有 Key</p>
            <p class="mt-1 text-xs leading-relaxed text-muted-foreground">第三方接口凭据只保存在后台私有设置，不写进 Nuxt 前台包。</p>
          </div>
        </div>
        <div class="flex items-start gap-3">
          <RefreshCw class="mt-0.5 size-4 flex-none text-admin-selected" />
          <div>
            <p class="text-xs font-black text-foreground">缓存与内置规则</p>
            <p class="mt-1 text-xs leading-relaxed text-muted-foreground">汇率接口继续走缓存；时区使用内置规则，不再配置外部时间 API。</p>
          </div>
        </div>
      </div>
    </section>
  </div>
</template>

<script setup lang="ts">
import { KeyRound, RefreshCw, ShieldCheck } from '@lucide/vue'
import CustomsLookupApiSettingsCard from '@/components/admin/settings/CustomsLookupApiSettingsCard.vue'
import TimeApiSettingsCard from '@/components/admin/settings/TimeApiSettingsCard.vue'
import type { ApiManagementSettings } from '@/modules/settings/types'

withDefaults(defineProps<{
  apiSettings: ApiManagementSettings
  canEdit?: boolean
  savingApiSettings?: boolean
}>(), {
  canEdit: false,
  savingApiSettings: false,
})
</script>
