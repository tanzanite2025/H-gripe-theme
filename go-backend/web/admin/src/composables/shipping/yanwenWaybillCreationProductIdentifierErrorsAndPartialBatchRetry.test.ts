import { beforeEach, describe, expect, it, vi } from 'vitest'

const apiMocks = vi.hoisted(() => ({
  createYanwenWaybill: vi.fn(),
  createYanwenWaybills: vi.fn(),
  listYanwenWaybills: vi.fn().mockResolvedValue([]),
}))
vi.mock('@/api/yanwenLogisticsAdminApi', () => ({ yanwenLogisticsAdminApi: apiMocks }))
vi.mock('@/stores/auth', () => ({ useAuthStore: () => ({ hasPermission: () => true }) }))
vi.mock('vue', async (importOriginal) => ({ ...await importOriginal<typeof import('vue')>(), onMounted: vi.fn() }))

import { useYanwenWaybillOperations } from './useYanwenWaybillOperations'

describe('Yanwen product identifier failures', () => {
  beforeEach(() => { vi.clearAllMocks() })

  it('preserves the single-order form and exposes the actionable product error', async () => {
    const operations = useYanwenWaybillOperations()
    operations.createWaybillDialogOpen.value = true
    Object.assign(operations.createWaybillForm.value, { orderId: '11', channelId: '1', warehouseCode: 'WH01', hasBattery: 'false' })
    apiMocks.createYanwenWaybill.mockRejectedValue(new Error('SKU-A 缺少制造商产品识别码（NSPID），本次未提交燕文'))
    await operations.createWaybill()
    expect(operations.createWaybillDialogOpen.value).toBe(true)
    expect(operations.createWaybillForm.value.orderId).toBe('11')
    expect(operations.waybillActionMessage.value).toContain('NSPID')
    expect(operations.createWaybillSaving.value).toBe(false)
  })

  it('reports every failed order and retries only rows without successful waybills', async () => {
    const operations = useYanwenWaybillOperations()
    operations.curatedChannels.value = [{ id: 1, enabled: true }] as any
    operations.batchCreateWaybillDialogOpen.value = true
    operations.batchCreateWaybillForm.value = { channelId: '1', warehouseCode: 'WH01' }
    operations.batchCreateWaybillRows.value = ['11', '12', '13'].map((orderId) => ({ orderId, hasBattery: 'false', receiverTaxNumber: '', ioss: '', eori: '' }))
    apiMocks.createYanwenWaybills.mockResolvedValueOnce({
      succeeded: 1, failed: 2,
      items: [
        { request_index: 0, order_id: 11, error: 'SKU-A 缺少 MPID' },
        { request_index: 1, order_id: 12, waybill_id: 100 },
        { request_index: 2, order_id: 13, error: 'SKU-C 缺少 NSPID' },
      ],
    })
    await operations.createYanwenBatchWaybills()
    expect(operations.batchCreateWaybillDialogOpen.value).toBe(true)
    expect(operations.batchCreateWaybillRows.value.map((row) => row.orderId)).toEqual(['11', '13'])
    expect(operations.waybillActionMessage.value).toContain('订单 ID 11：SKU-A 缺少 MPID')
    expect(operations.waybillActionMessage.value).toContain('订单 ID 13：SKU-C 缺少 NSPID')
    apiMocks.createYanwenWaybills.mockResolvedValueOnce({ succeeded: 2, failed: 0, items: [
      { request_index: 0, order_id: 11, waybill_id: 101 },
      { request_index: 1, order_id: 13, waybill_id: 102 },
    ] })
    await operations.createYanwenBatchWaybills()
    expect(apiMocks.createYanwenWaybills.mock.calls[1][0].map((request: { order_id: number }) => request.order_id)).toEqual([11, 13])
    expect(operations.batchCreateWaybillRows.value).toEqual([])
    expect(operations.batchCreateWaybillDialogOpen.value).toBe(false)
  })

  it('keeps rows when the server response does not confirm their outcome', async () => {
    const operations = useYanwenWaybillOperations()
    operations.curatedChannels.value = [{ id: 1, enabled: true }] as any
    operations.batchCreateWaybillDialogOpen.value = true
    operations.batchCreateWaybillForm.value = { channelId: '1', warehouseCode: 'WH01' }
    operations.batchCreateWaybillRows.value = [{ orderId: '11', hasBattery: 'false', receiverTaxNumber: '', ioss: '', eori: '' }]
    apiMocks.createYanwenWaybills.mockResolvedValueOnce({ succeeded: 0, failed: 0, items: [] })
    await operations.createYanwenBatchWaybills()
    expect(operations.batchCreateWaybillRows.value[0]?.orderId).toBe('11')
    expect(operations.batchCreateWaybillDialogOpen.value).toBe(true)
    expect(operations.waybillActionMessage.value).toContain('未确认成功')
  })
})
