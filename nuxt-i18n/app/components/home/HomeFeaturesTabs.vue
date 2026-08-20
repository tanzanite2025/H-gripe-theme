<template>
  <section id="home-why-choose-us" class="home-features-tabs bg-transparent text-white pt-6 pb-2">
    <div class="page-content-shell px-0 md:px-4">
      <section
        class="home-features-tabs__group"
        aria-labelledby="why-us-features-title"
      >
        <h2 id="why-us-features-title" class="home-features-tabs__group-title">
          {{ $t('home.whyChooseUs.title') }}
        </h2>

        <div class="home-features-tabs__grid home-features-tabs__grid--why_us">
          <article
            v-for="(item, index) in whyUsCards"
            :key="item.titleKey"
            class="home-features-tabs__card"
          >
            <div class="home-features-tabs__card-heading">
              <span class="home-features-tabs__card-index">{{ String(index + 1).padStart(2, '0') }}</span>
              <h3>{{ $t(item.titleKey) }}</h3>
            </div>

            <p v-if="hasContent(item.descriptionKey)" class="home-features-tabs__description">
              {{ $t(item.descriptionKey) }}
            </p>

            <ul v-if="visibleBullets(item).length" class="home-features-tabs__bullets">
              <li v-for="bulletKey in visibleBullets(item)" :key="bulletKey">
                <span aria-hidden="true">•</span>
                <span>{{ $t(bulletKey) }}</span>
              </li>
            </ul>
          </article>
        </div>
      </section>
    </div>
  </section>
</template>

<script setup lang="ts">
import { useI18n } from '#imports'

const { t } = useI18n()

type FeatureItem = {
  titleKey: string
  descriptionKey: string
  bullets: string[]
}

const whyUsCards: FeatureItem[] = Array.from({ length: 6 }, (_, index) => ({
  titleKey: `home.whyChooseUs.items.${index}.title`,
  descriptionKey: `home.whyChooseUs.items.${index}.description`,
  bullets: [0, 1, 2].map((bulletIndex) => `home.whyChooseUs.items.${index}.bullets.${bulletIndex}`),
}))

const hasContent = (key: string) => {
  return !/placeholder/i.test(String(t(key)))
}

const visibleBullets = (item: { bullets?: string[] }) => {
  return (item.bullets || []).filter((key) => hasContent(key))
}

</script>

<style scoped>
#home-why-choose-us {
  scroll-margin-top: calc(var(--tz-site-header-spacer-height) + 1rem);
}

.home-features-tabs__group-title {
  margin: 0 0 18px;
  color: #f8fafc;
  font-size: 24px;
  font-weight: 800;
  line-height: 1.2;
  text-align: center;
}

.home-features-tabs__grid {
  display: grid;
  gap: 14px;
  align-items: start;
}

.home-features-tabs__grid--why_us {
  grid-template-columns: repeat(3, minmax(0, 1fr));
}

.home-features-tabs__card {
  display: flex;
  min-width: 0;
  min-height: 0;
  flex-direction: column;
  gap: 12px;
  padding: 18px;
  border: 1px solid rgba(255, 255, 255, 0.08);
  border-radius: 16px;
  background: var(--tz-card-surface, #111116);
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.05),
    0 16px 36px -28px rgba(0, 0, 0, 0.96);
}

.home-features-tabs__card-heading {
  display: flex;
  min-width: 0;
  align-items: flex-start;
  gap: 10px;
}

.home-features-tabs__card-index {
  flex: 0 0 auto;
  padding-top: 2px;
  color: var(--tz-brand-primary, #b5ff6d);
  font-size: 11px;
  font-weight: 800;
  letter-spacing: 0.12em;
}

.home-features-tabs__card-heading h3 {
  min-width: 0;
  margin: 0;
  color: #f8fafc;
  font-size: 16px;
  font-weight: 800;
  line-height: 1.25;
}

.home-features-tabs__description {
  margin: 0;
  color: var(--tz-text-secondary, rgba(203, 213, 225, 0.72));
  font-size: 14px;
  line-height: 1.45;
}

.home-features-tabs__bullets {
  display: grid;
  gap: 7px;
  margin: 0;
  padding: 0;
  color: var(--tz-text-secondary, rgba(203, 213, 225, 0.72));
  font-size: 13px;
  line-height: 1.35;
  list-style: none;
}

.home-features-tabs__bullets li {
  display: grid;
  grid-template-columns: auto minmax(0, 1fr);
  gap: 8px;
}

.home-features-tabs__bullets li > span:first-child {
  color: var(--tz-brand-primary, #b5ff6d);
}

@media (max-width: 900px) {
  .home-features-tabs__grid--why_us {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (max-width: 640px) {
  .home-features-tabs__group-title {
    margin-bottom: 14px;
    font-size: 20px;
  }

  .home-features-tabs__grid {
    gap: 10px;
  }

  .home-features-tabs__grid--why_us {
    grid-template-columns: minmax(0, 1fr);
  }

  .home-features-tabs__card {
    gap: 9px;
    padding: 13px;
    border-radius: 14px;
  }

  .home-features-tabs__card-heading {
    gap: 7px;
  }

  .home-features-tabs__card-heading h3 {
    font-size: 14px;
  }

  .home-features-tabs__description,
  .home-features-tabs__bullets {
    font-size: 12px;
  }

}
</style>
