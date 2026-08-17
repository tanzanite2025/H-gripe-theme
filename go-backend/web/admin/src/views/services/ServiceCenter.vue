<template>
 <div class="space-y-4">
    <AdminPageHeader title="服务中心" description="平台授权与资源绑定">
      <template #actions>
        <Button size="icon" variant="outline" title="刷新服务状态" :disabled="loading" @click="loadOverview">
 <RefreshCw :class="['size-4', loading ? 'animate-spin': '']" />
        </Button>
      </template>
    </AdminPageHeader>

 <section class="grid gap-3 md:grid-cols-2 xl:grid-cols-3">
      <button
        v-for="provider in providers"
        :key="provider.id"
        type="button"
 class="group min-h-44 border bg-card p-4 text-left transition-colors"
 :class="provider.route ? 'hover:border-primary/60 hover:bg-muted/30': 'cursor-default opacity-80'"
        :disabled="!provider.route"
        @click="openProvider(provider)"
      >
 <div class="flex items-start justify-between gap-3">
 <div class="flex min-w-0 items-center gap-3">
 <div class="flex size-10 shrink-0 items-center justify-center border bg-muted/40">
 <component :is="providerIcon(provider.id)" class="size-5" />
            </div>
 <div class="min-w-0">
 <p class="truncate text-sm font-black">{{ provider.label }}</p>
 <p class="mt-1 text-[10px] font-bold uppercase tracking-wider text-muted-foreground">
                {{ statusLabel(provider.status) }}
              </p>
            </div>
          </div>
 <ChevronRight v-if="provider.route" class="mt-1 size-4 text-muted-foreground transition-transform group-hover:translate-x-0.5" />
        </div>

 <div class="mt-7 grid grid-cols-3 gap-3 border-t border-dashed pt-3">
          <div>
 <p class="text-[9px] font-black uppercase tracking-widest text-muted-foreground/70">连接</p>
 <p class="mt-1 text-lg font-black">{{ provider.connection_count }}</p>
          </div>
          <div>
 <p class="text-[9px] font-black uppercase tracking-widest text-muted-foreground/70">可用</p>
 <p class="mt-1 text-lg font-black text-emerald-600">{{ provider.active_connection_count }}</p>
          </div>
          <div>
 <p class="text-[9px] font-black uppercase tracking-widest text-muted-foreground/70">资源</p>
 <p class="mt-1 text-lg font-black">{{ provider.resource_count }}</p>
          </div>
        </div>
      </button>
    </section>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { ChevronRight, Cloud, RefreshCw, Server } from '@lucide/vue'
import { toast } from 'vue-sonner'
import AdminPageHeader from '@/components/admin/AdminPageHeader.vue'
import { Button } from '@/components/ui/button'
import servicesApi, { type ServiceCenterProvider } from '@/api/services'

const router = useRouter()
const loading = ref(false)
const providers = ref<ServiceCenterProvider[]>([])

const loadOverview = async (): Promise<void> => {
  loading.value = true
  try {
    const data = await servicesApi.getOverview()
    providers.value = Array.isArray(data.providers) ? data.providers : []
  } catch (error: any) {
    toast.error(error?.response?.data?.message || error?.response?.data?.error || '服务中心加载失败')
  } finally {
    loading.value = false
  }
}

const openProvider = (provider: ServiceCenterProvider): void => {
  if (provider.route) void router.push(provider.route)
}

const providerIcon = (provider: string) => ({
  cloudflare: Cloud,
  hostinger: Server,
}[provider] || Cloud)

const statusLabel = (status: ServiceCenterProvider['status']): string => ({
  active: '已连接',
  attention: '需要处理',
  pending: '待验证',
  not_connected: '未连接',
  not_configured: '未接入',
}[status] || status)

onMounted(() => {
  void loadOverview()
})
</script>
