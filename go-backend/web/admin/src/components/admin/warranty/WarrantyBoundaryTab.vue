<template>
  <TabsContent value="boundary" class="space-y-4">
    <section class="grid gap-4 lg:grid-cols-3">
      <div class="rounded-[24px] border border-dashed bg-card p-4">
        <ShieldCheck class="size-5 text-emerald-600" />
        <h2 class="mt-3 text-sm font-black tracking-tighter italic uppercase">单一事实源</h2>
        <p class="mt-2 text-xs leading-5 text-muted-foreground">
          后台只读取 Go 后端 `product_registrations`、`warranty_claims` 和 `warranty_service_records`，不接 WordPress，也不在前端保留影子数据。
        </p>
      </div>
      <div class="rounded-[24px] border border-dashed bg-card p-4">
        <GitBranch class="size-5 text-blue-600" />
        <h2 class="mt-3 text-sm font-black tracking-tighter italic uppercase">订单行边界</h2>
        <p class="mt-2 text-xs leading-5 text-muted-foreground">
          订单型申请通过 `warranty_claims.order_item_id` 绑定具体购买行；没有订单号时保持为空，不用处理备注伪造绑定。
        </p>
      </div>
      <div class="rounded-[24px] border border-dashed bg-card p-4">
        <FileWarning class="size-5 text-amber-600" />
        <h2 class="mt-3 text-sm font-black tracking-tighter italic uppercase">服务记录边界</h2>
        <p class="mt-2 text-xs leading-5 text-muted-foreground">
          检测、维修、更换、退款、物流处理等动作进入 `warranty_service_records`；`resolution` 只保留处理备注。
        </p>
      </div>
    </section>
  </TabsContent>
</template>

<script setup>
import { FileWarning, GitBranch, ShieldCheck } from '@lucide/vue'
import { TabsContent } from '@/components/ui/tabs'
</script>
