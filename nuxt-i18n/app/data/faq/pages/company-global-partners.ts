import type { PageFaqData } from '../types'

export const companyGlobalPartnersFaq: PageFaqData = {
    id: 'company-global-partners',
    title: 'Global Partnerships FAQ',
    subtitle: 'Common questions about becoming a global partner',
    categories: [
        {
            id: 'partnership',
            name: 'Partnership Program',
            items: [
                {
                    id: 'partnership-criteria',
                    question: 'What are the criteria for becoming a distributor?',
                    answer: 'We look for partners with established distribution channels, technical service capabilities, and a commitment to brand building. A background in premium cycling components is preferred.',
                },
                {
                    id: 'oem-support',
                    question: 'Do you offer OEM/ODM services for global brands?',
                    answer: 'Yes, we provide comprehensive OEM/ODM solutions, including mold design, layup optimization, and private labeling for established bicycle brands.',
                },
                {
                    id: 'shipping-logistics',
                    question: 'How do you handle global evaluations and logistics?',
                    answer: 'We have extensive experience in global logistics, including DDP shipping to key markets. We can also arrange sample evaluations for potential partners to verify our quality standards.',
                },
                {
                    id: 'exclusivity',
                    question: 'Is regional exclusivity available?',
                    answer: 'Regional exclusivity is negotiable based on projected volume, market coverage, and a proven track record of sales performance.',
                },
            ],
        },
    ],
}
