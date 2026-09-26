<script setup lang="ts">
import { computed, ref } from 'vue'
import { Activity, CheckCircle2, RefreshCw, Settings2 } from '@lucide/vue'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { useAuthStore } from '@/stores/auth'

const authStore = useAuthStore()
const canManage = computed(() => authStore.hasPermission('logistics:fpx:manage'))

const configForm = ref({
  environment: 'sandbox',
  appKey: '',
  appSecret: '',
})
const configSecretVisible = ref(false)
const configSaved = ref(false)
const configTesting = ref(false)
const configAttempted = ref(false)
const configTested = ref(false)
const configCanSave = computed(() =>
  canManage.value
  && configForm.value.appKey.trim().length > 0
  && configForm.value.appSecret.trim().length > 0,
)
const configEndpoint = computed(() =>
  configForm.value.environment === 'production'
    ? 'https://open.4px.com/router/api/service'
    : 'https://open-test.4px.com/router/api/service',
)
const saveConfigDraft = () => {
  if (!configCanSave.value) return
  configSaved.value = true
  configAttempted.value = false
  configTested.value = false
}
const testConfigConnection = async () => {
  if (!configCanSave.value) return
  configTesting.value = true
  configAttempted.value = false
  configTested.value = false
  await Promise.resolve()
  configTesting.value = false
  configAttempted.value = true
}
</script>

<template>
      <form class="space-y-5 rounded-[28px] border border-dashed border-border/80 bg-card p-5 shadow-sm sm:p-6" @submit.prevent>
        <div class="flex flex-col gap-4 lg:flex-row lg:items-start lg:justify-between">
          <div>
            <p class="text-[10px] font-black uppercase tracking-[0.18em] text-slate-600 dark:text-slate-300">API Credentials & Master Config</p>
            <h2 class="mt-1 text-lg font-black tracking-tight">接口凭据与主数据同步</h2>
            <p class="mt-1 max-w-3xl text-xs leading-5 text-muted-foreground">管理 4PX 开放平台环境与应用凭据，执行签名连通性自检。凭据保存操作仅允许 logistics:fpx:manage 权限。</p>
          </div>
          <span class="inline-flex w-fit items-center gap-2 rounded-full border border-border bg-muted px-3 py-1.5 text-[11px] font-bold text-muted-foreground">
            <Settings2 class="size-3.5" />
            {{ configForm.environment === 'production' ? '生产环境' : '沙箱环境' }}
          </span>
        </div>

        <div class="grid gap-4 xl:grid-cols-[minmax(0,1.1fr)_minmax(320px,0.9fr)]">
          <div class="space-y-4 rounded-[24px] border border-dashed border-border/80 bg-muted/20 p-4">
            <div class="flex items-center justify-between gap-3">
              <div><h3 class="text-sm font-black">应用凭据配置</h3><p class="mt-1 text-[11px] text-muted-foreground">App Secret 不会在台账中明文回显。</p></div>
              <span :class="configSaved ? 'bg-emerald-500/10 text-emerald-700 dark:text-emerald-300' : 'bg-amber-500/10 text-amber-700 dark:text-amber-300'" class="rounded-full px-2.5 py-1 text-[10px] font-bold">{{ configSaved ? '会话草稿已记录' : '尚未保存' }}</span>
            </div>
            <label class="space-y-1.5">
              <span class="text-xs font-bold">运行环境</span>
              <select v-model="configForm.environment" :disabled="!canManage" class="h-9 w-full rounded-md border border-border bg-background px-3 text-sm" @change="configTested = false">
                <option value="sandbox">测试环境 · open-test.4px.com</option>
                <option value="production">生产环境 · open.4px.com</option>
              </select>
            </label>
            <label class="space-y-1.5"><span class="text-xs font-bold">App Key</span><Input v-model="configForm.appKey" :disabled="!canManage" autocomplete="off" placeholder="输入 4PX app_key" @input="configSaved = false; configTested = false" /></label>
            <label class="space-y-1.5">
              <span class="text-xs font-bold">App Secret</span>
              <div class="flex gap-2">
                <Input v-model="configForm.appSecret" :type="configSecretVisible ? 'text' : 'password'" :disabled="!canManage" autocomplete="new-password" placeholder="输入 4PX app_secret" @input="configSaved = false; configTested = false" />
                <Button type="button" variant="outline" :disabled="!canManage" @click="configSecretVisible = !configSecretVisible">{{ configSecretVisible ? '隐藏' : '显示' }}</Button>
              </div>
            </label>
            <div class="rounded-2xl border border-dashed border-border/80 bg-background px-4 py-3"><p class="text-[10px] font-black uppercase tracking-wider text-muted-foreground">当前网关</p><p class="mt-1 break-all font-mono text-xs font-bold">{{ configEndpoint }}</p></div>
            <div class="flex flex-wrap gap-2">
              <Button :disabled="!configCanSave" @click="saveConfigDraft">保存凭据草稿</Button>
              <Button variant="outline" :disabled="!configCanSave || configTesting" @click="testConfigConnection"><RefreshCw :class="['size-3.5', { 'animate-spin': configTesting }]" />连通性自检</Button>
            </div>
          </div>

          <div class="space-y-3">
            <article class="rounded-[24px] border border-dashed border-border/80 bg-card p-4">
              <p class="text-[10px] font-black uppercase tracking-wider text-muted-foreground">签名与网络状态</p>
              <div class="mt-3 flex items-center gap-3">
                <span :class="configTested ? 'bg-emerald-500/10 text-emerald-600' : 'bg-muted text-muted-foreground'" class="flex size-10 items-center justify-center rounded-2xl"><CheckCircle2 v-if="configTested" class="size-4" /><Activity v-else class="size-4" /></span>
                <div><p class="text-sm font-black">{{ configAttempted ? '自检已执行，等待真实响应' : '尚未执行自检' }}</p><p class="mt-1 text-[11px] leading-5 text-muted-foreground">{{ configAttempted ? '真实 ds.xms.logistics_product.getlist 后端接口尚未接入，当前不伪造成功响应。' : '保存凭据后调用测试连接，验证 MD5 签名、DNS、TLS 与 4PX 网关响应。' }}</p></div>
              </div>
            </article>
            <article class="rounded-[24px] border border-dashed border-border/80 bg-card p-4">
              <p class="text-[10px] font-black uppercase tracking-wider text-muted-foreground">安全边界</p>
              <ul class="mt-3 space-y-2 text-xs leading-5 text-muted-foreground">
                <li class="flex gap-2"><CheckCircle2 class="mt-0.5 size-3.5 shrink-0 text-emerald-600" />App Secret 仅提交给后端加密存储，不进入列表响应。</li>
                <li class="flex gap-2"><CheckCircle2 class="mt-0.5 size-3.5 shrink-0 text-emerald-600" />生产环境切换要求管理权限并重新执行连通性自检。</li>
                <li class="flex gap-2"><CheckCircle2 class="mt-0.5 size-3.5 shrink-0 text-emerald-600" />不在浏览器日志、URL 或本地台账中暴露明文密钥。</li>
              </ul>
            </article>
          </div>
        </div>
      </form>

</template>
