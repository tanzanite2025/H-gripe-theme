<template>
  <div class="flex flex-col gap-8 animate-in fade-in duration-700 min-h-screen bg-[#0a0a0a] text-white font-sans p-8">
    
    <!-- Header Module -->
    <header class="rounded-[32px] border border-dashed border-white/20 bg-muted/5 p-8 flex justify-between items-center relative overflow-hidden">
      <div class="absolute inset-0 bg-gradient-to-br from-emerald-500/5 via-transparent pointer-events-none"></div>
      <div>
        <h1 class="text-lg font-black tracking-tighter italic uppercase text-white">SYSTEM DASHBOARD</h1>
        <p class="text-[9px] font-black uppercase tracking-widest opacity-60 mt-1">Live Overview of ERP Metrics</p>
      </div>
      <div class="flex gap-4">
        <button @click="refreshAll" class="rounded-full h-11 px-8 bg-emerald-500 hover:bg-emerald-400 text-black font-black text-[10px] uppercase tracking-widest transition-colors shadow-[0_0_15px_rgba(16,185,129,0.3)]">
          REFRESH DATA
        </button>
      </div>
    </header>

    <div v-if="loading" class="flex-grow flex items-center justify-center text-emerald-500/50 text-[10px] font-mono animate-pulse">
      GATHERING_TELEMETRY...
    </div>

    <!-- Stats Grid -->
    <div v-else class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
      
      <!-- Orders Module -->
      <div class="rounded-[24px] border border-dashed border-white/20 bg-muted/5 p-6 relative flex flex-col group overflow-hidden">
        <div class="absolute inset-0 bg-gradient-to-br from-sky-500/5 via-transparent pointer-events-none group-hover:from-sky-500/10 transition-colors"></div>
        <h2 class="text-sm font-black tracking-tighter italic text-sky-400 mb-1">SALES & ORDERS</h2>
        <div class="text-[8px] font-mono text-white/40 mb-6">ALL-TIME METRICS</div>
        
        <div class="flex flex-col gap-4">
           <div>
             <div class="text-[10px] font-black uppercase tracking-widest text-white/50 mb-1">TOTAL REVENUE</div>
             <div class="text-3xl font-mono text-white">${{ formatNum(orderStats.total_revenue) }}</div>
           </div>
           <div class="grid grid-cols-2 gap-4">
             <div>
               <div class="text-[8px] font-black uppercase tracking-widest text-white/50 mb-1">TOTAL ORDERS</div>
               <div class="text-lg font-mono text-white/90">{{ orderStats.total_orders || 0 }}</div>
             </div>
             <div>
               <div class="text-[8px] font-black uppercase tracking-widest text-white/50 mb-1">PENDING</div>
               <div class="text-lg font-mono text-amber-500">{{ orderStats.pending_orders || 0 }}</div>
             </div>
           </div>
        </div>
      </div>

      <!-- Tickets Module -->
      <div class="rounded-[24px] border border-dashed border-white/20 bg-muted/5 p-6 relative flex flex-col group overflow-hidden">
        <div class="absolute inset-0 bg-gradient-to-br from-rose-500/5 via-transparent pointer-events-none group-hover:from-rose-500/10 transition-colors"></div>
        <h2 class="text-sm font-black tracking-tighter italic text-rose-400 mb-1">CUSTOMER SUPPORT</h2>
        <div class="text-[8px] font-mono text-white/40 mb-6">ACTIVE TICKETS & INQUIRIES</div>
        
        <div class="flex flex-col gap-4">
           <div>
             <div class="text-[10px] font-black uppercase tracking-widest text-white/50 mb-1">OPEN TICKETS</div>
             <div class="text-3xl font-mono text-rose-500">{{ ticketStats.open_tickets || 0 }}</div>
           </div>
           <div class="grid grid-cols-2 gap-4">
             <div>
               <div class="text-[8px] font-black uppercase tracking-widest text-white/50 mb-1">PENDING REPLY</div>
               <div class="text-lg font-mono text-white/90">{{ ticketStats.pending_reply || 0 }}</div>
             </div>
             <div>
               <div class="text-[8px] font-black uppercase tracking-widest text-white/50 mb-1">UNASSIGNED</div>
               <div class="text-lg font-mono text-amber-500">{{ ticketStats.unassigned || 0 }}</div>
             </div>
           </div>
        </div>
      </div>

      <!-- Warranties Module -->
      <div class="rounded-[24px] border border-dashed border-white/20 bg-muted/5 p-6 relative flex flex-col group overflow-hidden">
        <div class="absolute inset-0 bg-gradient-to-br from-emerald-500/5 via-transparent pointer-events-none group-hover:from-emerald-500/10 transition-colors"></div>
        <h2 class="text-sm font-black tracking-tighter italic text-emerald-400 mb-1">WARRANTIES</h2>
        <div class="text-[8px] font-mono text-white/40 mb-6">PRODUCT REGISTRATIONS</div>
        
        <div class="flex flex-col gap-4">
           <div>
             <div class="text-[10px] font-black uppercase tracking-widest text-white/50 mb-1">TOTAL REGISTERED</div>
             <div class="text-3xl font-mono text-emerald-500">{{ regStats.total_registrations || 0 }}</div>
           </div>
           <div class="grid grid-cols-2 gap-4">
             <div>
               <div class="text-[8px] font-black uppercase tracking-widest text-white/50 mb-1">ACTIVE</div>
               <div class="text-lg font-mono text-white/90">{{ regStats.active_warranties || 0 }}</div>
             </div>
             <div>
               <div class="text-[8px] font-black uppercase tracking-widest text-white/50 mb-1">PENDING REVIEW</div>
               <div class="text-lg font-mono text-amber-500">{{ regStats.pending_review || 0 }}</div>
             </div>
           </div>
        </div>
      </div>

      <!-- Quick Actions Module -->
      <div class="rounded-[24px] border border-dashed border-white/20 bg-black p-6 relative flex flex-col justify-between">
        <div>
          <h2 class="text-sm font-black tracking-tighter italic text-white mb-1">QUICK LINKS</h2>
          <div class="text-[8px] font-mono text-white/40 mb-6">SHORTCUTS TO MODULES</div>
        </div>
        
        <div class="flex flex-col gap-3">
          <router-link to="/orders" class="flex items-center justify-between p-3 rounded-xl bg-white/5 hover:bg-white/10 border border-transparent hover:border-white/10 transition-all group">
            <span class="text-[10px] font-black uppercase tracking-widest text-white/80 group-hover:text-sky-400">Process Orders</span>
            <span class="text-[10px] font-mono text-white/30">→</span>
          </router-link>
          
          <router-link to="/loyalty" class="flex items-center justify-between p-3 rounded-xl bg-white/5 hover:bg-white/10 border border-transparent hover:border-white/10 transition-all group">
            <span class="text-[10px] font-black uppercase tracking-widest text-white/80 group-hover:text-emerald-400">Member Adjustments</span>
            <span class="text-[10px] font-mono text-white/30">→</span>
          </router-link>

          <router-link to="/showcase" class="flex items-center justify-between p-3 rounded-xl bg-white/5 hover:bg-white/10 border border-transparent hover:border-white/10 transition-all group">
            <span class="text-[10px] font-black uppercase tracking-widest text-white/80 group-hover:text-amber-400">Approve Uploads</span>
            <span class="text-[10px] font-mono text-white/30">→</span>
          </router-link>
        </div>
      </div>

    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import http from '@/api/http'

const loading = ref(true)

const orderStats = ref({})
const ticketStats = ref({})
const regStats = ref({})

const fetchOrderStats = async () => {
  try {
    const data = await http('/orders/stats')
    orderStats.value = data || {}
  } catch (err) {
    console.warn("Could not load order stats", err)
  }
}

const fetchTicketStats = async () => {
  try {
    const data = await http('/tickets/stats')
    ticketStats.value = data || {}
  } catch (err) {
    console.warn("Could not load ticket stats", err)
  }
}

const fetchRegStats = async () => {
  try {
    const data = await http('/registrations/stats')
    regStats.value = data || {}
  } catch (err) {
    console.warn("Could not load registration stats", err)
  }
}

const refreshAll = async () => {
  loading.value = true
  await Promise.all([
    fetchOrderStats(),
    fetchTicketStats(),
    fetchRegStats()
  ])
  loading.value = false
}

const formatNum = (num) => {
  if (num == null) return "0.00"
  return Number(num).toFixed(2)
}

onMounted(() => {
  refreshAll()
})
</script>
