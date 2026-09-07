import type { RefundCancellationPolicy } from '~/types/refundCancellationPolicy'

export const refundCancellationPolicyFallback: RefundCancellationPolicy = {
  title: 'Refund & Cancellation Policy',
  intro: 'How we handle cancellations, refunds, returns, and exchanges to keep your experience predictable and fair.',
  sections: [
    {
      id: 'eligibility',
      title: 'Eligibility',
      body: 'For stocked, non-custom items, we accept return requests within 30 days of delivery when the item is unused and in original packaging. Made-to-order and custom-configured items follow the cancellation rules in Special Orders.',
    },
    {
      id: 'condition',
      title: 'Condition of Items',
      body: 'Returned items must be unused, undamaged, and include all accessories/manuals. We reserve the right to refuse returns that do not meet these conditions.',
    },
    {
      id: 'special-orders',
      title: 'Special Orders',
      body: 'Non-stock or custom-configured products (special orders) are made to order. Cancellation may be requested before production or material cutting starts; after production begins, cancellation may be unavailable. If a post-production cancellation is approved, a 15%-20% custom handling fee may be deducted for committed materials and labor.',
    },
    {
      id: 'high-value-signature',
      title: 'High-Value Delivery',
      body: 'Orders totaling $750 USD or more require a direct signature at delivery. The signature requirement is part of our delivery and dispute-protection process.',
    },
    {
      id: 'process',
      title: 'Process',
      bullets: [
        'Contact our support team with your order number and issue details.',
        'We will provide return instructions and, if applicable, a return authorization.',
        'Ship the item using a trackable method; retain proof of shipment.',
      ],
    },
    {
      id: 'refund-timing',
      title: 'Refund Method & Timing',
      body: 'Refunds are processed to the original payment method within 5-10 business days after we receive and inspect the returned item.',
    },
    {
      id: 'shipping-costs',
      title: 'Shipping Costs',
      body: "Return shipping is the customer's responsibility unless the item is defective or incorrect. Original shipping fees are non-refundable unless required by law.",
    },
    {
      id: 'restocking-fee',
      title: 'Restocking & Refurbishment',
      body: 'If an eligible return requires refurbishment or is missing original packaging or accessories, a fee of up to 20% of the original purchase value, with a minimum of USD $100, may be deducted after review.',
    },
    {
      id: 'other-costs',
      title: 'Other Costs',
      body: 'Unless otherwise specified under the warranty policy, shipping fees, duties, taxes, and any additional charges are borne by the customer.',
    },
    {
      id: 'exchanges',
      title: 'Exchanges',
      body: 'For exchanges, please initiate a return first, then place a new order once the return is approved. This ensures availability and faster processing.',
    },
  ],
  contact_label: 'For refund, cancellation, or return questions, contact our support team through the contact page.',
  contact_url: '/company/contact',
  updated_at: '2026-09-04T00:00:00Z',
}
