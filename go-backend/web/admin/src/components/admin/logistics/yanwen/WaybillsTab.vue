<script setup lang="ts">
import { useYanwenWaybillOperations } from '@/composables/shipping/useYanwenWaybillOperations'

const {
  waybillSearch,
  waybillEnvironment,
  waybillChannel,
  waybillWarehouse,
  waybillStatus,
  waybillRefreshing,
  waybillSyncingId,
  waybillBatchSyncing,
  waybillBatchCancelling,
  waybillBatchLabelDownloading,
  waybillLabelDownloadingId,
  waybillCancellingId,
  waybillSelectedIds,
  waybillActionMessage,
  createWaybillDialogOpen,
  createWaybillSaving,
  createWaybillForm,
  batchCreateWaybillDialogOpen,
  batchCreateWaybillSaving,
  batchCreateWaybillForm,
  batchCreateWaybillRows,
  curatedChannels,
  warehouses,
  selectedYanwenChannel,
  selectedBatchYanwenChannel,
  selectedYanwenChannelRequiredFields,
  selectedBatchYanwenChannelRequiredFields,
  canShip,
  canCancel,
  waybillRecords,
  filteredWaybills,
  waybillChannelOptions,
  hasSelectedWaybills,
  refreshWaybills,
  resetWaybillFilters,
  toggleWaybill,
  downloadSelectedWaybillLabels,
  cancelWaybill,
  cancelSelectedWaybills,
  syncWaybillOfficialDetails,
  syncSelectedWaybillsOfficialDetails,
  downloadWaybillLabel,
  openCreateWaybillDialog,
  openBatchCreateWaybillDialog,
  addYanwenBatchWaybillCreationFormRow,
  removeYanwenBatchWaybillCreationFormRow,
  createYanwenBatchWaybills,
  createWaybill,
} = useYanwenWaybillOperations()
</script>

<template>
      <section class="space-y-4 rounded-[28px] border border-dashed border-border/80 bg-card p-5 shadow-sm sm:p-6">
        <div class="flex flex-col gap-3 lg:flex-row lg:items-start lg:justify-between">
          <div>
            <p class="text-[10px] font-black uppercase tracking-[0.18em] text-emerald-600">Waybill Fulfillment &amp; Labels</p>
            <h2 class="mt-1 text-lg font-black tracking-tight">专线运单与面单中心</h2>
            <p class="mt-1 max-w-3xl text-xs leading-5 text-muted-foreground">台账只读取已成功创建的真实燕文运单；可逐票或批量推单、同步官方详情、下载真实 PDF 面单，并逐票或批量提交取消。批量面单会下载包含官方 PDF 和逐票 manifest 的压缩包。</p>
          </div>
          <div class="flex flex-wrap gap-2">
            <Button variant="outline" size="sm" :disabled="waybillRefreshing" @click="refreshWaybills"><RefreshCw :class="['size-3.5', { 'animate-spin': waybillRefreshing }]" />刷新</Button>
            <Button variant="outline" size="sm" :disabled="!canShip || !hasSelectedWaybills || waybillBatchSyncing" :title="canShip ? undefined : '需要 logistics:yanwen:ship 权限'" @click="syncSelectedWaybillsOfficialDetails"><RefreshCw :class="['size-3.5', { 'animate-spin': waybillBatchSyncing }]" />{{ waybillBatchSyncing ? '同步中…' : '批量同步官方状态' }}</Button>
            <Button variant="outline" size="sm" :disabled="!canCancel || !hasSelectedWaybills || waybillBatchCancelling" :title="canCancel ? undefined : '需要 logistics:yanwen:cancel 权限'" @click="cancelSelectedWaybills"><Ban :class="['size-3.5', { 'animate-pulse': waybillBatchCancelling }]" />{{ waybillBatchCancelling ? '取消中…' : '批量取消' }}</Button>
            <Button size="sm" :disabled="!canShip" :title="canShip ? undefined : '需要 logistics:yanwen:ship 权限'" @click="openCreateWaybillDialog"><Plus class="size-3.5" />创建真实运单</Button>
            <Button variant="outline" size="sm" :disabled="!canShip" :title="canShip ? undefined : '需要 logistics:yanwen:ship 权限'" @click="openBatchCreateWaybillDialog"><Plus class="size-3.5" />批量推单</Button>
            <Button variant="outline" size="sm" :disabled="!canShip || !hasSelectedWaybills || waybillBatchLabelDownloading" :title="canShip ? undefined : '需要 logistics:yanwen:ship 权限'" @click="downloadSelectedWaybillLabels"><Printer :class="['size-3.5', { 'animate-pulse': waybillBatchLabelDownloading }]" />{{ waybillBatchLabelDownloading ? '下载中…' : '批量打面单' }}</Button>
          </div>
        </div>

        <div class="grid gap-3 rounded-[24px] border border-dashed border-border/80 bg-muted/20 p-4 md:grid-cols-2 xl:grid-cols-[minmax(0,1.4fr)_repeat(3,minmax(0,1fr))_auto]">
          <label class="space-y-1.5"><span class="flex items-center gap-1.5 text-[10px] font-black uppercase tracking-wider text-muted-foreground/80"><Search class="size-3" />搜索运单台账</span><Input v-model="waybillSearch" placeholder="商城单号 / 燕文单号 / 买家" /></label>
          <label class="space-y-1.5"><span class="text-[10px] font-black uppercase tracking-wider text-muted-foreground/80">环境</span><select v-model="waybillEnvironment" class="h-9 w-full rounded-md border border-dashed border-border bg-background px-3 text-sm"><option value="fat">FAT</option><option value="production">PRD</option></select></label>
          <label class="space-y-1.5"><span class="text-[10px] font-black uppercase tracking-wider text-muted-foreground/80">渠道</span><select v-model="waybillChannel" class="h-9 w-full rounded-md border border-dashed border-border bg-background px-3 text-sm"><option value="">全部渠道</option><option v-for="channel in waybillChannelOptions" :key="channel" :value="channel">{{ channel }}</option></select></label>
          <label class="space-y-1.5"><span class="text-[10px] font-black uppercase tracking-wider text-muted-foreground/80">交货仓</span><Input v-model="waybillWarehouse" placeholder="官方仓库代码" /></label>
          <label class="space-y-1.5"><span class="text-[10px] font-black uppercase tracking-wider text-muted-foreground/80">状态</span><select v-model="waybillStatus" class="h-9 w-full rounded-md border border-dashed border-border bg-background px-3 text-sm"><option value="">全部状态</option><option value="已建单">已建单</option></select></label>
          <div class="flex items-end"><Button variant="ghost" size="sm" class="w-full" @click="resetWaybillFilters">重置筛选</Button></div>
        </div>

        <p v-if="waybillActionMessage" class="rounded-2xl border border-amber-500/30 bg-amber-500/10 px-4 py-3 text-xs leading-5 text-amber-700 dark:text-amber-300">{{ waybillActionMessage }}</p>

        <section class="overflow-hidden rounded-[24px] border border-dashed border-border/80">
          <div class="flex items-center justify-between gap-3 border-b border-dashed border-border/80 px-5 py-4">
            <div><h3 class="text-sm font-black">专线运单全透明台账</h3><p class="mt-1 text-[11px] text-muted-foreground">{{ filteredWaybills.length }} 条记录 · 面单和轨迹只展示真实接口返回</p></div>
            <span class="rounded-full bg-muted px-2.5 py-1 text-[10px] font-bold text-muted-foreground">{{ waybillSelectedIds.length }} 已选择</span>
          </div>
          <div v-if="filteredWaybills.length === 0" class="flex min-h-64 flex-col items-center justify-center gap-3 px-6 py-12 text-center">
            <span class="flex size-12 items-center justify-center rounded-2xl bg-muted text-muted-foreground"><PackageCheck class="size-5" /></span>
            <div><p class="text-sm font-black">暂无燕文真实运单记录</p><p class="mt-1 max-w-lg text-xs leading-5 text-muted-foreground">创建成功后，订单、燕文单号、官方渠道、交货仓和申报信息会从本地持久化台账读取。</p></div>
          </div>
          <div v-else class="overflow-x-auto">
            <table class="w-full min-w-[1380px] text-left text-sm">
              <thead class="bg-muted/30 text-[10px] font-black uppercase tracking-wider text-muted-foreground"><tr><th class="w-12 px-5 py-3"><span class="sr-only">选择</span></th><th class="px-4 py-3">订单编号 / 时间</th><th class="px-4 py-3">燕文单号 / 转单号</th><th class="px-4 py-3">渠道 / 交货仓</th><th class="px-4 py-3">目的国 / 买家</th><th class="px-4 py-3">申报品名 / 重量</th><th class="px-4 py-3">官方状态 / 打印</th><th class="px-5 py-3 text-right">操作</th></tr></thead>
              <tbody class="divide-y divide-dashed divide-border/70"><tr v-for="record in filteredWaybills" :key="record.id"><td class="px-5 py-4"><input type="checkbox" :checked="waybillSelectedIds.includes(record.id)" class="size-4 rounded border-border" @change="toggleWaybill(record.id)" /></td><td class="px-4 py-4"><p class="font-mono text-xs font-bold">{{ record.orderNo }}</p><p class="mt-1 text-[11px] text-muted-foreground">{{ record.createdAt }}</p></td><td class="px-4 py-4"><p class="font-mono text-xs font-bold">{{ record.waybill || '尚未推单' }}</p><p class="mt-1 font-mono text-[11px] text-muted-foreground">转单号：{{ record.referenceNumber || '—' }}</p><p class="mt-1 font-mono text-[11px] text-muted-foreground">燕文流水号：{{ record.lastMile || '—' }}</p></td><td class="px-4 py-4"><p class="font-semibold">{{ record.channel }}</p><p class="mt-1 text-[11px] text-muted-foreground">{{ record.warehouse }}</p></td><td class="px-4 py-4"><p class="font-semibold">{{ record.destination }}</p><p class="mt-1 text-[11px] text-muted-foreground">{{ record.consignee }}</p></td><td class="px-4 py-4"><p>{{ record.description }}</p><p class="mt-1 text-[11px] text-muted-foreground">{{ record.weight }}</p></td><td class="px-4 py-4"><span class="rounded-full bg-muted px-2.5 py-1 text-[10px] font-bold text-muted-foreground">{{ record.officialStatusLabel }}</span><p class="mt-2 text-[11px] text-muted-foreground">{{ record.officialSyncedAt ? (record.isPrinted ? '已打印' : '未打印') : '打印状态未同步' }}</p><p v-if="record.officialSyncedAt" class="mt-1 text-[10px] text-muted-foreground">{{ record.officialSyncedAt }}</p></td><td class="px-5 py-4 text-right"><div class="flex justify-end gap-1"><Button variant="outline" size="sm" :disabled="!canShip || waybillSyncingId === record.id" :title="canShip ? undefined : '需要 logistics:yanwen:ship 权限'" @click="syncWaybillOfficialDetails(record.id)"><RefreshCw :class="['size-3.5', { 'animate-spin': waybillSyncingId === record.id }]" />同步官方状态</Button><Button variant="outline" size="sm" :disabled="!canShip || waybillLabelDownloadingId === record.id || waybillBatchLabelDownloading" :title="canShip ? undefined : '需要 logistics:yanwen:ship 权限'" @click="downloadWaybillLabel(record.id)"><Download class="size-3.5" />{{ waybillLabelDownloadingId === record.id ? '下载中…' : '下载面单' }}</Button><Button variant="ghost" size="sm" :disabled="!canCancel || waybillCancellingId === record.id || waybillBatchCancelling" :title="canCancel ? undefined : '需要 logistics:yanwen:cancel 权限'" @click="cancelWaybill(record.id)">{{ waybillCancellingId === record.id || waybillBatchCancelling ? '取消中…' : '取消' }}</Button></div></td></tr></tbody>
            </table>
          </div>
        </section>
      </section>
      <Dialog v-model:open="createWaybillDialogOpen">
        <DialogContent>
          <DialogHeader><DialogTitle>创建真实燕文运单</DialogTitle><DialogDescription>请求会使用订单快照、已启用服务集合和官方交货仓调用 express.order.create。韩国和美国订单会在建单前由服务端重新调用燕文官方关务校验；关务 TAB 的浏览器状态不会直接授权建单。</DialogDescription></DialogHeader>
          <div class="grid gap-4 py-2">
            <label class="space-y-1.5"><span class="text-xs font-semibold">订单 ID</span><Input v-model="createWaybillForm.orderId" inputmode="numeric" placeholder="真实订单 ID" /></label>
            <label class="space-y-1.5"><span class="text-xs font-semibold">已启用燕文渠道</span><select v-model="createWaybillForm.channelId" class="h-9 w-full rounded-md border border-border bg-background px-3 text-sm"><option value="">选择服务集合渠道</option><option v-for="channel in curatedChannels" :key="channel.id" :value="String(channel.id)">{{ channel.display_name }} · {{ channel.product_code }}</option></select></label>
            <label class="space-y-1.5"><span class="text-xs font-semibold">官方交货仓</span><select v-model="createWaybillForm.warehouseCode" class="h-9 w-full rounded-md border border-border bg-background px-3 text-sm"><option value="">选择官方交货仓</option><option v-for="warehouse in warehouses" :key="warehouse.warehouse_code" :value="warehouse.warehouse_code">{{ warehouse.name }} · {{ warehouse.warehouse_code }}</option></select></label>
            <label class="space-y-1.5"><span class="text-xs font-semibold">包裹是否带电</span><select v-model="createWaybillForm.hasBattery" class="h-9 w-full rounded-md border border-border bg-background px-3 text-sm"><option value="">明确选择</option><option value="false">否</option><option value="true">是</option></select></label>
            <div class="space-y-3 rounded-2xl border border-dashed border-emerald-500/30 bg-emerald-500/5 p-3">
              <div><p class="text-xs font-black">燕文关务申报字段</p><p class="mt-1 text-[11px] leading-5 text-muted-foreground">只按燕文 `express.order.create` 官方字段发送：收件人税号→`receiverInfo.taxNumber`、IOSS→`parcelInfo.ioss`、EORI→`importCustomsInfo.eori`。是否必填由当前所选燕文渠道配置；燕文没有独立的 IOSS/EORI 校验接口，必填规则不代表号码有效性校验。</p></div>
              <p v-if="selectedYanwenChannelRequiredFields.length" class="rounded-lg border border-amber-500/30 bg-amber-500/5 px-3 py-2 text-[11px] text-amber-700 dark:text-amber-300">当前渠道要求填写：{{ selectedYanwenChannelRequiredFields.join('、') }}。服务端建单前也会再次校验。</p>
              <p v-else class="text-[11px] text-muted-foreground">{{ selectedYanwenChannel ? '当前渠道未配置额外必填项。' : '请选择渠道后查看该渠道的必填要求。' }}</p>
              <label class="space-y-1.5"><span class="text-[11px] font-semibold">收件人税号（燕文 taxNumber）<span v-if="selectedYanwenChannel?.require_receiver_tax_number" class="text-red-600"> · 必填</span></span><Input v-model="createWaybillForm.receiverTaxNumber" maxlength="50" autocomplete="off" placeholder="按渠道规则及适用情形填写" /></label>
              <div class="grid gap-3 sm:grid-cols-2"><label class="space-y-1.5"><span class="text-[11px] font-semibold">IOSS<span v-if="selectedYanwenChannel?.require_ioss" class="text-red-600"> · 必填</span></span><Input v-model="createWaybillForm.ioss" maxlength="50" autocomplete="off" placeholder="按渠道规则及适用情形填写" /></label><label class="space-y-1.5"><span class="text-[11px] font-semibold">EORI<span v-if="selectedYanwenChannel?.require_eori" class="text-red-600"> · 必填</span></span><Input v-model="createWaybillForm.eori" maxlength="64" autocomplete="off" placeholder="按渠道规则及适用情形填写" /></label></div>
            </div>
          </div>
          <DialogFooter><Button variant="outline" @click="createWaybillDialogOpen = false">取消</Button><Button :disabled="createWaybillSaving" @click="createWaybill">{{ createWaybillSaving ? '提交中…' : '调用官方建单' }}</Button></DialogFooter>
        </DialogContent>
      </Dialog>
      <Dialog v-model:open="batchCreateWaybillDialogOpen">
        <DialogContent class="max-w-5xl">
          <DialogHeader><DialogTitle>批量推送燕文真实运单</DialogTitle><DialogDescription>本批次只选择一个已启用燕文渠道和一个官方交货仓，最多提交 50 个订单。服务端会先校验整个批次，再逐票调用燕文 express.order.create；失败订单会逐票返回原因，成功订单会写入燕文运单台账。</DialogDescription></DialogHeader>
          <div class="grid gap-4 py-2">
            <div class="grid gap-3 sm:grid-cols-2">
              <label class="space-y-1.5"><span class="text-xs font-semibold">本批次燕文渠道</span><select v-model="batchCreateWaybillForm.channelId" class="h-9 w-full rounded-md border border-border bg-background px-3 text-sm"><option value="">选择服务集合渠道</option><option v-for="channel in curatedChannels" :key="channel.id" :value="String(channel.id)">{{ channel.display_name }} · {{ channel.product_code }}</option></select></label>
              <label class="space-y-1.5"><span class="text-xs font-semibold">本批次官方交货仓</span><select v-model="batchCreateWaybillForm.warehouseCode" class="h-9 w-full rounded-md border border-border bg-background px-3 text-sm"><option value="">选择官方交货仓</option><option v-for="warehouse in warehouses" :key="warehouse.warehouse_code" :value="warehouse.warehouse_code">{{ warehouse.name }} · {{ warehouse.warehouse_code }}</option></select></label>
            </div>
            <p v-if="selectedBatchYanwenChannelRequiredFields.length" class="rounded-lg border border-amber-500/30 bg-amber-500/5 px-3 py-2 text-[11px] text-amber-700 dark:text-amber-300">当前渠道要求每票填写：{{ selectedBatchYanwenChannelRequiredFields.join('、') }}。燕文没有独立的 IOSS/EORI 有效性接口，服务端只执行已配置的必填规则。</p>
            <div class="overflow-x-auto rounded-2xl border border-dashed border-border/80">
              <table class="w-full min-w-[960px] text-left text-xs">
                <thead class="bg-muted/30 text-[10px] font-black uppercase tracking-wider text-muted-foreground"><tr><th class="px-3 py-2">订单 ID</th><th class="px-3 py-2">带电</th><th class="px-3 py-2">收件人税号</th><th class="px-3 py-2">IOSS</th><th class="px-3 py-2">EORI</th><th class="w-20 px-3 py-2 text-right">操作</th></tr></thead>
                <tbody class="divide-y divide-dashed divide-border/70">
                  <tr v-for="(row, rowIndex) in batchCreateWaybillRows" :key="rowIndex">
                    <td class="px-3 py-2"><Input v-model="row.orderId" inputmode="numeric" placeholder="订单 ID" /></td>
                    <td class="px-3 py-2"><select v-model="row.hasBattery" class="h-9 w-full rounded-md border border-border bg-background px-2 text-sm"><option value="">选择</option><option value="false">否</option><option value="true">是</option></select></td>
                    <td class="px-3 py-2"><Input v-model="row.receiverTaxNumber" maxlength="50" placeholder="可选/按渠道规则" /></td>
                    <td class="px-3 py-2"><Input v-model="row.ioss" maxlength="50" placeholder="可选/按渠道规则" /></td>
                    <td class="px-3 py-2"><Input v-model="row.eori" maxlength="64" placeholder="可选/按渠道规则" /></td>
                    <td class="px-3 py-2 text-right"><Button variant="ghost" size="sm" :disabled="batchCreateWaybillRows.length <= 1" @click="removeYanwenBatchWaybillCreationFormRow(rowIndex)">移除</Button></td>
                  </tr>
                </tbody>
              </table>
            </div>
            <div class="flex items-center justify-between gap-3"><p class="text-[11px] text-muted-foreground">{{ batchCreateWaybillRows.length }} / 50 条 · 所有订单使用同一环境、渠道和交货仓</p><Button variant="outline" size="sm" :disabled="batchCreateWaybillRows.length >= 50" @click="addYanwenBatchWaybillCreationFormRow"><Plus class="size-3.5" />增加订单</Button></div>
          </div>
          <DialogFooter><Button variant="outline" @click="batchCreateWaybillDialogOpen = false">取消</Button><Button :disabled="batchCreateWaybillSaving" @click="createYanwenBatchWaybills">{{ batchCreateWaybillSaving ? '推单中…' : '调用官方批量推单' }}</Button></DialogFooter>
        </DialogContent>
      </Dialog>
</template>
