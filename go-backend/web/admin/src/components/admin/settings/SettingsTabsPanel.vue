<template>
  <Tabs :model-value="activeTab" class="gap-5" @update:model-value="emit('update:activeTab', $event)">
    <TabsList variant="line" class="h-10 w-full justify-start overflow-x-auto rounded-none border-b bg-transparent p-0">
      <TabsTrigger value="site" class="h-9 flex-none px-3"><Globe2 class="size-4" />站点</TabsTrigger>
      <TabsTrigger value="email" class="h-9 flex-none px-3"><Mail class="size-4" />邮件</TabsTrigger>
      <TabsTrigger value="seo" class="h-9 flex-none px-3"><SearchCheck class="size-4" />SEO</TabsTrigger>
      <TabsTrigger value="social" class="h-9 flex-none px-3"><Share2 class="size-4" />社交媒体</TabsTrigger>
      <TabsTrigger value="payment" class="h-9 flex-none px-3"><CreditCard class="size-4" />支付</TabsTrigger>
      <TabsTrigger value="public_chat" class="h-9 flex-none px-3"><Headset class="size-4" />Public Chat</TabsTrigger>
    </TabsList>

    <TabsContent value="site" class="space-y-6">
      <SettingsSection title="站点资料" description="前台使用的品牌和联系信息。">
        <div class="grid gap-4 md:grid-cols-2">
          <AdminFormField label="站点名称">
            <Input v-model="siteSettings.site_name" />
          </AdminFormField>
          <AdminFormField label="顶部品牌标题" description="控制 Nuxt 顶部特殊字体渐变字样；留空则前台不显示标题。">
            <Input v-model="siteSettings.brand_title" />
          </AdminFormField>
          <AdminFormField label="联系邮箱">
            <Input v-model="siteSettings.contact_email" type="email" />
          </AdminFormField>
          <AdminFormField label="联系电话">
            <Input v-model="siteSettings.contact_phone" type="tel" />
          </AdminFormField>
          <AdminFormField label="站点 URL">
            <Input v-model="siteSettings.site_url" type="url" placeholder="https://example.com" />
          </AdminFormField>
          <AdminFormField label="站点 Logo">
            <Input v-model="siteSettings.site_logo" placeholder="Logo URL" />
          </AdminFormField>
          <AdminFormField label="站点描述" class="md:col-span-2">
            <Textarea v-model="siteSettings.site_description" class="min-h-24" />
          </AdminFormField>
          <div v-if="siteSettings.site_logo" class="flex h-28 items-center justify-center overflow-hidden rounded-lg border bg-muted md:col-span-2">
            <img :src="siteSettings.site_logo" alt="站点 Logo 预览" class="max-h-20 max-w-[80%] object-contain" />
          </div>
        </div>
      </SettingsSection>

      <SettingsSection title="后台品牌" description="登录页、后台面板与浏览器标题使用的白标文案。">
        <div class="grid gap-4 md:grid-cols-2">
          <AdminFormField label="后台品牌名称" description="登录页左上角展示；留空则不显示品牌名。">
            <Input v-model="siteSettings.admin_brand_name" />
          </AdminFormField>
          <AdminFormField label="后台品牌缩写" description="圆形标识内的字母或短文本；留空则不显示缩写。">
            <Input v-model="siteSettings.admin_brand_initial" maxlength="4" />
          </AdminFormField>
          <AdminFormField label="面板标签" description="登录页右上角和后台顶部的小标签。">
            <Input v-model="siteSettings.admin_panel_label" />
          </AdminFormField>
          <AdminFormField label="登录标题">
            <Input v-model="siteSettings.admin_login_title" />
          </AdminFormField>
          <AdminFormField label="浏览器标题">
            <Input v-model="siteSettings.admin_html_title" />
          </AdminFormField>
          <AdminFormField label="登录页脚" description="留空则隐藏登录页底部文案。">
            <Input v-model="siteSettings.admin_footer_text" />
          </AdminFormField>
        </div>
      </SettingsSection>
    </TabsContent>

    <TabsContent value="email">
      <SettingsSection title="SMTP 配置" description="用于系统通知与业务邮件发送。">
        <div class="grid gap-4 md:grid-cols-2">
          <AdminFormField label="SMTP 主机">
            <Input v-model="emailSettings.smtp_host" placeholder="smtp.example.com" />
          </AdminFormField>
          <AdminFormField label="SMTP 端口">
            <Input v-model.number="emailSettings.smtp_port" type="number" min="1" max="65535" />
          </AdminFormField>
          <AdminFormField label="SMTP 用户名">
            <Input v-model="emailSettings.smtp_username" autocomplete="off" />
          </AdminFormField>
          <AdminFormField label="SMTP 密码">
            <div class="relative">
              <Input
                v-model="emailSettings.smtp_password"
                :type="showSmtpPassword ? 'text' : 'password'"
                class="pr-10"
                autocomplete="new-password"
              />
              <Button
                type="button"
                variant="ghost"
                size="icon"
                class="absolute right-0 top-0"
                :aria-label="showSmtpPassword ? '隐藏 SMTP 密码' : '显示 SMTP 密码'"
                @click="emit('update:showSmtpPassword', !showSmtpPassword)"
              >
                <EyeOff v-if="showSmtpPassword" class="size-4" />
                <Eye v-else class="size-4" />
              </Button>
            </div>
          </AdminFormField>
          <AdminFormField label="发件人邮箱">
            <Input v-model="emailSettings.from_email" type="email" />
          </AdminFormField>
          <AdminFormField label="发件人名称">
            <Input v-model="emailSettings.from_name" />
          </AdminFormField>
        </div>
      </SettingsSection>
    </TabsContent>

    <TabsContent value="seo">
      <SettingsSection title="默认搜索信息" description="未单独配置页面 SEO 时使用的默认值。">
        <div class="grid gap-4 md:grid-cols-2">
          <AdminFormField label="Meta 标题" class="md:col-span-2">
            <Input v-model="seoSettings.meta_title" />
          </AdminFormField>
          <AdminFormField label="Meta 描述" class="md:col-span-2">
            <Textarea v-model="seoSettings.meta_description" class="min-h-24" />
          </AdminFormField>
          <AdminFormField label="Meta 关键词" class="md:col-span-2">
            <Input v-model="seoSettings.meta_keywords" placeholder="用逗号分隔" />
          </AdminFormField>
          <AdminFormField label="Google Analytics">
            <Input v-model="seoSettings.google_analytics" placeholder="GA 跟踪 ID" />
          </AdminFormField>
          <AdminFormField label="Google Tag Manager">
            <Input v-model="seoSettings.google_tag_manager" placeholder="GTM ID" />
          </AdminFormField>
        </div>
      </SettingsSection>
    </TabsContent>

    <TabsContent value="social">
      <SettingsSection title="社交媒体" description="前台展示的官方账号与页面链接。">
        <div class="grid gap-4 md:grid-cols-2">
          <AdminFormField v-for="field in socialFields" :key="field.key" :label="field.label">
            <Input v-model="socialSettings[field.key]" type="url" :placeholder="field.placeholder" />
          </AdminFormField>
        </div>
      </SettingsSection>
    </TabsContent>

    <TabsContent value="payment">
      <div class="space-y-6">
        <SettingsSection title="支付网关" description="后端 runtime 就绪检查，不回显生产密钥。">
          <div class="space-y-4">
            <div class="flex items-start gap-3 rounded-xl border border-amber-500/20 bg-amber-500/10 p-3 text-amber-900 dark:text-amber-100">
              <AlertTriangle class="mt-0.5 size-4 flex-none" />
              <div class="space-y-1">
                <p class="text-sm font-black">支付密钥不在普通 settings 明文管理</p>
                <p class="text-xs leading-relaxed text-amber-800/80 dark:text-amber-100/75">
                  生产密钥通过下方加密写入口保存；接口只回配置状态，不回显明文。
                </p>
              </div>
            </div>

            <div class="grid gap-4 xl:grid-cols-[minmax(0,1fr)_280px]">
              <div class="rounded-2xl border bg-card/75 p-4 shadow-sm">
                <div class="flex flex-wrap items-start justify-between gap-3">
                  <div>
                    <p class="text-xs font-black uppercase tracking-widest text-muted-foreground/60">Gateway Runtime</p>
                    <h3 class="mt-1 text-base font-black text-foreground">支付服务商就绪状态</h3>
                  </div>
                  <Button
                    type="button"
                    variant="outline"
                    size="sm"
                    :disabled="loadingPaymentRuntime"
                    @click="emit('refresh-payment-runtime')"
                  >
                    <RefreshCw :class="['size-3.5', loadingPaymentRuntime ? 'animate-spin' : '']" />
                    刷新
                  </Button>
                </div>

                <div class="mt-4 grid gap-3 sm:grid-cols-2 xl:grid-cols-4">
                  <button
                    v-for="gateway in paymentGatewayOptions"
                    :key="gateway.value"
                    type="button"
                    class="group min-h-28 rounded-xl border p-3 text-left transition-all hover:-translate-y-0.5"
                    :class="gatewayCardClass(paymentRuntime, gateway.value, paymentSettings.gateway === gateway.value)"
                    @click="paymentSettings.gateway = gateway.value"
                  >
                    <div class="flex items-center justify-between gap-2">
                      <span
                        class="text-sm font-black"
                        :class="paymentSettings.gateway === gateway.value ? 'text-orange-500' : 'text-foreground'"
                      >
                        {{ gateway.label }}
                      </span>
                      <CheckCircle2
                        v-if="gatewayRuntimeStatus(paymentRuntime, gateway.value)?.production_ready"
                        class="size-4 text-emerald-500"
                      />
                      <XCircle
                        v-else-if="gatewayRuntimeStatus(paymentRuntime, gateway.value) && !gatewayRuntimeStatus(paymentRuntime, gateway.value)?.webhook_supported"
                        class="size-4 text-rose-500"
                      />
                      <AlertTriangle v-else class="size-4 text-amber-500" />
                    </div>
                    <p class="mt-2 text-xs leading-relaxed text-muted-foreground">
                      {{ gateway.description }}
                    </p>
                    <span
                      class="mt-3 inline-flex rounded-full border px-2 py-0.5 text-[11px] font-black"
                      :class="runtimeStatusBadgeClass(gatewayRuntimeStatus(paymentRuntime, gateway.value))"
                    >
                      {{ runtimeStatusLabel(gatewayRuntimeStatus(paymentRuntime, gateway.value), loadingPaymentRuntime) }}
                    </span>
                  </button>
                </div>
              </div>

              <div class="rounded-2xl border bg-muted/30 p-4">
                <p class="text-xs font-black uppercase tracking-widest text-muted-foreground/60">Selected Gateway</p>
                <div class="mt-3 space-y-3">
                  <div>
                    <div class="text-sm font-black text-foreground">{{ paymentGatewayLabel(paymentSettings.gateway) }}</div>
                    <p class="mt-1 text-xs leading-relaxed text-muted-foreground">
                      {{ paymentGatewayDescription(paymentSettings.gateway) }}
                    </p>
                  </div>

                  <div class="flex items-center justify-between gap-3 rounded-xl border bg-background/70 px-3 py-2.5">
                    <div>
                      <span class="text-xs font-bold text-foreground">后台首选测试模式</span>
                      <p class="mt-0.5 text-xs text-muted-foreground">保存为后台记录；真实支付环境看 runtime。</p>
                    </div>
                    <Switch v-model="paymentSettings.test_mode" aria-label="支付测试模式" />
                  </div>

                  <div class="flex items-center justify-between text-xs">
                    <span class="text-muted-foreground">后台记录</span>
                    <span
                      class="rounded-full px-2 py-0.5 font-black"
                      :class="paymentSettings.test_mode ? 'bg-amber-500/10 text-amber-600 dark:text-amber-300' : 'bg-emerald-500/10 text-emerald-600 dark:text-emerald-300'"
                    >
                      {{ paymentSettings.test_mode ? '测试' : '生产' }}
                    </span>
                  </div>

                  <div class="flex items-center justify-between text-xs">
                    <span class="text-muted-foreground">Runtime Source</span>
                    <span class="font-bold text-foreground">{{ paymentRuntime?.runtime_source || 'environment' }}</span>
                  </div>
                </div>
              </div>
            </div>

            <div class="rounded-2xl border bg-card/75 p-4 shadow-sm">
              <div class="flex flex-wrap items-start justify-between gap-3">
                <div>
                  <p class="text-xs font-black uppercase tracking-widest text-muted-foreground/60">Readiness Detail</p>
                  <h3 class="mt-1 text-base font-black text-foreground">生产就绪检查</h3>
                </div>
                <span
                  class="rounded-full border px-2.5 py-1 text-[11px] font-black"
                  :class="runtimeStatusBadgeClass(gatewayRuntimeStatus(paymentRuntime, paymentSettings.gateway))"
                >
                  {{ runtimeStatusLabel(gatewayRuntimeStatus(paymentRuntime, paymentSettings.gateway), loadingPaymentRuntime) }}
                </span>
              </div>

              <div v-if="loadingPaymentRuntime" class="mt-4 flex h-28 items-center justify-center text-xs text-muted-foreground">
                <RefreshCw class="mr-2 size-4 animate-spin" />
                正在检查支付运行配置
              </div>

              <div v-else-if="gatewayRuntimeStatus(paymentRuntime, paymentSettings.gateway)" class="mt-4 space-y-4">
                <div class="grid gap-3 md:grid-cols-3">
                  <div class="rounded-xl border bg-background/70 p-3">
                    <p class="text-[11px] font-black uppercase tracking-widest text-muted-foreground/60">Environment</p>
                    <p class="mt-1 text-sm font-black text-foreground">
                      {{ gatewayRuntimeStatus(paymentRuntime, paymentSettings.gateway).environment || 'unknown' }}
                    </p>
                  </div>
                  <div class="rounded-xl border bg-background/70 p-3">
                    <p class="text-[11px] font-black uppercase tracking-widest text-muted-foreground/60">Credentials</p>
                    <p class="mt-1 text-sm font-black" :class="gatewayRuntimeStatus(paymentRuntime, paymentSettings.gateway).configured ? 'text-emerald-500' : 'text-amber-500'">
                      {{ gatewayRuntimeStatus(paymentRuntime, paymentSettings.gateway).configured ? '已配置' : '缺字段' }}
                    </p>
                  </div>
                  <div class="rounded-xl border bg-background/70 p-3">
                    <p class="text-[11px] font-black uppercase tracking-widest text-muted-foreground/60">Webhook</p>
                    <p class="mt-1 text-sm font-black" :class="gatewayRuntimeStatus(paymentRuntime, paymentSettings.gateway).webhook_configured ? 'text-emerald-500' : 'text-amber-500'">
                      {{ gatewayRuntimeStatus(paymentRuntime, paymentSettings.gateway).webhook_configured ? '已配置' : '缺配置' }}
                    </p>
                  </div>
                </div>

                <div class="rounded-xl border bg-background/70 p-3">
                  <div class="flex flex-wrap items-center justify-between gap-2">
                    <p class="text-[11px] font-black uppercase tracking-widest text-muted-foreground/60">Callback URL</p>
                    <Button
                      type="button"
                      variant="ghost"
                      size="icon"
                      class="size-8"
                      aria-label="复制支付回调地址"
                      @click="copyText(gatewayRuntimeStatus(paymentRuntime, paymentSettings.gateway).callback_url)"
                    >
                      <Copy class="size-3.5" />
                    </Button>
                  </div>
                  <p class="mt-1 break-all font-mono text-xs text-foreground">
                    {{ gatewayRuntimeStatus(paymentRuntime, paymentSettings.gateway).callback_url }}
                  </p>
                </div>

                <div class="grid gap-3 md:grid-cols-2">
                  <div class="rounded-xl border bg-background/70 p-3">
                    <p class="text-[11px] font-black uppercase tracking-widest text-muted-foreground/60">Required Runtime Fields</p>
                    <div class="mt-2 flex flex-wrap gap-1.5">
                      <span
                        v-for="field in gatewayRuntimeStatus(paymentRuntime, paymentSettings.gateway).required_fields"
                        :key="field"
                        class="rounded-full border bg-muted px-2 py-0.5 font-mono text-[11px] font-bold text-foreground"
                      >
                        {{ field }}
                      </span>
                    </div>
                  </div>
                  <div class="rounded-xl border bg-background/70 p-3">
                    <p class="text-[11px] font-black uppercase tracking-widest text-muted-foreground/60">Configured Fields</p>
                    <div class="mt-2 flex flex-wrap gap-1.5">
                      <span
                        v-for="field in gatewayRuntimeStatus(paymentRuntime, paymentSettings.gateway).configured_fields"
                        :key="field"
                        class="rounded-full border border-emerald-500/20 bg-emerald-500/10 px-2 py-0.5 font-mono text-[11px] font-bold text-emerald-700 dark:text-emerald-200"
                      >
                        {{ field }}
                      </span>
                      <span
                        v-if="gatewayRuntimeStatus(paymentRuntime, paymentSettings.gateway).configured_fields.length === 0"
                        class="text-xs text-muted-foreground"
                      >
                        暂无
                      </span>
                    </div>
                  </div>
                </div>

                <div
                  v-if="gatewayRuntimeStatus(paymentRuntime, paymentSettings.gateway).missing.length"
                  class="rounded-xl border border-amber-500/20 bg-amber-500/10 p-3"
                >
                  <p class="text-xs font-black text-amber-800 dark:text-amber-100">缺失字段</p>
                  <p class="mt-1 text-xs leading-relaxed text-amber-800/80 dark:text-amber-100/75">
                    {{ gatewayRuntimeStatus(paymentRuntime, paymentSettings.gateway).missing.join(', ') }}
                  </p>
                </div>

                <div
                  v-if="gatewayRuntimeStatus(paymentRuntime, paymentSettings.gateway).blockers.length"
                  class="rounded-xl border border-rose-500/20 bg-rose-500/10 p-3"
                >
                  <p class="text-xs font-black text-rose-700 dark:text-rose-100">生产阻塞</p>
                  <ul class="mt-1 space-y-1 text-xs leading-relaxed text-rose-700/85 dark:text-rose-100/80">
                    <li v-for="blocker in gatewayRuntimeStatus(paymentRuntime, paymentSettings.gateway).blockers" :key="blocker">
                      {{ blocker }}
                    </li>
                  </ul>
                </div>

                <a
                  v-if="gatewayRuntimeStatus(paymentRuntime, paymentSettings.gateway).documentation_url"
                  :href="gatewayRuntimeStatus(paymentRuntime, paymentSettings.gateway).documentation_url"
                  target="_blank"
                  rel="noreferrer"
                  class="inline-flex items-center gap-2 text-xs font-black text-orange-500 hover:text-orange-400"
                >
                  <ShieldCheck class="size-3.5" />
                  {{ gatewayRuntimeStatus(paymentRuntime, paymentSettings.gateway).documentation_label }}
                </a>
              </div>

              <div v-else class="mt-4 rounded-xl border border-dashed p-5 text-sm text-muted-foreground">
                选择一个支付服务商后查看生产就绪检查。
              </div>
            </div>

            <PaymentGatewaySecureConfigPanel
              :selected-gateway="paymentSettings.gateway"
              :status="gatewayRuntimeStatus(paymentRuntime, paymentSettings.gateway)"
              :secret-store-configured="paymentRuntime?.secret_store_configured === true"
              :can-edit="canEdit"
              @saved="emit('refresh-payment-runtime')"
            />
          </div>
        </SettingsSection>

        <PaymentMethodsSettingsPanel :can-edit="canEdit" />
      </div>
    </TabsContent>

    <TabsContent value="public_chat" class="space-y-4">
      <PublicChatSettingsPanel
        :loading-agents="loadingPublicChatAgents"
        :loading-candidates="loadingPublicChatAgentCandidates"
        :summary="publicChatAgentsSummary"
        :agents="publicChatAgents"
        :warnings="publicChatAgentWarnings"
        :can-edit="canEdit"
        @open-agent-dialog="emit('open-agent-dialog')"
        @refresh="emit('refresh-public-chat')"
      />
    </TabsContent>
  </Tabs>
</template>

<script setup>
import { defineComponent, h } from 'vue'
import {
  AlertTriangle,
  CheckCircle2,
  Copy,
  CreditCard,
  Eye,
  EyeOff,
  Globe2,
  Headset,
  Mail,
  RefreshCw,
  SearchCheck,
  ShieldCheck,
  Share2,
  XCircle,
} from '@lucide/vue'
import AdminFormField from '@/components/admin/AdminFormField.vue'
import PaymentGatewaySecureConfigPanel from '@/components/admin/settings/PaymentGatewaySecureConfigPanel.vue'
import PaymentMethodsSettingsPanel from '@/components/admin/settings/PaymentMethodsSettingsPanel.vue'
import PublicChatSettingsPanel from '@/components/admin/settings/PublicChatSettingsPanel.vue'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Switch } from '@/components/ui/switch'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Textarea } from '@/components/ui/textarea'

defineProps({
  activeTab: { type: String, default: 'site' },
  siteSettings: { type: Object, required: true },
  emailSettings: { type: Object, required: true },
  seoSettings: { type: Object, required: true },
  socialSettings: { type: Object, required: true },
  paymentSettings: { type: Object, required: true },
  paymentRuntime: { type: Object, default: null },
  loadingPaymentRuntime: { type: Boolean, default: false },
  socialFields: { type: Array, default: () => [] },
  showSmtpPassword: { type: Boolean, default: false },
  showPaymentSecrets: { type: Boolean, default: false },
  loadingPublicChatAgents: { type: Boolean, default: false },
  loadingPublicChatAgentCandidates: { type: Boolean, default: false },
  publicChatAgentsSummary: { type: Object, default: () => ({}) },
  publicChatAgents: { type: Array, default: () => [] },
  publicChatAgentWarnings: { type: Array, default: () => [] },
  canEdit: { type: Boolean, default: false },
})

const emit = defineEmits([
  'update:activeTab',
  'update:showSmtpPassword',
  'update:showPaymentSecrets',
  'open-agent-dialog',
  'refresh-public-chat',
  'refresh-payment-runtime',
])

const paymentGatewayOptions = [
  { value: 'stripe', label: 'Stripe', description: 'Cards, wallets and international card checkout.' },
  { value: 'paypal', label: 'PayPal', description: 'PayPal account checkout and express payments.' },
  { value: 'alipay', label: '支付宝', description: '适合人民币和跨境支付宝收款场景。' },
  { value: 'wechat', label: '微信支付', description: '适合微信生态内的扫码和小程序支付。' },
]

const paymentGatewayOption = (value) =>
  paymentGatewayOptions.find((gateway) => gateway.value === value)

const paymentGatewayLabel = (value) =>
  paymentGatewayOption(value)?.label || '未选择支付网关'

const paymentGatewayDescription = (value) =>
  paymentGatewayOption(value)?.description || '请选择一个支付服务商后查看 runtime 检查。'

const gatewayRuntimeStatus = (runtime, value) =>
  (runtime?.gateways || []).find((gateway) => gateway.provider === value)

const runtimeStatusLabel = (status, loading) => {
  if (loading) return '检查中'
  if (!status) return '未知'
  if (status.production_ready) return '生产就绪'
  if (!status.webhook_supported) return '锁定'
  if (status.configured || status.webhook_configured) return '需补配置'
  return '缺配置'
}

const runtimeStatusBadgeClass = (status) => {
  if (!status) return 'border-border bg-muted text-muted-foreground'
  if (status.production_ready) return 'border-emerald-500/20 bg-emerald-500/10 text-emerald-700 dark:text-emerald-200'
  if (!status.webhook_supported) return 'border-rose-500/20 bg-rose-500/10 text-rose-700 dark:text-rose-200'
  if (status.configured || status.webhook_configured) return 'border-amber-500/20 bg-amber-500/10 text-amber-700 dark:text-amber-200'
  return 'border-border bg-muted text-muted-foreground'
}

const gatewayCardClass = (runtime, value, isSelected) => {
  const status = gatewayRuntimeStatus(runtime, value)
  const selectedClass = isSelected ? 'border-orange-500/55 bg-orange-500/10 shadow-[0_16px_34px_rgba(255,90,0,0.10)]' : ''
  if (selectedClass) return selectedClass
  if (status?.production_ready) return 'border-emerald-500/20 bg-emerald-500/5 hover:border-emerald-500/35'
  if (status && !status.webhook_supported) return 'border-rose-500/20 bg-rose-500/5 hover:border-rose-500/35'
  if (status?.configured || status?.webhook_configured) return 'border-amber-500/20 bg-amber-500/5 hover:border-amber-500/35'
  return 'border-border bg-background/70 hover:border-orange-500/35 hover:bg-orange-500/5'
}

const copyText = (value) => {
  if (!value || typeof navigator === 'undefined' || !navigator.clipboard) return
  navigator.clipboard.writeText(value)
}

const SettingsSection = defineComponent({
  props: {
    title: { type: String, required: true },
    description: { type: String, default: '' }
  },
  setup(props, { slots }) {
    return () => h('section', { class: 'grid max-w-5xl gap-5 lg:grid-cols-[190px_minmax(0,1fr)]' }, [
      h('div', {}, [
        h('h2', { class: 'text-sm font-black tracking-tighter italic uppercase text-foreground' }, props.title),
        props.description ? h('p', { class: 'mt-1 text-[9px] font-black uppercase tracking-widest text-muted-foreground/60' }, props.description) : null
      ]),
      h('div', { class: 'min-w-0' }, slots.default?.())
    ])
  }
})
</script>
