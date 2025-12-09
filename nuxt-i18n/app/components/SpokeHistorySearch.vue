<template>
  <section class="mt-8 rounded-2xl p-6 bg-[radial-gradient(circle_at_top_left,rgba(31,41,55,0.96),rgba(15,23,42,0.98))] shadow-[0_10px_26px_-14px_rgba(0,0,0,0.95)]">
    <header class="flex flex-col gap-1 mb-4">
      <h2 class="text-sm font-semibold text-slate-100">Spoke length history</h2>
      <p class="text-xs text-slate-400">
        Search past wheel builds by hub model, hub brand, or rim model. Results are limited to the most recent 5
        matches.
      </p>
    </header>

    <div class="flex flex-col gap-3 md:flex-row md:items-center">
      <input
        v-model="searchTextLocal"
        type="text"
        class="block w-full rounded-full border-none bg-[linear-gradient(135deg,rgba(15,23,42,0.98),rgba(15,23,42,0.96))] px-3 py-2.5 text-sm text-slate-50 placeholder:text-slate-500 shadow-[0_2px_6px_-3px_rgba(0,0,0,0.9),0_0_6px_rgba(15,23,42,0.7)] focus:outline-none focus:ring-0 focus:[box-shadow:0_0_0_1px_rgba(45,212,191,0.75),0_0_14px_rgba(45,212,191,0.35)]"
        placeholder="Search by hub model (e.g. '350', '240', 'D791')"
        @keyup.enter="onSearch"
      />
      <button
        type="button"
        class="inline-flex items-center justify-center rounded-full bg-[linear-gradient(135deg,#2dd4bf_0%,#3b82f6_100%)] px-4 py-2 text-sm font-medium text-slate-950 shadow-[0_4px_12px_rgba(45,212,191,0.3)] hover:shadow-[0_6px_18px_rgba(45,212,191,0.4)] hover:-translate-y-[1px] transition-all focus:outline-none focus-visible:ring-0 disabled:opacity-50 disabled:cursor-not-allowed md:w-auto"
        :disabled="loading"
        @click="onSearch"
      >
        <span v-if="loading">Searching…</span>
        <span v-else>Search history</span>
      </button>
    </div>

    <p v-if="error" class="mt-3 text-xs text-rose-400">
      {{ error }}
    </p>

    <div class="mt-4">
      <p v-if="!loading && !error && !items.length" class="text-xs text-slate-500">
        No matching history yet. Try a different keyword or check back after you have saved more wheel builds.
      </p>

      <ul v-else class="space-y-3">
        <li
          v-for="item in items"
          :key="item.id"
          class="rounded-xl border border-slate-800/70 bg-slate-950/70 px-4 py-3 text-xs text-slate-200 shadow-[0_4px_10px_-4px_rgba(0,0,0,0.95)]"
        >
          <div class="flex flex-col gap-1">
            <div class="flex flex-wrap items-center justify-between gap-2">
              <div class="font-semibold text-slate-100">
                <span v-if="item.hub_brand">{{ item.hub_brand }}</span>
                <span v-if="item.hub_model" class="ml-1">{{ item.hub_model }}</span>
                <span v-if="!item.hub_brand && !item.hub_model" class="text-slate-400">Unknown hub</span>
              </div>
              <div class="text-[11px] text-slate-400">
                <span v-if="item.spoke_count">{{ item.spoke_count }} spokes</span>
                <span v-if="item.lacing_pattern" class="ml-1">· {{ item.lacing_pattern }}</span>
                <span v-if="item.wheel_type" class="ml-1">· {{ item.wheel_type }}</span>
              </div>
            </div>

            <div class="text-[11px] text-slate-400">
              <span>Rim:</span>
              <span v-if="item.rim_brand" class="ml-1">{{ item.rim_brand }}</span>
              <span v-if="item.rim_model" class="ml-1">{{ item.rim_model }}</span>
              <span v-if="!item.rim_brand && !item.rim_model" class="ml-1">Unknown</span>
              <span v-if="item.erd_mm" class="ml-2">· ERD {{ item.erd_mm }} mm</span>
            </div>

            <div class="text-[11px] text-slate-300">
              <span>Left {{ item.left_length_mm ?? '—' }} mm</span>
              <span class="ml-2">· Right {{ item.right_length_mm ?? '—' }} mm</span>
            </div>
          </div>
        </li>
      </ul>

      <div v-if="isLoggedIn && canLoadMore" class="mt-3">
        <button
          type="button"
          class="inline-flex items-center justify-center rounded-full border-none bg-[radial-gradient(circle_at_top_left,rgba(31,41,55,0.96),rgba(15,23,42,0.98))] px-4 py-2 text-xs font-medium text-slate-100 shadow-[0_3px_9px_rgba(0,0,0,0.9)] hover:shadow-[0_4px_12px_rgba(0,0,0,1)] hover:-translate-y-[1px] transition-all focus:outline-none focus-visible:ring-0 disabled:opacity-50 disabled:cursor-not-allowed"
          :disabled="loading"
          @click="onLoadMore"
        >
          <span v-if="loading">Loading…</span>
          <span v-else>Load more results</span>
        </button>
      </div>

      <p v-if="showLoginHintForMore" class="mt-2 text-[11px] text-slate-400">
        There are more matching records for this search. Log in with your member account to view the full history.
      </p>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useSpokeHistory } from '~/composables/useSpokeHistory'
import { useAuth } from '~/composables/useAuth'

const auth = useAuth()
const isLoggedIn = computed(() => !!auth.user.value)

const { items, meta, loading, error, searchText, fetchHistory } = useSpokeHistory()

const searchTextLocal = ref('')

watch(
  () => searchText.value,
  (val) => {
    if (val !== searchTextLocal.value) {
      searchTextLocal.value = val || ''
    }
  },
  { immediate: true },
)

const onSearch = async () => {
  searchText.value = searchTextLocal.value.trim()
  await fetchHistory({ page: 1 })
}

onMounted(() => {
  // 初次加载时请求最近 5 条记录（游客也可以预览前 5 条）
  fetchHistory({ page: 1 })
})

const canLoadMore = computed(() => {
  if (!isLoggedIn.value) return false
  const m = meta.value
  if (!m) return false
  if (!items.value || !items.value.length) return false
  // 还有更多页，且当前已加载条数小于总数
  return m.page < m.total_pages && items.value.length < m.total
})

const onLoadMore = async () => {
  if (!canLoadMore.value || loading.value) return
  const m = meta.value
  if (!m) return
  await fetchHistory({ page: m.page + 1, append: true })
}

const showLoginHintForMore = computed(() => {
  if (isLoggedIn.value) return false
  const m = meta.value
  if (!m) return false
  // 当总数大于当前返回条数（默认 5 条）时，提示登录查看更多
  return m.total > items.value.length
})
</script>
