<template>
  <div class="company-page">
    <h2 class="company-page__title">Our Story</h2>
    <p class="company-page__intro">
      A brief overview of how Tanzanite grew from a wheelbuilding idea into a complete
      tools-and-support ecosystem for builders.
    </p>

    <section class="company-section company-section--overview">
      <h3 class="company-section__title">From first wheel to full platform</h3>
      <p class="company-section__body">
        Use this space to describe where the brand started, the type of riders or
        customers you focus on, and why wheels and tools became the core of the
        business.
      </p>
      <p class="company-section__body">
        You can later replace this text with concrete milestones, locations, or
        founders' motivations so visitors quickly understand what makes Tanzanite
        different from another generic wheel brand.
      </p>
    </section>

    <section class="company-section company-section--values">
      <h3 class="company-section__title">What we believe in</h3>
      <div class="company-values">
        <div class="company-values__item">
          <h4 class="company-values__title">Precision for every build</h4>
          <p class="company-values__body">
            Describe your view on accuracy, repeatability, and why calculators,
            charts, and documentation matter so much for wheelbuilding.
          </p>
        </div>
        <div class="company-values__item">
          <h4 class="company-values__title">Accessible tools</h4>
          <p class="company-values__body">
            Explain how you want tools, guides, and support to be usable for both
            first-time builders and experienced wheel builders.
          </p>
        </div>
        <div class="company-values__item">
          <h4 class="company-values__title">Support that feels close</h4>
          <p class="company-values__body">
            Connect this idea to the Support and Guides areas: how people can get
            answers, feedback, or quick help when a build or calculation goes wrong.
          </p>
        </div>
      </div>
    </section>

    <section class="company-section company-section--timeline">
      <h3 class="company-section__title">Milestones</h3>
      <ul class="company-timeline">
        <li class="company-timeline__item">
          <div class="company-timeline__year">20XX</div>
          <div class="company-timeline__content">
            First prototype tools or internal calculators. Replace this with your
            real starting point.
          </div>
        </li>
        <li class="company-timeline__item">
          <div class="company-timeline__year">20XX</div>
          <div class="company-timeline__content">
            Launch of the first public spoke calculator or wheelbuilding resource
            under the Tanzanite name.
          </div>
        </li>
        <li class="company-timeline__item">
          <div class="company-timeline__year">20XX</div>
          <div class="company-timeline__content">
            Expansion into documentation, guides, or new product lines. Add any
            other turning points you want visitors to remember.
          </div>
        </li>
      </ul>
    </section>

    <section class="company-section company-section--cta">
      <h3 class="company-section__title">What comes next</h3>
      <p class="company-section__body">
        Use this closing block to talk about where you want Tanzanite to go next:
        new tools, deeper integrations, or better support experiences for
        builders.
      </p>
    </section>
  </div>
</template>

<script setup lang="ts">
definePageMeta({
  layout: 'products',
})

useHead({
  title: 'Our Story',
})
</script>

<style scoped>
.company-page {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

.company-page__title {
  margin: 0 0 0.75rem;
  font-size: 1.5rem;
  font-weight: 600;
  color: #f9fafb;
}

.company-page__intro {
  margin: 0 0 0.75rem;
  font-size: 0.95rem;
  color: rgba(148, 163, 184, 0.9);
}

.company-section {
  padding-top: 0.5rem;
}

.company-section__title {
  margin: 0 0 0.5rem;
  font-size: 1.1rem;
  font-weight: 600;
  color: #e5e7eb;
}

.company-section__body {
  margin: 0 0 0.5rem;
  font-size: 0.95rem;
  color: rgba(148, 163, 184, 0.95);
}

.company-section--values {
  border-top: 1px solid rgba(148, 163, 184, 0.25);
  padding-top: 1.25rem;
}

.company-section--timeline {
  border-top: 1px solid rgba(148, 163, 184, 0.25);
  padding-top: 1.25rem;
}

.company-section--cta {
  border-top: 1px solid rgba(148, 163, 184, 0.25);
  padding-top: 1.25rem;
}

.company-values {
  display: grid;
  grid-template-columns: 1fr;
  gap: 0.75rem;
}

.company-values__item {
  padding: 0.75rem 0.9rem;
  border-radius: 0.75rem;
  background: rgba(15, 23, 42, 0.8);
  border: 1px solid rgba(148, 163, 184, 0.24);
}

.company-values__title {
  margin: 0 0 0.35rem;
  font-size: 0.95rem;
  font-weight: 600;
  color: #f9fafb;
}

.company-values__body {
  margin: 0;
  font-size: 0.9rem;
  color: rgba(148, 163, 184, 0.95);
}

.company-timeline {
  list-style: none;
  padding: 0;
  margin: 0;
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.company-timeline__item {
  display: grid;
  grid-template-columns: auto 1fr;
  column-gap: 0.75rem;
  row-gap: 0.25rem;
  align-items: flex-start;
}

.company-timeline__year {
  font-size: 0.85rem;
  font-weight: 600;
  color: #e5e7eb;
  min-width: 3.5rem;
}

.company-timeline__content {
  font-size: 0.9rem;
  color: rgba(148, 163, 184, 0.95);
}

@media (min-width: 768px) {
  .company-page__title {
    font-size: 1.75rem;
  }

  .company-values {
    grid-template-columns: repeat(3, minmax(0, 1fr));
  }
}
</style>
