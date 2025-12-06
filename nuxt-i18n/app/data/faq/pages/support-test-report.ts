import type { PageFaqData } from '../types'

/**
 * FAQ data for the Test Report support page
 */
export const supportTestReportFaq: PageFaqData = {
  pageId: 'support-test-report',
  title: 'Test Report FAQs',
  subtitle: 'Common questions about product testing and quality assurance',
  categories: [
    {
      id: 'testing',
      name: 'Product Testing',
      icon: '🔬',
      items: [
        {
          id: 'test-1',
          question: 'What tests do you perform on your products?',
          answer: `Our products undergo rigorous testing including:
            <ul>
              <li><strong>Impact testing</strong> - Simulating real-world impacts</li>
              <li><strong>Fatigue testing</strong> - Thousands of cycles to ensure durability</li>
              <li><strong>Brake heat testing</strong> - For rim brake compatibility</li>
              <li><strong>Spoke tension testing</strong> - Ensuring consistent wheel build quality</li>
              <li><strong>Weight verification</strong> - Confirming advertised specifications</li>
            </ul>`,
          tags: ['testing', 'quality', 'standards'],
        },
        {
          id: 'test-2',
          question: 'Are your products tested by third parties?',
          answer: `Yes, we work with independent testing laboratories to verify our products meet international standards. Third-party test reports are available for many of our products upon request.`,
          tags: ['third party', 'independent', 'certification'],
        },
        {
          id: 'test-3',
          question: 'What standards do your products meet?',
          answer: `Our products are designed and tested to meet or exceed relevant industry standards including UCI regulations for competitive cycling and ISO standards for bicycle components.`,
          tags: ['standards', 'UCI', 'ISO', 'regulations'],
        },
      ],
    },
    {
      id: 'reports',
      name: 'Accessing Reports',
      icon: '📄',
      items: [
        {
          id: 'rep-1',
          question: 'How can I get a test report for a specific product?',
          answer: `Test reports for specific products can be requested by contacting our support team. Please provide the product name or SKU, and we will share available documentation. Some reports are also available for download on individual product pages.`,
          tags: ['report', 'download', 'request'],
        },
        {
          id: 'rep-2',
          question: 'Are test reports available in multiple languages?',
          answer: `Most test reports are available in English. For major markets, we may have translated versions available. Please contact our support team to inquire about specific language availability.`,
          tags: ['report', 'language', 'translation'],
        },
      ],
    },
    {
      id: 'quality',
      name: 'Quality Assurance',
      icon: '✅',
      items: [
        {
          id: 'qa-1',
          question: 'How do you ensure consistent quality?',
          answer: `We maintain strict quality control throughout our manufacturing process:
            <ul>
              <li>Incoming material inspection</li>
              <li>In-process quality checks at each production stage</li>
              <li>Final inspection before packaging</li>
              <li>Random sampling for destructive testing</li>
            </ul>`,
          tags: ['quality', 'control', 'inspection'],
        },
        {
          id: 'qa-2',
          question: 'What if I receive a defective product?',
          answer: `If you receive a product that doesn't meet our quality standards, please contact our support team immediately. We will arrange for a replacement or refund. Defective products are covered under our warranty policy.`,
          tags: ['defective', 'quality', 'replacement'],
        },
      ],
    },
  ],
}
