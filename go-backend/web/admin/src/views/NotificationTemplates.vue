<template>
  <div class="space-y-6">
    <AdminPageHeader title="邮件模板" description="维护订单与售后事务邮件模板，保存会生成不可变版本快照。" />

    <div class="grid gap-6 lg:grid-cols-[18rem_minmax(0,1fr)]">
      <section class="rounded-lg border bg-card p-4">
        <div class="mb-3 flex items-center justify-between">
          <h2 class="font-semibold">模板列表</h2>
          <Button size="icon" variant="ghost" :disabled="loading" title="刷新" @click="loadTemplates">
            <RefreshCw class="size-4" :class="loading ? 'animate-spin' : ''" />
          </Button>
        </div>
        <div v-if="loading" class="py-8 text-center text-sm text-muted-foreground">正在加载...</div>
        <div v-else class="space-y-1">
          <button
            v-for="template in templates"
            :key="template.id"
            type="button"
            class="w-full rounded-md border px-3 py-2 text-left text-sm transition-colors hover:bg-muted"
            :class="selected?.id === template.id ? 'border-primary bg-muted' : 'border-transparent'"
            @click="selectTemplate(template)"
          >
            <div class="flex items-center justify-between gap-2">
              <span class="truncate font-medium">{{ template.code }}</span>
              <span class="text-xs text-muted-foreground">{{ template.locale }}</span>
            </div>
            <div class="mt-1 flex items-center justify-between text-xs text-muted-foreground">
              <span>v{{ template.version }}</span>
              <span>{{ template.is_enabled ? '启用' : '停用' }}</span>
            </div>
          </button>
        </div>
      </section>

      <section v-if="selected" class="space-y-4 rounded-lg border bg-card p-6">
        <div class="flex flex-wrap items-start justify-between gap-3">
          <div>
            <h2 class="text-lg font-semibold">{{ selected.code }} / {{ selected.locale }}</h2>
            <p class="text-sm text-muted-foreground">当前版本 v{{ selected.version }} · {{ selected.category }}</p>
          </div>
          <div class="flex gap-2">
            <Button variant="outline" :disabled="previewing" @click="previewTemplate">
              <Eye class="mr-2 size-4" />预览
            </Button>
            <Button :disabled="saving" @click="saveTemplate">
              <LoaderCircle v-if="saving" class="mr-2 size-4 animate-spin" />
              <Save v-else class="mr-2 size-4" />保存
            </Button>
          </div>
        </div>

        <div class="grid gap-4 md:grid-cols-2">
          <label class="space-y-1 text-sm"><span>名称</span><Input v-model="draft.name" /></label>
          <label class="flex items-center gap-2 pt-6 text-sm"><input v-model="draft.is_enabled" type="checkbox" />启用模板</label>
        </div>
        <label class="block space-y-1 text-sm"><span>邮件标题</span><Input v-model="draft.subject_template" /></label>
        <div class="grid gap-4 xl:grid-cols-2">
          <label class="space-y-1 text-sm"><span>HTML 正文</span><Textarea v-model="draft.body_html" class="min-h-72 font-mono text-xs" /></label>
          <label class="space-y-1 text-sm"><span>纯文本正文</span><Textarea v-model="draft.body_text" class="min-h-72 font-mono text-xs" /></label>
        </div>
        <label class="block space-y-1 text-sm"><span>变更原因</span><Input v-model="draft.change_reason" placeholder="例如：更新售后寄回说明" /></label>

        <div class="grid gap-4 border-t pt-4 xl:grid-cols-2">
          <div><h3 class="mb-2 text-sm font-medium">必填变量</h3><div class="flex flex-wrap gap-1"> <code v-for="variable in selected.required_variables" :key="variable" class="rounded bg-muted px-2 py-1 text-xs">{{ variableLabel(variable) }}</code></div></div>
          <div><h3 class="mb-2 text-sm font-medium">允许变量</h3><div class="flex flex-wrap gap-1"> <code v-for="variable in selected.allowed_variables" :key="variable" class="rounded bg-muted px-2 py-1 text-xs">{{ variableLabel(variable) }}</code></div></div>
        </div>

        <div v-if="versions.length" class="border-t pt-4">
          <h3 class="mb-2 text-sm font-medium">版本历史</h3>
          <div class="space-y-1 text-sm">
            <div v-for="version in versions" :key="version.id" class="flex justify-between rounded bg-muted/50 px-3 py-2">
              <span class="min-w-0 truncate">v{{ version.version }} {{ version.change_reason || '无说明' }}</span>
              <span class="flex shrink-0 items-center gap-2 text-muted-foreground">
                <span>{{ formatDate(version.created_at) }}</span>
                <Button
                  v-if="version.version < selected.version"
                  size="icon"
                  variant="ghost"
                  class="size-7"
                  :disabled="rollingBack"
                  title="回滚为新版本"
                  @click.stop="rollbackVersion(version.version)"
                >
                  <RotateCcw class="size-3.5" :class="rollingBack ? 'animate-spin' : ''" />
                </Button>
              </span>
            </div>
          </div>
        </div>
      </section>
      <section v-else class="rounded-lg border bg-card p-10 text-center text-muted-foreground">请选择一个模板</section>
    </div>

    <Dialog v-model:open="previewOpen">
      <DialogContent class="max-w-4xl">
        <DialogHeader><DialogTitle>邮件预览</DialogTitle></DialogHeader>
        <div v-if="preview" class="space-y-4">
          <div><div class="mb-1 text-xs font-medium text-muted-foreground">标题</div><div class="rounded border p-3">{{ preview.subject }}</div></div>
          <div class="grid gap-4 lg:grid-cols-2"><div><div class="mb-1 text-xs font-medium text-muted-foreground">HTML</div><iframe :srcdoc="preview.html" class="h-80 w-full rounded border bg-white" sandbox="" /></div><div><div class="mb-1 text-xs font-medium text-muted-foreground">纯文本</div><pre class="h-80 overflow-auto whitespace-pre-wrap rounded border p-3 text-sm">{{ preview.text }}</pre></div></div>
        </div>
      </DialogContent>
    </Dialog>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { toast } from 'vue-sonner'
import { Eye, LoaderCircle, RefreshCw, RotateCcw, Save } from '@lucide/vue'
import AdminPageHeader from '@/components/admin/AdminPageHeader.vue'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import { Dialog, DialogContent, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import notificationTemplatesApi, { type NotificationTemplate, type NotificationTemplatePreview, type NotificationTemplateVersion } from '@/api/notificationTemplates'

const templates = ref<NotificationTemplate[]>([])
const selected = ref<NotificationTemplate | null>(null)
const versions = ref<NotificationTemplateVersion[]>([])
const loading = ref(false)
const saving = ref(false)
const previewing = ref(false)
const rollingBack = ref(false)
const previewOpen = ref(false)
const preview = ref<NotificationTemplatePreview | null>(null)
const draft = reactive({ name: '', subject_template: '', body_html: '', body_text: '', is_enabled: true, change_reason: '' })

const applyDraft = (template: NotificationTemplate) => {
  draft.name = template.name; draft.subject_template = template.subject_template; draft.body_html = template.body_html; draft.body_text = template.body_text; draft.is_enabled = template.is_enabled; draft.change_reason = ''
}

const loadTemplates = async () => {
  loading.value = true
  try { templates.value = await notificationTemplatesApi.list(); if (!selected.value && templates.value.length) await selectTemplate(templates.value[0]) } catch (error) { toast.error(error instanceof Error ? error.message : '模板加载失败') } finally { loading.value = false }
}

const selectTemplate = async (template: NotificationTemplate) => {
  selected.value = template; applyDraft(template)
  try { versions.value = await notificationTemplatesApi.versions(template.id) } catch (error) { toast.error(error instanceof Error ? error.message : '版本历史加载失败') }
}

const saveTemplate = async () => {
  if (!selected.value) return
  saving.value = true
  try { const updated = await notificationTemplatesApi.save({ id: selected.value.id, code: selected.value.code, locale: selected.value.locale, category: selected.value.category, version: selected.value.version + 1, allowed_variables: selected.value.allowed_variables, required_variables: selected.value.required_variables, ...draft }); selected.value = updated; templates.value = templates.value.map((item) => item.id === updated.id ? updated : item); await selectTemplate(updated); toast.success('模板已保存') } catch (error) { toast.error(error instanceof Error ? error.message : '模板保存失败') } finally { saving.value = false }
}

const previewTemplate = async () => {
  if (!selected.value) return
  previewing.value = true
  try { const variables = Object.fromEntries(selected.value.allowed_variables.map((variable) => [variable, `[${variable}]`])); preview.value = await notificationTemplatesApi.preview({ code: selected.value.code, locale: selected.value.locale, variables }); previewOpen.value = true } catch (error) { toast.error(error instanceof Error ? error.message : '模板预览失败') } finally { previewing.value = false }
}

const rollbackVersion = async (version: number) => {
  if (!selected.value || !window.confirm(`确定将模板回滚到 v${version} 吗？这会生成一个新的当前版本。`)) return
  rollingBack.value = true
  try {
    const updated = await notificationTemplatesApi.rollback(selected.value.id, version, `rollback to version ${version}`)
    selected.value = updated
    templates.value = templates.value.map((item) => item.id === updated.id ? updated : item)
    await selectTemplate(updated)
    toast.success('模板已回滚并生成新版本')
  } catch (error) {
    toast.error(error instanceof Error ? error.message : '模板回滚失败')
  } finally {
    rollingBack.value = false
  }
}

const formatDate = (value: string) => value ? new Date(value).toLocaleString() : ''
const variableLabel = (value: string) => `{{${value}}}`
onMounted(loadTemplates)
</script>
