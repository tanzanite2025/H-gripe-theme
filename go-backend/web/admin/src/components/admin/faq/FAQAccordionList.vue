<template>
  <AdminTablePanel :loading="loading || structureLoading" :batch-visible="selectedFaqs.length > 0">
    <template #batch>
      <div class="flex flex-wrap items-center justify-between gap-2">
        <span class="text-xs font-medium">已选择 {{ selectedFaqs.length }} 个 FAQ</span>
        <Button v-if="hasPermission('faq:delete')" variant="destructive" size="sm" @click="$emit('batch-delete')">
          <Trash2 class="size-3.5" />
          批量删除
        </Button>
      </div>
    </template>

    <div class="border-b border-border/70 px-4 pt-3">
      <Tabs
        :model-value="activeStructureLocale"
        class="min-w-0 gap-0"
        @update:model-value="$emit('switch-locale', $event)"
      >
        <TabsList variant="line" class="h-10 w-full justify-start overflow-x-auto rounded-none border-b bg-transparent p-0">
          <TabsTrigger
            v-for="locale in structureLocales"
            :key="locale.value"
            :value="locale.value"
            class="h-9 flex-none px-3 text-[11px] font-bold normal-case tracking-normal"
          >
            {{ locale.label }}
          </TabsTrigger>
        </TabsList>
      </Tabs>
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

<script setup>
import { ref, watch } from 'vue'
import { ChevronDown, FolderPlus, Pencil, Plus, Trash2 } from '@lucide/vue'
import AdminStatusBadge from '@/components/admin/AdminStatusBadge.vue'
import AdminTablePanel from '@/components/admin/AdminTablePanel.vue'
import { Button } from '@/components/ui/button'
import { Checkbox } from '@/components/ui/checkbox'
import { Tabs, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Tooltip, TooltipContent, TooltipTrigger } from '@/components/ui/tooltip'

const props = defineProps({
  loading: { type: Boolean, default: false },
  structureLoading: { type: Boolean, default: false },
  faqGroups: { type: Array, required: true },
  selectedFaqs: { type: Array, required: true },
  pagination: { type: Object, required: true },
  structureLocales: { type: Array, required: true },
  activeStructureLocale: { type: String, required: true },
  hasPermission: { type: Function, required: true },
  isSelected: { type: Function, required: true },
  plainText: { type: Function, required: true },
  statusTone: { type: Function, required: true },
  statusName: { type: Function, required: true },
  visibilityName: { type: Function, required: true },
  visibilityTone: { type: Function, required: true },
  domainName: { type: Function, required: true }
})

const expandedPages = ref(new Set())

const pageKey = (page) => `${page.page_id || ''}\u0000${page.locale || ''}`
const isPageExpanded = (key) => expandedPages.value.has(key)
const togglePage = (key) => {
  const next = new Set(expandedPages.value)
  if (next.has(key)) next.delete(key)
  else next.add(key)
  expandedPages.value = next
}

const pageFAQCount = (page) => (
  (page.categories || []).reduce((total, category) => total + (category.faqs?.length || 0), 0)
)

watch(() => props.faqGroups, (groups) => {
  const keys = new Set(groups.map(pageKey))
  const next = new Set([...expandedPages.value].filter((key) => keys.has(key)))
  if (next.size === 0 && groups.length === 1) next.add(pageKey(groups[0]))
  expandedPages.value = next
}, { immediate: true })

defineEmits([
  'switch-locale',
  'toggle-faq',
  'edit',
  'delete',
  'batch-delete',
  'edit-page',
  'create-category',
  'edit-category',
  'delete-category',
  'create-faq'
])
</script>
