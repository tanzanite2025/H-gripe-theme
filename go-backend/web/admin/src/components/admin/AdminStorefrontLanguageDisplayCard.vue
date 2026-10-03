<template>
  <Card :class="['shrink-0', props.class]">
    <CardContent class="p-3 sm:p-4">
      <div class="flex min-w-0 items-center gap-2" :aria-label="ariaLabel">
        <div
          ref="languageScrollArea"
          class="min-w-0 flex-1 overflow-x-auto [scrollbar-width:none] [-ms-overflow-style:none] [&::-webkit-scrollbar]:hidden"
        >
          <Tabs
            :model-value="modelValue"
            class="w-max min-w-full gap-0"
            @update:model-value="selectStorefrontLanguage"
          >
            <TabsList
              variant="default"
              class="h-11 w-max min-w-full max-w-none flex-nowrap justify-start gap-2 rounded-2xl bg-muted/50 p-1.5"
            >
              <TabsTrigger
                v-for="(language, index) in languageOptions"
                :key="language.value"
                :ref="(element) => setStorefrontLanguageTabTriggerElement(language.value, element)"
                :value="language.value"
                :disabled="disabled || loading"
                class="h-8 flex-none gap-1.5 px-3.5 text-xs font-bold normal-case tracking-normal"
              >
                <span class="font-mono text-[11px] opacity-60">{{ formatStorefrontLanguageTabNumber(index) }}</span>
                <span>{{ language.label }}</span>
              </TabsTrigger>
            </TabsList>
          </Tabs>
        </div>

        <DropdownMenu v-if="hasStorefrontLanguageTabOverflow">
          <DropdownMenuTrigger as-child>
            <Button
              type="button"
              variant="outline"
              size="icon"
              class="size-9 shrink-0 rounded-full"
              :disabled="disabled || loading"
              aria-label="选择更多语言"
              title="更多语言"
            >
              <Ellipsis class="size-4" />
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end" class="max-h-80 w-64">
            <DropdownMenuLabel>全部语言</DropdownMenuLabel>
            <DropdownMenuItem
              v-for="(language, index) in languageOptions"
              :key="language.value"
              class="gap-2"
              :disabled="disabled || loading"
              @select="selectStorefrontLanguage(language.value)"
            >
              <span class="w-5 shrink-0 text-center font-mono text-[10px] text-muted-foreground">
                {{ formatStorefrontLanguageTabNumber(index) }}
              </span>
              <span class="min-w-0 flex-1 truncate">{{ language.label }}</span>
              <Check
                v-if="language.value === modelValue"
                class="size-3.5 shrink-0 text-primary"
              />
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      </div>
    </CardContent>
  </Card>
</template>

<script setup lang="ts">
import { nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import type { ComponentPublicInstance, HTMLAttributes } from 'vue'
import { Check, Ellipsis } from '@lucide/vue'
import type { LanguageOption } from '@/lib/languages'
import { Button } from '@/components/ui/button'
import { Card, CardContent } from '@/components/ui/card'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { Tabs, TabsList, TabsTrigger } from '@/components/ui/tabs'

type StorefrontLanguageTabTriggerRef = Element | ComponentPublicInstance | null

const props = withDefaults(defineProps<{
  modelValue: string
  languageOptions: LanguageOption[]
  disabled?: boolean
  loading?: boolean
  ariaLabel?: string
  class?: HTMLAttributes['class']
}>(), {
  modelValue: '',
  disabled: false,
  loading: false,
  ariaLabel: '前台内容语言',
  class: undefined,
})

const emit = defineEmits<{
  (event: 'update:modelValue', value: string): void
}>()

const languageScrollArea = ref<HTMLElement | null>(null)
const languageTabTriggerElements = new Map<string, HTMLElement>()
const hasStorefrontLanguageTabOverflow = ref(false)
let languageResizeObserver: ResizeObserver | null = null

const formatStorefrontLanguageTabNumber = (index: number): string => String(index + 1).padStart(2, '0')

const setStorefrontLanguageTabTriggerElement = (
  languageCode: string,
  element: StorefrontLanguageTabTriggerRef,
): void => {
  const target = element && '$el' in element ? element.$el : element
  if (target instanceof HTMLElement) languageTabTriggerElements.set(languageCode, target)
  else languageTabTriggerElements.delete(languageCode)
}

const updateStorefrontLanguageDisplayCardOverflow = (): void => {
  const scrollArea = languageScrollArea.value
  hasStorefrontLanguageTabOverflow.value = Boolean(
    scrollArea && scrollArea.scrollWidth > scrollArea.clientWidth + 1,
  )
}

const centerActiveStorefrontLanguageTab = async (
  languageCode = props.modelValue,
  behavior: ScrollBehavior = 'smooth',
): Promise<void> => {
  await nextTick()
  const scrollArea = languageScrollArea.value
  const target = languageTabTriggerElements.get(languageCode)
  if (!scrollArea || !target) return

  const scrollAreaRect = scrollArea.getBoundingClientRect()
  const targetRect = target.getBoundingClientRect()
  const desiredLeft = scrollArea.scrollLeft
    + targetRect.left
    - scrollAreaRect.left
    - (scrollArea.clientWidth - targetRect.width) / 2
  const maxLeft = Math.max(0, scrollArea.scrollWidth - scrollArea.clientWidth)

  scrollArea.scrollTo({
    left: Math.max(0, Math.min(desiredLeft, maxLeft)),
    behavior,
  })
}

const selectStorefrontLanguage = (languageCode: unknown): void => {
  if (props.disabled || props.loading || typeof languageCode !== 'string') return
  emit('update:modelValue', languageCode)
}

watch(
  () => [props.languageOptions, props.modelValue],
  async () => {
    await nextTick()
    updateStorefrontLanguageDisplayCardOverflow()
    await centerActiveStorefrontLanguageTab(props.modelValue)
  },
  { deep: true, immediate: true },
)

onMounted(() => {
  languageResizeObserver = new ResizeObserver(() => {
    updateStorefrontLanguageDisplayCardOverflow()
    void centerActiveStorefrontLanguageTab(props.modelValue, 'auto')
  })
  if (languageScrollArea.value) languageResizeObserver.observe(languageScrollArea.value)
  updateStorefrontLanguageDisplayCardOverflow()
  void centerActiveStorefrontLanguageTab(props.modelValue, 'auto')
})

onBeforeUnmount(() => {
  languageResizeObserver?.disconnect()
  languageTabTriggerElements.clear()
})
</script>
