<template>
  <AdminTablePanel class="h-full min-h-0" :loading="loading || structureLoading" :batch-visible="selectedFaqs.length > 0" scroll-body>
    <template #batch>
      <div class="flex flex-wrap items-center justify-between gap-2">
        <span class="text-xs font-medium">已选择 {{ selectedFaqs.length }} 个 FAQ</span>
        <Button v-if="hasPermission('faq:delete')" variant="destructive" size="sm" @click="$emit('batch-delete')">
          <Trash2 class="size-3.5" />
          批量删除
        </Button>
      </div>
    </template>

    <div class="border-b border-border/70 px-4 py-3">
      <div class="flex min-w-0 items-center gap-2">
        <div
          ref="localeScrollArea"
          class="min-w-0 flex-1 overflow-x-auto [scrollbar-width:none] [-ms-overflow-style:none] [&::-webkit-scrollbar]:hidden"
        >
          <Tabs
            :model-value="activeStructureLocale"
            class="w-max min-w-full gap-0"
            @update:model-value="selectLocale"
          >
            <TabsList
              variant="default"
              class="h-11 w-max min-w-full max-w-none flex-nowrap justify-start gap-2 rounded-2xl bg-muted/50 p-1.5"
            >
              <TabsTrigger
                v-for="(locale, index) in structureLocales"
                :key="locale.value"
                :ref="(element) => setLocaleTriggerRef(locale.value, element)"
                :value="locale.value"
          class="h-8 flex-none gap-1.5 px-3.5 text-xs font-bold normal-case tracking-normal"
              >
                <span class="font-mono text-[11px] opacity-60">{{ localeNumber(index) }}</span>
                <span>{{ locale.label }}</span>
              </TabsTrigger>
            </TabsList>
          </Tabs>
        </div>

        <DropdownMenu v-if="hasLocaleOverflow">
          <DropdownMenuTrigger as-child>
            <Button
              type="button"
              variant="outline"
              size="icon"
              class="size-9 shrink-0 rounded-full"
              aria-label="选择更多语言"
              title="更多语言"
            >
              <Ellipsis class="size-4" />
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end" class="max-h-80 w-64">
            <DropdownMenuLabel>全部语言</DropdownMenuLabel>
            <DropdownMenuItem
              v-for="(locale, index) in structureLocales"
              :key="locale.value"
              class="gap-2"
              @select="selectLocale(locale.value)"
            >
              <span class="w-5 shrink-0 text-center font-mono text-[10px] text-muted-foreground">
                {{ localeNumber(index) }}
              </span>
              <span class="min-w-0 flex-1 truncate">{{ locale.label }}</span>
              <Check
                v-if="locale.value === activeStructureLocale"
                class="size-3.5 shrink-0 text-primary"
              />
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      </div>
    </div>

    <div v-if="!loading && !structureLoading && faqGroups.length === 0" class="p-10 text-center text-sm text-muted-foreground">
      当前语言暂无 FAQ 页面。
    </div>

    <div v-else class="divide-y divide-border/70">
      <section v-for="page in faqGroups" :key="pageKey(page)" class="bg-card">
        <div class="flex items-center gap-2 px-4 py-3">
          <button
            type="button"
            class="flex min-w-0 flex-1 items-center gap-3 text-left"
            :aria-expanded="isPageExpanded(pageKey(page))"
            @click="togglePage(pageKey(page))"
          >
            <ChevronDown
              class="size-4 shrink-0 text-muted-foreground transition-transform"
              :class="{ 'rotate-180': isPageExpanded(pageKey(page)) }"
            />
            <span class="flex size-9 shrink-0 items-center justify-center rounded-full bg-primary/10 text-xs font-black text-primary">
              {{ pageFAQCount(page) }}
            </span>
            <span class="min-w-0">
              <span class="mb-1 flex flex-wrap items-center gap-2">
                <AdminStatusBadge tone="blue">{{ domainName(page.domain) }}</AdminStatusBadge>
                <AdminStatusBadge :tone="visibilityTone(page.status)">{{ visibilityName(page.status) }}</AdminStatusBadge>
              </span>
              <span class="block truncate text-sm font-black">{{ page.title || page.page_id }}</span>
              <span class="mt-1 block truncate font-mono text-[11px] text-muted-foreground">
                {{ page.route_path || page.page_id }}
              </span>
            </span>
          </button>

          <span class="hidden shrink-0 text-[11px] font-medium text-muted-foreground sm:inline">
            {{ page.categories?.length || 0 }} 个分类 · {{ pageFAQCount(page) }} 个 FAQ
          </span>
          <Button
            v-if="hasPermission('faq:create')"
            type="button"
            variant="ghost"
            size="sm"
            class="h-8 shrink-0 px-2 text-xs font-bold"
            :aria-label="`添加分类到 ${page.title || page.page_id}`"
            @click.stop="$emit('create-category', page)"
          >
            <FolderPlus class="size-3.5" />
            <span class="hidden sm:inline">添加分类</span>
          </Button>
          <Button
            v-if="hasPermission('faq:edit')"
            type="button"
            variant="ghost"
            size="icon"
            class="size-8 shrink-0"
            :aria-label="`编辑页面 ${page.title || page.page_id}`"
            @click.stop="$emit('edit-page', page)"
          >
            <Pencil class="size-3.5" />
          </Button>
        </div>

        <div v-if="isPageExpanded(pageKey(page))" class="border-t border-border/70 bg-muted/10 px-3 pb-3">
          <section
            v-for="category in page.categories || []"
            :key="category.id"
            class="mt-3 overflow-hidden rounded-xl border border-border/70 bg-background/70"
          >
            <div class="flex items-center gap-2 px-3 py-2.5">
              <span class="flex size-8 shrink-0 items-center justify-center rounded-full bg-muted text-base">
                {{ category.icon || 'FAQ' }}
              </span>
              <div class="min-w-0 flex-1">
                <div class="flex flex-wrap items-center gap-2">
                  <p class="truncate text-xs font-black">{{ category.name || category.category_key }}</p>
                  <AdminStatusBadge :tone="visibilityTone(category.status)">{{ visibilityName(category.status) }}</AdminStatusBadge>
                </div>
                <p class="mt-0.5 truncate font-mono text-[11px] text-muted-foreground">
                  {{ category.category_key }} · {{ category.faqs?.length || 0 }} 个 FAQ
                </p>
              </div>
              <Button
                v-if="hasPermission('faq:create') && category.status !== 'hidden'"
                type="button"
                variant="ghost"
                size="icon"
                class="size-8 shrink-0"
                :aria-label="`添加 FAQ 到分类 ${category.name || category.category_key}`"
                @click="$emit('create-faq', page, category)"
              >
                <Plus class="size-3.5" />
              </Button>
              <Button
                v-if="hasPermission('faq:edit')"
                type="button"
                variant="ghost"
                size="icon"
                class="size-8 shrink-0"
                :aria-label="`编辑分类 ${category.name || category.category_key}`"
                @click="$emit('edit-category', page, category)"
              >
                <Pencil class="size-3.5" />
              </Button>
              <Button
                v-if="hasPermission('faq:delete')"
                type="button"
                variant="ghost"
                size="icon"
                class="size-8 shrink-0 text-destructive hover:text-destructive"
                :aria-label="`删除分类 ${category.name || category.category_key}`"
                @click="$emit('delete-category', category)"
              >
                <Trash2 class="size-3.5" />
              </Button>
            </div>

            <div v-if="category.faqs?.length" class="border-t border-border/70">
              <article
                v-for="faq in category.faqs"
                :key="faq.id"
                class="grid gap-3 border-b border-border/60 px-3 py-3 last:border-b-0 lg:grid-cols-[auto_minmax(0,1fr)_minmax(18rem,0.9fr)_auto] lg:items-start"
              >
                <Checkbox
                  :model-value="isSelected(faq.id)"
                  :aria-label="`选择 FAQ ${faq.question}`"
                  class="mt-1"
                  @update:model-value="$emit('toggle-faq', faq, $event)"
                />

                <Tooltip>
                  <TooltipTrigger as-child>
                    <div tabindex="0" class="min-w-0 cursor-help rounded-sm text-left focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-primary/35">
                      <p class="line-clamp-2 break-words text-xs font-bold leading-5">{{ faq.question }}</p>
                      <p class="mt-1 font-mono text-[10px] text-muted-foreground">#{{ faq.id }}</p>
                    </div>
                  </TooltipTrigger>
                  <TooltipContent
                    side="top"
                    align="start"
                    class="max-w-[34rem] items-start whitespace-normal rounded-lg px-3 py-2 text-left font-sans text-xs font-medium normal-case leading-5 tracking-normal"
                  >
                    {{ faq.question }}
                  </TooltipContent>
                </Tooltip>

                <Tooltip>
                  <TooltipTrigger as-child>
                    <div tabindex="0" class="min-w-0 cursor-help rounded-sm text-left focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-primary/35">
                      <p class="line-clamp-2 break-words text-xs leading-5 text-muted-foreground">
                        {{ plainText(faq.answer) || '-' }}
                      </p>
                      <p v-if="faq.answer_image_url" class="mt-1 text-[10px] font-bold text-sky-600">含 FAQ 图</p>
                    </div>
                  </TooltipTrigger>
                  <TooltipContent
                    side="top"
                    align="start"
                    class="max-h-72 max-w-[42rem] items-start overflow-y-auto whitespace-normal rounded-lg px-3 py-2 text-left font-sans text-xs font-medium normal-case leading-5 tracking-normal"
                  >
                    {{ plainText(faq.answer) || '-' }}
                  </TooltipContent>
                </Tooltip>

                <div class="flex items-center justify-between gap-2 lg:justify-end">
                  <AdminStatusBadge :tone="statusTone(faq.status)">{{ statusName(faq.status) }}</AdminStatusBadge>
                  <div class="flex items-center gap-1.5">
                    <Button
                      v-if="hasPermission('faq:edit')"
                      type="button"
                      variant="outline"
                      size="sm"
                      class="h-8 px-2.5 text-xs"
                      :aria-label="`编辑 FAQ ${faq.question}`"
                      @click="$emit('edit', faq)"
                    >
                      <Pencil class="size-3.5" />
                      编辑
                    </Button>
                    <Button
                      v-if="hasPermission('faq:delete')"
                      type="button"
                      variant="ghost"
                      size="icon"
                      class="size-8 text-destructive hover:text-destructive"
                      :aria-label="`删除 FAQ ${faq.question}`"
                      @click="$emit('delete', faq)"
                    >
                      <Trash2 class="size-4" />
                    </Button>
                  </div>
                </div>
              </article>
            </div>
            <p v-else class="border-t border-dashed border-border/70 px-3 py-4 text-center text-xs text-muted-foreground">
              当前分类暂无 FAQ
            </p>
          </section>

          <div
            v-if="!page.categories?.length"
            class="mt-3 flex flex-col items-center justify-center gap-3 rounded-xl border border-dashed border-border/80 bg-background/70 px-4 py-8 text-center"
          >
            <p class="text-xs font-medium text-muted-foreground">当前页面暂无 FAQ 分类</p>
            <Button
              v-if="hasPermission('faq:create')"
              type="button"
              size="sm"
              class="h-8 text-xs font-bold"
              @click="$emit('create-category', page)"
            >
              <FolderPlus class="size-3.5" />
              添加分类
            </Button>
          </div>
        </div>
      </section>
    </div>

    <template #footer>
      <div class="flex items-center justify-between gap-3 text-[10px] font-mono font-bold uppercase tracking-wider text-muted-foreground/70">
        <span>当前语言 FAQ：{{ pagination.total }}</span>
      </div>
    </template>
  </AdminTablePanel>
</template>

<script setup lang="ts">
import { nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import type { ComponentPublicInstance } from 'vue'
import { Check, ChevronDown, Ellipsis, FolderPlus, Pencil, Plus, Trash2 } from '@lucide/vue'
import type { LanguageOption } from '@/lib/languages'
import type { FAQCategory, FAQID, FAQItemLike, FAQStatusTone, FAQStructurePage } from '@/lib/faqAdminPresentation'
import AdminStatusBadge from '@/components/admin/AdminStatusBadge.vue'
import AdminTablePanel from '@/components/admin/AdminTablePanel.vue'
import { Button } from '@/components/ui/button'
import { Checkbox } from '@/components/ui/checkbox'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuTrigger
} from '@/components/ui/dropdown-menu'
import { Tabs, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Tooltip, TooltipContent, TooltipTrigger } from '@/components/ui/tooltip'
import type { FAQPagination } from '@/composables/faq/useFaqList'

type FAQSelectionState = boolean | string
type TemplateRefTarget = Element | ComponentPublicInstance | null

const props = withDefaults(defineProps<{
  loading?: boolean
  structureLoading?: boolean
  faqGroups: FAQStructurePage[]
  selectedFaqs: FAQItemLike[]
  pagination: FAQPagination
  structureLocales: LanguageOption[]
  activeStructureLocale: string
  hasPermission: (permission: string) => boolean
  isSelected: (faqID?: FAQID | null) => boolean
  plainText: (value?: string | null) => string
  statusTone: (status?: string | null) => FAQStatusTone
  statusName: (status?: string | null) => string
  visibilityName: (status?: string | null) => string
  visibilityTone: (status?: string | null) => FAQStatusTone
  domainName: (domain?: string | null) => string
}>(), {
  loading: false,
  structureLoading: false
})

const expandedPages = ref<Set<string>>(new Set())
const localeScrollArea = ref<HTMLElement | null>(null)
const localeTriggerRefs = new Map<string, HTMLElement>()
const hasLocaleOverflow = ref(false)
let localeResizeObserver: ResizeObserver | null = null

const pageKey = (page: FAQStructurePage): string => `${page.page_id || ''}\u0000${page.locale || ''}`
const localeNumber = (index: number): string => String(index + 1).padStart(2, '0')
const setLocaleTriggerRef = (locale: string, element: TemplateRefTarget): void => {
  const target = element && '$el' in element ? element.$el : element
  if (target instanceof HTMLElement) localeTriggerRefs.set(locale, target)
  else localeTriggerRefs.delete(locale)
}
const updateLocaleOverflow = (): void => {
  const area = localeScrollArea.value
  hasLocaleOverflow.value = Boolean(area && area.scrollWidth > area.clientWidth + 1)
}
const centerActiveLocale = async (locale = props.activeStructureLocale, behavior: ScrollBehavior = 'smooth'): Promise<void> => {
  await nextTick()
  const area = localeScrollArea.value
  const target = localeTriggerRefs.get(locale)
  if (!area || !target) return

  const areaRect = area.getBoundingClientRect()
  const targetRect = target.getBoundingClientRect()
  const desiredLeft = area.scrollLeft
    + targetRect.left
    - areaRect.left
    - (area.clientWidth - targetRect.width) / 2
  const maxLeft = Math.max(0, area.scrollWidth - area.clientWidth)

  area.scrollTo({
    left: Math.max(0, Math.min(desiredLeft, maxLeft)),
    behavior
  })
}
const selectLocale = (locale: unknown): void => {
  if (typeof locale === 'string') emit('switch-locale', locale)
}
const isPageExpanded = (key: string): boolean => expandedPages.value.has(key)
const togglePage = (key: string): void => {
  const next = new Set(expandedPages.value)
  if (next.has(key)) next.delete(key)
  else next.add(key)
  expandedPages.value = next
}

const pageFAQCount = (page: FAQStructurePage): number => (
  (page.categories || []).reduce((total, category) => total + (category.faqs?.length || 0), 0)
)

watch(() => props.faqGroups, (groups) => {
  const keys = new Set(groups.map(pageKey))
  const next = new Set([...expandedPages.value].filter((key) => keys.has(key)))
  if (next.size === 0 && groups.length === 1) next.add(pageKey(groups[0]))
  expandedPages.value = next
}, { immediate: true })

const emit = defineEmits<{
  (event: 'switch-locale', locale: string): void
  (event: 'toggle-faq', faq: FAQItemLike, checked: FAQSelectionState): void
  (event: 'edit', faq: FAQItemLike): void
  (event: 'delete', faq: FAQItemLike): void
  (event: 'batch-delete'): void
  (event: 'edit-page', page: FAQStructurePage): void
  (event: 'create-category', page: FAQStructurePage): void
  (event: 'edit-category', page: FAQStructurePage, category: FAQCategory): void
  (event: 'delete-category', category: FAQCategory): void
  (event: 'create-faq', page: FAQStructurePage, category: FAQCategory): void
}>()

watch(
  () => [props.structureLocales, props.activeStructureLocale],
  async () => {
    await nextTick()
    updateLocaleOverflow()
    centerActiveLocale(props.activeStructureLocale)
  },
  { deep: true, immediate: true }
)

onMounted(() => {
  localeResizeObserver = new ResizeObserver(() => {
    updateLocaleOverflow()
    centerActiveLocale(props.activeStructureLocale, 'auto')
  })
  if (localeScrollArea.value) localeResizeObserver.observe(localeScrollArea.value)
  updateLocaleOverflow()
  centerActiveLocale(props.activeStructureLocale, 'auto')
})

onBeforeUnmount(() => {
  localeResizeObserver?.disconnect()
  localeTriggerRefs.clear()
})
</script>
