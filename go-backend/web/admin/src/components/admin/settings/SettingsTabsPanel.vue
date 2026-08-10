<template>
  <Tabs :model-value="activeTab" class="gap-5">
    <TabsContent value="site" class="space-y-6">
      <SettingsSection title="站点资料" description="前台使用的品牌和联系信息。">
        <div class="grid gap-4 md:grid-cols-2">
          <AdminFormField label="品牌标题" description="控制 Nuxt 顶部品牌字样、SEO 站点名和页脚品牌名。">
            <Input v-model="siteSettings.brand_title" />
          </AdminFormField>
          <AdminFormField label="联系邮箱">
            <Input v-model="siteSettings.contact_email" type="email" />
          </AdminFormField>
          <AdminFormField label="联系电话">
            <Input v-model="siteSettings.contact_phone" type="tel" />
          </AdminFormField>
          <AdminFormField label="站点 Logo" description="支持输入图片 URL，或上传到媒体库后自动填入。">
            <div class="flex min-w-0 items-center gap-2">
              <Input v-model="siteSettings.site_logo" type="url" placeholder="Logo URL" :disabled="uploadingSiteLogo" />
              <Button type="button" variant="outline" size="icon" :disabled="!canEdit || uploadingSiteLogo" title="上传 Logo" @click="chooseSiteLogo">
                <LoaderCircle v-if="uploadingSiteLogo" class="size-4 animate-spin" />
                <ImagePlus v-else class="size-4" />
              </Button>
              <Button v-if="siteSettings.site_logo" type="button" variant="ghost" size="icon" :disabled="!canEdit || uploadingSiteLogo" title="清空 Logo" @click="clearSiteLogo">
                <Trash2 class="size-4" />
              </Button>
            </div>
            <input ref="siteLogoInput" type="file" class="sr-only" accept="image/jpeg,image/png,image/webp,image/gif" :disabled="!canEdit || uploadingSiteLogo" @change="uploadSiteLogo" />
          </AdminFormField>
          <AdminFormField label="站点描述" class="md:col-span-2">
            <Textarea v-model="siteSettings.site_description" class="min-h-24" />
          </AdminFormField>
          <div v-if="siteSettings.site_logo" class="flex h-28 items-center justify-center overflow-hidden rounded-lg border bg-muted md:col-span-2">
            <img :src="siteSettings.site_logo" alt="站点 Logo 预览" class="max-h-full max-w-full object-contain" />
          </div>
        </div>
      </SettingsSection>

      <SettingsSection title="图片版权取证" description="上传时冻结到图片版权证据包，不会因后续修改设置而改写历史记录。">
        <div class="grid gap-4 md:grid-cols-2">
          <AdminFormField label="权利人">
            <Input v-model="siteSettings.copyright_holder" placeholder="公司或版权主体名称" />
          </AdminFormField>
          <AdminFormField label="版权政策 URL">
            <Input v-model="siteSettings.copyright_url" type="url" placeholder="https://example.com/policies/copyright" />
          </AdminFormField>
          <AdminFormField label="版权声明" class="md:col-span-2">
            <Textarea
              v-model="siteSettings.copyright_notice"
              class="min-h-20"
              placeholder="Copyright 2026 Example. All rights reserved."
            />
          </AdminFormField>
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

    <TabsContent value="social">
      <SettingsSection title="社交媒体" description="前台展示的官方账号与页面链接。">
        <div class="grid gap-4 md:grid-cols-2">
          <AdminFormField v-for="field in socialFields" :key="field.key" :label="field.label">
            <Input v-model="socialSettings[field.key]" type="url" :placeholder="field.placeholder" />
          </AdminFormField>
        </div>
      </SettingsSection>
    </TabsContent>

    <TabsContent value="currency">
      <CurrencyPolicySettingsCard :can-edit="canEdit" @saved="emit('currency-policy-saved', $event)" />
    </TabsContent>

    <TabsContent value="markets">
      <StorefrontMarketsSettingsPanel :can-edit="canEdit" />
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

            <PaymentGatewayRuntimePanel
              v-model:selected-gateway="paymentSettings.gateway"
              v-model:test-mode="paymentSettings.test_mode"
              :runtime="paymentRuntime"
              :loading="loadingPaymentRuntime"
              :can-edit="canEdit"
              @refresh="emit('refresh-payment-runtime')"
            />
          </div>
        </SettingsSection>

        <PaymentMethodsSettingsPanel :can-edit="canEdit" />
      </div>
    </TabsContent>

    <TabsContent value="api">
      <SettingsSection title="API 管理" description="统一管理第三方接口配置、刷新策略和凭据引用。">
        <ApiManagementSettingsPanel
          :api-settings="apiSettings"
          :primary-pricing-currency="primaryPricingCurrency"
          :can-edit="canEdit"
          :syncing-exchange-rates="syncingExchangeRates"
          :saving-api-settings="savingApiSettings"
          @sync-exchange-rates="emit('sync-exchange-rates')"
        />
      </SettingsSection>
    </TabsContent>

    <TabsContent value="commercial_crawler">
      <CommercialCrawlerProtectionPanel
        :protection="commercialCrawlerProtection"
        :loading="loadingCommercialCrawlerProtection"
        @refresh="emit('refresh-commercial-crawler-protection')"
      />
    </TabsContent>

    <TabsContent value="public_chat" class="space-y-4">
      <PublicChatSettingsPanel
        :loading-agents="loadingPublicChatAgents"
        :loading-groups="loadingPublicChatGroups"
        :loading-candidates="loadingPublicChatAgentCandidates"
        :summary="publicChatAgentsSummary"
        :agents="publicChatAgents"
        :groups="publicChatGroups"
        :warnings="publicChatAgentWarnings"
        :can-edit="canEdit"
        @open-agent-dialog="emit('open-agent-dialog')"
        @open-group-dialog="emit('open-group-dialog')"
        @edit-group="emit('edit-group', $event)"
        @delete-group="emit('delete-group', $event)"
        @refresh="emit('refresh-public-chat')"
      />
    </TabsContent>
  </Tabs>
</template>

<script setup lang="ts">
import { defineComponent, h, ref } from 'vue'
import type { PropType } from 'vue'
import {
  AlertTriangle,
  Eye,
  EyeOff,
  ImagePlus,
  LoaderCircle,
  Trash2,
} from '@lucide/vue'
import AdminFormField from '@/components/admin/AdminFormField.vue'
import ApiManagementSettingsPanel from '@/components/admin/settings/ApiManagementSettingsPanel.vue'
import CurrencyPolicySettingsCard from '@/components/admin/settings/CurrencyPolicySettingsCard.vue'
import PaymentGatewayRuntimePanel from '@/components/admin/settings/PaymentGatewayRuntimePanel.vue'
import PaymentMethodsSettingsPanel from '@/components/admin/settings/PaymentMethodsSettingsPanel.vue'
import PublicChatSettingsPanel from '@/components/admin/settings/PublicChatSettingsPanel.vue'
import StorefrontMarketsSettingsPanel from '@/components/admin/settings/StorefrontMarketsSettingsPanel.vue'
import { Button } from '@/components/ui/button'
import CommercialCrawlerProtectionPanel from '@/components/admin/settings/CommercialCrawlerProtectionPanel.vue'
import { Input } from '@/components/ui/input'
import { Tabs, TabsContent } from '@/components/ui/tabs'
import { Textarea } from '@/components/ui/textarea'
import type {
  CommercialCrawlerProtection,
  PublicChatAgent,
  PublicChatGroup,
  PublicChatSummary,
} from './settingsTypes'

interface SocialField {
  key: string
  label: string
  placeholder: string
}

interface APISettings {
  exchange_rate_enabled: boolean | string | number
  exchange_rate_provider: string
  exchange_rate_endpoint: string
  exchange_rate_query_template: string
  exchange_rate_refresh_minutes: number
  exchange_rate_api_key: string
  time_api_enabled: boolean | string | number
  time_api_provider: string
  time_api_endpoint: string
  time_api_query_template: string
  time_api_default_timezone: string
  time_api_refresh_minutes: number
  time_api_key_ref: string
}

const props = defineProps({
  activeTab: { type: String, default: 'site' },
  siteSettings: { type: Object, required: true },
  emailSettings: { type: Object, required: true },
  socialSettings: { type: Object, required: true },
  paymentSettings: { type: Object, required: true },
  apiSettings: { type: Object as PropType<APISettings>, required: true },
  primaryPricingCurrency: { type: String, default: '' },
  commercialCrawlerProtection: { type: Object as PropType<CommercialCrawlerProtection | null>, default: null },
  loadingCommercialCrawlerProtection: { type: Boolean, default: false },
  uploadingSiteLogo: { type: Boolean, default: false },
  paymentRuntime: { type: Object, default: null },
  loadingPaymentRuntime: { type: Boolean, default: false },
  syncingExchangeRates: { type: Boolean, default: false },
  savingApiSettings: { type: Boolean, default: false },
  socialFields: { type: Array as PropType<SocialField[]>, default: () => [] },
  showSmtpPassword: { type: Boolean, default: false },
  showPaymentSecrets: { type: Boolean, default: false },
  loadingPublicChatAgents: { type: Boolean, default: false },
  loadingPublicChatGroups: { type: Boolean, default: false },
  loadingPublicChatAgentCandidates: { type: Boolean, default: false },
  publicChatAgentsSummary: { type: Object as PropType<PublicChatSummary>, default: () => ({}) },
  publicChatAgents: { type: Array as PropType<PublicChatAgent[]>, default: () => [] },
  publicChatGroups: { type: Array as PropType<PublicChatGroup[]>, default: () => [] },
  publicChatAgentWarnings: { type: Array as PropType<string[]>, default: () => [] },
  canEdit: { type: Boolean, default: false },
})

const emit = defineEmits([
  'update:showSmtpPassword',
  'update:showPaymentSecrets',
  'upload-site-logo',
  'open-agent-dialog',
  'open-group-dialog',
  'edit-group',
  'delete-group',
  'refresh-public-chat',
  'refresh-payment-runtime',
  'currency-policy-saved',
  'sync-exchange-rates',
  'refresh-commercial-crawler-protection',
])

const siteLogoInput = ref<HTMLInputElement | null>(null)

const chooseSiteLogo = () => {
  if (!props.canEdit || props.uploadingSiteLogo) return
  siteLogoInput.value?.click()
}

const uploadSiteLogo = (event: Event) => {
  const input = event.target as HTMLInputElement
  const file = input.files?.[0]
  input.value = ''
  if (!file) return
  emit('upload-site-logo', file)
}

const clearSiteLogo = () => {
  if (props.uploadingSiteLogo) return
  props.siteSettings.site_logo = ''
}

const SettingsSection = defineComponent({
  props: {
    title: { type: String, required: true },
    description: { type: String, default: '' }
  },
  setup(props, { slots }) {
    return () => h('section', { class: 'grid w-full max-w-none gap-6 lg:grid-cols-[200px_minmax(0,1fr)]' }, [
      h('div', {}, [
        h('h2', { class: 'text-sm font-black tracking-tighter italic uppercase text-foreground' }, props.title),
        props.description ? h('p', { class: 'mt-1 text-[9px] font-black uppercase tracking-widest text-muted-foreground/60' }, props.description) : null
      ]),
      h('div', { class: 'min-w-0' }, slots.default?.())
    ])
  }
})
</script>
