<template>
  <div class="space-y-4">
    <AdminPageHeader title="SEO / 文章" description="管理 Blog 文章资源的搜索元数据">
      <template #actions>
        <Button variant="outline" as-child>
          <a :href="blogHubHref" target="_blank" rel="noreferrer">
            <ExternalLink class="size-4" />
            打开 Blog
          </a>
        </Button>
        <Button variant="outline" :disabled="loading" @click="load">
          <RefreshCw :class="['size-4', loading ? 'animate-spin' : '']" />
          刷新
        </Button>
      </template>
    </AdminPageHeader>

    <section class="rounded-2xl border bg-muted/20 p-4">
      <div class="flex flex-col gap-3 lg:flex-row lg:items-end lg:justify-between">
        <div>
          <p class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/60">Article Resources</p>
          <h2 class="mt-1 text-sm font-black">文章标题与实际路由</h2>
        </div>
        <div class="flex flex-col gap-2 sm:flex-row">
          <Input
            v-model="filters.search"
            class="w-full sm:w-64"
            placeholder="搜索标题或 slug"
            @keydown.enter.prevent="applyFilters"
          />
          <Select v-model="filters.locale">
            <SelectTrigger class="w-full sm:w-48"><SelectValue placeholder="全部语言" /></SelectTrigger>
            <SelectContent>
              <SelectItem v-for="option in localeFilterOptions" :key="option.value" :value="option.value">
                {{ option.label }}
              </SelectItem>
            </SelectContent>
          </Select>
          <Button variant="outline" @click="applyFilters">
            <Search class="size-4" />
            查询
          </Button>
        </div>
      </div>
    </section>

    <SEOResourceTable
      resource-label="文章"
      :items="items"
      :pagination="pagination"
      :loading="loading"
      :can-edit="canEdit"
      @edit="openEditor"
      @update-page="updatePage"
      @update-page-size="updatePageSize"
    />

    <SEOResourceEditorDialog
      v-model:open="dialogOpen"
      kind="article"
      :resource="selectedResource"
      :saving="saving"
      :can-edit="canEdit"
      @save="saveSeo"
    />
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { toast } from 'vue-sonner'
import { ExternalLink, RefreshCw, Search } from '@lucide/vue'
import AdminPageHeader from '@/components/admin/AdminPageHeader.vue'
import SEOResourceEditorDialog from '@/components/admin/seo/SEOResourceEditorDialog.vue'
import SEOResourceTable from '@/components/admin/seo/SEOResourceTable.vue'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { useSupportedLanguages } from '@/composables/useSupportedLanguages'
import { useAuthStore } from '@/stores/auth'
import { seoArticlesApi } from '@/modules/seo/resources'
import { buildArticlePath, buildBlogHubPath, storefrontHref } from '@/modules/seo/routes'
import type {
  SEOArticleResource,
  SEOResourceEditorValues,
  SEOResourceItem,
  SEOResourcePagination,
} from '@/modules/seo/types'

const authStore = useAuthStore()
const supportedLanguages = useSupportedLanguages()
const canEdit = authStore.hasPermission('seo:edit')
const localeFilterOptions = supportedLanguages.localeFilterOptions
const blogHubHref = computed(() => storefrontHref(buildBlogHubPath()))

const items = ref<SEOResourceItem[]>([])
const selectedResource = ref<SEOResourceItem | null>(null)
const dialogOpen = ref(false)
const loading = ref(false)
const saving = ref(false)
const filters = reactive({ search: '', locale: 'all' })
const pagination = reactive<SEOResourcePagination>({
  page: 1,
  page_size: 20,
  total: 0,
  total_pages: 0,
})

const normalizeArticle = (article: SEOArticleResource): SEOResourceItem => {
  const routePath = buildArticlePath(article)
  return {
    id: article.id,
    title: article.title?.trim() || '(未命名文章)',
    routePath,
    href: routePath ? storefrontHref(routePath) : '',
    localeLabel: supportedLanguages.localeName(article.locale),
    status: article.status || 'draft',
    metaTitle: article.meta_title?.trim() || '',
    metaDescription: article.meta_description?.trim() || '',
    canonicalUrl: article.canonical_url?.trim() || '',
  }
}

const load = async (): Promise<void> => {
  loading.value = true
  try {
    const response = await seoArticlesApi.list({
      page: pagination.page,
      page_size: pagination.page_size,
      ...(filters.search.trim() ? { search: filters.search.trim() } : {}),
      ...(filters.locale !== 'all' ? { locale: filters.locale } : {}),
    })
    items.value = response.items.map(normalizeArticle)
    Object.assign(pagination, response.pagination)
  } catch (error) {
    console.error('Failed to load article SEO resources:', error)
    toast.error('文章 SEO 资源加载失败')
  } finally {
    loading.value = false
  }
}

const applyFilters = (): void => {
  pagination.page = 1
  void load()
}

const updatePage = (page: number): void => {
  pagination.page = page
  void load()
}

const updatePageSize = (pageSize: number): void => {
  pagination.page_size = pageSize
  pagination.page = 1
  void load()
}

const openEditor = (resource: SEOResourceItem): void => {
  selectedResource.value = resource
  dialogOpen.value = true
}

const saveSeo = async (values: SEOResourceEditorValues): Promise<void> => {
  if (!selectedResource.value || !canEdit) return
  saving.value = true
  try {
    const updated = await seoArticlesApi.update(selectedResource.value.id, values)
    const normalized = normalizeArticle(updated)
    items.value = items.value.map((item) => item.id === normalized.id ? normalized : item)
    selectedResource.value = normalized
    dialogOpen.value = false
    toast.success('文章 SEO 已保存')
  } catch (error) {
    console.error('Failed to save article SEO:', error)
    toast.error('文章 SEO 保存失败')
  } finally {
    saving.value = false
  }
}

onMounted(async () => {
  await supportedLanguages.fetchLanguages()
  await load()
})
</script>
