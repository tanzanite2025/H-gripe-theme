<template>
  <DialogContent size="md" class="max-h-[min(85dvh,46rem)] grid-rows-[auto_auto_minmax(0,1fr)_auto] gap-0 rounded-3xl p-0">
    <DialogHeader class="border-0 px-5 pb-4 pt-5 pr-14 text-left">
      <DialogTitle>{{ t('adminGlobalSearch.title') }}</DialogTitle>
      <DialogDescription>{{ t('adminGlobalSearch.description') }}</DialogDescription>
    </DialogHeader>

    <div class="px-5 pb-4">
      <div class="relative">
        <Search class="pointer-events-none absolute left-3.5 top-1/2 size-5 -translate-y-1/2 text-muted-foreground" aria-hidden="true" />
        <Input
          ref="globalSearchInput"
          :model-value="query"
          type="search"
          class="h-12 pl-11 pr-11 text-base font-normal sm:text-sm [&::-webkit-search-cancel-button]:appearance-none"
          :aria-label="t('adminGlobalSearch.inputLabel')"
          :placeholder="t('adminGlobalSearch.placeholder')"
          autocomplete="off"
          spellcheck="false"
          @update:model-value="emit('update:query', String($event))"
          @keydown.enter="emit('open-first-result', $event)"
        />
        <button
          v-if="query"
          type="button"
          class="absolute right-1 top-1/2 grid size-10 -translate-y-1/2 place-items-center rounded-full text-muted-foreground hover:text-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
          :aria-label="t('adminGlobalSearch.clear')"
          @click="clearAdminGlobalSearchQueryAndFocusInput"
        >
          <X class="size-4" aria-hidden="true" />
        </button>
      </div>
    </div>

    <section class="min-h-0 overflow-y-auto overscroll-contain border-t px-3 pb-3" :aria-label="t('adminGlobalSearch.results')" :aria-busy="loading">
      <div class="sticky top-0 z-10 flex items-center justify-between bg-popover px-2 py-3 text-xs text-muted-foreground">
        <h2 class="font-semibold">{{ resultGroups.length === 1 ? resultGroups[0].label : t(query.trim() ? 'adminGlobalSearch.results' : 'adminGlobalSearch.commonLinks') }}</h2>
        <span role="status" aria-live="polite">
          {{ t('adminGlobalSearch.resultCount', { count: resultCount }) }}
          <span v-if="loading" class="ml-2">{{ t('common.loading') }}</span>
        </span>
      </div>

      <div v-if="failedSourceLabels.length" role="alert" class="px-2 pb-3 text-xs text-destructive">
        {{ t('adminGlobalSearch.sourceUnavailable', { sources: failedSourceLabels.join('、') }) }}
        <button type="button" class="ml-2 rounded underline focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring" @click="emit('retry-search')">{{ t('adminGlobalSearch.retry') }}</button>
      </div>
      <p v-if="navigationFailed" role="alert" class="px-2 pb-3 text-xs text-destructive">{{ t('adminGlobalSearch.navigationFailed') }}</p>

      <template v-for="group in resultGroups" :key="group.sourceId">
        <div v-if="resultGroups.length > 1 && (group.results.length || group.loading)" class="flex items-center justify-between px-2 py-2 text-xs text-muted-foreground">
          <h3 class="font-semibold">{{ group.label }}</h3>
          <span v-if="group.loading" role="status">{{ t('common.loading') }}</span>
        </div>
        <ul v-if="group.results.length" class="space-y-1">
          <li v-for="entry in group.results" :key="entry.id">
            <RouterLink
              :to="entry.routeTarget"
              class="group flex min-w-0 items-center gap-3 rounded-2xl border border-transparent p-3 transition-colors hover:border-border/60 hover:bg-muted focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-inset"
              @click="emit('activate-link', $event)"
            >
              <span class="grid size-10 shrink-0 place-items-center rounded-xl bg-muted text-slate-500 group-hover:bg-background group-hover:text-foreground" aria-hidden="true">
                <component :is="entry.icon || FileText" class="size-5" />
              </span>
              <span class="min-w-0 flex-1">
                <span class="flex flex-wrap items-center gap-x-2 gap-y-0.5">
                  <span class="text-sm font-semibold">{{ entry.title }}</span>
                  <span v-if="entry.categoryLabel" class="text-[10px] text-muted-foreground">{{ entry.categoryLabel }}</span>
                </span>
                <span class="mt-0.5 block text-xs leading-relaxed text-muted-foreground">{{ entry.description }}</span>
                <span class="mt-1 block break-all text-[10px] text-muted-foreground/75">{{ entry.linkLabel }}</span>
              </span>
              <ArrowUpRight class="size-4 shrink-0 text-muted-foreground/60 group-hover:text-foreground" aria-hidden="true" />
            </RouterLink>
          </li>
        </ul>
      </template>

      <div v-if="!resultCount && !loading && !failedSourceLabels.length" class="px-4 py-12 text-center">
        <Search class="mx-auto mb-3 size-8 text-muted-foreground/50" aria-hidden="true" />
        <p class="text-sm font-semibold">{{ t('adminGlobalSearch.noResults') }}</p>
        <p class="mt-1 text-xs text-muted-foreground">{{ t('adminGlobalSearch.tryAnotherQuery') }}</p>
      </div>
    </section>

    <div class="flex flex-wrap items-center justify-between gap-2 border-t px-5 py-3 text-[10px] text-muted-foreground">
      <span>{{ t('adminGlobalSearch.openLinkHint') }}</span>
      <span>{{ t('adminGlobalSearch.keyboardHint') }}</span>
    </div>
  </DialogContent>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { ArrowUpRight, FileText, Search, X } from '@lucide/vue'
import { DialogContent, DialogDescription, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'
import type { AdminGlobalSearchResultGroup } from '@/modules/search/adminGlobalSearchResultAndSourceContracts'
import { useAdminI18n } from '@/i18n'

const props = defineProps<{
  query: string
  resultGroups: AdminGlobalSearchResultGroup[]
  loading: boolean
  failedSourceLabels: string[]
  navigationFailed: boolean
}>()
const emit = defineEmits<{
  (event: 'update:query', query: string): void
  (event: 'activate-link', mouseEvent: MouseEvent): void
  (event: 'open-first-result', keyboardEvent: KeyboardEvent): void
  (event: 'retry-search'): void
}>()
const { t } = useAdminI18n()
const globalSearchInput = ref<InstanceType<typeof Input> | null>(null)
const resultCount = computed(() => props.resultGroups.reduce((count, group) => count + group.results.length, 0))

function clearAdminGlobalSearchQueryAndFocusInput(): void {
  emit('update:query', '')
  globalSearchInput.value?.$el.focus()
}
</script>
