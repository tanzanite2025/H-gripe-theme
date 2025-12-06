import type { PageFaqData } from '../types'

/**
 * FAQ data for the Warranty Check page
 */
export const supportWarrantyCheckFaq: PageFaqData = {
  pageId: 'support-warranty-check',
  title: 'Warranty Check FAQs',
  subtitle: 'Common questions about checking your product warranty',
  categories: [
    {
      id: 'how-to-check',
      name: 'How to Check Warranty',
      icon: '🔍',
      items: [
        {
          id: 'wc-1',
          question: 'Where can I find my product code?',
          answer: `Your product code can be found in several places:
            <ul>
              <li><strong>Product packaging</strong>: On the original box or packaging</li>
              <li><strong>Warranty card</strong>: Included with your purchase</li>
              <li><strong>Product label</strong>: On the rim or hub sticker</li>
              <li><strong>Order confirmation email</strong>: Listed in your order details</li>
            </ul>
            The code is typically a combination of letters and numbers (e.g., TZ-2024-ABC123).`,
          tags: ['product code', 'find', 'location'],
        },
        {
          id: 'wc-2',
          question: 'Why do I need to log in to check warranty?',
          answer: `Logging in helps us:
            <ul>
              <li>Verify your ownership of the product</li>
              <li>Link warranty information to your account</li>
              <li>Provide personalized support if needed</li>
              <li>Keep your warranty history in one place</li>
            </ul>
            Your account information is kept secure and private.`,
          tags: ['login', 'account', 'security'],
        },
        {
          id: 'wc-3',
          question: 'What information will I see after checking?',
          answer: `After a successful warranty check, you'll see:
            <ul>
              <li><strong>Product details</strong>: Name, type, and specifications</li>
              <li><strong>Warranty status</strong>: Valid or expired</li>
              <li><strong>Ship date</strong>: When the product was shipped</li>
              <li><strong>Warranty period</strong>: Length of coverage</li>
              <li><strong>Expiration date</strong>: When warranty ends</li>
              <li><strong>Remaining time</strong>: Days/months left on warranty</li>
            </ul>`,
          tags: ['results', 'information', 'details'],
        },
      ],
    },
    {
      id: 'troubleshooting',
      name: 'Troubleshooting',
      icon: '🔧',
      items: [
        {
          id: 'wc-4',
          question: 'What if my product code is not found?',
          answer: `If your product code is not found, please check:
            <ul>
              <li>The code is entered correctly (no typos)</li>
              <li>You're using the correct format</li>
              <li>The product is a genuine Tanzanite product</li>
            </ul>
            If the issue persists, please contact our support team with your order details.`,
          tags: ['not found', 'error', 'troubleshoot'],
        },
        {
          id: 'wc-5',
          question: 'My warranty shows as expired but I just bought it. What should I do?',
          answer: `If your warranty appears expired incorrectly:
            <ol>
              <li>Check your purchase date against the displayed ship date</li>
              <li>Verify the warranty period for your product type</li>
              <li>Contact support with your order confirmation and receipt</li>
            </ol>
            We'll investigate and correct any discrepancies in our system.`,
          tags: ['expired', 'incorrect', 'dispute'],
        },
        {
          id: 'wc-6',
          question: 'Can I transfer warranty to a new owner?',
          answer: `Warranty transfer policies:
            <ul>
              <li>Warranty is generally tied to the original purchaser</li>
              <li>For second-hand purchases, contact us with proof of original purchase</li>
              <li>Transfer may be possible with proper documentation</li>
            </ul>
            Please contact our support team for warranty transfer requests.`,
          tags: ['transfer', 'second-hand', 'new owner'],
        },
      ],
    },
  ],
}
