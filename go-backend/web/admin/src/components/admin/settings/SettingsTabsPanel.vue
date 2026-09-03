<template>
  <Tabs :model-value="activeTab" class="gap-5">
    <TabsContent value="site" class="space-y-6">
      <SettingsSection :title="t('settings.siteProfile')" :description="t('settings.siteProfileDescription')">
        <div class="grid gap-4 md:grid-cols-2">
          <AdminFormField :label="t('settings.siteName')" description="前台顶部、SEO 和页脚使用的站点名称。">
            <Input v-model="siteSettings.site_name" />
          </AdminFormField>
          <AdminFormField :label="t('settings.contactEmail')">
            <Input v-model="siteSettings.contact_email" type="email" />
          </AdminFormField>
          <AdminFormField :label="t('settings.contactPhone')">
            <Input v-model="siteSettings.contact_phone" type="tel" />
          </AdminFormField>
          <AdminFormField :label="t('settings.siteLogo')" description="仅支持上传 512×512 WebP；更换时自动删除旧 Logo。">
            <div class="flex min-w-0 items-center gap-2">
              <Input :model-value="siteSettings.site_logo" type="text" placeholder="尚未上传站点 Logo" readonly :disabled="uploadingSiteLogo" />
              <Button type="button" variant="outline" size="icon" :disabled="!canEdit || uploadingSiteLogo" :title="t('settings.uploadLogo')" @click="chooseSiteLogo">
                <LoaderCircle v-if="uploadingSiteLogo" class="size-4 animate-spin" />
                <ImagePlus v-else class="size-4" />
              </Button>
              <Button v-if="siteSettings.site_logo" type="button" variant="ghost" size="icon" :disabled="!canEdit || uploadingSiteLogo" :title="t('settings.clearLogo')" @click="clearSiteLogo">
                <Trash2 class="size-4" />
              </Button>
            </div>
            <input ref="siteLogoInput" type="file" class="sr-only" :accept="uploadSpecAccept('site_logo')" :disabled="!canEdit || uploadingSiteLogo" @change="uploadSiteLogo" />
            <UploadSpecHint code="site_logo" />
          </AdminFormField>
          <AdminFormField :label="t('settings.siteFavicon')" description="用于浏览器 TAB、收藏夹和 PWA 图标；支持直接填写图片 URL 或上传图标。">
            <div class="flex min-w-0 items-center gap-2">
              <Input v-model="siteSettings.site_favicon" type="url" placeholder="Favicon URL" :disabled="uploadingSiteFavicon" />
              <Button type="button" variant="outline" size="icon" :disabled="!canEdit || uploadingSiteFavicon" :title="t('settings.uploadFavicon')" @click="chooseSiteFavicon">
                <LoaderCircle v-if="uploadingSiteFavicon" class="size-4 animate-spin" />
                <ImagePlus v-else class="size-4" />
              </Button>
              <Button v-if="siteSettings.site_favicon" type="button" variant="ghost" size="icon" :disabled="!canEdit || uploadingSiteFavicon" :title="t('settings.clearFavicon')" @click="clearSiteFavicon">
                <Trash2 class="size-4" />
              </Button>
            </div>
            <input ref="siteFaviconInput" type="file" class="sr-only" :accept="uploadSpecAccept('site_favicon')" :disabled="!canEdit || uploadingSiteFavicon" @change="uploadSiteFavicon" />
            <UploadSpecHint code="site_favicon" />
          </AdminFormField>
          <AdminFormField :label="t('settings.siteDescription')" class="md:col-span-2">
            <Textarea v-model="siteSettings.site_description" class="min-h-24" />
          </AdminFormField>
          <div v-if="siteSettings.site_logo" class="flex h-28 items-center justify-center overflow-hidden rounded-lg border bg-muted md:col-span-2">
            <img :src="siteSettings.site_logo" :alt="t('settings.siteLogoPreview')" class="max-h-full max-w-full object-contain" />
          </div>
          <div v-if="siteSettings.site_favicon" class="flex h-20 items-center gap-3 overflow-hidden rounded-lg border bg-muted px-4 md:col-span-2">
            <img :src="siteSettings.site_favicon" :alt="t('settings.siteFaviconPreview')" class="size-12 object-contain" />
            <span class="text-sm text-muted-foreground">{{ t('settings.siteFaviconPreview') }}</span>
          </div>
        </div>
      </SettingsSection>

      <SettingsSection :title="t('settings.adminBrand')" :description="t('settings.adminBrandDescription')">
        <div class="grid gap-4 md:grid-cols-2">
          <AdminFormField :label="t('settings.adminBrandName')" description="登录页左上角展示；留空则不显示品牌名。">
            <Input v-model="siteSettings.admin_brand_name" />
          </AdminFormField>
          <AdminFormField :label="t('settings.adminBrandInitial')" description="圆形标识内的字母或短文本；留空则不显示缩写。">
            <Input v-model="siteSettings.admin_brand_initial" maxlength="4" />
          </AdminFormField>
          <AdminFormField :label="t('settings.panelLabel')" description="登录页右上角和后台顶部的小标签。">
            <Input v-model="siteSettings.admin_panel_label" />
          </AdminFormField>
          <AdminFormField :label="t('settings.loginTitle')">
            <Input v-model="siteSettings.admin_login_title" />
          </AdminFormField>
          <AdminFormField :label="t('settings.browserTitle')">
            <Input v-model="siteSettings.admin_html_title" />
          </AdminFormField>
          <AdminFormField :label="t('settings.loginFooter')" description="留空则隐藏登录页底部文案。">
            <Input v-model="siteSettings.admin_footer_text" />
          </AdminFormField>
        </div>
      </SettingsSection>
    </TabsContent>

    <TabsContent value="email">
      <SettingsSection :title="t('settings.emailSettings')" :description="t('settings.emailSettingsDescription')">
        <div class="grid gap-4 md:grid-cols-2">
          <AdminFormField :label="t('settings.smtpHost')">
            <Input v-model="emailSettings.smtp_host" placeholder="smtp.example.com" />
          </AdminFormField>
          <AdminFormField :label="t('settings.smtpPort')">
            <Input v-model.number="emailSettings.smtp_port" type="number" min="1" max="65535" />
          </AdminFormField>
          <AdminFormField :label="t('settings.smtpUsername')">
            <Input v-model="emailSettings.smtp_username" autocomplete="off" />
          </AdminFormField>
          <AdminFormField :label="t('settings.smtpPassword')">
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
                :aria-label="showSmtpPassword ? t('settings.hideSMTPPassword') : t('settings.showSMTPPassword')"
                @click="emit('update:showSmtpPassword', !showSmtpPassword)"
              >
                <EyeOff v-if="showSmtpPassword" class="size-4" />
                <Eye v-else class="size-4" />
              </Button>
            </div>
          </AdminFormField>
          <AdminFormField :label="t('settings.senderEmail')">
            <Input v-model="emailSettings.from_email" type="email" />
          </AdminFormField>
          <AdminFormField :label="t('settings.senderName')">
            <Input v-model="emailSettings.from_name" />
          </AdminFormField>
        </div>
      </SettingsSection>
    </TabsContent>

    <TabsContent value="markets">
      <StorefrontMarketsSettingsPanel :can-edit="canEdit" />
    </TabsContent>

    <TabsContent value="api">
      <SettingsSection title="API 管理" description="统一管理第三方接口配置、缓存策略和系统时区规则。">
        <ApiManagementSettingsPanel
          :api-settings="apiSettings"
          :can-edit="canEdit"
          :saving-api-settings="savingApiSettings"
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

    <TabsContent value="refund_cancellation" class="space-y-4">
      <RefundCancellationPolicySettingsPanel
        :policy="refundCancellationPolicy"
        :locale="refundCancellationPolicyLocale"
        :fallback="refundCancellationPolicyFallback"
        :loading="loadingRefundCancellationPolicy"
        :saving="savingRefundCancellationPolicy"
        :can-edit="canEdit"
        :uploading-section="uploadingRefundCancellationSection"
        @locale-change="emit('refund-cancellation-locale-change', $event)"
        @save="emit('save-refund-cancellation-policy')"
        @upload-image="emit('upload-refund-cancellation-image', $event)"
      />
    </TabsContent>
  </Tabs>
</template>

<script setup lang="ts">
import { defineComponent, h, ref } from 'vue'
import type { PropType } from 'vue'
import {
  Eye,
  EyeOff,
  ImagePlus,
  LoaderCircle,
  Trash2,
} from '@lucide/vue'
import AdminFormField from '@/components/admin/AdminFormField.vue'
import ApiManagementSettingsPanel from '@/components/admin/settings/ApiManagementSettingsPanel.vue'
import PublicChatSettingsPanel from '@/components/admin/settings/PublicChatSettingsPanel.vue'
import RefundCancellationPolicySettingsPanel from '@/components/admin/settings/RefundCancellationPolicySettingsPanel.vue'
import StorefrontMarketsSettingsPanel from '@/components/admin/settings/StorefrontMarketsSettingsPanel.vue'
import UploadSpecHint from '@/components/admin/UploadSpecHint.vue'
import { uploadSpecAccept } from '@/lib/uploadSpecs'
import { Button } from '@/components/ui/button'
import CommercialCrawlerProtectionPanel from '@/components/admin/settings/CommercialCrawlerProtectionPanel.vue'
import { Input } from '@/components/ui/input'
import { Tabs, TabsContent } from '@/components/ui/tabs'
import { Textarea } from '@/components/ui/textarea'
import { useAdminI18n } from '@/i18n'
import type {
  ApiManagementSettings,
  CommercialCrawlerProtection,
  PublicChatAgent,
  PublicChatGroup,
  PublicChatSummary,
} from '@/modules/settings/types'
import type { RefundCancellationPolicyEditor } from '@/api/refundCancellationPolicy'

const props = defineProps({
  activeTab: { type: String, default: 'site' },
  siteSettings: { type: Object, required: true },
  emailSettings: { type: Object, required: true },
  apiSettings: { type: Object as PropType<ApiManagementSettings>, required: true },
  commercialCrawlerProtection: { type: Object as PropType<CommercialCrawlerProtection | null>, default: null },
  loadingCommercialCrawlerProtection: { type: Boolean, default: false },
  uploadingSiteLogo: { type: Boolean, default: false },
  uploadingSiteFavicon: { type: Boolean, default: false },
  savingApiSettings: { type: Boolean, default: false },
  showSmtpPassword: { type: Boolean, default: false },
  loadingPublicChatAgents: { type: Boolean, default: false },
  loadingPublicChatGroups: { type: Boolean, default: false },
  loadingPublicChatAgentCandidates: { type: Boolean, default: false },
  publicChatAgentsSummary: { type: Object as PropType<PublicChatSummary>, default: () => ({}) },
  publicChatAgents: { type: Array as PropType<PublicChatAgent[]>, default: () => [] },
  publicChatGroups: { type: Array as PropType<PublicChatGroup[]>, default: () => [] },
  publicChatAgentWarnings: { type: Array as PropType<string[]>, default: () => [] },
  refundCancellationPolicy: { type: Object as PropType<RefundCancellationPolicyEditor>, required: true },
  refundCancellationPolicyLocale: { type: String, default: 'en' },
  refundCancellationPolicyFallback: { type: Boolean, default: false },
  loadingRefundCancellationPolicy: { type: Boolean, default: false },
  savingRefundCancellationPolicy: { type: Boolean, default: false },
  uploadingRefundCancellationSection: { type: Number as PropType<number | null>, default: null },
  canEdit: { type: Boolean, default: false },
})

const emit = defineEmits([
  'update:showSmtpPassword',
  'upload-site-logo',
  'clear-site-logo',
  'upload-site-favicon',
  'open-agent-dialog',
  'open-group-dialog',
  'edit-group',
  'delete-group',
  'refresh-public-chat',
  'refresh-commercial-crawler-protection',
  'refund-cancellation-locale-change',
  'save-refund-cancellation-policy',
  'upload-refund-cancellation-image',
])

const siteLogoInput = ref<HTMLInputElement | null>(null)
const siteFaviconInput = ref<HTMLInputElement | null>(null)
const { t } = useAdminI18n()

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

const chooseSiteFavicon = () => {
  if (!props.canEdit || props.uploadingSiteFavicon) return
  siteFaviconInput.value?.click()
}

const uploadSiteFavicon = (event: Event) => {
  const input = event.target as HTMLInputElement
  const file = input.files?.[0]
  input.value = ''
  if (!file) return
  emit('upload-site-favicon', file)
}

const clearSiteLogo = () => {
  if (props.uploadingSiteLogo) return
  emit('clear-site-logo')
}

const clearSiteFavicon = () => {
  if (props.uploadingSiteFavicon) return
  props.siteSettings.site_favicon = ''
}

const SettingsSection = defineComponent({
  props: {
    title: { type: String, required: true },
    description: { type: String, default: '' }
  },
  setup(props, { slots }) {
 return () => h('section', { class: 'w-full max-w-none space-y-3'}, [
 h('div', { class: 'max-w-3xl'}, [
 h('h2', { class: 'text-sm font-black tracking-tighter uppercase text-foreground'}, props.title),
 props.description ? h('p', { class: 'mt-1 text-[9px] font-black uppercase tracking-widest text-muted-foreground/60'}, props.description) : null
      ]),
 h('div', { class: 'min-w-0'}, slots.default?.())
    ])
  }
})
</script>
