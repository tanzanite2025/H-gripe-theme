export interface PaymentMethodAvailability {
  id?: number
  name?: string
  code?: string
  provider?: string
  icon?: string
  description?: string
  fee_type?: string
  fee_value?: number
  min_amount?: number
  max_amount?: number
  enabled?: boolean
  available?: boolean
  unavailable_reason?: string
  unavailableReason?: string
  sort_order?: number
}

export interface CheckoutPaymentOption {
  id: string
  title: string
  subtitle: string
  description: string
  points?: string[]
  code?: string
  provider?: string
  icon?: string
  enabled?: boolean
  available?: boolean
  unavailableReason?: string
  unavailable_reason?: string
}
