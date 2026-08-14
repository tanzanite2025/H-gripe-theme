<template>
  <div class="space-y-4">
    <AdminPageHeader :title="pageTitle" :description="pageDescription">
      <template #actions>
        <Button
          v-if="activeTab === 'profiles' && canEdit"
          :disabled="saving || loading"
          @click="savePublicLinks"
        >
          <LoaderCircle v-if="saving" class="size-4 animate-spin" />
          <Save v-else class="size-4" />
          {{ saving ? '保存中' : '保存链接' }}
        </Button>
        <Button v-if="activeTab === 'profiles'" variant="outline" :disabled="loading || saving" @click="refresh">
          <RefreshCw :class="['size-4', loading ? 'animate-spin' : '']" />
          刷新
        </Button>
      </template>
    </AdminPageHeader>

    <template v-if="activeTab === 'overview'">
      <section class="grid gap-3 md:grid-cols-2 xl:grid-cols-3">
        <article
          v-for="platform in platformCards"
          :key="platform.key"
          class="rounded-2xl border border-dashed bg-background p-4"
        >
          <div class="flex items-start justify-between gap-3">
            <div class="flex min-w-0 items-center gap-3">
              <span class="flex size-9 shrink-0 items-center justify-center rounded-xl bg-muted text-foreground">
                <component :is="platform.icon" class="size-4" />
              </span>
              <div class="min-w-0">
                <h2 class="truncate text-sm font-black">{{ platform.label }}</h2>
                <p class="mt-1 truncate text-[10px] text-muted-foreground">{{ platform.description }}</p>
              </div>
            </div>
            <AdminStatusBadge :tone="platform.tone">{{ platform.status }}</AdminStatusBadge>
          </div>
          <div class="mt-4 flex items-center justify-between gap-2 border-t border-dashed pt-3">
            <span class="text-[10px] font-bold text-muted-foreground">{{ platform.capability }}</span>
            <Button size="sm" variant="outline" @click="openPlatform(platform.routeName)">
              查看
            </Button>
          </div>
        </article>
      </section>

      <section class="border-y bg-muted/10 px-1 py-5">
        <div class="flex flex-wrap items-start justify-between gap-4">
          <div>
            <p class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/60">Social Workspace</p>
            <h2 class="mt-1 text-sm font-black">社交账号与内容发布工作区</h2>
            <p class="mt-1 max-w-3xl text-xs leading-5 text-muted-foreground">
              公开链接、平台账号连接和发布任务分开管理。后续绑定 YouTube 后，可以从媒体库选择视频，再在对应平台 TAB 中完成发布。
            </p>
          </div>
          <div class="flex items-center gap-2">
            <AdminStatusBadge tone="blue">域结构已就绪</AdminStatusBadge>
            <AdminStatusBadge tone="amber">OAuth 待接入</AdminStatusBadge>
          </div>
        </div>

        <div class="mt-5 grid gap-3 md:grid-cols-3">
          <div class="rounded-xl border bg-background/70 p-3">
            <div class="flex items-center gap-2 text-xs font-black">
              <Link2 class="size-4 text-sky-600" />
              前台展示
            </div>
            <p class="mt-2 text-xs leading-5 text-muted-foreground">保存官网页脚、联系页和二维码使用的官方账号链接。</p>
          </div>
          <div class="rounded-xl border bg-background/70 p-3">
            <div class="flex items-center gap-2 text-xs font-black">
              <ShieldCheck class="size-4 text-emerald-600" />
              账号连接
            </div>
            <p class="mt-2 text-xs leading-5 text-muted-foreground">未来单独保存账号、频道、权限范围和加密 Token 状态。</p>
          </div>
          <div class="rounded-xl border bg-background/70 p-3">
            <div class="flex items-center gap-2 text-xs font-black">
              <Send class="size-4 text-violet-600" />
              内容发布
            </div>
            <p class="mt-2 text-xs leading-5 text-muted-foreground">发布任务、定时任务和平台返回结果统一进入发布记录。</p>
          </div>
        </div>
      </section>
    </template>

    <template v-else-if="activeTab === 'profiles'">
      <section class="border-y bg-background px-1 py-5">
        <div class="max-w-3xl">
          <p class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/60">Storefront Profiles</p>
          <h2 class="mt-1 text-sm font-black">前台展示链接</h2>
          <p class="mt-1 text-xs leading-5 text-muted-foreground">
            这里保留原系统设置中的公开账号配置。它只负责展示，不代表已经绑定了可发布内容的平台账号。
          </p>
        </div>

        <div v-if="loading" class="flex min-h-48 items-center justify-center text-xs font-bold text-muted-foreground">
          正在加载社交媒体链接
        </div>
        <div v-else class="mt-5 grid gap-4 md:grid-cols-2">
          <AdminFormField v-for="field in socialFields" :key="field.key" :label="field.label">
            <Input v-model="socialSettings[field.key]" type="url" :placeholder="field.placeholder" :disabled="!canEdit" />
          </AdminFormField>
        </div>
      </section>

      <section class="border-y bg-muted/10 px-1 py-5">
        <div class="flex items-start gap-3">
          <Link2 class="mt-0.5 size-4 shrink-0 text-sky-600" />
          <div>
            <h2 class="text-sm font-black">数据边界</h2>
            <p class="mt-1 max-w-3xl text-xs leading-5 text-muted-foreground">
              这些 URL 仍然兼容原来的 `settings.social` 数据。YouTube 等平台的 OAuth 账号、频道和发布凭据不会写入这里。
            </p>
          </div>
        </div>
      </section>
    </template>

    <template v-else-if="activeTab === 'publications'">
      <section class="border-y bg-background px-1 py-6">
        <div class="flex flex-col items-center justify-center py-8 text-center">
          <span class="flex size-12 items-center justify-center rounded-2xl bg-muted">
            <Send class="size-5 text-muted-foreground" />
          </span>
          <h2 class="mt-4 text-sm font-black">还没有发布任务</h2>
          <p class="mt-2 max-w-md text-xs leading-5 text-muted-foreground">
            绑定平台账号并从媒体库选择视频后，发布任务、计划发布时间、平台返回链接和失败原因会显示在这里。
          </p>
        </div>
      </section>

      <section class="grid gap-3 md:grid-cols-3">
        <div class="rounded-xl border border-dashed bg-background p-4">
          <Clock3 class="size-4 text-amber-600" />
          <h3 class="mt-3 text-xs font-black">定时发布</h3>
          <p class="mt-1 text-xs leading-5 text-muted-foreground">为不同平台使用各自的发布时间和发布策略。</p>
        </div>
        <div class="rounded-xl border border-dashed bg-background p-4">
          <ShieldCheck class="size-4 text-emerald-600" />
          <h3 class="mt-3 text-xs font-black">状态追踪</h3>
          <p class="mt-1 text-xs leading-5 text-muted-foreground">记录平台任务 ID、发布链接、失败原因和重试次数。</p>
        </div>
        <div class="rounded-xl border border-dashed bg-background p-4">
          <Link2 class="size-4 text-sky-600" />
          <h3 class="mt-3 text-xs font-black">资源关联</h3>
          <p class="mt-1 text-xs leading-5 text-muted-foreground">每条发布记录都关联媒体库资源和实际使用的账号。</p>
        </div>
      </section>
    </template>

    <template v-else>
      <section class="border-y bg-background px-1 py-5">
        <div class="flex flex-wrap items-start justify-between gap-4">
          <div class="flex min-w-0 items-start gap-3">
            <span class="flex size-10 shrink-0 items-center justify-center rounded-xl bg-muted">
              <component :is="currentPlatform.icon" class="size-5" />
            </span>
            <div class="min-w-0">
              <p class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/60">Platform Connection</p>
              <h2 class="mt-1 text-sm font-black">{{ currentPlatform.label }} 账号连接</h2>
              <p class="mt-1 max-w-2xl text-xs leading-5 text-muted-foreground">{{ currentPlatform.description }}</p>
            </div>
          </div>
          <AdminStatusBadge tone="amber">待接入</AdminStatusBadge>
        </div>

        <div class="mt-5 grid gap-3 md:grid-cols-3">
          <div class="rounded-xl border bg-muted/20 p-4">
            <p class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/60">账号状态</p>
            <p class="mt-2 text-sm font-black">尚未绑定</p>
            <p class="mt-1 text-xs leading-5 text-muted-foreground">后端 OAuth 连接接口接入后显示账号名称和连接状态。</p>
          </div>
          <div class="rounded-xl border bg-muted/20 p-4">
            <p class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/60">发布能力</p>
            <p class="mt-2 text-sm font-black">{{ currentPlatform.capability }}</p>
            <p class="mt-1 text-xs leading-5 text-muted-foreground">平台发布参数会在绑定账号后按平台单独配置。</p>
          </div>
          <div class="rounded-xl border bg-muted/20 p-4">
            <p class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/60">安全边界</p>
            <p class="mt-2 text-sm font-black">独立凭据存储</p>
            <p class="mt-1 text-xs leading-5 text-muted-foreground">不会复用公开链接配置，也不会把 Token 写进通用 settings 字段。</p>
          </div>
        </div>

        <div class="mt-5 flex flex-wrap items-center justify-between gap-3 border-t border-dashed pt-4">
          <p class="text-xs text-muted-foreground">
            {{ activeTab === 'youtube' ? '下一步将接入 Google OAuth、频道选择和视频发布参数。' : '当前先完成平台域结构，后续按平台逐个接入账号绑定。' }}
          </p>
          <Button variant="outline" disabled :title="connectionButtonTitle">
            <Link2 class="size-4" />
            {{ activeTab === 'youtube' ? '绑定 YouTube 账号' : '绑定平台账号' }}
          </Button>
        </div>
      </section>
    </template>
  </div>
</template>

<script setup lang="ts">
import { computed, reactive, ref, watch } from 'vue'
import type { Component } from 'vue'
import {
  AtSign,
  BriefcaseBusiness,
  Clock3,
  Link2,
  LoaderCircle,
  MessageCircle,
  Music2,
  RefreshCw,
  Save,
  Send,
  Share2,
  ShieldCheck,
  Video,
} from '@lucide/vue'
import { toast } from 'vue-sonner'
import { useRouter } from 'vue-router'
import AdminFormField from '@/components/admin/AdminFormField.vue'
import AdminPageHeader from '@/components/admin/AdminPageHeader.vue'
import AdminStatusBadge from '@/components/admin/AdminStatusBadge.vue'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { useRouteTab } from '@/composables/useRouteTab'
import { useAuthStore } from '@/stores/auth'
import axios from '@/utils/axios'

type SocialTab = 'overview' | 'profiles' | 'youtube' | 'meta' | 'tiktok' | 'linkedin' | 'x' | 'wechat' | 'publications'
type SocialFieldKey = 'facebook' | 'twitter' | 'instagram' | 'linkedin' | 'youtube' | 'wechat'

const authStore = useAuthStore()
const router = useRouter()
const activeTab = useRouteTab<SocialTab>({
  defaultValue: 'overview',
  values: ['overview', 'profiles', 'youtube', 'meta', 'tiktok', 'linkedin', 'x', 'wechat', 'publications'],
  routes: {
    overview: 'SocialOverview',
    profiles: 'SocialProfiles',
    youtube: 'SocialYouTube',
    meta: 'SocialMeta',
    tiktok: 'SocialTikTok',
    linkedin: 'SocialLinkedIn',
    x: 'SocialX',
    wechat: 'SocialWeChat',
    publications: 'SocialPublications',
  },
})

const canEdit = computed(() => authStore.hasPermission('settings:edit'))
const loading = ref(false)
const saving = ref(false)

const socialFields: Array<{ key: SocialFieldKey; label: string; placeholder: string }> = [
  { key: 'facebook', label: 'Facebook', placeholder: 'Facebook 页面 URL' },
  { key: 'twitter', label: 'Twitter / X', placeholder: '账号 URL' },
  { key: 'instagram', label: 'Instagram', placeholder: '账号 URL' },
  { key: 'linkedin', label: 'LinkedIn', placeholder: '页面 URL' },
  { key: 'youtube', label: 'YouTube', placeholder: '频道 URL' },
  { key: 'wechat', label: '微信', placeholder: '二维码 URL' },
]

const socialSettings = reactive<Record<SocialFieldKey, string>>({
  facebook: '',
  twitter: '',
  instagram: '',
  linkedin: '',
  youtube: '',
  wechat: '',
})

const platformCards = [
  { key: 'youtube', label: 'YouTube', description: '绑定频道并发布视频', capability: '视频发布', status: '待绑定', tone: 'amber' as const, routeName: 'SocialYouTube', icon: Video },
  { key: 'meta', label: 'Facebook / Instagram', description: '管理 Meta 平台账号', capability: '图文 / 视频', status: '待绑定', tone: 'amber' as const, routeName: 'SocialMeta', icon: Share2 },
  { key: 'tiktok', label: 'TikTok', description: '准备短视频发布通道', capability: '短视频发布', status: '规划中', tone: 'gray' as const, routeName: 'SocialTikTok', icon: Music2 },
  { key: 'linkedin', label: 'LinkedIn', description: '管理品牌内容分发', capability: '内容发布', status: '规划中', tone: 'gray' as const, routeName: 'SocialLinkedIn', icon: BriefcaseBusiness },
  { key: 'x', label: 'X', description: '管理品牌内容分发', capability: '内容发布', status: '规划中', tone: 'gray' as const, routeName: 'SocialX', icon: AtSign },
  { key: 'wechat', label: '微信', description: '管理微信内容入口', capability: '内容发布', status: '规划中', tone: 'gray' as const, routeName: 'SocialWeChat', icon: MessageCircle },
]

const platformTabs: Record<Exclude<SocialTab, 'overview' | 'profiles' | 'publications'>, {
  label: string
  description: string
  capability: string
  icon: Component
}> = {
  youtube: { label: 'YouTube', description: '绑定 Google 账号和 YouTube Channel，为后续视频发布提供账号上下文。', capability: '视频发布', icon: Video },
  meta: { label: 'Facebook / Instagram', description: '统一维护 Meta 账号连接，并为 Facebook 页面与 Instagram 账号预留发布能力。', capability: '图文 / 视频', icon: Share2 },
  tiktok: { label: 'TikTok', description: '为短视频内容准备平台账号、发布参数和任务状态管理。', capability: '短视频发布', icon: Music2 },
  linkedin: { label: 'LinkedIn', description: '为品牌动态、视频和内容分发预留账号连接能力。', capability: '内容发布', icon: BriefcaseBusiness },
  x: { label: 'X', description: '为品牌内容、媒体资源和发布记录预留平台连接能力。', capability: '内容发布', icon: AtSign },
  wechat: { label: '微信', description: '为公众号、视频号或二维码展示配置预留独立管理入口。', capability: '内容发布', icon: MessageCircle },
}

const currentPlatform = computed(() => platformTabs[activeTab.value as keyof typeof platformTabs] || platformTabs.youtube)
const pageTitle = computed(() => {
  if (activeTab.value === 'overview') return '社交媒体'
  if (activeTab.value === 'profiles') return '前台展示'
  if (activeTab.value === 'publications') return '发布记录'
  return currentPlatform.value.label
})
const pageDescription = computed(() => {
  if (activeTab.value === 'overview') return '统一管理社交账号连接、前台链接和内容发布'
  if (activeTab.value === 'profiles') return '管理前台展示的官方账号与页面链接'
  if (activeTab.value === 'publications') return '查看平台发布任务、结果和失败记录'
  return currentPlatform.value.description
})
const connectionButtonTitle = '当前阶段先完成社交媒体域结构，平台 OAuth 接口将在对应平台接入时启用'

const applySetting = (key: string, value: unknown): void => {
  const normalizedKey = key.startsWith('social_') ? key.slice('social_'.length) : key
  if (normalizedKey in socialSettings) {
    socialSettings[normalizedKey as SocialFieldKey] = String(value ?? '')
  }
}

const fetchPublicLinks = async (): Promise<void> => {
  loading.value = true
  try {
    const response = await axios.get('/api/admin/settings/social', { params: { locale: 'en' } })
    const settings = Array.isArray(response.data?.settings) ? response.data.settings : []
    settings.forEach((setting: { key?: string; value?: unknown }) => {
      if (setting.key) applySetting(setting.key, setting.value)
    })
  } catch (error) {
    console.error('Failed to fetch social media links:', error)
    toast.error('社交媒体链接加载失败')
  } finally {
    loading.value = false
  }
}

const savePublicLinks = async (): Promise<void> => {
  if (!canEdit.value || saving.value) return
  saving.value = true
  try {
    const settings = socialFields.map((field) => ({
      key: field.key,
      value: socialSettings[field.key],
      type: 'string',
      group: 'social',
      locale: 'en',
      is_public: true,
      description: field.label,
    }))
    const response = await axios.post('/api/admin/settings/batch', { settings })
    toast.success(`已保存 ${response.data?.count ?? settings.length} 项社交媒体链接`)
  } catch (error) {
    console.error('Failed to save social media links:', error)
    toast.error('社交媒体链接保存失败')
  } finally {
    saving.value = false
  }
}

const refresh = async (): Promise<void> => {
  if (activeTab.value === 'profiles') {
    await fetchPublicLinks()
  }
}

const openPlatform = (routeName: string): void => {
  void router.push({ name: routeName })
}

watch(activeTab, (tab) => {
  if (tab === 'profiles') void fetchPublicLinks()
}, { immediate: true })
</script>
