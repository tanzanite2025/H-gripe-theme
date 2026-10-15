import { afterEach, describe, expect, it, vi } from 'vitest'
import axios from '@/utils/axios'
import opsApi from './ops'
import { fetchDashboardBusinessFollowUpDetail, parseDashboardCommandCenterOutboxFailureResponse } from './dashboardCommandCenterOutboxFailures'

afterEach(() => vi.restoreAllMocks())

function createReadyDashboardOutboxFailureContractFixture() {
  return {
    timestamp: '2026-10-15T08:00:00Z',
    action_center: {
      outbox_failures: {
        status: 'ready', total_count: 8, failed_count: 6, dead_letter_count: 2,
        manual_attention_count: 2, automatic_retry_count: 6, preview_limit: 5,
        items: Array.from({ length: 5 }, (_, index) => ({
          id: index + 1, title: '付款确认邮件发送未完成', category: '客户通知',
          business_impact: '客户可能未收到通知，付款状态以订单记录为准。', responsible_team: '邮件服务负责人',
          sales_next_steps: ['打开订单核查付款记录。'], reason_summary: '服务响应超时。',
          processing_state: index < 2 ? 'manual_attention' : 'automatic_retry_scheduled',
          attempts: 3, max_attempts: 10, updated_at: '2026-10-15T07:00:00Z',
          created_at: '2026-10-15T06:00:00Z', last_attempt_at: '2026-10-15T07:00:00Z',
          next_retry_at: index < 2 ? null : '2026-10-15T08:30:00Z',
          related_record: { label: '订单 SALES-1', path: '/orders/list?order_id=1', required_permission: 'order:view' },
          related_record_notice: '',
        })),
      },
    },
    system_degraded_nodes: [] as string[],
  }
}

describe('dashboard Outbox availability and count contract', () => {
  it('rejects a valid business detail that belongs to a different requested event', async () => {
    const item = createReadyDashboardOutboxFailureContractFixture().action_center.outbox_failures.items[0]
    vi.spyOn(axios, 'get').mockResolvedValue({ data: { ...item, id: 2 } })
    await expect(fetchDashboardBusinessFollowUpDetail(1)).rejects.toThrow('编号与请求不匹配')
  })

  it('rejects a technical detail for another event or an event outside the failure queue', async () => {
    const mockedRequest = vi.spyOn(axios, 'get')
    mockedRequest.mockResolvedValueOnce({ data: { id: 2, status: 'failed' } })
    await expect(opsApi.getOutboxFailure(1)).rejects.toThrow('编号或失败状态与请求不匹配')
    mockedRequest.mockResolvedValueOnce({ data: { id: 1, status: 'processed' } })
    await expect(opsApi.getOutboxFailure(1)).rejects.toThrow('编号或失败状态与请求不匹配')
  })
  it('preserves full queue count when the preview is limited to five items', () => {
    const response = parseDashboardCommandCenterOutboxFailureResponse(createReadyDashboardOutboxFailureContractFixture())
    expect(response.action_center.outbox_failures.total_count).toBe(8)
    expect(response.action_center.outbox_failures.items).toHaveLength(5)
  })

  it('represents an unavailable query with unknown counts and an explicit degraded node', () => {
    const fixture = createReadyDashboardOutboxFailureContractFixture()
    const response = parseDashboardCommandCenterOutboxFailureResponse({
      ...fixture,
      action_center: { outbox_failures: { status: 'unavailable', total_count: null, failed_count: null, dead_letter_count: null, manual_attention_count: null, automatic_retry_count: null, preview_limit: 5, items: [] } },
      system_degraded_nodes: ['outbox_failures'],
    })
    expect(response.action_center.outbox_failures.total_count).toBeNull()
  })

  it('rejects missing sales guidance, contradictory retry schedules and impossible processing counts', () => {
    const fixture = createReadyDashboardOutboxFailureContractFixture()
    fixture.action_center.outbox_failures.items[0].sales_next_steps = []
    expect(() => parseDashboardCommandCenterOutboxFailureResponse(fixture)).toThrow()
    const retryFixture = createReadyDashboardOutboxFailureContractFixture()
    retryFixture.action_center.outbox_failures.items[2].next_retry_at = null
    expect(() => parseDashboardCommandCenterOutboxFailureResponse(retryFixture)).toThrow()
    const countFixture = createReadyDashboardOutboxFailureContractFixture()
    countFixture.action_center.outbox_failures.manual_attention_count = 1
    countFixture.action_center.outbox_failures.automatic_retry_count = 7
    expect(() => parseDashboardCommandCenterOutboxFailureResponse(countFixture)).toThrow()
  })

  it('preserves valid counts when related business references are unavailable', () => {
    const fixture = createReadyDashboardOutboxFailureContractFixture()
    fixture.system_degraded_nodes = ['outbox_business_references']
    const response = parseDashboardCommandCenterOutboxFailureResponse({
      ...fixture,
      action_center: { outbox_failures: {
        ...fixture.action_center.outbox_failures,
        items: fixture.action_center.outbox_failures.items.map((item) => ({ ...item, related_record: null, related_record_notice: '关联资料暂时无法读取，待办数量仍有效。' })),
      } },
    })
    expect(response.action_center.outbox_failures.total_count).toBe(8)
    expect(response.action_center.outbox_failures.items[0].related_record).toBeNull()
  })

  it('rejects unavailable data represented as healthy zero counts', () => {
    const fixture = createReadyDashboardOutboxFailureContractFixture()
    expect(() => parseDashboardCommandCenterOutboxFailureResponse({
      ...fixture,
      action_center: { outbox_failures: { status: 'unavailable', total_count: 0, failed_count: 0, dead_letter_count: 0, manual_attention_count: 0, automatic_retry_count: 0, preview_limit: 5, items: [] } },
      system_degraded_nodes: ['outbox_failures'],
    })).toThrow()
  })

  it('rejects missing counts, missing degradation markers and inconsistent totals', () => {
    const fixture = createReadyDashboardOutboxFailureContractFixture()
    const { total_count: _omittedCount, ...incompleteFailures } = fixture.action_center.outbox_failures
    expect(() => parseDashboardCommandCenterOutboxFailureResponse({ ...fixture, action_center: { outbox_failures: incompleteFailures } })).toThrow()
    expect(() => parseDashboardCommandCenterOutboxFailureResponse({
      ...fixture, action_center: { outbox_failures: { status: 'unavailable', total_count: null, failed_count: null, dead_letter_count: null, manual_attention_count: null, automatic_retry_count: null, preview_limit: 5, items: [] } },
    })).toThrow()
    fixture.action_center.outbox_failures.total_count = 5
    expect(() => parseDashboardCommandCenterOutboxFailureResponse(fixture)).toThrow()
  })
})
