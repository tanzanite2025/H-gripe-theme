<template>
  <div class="website-page">
    <section class="website-identity" aria-labelledby="website-page-title">
      <div class="website-identity__copy">
        <p class="website-eyebrow">{{ copy.eyebrow }}</p>
        <h1 id="website-page-title" class="website-title">{{ copy.title }}</h1>
        <p class="website-lead">{{ copy.lead }}</p>

        <div class="website-identity__scope">
          <span class="website-identity__scope-dot" aria-hidden="true"></span>
          <span>{{ copy.scope }}</span>
        </div>
      </div>

      <div class="website-profile" :aria-label="copy.profileLabel">
        <div class="website-profile__avatar" role="img" :aria-label="copy.avatarLabel">
          <StorefrontImage v-if="copy.avatarUrl" :src="copy.avatarUrl" :alt="copy.avatarLabel" preset="avatar" />
          <span v-else>{{ copy.avatarMark }}</span>
        </div>
        <div class="website-profile__meta">
          <span class="website-profile__label">{{ copy.profileLabel }}</span>
          <strong>{{ copy.profileRole }}</strong>
          <span>{{ copy.profileContext }}</span>
        </div>
      </div>
    </section>

    <section class="website-statement" aria-labelledby="website-statement-title">
      <div class="website-section-heading">
        <p class="website-eyebrow">{{ copy.statementEyebrow }}</p>
        <h2 id="website-statement-title">{{ copy.statementTitle }}</h2>
      </div>

      <div class="website-statement__body">
        <p v-for="paragraph in copy.statementParagraphs" :key="paragraph">
          {{ paragraph }}
        </p>
      </div>
    </section>

    <section class="website-factory" aria-labelledby="website-factory-title">
      <figure class="website-factory__visual">
        <StorefrontImage
          :src="copy.factoryImageUrl"
          :alt="copy.factoryImageAlt"
          preset="content"
        />
        <figcaption>{{ copy.factoryImageCaption }}</figcaption>
      </figure>

      <div class="website-factory__copy">
        <p class="website-eyebrow">{{ copy.factoryEyebrow }}</p>
        <h2 id="website-factory-title">{{ copy.factoryTitle }}</h2>
        <p>{{ copy.factoryBody }}</p>

        <NuxtLink class="website-factory__link" :to="factoryLink">
          <span>{{ copy.factoryCta }}</span>
          <Icon name="lucide:arrow-up-right" aria-hidden="true" />
        </NuxtLink>
      </div>
    </section>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { definePageMeta, useHead, useI18n, useLocalePath } from '#imports'
import { useWebsiteProfileSettings } from '~/composables/useWebsiteProfileSettings'

const { locale } = useI18n()
const localePath = useLocalePath()
const { websiteProfileSettings: copy } = useWebsiteProfileSettings(locale)
const factoryLink = computed(() =>
  localePath(copy.value.factoryLink || '/company/about/factory')
)

definePageMeta({
  layout: 'products',
})

useHead(() => ({
  title: copy.value.title,
}))
</script>

<style scoped>
.website-page {
  width: 100%;
  color: var(--tz-text-primary);
}

.website-identity,
.website-statement,
.website-factory {
  width: min(100%, 76rem);
  margin-inline: auto;
}

.website-identity {
  display: grid;
  grid-template-columns: minmax(0, 1.25fr) minmax(14rem, 0.75fr);
  align-items: center;
  gap: 3rem;
  padding: clamp(1.25rem, 4vw, 3.5rem) 0 clamp(2.25rem, 6vw, 5rem);
  border-bottom: 1px solid rgba(255, 255, 255, 0.12);
}

.website-identity__copy {
  max-width: 48rem;
}

.website-eyebrow {
  margin: 0;
  color: var(--tz-text-accent);
  font-size: 0.66rem;
  font-weight: 850;
  letter-spacing: 0.18em;
  line-height: 1.4;
  text-transform: uppercase;
}

.website-title {
  margin: 0.65rem 0 0;
  color: var(--tz-text-primary);
  font-size: 5rem;
  font-weight: 800;
  letter-spacing: 0;
  line-height: 0.98;
}

.website-lead {
  max-width: 44rem;
  margin: 1.25rem 0 0;
  color: var(--tz-text-secondary);
  font-size: 1.25rem;
  line-height: 1.7;
}

.website-identity__scope {
  display: inline-flex;
  align-items: center;
  gap: 0.55rem;
  margin-top: 1.5rem;
  color: var(--tz-text-muted);
  font-size: 0.74rem;
  font-weight: 700;
  line-height: 1.4;
}

.website-identity__scope-dot {
  width: 0.55rem;
  height: 0.55rem;
  flex: 0 0 auto;
  border-radius: 9999px;
  background: var(--tz-text-accent);
  box-shadow: 0 0 0 5px rgba(181, 255, 109, 0.1);
}

.website-profile {
  display: flex;
  align-items: center;
  gap: 1rem;
  justify-self: end;
  width: min(100%, 19rem);
  padding: 1rem 0 1rem 1.25rem;
  border-left: 1px solid rgba(255, 255, 255, 0.16);
}

.website-profile__avatar {
  display: inline-flex;
  width: 5.25rem;
  height: 5.25rem;
  flex: 0 0 auto;
  align-items: center;
  justify-content: center;
  border: 1px solid rgba(181, 255, 109, 0.52);
  border-radius: 50%;
  background: #111116;
  color: var(--tz-text-accent);
  box-shadow:
    0 0 0 7px rgba(181, 255, 109, 0.05),
    0 18px 40px rgba(0, 0, 0, 0.35);
  font-size: 1.15rem;
  font-weight: 850;
  letter-spacing: 0.04em;
}

.website-profile__meta {
  display: grid;
  min-width: 0;
  gap: 0.18rem;
}

.website-profile__label {
  color: var(--tz-text-muted);
  font-size: 0.62rem;
  font-weight: 800;
  letter-spacing: 0.14em;
  line-height: 1.3;
  text-transform: uppercase;
}

.website-profile__meta strong {
  color: var(--tz-text-primary);
  font-size: 1rem;
  font-weight: 750;
  line-height: 1.25;
}

.website-profile__meta > span:last-child {
  color: var(--tz-text-secondary);
  font-size: 0.78rem;
  line-height: 1.45;
}

.website-statement {
  display: grid;
  grid-template-columns: minmax(11rem, 0.4fr) minmax(0, 1fr);
  gap: 3rem;
  padding: clamp(2.5rem, 7vw, 6rem) 0;
}

.website-section-heading h2,
.website-factory__copy h2 {
  margin: 0.5rem 0 0;
  color: var(--tz-text-primary);
  font-size: 2.35rem;
  font-weight: 750;
  letter-spacing: 0;
  line-height: 1.1;
}

.website-statement__body {
  max-width: 48rem;
}

.website-statement__body p {
  margin: 0;
  color: var(--tz-text-secondary);
  font-size: 1.16rem;
  line-height: 1.85;
}

.website-statement__body p + p {
  margin-top: 1.35rem;
}

.website-factory {
  display: grid;
  grid-template-columns: minmax(0, 1fr) minmax(0, 0.85fr);
  align-items: center;
  gap: clamp(1.5rem, 5vw, 4rem);
  padding: clamp(1.25rem, 4vw, 2.5rem) 0 clamp(2.5rem, 7vw, 5rem);
  border-top: 1px solid rgba(255, 255, 255, 0.12);
}

.website-factory__visual {
  margin: 0;
}

.website-factory__visual img {
  display: block;
  aspect-ratio: 16 / 10;
  width: 100%;
  border: 1px solid rgba(255, 255, 255, 0.12);
  border-radius: 1rem;
  object-fit: cover;
  box-shadow: 0 24px 56px rgba(0, 0, 0, 0.36);
}

.website-factory__visual figcaption {
  margin-top: 0.65rem;
  color: var(--tz-text-muted);
  font-size: 0.72rem;
  line-height: 1.5;
}

.website-factory__copy {
  max-width: 34rem;
}

.website-factory__copy > p:not(.website-eyebrow) {
  margin: 1rem 0 0;
  color: var(--tz-text-secondary);
  font-size: 0.98rem;
  line-height: 1.75;
}

.website-factory__link {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
  margin-top: 1.35rem;
  border: 1px solid rgba(181, 255, 109, 0.28);
  border-radius: 9999px;
  background: rgba(181, 255, 109, 0.08);
  padding: 0.68rem 1rem;
  color: var(--tz-text-primary);
  font-size: 0.78rem;
  font-weight: 750;
  line-height: 1.2;
  text-decoration: none;
  transition:
    background-color 180ms ease,
    border-color 180ms ease,
    color 180ms ease,
    transform 180ms ease;
}

.website-factory__link:hover,
.website-factory__link:focus-visible {
  border-color: rgba(181, 255, 109, 0.68);
  background: rgba(181, 255, 109, 0.16);
  color: #ffffff;
  transform: translateY(-1px);
}

.website-factory__link:focus-visible {
  outline: 1px solid rgba(181, 255, 109, 0.72);
  outline-offset: 0.18rem;
}

.website-factory__link :deep(svg) {
  width: 0.95rem;
  height: 0.95rem;
  color: var(--tz-text-accent);
}

@media (max-width: 767px) {
  .website-identity,
  .website-statement,
  .website-factory {
    display: block;
  }

  .website-identity {
    padding-top: 0.5rem;
  }

  .website-title {
    font-size: 2.25rem;
  }

  .website-lead {
    font-size: 1rem;
  }

  .website-profile {
    width: 100%;
    margin-top: 2rem;
    padding: 1rem 0 0;
    border-top: 1px solid rgba(255, 255, 255, 0.12);
    border-left: 0;
  }

  .website-statement {
    padding: 2.75rem 0;
  }

  .website-section-heading {
    margin-bottom: 1.5rem;
  }

  .website-section-heading h2,
  .website-factory__copy h2 {
    font-size: 1.45rem;
  }

  .website-statement__body p {
    font-size: 1rem;
  }

  .website-factory {
    padding-top: 1.25rem;
  }

  .website-factory__copy {
    margin-top: 1.5rem;
  }
}
</style>
