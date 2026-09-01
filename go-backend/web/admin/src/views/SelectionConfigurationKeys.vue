<template>
  <div class="space-y-4">
    <AdminPageHeader title="选型配置 Key 管理" description="统一维护问题 KEY 和回答 KEY，问卷里只从这里选。">
      <template #actions>
        <Button variant="outline" :disabled="loading" @click="reload">
          <RefreshCcw class="size-4" />
          重新加载
        </Button>
        <Button variant="outline" :disabled="loading" @click="openCreateDialog('question_key')">
          <Plus class="size-4" />
          新增问题 KEY
        </Button>
        <Button :disabled="loading" @click="openCreateDialog('answer_key')">
          <Plus class="size-4" />
          新增回答 KEY
        </Button>
      </template>
    </AdminPageHeader>

    <Tabs v-model:model-value="activeKind" class="space-y-4">
      <TabsList class="h-auto flex-wrap justify-start rounded-xl">
        <TabsTrigger v-for="item in kindOptions" :key="item.value" :value="item.value">{{ item.label }}</TabsTrigger>
      </TabsList>

      <TabsContent v-for="item in kindOptions" :key="item.value" :value="item.value" class="space-y-4">
        <Card>
          <CardHeader class="border-b">
            <div class="flex flex-wrap items-center justify-between gap-3">
              <div>
                <CardTitle class="flex items-center gap-2">
                  <ListChecks class="size-5" />
                  {{ item.label }}
                </CardTitle>
                <CardDescription>{{ selectionConfigurationKeyKindDescriptions[item.value] }}</CardDescription>
              </div>
              <Button variant="outline" :disabled="loading" @click="openCreateDialog(item.value)">
                <Plus class="size-4" />
                新增
              </Button>
            </div>
          </CardHeader>
          <CardContent class="p-0">
            <div v-if="loading" class="flex items-center justify-center gap-2 p-8 text-sm text-muted-foreground">
              <LoaderCircle class="size-4 animate-spin" />
              加载中...
            </div>
            <div v-else-if="keysByKind(item.value).length" class="divide-y">
              <div
                v-for="key in keysByKind(item.value)"
                :key="key.id"
                class="grid gap-3 px-4 py-4 lg:grid-cols-[1.1fr_1fr_1fr_0.6fr_0.6fr_auto]"
              >
                <div>
                  <div class="text-xs font-black uppercase tracking-widest text-muted-foreground/70">Code</div>
                  <div class="mt-1 font-mono text-sm font-bold">{{ key.code }}</div>
                </div>
                <div>
                  <div class="text-xs font-black uppercase tracking-widest text-muted-foreground/70">显示名称</div>
                  <div class="mt-1 text-sm font-bold">{{ key.display_label }}</div>
                </div>
                <div>
                  <div class="text-xs font-black uppercase tracking-widest text-muted-foreground/70">说明</div>
                  <div class="mt-1 text-sm text-muted-foreground">{{ key.description || '无' }}</div>
                </div>
                <div>
                  <div class="text-xs font-black uppercase tracking-widest text-muted-foreground/70">状态</div>
                  <Badge class="mt-1" :variant="key.is_enabled ? 'default' : 'secondary'">
                    {{ key.is_enabled ? '启用' : '停用' }}
                  </Badge>
                </div>
                <div>
                  <div class="text-xs font-black uppercase tracking-widest text-muted-foreground/70">排序</div>
                  <div class="mt-1 font-mono text-sm font-bold">{{ key.sort_order }}</div>
                </div>
                <div class="flex items-center gap-2 lg:justify-end">
                  <Button size="sm" variant="outline" @click="openEditDialog(key)">编辑</Button>
                </div>
              </div>
            </div>
            <div v-else class="p-8 text-center text-sm text-muted-foreground">暂无配置 Key</div>
          </CardContent>
        </Card>
      </TabsContent>
    </Tabs>

    <SelectionConfigurationKeyEditorDialog
      v-model:open="editorOpen"
      :mode="editorMode"
      :form="editorForm"
      :kind-options="kindOptions"
      :disabled="saving"
      :saving="saving"
      @submit="saveKey"
    />
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { toast } from 'vue-sonner'
import { ListChecks, LoaderCircle, Plus, RefreshCcw } from '@lucide/vue'
import AdminPageHeader from '@/components/admin/AdminPageHeader.vue'
import SelectionConfigurationKeyEditorDialog from '@/components/admin/selection-configuration/SelectionConfigurationKeyEditorDialog.vue'
import selectionConfigurationKeyApi, { type SelectionConfigurationKeyRecord } from '@/api/selectionConfigurationKeys'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import {
  selectionConfigurationKeyKindAnswerKey,
  selectionConfigurationKeyKindDescriptions,
  selectionConfigurationKeyKindOptions,
  selectionConfigurationKeyKindQuestionKey,
  type SelectionConfigurationKeyEditorForm,
  type SelectionConfigurationKeyKind,
} from '@/modules/selection-configuration/selectionConfigurationKeys'

const loading = ref(false)
const saving = ref(false)
const activeKind = ref<SelectionConfigurationKeyKind>(selectionConfigurationKeyKindQuestionKey)
const keyRecords = ref<SelectionConfigurationKeyRecord[]>([])
const editorOpen = ref(false)
const editorMode = ref<'create' | 'edit'>('create')
const editorForm = reactive<SelectionConfigurationKeyEditorForm>({
  kind: selectionConfigurationKeyKindQuestionKey,
  code: '',
  display_label: '',
  description: '',
  is_enabled: true,
  sort_order: 10,
})

const kindOptions = selectionConfigurationKeyKindOptions

const keysByKind = (kind: SelectionConfigurationKeyKind) => (
  keyRecords.value.filter((item) => item.kind === kind)
)

const resetEditorForm = (kind: SelectionConfigurationKeyKind) => {
  editorForm.id = undefined
  editorForm.kind = kind
  editorForm.code = ''
  editorForm.display_label = ''
  editorForm.description = ''
  editorForm.is_enabled = true
  editorForm.sort_order = 10
}

const openCreateDialog = (kind: SelectionConfigurationKeyKind) => {
  editorMode.value = 'create'
  resetEditorForm(kind)
  editorOpen.value = true
}

const openEditDialog = (record: SelectionConfigurationKeyRecord) => {
  editorMode.value = 'edit'
  editorForm.id = record.id
  editorForm.kind = record.kind as SelectionConfigurationKeyKind
  editorForm.code = record.code
  editorForm.display_label = record.display_label
  editorForm.description = record.description
  editorForm.is_enabled = record.is_enabled
  editorForm.sort_order = record.sort_order
  editorOpen.value = true
}

const reload = async () => {
  loading.value = true
  try {
    const [questionKeys, answerKeys] = await Promise.all([
      selectionConfigurationKeyApi.listKeys(selectionConfigurationKeyKindQuestionKey, true),
      selectionConfigurationKeyApi.listKeys(selectionConfigurationKeyKindAnswerKey, true),
    ])
    keyRecords.value = [...questionKeys, ...answerKeys]
  } catch (error) {
    console.error('Failed to load selection configuration keys:', error)
    toast.error('选型配置 Key 加载失败')
  } finally {
    loading.value = false
  }
}

const saveKey = async () => {
  saving.value = true
  try {
    const payload = {
      kind: editorForm.kind,
      code: editorForm.code.trim(),
      display_label: editorForm.display_label.trim(),
      description: editorForm.description.trim(),
      is_enabled: editorForm.is_enabled,
      sort_order: editorForm.sort_order,
    }
    const saved = editorForm.id
      ? await selectionConfigurationKeyApi.updateKey(editorForm.id, payload)
      : await selectionConfigurationKeyApi.createKey(payload)
    await reload()
    editorOpen.value = false
    toast.success(editorForm.id ? 'Key 已更新' : 'Key 已创建')
    editorForm.id = saved.id
  } catch (error) {
    console.error('Failed to save selection config key:', error)
    toast.error('保存 Key 失败')
  } finally {
    saving.value = false
  }
}

onMounted(async () => {
  await reload()
})
</script>
