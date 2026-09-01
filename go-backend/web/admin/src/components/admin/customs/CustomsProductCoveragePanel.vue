<template>
  <Card class="overflow-hidden">
    <CardHeader class="border-b">
      <div class="flex flex-wrap items-start justify-between gap-3">
        <div>
          <CardTitle>商品清关完整度</CardTitle>
          <CardDescription>默认聚焦有缺失字段的商品，逐个补齐后即可进入正常结算和清关流程。</CardDescription>
        </div>
        <div class="flex items-center gap-2">
          <span class="rounded-full bg-amber-500/10 px-2.5 py-1 text-xs font-semibold text-amber-700">
            当前 {{ productTotal }} 件
          </span>
          <Button variant="ghost" size="icon" aria-label="刷新商品清关状态" @click="emit('refresh')">
            <RefreshCw class="size-4" :class="{ 'animate-spin': productLoading }" />
          </Button>
        </div>
      </div>
    </CardHeader>
    <CardContent class="space-y-3 p-4">
      <div class="grid gap-3 md:grid-cols-[minmax(0,1fr)_12rem_12rem_auto]">
        <div class="relative">
          <Search class="pointer-events-none absolute left-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" />
          <Input
            v-model="filters.search"
            class="pl-9"
            placeholder="搜索商品名称、SKU"
            @keyup.enter="emit('apply')"
          />
        </div>
        <Select v-model="filters.product_specification_template_id">
          <SelectTrigger><SelectValue placeholder="商品规格模板" /></SelectTrigger>
          <SelectContent>
            <SelectItem value="all">全部商品规格模板</SelectItem>
            <SelectItem v-for="type in productSpecTemplates" :key="type.id" :value="String(type.id)">
              {{ type.name }}
            </SelectItem>
          </SelectContent>
        </Select>
        <Select v-model="filters.customs_status">
          <SelectTrigger><SelectValue placeholder="清关状态" /></SelectTrigger>
          <SelectContent>
            <SelectItem value="incomplete">有缺失字段</SelectItem>
            <SelectItem value="complete">资料完整</SelectItem>
            <SelectItem value="missing_hs_code">缺 HS Code</SelectItem>
            <SelectItem value="missing_cn_code">缺 CN Code</SelectItem>
            <SelectItem value="missing_country_of_origin">缺原产国</SelectItem>
            <SelectItem value="missing_customs_description">缺英文品名</SelectItem>
            <SelectItem value="all">全部商品</SelectItem>
          </SelectContent>
        </Select>
        <Button :disabled="productLoading" @click="emit('apply')">
          <Search class="size-4" />
          筛选
        </Button>
      </div>

      <div class="overflow-x-auto rounded-lg border">
        <table class="w-full min-w-[760px] text-sm">
          <thead class="bg-muted/40 text-left text-xs text-muted-foreground">
            <tr>
              <th class="px-3 py-2.5 font-medium">商品</th>
              <th class="px-3 py-2.5 font-medium">商品规格模板</th>
              <th class="px-3 py-2.5 font-medium">资料状态</th>
              <th class="px-3 py-2.5 font-medium">缺失字段</th>
              <th class="px-3 py-2.5 text-right font-medium">操作</th>
            </tr>
          </thead>
          <tbody class="divide-y">
            <tr v-if="productLoading">
              <td colspan="5" class="px-3 py-10 text-center text-sm text-muted-foreground">加载中...</td>
            </tr>
            <tr v-else-if="!productRows.length">
              <td colspan="5" class="px-3 py-10 text-center text-sm text-muted-foreground">没有符合条件的商品。</td>
            </tr>
            <tr v-for="product in productRows" v-else :key="product.id" class="align-top">
              <td class="px-3 py-3">
                <p class="font-semibold">{{ product.name }}</p>
                <p class="mt-1 font-mono text-xs text-muted-foreground">{{ product.sku }}</p>
              </td>
              <td class="px-3 py-3 text-muted-foreground">
                {{ product.product_specification_template?.name || '未绑定商品规格模板' }}
              </td>
              <td class="px-3 py-3">
                <span
                  class="inline-flex items-center gap-1 rounded-full px-2 py-1 text-xs font-medium"
 :class="missingFields(product).length ? 'bg-amber-500/10 text-amber-700': 'bg-emerald-500/10 text-emerald-700'"
                >
                  <CircleAlert v-if="missingFields(product).length" class="size-3.5" />
                  <CheckCircle2 v-else class="size-3.5" />
                  {{ missingFields(product).length ? '需补充' : '完整' }}
                </span>
              </td>
              <td class="px-3 py-3">
                <div v-if="missingFields(product).length" class="flex flex-wrap gap-1.5">
                  <span
                    v-for="field in missingFields(product)"
                    :key="field"
                    class="rounded-full bg-amber-500/10 px-2 py-0.5 text-[11px] text-amber-700"
                  >
                    {{ field }}
                  </span>
                </div>
                <span v-else class="text-xs text-muted-foreground">HS / CN / 原产国 / 英文品名</span>
              </td>
              <td class="px-3 py-3 text-right">
                <Button variant="outline" size="sm" @click="emit('edit-product', product)">
                  <Pencil class="size-3.5" />
                  补齐资料
                </Button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <AdminPagination
        :page="productPage"
        :page-size="productPageSize"
        :total="productTotal"
        @update:page="emit('update:page', $event)"
        @update:page-size="emit('update:page-size', $event)"
      />
    </CardContent>
  </Card>
</template>

<script setup lang="ts">
import { CheckCircle2, CircleAlert, Pencil, RefreshCw, Search } from '@lucide/vue'
import AdminPagination from '@/components/admin/AdminPagination.vue'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { missingCustomsFields, type CustomsProductFilters } from '@/modules/customs/customsClassificationTypes'

withDefaults(defineProps<{
  productRows?: Record<string, any>[]
  productLoading?: boolean
  productPage?: number
  productPageSize?: number
  productTotal?: number
  filters: CustomsProductFilters
  productSpecTemplates?: any[]
}>(), {
  productRows: () => [],
  productLoading: false,
  productPage: 1,
  productPageSize: 20,
  productTotal: 0,
  productSpecTemplates: () => [],
})

const emit = defineEmits<{
  (event: 'refresh'): void
  (event: 'apply'): void
  (event: 'update:page', page: number): void
  (event: 'update:page-size', pageSize: number): void
  (event: 'edit-product', product: Record<string, any>): void
}>()

const missingFields = missingCustomsFields
</script>

