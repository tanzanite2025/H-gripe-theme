import type { PageFaqData } from '../types'

/**
 * FAQ data for the Wheelset Buyers Guide page
 */
export const guidesWheelsetBuyersFaq: PageFaqData = {
  pageId: 'guides-wheelset-buyers',
  title: 'Wheelset Buyers Guide FAQs',
  subtitle: 'Common questions about choosing and customizing wheelsets',
  categories: [
    {
      id: 'choosing',
      name: 'Choosing Wheelsets',
      icon: '🎯',
      items: [
        {
          id: 'choose-1',
          question: 'How do I choose the right wheelset for my riding style?',
          answer: `Consider these factors when choosing a wheelset:
            <ul>
              <li><strong>Riding discipline</strong> - Road, gravel, MTB, or mixed use</li>
              <li><strong>Rim depth</strong> - Deeper rims for aero, shallower for climbing</li>
              <li><strong>Rim width</strong> - Wider for better tire support and comfort</li>
              <li><strong>Weight</strong> - Lighter for climbing, durability for rough terrain</li>
              <li><strong>Brake type</strong> - Rim brake or disc brake compatibility</li>
            </ul>`,
          tags: ['choose', 'selection', 'riding style'],
        },
        {
          id: 'choose-2',
          question: 'What is a mullet wheelset?',
          answer: `A mullet wheelset uses different wheel sizes front and rear - typically a 29" front wheel paired with a 27.5" rear wheel. This combination offers the rolling efficiency and obstacle clearance of a larger front wheel with the agility and acceleration of a smaller rear wheel. Popular for enduro and trail riding.`,
          tags: ['mullet', 'mixed', 'wheel size'],
        },
        {
          id: 'choose-3',
          question: 'Should I choose carbon or aluminum rims?',
          answer: `<strong>Carbon rims</strong> offer better stiffness-to-weight ratio, aerodynamics, and can be shaped more precisely. They're ideal for performance-focused riders.<br><br><strong>Aluminum rims</strong> are more affordable, easier to repair, and handle impacts well. They're great for everyday riding and rough conditions.`,
          tags: ['carbon', 'aluminum', 'material'],
        },
      ],
    },
    {
      id: 'customization',
      name: 'Customization',
      icon: '🎨',
      items: [
        {
          id: 'custom-1',
          question: 'Can I customize the appearance of my wheelset?',
          answer: `Yes! We offer several customization options:
            <ul>
              <li><strong>Laser engraving</strong> - Eco-friendly, precise designs in light gray</li>
              <li><strong>Waterslide decals</strong> - Full color graphics</li>
              <li><strong>Vinyl stickers</strong> - Removable, laser-cut options</li>
            </ul>
            Our graphics team will work with you to design the perfect look.`,
          tags: ['customize', 'decals', 'graphics', 'logo'],
        },
        {
          id: 'custom-2',
          question: 'What is laser engraving and why choose it?',
          answer: `Laser engraving uses precision laser technology to etch designs directly into the carbon rim surface. Benefits include:
            <ul>
              <li>Eco-friendly - reduces plastic waste from stickers</li>
              <li>Permanent - won't peel or fade</li>
              <li>Sleek appearance - subtle light gray finish</li>
              <li>No added weight</li>
            </ul>`,
          tags: ['laser', 'engraving', 'eco-friendly'],
        },
      ],
    },
    {
      id: 'specs',
      name: 'Specifications',
      icon: '📐',
      items: [
        {
          id: 'spec-1',
          question: 'What spoke count should I choose?',
          answer: `Spoke count affects strength, weight, and aerodynamics:
            <ul>
              <li><strong>20-24 spokes</strong> - Lighter, more aero, best for lighter riders and smooth roads</li>
              <li><strong>28-32 spokes</strong> - Stronger, more durable, better for heavier riders and rough terrain</li>
            </ul>
            We can help you choose the right spoke count based on your weight and riding style.`,
          tags: ['spokes', 'count', 'strength'],
        },
        {
          id: 'spec-2',
          question: 'What hub options are available?',
          answer: `We offer various hub options to match your needs:
            <ul>
              <li>Different engagement points (36T, 54T, etc.)</li>
              <li>Various axle standards (QR, thru-axle)</li>
              <li>Multiple freehub body options (Shimano, SRAM XD, Campagnolo)</li>
            </ul>
            Contact us for specific hub recommendations.`,
          tags: ['hub', 'freehub', 'axle'],
        },
      ],
    },
  ],
}
