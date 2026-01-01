import type { PageFaqData } from '../types'

export const companyOurStoryFaq: PageFaqData = {
    id: 'company-ourstory',
    title: 'Our Story & Brand FAQ',
    subtitle: 'Learn more about Tanzanite and our mission',
    categories: [
        {
            id: 'brand',
            name: 'Brand & Mission',
            items: [
                {
                    id: 'origin',
                    question: 'Where is Tanzanite based?',
                    answer: 'Tanzanite is a brand of Top Sports Co., Limited, with our Global Headquarters in Hong Kong and our state-of-the-art Manufacturing & R&D Base in Xiamen, China.',
                },
                {
                    id: 'mission',
                    question: 'What is Tanzanite’s core mission?',
                    answer: 'Our mission is to democratize high-performance cycling components by leveraging advanced manufacturing and direct-to-consumer efficiency, delivering premium carbon wheels without the traditional markup.',
                },
                {
                    id: 'sustainability',
                    question: 'How does Tanzanite approach sustainability?',
                    answer: 'We are committed to sustainable manufacturing practices, minimizing waste in our carbon layup process, and designing durable products that stand the test of time, reducing the need for frequent replacements.',
                },
            ],
        },
        {
            id: 'products',
            name: 'Product Philosophy',
            items: [
                {
                    id: 'design-philosophy',
                    question: 'What drives your product design?',
                    answer: 'We believe in data-driven design backed by rigorous testing. Every rim profile and layup schedule is optimized for the specific demands of its discipline, whether it’s aerodynamics for road or impact resistance for MTB.',
                },
            ],
        },
    ],
}
