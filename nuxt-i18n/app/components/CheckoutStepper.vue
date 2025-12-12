<template>
  <section class="flex flex-col gap-3 h-full px-1.5 md:px-6 py-2 md:py-3 max-w-none md:max-w-[720px] mx-auto w-full">
    <!-- Stepper header -->
    <div class="flex items-center gap-2 text-[10px] md:text-xs uppercase tracking-[0.2em] text-white/60">
      <StepperDot :active="currentStep === 1">1</StepperDot>
      <div class="hidden md:block" :class="currentStep === 1 ? 'text-white/90' : 'text-white/60'">Choose payment method</div>
      <div class="flex-1 h-px bg-white/10"></div>
      <StepperDot :active="currentStep === 2">2</StepperDot>
      <div class="hidden md:block" :class="currentStep === 2 ? 'text-white/90' : 'text-white/60'">Shipping info</div>
      <div class="flex-1 h-px bg-white/10"></div>
      <StepperDot :active="currentStep === 3">3</StepperDot>
      <div class="hidden md:block" :class="currentStep === 3 ? 'text-white/90' : 'text-white/60'">Review &amp; pay</div>
    </div>

    <!-- Step 1 -->
    <div v-if="currentStep === 1" class="flex flex-col gap-3 flex-1">
      <header class="flex items-center justify-between">
        <div class="text-xs font-semibold text-white/80 flex items-center gap-2">
          Pick a provider
          <span class="px-2 py-0.5 rounded-full text-[10px] uppercase tracking-[0.2em] bg-white/10 text-white/70">Step 1 of 3</span>
        </div>
        <span class="text-[11px] text-white/60">Tap a card to view details</span>
      </header>

      <div class="space-y-2 flex-1 overflow-y-auto pr-1 md:pr-2 max-w-[620px] mx-auto w-full">
        <article
          v-for="option in paymentOptions"
          :key="option.id"
          class="rounded-2xl bg-[radial-gradient(circle_at_top_left,rgba(31,41,55,0.96),rgba(15,23,42,0.98))] shadow-[6px_10px_30px_rgba(0,0,0,0.55)] overflow-hidden transition-colors"
        >
          <button
            type="button"
            class="w-full px-4 py-3 flex items-center justify-between gap-3 text-left"
            :class="activeMethod === option.id ? 'text-white' : 'text-white/80'"
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
                    :alt="`${option.title} icon`"
                    class="h-4 w-6 object-contain"
                    loading="lazy"
                    decoding="async"
                  />
                </div>
              </div>
              <span class="text-[11px] text-white/60">{{ option.subtitle }}</span>
            </div>
            <div
              class="inline-flex items-center justify-center rounded-full px-2 py-0.5 text-[10px] font-semibold uppercase tracking-[0.2em] shadow-[3px_3px_12px_rgba(0,0,0,0.35)]"
              :class="activeMethod === option.id ? 'bg-white text-slate-900' : 'bg-white/10 text-white/60'"
            >
              {{ activeMethod === option.id ? 'Selected' : 'Tap to view' }}
            </div>
          </button>

          <div
            v-if="activeMethod === option.id"
            class="px-4 pb-4 space-y-3 border-t border-white/5 text-[12px] text-white/70"
          >
            <p class="leading-relaxed">
              {{ option.description }}
            </p>
            <ul v-if="option.points?.length" class="space-y-1 text-white/70">
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
              Continue with this method →
            </button>
          </div>
        </article>
      </div>
    </div>

    <!-- Step 2 -->
    <div v-else-if="currentStep === 2" class="flex flex-col gap-3 flex-1">
      <button
        type="button"
        class="inline-flex items-center gap-2 text-[11px] font-semibold uppercase tracking-[0.2em] text-white/70 hover:text-white transition"
        @click="handleBackToMethods"
      >
        <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 19l-7-7 7-7" />
        </svg>
        Change payment method
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
            Shipping address
          </header>
          <p class="text-xs text-white/70">
            If shipping to your area isn't available yet, it doesn't always mean we can't deliver. Contact us and we'll check options for you.
          </p>

          <div class="space-y-3">
            <div class="grid grid-cols-1 gap-2.5 min-[420px]:grid-cols-2">
              <div class="min-[420px]:col-span-2 space-y-1.5">
                <label class="block text-sm font-medium text-white/80">Country / Region <span class="text-red-400">*</span></label>
                <input
                  :value="countrySearchValue"
                  type="text"
                  placeholder="Search country or region"
                  class="w-full px-3 py-1.5 rounded-lg border-none text-xs text-white placeholder:text-white/40 bg-[linear-gradient(135deg,rgba(15,23,42,0.98),rgba(15,23,42,0.96))] shadow-[0_2px_6px_-3px_rgba(0,0,0,0.9),0_0_6px_rgba(15,23,42,0.7)] focus:outline-none focus:[box-shadow:0_0_0_1px_rgba(56,189,248,0.85)]"
                  @input="handleCountrySearchInput"
                />
                <select
                  :value="shippingForm?.country || ''"
                  class="w-full px-4 py-1.5 rounded-lg border-none text-white bg-slate-900/80 shadow-[0_2px_6px_-3px_rgba(0,0,0,0.9),0_0_6px_rgba(15,23,42,0.7)] focus:outline-none focus:[box-shadow:0_0_0_1px_rgba(56,189,248,0.85)] [color-scheme:dark]"
                  :class="{ 'focus:[box-shadow:0_0_0_1px_rgba(248,113,113,0.9)]': shippingForm?.country && shippingValidation && !shippingValidation.isShippable }"
                  @change="handleShippingInput('country', $event)"
                >
                  <option value="" disabled>Select a country</option>
                  <optgroup label="Available for shipping">
                    <option
                      v-for="country in shippableCountries"
                      :key="country.code"
                      :value="country.code"
                    >
                      {{ country.name }}
                    </option>
                  </optgroup>
                  <optgroup label="Other countries">
                    <option
                      v-for="country in nonShippableCountries"
                      :key="country.code"
                      :value="country.code"
                      class="text-white/50"
                    >
                      {{ country.name }}
                    </option>
                  </optgroup>
                </select>
              </div>

              <div>
                <label class="block text-sm font-medium text-white/80 mb-1">Recipient <span class="text-red-400">*</span></label>
                <input
                  :value="shippingForm?.name || ''"
                  type="text"
                  placeholder="Enter recipient name"
                  class="w-full px-4 py-1.5 rounded-lg border-none text-white placeholder:text-white/40 bg-[linear-gradient(135deg,rgba(15,23,42,0.98),rgba(15,23,42,0.96))] shadow-[0_2px_6px_-3px_rgba(0,0,0,0.9),0_0_6px_rgba(15,23,42,0.7)] focus:outline-none focus:[box-shadow:0_0_0_1px_rgba(56,189,248,0.85)]"
                  @input="handleShippingInput('name', $event)"
                />
              </div>

              <div>
                <label class="block text-sm font-medium text-white/80 mb-1">Phone <span class="text-red-400">*</span></label>
                <input
                  :value="shippingForm?.phone || ''"
                  type="tel"
                  placeholder="Enter phone number"
                  class="w-full px-4 py-1.5 rounded-lg border-none text-white placeholder:text-white/40 bg-[linear-gradient(135deg,rgba(15,23,42,0.98),rgba(15,23,42,0.96))] shadow-[0_2px_6px_-3px_rgba(0,0,0,0.9),0_0_6px_rgba(15,23,42,0.7)] focus:outline-none focus:[box-shadow:0_0_0_1px_rgba(56,189,248,0.85)]"
                  @input="handleShippingInput('phone', $event)"
                />
              </div>

              <div class="min-[420px]:col-span-2">
                <label class="block text-sm font-medium text-white/80 mb-1">Address <span class="text-red-400">*</span></label>
                <textarea
                  :value="shippingForm?.address || ''"
                  rows="2"
                  placeholder="Enter detailed address"
                  class="w-full px-4 py-1.5 rounded-lg border-none text-white placeholder:text-white/40 bg-[linear-gradient(135deg,rgba(15,23,42,0.98),rgba(15,23,42,0.96))] shadow-[0_2px_6px_-3px_rgba(0,0,0,0.9),0_0_6px_rgba(15,23,42,0.7)] focus:outline-none focus:[box-shadow:0_0_0_1px_rgba(56,189,248,0.85)] resize-none"
                  @input="handleShippingInput('address', $event)"
                ></textarea>
              </div>
            </div>

            <div
              v-if="shippingForm?.country && shippingValidation && !shippingValidation.isShippable"
              class="px-4 py-2 bg-red-500/10 border border-red-500/30 rounded-lg space-y-2"
            >
              <p class="text-red-300 text-sm font-semibold">{{ shippingValidation.reason || 'Shipping unavailable for this country.' }}</p>
              <p class="text-white/60 text-xs">Don’t worry, we have options for you:</p>
              <div class="space-y-2 text-xs">
                <button type="button" class="flex items-center gap-2 text-blue-400 hover:text-blue-300" @click="handleOpenContact">
                  <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z" />
                  </svg>
                  Contact us for special shipping arrangements
                </button>
                <button type="button" class="flex items-center gap-2 text-white/70 hover:text-white" @click="handleOpenFreight">
                  <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7h12m0 0l-4-4m4 4l-4 4m0 6H4m0 0l4 4m-4-4l4-4" />
                  </svg>
                  Use a freight forwarder service
                </button>
                <button type="button" class="flex items-center gap-2 text-white/70 hover:text-white" @click="handleSaveCart">
                  <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 5a2 2 0 012-2h10a2 2 0 012 2v16l-7-3.5L5 21V5z" />
                  </svg>
                  Save cart and check back later
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
              Estimated delivery: {{ estimatedDelivery }}
            </div>

            <div class="grid grid-cols-2 gap-2.5">
              <div>
                <label class="block text-sm font-medium text-white/80 mb-1">City <span class="text-red-400">*</span></label>
                <input
                  :value="shippingForm?.city || ''"
                  type="text"
                  placeholder="City"
                  class="w-full px-4 py-1.5 rounded-lg border-none text-white placeholder:text-white/40 bg-[linear-gradient(135deg,rgba(15,23,42,0.98),rgba(15,23,42,0.96))] shadow-[0_2px_6px_-3px_rgba(0,0,0,0.9),0_0_6px_rgba(15,23,42,0.7)] focus:outline-none focus:[box-shadow:0_0_0_1px_rgba(56,189,248,0.85)]"
                  @input="handleShippingInput('city', $event)"
                />
              </div>
              <div>
                <label class="block text-sm font-medium text-white/80 mb-1">Zip code</label>
                <input
                  :value="shippingForm?.zip || ''"
                  type="text"
                  :placeholder="zipPlaceholder"
                  class="w-full px-4 py-1.5 rounded-lg border-none text-white placeholder:text-white/40 bg-[linear-gradient(135deg,rgba(15,23,42,0.98),rgba(15,23,42,0.96))] shadow-[0_2px_6px_-3px_rgba(0,0,0,0.9),0_0_6px_rgba(15,23,42,0.7)] focus:outline-none focus:[box-shadow:0_0_0_1px_rgba(56,189,248,0.85)]"
                  @input="handleShippingInput('zip', $event)"
                />
                <p v-if="zipHint" class="text-xs text-white/50 mt-1">{{ zipHint }}</p>
              </div>
            </div>

            <div class="flex justify-center md:justify-end pt-2">
              <button
                type="button"
                class="px-4 py-2 rounded-xl text-sm font-semibold transition w-full max-w-[280px] md:w-auto md:max-w-none"
                :class="[
                  canProceedToReview ? 'bg-white text-slate-900 hover:brightness-95' : 'bg-white/10 text-white/40 cursor-not-allowed'
                ]"
                :disabled="!canProceedToReview"
                @click="handleContinueToReview"
              >
                {{ canProceedToReview ? 'Continue to review →' : 'Complete shipping info first' }}
              </button>
            </div>
          </div>
        </section>
        <div v-else class="rounded-2xl border border-dashed border-white/20 px-4 py-6 text-center text-sm text-white/60 bg-white/5">
          Shipping details are not required for this method.
          <div class="mt-4 flex justify-center md:justify-end">
            <button
              type="button"
              class="px-4 py-2 rounded-xl text-sm font-semibold transition w-full max-w-[280px] md:w-auto md:max-w-none"
              :class="[
                canProceedToReview ? 'bg-white text-slate-900 hover:brightness-95' : 'bg-white/10 text-white/40 cursor-not-allowed'
              ]"
              :disabled="!canProceedToReview"
              @click="handleContinueToReview"
            >
              {{ canProceedToReview ? 'Continue to review →' : 'Complete shipping info first' }}
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- Step 3 -->
    <div v-else-if="currentStep === 3" class="flex flex-col gap-3 flex-1">
      <button
        type="button"
        class="inline-flex items-center gap-2 text-[11px] font-semibold uppercase tracking-[0.2em] text-white/70 hover:text-white transition"
        @click="handleBackToShipping"
      >
        <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 19l-7-7 7-7" />
        </svg>
        Back to shipping
      </button>

      <div class="space-y-3 flex-1 overflow-y-auto pr-1 md:pr-2">
        <div class="grid gap-3 md:grid-cols-3">
          <section class="rounded-2xl bg-[radial-gradient(circle_at_top_left,rgba(31,41,55,0.96),rgba(15,23,42,0.98))] px-4 py-4 shadow-[6px_10px_30px_rgba(0,0,0,0.55)] space-y-3">
            <header class="text-sm font-semibold text-white flex items-center gap-2">
              Coupon
              <span class="text-[11px] text-white/50">Optional</span>
            </header>
            <div class="flex flex-col sm:flex-row sm:flex-wrap gap-2">
              <input
                :value="couponInput"
                type="text"
                placeholder="Enter code"
                class="flex-1 px-3 py-2 rounded-xl bg-white/5 text-sm text-white placeholder:text-white/40 focus:outline-none focus:ring-2 focus:ring-white/30 shadow-[0_2px_6px_-3px_rgba(0,0,0,0.9),0_0_6px_rgba(15,23,42,0.7)]"
                @input="handleCouponInput"
              />
              <button
                type="button"
                class="px-3 py-2 rounded-xl text-sm font-semibold bg-white text-slate-900 hover:brightness-95 transition disabled:opacity-50 disabled:cursor-not-allowed w-full sm:w-auto sm:flex-none"
                :disabled="!couponInput || isApplyingCoupon"
                @click="handleApplyCouponClick"
              >
                {{ isApplyingCoupon ? 'Applying…' : 'Apply' }}
              </button>
            </div>
            <p v-if="appliedCouponDisplay" class="text-xs text-emerald-300 flex items-center gap-1">
              <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
              </svg>
              Coupon applied: {{ appliedCouponDisplay }}
            </p>
          </section>

          <section
            class="rounded-2xl bg-[radial-gradient(circle_at_top_left,rgba(31,41,55,0.96),rgba(15,23,42,0.98))] px-4 py-4 shadow-[6px_10px_30px_rgba(0,0,0,0.55)] space-y-3"
          >
            <header class="text-sm font-semibold text-white flex items-center gap-2">
              Use Points Discount
              <span class="text-[11px] text-white/50">
                {{ pointsAvailable > 0 ? `Available: ${pointsAvailable} pts` : 'No points available yet' }}
              </span>
            </header>
            <div class="flex items-center gap-2.5 mb-1 text-sm text-white/80">
              <input
                type="checkbox"
                class="w-4 h-4 rounded border-white/30 bg-white/10 text-[#6b73ff]"
                :checked="isUsingPoints"
                @change="handlePointsToggle"
              />
              <span>{{ pointsAvailable > 0 ? 'Use points' : 'Points unavailable' }}</span>
            </div>
            <div v-if="isUsingPoints" class="space-y-2">
              <label class="block text-xs text-white/70">Points to use</label>
              <input
                :value="pointsToUse"
                type="number"
                min="0"
                :max="maxPointsToUse"
                class="w-full px-3 py-2 rounded-xl bg-white/5 text-sm text-white border border-white/10 focus:outline-none focus:ring-2 focus:ring-white/30"
                @input="handlePointsInputChange"
              />
              <p class="text-xs text-white/60">{{ pointsHint }}</p>
            </div>
            <p v-if="pointsAvailable <= 0" class="text-xs text-white/50">
              Earn points on your next order and apply them here for instant discounts.
            </p>
          </section>

          <section class="rounded-2xl border border-white/10 bg-[radial-gradient(circle_at_top_left,rgba(31,41,55,0.96),rgba(15,23,42,0.98))] px-4 py-6 text-center text-sm text-white/60 shadow-[6px_10px_30px_rgba(0,0,0,0.55)]">
            <header class="text-sm font-semibold text-white flex items-center gap-2">
              <svg class="w-4 h-4 text-purple-300" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 10h.01M12 10h.01M16 10h.01M21 16c0 3.866-4.03 7-9 7s-9-3.134-9-7V8a4 4 0 014-4h10a4 4 0 014 4v8z" />
              </svg>
              Order notes <span class="text-[11px] text-white/50">(optional)</span>
            </header>
            <textarea
              :value="shippingForm?.notes || ''"
              rows="3"
              placeholder="Add any delivery instructions or special requests…"
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
            Order summary
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
                <p class="text-xs text-white/60 mt-0.5">Qty: {{ item.quantity }}</p>
                <p class="text-sm font-semibold text-white mt-1">{{ formatPrice(item.price * item.quantity) }}</p>
              </div>
            </article>
          </div>
          <div v-else class="text-xs text-white/50 text-center py-2 border border-dashed border-white/10 rounded-xl">
            Order details will appear here once items are in your cart.
          </div>

          <div class="space-y-2 text-sm text-white/70 pt-2 border-t border-white/10">
            <div class="flex justify-between">
              <span>Subtotal</span>
              <span class="font-medium text-white">{{ formatPrice(orderSummary?.totals?.subtotal ?? 0) }}</span>
            </div>
            <div class="flex justify-between items-start gap-4">
              <span>
                Shipping
                <span v-if="orderSummary?.totals?.shippingLabel" class="text-xs text-white/50">
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
              <span>Tax</span>
              <span class="font-medium text-white">{{ formatPrice(orderSummary?.totals?.tax ?? 0) }}</span>
            </div>
            <div class="flex justify-between" :class="orderSummary?.totals?.pointsDiscount ? 'text-emerald-300' : 'text-white/60'">
              <span>Points discount</span>
              <span>- {{ formatPrice(orderSummary?.totals?.pointsDiscount ?? 0) }}</span>
            </div>
            <div class="flex justify-between" :class="orderSummary?.totals?.couponDiscount ? 'text-emerald-300' : 'text-white/60'">
              <span>Coupon discount</span>
              <span>- {{ formatPrice(orderSummary?.totals?.couponDiscount ?? 0) }}</span>
            </div>
            <div class="flex justify-between text-white/60">
              <span>Gift card discount</span>
              <span>- {{ formatPrice(giftCardDiscount) }}</span>
            </div>
            <div class="flex justify-between text-lg font-bold pt-3 border-t border-white/20">
              <span>Total</span>
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
          <p v-if="ctaDescription" class="text-[11px] text-white/60 text-center">
            {{ ctaDescription }}
          </p>
        </div>
      </div>

      <div class="rounded-2xl bg-[radial-gradient(circle_at_bottom,rgba(15,23,42,0.98),rgba(2,6,23,0.95))] px-4 py-4 space-y-2 md:hidden shadow-[6px_10px_30px_rgba(0,0,0,0.55)]">
        <div class="text-center">
          <p class="text-xs font-semibold text-white">{{ mobilePaymentTitle }}</p>
          <p v-if="mobilePaymentDescription" class="text-[11px] text-white/60 mt-1">
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

const currentStep = ref<Step>(props.initialStep ?? 1)
const activeMethod = ref(props.initialMethod ?? 'card')

const paymentOptions: PaymentOption[] = [
  {
    id: 'card',
    title: 'Credit / Debit cards',
    subtitle: '≈ $11.00 · Secure card page',
    description: 'Card numbers are entered on a PCI-compliant page; we only need your full shipping information and phone so carriers can reach you.',
    points: ['Required: full shipping info + contact number', 'Use when you want to pay immediately with Visa / Mastercard.'],
  },
  {
    id: 'alipay',
    title: 'Alipay / WeChat',
    subtitle: '≈ $11.00 · Local wallets',
    description: 'Approve payment inside your usual wallet. We still gather shipping details here to prepare dispatch.',
    points: ['Required: recipient name + phone number', 'Great for China-mainland wallets and UnionPay.'],
  },
  {
    id: 'bank',
    title: 'Bank transfer',
    subtitle: '≈ $11.00 · Manual invoice',
    description: 'Place the order first, then send a bank transfer using the instructions on the next step.',
    points: ['Include your order number as payment reference.'],
  },
  {
    id: 'paypal',
    title: 'PayPal',
    subtitle: '≈ $11.00 · Express checkout',
    description: 'We rely on your PayPal address whenever possible; pick a country we can ship to before you continue.',
    points: ['Required: select a shippable country before continuing.'],
  },
  {
    id: 'stripe',
    title: 'Stripe',
    subtitle: '≈ $11.00 · 3D Secure ready',
    description: 'Stripe handles all card details and optional 3D Secure checks.',
    points: ['Great for customers that require SCA/3DS.'],
  },
  {
    id: 'worldfirst',
    title: 'WorldFirst',
    subtitle: '≈ $11.00 · Cross-border',
    description: 'Ideal for international business orders. Pay into a local WorldFirst account; they settle the funds to us.',
    points: ['Recommended for B2B cross-border payments.'],
  },
]

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
const pointsHint = computed(() => props.pointsHint ?? '1 point = $0.01, max 50% of order')
const orderSummary = computed(() => props.orderSummary)
const currency = computed(() => props.currency ?? 'USD')
const showShippingForm = computed(() => props.showShippingForm !== false && Boolean(props.shippingForm))
const shippingForm = computed(() => props.shippingForm)
const countrySearchValue = computed(() => props.countrySearch ?? '')
const shippableCountries = computed(() => props.shippableCountries ?? [])
const nonShippableCountries = computed(() => props.nonShippableCountries ?? [])
const shippingValidation = computed(() => props.shippingValidation)
const estimatedDelivery = computed(() => props.estimatedDelivery ?? null)
const zipPlaceholder = computed(() => props.zipPlaceholder ?? 'Zip code')
const zipHint = computed(() => props.zipHint ?? '')
const desktopCtaLabel = computed(() => props.desktopCtaLabel ?? 'Continue to secure payment')
const ctaDescription = computed(() => props.ctaDescription ?? '')
const mobilePaymentTitle = computed(() => props.mobilePaymentTitle ?? 'Continue to checkout')
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
    return new Intl.NumberFormat('en-US', {
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
    return 'Select country'
  }
  if (totals.shippingState === 'unavailable') {
    return 'Unavailable'
  }
  if (totals.shipping === null) {
    return 'Calculating...'
  }
  if (totals.shipping === 0) {
    return 'Free'
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
