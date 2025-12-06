import type { PageFaqData } from '../types'

/**
 * FAQ data for the Membership and Points page
 */
export const companyMembershipFaq: PageFaqData = {
  pageId: 'company-membership',
  title: 'Membership & Points FAQs',
  subtitle: 'Common questions about membership tiers, benefits, and points',
  categories: [
    {
      id: 'membership',
      name: 'Membership Tiers',
      icon: '🏆',
      items: [
        {
          id: 'mem-1',
          question: 'What membership tiers are available?',
          answer: `We offer several membership tiers based on your purchase history:
            <ul>
              <li><strong>Bronze</strong> - New members (0-499 points)</li>
              <li><strong>Silver</strong> - 500-1,999 points</li>
              <li><strong>Gold</strong> - 2,000-4,999 points</li>
              <li><strong>Platinum</strong> - 5,000+ points</li>
            </ul>
            Higher tiers unlock better discounts and exclusive benefits.`,
          tags: ['membership', 'tiers', 'levels'],
        },
        {
          id: 'mem-2',
          question: 'How do I upgrade my membership tier?',
          answer: `Your membership tier is automatically upgraded based on your accumulated points. Points are earned through purchases, reviews, and other activities. Once you reach the required points threshold, your tier will be upgraded immediately.`,
          tags: ['upgrade', 'tier', 'points'],
        },
        {
          id: 'mem-3',
          question: 'Do membership tiers expire?',
          answer: `Membership tiers are evaluated annually. To maintain your current tier, you need to earn at least 50% of the tier's point threshold within each 12-month period. If you don't meet this requirement, you may be moved to a lower tier.`,
          tags: ['expire', 'maintain', 'tier'],
        },
      ],
    },
    {
      id: 'points',
      name: 'Points System',
      icon: '💎',
      items: [
        {
          id: 'pts-1',
          question: 'How do I earn points?',
          answer: `You can earn points through various activities:
            <ul>
              <li><strong>Purchases</strong> - 1 point per $1 spent</li>
              <li><strong>Product reviews</strong> - 50 points per review</li>
              <li><strong>Referrals</strong> - 200 points when a friend makes their first purchase</li>
              <li><strong>Daily login</strong> - 1 point per day (30-day validity)</li>
              <li><strong>Birthday bonus</strong> - Double points during your birthday month</li>
            </ul>`,
          tags: ['earn', 'points', 'rewards'],
        },
        {
          id: 'pts-2',
          question: 'How do I redeem my points?',
          answer: `Points can be redeemed at checkout for discounts on your order. The redemption rate is typically 100 points = $1 discount. You can choose how many points to redeem, up to a maximum of 20% of your order total.`,
          tags: ['redeem', 'points', 'discount'],
        },
        {
          id: 'pts-3',
          question: 'Do points expire?',
          answer: `Points earned from purchases are valid for 24 months from the date earned. Daily login points expire after 30 days. Bonus points from promotions may have different expiration periods - check the promotion details for specifics.`,
          tags: ['expire', 'points', 'validity'],
        },
        {
          id: 'pts-4',
          question: 'Can I transfer points to another account?',
          answer: `Points are non-transferable and can only be used by the account holder. Each account must earn and redeem their own points.`,
          tags: ['transfer', 'points', 'account'],
        },
      ],
    },
    {
      id: 'benefits',
      name: 'Member Benefits',
      icon: '🎁',
      items: [
        {
          id: 'ben-1',
          question: 'What benefits do members receive?',
          answer: `Member benefits vary by tier and include:
            <ul>
              <li>Exclusive member-only discounts</li>
              <li>Early access to new products and sales</li>
              <li>Free shipping on orders over a certain amount</li>
              <li>Birthday rewards and special promotions</li>
              <li>Priority customer support (Gold and above)</li>
            </ul>`,
          tags: ['benefits', 'perks', 'rewards'],
        },
        {
          id: 'ben-2',
          question: 'How do I access member-only deals?',
          answer: `Member-only deals are automatically displayed when you're logged into your account. You'll also receive email notifications about exclusive offers. Make sure your email preferences are set to receive promotional emails.`,
          tags: ['deals', 'exclusive', 'access'],
        },
      ],
    },
  ],
}
