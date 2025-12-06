import type { PageFaqData } from '../types'

/**
 * FAQ data for the Product Feedback support page
 */
export const supportProductFeedbackFaq: PageFaqData = {
  pageId: 'support-product-feedback',
  title: 'Product Feedback FAQs',
  subtitle: 'Common questions about sharing feedback and suggestions',
  categories: [
    {
      id: 'feedback',
      name: 'Submitting Feedback',
      icon: '💬',
      items: [
        {
          id: 'fb-1',
          question: 'How can I submit product feedback?',
          answer: `You can submit feedback through several channels:
            <ul>
              <li>Use the feedback form on this page</li>
              <li>Email our support team directly</li>
              <li>Leave a review on your order</li>
              <li>Contact us via WhatsApp</li>
            </ul>
            We read and consider all feedback to improve our products.`,
          tags: ['feedback', 'submit', 'how to'],
        },
        {
          id: 'fb-2',
          question: 'What kind of feedback are you looking for?',
          answer: `We welcome all types of feedback including:
            <ul>
              <li>Product performance and durability</li>
              <li>Build quality and finish</li>
              <li>Suggestions for new products or features</li>
              <li>Issues or problems you've encountered</li>
              <li>Comparison with other products</li>
            </ul>`,
          tags: ['feedback', 'types', 'suggestions'],
        },
        {
          id: 'fb-3',
          question: 'Will I receive a response to my feedback?',
          answer: `We read all feedback but may not respond to every submission individually. If your feedback requires follow-up or contains a specific question, our team will reach out to you via email.`,
          tags: ['feedback', 'response', 'reply'],
        },
      ],
    },
    {
      id: 'reviews',
      name: 'Product Reviews',
      icon: '⭐',
      items: [
        {
          id: 'rev-1',
          question: 'How can I leave a product review?',
          answer: `After receiving your order, you can leave a review by:
            <ul>
              <li>Clicking the review link in your order confirmation email</li>
              <li>Logging into your account and visiting your order history</li>
              <li>Visiting the product page and clicking "Write a Review"</li>
            </ul>`,
          tags: ['review', 'how to', 'rating'],
        },
        {
          id: 'rev-2',
          question: 'Can I edit or delete my review?',
          answer: `Yes, you can edit or delete your review by logging into your account and visiting your order history. Find the order containing the reviewed product and click "Edit Review" or "Delete Review".`,
          tags: ['review', 'edit', 'delete'],
        },
      ],
    },
  ],
}
