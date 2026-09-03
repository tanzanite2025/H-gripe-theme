<template>
  <div class="space-y-4">
    <AdminPageHeader title="商品品牌" description="维护所有商品共用的品牌主数据，商品详情和 SEO 会直接读取这里的品牌。">
      <template #actions>
        <Button variant="outline" as-child>
          <RouterLink to="/catalog/products">商品管理</RouterLink>
        </Button>
      </template>
    </AdminPageHeader>

    <ProductBrandManager />
  </div>
</template>

<script setup lang="ts">
import { RouterLink } from 'vue-router'
import AdminPageHeader from '@/components/admin/AdminPageHeader.vue'
import ProductBrandManager from '@/components/admin/product/ProductBrandManager.vue'
import { Button } from '@/components/ui/button'
</script>
