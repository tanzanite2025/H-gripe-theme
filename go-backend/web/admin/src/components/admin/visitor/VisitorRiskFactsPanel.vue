<template>
  <AdminTablePanel class="h-full min-h-0" :loading="loading" scroll-body>
    <Table class="min-w-[1440px]">
      <TableHeader>
        <TableRow>
          <TableHead class="w-24">日期</TableHead>
          <TableHead class="w-40">风险</TableHead>
          <TableHead class="w-52">请求身份</TableHead>
          <TableHead class="w-44">请求量</TableHead>
          <TableHead class="w-56">异常计数</TableHead>
          <TableHead class="w-44">身份扩散</TableHead>
          <TableHead>样例路径</TableHead>
          <TableHead class="w-44">最后出现</TableHead>
          <TableHead class="w-28 text-right">决策</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        <TableEmpty v-if="facts.length === 0" :colspan="9">
          <div class="flex flex-col items-center text-muted-foreground">
            <ShieldAlert class="mb-2 size-7 opacity-55" />
            <span class="text-xs">暂无风险事实</span>
          </div>
        </TableEmpty>

        <TableRow v-for="fact in facts" :key="fact.id">
          <TableCell class="font-mono text-xs font-bold text-muted-foreground">
            {{ formatDay(fact.day) }}
          </TableCell>
          <TableCell>
            <div class="flex flex-wrap items-center gap-1.5">
              <AdminStatusBadge :tone="riskTone(fact.risk_level)">
                {{ riskLabel(fact.risk_level) }}
              </AdminStatusBadge>
              <span class="font-mono text-xs font-black text-foreground">R{{ fact.risk_score || 0 }}</span>
            </div>
            <p class="mt-1 font-mono text-[11px] text-muted-foreground">#{{ fact.id }}</p>
            <p v-if="fact.decision" class="mt-1 flex items-center gap-1 text-[10px] font-bold text-emerald-500">
              <ShieldCheck class="size-3" />
              人工：{{ decisionActionLabel(fact.decision.action) }}
            </p>
          </TableCell>
          <TableCell>
            <p class="break-all font-mono text-[11px] font-bold text-foreground">{{ fact.ip_hash_preview }}</p>
 <p class="mt-1 break-all font-mono text-[10px] text-muted-foreground">{{ fact.user_agent_hash_preview || 'no-ua'}}</p>
            <p v-if="fact.country_code" class="mt-1 font-mono text-[10px] text-muted-foreground">{{ fact.country_code }}</p>
          </TableCell>
          <TableCell>
            <p class="font-mono text-sm font-black">{{ fact.request_count || 0 }}</p>
            <p class="mt-1 text-[11px] text-muted-foreground">无 Cookie：{{ fact.no_cookie_request_count || 0 }}</p>
          </TableCell>
          <TableCell>
            <div class="grid grid-cols-2 gap-1.5 text-[11px] text-muted-foreground">
              <span>4xx/异常 {{ fact.invalid_request_count || 0 }}</span>
              <span>认证 {{ fact.auth_failure_count || 0 }}</span>
              <span>结账 {{ fact.checkout_failure_count || 0 }}</span>
              <span>Bot UA {{ fact.bot_like_user_agent_count || 0 }}</span>
            </div>
          </TableCell>
          <TableCell>
            <div class="grid grid-cols-1 gap-1 text-[11px] text-muted-foreground">
              <span>路径 {{ fact.unique_path_count || 0 }}</span>
              <span>匿名 ID {{ fact.unique_anonymous_count || 0 }}</span>
              <span>Session {{ fact.unique_session_count || 0 }}</span>
              <span>有效动作 {{ fact.meaningful_action_count || 0 }}</span>
            </div>
          </TableCell>
          <TableCell>
            <div class="flex max-w-[420px] flex-wrap gap-1.5">
              <span
                v-for="path in (fact.sample_paths || []).slice(0, 5)"
                :key="path"
                class="max-w-full truncate rounded-full border border-border bg-muted/50 px-2 py-1 font-mono text-[10px] text-muted-foreground"
              >
                {{ path }}
              </span>
              <span v-if="!fact.sample_paths?.length" class="text-[11px] text-muted-foreground">no-sample-path</span>
            </div>
          </TableCell>
          <TableCell class="text-xs text-muted-foreground">{{ formatDate(fact.last_seen_at) }}</TableCell>
          <TableCell class="text-right">
            <DropdownMenu>
              <DropdownMenuTrigger as-child>
                <Button variant="ghost" size="icon" :aria-label="`管理风险事实 ${fact.id}`">
                  <MoreHorizontal class="size-4" />
                </Button>
              </DropdownMenuTrigger>
              <DropdownMenuContent align="end" class="w-44">
                <DropdownMenuItem @select="openDecisionDialog(fact)">
                  <ShieldCheck class="size-4" />
                  {{ fact.decision ? '更新人工决策' : '记录人工决策' }}
                </DropdownMenuItem>
              </DropdownMenuContent>
            </DropdownMenu>
          </TableCell>
        </TableRow>
      </TableBody>
    </Table>

    <template #footer>
      <AdminPagination
        :page="pagination.page"
        :page-size="pagination.pageSize"
        :total="pagination.total"
        @update:page="emit('update-page', $event)"
        @update:page-size="emit('update-page-size', $event)"
      />
    </template>
  </AdminTablePanel>

  <Dialog v-model:open="decisionDialogOpen">
    <DialogContent size="md" class="max-h-[90dvh] overflow-y-auto" @open-auto-focus.prevent>
      <form class="space-y-5" @submit.prevent="submitDecision">
        <DialogHeader>
          <DialogTitle>记录人工风险决策</DialogTitle>
          <DialogDescription>
            这只会保存人工判断，当前不会直接触发封禁或限流。
          </DialogDescription>
        </DialogHeader>

        <div v-if="selectedFact" class="grid grid-cols-2 gap-3 rounded-2xl border border-border/70 bg-muted/30 p-3 text-xs">
          <div>
            <p class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/70">FACT</p>
            <p class="mt-1 font-mono font-bold text-foreground">#{{ selectedFact.id }}</p>
          </div>
          <div>
            <p class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/70">SCOPE</p>
            <p class="mt-1 font-mono font-bold text-foreground">{{ decisionScopeLabel(selectedFact) }}</p>
          </div>
        </div>

        <div class="space-y-4">
          <AdminFormField label="人工动作" required description="动作只作为后续风控执行层的输入">
            <Select v-model="decisionForm.action">
              <SelectTrigger class="w-full"><SelectValue /></SelectTrigger>
              <SelectContent>
                <SelectItem value="ignore">忽略</SelectItem>
                <SelectItem value="watch">观察</SelectItem>
                <SelectItem value="temporary_block">临时封禁（仅记录）</SelectItem>
                <SelectItem value="block_candidate">封禁候选</SelectItem>
              </SelectContent>
            </Select>
          </AdminFormField>

          <AdminFormField label="判断原因" required description="保留给后续复核和审计使用">
            <Textarea v-model="decisionForm.reason" class="min-h-24" maxlength="500" placeholder="例如：同一 IP/UA 在结账接口连续失败，需要人工观察。" />
          </AdminFormField>

          <AdminFormField
            label="到期时间"
            :required="decisionForm.action === 'temporary_block'"
            :description="decisionForm.action === 'temporary_block' ? '临时封禁必须设置未来时间' : '留空表示持续有效，直到后续人工决策覆盖'"
          >
            <Input v-model="decisionForm.expiresAt" type="datetime-local" />
          </AdminFormField>
        </div>

        <p v-if="decisionError" class="rounded-xl border border-red-500/30 bg-red-500/10 px-3 py-2 text-xs font-bold text-red-300">
          {{ decisionError }}
        </p>

        <DialogFooter>
          <Button type="button" variant="outline" @click="decisionDialogOpen = false">取消</Button>
          <Button type="submit" :disabled="decisionSubmitting">
            <LoaderCircle v-if="decisionSubmitting" class="size-4 animate-spin" />
            {{ decisionSubmitting ? '保存中' : '保存决策' }}
          </Button>
        </DialogFooter>
      </form>
    </DialogContent>
  </Dialog>
</template>

<script setup lang="ts">
import { reactive, ref } from 'vue'
import { LoaderCircle, MoreHorizontal, ShieldAlert, ShieldCheck } from '@lucide/vue'
import AdminFormField from '@/components/admin/AdminFormField.vue'
import AdminPagination from '@/components/admin/AdminPagination.vue'
import AdminStatusBadge, { type AdminStatusTone } from '@/components/admin/AdminStatusBadge.vue'
import AdminTablePanel from '@/components/admin/AdminTablePanel.vue'
import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle
} from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Table, TableBody, TableCell, TableEmpty, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Textarea } from '@/components/ui/textarea'
import type {
  VisitorPagination,
  VisitorRiskDecisionPayload,
  VisitorRiskFact,
} from './visitorTypes'

const props = withDefaults(defineProps<{
  loading?: boolean
  facts?: VisitorRiskFact[]
  pagination: VisitorPagination
  formatDate: (value: unknown) => string
  saveDecision: (fact: VisitorRiskFact, payload: VisitorRiskDecisionPayload) => Promise<void>
}>(), {
  loading: false,
  facts: () => [],
})

const emit = defineEmits<{
  (event: 'update-page', page: number): void
  (event: 'update-page-size', pageSize: number): void
}>()
const decisionDialogOpen = ref(false)
const selectedFact = ref<VisitorRiskFact | null>(null)
const decisionSubmitting = ref(false)
const decisionError = ref('')
const decisionForm = reactive({
  action: 'watch',
  reason: '',
  expiresAt: '',
})

const riskLabel = (level?: string): string => ({
  normal: '正常',
  watch: '观察',
  suspicious: '可疑',
  block: '封禁候选',
} as Record<string, string>)[level || ''] || '正常'

const riskTone = (level?: string): AdminStatusTone => ({
  normal: 'gray',
  watch: 'amber',
  suspicious: 'coral',
  block: 'coral',
} as Record<string, AdminStatusTone>)[level || ''] || 'gray'

const decisionActionLabel = (action?: string): string => ({
  ignore: '忽略',
  watch: '观察',
  temporary_block: '临时封禁',
  block_candidate: '封禁候选',
} as Record<string, string>)[action || ''] || action || '未知'

const decisionScopeLabel = (fact?: VisitorRiskFact | null): string => fact?.user_agent_hash_preview ? 'IP + UA' : 'IP'

const formatDateTimeLocal = (value: unknown): string => {
  if (!value) return ''
  const date = new Date(value as string | number | Date)
  if (Number.isNaN(date.getTime())) return ''
  const pad = (part) => String(part).padStart(2, '0')
  return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())}T${pad(date.getHours())}:${pad(date.getMinutes())}`
}

const openDecisionDialog = (fact: VisitorRiskFact): void => {
  selectedFact.value = fact
  decisionError.value = ''
  decisionForm.action = fact.decision?.action || 'watch'
  decisionForm.reason = fact.decision?.reason || ''
  decisionForm.expiresAt = formatDateTimeLocal(fact.decision?.expires_at)
  decisionDialogOpen.value = true
}

interface ErrorLike {
  response?: { data?: { error?: string; message?: string } }
}

const submitDecision = async (): Promise<void> => {
  if (!selectedFact.value) return
  const reason = decisionForm.reason.trim()
  if (!reason) {
    decisionError.value = '请填写判断原因。'
    return
  }
  if (decisionForm.action === 'temporary_block' && !decisionForm.expiresAt) {
    decisionError.value = '临时封禁必须设置到期时间。'
    return
  }

  decisionSubmitting.value = true
  decisionError.value = ''
  try {
    await props.saveDecision(selectedFact.value, {
      action: decisionForm.action,
      reason,
      expires_at: decisionForm.expiresAt ? new Date(decisionForm.expiresAt).toISOString() : null,
    })
    decisionDialogOpen.value = false
  } catch (error) {
    decisionError.value = error?.response?.data?.error || error?.response?.data?.message || '保存失败，请稍后重试。'
  } finally {
    decisionSubmitting.value = false
  }
}

const formatDay = (day: unknown): string => {
  if (!day) return '-'
  return new Date(day as string | number | Date).toLocaleDateString('zh-CN')
}
</script>
