<template>
  <TabsContent value="overview" class="space-y-4">
    <section class="grid gap-4 xl:grid-cols-3">
      <div class="rounded-lg border bg-card p-4 shadow-xs">
        <div class="flex items-start justify-between gap-3">
          <div>
            <h2 class="text-sm font-black tracking-tighter italic uppercase">当前已接入</h2>
            <p class="mt-1 text-xs text-muted-foreground">已把核心配置入口放到独立物流后台里。</p>
          </div>
          <Badge variant="outline" class="border-emerald-200 bg-emerald-50 text-emerald-700">Phase 1/2</Badge>
        </div>
        <ul class="mt-4 space-y-2 text-sm text-muted-foreground">
          <li class="flex gap-2"><CircleCheck class="mt-0.5 size-4 text-emerald-600" />运费模板和规则矩阵列表、新增、编辑、删除</li>
          <li class="flex gap-2"><CircleCheck class="mt-0.5 size-4 text-emerald-600" />配送区域列表、新增、编辑、删除</li>
          <li class="flex gap-2"><CircleCheck class="mt-0.5 size-4 text-emerald-600" />默认、产品类型、产品和 SKU 模板绑定</li>
          <li class="flex gap-2"><CircleCheck class="mt-0.5 size-4 text-emerald-600" />承运商列表、新增、编辑、删除</li>
          <li class="flex gap-2"><CircleCheck class="mt-0.5 size-4 text-emerald-600" />追踪 Provider 配置独立维护，不和承运商档案混用</li>
          <li class="flex gap-2"><CircleCheck class="mt-0.5 size-4 text-emerald-600" />包装规则列表、新增、编辑、删除</li>
        </ul>
      </div>

      <div class="rounded-lg border bg-card p-4 shadow-xs">
        <div class="flex items-start justify-between gap-3">
          <div>
            <h2 class="text-sm font-black tracking-tighter italic uppercase">下一阶段</h2>
            <p class="mt-1 text-xs text-muted-foreground">下一步进入真正报价链路和追踪集成。</p>
          </div>
          <Badge variant="outline" class="border-amber-200 bg-amber-50 text-amber-700">Phase 4</Badge>
        </div>
        <ul class="mt-4 space-y-2 text-sm text-muted-foreground">
          <li class="flex gap-2"><CircleCheck class="mt-0.5 size-4 text-emerald-600" />承运商线路服务基础档案、首续重和体积重参数</li>
          <li class="flex gap-2"><CircleCheck class="mt-0.5 size-4 text-emerald-600" />运费试算器和结算报价 API</li>
          <li class="flex gap-2"><CircleCheck class="mt-0.5 size-4 text-emerald-600" />17TRACK 等追踪 Provider API 凭证、Webhook、同步策略配置</li>
          <li class="flex gap-2"><CircleCheck class="mt-0.5 size-4 text-emerald-600" />追踪任务面板已承接手动同步、自动轮询和 Webhook 状态落库</li>
          <li class="flex gap-2"><CircleDashed class="mt-0.5 size-4 text-amber-600" />Nuxt 单品页、购物车、结算接入报价</li>
        </ul>
      </div>

      <div class="rounded-lg border bg-card p-4 shadow-xs">
        <div class="flex items-start justify-between gap-3">
          <div>
            <h2 class="text-sm font-black tracking-tighter italic uppercase">前台原则</h2>
            <p class="mt-1 text-xs text-muted-foreground">Nuxt 不再用硬编码运费，后续统一调用后端报价。</p>
          </div>
          <Badge variant="outline" class="border-blue-200 bg-blue-50 text-blue-700">Nuxt</Badge>
        </div>
        <ul class="mt-4 space-y-2 text-sm text-muted-foreground">
          <li class="flex gap-2"><CircleCheck class="mt-0.5 size-4 text-blue-600" />单品页未选 SKU 时显示重量范围</li>
          <li class="flex gap-2"><CircleCheck class="mt-0.5 size-4 text-blue-600" />选中 SKU 后使用该 SKU 的 weight_grams</li>
          <li class="flex gap-2"><CircleCheck class="mt-0.5 size-4 text-blue-600" />购物车不再硬显示免运，结算读取后端 shipping_quote 明细</li>
        </ul>
      </div>
    </section>
  </TabsContent>
</template>

<script setup>
import { CircleCheck, CircleDashed } from '@lucide/vue'
import { Badge } from '@/components/ui/badge'
import { TabsContent } from '@/components/ui/tabs'
</script>
