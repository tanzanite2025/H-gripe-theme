<template>
  <Dialog :open="open" @update:open="emit('update:open', $event)">
    <DialogContent class="sm:max-w-xl">
      <DialogHeader>
        <DialogTitle>选择商品类型</DialogTitle>
        <DialogDescription>先选择要创建的商品类型，随后会打开对应的商品编辑窗口。</DialogDescription>
      </DialogHeader>

      <RadioGroup v-model="selectedType" class="grid gap-3 py-2">
        <label
          v-for="option in options"
          :key="option.value"
          class="flex cursor-pointer items-start gap-3 rounded-lg border p-4 transition-colors hover:border-primary/60"
          :class="selectedType === option.value ? 'border-primary bg-primary/5' : 'border-border'"
        >
          <RadioGroupItem class="mt-0.5" :value="option.value" />
          <span class="min-w-0">
            <span class="block text-sm font-semibold text-foreground">{{ option.title }}</span>
            <span class="mt-1 block text-xs leading-5 text-muted-foreground">{{ option.description }}</span>
          </span>
        </label>
      </RadioGroup>

      <DialogFooter>
        <Button type="button" variant="outline" @click="emit('update:open', false)">取消</Button>
        <Button type="button" :disabled="!selectedType" @click="handleContinueProductCreation">继续</Button>
      </DialogFooter>
    </DialogContent>
  </Dialog>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { Button } from '@/components/ui/button'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { RadioGroup, RadioGroupItem } from '@/components/ui/radio-group'

export type ProductCreationType = 'standard' | 'spoke-repair-kit'

const props = withDefaults(defineProps<{
  open?: boolean
  defaultType?: ProductCreationType
}>(), {
  open: false,
  defaultType: 'standard'
})

const emit = defineEmits<{
  (event: 'update:open', value: boolean): void
  (event: 'continue', value: ProductCreationType): void
}>()

const options: Array<{ value: ProductCreationType; title: string; description: string }> = [
  {
    value: 'standard',
    title: '普通商品',
    description: '打开现有的商品编辑窗口，按现有商品流程填写。',
  },
  {
    value: 'spoke-repair-kit',
    title: '辐条修补件',
    description: '打开专用修补件窗口，只维护商品资料、SKU、价格、库存和适配轮组型号。',
  },
]

const selectedType = ref<ProductCreationType>(props.defaultType)

watch(() => props.open, (open) => {
  if (open) selectedType.value = props.defaultType
})

const handleContinueProductCreation = (): void => {
  if (!selectedType.value) return
  emit('continue', selectedType.value)
  emit('update:open', false)
}
</script>
