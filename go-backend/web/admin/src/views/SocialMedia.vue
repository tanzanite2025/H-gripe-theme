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
        <Button v-if="activeTab === 'profiles'" variant="outline" :disabled="loading || saving || oauthLoading" @click="refresh">
 <RefreshCw :class="['size-4', loading || oauthLoading ? 'animate-spin': '']" />
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
            <AdminStatusBadge tone="green">OAuth 已接入</AdminStatusBadge>
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
            <p class="mt-2 text-xs leading-5 text-muted-foreground">账号、频道、权限范围和加密 Token 独立保存。</p>
          </div>
          <div class="rounded-xl border bg-background/70 p-3">
            <div class="flex items-center gap-2 text-xs font-black">
              <Send class="size-4 text-emerald-600" />
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
              这些 URL 写入 `settings.social`，只用于前台展示。YouTube 等平台的 OAuth 账号、频道和发布凭据不会写入这里。
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
          <AdminStatusBadge :tone="connectionTone(currentConnection)">{{ connectionStatusLabel(currentConnection) }}</AdminStatusBadge>
        </div>

        <div class="mt-5 grid gap-3 md:grid-cols-3">
          <div class="rounded-xl border bg-muted/20 p-4">
            <p class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/60">账号状态</p>
            <p class="mt-2 truncate text-sm font-black">{{ currentConnection.connected ? currentConnection.provider_account_name : '尚未绑定' }}</p>
            <p class="mt-1 break-words text-xs leading-5 text-muted-foreground">
              {{ currentConnection.connected ? (currentConnection.provider_account_email || currentConnection.provider_account_url || '账号已授权') : currentConnection.configured ? '点击授权按钮绑定官方账号。' : '后台尚未配置该平台 OAuth 应用。' }}
            </p>
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

        <div v-if="activeTab === 'meta' && currentConnection.connected" class="mt-4 grid gap-3 md:grid-cols-2">
          <div class="rounded-xl border bg-muted/20 p-4">
            <p class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/60">Facebook Page</p>
            <template v-if="metaResources.pages?.length">
              <a v-for="page in metaResources.pages" :key="page.id" :href="page.url" target="_blank" rel="noreferrer" class="mt-2 block text-sm font-black text-sky-700 hover:underline">
                {{ page.name || page.id }}
              </a>
            </template>
            <p v-else class="mt-2 text-xs text-muted-foreground">未发现可管理的 Facebook Page。</p>
          </div>
          <div class="rounded-xl border bg-muted/20 p-4">
            <p class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/60">Instagram</p>
            <template v-if="metaResources.instagram_accounts?.length">
              <a v-for="account in metaResources.instagram_accounts" :key="account.id" :href="account.url" target="_blank" rel="noreferrer" class="mt-2 block text-sm font-black text-pink-700 hover:underline">
                {{ account.username ? `@${account.username}` : account.name || account.id }}
              </a>
            </template>
            <p v-else class="mt-2 text-xs text-muted-foreground">未发现已关联的 Instagram Professional Account。</p>
          </div>
        </div>

        <div class="mt-5 flex flex-wrap items-center justify-between gap-3 border-t border-dashed pt-4">
          <p class="text-xs text-muted-foreground">
            {{ currentConnection.connected ? `已连接 ${currentConnection.provider_account_name || currentPlatform.label}` : currentConnection.configured ? '授权后会回到当前后台页面。' : '请先在后端配置该平台 OAuth 应用。' }}
          </p>
          <div class="flex items-center gap-2">
            <Button v-if="currentConnection.connected" variant="outline" :disabled="oauthActionLoading || !canEdit" @click="disconnectProvider">
              <LoaderCircle v-if="oauthActionLoading" class="size-4 animate-spin" />
              <Unplug v-else class="size-4" />
              解绑账号
            </Button>
            <Button v-else variant="outline" :disabled="oauthActionLoading || !canEdit || !currentConnection.configured" :title="currentConnection.configured ? '跳转官方授权页面' : '请先配置 OAuth 应用'" @click="startOAuth">
              <LoaderCircle v-if="oauthActionLoading" class="size-4 animate-spin" />
              <Link2 v-else class="size-4" />
              {{ currentPlatform.provider === 'youtube' ? '绑定 YouTube 账号' : `绑定 ${currentPlatform.label}` }}
            </Button>
          </div>
        </div>
      </section>
    </template>
  </div>
</template>

<script setup lang="ts">
import { computed, reactive, ref, watch } from 'vue'
import {
  Clock3,
  Link2,
  LoaderCircle,
  RefreshCw,
  Save,
  Send,
  ShieldCheck,
  Unplug,
} from '@lucide/vue'
import { toast } from 'vue-sonner'
import { useRoute, useRouter } from 'vue-router'
import AdminFormField from '@/components/admin/AdminFormField.vue'
import AdminPageHeader from '@/components/admin/AdminPageHeader.vue'
import AdminStatusBadge from '@/components/admin/AdminStatusBadge.vue'
import socialApi, {
  type SocialConnection,
  type SocialPublicLinkUpdate,
} from '@/api/social'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { useRouteTab } from '@/composables/useRouteTab'
import { useAuthStore } from '@/stores/auth'
import {
  socialFields,
  socialPlatforms,
  socialProviderList,
  type SocialFieldKey,
  type SocialProvider,
} from '@/utils/socialPlatforms'

type SocialTab = 'overview' | 'profiles' | 'youtube' | 'meta' | 'x' | 'reddit' | 'publications'

const makeEmptyConnection = (provider: SocialProvider): SocialConnection => ({
  provider,
  label: socialPlatforms[provider].label,
  configured: false,
  connected: false,
  status: 'disconnected',
  provider_resources: {},
})

const authStore = useAuthStore()
const route = useRoute()
const router = useRouter()
const activeTab = useRouteTab<SocialTab>({
  defaultValue: 'overview',
  values: ['overview', 'profiles', 'youtube', 'meta', 'x', 'reddit', 'publications'],
  routes: {
    overview: 'SocialOverview',
    profiles: 'SocialProfiles',
    youtube: 'SocialYouTube',
    meta: 'SocialMeta',
    x: 'SocialX',
    reddit: 'SocialReddit',
    publications: 'SocialPublications',
  },
})

const canEdit = computed(() => authStore.hasPermission('settings:edit'))
const loading = ref(false)
const saving = ref(false)
const oauthLoading = ref(false)
const oauthActionLoading = ref(false)
const oauthConnections = ref<SocialConnection[]>([])

const socialSettings = reactive<Record<SocialFieldKey, string>>({
  facebook: '',
  instagram: '',
  x: '',
  youtube: '',
  reddit: '',
})

const connectionByProvider = computed<Record<SocialProvider, SocialConnection>>(() => {
  const result = Object.fromEntries(
    socialProviderList.map((provider) => [provider, makeEmptyConnection(provider)])
  ) as Record<SocialProvider, SocialConnection>
  oauthConnections.value.forEach((connection) => {
    if (connection.provider in result) {
      result[connection.provider] = { ...result[connection.provider], ...connection }
    }
  })
  return result
})

const connectionStatusLabel = (connection: SocialConnection): string => {
  if (connection.connected) return '已连接'
  if (connection.status === 'error') return '连接异常'
  if (!connection.configured) return '未配置'
  return '待绑定'
}

const connectionTone = (connection: SocialConnection): 'green' | 'coral' | 'amber' | 'gray' => {
  if (connection.connected) return 'green'
  if (connection.status === 'error') return 'coral'
  if (!connection.configured) return 'gray'
  return 'amber'
}

const platformCards = computed(() => socialProviderList.map((provider) => {
  const platform = socialPlatforms[provider]
  const connection = connectionByProvider.value[provider]
  return {
    key: provider,
    label: platform.label,
    description: platform.overviewDescription,
    capability: platform.capability,
    status: connectionStatusLabel(connection),
    tone: connectionTone(connection),
    routeName: platform.routeName,
    icon: platform.icon,
  }
}))

const currentPlatform = computed(() => (
  socialPlatforms[activeTab.value as SocialProvider] || socialPlatforms.youtube
))
const currentConnection = computed(() => connectionByProvider.value[currentPlatform.value.provider])
const metaResources = computed(() => currentConnection.value.provider_resources || {})
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

const applySetting = (key: string, value: unknown): void => {
  const settingKey = key.startsWith('social_') ? key.slice('social_'.length) : key
  if (settingKey in socialSettings) {
    socialSettings[settingKey as SocialFieldKey] = String(value ?? '')
  }
}

const fetchPublicLinks = async (): Promise<void> => {
  loading.value = true
  try {
    const settings = await socialApi.listPublicLinks('en')
    settings.forEach((setting) => {
      applySetting(setting.key, setting.value)
    })
  } catch (error) {
    console.error('Failed to fetch social media links:', error)
    toast.error('社交媒体链接加载失败')
  } finally {
    loading.value = false
  }
}

const fetchOAuthConnections = async (): Promise<void> => {
  oauthLoading.value = true
  try {
    oauthConnections.value = await socialApi.listOAuthConnections()
  } catch (error) {
    console.error('Failed to fetch social OAuth connections:', error)
    toast.error('社交账号状态加载失败')
  } finally {
    oauthLoading.value = false
  }
}

const startOAuth = async (): Promise<void> => {
  if (!canEdit.value || oauthActionLoading.value || !currentConnection.value.configured) return
  oauthActionLoading.value = true
  try {
    const authorizationURL = await socialApi.startOAuth(currentPlatform.value.provider, route.fullPath)
    window.location.assign(authorizationURL)
  } catch (error) {
    console.error('Failed to start social OAuth:', error)
    toast.error('无法开始平台授权')
    oauthActionLoading.value = false
  }
}

const disconnectProvider = async (): Promise<void> => {
  if (!canEdit.value || oauthActionLoading.value) return
  if (!window.confirm(`确定解绑 ${currentPlatform.value.label} 账号吗？`)) return
  oauthActionLoading.value = true
  try {
    await socialApi.disconnect(currentPlatform.value.provider)
    toast.success(`${currentPlatform.value.label} 已解绑`)
    await fetchOAuthConnections()
  } catch (error) {
    console.error('Failed to disconnect social OAuth:', error)
    toast.error('社交账号解绑失败')
  } finally {
    oauthActionLoading.value = false
  }
}

const handleOAuthCallbackMessage = async (): Promise<void> => {
  const status = String(route.query.social_oauth_status || '')
  if (!status) return
  const message = String(route.query.social_oauth_message || '')
  if (status === 'connected') {
    toast.success(message || '社交账号已连接')
  } else if (status === 'error') {
    toast.error(message || '社交账号授权失败')
  }
  await fetchOAuthConnections()
  const query = { ...route.query }
  delete query.social_oauth_status
  delete query.social_oauth_provider
  delete query.social_oauth_message
  await router.replace({ query })
}

const savePublicLinks = async (): Promise<void> => {
  if (!canEdit.value || saving.value) return
  saving.value = true
  try {
    const settings: SocialPublicLinkUpdate[] = socialFields.map((field) => ({
      key: field.key,
      value: socialSettings[field.key],
      type: 'string',
      group: 'social',
      locale: 'en',
      is_public: true,
      description: field.label,
    }))
    const count = await socialApi.savePublicLinks(settings)
    toast.success(`已保存 ${count} 项社交媒体链接`)
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
  await fetchOAuthConnections()
}

const openPlatform = (routeName: string): void => {
  void router.push({ name: routeName })
}

watch(activeTab, (tab) => {
  if (tab === 'profiles') void fetchPublicLinks()
  void fetchOAuthConnections()
}, { immediate: true })

watch(() => route.query.social_oauth_status, () => {
  void handleOAuthCallbackMessage()
}, { immediate: true })
</script>
