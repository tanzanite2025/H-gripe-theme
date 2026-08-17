<template>
 <div class="space-y-4">
    <AdminPageHeader
      title="URL 管理 / 重定向"
      description="管理已退役路径的精确永久跳转规则"
    >
      <template #actions>
        <Button variant="outline" :disabled="loading" @click="load">
 <RefreshCw :class="['size-4', loading ? 'animate-spin': '']" />
          刷新
        </Button>
        <Button :disabled="!canEdit" @click="openCreate">
 <Plus class="size-4" />
          新建重定向
        </Button>
      </template>
    </AdminPageHeader>

    <AdminTablePanel :loading="loading">
 <Table class="min-w-[980px]">
        <TableHeader>
          <TableRow>
 <TableHead class="w-[280px]">来源路径</TableHead>
 <TableHead class="w-[280px]">目标路径</TableHead>
 <TableHead class="w-28">状态码</TableHead>
 <TableHead class="w-32">发布状态</TableHead>
            <TableHead>变更原因</TableHead>
 <TableHead class="w-44">发布时间</TableHead>
 <TableHead class="w-24 text-right">操作</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          <TableRow v-if="rules.length === 0">
 <TableCell colspan="7" class="h-40 text-center text-sm text-muted-foreground">
              {{ loading ? '正在加载重定向规则' : '暂无受管重定向规则' }}
            </TableCell>
          </TableRow>
          <TableRow v-for="rule in rules" :key="rule.id">
 <TableCell class="font-mono text-xs">{{ rule.source_path }}</TableCell>
 <TableCell class="font-mono text-xs">{{ rule.target_path }}</TableCell>
 <TableCell class="font-mono text-xs">{{ rule.status_code }}</TableCell>
            <TableCell>
              <AdminStatusBadge :tone="stateTone(rule.state)">
                {{ stateLabel(rule.state) }}
              </AdminStatusBadge>
            </TableCell>
 <TableCell class="max-w-72 truncate text-xs" :title="rule.reason">{{ rule.reason }}</TableCell>
 <TableCell class="text-xs text-muted-foreground">{{ formatRouteCatalogDate(rule.published_at) }}</TableCell>
 <TableCell class="text-right">
 <div class="inline-flex gap-1">
                <Button
                  v-if="rule.state === 'draft'"
                  variant="ghost"
                  size="icon"
                  title="发布重定向"
                  aria-label="发布重定向"
                  :disabled="!canEdit || actionID === rule.id"
                  @click="publish(rule)"
                >
 <Upload class="size-4" />
                </Button>
                <Button
                  v-if="rule.state !== 'disabled'"
                  variant="ghost"
                  size="icon"
                  title="停用重定向"
                  aria-label="停用重定向"
                  :disabled="!canEdit || actionID === rule.id"
                  @click="disable(rule)"
                >
 <Ban class="size-4" />
                </Button>
              </div>
            </TableCell>
          </TableRow>
        </TableBody>
      </Table>
    </AdminTablePanel>

    <Dialog v-model:open="createOpen">
      <DialogContent>
        <DialogHeader>
          <DialogTitle>新建重定向</DialogTitle>
          <DialogDescription>仅用于已退役的前台路径；目标必须是当前有效的规范路径。</DialogDescription>
        </DialogHeader>

 <form class="space-y-4" @submit.prevent="create">
 <label class="block space-y-1">
 <span class="text-xs font-bold">来源路径</span>
            <Input v-model="form.source_path" autocomplete="off" placeholder="/legacy-page" />
          </label>
 <label class="block space-y-1">
 <span class="text-xs font-bold">目标路径</span>
            <Input v-model="form.target_path" autocomplete="off" placeholder="/support/faqs" />
          </label>
 <label class="block space-y-1">
 <span class="text-xs font-bold">永久跳转状态码</span>
            <Select v-model="form.status_code">
              <SelectTrigger><SelectValue /></SelectTrigger>
              <SelectContent>
                <SelectItem value="301">301</SelectItem>
                <SelectItem value="308">308</SelectItem>
              </SelectContent>
            </Select>
          </label>
 <label class="block space-y-1">
 <span class="text-xs font-bold">变更原因</span>
            <Input v-model="form.reason" autocomplete="off" placeholder="例如：旧版帮助中心路径迁移" />
          </label>

          <DialogFooter>
            <Button type="button" variant="outline" :disabled="saving" @click="createOpen = false">取消</Button>
            <Button type="submit" :disabled="saving">
 <Plus class="size-4" />
              {{ saving ? '创建中' : '创建草稿' }}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref, watch } from 'vue'
import { Ban, Plus, RefreshCw, Upload } from '@lucide/vue'
import { toast } from 'vue-sonner'
import { useRoute, useRouter } from 'vue-router'
import AdminPageHeader from '@/components/admin/AdminPageHeader.vue'
import AdminStatusBadge, { type AdminStatusTone } from '@/components/admin/AdminStatusBadge.vue'
import AdminTablePanel from '@/components/admin/AdminTablePanel.vue'
import { Button } from '@/components/ui/button'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { formatRouteCatalogDate } from '@/components/admin/url-management/route-catalog/routeCatalogPresentation'
import {
  storefrontRedirectRulesApi,
  type StorefrontRedirectRule,
  type StorefrontRedirectRuleState,
} from '@/modules/url-management/redirectRules'
import { storefrontURLIssuesApi } from '@/modules/url-management/urlIssues'
import { useAuthStore } from '@/stores/auth'

const authStore = useAuthStore()
const route = useRoute()
const router = useRouter()
const canEdit = authStore.hasPermission('url:edit')
const loading = ref(false)
const saving = ref(false)
const actionID = ref<number | null>(null)
const createOpen = ref(false)
const rules = ref<StorefrontRedirectRule[]>([])
const sourceIssueID = ref<number | null>(null)
const form = reactive({
  source_path: '',
  target_path: '',
  status_code: '301',
  reason: '',
})

const stateLabel = (state: StorefrontRedirectRuleState): string => ({
  draft: '草稿',
  published: '已发布',
  disabled: '已停用',
}[state])

const stateTone = (state: StorefrontRedirectRuleState): AdminStatusTone => {
  switch (state) {
    case 'published':
      return 'green'
    case 'disabled':
      return 'gray'
    default:
      return 'amber'
  }
}

const load = async (): Promise<void> => {
  loading.value = true
  try {
    rules.value = await storefrontRedirectRulesApi.list()
  } catch (error) {
    console.error('Failed to load storefront redirect rules:', error)
    toast.error('重定向规则加载失败')
  } finally {
    loading.value = false
  }
}

const resetForm = (): void => {
  form.source_path = ''
  form.target_path = ''
  form.status_code = '301'
  form.reason = ''
}

const openCreate = (): void => {
  resetForm()
  sourceIssueID.value = null
  void router.replace({ query: {} })
  createOpen.value = true
}

const openCreateForSourcePath = (value: unknown): void => {
  if (typeof value !== 'string' || !value.trim()) return
  resetForm()
  form.source_path = value
  const issueID = Number(route.query.issue_id)
  sourceIssueID.value = Number.isInteger(issueID) && issueID > 0 ? issueID : null
  createOpen.value = true
}

const create = async (): Promise<void> => {
  if (!canEdit || saving.value) return
  saving.value = true
  try {
    const created = await storefrontRedirectRulesApi.create({
      source_path: form.source_path,
      target_path: form.target_path,
      status_code: Number(form.status_code) as 301 | 308,
      reason: form.reason,
    })
    if (sourceIssueID.value) {
      try {
        await storefrontURLIssuesApi.linkRedirect(sourceIssueID.value, created.id)
        toast.success('重定向草稿已创建并关联问题')
      } catch (error) {
        console.error('Failed to link storefront redirect rule to URL issue:', error)
        toast.error('重定向草稿已创建，但关联问题失败')
      }
    } else {
      toast.success('重定向草稿已创建')
    }
    createOpen.value = false
    sourceIssueID.value = null
    await router.replace({ query: {} })
    await load()
  } catch (error) {
    console.error('Failed to create storefront redirect rule:', error)
    toast.error('重定向草稿创建失败')
  } finally {
    saving.value = false
  }
}

const publish = async (rule: StorefrontRedirectRule): Promise<void> => {
  if (!canEdit || actionID.value) return
  actionID.value = rule.id
  try {
    await storefrontRedirectRulesApi.publish(rule.id)
    toast.success('重定向已发布')
    await load()
  } catch (error) {
    console.error('Failed to publish storefront redirect rule:', error)
    toast.error('重定向发布失败')
  } finally {
    actionID.value = null
  }
}

const disable = async (rule: StorefrontRedirectRule): Promise<void> => {
  if (!canEdit || actionID.value) return
  actionID.value = rule.id
  try {
    await storefrontRedirectRulesApi.disable(rule.id)
    toast.success('重定向已停用')
    await load()
  } catch (error) {
    console.error('Failed to disable storefront redirect rule:', error)
    toast.error('重定向停用失败')
  } finally {
    actionID.value = null
  }
}

onMounted(() => {
  void load()
})

watch(
  () => route.query.source_path,
  openCreateForSourcePath,
  { immediate: true },
)
</script>
