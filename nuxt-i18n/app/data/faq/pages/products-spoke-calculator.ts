import type { PageFaqData } from '../types'

/**
 * FAQ data for the Spoke Calculator page
 */
export const productsSpokeCalculatorFaq: PageFaqData = {
    pageId: 'products-spoke-calculator',
    title: 'Spoke Calculator FAQs',
    subtitle: 'Common questions about using our calculator and obtaining accurate measurements',
    categories: [
        {
            id: 'usage',
            name: 'Calculator Usage',
            icon: '🧮',
            items: [
                {
                    id: 'spoke-1',
                    question: 'How do I use this calculator?',
                    answer: `To calculate the correct spoke length:
            <ol>
              <li><strong>Select your Rim</strong>: Choose from our product list or enter ERD manually.</li>
              <li><strong>Select your Hub</strong>: Choose from our product list or enter flange dimensions manually.</li>
              <li><strong>Configure setup</strong>: Set spoke count, lacing pattern (e.g., 3-cross), and nipple type.</li>
              <li><strong>Calculate</strong>: Click the button to get the recommended spoke lengths for left and right sides.</li>
            </ol>`,
                    tags: ['how to', 'guide', 'steps'],
                },
                {
                    id: 'spoke-2',
                    question: 'What is ERD and why is it important?',
                    answer: `ERD (Effective Rim Diameter) is the diameter of the rim at the point where the spoke nipples seat. It is the most critical dimension for spoke calculation. An incorrect ERD is the most common cause of wrong spoke lengths. We recommend measuring your specific rim's ERD yourself using two cut spokes and nipples to be absolutely sure along with the manufacturer's spec.`,
                    tags: ['erd', 'measurement', 'rim'],
                },
                {
                    id: 'spoke-3',
                    question: 'How accurate are the results?',
                    answer: `The results are mathematically precise based on the inputs provided. However, real-world variations in rim roundness, hub manufacturing tolerances, and nipple dimensions mean the calculated length is a theoretical ideal. We recommend rounding to the nearest available even millimeter length.`,
                    tags: ['accuracy', 'tolerance', 'result'],
                },
            ],
        },
        {
            id: 'troubleshooting',
            name: 'Troubleshooting',
            icon: '🔧',
            items: [
                {
                    id: 'trouble-1',
                    question: 'Why are my calculated lengths different from another calculator?',
                    answer: `Different calculators might use slightly different formulas or assumptions about spoke stretch. Our calculator uses standard trigonometric formulas including compensation for spoke hole offset (if applicable). Small differences of +/- 1mm are normal.`,
                    tags: ['difference', 'formula', 'comparison'],
                },
                {
                    id: 'trouble-2',
                    question: 'What if my hub is not in the list?',
                    answer: `If your hub is not in our dropdown list, you can manually enter the flange dimensions. You will need: Left/Right Flange Distance (center to flange), and Left/Right Flange Diameter (PCD). Consult your hub manufacturer's technical manual for these specs.`,
                    tags: ['hub', 'missing', 'manual input'],
                },
            ],
        },
    ],
}
