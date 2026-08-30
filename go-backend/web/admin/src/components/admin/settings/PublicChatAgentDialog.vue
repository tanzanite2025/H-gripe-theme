<template>
  <Dialog :open="open" @update:open="emit('update:open', $event)">
    <DialogContent size="lg" class="max-h-[90dvh] overflow-y-auto" @open-auto-focus.prevent>
      <form class="space-y-4" @submit.prevent="emit('save')">
        <DialogHeader>
          <DialogTitle>添加 Public Chat 客服 Profile</DialogTitle>
          <DialogDescription>
            选择已存在的后台用户，保存后该用户会作为公开客服出现在 Public Chat 的客服列表里。
            客服组只作为前台展示标签，不参与会话分配。
          </DialogDescription>
        </DialogHeader>

        <Alert v-if="selectedCandidate?.has_profile" class="border-border bg-muted/40 text-foreground">
          <Info class="size-4" />
          <AlertTitle>该用户已有 Profile</AlertTitle>
          <AlertDescription>再次保存会更新现有 Profile，不会重复创建。</AlertDescription>
        </Alert>

        <div v-if="candidates.length === 0 && !loadingCandidates" class="rounded-lg border border-dashed p-4 text-sm text-muted-foreground">
          暂无可绑定的活跃后台用户。候选用户必须是 active 且角色为 admin、manager 或 support。
        </div>

        <div class="grid gap-4 md:grid-cols-2">
          <AdminFormField label="后台用户" required description="候选来自 active 的 admin、manager、support 用户。">
            <Select v-model="form.user_id" :disabled="loadingCandidates || saving">
              <SelectTrigger class="w-full">
                <SelectValue :placeholder="loadingCandidates ? '正在加载候选用户' : '选择已有后台用户'" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem v-for="candidate in candidates" :key="candidate.user_id" :value="String(candidate.user_id)">
                  {{ candidateLabel(candidate) }}
                </SelectItem>
              </SelectContent>
            </Select>
          </AdminFormField>

          <AdminFormField label="Profile 状态" required>
            <Select v-model="form.status" :disabled="saving">
              <SelectTrigger class="w-full"><SelectValue placeholder="请选择状态" /></SelectTrigger>
              <SelectContent>
                <SelectItem value="active">active · 对外展示</SelectItem>
                <SelectItem value="inactive">inactive · 暂不展示</SelectItem>
              </SelectContent>
            </Select>
          </AdminFormField>

          <AdminFormField label="在线状态" required description="前台只显示状态点，不再显示 ONLINE 文案。">
            <Select v-model="form.online_status" :disabled="saving">
              <SelectTrigger class="w-full"><SelectValue placeholder="请选择在线状态" /></SelectTrigger>
              <SelectContent>
                <SelectItem value="online">online · 在线</SelectItem>
                <SelectItem value="busy">busy · 忙碌</SelectItem>
                <SelectItem value="away">away · 暂离</SelectItem>
                <SelectItem value="offline">offline · 离线</SelectItem>
              </SelectContent>
            </Select>
          </AdminFormField>

          <AdminFormField label="Agent ID" description="不填写时自动使用 user-用户ID；如需对接外部系统可手动改。">
            <Input v-model="form.agent_id" :disabled="saving" placeholder="user-1" maxlength="50" />
          </AdminFormField>

          <AdminFormField label="公开名称" description="不填写时使用后台用户显示名。">
            <Input v-model="form.name" :disabled="saving" placeholder="客服名称" />
          </AdminFormField>

          <AdminFormField label="公开邮箱" description="邮箱和 WhatsApp 至少填一个，前台会展示可用的联系方式。">
            <Input v-model="form.email" :disabled="saving" type="email" placeholder="support@example.com" />
          </AdminFormField>

          <AdminFormField label="WhatsApp" description="邮箱和 WhatsApp 至少填一个，前台会展示可用的联系方式。">
            <Input v-model="form.whatsapp" :disabled="saving" placeholder="+1 000 000 0000" />
          </AdminFormField>

          <AdminFormField label="客服头像" class="md:col-span-2">
            <PublicChatAgentAvatarField
              :user-id="form.user_id"
              :avatar="form.avatar"
              :profile-ready="Boolean(selectedCandidate?.has_profile)"
              :disabled="saving"
              @uploaded="emit('avatar-uploaded', $event)"
              @removed="emit('avatar-removed')"
            />
          </AdminFormField>

          <AdminFormField label="展示标签" class="md:col-span-2" description="只用于前台客服卡片显示，方便客户识别，不参与会话分配。">
            <div v-if="groups.length" class="grid gap-2 sm:grid-cols-2">
              <label
                v-for="group in groups"
                :key="group.id"
                class="flex items-start gap-2 rounded-xl border border-border/70 bg-muted/20 px-3 py-2 text-xs"
              >
                <input v-model="form.group_ids" type="checkbox" :value="group.id" class="mt-0.5 size-4 accent-[var(--admin-selected)]" :disabled="saving || group.status !== 'active'" />
                <span class="min-w-0">
                  <strong class="block truncate">{{ group.name }}</strong>
                  <span class="block truncate text-muted-foreground">{{ group.code }}</span>
                </span>
              </label>
            </div>
            <p v-else class="text-xs text-muted-foreground">暂无 active 展示标签，请先创建客服组。</p>
          </AdminFormField>
        </div>

        <DialogFooter>
          <Button type="button" variant="outline" :disabled="saving" @click="emit('update:open', false)">取消</Button>
          <Button type="submit" :disabled="saving || !form.user_id">
            <LoaderCircle v-if="saving" class="size-4 animate-spin" />
            保存 Profile
          </Button>
        </DialogFooter>
      </form>
    </DialogContent>
  </Dialog>
</template>

<script setup lang="ts">
import { Info, LoaderCircle } from '@lucide/vue'
import AdminFormField from '@/components/admin/AdminFormField.vue'
import PublicChatAgentAvatarField from '@/components/admin/settings/PublicChatAgentAvatarField.vue'
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { Button } from '@/components/ui/button'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'

interface PublicChatAgentForm {
  user_id: string | number | null
  status: string
  online_status: string
  agent_id: string
  name: string
  email: string
  whatsapp: string
  avatar: string
  group_ids: Array<string | number>
}

interface PublicChatCandidate {
  user_id: string | number
  display_name?: string
  username?: string
  email?: string
  normalized_role?: string
  raw_role?: string
  has_profile?: boolean
}

interface PublicChatGroup {
  id: string | number
  name: string
  code: string
  status: string
}

const props = withDefaults(defineProps<{
  open?: boolean
  form: PublicChatAgentForm
  candidates?: PublicChatCandidate[]
  selectedCandidate?: PublicChatCandidate | null
  groups?: PublicChatGroup[]
  loadingCandidates?: boolean
  saving?: boolean
}>(), {
  open: false,
  candidates: () => [],
  selectedCandidate: null,
  groups: () => [],
  loadingCandidates: false,
  saving: false,
})

const emit = defineEmits<{
  (event: 'update:open', value: boolean): void
  (event: 'save'): void
  (event: 'avatar-uploaded', avatar: string): void
  (event: 'avatar-removed'): void
}>()

const candidateLabel = (candidate: PublicChatCandidate): string => {
  const name = candidate.display_name || candidate.username || candidate.email || `User #${candidate.user_id}`
  const role = candidate.normalized_role || candidate.raw_role || 'unknown'
  const profileSuffix = candidate.has_profile ? ' · 已有 Profile' : ''
  return `${name} · ${role} · #${candidate.user_id}${profileSuffix}`
}
</script>
