<template>
  <div class="space-y-4">
    <AdminPageHeader title="物流管理" description="管理承运商、包装规则、运费模板、配送区域和 17TRACK 追踪配置。">
      <template #actions>
        <Button variant="outline" :disabled="refreshing" @click="refreshCurrentTab">
          <RefreshCw :class="['size-3.5', { 'animate-spin': refreshing }]" />
          刷新
        </Button>
      </template>
    </AdminPageHeader>

    <AdminStatsGrid :items="statItems" />

    <Tabs v-model="activeTab" class="gap-4">
      <TabsList variant="line" class="h-10 w-full justify-start overflow-x-auto rounded-none border-b bg-transparent p-0">
        <TabsTrigger value="overview" class="h-9 flex-none px-3">
          <ClipboardList class="size-4" />
          概览
        </TabsTrigger>
        <TabsTrigger value="templates" class="h-9 flex-none px-3">
          <Calculator class="size-4" />
          运费模板
        </TabsTrigger>
        <TabsTrigger value="zones" class="h-9 flex-none px-3">
          <MapPin class="size-4" />
          配送区域
        </TabsTrigger>
        <TabsTrigger value="carriers" class="h-9 flex-none px-3">
          <Truck class="size-4" />
          承运商
        </TabsTrigger>
        <TabsTrigger value="packaging" class="h-9 flex-none px-3">
          <Package class="size-4" />
          包装规则
        </TabsTrigger>
        <TabsTrigger value="bindings" class="h-9 flex-none px-3">
          <Link2 class="size-4" />
          产品/SKU 绑定
        </TabsTrigger>
        <TabsTrigger value="tracking" class="h-9 flex-none px-3">
          <Radar class="size-4" />
          17TRACK
        </TabsTrigger>
      </TabsList>

      <TabsContent value="overview" class="space-y-4">
        <section class="grid gap-4 xl:grid-cols-3">
          <div class="rounded-lg border bg-card p-4 shadow-xs">
            <div class="flex items-start justify-between gap-3">
              <div>
                <h2 class="text-sm font-semibold">当前已接入</h2>
                <p class="mt-1 text-xs text-muted-foreground">已把核心配置入口放到独立物流后台里。</p>
              </div>
              <Badge variant="outline" class="border-emerald-200 bg-emerald-50 text-emerald-700">Phase 1/2</Badge>
            </div>
            <ul class="mt-4 space-y-2 text-sm text-muted-foreground">
              <li class="flex gap-2"><CircleCheck class="mt-0.5 size-4 text-emerald-600" />运费模板和规则矩阵列表、新增、编辑、删除</li>
              <li class="flex gap-2"><CircleCheck class="mt-0.5 size-4 text-emerald-600" />配送区域列表、新增、编辑、删除</li>
              <li class="flex gap-2"><CircleCheck class="mt-0.5 size-4 text-emerald-600" />默认、产品类型、产品和 SKU 模板绑定</li>
              <li class="flex gap-2"><CircleCheck class="mt-0.5 size-4 text-emerald-600" />承运商列表、新增、编辑、删除</li>
              <li class="flex gap-2"><CircleCheck class="mt-0.5 size-4 text-emerald-600" />包装规则列表、新增、编辑、删除</li>
            </ul>
          </div>

          <div class="rounded-lg border bg-card p-4 shadow-xs">
            <div class="flex items-start justify-between gap-3">
              <div>
                <h2 class="text-sm font-semibold">下一阶段</h2>
                <p class="mt-1 text-xs text-muted-foreground">下一步进入真正报价链路和追踪集成。</p>
              </div>
              <Badge variant="outline" class="border-amber-200 bg-amber-50 text-amber-700">Phase 4</Badge>
            </div>
            <ul class="mt-4 space-y-2 text-sm text-muted-foreground">
              <li class="flex gap-2"><CircleDashed class="mt-0.5 size-4 text-amber-600" />承运商线路服务和体积重/首续重规则增强</li>
              <li class="flex gap-2"><CircleDashed class="mt-0.5 size-4 text-amber-600" />运费试算器和结算报价 API</li>
              <li class="flex gap-2"><CircleDashed class="mt-0.5 size-4 text-amber-600" />Nuxt 单品页、购物车、结算接入报价</li>
            </ul>
          </div>

          <div class="rounded-lg border bg-card p-4 shadow-xs">
            <div class="flex items-start justify-between gap-3">
              <div>
                <h2 class="text-sm font-semibold">前台原则</h2>
                <p class="mt-1 text-xs text-muted-foreground">Nuxt 不再用硬编码运费，后续统一调用后端报价。</p>
              </div>
              <Badge variant="outline" class="border-blue-200 bg-blue-50 text-blue-700">Nuxt</Badge>
            </div>
            <ul class="mt-4 space-y-2 text-sm text-muted-foreground">
              <li class="flex gap-2"><CircleCheck class="mt-0.5 size-4 text-blue-600" />单品页未选 SKU 时显示重量范围</li>
              <li class="flex gap-2"><CircleCheck class="mt-0.5 size-4 text-blue-600" />选中 SKU 后使用该 SKU 的 weight_grams</li>
              <li class="flex gap-2"><CircleCheck class="mt-0.5 size-4 text-blue-600" />购物车和结算共用同一个 shipping quote composable</li>
            </ul>
          </div>
        </section>
      </TabsContent>

      <TabsContent value="templates" class="space-y-3">
        <div class="flex flex-wrap items-end justify-between gap-3">
          <div>
            <h2 class="text-sm font-semibold">运费模板</h2>
            <p class="mt-1 text-xs text-muted-foreground">维护基础计费方式、默认费用、免运门槛和区域规则矩阵。</p>
          </div>
          <Button v-if="hasPermission('shipping:create')" size="sm" @click="showCreateTemplateDialog">
            <Plus class="size-3.5" />
            新增运费模板
          </Button>
        </div>

        <AdminTablePanel :loading="loading.templates">
          <Table class="min-w-[1120px]">
            <TableHeader>
              <TableRow>
                <TableHead>模板名称</TableHead>
                <TableHead class="w-28">计费类型</TableHead>
                <TableHead class="w-28 text-right">默认运费</TableHead>
                <TableHead class="w-36 text-right">免运门槛</TableHead>
                <TableHead>规则摘要</TableHead>
                <TableHead class="w-24">状态</TableHead>
                <TableHead class="w-32 text-right">操作</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              <TableEmpty v-if="templates.length === 0" :colspan="7">
                <div class="flex flex-col items-center text-muted-foreground">
                  <Calculator class="mb-2 size-7 opacity-55" />
                  <span class="text-xs">暂无运费模板</span>
                </div>
              </TableEmpty>
              <TableRow v-for="template in templates" :key="template.id">
                <TableCell>
                  <span class="block font-medium">{{ template.name || '-' }}</span>
                  <span class="block max-w-96 truncate text-xs text-muted-foreground">{{ template.description || '暂无说明' }}</span>
                </TableCell>
                <TableCell>{{ templateTypeLabel(template.type) }}</TableCell>
                <TableCell class="text-right tabular-nums">{{ formatMoney(template.default_fee) }}</TableCell>
                <TableCell class="text-right tabular-nums">
                  {{ template.free_shipping ? formatMoney(template.free_threshold) : '未开启' }}
                </TableCell>
                <TableCell class="max-w-[28rem] truncate text-xs text-muted-foreground">{{ formatRuleSummary(template.rules) }}</TableCell>
                <TableCell>
                  <AdminStatusBadge :tone="template.enabled ? 'green' : 'gray'">
                    {{ template.enabled ? '启用' : '停用' }}
                  </AdminStatusBadge>
                </TableCell>
                <TableCell class="text-right">
                  <div class="inline-flex items-center gap-1">
                    <Button
                      v-if="hasPermission('shipping:edit')"
                      variant="ghost"
                      size="icon-sm"
                      :aria-label="`编辑运费模板 ${template.name}`"
                      @click="showEditTemplateDialog(template)"
                    >
                      <Pencil class="size-4" />
                    </Button>
                    <Button
                      v-if="hasPermission('shipping:delete')"
                      variant="ghost"
                      size="icon-sm"
                      class="text-destructive hover:text-destructive"
                      :aria-label="`删除运费模板 ${template.name}`"
                      @click="requestDelete('template', template)"
                    >
                      <Trash2 class="size-4" />
                    </Button>
                  </div>
                </TableCell>
              </TableRow>
            </TableBody>
          </Table>
        </AdminTablePanel>
      </TabsContent>

      <TabsContent value="zones" class="space-y-3">
        <div class="flex flex-wrap items-end justify-between gap-3">
          <div>
            <h2 class="text-sm font-semibold">配送区域</h2>
            <p class="mt-1 text-xs text-muted-foreground">按国家/地区、州省和邮编范围组织区域，供运费模板匹配使用。</p>
          </div>
          <Button v-if="hasPermission('shipping:create')" size="sm" @click="showCreateZoneDialog">
            <Plus class="size-3.5" />
            新增配送区域
          </Button>
        </div>

        <AdminTablePanel :loading="loading.zones">
          <Table class="min-w-[980px]">
            <TableHeader>
              <TableRow>
                <TableHead>区域名称</TableHead>
                <TableHead>国家/地区</TableHead>
                <TableHead>州/省</TableHead>
                <TableHead>邮编</TableHead>
                <TableHead class="w-24">状态</TableHead>
                <TableHead class="w-32 text-right">操作</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              <TableEmpty v-if="zones.length === 0" :colspan="6">
                <div class="flex flex-col items-center text-muted-foreground">
                  <MapPin class="mb-2 size-7 opacity-55" />
                  <span class="text-xs">暂无配送区域</span>
                </div>
              </TableEmpty>
              <TableRow v-for="zone in zones" :key="zone.id">
                <TableCell class="font-medium">{{ zone.name || '-' }}</TableCell>
                <TableCell class="max-w-72 truncate text-xs text-muted-foreground">{{ compactListLabel(zone.countries) }}</TableCell>
                <TableCell class="max-w-56 truncate text-xs text-muted-foreground">{{ compactListLabel(zone.states) }}</TableCell>
                <TableCell class="max-w-56 truncate text-xs text-muted-foreground">{{ compactListLabel(zone.postal_codes) }}</TableCell>
                <TableCell>
                  <AdminStatusBadge :tone="zone.enabled ? 'green' : 'gray'">
                    {{ zone.enabled ? '启用' : '停用' }}
                  </AdminStatusBadge>
                </TableCell>
                <TableCell class="text-right">
                  <div class="inline-flex items-center gap-1">
                    <Button
                      v-if="hasPermission('shipping:edit')"
                      variant="ghost"
                      size="icon-sm"
                      :aria-label="`编辑配送区域 ${zone.name}`"
                      @click="showEditZoneDialog(zone)"
                    >
                      <Pencil class="size-4" />
                    </Button>
                    <Button
                      v-if="hasPermission('shipping:delete')"
                      variant="ghost"
                      size="icon-sm"
                      class="text-destructive hover:text-destructive"
                      :aria-label="`删除配送区域 ${zone.name}`"
                      @click="requestDelete('zone', zone)"
                    >
                      <Trash2 class="size-4" />
                    </Button>
                  </div>
                </TableCell>
              </TableRow>
            </TableBody>
          </Table>
        </AdminTablePanel>
      </TabsContent>

      <TabsContent value="carriers" class="space-y-3">
        <div class="flex flex-wrap items-end justify-between gap-3">
          <div>
            <h2 class="text-sm font-semibold">承运商</h2>
            <p class="mt-1 text-xs text-muted-foreground">维护 DHL、FedEx、UPS、邮政小包、专线等物流公司基础资料。</p>
          </div>
          <Button v-if="hasPermission('shipping:create')" size="sm" @click="showCreateCarrierDialog">
            <Plus class="size-3.5" />
            新增承运商
          </Button>
        </div>

        <AdminTablePanel :loading="loading.carriers">
          <Table class="min-w-[1080px]">
            <TableHeader>
              <TableRow>
                <TableHead class="w-32">代码</TableHead>
                <TableHead>名称</TableHead>
                <TableHead class="w-44">联系人</TableHead>
                <TableHead>服务区域</TableHead>
                <TableHead class="w-24 text-right">排序</TableHead>
                <TableHead class="w-24">状态</TableHead>
                <TableHead class="w-32 text-right">操作</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              <TableEmpty v-if="carriers.length === 0" :colspan="7">
                <div class="flex flex-col items-center text-muted-foreground">
                  <Truck class="mb-2 size-7 opacity-55" />
                  <span class="text-xs">暂无承运商</span>
                </div>
              </TableEmpty>
              <TableRow v-for="carrier in carriers" :key="carrier.id">
                <TableCell class="font-mono text-xs font-semibold">{{ carrier.code || '-' }}</TableCell>
                <TableCell>
                  <span class="block font-medium">{{ carrier.name || '-' }}</span>
                  <span class="block truncate text-xs text-muted-foreground">{{ carrier.tracking_url || '未配置查询链接' }}</span>
                </TableCell>
                <TableCell>
                  <span class="block truncate text-sm">{{ carrier.contact || '-' }}</span>
                  <span class="block truncate text-xs text-muted-foreground">{{ carrier.email || carrier.phone || '-' }}</span>
                </TableCell>
                <TableCell class="max-w-80 truncate text-xs text-muted-foreground">{{ serviceAreaLabel(carrier.service_area) }}</TableCell>
                <TableCell class="text-right tabular-nums">{{ carrier.sort_order || 0 }}</TableCell>
                <TableCell>
                  <AdminStatusBadge :tone="carrier.enabled ? 'green' : 'gray'">
                    {{ carrier.enabled ? '启用' : '停用' }}
                  </AdminStatusBadge>
                </TableCell>
                <TableCell class="text-right">
                  <div class="inline-flex items-center gap-1">
                    <Button
                      v-if="hasPermission('shipping:edit')"
                      variant="ghost"
                      size="icon-sm"
                      :aria-label="`编辑承运商 ${carrier.name}`"
                      @click="showEditCarrierDialog(carrier)"
                    >
                      <Pencil class="size-4" />
                    </Button>
                    <Button
                      v-if="hasPermission('shipping:delete')"
                      variant="ghost"
                      size="icon-sm"
                      class="text-destructive hover:text-destructive"
                      :aria-label="`删除承运商 ${carrier.name}`"
                      @click="requestDelete('carrier', carrier)"
                    >
                      <Trash2 class="size-4" />
                    </Button>
                  </div>
                </TableCell>
              </TableRow>
            </TableBody>
          </Table>
        </AdminTablePanel>
      </TabsContent>

      <TabsContent value="packaging" class="space-y-3">
        <div class="flex flex-wrap items-end justify-between gap-3">
          <div>
            <h2 class="text-sm font-semibold">包装规则</h2>
            <p class="mt-1 text-xs text-muted-foreground">先管理包装箱重量、尺寸和最大承重；产品/SKU 绑定后续单独做。</p>
          </div>
          <Button v-if="hasPermission('shipping:create')" size="sm" @click="showCreatePackagingDialog">
            <Plus class="size-3.5" />
            新增包装规则
          </Button>
        </div>

        <AdminTablePanel :loading="loading.packaging">
          <Table class="min-w-[980px]">
            <TableHeader>
              <TableRow>
                <TableHead>规则名称</TableHead>
                <TableHead class="w-28 text-right">包装重量</TableHead>
                <TableHead class="w-44">尺寸</TableHead>
                <TableHead class="w-28 text-right">最大承重</TableHead>
                <TableHead class="w-28 text-right">绑定商品</TableHead>
                <TableHead class="w-24">状态</TableHead>
                <TableHead class="w-32 text-right">操作</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              <TableEmpty v-if="packagingRules.length === 0" :colspan="7">
                <div class="flex flex-col items-center text-muted-foreground">
                  <Package class="mb-2 size-7 opacity-55" />
                  <span class="text-xs">暂无包装规则</span>
                </div>
              </TableEmpty>
              <TableRow v-for="rule in packagingRules" :key="rule.id">
                <TableCell>
                  <span class="block font-medium">{{ rule.rule_name || '-' }}</span>
                  <span class="block max-w-96 truncate text-xs text-muted-foreground">{{ rule.description || '暂无说明' }}</span>
                </TableCell>
                <TableCell class="text-right tabular-nums">{{ formatWeight(rule.box_weight) }}</TableCell>
                <TableCell class="font-mono text-xs text-muted-foreground">{{ formatDimensions(rule) }}</TableCell>
                <TableCell class="text-right tabular-nums">{{ formatWeight(rule.max_weight) }}</TableCell>
                <TableCell class="text-right tabular-nums">{{ appliesCount(rule) }}</TableCell>
                <TableCell>
                  <AdminStatusBadge :tone="rule.is_active ? 'green' : 'gray'">
                    {{ rule.is_active ? '启用' : '停用' }}
                  </AdminStatusBadge>
                </TableCell>
                <TableCell class="text-right">
                  <div class="inline-flex items-center gap-1">
                    <Button
                      v-if="hasPermission('shipping:edit')"
                      variant="ghost"
                      size="icon-sm"
                      :aria-label="`编辑包装规则 ${rule.rule_name}`"
                      @click="showEditPackagingDialog(rule)"
                    >
                      <Pencil class="size-4" />
                    </Button>
                    <Button
                      v-if="hasPermission('shipping:delete')"
                      variant="ghost"
                      size="icon-sm"
                      class="text-destructive hover:text-destructive"
                      :aria-label="`删除包装规则 ${rule.rule_name}`"
                      @click="requestDelete('packaging', rule)"
                    >
                      <Trash2 class="size-4" />
                    </Button>
                  </div>
                </TableCell>
              </TableRow>
            </TableBody>
          </Table>
        </AdminTablePanel>
      </TabsContent>

      <TabsContent value="bindings" class="space-y-3">
        <div class="flex flex-wrap items-end justify-between gap-3">
          <div>
            <h2 class="text-sm font-semibold">产品/SKU 绑定</h2>
            <p class="mt-1 text-xs text-muted-foreground">设置默认、产品类型、产品或 SKU 应使用的运费模板。</p>
          </div>
          <Button v-if="hasPermission('shipping:create')" size="sm" :disabled="templates.length === 0" @click="showCreateBindingDialog">
            <Plus class="size-3.5" />
            新增绑定
          </Button>
        </div>

        <AdminTablePanel :loading="loading.bindings">
          <Table class="min-w-[980px]">
            <TableHeader>
              <TableRow>
                <TableHead class="w-32">范围</TableHead>
                <TableHead>目标</TableHead>
                <TableHead>运费模板</TableHead>
                <TableHead class="w-24 text-right">优先级</TableHead>
                <TableHead class="w-24">状态</TableHead>
                <TableHead class="w-32 text-right">操作</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              <TableEmpty v-if="templateBindings.length === 0" :colspan="6">
                <div class="flex flex-col items-center text-muted-foreground">
                  <Link2 class="mb-2 size-7 opacity-55" />
                  <span class="text-xs">暂无模板绑定</span>
                </div>
              </TableEmpty>
              <TableRow v-for="binding in templateBindings" :key="binding.id">
                <TableCell>{{ bindingScopeLabel(binding.scope) }}</TableCell>
                <TableCell class="font-mono text-xs text-muted-foreground">{{ bindingTargetLabel(binding) }}</TableCell>
                <TableCell>
                  <span class="block font-medium">{{ bindingTemplateName(binding) }}</span>
                  <span class="block text-xs text-muted-foreground">ID: {{ binding.template_id }}</span>
                </TableCell>
                <TableCell class="text-right tabular-nums">{{ binding.priority || 0 }}</TableCell>
                <TableCell>
                  <AdminStatusBadge :tone="binding.enabled ? 'green' : 'gray'">
                    {{ binding.enabled ? '启用' : '停用' }}
                  </AdminStatusBadge>
                </TableCell>
                <TableCell class="text-right">
                  <div class="inline-flex items-center gap-1">
                    <Button
                      v-if="hasPermission('shipping:edit')"
                      variant="ghost"
                      size="icon-sm"
                      :aria-label="`编辑模板绑定 ${binding.id}`"
                      @click="showEditBindingDialog(binding)"
                    >
                      <Pencil class="size-4" />
                    </Button>
                    <Button
                      v-if="hasPermission('shipping:delete')"
                      variant="ghost"
                      size="icon-sm"
                      class="text-destructive hover:text-destructive"
                      :aria-label="`删除模板绑定 ${binding.id}`"
                      @click="requestDelete('binding', binding)"
                    >
                      <Trash2 class="size-4" />
                    </Button>
                  </div>
                </TableCell>
              </TableRow>
            </TableBody>
          </Table>
        </AdminTablePanel>
      </TabsContent>

      <TabsContent value="tracking">
        <RoadmapPanel
          title="17TRACK 追踪"
          badge="Phase 4"
          description="17TRACK 应作为全局物流追踪 Provider 配置，不和单个承运商的 API Key 混在一起。"
          :items="[
            '17TRACK API Key、Base URL、Webhook Secret',
            '承运商代码映射和追踪号注册',
            '订单页展示 tracking_events 同步结果',
          ]"
        />
      </TabsContent>
    </Tabs>

    <ShippingTemplateEditorDialog
      v-model:open="templateDialogOpen"
      :mode="templateDialogMode"
      :form="templateForm"
      :errors="templateErrors"
      :submitting="templateSubmitting"
      @submit="saveTemplate"
      @clear-error="clearTemplateError"
    />

    <ShippingZoneEditorDialog
      v-model:open="zoneDialogOpen"
      :mode="zoneDialogMode"
      :form="zoneForm"
      :errors="zoneErrors"
      :submitting="zoneSubmitting"
      @submit="saveZone"
      @clear-error="clearZoneError"
    />

    <ShippingTemplateBindingEditorDialog
      v-model:open="bindingDialogOpen"
      :mode="bindingDialogMode"
      :form="bindingForm"
      :errors="bindingErrors"
      :templates="templates"
      :submitting="bindingSubmitting"
      @submit="saveBinding"
      @clear-error="clearBindingError"
    />

    <CarrierEditorDialog
      v-model:open="carrierDialogOpen"
      :mode="carrierDialogMode"
      :form="carrierForm"
      :errors="carrierErrors"
      :submitting="carrierSubmitting"
      @submit="saveCarrier"
      @clear-error="clearCarrierError"
    />

    <PackagingRuleEditorDialog
      v-model:open="packagingDialogOpen"
      :mode="packagingDialogMode"
      :form="packagingForm"
      :errors="packagingErrors"
      :submitting="packagingSubmitting"
      @submit="savePackagingRule"
      @clear-error="clearPackagingError"
    />

    <AdminConfirmDialog
      v-model:open="deleteDialogOpen"
      :title="deleteDialogTitle"
      :description="deleteDialogDescription"
      confirm-label="确认删除"
      destructive
      @confirm="confirmDelete"
    />
  </div>
</template>

<script setup>
import { computed, defineComponent, h, onMounted, reactive, ref } from 'vue'
import { toast } from 'vue-sonner'
import {
  Calculator,
  CircleCheck,
  CircleDashed,
  ClipboardList,
  Link2,
  MapPin,
  Package,
  Pencil,
  Plus,
  Radar,
  RefreshCw,
  Trash2,
  Truck,
} from '@lucide/vue'
import shippingApi from '@/api/shipping'
import AdminConfirmDialog from '@/components/admin/AdminConfirmDialog.vue'
import AdminPageHeader from '@/components/admin/AdminPageHeader.vue'
import AdminStatsGrid from '@/components/admin/AdminStatsGrid.vue'
import AdminStatusBadge from '@/components/admin/AdminStatusBadge.vue'
import AdminTablePanel from '@/components/admin/AdminTablePanel.vue'
import CarrierEditorDialog from '@/components/admin/shipping/CarrierEditorDialog.vue'
import PackagingRuleEditorDialog from '@/components/admin/shipping/PackagingRuleEditorDialog.vue'
import ShippingTemplateBindingEditorDialog from '@/components/admin/shipping/ShippingTemplateBindingEditorDialog.vue'
import ShippingTemplateEditorDialog from '@/components/admin/shipping/ShippingTemplateEditorDialog.vue'
import ShippingZoneEditorDialog from '@/components/admin/shipping/ShippingZoneEditorDialog.vue'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Table, TableBody, TableCell, TableEmpty, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { useAuthStore } from '@/stores/auth'

const RoadmapPanel = defineComponent({
  name: 'RoadmapPanel',
  props: {
    title: { type: String, required: true },
    badge: { type: String, required: true },
    description: { type: String, required: true },
    items: { type: Array, required: true },
  },
  setup(props) {
    return () => h('section', { class: 'rounded-lg border bg-card p-5 shadow-xs' }, [
      h('div', { class: 'flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between' }, [
        h('div', { class: 'min-w-0' }, [
          h('h2', { class: 'text-base font-semibold' }, props.title),
          h('p', { class: 'mt-1 max-w-3xl text-sm text-muted-foreground' }, props.description),
        ]),
        h(Badge, { variant: 'outline', class: 'w-fit border-amber-200 bg-amber-50 text-amber-700' }, () => props.badge),
      ]),
      h('ul', { class: 'mt-5 grid gap-3 md:grid-cols-3' }, props.items.map((item) =>
        h('li', { class: 'rounded-lg border bg-muted/35 p-3 text-sm text-muted-foreground' }, [
          h(CircleDashed, { class: 'mb-2 size-4 text-amber-600' }),
          h('span', item),
        ])
      )),
    ])
  },
})

const authStore = useAuthStore()
const activeTab = ref('overview')
const templates = ref([])
const zones = ref([])
const templateBindings = ref([])
const carriers = ref([])
const packagingRules = ref([])
const refreshing = ref(false)
const loading = reactive({
  templates: false,
  zones: false,
  bindings: false,
  carriers: false,
  packaging: false,
})

const templateDialogOpen = ref(false)
const templateDialogMode = ref('create')
const templateSubmitting = ref(false)
const templateErrors = reactive({})
const templateForm = reactive(defaultTemplateForm())

const zoneDialogOpen = ref(false)
const zoneDialogMode = ref('create')
const zoneSubmitting = ref(false)
const zoneErrors = reactive({})
const zoneForm = reactive(defaultZoneForm())

const bindingDialogOpen = ref(false)
const bindingDialogMode = ref('create')
const bindingSubmitting = ref(false)
const bindingErrors = reactive({})
const bindingForm = reactive(defaultBindingForm())

const carrierDialogOpen = ref(false)
const carrierDialogMode = ref('create')
const carrierSubmitting = ref(false)
const carrierErrors = reactive({})
const carrierForm = reactive(defaultCarrierForm())

const packagingDialogOpen = ref(false)
const packagingDialogMode = ref('create')
const packagingSubmitting = ref(false)
const packagingErrors = reactive({})
const packagingForm = reactive(defaultPackagingForm())

const deleteDialogOpen = ref(false)
const deleteTarget = ref(null)
const deleteType = ref('')
const deleteLoading = ref(false)

const hasPermission = (permission) => authStore.hasPermission(permission)

const statItems = computed(() => [
  {
    key: 'templates',
    label: '运费模板',
    value: templates.value.length,
    icon: Calculator,
    tone: 'blue',
  },
  {
    key: 'zones',
    label: '配送区域',
    value: zones.value.length,
    icon: MapPin,
    tone: 'amber',
  },
  {
    key: 'carriers',
    label: '承运商',
    value: carriers.value.length,
    icon: Truck,
    tone: 'blue',
  },
  {
    key: 'activePackaging',
    label: '启用包装规则',
    value: packagingRules.value.filter((rule) => rule.is_active).length,
    icon: CircleCheck,
    tone: 'green',
  },
])

const deleteDialogTitle = computed(() => {
  if (deleteType.value === 'template') return '删除运费模板？'
  if (deleteType.value === 'zone') return '删除配送区域？'
  if (deleteType.value === 'binding') return '删除模板绑定？'
  if (deleteType.value === 'carrier') return '删除承运商？'
  if (deleteType.value === 'packaging') return '删除包装规则？'
  return '确认删除？'
})

const deleteDialogDescription = computed(() => {
  const name = deleteTarget.value?.name || deleteTarget.value?.rule_name || '当前记录'
  if (deleteLoading.value) return '正在删除，请稍候。'
  return `将删除「${name}」。这个操作不可撤销，请确认没有正在使用它的运费规则或订单流程。`
})

function defaultCarrierForm() {
  return {
    id: null,
    name: '',
    code: '',
    tracking_url: '',
    api_endpoint: '',
    api_key: '',
    api_secret: '',
    contact: '',
    phone: '',
    email: '',
    service_area: '',
    enabled: true,
    sort_order: 0,
  }
}

function defaultTemplateForm() {
  return {
    id: null,
    name: '',
    type: 'weight',
    free_shipping: false,
    free_threshold: 0,
    default_fee: 0,
    description: '',
    enabled: true,
    rules: [],
  }
}

function defaultZoneForm() {
  return {
    id: null,
    name: '',
    countries: '[]',
    states: '[]',
    postal_codes: '[]',
    enabled: true,
  }
}

function defaultBindingForm() {
  return {
    id: null,
    template_id: '',
    scope: 'default',
    product_type_id: '',
    product_id: '',
    variant_id: '',
    priority: 0,
    enabled: true,
  }
}

function defaultPackagingForm() {
  return {
    id: null,
    rule_name: '',
    description: '',
    box_weight: 0,
    box_length: 0,
    box_width: 0,
    box_height: 0,
    max_weight: 0,
    is_active: true,
  }
}

function resetReactive(target, defaults) {
  Object.keys(target).forEach((key) => delete target[key])
  Object.assign(target, defaults)
}

function clearErrors(errors) {
  Object.keys(errors).forEach((key) => delete errors[key])
}

const clearTemplateError = (field) => {
  delete templateErrors[field]
}

const clearZoneError = (field) => {
  delete zoneErrors[field]
}

const clearBindingError = (field) => {
  delete bindingErrors[field]
}

const clearCarrierError = (field) => {
  delete carrierErrors[field]
}

const clearPackagingError = (field) => {
  delete packagingErrors[field]
}

const fetchTemplates = async () => {
  loading.templates = true
  try {
    templates.value = await shippingApi.listTemplates()
  } catch (error) {
    console.error('Failed to fetch shipping templates:', error)
  } finally {
    loading.templates = false
  }
}

const fetchZones = async () => {
  loading.zones = true
  try {
    zones.value = await shippingApi.listZones()
  } catch (error) {
    console.error('Failed to fetch shipping zones:', error)
  } finally {
    loading.zones = false
  }
}

const fetchTemplateBindings = async () => {
  loading.bindings = true
  try {
    templateBindings.value = await shippingApi.listTemplateBindings()
  } catch (error) {
    console.error('Failed to fetch shipping template bindings:', error)
  } finally {
    loading.bindings = false
  }
}

const fetchCarriers = async () => {
  loading.carriers = true
  try {
    carriers.value = await shippingApi.listCarriers()
  } catch (error) {
    console.error('Failed to fetch carriers:', error)
  } finally {
    loading.carriers = false
  }
}

const fetchPackagingRules = async () => {
  loading.packaging = true
  try {
    packagingRules.value = await shippingApi.listPackagingRules()
  } catch (error) {
    console.error('Failed to fetch packaging rules:', error)
  } finally {
    loading.packaging = false
  }
}

const refreshCurrentTab = async () => {
  refreshing.value = true
  try {
    if (activeTab.value === 'templates') {
      await fetchTemplates()
    } else if (activeTab.value === 'zones') {
      await fetchZones()
    } else if (activeTab.value === 'bindings') {
      await Promise.all([fetchTemplateBindings(), fetchTemplates()])
    } else if (activeTab.value === 'carriers') {
      await fetchCarriers()
    } else if (activeTab.value === 'packaging') {
      await fetchPackagingRules()
    } else {
      await Promise.all([fetchTemplates(), fetchZones(), fetchTemplateBindings(), fetchCarriers(), fetchPackagingRules()])
    }
  } finally {
    refreshing.value = false
  }
}

const showCreateTemplateDialog = () => {
  templateDialogMode.value = 'create'
  resetReactive(templateForm, defaultTemplateForm())
  clearErrors(templateErrors)
  templateDialogOpen.value = true
}

const showEditTemplateDialog = (template) => {
  templateDialogMode.value = 'edit'
  resetReactive(templateForm, {
    ...defaultTemplateForm(),
    ...template,
    free_threshold: Number(template.free_threshold || 0),
    default_fee: Number(template.default_fee || 0),
    enabled: template.enabled !== false,
    rules: Array.isArray(template.rules) ? template.rules.map((rule) => ({
      id: rule.id,
      region: rule.region || '',
      min_value: Number(rule.min_value || 0),
      max_value: Number(rule.max_value || 0),
      fee: Number(rule.fee || 0),
      additional: Number(rule.additional || 0),
    })) : [],
  })
  clearErrors(templateErrors)
  templateDialogOpen.value = true
}

const normalizeTemplateRules = () => (Array.isArray(templateForm.rules) ? templateForm.rules : [])
  .map((rule) => ({
    region: String(rule.region || '').trim().toUpperCase(),
    min_value: Number(rule.min_value || 0),
    max_value: Number(rule.max_value || 0),
    fee: Number(rule.fee || 0),
    additional: Number(rule.additional || 0),
  }))

const validateTemplate = () => {
  clearErrors(templateErrors)
  if (!templateForm.name?.trim()) templateErrors.name = '请输入模板名称'
  if (!['weight', 'quantity', 'price'].includes(templateForm.type)) templateErrors.type = '请选择计费类型'
  if (Number(templateForm.default_fee) < 0) templateErrors.default_fee = '默认运费不能小于 0'

  const invalidRule = normalizeTemplateRules().find((rule) =>
    !rule.region || rule.min_value < 0 || rule.max_value < 0 || rule.fee < 0 || rule.additional < 0 || (rule.max_value > 0 && rule.max_value < rule.min_value)
  )
  if (invalidRule) {
    toast.error('请检查规则矩阵：Region 必填，数值不能小于 0，最大值不能小于最小值')
    return false
  }

  return Object.keys(templateErrors).length === 0
}

const saveTemplate = async () => {
  if (!validateTemplate()) return

  templateSubmitting.value = true
  try {
    const payload = {
      name: templateForm.name.trim(),
      type: templateForm.type,
      free_shipping: Boolean(templateForm.free_shipping),
      free_threshold: Number(templateForm.free_threshold || 0),
      default_fee: Number(templateForm.default_fee || 0),
      description: templateForm.description || '',
      enabled: Boolean(templateForm.enabled),
      rules: normalizeTemplateRules(),
    }

    if (templateDialogMode.value === 'create') {
      await shippingApi.createTemplate(payload)
      toast.success('运费模板已创建')
    } else {
      await shippingApi.updateTemplate(templateForm.id, payload)
      toast.success('运费模板已更新')
    }

    templateDialogOpen.value = false
    await fetchTemplates()
  } catch (error) {
    console.error('Failed to save shipping template:', error)
  } finally {
    templateSubmitting.value = false
  }
}

const showCreateZoneDialog = () => {
  zoneDialogMode.value = 'create'
  resetReactive(zoneForm, defaultZoneForm())
  clearErrors(zoneErrors)
  zoneDialogOpen.value = true
}

const showEditZoneDialog = (zone) => {
  zoneDialogMode.value = 'edit'
  resetReactive(zoneForm, {
    ...defaultZoneForm(),
    ...zone,
    countries: zone.countries || '[]',
    states: zone.states || '[]',
    postal_codes: zone.postal_codes || '[]',
    enabled: zone.enabled !== false,
  })
  clearErrors(zoneErrors)
  zoneDialogOpen.value = true
}

const validateZone = () => {
  clearErrors(zoneErrors)
  if (!zoneForm.name?.trim()) zoneErrors.name = '请输入区域名称'
  if (!zoneForm.countries?.trim() || zoneForm.countries.trim() === '[]') zoneErrors.countries = '请输入至少一个国家/地区代码'
  return Object.keys(zoneErrors).length === 0
}

const saveZone = async () => {
  if (!validateZone()) return

  zoneSubmitting.value = true
  try {
    const payload = {
      name: zoneForm.name.trim(),
      countries: zoneForm.countries.trim(),
      states: zoneForm.states?.trim() || '[]',
      postal_codes: zoneForm.postal_codes?.trim() || '[]',
      enabled: Boolean(zoneForm.enabled),
    }

    if (zoneDialogMode.value === 'create') {
      await shippingApi.createZone(payload)
      toast.success('配送区域已创建')
    } else {
      await shippingApi.updateZone(zoneForm.id, payload)
      toast.success('配送区域已更新')
    }

    zoneDialogOpen.value = false
    await fetchZones()
  } catch (error) {
    console.error('Failed to save shipping zone:', error)
  } finally {
    zoneSubmitting.value = false
  }
}

const showCreateBindingDialog = () => {
  bindingDialogMode.value = 'create'
  resetReactive(bindingForm, {
    ...defaultBindingForm(),
    template_id: templates.value[0]?.id ? String(templates.value[0].id) : '',
  })
  clearErrors(bindingErrors)
  bindingDialogOpen.value = true
}

const showEditBindingDialog = (binding) => {
  bindingDialogMode.value = 'edit'
  resetReactive(bindingForm, {
    ...defaultBindingForm(),
    ...binding,
    template_id: binding.template_id ? String(binding.template_id) : '',
    product_type_id: binding.product_type_id || '',
    product_id: binding.product_id || '',
    variant_id: binding.variant_id || '',
    priority: Number(binding.priority || 0),
    enabled: binding.enabled !== false,
  })
  clearErrors(bindingErrors)
  bindingDialogOpen.value = true
}

const nullablePositiveID = (value) => {
  const numberValue = Number(value || 0)
  return numberValue > 0 ? numberValue : null
}

const validateBinding = () => {
  clearErrors(bindingErrors)
  if (!nullablePositiveID(bindingForm.template_id)) bindingErrors.template_id = '请选择运费模板'
  if (!['default', 'product_type', 'product', 'variant'].includes(bindingForm.scope)) bindingErrors.scope = '请选择绑定范围'
  if (bindingForm.scope === 'product_type' && !nullablePositiveID(bindingForm.product_type_id)) bindingErrors.product_type_id = '请输入产品类型 ID'
  if (bindingForm.scope === 'product' && !nullablePositiveID(bindingForm.product_id)) bindingErrors.product_id = '请输入产品 ID'
  if (bindingForm.scope === 'variant' && !nullablePositiveID(bindingForm.variant_id)) bindingErrors.variant_id = '请输入 SKU / 变体 ID'
  return Object.keys(bindingErrors).length === 0
}

const buildBindingPayload = () => {
  const payload = {
    template_id: nullablePositiveID(bindingForm.template_id),
    scope: bindingForm.scope,
    product_type_id: null,
    product_id: null,
    variant_id: null,
    priority: Number(bindingForm.priority || 0),
    enabled: Boolean(bindingForm.enabled),
  }

  if (bindingForm.scope === 'product_type') payload.product_type_id = nullablePositiveID(bindingForm.product_type_id)
  if (bindingForm.scope === 'product') payload.product_id = nullablePositiveID(bindingForm.product_id)
  if (bindingForm.scope === 'variant') payload.variant_id = nullablePositiveID(bindingForm.variant_id)

  return payload
}

const saveBinding = async () => {
  if (!validateBinding()) return

  bindingSubmitting.value = true
  try {
    const payload = buildBindingPayload()
    if (bindingDialogMode.value === 'create') {
      await shippingApi.createTemplateBinding(payload)
      toast.success('模板绑定已创建')
    } else {
      await shippingApi.updateTemplateBinding(bindingForm.id, payload)
      toast.success('模板绑定已更新')
    }

    bindingDialogOpen.value = false
    await fetchTemplateBindings()
  } catch (error) {
    console.error('Failed to save shipping template binding:', error)
  } finally {
    bindingSubmitting.value = false
  }
}

const showCreateCarrierDialog = () => {
  carrierDialogMode.value = 'create'
  resetReactive(carrierForm, defaultCarrierForm())
  clearErrors(carrierErrors)
  carrierDialogOpen.value = true
}

const showEditCarrierDialog = (carrier) => {
  carrierDialogMode.value = 'edit'
  resetReactive(carrierForm, {
    ...defaultCarrierForm(),
    ...carrier,
    enabled: carrier.enabled !== false,
    sort_order: Number(carrier.sort_order || 0),
  })
  clearErrors(carrierErrors)
  carrierDialogOpen.value = true
}

const validateCarrier = () => {
  clearErrors(carrierErrors)
  if (!carrierForm.name?.trim()) carrierErrors.name = '请输入承运商名称'
  if (!carrierForm.code?.trim()) carrierErrors.code = '请输入承运商代码'
  return Object.keys(carrierErrors).length === 0
}

const saveCarrier = async () => {
  if (!validateCarrier()) return

  carrierSubmitting.value = true
  try {
    const payload = {
      name: carrierForm.name.trim(),
      code: carrierForm.code.trim().toUpperCase(),
      tracking_url: carrierForm.tracking_url?.trim() || '',
      api_endpoint: carrierForm.api_endpoint?.trim() || '',
      api_key: carrierForm.api_key?.trim() || '',
      api_secret: carrierForm.api_secret?.trim() || '',
      contact: carrierForm.contact?.trim() || '',
      phone: carrierForm.phone?.trim() || '',
      email: carrierForm.email?.trim() || '',
      service_area: carrierForm.service_area || '',
      enabled: Boolean(carrierForm.enabled),
      sort_order: Number(carrierForm.sort_order || 0),
    }

    if (carrierDialogMode.value === 'create') {
      await shippingApi.createCarrier(payload)
      toast.success('承运商已创建')
    } else {
      await shippingApi.updateCarrier(carrierForm.id, payload)
      toast.success('承运商已更新')
    }

    carrierDialogOpen.value = false
    await fetchCarriers()
  } catch (error) {
    console.error('Failed to save carrier:', error)
  } finally {
    carrierSubmitting.value = false
  }
}

const showCreatePackagingDialog = () => {
  packagingDialogMode.value = 'create'
  resetReactive(packagingForm, defaultPackagingForm())
  clearErrors(packagingErrors)
  packagingDialogOpen.value = true
}

const showEditPackagingDialog = (rule) => {
  packagingDialogMode.value = 'edit'
  resetReactive(packagingForm, {
    ...defaultPackagingForm(),
    ...rule,
    box_weight: Number(rule.box_weight || 0),
    box_length: Number(rule.box_length || 0),
    box_width: Number(rule.box_width || 0),
    box_height: Number(rule.box_height || 0),
    max_weight: Number(rule.max_weight || 0),
    is_active: rule.is_active !== false,
  })
  clearErrors(packagingErrors)
  packagingDialogOpen.value = true
}

const validatePackaging = () => {
  clearErrors(packagingErrors)
  if (!packagingForm.rule_name?.trim()) packagingErrors.rule_name = '请输入规则名称'
  if (Number(packagingForm.box_weight) < 0) packagingErrors.box_weight = '包装重量不能小于 0'
  if (Number(packagingForm.max_weight) < 0) packagingErrors.max_weight = '最大承重不能小于 0'
  return Object.keys(packagingErrors).length === 0
}

const savePackagingRule = async () => {
  if (!validatePackaging()) return

  packagingSubmitting.value = true
  try {
    const payload = {
      rule_name: packagingForm.rule_name.trim(),
      description: packagingForm.description || '',
      box_weight: Number(packagingForm.box_weight || 0),
      box_length: Number(packagingForm.box_length || 0),
      box_width: Number(packagingForm.box_width || 0),
      box_height: Number(packagingForm.box_height || 0),
      max_weight: Number(packagingForm.max_weight || 0),
      is_active: Boolean(packagingForm.is_active),
    }

    if (packagingDialogMode.value === 'create') {
      await shippingApi.createPackagingRule(payload)
      toast.success('包装规则已创建')
    } else {
      await shippingApi.updatePackagingRule(packagingForm.id, payload)
      toast.success('包装规则已更新')
    }

    packagingDialogOpen.value = false
    await fetchPackagingRules()
  } catch (error) {
    console.error('Failed to save packaging rule:', error)
  } finally {
    packagingSubmitting.value = false
  }
}

const requestDelete = (type, target) => {
  deleteType.value = type
  deleteTarget.value = target
  deleteDialogOpen.value = true
}

const confirmDelete = async () => {
  if (!deleteTarget.value || deleteLoading.value) return

  deleteLoading.value = true
  try {
    if (deleteType.value === 'template') {
      await shippingApi.deleteTemplate(deleteTarget.value.id)
      toast.success('运费模板已删除')
      await fetchTemplates()
    } else if (deleteType.value === 'zone') {
      await shippingApi.deleteZone(deleteTarget.value.id)
      toast.success('配送区域已删除')
      await fetchZones()
    } else if (deleteType.value === 'binding') {
      await shippingApi.deleteTemplateBinding(deleteTarget.value.id)
      toast.success('模板绑定已删除')
      await fetchTemplateBindings()
    } else if (deleteType.value === 'carrier') {
      await shippingApi.deleteCarrier(deleteTarget.value.id)
      toast.success('承运商已删除')
      await fetchCarriers()
    } else if (deleteType.value === 'packaging') {
      await shippingApi.deletePackagingRule(deleteTarget.value.id)
      toast.success('包装规则已删除')
      await fetchPackagingRules()
    }
    deleteDialogOpen.value = false
  } catch (error) {
    console.error('Failed to delete shipping record:', error)
  } finally {
    deleteLoading.value = false
  }
}

const bindingScopeLabel = (scope) => {
  const labels = {
    default: '默认',
    product_type: '产品类型',
    product: '产品',
    variant: 'SKU / 变体',
  }
  return labels[scope] || scope || '-'
}

const bindingTargetLabel = (binding) => {
  if (binding.scope === 'default') return '全局默认'
  if (binding.scope === 'product_type') return `product_type_id=${binding.product_type_id || '-'}`
  if (binding.scope === 'product') return `product_id=${binding.product_id || '-'}`
  if (binding.scope === 'variant') return `variant_id=${binding.variant_id || '-'}`
  return '-'
}

const bindingTemplateName = (binding) => {
  if (binding.template?.name) return binding.template.name
  const template = templates.value.find((item) => Number(item.id) === Number(binding.template_id))
  return template?.name || '未知模板'
}

const templateTypeLabel = (type) => {
  const labels = {
    weight: '按重量',
    quantity: '按数量',
    price: '按金额',
  }
  return labels[type] || type || '-'
}

const formatMoney = (value) => Number(value || 0).toFixed(2)

const formatRuleSummary = (rules) => {
  if (!Array.isArray(rules) || rules.length === 0) return '无规则，使用默认运费'
  return rules
    .slice(0, 4)
    .map((rule) => `${rule.region || '-'} ${Number(rule.min_value || 0)}-${Number(rule.max_value || 0) || '∞'}: ${formatMoney(rule.fee)}`)
    .join('；')
}

const compactListLabel = (value) => {
  if (!value) return '-'
  try {
    const parsed = JSON.parse(value)
    if (Array.isArray(parsed)) {
      return parsed.length ? parsed.join(', ') : '-'
    }
  } catch {
    // keep raw value
  }
  return String(value)
}

const serviceAreaLabel = (value) => {
  if (!value) return '-'
  try {
    const parsed = JSON.parse(value)
    if (Array.isArray(parsed)) return parsed.join(', ')
  } catch {
    // keep raw value
  }
  return String(value)
}

const formatWeight = (value) => `${Number(value || 0).toFixed(3)} kg`

const formatDimensions = (rule) => {
  const length = Number(rule.box_length || 0).toFixed(2)
  const width = Number(rule.box_width || 0).toFixed(2)
  const height = Number(rule.box_height || 0).toFixed(2)
  return `${length} × ${width} × ${height} cm`
}

const appliesCount = (rule) => Array.isArray(rule.applies) ? rule.applies.length : 0

onMounted(() => Promise.all([fetchTemplates(), fetchZones(), fetchTemplateBindings(), fetchCarriers(), fetchPackagingRules()]))
</script>
