<script setup lang="ts">
import { computed, ref } from 'vue'
import { CheckCircle2, RefreshCw, Settings2 } from '@lucide/vue'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { useAuthStore } from '@/stores/auth'

const authStore = useAuthStore()
const canManage = computed(() => authStore.hasPermission('logistics:yanwen:manage'))

const configForm = ref({ environment: 'fat', userId: '', apiToken: '' })
const configSecretVisible = ref(false)
const configSaved = ref(false)
const configTesting = ref(false)
const configAttempted = ref(false)
const configTested = ref(false)
const configSyncing = ref(false)
const configMessage = ref('')
const configEndpoint = computed(() => configForm.value.environment === 'production' ? 'PRD · 燕文生产网关' : 'FAT · 燕文测试网关')
const configCanSave = computed(() => canManage.value && configForm.value.userId.trim().length > 0 && configForm.value.apiToken.trim().length > 0)
const saveGatewayDraft = () => {
  if (!configCanSave.value) return
  configSaved.value = true
  configAttempted.value = false
  configTested.value = false
  configMessage.value = '网关凭据仅记录在当前会话草稿中，后端安全存储接口尚未接入。'
}
const pingGateway = async () => {
  if (!configCanSave.value) return
  configTesting.value = true
  configAttempted.value = false
  configTested.value = false
  configMessage.value = ''
  await Promise.resolve()
  configTesting.value = false
  configAttempted.value = true
  configMessage.value = 'common.country.getlist 尚未接入，未产生真实成功结论。'
}
const syncYanwenMasterData = async () => {
  if (!canManage.value || !configTested.value) return
  configSyncing.value = true
  configMessage.value = ''
  await Promise.resolve()
  configSyncing.value = false
  configMessage.value = '国家、交货仓和已开通产品同步接口尚未接入，未写入模拟主数据。'
}

</script>

<template>
      <section class="space-y-5 rounded-[28px] border border-dashed border-border/80 bg-card p-5 shadow-sm sm:p-6">
        <div class="flex flex-col gap-4 lg:flex-row lg:items-start lg:justify-between">
          <div>
            <p class="text-[10px] font-black uppercase tracking-[0.18em] text-emerald-600">Channel Master Data &amp; Gateway Config</p>
            <h2 class="mt-1 text-lg font-black tracking-tight">渠道主数据与网关配置</h2>
            <p class="mt-1 max-w-3xl text-xs leading-5 text-muted-foreground">配置燕文客户身份与制单秘钥，执行 common.country.getlist Ping，并同步国家、交货仓和已开通产品。所有管理操作受 logistics:yanwen:manage 控制。</p>
          </div>
          <span class="inline-flex w-fit items-center gap-2 rounded-full border border-border bg-muted px-3 py-1.5 text-[11px] font-bold text-muted-foreground"><Settings2 class="size-3.5" />{{ configEndpoint }}</span>
        </div>

        <div class="grid gap-4 xl:grid-cols-[minmax(0,1.1fr)_minmax(320px,0.9fr)]">
          <section class="space-y-4 rounded-[24px] border border-dashed border-border/80 bg-muted/20 p-4">
            <div class="flex items-center justify-between gap-3"><div><h3 class="text-sm font-black">API 身份凭据</h3><p class="mt-1 text-[11px] text-muted-foreground">apitoken 只允许提交到后端安全存储，不在列表或 URL 中展示。</p></div><span :class="configSaved ? 'bg-emerald-500/10 text-emerald-700 dark:text-emerald-300' : 'bg-muted text-muted-foreground'" class="rounded-full px-2.5 py-1 text-[10px] font-bold">{{ configSaved ? '草稿已保存' : '尚未保存' }}</span></div>
            <label class="space-y-1.5"><span class="text-xs font-bold">运行环境</span><select v-model="configForm.environment" :disabled="!canManage" class="h-9 w-full rounded-md border border-border bg-background px-3 text-sm" @change="configTested = false"><option value="fat">FAT 测试环境</option><option value="production">PRD 生产环境</option></select></label>
            <label class="space-y-1.5"><span class="text-xs font-bold">user_id（客户商户号）</span><Input v-model="configForm.userId" :disabled="!canManage" autocomplete="off" placeholder="输入燕文客户商户号" @input="configSaved = false; configTested = false" /></label>
            <label class="space-y-1.5"><span class="text-xs font-bold">apitoken（制单账号秘钥）</span><div class="flex gap-2"><Input v-model="configForm.apiToken" :type="configSecretVisible ? 'text' : 'password'" :disabled="!canManage" autocomplete="new-password" placeholder="输入燕文 apitoken" @input="configSaved = false; configTested = false" /><Button type="button" variant="outline" :disabled="!canManage" @click="configSecretVisible = !configSecretVisible">{{ configSecretVisible ? '隐藏' : '显示' }}</Button></div></label>
            <div class="flex flex-wrap gap-2"><Button :disabled="!configCanSave" @click="saveGatewayDraft">保存凭据草稿</Button><Button variant="outline" :disabled="!configCanSave || configTesting" @click="pingGateway"><RefreshCw :class="['size-3.5', { 'animate-spin': configTesting }]" />执行网关 Ping</Button></div>
            <p v-if="configMessage" class="rounded-2xl border border-amber-500/30 bg-amber-500/10 px-4 py-3 text-xs leading-5 text-amber-700 dark:text-amber-300">{{ configMessage }}</p>
          </section>

          <section class="space-y-3 rounded-[24px] border border-dashed border-border/80 bg-card p-4"><div class="flex items-center gap-3"><span :class="configAttempted ? 'bg-amber-500/10 text-amber-600' : 'bg-muted text-muted-foreground'" class="flex size-10 items-center justify-center rounded-2xl"><CheckCircle2 class="size-4" /></span><div><h3 class="text-sm font-black">网关状态</h3><p class="mt-1 text-[11px] text-muted-foreground">{{ configAttempted ? 'Ping 流程已执行，等待真实接口响应' : '尚未执行 Ping' }}</p></div></div><div class="grid gap-2 sm:grid-cols-2"><div class="rounded-2xl bg-muted/50 p-3"><p class="text-[10px] font-bold text-muted-foreground">环境</p><p class="mt-1 text-xs font-black">{{ configForm.environment === 'production' ? 'PRD 生产' : 'FAT 测试' }}</p></div><div class="rounded-2xl bg-muted/50 p-3"><p class="text-[10px] font-bold text-muted-foreground">响应延迟</p><p class="mt-1 font-mono text-xs font-black">—</p></div></div><p class="text-xs leading-5 text-muted-foreground">真实 common.country.getlist 返回前，不显示 200 OK、延迟或 Token 剩余天数等推测数据。</p></section>
        </div>
      </section>

      <section class="space-y-4 rounded-[28px] border border-dashed border-border/80 bg-card p-5 shadow-sm sm:p-6">
        <div class="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between"><div><h2 class="text-base font-black">主数据一键同步</h2><p class="mt-1 text-xs text-muted-foreground">同步国家、交货仓和商户已开通产品，供燕文域内部创建运单与运价试算使用。</p></div><Button size="sm" :disabled="!canManage || !configTested || configSyncing" @click="syncYanwenMasterData"><RefreshCw :class="['size-3.5', { 'animate-spin': configSyncing }]" />同步主数据</Button></div>
        <div class="grid gap-3 md:grid-cols-3"><article v-for="item in [{ label: '通达国家', api: 'common.country.getlist', detail: '国家代码与中英文名称' }, { label: '交货仓', api: 'common.warehouse.getlist', detail: '接口返回的燕文揽收仓代码和区域' }, { label: '已开通产品', api: 'express.channel.getlist', detail: '当前商户协议下可用的产品 ID' }]" :key="item.api" class="rounded-[24px] border border-dashed border-border/80 bg-muted/20 p-4"><div class="flex items-center justify-between gap-2"><p class="text-sm font-black">{{ item.label }}</p><span class="rounded-full bg-muted px-2 py-1 text-[10px] font-bold text-muted-foreground">待同步</span></div><p class="mt-2 font-mono text-[10px] text-muted-foreground">{{ item.api }}</p><p class="mt-2 text-xs leading-5 text-muted-foreground">{{ item.detail }}</p><p class="mt-3 font-mono text-lg font-black">—</p></article></div>
      </section>

</template>
