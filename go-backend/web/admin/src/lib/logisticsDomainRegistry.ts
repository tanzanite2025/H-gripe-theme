export type LogisticsDomainKey = 'fpx' | 'yanwen'

export interface LogisticsTabDefinition {
  key: string
  id: string
  path: string
  routeName: string
  label: string
  title: string
  description: string
  permission: string
}

export interface LogisticsDomainDefinition {
  key: LogisticsDomainKey
  id: string
  code: string
  label: string
  path: string
  permission: string
  defaultRouteName: string
  tabs: readonly LogisticsTabDefinition[]
}

const viewPermissions = {
  fpx: 'logistics:fpx:view',
  yanwen: 'logistics:yanwen:view',
} as const

const managePermissions = {
  fpx: 'logistics:fpx:manage',
  yanwen: 'logistics:yanwen:manage',
} as const

export const fpxLogisticsTabs = [
  { key: 'overview', id: 'fpx-overview', path: '/logistics/4px/overview', routeName: 'FPXOverview', label: '运营大盘', title: '4PX大盘', description: '发货吞吐与履约态势', permission: viewPermissions.fpx },
  { key: 'direct', id: 'fpx-direct', path: '/logistics/4px/direct', routeName: 'FPXDirectShipping', label: '大件直发', title: '4PX大件直发', description: '专线下单、预报与面单', permission: viewPermissions.fpx },
  { key: 'collection', id: 'fpx-collection', path: '/logistics/4px/collection', routeName: 'FPXCollection', label: '服务集合', title: '4PX服务集合', description: '发布集合与短链路', permission: viewPermissions.fpx },
  { key: 'calculator', id: 'fpx-calculator', path: '/logistics/4px/calculator', routeName: 'FPXCalculator', label: '运费试算', title: '4PX运费试算', description: '材积重与渠道选优', permission: viewPermissions.fpx },
  { key: 'tracking', id: 'fpx-tracking', path: '/logistics/4px/tracking', routeName: 'FPXTracking', label: '全球追踪', title: '4PX全球追踪', description: '轨迹查询与授权证明', permission: viewPermissions.fpx },
  { key: 'rma', id: 'fpx-rma', path: '/logistics/4px/rma', routeName: 'FPXReverseLogistics', label: '逆向退件', title: '4PX逆向退件', description: '内部台账与承运商状态', permission: viewPermissions.fpx },
  { key: 'config', id: 'fpx-config', path: '/logistics/4px/config', routeName: 'FPXConfig', label: '接口配置', title: '4PX接口配置', description: '凭据与主数据同步', permission: managePermissions.fpx },
] as const satisfies readonly LogisticsTabDefinition[]

export const yanwenLogisticsTabs = [
  { key: 'overview', id: 'yanwen-overview', path: '/logistics/yanwen/overview', routeName: 'YanwenOverview', label: '运营大盘', title: '燕文大盘', description: '出口吞吐与链路健康', permission: viewPermissions.yanwen },
  { key: 'waybills', id: 'yanwen-waybills', path: '/logistics/yanwen/waybills', routeName: 'YanwenWaybills', label: '专线运单', title: '燕文专线运单', description: '推单、面单与交运', permission: viewPermissions.yanwen },
  { key: 'collection', id: 'yanwen-collection', path: '/logistics/yanwen/collection', routeName: 'YanwenCollection', label: '服务集合', title: '燕文服务集合', description: '渠道池与发布集合', permission: viewPermissions.yanwen },
  { key: 'calculator', id: 'yanwen-calculator', path: '/logistics/yanwen/calculator', routeName: 'YanwenCalculator', label: '运价试算', title: '燕文运价试算', description: '报价与路由选优', permission: viewPermissions.yanwen },
  { key: 'tracking', id: 'yanwen-tracking', path: '/logistics/yanwen/tracking', routeName: 'YanwenTracking', label: '轨迹监控', title: '燕文轨迹监控', description: '全链路轨迹与时效', permission: viewPermissions.yanwen },
  { key: 'customs', id: 'yanwen-customs', path: '/logistics/yanwen/customs', routeName: 'YanwenCustoms', label: '关务合规', title: '燕文关务合规', description: '前置校验与异常拦截', permission: viewPermissions.yanwen },
  { key: 'config', id: 'yanwen-config', path: '/logistics/yanwen/config', routeName: 'YanwenConfig', label: '网关配置', title: '燕文网关配置', description: '凭据与主数据同步', permission: managePermissions.yanwen },
] as const satisfies readonly LogisticsTabDefinition[]

export const logisticsDomains = {
  fpx: {
    key: 'fpx',
    id: 'fpx-logistics',
    code: 'FPX',
    label: '4PX 物流域',
    path: '/logistics/4px',
    permission: viewPermissions.fpx,
    defaultRouteName: 'FPXOverview',
    tabs: fpxLogisticsTabs,
  },
  yanwen: {
    key: 'yanwen',
    id: 'yanwen-logistics',
    code: 'YANWEN',
    label: '燕文物流域',
    path: '/logistics/yanwen',
    permission: viewPermissions.yanwen,
    defaultRouteName: 'YanwenOverview',
    tabs: yanwenLogisticsTabs,
  },
} as const satisfies Record<LogisticsDomainKey, LogisticsDomainDefinition>

export const tabRouteMap = (tabs: readonly LogisticsTabDefinition[]): Record<string, string> => (
  Object.fromEntries(tabs.map((tab) => [tab.key, tab.routeName]))
)
