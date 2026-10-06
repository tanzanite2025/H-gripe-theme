export type CustomsReviewLifecycleStatus = 'healthy' | 'expiring_soon' | 'expired' | 'pending'

export const CUSTOMS_REVIEW_EXPIRING_SOON_DAYS = 30

const MILLISECONDS_PER_DAY = 86_400_000

interface CustomsReviewDateParts {
  year: number
  month: number
  day: number
}

const parseCustomsReviewDateParts = (value: string | null | undefined): CustomsReviewDateParts | null => {
  const dateText = value?.trim().slice(0, 10) || ''
  const match = /^(\d{4})-(\d{2})-(\d{2})$/.exec(dateText)
  if (!match) return null

  const year = Number(match[1])
  const month = Number(match[2])
  const day = Number(match[3])
  const date = new Date(year, month - 1, day)
  if (
    date.getFullYear() !== year
    || date.getMonth() !== month - 1
    || date.getDate() !== day
  ) {
    return null
  }

  return { year, month, day }
}

const getReferenceDateParts = (referenceDate: Date): CustomsReviewDateParts => ({
  year: referenceDate.getFullYear(),
  month: referenceDate.getMonth() + 1,
  day: referenceDate.getDate(),
})

const getCalendarDayDifference = (
  fromDate: CustomsReviewDateParts,
  toDate: CustomsReviewDateParts,
): number => {
  const fromUtc = Date.UTC(fromDate.year, fromDate.month - 1, fromDate.day)
  const toUtc = Date.UTC(toDate.year, toDate.month - 1, toDate.day)
  return Math.round((toUtc - fromUtc) / MILLISECONDS_PER_DAY)
}

export const getCustomsReviewLifecycleStatus = (
  verifiedAt: string | null | undefined,
  reviewDueAt: string | null | undefined,
  referenceDate: Date = new Date(),
): CustomsReviewLifecycleStatus => {
  if (Number.isNaN(referenceDate.getTime())) return 'pending'

  const verifiedDate = parseCustomsReviewDateParts(verifiedAt)
  const reviewDueDate = parseCustomsReviewDateParts(reviewDueAt)
  if (!verifiedDate || !reviewDueDate) return 'pending'

  const daysUntilReview = getCalendarDayDifference(
    getReferenceDateParts(referenceDate),
    reviewDueDate,
  )
  if (daysUntilReview < 0) return 'expired'
  if (daysUntilReview <= CUSTOMS_REVIEW_EXPIRING_SOON_DAYS) return 'expiring_soon'
  return 'healthy'
}

export const getCustomsReviewLifecycleStatusLabel = (
  status: CustomsReviewLifecycleStatus,
): string => {
  const labels: Record<CustomsReviewLifecycleStatus, string> = {
    healthy: '健康',
    expiring_soon: '临期复核',
    expired: '已过期',
    pending: '待核验',
  }
  return labels[status]
}

export const getCustomsReviewLifecycleStatusBadgeClass = (
  status: CustomsReviewLifecycleStatus,
): string => {
  const classes: Record<CustomsReviewLifecycleStatus, string> = {
    healthy: 'border-emerald-500/30 bg-emerald-500/10 text-emerald-700 dark:text-emerald-300',
    expiring_soon: 'border-amber-500/30 bg-amber-500/10 text-amber-700 dark:text-amber-300',
    expired: 'border-red-500/40 bg-red-500/10 text-red-700 dark:text-red-300',
    pending: 'border-slate-500/30 bg-slate-500/10 text-slate-700 dark:text-slate-300',
  }
  return classes[status]
}

export const customsReviewLifecycleStatusRequiresAttention = (
  status: CustomsReviewLifecycleStatus,
): boolean => status !== 'healthy'
