<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { Languages, RefreshCw, ShieldCheck, Type } from '@lucide/vue'
import { toast } from 'vue-sonner'
import preflightApi, {
  type FontPreflightCheck,
  type FontPreflightFace,
  type FontPreflightLocaleCoverage,
  type FontPreflightReport,
  type FontPreflightStatus,
} from '@/api/preflight'
import AdminPageHeader from '@/components/admin/AdminPageHeader.vue'
import AdminStatusBadge, { type AdminStatusTone } from '@/components/admin/AdminStatusBadge.vue'
import { Button } from '@/components/ui/button'

const loading = ref(false)
const report = ref<FontPreflightReport | null>(null)
const loadError = ref('')

const summaryItems = computed(() => {
  if (!report.value) return []
  return [
    { label: '上线门禁', value: statusLabel(report.value.overall_status), tone: statusTextTone(report.value.overall_status) },
    { label: '语言', value: report.value.coverage.locale_count, tone: '' },
    { label: '受控字体', value: report.value.faces.length, tone: '' },
    { label: '缺失字形', value: report.value.coverage.missing_characters, tone: report.value.coverage.missing_characters ? 'text-rose-600' : 'text-emerald-600' },
  ]
})

const blockingChecks = computed(() => report.value?.checks.filter((check) => check.status === 'block').length || 0)

const loadReport = async (): Promise<void> => {
  loading.value = true
  loadError.value = ''
  try {
    report.value = await preflightApi.getFontPreflight()
  } catch (error: any) {
    report.value = null
    loadError.value = error?.response?.data?.error || '已部署店铺没有可用的字体预检清单'
    toast.error(loadError.value)
  } finally {
    loading.value = false
  }
}

const statusTone = (status: FontPreflightStatus): AdminStatusTone => ({
  pass: 'green',
  warning: 'amber',
  block: 'coral',
} as const)[status]

const statusTextTone = (status: FontPreflightStatus): string => ({
  pass: 'text-emerald-600',
  warning: 'text-amber-600',
  block: 'text-rose-600',
}[status])

const statusLabel = (status: FontPreflightStatus): string => ({
  pass: '通过',
  warning: '注意',
  block: '阻断',
}[status])

const faceStatus = (face: FontPreflightFace): FontPreflightStatus => (
  face.self_hosted && face.font_display === 'block' ? 'pass' : 'block'
)

const localeStatus = (locale: FontPreflightLocaleCoverage): FontPreflightStatus => locale.status

const formatBytes = (value: number): string => {
  if (value < 1024) return `${value} B`
  if (value < 1024 * 1024) return `${(value / 1024).toFixed(1)} KiB`
  return `${(value / (1024 * 1024)).toFixed(2)} MiB`
}

const formatDate = (value: string): string => new Date(value).toLocaleString('zh-CN')

const detailText = (check: FontPreflightCheck): string => check.details.join(' ')

onMounted(() => {
  void loadReport()
})
</script>

<template>
  <div class="space-y-4">
    <AdminPageHeader title="上线前检查 / 字体" description="部署字体合同">
      <template #actions>
        <Button
          size="icon"
          variant="outline"
          title="刷新字体预检"
          :disabled="loading"
          @click="loadReport"
        >
          <RefreshCw :class="['size-4', { 'animate-spin': loading }]" />
        </Button>
      </template>
    </AdminPageHeader>

    <section v-if="loading && !report" class="border bg-card px-4 py-12 text-center text-sm text-muted-foreground">
      正在读取已部署店铺的字体预检清单
    </section>

    <section v-else-if="loadError" class="border border-rose-500/30 bg-card px-4 py-5">
      <div class="flex items-start gap-3">
        <ShieldCheck class="mt-0.5 size-5 shrink-0 text-rose-600" />
        <div>
          <h2 class="text-sm font-black text-rose-700">字体预检不可用</h2>
          <p class="mt-1 text-sm text-muted-foreground">{{ loadError }}</p>
        </div>
      </div>
    </section>

    <template v-else-if="report">
      <section class="grid gap-3 sm:grid-cols-2 xl:grid-cols-4">
        <div v-for="item in summaryItems" :key="item.label" class="border bg-card px-4 py-3">
          <p class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/70">{{ item.label }}</p>
          <p class="mt-2 text-2xl font-black" :class="item.tone">{{ item.value }}</p>
        </div>
      </section>

      <section class="border bg-card">
        <div class="flex flex-wrap items-start justify-between gap-4 px-4 py-4">
          <div class="flex min-w-0 gap-3">
            <ShieldCheck class="mt-0.5 size-5 shrink-0" :class="statusTextTone(report.overall_status)" />
            <div>
              <div class="flex flex-wrap items-center gap-2">
                <h2 class="text-sm font-black">自托管字体，禁止非预期回退</h2>
                <AdminStatusBadge :tone="statusTone(report.overall_status)">{{ statusLabel(report.overall_status) }}</AdminStatusBadge>
              </div>
              <p class="mt-1 text-xs text-muted-foreground">{{ report.project }} · 构建于 {{ formatDate(report.generated_at) }}</p>
            </div>
          </div>
          <p class="font-mono text-xs font-bold text-muted-foreground">schema v{{ report.schema_version }}</p>
        </div>
        <div class="grid divide-y border-t sm:grid-cols-3 sm:divide-x sm:divide-y-0">
          <div class="px-4 py-3">
            <p class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/70">字体加载</p>
            <p class="mt-1 font-mono text-sm font-black">{{ report.baseline.font_display }}</p>
          </div>
          <div class="px-4 py-3">
            <p class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/70">阻断项</p>
            <p class="mt-1 font-mono text-sm font-black" :class="blockingChecks ? 'text-rose-600' : 'text-emerald-600'">{{ blockingChecks }}</p>
          </div>
          <div class="px-4 py-3">
            <p class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/70">策略</p>
            <p class="mt-1 text-sm font-black">{{ report.strategy.label }}</p>
          </div>
        </div>
        <ul class="divide-y border-t">
          <li v-for="rule in report.baseline.rules" :key="rule" class="px-4 py-2.5 text-xs text-muted-foreground">{{ rule }}</li>
        </ul>
      </section>

      <section class="border bg-card">
        <div class="flex flex-wrap items-start justify-between gap-3 border-b px-4 py-3">
          <div class="flex gap-3">
            <Type class="mt-0.5 size-4 shrink-0 text-muted-foreground" />
            <div>
              <h2 class="text-sm font-black">分片策略与布局一致性</h2>
              <p class="mt-1 text-xs text-muted-foreground">{{ report.strategy.rationale }}</p>
            </div>
          </div>
          <AdminStatusBadge :tone="statusTone(report.strategy.status)">{{ statusLabel(report.strategy.status) }}</AdminStatusBadge>
        </div>
        <div class="grid divide-y sm:grid-cols-3 sm:divide-x sm:divide-y-0">
          <div class="px-4 py-3">
            <p class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/70">默认字体栈</p>
            <p class="mt-1 break-words font-mono text-xs font-bold">{{ report.strategy.default_stack.join(', ') }}</p>
          </div>
          <div class="px-4 py-3">
            <p class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/70">Latin 子集</p>
            <p class="mt-1 font-mono text-xs font-bold">{{ formatBytes(report.strategy.latin_bytes) }} / {{ formatBytes(report.strategy.latin_budget_bytes) }}</p>
          </div>
          <div class="px-4 py-3">
            <p class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/70">切换布局</p>
            <p class="mt-1 text-xs font-black" :class="report.strategy.layout_parity_verified ? 'text-emerald-600' : 'text-rose-600'">
              {{ report.strategy.layout_parity_verified ? '已验证同度量与同轮廓' : '未通过一致性验证' }}
            </p>
          </div>
        </div>
      </section>

      <section class="overflow-hidden border bg-card">
        <div class="flex flex-wrap items-center justify-between gap-3 border-b px-4 py-3">
          <div>
            <h2 class="text-sm font-black">字体资源</h2>
            <p class="mt-1 text-xs text-muted-foreground">仅列出当前店铺构建中声明并验证的自托管字体。</p>
          </div>
          <span class="font-mono text-xs font-bold text-muted-foreground">{{ report.faces.length }} 个字体面</span>
        </div>
        <div class="overflow-x-auto">
          <table class="w-full min-w-[880px] text-sm">
            <thead class="border-b bg-muted/30 text-left text-[10px] font-black uppercase tracking-widest text-muted-foreground/70">
              <tr>
                <th class="px-4 py-3">字体族</th>
                <th class="px-4 py-3">脚本 / 用途</th>
                <th class="px-4 py-3">文件</th>
                <th class="px-4 py-3">大小</th>
                <th class="px-4 py-3">加载</th>
                <th class="px-4 py-3">状态</th>
              </tr>
            </thead>
            <tbody class="divide-y">
              <tr v-for="face in report.faces" :key="face.family">
                <td class="px-4 py-3 font-mono text-xs font-bold">{{ face.family }}</td>
                <td class="px-4 py-3">
                  <p class="font-bold">{{ face.script }}</p>
                  <p class="mt-1 text-xs text-muted-foreground">{{ face.role }}</p>
                </td>
                <td class="px-4 py-3 font-mono text-[11px] text-muted-foreground">{{ face.filename }}</td>
                <td class="px-4 py-3 font-mono text-xs">{{ formatBytes(face.bytes) }}</td>
                <td class="px-4 py-3 font-mono text-xs">{{ face.font_display }}</td>
                <td class="px-4 py-3"><AdminStatusBadge :tone="statusTone(faceStatus(face))">{{ statusLabel(faceStatus(face)) }}</AdminStatusBadge></td>
              </tr>
            </tbody>
          </table>
        </div>
      </section>

      <section class="border bg-card">
        <div class="flex flex-wrap items-start justify-between gap-3 border-b px-4 py-3">
          <div class="flex gap-3">
            <Languages class="mt-0.5 size-4 shrink-0 text-muted-foreground" />
            <div>
              <h2 class="text-sm font-black">语言资源覆盖</h2>
              <p class="mt-1 text-xs text-muted-foreground">{{ report.coverage.source_file_count }} 个语言资源文件，共检查 {{ report.coverage.checked_characters }} 个去重字符。</p>
            </div>
          </div>
          <AdminStatusBadge :tone="report.coverage.missing_characters ? 'coral' : 'green'">
            {{ report.coverage.missing_characters ? `${report.coverage.missing_characters} 缺失` : '覆盖完整' }}
          </AdminStatusBadge>
        </div>
        <div class="overflow-x-auto">
          <table class="w-full min-w-[820px] text-sm">
            <thead class="border-b bg-muted/30 text-left text-[10px] font-black uppercase tracking-widest text-muted-foreground/70">
              <tr>
                <th class="px-4 py-3">语言</th>
                <th class="px-4 py-3">资源文件</th>
                <th class="px-4 py-3">去重字符</th>
                <th class="px-4 py-3">字体栈</th>
                <th class="px-4 py-3">缺失</th>
                <th class="px-4 py-3">状态</th>
              </tr>
            </thead>
            <tbody class="divide-y">
              <tr v-for="locale in report.coverage.locales" :key="locale.locale">
                <td class="px-4 py-3 font-mono text-xs font-bold">{{ locale.locale }}</td>
                <td class="px-4 py-3 font-mono text-xs">{{ locale.source_files }}</td>
                <td class="px-4 py-3 font-mono text-xs">{{ locale.checked_characters }}</td>
                <td class="max-w-[22rem] px-4 py-3 font-mono text-[11px] text-muted-foreground">{{ locale.font_stack.join(', ') }}</td>
                <td class="px-4 py-3 font-mono text-xs" :class="locale.missing_characters ? 'text-rose-600' : 'text-emerald-600'">
                  {{ locale.missing_characters }}
                  <span v-if="locale.missing_sample.length" class="block text-[10px]">{{ locale.missing_sample.join(', ') }}</span>
                </td>
                <td class="px-4 py-3"><AdminStatusBadge :tone="statusTone(localeStatus(locale))">{{ statusLabel(localeStatus(locale)) }}</AdminStatusBadge></td>
              </tr>
            </tbody>
          </table>
        </div>
      </section>

      <section class="border bg-card">
        <div class="border-b px-4 py-3">
          <h2 class="text-sm font-black">检查结果</h2>
        </div>
        <div class="divide-y">
          <div v-for="check in report.checks" :key="check.key" class="flex flex-wrap items-start justify-between gap-3 px-4 py-3">
            <div class="min-w-0">
              <div class="flex flex-wrap items-center gap-2">
                <p class="font-bold">{{ check.label }}</p>
                <span class="font-mono text-[10px] text-muted-foreground">{{ check.key }}</span>
              </div>
              <p class="mt-1 text-xs text-muted-foreground">{{ check.message }}</p>
              <p v-if="check.details.length" class="mt-2 text-xs text-rose-600">{{ detailText(check) }}</p>
            </div>
            <AdminStatusBadge :tone="statusTone(check.status)">{{ statusLabel(check.status) }}</AdminStatusBadge>
          </div>
        </div>
      </section>
    </template>
  </div>
</template>
