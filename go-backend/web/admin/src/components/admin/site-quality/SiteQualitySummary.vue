<template>
 <div class="space-y-3">
 <section class="grid gap-3 sm:grid-cols-2 xl:grid-cols-7">
 <div v-for="item in items" :key="item.label" class="border bg-card p-3">
 <p class="text-[9px] font-black uppercase tracking-widest text-muted-foreground/70">{{ item.label }}</p>
 <div class="mt-2 flex min-h-10 items-center gap-2">
 <p class="min-w-0 text-2xl font-black" :class="item.tone">{{ item.value }}</p>
          <AdminStatusBadge v-if="item.badgeTone" :tone="item.badgeTone">
            {{ item.badge }}
          </AdminStatusBadge>
        </div>
 <p v-if="item.hint" class="mt-1 text-[11px] leading-5 text-muted-foreground">{{ item.hint }}</p>
      </div>
    </section>
 <section v-if="warnings.length > 0" class="border border-amber-300 bg-amber-50 px-4 py-3 text-xs text-amber-900">
 <p class="font-black">运行态势警告</p>
 <p v-for="warning in warnings" :key="warning" class="mt-1">{{ warning }}</p>
    </section>
  </div>
</template>

<script setup lang="ts">
import AdminStatusBadge, { type AdminStatusTone } from '@/components/admin/AdminStatusBadge.vue'

interface SummaryItem {
  label: string
  value: string
  tone?: string
  badge?: string
  badgeTone?: AdminStatusTone
  hint?: string
}

defineProps<{
  items: SummaryItem[]
  warnings: string[]
}>()
</script>
