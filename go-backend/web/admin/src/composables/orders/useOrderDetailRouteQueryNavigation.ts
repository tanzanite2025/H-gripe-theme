import { onMounted, watch } from 'vue'
import { parsePositiveUint32RouteQueryParameter } from '@/lib/parsePositiveUint32RouteQueryParameter'

interface OrderDetailRouteQueryNavigationOptions {
  orderIdQuery: () => unknown
  openOrderDetailById: (orderId: number) => Promise<void>
  clearOrderDetailSelection: () => void
  reportInvalidOrderId: () => void
}

export function useOrderDetailRouteQueryNavigation(options: OrderDetailRouteQueryNavigationOptions): void {
  async function openOrderDetailFromRouteQuery(): Promise<void> {
    const rawOrderId = options.orderIdQuery()
    if (!rawOrderId) {
      options.clearOrderDetailSelection()
      return
    }
    const orderId = parsePositiveUint32RouteQueryParameter(rawOrderId)
    if (orderId === null) {
      options.clearOrderDetailSelection()
      options.reportInvalidOrderId()
      return
    }
    await options.openOrderDetailById(orderId)
  }
  onMounted(() => { void openOrderDetailFromRouteQuery() })
  watch(options.orderIdQuery, () => { void openOrderDetailFromRouteQuery() })
}
