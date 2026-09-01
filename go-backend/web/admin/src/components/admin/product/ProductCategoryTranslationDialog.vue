<template>
  <Dialog :open="open" @update:open="emit('update:open', $event)">
    <DialogContent size="lg" class="max-h-[90dvh] overflow-y-auto">
      <form @submit.prevent="emit('save', payload())">
        <DialogHeader>
          <DialogTitle>分类翻译</DialogTitle>
          <DialogDescription>
            {{ categoryName || '商品分类' }}
          </DialogDescription>
        </DialogHeader>

        <div class="mt-4 space-y-2">
          <div v-if="loading" class="flex items-center justify-center gap-2 rounded-lg border border-dashed p-8 text-sm text-muted-foreground">
            <LoaderCircle class="size-4 animate-spin" />
            加载翻译中...
          </div>

          <template v-else>
            <div class="hidden grid-cols-[minmax(150px,0.6fr)_minmax(220px,1fr)_minmax(220px,1fr)] gap-2 px-3 text-xs font-medium text-muted-foreground md:grid">
              <div>语言</div>
              <div>分类名称</div>
              <div>分类描述</div>
            </div>
            <div
              v-for="translation in rows"
              :key="translation.locale"
              class="grid gap-2 rounded-lg border p-3 md:grid-cols-[minmax(150px,0.6fr)_minmax(220px,1fr)_minmax(220px,1fr)] md:items-center"
            >
              <div class="min-w-0">
                <div class="truncate text-sm font-semibold">{{ languageLabel(translation.locale) }}</div>
                <div class="font-mono text-[10px] text-muted-foreground">{{ translation.locale }}</div>
              </div>
              <Input
                v-model="translation.name"
                :placeholder="`请输入${languageLabel(translation.locale)}分类名称`"
              />
              <Input
                v-model="translation.description"
                :placeholder="`请输入${languageLabel(translation.locale)}分类描述（可选）`"
              />
            </div>
          </template>
        </div>

        <DialogFooter class="mt-5">
          <Button type="button" variant="outline" :disabled="saving" @click="emit('update:open', false)">
            取消
          </Button>
          <Button type="submit" :disabled="loading || saving">
            <LoaderCircle v-if="saving" class="size-4 animate-spin" />
            {{ saving ? '保存中' : '保存翻译' }}
          </Button>
        </DialogFooter>
      </form>
    </DialogContent>
  </Dialog>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { LoaderCircle } from '@lucide/vue'
import type { LanguageOption } from '@/lib/languages'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import type {
  ProductCategoryTranslationForm,
  ProductCategoryTranslationPayload,
} from '@/modules/product/productCategoryTypes'
import type { ProductCategoryTranslationRecord } from '@/api/productCategories'

const props = withDefaults(defineProps<{
  open?: boolean
  categoryName?: string
  translations?: ProductCategoryTranslationRecord[]
  languageOptions?: LanguageOption[]
  loading?: boolean
  saving?: boolean
}>(), {
  open: false,
  categoryName: '',
  translations: () => [],
  languageOptions: () => [],
  loading: false,
  saving: false,
})

const emit = defineEmits<{
  (event: 'update:open', value: boolean): void
  (event: 'save', translations: ProductCategoryTranslationPayload[]): void
}>()

const rows = ref<ProductCategoryTranslationForm[]>([])

const languageLabel = (locale: string): string => (
  props.languageOptions.find((option) => option.value === locale)?.label || locale
)

const syncRows = (): void => {
  const existing = new Map(
    props.translations.map((translation) => [translation.locale, translation]),
  )
  const nextRows = props.languageOptions.map((option) => {
    const translation = existing.get(option.value)
    return {
      id: translation?.id ?? null,
      locale: option.value,
      name: translation?.name || '',
      description: translation?.description || '',
    }
  })
  const displayedLocales = new Set(nextRows.map((translation) => translation.locale))

  props.translations.forEach((translation) => {
    if (displayedLocales.has(translation.locale)) return
    nextRows.push({
      id: translation.id,
      locale: translation.locale,
      name: translation.name || '',
      description: translation.description || '',
    })
  })

  rows.value = nextRows
}

const payload = (): ProductCategoryTranslationPayload[] => rows.value
  .map((translation) => ({
    locale: translation.locale.trim(),
    name: translation.name.trim(),
    description: translation.description.trim(),
  }))
  .filter((translation) => translation.locale && translation.name)

watch(
  () => [props.open, props.translations, props.languageOptions],
  syncRows,
  { immediate: true, deep: true },
)
</script>

