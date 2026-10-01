export type HomePurchasePathActionKind = 'route' | 'contact-service' | 'quickbuy-options'

export interface HomePurchasePathQuickBuyAction {
  id: 'direct-select' | 'wheelset-selection-assistant'
  label: string
  icon: string
}

export interface HomePurchasePathCard {
  id: string
  title: string
  description: string
  highlights?: string[]
  quickBuyActions?: HomePurchasePathQuickBuyAction[]
  actionLabel?: string
  icon: string
  kind: HomePurchasePathActionKind
  route?: string
  chatSource?: string
}

export interface HomePurchasePathSection {
  eyebrow: string
  title: string
  cards: HomePurchasePathCard[]
}

export const homePurchasePathSection: HomePurchasePathSection = {
  eyebrow: 'BUYING PATH',
  title: 'Not sure which wheelset fits?',
  cards: [
    {
      id: 'fit-guide',
      title: 'Need a fit check?',
      description: 'Use the buyers guide when you want a clearer match before spending.',
      actionLabel: 'Open guide',
      icon: 'lucide:route',
      kind: 'route',
      route: '/guides/wheelset-buyers',
    },
    {
      id: 'talk-to-support',
      title: 'Contact support',
      description: 'Open QuickBuy support options to email us or continue in a service chat about your wheelset configuration.',
      actionLabel: 'Email or chat with support',
      icon: 'lucide:message-circle-more',
      kind: 'contact-service',
    },
    {
      id: 'quickbuy-options',
      title: 'Choose your QuickBuy path',
      description: 'Know your specs already, or need help finding them? Choose the route that matches your situation.',
      quickBuyActions: [
        {
          id: 'direct-select',
          label: 'I know my specs',
          icon: 'lucide:sliders-horizontal',
        },
        {
          id: 'wheelset-selection-assistant',
          label: 'Find my bike specs',
          icon: 'lucide:circle-help',
        },
      ],
      icon: 'lucide:zap',
      kind: 'quickbuy-options',
    },
  ],
}
