import type { PageFaqData } from '../types'

/**
 * FAQ data for the After-sales support page
 */
export const supportAfterSalesFaq: PageFaqData = {
  pageId: 'support-after-sales',
  title: 'After-sales FAQs',
  subtitle: 'Common questions about returns, exchanges, and repairs',
  categories: [
    {
      id: 'returns',
      name: 'Returns & Refunds',
      icon: '↩️',
      items: [
        {
          id: 'ret-1',
          question: 'What is your return policy?',
          answer: `We accept returns within 30 days of delivery for unused items in original packaging. Custom-built wheelsets may have different return conditions. Please contact our support team to initiate a return request.`,
          tags: ['return', 'refund', 'policy'],
        },
        {
          id: 'ret-2',
          question: 'How do I request a refund?',
          answer: `To request a refund, please contact our support team with your order number and reason for return. Once approved, we will provide return shipping instructions. Refunds are processed within 5-10 business days after we receive the returned item.`,
          tags: ['refund', 'money back'],
        },
        {
          id: 'ret-3',
          question: 'Who pays for return shipping?',
          answer: `For defective products or shipping errors, we cover return shipping costs. For other returns (change of mind, wrong size ordered, etc.), the customer is responsible for return shipping fees.`,
          tags: ['return', 'shipping', 'cost'],
        },
      ],
    },
    {
      id: 'exchanges',
      name: 'Exchanges',
      icon: '🔄',
      items: [
        {
          id: 'exc-1',
          question: 'Can I exchange my product for a different one?',
          answer: `Yes, exchanges are available within 30 days of delivery for unused items. Please contact our support team with your order number and the item you would like to exchange for. Availability and price differences may apply.`,
          tags: ['exchange', 'swap'],
        },
        {
          id: 'exc-2',
          question: 'How long does an exchange take?',
          answer: `Once we receive your returned item, we will ship the replacement within 2-3 business days. Total exchange time depends on shipping to and from your location.`,
          tags: ['exchange', 'time'],
        },
      ],
    },
    {
      id: 'repairs',
      name: 'Repairs & Service',
      icon: '🔧',
      items: [
        {
          id: 'rep-1',
          question: 'Do you offer repair services?',
          answer: `Yes, we offer repair services for our products. Common repairs include spoke replacement, hub servicing, and rim truing. Please contact our support team to discuss your repair needs and get a quote.`,
          tags: ['repair', 'service', 'fix'],
        },
        {
          id: 'rep-2',
          question: 'How much do repairs cost?',
          answer: `Repair costs vary depending on the type of repair needed. Minor repairs like spoke replacement start from $10-20. More extensive repairs will be quoted after assessment. Repairs covered under warranty are free of charge.`,
          tags: ['repair', 'cost', 'price'],
        },
        {
          id: 'rep-3',
          question: 'How do I send my product for repair?',
          answer: `Contact our support team first to describe the issue and get a repair authorization. We will provide shipping instructions and, if applicable, a prepaid shipping label. Please pack the item securely to prevent further damage during transit.`,
          tags: ['repair', 'shipping', 'send'],
        },
      ],
    },
  ],
}
