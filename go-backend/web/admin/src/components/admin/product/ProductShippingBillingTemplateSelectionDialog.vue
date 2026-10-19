<template>
  <Dialog :open="selectionDialogOpen" @update:open="updateShippingTemplateSelectionDialogOpen">
    <DialogTrigger as-child>
      <Button
        type="button"
        variant="outline"
        class="h-auto min-h-9 w-full min-w-0 justify-between gap-2 whitespace-normal px-3 py-2 text-left font-normal"
        :aria-label="`${selectionContext}：选择线路计费模板`"
      >
        <span class="min-w-0 break-words">{{ selectedTemplateSummary || '选择线路计费模板' }}</span>
        <ChevronDown class="size-4 shrink-0 text-muted-foreground" />
      </Button>
    </DialogTrigger>
    <DialogContent size="md" class="!flex flex-col gap-4" data-shipping-template-selection-dialog>
      <DialogHeader class="pr-8">
        <DialogTitle>选择线路计费模板</DialogTitle>
        <DialogDescription>{{ selectionContext }} · 可选择一个或多个模板，确认后生效。</DialogDescription>
      </DialogHeader>
      <Input
        v-model="templateSearchKeyword"
        placeholder="搜索模板名称或 ID"
        aria-label="搜索线路计费模板"
        class="shrink-0"
        @keydown.enter.prevent
      />
      <div class="min-h-0 space-y-2 overflow-y-auto overscroll-contain" role="group" aria-label="可选线路计费模板">
        <label
          v-for="template in filteredShippingTemplateChoices"
          :key="template.id"
          class="flex cursor-pointer items-start gap-3 rounded-lg border px-3 py-3 hover:bg-muted/40"
          :class="draftSelectedTemplateIDs.includes(template.id) ? 'border-primary bg-primary/5' : ''"
        >
          <input
            v-model="draftSelectedTemplateIDs"
            type="checkbox"
            :value="template.id"
            class="mt-0.5 size-4 shrink-0 accent-primary"
          >
          <span class="min-w-0 break-words">
            <span class="block text-sm font-medium">{{ template.name }}</span>
            <span class="mt-1 block text-xs text-muted-foreground">
              #{{ template.id }}{{ template.unavailable ? ' · 不在当前模板列表中' : template.enabled === false ? ' · 已停用' : '' }}
            </span>
          </span>
        </label>
        <p v-if="filteredShippingTemplateChoices.length === 0" class="py-8 text-center text-sm text-muted-foreground">
          {{ templateSearchKeyword.trim() ? '没有匹配的线路计费模板' : '暂无线路计费模板，请先在物流设置中创建。' }}
        </p>
      </div>
      <div class="flex shrink-0 items-center justify-between gap-3 border-t pt-3">
        <span class="text-xs text-muted-foreground" aria-live="polite">已选 {{ draftSelectedTemplateIDs.length }} 项</span>
        <Button type="button" variant="ghost" size="sm" :disabled="draftSelectedTemplateIDs.length === 0" @click="draftSelectedTemplateIDs = []">清空选择</Button>
      </div>
      <DialogFooter class="shrink-0">
        <Button type="button" variant="outline" @click="selectionDialogOpen = false">取消</Button>
        <Button type="button" @click="confirmShippingTemplateSelection">确认选择</Button>
      </DialogFooter>
    </DialogContent>
  </Dialog>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { ChevronDown } from '@lucide/vue'
import { Button } from '@/components/ui/button'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle, DialogTrigger } from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'
import type { ShippingTemplateRecord } from '@/modules/product/productEditorTypes'

const props = withDefaults(defineProps<{
  modelValue?: Array<number | string>
  shippingTemplates?: ShippingTemplateRecord[]
  selectionContext?: string
}>(), {
  modelValue: () => [],
  shippingTemplates: () => [],
  selectionContext: 'SKU',
})

const emit = defineEmits<{ (event: 'update:modelValue', value: number[]): void }>()
const selectionDialogOpen = ref(false)
const templateSearchKeyword = ref('')
const draftSelectedTemplateIDs = ref<number[]>([])

const normalizeShippingTemplateSelectionIDs = (values: Array<number | string>): number[] =>
  Array.from(new Set(values.map(Number).filter(value => Number.isSafeInteger(value) && value > 0)))

const selectedTemplateIDs = computed(() => normalizeShippingTemplateSelectionIDs(props.modelValue))
const shippingTemplateChoices = computed(() => {
  const choices = new Map<number, { id: number; name: string; enabled?: boolean; unavailable?: boolean }>()
  for (const template of props.shippingTemplates) {
    const id = Number(template.id)
    if (Number.isSafeInteger(id) && id > 0) choices.set(id, { ...template, id })
  }
  // Keep historical selections visible and removable even when the catalog omits them.
  for (const id of new Set([...selectedTemplateIDs.value, ...draftSelectedTemplateIDs.value])) {
    if (!choices.has(id)) choices.set(id, { id, name: `模板 #${id}`, unavailable: true })
  }
  return Array.from(choices.values())
})
const selectedTemplateSummary = computed(() => selectedTemplateIDs.value.map(id => {
  const template = shippingTemplateChoices.value.find(choice => choice.id === id)
  return `${template?.name || `模板 #${id}`}${template?.unavailable ? '（未在列表中）' : template?.enabled === false ? '（停用）' : ''}`
}).join('、'))
const filteredShippingTemplateChoices = computed(() => {
  const keyword = templateSearchKeyword.value.trim().toLocaleLowerCase()
  return shippingTemplateChoices.value.filter(template => `${template.name} ${template.id}`.toLocaleLowerCase().includes(keyword))
})

const updateShippingTemplateSelectionDialogOpen = (open: boolean): void => {
  if (open) {
    draftSelectedTemplateIDs.value = [...selectedTemplateIDs.value]
    templateSearchKeyword.value = ''
  }
  selectionDialogOpen.value = open
}
const confirmShippingTemplateSelection = (): void => {
  emit('update:modelValue', normalizeShippingTemplateSelectionIDs(draftSelectedTemplateIDs.value))
  selectionDialogOpen.value = false
}
</script>
