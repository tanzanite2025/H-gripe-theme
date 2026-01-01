import type { PageFaqData } from '../types'

export const companyOemOdmFaq: PageFaqData = {
    id: 'company-oem-odm',
    title: 'OEM/ODM Services FAQ',
    subtitle: 'Common questions about our manufacturing services',
    categories: [
        {
            id: 'general',
            name: 'General Inquiries',
            items: [
                {
                    id: 'moq',
                    question: 'What is the Minimum Order Quantity (MOQ) for OEM/ODM?',
                    answer: 'Our MOQ is flexible to support growing brands. For standard OEM rims, MOQ can be as low as 20 pairs. For ODM (custom mold) projects, we typically require an initial commitment of 50-100 rims to amortize the mold costs effectively.',
                },
                {
                    id: 'design',
                    question: 'Do you offer design services for custom molds?',
                    answer: 'Yes, our in-house R&D team provides comprehensive design services, including 3D modeling, aerodynamic layout analysis, and graphic design for decals. We can work from a simple sketch or a detailed CAD file.',
                },
                {
                    id: 'lead-time',
                    question: 'What is the typical lead time for production?',
                    answer: 'For standard OEM orders, lead time is typically 25-35 days. For ODM projects involving new mold creation, the timeline is usually: 15 days for design confirmation, 25 days for mold opening, and 10 days for prototyping/testing.',
                },
                {
                    id: 'confidentiality',
                    question: 'Is my design confidential?',
                    answer: 'Absolutely. We sign strict Non-Disclosure Agreements (NDAs) with all our ODM partners. Your private molds and layup schedules are exclusive to your brand and will never be shared with or sold to other clients.',
                },
                {
                    id: 'warranty',
                    question: 'Do OEM/ODM wheels come with a warranty?',
                    answer: 'Yes, we stand behind our manufacturing quality. We offer a standard 2-year warranty on all OEM/ODM rims against manufacturing defects, with options to extend coverage based on specific partnership agreements.',
                },
            ],
        },
    ],
}
