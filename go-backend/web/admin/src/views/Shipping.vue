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
        <TabsTrigger value="services" class="h-9 flex-none px-3">
          <Route class="size-4" />
          线路服务
        </TabsTrigger>
        <TabsTrigger value="quote" class="h-9 flex-none px-3">
          <Calculator class="size-4" />
          试算器
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
          追踪配置
        </TabsTrigger>
        <TabsTrigger value="trackingShipments" class="h-9 flex-none px-3">
          <RefreshCw class="size-4" />
          追踪任务
        </TabsTrigger>
      </TabsList>

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

      <TabsContent value="templates" class="space-y-3">
        <div class="flex flex-wrap items-end justify-between gap-3">
          <div>
            <h2 class="text-sm font-black tracking-tighter italic uppercase">运费模板</h2>
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
                  <span class="block font-bold text-xs">{{ template.name || '-' }}</span>
                  <span class="block max-w-96 truncate text-[10px] text-muted-foreground/70">{{ template.description || '暂无说明' }}</span>
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

        <div class="flex flex-wrap items-end justify-between gap-3 pt-3">
          <div>
            <h3 class="text-sm font-black tracking-tighter italic uppercase">承运商代码映射</h3>
            <p class="mt-1 text-xs text-muted-foreground">把本地承运商或线路服务映射到 Provider carrier code，后续追踪号注册和轨迹同步统一读取这里。</p>
          </div>
          <Button
            v-if="hasPermission('shipping:create')"
            size="sm"
            :disabled="trackingProviders.length === 0 || (carriers.length === 0 && carrierServices.length === 0)"
            @click="showCreateTrackingCarrierMappingDialog"
          >
            <Plus class="size-3.5" />
            新增承运商映射
          </Button>
        </div>

        <AdminTablePanel :loading="loading.trackingMappings">
          <Table class="min-w-[1120px]">
            <TableHeader>
              <TableRow>
                <TableHead>追踪 Provider</TableHead>
                <TableHead>本地对象</TableHead>
                <TableHead>Provider Carrier</TableHead>
                <TableHead class="w-24 text-right">优先级</TableHead>
                <TableHead class="w-24">状态</TableHead>
                <TableHead class="w-32 text-right">操作</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              <TableEmpty v-if="trackingCarrierMappings.length === 0" :colspan="6">
                <div class="flex flex-col items-center text-muted-foreground">
                  <Radar class="mb-2 size-7 opacity-55" />
                  <span class="text-xs">暂无承运商代码映射</span>
                </div>
              </TableEmpty>
              <TableRow v-for="mapping in trackingCarrierMappings" :key="mapping.id">
                <TableCell>
                  <span class="block font-bold text-xs">{{ trackingProviderName(mapping) }}</span>
                  <span class="block font-mono text-[10px] text-muted-foreground/70">
                    provider_id={{ mapping.provider_id || '-' }}
                  </span>
                </TableCell>
                <TableCell>
                  <span class="block font-bold text-xs">{{ trackingMappingLocalTargetLabel(mapping) }}</span>
                  <span class="block text-[10px] text-muted-foreground/70">{{ trackingMappingScopeLabel(mapping.scope) }}</span>
                </TableCell>
                <TableCell>
                  <span class="block font-mono text-xs font-bold">{{ mapping.provider_carrier_code || '-' }}</span>
                  <span class="block max-w-80 truncate text-[10px] text-muted-foreground/70">
                    {{ mapping.provider_carrier_name || mapping.description || '暂无 Provider 名称' }}
                  </span>
                </TableCell>
                <TableCell class="text-right tabular-nums">{{ mapping.priority || 0 }}</TableCell>
                <TableCell>
                  <AdminStatusBadge :tone="mapping.enabled ? 'green' : 'gray'">
                    {{ mapping.enabled ? '启用' : '停用' }}
                  </AdminStatusBadge>
                </TableCell>
                <TableCell class="text-right">
                  <div class="inline-flex items-center gap-1">
                    <Button
                      v-if="hasPermission('shipping:edit')"
                      variant="ghost"
                      size="icon-sm"
                      :aria-label="`编辑承运商映射 ${mapping.provider_carrier_code}`"
                      @click="showEditTrackingCarrierMappingDialog(mapping)"
                    >
                      <Pencil class="size-4" />
                    </Button>
                    <Button
                      v-if="hasPermission('shipping:delete')"
                      variant="ghost"
                      size="icon-sm"
                      class="text-destructive hover:text-destructive"
                      :aria-label="`删除承运商映射 ${mapping.provider_carrier_code}`"
                      @click="requestDelete('trackingCarrierMapping', mapping)"
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
            <h2 class="text-sm font-black tracking-tighter italic uppercase">配送区域</h2>
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
                <TableCell class="font-bold text-xs">{{ zone.name || '-' }}</TableCell>
                <TableCell class="max-w-72 truncate text-[10px] text-muted-foreground/70">{{ compactListLabel(zone.countries) }}</TableCell>
                <TableCell class="max-w-56 truncate text-[10px] text-muted-foreground/70">{{ compactListLabel(zone.states) }}</TableCell>
                <TableCell class="max-w-56 truncate text-[10px] text-muted-foreground/70">{{ compactListLabel(zone.postal_codes) }}</TableCell>
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
            <h2 class="text-sm font-black tracking-tighter italic uppercase">承运商</h2>
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
                <TableCell class="font-mono text-xs font-bold">{{ carrier.code || '-' }}</TableCell>
                <TableCell>
                  <span class="block font-bold text-xs">{{ carrier.name || '-' }}</span>
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

      <TabsContent value="services" class="space-y-3">
        <div class="flex flex-wrap items-end justify-between gap-3">
          <div>
            <h2 class="text-sm font-black tracking-tighter italic uppercase">线路服务</h2>
            <p class="mt-1 text-xs text-muted-foreground">维护承运商下的国际线路、计费口径、首续重、体积重和预计时效。</p>
          </div>
          <Button
            v-if="hasPermission('shipping:create')"
            size="sm"
            :disabled="carriers.length === 0"
            @click="showCreateCarrierServiceDialog"
          >
            <Plus class="size-3.5" />
            新增线路服务
          </Button>
        </div>

        <AdminTablePanel :loading="loading.services">
          <Table class="min-w-[1280px]">
            <TableHeader>
              <TableRow>
                <TableHead>承运商 / 线路</TableHead>
                <TableHead>关联模板</TableHead>
                <TableHead>国家/区域</TableHead>
                <TableHead class="w-40">计费模式</TableHead>
                <TableHead class="w-48">首续重</TableHead>
                <TableHead class="w-40">体积重</TableHead>
                <TableHead class="w-28">时效</TableHead>
                <TableHead class="w-24">状态</TableHead>
                <TableHead class="w-32 text-right">操作</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              <TableEmpty v-if="carrierServices.length === 0" :colspan="9">
                <div class="flex flex-col items-center text-muted-foreground">
                  <Route class="mb-2 size-7 opacity-55" />
                  <span class="text-xs">暂无线路服务</span>
                </div>
              </TableEmpty>
              <TableRow v-for="service in carrierServices" :key="service.id">
                <TableCell>
                  <span class="block font-bold text-xs">{{ service.service_name || '-' }}</span>
                  <span class="block font-mono text-[10px] text-muted-foreground/70">
                    {{ carrierServiceCarrierName(service) }} · {{ service.service_code || '-' }}
                  </span>
                </TableCell>
                <TableCell>
                  <span class="block font-bold text-xs">{{ carrierServiceTemplateName(service) }}</span>
                  <span class="block text-[10px] text-muted-foreground/70">{{ service.route_name || '未填写线路渠道' }}</span>
                </TableCell>
                <TableCell class="max-w-72 truncate text-xs text-muted-foreground">{{ compactListLabel(service.countries) }}</TableCell>
                <TableCell>{{ billingModeLabel(service.billing_mode) }}</TableCell>
                <TableCell class="font-mono text-xs text-muted-foreground">
                  {{ formatServiceWeightStep(service) }}
                </TableCell>
                <TableCell class="font-mono text-xs text-muted-foreground">
                  {{ formatVolumetricDivisor(service) }}
                </TableCell>
                <TableCell class="text-xs text-muted-foreground">{{ formatEta(service) }}</TableCell>
                <TableCell>
                  <AdminStatusBadge :tone="service.enabled ? 'green' : 'gray'">
                    {{ service.enabled ? '启用' : '停用' }}
                  </AdminStatusBadge>
                </TableCell>
                <TableCell class="text-right">
                  <div class="inline-flex items-center gap-1">
                    <Button
                      v-if="hasPermission('shipping:edit')"
                      variant="ghost"
                      size="icon-sm"
                      :aria-label="`编辑线路服务 ${service.service_name}`"
                      @click="showEditCarrierServiceDialog(service)"
                    >
                      <Pencil class="size-4" />
                    </Button>
                    <Button
                      v-if="hasPermission('shipping:delete')"
                      variant="ghost"
                      size="icon-sm"
                      class="text-destructive hover:text-destructive"
                      :aria-label="`删除线路服务 ${service.service_name}`"
                      @click="requestDelete('carrierService', service)"
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

      <TabsContent value="quote" class="space-y-3">
        <ShippingQuoteCalculator />
      </TabsContent>

      <TabsContent value="packaging" class="space-y-3">
        <div class="flex flex-wrap items-end justify-between gap-3">
          <div>
            <h2 class="text-sm font-black tracking-tighter italic uppercase">包装规则</h2>
            <p class="mt-1 text-xs text-muted-foreground">管理包装箱重量、尺寸、最大承重和适用商品；当前先保持产品级事实源，SKU 级包装规则后续统一升级。</p>
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
                  <span class="block font-bold text-xs">{{ rule.rule_name || '-' }}</span>
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
                      :aria-label="`维护包装规则 ${rule.rule_name} 的适用商品`"
                      @click="showPackagingAppliesDialog(rule)"
                    >
                      <Link2 class="size-4" />
                    </Button>
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
            <h2 class="text-sm font-black tracking-tighter italic uppercase">产品/SKU 绑定</h2>
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
                  <span class="block font-bold text-xs">{{ bindingTemplateName(binding) }}</span>
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

      <TabsContent value="tracking" class="space-y-3">
        <div class="flex flex-wrap items-end justify-between gap-3">
          <div>
            <h2 class="text-sm font-black tracking-tighter italic uppercase">追踪配置</h2>
            <p class="mt-1 text-xs text-muted-foreground">统一维护 17TRACK / AfterShip 等追踪 Provider 的 API 凭证、Webhook 和同步策略。</p>
          </div>
          <Button v-if="hasPermission('shipping:create')" size="sm" @click="showCreateTrackingProviderDialog">
            <Plus class="size-3.5" />
            新增追踪配置
          </Button>
        </div>

        <AdminTablePanel :loading="loading.tracking">
          <Table class="min-w-[1120px]">
            <TableHeader>
              <TableRow>
                <TableHead>Provider</TableHead>
                <TableHead>接口地址</TableHead>
                <TableHead class="w-44">同步策略</TableHead>
                <TableHead class="w-44">凭证状态</TableHead>
                <TableHead class="w-24 text-right">排序</TableHead>
                <TableHead class="w-24">状态</TableHead>
                <TableHead class="w-32 text-right">操作</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              <TableEmpty v-if="trackingProviders.length === 0" :colspan="7">
                <div class="flex flex-col items-center text-muted-foreground">
                  <Radar class="mb-2 size-7 opacity-55" />
                  <span class="text-xs">暂无追踪 Provider 配置</span>
                </div>
              </TableEmpty>
              <TableRow v-for="provider in trackingProviders" :key="provider.id">
                <TableCell>
                  <span class="block font-bold text-xs">{{ provider.provider_name || '-' }}</span>
                  <span class="block font-mono text-[10px] text-muted-foreground/70">
                    {{ provider.provider_code || '-' }} · {{ trackingEnvironmentLabel(provider.environment) }}
                  </span>
                </TableCell>
                <TableCell>
                  <span class="block max-w-96 truncate font-mono text-xs">{{ provider.base_url || '未配置 Base URL' }}</span>
                  <div class="mt-1 flex max-w-96 items-center gap-1 rounded-md bg-muted/35 px-2 py-1">
                    <span
                      class="min-w-0 flex-1 truncate font-mono text-[10px] text-muted-foreground"
                      :title="trackingWebhookUrl(provider)"
                    >
                      {{ trackingWebhookUrl(provider) || '填写 Provider 代码后生成 Webhook URL' }}
                    </span>
                    <Button
                      v-if="trackingWebhookUrl(provider)"
                      variant="ghost"
                      size="icon-sm"
                      class="size-6 shrink-0"
                      :aria-label="`复制 ${provider.provider_name || provider.provider_code} Webhook 地址`"
                      @click="copyTrackingWebhookUrl(provider)"
                    >
                      <Copy class="size-3.5" />
                    </Button>
                  </div>
                  <span class="mt-1 block max-w-96 truncate text-[10px] text-muted-foreground/70">{{ provider.description || '暂无说明' }}</span>
                </TableCell>
                <TableCell class="text-xs text-muted-foreground">{{ formatTrackingSyncPolicy(provider) }}</TableCell>
                <TableCell>
                  <div class="flex flex-wrap gap-1.5">
                    <AdminStatusBadge :tone="trackingProviderHasApiKey(provider) ? 'green' : 'gray'">
                      API {{ trackingProviderHasApiKey(provider) ? '已配置' : '未配置' }}
                    </AdminStatusBadge>
                    <AdminStatusBadge :tone="trackingProviderHasWebhookSecret(provider) ? 'green' : 'gray'">
                      WEBHOOK {{ trackingProviderHasWebhookSecret(provider) ? '已配置' : '未配置' }}
                    </AdminStatusBadge>
                  </div>
                </TableCell>
                <TableCell class="text-right tabular-nums">{{ provider.sort_order || 0 }}</TableCell>
                <TableCell>
                  <AdminStatusBadge :tone="provider.enabled ? 'green' : 'gray'">
                    {{ provider.enabled ? '启用' : '停用' }}
                  </AdminStatusBadge>
                </TableCell>
                <TableCell class="text-right">
                  <div class="inline-flex items-center gap-1">
                    <Button
                      v-if="hasPermission('shipping:edit')"
                      variant="ghost"
                      size="icon-sm"
                      :aria-label="`编辑追踪配置 ${provider.provider_name}`"
                      @click="showEditTrackingProviderDialog(provider)"
                    >
                      <Pencil class="size-4" />
                    </Button>
                    <Button
                      v-if="hasPermission('shipping:delete')"
                      variant="ghost"
                      size="icon-sm"
                      class="text-destructive hover:text-destructive"
                      :aria-label="`删除追踪配置 ${provider.provider_name}`"
                      @click="requestDelete('trackingProvider', provider)"
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

      <TabsContent value="trackingShipments">
        <TrackingShipmentsPanel
          ref="trackingShipmentsPanelRef"
          :tracking-providers="trackingProviders"
          :carriers="carriers"
          :carrier-services="carrierServices"
          :can-edit="hasPermission('shipping:edit')"
          @count-change="handleTrackingShipmentsCountChange"
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

    <CarrierServiceEditorDialog
      v-model:open="carrierServiceDialogOpen"
      :mode="carrierServiceDialogMode"
      :form="carrierServiceForm"
      :errors="carrierServiceErrors"
      :carriers="carriers"
      :templates="templates"
      :submitting="carrierServiceSubmitting"
      @submit="saveCarrierService"
      @clear-error="clearCarrierServiceError"
    />

    <TrackingProviderEditorDialog
      v-model:open="trackingProviderDialogOpen"
      :mode="trackingProviderDialogMode"
      :form="trackingProviderForm"
      :errors="trackingProviderErrors"
      :webhook-url="trackingWebhookUrl(trackingProviderForm)"
      :submitting="trackingProviderSubmitting"
      @submit="saveTrackingProvider"
      @clear-error="clearTrackingProviderError"
    />

    <TrackingCarrierMappingEditorDialog
      v-model:open="trackingCarrierMappingDialogOpen"
      :mode="trackingCarrierMappingDialogMode"
      :form="trackingCarrierMappingForm"
      :errors="trackingCarrierMappingErrors"
      :providers="trackingProviders"
      :carriers="carriers"
      :carrier-services="carrierServices"
      :submitting="trackingCarrierMappingSubmitting"
      @submit="saveTrackingCarrierMapping"
      @clear-error="clearTrackingCarrierMappingError"
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

    <PackagingRuleAppliesDialog
      v-model:open="packagingAppliesDialogOpen"
      :rule="packagingAppliesRule"
      @updated="fetchPackagingRules"
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
import { computed, onMounted, reactive, ref } from 'vue'
import { toast } from 'vue-sonner'
import {
  Calculator,
  CircleCheck,
  CircleDashed,
  ClipboardList,
  Copy,
  Link2,
  MapPin,
  Package,
  Pencil,
  Plus,
  Radar,
  RefreshCw,
  Route,
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
import CarrierServiceEditorDialog from '@/components/admin/shipping/CarrierServiceEditorDialog.vue'
import PackagingRuleAppliesDialog from '@/components/admin/shipping/PackagingRuleAppliesDialog.vue'
import PackagingRuleEditorDialog from '@/components/admin/shipping/PackagingRuleEditorDialog.vue'
import ShippingQuoteCalculator from '@/components/admin/shipping/ShippingQuoteCalculator.vue'
import ShippingTemplateBindingEditorDialog from '@/components/admin/shipping/ShippingTemplateBindingEditorDialog.vue'
import ShippingTemplateEditorDialog from '@/components/admin/shipping/ShippingTemplateEditorDialog.vue'
import ShippingZoneEditorDialog from '@/components/admin/shipping/ShippingZoneEditorDialog.vue'
import TrackingCarrierMappingEditorDialog from '@/components/admin/shipping/TrackingCarrierMappingEditorDialog.vue'
import TrackingProviderEditorDialog from '@/components/admin/shipping/TrackingProviderEditorDialog.vue'
import TrackingShipmentsPanel from '@/components/admin/shipping/TrackingShipmentsPanel.vue'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Table, TableBody, TableCell, TableEmpty, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { useAuthStore } from '@/stores/auth'

const authStore = useAuthStore()
const activeTab = ref('overview')
const templates = ref([])
const zones = ref([])
const templateBindings = ref([])
const carriers = ref([])
const carrierServices = ref([])
const trackingProviders = ref([])
const trackingCarrierMappings = ref([])
const trackingShipmentsCount = ref(0)
const packagingRules = ref([])
const refreshing = ref(false)
const trackingShipmentsPanelRef = ref(null)
const loading = reactive({
  templates: false,
  zones: false,
  bindings: false,
  carriers: false,
  services: false,
  tracking: false,
  trackingMappings: false,
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

const carrierServiceDialogOpen = ref(false)
const carrierServiceDialogMode = ref('create')
const carrierServiceSubmitting = ref(false)
const carrierServiceErrors = reactive({})
const carrierServiceForm = reactive(defaultCarrierServiceForm())

const trackingProviderDialogOpen = ref(false)
const trackingProviderDialogMode = ref('create')
const trackingProviderSubmitting = ref(false)
const trackingProviderErrors = reactive({})
const trackingProviderForm = reactive(defaultTrackingProviderForm())

const trackingCarrierMappingDialogOpen = ref(false)
const trackingCarrierMappingDialogMode = ref('create')
const trackingCarrierMappingSubmitting = ref(false)
const trackingCarrierMappingErrors = reactive({})
const trackingCarrierMappingForm = reactive(defaultTrackingCarrierMappingForm())

const packagingDialogOpen = ref(false)
const packagingDialogMode = ref('create')
const packagingSubmitting = ref(false)
const packagingErrors = reactive({})
const packagingForm = reactive(defaultPackagingForm())
const packagingAppliesDialogOpen = ref(false)
const packagingAppliesRule = ref(null)

const deleteDialogOpen = ref(false)
const deleteTarget = ref(null)
const deleteType = ref('')
const deleteLoading = ref(false)

const hasPermission = (permission) => authStore.hasPermission(permission)

const handleTrackingShipmentsCountChange = (count) => {
  trackingShipmentsCount.value = Number(count || 0)
}

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
    key: 'trackingProviders',
    label: '追踪配置',
    value: trackingProviders.value.length,
    icon: Radar,
    tone: 'amber',
  },
  {
    key: 'trackingMappings',
    label: '承运商映射',
    value: trackingCarrierMappings.value.length,
    icon: Link2,
    tone: 'blue',
  },
  {
    key: 'trackingShipments',
    label: '追踪任务',
    value: trackingShipmentsCount.value,
    icon: RefreshCw,
    tone: 'amber',
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
  if (deleteType.value === 'carrierService') return '删除线路服务？'
  if (deleteType.value === 'trackingProvider') return '删除追踪配置？'
  if (deleteType.value === 'trackingCarrierMapping') return '删除承运商映射？'
  if (deleteType.value === 'packaging') return '删除包装规则？'
  return '确认删除？'
})

const deleteDialogDescription = computed(() => {
  const name = deleteTarget.value?.name || deleteTarget.value?.provider_name || deleteTarget.value?.provider_carrier_code || deleteTarget.value?.service_name || deleteTarget.value?.rule_name || '当前记录'
  if (deleteLoading.value) return '正在删除，请稍候。'
  return `将删除「${name}」。这个操作不可撤销，请确认没有正在使用它的运费规则或订单流程。`
})

function defaultCarrierForm() {
  return {
    id: null,
    name: '',
    code: '',
    tracking_url: '',
    contact: '',
    phone: '',
    email: '',
    service_area: '',
    enabled: true,
    sort_order: 0,
  }
}

function defaultCarrierServiceForm() {
  return {
    id: null,
    carrier_id: carriers.value[0]?.id ? String(carriers.value[0].id) : '',
    template_id: 'none',
    service_code: '',
    service_name: '',
    route_name: '',
    countries: '[]',
    currency: 'USD',
    billing_mode: 'actual_weight',
    first_weight_grams: 0,
    additional_weight_grams: 0,
    min_charge_weight_grams: 0,
    volumetric_divisor: 6000,
    fuel_surcharge_percent: 0,
    remote_surcharge: 0,
    eta_min_days: 0,
    eta_max_days: 0,
    enabled: true,
    sort_order: 0,
    description: '',
  }
}

function defaultTrackingProviderForm() {
  return {
    id: null,
    provider_code: '17TRACK',
    provider_name: '17TRACK',
    environment: 'production',
    base_url: '',
    api_key: '',
    webhook_secret: '',
    webhook_enabled: false,
    auto_register: false,
    polling_enabled: false,
    polling_interval_minutes: 60,
    request_timeout_seconds: 15,
    enabled: true,
    sort_order: 0,
    description: '',
  }
}

function defaultTrackingCarrierMappingForm() {
  return {
    id: null,
    provider_id: trackingProviders.value[0]?.id ? String(trackingProviders.value[0].id) : '',
    scope: 'carrier',
    carrier_id: carriers.value[0]?.id ? String(carriers.value[0].id) : '',
    carrier_service_id: carrierServices.value[0]?.id ? String(carrierServices.value[0].id) : '',
    provider_carrier_code: '',
    provider_carrier_name: '',
    enabled: true,
    priority: 0,
    description: '',
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

const clearCarrierServiceError = (field) => {
  delete carrierServiceErrors[field]
}

const clearTrackingProviderError = (field) => {
  delete trackingProviderErrors[field]
}

const clearTrackingCarrierMappingError = (field) => {
  delete trackingCarrierMappingErrors[field]
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

const fetchCarrierServices = async () => {
  loading.services = true
  try {
    carrierServices.value = await shippingApi.listCarrierServices()
  } catch (error) {
    console.error('Failed to fetch carrier services:', error)
  } finally {
    loading.services = false
  }
}

const fetchTrackingProviders = async () => {
  loading.tracking = true
  try {
    trackingProviders.value = await shippingApi.listTrackingProviders()
  } catch (error) {
    console.error('Failed to fetch tracking providers:', error)
  } finally {
    loading.tracking = false
  }
}

const fetchTrackingCarrierMappings = async () => {
  loading.trackingMappings = true
  try {
    trackingCarrierMappings.value = await shippingApi.listTrackingCarrierMappings()
  } catch (error) {
    console.error('Failed to fetch tracking carrier mappings:', error)
  } finally {
    loading.trackingMappings = false
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
    } else if (activeTab.value === 'services') {
      await Promise.all([fetchCarrierServices(), fetchCarriers(), fetchTemplates()])
    } else if (activeTab.value === 'tracking') {
      await Promise.all([fetchTrackingProviders(), fetchTrackingCarrierMappings(), fetchCarriers(), fetchCarrierServices()])
    } else if (activeTab.value === 'trackingShipments') {
      await Promise.all([
        trackingShipmentsPanelRef.value?.refresh?.(),
        fetchTrackingProviders(),
        fetchCarriers(),
        fetchCarrierServices(),
      ])
    } else if (activeTab.value === 'packaging') {
      await fetchPackagingRules()
    } else {
      await Promise.all([
        fetchTemplates(),
        fetchZones(),
        fetchTemplateBindings(),
        fetchCarriers(),
        fetchCarrierServices(),
        fetchTrackingProviders(),
        fetchTrackingCarrierMappings(),
        fetchPackagingRules(),
      ])
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

const showCreateTrackingProviderDialog = () => {
  trackingProviderDialogMode.value = 'create'
  resetReactive(trackingProviderForm, defaultTrackingProviderForm())
  clearErrors(trackingProviderErrors)
  trackingProviderDialogOpen.value = true
}

const showEditTrackingProviderDialog = (provider) => {
  trackingProviderDialogMode.value = 'edit'
  resetReactive(trackingProviderForm, {
    ...defaultTrackingProviderForm(),
    ...provider,
    webhook_enabled: provider.webhook_enabled === true,
    auto_register: provider.auto_register === true,
    polling_enabled: provider.polling_enabled === true,
    enabled: provider.enabled !== false,
    polling_interval_minutes: Number(provider.polling_interval_minutes || 60),
    request_timeout_seconds: Number(provider.request_timeout_seconds || 15),
    sort_order: Number(provider.sort_order || 0),
  })
  clearErrors(trackingProviderErrors)
  trackingProviderDialogOpen.value = true
}

const validateTrackingProvider = () => {
  clearErrors(trackingProviderErrors)
  if (!trackingProviderForm.provider_name?.trim()) trackingProviderErrors.provider_name = '请输入 Provider 名称'
  if (!trackingProviderForm.provider_code?.trim()) trackingProviderErrors.provider_code = '请输入 Provider 代码'
  if (!['production', 'sandbox'].includes(trackingProviderForm.environment)) trackingProviderErrors.environment = '请选择 Provider 环境'

  const numericFields = ['polling_interval_minutes', 'request_timeout_seconds', 'sort_order']
  const hasInvalidNumber = numericFields.some((field) => Number(trackingProviderForm[field] || 0) < 0)
  if (hasInvalidNumber) {
    toast.error('追踪配置的数字字段不能小于 0')
    return false
  }

  return Object.keys(trackingProviderErrors).length === 0
}

const buildTrackingProviderPayload = () => ({
  provider_code: trackingProviderForm.provider_code?.trim().toUpperCase() || '',
  provider_name: trackingProviderForm.provider_name?.trim() || '',
  environment: trackingProviderForm.environment || 'production',
  base_url: trackingProviderForm.base_url?.trim() || '',
  api_key: trackingProviderForm.api_key?.trim() || '',
  webhook_secret: trackingProviderForm.webhook_secret?.trim() || '',
  webhook_enabled: Boolean(trackingProviderForm.webhook_enabled),
  auto_register: Boolean(trackingProviderForm.auto_register),
  polling_enabled: Boolean(trackingProviderForm.polling_enabled),
  polling_interval_minutes: Number(trackingProviderForm.polling_interval_minutes || 60),
  request_timeout_seconds: Number(trackingProviderForm.request_timeout_seconds || 15),
  enabled: Boolean(trackingProviderForm.enabled),
  sort_order: Number(trackingProviderForm.sort_order || 0),
  description: trackingProviderForm.description || '',
})

const saveTrackingProvider = async () => {
  if (!validateTrackingProvider()) return

  trackingProviderSubmitting.value = true
  try {
    const payload = buildTrackingProviderPayload()
    if (trackingProviderDialogMode.value === 'create') {
      await shippingApi.createTrackingProvider(payload)
      toast.success('追踪配置已创建')
    } else {
      await shippingApi.updateTrackingProvider(trackingProviderForm.id, payload)
      toast.success('追踪配置已更新')
    }

    trackingProviderDialogOpen.value = false
    await Promise.all([fetchTrackingProviders(), fetchTrackingCarrierMappings()])
  } catch (error) {
    console.error('Failed to save tracking provider:', error)
  } finally {
    trackingProviderSubmitting.value = false
  }
}

const showCreateTrackingCarrierMappingDialog = () => {
  trackingCarrierMappingDialogMode.value = 'create'
  const form = defaultTrackingCarrierMappingForm()
  if (!form.carrier_id && form.carrier_service_id) {
    form.scope = 'carrier_service'
  }
  resetReactive(trackingCarrierMappingForm, form)
  clearErrors(trackingCarrierMappingErrors)
  trackingCarrierMappingDialogOpen.value = true
}

const showEditTrackingCarrierMappingDialog = (mapping) => {
  trackingCarrierMappingDialogMode.value = 'edit'
  resetReactive(trackingCarrierMappingForm, {
    ...defaultTrackingCarrierMappingForm(),
    ...mapping,
    provider_id: mapping.provider_id ? String(mapping.provider_id) : '',
    carrier_id: mapping.carrier_id ? String(mapping.carrier_id) : '',
    carrier_service_id: mapping.carrier_service_id ? String(mapping.carrier_service_id) : '',
    priority: Number(mapping.priority || 0),
    enabled: mapping.enabled !== false,
  })
  clearErrors(trackingCarrierMappingErrors)
  trackingCarrierMappingDialogOpen.value = true
}

const validateTrackingCarrierMapping = () => {
  clearErrors(trackingCarrierMappingErrors)
  if (!nullablePositiveID(trackingCarrierMappingForm.provider_id)) trackingCarrierMappingErrors.provider_id = '请选择追踪 Provider'
  if (!['carrier', 'carrier_service'].includes(trackingCarrierMappingForm.scope)) trackingCarrierMappingErrors.scope = '请选择映射层级'
  if (trackingCarrierMappingForm.scope === 'carrier' && !nullablePositiveID(trackingCarrierMappingForm.carrier_id)) {
    trackingCarrierMappingErrors.carrier_id = '请选择本地承运商'
  }
  if (trackingCarrierMappingForm.scope === 'carrier_service' && !nullablePositiveID(trackingCarrierMappingForm.carrier_service_id)) {
    trackingCarrierMappingErrors.carrier_service_id = '请选择本地线路服务'
  }
  if (!trackingCarrierMappingForm.provider_carrier_code?.trim()) {
    trackingCarrierMappingErrors.provider_carrier_code = '请输入 Provider Carrier Code'
  }
  if (Number(trackingCarrierMappingForm.priority || 0) < 0) {
    toast.error('承运商映射优先级不能小于 0')
    return false
  }
  return Object.keys(trackingCarrierMappingErrors).length === 0
}

const buildTrackingCarrierMappingPayload = () => ({
  provider_id: nullablePositiveID(trackingCarrierMappingForm.provider_id),
  scope: trackingCarrierMappingForm.scope || 'carrier',
  carrier_id: trackingCarrierMappingForm.scope === 'carrier' ? nullablePositiveID(trackingCarrierMappingForm.carrier_id) : null,
  carrier_service_id: trackingCarrierMappingForm.scope === 'carrier_service' ? nullablePositiveID(trackingCarrierMappingForm.carrier_service_id) : null,
  provider_carrier_code: trackingCarrierMappingForm.provider_carrier_code?.trim() || '',
  provider_carrier_name: trackingCarrierMappingForm.provider_carrier_name?.trim() || '',
  enabled: Boolean(trackingCarrierMappingForm.enabled),
  priority: Number(trackingCarrierMappingForm.priority || 0),
  description: trackingCarrierMappingForm.description || '',
})

const saveTrackingCarrierMapping = async () => {
  if (!validateTrackingCarrierMapping()) return

  trackingCarrierMappingSubmitting.value = true
  try {
    const payload = buildTrackingCarrierMappingPayload()
    if (trackingCarrierMappingDialogMode.value === 'create') {
      await shippingApi.createTrackingCarrierMapping(payload)
      toast.success('承运商映射已创建')
    } else {
      await shippingApi.updateTrackingCarrierMapping(trackingCarrierMappingForm.id, payload)
      toast.success('承运商映射已更新')
    }

    trackingCarrierMappingDialogOpen.value = false
    await fetchTrackingCarrierMappings()
  } catch (error) {
    console.error('Failed to save tracking carrier mapping:', error)
  } finally {
    trackingCarrierMappingSubmitting.value = false
  }
}

const showCreateCarrierServiceDialog = () => {
  carrierServiceDialogMode.value = 'create'
  resetReactive(carrierServiceForm, defaultCarrierServiceForm())
  clearErrors(carrierServiceErrors)
  carrierServiceDialogOpen.value = true
}

const showEditCarrierServiceDialog = (service) => {
  carrierServiceDialogMode.value = 'edit'
  resetReactive(carrierServiceForm, {
    ...defaultCarrierServiceForm(),
    ...service,
    carrier_id: service.carrier_id ? String(service.carrier_id) : '',
    template_id: service.template_id ? String(service.template_id) : 'none',
    first_weight_grams: Number(service.first_weight_grams || 0),
    additional_weight_grams: Number(service.additional_weight_grams || 0),
    min_charge_weight_grams: Number(service.min_charge_weight_grams || 0),
    volumetric_divisor: Number(service.volumetric_divisor || 6000),
    fuel_surcharge_percent: Number(service.fuel_surcharge_percent || 0),
    remote_surcharge: Number(service.remote_surcharge || 0),
    eta_min_days: Number(service.eta_min_days || 0),
    eta_max_days: Number(service.eta_max_days || 0),
    enabled: service.enabled !== false,
    sort_order: Number(service.sort_order || 0),
  })
  clearErrors(carrierServiceErrors)
  carrierServiceDialogOpen.value = true
}

const validateCarrierService = () => {
  clearErrors(carrierServiceErrors)
  if (!nullablePositiveID(carrierServiceForm.carrier_id)) carrierServiceErrors.carrier_id = '请选择承运商'
  if (!carrierServiceForm.service_code?.trim()) carrierServiceErrors.service_code = '请输入线路代码'
  if (!carrierServiceForm.service_name?.trim()) carrierServiceErrors.service_name = '请输入线路名称'
  if (!['actual_weight', 'volumetric_weight', 'greater_of_actual_and_volumetric'].includes(carrierServiceForm.billing_mode)) {
    carrierServiceErrors.billing_mode = '请选择计费模式'
  }
  if (Number(carrierServiceForm.eta_max_days || 0) > 0 && Number(carrierServiceForm.eta_min_days || 0) > 0 && Number(carrierServiceForm.eta_max_days) < Number(carrierServiceForm.eta_min_days)) {
    carrierServiceErrors.eta_max_days = '最长时效不能小于最短时效'
  }

  const numericFields = [
    'first_weight_grams',
    'additional_weight_grams',
    'min_charge_weight_grams',
    'volumetric_divisor',
    'fuel_surcharge_percent',
    'remote_surcharge',
    'eta_min_days',
    'eta_max_days',
    'sort_order',
  ]
  const hasNegative = numericFields.some((field) => Number(carrierServiceForm[field] || 0) < 0)
  if (hasNegative) {
    toast.error('线路服务的数字字段不能小于 0')
    return false
  }

  return Object.keys(carrierServiceErrors).length === 0
}

const buildCarrierServicePayload = () => ({
  carrier_id: nullablePositiveID(carrierServiceForm.carrier_id),
  template_id: carrierServiceForm.template_id === 'none' ? null : nullablePositiveID(carrierServiceForm.template_id),
  service_code: carrierServiceForm.service_code?.trim().toUpperCase() || '',
  service_name: carrierServiceForm.service_name?.trim() || '',
  route_name: carrierServiceForm.route_name?.trim() || '',
  countries: carrierServiceForm.countries?.trim() || '[]',
  currency: carrierServiceForm.currency?.trim().toUpperCase() || 'USD',
  billing_mode: carrierServiceForm.billing_mode || 'actual_weight',
  first_weight_grams: Number(carrierServiceForm.first_weight_grams || 0),
  additional_weight_grams: Number(carrierServiceForm.additional_weight_grams || 0),
  min_charge_weight_grams: Number(carrierServiceForm.min_charge_weight_grams || 0),
  volumetric_divisor: Number(carrierServiceForm.volumetric_divisor || 6000),
  fuel_surcharge_percent: Number(carrierServiceForm.fuel_surcharge_percent || 0),
  remote_surcharge: Number(carrierServiceForm.remote_surcharge || 0),
  eta_min_days: Number(carrierServiceForm.eta_min_days || 0),
  eta_max_days: Number(carrierServiceForm.eta_max_days || 0),
  enabled: Boolean(carrierServiceForm.enabled),
  sort_order: Number(carrierServiceForm.sort_order || 0),
  description: carrierServiceForm.description || '',
})

const saveCarrierService = async () => {
  if (!validateCarrierService()) return

  carrierServiceSubmitting.value = true
  try {
    const payload = buildCarrierServicePayload()
    if (carrierServiceDialogMode.value === 'create') {
      await shippingApi.createCarrierService(payload)
      toast.success('线路服务已创建')
    } else {
      await shippingApi.updateCarrierService(carrierServiceForm.id, payload)
      toast.success('线路服务已更新')
    }

    carrierServiceDialogOpen.value = false
    await fetchCarrierServices()
  } catch (error) {
    console.error('Failed to save carrier service:', error)
  } finally {
    carrierServiceSubmitting.value = false
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

const showPackagingAppliesDialog = (rule) => {
  packagingAppliesRule.value = rule
  packagingAppliesDialogOpen.value = true
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
    } else if (deleteType.value === 'carrierService') {
      await shippingApi.deleteCarrierService(deleteTarget.value.id)
      toast.success('线路服务已删除')
      await fetchCarrierServices()
    } else if (deleteType.value === 'trackingProvider') {
      await shippingApi.deleteTrackingProvider(deleteTarget.value.id)
      toast.success('追踪配置已删除')
      await Promise.all([fetchTrackingProviders(), fetchTrackingCarrierMappings()])
    } else if (deleteType.value === 'trackingCarrierMapping') {
      await shippingApi.deleteTrackingCarrierMapping(deleteTarget.value.id)
      toast.success('承运商映射已删除')
      await fetchTrackingCarrierMappings()
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
const formatDate = (value) => value ? new Date(value).toLocaleString('zh-CN') : '-'

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

const carrierServiceCarrierName = (service) => {
  if (service.carrier?.name) return service.carrier.name
  const carrier = carriers.value.find((item) => Number(item.id) === Number(service.carrier_id))
  return carrier?.name || `Carrier #${service.carrier_id || '-'}`
}

const carrierServiceTemplateName = (service) => {
  if (service.template?.name) return service.template.name
  if (!service.template_id) return '未绑定模板'
  const template = templates.value.find((item) => Number(item.id) === Number(service.template_id))
  return template?.name || `Template #${service.template_id}`
}

const billingModeLabel = (mode) => {
  const labels = {
    actual_weight: '实重计费',
    volumetric_weight: '体积重计费',
    greater_of_actual_and_volumetric: '实重/体积重取大',
  }
  return labels[mode] || mode || '-'
}

const trackingEnvironmentLabel = (environment) => {
  const labels = {
    production: 'Production',
    sandbox: 'Sandbox',
  }
  return labels[environment] || environment || '-'
}

const apiBaseUrl = () => {
  const configured = String(import.meta.env.VITE_API_BASE_URL || '').trim().replace(/\/+$/, '')
  if (/^https?:\/\//i.test(configured)) {
    try {
      return new URL(configured).origin
    } catch {
      return configured
    }
  }
  if (typeof window !== 'undefined' && window.location?.origin) return window.location.origin
  return ''
}

const trackingWebhookUrl = (provider) => {
  const providerCode = String(provider?.provider_code || '').trim()
  if (!providerCode) return ''

  const path = `/api/v1/shipping/webhook/${encodeURIComponent(providerCode)}`
  const base = apiBaseUrl()
  return base ? `${base}${path}` : path
}

const trackingProviderHasApiKey = (provider) => provider?.api_key_configured === true || Boolean(provider?.api_key)
const trackingProviderHasWebhookSecret = (provider) => provider?.webhook_secret_configured === true || Boolean(provider?.webhook_secret)

const copyTrackingWebhookUrl = async (provider) => {
  const url = trackingWebhookUrl(provider)
  if (!url) return

  try {
    await navigator.clipboard.writeText(url)
    toast.success('Webhook 地址已复制')
  } catch (error) {
    console.error('Failed to copy tracking webhook URL:', error)
    toast.error('复制失败，请手动复制')
  }
}

const formatTrackingSyncPolicy = (provider) => {
  const policies = []
  if (provider.auto_register) policies.push('自动注册追踪号')
  if (provider.webhook_enabled) policies.push('Webhook 推送')
  if (provider.polling_enabled) policies.push(`轮询 ${Number(provider.polling_interval_minutes || 60)} 分钟`)
  if (Number(provider.request_timeout_seconds || 0) > 0) policies.push(`超时 ${Number(provider.request_timeout_seconds)} 秒`)
  return policies.length ? policies.join(' / ') : '暂未启用同步策略'
}

const trackingProviderName = (mapping) => {
  if (mapping.provider?.provider_name) return mapping.provider.provider_name
  const provider = trackingProviders.value.find((item) => Number(item.id) === Number(mapping.provider_id))
  return provider?.provider_name || `Provider #${mapping.provider_id || '-'}`
}

const trackingMappingScopeLabel = (scope) => {
  const labels = {
    carrier: '承运商映射',
    carrier_service: '线路服务映射',
  }
  return labels[scope] || scope || '-'
}

const trackingMappingLocalTargetLabel = (mapping) => {
  if (mapping.scope === 'carrier_service') {
    if (mapping.carrier_service?.service_name) return mapping.carrier_service.service_name
    const service = carrierServices.value.find((item) => Number(item.id) === Number(mapping.carrier_service_id))
    return service?.service_name || `Carrier service #${mapping.carrier_service_id || '-'}`
  }

  if (mapping.carrier?.name) return mapping.carrier.name
  const carrier = carriers.value.find((item) => Number(item.id) === Number(mapping.carrier_id))
  return carrier?.name || `Carrier #${mapping.carrier_id || '-'}`
}

const formatGrams = (value) => `${Number(value || 0).toLocaleString()} g`

const formatServiceWeightStep = (service) => {
  const first = Number(service.first_weight_grams || 0)
  const additional = Number(service.additional_weight_grams || 0)
  const min = Number(service.min_charge_weight_grams || 0)
  const parts = []
  if (first > 0) parts.push(`首 ${formatGrams(first)}`)
  if (additional > 0) parts.push(`续 ${formatGrams(additional)}`)
  if (min > 0) parts.push(`最低 ${formatGrams(min)}`)
  return parts.length ? parts.join(' / ') : '未设置'
}

const formatVolumetricDivisor = (service) => {
  const divisor = Number(service.volumetric_divisor || 0)
  const surcharge = Number(service.fuel_surcharge_percent || 0)
  const remote = Number(service.remote_surcharge || 0)
  const parts = []
  if (divisor > 0) parts.push(`÷${divisor}`)
  if (surcharge > 0) parts.push(`燃油 ${surcharge.toFixed(3)}%`)
  if (remote > 0) parts.push(`偏远 ${formatMoney(remote)}`)
  return parts.length ? parts.join(' / ') : '未设置'
}

const formatEta = (service) => {
  const min = Number(service.eta_min_days || 0)
  const max = Number(service.eta_max_days || 0)
  if (min > 0 && max > 0) return `${min}-${max} 天`
  if (min > 0) return `${min}+ 天`
  if (max > 0) return `≤ ${max} 天`
  return '-'
}

const formatWeight = (value) => `${Number(value || 0).toFixed(3)} kg`

const formatDimensions = (rule) => {
  const length = Number(rule.box_length || 0).toFixed(2)
  const width = Number(rule.box_width || 0).toFixed(2)
  const height = Number(rule.box_height || 0).toFixed(2)
  return `${length} × ${width} × ${height} cm`
}

const appliesCount = (rule) => Array.isArray(rule.applies) ? rule.applies.length : 0

onMounted(() => Promise.all([
  fetchTemplates(),
  fetchZones(),
  fetchTemplateBindings(),
  fetchCarriers(),
  fetchCarrierServices(),
  fetchTrackingProviders(),
  fetchTrackingCarrierMappings(),
  fetchPackagingRules(),
]))
</script>
