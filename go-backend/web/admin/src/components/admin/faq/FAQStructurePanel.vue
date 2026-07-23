<template>
  <section class="rounded-3xl border border-border/70 bg-card/95 p-4 shadow-sm sm:p-5">
    <div class="mb-4 flex flex-wrap items-start justify-between gap-3">
      <div>
        <p class="text-[10px] font-black uppercase tracking-[0.22em] text-muted-foreground/70">STOREFRONT FAQ STRUCTURE</p>
        <h2 class="mt-1 text-lg font-black tracking-tight">前端 FAQ 页面分类</h2>
        <p class="mt-1 text-xs text-muted-foreground">这里对应 Nuxt 单页 FAQ 的 pageId 和分类。分类标识用于 FAQ 归属，分类名称用于前端展示。</p>
      </div>
      <div class="flex items-center gap-2 rounded-full border bg-muted/30 p-1">
        <Button
          v-for="locale in structureLocales"
          :key="locale.value"
          type="button"
          size="sm"
          :variant="activeStructureLocale === locale.value ? 'default' : 'ghost'"
          class="h-8 rounded-full px-3 text-xs font-black"
          @click="$emit('switch-locale', locale.value)"
        >
          {{ locale.label }}
        </Button>
      </div>
    </div>

    <div v-if="loading" class="flex min-h-36 items-center justify-center text-sm text-muted-foreground">
      <LoaderCircle class="mr-2 size-4 animate-spin" />
      正在加载 FAQ 页面分类
    </div>

    <div v-else-if="faqStructure.length === 0" class="rounded-2xl border border-dashed p-8 text-center text-sm text-muted-foreground">
      当前语言还没有 FAQ 页面分类。请先添加页面结构或检查后端种子数据。
    </div>

    <div v-else class="grid gap-3 xl:grid-cols-2 2xl:grid-cols-3">
      <article
        v-for="page in faqStructure"
        :key="page.page_id"
        class="flex min-h-0 flex-col rounded-2xl border border-border/70 bg-background/60 p-3"
      >
        <div class="mb-3 flex items-start justify-between gap-3">
          <div class="min-w-0">
            <div class="mb-1 flex flex-wrap items-center gap-2">
              <AdminStatusBadge tone="blue">{{ domainName(page.domain) }}</AdminStatusBadge>
              <AdminStatusBadge :tone="visibilityTone(page.status)">{{ visibilityName(page.status) }}</AdminStatusBadge>
            </div>
            <h3 class="truncate text-sm font-black">{{ page.title || page.page_id }}</h3>
            <p class="mt-1 truncate font-mono text-[11px] text-muted-foreground">{{ page.route_path || page.page_id }}</p>
          </div>
          <Button
            v-if="hasPermission('faq:edit')"
            type="button"
            variant="ghost"
            size="sm"
            class="h-8 rounded-full px-2 text-xs"
            @click="$emit('edit-page', page)"
          >
            <Pencil class="size-3.5" />
            页面
          </Button>
        </div>

        <div class="space-y-2">
          <div
            v-for="category in page.categories"
            :key="category.id"
            class="grid grid-cols-[auto_minmax(0,1fr)_auto] items-center gap-2 rounded-xl border bg-card/70 px-3 py-2"
          >
            <span class="flex size-8 items-center justify-center rounded-full bg-muted text-base">{{ category.icon || 'FAQ' }}</span>
            <div class="min-w-0">
              <div class="flex flex-wrap items-center gap-2">
                <p class="truncate text-xs font-black">{{ category.name }}</p>
                <AdminStatusBadge :tone="visibilityTone(category.status)">{{ visibilityName(category.status) }}</AdminStatusBadge>
              </div>
              <p class="mt-0.5 truncate font-mono text-[11px] text-muted-foreground">
                {{ category.category_key }} · {{ category.faq_count || 0 }} 条 FAQ
              </p>
            </div>
            <div class="flex items-center justify-end gap-1">
              <Button
                v-if="hasPermission('faq:edit')"
                type="button"
                variant="ghost"
                size="icon"
                class="size-8"
                :aria-label="`编辑分类 ${category.name}`"
                @click="$emit('edit-category', page, category)"
              >
                <Pencil class="size-3.5" />
              </Button>
              <Button
                v-if="hasPermission('faq:delete')"
                type="button"
                variant="ghost"
                size="icon"
                class="size-8 text-destructive hover:text-destructive"
                :aria-label="`删除分类 ${category.name}`"
                @click="$emit('delete-category', category)"
              >
                <Trash2 class="size-3.5" />
              </Button>
            </div>
          </div>
        </div>

        <Button
          v-if="hasPermission('faq:create')"
          type="button"
          variant="outline"
          size="sm"
          class="mt-3 h-9 rounded-full text-xs font-black"
          @click="$emit('create-category', page)"
        >
          <Plus class="size-3.5" />
          添加分类
        </Button>
      </article>
    </div>
  </section>
</template>

<script setup>
import { LoaderCircle, Pencil, Plus, Trash2 } from '@lucide/vue'
import AdminStatusBadge from '@/components/admin/AdminStatusBadge.vue'
import { Button } from '@/components/ui/button'

defineProps({
  structureLocales: { type: Array, required: true },
  activeStructureLocale: { type: String, required: true },
  loading: { type: Boolean, default: false },
  faqStructure: { type: Array, required: true },
  hasPermission: { type: Function, required: true },
  domainName: { type: Function, required: true },
  visibilityName: { type: Function, required: true },
  visibilityTone: { type: Function, required: true }
})

defineEmits(['switch-locale', 'edit-page', 'create-category', 'edit-category', 'delete-category'])
</script>
