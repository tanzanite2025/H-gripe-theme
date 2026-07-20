<template>
  <div class="space-y-3">
    <Alert>
      <Info class="size-4" />
      <AlertTitle>价格与库存按 SKU 维护</AlertTitle>
      <AlertDescription>
        每个 SKU 都可以单独维护重量、价格和库存，前台会按当前选中的 SKU 显示。
      </AlertDescription>
    </Alert>

    <RadioGroup
      :model-value="String(defaultIndex)"
      @update:model-value="emit('set-default', Number($event))"
    >
      <Table class="min-w-[980px]">
        <TableHeader>
          <TableRow>
            <TableHead class="w-16 text-center">默认</TableHead>
            <TableHead class="min-w-40">SKU</TableHead>
            <TableHead v-for="spec in specDefinitions" :key="spec.id" class="min-w-36">
              {{ specLabel(spec) }}
            </TableHead>
            <TableHead class="w-32">价格</TableHead>
            <TableHead class="w-32">促销价</TableHead>
            <TableHead class="w-28">重量（克）</TableHead>
            <TableHead class="w-24">库存</TableHead>
            <TableHead class="w-20 text-center">启用</TableHead>
            <TableHead class="w-16 text-right">操作</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          <TableRow v-for="(variant, index) in variants" :key="variant.id || `variant-${index}`">
            <TableCell class="text-center">
              <RadioGroupItem :value="String(index)" :aria-label="`设为默认变体 ${index + 1}`" />
            </TableCell>
            <TableCell>
              <Input v-model="variant.sku" placeholder="变体 SKU" />
            </TableCell>
            <TableCell v-for="spec in specDefinitions" :key="spec.id">
              <Input
                v-if="spec.field_type === 'number'"
                v-model.number="variant.option_values[spec.slug]"
                type="number"
                min="0"
              />
              <Select
                v-else-if="spec.field_type === 'select'"
                :model-value="selectValue(variant.option_values[spec.slug])"
                @update:model-value="setSelectValue(variant, spec.slug, $event)"
              >
                <SelectTrigger class="w-full">
                  <SelectValue placeholder="请选择" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="__empty__">未设置</SelectItem>
                  <SelectItem v-for="option in specOptions(spec)" :key="String(option)" :value="String(option)">
                    {{ formatOption(option) }}
                  </SelectItem>
                </SelectContent>
              </Select>
              <div v-else-if="spec.field_type === 'boolean'" class="flex h-8 items-center">
                <Switch v-model="variant.option_values[spec.slug]" :aria-label="spec.name" />
              </div>
              <Input v-else v-model="variant.option_values[spec.slug]" :placeholder="spec.name" />
            </TableCell>
            <TableCell>
              <Input v-model.number="variant.price" type="number" min="0" step="0.01" />
            </TableCell>
            <TableCell>
              <Input v-model.number="variant.sale_price" type="number" min="0" step="0.01" placeholder="可选" />
            </TableCell>
            <TableCell>
              <Input v-model.number="variant.weight_grams" type="number" min="0" step="1" placeholder="克" />
            </TableCell>
            <TableCell>
              <Input v-model.number="variant.stock" type="number" min="0" step="1" />
            </TableCell>
            <TableCell class="text-center">
              <Switch v-model="variant.is_active" :aria-label="`启用变体 ${variant.sku || index + 1}`" />
            </TableCell>
            <TableCell class="text-right">
              <Tooltip>
                <TooltipTrigger as-child>
                  <Button
                    type="button"
                    variant="ghost"
                    size="icon"
                    class="text-destructive hover:text-destructive"
                    :disabled="variants.length <= 1"
                    :aria-label="`删除变体 ${variant.sku || index + 1}`"
                    @click="emit('remove', index)"
                  >
                    <Trash2 class="size-4" />
                  </Button>
                </TooltipTrigger>
                <TooltipContent>删除变体</TooltipContent>
              </Tooltip>
            </TableCell>
          </TableRow>
        </TableBody>
      </Table>
    </RadioGroup>

    <Button type="button" variant="outline" size="sm" @click="emit('add')">
      <Plus class="size-3.5" />
      添加变体
    </Button>
  </div>
</template>

<script setup>
import { Info, Plus, Trash2 } from '@lucide/vue'
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { RadioGroup, RadioGroupItem } from '@/components/ui/radio-group'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Switch } from '@/components/ui/switch'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Tooltip, TooltipContent, TooltipTrigger } from '@/components/ui/tooltip'

defineProps({
  variants: {
    type: Array,
    required: true
  },
  specDefinitions: {
    type: Array,
    default: () => []
  },
  defaultIndex: {
    type: Number,
    default: 0
  }
})

const emit = defineEmits(['add', 'remove', 'set-default'])

const specOptions = (spec) => {
  if (!spec?.options) return []
  try {
    const options = JSON.parse(spec.options)
    return Array.isArray(options) ? options : []
  } catch {
    return []
  }
}

const specLabel = (spec) => spec.unit ? `${spec.name} (${spec.unit})` : spec.name
const formatOption = (option) => String(option).replace(/_/g, ' ')
const selectValue = (value) => value === undefined || value === null || value === '' ? '__empty__' : String(value)

const setSelectValue = (variant, slug, value) => {
  variant.option_values[slug] = value === '__empty__' ? '' : value
}
</script>
