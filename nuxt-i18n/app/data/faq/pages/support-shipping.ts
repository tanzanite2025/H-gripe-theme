import type { PageFaqData } from '../types'

/**
 * FAQ data for the Shipping support page
 */
export const supportShippingFaq: PageFaqData = {
  pageId: 'support-shipping',
  title: 'Shipping FAQs',
  subtitle: 'Common questions about shipping, delivery times, and tracking',
  categories: [
    {
      id: 'delivery',
      name: 'Delivery & Timing',
      icon: '📦',
      items: [
        {
          id: 'ship-1',
          question: 'How long does shipping take?',
          answer: `Shipping times vary by destination:
            <ul>
              <li><strong>Domestic (China)</strong>: 3-5 business days</li>
              <li><strong>Asia Pacific</strong>: 7-14 business days</li>
              <li><strong>Europe</strong>: 10-20 business days</li>
              <li><strong>North America</strong>: 10-20 business days</li>
              <li><strong>Other regions</strong>: 14-30 business days</li>
            </ul>
            Note: Custom-built wheelsets may require additional 3-7 days for assembly before shipping.`,
          tags: ['delivery', 'time', 'duration'],
        },
        {
          id: 'ship-2',
          question: 'What is the order processing time?',
          answer: `Order processing typically takes 1-3 business days for in-stock items. For custom wheel builds, please allow an additional 3-7 business days for assembly and quality inspection before shipping.`,
          tags: ['processing', 'handling', 'time'],
        },
        {
          id: 'ship-3',
          question: 'Can I get express shipping?',
          answer: `Yes, we offer express shipping options for most destinations. Express shipping typically reduces delivery time by 50% or more. Please contact our support team for express shipping quotes and availability for your location.`,
          tags: ['express', 'fast', 'urgent'],
        },
      ],
    },
    {
      id: 'tracking',
      name: 'Tracking & Updates',
      icon: '🔍',
      items: [
        {
          id: 'track-1',
          question: 'How can I track my order?',
          answer: `Once your order ships, you will receive an email with a tracking number and link. You can also log into your account on our website to view the latest shipping status. For any tracking issues, please contact our support team.`,
          tags: ['tracking', 'status', 'updates'],
        },
        {
          id: 'track-2',
          question: 'My tracking hasn\'t updated in several days. Is this normal?',
          answer: `Yes, this can be normal, especially for international shipments. Tracking updates may pause during customs clearance or when packages transfer between carriers. If there's no update for more than 7 days, please contact our support team.`,
          tags: ['tracking', 'delay', 'customs'],
        },
      ],
    },
    {
      id: 'international',
      name: 'International Shipping',
      icon: '🌍',
      items: [
        {
          id: 'intl-1',
          question: 'Do you ship internationally?',
          answer: `Yes, we ship to most countries worldwide. Shipping costs and delivery times vary by destination. Some remote areas may have limited shipping options or longer delivery times.`,
          tags: ['international', 'worldwide', 'global'],
        },
        {
          id: 'intl-2',
          question: 'Will I have to pay customs duties or import taxes?',
          answer: `Depending on your country's regulations, you may be required to pay customs duties, import taxes, or VAT upon delivery. These charges are determined by your local customs authority and are the responsibility of the recipient. We recommend checking with your local customs office for more information.`,
          tags: ['customs', 'duties', 'tax', 'import'],
        },
        {
          id: 'intl-3',
          question: 'What documents are included with international shipments?',
          answer: `All international shipments include a commercial invoice and packing list. For certain destinations, we may also include a certificate of origin or other required documentation. If you need specific documents for customs clearance, please let us know before shipping.`,
          tags: ['documents', 'invoice', 'customs'],
        },
      ],
    },
    {
      id: 'issues',
      name: 'Shipping Issues',
      icon: '⚠️',
      items: [
        {
          id: 'issue-1',
          question: 'What if my package is damaged during shipping?',
          answer: `If your package arrives damaged, please take photos of the packaging and contents immediately. Contact our support team within 48 hours of delivery with the photos and your order number. We will work with the carrier to file a claim and arrange a replacement or refund.`,
          tags: ['damaged', 'broken', 'claim'],
        },
        {
          id: 'issue-2',
          question: 'What if my package is lost?',
          answer: `If your tracking shows no updates for an extended period or indicates the package is lost, please contact our support team. We will investigate with the carrier and, if the package cannot be located, arrange a replacement shipment or refund.`,
          tags: ['lost', 'missing', 'claim'],
        },
      ],
    },
  ],
}
