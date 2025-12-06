import type { PageFaqData } from '../types'

/**
 * FAQ data for the Warranty support page
 */
export const supportWarrantyFaq: PageFaqData = {
  pageId: 'support-warranty',
  title: 'Warranty FAQs',
  subtitle: 'Common questions about warranty coverage and claims',
  categories: [
    {
      id: 'coverage',
      name: 'Warranty Coverage',
      icon: '🛡️',
      items: [
        {
          id: 'war-1',
          question: 'What is covered under warranty?',
          answer: `Our warranty covers manufacturing defects in materials and workmanship. This includes:
            <ul>
              <li>Rim structural failures under normal use</li>
              <li>Hub bearing defects</li>
              <li>Spoke breakage due to manufacturing defects</li>
              <li>Freehub mechanism failures</li>
            </ul>
            Normal wear and tear, crash damage, and improper use are not covered.`,
          tags: ['warranty', 'coverage', 'defects'],
        },
        {
          id: 'war-2',
          question: 'How long is the warranty period?',
          answer: `Our standard warranty periods are:
            <ul>
              <li><strong>Carbon rims</strong>: 3 years from purchase date</li>
              <li><strong>Aluminum rims</strong>: 2 years from purchase date</li>
              <li><strong>Hubs</strong>: 2 years from purchase date</li>
              <li><strong>Spokes & nipples</strong>: 1 year from purchase date</li>
            </ul>
            Please keep your purchase receipt as proof of purchase date.`,
          tags: ['warranty', 'period', 'duration'],
        },
        {
          id: 'war-3',
          question: 'What is NOT covered under warranty?',
          answer: `The following are not covered:
            <ul>
              <li>Damage from crashes, accidents, or impacts</li>
              <li>Normal wear and tear (brake tracks, bearings, etc.)</li>
              <li>Damage from improper installation or maintenance</li>
              <li>Modifications or repairs by unauthorized parties</li>
              <li>Cosmetic damage that doesn't affect function</li>
            </ul>`,
          tags: ['warranty', 'exclusions', 'not covered'],
        },
      ],
    },
    {
      id: 'claims',
      name: 'Warranty Claims',
      icon: '📋',
      items: [
        {
          id: 'claim-1',
          question: 'How do I make a warranty claim?',
          answer: `To make a warranty claim:
            <ol>
              <li>Contact our support team with your order number</li>
              <li>Provide photos of the defect and the product</li>
              <li>Describe when and how the issue occurred</li>
              <li>Wait for our team to review and respond (usually within 2-3 business days)</li>
            </ol>
            If approved, we will provide instructions for return or replacement.`,
          tags: ['claim', 'process', 'how to'],
        },
        {
          id: 'claim-2',
          question: 'What happens if my warranty claim is approved?',
          answer: `If your claim is approved, we will either repair or replace the defective component at no charge. In some cases, we may offer a partial refund or credit toward a new product. Shipping costs for warranty replacements are covered by us.`,
          tags: ['claim', 'approved', 'replacement'],
        },
        {
          id: 'claim-3',
          question: 'How long does the warranty claim process take?',
          answer: `The typical timeline is:
            <ul>
              <li><strong>Initial review</strong>: 2-3 business days</li>
              <li><strong>Return shipping</strong>: Varies by location</li>
              <li><strong>Inspection & decision</strong>: 3-5 business days after receipt</li>
              <li><strong>Replacement shipping</strong>: 2-3 business days after approval</li>
            </ul>`,
          tags: ['claim', 'time', 'duration'],
        },
      ],
    },
  ],
}
