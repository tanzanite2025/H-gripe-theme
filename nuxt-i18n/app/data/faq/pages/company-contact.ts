import type { PageFaqData } from '../types'

export const companyContactFaq: PageFaqData = {
    id: 'company-contact',
    title: 'Contact & Support FAQ',
    subtitle: 'Common questions about reaching our team',
    categories: [
        {
            id: 'general',
            name: 'General Support',
            items: [
                {
                    id: 'response-time',
                    question: 'What is your typical response time?',
                    answer: 'We strive to respond to all inquiries within 24 hours during business days (Mon-Fri). For urgent technical support, please use our WhatsApp channel for faster assistance.',
                },
                {
                    id: 'factory-visit',
                    question: 'Can I visit the Xiamen factory?',
                    answer: 'Yes, we welcome visits from our OEM/ODM partners. Please contact our sales team at least 2 weeks in advance to schedule a tour and ensure the appropriate staff are available to meet you.',
                },
                {
                    id: 'local-distributors',
                    question: 'Do you have local distributors in my country?',
                    answer: 'We operate primarily through a direct-to-consumer model to offer the best prices. However, we have a growing network of service partners in key regions. Please contact us to find the nearest partner.',
                },
                {
                    id: 'sponsorship',
                    question: 'How can I apply for sponsorship?',
                    answer: 'We are always looking for passionate riders and teams. Please send your racing resume and proposal to our support email with the subject line "Sponsorship Application".',
                },
            ],
        },
    ],
}
