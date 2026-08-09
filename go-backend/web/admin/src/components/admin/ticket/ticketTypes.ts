export type TicketId = number | string
export type TicketBadgeTone = 'blue' | 'green' | 'amber' | 'coral' | 'gray'

export interface TicketUser {
  id?: TicketId | null
  username?: string | null
  email?: string | null
}

export interface TicketMessage {
  id: TicketId
  sender_name?: string | null
  user?: TicketUser | null
  is_staff?: boolean
  content?: string | null
  message?: string | null
  created_at?: string | null
}

export interface TicketRecord {
  id: TicketId
  ticket_number?: string | null
  subject?: string | null
  category?: string | null
  status?: string | null
  priority?: string | null
  user_name?: string | null
  user_id?: TicketId | null
  user?: TicketUser | null
  assigned_to?: TicketId | null
  created_at?: string | null
  updated_at?: string | null
  tags?: string | null
  messages?: TicketMessage[]
}

export interface TicketFilters {
  search: string
  status: string
  priority: string
}

export interface TicketPagination {
  page: number
  pageSize: number
  total: number
}

export interface TicketStats {
  total?: number
  open?: number
  in_progress?: number
  resolved?: number
  closed?: number
}

export interface TicketListPayload {
  tickets?: TicketRecord[]
  pagination?: {
    total?: number
  }
}

export interface TicketStatsPayload extends TicketStats {}

export interface TicketDetailPayload {
  ticket?: TicketRecord
}

export interface TicketMessagesPayload {
  messages?: TicketMessage[]
}

export interface TicketSupportUsersPayload {
  users?: TicketUser[]
}

export interface TicketConfirmation {
  open: boolean
  target: TicketRecord | null
  title: string
  description: string
}

export type TicketLabelResolver = (value?: string | null) => string
export type TicketToneResolver = (value?: string | null) => TicketBadgeTone
export type TicketDateFormatter = (value?: string | null) => string
export type TicketCustomerResolver = (ticket: TicketRecord) => string
export type TicketAssigneeResolver = (assignedTo?: TicketId | null) => string
export type TicketMessageSenderResolver = (message: TicketMessage) => string
