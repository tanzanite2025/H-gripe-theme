import type { RefundCancellationPolicy } from '~/types/refundCancellationPolicy'

export const refundCancellationPolicyFallback: RefundCancellationPolicy = {
  title: 'Refund & Cancellation Policy',
  intro: 'How we handle cancellations, refunds, returns, and exchanges to keep your experience predictable and fair.',
  sections: [
    {
      id: 'eligibility',
      title: 'Eligibility',
      body: 'We accept returns within 30 days of delivery for unused items in original packaging. Custom or personalized items are non-refundable unless defective.',
    },
    {
      id: 'condition',
      title: 'Condition of Items',
      body: 'Returned items must be unused, undamaged, and include all accessories/manuals. We reserve the right to refuse returns that do not meet these conditions.',
    },
    {
      id: 'special-orders',
      title: 'Special Orders',
      body: 'Non-stock or custom-configured products (special orders) are not eligible for return or refund, unless the issue is caused by our error.',
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
      title: 'Restocking & Refurbishment Fee',
      body: 'A restocking and refurbishment fee may be charged at 20% of the original purchase value, with a minimum of USD $100.',
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
  updated_at: '2024-12-01T00:00:00Z',
}
