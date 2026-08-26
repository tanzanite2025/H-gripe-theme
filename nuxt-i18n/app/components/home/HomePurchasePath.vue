<template>
  <section id="home-buying-path" class="bg-transparent py-8 tz-text-primary sm:py-10 lg:py-12">
      <div class="page-content-shell px-0 md:px-6">
        <div class="flex flex-col gap-4 sm:gap-5 lg:flex-row lg:items-end lg:justify-between">
          <div class="max-w-3xl">
            <span
              class="inline-flex items-center gap-2 rounded-full border border-[#059669]/30 bg-[#059669]/10 px-3 py-1 text-[11px] font-medium uppercase tracking-[0.16em] text-[#059669]"
            >
              <Icon name="lucide:route" class="h-3.5 w-3.5" aria-hidden="true" />
              {{ section.eyebrow }}
            </span>
            <h2 class="mt-3 text-2xl font-semibold leading-tight tz-text-primary sm:text-3xl">
              {{ section.title }}
            </h2>
          </div>
        </div>

        <div class="mt-5 grid gap-4 md:grid-cols-2 xl:grid-cols-4">
          <article
            v-for="card in section.cards"
            :key="card.id"
            class="group flex min-h-[218px] flex-col justify-between rounded-2xl premium-card p-5 transition-transform duration-200 hover:-translate-y-1"
          >
            <div>
              <div class="flex items-center gap-3">
                <div
                  class="grid h-11 w-11 shrink-0 place-items-center text-[#059669]"
                  aria-hidden="true"
                >
                  <Icon :name="card.icon" class="h-6 w-6" />
                </div>
                <h3 class="min-w-0 text-lg font-semibold leading-tight tz-text-primary">{{ card.title }}</h3>
              </div>
              <p class="mt-4 text-sm leading-relaxed tz-text-secondary">
                {{ card.description }}
              </p>
              <div
                v-if="card.highlights?.length"
                class="mt-4 flex flex-wrap gap-2"
                :aria-label="`${card.title} topics`"
              >
                <span
                  v-for="highlight in card.highlights"
                  :key="highlight"
                  class="rounded-full border border-slate-200 bg-slate-50 px-2.5 py-1 text-[11px] font-semibold uppercase leading-none tracking-[0.12em] tz-text-muted"
                >
                  {{ highlight }}
                </span>
              </div>
            </div>

            <div class="mt-6">
              <NuxtLink
                v-if="card.kind === 'route'"
                :to="localePath(card.route || '/')"
                class="premium-button w-full justify-center"
              >
                <Icon name="lucide:arrow-right" class="mr-2 h-4 w-4" aria-hidden="true" />
                {{ card.actionLabel }}
              </NuxtLink>

              <div
                v-else-if="card.kind === 'quickbuy-options'"
              >
                <ClientOnly>
                  <LazyHomePurchasePathQuickBuyActions :actions="card.quickBuyActions || []" />
                  <template #fallback>
                    <div class="grid gap-2">
                      <button
                        v-for="action in card.quickBuyActions || []"
                        :key="action.id"
                        type="button"
                        class="premium-button w-full justify-center"
                        disabled
                      >
                        <Icon :name="action.icon" class="mr-2 h-4 w-4" aria-hidden="true" />
                        {{ action.label }}
                      </button>
                    </div>
                  </template>
                </ClientOnly>
              </div>

              <button
                v-else
                type="button"
                class="premium-button w-full justify-center"
                @click="handleCardAction(card)"
              >
                <Icon name="lucide:message-circle" class="mr-2 h-4 w-4" aria-hidden="true" />
                {{ card.actionLabel }}
              </button>
            </div>
          </article>
        </div>
      </div>
  </section>

    <LazyQuickBuyContactServiceModal
      v-if="contactServiceOpen"
      @close="contactServiceOpen = false"
    />

</template>

<script setup lang="ts">
import { ref, useLocalePath } from '#imports'
import {
  homePurchasePathSection,
  type HomePurchasePathCard,
} from '~/utils/homePurchasePath'

const localePath = useLocalePath()
const contactServiceOpen = ref(false)
const section = homePurchasePathSection

const handleCardAction = (card: HomePurchasePathCard) => {
  if (card.kind === 'contact-service') {
    contactServiceOpen.value = true
  }
}
</script>

<style scoped>
#home-buying-path {
  scroll-margin-top: calc(var(--tz-site-header-spacer-height) + 1rem);
}
</style>
