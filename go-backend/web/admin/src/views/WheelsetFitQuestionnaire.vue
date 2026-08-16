<template>
  <div class="space-y-4">
    <AdminPageHeader
      title="轮组选型问卷"
      description="固定顺序问答、帮助说明、商品筛选和客服结构化信息的独立管理入口"
    />

    <AdminTablePanel>
      <template #header>
        <div>
          <h2 class="text-sm font-black uppercase tracking-tight">问卷配置</h2>
          <p class="text-[11px] font-bold text-muted-foreground">一个轮组选型问卷，不包含助手列表或分支图编辑。</p>
        </div>
      </template>

      <div class="p-4">
        <section class="grid min-h-64 place-items-center rounded-lg border border-dashed border-border/80 bg-muted/20 p-6 text-center">
          <div class="max-w-md space-y-2">
            <ListChecks class="mx-auto size-6 text-muted-foreground" />
            <h3 class="text-sm font-black">固定顺序问卷正在接入</h3>
            <p class="text-xs font-bold leading-5 text-muted-foreground">
              旧分支图编辑器已经下线。新的有序问题列表、问题编辑弹窗、多语言内容、HELP 和商品筛选规则会基于独立问卷数据模型接入此处。
            </p>
          </div>
        </section>
      </div>
    </AdminTablePanel>
  </div>
</template>

<script setup lang="ts">
import { ListChecks } from '@lucide/vue'
import AdminPageHeader from '@/components/admin/AdminPageHeader.vue'
import AdminTablePanel from '@/components/admin/AdminTablePanel.vue'
</script>
