<template>
  <section class="home-hero bg-transparent text-white">
    <div class="page-content-shell px-0 md:px-4 pb-6 pt-1 sm:pt-2 lg:pb-8 lg:pt-3">
      <div class="flex flex-col items-center text-center">
        <div class="home-hero__content mt-3 grid w-full gap-7 sm:mt-5 lg:mt-2 lg:grid-cols-[minmax(0,3fr)_minmax(0,7fr)] lg:items-center lg:gap-12">
          <div class="home-hero__copy flex flex-col items-center px-2 text-center lg:h-full lg:justify-center lg:items-start lg:px-0 lg:text-left">
            <div class="mb-4 flex items-center justify-center gap-2.5 text-base font-medium leading-tight text-white sm:gap-3 sm:text-lg lg:justify-start">
              <span class="welcome-hand-wave inline-flex shrink-0" aria-hidden="true">
                <Icon
                  name="lucide:hand"
                  class="h-7 w-7 text-[#B5FF6D] sm:h-8 sm:w-8"
                />
              </span>
              <span>{{ welcomeText }}</span>
            </div>
            <h1 class="text-xl font-semibold leading-[1.05] tracking-tight sm:text-3xl lg:text-4xl">
              <template v-if="heroTitle.accent">
                {{ heroTitle.before }}<span class="text-[#B5FF6D]">{{ heroTitle.accent }}</span>{{ heroTitle.after }}
              </template>
              <template v-else>
                {{ heroTitle.full }}
              </template>
            </h1>
            <p class="mt-3 max-w-2xl text-sm leading-relaxed tz-text-secondary sm:text-lg">
              {{ t('home.hero.subtitle') }}
            </p>

            <div class="mt-5 flex w-full flex-row items-center justify-center gap-2 sm:w-auto sm:gap-3 lg:justify-start">
              <button
                type="button"
                class="premium-button premium-button--active home-hero__cta inline-flex min-w-0 flex-1 items-center justify-center sm:w-auto sm:flex-none"
                @click="scrollToSection('home-buying-path')"
              >
                <Icon name="lucide:route" class="mr-2 h-4 w-4" aria-hidden="true" />
                Find my wheelset
              </button>
              <button
                type="button"
                class="premium-button home-hero__cta inline-flex min-w-0 flex-1 items-center justify-center sm:w-auto sm:flex-none"
                @click="scrollToSection('featured-products')"
              >
                <Icon name="lucide:shopping-bag" class="mr-2 h-4 w-4" aria-hidden="true" />
                Shop best sellers
              </button>
            </div>
          </div>

          <div class="home-hero__media relative hidden w-full lg:block lg:justify-self-stretch">
            <HomeHeroVisualShowcaseDesktopCollage
              :items="homeHeroVisualShowcaseItems"
              :ariaLabel="t('home.hero.stackAriaLabel')"
            />
            <StorefrontDataNotice
              v-if="heroVisualNotice"
              class="home-hero__visual-notice"
              :tone="heroVisualNotice.tone"
              :title="heroVisualNotice.title"
              :description="heroVisualNotice.description"
              :role="heroVisualNotice.role"
              compact
            >
              <template v-if="showHeroVisualRetry" #actions>
                <button
                  type="button"
                  class="storefront-data-notice-action"
                  :disabled="homeHeroVisualShowcasePending"
                  @click="retryHomeHeroVisualShowcase"
                >
                  <Icon name="lucide:refresh-cw" aria-hidden="true" />
                  {{ t('common.retry') }}
                </button>
              </template>
            </StorefrontDataNotice>
          </div>

          <div class="home-hero__mobile-media relative mx-auto w-full lg:hidden">
            <HomeHeroVisualShowcaseMobilePairGallery
              :items="homeHeroVisualShowcaseItems"
              :ariaLabel="t('home.hero.stackAriaLabel')"
            />
            <StorefrontDataNotice
              v-if="heroVisualNotice"
              class="home-hero__visual-notice"
              :tone="heroVisualNotice.tone"
              :title="heroVisualNotice.title"
              :description="heroVisualNotice.description"
              :role="heroVisualNotice.role"
              compact
            >
              <template v-if="showHeroVisualRetry" #actions>
                <button
                  type="button"
                  class="storefront-data-notice-action"
                  :disabled="homeHeroVisualShowcasePending"
                  @click="retryHomeHeroVisualShowcase"
                >
                  <Icon name="lucide:refresh-cw" aria-hidden="true" />
                  {{ t('common.retry') }}
                </button>
              </template>
            </StorefrontDataNotice>
          </div>
        </div>

      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed, useI18n } from '#imports'
import StorefrontDataNotice from '~/components/StorefrontDataNotice.vue'
import HomeHeroVisualShowcaseDesktopCollage from '~/components/home/HomeHeroVisualShowcaseDesktopCollage.vue'
import HomeHeroVisualShowcaseMobilePairGallery from '~/components/home/HomeHeroVisualShowcaseMobilePairGallery.vue'
import { useHomeHeroVisualShowcase } from '~/composables/useHomeHeroVisualShowcase'

const { t } = useI18n()
const {
  homeHeroVisualShowcaseItems,
  homeHeroVisualShowcaseSource,
  homeHeroVisualShowcasePending,
  refreshHomeHeroVisualShowcase,
} = await useHomeHeroVisualShowcase()

const welcomeText = computed(() => {
  const translated = t('home.hero.welcome')
  return translated === 'home.hero.welcome' ? 'WELCOME!' : translated
})

const heroTitle = computed(() => {
  const full = t('home.hero.title')
  const accent = 'Factory-Direct'
  const accentIndex = full.indexOf(accent)

  if (accentIndex === -1) {
    return { full, before: '', accent: '', after: '' }
  }

  return {
    full,
    before: full.slice(0, accentIndex),
    accent,
    after: full.slice(accentIndex + accent.length),
  }
})

const scrollToSection = (targetId: string) => {
  if (!import.meta.client) return

  document.getElementById(targetId)?.scrollIntoView({
    behavior: 'smooth',
    block: 'start',
  })
}

const heroVisualNotice = computed(() => {
  switch (homeHeroVisualShowcaseSource.value) {
    case 'loading':
      return {
        tone: 'empty' as const,
        role: 'status' as const,
        title: t('storefrontDataNotice.hero.loading.title'),
        description: t('storefrontDataNotice.hero.loading.description'),
      }
    case 'error':
      return {
        tone: 'error' as const,
        role: 'alert' as const,
        title: t('storefrontDataNotice.hero.error.title'),
        description: t('storefrontDataNotice.hero.error.description'),
      }
    case 'locale-fallback':
      return {
        tone: 'fallback' as const,
        role: 'status' as const,
        title: t('storefrontDataNotice.hero.localeFallback.title'),
        description: t('storefrontDataNotice.hero.localeFallback.description'),
      }
    case 'built-in-fallback':
      return {
        tone: 'fallback' as const,
        role: 'status' as const,
        title: t('storefrontDataNotice.hero.builtInFallback.title'),
        description: t('storefrontDataNotice.hero.builtInFallback.description'),
      }
    default:
      return null
  }
})

const showHeroVisualRetry = computed(() => homeHeroVisualShowcaseSource.value === 'error')

const retryHomeHeroVisualShowcase = () => {
  void refreshHomeHeroVisualShowcase()
}
</script>

<style scoped>
.home-hero__visual-notice {
  margin-top: 0.7rem;
}

.home-hero__cta {
  padding-inline: 0.85rem;
  white-space: nowrap;
}

@media (max-width: 380px) {
  .home-hero__cta {
    font-size: 0.72rem;
    padding-inline: 0.65rem;
  }

  .home-hero__cta :deep(svg) {
    width: 0.9rem;
    height: 0.9rem;
    margin-right: 0.35rem;
  }
}

@media (min-width: 640px) {
  .home-hero__cta {
    padding-inline: 1.25rem;
  }
}

@keyframes welcome-hand-wave {
  0%,
  72%,
  100% {
    transform: rotate(0deg);
  }

  76% {
    transform: rotate(12deg);
  }

  80% {
    transform: rotate(-8deg);
  }

  84% {
    transform: rotate(10deg);
  }

  88% {
    transform: rotate(-4deg);
  }

  92% {
    transform: rotate(0deg);
  }
}

.welcome-hand-wave {
  transform-origin: 70% 90%;
  animation: welcome-hand-wave 4.5s ease-in-out infinite;
}

@media (prefers-reduced-motion: reduce) {
  .welcome-hand-wave {
    animation: none;
  }
}
</style>
