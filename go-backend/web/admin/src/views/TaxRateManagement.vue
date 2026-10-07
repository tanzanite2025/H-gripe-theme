<template>
  <div class="space-y-5">
    <AdminPageHeader
      title="税率管理"
      description="分开管理税率数据源、参考快照、地区覆盖和 Checkout 实际使用的规则。"
    >
      <template #actions>
        <Button
          v-if="activeTab === 'config'"
          :disabled="loadingConfiguration || savingConfiguration || !configurationDirty || !canEdit"
          @click="saveTaxRateSourceConfiguration"
        >
          <LoaderCircle v-if="savingConfiguration" class="size-4 animate-spin" />
          <Save v-else class="size-4" />
          {{ savingConfiguration ? '保存中' : '保存配置' }}
        </Button>
        <Button
          v-else-if="activeTab === 'snapshot'"
          variant="outline"
          :disabled="loadingSnapshot"
          @click="loadCurrentTaxRateSourceSnapshot"
        >
          <LoaderCircle v-if="loadingSnapshot" class="size-4 animate-spin" />
          <RefreshCw v-else class="size-4" />
          刷新快照
        </Button>
        <Button
          v-else-if="activeTab === 'coverage'"
          variant="outline"
          :disabled="loadingTaxRateCoverage"
          @click="loadTaxRateCoverage"
        >
          <LoaderCircle v-if="loadingTaxRateCoverage" class="size-4 animate-spin" />
          <RefreshCw v-else class="size-4" />
          刷新覆盖情况
        </Button>
        <Button
          v-else
          :disabled="!canEdit"
          @click="startCreatingTaxRateRule"
        >
          <Plus class="size-4" />
          新建结算规则
        </Button>
      </template>
    </AdminPageHeader>

    <nav class="flex flex-wrap gap-2" aria-label="税率管理页签">
      <button
        v-for="tab in taxRateManagementTabs"
        :key="tab.value"
        type="button"
        class="rounded-full border px-4 py-2 text-xs font-black transition-colors"
        :class="activeTab === tab.value
          ? 'border-primary bg-primary text-primary-foreground shadow-sm'
          : 'border-border bg-card text-muted-foreground hover:bg-muted hover:text-foreground'"
        @click="activeTab = tab.value"
      >
        {{ tab.label }}
      </button>
    </nav>

    <template v-if="activeTab === 'config'">
      <section class="rounded-3xl border border-sky-500/20 bg-sky-50/70 p-5 dark:bg-sky-950/20">
        <div class="flex flex-col gap-5 xl:flex-row xl:items-start xl:justify-between">
          <div class="flex min-w-0 items-start gap-3">
            <span class="flex size-11 shrink-0 items-center justify-center rounded-2xl bg-slate-950 text-white">
              <Database class="size-5" />
            </span>
            <div class="min-w-0">
              <p class="text-[10px] font-black uppercase tracking-[0.18em] text-sky-800/70 dark:text-sky-200/70">Tax rate data source</p>
              <h2 class="mt-1 text-xl font-black tracking-tight text-slate-950 dark:text-white">{{ sourceConfiguration.provider_name || 'VATcomply' }}</h2>
              <p class="mt-1 max-w-2xl text-xs leading-relaxed text-slate-600 dark:text-slate-300">
                当前接入欧盟 VAT 参考数据接口，无需 API Key。接口地址由系统固定维护，不接受任意 URL。
              </p>
            </div>
          </div>
          <div class="flex flex-wrap items-center gap-2">
            <span class="rounded-full border border-emerald-600/20 bg-white/80 px-3 py-1.5 text-[10px] font-black text-emerald-700 dark:bg-card dark:text-emerald-300">
              无需 API Key
            </span>
            <a
              v-if="sourceConfiguration.endpoint"
              :href="sourceConfiguration.endpoint"
              target="_blank"
              rel="noreferrer"
              class="inline-flex items-center gap-1 rounded-full border border-sky-600/20 bg-white/80 px-3 py-1.5 text-[10px] font-black text-sky-800 hover:bg-white dark:bg-card dark:text-sky-200"
            >
              查看数据接口 <ExternalLink class="size-3" />
            </a>
            <a
              v-if="sourceConfiguration.documentation_url"
              :href="sourceConfiguration.documentation_url"
              target="_blank"
              rel="noreferrer"
              class="inline-flex items-center gap-1 rounded-full border border-sky-600/20 bg-white/80 px-3 py-1.5 text-[10px] font-black text-sky-800 hover:bg-white dark:bg-card dark:text-sky-200"
            >
              接口说明 <ExternalLink class="size-3" />
            </a>
          </div>
        </div>
      </section>

      <section class="grid gap-4 xl:grid-cols-[minmax(0,1.1fr)_minmax(320px,0.9fr)]">
        <div class="rounded-3xl border bg-card p-5">
          <div class="flex items-start justify-between gap-4">
            <div>
              <h2 class="text-base font-black tracking-tight text-foreground">采集计划</h2>
              <p class="mt-1 text-xs leading-relaxed text-muted-foreground">启用后按所选间隔创建新的数据快照；也可以随时手动同步。</p>
            </div>
            <span
              class="shrink-0 rounded-full border px-3 py-1.5 text-[10px] font-black"
              :class="sourceConfiguration.enabled
                ? 'border-emerald-500/20 bg-emerald-500/10 text-emerald-700 dark:text-emerald-300'
                : 'border-border bg-muted text-muted-foreground'"
            >
              {{ sourceConfiguration.enabled ? '自动采集已启用' : '自动采集已停用' }}
            </span>
          </div>

          <div class="mt-5 grid gap-4 sm:grid-cols-2">
            <label class="flex items-center gap-3 rounded-2xl border bg-background p-4">
              <input
                v-model="sourceConfiguration.enabled"
                type="checkbox"
                :disabled="!canEdit || loadingConfiguration"
                class="size-4 rounded border-input accent-primary"
              >
              <span>
                <span class="block text-xs font-black text-foreground">启用定时采集</span>
                <span class="mt-1 block text-[11px] leading-relaxed text-muted-foreground">只生成来源快照，不发布到结算。</span>
              </span>
            </label>

            <label class="rounded-2xl border bg-background p-4">
              <span class="block text-xs font-black text-foreground">刷新间隔</span>
              <select
                v-model.number="sourceConfiguration.refresh_interval_hours"
                :disabled="!canEdit || loadingConfiguration"
                class="mt-2 h-9 w-full rounded-xl border border-input bg-card px-3 text-xs font-bold text-foreground outline-none focus-visible:ring-2 focus-visible:ring-ring/50"
              >
                <option v-for="option in refreshIntervalOptions" :key="option.hours" :value="option.hours">{{ option.label }}</option>
              </select>
            </label>
          </div>

          <div class="mt-4 flex flex-col gap-3 rounded-2xl border border-dashed p-4 sm:flex-row sm:items-center sm:justify-between">
            <p class="text-xs leading-relaxed text-muted-foreground">
              {{ sourceConfiguration.last_successful_sync_at
                ? `上次成功采集：${formatTaxRateSnapshotTimestamp(sourceConfiguration.last_successful_sync_at)}`
                : '尚未成功采集；启用后后台会在下一次检查时开始采集。' }}
            </p>
            <Button
              v-if="canEdit"
              variant="outline"
              :disabled="syncingSnapshot || loadingConfiguration || !sourceConfiguration.enabled || configurationDirty"
              @click="syncTaxRateSourceSnapshot"
            >
              <LoaderCircle v-if="syncingSnapshot" class="size-4 animate-spin" />
              <RefreshCw v-else class="size-4" />
              {{ syncingSnapshot ? '同步中' : '立即同步' }}
            </Button>
          </div>

          <p v-if="sourceConfiguration.last_sync_error" class="mt-4 rounded-2xl border border-rose-500/20 bg-rose-500/5 px-4 py-3 text-xs leading-relaxed text-rose-700 dark:text-rose-300">
            最近一次采集失败：{{ sourceConfiguration.last_sync_error }}
          </p>
          <p v-if="loadingConfiguration" class="mt-4 text-xs text-muted-foreground">正在读取税率数据源配置…</p>
        </div>

        <div class="space-y-4 rounded-3xl border bg-card p-5">
          <div>
            <h2 class="text-base font-black tracking-tight text-foreground">数据范围与职责</h2>
            <p class="mt-2 text-xs leading-relaxed text-muted-foreground">{{ sourceConfiguration.coverage_description || '当前数据源提供欧盟成员国 VAT 参考税率。' }}</p>
          </div>
          <div class="rounded-2xl border border-amber-500/20 bg-amber-500/5 p-4">
            <div class="flex items-center gap-2 text-amber-800 dark:text-amber-200">
              <ShieldCheck class="size-4" />
              <span class="text-xs font-black">快照只供核对</span>
            </div>
            <p class="mt-2 text-xs leading-relaxed text-muted-foreground">
              这份来源数据不会覆盖现有税率记录，也不会改变 Checkout 的税费计算或缺失税率处理。它目前不覆盖美国州税、英国、加拿大、中国等市场。
            </p>
          </div>
          <div class="rounded-2xl border bg-muted/40 p-4">
            <p class="text-[10px] font-black uppercase tracking-widest text-muted-foreground">最近检查时间</p>
            <p class="mt-1 text-sm font-black text-foreground">{{ formatTaxRateSnapshotTimestamp(sourceConfiguration.last_checked_at) || '尚无记录' }}</p>
            <p class="mt-3 text-[10px] font-black uppercase tracking-widest text-muted-foreground">自动采集频率</p>
            <p class="mt-1 text-sm font-black text-foreground">{{ formatRefreshInterval(sourceConfiguration.refresh_interval_hours) }}</p>
          </div>
        </div>
      </section>
    </template>

    <template v-else-if="activeTab === 'snapshot'">
      <section class="rounded-3xl border border-amber-500/20 bg-amber-500/5 p-4">
        <div class="flex items-start gap-3">
          <ShieldCheck class="mt-0.5 size-4 shrink-0 text-amber-700 dark:text-amber-300" />
          <p class="text-xs leading-relaxed text-muted-foreground">
            这里展示 VATcomply 最近一次有效采集的只读快照。该快照用于核对外部数据，不会被 Checkout 读取，也不会自动写入当前税率规则。
          </p>
        </div>
      </section>

      <div v-if="loadingSnapshot" class="rounded-3xl border bg-card p-10 text-center text-sm text-muted-foreground">
        正在读取税率快照…
      </div>

      <template v-else-if="currentSnapshot.snapshot">
        <section class="grid gap-3 sm:grid-cols-2 xl:grid-cols-4">
          <div class="rounded-2xl border bg-card p-4">
            <p class="text-[10px] font-black uppercase tracking-widest text-muted-foreground">快照版本</p>
            <p class="mt-2 break-all font-mono text-xs font-black text-foreground">{{ currentSnapshot.snapshot.version }}</p>
          </div>
          <div class="rounded-2xl border bg-card p-4">
            <p class="text-[10px] font-black uppercase tracking-widest text-muted-foreground">采集时间</p>
            <p class="mt-2 text-sm font-black text-foreground">{{ formatTaxRateSnapshotTimestamp(currentSnapshot.snapshot.captured_at) }}</p>
          </div>
          <div class="rounded-2xl border bg-card p-4">
            <p class="text-[10px] font-black uppercase tracking-widest text-muted-foreground">国家 / 地区</p>
            <p class="mt-2 text-sm font-black text-foreground">{{ currentSnapshot.snapshot.country_count }}</p>
          </div>
          <div class="rounded-2xl border bg-card p-4">
            <p class="text-[10px] font-black uppercase tracking-widest text-muted-foreground">税率记录</p>
            <p class="mt-2 text-sm font-black text-foreground">{{ currentSnapshot.snapshot.rate_count }}</p>
          </div>
        </section>

        <section class="overflow-hidden rounded-3xl border bg-card">
          <div class="flex flex-col gap-3 border-b p-4 sm:flex-row sm:items-center sm:justify-between">
            <div>
              <h2 class="text-sm font-black text-foreground">各国税率快照</h2>
              <p class="mt-1 text-xs text-muted-foreground">标准税率、其他税率及 API 返回的商品类别参考税率。</p>
            </div>
            <input
              v-model="countrySearchText"
              type="search"
              aria-label="搜索国家或地区"
              placeholder="搜索国家或代码"
              class="h-9 w-full rounded-full border border-input bg-background px-4 text-xs outline-none focus-visible:ring-2 focus-visible:ring-ring/50 sm:max-w-xs"
            >
          </div>

          <div class="overflow-x-auto">
            <table class="w-full min-w-[920px] text-left text-xs">
              <thead class="bg-muted/50 text-[10px] font-black uppercase tracking-wider text-muted-foreground">
                <tr>
                  <th class="px-4 py-3">国家 / 地区</th>
                  <th class="px-4 py-3">币种</th>
                  <th class="px-4 py-3">标准税率</th>
                  <th class="px-4 py-3">其他税率</th>
                  <th class="px-4 py-3">商品类别参考税率</th>
                </tr>
              </thead>
              <tbody class="divide-y divide-border">
                <tr v-for="country in filteredTaxRateCountryRows" :key="country.countryCode" class="align-top">
                  <td class="px-4 py-4">
                    <p class="font-black text-foreground">{{ country.countryName }}</p>
                    <p class="mt-1 font-mono text-[10px] text-muted-foreground">
                      {{ country.countryCode }}<span v-if="country.sourceCountryCode !== country.countryCode"> · API {{ country.sourceCountryCode }}</span>
                    </p>
                  </td>
                  <td class="px-4 py-4 font-mono font-bold text-foreground">{{ country.currency || '—' }}</td>
                  <td class="px-4 py-4 font-mono font-black text-foreground">{{ formatTaxRatePercentage(country.standardRate) }}</td>
                  <td class="px-4 py-4">
                    <div v-if="country.otherRates.length" class="flex flex-wrap gap-1.5">
                      <span v-for="rate in country.otherRates" :key="`${rate.rateType}-${rate.rateDecimal}`" class="rounded-full border bg-background px-2 py-1 font-mono text-[10px] font-bold text-foreground">
                        {{ formatTaxRateType(rate.rateType) }} {{ formatTaxRatePercentage(rate.rateDecimal) }}
                      </span>
                    </div>
                    <span v-else class="text-muted-foreground">—</span>
                  </td>
                  <td class="max-w-lg px-4 py-4">
                    <details v-if="country.categoryRates.length" class="group">
                      <summary class="cursor-pointer list-none font-bold text-sky-700 marker:hidden hover:text-sky-900 dark:text-sky-300 dark:hover:text-sky-100">
                        查看 {{ country.categoryRates.length }} 个类别
                      </summary>
                      <div class="mt-2 grid gap-1.5 rounded-xl border bg-muted/30 p-3 sm:grid-cols-2">
                        <p v-for="category in country.categoryRates" :key="category.name" class="leading-relaxed text-foreground">
                          <span class="font-semibold">{{ humanizeTaxRateCategory(category.name) }}</span>
                          <span class="ml-1 font-mono text-muted-foreground">{{ category.rates.map(formatTaxRatePercentage).join(' / ') }}</span>
                        </p>
                      </div>
                    </details>
                    <span v-else class="text-muted-foreground">—</span>
                  </td>
                </tr>
                <tr v-if="filteredTaxRateCountryRows.length === 0">
                  <td colspan="5" class="px-4 py-10 text-center text-muted-foreground">没有符合搜索条件的国家或地区。</td>
                </tr>
              </tbody>
            </table>
          </div>
          <div class="border-t px-4 py-3">
            <p class="break-all font-mono text-[10px] leading-relaxed text-muted-foreground">SHA-256 · {{ currentSnapshot.snapshot.content_sha256 }}</p>
          </div>
        </section>
      </template>

      <section v-else class="rounded-3xl border border-dashed bg-card p-10 text-center">
        <Database class="mx-auto size-8 text-muted-foreground/60" />
        <h2 class="mt-3 text-base font-black text-foreground">还没有可核对的快照</h2>
        <p class="mx-auto mt-2 max-w-lg text-xs leading-relaxed text-muted-foreground">请先在“数据源配置”页启用 VATcomply，并手动同步一次或等待定时采集。</p>
        <Button class="mt-4" @click="activeTab = 'config'">前往数据源配置</Button>
      </section>
    </template>

    <template v-else-if="activeTab === 'coverage'">
      <section class="rounded-3xl border border-amber-500/20 bg-amber-500/5 p-4">
        <div class="flex items-start gap-3">
          <ShieldCheck class="mt-0.5 size-4 shrink-0 text-amber-700 dark:text-amber-300" />
          <div class="space-y-1 text-xs leading-relaxed text-muted-foreground">
            <p>这里对照后台市场、当前快照和 Checkout 已启用规则。快照只用于发现数据缺口或数值差异；“数值一致”不代表该税率已被确认适用于本店商品和当地纳税义务。</p>
            <p>Checkout 只读取下方已启用规则。未匹配税率会阻止报价和下单；显式配置的 0% 仍是一条有效规则。</p>
          </div>
        </div>
      </section>

      <section v-if="taxRateCoverageError" class="rounded-2xl border border-rose-500/20 bg-rose-500/5 p-4 text-xs leading-relaxed text-rose-700 dark:text-rose-300">
        {{ taxRateCoverageError }}
      </section>

      <section v-if="enabledMarketCountryCount === 0 && !loadingTaxRateCoverage && !taxRateCoverageError" class="rounded-2xl border border-rose-500/20 bg-rose-500/5 p-4 text-xs leading-relaxed text-rose-700 dark:text-rose-300">
        后台目前没有启用且配置国家的销售市场。本页无法据此判断完整销售覆盖范围，请先检查“市场与本地化语种”配置。
      </section>

      <section v-if="!taxRateCoverageError" class="grid gap-3 sm:grid-cols-2 xl:grid-cols-4">
        <div class="rounded-2xl border bg-card p-4">
          <p class="text-[10px] font-black uppercase tracking-widest text-muted-foreground">启用市场国家</p>
          <p class="mt-2 text-2xl font-black text-foreground">{{ enabledMarketCountryCount }}</p>
        </div>
        <div class="rounded-2xl border bg-card p-4">
          <p class="text-[10px] font-black uppercase tracking-widest text-muted-foreground">有国家默认规则</p>
          <p class="mt-2 text-2xl font-black text-foreground">{{ enabledMarketCountriesWithDefaultRuleCount }}</p>
        </div>
        <div class="rounded-2xl border bg-card p-4">
          <p class="text-[10px] font-black uppercase tracking-widest text-muted-foreground">仅部分地区有规则</p>
          <p class="mt-2 text-2xl font-black text-foreground">{{ enabledMarketCountriesWithRegionalRulesOnlyCount }}</p>
        </div>
        <div class="rounded-2xl border bg-card p-4">
          <p class="text-[10px] font-black uppercase tracking-widest text-muted-foreground">快照含标准税率</p>
          <p class="mt-2 text-2xl font-black text-foreground">{{ enabledMarketCountriesWithSnapshotStandardRateCount }}</p>
        </div>
      </section>

      <section class="overflow-hidden rounded-3xl border bg-card">
        <div class="flex flex-col gap-3 border-b p-4 sm:flex-row sm:items-center sm:justify-between">
          <div>
            <h2 class="text-sm font-black text-foreground">地区规则覆盖核对</h2>
            <p class="mt-1 text-xs text-muted-foreground">国家默认规则会作为未命中更具体州省或邮编规则时的回退；仅有州省 / 邮编规则属于部分覆盖。</p>
            <p class="mt-1 text-[10px] text-muted-foreground">
              核对快照：{{ currentSnapshot.snapshot
                ? `${currentSnapshot.snapshot.version} · ${formatTaxRateSnapshotTimestamp(currentSnapshot.snapshot.captured_at)}`
                : '尚未采集' }}
            </p>
          </div>
          <input
            v-model="coverageSearchText"
            type="search"
            aria-label="搜索税率覆盖国家或地区"
            placeholder="搜索国家或代码"
            class="h-9 w-full rounded-full border border-input bg-background px-4 text-xs outline-none focus-visible:ring-2 focus-visible:ring-ring/50 sm:max-w-xs"
          >
        </div>

        <div v-if="loadingTaxRateCoverage" class="p-10 text-center text-sm text-muted-foreground">正在读取市场、快照和 Checkout 规则…</div>
        <div v-else-if="!taxRateCoverageError" class="overflow-x-auto">
          <table class="w-full min-w-[1100px] text-left text-xs">
            <thead class="bg-muted/50 text-[10px] font-black uppercase tracking-wider text-muted-foreground">
              <tr>
                <th class="px-4 py-3">国家 / 市场</th>
                <th class="px-4 py-3">快照标准税率</th>
                <th class="px-4 py-3">Checkout 覆盖</th>
                <th class="px-4 py-3">国家默认规则</th>
                <th class="px-4 py-3">数值核对</th>
                <th class="px-4 py-3 text-right">操作</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-border">
              <tr v-for="row in filteredTaxRateCoverageRows" :key="row.countryCode" class="align-top">
                <td class="px-4 py-4">
                  <p class="font-black text-foreground">{{ row.countryName }}</p>
                  <p class="mt-1 font-mono text-[10px] text-muted-foreground">{{ row.countryCode }}</p>
                  <div v-if="row.enabledMarketNames.length || row.disabledMarketNames.length" class="mt-2 flex flex-wrap gap-1">
                    <span v-for="marketName in row.enabledMarketNames" :key="`enabled-${marketName}`" class="rounded-full border border-emerald-500/20 bg-emerald-500/10 px-2 py-0.5 text-[9px] font-bold text-emerald-700 dark:text-emerald-300">{{ marketName }}</span>
                    <span v-for="marketName in row.disabledMarketNames" :key="`disabled-${marketName}`" class="rounded-full border bg-muted px-2 py-0.5 text-[9px] font-bold text-muted-foreground">{{ marketName }} · 停用</span>
                  </div>
                  <p v-else class="mt-2 text-[10px] text-muted-foreground">不在启用市场中</p>
                </td>
                <td class="px-4 py-4">
                  <p v-if="row.snapshotStandardRate !== null" class="font-mono font-black text-foreground">{{ formatTaxRatePercentage(row.snapshotStandardRate) }}</p>
                  <p v-else class="text-muted-foreground">{{ row.hasSnapshotData ? '有快照数据，但无标准税率' : '快照未覆盖' }}</p>
                  <p v-if="row.hasSnapshotData" class="mt-1 text-[10px] text-muted-foreground">{{ currentSnapshot.snapshot?.version || '当前快照' }}</p>
                </td>
                <td class="px-4 py-4">
                  <span class="rounded-full border px-2.5 py-1 text-[10px] font-black" :class="taxRateCoverageStatusClass(row)">{{ taxRateCoverageStatusLabel(row) }}</span>
                  <p v-if="row.enabledRegionalRuleCount > 0" class="mt-2 text-[10px] text-muted-foreground">{{ row.enabledRegionalRuleCount }} 条州省 / 邮编规则</p>
                  <p v-if="row.disabledRuleCount > 0" class="mt-1 text-[10px] text-muted-foreground">另有 {{ row.disabledRuleCount }} 条停用规则</p>
                </td>
                <td class="px-4 py-4 font-mono font-black text-foreground">
                  <template v-if="row.enabledCountryDefaultRule">{{ formatTaxRatePercentage(row.enabledCountryDefaultRule.rate_decimal) }}</template>
                  <span v-else class="font-sans font-normal text-muted-foreground">无</span>
                </td>
                <td class="px-4 py-4">
                  <span class="rounded-full border px-2.5 py-1 text-[10px] font-black" :class="taxRateSnapshotComparisonClass(row)">{{ taxRateSnapshotComparisonLabel(row) }}</span>
                </td>
                <td class="px-4 py-4 text-right">
                  <Button variant="outline" size="sm" @click="activeTab = 'rules'">查看结算规则</Button>
                </td>
              </tr>
              <tr v-if="filteredTaxRateCoverageRows.length === 0">
                <td colspan="6" class="px-4 py-10 text-center text-muted-foreground">没有符合搜索条件的地区。核对范围来自后台市场国家和已有结算规则。</td>
              </tr>
            </tbody>
          </table>
        </div>
      </section>

      <section class="rounded-3xl border bg-card p-5">
        <h2 class="text-sm font-black text-foreground">怎么看这个结果</h2>
        <div class="mt-3 grid gap-3 text-xs leading-relaxed text-muted-foreground lg:grid-cols-3">
          <p class="rounded-2xl border bg-background p-4"><strong class="text-foreground">未配置：</strong>当前没有能匹配该国的启用规则；Checkout 会阻止未匹配地区的报价和下单。</p>
          <p class="rounded-2xl border bg-background p-4"><strong class="text-foreground">部分覆盖：</strong>只有州省或邮编规则；未命中这些细分规则的地址会被阻止。</p>
          <p class="rounded-2xl border bg-background p-4"><strong class="text-foreground">数值一致：</strong>只说明结算国家默认率与快照标准率相同。仍需确认销售方纳税义务、商品税务类别、地区范围和生效日期。</p>
        </div>
      </section>
    </template>

    <template v-else>
      <section class="rounded-3xl border border-amber-500/20 bg-amber-500/5 p-4">
        <div class="flex items-start gap-3">
          <ShieldCheck class="mt-0.5 size-4 shrink-0 text-amber-700 dark:text-amber-300" />
          <div class="space-y-1 text-xs leading-relaxed text-muted-foreground">
            <p>这里维护 Checkout 实际读取的税率。匹配顺序为精确邮编、州默认、国家默认；外部快照不会自动写入这些规则。</p>
            <p>显式配置 0% 是有效规则；未匹配到规则时 Checkout 报价和下单会被阻止，并返回“税率不可用”。</p>
          </div>
        </div>
      </section>

      <section v-if="editingTaxRateRule" class="rounded-3xl border bg-card p-5">
        <div class="flex flex-col gap-1 sm:flex-row sm:items-center sm:justify-between">
          <div>
            <h2 class="text-base font-black text-foreground">{{ editingTaxRateRule.id ? '编辑结算规则' : '新建结算规则' }}</h2>
            <p class="mt-1 text-xs text-muted-foreground">税率为百分比数值，例如 7.25 表示 7.25%，0 表示明确免税。</p>
          </div>
          <button type="button" class="text-xs font-bold text-muted-foreground hover:text-foreground" @click="cancelTaxRateRuleEdit">取消</button>
        </div>

        <form class="mt-5 grid gap-4 sm:grid-cols-2 xl:grid-cols-4" @submit.prevent="saveTaxRateRule">
          <label class="sm:col-span-2">
            <span class="block text-xs font-bold text-foreground">规则名称</span>
            <input v-model.trim="editingTaxRateRule.name" required maxlength="255" type="text" class="mt-1.5 h-10 w-full rounded-xl border border-input bg-background px-3 text-sm outline-none focus-visible:ring-2 focus-visible:ring-ring/50">
          </label>
          <label>
            <span class="block text-xs font-bold text-foreground">国家</span>
            <select v-model="editingTaxRateRule.country" required class="mt-1.5 h-10 w-full rounded-xl border border-input bg-background px-3 text-sm outline-none focus-visible:ring-2 focus-visible:ring-ring/50" @change="clearLocationFieldsAfterTaxRateRuleCountryChange">
              <option value="" disabled>选择国家</option>
              <option v-for="country in taxRateRuleCountryOptions" :key="country.code" :value="country.code">{{ country.name }} ({{ country.code }})</option>
            </select>
          </label>
          <label>
            <span class="block text-xs font-bold text-foreground">州 / 省代码（可空）</span>
            <select v-model="editingTaxRateRule.state" :disabled="!editingTaxRateRule.country" class="mt-1.5 h-10 w-full rounded-xl border border-input bg-background px-3 font-mono text-sm outline-none focus-visible:ring-2 focus-visible:ring-ring/50" @change="clearPostalCodeAfterTaxRateRuleStateChange">
              <option value="">不限定州 / 省（国家默认）</option>
              <option v-for="region in selectedTaxRateRuleRegionOptions" :key="region.code" :value="region.code">{{ region.name }} ({{ region.code }})</option>
            </select>
            <span v-if="editingTaxRateRule.country && selectedTaxRateRuleRegionOptions.length === 0" class="mt-1.5 block text-[10px] leading-relaxed text-muted-foreground">当前目录没有该国家的州 / 省代码；可配置国家默认税率。</span>
            <span v-else-if="editingTaxRateRule.country" class="mt-1.5 block text-[10px] leading-relaxed text-muted-foreground">州 / 省保存为标准代码（如美国 CA）；税率仍按结账地址提交的值精确匹配。</span>
          </label>
          <label>
            <span class="block text-xs font-bold text-foreground">邮编（可空）</span>
            <input v-model.trim="editingTaxRateRule.postal_code" list="tax-rate-rule-postal-code-options" maxlength="100" placeholder="精确邮编，例如 90210" type="text" class="mt-1.5 h-10 w-full rounded-xl border border-input bg-background px-3 font-mono text-sm uppercase outline-none focus-visible:ring-2 focus-visible:ring-ring/50">
            <datalist id="tax-rate-rule-postal-code-options">
              <option v-for="postalCode in selectedTaxRateRulePostalCodeSuggestions" :key="postalCode" :value="postalCode" />
            </datalist>
            <span class="mt-1.5 block text-[10px] leading-relaxed text-muted-foreground">邮编没有全球完整的可选清单；这里提供当前国家 / 州省已配置的邮编建议。新增邮编请按税务资料填写，系统按完整邮编精确匹配。</span>
          </label>
          <label>
            <span class="block text-xs font-bold text-foreground">税率（%）</span>
            <input v-model.trim="editingTaxRateRule.rate_decimal" required inputmode="decimal" placeholder="7.25" type="text" class="mt-1.5 h-10 w-full rounded-xl border border-input bg-background px-3 font-mono text-sm outline-none focus-visible:ring-2 focus-visible:ring-ring/50">
          </label>
          <label>
            <span class="block text-xs font-bold text-foreground">优先级</span>
            <input v-model.number="editingTaxRateRule.priority" type="number" step="1" class="mt-1.5 h-10 w-full rounded-xl border border-input bg-background px-3 font-mono text-sm outline-none focus-visible:ring-2 focus-visible:ring-ring/50">
          </label>
          <label class="flex items-center gap-2 self-end rounded-xl border bg-background px-3 py-2.5 text-xs font-bold text-foreground">
            <input v-model="editingTaxRateRule.enabled" type="checkbox" class="size-4 accent-primary">
            启用此规则
          </label>
          <div class="flex items-end gap-2 sm:col-span-2 xl:col-span-4">
            <Button type="submit" :disabled="savingTaxRateRule || !canEdit">
              <LoaderCircle v-if="savingTaxRateRule" class="size-4 animate-spin" />
              <Save v-else class="size-4" />
              {{ savingTaxRateRule ? '保存中' : '保存结算规则' }}
            </Button>
            <Button type="button" variant="outline" :disabled="savingTaxRateRule" @click="cancelTaxRateRuleEdit">取消</Button>
          </div>
        </form>
      </section>

      <section class="overflow-hidden rounded-3xl border bg-card">
        <div class="flex flex-col gap-3 border-b p-4 sm:flex-row sm:items-center sm:justify-between">
          <div>
            <h2 class="text-sm font-black text-foreground">结算税率规则</h2>
            <p class="mt-1 text-xs text-muted-foreground">规则变更会立即影响新的结算报价；已创建订单保留其订单金额快照。</p>
          </div>
          <Button variant="outline" :disabled="loadingTaxRateRules" @click="loadTaxRateRules">
            <LoaderCircle v-if="loadingTaxRateRules" class="size-4 animate-spin" />
            <RefreshCw v-else class="size-4" />
            刷新规则
          </Button>
        </div>

        <div v-if="loadingTaxRateRules" class="p-10 text-center text-sm text-muted-foreground">正在读取结算税率规则…</div>
        <div v-else class="overflow-x-auto">
          <table class="w-full min-w-[900px] text-left text-xs">
            <thead class="bg-muted/50 text-[10px] font-black uppercase tracking-wider text-muted-foreground">
              <tr>
                <th class="px-4 py-3">规则</th>
                <th class="px-4 py-3">地区</th>
                <th class="px-4 py-3">税率</th>
                <th class="px-4 py-3">优先级</th>
                <th class="px-4 py-3">状态</th>
                <th class="px-4 py-3 text-right">操作</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-border">
              <tr v-for="rule in taxRateRules" :key="rule.id" class="align-top">
                <td class="px-4 py-4">
                  <p class="font-black text-foreground">{{ rule.name }}</p>
                  <p class="mt-1 font-mono text-[10px] text-muted-foreground">#{{ rule.id }}</p>
                </td>
                <td class="px-4 py-4 font-mono font-bold text-foreground">
                  {{ rule.country }}<span v-if="rule.state"> · {{ rule.state }}</span><span v-if="rule.postal_code"> · {{ rule.postal_code }}</span>
                  <p class="mt-1 font-sans text-[10px] font-normal text-muted-foreground">{{ rule.postal_code ? '精确邮编' : rule.state ? '州 / 省默认' : '国家默认' }}</p>
                </td>
                <td class="px-4 py-4 font-mono font-black text-foreground">{{ formatTaxRatePercentage(rule.rate_decimal) }}</td>
                <td class="px-4 py-4 font-mono text-foreground">{{ rule.priority }}</td>
                <td class="px-4 py-4">
                  <span class="rounded-full border px-2.5 py-1 text-[10px] font-black" :class="rule.enabled ? 'border-emerald-500/20 bg-emerald-500/10 text-emerald-700 dark:text-emerald-300' : 'border-border bg-muted text-muted-foreground'">
                    {{ rule.enabled ? '启用' : '停用' }}
                  </span>
                </td>
                <td class="px-4 py-4 text-right">
                  <div v-if="canEdit" class="flex justify-end gap-1.5">
                    <button type="button" class="inline-flex size-8 items-center justify-center rounded-lg border text-foreground hover:bg-muted" :aria-label="`编辑${rule.name}`" @click="startEditingTaxRateRule(rule)">
                      <Pencil class="size-3.5" />
                    </button>
                    <button type="button" class="inline-flex size-8 items-center justify-center rounded-lg border border-destructive/20 text-destructive hover:bg-destructive/5" :aria-label="`删除${rule.name}`" @click="deleteTaxRateRule(rule)">
                      <Trash2 class="size-3.5" />
                    </button>
                  </div>
                  <span v-else class="text-muted-foreground">只读</span>
                </td>
              </tr>
              <tr v-if="taxRateRules.length === 0">
                <td colspan="6" class="px-4 py-10 text-center text-muted-foreground">当前没有结算税率规则。外部快照不会自动创建规则。</td>
              </tr>
            </tbody>
          </table>
        </div>
      </section>
    </template>
  </div>
</template>

<script setup lang="ts">
import { computed, reactive, ref, watch } from 'vue'
import { toast } from 'vue-sonner'
import { Database, ExternalLink, LoaderCircle, Pencil, Plus, RefreshCw, Save, ShieldCheck, Trash2 } from '@lucide/vue'
import { allCountries } from 'country-region-data'
import AdminPageHeader from '@/components/admin/AdminPageHeader.vue'
import { Button } from '@/components/ui/button'
import { useRouteTab } from '@/composables/useRouteTab'
import { useAuthStore } from '@/stores/auth'
import axios from '@/utils/axios'

interface TaxRateSourceConfigurationView {
  provider_code: string
  provider_name: string
  endpoint: string
  documentation_url: string
  enabled: boolean
  refresh_interval_hours: number
  requires_api_key: boolean
  last_checked_at?: string | null
  last_successful_sync_at?: string | null
  last_sync_error?: string
  coverage_description: string
}

interface TaxRateSourceSnapshotView {
  id: number
  provider_code: string
  source_endpoint: string
  version: string
  content_sha256: string
  captured_at: string
  country_count: number
  rate_count: number
}

interface TaxRateSourceSnapshotEntryView {
  snapshot_id: number
  country_code: string
  source_country_code: string
  country_name: string
  currency: string
  rate_type: string
  rate_category: string
  rate_decimal: string
}

interface TaxRateSnapshotCountryRow {
  countryCode: string
  sourceCountryCode: string
  countryName: string
  currency: string
  standardRate: string | null
  otherRates: Array<{ rateType: string; rateDecimal: string }>
  categoryRates: Array<{ name: string; rates: string[] }>
}

interface TaxRateRuleView {
  id: number
  name: string
  country: string
  state: string
  postal_code: string
  rate_decimal: string
  priority: number
  enabled: boolean
}

interface TaxRateCoverageMarketView {
  code: string
  name: string
  countries: string[]
  enabled: boolean
}

interface TaxRateCoverageCountryRow {
  countryCode: string
  countryName: string
  enabledMarketNames: string[]
  disabledMarketNames: string[]
  hasSnapshotData: boolean
  snapshotStandardRate: string | null
  enabledCountryDefaultRule: TaxRateRuleView | null
  enabledCountryDefaultRuleCount: number
  enabledRegionalRuleCount: number
  disabledRuleCount: number
}

interface TaxRateRuleDraft {
  id: number | null
  name: string
  country: string
  state: string
  postal_code: string
  rate_decimal: string
  priority: number
  enabled: boolean
}

interface TaxRateRuleCountryOption {
  code: string
  name: string
}

interface TaxRateRuleRegionOption {
  code: string
  name: string
}

const authStore = useAuthStore()
const refreshIntervalOptions = [
  { hours: 720, label: '每 30 天' },
  { hours: 1440, label: '每 60 天' },
  { hours: 2160, label: '每 90 天' },
  { hours: 4320, label: '每 180 天' },
  { hours: 8760, label: '每 365 天' },
]
const taxRateManagementTabs = [
  { value: 'config' as const, label: '数据源配置' },
  { value: 'snapshot' as const, label: '当前快照' },
  { value: 'coverage' as const, label: '地区覆盖核对' },
  { value: 'rules' as const, label: '结算规则' },
]

const activeTab = useRouteTab({
  defaultValue: 'config',
  values: ['config', 'snapshot', 'coverage', 'rules'],
  routes: {
    config: 'TaxRateManagementConfig',
    snapshot: 'TaxRateManagementSnapshot',
    coverage: 'TaxRateManagementCoverage',
    rules: 'TaxRateManagementRules',
  },
})

const canEdit = computed(() => authStore.hasPermission('settings:edit'))
const loadingConfiguration = ref(false)
const savingConfiguration = ref(false)
const syncingSnapshot = ref(false)
const loadingSnapshot = ref(false)
const loadingTaxRateRules = ref(false)
const loadingTaxRateCoverage = ref(false)
const savingTaxRateRule = ref(false)
const savedEnabled = ref(false)
const savedRefreshIntervalHours = ref(720)
const countrySearchText = ref('')
const coverageSearchText = ref('')
const taxRateRules = ref<TaxRateRuleView[]>([])
const coverageMarkets = ref<TaxRateCoverageMarketView[]>([])
const taxRateCoverageError = ref('')
const editingTaxRateRule = ref<TaxRateRuleDraft | null>(null)
const taxRateRuleCountryOptions: TaxRateRuleCountryOption[] = allCountries
  .map(([name, code]) => ({ code, name }))
  .sort((left, right) => left.name.localeCompare(right.name))

const selectedTaxRateRuleRegionOptions = computed<TaxRateRuleRegionOption[]>(() => {
  const selectedCountryCode = editingTaxRateRule.value?.country
  if (!selectedCountryCode) return []

  const selectedCountry = allCountries.find(([, countryCode]) => countryCode === selectedCountryCode)
  return (selectedCountry?.[2] || [])
    .filter(([, regionCode]) => String(regionCode || '').trim() !== '')
    .map(([name, code]) => ({ code: String(code || '').trim().toUpperCase(), name }))
})

const selectedTaxRateRulePostalCodeSuggestions = computed(() => {
  const draft = editingTaxRateRule.value
  if (!draft) return []
  const countryCode = draft.country.trim().toUpperCase()
  const stateCode = draft.state.trim().toUpperCase()

  return Array.from(new Set(taxRateRules.value
    .filter(rule => (
      rule.country.trim().toUpperCase() === countryCode &&
      rule.state.trim().toUpperCase() === stateCode &&
      rule.postal_code.trim() !== ''
    ))
    .map(rule => rule.postal_code.trim().toUpperCase())))
    .sort((left, right) => left.localeCompare(right))
})

const sourceConfiguration = reactive<TaxRateSourceConfigurationView>({
  provider_code: 'vatcomply',
  provider_name: 'VATcomply',
  endpoint: 'https://api.vatcomply.com/vat_rates',
  documentation_url: 'https://www.vatcomply.com/api/vat-rates/',
  enabled: false,
  refresh_interval_hours: 720,
  requires_api_key: false,
  coverage_description: 'VATcomply currently provides VAT rates for EU member states. It does not cover every storefront market or determine which product-specific rate applies.',
})

const currentSnapshot = reactive<{
  snapshot: TaxRateSourceSnapshotView | null
  entries: TaxRateSourceSnapshotEntryView[]
}>({ snapshot: null, entries: [] })

const configurationDirty = computed(() => (
  sourceConfiguration.enabled !== savedEnabled.value ||
  sourceConfiguration.refresh_interval_hours !== savedRefreshIntervalHours.value
))

const taxRateCountryRows = computed(() => buildTaxRateSnapshotCountryRows(currentSnapshot.entries))
const filteredTaxRateCountryRows = computed(() => {
  const query = countrySearchText.value.trim().toLowerCase()
  if (!query) return taxRateCountryRows.value
  return taxRateCountryRows.value.filter((country) => (
    country.countryName.toLowerCase().includes(query) ||
    country.countryCode.toLowerCase().includes(query) ||
    country.sourceCountryCode.toLowerCase().includes(query)
  ))
})

const taxRateCoverageRows = computed<TaxRateCoverageCountryRow[]>(() => {
  const countries = new Map<string, TaxRateCoverageCountryRow>()
  const snapshotByCountry = new Map<string, { hasData: boolean; standardRate: string | null }>()
  for (const entry of currentSnapshot.entries) {
    const countryCode = entry.country_code.trim().toUpperCase()
    if (!countryCode) continue
    const snapshotCountry = snapshotByCountry.get(countryCode) || { hasData: false, standardRate: null }
    snapshotCountry.hasData = true
    if (entry.rate_type === 'standard') snapshotCountry.standardRate = entry.rate_decimal
    snapshotByCountry.set(countryCode, snapshotCountry)
  }

  for (const market of coverageMarkets.value) {
    const marketName = market.name?.trim() || market.code.trim().toUpperCase()
    for (const rawCountryCode of market.countries || []) {
      const countryCode = rawCountryCode.trim().toUpperCase()
      if (!countryCode) continue
      let row = countries.get(countryCode)
      if (!row) {
        row = createTaxRateCoverageCountryRow(countryCode, snapshotByCountry)
        countries.set(countryCode, row)
      }
      const marketNames = market.enabled ? row.enabledMarketNames : row.disabledMarketNames
      if (!marketNames.includes(marketName)) marketNames.push(marketName)
    }
  }

  for (const rule of taxRateRules.value) {
    const countryCode = rule.country.trim().toUpperCase()
    if (!countryCode) continue
    let row = countries.get(countryCode)
    if (!row) {
      row = createTaxRateCoverageCountryRow(countryCode, snapshotByCountry)
      countries.set(countryCode, row)
    }
    if (!rule.enabled) {
      row.disabledRuleCount++
      continue
    }
    if (!rule.state.trim() && !rule.postal_code.trim()) {
      row.enabledCountryDefaultRuleCount++
      if (!row.enabledCountryDefaultRule || compareTaxRateRulePriority(rule, row.enabledCountryDefaultRule) < 0) {
        row.enabledCountryDefaultRule = rule
      }
      continue
    }
    row.enabledRegionalRuleCount++
  }

  return Array.from(countries.values()).sort((left, right) => left.countryName.localeCompare(right.countryName))
})

const filteredTaxRateCoverageRows = computed(() => {
  const query = coverageSearchText.value.trim().toLowerCase()
  if (!query) return taxRateCoverageRows.value
  return taxRateCoverageRows.value.filter((row) => (
    row.countryName.toLowerCase().includes(query) ||
    row.countryCode.toLowerCase().includes(query) ||
    row.enabledMarketNames.some(name => name.toLowerCase().includes(query)) ||
    row.disabledMarketNames.some(name => name.toLowerCase().includes(query))
  ))
})

const enabledMarketCountryCount = computed(() => taxRateCoverageRows.value.filter(row => row.enabledMarketNames.length > 0).length)
const enabledMarketCountriesWithDefaultRuleCount = computed(() => taxRateCoverageRows.value.filter(row => row.enabledMarketNames.length > 0 && row.enabledCountryDefaultRule).length)
const enabledMarketCountriesWithRegionalRulesOnlyCount = computed(() => taxRateCoverageRows.value.filter(row => row.enabledMarketNames.length > 0 && !row.enabledCountryDefaultRule && row.enabledRegionalRuleCount > 0).length)
const enabledMarketCountriesWithSnapshotStandardRateCount = computed(() => taxRateCoverageRows.value.filter(row => row.enabledMarketNames.length > 0 && row.snapshotStandardRate !== null).length)

const createTaxRateCoverageCountryRow = (
  countryCode: string,
  snapshotByCountry: Map<string, { hasData: boolean; standardRate: string | null }>,
): TaxRateCoverageCountryRow => {
  const snapshotCountry = snapshotByCountry.get(countryCode)
  const countryName = taxRateRuleCountryOptions.find(country => country.code === countryCode)?.name ||
    currentSnapshot.entries.find(entry => entry.country_code.trim().toUpperCase() === countryCode)?.country_name ||
    countryCode
  return {
    countryCode,
    countryName,
    enabledMarketNames: [],
    disabledMarketNames: [],
    hasSnapshotData: snapshotCountry?.hasData || false,
    snapshotStandardRate: snapshotCountry?.standardRate || null,
    enabledCountryDefaultRule: null,
    enabledCountryDefaultRuleCount: 0,
    enabledRegionalRuleCount: 0,
    disabledRuleCount: 0,
  }
}

const compareTaxRateRulePriority = (left: TaxRateRuleView, right: TaxRateRuleView): number => (
  right.priority - left.priority || left.id - right.id
)

const taxRateCoverageStatusLabel = (row: TaxRateCoverageCountryRow): string => {
  if (row.enabledCountryDefaultRuleCount > 1) return `多条国家默认 · ${row.enabledCountryDefaultRuleCount} 条`
  if (row.enabledCountryDefaultRule && row.enabledRegionalRuleCount > 0) return `国家默认 + ${row.enabledRegionalRuleCount} 条分区规则`
  if (row.enabledCountryDefaultRule) return '已配置国家默认'
  if (row.enabledRegionalRuleCount > 0) return '部分地区已配置'
  if (row.disabledRuleCount > 0) return '仅有停用规则'
  return '未配置'
}

const taxRateCoverageStatusClass = (row: TaxRateCoverageCountryRow): string => {
  if (row.enabledCountryDefaultRuleCount > 1) return 'border-amber-500/20 bg-amber-500/10 text-amber-700 dark:text-amber-300'
  if (row.enabledCountryDefaultRule) return 'border-emerald-500/20 bg-emerald-500/10 text-emerald-700 dark:text-emerald-300'
  if (row.enabledRegionalRuleCount > 0) return 'border-amber-500/20 bg-amber-500/10 text-amber-700 dark:text-amber-300'
  return 'border-rose-500/20 bg-rose-500/10 text-rose-700 dark:text-rose-300'
}

const normalizeTaxRateDecimalForComparison = (rate: string): string => {
  const [rawInteger, rawFraction = ''] = rate.trim().split('.')
  const integer = rawInteger.replace(/^0+(?=\d)/, '') || '0'
  const fraction = rawFraction.replace(/0+$/, '')
  return fraction ? `${integer}.${fraction}` : integer
}

const taxRateSnapshotComparisonLabel = (row: TaxRateCoverageCountryRow): string => {
  if (row.snapshotStandardRate === null) return '无法与标准税率对比'
  if (!row.enabledCountryDefaultRule) return '缺少国家默认规则'
  return normalizeTaxRateDecimalForComparison(row.enabledCountryDefaultRule.rate_decimal) === normalizeTaxRateDecimalForComparison(row.snapshotStandardRate)
    ? '数值一致 · 待核验适用性'
    : '数值不同 · 需核对'
}

const taxRateSnapshotComparisonClass = (row: TaxRateCoverageCountryRow): string => {
  if (row.snapshotStandardRate === null || !row.enabledCountryDefaultRule) return 'border-border bg-muted text-muted-foreground'
  return normalizeTaxRateDecimalForComparison(row.enabledCountryDefaultRule.rate_decimal) === normalizeTaxRateDecimalForComparison(row.snapshotStandardRate)
    ? 'border-sky-500/20 bg-sky-500/10 text-sky-700 dark:text-sky-300'
    : 'border-amber-500/20 bg-amber-500/10 text-amber-700 dark:text-amber-300'
}

const extractTaxRateAdminResponsePayload = (response: { data?: any }): Record<string, any> => (
  response.data?.data?.data || response.data?.data || response.data || {}
)

const loadTaxRateSourceConfiguration = async (): Promise<void> => {
  loadingConfiguration.value = true
  try {
    const response = await axios.get('/api/admin/settings/tax-rates/config')
    const configuration = response.data?.data?.config || response.data?.config
    if (!configuration) throw new Error('税率数据源配置响应为空')
    Object.assign(sourceConfiguration, configuration)
    savedEnabled.value = Boolean(configuration.enabled)
    savedRefreshIntervalHours.value = Number(configuration.refresh_interval_hours) || 720
  } catch (error) {
    console.error('Failed to load tax rate source configuration:', error)
    toast.error('税率数据源配置读取失败')
  } finally {
    loadingConfiguration.value = false
  }
}

const saveTaxRateSourceConfiguration = async (): Promise<void> => {
  savingConfiguration.value = true
  try {
    const response = await axios.put('/api/admin/settings/tax-rates/config', {
      enabled: sourceConfiguration.enabled,
      refresh_interval_hours: sourceConfiguration.refresh_interval_hours,
    })
    const configuration = response.data?.data?.config || response.data?.config
    if (configuration) Object.assign(sourceConfiguration, configuration)
    savedEnabled.value = Boolean(sourceConfiguration.enabled)
    savedRefreshIntervalHours.value = Number(sourceConfiguration.refresh_interval_hours)
    toast.success('税率数据源配置已保存')
  } catch (error) {
    console.error('Failed to save tax rate source configuration:', error)
    toast.error('税率数据源配置保存失败')
  } finally {
    savingConfiguration.value = false
  }
}

const loadCurrentTaxRateSourceSnapshot = async (): Promise<void> => {
  loadingSnapshot.value = true
  try {
    const response = await axios.get('/api/admin/settings/tax-rates/snapshot')
    const snapshotData = response.data?.data || response.data || {}
    currentSnapshot.snapshot = snapshotData.snapshot || null
    currentSnapshot.entries = Array.isArray(snapshotData.entries) ? snapshotData.entries : []
  } catch (error) {
    console.error('Failed to load current tax rate source snapshot:', error)
    toast.error('税率快照读取失败')
  } finally {
    loadingSnapshot.value = false
  }
}

const loadTaxRateRules = async (): Promise<void> => {
  loadingTaxRateRules.value = true
  try {
    const response = await axios.get('/api/admin/settings/tax-rates/rules')
    const payload = response.data?.data || response.data || {}
    const rules = payload.rules || payload.data
    taxRateRules.value = Array.isArray(rules) ? rules : []
  } catch (error) {
    console.error('Failed to load checkout tax rate rules:', error)
    toast.error('结算税率规则读取失败')
  } finally {
    loadingTaxRateRules.value = false
  }
}

const loadTaxRateCoverage = async (): Promise<void> => {
  loadingTaxRateCoverage.value = true
  taxRateCoverageError.value = ''
  try {
    const [marketResponse, ruleResponse, snapshotResponse] = await Promise.all([
      axios.get('/api/admin/storefront/markets'),
      axios.get('/api/admin/settings/tax-rates/rules'),
      axios.get('/api/admin/settings/tax-rates/snapshot'),
    ])
    const marketPayload = extractTaxRateAdminResponsePayload(marketResponse)
    const marketRecords = Array.isArray(marketPayload)
      ? marketPayload
      : Array.isArray(marketPayload.markets)
        ? marketPayload.markets
        : []
    const rulePayload = extractTaxRateAdminResponsePayload(ruleResponse)
    const ruleRecords = rulePayload.rules || rulePayload.data
    const snapshotPayload = extractTaxRateAdminResponsePayload(snapshotResponse)
    const snapshotRecord = snapshotPayload.snapshot || null
    const snapshotEntries = Array.isArray(snapshotPayload.entries) ? snapshotPayload.entries : []

    coverageMarkets.value = marketRecords.map((market: any) => ({
      code: String(market.code || '').trim().toUpperCase(),
      name: String(market.name || market.code || '').trim(),
      countries: Array.isArray(market.countries) ? market.countries.map((country: unknown) => String(country).trim().toUpperCase()) : [],
      enabled: market.enabled === true,
    }))
    taxRateRules.value = Array.isArray(ruleRecords) ? ruleRecords : []
    currentSnapshot.snapshot = snapshotRecord
    currentSnapshot.entries = snapshotEntries
  } catch (error) {
    console.error('Failed to load tax rate market, snapshot, and checkout coverage:', error)
    taxRateCoverageError.value = '市场、快照或 Checkout 规则读取失败，因此当前覆盖结果不完整。请刷新重试；读取成功前不会把缺失数据显示为“未配置”。'
  } finally {
    loadingTaxRateCoverage.value = false
  }
}

const startCreatingTaxRateRule = (): void => {
  editingTaxRateRule.value = {
    id: null,
    name: '',
    country: '',
    state: '',
    postal_code: '',
    rate_decimal: '',
    priority: 0,
    enabled: true,
  }
}

const startEditingTaxRateRule = (rule: TaxRateRuleView): void => {
  editingTaxRateRule.value = {
    id: rule.id,
    name: rule.name,
    country: rule.country.trim().toUpperCase(),
    state: rule.state.trim().toUpperCase(),
    postal_code: rule.postal_code,
    rate_decimal: rule.rate_decimal,
    priority: rule.priority,
    enabled: rule.enabled,
  }
}

const clearLocationFieldsAfterTaxRateRuleCountryChange = (): void => {
  if (!editingTaxRateRule.value) return
  editingTaxRateRule.value.state = ''
  editingTaxRateRule.value.postal_code = ''
}

const clearPostalCodeAfterTaxRateRuleStateChange = (): void => {
  if (editingTaxRateRule.value) editingTaxRateRule.value.postal_code = ''
}

const cancelTaxRateRuleEdit = (): void => {
  editingTaxRateRule.value = null
}

const saveTaxRateRule = async (): Promise<void> => {
  const draft = editingTaxRateRule.value
  if (!draft) return

  const countryCode = draft.country.trim().toUpperCase()
  if (!taxRateRuleCountryOptions.some(country => country.code === countryCode)) {
    toast.error('请选择标准国家代码')
    return
  }
  const stateCode = draft.state.trim().toUpperCase()
  if (stateCode && !selectedTaxRateRuleRegionOptions.value.some(region => region.code.toUpperCase() === stateCode)) {
    toast.error('州 / 省必须从标准代码列表中选择')
    return
  }

  savingTaxRateRule.value = true
  const requestBody = {
    name: draft.name,
    country: countryCode,
    state: stateCode,
    postal_code: draft.postal_code,
    rate_decimal: draft.rate_decimal,
    priority: draft.priority,
    enabled: draft.enabled,
  }
  try {
    if (draft.id === null) {
      await axios.post('/api/admin/settings/tax-rates/rules', requestBody)
      toast.success('结算税率规则已创建')
    } else {
      await axios.put(`/api/admin/settings/tax-rates/rules/${draft.id}`, requestBody)
      toast.success('结算税率规则已更新')
    }
    editingTaxRateRule.value = null
    await loadTaxRateRules()
  } catch (error) {
    console.error('Failed to save checkout tax rate rule:', error)
    toast.error('结算税率规则保存失败，请检查国家代码和税率格式')
  } finally {
    savingTaxRateRule.value = false
  }
}

const deleteTaxRateRule = async (rule: TaxRateRuleView): Promise<void> => {
  if (!window.confirm(`确定删除结算规则“${rule.name}”吗？删除后新订单将不再匹配此规则。`)) return
  try {
    await axios.delete(`/api/admin/settings/tax-rates/rules/${rule.id}`)
    toast.success('结算税率规则已删除')
    if (editingTaxRateRule.value?.id === rule.id) editingTaxRateRule.value = null
    await loadTaxRateRules()
  } catch (error) {
    console.error('Failed to delete checkout tax rate rule:', error)
    toast.error('结算税率规则删除失败')
  }
}

const syncTaxRateSourceSnapshot = async (): Promise<void> => {
  if (!sourceConfiguration.enabled) {
    toast.error('请先启用税率数据源')
    return
  }
  if (configurationDirty.value) {
    toast.error('请先保存数据源配置')
    return
  }
  syncingSnapshot.value = true
  try {
    const response = await axios.post('/api/admin/settings/tax-rates/sync')
    const result = response.data?.data || response.data || {}
    toast.success(result.changed ? '新税率快照已创建' : '数据未变化，当前快照仍然有效')
    await Promise.all([loadTaxRateSourceConfiguration(), loadCurrentTaxRateSourceSnapshot()])
  } catch (error) {
    console.error('Failed to synchronize tax rate source snapshot:', error)
    toast.error('税率数据同步失败，请检查数据源状态')
    await loadTaxRateSourceConfiguration()
  } finally {
    syncingSnapshot.value = false
  }
}

const buildTaxRateSnapshotCountryRows = (entries: TaxRateSourceSnapshotEntryView[]): TaxRateSnapshotCountryRow[] => {
  const countryRowsByCode = new Map<string, TaxRateSnapshotCountryRow>()
  for (const entry of entries) {
    let country = countryRowsByCode.get(entry.country_code)
    if (!country) {
      country = {
        countryCode: entry.country_code,
        sourceCountryCode: entry.source_country_code,
        countryName: entry.country_name,
        currency: entry.currency,
        standardRate: null,
        otherRates: [],
        categoryRates: [],
      }
      countryRowsByCode.set(entry.country_code, country)
    }
    if (entry.rate_type === 'standard') {
      country.standardRate = entry.rate_decimal
    } else if (entry.rate_type === 'product_category') {
      let category = country.categoryRates.find((item) => item.name === entry.rate_category)
      if (!category) {
        category = { name: entry.rate_category, rates: [] }
        country.categoryRates.push(category)
      }
      category.rates.push(entry.rate_decimal)
    } else {
      country.otherRates.push({ rateType: entry.rate_type, rateDecimal: entry.rate_decimal })
    }
  }
  return Array.from(countryRowsByCode.values())
    .map((country) => ({
      ...country,
      otherRates: country.otherRates.sort((left, right) => left.rateType.localeCompare(right.rateType) || Number(left.rateDecimal) - Number(right.rateDecimal)),
      categoryRates: country.categoryRates
        .map((category) => ({ ...category, rates: category.rates.sort((left, right) => Number(left) - Number(right)) }))
        .sort((left, right) => left.name.localeCompare(right.name)),
    }))
    .sort((left, right) => left.countryName.localeCompare(right.countryName))
}

const formatTaxRatePercentage = (rate: string | null): string => rate === null ? '无标准税率' : `${Number(rate)}%`

const formatTaxRateType = (rateType: string): string => ({
  reduced: '优惠',
  super_reduced: '超优惠',
  parking: '停车',
}[rateType] || rateType)

const humanizeTaxRateCategory = (category: string): string => category.replace(/_/g, ' ')

const formatTaxRateSnapshotTimestamp = (value?: string | null): string => {
  if (!value) return ''
  const date = new Date(value)
  return Number.isNaN(date.getTime()) ? '' : new Intl.DateTimeFormat('zh-CN', {
    dateStyle: 'medium',
    timeStyle: 'short',
  }).format(date)
}

const formatRefreshInterval = (hours: number): string => {
  const matchingOption = refreshIntervalOptions.find((option) => option.hours === hours)
  if (matchingOption) return matchingOption.label
  return `每 ${Math.max(1, Math.round(hours / 24))} 天`
}

watch(activeTab, async (tab) => {
  if (tab === 'config') {
    await loadTaxRateSourceConfiguration()
    return
  }
  if (tab === 'snapshot') {
    await loadCurrentTaxRateSourceSnapshot()
    return
  }
  if (tab === 'coverage') {
    await loadTaxRateCoverage()
    return
  }
  await loadTaxRateRules()
}, { immediate: true })
</script>
