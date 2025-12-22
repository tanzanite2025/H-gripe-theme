<template>
  <div>
    <h2 class="products-page__title">Wheelbuilding Guides</h2>
    <p class="products-page__intro">
      Explore tools, charts, and documentation to help you design and build wheelsets.
    </p>

    <section class="mt-4 grid gap-4 md:grid-cols-2 lg:gap-5">
      <NuxtLink
        v-for="guide in guides"
        :key="guide.to"
        :to="localePath(guide.to)"
        class="group flex items-start gap-4 rounded-2xl px-4 py-4 bg-[radial-gradient(circle_at_top_left,rgba(31,41,55,0.96),rgba(15,23,42,0.98))] shadow-[0_10px_26px_-14px_rgba(0,0,0,0.95)] transition-all duration-200 hover:bg-[radial-gradient(circle_at_top_left,rgba(31,41,55,0.98),rgba(15,23,42,0.99))] hover:shadow-[0_14px_32px_-16px_rgba(0,0,0,1)] hover:-translate-y-[1px] sm:px-5 sm:py-5"
      >
        <div
          class="mt-0.5 flex h-10 w-10 items-center justify-center rounded-full bg-gradient-to-br from-sky-400 to-emerald-400 text-xs font-semibold text-slate-950 shadow-[0_0_20px_rgba(56,189,248,0.35)]"
        >
          {{ guide.icon }}
        </div>
        <div class="flex-1">
          <h3 class="mb-1 text-sm font-semibold text-white sm:text-base">
            {{ guide.title }}
          </h3>
          <p class="text-xs text-slate-300 sm:text-sm">
            {{ guide.description }}
          </p>
          <span
            class="mt-2 inline-flex items-center text-[11px] font-medium text-sky-300 group-hover:text-sky-200"
          >
            {{ guide.cta }}
          </span>
        </div>
      </NuxtLink>
    </section>
  </div>
</template>

<script setup lang="ts">
import { useLocalePath } from '#imports'

const localePath = useLocalePath()

const guides = [
  {
    to: '/guides/tools',
    icon: 'T',
    title: 'Tools',
    description: 'Overview of tools and utilities that support your wheelbuilding workflow.',
    cta: 'Open tools guide',
  },
  {
    to: '/guides/tireguides',
    icon: 'C',
    title: 'Tire Guides',
    description:
      'Guides for tire sizing, pressure, inner tubes and tubeless setups, built on Schwalbe references.',
    cta: 'Open tire guides',
  },
  {
    to: '/guides/technical',
    icon: 'D',
    title: 'Technical',
    description: 'Technical notes, formulas, and reference docs for advanced wheelbuilding.',
    cta: 'Open technical docs',
  },
  {
    to: '/guides/wheelset-buyers',
    icon: 'G',
    title: 'Wheelset Buyers Guide',
    description: 'A high-level guide to choosing the right wheelset, to be expanded with steps.',
    cta: 'Read buyers guide',
  },
]

definePageMeta({
  layout: 'products',
})

useHead({
  title: 'Guides',
})
</script>

<style scoped>
.products-page__title {
  margin: 0 0 0.75rem;
  font-size: 1.5rem;
  font-weight: 600;
  color: #f9fafb;
}

.products-page__intro {
  margin: 0 0 0.75rem;
  font-size: 0.95rem;
  color: rgba(148, 163, 184, 0.9);
}
</style>