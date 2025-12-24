import type { PageFaqData } from '../types'

/**
 * FAQ data for the Tire Guides page
 */
export const guidesTireguidesFaq: PageFaqData = {
  pageId: 'guides-tireguides',
  title: 'Tire Guides FAQs',
  subtitle: 'Key questions about tire sizing, pressure, and inner tubes',
  categories: [
    {
      id: 'sizing',
      name: 'Sizing & compatibility',
      icon: '📏',
      items: [
        {
          id: 'size-1',
          question: 'How do I read tire size and match an inner tube?',
          answer: `Check the size printed on your tire sidewall (e.g., 700x28C or 29x2.2, plus the ETRTO number). Choose an inner tube whose width range covers your tire size and with the correct diameter.`,
          tags: ['size', 'inner tube', 'etrto'],
        },
        {
          id: 'size-2',
          question: 'Which valve type should I pick for my rim?',
          answer: `Match the valve hole and use case:<ul>
            <li><strong>AV (Schrader)</strong> - common on many rims; larger hole.</li>
            <li><strong>DV (Dunlop/Woods)</strong> - traditional bicycle valve.</li>
            <li><strong>SV (Presta)</strong> - narrow hole, popular on road/MTB.</li>
            <li><strong>SV Extra long</strong> - for deep rims; choose 50/60/80 mm as needed.</li>
          </ul>`,
          tags: ['valve', 'presta', 'schrader', 'rim'],
        },
      ],
    },
    {
      id: 'pressure',
      name: 'Tire pressure & setup',
      icon: '⚙️',
      items: [
        {
          id: 'pressure-1',
          question: 'How should I set tire pressure for road vs gravel?',
          answer: `Start from the manufacturer’s recommended range and adjust for rider weight and terrain. Road tires typically run higher pressure for speed; gravel and wider tires use lower pressure for comfort and grip. Avoid exceeding max pressure printed on the tire.`,
          tags: ['pressure', 'road', 'gravel'],
        },
        {
          id: 'tubeless-1',
          question: 'Can I run tubeless on any tire and rim?',
          answer: `Use tubeless-ready tires and rims with proper tape and valves. Non-tubeless components may not seal reliably and can burp air. Always check the rim/tire manufacturer’s tubeless compatibility.`,
          tags: ['tubeless', 'compatibility', 'setup'],
        },
      ],
    },
    {
      id: 'maintenance',
      name: 'Maintenance',
      icon: '🛠️',
      items: [
        {
          id: 'flat-1',
          question: 'How can I reduce pinch flats with tubes?',
          answer: `Ensure correct tube size, keep pressure within the recommended range, and avoid trapping the tube during installation. Check rim tape for damage and inspect the tire for debris before mounting.`,
          tags: ['pinch flat', 'tube', 'maintenance'],
        },
        {
          id: 'storage-1',
          question: 'How should I store spare tubes?',
          answer: `Keep tubes in a cool, dry place away from direct sunlight and ozone sources. Avoid sharp folds; a loose roll or small pouch helps prevent creases that can weaken rubber over time.`,
          tags: ['storage', 'tube', 'care'],
        },
      ],
    },
  ],
}
