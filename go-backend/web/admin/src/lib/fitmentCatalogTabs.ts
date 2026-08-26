import type { AdminNavigationItem } from '@/lib/adminNavigation'
import type { useAdminI18n } from '@/i18n'

type Translate = ReturnType<typeof useAdminI18n>['t']

export const buildFitmentCatalogTabs = (t: Translate): AdminNavigationItem[] => [
  {
    id: 'fitment-frame-entries',
    path: '/fitment-catalog/frame-entries',
    routeName: 'FitmentFrameEntries',
    label: t('fitmentCatalog.frame.title'),
  },
  {
    id: 'fitment-fork-entries',
    path: '/fitment-catalog/fork-entries',
    routeName: 'FitmentForkEntries',
    label: t('fitmentCatalog.fork.title'),
  },
  {
    id: 'fitment-hub-specifications',
    path: '/fitment-catalog/hub-specifications',
    routeName: 'FitmentHubSpecifications',
    label: t('fitmentCatalog.hub.title'),
  },
]
