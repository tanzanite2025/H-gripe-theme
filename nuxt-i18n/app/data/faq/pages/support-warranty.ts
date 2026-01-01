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
      id: 'policy',
      name: 'Warranty Policy',
      icon: '🛡️',
      items: [
        {
          id: 'war-1',
          question: 'What is the warranty period for Tanzanite wheels?',
          answer: `We offer two coverage options for all Tanzanite series Wheels/Rims:
            <ul>
              <li><strong>Standard Warranty</strong>: 5 Years from the date of purchase (included by default).</li>
              <li><strong>Lifetime Warranty</strong>: Available as an upgrade for USD $100 per rim.</li>
            </ul>`,
          tags: ['warranty', 'period', 'duration'],
        },
        {
          id: 'war-2',
          question: 'What does the warranty cover?',
          answer: `Our warranty covers <strong>manufacturing defects</strong> in materials and workmanship. If a structural failure occurs due to a defect within the warranty period, we will replace the rim.`,
          tags: ['coverage', 'defects'],
        },
        {
          id: 'war-3',
          question: 'Is the warranty transferable?',
          answer: `No, the warranty applies only to the <strong>original purchaser</strong> and is non-transferable. Valid proof of purchase is required for all claims.`,
          tags: ['transferable', 'original owner'],
        },
        {
          id: 'war-4',
          question: 'What isn\'t covered by the warranty?',
          answer: `The warranty does not cover damage caused by <strong>improper assembly</strong>, use of incompatible parts, unauthorized modifications, or normal wear and tear. Accidental damage (crashes) is covered under our Crash Replacement Policy.`,
          tags: ['exclusions', 'coverage', 'not covered'],
        },
        {
          id: 'war-5',
          question: 'What happens if my rim is discontinued?',
          answer: `If a warranty replacement is approved but your specific rim model is discontinued, we will upgrade you to the <strong>latest equivalent model</strong> at no additional cost.`,
          tags: ['discontinued', 'upgrade', 'replacement'],
        },
      ],
    },
    {
      id: 'crash',
      name: 'Accidental Damage',
      icon: '💥',
      items: [
        {
          id: 'crash-1',
          question: 'Do you offer crash replacement?',
          answer: `Yes, we offer a <strong>Crash Replacement Policy</strong> for accidental damage (e.g., crashes, jumps, rock impacts) that is not covered under the standard warranty.`,
          tags: ['crash replacement', 'accidental'],
        },
        {
          id: 'crash-2',
          question: 'What are the terms for crash replacement?',
          answer: `This coverage is valid for <strong>3 years</strong> from the date of purchase. You can receive a replacement rim at a <strong>10% discount</strong>.`,
          tags: ['discount', 'crash terms'],
        },
      ],
    },
    {
      id: 'claims',
      name: 'Claims & Process',
      icon: '📋',
      items: [
        {
          id: 'claim-1',
          question: 'How do I submit a warranty claim?',
          answer: `You can submit a claim directly on this page under the <strong>"Submit Warranty"</strong> tab. Please have your order number, photos, and a description of the issue ready.`,
          tags: ['submit claim', 'process'],
        },
        {
          id: 'claim-2',
          question: 'Who covers shipping costs?',
          answer: `Shipping responsibility depends on when the claim is made:
            <ul>
              <li><strong>Within 30 days</strong> of receipt: Tanzanite covers shipping costs.</li>
              <li><strong>After 30 days</strong>: The customer is responsible for shipping costs.</li>
            </ul>`,
          tags: ['shipping', 'costs'],
        },
        {
          id: 'claim-3',
          question: 'Do I need to return the damaged product?',
          answer: `In most cases, <strong>no</strong>. We typically require clear photos and videos of the damage. If we do need the item returned for inspection, we will issue a return authorization.`,
          tags: ['return', 'damaged item'],
        },
      ],
    },
  ],
}
