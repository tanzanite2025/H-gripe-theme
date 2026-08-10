<template>
  <div class="w-full">
    <h1 class="sr-only">Our Story</h1>

    <section id="story" class="w-full max-w-none">
      
      <!-- Premium Container -->
      <div class="ourstory-container relative rounded-2xl md:rounded-3xl px-4 md:px-10 pb-5 pt-4 md:pb-8 md:pt-6 overflow-hidden">
        
        <!-- Background Decor -->
        <div class="ourstory-container__glow ourstory-container__glow--top"></div>
        <div class="ourstory-container__glow ourstory-container__glow--bottom"></div>

        <div class="relative z-10">
          <!-- Header -->
          <div class="text-center mb-5">
            <h2 class="text-2xl md:text-3xl font-bold tz-text-primary mb-3">{{ t('company.ourStory.story.title') }}</h2>
            
            <NuxtLink
              class="ourstory-factory-link inline-flex items-center gap-2 px-6 py-2 rounded-full tz-text-primary text-sm font-medium transition-all group"
              :to="factoryTabTo"
            >
              <span>{{ t('company.ourStory.story.factoryButton') }}</span>
              <svg class="ourstory-factory-link__icon w-4 h-4 group-hover:translate-x-0.5 transition-transform" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" /></svg>
            </NuxtLink>
          </div>

          <!-- Hero Image -->
          <div class="mb-6 group">
            <div class="aspect-[21/9] w-full overflow-hidden rounded-xl border border-white/10 shadow-2xl relative">
              <div class="ourstory-image__scrim absolute inset-0 z-10"></div>
              <img
                class="h-full w-full object-cover transition-transform duration-700 group-hover:scale-105"
                src="/company/ourstory/ourstory/tanzanite-ourstory.webp"
                alt="Tanzanite Our Story"
                loading="lazy"
              />
            </div>
          </div>

          <!-- Story Text -->
          <div class="space-y-4 text-base tz-text-secondary leading-relaxed font-light tracking-wide">
            <template v-for="(paragraph, index) in storyParagraphs" :key="index">
              <p class="ourstory-lede" v-if="index === 0">
                {{ paragraph }}
              </p>
              <p v-else>
                {{ paragraph }}
              </p>
            </template>
          </div>
        </div>

      </div>
    </section>

    <!-- Feedback Section -->
    <div class="w-full max-w-none px-0 mt-4 mb-6">
      <UserFeedbackThread
        threadKey="company-ourstory"
        title="Share your feedback about Our Story"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, watch } from 'vue'
import { useHead, definePageMeta, navigateTo, useI18n, useLocalePath, useRoute } from '#imports'
import UserFeedbackThread from '~/components/UserFeedbackThread.vue'
import {
  companyAboutTabs,
  isPageSubNavigationTabId,
  type CompanyAboutTabId,
} from '~/utils/pageSubNavigation'

const { t } = useI18n()
const localePath = useLocalePath()
const route = useRoute()

const storyParagraphs = computed(() => {
  const body = t('company.ourStory.story.body')
  return body.split('\n\n').filter(Boolean)
})

const factoryTabTo = computed(() => {
  return `${localePath('/company/about')}#factory`
})

const getCompanyAboutTabFromHash = (hash: string): CompanyAboutTabId | null => {
  const raw = String(hash || '').replace(/^#/, '')
  return isPageSubNavigationTabId(companyAboutTabs, raw) ? raw : null
}

watch(
  () => route.hash,
  (hash) => {
    const tab = getCompanyAboutTabFromHash(hash)
    if (!tab) return

    navigateTo(`${localePath('/company/about')}#${tab}`, { replace: true })
  },
  { immediate: true }
)

definePageMeta({
  layout: 'products',
})

useHead({
  title: 'Our Story',
})
</script>

<style scoped>
.ourstory-container {
  border: 1px solid rgba(255, 255, 255, 0.1);
  background:
    radial-gradient(circle at 86% 0%, rgba(181, 255, 109, 0.08), transparent 34rem),
    linear-gradient(180deg, rgba(255, 255, 255, 0.035), rgba(255, 255, 255, 0.012)),
    var(--tz-card-surface);
  box-shadow: 0 24px 60px rgba(0, 0, 0, 0.46);
}

.ourstory-container__glow {
  position: absolute;
  border-radius: 9999px;
  pointer-events: none;
  filter: blur(88px);
}

.ourstory-container__glow--top {
  top: 0;
  right: 0;
  width: 24rem;
  height: 24rem;
  background: rgba(181, 255, 109, 0.07);
  transform: translate(50%, -50%);
}

.ourstory-container__glow--bottom {
  bottom: 0;
  left: 0;
  width: 16rem;
  height: 16rem;
  background: rgba(255, 255, 255, 0.035);
  transform: translate(-50%, 50%);
}

.ourstory-factory-link {
  border: 1px solid rgba(255, 255, 255, 0.12);
  background: rgba(17, 17, 22, 0.86);
  box-shadow: 0 10px 28px rgba(0, 0, 0, 0.36);
}

.ourstory-factory-link:hover {
  border-color: rgba(181, 255, 109, 0.36);
  background: rgba(26, 26, 31, 0.92);
}

.ourstory-factory-link__icon,
.ourstory-lede::first-letter {
  color: var(--tz-text-accent);
}

.ourstory-image__scrim {
  background: linear-gradient(to top, rgba(0, 0, 0, 0.42), transparent);
}

.ourstory-lede::first-letter {
  float: left;
  margin-top: -6px;
  margin-right: 0.75rem;
  font-family: ui-serif, Georgia, Cambria, 'Times New Roman', Times, serif;
  font-size: 3rem;
  font-weight: 700;
  line-height: 1;
}
</style>
