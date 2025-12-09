<template>
  <teleport to="body">
    <!-- 遮罩层 -->
    <transition name="fade">
      <div
        v-if="isCheckoutOpen"
        class="fixed inset-0 z-[9998] flex items-center justify-center p-4"
        @click.self="closeCheckout"
      >
        <!-- 半透明背景遮罩 -->
        <div class="absolute inset-0 bg-black/80 backdrop-blur-sm"></div>
        <!-- 结账弹窗 -->
        <transition name="scale">
          <div v-if="isCheckoutOpen" class="relative flex flex-col w-full max-w-[1400px] h-[90vh] md:h-[700px] max-h-[85vh] overflow-hidden rounded-2xl bg-[radial-gradient(circle_at_top_left,rgba(15,23,42,0.98),rgba(0,0,0,1))] backdrop-blur-xl border border-white/10 shadow-[0_24px_56px_-18px_rgba(0,0,0,1)]">
            <!-- 头部：再次压缩高度，为下方内容留出更多空间 -->
            <div class="flex items-center justify-between px-6 py-2.5 border-b border-white/10">
              <div class="flex items-center gap-2">
                <button
                  @click="backToCart"
                  class="px-3 py-0.5 rounded-full text-[11px] font-medium text-white/80 bg-[linear-gradient(135deg,rgba(15,23,42,0.98),rgba(15,23,42,0.96))] shadow-[0_2px_6px_-3px_rgba(0,0,0,0.9),0_0_6px_rgba(15,23,42,0.7)] hover:text-white hover:-translate-y-0.5 transition-all"
                >
                  ← Back to Cart
                </button>
                <h2 class="text-lg font-bold text-white">
                  💳 Checkout
                </h2>
              </div>
              <button
                @click="closeCheckout"
                class="w-8 h-8 flex items-center justify-center rounded-full hover:bg-white/10 transition-colors"
                aria-label="Close checkout"
              >
                <svg class="w-5 h-5 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
                </svg>
              </button>
            </div>

            <!-- 内容区域 -->
            <div class="flex-1 overflow-y-auto">
              <div class="px-6 pt-1 pb-6 space-y-4">
                <div class="space-y-4">
                <!-- 右侧：支付方式 Tabs + Shipping Address + 订单摘要 -->
                <div class="space-y-4">
                  <!-- 支付方式 Tabs 头部 + Card 说明（静态布局，仅作为 HTML Demo 的第一步迁移） -->
                  <div class="space-y-4">
                    <!-- Tabs 头部 -->
                    <div
                      class="w-full mt-2 rounded-full bg-[linear-gradient(135deg,rgba(15,23,42,0.98),rgba(15,23,42,0.96))] px-1 py-2 shadow-[0_14px_36px_-20px_rgba(0,0,0,1)] flex flex-wrap items-center justify-center gap-1.5"
                    >
                      <!-- Active: Card -->
                      <button
                        type="button"
                        @click="setActivePaymentTab('card')"
                        :class="[
                          'inline-flex items-center gap-2 px-4 py-2 text-left rounded-full text-[11px] md:text-[12px]',
                          activePaymentTab === 'card'
                            ? 'bg-[linear-gradient(135deg,#4efce7_0%,#60a5fa_100%)] text-slate-950 shadow-[0_20px_46px_-20px_rgba(0,0,0,1)] font-semibold'
                            : 'bg-[linear-gradient(135deg,rgba(31,41,55,0.95),rgba(15,23,42,0.98))] text-white/90 shadow-[0_18px_40px_-18px_rgba(0,0,0,1)]'
                        ]"
                      >
                        <span>Card</span>
                        <span
                          :class="[
                            'text-[10px] md:text-[11px] font-normal',
                            activePaymentTab === 'card' ? 'text-slate-900/80' : 'text-emerald-200/80'
                          ]"
                        >
                          · ≈ {{ formatPrice(priceBreakdown.total) }}
                        </span>
                      </button>
                      <!-- PayPal -->
                      <button
                        type="button"
                        @click="setActivePaymentTab('paypal')"
                        :class="[
                          'inline-flex items-center gap-2 px-4 py-2 text-left rounded-full text-[11px] md:text-[12px]',
                          activePaymentTab === 'paypal'
                            ? 'bg-[linear-gradient(135deg,#4efce7_0%,#60a5fa_100%)] text-slate-950 shadow-[0_20px_46px_-20px_rgba(0,0,0,1)] font-semibold'
                            : 'bg-[linear-gradient(135deg,rgba(31,41,55,0.95),rgba(15,23,42,0.98))] text-white/90 shadow-[0_18px_40px_-18px_rgba(0,0,0,1)]'
                        ]"
                      >
                        <span>PayPal</span>
                        <span class="text-[10px] md:text-[11px] font-normal text-emerald-200/80">· ≈ {{ formatPrice(priceBreakdown.total) }}</span>
                      </button>
                      <!-- Alipay / WeChat / UnionPay -->
                      <button
                        type="button"
                        @click="setActivePaymentTab('alipay')"
                        :class="[
                          'inline-flex items-center gap-2 px-4 py-2 text-left rounded-full text-[11px] md:text-[12px]',
                          activePaymentTab === 'alipay'
                            ? 'bg-[linear-gradient(135deg,#4efce7_0%,#60a5fa_100%)] text-slate-950 shadow-[0_20px_46px_-20px_rgba(0,0,0,1)] font-semibold'
                            : 'bg-[linear-gradient(135deg,rgba(31,41,55,0.95),rgba(15,23,42,0.98))] text-white/90 shadow-[0_18px_40px_-18px_rgba(0,0,0,1)]'
                        ]"
                      >
                        <span>Alipay / WeChat / UnionPay</span>
                        <span class="text-[10px] md:text-[11px] font-normal text-emerald-200/80">· ≈ {{ formatPrice(priceBreakdown.total) }}</span>
                      </button>
                      <!-- Stripe -->
                      <button
                        type="button"
                        @click="setActivePaymentTab('stripe')"
                        :class="[
                          'inline-flex items-center gap-2 px-4 py-2 text-left rounded-full text-[11px] md:text-[12px]',
                          activePaymentTab === 'stripe'
                            ? 'bg-[linear-gradient(135deg,#4efce7_0%,#60a5fa_100%)] text-slate-950 shadow-[0_20px_46px_-20px_rgba(0,0,0,1)] font-semibold'
                            : 'bg-[linear-gradient(135deg,rgba(31,41,55,0.95),rgba(15,23,42,0.98))] text-white/90 shadow-[0_18px_40px_-18px_rgba(0,0,0,1)]'
                        ]"
                      >
                        <span>Stripe</span>
                        <span class="text-[10px] md:text-[11px] font-normal text-emerald-200/80">· ≈ {{ formatPrice(priceBreakdown.total) }}</span>
                      </button>
                      <!-- Bank transfer -->
                      <button
                        type="button"
                        @click="setActivePaymentTab('bank')"
                        :class="[
                          'inline-flex items-center gap-2 px-4 py-2 text-left rounded-full text-[11px] md:text-[12px]',
                          activePaymentTab === 'bank'
                            ? 'bg-[linear-gradient(135deg,#4efce7_0%,#60a5fa_100%)] text-slate-950 shadow-[0_20px_46px_-20px_rgba(0,0,0,1)] font-semibold'
                            : 'bg-[linear-gradient(135deg,rgba(31,41,55,0.95),rgba(15,23,42,0.98))] text-white/90 shadow-[0_18px_40px_-18px_rgba(0,0,0,1)]'
                        ]"
                      >
                        <span>Bank transfer</span>
                        <span class="text-[10px] md:text-[11px] font-normal text-emerald-200/80">· ≈ {{ formatPrice(priceBreakdown.total) }}</span>
                      </button>
                      <!-- WorldFirst -->
                      <button
                        type="button"
                        @click="setActivePaymentTab('worldfirst')"
                        :class="[
                          'inline-flex items-center gap-2 px-4 py-2 text-left rounded-full text-[11px] md:text-[12px]',
                          activePaymentTab === 'worldfirst'
                            ? 'bg-[linear-gradient(135deg,#4efce7_0%,#60a5fa_100%)] text-slate-950 shadow-[0_20px_46px_-20px_rgba(0,0,0,1)] font-semibold'
                            : 'bg-[linear-gradient(135deg,rgba(31,41,55,0.95),rgba(15,23,42,0.98))] text-white/90 shadow-[0_18px_40px_-18px_rgba(0,0,0,1)]'
                        ]"
                      >
                        <span>WorldFirst</span>
                        <span class="text-[10px] md:text-[11px] font-normal text-emerald-200/80">· ≈ {{ formatPrice(priceBreakdown.total) }}</span>
                      </button>
                    </div>

                  </div>

                  <!-- 安全说明 + Coupon + Points + Notes 一行四卡（桌面端等宽 3:3:3:3） -->
                  <div class="grid gap-3 md:grid-cols-12">
                    <!-- Secure payment & data protection -->
                    <section class="md:col-span-3">
                      <div class="flex items-start gap-2 p-2.5 rounded-xl bg-[linear-gradient(135deg,rgba(15,23,42,0.98),rgba(15,23,42,0.96))] shadow-[0_4px_16px_-10px_rgba(0,0,0,1)]">
                        <div class="mt-0.5 flex h-8 w-8 items-center justify-center rounded-full bg-emerald-500/20 text-emerald-300">
                          <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4.5l6 2.25v5.25c0 3.5-2.55 6.75-6 7.5-3.45-.75-6-4-6-7.5V6.75L12 4.5z" />
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9.75 12.75L11.25 14.25 14.25 11.25" />
                          </svg>
                        </div>
                        <div class="space-y-1">
                          <p class="text-sm font-semibold text-emerald-300 leading-snug">
                            Secure payment &amp; data protection
                          </p>
                          <p class="text-xs text-white/70 leading-snug">
                            Pages use HTTPS with trusted SSL; payments are processed by providers like PayPal / Stripe / Alipay. We only store the result, not your card details.
                          </p>
                        </div>
                      </div>
                    </section>

                    <!-- Coupon -->
                    <div class="md:col-span-3 rounded-xl p-2.5 bg-[radial-gradient(circle_at_top_left,rgba(31,41,55,0.96),rgba(15,23,42,0.98))] shadow-[0_10px_26px_-14px_rgba(0,0,0,0.95)]">
                      <h3 class="text-sm font-semibold text-white mb-1 flex items-center gap-2">
                        <svg class="w-4 h-4 text-red-500" fill="currentColor" viewBox="0 0 20 20">
                          <path fill-rule="evenodd" d="M5 5a3 3 0 015-2.236A3 3 0 0114.83 6H16a2 2 0 110 4h-5V9a1 1 0 10-2 0v1H4a2 2 0 110-4h1.17C5.06 5.687 5 5.35 5 5zm4 1V5a1 1 0 10-1 1h1zm3 0a1 1 0 10-1-1v1h1z" clip-rule="evenodd" />
                        </svg>
                        Coupon
                      </h3>
                      <div class="flex gap-2">
                        <input
                          v-model="couponCode"
                          type="text"
                          placeholder="Enter coupon code"
                          class="flex-1 px-3 py-1.5 rounded-lg border-none text-sm text-white placeholder:text-white/40 bg-[linear-gradient(135deg,rgba(15,23,42,0.98),rgba(15,23,42,0.96))] shadow-[0_2px_6px_-3px_rgba(0,0,0,0.9),0_0_6px_rgba(15,23,42,0.7)] focus:outline-none focus:[box-shadow:0_0_0_1px_rgba(56,189,248,0.85)]"
                        />
                        <button
                          @click="handleApplyCoupon"
                          :disabled="!couponCode || isApplyingCoupon"
                          class="px-3.5 py-1.5 bg-[#6b73ff] text-white rounded-lg hover:brightness-110 transition-all text-sm font-medium disabled:opacity-50 disabled:cursor-not-allowed"
                        >
                          {{ isApplyingCoupon ? 'Applying...' : 'Apply' }}
                        </button>
                      </div>
                      <p v-if="calculation.appliedCoupon.value" class="mt-1 text-xs text-green-400 flex items-center gap-1">
                        <svg class="w-4 h-4" fill="currentColor" viewBox="0 0 20 20">
                          <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clip-rule="evenodd" />
                        </svg>
                        Coupon applied: {{ calculation.appliedCoupon.value.code }}
                      </p>
                    </div>

                    <!-- Points Discount -->
                    <div v-if="calculation.userPoints.value" class="md:col-span-3 rounded-xl p-2.5 bg-[radial-gradient(circle_at_top_left,rgba(31,41,55,0.96),rgba(15,23,42,0.98))] shadow-[0_10px_26px_-14px_rgba(0,0,0,0.95)]">
                      <h3 class="text-sm font-semibold text-white mb-1 flex items-center gap-2">
                        <svg class="w-4 h-4 text-purple-500" fill="currentColor" viewBox="0 0 20 20">
                          <path d="M8.433 7.418c.155-.103.346-.196.567-.267v1.698a2.305 2.305 0 01-.567-.267C8.07 8.34 8 8.114 8 8c0-.114.07-.34.433-.582zM11 12.849v-1.698c.22.071.412.164.567.267.364.243.433.468.433.582 0 .114-.07.34-.433.582a2.305 2.305 0 01-.567.267z" />
                          <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm1-13a1 1 0 10-2 0v.092a4.535 4.535 0 00-1.676.662C6.602 6.234 6 7.009 6 8c0 .99.602 1.765 1.324 2.246.48.32 1.054.545 1.676.662v1.941c-.391-.127-.68-.317-.843-.504a1 1 0 10-1.51 1.31c.562.649 1.413 1.076 2.353 1.253V15a1 1 0 102 0v-.092a4.535 4.535 0 001.676-.662C13.398 13.766 14 12.991 14 12c0-.99-.602-1.765-1.324-2.246A4.535 4.535 0 0011 9.092V7.151c.391.127.68.317.843.504a1 1 0 101.511-1.31c-.563-.649-1.413-1.076-2.354-1.253V5z" clip-rule="evenodd" />
                        </svg>
                        Use Points Discount
                      </h3>
                      <div class="flex items-center gap-2.5 mb-1">
                        <input
                          v-model="calculation.usePointsDiscount.value"
                          type="checkbox"
                          class="w-4 h-4 text-[#6b73ff] rounded"
                        />
                        <span class="text-sm text-white/80">Use points (Available: {{ calculation.userPoints.value.available }} pts)</span>
                      </div>
                      <div v-if="calculation.usePointsDiscount.value" class="mt-2">
                        <label class="block text-xs text-white/70 mb-1">Points to use</label>
                        <input
                          :value="calculation.pointsToUse.value"
                          @input="calculation.setPointsUsage(parseInt(($event.target && $event.target.value) || '0', 10) || 0)"
                          type="number"
                          :max="calculation.userPoints.value.available"
                          min="0"
                          class="w-full px-3 py-1.5 rounded-lg border-none text-sm text-white bg-[linear-gradient(135deg,rgba(15,23,42,0.98),rgba(15,23,42,0.96))] shadow-[0_2px_6px_-3px_rgba(0,0,0,0.9),0_0_6px_rgba(15,23,42,0.7)] focus:outline-none focus:[box-shadow:0_0_0_1px_rgba(56,189,248,0.85)]"
                        />
                        <p class="mt-1 text-xs text-white/60">1 point = $0.01, max 50% of order</p>
                      </div>
                    </div>

                    <!-- Notes -->
                    <div class="md:col-span-3 rounded-xl p-2.5 bg-[radial-gradient(circle_at_top_left,rgba(31,41,55,0.96),rgba(15,23,42,0.98))] shadow-[0_10px_26px_-14px_rgba(0,0,0,0.95)]">
                      <label class="block text-sm font-medium text-white/80 mb-1">Order Notes (Optional)</label>
                      <textarea
                        v-model="form.notes"
                        rows="2"
                        placeholder="Any special requests..."
                        class="w-full px-3.5 py-1.5 rounded-lg border-none text-white placeholder:text-white/40 bg-[linear-gradient(135deg,rgba(15,23,42,0.98),rgba(15,23,42,0.96))] shadow-[0_2px_6px_-3px_rgba(0,0,0,0.9),0_0_6px_rgba(15,23,42,0.7)] focus:outline-none focus:[box-shadow:0_0_0_1px_rgba(56,189,248,0.85)] resize-none"
                      ></textarea>
                    </div>
                  </div>

                  <!-- Card 说明 + Shipping Address + Order Summary 并排（桌面端三列，移动端一列） -->
                  <div class="grid gap-3 md:grid-cols-12">
                    <!-- Card 说明块（不含礼品卡与积分） -->
                    <section
                      class="md:col-span-3 md:h-[420px] space-y-3 rounded-2xl p-4 bg-[radial-gradient(circle_at_top_left,rgba(31,41,55,0.96),rgba(15,23,42,0.98))] shadow-[0_10px_26px_-14px_rgba(0,0,0,0.95)]"
                    >
                      <header class="space-y-1">
                        <h3 class="text-sm font-semibold text-white flex items-center gap-2">
                          <span class="inline-flex h-6 w-6 items-center justify-center rounded-full bg-sky-500/20 text-sky-300">
                            💳
                          </span>
                          <span>Pay with credit / debit card</span>
                        </h3>
                      </header>

                      <div class="mt-3 grid grid-cols-1 md:grid-cols-1 gap-2 text-sm text-white/60">
                        <div class="space-y-1">
                          <p class="font-medium text-white/80 flex items-center gap-1">
                            <span class="inline-flex h-1 w-4 rounded-full bg-emerald-400/80"></span>
                            Why we ask for your phone number
                          </p>
                          <p>
                            For international shipments the carrier may need to contact you for customs clearance or final
                            delivery. We only share this with the logistics providers handling your order.
                          </p>
                        </div>
                        <div class="space-y-1">
                          <p class="font-medium text-white/80 flex items-center gap-1">
                            <span class="inline-flex h-1 w-4 rounded-full bg-sky-400/80"></span>
                            When to choose card payments
                          </p>
                          <p>
                            Good if you prefer to pay directly with Visa / MasterCard and keep everything in one checkout
                            flow without being redirected.
                          </p>
                        </div>
                      </div>
                    </section>

                    <!-- Shipping Address（Card Tab 所需的收货信息） -->
                    <div class="md:col-span-6 md:h-[420px] rounded-xl p-5 bg-[radial-gradient(circle_at_top_left,rgba(31,41,55,0.96),rgba(15,23,42,0.98))] shadow-[0_10px_26px_-14px_rgba(0,0,0,0.95)] md:overflow-y-auto">
                      <h3 class="text-base font-semibold text-white mb-2 flex items-center gap-2">
                        <svg class="w-5 h-5 text-blue-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17.657 16.657L13.414 20.9a1.998 1.998 0 01-2.827 0l-4.244-4.243a8 8 0 1111.314 0z" />
                          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 11a3 3 0 11-6 0 3 3 0 016 0z" />
                        </svg>
                        Shipping Address
                      </h3>

                      <div class="space-y-3">
                        <!-- 4 个主要字段：Country / Recipient / Phone / Address，在桌面端两行两列排列 -->
                        <div class="grid md:grid-cols-2 gap-3">
                          <!-- Country Selection -->
                          <div>
                            <label class="block text-sm font-medium text-white/80 mb-1">Country / Region <span class="text-red-400">*</span></label>
                            <select
                              v-model="form.country"
                              class="w-full px-4 py-2.5 rounded-lg border-none text-white bg-[linear-gradient(135deg,rgba(15,23,42,0.98),rgba(15,23,42,0.96))] shadow-[0_2px_6px_-3px_rgba(0,0,0,0.9),0_0_6px_rgba(15,23,42,0.7)] focus:outline-none focus:[box-shadow:0_0_0_1px_rgba(56,189,248,0.85)]"
                              :class="{ 'focus:[box-shadow:0_0_0_1px_rgba(248,113,113,0.9)]': form.country && !shippingValidation.isShippable }"
                            >
                              <option value="" disabled class="bg-gray-900">Select a country</option>
                              <optgroup label="Available for Shipping" class="bg-gray-900">
                                <option
                                  v-for="country in shippableCountries"
                                  :key="country.code"
                                  :value="country.code"
                                  class="bg-gray-900"
                                >
                                  {{ country.name }}
                                </option>
                              </optgroup>
                              <optgroup label="Other Countries" class="bg-gray-900">
                                <option
                                  v-for="country in nonShippableCountries"
                                  :key="country.code"
                                  :value="country.code"
                                  class="bg-gray-900 text-white/50"
                                >
                                  {{ country.name }}
                                </option>
                              </optgroup>
                            </select>
                          </div>

                          <div>
                            <label class="block text-sm font-medium text-white/80 mb-1">Recipient <span class="text-red-400">*</span></label>
                            <input
                              v-model="form.name"
                              type="text"
                              placeholder="Enter recipient name"
                              class="w-full px-4 py-2.5 rounded-lg border-none text-white placeholder:text-white/40 bg-[linear-gradient(135deg,rgba(15,23,42,0.98),rgba(15,23,42,0.96))] shadow-[0_2px_6px_-3px_rgba(0,0,0,0.9),0_0_6px_rgba(15,23,42,0.7)] focus:outline-none focus:[box-shadow:0_0_0_1px_rgba(56,189,248,0.85)]"
                            />
                          </div>

                          <div>
                            <label class="block text-sm font-medium text-white/80 mb-1">Phone <span class="text-red-400">*</span></label>
                            <input
                              v-model="form.phone"
                              type="tel"
                              placeholder="Enter phone number"
                              class="w-full px-4 py-2.5 rounded-lg border-none text-white placeholder:text-white/40 bg-[linear-gradient(135deg,rgba(15,23,42,0.98),rgba(15,23,42,0.96))] shadow-[0_2px_6px_-3px_rgba(0,0,0,0.9),0_0_6px_rgba(15,23,42,0.7)] focus:outline-none focus:[box-shadow:0_0_0_1px_rgba(56,189,248,0.85)]"
                            />
                          </div>

                          <div>
                            <label class="block text-sm font-medium text-white/80 mb-1">Address <span class="text-red-400">*</span></label>
                            <textarea
                              v-model="form.address"
                              rows="1"
                              placeholder="Enter detailed address"
                              class="w-full px-4 py-2.5 rounded-lg border-none text-white placeholder:text-white/40 bg-[linear-gradient(135deg,rgba(15,23,42,0.98),rgba(15,23,42,0.96))] shadow-[0_2px_6px_-3px_rgba(0,0,0,0.9),0_0_6px_rgba(15,23,42,0.7)] focus:outline-none focus:[box-shadow:0_0_0_1px_rgba(56,189,248,0.85)] resize-none"
                            ></textarea>
                          </div>
                        </div>

                        <!-- Shipping Unavailable Warning -->
                        <div
                          v-if="form.country && !shippingValidation.isShippable"
                          class="p-4 bg-red-500/10 border border-red-500/30 rounded-lg"
                        >
                          <div class="flex items-start gap-3">
                            <svg class="w-6 h-6 text-red-400 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
                            </svg>
                            <div class="flex-1">
                              <p class="text-red-400 font-medium">{{ shippingValidation.reason }}</p>
                              <p class="text-white/60 text-sm mt-2">Don't worry, we have options for you:</p>
                              <div class="mt-3 space-y-2">
                                <button
                                  type="button"
                                  @click="openContactSupport"
                                  class="flex items-center gap-2 text-sm text-blue-400 hover:text-blue-300 transition-colors"
                                >
                                  <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z" />
                                  </svg>
                                  Contact us for special shipping arrangements
                                </button>
                                <button
                                  type="button"
                                  @click="openFreightForwarder"
                                  class="flex items-center gap-2 text-sm text-white/70 hover:text-white transition-colors"
                                >
                                  <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7h12m0 0l-4-4m4 4l-4 4m0 6H4m0 0l4 4m-4-4l4-4" />
                                  </svg>
                                  Use a freight forwarder service
                                </button>
                                <button
                                  type="button"
                                  @click="saveCartForLater"
                                  class="flex items-center gap-2 text-sm text-white/70 hover:text-white transition-colors"
                                >
                                  <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 5a2 2 0 012-2h10a2 2 0 012 2v16l-7-3.5L5 21V5z" />
                                  </svg>
                                  Save cart and check back later
                                </button>
                              </div>
                            </div>
                          </div>
                        </div>

                        <!-- Estimated Delivery -->
                        <div
                          v-if="form.country && shippingValidation.isShippable && estimatedDelivery"
                          class="p-2 bg-green-500/10 border border-green-500/30 rounded-lg"
                        >
                          <p class="text-green-400 text-sm flex items-center gap-2">
                            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
                            </svg>
                            Estimated delivery: {{ estimatedDelivery }}
                          </p>
                        </div>

                        <div class="grid grid-cols-2 gap-3">
                          <div>
                            <label class="block text-sm font-medium text-white/80 mb-1">City <span class="text-red-400">*</span></label>
                            <input
                              v-model="form.city"
                              type="text"
                              placeholder="City"
                              class="w-full px-4 py-2.5 rounded-lg border-none text-white placeholder:text-white/40 bg-[linear-gradient(135deg,rgba(15,23,42,0.98),rgba(15,23,42,0.96))] shadow-[0_2px_6px_-3px_rgba(0,0,0,0.9),0_0_6px_rgba(15,23,42,0.7)] focus:outline-none focus:[box-shadow:0_0_0_1px_rgba(56,189,248,0.85)]"
                            />
                          </div>
                          <div>
                            <label class="block text-sm font-medium text-white/80 mb-1">Zip Code</label>
                            <input
                              v-model="form.zip"
                              type="text"
                              :placeholder="zipPlaceholder"
                              class="w-full px-4 py-2.5 rounded-lg border-none text-white placeholder:text-white/40 bg-[linear-gradient(135deg,rgba(15,23,42,0.98),rgba(15,23,42,0.96))] shadow-[0_2px_6px_-3px_rgba(0,0,0,0.9),0_0_6px_rgba(15,23,42,0.7)] focus:outline-none focus:[box-shadow:0_0_0_1px_rgba(56,189,248,0.85)]"
                            />
                            <p v-if="zipHint" class="text-xs text-white/50 mt-1">{{ zipHint }}</p>
                          </div>
                        </div>
                      </div>
                    </div>

                    <!-- 现有 Order Summary 卡片：暂时代表 Card Tab 的金额汇总，先保持逻辑不变 -->
                    <div class="md:col-span-3 md:h-[420px] rounded-xl p-4 bg-[radial-gradient(circle_at_top_left,rgba(31,41,55,0.96),rgba(15,23,42,0.98))] shadow-[0_10px_26px_-14px_rgba(0,0,0,0.95)] flex flex-col">
                      <h3 class="text-sm font-semibold text-white mb-1 flex items-center gap-1">
                        <svg class="w-4 h-4 text-green-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2" />
                        </svg>
                        Order Summary
                      </h3>

                      <!-- 商品列表 -->
                      <div class="space-y-3 mb-4 max-h-40 overflow-y-auto">
                        <div
                          v-for="item in cartItems"
                          :key="item.id"
                          class="flex gap-3 p-3 rounded-lg bg-[radial-gradient(circle_at_top_left,rgba(15,23,42,0.98),rgba(15,23,42,0.96))] shadow-[0_6px_18px_-10px_rgba(0,0,0,0.95)]"
                        >
                          <div class="w-16 h-16 flex-shrink-0 rounded-lg overflow-hidden bg-[linear-gradient(135deg,rgba(15,23,42,0.98),rgba(15,23,42,0.96))] shadow-[0_4px_12px_-6px_rgba(0,0,0,0.95)]">
                            <img
                              v-if="item.thumbnail"
                              :src="item.thumbnail"
                              :alt="item.title"
                              class="w-full h-full object-cover"
                            />
                          </div>
                          <div class="flex-1 min-w-0">
                            <p class="text-sm font-medium text-white truncate">{{ item.title }}</p>
                            <p class="text-xs text-white/60 mt-1">Qty: {{ item.quantity }}</p>
                            <p class="text-sm font-semibold text-white mt-1">
                              {{ formatPrice(item.price * item.quantity) }}
                            </p>
                          </div>
                        </div>
                      </div>

                      <!-- 费用明细 -->
                      <div class="space-y-2 pt-4 border-t border-white/10">
                        <div class="flex justify-between text-sm">
                          <span class="text-white/70">Subtotal</span>
                          <span class="font-medium text-white">{{ formatPrice(priceBreakdown.subtotal) }}</span>
                        </div>
                        
                        <div class="flex justify-between text-sm">
                          <span class="text-white/70">
                            Shipping
                            <span v-if="form.country && shippingValidation.matchedRule?.service_label" class="text-xs text-white/50">
                              ({{ shippingValidation.matchedRule.service_label }})
                            </span>
                          </span>
                          <span class="font-medium text-white">
                            <template v-if="!form.country">
                              <span class="text-white/50 text-xs">Select country</span>
                            </template>
                            <template v-else-if="!shippingValidation.isShippable">
                              <span class="text-red-400 text-xs">Unavailable</span>
                            </template>
                            <template v-else>
                              {{ regionShippingFee === 0 ? 'Free' : formatPrice(regionShippingFee) }}
                            </template>
                          </span>
                        </div>
                        <!-- 免运费进度 -->
                        <div
                          v-if="form.country && shippingValidation.isShippable && shippingValidation.matchedRule?.free_over && regionShippingFee > 0"
                          class="text-xs text-green-400 mt-1"
                        >
                          Add {{ formatPrice(shippingValidation.matchedRule.free_over - priceBreakdown.subtotal) }} more for free shipping!
                        </div>
                        <div class="flex justify-between text-sm">
                          <span class="text-white/70">Tax</span>
                          <span class="font-medium text-white">{{ formatPrice(priceBreakdown.tax) }}</span>
                        </div>
                        <div class="flex justify-between text-sm">
                          <span class="text-white/70">Payment method fee</span>
                          <span class="font-medium text-emerald-300">{{ formatPrice(0) }}</span>
                        </div>
                        <div class="flex justify-between text-sm">
                          <span class="text-white/70">Points discount</span>
                          <span class="font-medium text-emerald-300">- {{ formatPrice(0) }}</span>
                        </div>
                        <div class="flex justify-between text-sm">
                          <span class="text-white/70">Gift card discount</span>
                          <span class="font-medium text-emerald-300">- {{ formatPrice(0) }}</span>
                        </div>
                        <div class="flex justify-between text-lg font-bold pt-3 border-t border-white/20">
                          <span class="text-white">Total Amount</span>
                          <span class="text-[#6b73ff]">{{ formatPrice(priceBreakdown.total) }}</span>
                        </div>
                      </div>

                      <div class="mt-4 space-y-2">
                        <ChatStartButton
                          class="w-full text-sm"
                          :label="activePaymentTab === 'card' ? 'Confirm & Pay' : 'Pay now'"
                        />
                        <button
                          type="button"
                          class="w-full inline-flex items-center justify-center px-3 py-2 rounded-lg border border-white/10 text-xs font-medium text-white/70 hover:text-white hover:bg-white/5 transition-colors"
                        >
                          View cart
                        </button>
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
        </transition>
      </div>
    </transition>
  </teleport>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import ChatStartButton from '~/components/ChatStartButton.vue'
import { COUNTRIES } from '~/data/countries'
import { useShippingValidation } from '~/composables/useShippingValidation'

const {
  cartItems,
  isCheckoutOpen,
  priceBreakdown,
  closeCheckout,
  backToCart,
  formatPrice,
  clearCart,
  calculation,
} = useCart()

// 配送验证
const {
  loadShippingTemplates,
  validateShipping,
  getShippableCountries,
  getEstimatedDeliveryText,
  getZipFormatHint,
} = useShippingValidation()

// 初始化计算系统和配送模板
onMounted(async () => {
  calculation.initialize()
  await loadShippingTemplates()
})

// 支付方式列表（暂未直接用于 Tabs，但保留给后续逻辑使用）
const paymentMethods = [
  { id: 'credit_card', name: 'Credit Card', icon: '💳' },
  { id: 'paypal', name: 'PayPal', icon: '🅿️' },
  { id: 'alipay', name: 'Alipay', icon: '💙' },
  { id: 'wechat', name: 'WeChat Pay', icon: '💚' },
]

const activePaymentTab = ref<'card' | 'paypal' | 'alipay' | 'stripe' | 'bank' | 'worldfirst'>('card')

const setActivePaymentTab = (tab: 'card' | 'paypal' | 'alipay' | 'stripe' | 'bank' | 'worldfirst') => {
  activePaymentTab.value = tab
}

// 表单数据
const form = ref({
  country: '',
  name: '',
  phone: '',
  address: '',
  city: '',
  zip: '',
  paymentMethod: 'credit_card',
  notes: '',
})

const isSubmitting = ref(false)
const couponCode = ref('')
const isApplyingCoupon = ref(false)

// 配送验证结果
const shippingValidation = computed(() => {
  return validateShipping(form.value.country, form.value.zip)
})

// 可配送国家列表
const shippableCountryCodes = computed(() => getShippableCountries())

const shippableCountries = computed(() => {
  return COUNTRIES.filter(c => shippableCountryCodes.value.includes(c.code))
})

const nonShippableCountries = computed(() => {
  return COUNTRIES.filter(c => !shippableCountryCodes.value.includes(c.code))
})

// 预计送达时间
const estimatedDelivery = computed(() => {
  if (!shippingValidation.value.isShippable) return null
  return getEstimatedDeliveryText(shippingValidation.value.matchedRule)
})

// 邮编格式提示
const zipFormatHint = computed(() => {
  if (!form.value.country) return null
  return getZipFormatHint(form.value.country)
})

const zipPlaceholder = computed(() => {
  return zipFormatHint.value?.placeholder || 'Zip code'
})

const zipHint = computed(() => {
  return zipFormatHint.value?.hint || ''
})

// 联系客服
const openContactSupport = () => {
  window.open('/contact', '_blank')
}

// 货运代理服务
const openFreightForwarder = () => {
  window.open('/help/freight-forwarder', '_blank')
}

// 保存购物车稍后再买
const saveCartForLater = () => {
  // 可以保存到 localStorage 或用户账户
  alert('Your cart has been saved. We\'ll notify you when shipping becomes available to your region.')
}

// 基于地区的运费计算
const regionShippingFee = computed(() => {
  if (!form.value.country || !shippingValidation.value.isShippable) {
    return 0
  }
  
  const rule = shippingValidation.value.matchedRule
  if (!rule) return 0
  
  // 检查免运费门槛
  if (rule.free_over && priceBreakdown.value.subtotal >= rule.free_over) {
    return 0
  }
  
  return rule.fee || 0
})

// 应用优惠券
const handleApplyCoupon = async () => {
  if (!couponCode.value) return
  
  isApplyingCoupon.value = true
  const result = await calculation.applyCoupon(couponCode.value)
  isApplyingCoupon.value = false
  
  if (result.success) {
    alert('Coupon applied successfully!')
  } else {
    alert(result.message)
  }
}

// 表单验证
const isFormValid = computed(() => {
  return (
    form.value.country !== '' &&
    shippingValidation.value.isShippable &&
    form.value.name.trim() !== '' &&
    form.value.phone.trim() !== '' &&
    form.value.address.trim() !== '' &&
    form.value.city.trim() !== '' &&
    form.value.paymentMethod !== ''
  )
})

// 提交订单
const handleSubmit = async () => {
  if (!isFormValid.value || isSubmitting.value) return

  isSubmitting.value = true

  try {
    // 这里调用你的订单 API
    // const response = await $fetch('/wp-json/tanzanite/v1/orders', {
    //   method: 'POST',
    //   body: {
    //     items: cartItems.value,
    //     shipping: form.value,
    //     payment_method: form.value.paymentMethod,
    //     notes: form.value.notes,
    //     total: total.value,
    //   }
    // })

    // 模拟 API 调用
    await new Promise(resolve => setTimeout(resolve, 2000))

    // 成功后清空购物车
    clearCart()
    closeCheckout()

    // 显示成功消息
    alert('Order submitted successfully!')

    // 重置表单
    form.value = {
      country: '',
      name: '',
      phone: '',
      address: '',
      city: '',
      zip: '',
      paymentMethod: 'credit_card',
      notes: '',
    }
  } catch (error) {
    console.error('Order submission failed:', error)
    alert('Order submission failed, please try again')
  } finally {
    isSubmitting.value = false
  }
}
</script>

<style scoped>
/* 淡入淡出动画 */
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.3s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}

/* 缩放动画 */
.scale-enter-active,
.scale-leave-active {
  transition: all 0.3s ease;
}

.scale-enter-from,
.scale-leave-to {
  opacity: 0;
  transform: scale(0.95);
}

/* 自定义滚动条 */
.overflow-y-auto::-webkit-scrollbar {
  width: 6px;
}

.overflow-y-auto::-webkit-scrollbar-track {
  background: rgba(255, 255, 255, 0.1);
  border-radius: 10px;
}

.overflow-y-auto::-webkit-scrollbar-thumb {
  background: rgba(255, 255, 255, 0.3);
  border-radius: 10px;
}

.overflow-y-auto::-webkit-scrollbar-thumb:hover {
  background: rgba(255, 255, 255, 0.5);
}
</style>
