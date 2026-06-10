<script setup>
import { ref } from 'vue'
import LoyaltyManagement from './views/LoyaltyManagement.vue'
import CouponManagement from './views/CouponManagement.vue'
import PictureWarehouseApproval from './views/PictureWarehouseApproval.vue'
import FaqManagement from './views/FaqManagement.vue'

const currentView = ref('loyalty')
</script>

<template>
  <div class="min-h-screen bg-[#0a0a0a]">
    <!-- Simple Top Nav for Demo Purposes -->
    <div class="p-4 border-b border-white/10 flex gap-4 justify-center bg-black">
      <button 
        @click="currentView = 'loyalty'"
        :class="['px-6 py-2 rounded-full text-[10px] font-black uppercase tracking-widest transition-colors', currentView === 'loyalty' ? 'bg-white text-black' : 'text-white/50 hover:bg-white/10']">
        Loyalty & Points
      </button>
      <button 
        @click="currentView = 'coupons'"
        :class="['px-6 py-2 rounded-full text-[10px] font-black uppercase tracking-widest transition-colors', currentView === 'coupons' ? 'bg-white text-black' : 'text-white/50 hover:bg-white/10']">
        Coupons & Promos
      </button>
      <button 
        @click="currentView = 'showcase'"
        :class="['px-6 py-2 rounded-full text-[10px] font-black uppercase tracking-widest transition-colors', currentView === 'showcase' ? 'bg-white text-black' : 'text-white/50 hover:bg-white/10']">
        Picture Warehouse
      </button>
      <button 
        @click="currentView = 'faq'"
        :class="['px-6 py-2 rounded-full text-[10px] font-black uppercase tracking-widest transition-colors', currentView === 'faq' ? 'bg-white text-black' : 'text-white/50 hover:bg-white/10']">
        FAQ Management
      </button>
    </div>

    <LoyaltyManagement v-if="currentView === 'loyalty'" />
    <CouponManagement v-if="currentView === 'coupons'" />
    <PictureWarehouseApproval v-if="currentView === 'showcase'" />
    <FaqManagement v-if="currentView === 'faq'" />
  </div>
</template>
