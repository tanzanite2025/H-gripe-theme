import type { PageFaqData } from '../types'

export const companyCertificatesFaq: PageFaqData = {
    id: 'company-certificates',
    title: 'Certificates & Testing FAQ',
    subtitle: 'Common questions about our quality standards and certifications',
    categories: [
        {
            id: 'general',
            name: 'General Inquiries',
            items: [
                {
                    id: 'uci-approval',
                    question: 'Are all TANZANITE wheels UCI approved?',
                    answer: 'Yes, our core road and track wheelsets are UCI approved and listed on the official Union Cycliste Internationale website. This means they are certified for use in all professional and amateur racing events worldwide.',
                },
                {
                    id: 'iso-4210',
                    question: 'What is ISO 4210 certification?',
                    answer: 'ISO 4210 is the international safety standard for bicycles and components. Meeting this standard ensures that our wheels have passed rigorous impact, fatigue, and environmental tests to guarantee rider safety.',
                },
                {
                    id: 'internal-testing',
                    question: 'How does your internal testing differ from standard requirements?',
                    answer: 'We believe standard compliance is just the starting point. Our internal laboratory tests subjects rims to impact energies and fatigue cycles that are 120% to 150% higher than ISO requirements to ensure durability under extreme conditions.',
                },
                {
                    id: 'test-reports',
                    question: 'Can I request specific test reports?',
                    answer: 'Yes, we can provide detailed test reports for specific production batches upon request for our ODM/OEM partners. Please contact our support team for more information.',
                },
            ],
        },
    ],
}
