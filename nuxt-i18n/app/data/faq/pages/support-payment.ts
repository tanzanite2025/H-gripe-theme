import type { PageFaqData } from '../types'

/**
 * FAQ data for the Payment support page
 */
export const supportPaymentFaq: PageFaqData = {
  pageId: 'support-payment',
  title: 'Payment FAQs',
  subtitle: 'Payment methods, payment security, and what to do when something goes wrong',
  categories: [
    {
      id: 'payment-methods',
      name: 'Payment Methods',
      icon: '💳',
      items: [
        {
          id: 'pm-1',
          question: 'What payment methods do you accept?',
          answer: `The checkout currently supports:
            <ul>
              <li><strong>Credit and debit cards</strong> - Visa, Mastercard, American Express, and other major cards through Stripe</li>
              <li><strong>PayPal</strong> - Hosted PayPal checkout</li>
              <li><strong>Alipay</strong> - Redirect payment through Alipay</li>
              <li><strong>WeChat Pay</strong> - Native QR payment</li>
            </ul>
            The methods shown at checkout are the methods currently enabled for your order and region.`,
          tags: ['payment', 'methods', 'cards', 'paypal', 'alipay', 'wechat pay'],
        },
        {
          id: 'pm-2',
          question: 'Why do I not see every payment method?',
          answer: `Payment availability can depend on the order, region, provider configuration, and temporary provider status. The checkout page is the final source for the payment methods available for your purchase.`,
          tags: ['availability', 'region', 'checkout', 'methods'],
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
          answer: `Yes. Our pages use HTTPS/TLS encryption, and payment details are handled by trusted providers including Stripe, PayPal, Alipay, and WeChat Pay. We verify provider webhooks or payment status before treating an order as paid, and we match the result to the order amount and currency.`,
          tags: ['security', 'https', 'tls', 'encryption', 'privacy'],
        },
        {
          id: 'sec-2',
          question: 'Do you store my card number or CVV?',
          answer: `We do not store your full card number, CVV, or card PIN. We receive the payment result and the provider reference needed to reconcile your order and help with support.`,
          tags: ['card number', 'cvv', 'pin', 'privacy', 'data'],
        },
        {
          id: 'sec-3',
          question: 'What is 3D Secure verification?',
          answer: `3D Secure is an additional security check from your bank or card issuer. Depending on your card and region, you may be asked to approve the payment in your bank app or enter a one-time code. This helps protect against unauthorized card use.`,
          tags: ['3d secure', 'verification', 'security', 'bank'],
        },
      ],
    },
    {
      id: 'payment-confirmation',
      name: 'Payment Confirmation',
      icon: '✓',
      items: [
        {
          id: 'confirm-1',
          question: 'How is my payment confirmed?',
          answer: `Each provider has its own confirmation flow. Stripe uses its payment confirmation flow, PayPal uses approval and provider confirmation, and Alipay and WeChat Pay use provider-side status or notification checks. Our server verifies the provider result against the order before marking it as paid.`,
          tags: ['confirmation', 'provider', 'webhook', 'status'],
        },
        {
          id: 'confirm-2',
          question: 'What if I close the payment page too early?',
          answer: `A browser redirect is not the only source of truth. If the provider has completed the payment, the server may still receive or verify the result. Do not submit the payment repeatedly; keep your receipt or provider reference and contact support if the order status needs checking.`,
          tags: ['redirect', 'closed page', 'status', 'receipt'],
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
              <li>Try again or use another method shown at checkout</li>
            </ul>
            Your bank or payment provider can often provide more detail through its app or support team.`,
          tags: ['declined', 'rejected', 'failed'],
        },
        {
          id: 'ts-2',
          question: 'I was charged but my order was not created. What happened?',
          answer: `A network or redirect issue can interrupt the return to our site after payment. Do not pay again immediately. Contact support with your order details, payment receipt, or provider reference so we can verify the payment and order status.`,
          tags: ['charged', 'no order', 'missing order', 'redirect'],
        },
        {
          id: 'ts-3',
          question: 'I see multiple pending charges on my card. Is this normal?',
          answer: `If you attempted payment several times, your bank may show multiple pending authorizations. Unused pending authorizations are normally released by the bank or payment provider. If an unfamiliar charge remains or you are concerned, contact your bank and our support team.`,
          tags: ['multiple charges', 'pending', 'authorization'],
        },
      ],
    },
  ],
}
