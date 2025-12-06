import type { PageFaqData } from '../types'

/**
 * FAQ data for the Payment support page
 */
export const supportPaymentFaq: PageFaqData = {
  pageId: 'support-payment',
  title: 'Payment FAQs',
  subtitle: 'Common questions about payment methods, fees, and troubleshooting',
  categories: [
    {
      id: 'payment-methods',
      name: 'Payment Methods',
      icon: '💳',
      items: [
        {
          id: 'pm-1',
          question: 'What payment methods do you accept?',
          answer: `We accept a variety of payment methods including:
            <ul>
              <li><strong>Credit & Debit Cards</strong> - Visa, MasterCard, and other major cards</li>
              <li><strong>PayPal</strong> - For faster checkout with buyer protection</li>
              <li><strong>Stripe</strong> - Secure card processing</li>
              <li><strong>WeChat Pay & Alipay</strong> - For customers in China</li>
              <li><strong>Bank Transfer</strong> - For larger or custom orders (by request)</li>
              <li><strong>WorldFirst</strong> - For selected international orders</li>
            </ul>`,
          tags: ['payment', 'methods', 'cards', 'paypal'],
        },
        {
          id: 'pm-2',
          question: 'Are there any payment processing fees?',
          answer: `Yes, different payment methods have different processing fees:
            <ul>
              <li><strong>PayPal & Credit Cards</strong>: 3.5% of the order amount</li>
              <li><strong>Stripe</strong>: 3.5% of the order amount</li>
              <li><strong>WeChat Pay</strong>: 0.6% of the order amount</li>
              <li><strong>Alipay</strong>: 1% of the order amount</li>
              <li><strong>WorldFirst</strong>: 1% of the order amount</li>
              <li><strong>Bank Transfer</strong>: $45 USD bank fee (your bank may charge additional fees)</li>
            </ul>`,
          tags: ['fees', 'processing', 'charges'],
        },
        {
          id: 'pm-3',
          question: 'Can I pay via bank transfer?',
          answer: `Yes, for larger or custom orders we can arrange payment via bank transfer. Please contact our support team before placing the order so we can provide the correct account details and reserve your items. Note that bank transfers incur a $45 USD bank fee, and your bank may charge additional fees.`,
          tags: ['bank', 'transfer', 'wire'],
        },
      ],
    },
    {
      id: 'security',
      name: 'Security & Privacy',
      icon: '🔒',
      items: [
        {
          id: 'sec-1',
          question: 'Is my payment information secure?',
          answer: `Absolutely. All pages on our site use HTTPS with SSL certificates issued by trusted certificate authorities. Your payment information is transmitted over encrypted channels. We use reputable payment providers (PayPal, Stripe, Alipay, WeChat Pay) to process payments - we only receive the payment result and <strong>never store</strong> your card number, CVV, or other sensitive data.`,
          tags: ['security', 'ssl', 'encryption', 'privacy'],
        },
        {
          id: 'sec-2',
          question: 'What is 3D Secure verification?',
          answer: `3D Secure is an additional security layer for online card payments. Depending on your card issuer and region, you may be asked to verify your identity through your bank's app or a one-time code sent to your phone. This helps protect against unauthorized use of your card.`,
          tags: ['3d secure', 'verification', 'security'],
        },
      ],
    },
    {
      id: 'billing',
      name: 'Billing & Charges',
      icon: '🧾',
      items: [
        {
          id: 'bill-1',
          question: 'When will my card be charged?',
          answer: `For card and wallet payments, an authorization is created when you confirm the order. The final capture normally happens when the order is accepted and prepared for shipment. For bank transfers, we start processing your order after the funds have been received and matched to your order reference.`,
          tags: ['charge', 'capture', 'authorization'],
        },
        {
          id: 'bill-2',
          question: 'Why is the charged amount different from the displayed price?',
          answer: `Product prices are usually shown in a reference currency (e.g., USD). Your bank or payment provider may convert this amount into your local currency using their own exchange rate and may add conversion fees. Additionally, depending on your shipping country, local VAT or import duties may apply.`,
          tags: ['currency', 'conversion', 'exchange rate'],
        },
        {
          id: 'bill-3',
          question: 'What about taxes and import duties?',
          answer: `Depending on your shipping country, local VAT or import duties may apply. These charges are handled according to the shipping option shown at checkout. Please review the order summary carefully before confirming payment.`,
          tags: ['tax', 'vat', 'duties', 'import'],
        },
      ],
    },
    {
      id: 'troubleshooting',
      name: 'Troubleshooting',
      icon: '🔧',
      items: [
        {
          id: 'ts-1',
          question: 'My payment was declined. What should I do?',
          answer: `If your card or wallet is declined, please:
            <ul>
              <li>Check that your billing details match your bank records</li>
              <li>Ensure sufficient funds are available</li>
              <li>Try again or use a different payment method</li>
            </ul>
            Your bank can often provide more detail via their app or customer service.`,
          tags: ['declined', 'rejected', 'failed'],
        },
        {
          id: 'ts-2',
          question: 'I was charged but my order was not created. What happened?',
          answer: `On rare occasions, a network issue can interrupt the redirect back to our site after payment. If you see a charge but no order in your account, please contact our support team with your payment reference so we can investigate and resolve the issue.`,
          tags: ['charged', 'no order', 'missing order'],
        },
        {
          id: 'ts-3',
          question: 'I see multiple pending charges on my card. Is this normal?',
          answer: `If you attempted payment several times, your bank may show multiple pending authorizations. Normally, any unused authorizations are released automatically after a short period (usually 3-7 business days). If this does not happen, please contact your bank or our support team.`,
          tags: ['multiple charges', 'pending', 'authorization'],
        },
      ],
    },
  ],
}
