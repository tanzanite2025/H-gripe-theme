export interface OrderEvidenceOutboundWeightPackagingForm {
  gross_weight_g: number | null
  package_count: number | null
  packaging_method: string
  packaging_note: string
}

export interface OrderEvidenceSignedPODForm {
  tracking_number: string
  delivered_at: string
  recipient_name: string
  proof_reference: string
  delivery_note: string
}
