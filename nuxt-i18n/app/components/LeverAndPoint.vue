<template>
  <div class="flex justify-center pt-0 pb-0 w-full">
    <div class="w-full max-w-[1400px] h-[90vh] md:h-[700px] max-h-[85vh] rounded-2xl border-2 border-[#6b73ff] bg-black shadow-[0_0_30px_rgba(107,115,255,0.3)] relative overflow-hidden flex flex-col" role="region" aria-label="Membership Levels and Points">
      <button class="absolute right-2 top-2 w-7 h-7 inline-flex items-center justify-center border border-[rgba(124,117,255,0.6)] rounded-md bg-[rgba(30,27,75,0.6)] text-[#e8e9ff] pointer-events-auto hover:brightness-110 transition-all" type="button" @click="$emit('close')">×</button>
      <!-- 移动端标签页 -->
      <div class="md:hidden flex gap-2 justify-center py-3 border-b border-white/10 px-3 pointer-events-auto">
        <button
          @click="mobileTab = 'info'"
          class="h-10 rounded-full text-sm font-semibold flex-1 border transition-all"
          :class="mobileTab === 'info' 
            ? 'bg-[#6b73ff] border-[#6b73ff] text-white shadow-[0_6px_24px_rgba(107,115,255,0.35)]' 
            : 'bg-white/5 border-white/15 text-white/70'"
        >
          My Info
        </button>
        <button
          @click="mobileTab = 'levels'"
          class="h-10 rounded-full text-sm font-semibold flex-1 border transition-all"
          :class="mobileTab === 'levels' 
            ? 'bg-[#6b73ff] border-[#6b73ff] text-white shadow-[0_6px_24px_rgba(107,115,255,0.35)]' 
            : 'bg-white/5 border-white/15 text-white/70'"
        >
          Levels & Points
        </button>
      </div>
      
      <div class="flex-1 flex p-4 px-5 pointer-events-auto overflow-hidden box-border">
        <div class="grid grid-cols-2 max-md:grid-cols-1 gap-5 w-full overflow-y-auto md:overflow-hidden">
          <!-- 左侧：认证表单或会员信息 -->
          <div v-show="mobileTab === 'info' || !isMobile" class="flex flex-col items-start max-md:items-center text-left max-md:text-center gap-4 w-full md:overflow-y-auto md:h-full">
            <!-- 显示会员信息 -->
            <div class="flex flex-col items-start max-md:items-center text-left max-md:text-center gap-4 w-full">
              <!-- 头像 + 登录/注册按钮 -->
              <div class="flex justify-center w-full">
                <div class="flex items-center gap-4 w-full max-w-[360px] md:flex-col md:gap-3">
                  <div class="w-[54px] h-[54px] overflow-hidden bg-transparent rounded-xl leading-[0]">
                    <BadgeAvatar :logged="isLogged" :level="levelName" :topTierImageUrl="topTierImage" />
                  </div>
                  <div class="flex-1 w-full" v-if="!isLogged">
                    <div class="flex gap-2 justify-end md:justify-center pointer-events-auto">
                      <button
                        class="h-10 px-5 rounded-full inline-flex items-center justify-center bg-gradient-to-r from-[#40ffaa] to-[#6b73ff] text-black text-sm font-semibold hover:brightness-110 transition-all pointer-events-auto"
                        type="button"
                        @click="openAuthForm('register')"
                      >
                        {{ $t('user.register') }}
                      </button>
                      <button
                        class="h-10 px-5 rounded-full inline-flex items-center justify-center border border-white/20 bg-white/10 text-white text-sm font-semibold hover:bg-white/20 transition-all pointer-events-auto"
                        type="button"
                        @click="openAuthForm('login')"
                      >
                        {{ $t('user.login') }}
                      </button>
                    </div>
                  </div>
                  <div class="flex-1 w-full flex justify-end md:justify-center" v-else-if="isLogged">
                    <button class="h-10 px-5 rounded-full inline-flex items-center justify-center bg-red-600 hover:bg-red-700 text-white text-sm font-semibold transition-all" type="button" @click="doLogout">
                      {{ $t('user.logout') }}
                    </button>
                  </div>
                </div>
              </div>
              
              <!-- 会员信息容器 - 美化版 -->
              <div class="w-full border-2 border-[#6e6ee9] rounded-xl bg-gradient-to-br from-white/[0.05] to-white/[0.02] p-4 backdrop-blur-sm">
                <!-- 基础信息网格 -->
                <div class="grid grid-cols-2 gap-3 mb-3 pb-3 border-b border-white/10">
                  <div class="flex items-center gap-2 bg-white/[0.03] rounded-lg p-2 border border-white/5">
                    <span class="text-2xl leading-none">👤</span>
                    <div class="flex flex-col">
                      <span class="text-[11px] text-white/50">{{ $t('member.brief.membershipLevel', 'Level') }}</span>
                      <span class="text-sm font-semibold" :class="isLogged ? 'text-white/90' : 'text-[#40ffaa]'">{{ isLogged ? (levelName || '0') : '?' }}</span>
                    </div>
                  </div>
                  <div class="flex items-center gap-2 bg-white/[0.03] rounded-lg p-2 border border-white/5">
                    <span class="text-2xl leading-none">🛍️</span>
                    <div class="flex flex-col">
                      <span class="text-[11px] text-white/50">{{ $t('member.brief.productDiscount', 'Product') }}</span>
                      <span class="text-sm font-semibold" :class="isLogged ? 'text-white/90' : 'text-[#40ffaa]'">{{ isLogged ? (levelDiscounts.product + '%') : '?' }}</span>
                    </div>
                  </div>
                  <div class="flex items-center gap-2 bg-white/[0.03] rounded-lg p-2 border border-white/5">
                    <span class="text-2xl leading-none">💎</span>
                    <div class="flex flex-col">
                      <span class="text-[11px] text-white/50">{{ $t('member.brief.pointsDiscount', 'Points') }}</span>
                      <span class="text-sm font-semibold" :class="isLogged ? 'text-white/90' : 'text-[#40ffaa]'">{{ isLogged ? (levelDiscounts.points + '%') : '?' }}</span>
                    </div>
                  </div>
                  <div class="flex items-center gap-2 bg-white/[0.03] rounded-lg p-2 border border-white/5">
                    <span class="text-2xl leading-none">📊</span>
                    <div class="flex flex-col">
                      <span class="text-[11px] text-white/50">{{ $t('member.brief.stackable', 'Stackable') }}</span>
                      <span class="text-sm font-semibold" :class="isLogged ? 'text-white/90' : 'text-[#40ffaa]'">{{ isLogged ? (levelDiscounts.stackable ? '✓' : '✗') : '?' }}</span>
                    </div>
                  </div>
                </div>
                
                <!-- 优惠券和积分卡 -->
                <div class="grid grid-cols-2 gap-3">
                  <div class="flex items-center gap-2 bg-white/[0.03] rounded-lg p-2 border border-white/5">
                    <span class="text-2xl">🎟️</span>
                    <div class="flex flex-col flex-1">
                      <span class="text-[11px] text-white/50">{{ $t('member.coupons', 'Coupons') }}</span>
                      <span class="text-base font-bold text-transparent bg-clip-text bg-gradient-to-r from-[#40ffaa] to-[#6b73ff]">
                        <span :class="isLogged ? '' : 'text-[#40ffaa]'">{{ isLogged ? `× ${userCoupons}` : '?' }}</span>
                      </span>
                    </div>
                  </div>
                  <div class="flex items-center gap-2 bg-white/[0.03] rounded-lg p-2 border border-white/5">
                    <span class="text-2xl">💳</span>
                    <div class="flex flex-col flex-1">
                      <span class="text-[11px] text-white/50">{{ $t('member.pointCards', 'Point Cards') }}</span>
                      <span class="text-base font-bold text-transparent bg-clip-text bg-gradient-to-r from-[#40ffaa] to-[#6b73ff]">
                        <span :class="isLogged ? '' : 'text-[#40ffaa]'">{{ isLogged ? `× ${userPointCards}` : '?' }}</span>
                      </span>
                    </div>
                  </div>
                </div>
              </div>
              <div class="w-full space-y-2" v-if="profileInfo">
                <div class="flex justify-between items-center py-2 px-3 bg-white/5 rounded-lg" v-if="profileInfo.fullName">
                  <span class="text-sm text-white/70">{{ $t?.('profile.fullName') || 'Full Name' }}</span>
                  <span class="text-sm text-white/90">{{ profileInfo.fullName }}</span>
                </div>
                <div class="flex justify-between items-center py-2 px-3 bg-white/5 rounded-lg" v-if="profileInfo.company">
                  <span class="text-sm text-white/70">{{ $t?.('profile.company') || 'Company' }}</span>
                  <span class="text-sm text-white/90">{{ profileInfo.company }}</span>
                </div>
                <div class="flex justify-between items-center py-2 px-3 bg-white/5 rounded-lg" v-if="profileInfo.country">
                  <span class="text-sm text-white/70">{{ $t?.('profile.country') || 'Country/Region' }}</span>
                  <span class="text-sm text-white/90">{{ profileInfo.country }}</span>
                </div>
                <div class="flex justify-between items-center py-2 px-3 bg-white/5 rounded-lg" v-if="profileInfo.phone">
                  <span class="text-sm text-white/70">{{ $t?.('profile.phone') || 'Phone' }}</span>
                  <span class="text-sm text-white/90">{{ profileInfo.phone }}</span>
                </div>
                <div class="flex justify-between items-center py-2 px-3 bg-white/5 rounded-lg" v-if="profileInfo.marketingOptIn !== undefined">
                  <span class="text-sm text-white/70">{{ $t?.('profile.marketingOptIn') || 'Marketing Subscription' }}</span>
                  <span class="text-sm text-white/90">{{ profileInfo.marketingOptIn ? ($t?.('common.yes') || 'Yes') : ($t?.('common.no') || 'No') }}</span>
                </div>
                <div class="flex flex-col gap-1 py-2 px-3 bg-white/5 rounded-lg" v-if="profileInfo.notes">
                  <span class="text-sm text-white/70">{{ $t?.('profile.notes') || 'Notes' }}</span>
                  <span class="text-sm text-white/90">{{ profileInfo.notes }}</span>
                </div>
              </div>

              <div class="w-full" v-if="isLogged">
                <div class="relative w-full h-2 bg-white/10 rounded-full overflow-hidden">
                  <div class="absolute top-0 left-0 h-full bg-gradient-to-r from-[#40ffaa] to-[#6b73ff] transition-all duration-300" :style="{ width: tierInfo.pct + '%' }"></div>
                </div>
                <div class="flex justify-between items-center mt-1 text-xs text-white/70">
                  <span>{{ tierInfo.current ? tierInfo.current.min : 0 }}</span>
                  <span class="font-semibold text-white/90">{{ tierInfo.pct }}%</span>
                  <span>{{ tierInfo.next ? tierInfo.next.min : (tierInfo.current && tierInfo.current.max !== -1 ? tierInfo.current.max : 'MAX') }}</span>
                </div>
              </div>

              <!-- 礼品卡兑换区域 -->
              <div class="w-full">
                <div class="text-sm font-semibold text-white/90 mb-3 flex items-center gap-2">
                  <span>🎁</span>
                  <span>{{ $t('giftcards.title', 'Redeem points for gift cards') }}</span>
                </div>
                
                <!-- 礼品卡内容容器 - 添加最大高度和滚动 -->
                <div class="max-h-[300px] overflow-y-auto pr-2 custom-scrollbar">
                  <!-- 加载状态 -->
                  <div v-if="giftcardsLoading" class="text-center py-4 text-white/50 text-sm">
                    {{ $t('common.loading', 'Loading...') }}
                  </div>
                  
                  <!-- 错误状态 -->
                  <div v-else-if="giftcardsError" class="text-center py-4 text-red-400 text-sm">
                    {{ giftcardsError }}
                  </div>
                  
                  <!-- 礼品卡列表 -->
                  <div v-else-if="availableGiftcards.length > 0" class="grid grid-cols-1 gap-3">
                  <div 
                    v-for="card in availableGiftcards" 
                    :key="card.id"
                    class="relative border border-white/10 rounded-xl overflow-hidden hover:border-[#6b73ff]/50 transition-all"
                  >
                    <!-- 背景图片或默认渐变 -->
                    <div 
                      v-if="card.cover_image" 
                      class="absolute inset-0 bg-cover bg-center opacity-30"
                      :style="{ backgroundImage: `url(${card.cover_image})` }"
                    ></div>
                    <div 
                      v-else
                      class="absolute inset-0 bg-gradient-to-br from-white/[0.08] to-white/[0.03]"
                    ></div>
                    
                    <!-- 内容层 -->
                    <div class="relative z-10 p-3 backdrop-blur-sm bg-black/20">
                      <div class="flex items-center justify-between mb-2">
                        <div class="flex items-center gap-2">
                          <span class="text-2xl">💳</span>
                          <div>
                            <div class="text-sm font-semibold text-white/90">{{ card.card_code }}</div>
                            <div class="text-xs text-white/50">{{ $t('giftcards.balance', 'Balance') }}</div>
                          </div>
                        </div>
                        <div class="text-right">
                          <div class="text-lg font-bold text-transparent bg-clip-text bg-gradient-to-r from-[#40ffaa] to-[#6b73ff]">
                            ${{ card.balance }}
                          </div>
                        </div>
                      </div>
                    
                    <div class="flex items-center justify-between pt-2 border-t border-white/10">
                      <div class="text-xs text-white/70">
                        {{ $t('giftcards.pointsRequired', 'Points required') }}: 
                        <span class="font-semibold text-white/90">{{ card.points_spent || 0 }}</span>
                      </div>
                      <button 
                        @click="handleRedeemGiftcard(card)"
                        :disabled="(isLogged && points < (card.points_spent || 0)) || redeemingCardId === card.id"
                        class="px-3 py-1 text-xs font-semibold rounded-lg bg-gradient-to-r from-[#40ffaa] to-[#6b73ff] text-white hover:brightness-110 transition-all disabled:opacity-50 disabled:cursor-not-allowed"
                      >
                        {{ redeemingCardId === card.id ? $t('giftcards.redeeming', 'Redeeming...') : $t('giftcards.redeem', 'Redeem') }}
                      </button>
                    </div>
                    </div>
                  </div>
                </div>
                
                <!-- 无可用礼品卡 -->
                <div v-else class="text-center py-4 text-white/50 text-sm">
                  {{ $t('giftcards.noCards', 'No gift cards available') }}
                </div>
              </div>
              
              <!-- 兑换结果消息 - 放在滚动容器外面 -->
              <div v-if="redeemMessage" class="mt-3 p-2 rounded-lg text-sm text-center" :class="redeemSuccess ? 'bg-green-500/20 text-green-300' : 'bg-red-500/20 text-red-300'">
                {{ redeemMessage }}
              </div>
            </div>
          </div>
          </div>
          <div v-show="mobileTab === 'levels' || !isMobile" class="flex flex-col overflow-hidden w-full">
            <!-- Levels & Points -->
            <div class="flex-1 w-full py-2 px-3 box-border overflow-y-auto">
              <div class="text-sm font-semibold text-white/90 my-1.5 mb-2">{{ $t('member.levels.title', 'Membership levels') }}</div>
              <div class="flex flex-col gap-2.5 w-full">
                <div class="hidden md:grid grid-cols-[1.1fr_1fr_1fr_1fr] items-center py-2 px-3 border border-[rgba(110,110,233,0.35)] rounded-[10px] bg-[rgba(110,110,233,0.08)] font-semibold">
                  <div class="text-[13px] text-white/90">{{ $t('member.levels.header.level', 'Level') }}</div>
                  <div class="text-[13px] text-white/90">{{ $t('member.levels.header.pointsRequired', 'Points required') }}</div>
                  <div class="text-[13px] text-white/90">{{ $t('member.levels.header.productDiscount', 'Product discount') }}</div>
                  <div class="text-[13px] text-white/90">{{ $t('member.levels.header.pointsDiscount', 'Points discount') }}</div>
                </div>
                <!-- 动态渲染会员等级表格 -->
                <div
                  v-for="tier in tierConfigs"
                  :key="tier.key"
                  class="grid grid-cols-2 md:grid-cols-[1.1fr_1fr_1fr_1fr] gap-1.5 md:gap-0 items-center py-2 px-3 border border-white/10 rounded-[10px] bg-white/[0.04] odd:bg-white/[0.03]"
                >
                  <div class="text-[13px] text-white/90 md:before:content-none before:content-['Level'] before:block before:text-[11px] before:opacity-70">
                    {{ tier.name }}
                  </div>
                  <div class="text-[13px] text-white/90 md:before:content-none before:content-['Points_required'] before:block before:text-[11px] before:opacity-70">
                    {{ tier.min }}{{ tier.max !== null ? '–' + tier.max : '+' }}
                  </div>
                  <div class="text-[13px] text-white/90 md:before:content-none before:content-['Product'] before:block before:text-[11px] before:opacity-70">
                    {{ tier.discount }}%
                  </div>
                  <div class="text-[13px] text-white/90 md:before:content-none before:content-['Points'] before:block before:text-[11px] before:opacity-70">
                    {{ tier.pointsDiscount }}%
                  </div>
                </div>
              </div>
              <div class="text-sm font-semibold text-white/90 mt-4 mb-2">{{ $t('member.points.title', 'How to get points?') }}</div>
              <div class="flex flex-col gap-2.5">
                <div class="grid grid-cols-[1.2fr_2fr] max-[480px]:grid-cols-1 gap-2.5 items-center py-2 px-3 border border-white/10 rounded-[10px] bg-white/[0.04] odd:bg-white/[0.03]">
                  <div class="text-[13px] text-white/85 font-semibold">{{ $t('member.points.invite', 'Invite new users') }}</div>
                  <div class="text-[13px] text-white/90">{{ $t('member.points.inviteDesc', '50 Points (invitee gets 30 Points)') }}</div>
                </div>
                <!-- Copy Link 按钮单独一行 -->
                <div class="flex items-center gap-3 py-2 px-3 border border-white/10 rounded-[10px] bg-white/[0.04]">
                  <button 
                    class="h-10 px-[18px] rounded-full border border-white/[0.14] bg-gradient-to-r from-[#40ffaa] to-[#6b73ff] text-white text-sm font-bold hover:brightness-110 transition-all disabled:opacity-50 disabled:cursor-not-allowed flex-shrink-0" 
                    @click="handleCopyLink" 
                    :disabled="inviteLoading"
                  >
                    {{ inviteLoading ? '...' : 'Copy Link' }}
                  </button>
                  <div class="text-left text-[#cfd6ff] text-xs min-h-[16px] flex-1">{{ inviteMsg || '\u00A0' }}</div>
                </div>
                <div class="grid grid-cols-[1.2fr_2fr] max-[480px]:grid-cols-1 gap-2.5 items-center py-2 px-3 border border-white/10 rounded-[10px] bg-white/[0.04] odd:bg-white/[0.03]"><div class="text-[13px] text-white/85 font-semibold">{{ $t('member.points.consume', 'Consumption currency') }}</div><div class="text-[13px] text-white/90">{{ $t('member.points.consumeDesc', '1 Dollar = 1 Point') }}</div></div>
                <div class="grid grid-cols-[1.2fr_2fr] max-[480px]:grid-cols-1 gap-2.5 items-center py-2 px-3 border border-white/10 rounded-[10px] bg-white/[0.04] odd:bg-white/[0.03]"><div class="text-[13px] text-white/85 font-semibold">{{ $t('member.points.daily', 'Daily login') }}</div><div class="text-[13px] text-white/90">{{ $t('member.points.dailyDesc', '1 Point (30 days validity)') }}</div></div>
              </div>
            </div>
          </div>
        </div>
      </div>
      <div class="flex flex-col items-center justify-center py-3 pb-4 pointer-events-auto gap-3">
        <div class="flex flex-wrap gap-2 md:gap-3 items-center justify-center">
          <button class="h-10 px-[18px] rounded-full border border-[#6b73ff] bg-[#6b73ff] text-white text-sm font-bold pointer-events-auto hover:brightness-110 transition-all" @click="handleSelectProducts">Products</button>
          <button class="h-10 px-[18px] rounded-full border border-[#6b73ff] bg-[#6b73ff] text-white text-sm font-bold pointer-events-auto hover:brightness-110 transition-all" @click="handleViewCart">Cart</button>
          <button class="h-10 px-[18px] rounded-full border border-[#6b73ff] bg-[#6b73ff] text-white text-sm font-bold pointer-events-auto hover:brightness-110 transition-all" @click="handleFAQ">FAQ</button>
          <button class="h-10 px-[18px] rounded-full border border-[#6b73ff] bg-[#6b73ff] text-white text-sm font-bold pointer-events-auto hover:brightness-110 transition-all" @click="handleWishlist">Wishlist</button>
          <button
            class="h-10 px-[18px] rounded-full inline-flex items-center justify-center bg-gradient-to-r from-[#40ffaa] to-[#6b73ff] text-black text-sm font-semibold pointer-events-auto hover:brightness-110 transition-all"
            type="button"
            @click="handlePrivacy"
          >
            {{ $t('privacy.button', 'Privacy statement') }}
          </button>
        </div>
      </div>
    </div>
    
    <!-- FAQ Modal -->
    <Teleport to="body">
      <Transition
        enter-active-class="transition duration-300 ease-out"
        leave-active-class="transition duration-300 ease-in"
        enter-from-class="translate-y-full opacity-0"
        enter-to-class="translate-y-0 opacity-100"
        leave-from-class="translate-y-0 opacity-100"
        leave-to-class="translate-y-full opacity-0"
      >
        <div
          v-if="showFaqModal"
          class="fixed inset-0 z-[12000] flex items-end justify-center p-0 md:p-4 pointer-events-none"
        >
          <div class="pointer-events-none w-full max-w-[1400px] h-[90vh] md:h-[700px] max-h-[80vh] md:max-h-[85vh]">
            <FaqModal class="pointer-events-auto" @close="closeFAQ" />
          </div>
        </div>
      </Transition>
    </Teleport>

    <!-- Privacy Statement Modal -->
    <Teleport to="body">
      <Transition
        enter-active-class="transition duration-300 ease-out"
        leave-active-class="transition duration-300 ease-in"
        enter-from-class="translate-y-full opacity-0"
        enter-to-class="translate-y-0 opacity-100"
        leave-from-class="translate-y-0 opacity-100"
        leave-to-class="translate-y-full opacity-0"
      >
        <div
          v-if="showPrivacyModal"
          class="fixed inset-0 z-[12000] flex items-end justify-center p-0 md:p-4 pointer-events-none"
        >
          <div class="pointer-events-none w-full max-w-[1400px] h-[90vh] md:h-[700px] max-h-[80vh] md:max-h-[85vh]">
            <PrivacyStatementModal class="pointer-events-auto" @close="closePrivacy" />
          </div>
        </div>
      </Transition>
    </Teleport>

    <AuthModal
      v-model="showAuthModal"
      :default-mode="authMode"
      embedded
      @mode-change="handleAuthModeChange"
      @success="handleAuthSuccess"
    />

    <WishlistDrawer v-model="wishlistDrawerVisible" />
  </div>
</template>

<script setup>
import { ref, computed, onMounted, watch, onBeforeUnmount } from 'vue'
import { useI18n, useLocalePath } from '#imports'
import { useAuth } from '~/composables/useAuth'
import { useCart } from '~/composables/useCart'
import BadgeAvatar from '~/components/BadgeAvatar.vue'
import FaqModal from '~/components/FaqModal.vue'
import AuthModal from '~/components/AuthModal.vue'
import WishlistDrawer from '~/components/WishlistDrawer.vue'
import PrivacyStatementModal from '~/components/PrivacyStatementModal.vue'
import { setSidebarHandlesHidden } from '~/utils/sidebarHandles'

const emit = defineEmits(['close'])
const cart = useCart()
const { t: $t } = useI18n()
const auth = useAuth()
const localePath = useLocalePath()

// 移动端标签页状态
const mobileTab = ref('info') // 'info' or 'levels'
const isMobile = ref(false)

// 检测是否为移动端
if (typeof window !== 'undefined') {
  isMobile.value = window.innerWidth < 768
  window.addEventListener('resize', () => {
    isMobile.value = window.innerWidth < 768
  })
}

// 认证表单状态
const showAuthModal = ref(false)
const authMode = ref('login')
const showFaqModal = ref(false)
const showPrivacyModal = ref(false)
const wishlistDrawerVisible = ref(false)

const SIDEBAR_TOKEN_MODAL = 'lever-modal'
const SIDEBAR_TOKEN_AUTH = 'lever-auth'
const SIDEBAR_TOKEN_FAQ = 'lever-faq'
const SIDEBAR_TOKEN_PRIVACY = 'lever-privacy'

onMounted(() => {
  setSidebarHandlesHidden(SIDEBAR_TOKEN_MODAL, true)
})

watch(showAuthModal, (open) => {
  setSidebarHandlesHidden(SIDEBAR_TOKEN_AUTH, open)
}, { immediate: true })

watch(showFaqModal, (open) => {
  setSidebarHandlesHidden(SIDEBAR_TOKEN_FAQ, open)
}, { immediate: true })

watch(showPrivacyModal, (open) => {
  setSidebarHandlesHidden(SIDEBAR_TOKEN_PRIVACY, open)
}, { immediate: true })

onBeforeUnmount(() => {
  setSidebarHandlesHidden(SIDEBAR_TOKEN_MODAL, false)
  setSidebarHandlesHidden(SIDEBAR_TOKEN_AUTH, false)
  setSidebarHandlesHidden(SIDEBAR_TOKEN_FAQ, false)
  setSidebarHandlesHidden(SIDEBAR_TOKEN_PRIVACY, false)
})

// 用户优惠券和积分卡数量
const userCoupons = ref(0)
const userPointCards = ref(0)

// 礼品卡相关状态
const availableGiftcards = ref([])
const giftcardsLoading = ref(false)
const giftcardsError = ref('')
const redeemingCardId = ref(null)
const redeemMessage = ref('')
const redeemSuccess = ref(false)

const userData = computed(() => auth.user.value)
const isLogged = computed(() => !!userData.value)
const levelName = computed(() => userData.value?.loyalty?.level || '—')
const topTierImage = computed(() => userData.value?.loyalty?.top_tier_image || '')
const points = computed(() => userData.value?.loyalty?.points ?? 0)
const profileInfo = computed(() => userData.value?.profile || null)

const tiers = computed(() => userData.value?.loyalty?.tiers || [])
const tierInfo = computed(() => {
  const pts = points.value
  let current = null, next = null
  for (let i = 0; i < tiers.value.length; i++) {
    const t = tiers.value[i]
    const min = Number(t.min)
    const max = Number(t.max)
    const inRange = (max === -1) ? (pts >= min) : (pts >= min && pts <= max)
    if (inRange) { current = t; next = tiers.value[i + 1] || null; break }
  }
  if (!current && tiers.value.length) { current = tiers.value[0]; next = tiers.value[1] || null }
  let pct = 100
  if (current) {
    if (next && Number(next.min) > 0) {
      const start = Number(current.min); const end = Number(next.min)
      pct = Math.max(0, Math.min(100, Math.floor(((pts - start) / (end - start)) * 100)))
    } else if (Number(current.max) !== -1) {
      const start = Number(current.min); const end = Number(current.max)
      pct = Math.max(0, Math.min(100, Math.floor(((pts - start) / Math.max(1, end - start)) * 100)))
    } else { pct = 100 }
  }
  return { current, next, pct }
})

// 从后台获取会员等级配置
const tierConfigs = ref([])

const loadTierConfigs = async () => {
  try {
    const response = await $fetch('/wp-json/tanzanite/v1/loyalty/settings')
    if (response?.tiers) {
      tierConfigs.value = Object.entries(response.tiers).map(([key, config]) => ({
        key,
        name: config.name,
        min: config.min,
        max: config.max,
        discount: config.discount,
        pointsDiscount: config.points_discount || 0,
        stackable: config.stackable !== false
      }))
    }
  } catch (error) {
    console.error('Failed to load tier configs:', error)
  }
}

onMounted(() => {
  loadTierConfigs()
})

const levelDiscounts = computed(() => {
  const lvl = (levelName.value || '').toString().toLowerCase()
  if (!lvl || lvl === '—') return { product: 0, points: 0, stackable: false }
  
  const config = tierConfigs.value.find(t => t.key === lvl)
  if (config) {
    return {
      product: config.discount,
      points: config.pointsDiscount,
      stackable: config.stackable
    }
  }
  
  return { product: 0, points: 0, stackable: false }
})

const doLogout = async () => {
  try { 
    await auth.logout()
    showAuthModal.value = false
  } catch {}
}

// 打开认证表单
const openAuthForm = (mode) => {
  authMode.value = mode
  showAuthModal.value = true
}

const handleAuthModeChange = (mode) => {
  authMode.value = mode
}

const handleAuthSuccess = async () => {
  showAuthModal.value = false
  await auth.ensureSession()
  await fetchUserAssets()
}

// 获取用户优惠券和积分卡数据
const fetchUserAssets = async () => {
  if (!isLogged.value) {
    userCoupons.value = 0
    userPointCards.value = 0
    return
  }
  
  try {
    const base = window.location.origin
    const res = await fetch(`${base}/wp-json/mytheme/v1/user/assets`, {
      method: 'GET',
      credentials: 'include',
      headers: { 'Content-Type': 'application/json' }
    })
    
    if (res.ok) {
      const data = await res.json()
      if (data.success) {
        userCoupons.value = data.data?.coupons || 0
        userPointCards.value = data.data?.point_cards || 0
      }
    }
  } catch (error) {
    console.error('获取用户资产失败:', error)
  }
}

// 获取可兑换的礼品卡列表
const fetchAvailableGiftcards = async () => {
  giftcardsLoading.value = true
  giftcardsError.value = ''
  
  try {
    const base = window.location.origin
    const res = await fetch(`${base}/wp-json/tanzanite/v1/giftcards`, {
      method: 'GET',
      credentials: 'include',
      headers: { 'Content-Type': 'application/json' }
    })
    
    if (res.ok) {
      const data = await res.json()
      console.log('Gift cards API response:', data)
      
      // 显示所有状态为 active 且有积分价格的礼品卡（作为可兑换的模板）
      const allCards = data.items || data || []
      console.log('All cards:', allCards)
      
      // 显示所有 status 为 active 的卡片（不检查 points_spent，因为后台可能返回字符串 '0'）
      availableGiftcards.value = allCards.filter(card => 
        card.status === 'active'
      )
      console.log('Filtered cards count:', availableGiftcards.value.length)
      console.log('Filtered cards:', availableGiftcards.value)
    } else {
      console.error('Failed to fetch gift cards, status:', res.status)
      giftcardsError.value = 'Failed to load gift cards'
    }
  } catch (error) {
    console.error('Failed to fetch gift cards:', error)
    giftcardsError.value = 'Network error'
  } finally {
    giftcardsLoading.value = false
  }
}

// 兑换礼品卡
const handleRedeemGiftcard = async (card) => {
  if (redeemingCardId.value) return
  
  // 检查是否登录
  if (!isLogged.value) {
    redeemSuccess.value = false
    redeemMessage.value = 'Please login to redeem gift cards'
    setTimeout(() => {
      redeemMessage.value = ''
    }, 3000)
    return
  }
  
  redeemingCardId.value = card.id
  redeemMessage.value = ''
  redeemSuccess.value = false
  
  try {
    const base = window.location.origin
    const res = await fetch(`${base}/wp-json/tanzanite/v1/redeem/exchange`, {
      method: 'POST',
      credentials: 'include',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        points_to_spend: card.points_spent,
        giftcard_value: parseFloat(card.balance)
      })
    })
    
    const data = await res.json()
    
    if (res.ok && data.success) {
      redeemSuccess.value = true
      redeemMessage.value = `Redeemed successfully! Card code: ${data.card_code}`
      
      // Refresh user data and gift card list
      await auth.ensureSession()
      await fetchAvailableGiftcards()
      await fetchUserAssets()
      
      // 3秒后清除消息
      setTimeout(() => {
        redeemMessage.value = ''
      }, 3000)
    } else {
      redeemSuccess.value = false
      redeemMessage.value = data.message || 'Redemption failed'
    }
  } catch (error) {
    console.error('Failed to redeem gift card:', error)
    redeemSuccess.value = false
    redeemMessage.value = 'Network error, please try again later'
  } finally {
    redeemingCardId.value = null
  }
}

onMounted(() => { 
  auth.ensureSession()
  fetchUserAssets()
  fetchAvailableGiftcards()
})

// copy link (migrated from dock share)
const inviteLoading = ref(false)
const inviteMsg = ref('')
const handleCopyLink = async () => {
  try {
    inviteLoading.value = true
    inviteMsg.value = ''
    const base = window.location.origin
    const res = await fetch(`${base}/wp-json/tanzanite/v1/loyalty/referral/generate`, {
      method: 'POST',
      credentials: 'include',
      headers: { 'Content-Type': 'application/json' }
    })
    const data = await res.json()
    if (!res.ok) throw new Error((data && data.message) || 'Failed to generate referral link')
    const url = String(data && data.url)
    if (navigator.share) {
      try { await navigator.share({ url }) } catch {}
    }
    await navigator.clipboard.writeText(url)
    inviteMsg.value = 'Invitation link copied'
  } catch (e) {
    inviteMsg.value = String(e instanceof Error ? e.message : 'Failed to generate referral link')
  } finally {
    inviteLoading.value = false
    setTimeout(() => { inviteMsg.value = '' }, 15000)
  }
}

// Products - 在新标签页打开 Shop 页面
const handleSelectProducts = () => {
  try {
    const target = localePath('/shop')
    if (typeof window !== 'undefined' && target) {
      window.open(String(target), '_blank')
    }
  } catch (e) {
    console.error('Failed to open shop page:', e)
  }
}

// Cart - 打开购物车弹窗（不关闭当前 LeverAndPoint）
const handleViewCart = () => {
  cart.openCart()
}

// FAQ - 常见问题
const handleFAQ = () => {
  showFaqModal.value = true
}

const closeFAQ = () => {
  showFaqModal.value = false
}

// Privacy statement
const handlePrivacy = () => {
  showPrivacyModal.value = true
}

const closePrivacy = () => {
  showPrivacyModal.value = false
}

// Wishlist - 心愿单抽屉
const handleWishlist = () => {
  wishlistDrawerVisible.value = true
}
</script>

<style scoped>
/* 自定义滚动条样式 */
.custom-scrollbar::-webkit-scrollbar {
  width: 6px;
}

.custom-scrollbar::-webkit-scrollbar-track {
  background: rgba(255, 255, 255, 0.05);
  border-radius: 3px;
}

.custom-scrollbar::-webkit-scrollbar-thumb {
  background: rgba(107, 115, 255, 0.5);
  border-radius: 3px;
}

.custom-scrollbar::-webkit-scrollbar-thumb:hover {
  background: rgba(107, 115, 255, 0.7);
}

/* Firefox 滚动条样式 */
.custom-scrollbar {
  scrollbar-width: thin;
  scrollbar-color: rgba(107, 115, 255, 0.5) rgba(255, 255, 255, 0.05);
}
</style>
