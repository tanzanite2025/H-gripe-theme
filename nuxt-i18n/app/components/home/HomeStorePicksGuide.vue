<template>
  <section id="home-store-picks-guide" class="home-store-picks-guide bg-transparent tz-text-primary">
    <div class="page-content-shell px-0 md:px-6">
      <div class="home-store-picks-guide__grid">
        <article
          v-for="card in cards"
          :key="card.id"
          class="home-store-picks-guide__card premium-card"
        >
          <div class="home-store-picks-guide__content">
            <div class="home-store-picks-guide__title-row">
              <span class="home-store-picks-guide__icon" aria-hidden="true">
                <Icon :name="card.icon" class="h-6 w-6" />
              </span>
              <h3>{{ tt(card.titleKey) }}</h3>
            </div>
            <p>{{ tt(card.descriptionKey) }}</p>
          </div>

          <NuxtLink
            v-if="card.kind === 'route'"
            :to="localePath(card.to)"
            class="home-store-picks-guide__button premium-button"
          >
            <Icon name="lucide:arrow-right" class="h-4 w-4" aria-hidden="true" />
            {{ tt(card.actionLabelKey) }}
          </NuxtLink>

          <button
            v-else
            type="button"
            class="home-store-picks-guide__button premium-button"
            @click="handleCardAction(card)"
          >
            <Icon :name="card.actionIcon" class="h-4 w-4" aria-hidden="true" />
            {{ tt(card.actionLabelKey) }}
          </button>
        </article>
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import { useI18n, useLocalePath } from '#imports'

type HomeStorePicksGuideCard =
  | {
      id: string
      titleKey: string
      descriptionKey: string
      icon: string
      actionLabelKey: string
      actionIcon: string
      kind: 'scroll'
      targetId: string
    }
  | {
      id: string
      titleKey: string
      descriptionKey: string
      icon: string
      actionLabelKey: string
      actionIcon: string
      kind: 'route'
      to: string
    }
  | {
      id: string
      titleKey: string
      descriptionKey: string
      icon: string
      actionLabelKey: string
      actionIcon: string
      kind: 'quickbuy'
    }

const localePath = useLocalePath()
const { t } = useI18n()

const tt = (key: string) => String(t(key))

const cards: HomeStorePicksGuideCard[] = [
  {
    id: 'precision-matching',
    titleKey: 'home.storePicksGuide.precisionMatching.title',
    descriptionKey: 'home.storePicksGuide.precisionMatching.description',
    icon: 'lucide:scan-search',
    actionLabelKey: 'home.storePicksGuide.precisionMatching.action',
    actionIcon: 'lucide:arrow-down',
    kind: 'scroll',
    targetId: 'home-buying-path',
  },
  {
    id: 'custom-production',
    titleKey: 'home.storePicksGuide.customProduction.title',
    descriptionKey: 'home.storePicksGuide.customProduction.description',
    icon: 'lucide:factory',
    actionLabelKey: 'home.storePicksGuide.customProduction.action',
    actionIcon: 'lucide:arrow-right',
    kind: 'route',
    to: '/company/oem-odm',
  },
  {
    id: 'product-gallery',
    titleKey: 'home.storePicksGuide.productGallery.title',
    descriptionKey: 'home.storePicksGuide.productGallery.description',
    icon: 'lucide:images',
    actionLabelKey: 'home.storePicksGuide.productGallery.action',
    actionIcon: 'lucide:arrow-down',
    kind: 'scroll',
    targetId: 'home-brand-photo-carousel',
  },
  {
    id: 'safe-shopping',
    titleKey: 'home.storePicksGuide.paymentSafety.title',
    descriptionKey: 'home.storePicksGuide.paymentSafety.description',
    icon: 'lucide:shield-check',
    actionLabelKey: 'home.storePicksGuide.paymentSafety.action',
    actionIcon: 'lucide:arrow-down',
    kind: 'scroll',
    targetId: 'home-shop-with-confidence',
  },
  {
    id: 'custom-component-build',
    titleKey: 'home.storePicksGuide.quickbuyBuild.title',
    descriptionKey: 'home.storePicksGuide.quickbuyBuild.description',
    icon: 'lucide:settings-2',
    actionLabelKey: 'home.storePicksGuide.quickbuyBuild.action',
    actionIcon: 'lucide:zap',
    kind: 'quickbuy',
  },
]

const scrollToSection = (targetId: string) => {
  if (!import.meta.client) return

  const scroll = () => {
    const target = document.getElementById(targetId)
    if (!target) return false

    target.scrollIntoView({
      behavior: 'smooth',
      block: 'start',
    })
    window.history.replaceState(
      window.history.state,
      '',
      `${window.location.pathname}${window.location.search}#${targetId}`,
    )
    return true
  }

  if (scroll()) return
  window.requestAnimationFrame(() => {
    if (scroll()) return
    window.setTimeout(scroll, 250)
  })
}

const openQuickBuyEntry = () => {
  if (!import.meta.client) return
  window.dispatchEvent(new CustomEvent('quickbuy:open-entry'))
}

const handleCardAction = (card: HomeStorePicksGuideCard) => {
  if (card.kind === 'scroll') {
    scrollToSection(card.targetId)
    return
  }

  if (card.kind === 'quickbuy') {
    openQuickBuyEntry()
  }
}
</script>

<style scoped>
.home-store-picks-guide {
  padding: 0.25rem 0 2rem;
  scroll-margin-top: calc(var(--tz-site-header-spacer-height) + 1rem);
}

.home-store-picks-guide__grid {
  display: grid;
  gap: 0.75rem;
  grid-template-columns: repeat(5, minmax(0, 1fr));
}

.home-store-picks-guide__card {
  display: flex;
  min-width: 0;
  min-height: 9.5rem;
  flex-direction: column;
  justify-content: space-between;
  gap: 0.75rem;
  padding: 1rem;
  border-radius: 8px;
}

.home-store-picks-guide__content {
  display: flex;
  min-width: 0;
  flex-direction: column;
  gap: 0.625rem;
}

.home-store-picks-guide__title-row {
  display: flex;
  min-width: 0;
  align-items: flex-start;
  gap: 0.75rem;
}

.home-store-picks-guide__icon {
  display: grid;
  flex: 0 0 auto;
  width: 2.75rem;
  height: 1.5rem;
  place-items: center;
  border-radius: 8px;
  color: var(--tz-site-accent, #059669);
}

.home-store-picks-guide__card h3 {
  margin: 0;
  min-width: 0;
  color: var(--tz-text-primary);
  font-size: 1rem;
  font-weight: 800;
  line-height: 1.25;
}

.home-store-picks-guide__card p {
  margin: 0;
  color: var(--tz-text-secondary, rgba(203, 213, 225, 0.72));
  font-size: 0.875rem;
  line-height: 1.5;
}

.home-store-picks-guide__button {
  width: 100%;
  justify-content: center;
  gap: 0.5rem;
}

@media (max-width: 1180px) {
  .home-store-picks-guide__grid {
    grid-template-columns: repeat(3, minmax(0, 1fr));
  }
}

@media (max-width: 760px) {
  .home-store-picks-guide {
    padding-bottom: 1.5rem;
  }

  .home-store-picks-guide__grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .home-store-picks-guide__card {
    min-height: 9.25rem;
    padding: 0.875rem;
  }

  .home-store-picks-guide__card h3 {
    font-size: 0.92rem;
  }

  .home-store-picks-guide__card p {
    font-size: 0.78rem;
  }
}

@media (max-width: 420px) {
  .home-store-picks-guide__grid {
    grid-template-columns: minmax(0, 1fr);
  }
}
</style>
