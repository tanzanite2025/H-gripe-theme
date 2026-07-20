<template>
  <div class="w-full">
    <h1 class="sr-only">Our Story</h1>

    <section id="story" class="w-full max-w-none mt-4 md:mt-8">
      
      <!-- Premium Container -->
      <div class="relative bg-slate-900/50 backdrop-blur-sm border border-white/10 rounded-2xl md:rounded-3xl px-4 md:px-10 pb-5 pt-4 md:pb-8 md:pt-6 shadow-2xl overflow-hidden">
        
        <!-- Background Decor -->
        <div class="absolute top-0 right-0 w-96 h-96 bg-blue-500/10 rounded-full blur-[100px] pointer-events-none -translate-y-1/2 translate-x-1/2"></div>
        <div class="absolute bottom-0 left-0 w-64 h-64 bg-teal-500/5 rounded-full blur-[80px] pointer-events-none translate-y-1/2 -translate-x-1/2"></div>

        <div class="relative z-10">
          <!-- Header -->
          <div class="text-center mb-5">
            <h2 class="text-2xl md:text-3xl font-bold text-white mb-3">{{ t('company.ourStory.story.title') }}</h2>
            
            <NuxtLink
              class="inline-flex items-center gap-2 px-6 py-2 rounded-full bg-slate-800/80 hover:bg-slate-700 text-white text-sm font-medium transition-all shadow-lg border border-white/10 hover:border-teal-500/30 group"
              :to="factoryTabTo"
            >
              <span>{{ t('company.ourStory.story.factoryButton') }}</span>
              <svg class="w-4 h-4 text-teal-400 group-hover:translate-x-0.5 transition-transform" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" /></svg>
            </NuxtLink>
          </div>

          <!-- Hero Image -->
          <div class="mb-6 group">
            <div class="aspect-[21/9] w-full overflow-hidden rounded-xl border border-white/10 shadow-2xl relative">
              <div class="absolute inset-0 bg-gradient-to-t from-slate-950/40 to-transparent z-10"></div>
              <img
                class="h-full w-full object-cover transition-transform duration-700 group-hover:scale-105"
                src="/company/ourstory/ourstory/tanzanite-ourstory.webp"
                alt="Tanzanite Our Story"
                loading="lazy"
              />
            </div>
          </div>

          <!-- Story Text -->
          <div class="space-y-4 text-base text-slate-300 leading-relaxed font-light tracking-wide">
            <template v-for="(paragraph, index) in storyParagraphs" :key="index">
              <p class="first-letter:text-5xl first-letter:font-serif first-letter:text-teal-500 first-letter:float-left first-letter:mr-3 first-letter:mt-[-6px]" v-if="index === 0">
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

    <!-- FAQ Section -->
    <div class="w-full max-w-none mt-4 px-0">
      <PageFaq pageId="company-ourstory" />
    </div>

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
import { computed } from 'vue'
import { useHead, definePageMeta, useI18n, useLocalePath } from '#imports'
import UserFeedbackThread from '~/components/UserFeedbackThread.vue'
import PageFaq from '~/components/PageFaq.vue'

const { t } = useI18n()
const localePath = useLocalePath()

const storyParagraphs = computed(() => {
  const body = t('company.ourStory.story.body')
  return body.split('\n\n').filter(Boolean)
})

const factoryTabTo = computed(() => {
  return `${localePath('/company/about')}#factory`
})

definePageMeta({
  layout: 'products',
})

useHead({
  title: 'Our Story',
})
</script>

<style scoped>
/* Scoped styles removed in favor of Tailwind utility classes for consistency and maintainability */
</style>
