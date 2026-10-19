<template>
  <div class="mt-3 border-t pt-3">
    <p v-if="comments == null" class="text-xs leading-relaxed text-muted-foreground">
      此快照尚未采集适用范围附注，请在数据源配置中同步最新快照。
    </p>
    <details v-else-if="commentGroups.length">
      <summary class="cursor-pointer font-bold text-sky-700 dark:text-sky-300">
        适用范围与附注 · {{ commentCount }} 条
      </summary>
      <div class="mt-2 space-y-3 rounded-xl border bg-muted/30 p-3">
        <p class="text-xs leading-relaxed text-muted-foreground">
          以下为数据源原文，按税率分组，包含具体商品、服务及地区限制；同一税率下的附注不一定适用于每个类别。
        </p>
        <section v-for="group in commentGroups" :key="group.rate">
          <h3 class="font-mono font-bold text-foreground">{{ Number(group.rate) }}%</h3>
          <ol class="mt-2 list-decimal space-y-2 pl-5">
            <li v-for="(comment, index) in group.comments" :key="index" class="whitespace-pre-wrap break-words leading-relaxed text-foreground">{{ comment }}</li>
          </ol>
        </section>
      </div>
    </details>
    <p v-else class="text-xs text-muted-foreground">本次采集的数据源未提供适用范围附注。</p>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{ comments?: Record<string, string[]> | null }>()
const commentGroups = computed(() => Object.entries(props.comments ?? {})
  .filter(([, comments]) => comments.length > 0)
  .map(([rate, comments]) => ({ rate, comments }))
  .sort((left, right) => Number(left.rate) - Number(right.rate)))
const commentCount = computed(() => commentGroups.value.reduce((total, group) => total + group.comments.length, 0))
</script>
