<template>
  <section class="flex flex-col gap-3 h-full px-1.5 md:px-6 py-2 md:py-3 max-w-none md:max-w-[720px] mx-auto w-full">
    <!-- Stepper header -->
    <div class="flex items-center gap-2 text-[10px] md:text-xs uppercase tracking-[0.2em] tz-text-secondary">
      <StepperDot :active="currentStep === 1">1</StepperDot>
      <div class="hidden md:block" :class="currentStep === 1 ? 'tz-text-primary' : 'tz-text-secondary'">{{ t('checkout.stepper.steps.payment') }}</div>
      <div class="flex-1 h-px bg-white/10"></div>
      <StepperDot :active="currentStep === 2">2</StepperDot>
      <div class="hidden md:block" :class="currentStep === 2 ? 'tz-text-primary' : 'tz-text-secondary'">{{ t('checkout.stepper.steps.shipping') }}</div>
      <div class="flex-1 h-px bg-white/10"></div>
      <StepperDot :active="currentStep === 3">3</StepperDot>
      <div class="hidden md:block" :class="currentStep === 3 ? 'tz-text-primary' : 'tz-text-secondary'">{{ t('checkout.stepper.steps.review') }}</div>
    </div>

    <!-- Step 1 -->
    <div v-if="currentStep === 1" class="flex flex-col gap-3 flex-1">
      <header class="flex items-center justify-between">
        <div class="text-xs font-semibold tz-text-primary flex items-center gap-2">
          {{ t('checkout.stepper.payment.pickProvider') }}
          <span class="px-2 py-0.5 rounded-full text-[10px] uppercase tracking-[0.2em] bg-white/10 tz-text-secondary">{{ t('checkout.stepper.payment.stepCount', { current: 1, total: 3 }) }}</span>
        </div>
        <span class="text-[11px] tz-text-secondary">{{ t('checkout.stepper.payment.tapCardHint') }}</span>
      </header>

      <div class="space-y-2 flex-1 overflow-y-auto pr-1 md:pr-2 max-w-[620px] mx-auto w-full">
        <article
          v-for="option in visiblePaymentOptions"
          :key="option.id"
          class="rounded-2xl bg-[radial-gradient(circle_at_top_left,rgba(31,41,55,0.96),rgba(15,23,42,0.98))] shadow-[6px_10px_30px_rgba(0,0,0,0.55)] overflow-hidden transition-colors"
        >
          <button
            type="button"
            class="w-full px-4 py-3 flex items-center justify-between gap-3 text-left"
            :class="activeMethod === option.id ? 'tz-text-primary' : 'tz-text-secondary'"
            @click="handleSelect(option.id)"
          >
            <div class="flex flex-col gap-1">
              <div class="flex items-center gap-2">
                <span class="text-sm font-semibold">{{ option.title }}</span>
                <div v-if="optionIcons[option.id]?.length" class="flex items-center gap-1">
                  <img
                    v-for="icon in optionIcons[option.id]"
                    :key="icon"
                    :src="icon"
                    :alt="`${option.title} ${t('checkout.stepper.payment.iconAltSuffix')}`"
                    class="h-4 w-6 object-contain"
                    loading="lazy"
                    decoding="async"
                  />
                </div>
              </div>
              <span class="text-[11px] tz-text-secondary">{{ option.subtitle }}</span>
            </div>
            <div
              class="inline-flex items-center justify-center rounded-full px-2 py-0.5 text-[10px] font-semibold uppercase tracking-[0.2em] shadow-[3px_3px_12px_rgba(0,0,0,0.35)]"
              :class="activeMethod === option.id ? 'bg-white text-slate-900' : 'bg-white/10 tz-text-secondary'"
            >
              {{ activeMethod === option.id ? t('checkout.stepper.payment.selected') : t('checkout.stepper.payment.tapToView') }}
            </div>
          </button>

          <div
            v-if="activeMethod === option.id"
            class="px-4 pb-4 space-y-3 border-t border-white/5 text-[12px] tz-text-secondary"
          >
            <p class="leading-relaxed">
              {{ option.description }}
            </p>
            <ul v-if="option.points?.length" class="space-y-1 tz-text-secondary">
              <li v-for="point in option.points" :key="point" class="flex items-start gap-2 text-[11px]">
                <span class="inline-flex h-1.5 w-1.5 rounded-full bg-emerald-300 translate-y-1"></span>
                <span>{{ point }}</span>
              </li>
            </ul>
            <button
              type="button"
              class="w-full inline-flex items-center justify-center px-4 py-2 rounded-xl text-xs font-semibold text-slate-900 bg-white hover:brightness-95 transition"
              @click="handleContinueFromStepOne"
            >
              {{ t('checkout.stepper.payment.continueWithMethod') }} →
            </button>
          </div>
        </article>
      </div>
    </div>

    <!-- Step 2 -->
    <div v-else-if="currentStep === 2" class="flex flex-col gap-3 flex-1">
      <button
        type="button"
        class="inline-flex items-center gap-2 text-[11px] font-semibold uppercase tracking-[0.2em] tz-text-secondary hover:text-white transition"
        @click="handleBackToMethods"
      >
        <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 19l-7-7 7-7" />
        </svg>
        {{ t('checkout.stepper.shipping.changePaymentMethod') }}
      </button>

      <div class="space-y-3 flex-1 overflow-y-auto pr-1 md:pr-2 max-w-[620px] mx-auto w-full">
        <section
          v-if="showShippingForm"
          class="rounded-2xl bg-[radial-gradient(circle_at_top_left,rgba(31,41,55,0.96),rgba(15,23,42,0.98))] px-4 py-4 shadow-[6px_10px_30px_rgba(0,0,0,0.55)] space-y-3"
        >
          <header class="text-sm font-semibold text-white flex items-center gap-2">
            <svg class="w-4 h-4 text-blue-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17.657 16.657L13.414 20.9a2 2 0 01-2.828 0l-4.243-4.243a8 8 0 1111.314 0z" />
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 11a3 3 0 11-6 0 3 3 0 016 0z" />
            </svg>
            {{ t('checkout.stepper.shipping.addressTitle') }}
          </header>
          <p class="text-xs tz-text-secondary">
            {{ t('checkout.stepper.shipping.addressHelp') }}
          </p>

          <div class="space-y-3">
            <div class="grid grid-cols-1 gap-2.5 min-[420px]:grid-cols-2">
              <div class="min-[420px]:col-span-2 space-y-1.5">
                <label class="block text-sm font-medium tz-text-secondary">{{ t('checkout.stepper.shipping.countryRegion') }} <span class="text-red-400">*</span></label>
                <input
                  :value="countrySearchValue"
                  type="text"
                  :placeholder="t('checkout.stepper.shipping.searchCountryPlaceholder')"
                  class="w-full px-3 py-1.5 rounded-lg border-none text-xs text-white placeholder:text-white/40 bg-[linear-gradient(135deg,rgba(15,23,42,0.98),rgba(15,23,42,0.96))] shadow-[0_2px_6px_-3px_rgba(0,0,0,0.9),0_0_6px_rgba(15,23,42,0.7)] focus:outline-none focus:[box-shadow:0_0_0_1px_rgba(56,189,248,0.85)]"
                  @input="handleCountrySearchInput"
                />
                <select
                  :value="shippingForm?.country || ''"
                  class="w-full px-4 py-1.5 rounded-lg border-none text-white bg-slate-900/80 shadow-[0_2px_6px_-3px_rgba(0,0,0,0.9),0_0_6px_rgba(15,23,42,0.7)] focus:outline-none focus:[box-shadow:0_0_0_1px_rgba(56,189,248,0.85)] [color-scheme:dark]"
                  :class="{ 'focus:[box-shadow:0_0_0_1px_rgba(248,113,113,0.9)]': shippingForm?.country && shippingValidation && !shippingValidation.isShippable }"
                  @change="handleShippingInput('country', $event)"
                >
                  <option value="" disabled>{{ t('checkout.stepper.shipping.selectCountry') }}</option>
                  <optgroup :label="t('checkout.stepper.shipping.availableForShipping')">
                    <option
                      v-for="country in shippableCountries"
                      :key="country.code"
                      :value="country.code"
                    >
                      {{ country.name }}
                    </option>
                  </optgroup>
                  <optgroup :label="t('checkout.stepper.shipping.otherCountries')">
                    <option
                      v-for="country in nonShippableCountries"
                      :key="country.code"
                      :value="country.code"
                      class="tz-text-muted"
                    >
                      {{ country.name }}
                    </option>
                  </optgroup>
                </select>
              </div>

              <div>
                <label class="block text-sm font-medium tz-text-secondary mb-1">{{ t('checkout.stepper.shipping.recipient') }} <span class="text-red-400">*</span></label>
                <input
                  :value="shippingForm?.name || ''"
                  type="text"
                  :placeholder="t('checkout.stepper.shipping.recipientPlaceholder')"
                  class="w-full px-4 py-1.5 rounded-lg border-none text-white placeholder:text-white/40 bg-[linear-gradient(135deg,rgba(15,23,42,0.98),rgba(15,23,42,0.96))] shadow-[0_2px_6px_-3px_rgba(0,0,0,0.9),0_0_6px_rgba(15,23,42,0.7)] focus:outline-none focus:[box-shadow:0_0_0_1px_rgba(56,189,248,0.85)]"
                  @input="handleShippingInput('name', $event)"
                />
              </div>

              <div>
                <label class="block text-sm font-medium tz-text-secondary mb-1">{{ t('checkout.stepper.shipping.phone') }} <span class="text-red-400">*</span></label>
                <input
                  :value="shippingForm?.phone || ''"
                  type="tel"
                  :placeholder="t('checkout.stepper.shipping.phonePlaceholder')"
                  class="w-full px-4 py-1.5 rounded-lg border-none text-white placeholder:text-white/40 bg-[linear-gradient(135deg,rgba(15,23,42,0.98),rgba(15,23,42,0.96))] shadow-[0_2px_6px_-3px_rgba(0,0,0,0.9),0_0_6px_rgba(15,23,42,0.7)] focus:outline-none focus:[box-shadow:0_0_0_1px_rgba(56,189,248,0.85)]"
                  @input="handleShippingInput('phone', $event)"
                />
              </div>

              <div class="min-[420px]:col-span-2">
                <label class="block text-sm font-medium tz-text-secondary mb-1">{{ t('checkout.stepper.shipping.address') }} <span class="text-red-400">*</span></label>
                <textarea
                  :value="shippingForm?.address || ''"
                  rows="2"
                  :placeholder="t('checkout.stepper.shipping.addressPlaceholder')"
                  class="w-full px-4 py-1.5 rounded-lg border-none text-white placeholder:text-white/40 bg-[linear-gradient(135deg,rgba(15,23,42,0.98),rgba(15,23,42,0.96))] shadow-[0_2px_6px_-3px_rgba(0,0,0,0.9),0_0_6px_rgba(15,23,42,0.7)] focus:outline-none focus:[box-shadow:0_0_0_1px_rgba(56,189,248,0.85)] resize-none"
                  @input="handleShippingInput('address', $event)"
                ></textarea>
              </div>
            </div>

            <div
              v-if="shippingForm?.country && shippingValidation && !shippingValidation.isShippable"
              class="px-4 py-2 bg-red-500/10 border border-red-500/30 rounded-lg space-y-2"
            >
              <p class="text-red-300 text-sm font-semibold">{{ shippingValidation.reason || t('checkout.stepper.shipping.unavailableFallback') }}</p>
              <p class="tz-text-secondary text-xs">{{ t('checkout.stepper.shipping.optionsIntro') }}</p>
              <div class="space-y-2 text-xs">
                <button type="button" class="flex items-center gap-2 text-blue-400 hover:text-blue-300" @click="handleOpenContact">
                  <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z" />
                  </svg>
                  {{ t('checkout.stepper.shipping.contactSpecialShipping') }}
                </button>
                <button type="button" class="flex items-center gap-2 tz-text-secondary hover:text-white" @click="handleOpenFreight">
                  <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7h12m0 0l-4-4m4 4l-4 4m0 6H4m0 0l4 4m-4-4l4-4" />
                  </svg>
                  {{ t('checkout.stepper.shipping.useFreightForwarder') }}
                </button>
                <button type="button" class="flex items-center gap-2 tz-text-secondary hover:text-white" @click="handleSaveCart">
                  <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 5a2 2 0 012-2h10a2 2 0 012 2v16l-7-3.5L5 21V5z" />
                  </svg>
                  {{ t('checkout.stepper.shipping.saveCartLater') }}
                </button>
              </div>
            </div>

            <div
              v-if="shippingForm?.country && shippingValidation?.isShippable && estimatedDelivery"
              class="px-3 py-2 bg-green-500/10 border border-green-500/30 rounded-lg text-sm text-green-300 flex items-center gap-2"
            >
              <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
              </svg>
              {{ t('checkout.stepper.shipping.estimatedDelivery', { estimatedDelivery }) }}
            </div>

            <div class="grid grid-cols-2 gap-2.5">
              <div>
                <label class="block text-sm font-medium tz-text-secondary mb-1">{{ t('checkout.stepper.shipping.city') }} <span class="text-red-400">*</span></label>
                <input
                  :value="shippingForm?.city || ''"
                  type="text"
                  :placeholder="t('checkout.stepper.shipping.cityPlaceholder')"
                  class="w-full px-4 py-1.5 rounded-lg border-none text-white placeholder:text-white/40 bg-[linear-gradient(135deg,rgba(15,23,42,0.98),rgba(15,23,42,0.96))] shadow-[0_2px_6px_-3px_rgba(0,0,0,0.9),0_0_6px_rgba(15,23,42,0.7)] focus:outline-none focus:[box-shadow:0_0_0_1px_rgba(56,189,248,0.85)]"
                  @input="handleShippingInput('city', $event)"
                />
              </div>
              <div>
                <label class="block text-sm font-medium tz-text-secondary mb-1">{{ t('checkout.stepper.shipping.zipCode') }}</label>
                <input
                  :value="shippingForm?.zip || ''"
                  type="text"
                  :placeholder="zipPlaceholder"
                  class="w-full px-4 py-1.5 rounded-lg border-none text-white placeholder:text-white/40 bg-[linear-gradient(135deg,rgba(15,23,42,0.98),rgba(15,23,42,0.96))] shadow-[0_2px_6px_-3px_rgba(0,0,0,0.9),0_0_6px_rgba(15,23,42,0.7)] focus:outline-none focus:[box-shadow:0_0_0_1px_rgba(56,189,248,0.85)]"
                  @input="handleShippingInput('zip', $event)"
                />
                <p v-if="zipHint" class="text-xs tz-text-muted mt-1">{{ zipHint }}</p>
              </div>
            </div>

            <div class="flex justify-center md:justify-end pt-2">
              <button
                type="button"
                class="px-4 py-2 rounded-xl text-sm font-semibold transition w-full max-w-[280px] md:w-auto md:max-w-none"
                :class="[
                  canProceedToReview ? 'bg-white text-slate-900 hover:brightness-95' : 'bg-white/10 tz-text-disabled cursor-not-allowed'
                ]"
                :disabled="!canProceedToReview"
                @click="handleContinueToReview"
              >
                {{ canProceedToReview ? t('checkout.stepper.shipping.continueToReview') : t('checkout.stepper.shipping.completeShippingFirst') }}
              </button>
            </div>
          </div>
        </section>
        <div v-else class="rounded-2xl border border-dashed border-white/20 px-4 py-6 text-center text-sm tz-text-secondary bg-white/5">
          {{ t('checkout.stepper.shipping.notRequiredForMethod') }}
          <div class="mt-4 flex justify-center md:justify-end">
            <button
              type="button"
              class="px-4 py-2 rounded-xl text-sm font-semibold transition w-full max-w-[280px] md:w-auto md:max-w-none"
              :class="[
                canProceedToReview ? 'bg-white text-slate-900 hover:brightness-95' : 'bg-white/10 tz-text-disabled cursor-not-allowed'
              ]"
              :disabled="!canProceedToReview"
              @click="handleContinueToReview"
            >
              {{ canProceedToReview ? t('checkout.stepper.shipping.continueToReview') : t('checkout.stepper.shipping.completeShippingFirst') }}
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- Step 3 -->
    <div v-else-if="currentStep === 3" class="flex flex-col gap-3 flex-1">
      <button
        type="button"
        class="inline-flex items-center gap-2 text-[11px] font-semibold uppercase tracking-[0.2em] tz-text-secondary hover:text-white transition"
        @click="handleBackToShipping"
      >
        <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 19l-7-7 7-7" />
        </svg>
        {{ t('checkout.stepper.review.backToShipping') }}
      </button>

      <div class="space-y-3 flex-1 overflow-y-auto pr-1 md:pr-2">
        <div class="grid gap-3 md:grid-cols-3">
          <section class="rounded-2xl bg-[radial-gradient(circle_at_top_left,rgba(31,41,55,0.96),rgba(15,23,42,0.98))] px-4 py-4 shadow-[6px_10px_30px_rgba(0,0,0,0.55)] space-y-3">
            <header class="text-sm font-semibold text-white flex items-center gap-2">
              {{ t('checkout.stepper.review.coupon') }}
              <span class="text-[11px] tz-text-muted">{{ t('checkout.stepper.common.optional') }}</span>
            </header>
            <div class="flex flex-col sm:flex-row sm:flex-wrap gap-2">
              <input
                :value="couponInput"
                type="text"
                :placeholder="t('checkout.stepper.review.couponPlaceholder')"
                class="flex-1 px-3 py-2 rounded-xl bg-white/5 text-sm text-white placeholder:text-white/40 focus:outline-none focus:ring-2 focus:ring-white/30 shadow-[0_2px_6px_-3px_rgba(0,0,0,0.9),0_0_6px_rgba(15,23,42,0.7)]"
                @input="handleCouponInput"
              />
              <button
                type="button"
                class="px-3 py-2 rounded-xl text-sm font-semibold bg-white text-slate-900 hover:brightness-95 transition disabled:opacity-50 disabled:cursor-not-allowed w-full sm:w-auto sm:flex-none"
                :disabled="!couponInput || isApplyingCoupon"
                @click="handleApplyCouponClick"
              >
                {{ isApplyingCoupon ? t('checkout.stepper.review.applyingCoupon') : t('checkout.stepper.review.applyCoupon') }}
              </button>
            </div>
            <p v-if="appliedCouponDisplay" class="text-xs text-emerald-300 flex items-center gap-1">
              <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
              </svg>
              {{ t('checkout.stepper.review.couponApplied', { coupon: appliedCouponDisplay }) }}
            </p>
          </section>

          <section
            class="rounded-2xl bg-[radial-gradient(circle_at_top_left,rgba(31,41,55,0.96),rgba(15,23,42,0.98))] px-4 py-4 shadow-[6px_10px_30px_rgba(0,0,0,0.55)] space-y-3"
          >
            <header class="text-sm font-semibold text-white flex items-center gap-2">
              {{ t('checkout.stepper.review.usePointsDiscount') }}
              <span class="text-[11px] tz-text-muted">
                {{ pointsAvailable > 0 ? t('checkout.stepper.review.pointsAvailable', { points: pointsAvailable }) : t('checkout.stepper.review.noPoints') }}
              </span>
            </header>
            <div class="flex items-center gap-2.5 mb-1 text-sm tz-text-secondary">
              <input
                type="checkbox"
                class="w-4 h-4 rounded border-white/30 bg-white/10 text-[#6b73ff]"
                :checked="isUsingPoints"
                @change="handlePointsToggle"
              />
              <span>{{ pointsAvailable > 0 ? t('checkout.stepper.review.usePoints') : t('checkout.stepper.review.pointsUnavailable') }}</span>
            </div>
            <div v-if="isUsingPoints" class="space-y-2">
              <label class="block text-xs tz-text-secondary">{{ t('checkout.stepper.review.pointsToUse') }}</label>
              <input
                :value="pointsToUse"
                type="number"
                min="0"
                :max="maxPointsToUse"
                class="w-full px-3 py-2 rounded-xl bg-white/5 text-sm text-white border border-white/10 focus:outline-none focus:ring-2 focus:ring-white/30"
                @input="handlePointsInputChange"
              />
              <p class="text-xs tz-text-secondary">{{ pointsHint }}</p>
            </div>
            <p v-if="pointsAvailable <= 0" class="text-xs tz-text-muted">
              {{ t('checkout.stepper.review.earnPointsHint') }}
            </p>
          </section>

          <section class="rounded-2xl border border-white/10 bg-[radial-gradient(circle_at_top_left,rgba(31,41,55,0.96),rgba(15,23,42,0.98))] px-4 py-6 text-center text-sm tz-text-secondary shadow-[6px_10px_30px_rgba(0,0,0,0.55)]">
            <header class="text-sm font-semibold text-white flex items-center gap-2">
              <svg class="w-4 h-4 text-purple-300" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 10h.01M12 10h.01M16 10h.01M21 16c0 3.866-4.03 7-9 7s-9-3.134-9-7V8a4 4 0 014-4h10a4 4 0 014 4v8z" />
              </svg>
            {{ t('checkout.stepper.review.orderNotes') }} <span class="text-[11px] tz-text-muted">({{ t('checkout.stepper.common.optional') }})</span>
            </header>
            <textarea
              :value="shippingForm?.notes || ''"
              rows="3"
              :placeholder="t('checkout.stepper.review.orderNotesPlaceholder')"
              class="w-full px-4 py-3 rounded-xl bg-white/5 text-sm text-white placeholder:text-white/40 focus:outline-none focus:ring-2 focus:ring-white/30 resize-none min-h-[96px] shadow-[0_2px_6px_-3px_rgba(0,0,0,0.9),0_0_6px_rgba(15,23,42,0.7)]"
              @input="handleShippingInput('notes', $event)"
            ></textarea>
          </section>
        </div>

        <section class="rounded-2xl border border-white/10 bg-[radial-gradient(circle_at_top_left,rgba(31,41,55,0.96),rgba(15,23,42,0.98))] px-4 py-4 shadow-[0_15px_44px_-30px_rgba(2,6,23,0.9)] space-y-3">
          <header class="text-sm font-semibold text-white flex items-center gap-2">
            <svg class="w-4 h-4 text-emerald-300" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
            </svg>
            {{ t('checkout.stepper.review.orderSummary') }}
          </header>

          <div v-if="orderSummary?.items?.length" class="space-y-3 max-h-48 overflow-y-auto pr-1">
            <article
              v-for="item in orderSummary.items"
              :key="item.id ?? item.title"
              class="flex gap-3 p-3 rounded-xl bg-[radial-gradient(circle_at_top_left,rgba(15,23,42,0.98),rgba(2,6,23,0.95))] border border-white/5"
            >
              <div class="w-14 h-14 rounded-lg overflow-hidden bg-white/5 flex-shrink-0">
                <img
                  v-if="item.thumbnail"
                  :src="item.thumbnail"
                  :alt="item.title"
                  class="w-full h-full object-cover"
                />
              </div>
              <div class="flex-1 min-w-0">
                <p class="text-sm font-medium text-white truncate">{{ item.title }}</p>
                <p class="text-xs tz-text-secondary mt-0.5">{{ t('checkout.stepper.summary.qty') }}: {{ item.quantity }}</p>
                <p class="text-sm font-semibold text-white mt-1">{{ formatPrice(item.price * item.quantity) }}</p>
              </div>
            </article>
          </div>
          <div v-else class="text-xs tz-text-muted text-center py-2 border border-dashed border-white/10 rounded-xl">
            {{ t('checkout.stepper.summary.empty') }}
          </div>

          <div class="space-y-2 text-sm tz-text-secondary pt-2 border-t border-white/10">
            <div class="flex justify-between">
              <span>{{ t('checkout.stepper.summary.subtotal') }}</span>
              <span class="font-medium text-white">{{ formatPrice(orderSummary?.totals?.subtotal ?? 0) }}</span>
            </div>
            <div class="flex justify-between items-start gap-4">
              <span>
                {{ t('checkout.stepper.summary.shipping') }}
                <span v-if="orderSummary?.totals?.shippingLabel" class="text-xs tz-text-muted">
                  ({{ orderSummary.totals.shippingLabel }})
                </span>
              </span>
              <span
                class="font-medium text-white text-right"
                :class="orderSummary?.totals?.shippingState === 'unavailable' ? 'text-red-400 text-xs' : ''"
              >
                {{ shippingLabel }}
              </span>
            </div>
            <div class="flex justify-between">
              <span>{{ t('checkout.stepper.summary.tax') }}</span>
              <span class="font-medium text-white">{{ formatPrice(orderSummary?.totals?.tax ?? 0) }}</span>
            </div>
            <div class="flex justify-between" :class="orderSummary?.totals?.pointsDiscount ? 'text-emerald-300' : 'tz-text-secondary'">
              <span>{{ t('checkout.stepper.summary.pointsDiscount') }}</span>
              <span>- {{ formatPrice(orderSummary?.totals?.pointsDiscount ?? 0) }}</span>
            </div>
            <div class="flex justify-between" :class="orderSummary?.totals?.couponDiscount ? 'text-emerald-300' : 'tz-text-secondary'">
              <span>{{ t('checkout.stepper.summary.couponDiscount') }}</span>
              <span>- {{ formatPrice(orderSummary?.totals?.couponDiscount ?? 0) }}</span>
            </div>
            <div class="flex justify-between tz-text-secondary">
              <span>{{ t('checkout.stepper.summary.giftCardDiscount') }}</span>
              <span>- {{ formatPrice(giftCardDiscount) }}</span>
            </div>
            <div class="flex justify-between text-lg font-bold pt-3 border-t border-white/20">
              <span>{{ t('checkout.stepper.summary.total') }}</span>
              <span class="text-[#4efce7]">{{ formatPrice(orderSummary?.totals?.total ?? 0) }}</span>
            </div>
          </div>
        </section>

        <div class="rounded-2xl bg-[radial-gradient(circle_at_bottom,rgba(15,23,42,0.98),rgba(2,6,23,0.95))] px-4 py-4 space-y-2 hidden md:block shadow-[6px_10px_30px_rgba(0,0,0,0.55)]">
          <ChatStartButton
            class="w-full text-sm"
            :label="desktopCtaLabel"
            :disabled="isSubmitting"
            @click="handleSubmitMock"
          />
          <p v-if="ctaDescription" class="text-[11px] tz-text-secondary text-center">
            {{ ctaDescription }}
          </p>
        </div>
      </div>

      <div class="rounded-2xl bg-[radial-gradient(circle_at_bottom,rgba(15,23,42,0.98),rgba(2,6,23,0.95))] px-4 py-4 space-y-2 md:hidden shadow-[6px_10px_30px_rgba(0,0,0,0.55)]">
        <div class="text-center">
          <p class="text-xs font-semibold text-white">{{ mobilePaymentTitle }}</p>
          <p v-if="mobilePaymentDescription" class="text-[11px] tz-text-secondary mt-1">
            {{ mobilePaymentDescription }}
          </p>
        </div>
        <ChatStartButton
          class="w-full justify-center text-xs"
          size="md"
          :label="desktopCtaLabel"
          :disabled="isSubmitting"
          @click="handleSubmitMock"
        />
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useI18n } from '#imports'
import StepperDot from './StepperDot.vue'
import ChatStartButton from '~/components/ChatStartButton.vue'

type Step = 1 | 2 | 3

interface PaymentOption {
  id: string
  title: string
  subtitle: string
  description: string
  points?: string[]
}

interface ShippingForm {
  country: string
  name: string
  phone: string
  address: string
  city: string
  zip: string
  notes?: string
}

interface OrderSummaryItem {
  id?: string | number
  title: string
  quantity: number
  price: number
  thumbnail?: string | null
}

interface OrderSummaryTotals {
  subtotal: number
  shipping: number | null
  shippingLabel?: string
  shippingState?: 'available' | 'unavailable' | 'checking' | 'select'
  tax: number
  pointsDiscount?: number
  couponDiscount?: number
  giftCardDiscount?: number
  total: number
}

const props = defineProps<{
  initialStep?: Step
  initialMethod?: string
  paymentOptions?: PaymentOption[]
  couponInput?: string
  isApplyingCoupon?: boolean
  appliedCoupon?: { code?: string; label?: string } | string | null
  pointsAvailable?: number
  isUsingPoints?: boolean
  pointsToUse?: number
  maxPointsToUse?: number
  pointsHint?: string
  orderSummary?: {
    items: OrderSummaryItem[]
    totals: OrderSummaryTotals
  }
  currency?: string
  showShippingForm?: boolean
  shippingForm?: ShippingForm
  countrySearch?: string
  shippableCountries?: Array<{ code: string; name: string }>
  nonShippableCountries?: Array<{ code: string; name: string }>
  shippingValidation?: {
    isShippable: boolean
    reason?: string
    matchedRule?: { service_label?: string; free_over?: number }
  } | null
  estimatedDelivery?: string | null
  zipPlaceholder?: string
  zipHint?: string
  desktopCtaLabel?: string
  ctaDescription?: string
  mobilePaymentTitle?: string
  mobilePaymentDescription?: string
  isSubmitting?: boolean
}>()

const emit = defineEmits<{
  (e: 'update:step', value: Step): void
  (e: 'update:method', value: string): void
  (e: 'submit'): void
  (e: 'coupon-input', value: string): void
  (e: 'apply-coupon'): void
  (e: 'toggle-points', value: boolean): void
  (e: 'points-input', value: number): void
  (e: 'update-shipping-field', payload: { field: keyof ShippingForm; value: string }): void
  (e: 'country-search', value: string): void
  (e: 'open-contact'): void
  (e: 'open-freight'): void
  (e: 'save-cart'): void
}>()

const { t, locale } = useI18n()

const currentStep = ref<Step>(props.initialStep ?? 1)
const activeMethod = ref(props.initialMethod ?? 'card')

const fallbackPaymentOptions = computed<PaymentOption[]>(() => {
  const priceText = formatPrice(orderSummary.value?.totals?.total ?? 0)
  return [
    {
      id: 'card',
      title: t('checkout.payment.card.title'),
      subtitle: `${priceText} · ${t('checkout.payment.card.subtitle')}`,
      description: t('checkout.payment.card.stepperDescription'),
      points: [
        t('checkout.payment.card.points.shipping'),
        t('checkout.payment.card.points.immediate'),
      ],
    },
    {
      id: 'alipay',
      title: t('checkout.payment.alipay.optionTitle'),
      subtitle: `${priceText} · ${t('checkout.payment.alipay.subtitle')}`,
      description: t('checkout.payment.alipay.stepperDescription'),
      points: [
        t('checkout.payment.alipay.points.recipient'),
        t('checkout.payment.alipay.points.wallets'),
      ],
    },
    {
      id: 'bank',
      title: t('checkout.payment.bank.title'),
      subtitle: `${priceText} · ${t('checkout.payment.bank.subtitle')}`,
      description: t('checkout.payment.bank.stepperDescription'),
      points: [
        t('checkout.payment.bank.points.reference'),
      ],
    },
    {
      id: 'paypal',
      title: t('checkout.payment.paypal.optionTitle'),
      subtitle: `${priceText} · ${t('checkout.payment.paypal.subtitle')}`,
      description: t('checkout.payment.paypal.stepperDescription'),
      points: [
        t('checkout.payment.paypal.points.country'),
      ],
    },
    {
      id: 'stripe',
      title: t('checkout.payment.stripe.optionTitle'),
      subtitle: `${priceText} · ${t('checkout.payment.stripe.subtitle')}`,
      description: t('checkout.payment.stripe.stepperDescription'),
      points: [
        t('checkout.payment.stripe.points.sca'),
      ],
    },
    {
      id: 'worldfirst',
      title: t('checkout.payment.worldfirst.optionTitle'),
      subtitle: `${priceText} · ${t('checkout.payment.worldfirst.subtitle')}`,
      description: t('checkout.payment.worldfirst.stepperDescription'),
      points: [
        t('checkout.payment.worldfirst.points.b2b'),
      ],
    },
  ]
})

const visiblePaymentOptions = computed(() =>
  props.paymentOptions?.length ? props.paymentOptions : fallbackPaymentOptions.value,
)

const optionIcons: Record<string, string[]> = {
  card: [
    '/checkoutstepper/step1/credit-debit-cards/visa.svg',
    '/checkoutstepper/step1/credit-debit-cards/mastercard.svg',
    '/checkoutstepper/step1/credit-debit-cards/amex.svg',
    '/checkoutstepper/step1/credit-debit-cards/applepay.svg',
    '/checkoutstepper/step1/credit-debit-cards/diners.svg',
    '/checkoutstepper/step1/credit-debit-cards/discover.svg',
    '/checkoutstepper/step1/credit-debit-cards/jcb.svg',
  ],
  alipay: [
    '/checkoutstepper/step1/alipay-wechat/alipay.svg',
    '/checkoutstepper/step1/alipay-wechat/wechatpay.svg',
    '/checkoutstepper/step1/alipay-wechat/unionpay.svg',
  ],
  bank: ['/checkoutstepper/step1/bank-transfer/bank-transfer.svg'],
  paypal: ['/checkoutstepper/step1/paypal/paypal.svg'],
  stripe: ['/checkoutstepper/step1/stripe/stripe.svg'],
  worldfirst: [],
}

const couponInput = computed({
  get: () => props.couponInput ?? '',
  set: value => emit('coupon-input', value),
})
const isApplyingCoupon = computed(() => props.isApplyingCoupon ?? false)
const appliedCoupon = computed(() => props.appliedCoupon ?? null)
const appliedCouponDisplay = computed(() => {
  const coupon = appliedCoupon.value
  if (coupon && typeof coupon === 'object') {
    const obj = coupon as { code?: string; label?: string }
    return obj.code ?? obj.label ?? ''
  }
  return typeof coupon === 'string' ? coupon : ''
})
const pointsAvailable = computed(() => props.pointsAvailable ?? 0)
const isUsingPoints = computed(() => props.isUsingPoints ?? false)
const pointsToUse = computed(() => props.pointsToUse ?? 0)
const maxPointsToUse = computed(() => props.maxPointsToUse ?? pointsAvailable.value)
const pointsHint = computed(() => props.pointsHint ?? t('checkout.modal.pointsHint'))
const orderSummary = computed(() => props.orderSummary)
const currency = computed(() => props.currency ?? 'USD')
const showShippingForm = computed(() => props.showShippingForm !== false && Boolean(props.shippingForm))
const shippingForm = computed(() => props.shippingForm)
const countrySearchValue = computed(() => props.countrySearch ?? '')
const shippableCountries = computed(() => { if (!props.shippableCountries) throw new Error("[CRITICAL] shippableCountries missing"); return props.shippableCountries; })
const nonShippableCountries = computed(() => { if (!props.nonShippableCountries) throw new Error("[CRITICAL] nonShippableCountries missing"); return props.nonShippableCountries; })
const shippingValidation = computed(() => props.shippingValidation)
const estimatedDelivery = computed(() => props.estimatedDelivery ?? null)
const zipPlaceholder = computed(() => props.zipPlaceholder ?? t('checkout.stepper.shipping.zipPlaceholder'))
const zipHint = computed(() => props.zipHint ?? '')
const desktopCtaLabel = computed(() => props.desktopCtaLabel ?? t('checkout.payment.card.cta'))
const ctaDescription = computed(() => props.ctaDescription ?? '')
const mobilePaymentTitle = computed(() => props.mobilePaymentTitle ?? t('checkout.stepper.review.continueToCheckout'))
const mobilePaymentDescription = computed(() => props.mobilePaymentDescription ?? '')
const isSubmitting = computed(() => props.isSubmitting ?? false)
const giftCardDiscount = computed(() => {
  const totals = orderSummary.value?.totals as unknown as { giftCardDiscount?: number } | undefined
  const val = totals?.giftCardDiscount
  return typeof val === 'number' ? val : 0
})

const isShippingComplete = computed(() => {
  if (!showShippingForm.value) return true
  const form = shippingForm.value
  if (!form) return false
  const requiredFields: Array<keyof ShippingForm> = ['country', 'name', 'phone', 'address', 'city']
  return requiredFields.every(field => {
    const value = form[field]
    if (typeof value === 'string') {
      return value.trim().length > 0
    }
    return Boolean(value)
  })
})

const canProceedToReview = computed(() => {
  if (!showShippingForm.value) return true
  if (!isShippingComplete.value) return false
  if (shippingValidation.value && shippingValidation.value.isShippable === false) return false
  return true
})

const setStep = (step: Step) => {
  currentStep.value = step
  emit('update:step', step)
}

const setMethod = (method: string) => {
  activeMethod.value = method
  emit('update:method', method)
}

const handleSelect = (id: string) => {
  if (activeMethod.value === id && currentStep.value === 1) return
  setMethod(id)
}

const handleContinueFromStepOne = () => {
  setStep(2)
}

const handleContinueToReview = () => {
  setStep(3)
}

const handleBackToMethods = () => {
  setStep(1)
}

const handleBackToShipping = () => {
  setStep(2)
}

const handleCouponInput = (event: Event) => {
  const target = event.target as HTMLInputElement | null
  couponInput.value = target?.value ?? ''
}

const handleApplyCouponClick = () => {
  if (!couponInput.value || isApplyingCoupon.value) return
  emit('apply-coupon')
}

const handlePointsToggle = (event: Event) => {
  const target = event.target as HTMLInputElement | null
  emit('toggle-points', Boolean(target?.checked))
}

const handlePointsInputChange = (event: Event) => {
  const target = event.target as HTMLInputElement | null
  const value = Number(target?.value ?? 0)
  emit('points-input', Number.isFinite(value) ? value : 0)
}

const emitShippingField = (field: keyof ShippingForm, value: string) => {
  emit('update-shipping-field', { field, value })
}

const handleShippingInput = (field: keyof ShippingForm, event: Event) => {
  const target = event.target as HTMLInputElement | HTMLTextAreaElement | HTMLSelectElement | null
  emitShippingField(field, target?.value ?? '')
}

const handleCountrySearchInput = (event: Event) => {
  const target = event.target as HTMLInputElement | null
  emit('country-search', target?.value ?? '')
}

const formatPrice = (value: number) => {
  try {
    return new Intl.NumberFormat(locale.value, {
      style: 'currency',
      currency: currency.value,
    }).format(value)
  } catch {
    return `$${value.toFixed(2)}`
  }
}

const shippingLabel = computed(() => {
  const totals = orderSummary.value?.totals
  if (!totals) return ''
  if (totals.shippingState === 'select') {
    return t('checkout.stepper.shipping.state.selectCountry')
  }
  if (totals.shippingState === 'unavailable') {
    return t('checkout.stepper.shipping.state.unavailable')
  }
  if (totals.shipping === null) {
    return t('checkout.stepper.shipping.state.calculating')
  }
  if (totals.shipping === 0) {
    return t('checkout.stepper.shipping.state.free')
  }
  return formatPrice(totals.shipping)
})

const handleOpenContact = () => emit('open-contact')
const handleOpenFreight = () => emit('open-freight')
const handleSaveCart = () => emit('save-cart')

const handleSubmitMock = () => {
  emit('submit')
}
</script>
