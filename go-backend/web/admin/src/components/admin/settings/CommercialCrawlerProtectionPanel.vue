<template>
  <div class="space-y-6">
    <section class="rounded-2xl border bg-muted/30 p-4">
      <div class="flex flex-wrap items-start justify-between gap-3">
        <div>
          <p class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/60">Commercial Intelligence Crawler Protection</p>
          <h2 class="mt-1 text-sm font-black text-foreground">商业爬虫防护</h2>
        </div>
        <div class="flex items-center gap-2">
          <span
            class="inline-flex items-center gap-1.5 rounded-full border px-2.5 py-1 text-xs font-black"
            :class="enabled ? 'border-emerald-500/20 bg-emerald-500/10 text-emerald-700 dark:text-emerald-200' : 'border-rose-500/20 bg-rose-500/10 text-rose-700 dark:text-rose-200'"
          >
            <ShieldCheck v-if="enabled" class="size-3.5" />
            <ShieldAlert v-else class="size-3.5" />
            {{ enabled ? '已启用' : '未启用' }}
          </span>
          <Button
            type="button"
            variant="outline"
            size="sm"
            :disabled="loading"
            @click="emit('refresh')"
          >
            <RefreshCw :class="['size-3.5', loading ? 'animate-spin' : '']" />
            刷新
          </Button>
        </div>
      </div>

      <div class="mt-4 grid gap-3 md:grid-cols-3">
        <div
          v-for="layer in enforcement"
          :key="layer.layer"
          class="rounded-xl border bg-background/70 p-3"
        >
          <p class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/60">{{ layer.layer }}</p>
          <p class="mt-1 text-sm font-black text-foreground">{{ layer.status === 'enabled' ? '生效' : layer.status }}</p>
        </div>
      </div>

      <div class="mt-4 rounded-xl border border-amber-500/20 bg-amber-500/10 px-3 py-2.5 text-xs text-amber-800 dark:text-amber-100">
        命中规则的请求返回 HTTP {{ responseStatus }}，前台页面在公网 Nginx 拦截，API 直连也由 Go 中间件拦截。
      </div>
    </section>

    <section class="rounded-2xl border bg-muted/30 p-4">
      <div class="flex items-center justify-between gap-3">
        <div>
          <p class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/60">Blocked User Agents</p>
          <h3 class="mt-1 text-sm font-black text-foreground">已封禁规则</h3>
        </div>
        <span class="text-xs font-bold text-muted-foreground">{{ rules.length }} 条</span>
      </div>

      <div v-if="loading && rules.length === 0" class="flex h-32 items-center justify-center text-sm text-muted-foreground">
        <RefreshCw class="mr-2 size-4 animate-spin" />
        正在读取防护状态
      </div>

      <div v-else class="mt-4 grid gap-3 md:grid-cols-2 xl:grid-cols-4">
        <div
          v-for="rule in rules"
          :key="rule.user_agent"
          class="rounded-xl border bg-background/70 p-3"
        >
          <p class="text-sm font-black text-foreground">{{ rule.provider }}</p>
          <code class="mt-2 block break-all text-xs font-bold text-admin-selected">{{ rule.user_agent }}</code>
        </div>
      </div>
    </section>

    <section class="border-t pt-5">
      <div class="flex flex-wrap items-end justify-between gap-3">
        <div>
          <p class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/60">Built-In Intelligence Seeds</p>
          <h3 class="mt-1 text-sm font-black text-foreground">内置商业情报种子</h3>
          <p class="mt-1 text-xs text-muted-foreground">种子为行为检测提供统一识别依据；标记为“已启用”或“已拦截”的规则正在执行。</p>
        </div>
        <span class="text-xs font-bold text-muted-foreground">{{ intelligenceSeeds.length }} 条</span>
      </div>

      <div v-if="loading && intelligenceSeeds.length === 0" class="flex h-32 items-center justify-center text-sm text-muted-foreground">
        <RefreshCw class="mr-2 size-4 animate-spin" />
        正在读取内置情报种子
      </div>

      <div v-else class="mt-4 grid gap-3 lg:grid-cols-2">
        <article
          v-for="seed in intelligenceSeeds"
          :key="seed.id"
          class="rounded-lg border bg-muted/20 p-4"
        >
          <div class="flex flex-wrap items-start justify-between gap-2">
            <div>
              <p class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/60">{{ seedCategoryLabel(seed.category) }}</p>
              <h4 class="mt-1 text-sm font-black text-foreground">{{ seed.name }}</h4>
            </div>
            <span
              class="inline-flex shrink-0 rounded-full border px-2 py-1 text-[11px] font-black"
              :class="seed.enforcement === 'enforced'
                ? 'border-emerald-500/20 bg-emerald-500/10 text-emerald-700 dark:text-emerald-200'
                : 'border-sky-500/20 bg-sky-500/10 text-sky-700 dark:text-sky-200'"
            >
              {{ seedEnforcementLabel(seed.enforcement, seed.action) }}
            </span>
          </div>

          <p class="mt-3 text-xs leading-5 text-muted-foreground">{{ seed.identification }}</p>

          <div v-if="seed.aliases?.length" class="mt-3">
            <p class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/60">Aliases</p>
            <p class="mt-1 text-xs font-semibold text-foreground">{{ seed.aliases.join(' · ') }}</p>
          </div>

          <div v-if="seed.detection_signals?.length" class="mt-3">
            <p class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/60">识别信号</p>
            <ul class="mt-1 space-y-1 text-xs leading-5 text-muted-foreground">
              <li v-for="signal in seed.detection_signals" :key="signal">{{ signal }}</li>
            </ul>
          </div>

          <p v-if="seed.threshold" class="mt-3 text-xs leading-5 text-muted-foreground">
            <span class="font-black text-foreground">触发阈值：</span>{{ seed.threshold }}
          </p>
          <p class="mt-3 text-xs font-bold text-admin-selected">{{ seedActionLabel(seed.action) }}</p>
        </article>
      </div>
    </section>

    <section v-if="orderNumberProtection" class="border-t pt-5">
      <div class="flex flex-wrap items-start justify-between gap-3">
        <div>
          <p class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/60">Opaque Order Numbers</p>
          <h3 class="mt-1 text-sm font-black text-foreground">订单号隐私防护</h3>
          <p class="mt-1 text-xs text-muted-foreground">客户侧、支付网关和客服订单卡仅使用公开订单号；数据库自增 ID 保留在受控后台边界内。</p>
        </div>
        <span
          class="inline-flex shrink-0 items-center gap-1.5 rounded-full border px-2.5 py-1 text-xs font-black"
          :class="orderNumberProtection.configured
            ? 'border-emerald-500/20 bg-emerald-500/10 text-emerald-700 dark:text-emerald-200'
            : 'border-rose-500/20 bg-rose-500/10 text-rose-700 dark:text-rose-200'"
        >
          <ShieldCheck v-if="orderNumberProtection.configured" class="size-3.5" />
          <ShieldAlert v-else class="size-3.5" />
          {{ orderNumberProtection.configured ? '已启用' : '未配置' }}
        </span>
      </div>

      <div class="mt-4 grid gap-3 lg:grid-cols-2 xl:grid-cols-4">
        <div class="rounded-lg border bg-muted/20 p-4">
          <p class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/60">Public Format</p>
          <code class="mt-2 block break-all text-xs font-bold text-admin-selected">{{ orderNumberProtection.format }}</code>
        </div>
        <div class="rounded-lg border bg-muted/20 p-4">
          <p class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/60">Verification</p>
          <p class="mt-2 text-xs leading-5 text-muted-foreground">{{ orderNumberProtection.verification }}</p>
        </div>
        <div class="rounded-lg border bg-muted/20 p-4">
          <p class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/60">Boundary</p>
          <p class="mt-2 text-xs leading-5 text-muted-foreground">公开 API 不返回订单数据库 ID；后台管理功能仍在鉴权后使用内部主键。</p>
        </div>
        <div class="rounded-lg border bg-muted/20 p-4">
          <p class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/60">Key Rotation</p>
          <p class="mt-2 text-xs leading-5 text-muted-foreground">{{ orderNumberProtection.key_rotation }}</p>
        </div>
      </div>

      <div class="mt-4 grid gap-3 lg:grid-cols-2">
        <div class="rounded-lg border bg-background/70 p-4">
          <p class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/60">Public Policy</p>
          <ul class="mt-2 space-y-1.5 text-xs leading-5 text-muted-foreground">
            <li v-for="item in orderNumberProtection.public_id_policy" :key="item">{{ item }}</li>
          </ul>
        </div>
        <div class="rounded-lg border bg-background/70 p-4">
          <p class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/60">Internal Boundary</p>
          <ul class="mt-2 space-y-1.5 text-xs leading-5 text-muted-foreground">
            <li v-for="item in orderNumberProtection.internal_id_policy" :key="item">{{ item }}</li>
          </ul>
        </div>
      </div>
    </section>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { RefreshCw, ShieldAlert, ShieldCheck } from '@lucide/vue'
import { Button } from '@/components/ui/button'
import type {
  CommercialCrawlerEnforcementLayer,
  CommercialCrawlerIntelligenceSeed,
  CommercialCrawlerProtection,
  CommercialCrawlerRule,
  OrderNumberProtection,
} from './settingsTypes'

const props = withDefaults(defineProps<{
  protection?: CommercialCrawlerProtection | null
  loading?: boolean
}>(), {
  protection: null,
  loading: false,
})

const emit = defineEmits<{
  (event: 'refresh'): void
}>()

const enabled = computed(() => props.protection?.enabled === true)
const responseStatus = computed(() => Number(props.protection?.response_status) || 403)
const rules = computed<CommercialCrawlerRule[]>(() => Array.isArray(props.protection?.rules) ? props.protection.rules : [])
const enforcement = computed<CommercialCrawlerEnforcementLayer[]>(() => Array.isArray(props.protection?.enforcement) ? props.protection.enforcement : [])
const intelligenceSeeds = computed<CommercialCrawlerIntelligenceSeed[]>(() => Array.isArray(props.protection?.intelligence_seeds) ? props.protection.intelligence_seeds : [])
const orderNumberProtection = computed<OrderNumberProtection | null>(() => props.protection?.order_number_protection || null)

const seedCategoryLabel = (category?: string): string => ({
  known_crawler: '已知商业爬虫',
  browser_extension: '浏览器插件',
  inventory_probe: '库存探测行为',
  order_enumeration: '订单枚举行为',
} as Record<string, string>)[category || ''] || '商业情报种子'

const seedEnforcementLabel = (enforcementMode?: string, action?: string): string => {
  if (enforcementMode === 'enforced' && action === 'block_403') return '已拦截'
  if (enforcementMode === 'enforced' && action === 'not_found_404') return '探测入口已封锁'
  if (enforcementMode === 'enforced') return '行为保护已启用'
  if (enforcementMode === 'seed_only') return '仅作情报种子'
  return enforcementMode
}

const seedActionLabel = (action?: string): string => ({
  block_403: '当前动作：命中后返回 HTTP 403',
  rate_limit_429: '当前动作：命中后返回 HTTP 429，并要求稍后重试',
  not_found_404: '当前动作：平台指纹探测路径返回 HTTP 404',
  behavior_detection_pending: '当前动作：待接入行为检测与分级限速',
} as Record<string, string>)[action || ''] || `当前动作：${action || '-'}`
</script>
