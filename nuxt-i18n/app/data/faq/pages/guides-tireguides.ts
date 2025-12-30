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
          question: 'Which valve type and length should I pick?',
          answer: `First, match the valve type to your rim hole (Presta/SV for road/MTB, Schrader/AV for wider holes).
          <br><br>
          <strong>Rule of thumb for length:</strong> The valve should extend at least 15mm above the rim.
          <ul>
            <li><strong>Low profile (≤30mm)</strong>: Standard 40mm valves.</li>
            <li><strong>Mid profile (30-45mm)</strong>: 60mm valves.</li>
            <li><strong>Deep profile (50-65mm)</strong>: 80mm valves.</li>
            <li><strong>Ultra deep (>70mm)</strong>: Use an 80mm valve with an extender.</li>
          </ul>`,
          tags: ['valve', 'presta', 'schrader', 'rim', 'length'],
        },
        {
          id: 'size-3',
          question: 'When do I need a valve extender and which type?',
          answer: `If your rims are deeper than 50-60mm, standard valves might be too short to pump.
          <ul>
            <li><strong>Internal (Core Removable)</strong>: Best choice. You remove the valve core, screw the extender into the valve shaft, and put the core back on top. Allows for easier pumping and pressure adjustments.</li>
            <li><strong>External</strong>: Screws onto the valve tip. Easier to install but requires the valve to remain open, which can sometimes leak air.</li>
          </ul>`,
          tags: ['valve extender', 'deep rim', 'components'],
        },
      ],
    },
    {
      id: 'installation',
      name: 'Installation & Setup',
      icon: '🔧',
      items: [
        {
          id: 'installation-1',
          question: 'Tips for seating difficult tubeless tires?',
          answer: `Seating tubeless tires can be tricky. Try these steps:
          <ol>
            <li><strong>Use soapy water</strong>: Apply generously to the tire beads and rim hook to help them slide into place.</li>
            <li><strong>Remove valve core</strong>: This allows air to enter much faster, creating the sudden pressure needed to "pop" the beads into place.</li>
            <li><strong>Massage the tire</strong>: Ensure the tire beads are sitting in the center channel (the lowest part) of the rim before inflating.</li>
            <li><strong>Use a booster</strong>: If a floor pump fails, use a compressor or a tubeless booster canister.</li>
          </ol>`,
          tags: ['tubeless', 'installation', 'tips'],
        },
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
