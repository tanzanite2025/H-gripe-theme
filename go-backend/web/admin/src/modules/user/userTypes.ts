export type UserId = number | string
export type UserDialogMode = 'create' | 'edit'
export type UserRole = 'admin' | 'manager' | 'editor' | 'support' | 'viewer'
export type UserStatus = 'active' | 'inactive' | 'suspended'
export type UserBadgeTone = 'blue' | 'green' | 'amber' | 'coral' | 'gray'
export type UserSelectionState = boolean | 'indeterminate'
export type UserConfirmationType = '' | 'toggle-status' | 'delete' | 'batch-delete'

export type UserLabelResolver = (value?: string | null) => string
export type UserToneResolver = (value?: string | null) => UserBadgeTone
export type UserDateFormatter = (value?: string | null) => string
export type UserNameFormatter = (user: UserRecord) => string

export interface UserRecord {
  id: UserId
  email?: string | null
  username?: string | null
  first_name?: string | null
  last_name?: string | null
  role?: string | null
  locale?: string | null
  status?: string | null
  created_at?: string | null
}

export interface UserFilters {
  search: string
  role: string
  status: string
}

export interface UserPagination {
  page: number
  pageSize: number
  total: number
}

export interface UserFormValues {
  email: string
  username: string
  password: string
  first_name: string
  last_name: string
  role: UserRole
  locale: string
  status: UserStatus
}

export interface UserListResponse {
  users?: UserRecord[]
  pagination?: {
    total?: number
  }
}

export interface UserConfirmation {
  open: boolean
  type: UserConfirmationType
  target: UserRecord | UserRecord[] | null
  title: string
  description: string
  confirmLabel: string
  destructive: boolean
}
