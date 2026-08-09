<template>
  <Dialog :open="open" @update:open="emit('update:open', $event)">
    <DialogContent size="xl" class="gap-0 p-0" @open-auto-focus.prevent>
      <DialogHeader class="border-b px-5 py-4 pr-12">
        <DialogTitle>引用 FAQ</DialogTitle>
        <DialogDescription>选择后台已发布 FAQ，自动写入客服自动回复的结构化 FAQ 卡片。</DialogDescription>
      </DialogHeader>

      <div class="flex min-h-[30rem] max-h-[72dvh] flex-col overflow-hidden">
        <div class="flex shrink-0 flex-wrap items-center justify-between gap-3 border-b px-4 py-3">
          <div class="min-w-0">
            <p class="text-[10px] font-black uppercase tracking-wider text-muted-foreground">
              当前语言：{{ normalizedLocale }}
            </p>
            <p class="mt-1 text-xs font-medium text-muted-foreground">
              只显示 published FAQ，避免引用草稿内容。
            </p>
          </div>
          <div class="flex min-w-0 items-center gap-2">
            <label class="relative block min-w-[16rem]">
              <Search class="pointer-events-none absolute left-3 top-1/2 size-3.5 -translate-y-1/2 text-muted-foreground" />
              <Input
                v-model="searchQuery"
                class="h-9 pl-8 text-xs font-bold"
                placeholder="搜索问题、答案、页面或分类"
              />
            </label>
            <Button variant="outline" size="sm" :disabled="loading" @click="loadFAQs">
              <RefreshCw :class="['size-3.5', { 'animate-spin': loading }]" />
              刷新
            </Button>
          </div>
        </div>

        <div v-if="loading" class="flex min-h-0 flex-1 items-center justify-center text-xs text-muted-foreground">
          正在加载 FAQ
        </div>

        <div v-else-if="filteredPages.length === 0" class="flex min-h-0 flex-1 flex-col items-center justify-center gap-2 text-muted-foreground">
          <HelpCircle class="size-8 opacity-50" />
          <span class="text-xs">没有找到可引用的已发布 FAQ</span>
        </div>

        <div v-else class="min-h-0 flex-1 overflow-y-auto p-4">
          <section
            v-for="page in filteredPages"
            :key="`${page.page_id}-${page.locale}`"
            class="mb-4 overflow-hidden rounded-2xl border border-border bg-background last:mb-0"
          >
            <div class="border-b bg-muted/30 px-4 py-3">
              <div class="flex flex-wrap items-center justify-between gap-2">
                <div class="min-w-0">
                  <h3 class="truncate text-sm font-black">{{ page.title || page.page_id }}</h3>
                  <p class="mt-1 truncate font-mono text-[11px] text-muted-foreground">
                    {{ page.route_path || page.page_id }}
                  </p>
                </div>
                <span class="rounded-full bg-muted px-2.5 py-1 text-[10px] font-black text-muted-foreground">
                  {{ pageFaqCount(page) }} FAQ
                </span>
              </div>
            </div>

            <div class="divide-y divide-border/70">
              <section
                v-for="category in page.categories"
                :key="`${page.page_id}-${category.category_key}`"
                class="px-4 py-3"
              >
                <div class="mb-2 flex items-center gap-2">
                  <span class="rounded-full border border-border bg-muted/50 px-2.5 py-1 text-[10px] font-black uppercase tracking-wider text-muted-foreground">
                    {{ category.name || category.category_key }}
                  </span>
                  <span class="text-[10px] font-bold text-muted-foreground">{{ category.faqs?.length || 0 }} 条</span>
                </div>

                <div class="grid gap-2 lg:grid-cols-2">
                  <button
                    v-for="faq in category.faqs"
                    :key="faq.id"
                    type="button"
                    class="group flex min-w-0 gap-3 rounded-xl border p-3 text-left transition hover:-translate-y-0.5 hover:border-[var(--admin-selected)] hover:shadow-sm focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
                    :class="isSelected(faq) ? 'border-[var(--admin-selected)] bg-[var(--admin-selected)]/5' : 'border-border bg-card'"
                    @click="selectFAQ(page, category, faq)"
                  >
                    <span class="mt-0.5 flex size-8 shrink-0 items-center justify-center rounded-full bg-muted text-muted-foreground group-hover:text-foreground">
                      <Check v-if="isSelected(faq)" class="size-4 text-[var(--admin-selected)]" />
                      <HelpCircle v-else class="size-4" />
                    </span>
                    <span class="min-w-0 flex-1">
                      <span class="line-clamp-2 break-words text-xs font-black leading-5 text-foreground">
                        {{ faq.question }}
                      </span>
                      <span class="mt-1 line-clamp-2 break-words text-[11px] leading-5 text-muted-foreground">
                        {{ plainText(faq.answer) || '-' }}
                      </span>
                      <span v-if="faq.answer_image_url" class="mt-2 inline-flex items-center gap-1 text-[10px] font-black text-sky-600">
                        <ImageIcon class="size-3" />
                        含 FAQ 图片
                      </span>
                    </span>
                  </button>
                </div>
              </section>
            </div>
          </section>
        </div>
      </div>
    </DialogContent>
  </Dialog>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { Check, HelpCircle, Image as ImageIcon, RefreshCw, Search } from '@lucide/vue'
import { Button } from '@/components/ui/button'
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'
import customerServiceApi from '@/api/customerService'
import type { FAQCategory, FAQItem, FAQPage, FAQSelection } from './customerServiceTypes'

const props = withDefaults(defineProps<{
  open?: boolean
  locale?: string
  selectedFaqId?: string | number
}>(), {
  open: false,
  locale: '',
  selectedFaqId: '',
})

const emit = defineEmits<{
  (event: 'update:open', value: boolean): void
  (event: 'select', selection: FAQSelection): void
}>()

const loading = ref(false)
const pages = ref<FAQPage[]>([])
const searchQuery = ref('')
const loadedLocale = ref('')

const normalizedLocale = computed(() => {
  const locale = String(props.locale || '').trim().toLowerCase().replace(/-/g, '_')
  return locale && locale !== '*' ? locale : ''
})

const normalizeText = (value: unknown): string => String(value || '').trim().toLowerCase()

const plainText = (value: unknown): string => String(value || '')
  .replace(/<[^>]+>/g, ' ')
  .replace(/&nbsp;/g, ' ')
  .replace(/\s+/g, ' ')
  .trim()

const pageFaqCount = (page: FAQPage): number => {
  return (page.categories || []).reduce((total, category) => total + (category.faqs?.length || 0), 0)
}

const filteredPages = computed(() => {
  const query = normalizeText(searchQuery.value)
  return pages.value
    .map((page) => {
      const pageText = normalizeText(`${page.title} ${page.page_id} ${page.route_path}`)
      const categories = (page.categories || [])
        .map((category) => {
          const categoryText = normalizeText(`${category.name} ${category.category_key}`)
          const faqs = (category.faqs || []).filter((faq) => {
            if (!query) return true
            const faqText = normalizeText(`${faq.question} ${plainText(faq.answer)}`)
            return pageText.includes(query) || categoryText.includes(query) || faqText.includes(query)
          })
          return { ...category, faqs }
        })
        .filter((category) => category.faqs.length > 0)
      return { ...page, categories }
    })
    .filter((page) => page.categories.length > 0)
})

const isSelected = (faq: FAQItem): boolean => {
  return props.selectedFaqId && String(props.selectedFaqId) === String(faq?.id)
}

const loadFAQs = async () => {
  if (!normalizedLocale.value) {
    pages.value = []
    loadedLocale.value = ''
    return
  }

  loading.value = true
  try {
    pages.value = await customerServiceApi.listAutoReplyFAQGroups({
      locale: normalizedLocale.value,
    })
    loadedLocale.value = normalizedLocale.value
  } catch (error) {
    console.error('Failed to load FAQs for auto reply picker:', error)
    pages.value = []
  } finally {
    loading.value = false
  }
}

const selectFAQ = (page: FAQPage, category: FAQCategory, faq: FAQItem): void => {
  emit('select', { page, category, faq })
  emit('update:open', false)
}

watch(
  () => props.open,
  (isOpen) => {
    if (isOpen && loadedLocale.value !== normalizedLocale.value) {
      loadFAQs()
    }
  }
)

watch(normalizedLocale, () => {
  if (props.open) loadFAQs()
})
</script>
