<template>
  <div class="website-hub">
    <p class="website-hub__eyebrow">{{ copy.eyebrow }}</p>
    <h1>{{ copy.title }}</h1>
    <p class="website-hub__lead">{{ copy.lead }}</p>

    <nav class="website-hub__links" :aria-label="copy.navigationLabel">
      <NuxtLink
        v-for="entry in entries"
        :key="entry.to"
        class="website-hub__link"
        :to="localePath(entry.to)"
      >
        <span>
          <strong>{{ entry.title }}</strong>
          <small>{{ entry.description }}</small>
        </span>
        <Icon name="lucide:arrow-up-right" aria-hidden="true" />
      </NuxtLink>
    </nav>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { definePageMeta, useHead, useI18n, useLocalePath } from '#imports'
import { isSimplifiedChineseStorefrontLocale } from '~/utils/storefrontLocales'

const { locale } = useI18n()
const localePath = useLocalePath()

const copy = computed(() => {
  if (isSimplifiedChineseStorefrontLocale(locale.value)) {
    return {
      eyebrow: 'ABOUT THIS WEBSITE',
      title: '关于这个网站',
      lead: '这里是关于网站本身的独立入口。它和“关于我们”属于不同的内容关系，下面的页面会分别说明网站背后的人，以及这个名字的由来。',
      navigationLabel: '网站页面',
      entries: [
        {
          to: '/website/me-this-website',
          title: 'Me & This Website',
          description: '网站管理者与这个网站的关系',
        },
        {
          to: '/website/why-this-name',
          title: 'Why This Name',
          description: '为什么使用这个名字',
        },
      ],
    }
  }

  return {
    eyebrow: 'ABOUT THIS WEBSITE',
    title: 'About This Website',
    lead: 'This is the independent entry point for information about the website itself. It is separate from “About Us”, with pages for the person behind the site and the reason for its name.',
    navigationLabel: 'Website pages',
    entries: [
      {
        to: '/website/me-this-website',
        title: 'Me & This Website',
        description: 'The relationship between the site manager and this website',
      },
      {
        to: '/website/why-this-name',
        title: 'Why This Name',
        description: 'Why this website uses its name',
      },
    ],
  }
})

const entries = computed(() => copy.value.entries)

definePageMeta({
  layout: 'products',
  footer: false,
  footerLabelFallback: 'Website',
})

useHead(() => ({
  title: copy.value.title,
}))
</script>

<style scoped>
.website-hub {
  width: min(100%, 76rem);
  margin-inline: auto;
  padding: clamp(1.25rem, 4vw, 3.5rem) 0 clamp(2.5rem, 7vw, 5rem);
  color: var(--tz-text-primary);
}

.website-hub__eyebrow {
  margin: 0;
  color: var(--tz-text-accent);
  font-size: 0.66rem;
  font-weight: 850;
  letter-spacing: 0.18em;
  line-height: 1.4;
  text-transform: uppercase;
}

.website-hub h1 {
  max-width: 48rem;
  margin: 0.65rem 0 0;
  color: var(--tz-text-primary);
  font-size: clamp(2.5rem, 6vw, 5rem);
  font-weight: 800;
  letter-spacing: 0;
  line-height: 0.98;
}

.website-hub__lead {
  max-width: 48rem;
  margin: 1.25rem 0 0;
  color: var(--tz-text-secondary);
  font-size: 1.1rem;
  line-height: 1.8;
}

.website-hub__links {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 1rem;
  margin-top: clamp(2.5rem, 6vw, 5rem);
}

.website-hub__link {
  display: flex;
  min-width: 0;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  border-top: 1px solid var(--tz-border-subtle);
  border-bottom: 1px solid var(--tz-border-subtle);
  padding: 1.15rem 0;
  color: var(--tz-text-primary);
  text-decoration: none;
  transition:
    border-color 180ms ease,
    color 180ms ease,
    transform 180ms ease;
}

.website-hub__link:hover,
.website-hub__link:focus-visible {
  border-color: rgba(5, 150, 105, 0.62);
  color: var(--tz-text-accent);
  transform: translateY(-2px);
}

.website-hub__link:focus-visible {
  outline: 1px solid rgba(5, 150, 105, 0.72);
  outline-offset: 0.2rem;
}

.website-hub__link span {
  display: grid;
  min-width: 0;
  gap: 0.4rem;
}

.website-hub__link strong {
  color: inherit;
  font-size: 1rem;
  font-weight: 800;
  line-height: 1.3;
}

.website-hub__link small {
  color: var(--tz-text-secondary);
  font-size: 0.82rem;
  line-height: 1.5;
}

.website-hub__link :deep(svg) {
  width: 1rem;
  height: 1rem;
  flex: 0 0 auto;
}

@media (max-width: 767px) {
  .website-hub {
    padding-top: 0.5rem;
  }

  .website-hub h1 {
    font-size: 2.35rem;
  }

  .website-hub__lead {
    font-size: 1rem;
  }

  .website-hub__links {
    display: block;
  }

  .website-hub__link + .website-hub__link {
    margin-top: 1rem;
  }
}
</style>
