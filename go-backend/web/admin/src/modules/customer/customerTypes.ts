export type CustomerStatusTone = 'green' | 'gray' | 'coral'
export type CustomerStatusNameResolver = (status?: string | null) => string
export type CustomerStatusToneResolver = (status?: string | null) => CustomerStatusTone
export type CustomerDateFormatter = (value?: string | null) => string
export type CustomerNameFormatter = (customer: CustomerAccount) => string

export interface CustomerAccount {
  id: number | string
  username?: string | null
  email?: string | null
  first_name?: string | null
  last_name?: string | null
  display_name?: string | null
  status?: string | null
  created_at?: string | null
}

export interface CustomerFilters {
  search: string
  status: string
}

export interface CustomerPagination {
  page: number
  pageSize: number
  total: number
}

export interface CustomerListResponse {
  customers?: CustomerAccount[]
  pagination?: {
    total?: number
  }
}
