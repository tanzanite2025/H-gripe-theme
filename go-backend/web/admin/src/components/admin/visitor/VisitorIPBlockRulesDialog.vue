<template>
  <Dialog :open="open" @update:open="emit('update:open', $event)">
    <DialogContent size="lg" class="max-h-[90dvh] overflow-y-auto" @open-auto-focus.prevent>
      <DialogHeader>
        <DialogTitle class="flex items-center gap-2">
          <ShieldAlert class="size-4 text-rose-500" />
          全局 IP 封禁规则
        </DialogTitle>
        <DialogDescription>
          规则会在所有 API 路由统一生效，访客画像、商业爬虫和后续风控自动化共用这套拦截引擎。
          如果多个用户共享出口 IP，封禁也会影响这些用户。
        </DialogDescription>
      </DialogHeader>

      <form v-if="canCreate" class="space-y-4 rounded-lg border border-rose-500/20 bg-rose-500/5 p-4" @submit.prevent="submit">
        <div class="flex items-start gap-2 rounded-lg border border-amber-500/25 bg-amber-500/10 px-3 py-2 text-xs leading-5 text-amber-800 dark:text-amber-200">
          <ShieldAlert class="mt-0.5 size-4 shrink-0" />
          <p>这是全局 IP/CIDR 封禁，不是单个访客封禁；共享网络、公司出口或 NAT 下的其他用户也可能无法访问 API。</p>
        </div>
        <div class="grid gap-3 md:grid-cols-[minmax(0,1fr)_minmax(0,1fr)]">
          <AdminFormField label="IP / CIDR" required description="例如 203.0.113.10 或 203.0.113.0/24">
            <Input
              v-model="form.cidr"
              :disabled="saving"
              autocomplete="off"
              placeholder="203.0.113.10 或 203.0.113.0/24"
            />
          </AdminFormField>
          <AdminFormField label="到期时间" description="留空表示持续有效">
            <Input v-model="form.expiresAt" :disabled="saving" type="datetime-local" />
          </AdminFormField>
        </div>

        <AdminFormField label="封禁原因" required description="规则会保留原因，审计日志只记录是否填写和长度">
          <Textarea
            v-model="form.reason"
            :disabled="saving"
            class="min-h-20"
            maxlength="500"
            placeholder="例如：同一 IP 连续探测库存接口。"
          />
        </AdminFormField>

        <div class="flex justify-end">
          <Button type="submit" variant="destructive" :disabled="saving">
            <LoaderCircle v-if="saving" class="size-4 animate-spin" />
            <ShieldAlert v-else class="size-4" />
            {{ saving ? '封禁中' : '新增封禁规则' }}
          </Button>
        </div>
      </form>

      <div v-else class="rounded-lg border border-amber-500/25 bg-amber-500/5 px-3 py-2 text-xs text-amber-700 dark:text-amber-300">
        当前账号只有规则查看权限，新增或解除封禁需要管理员角色。
      </div>

      <p v-if="validationError || error" class="rounded-lg border border-rose-500/25 bg-rose-500/5 px-3 py-2 text-xs font-bold text-rose-700 dark:text-rose-300">
        {{ validationError || error }}
      </p>

      <section class="min-h-0 space-y-2">
        <div class="flex items-center justify-between gap-3">
          <h3 class="text-xs font-black uppercase tracking-wider">规则清单</h3>
          <span class="font-mono text-[10px] text-muted-foreground">{{ pagination.total }} 条</span>
        </div>

        <div v-if="loading" class="flex min-h-32 items-center justify-center text-xs text-muted-foreground">
          <LoaderCircle class="mr-2 size-4 animate-spin" />
          正在读取规则
        </div>

        <div v-else-if="rules.length === 0" class="flex min-h-32 items-center justify-center rounded-lg border border-dashed text-xs text-muted-foreground">
          暂无全局 IP 封禁规则
        </div>

        <div v-else class="max-h-[36dvh] space-y-2 overflow-y-auto pr-1">
          <article
            v-for="rule in rules"
            :key="rule.id"
            class="rounded-lg border p-3"
            :class="rule.status === 'active' ? 'border-rose-500/20 bg-rose-500/5' : 'border-border/70 bg-muted/20'"
          >
            <div class="flex items-start justify-between gap-3">
              <div class="min-w-0">
                <div class="flex flex-wrap items-center gap-1.5">
                  <span class="break-all font-mono text-sm font-black">{{ rule.cidr || '-' }}</span>
                  <AdminStatusBadge :tone="statusTone(rule.status)">
                    {{ statusLabel(rule.status) }}
                  </AdminStatusBadge>
                </div>
                <p class="mt-1 text-[10px] text-muted-foreground">
                  来源：{{ sourceLabel(rule.source) }}
                  <span v-if="rule.source_reference"> · {{ rule.source_reference }}</span>
                </p>
                <p class="mt-2 break-words text-xs text-foreground">{{ rule.reason || '未填写原因' }}</p>
                <p class="mt-1 text-[10px] text-muted-foreground">
                  <template v-if="rule.status === 'disabled'">
                    {{ rule.disabled_at ? `解除：${formatDate(rule.disabled_at)}` : '已解除' }}
                  </template>
                  <template v-else-if="rule.expires_at">
                    到期：{{ formatDate(rule.expires_at) }}
                  </template>
                  <template v-else>
                    持续有效
                  </template>
                </p>
              </div>

              <Button
                v-if="canManage && rule.status === 'active'"
                type="button"
                variant="outline"
                size="xs"
                :disabled="saving"
                @click="emit('disable', rule)"
              >
                <ShieldCheck class="size-3" />
                解除
              </Button>
            </div>
          </article>
        </div>

        <AdminPagination
          v-if="pagination.total > 0"
          :page="pagination.page"
          :page-size="pagination.pageSize"
          :total="pagination.total"
          @update:page="emit('update:page', $event)"
          @update:page-size="emit('update:pageSize', $event)"
        />
      </section>

      <DialogFooter>
        <Button type="button" variant="outline" @click="emit('update:open', false)">关闭</Button>
      </DialogFooter>
    </DialogContent>
  </Dialog>
</template>

<script setup lang="ts">
import { reactive, ref, watch } from 'vue'
import { LoaderCircle, ShieldAlert, ShieldCheck } from '@lucide/vue'
import AdminFormField from '@/components/admin/AdminFormField.vue'
import AdminPagination from '@/components/admin/AdminPagination.vue'
import AdminStatusBadge, { type AdminStatusTone } from '@/components/admin/AdminStatusBadge.vue'
import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import type {
  VisitorGlobalIPBlockPayload,
  VisitorPagination,
  VisitorIPBlockRule,
} from '@/modules/visitor/visitorTypes'

const props = withDefaults(defineProps<{
  open: boolean
  rules?: VisitorIPBlockRule[]
  pagination?: VisitorPagination
  loading?: boolean
  saving?: boolean
  error?: string
  canCreate?: boolean
  canManage?: boolean
  formatDate: (value: unknown) => string
}>(), {
  rules: () => [],
  pagination: () => ({ page: 1, pageSize: 20, total: 0 }),
  loading: false,
  saving: false,
  error: '',
  canCreate: true,
  canManage: false,
})

const emit = defineEmits<{
  (event: 'update:open', value: boolean): void
  (event: 'create', payload: VisitorGlobalIPBlockPayload): void
  (event: 'disable', rule: VisitorIPBlockRule): void
  (event: 'update:page', value: number): void
  (event: 'update:pageSize', value: number): void
}>()

const form = reactive({
  cidr: '',
  reason: '',
  expiresAt: '',
})
const validationError = ref('')

const resetForm = (): void => {
  form.cidr = ''
  form.reason = ''
  form.expiresAt = ''
  validationError.value = ''
}

watch(() => props.open, (isOpen) => {
  if (!isOpen) resetForm()
})

const submit = (): void => {
  validationError.value = ''
  const cidr = form.cidr.trim()
  const reason = form.reason.trim()
  if (!cidr) {
    validationError.value = '请输入 IP 或 CIDR。'
    return
  }
  if (!reason) {
    validationError.value = '请填写封禁原因。'
    return
  }
  if (!window.confirm('确认创建这条全局 IP/CIDR 封禁规则吗？共享出口 IP 的其他用户也可能被拦截。')) {
    return
  }

  emit('create', {
    cidr,
    reason,
    expires_at: form.expiresAt ? new Date(form.expiresAt).toISOString() : null,
  })
}

const statusLabel = (status?: string): string => ({
  active: '有效',
  expired: '已到期',
  disabled: '已解除',
} as Record<string, string>)[status || ''] || '未知'

const statusTone = (status?: string): AdminStatusTone => ({
  active: 'coral',
  expired: 'amber',
  disabled: 'gray',
} as Record<string, AdminStatusTone>)[status || ''] || 'gray'

const sourceLabel = (source?: string): string => ({
  manual: '手动',
  visitor_profile: '访客画像',
  commercial_crawler: '商业爬虫',
  risk_automation: '风险自动化',
} as Record<string, string>)[source || ''] || source || '未知'
</script>

