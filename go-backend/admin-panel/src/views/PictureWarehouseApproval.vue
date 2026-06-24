<script setup>
import { ref, onMounted } from 'vue'
import http from '@/api/http'

const items = ref(null) // Deliberately null to trigger fast-fail if not loaded
const loading = ref(true)
const statusFilter = ref('pending')

// API fetching
const fetchItems = async () => {
  loading.value = true
  try {
    const json = await http(`/showcase?status=${statusFilter.value}`)
    items.value = json || []
  } catch (err) {
    alert(err.message) // loud fail
  } finally {
    loading.value = false
  }
}

// Approve item
const approveItem = async (id) => {
  // Optimistic update
  if (!items.value) throw new Error("[CRITICAL] Items list missing!")
  const original = items.value.find(i => i.id === id)
  if (!original) throw new Error(`[CRITICAL] Item ID ${id} not found in Index`)
  
  items.value = items.value.filter(i => i.id !== id)
  
  try {
    await http(`/showcase/${id}/approve`, {
      method: 'PUT'
    })
  } catch (err) {
    alert(err.message)
    // Revert optimistic
    items.value.push(original)
  }
}

// Reject item
const rejectItem = async (id) => {
  const reason = prompt("Enter rejection reason:")
  if (!reason) return
  
  if (!items.value) throw new Error("[CRITICAL] Items list missing!")
  const original = items.value.find(i => i.id === id)
  if (!original) throw new Error(`[CRITICAL] Item ID ${id} not found in Index`)
  
  items.value = items.value.filter(i => i.id !== id)

  try {
    await http(`/showcase/${id}/reject`, {
      method: 'PUT',
      body: JSON.stringify({ reason })
    })
  } catch (err) {
    alert(err.message)
    // Revert optimistic
    items.value.push(original)
  }
}

const changeFilter = (status) => {
  statusFilter.value = status
  fetchItems()
}

onMounted(() => {
  fetchItems()
})
</script>

<template>
  <div class="flex flex-col gap-8 animate-in fade-in duration-700 min-h-screen bg-[#0a0a0a] text-white font-sans p-8">
    
    <!-- Header Module -->
    <header class="rounded-[32px] border border-dashed border-white/20 bg-muted/5 p-8 flex justify-between items-center relative overflow-hidden">
      <div class="absolute inset-0 bg-gradient-to-br from-emerald-500/5 via-transparent pointer-events-none"></div>
      <div>
        <h1 class="text-lg font-black tracking-tighter italic uppercase text-white">PICTURE WAREHOUSE APPROVAL</h1>
        <p class="text-[9px] font-black uppercase tracking-widest opacity-60 mt-1">Review & Moderate User Uploaded Media</p>
      </div>
      <div class="flex bg-black/40 p-1 rounded-full border border-dashed border-white/20">
        <button @click="changeFilter('pending')" :class="['rounded-full h-11 px-6 font-black text-[10px] uppercase tracking-widest transition-colors', statusFilter === 'pending' ? 'bg-amber-500 text-black' : 'text-white/50 hover:text-white']">
          Pending
        </button>
        <button @click="changeFilter('approved')" :class="['rounded-full h-11 px-6 font-black text-[10px] uppercase tracking-widest transition-colors', statusFilter === 'approved' ? 'bg-emerald-500 text-black' : 'text-white/50 hover:text-white']">
          Approved
        </button>
        <button @click="changeFilter('rejected')" :class="['rounded-full h-11 px-6 font-black text-[10px] uppercase tracking-widest transition-colors', statusFilter === 'rejected' ? 'bg-rose-500 text-black' : 'text-white/50 hover:text-white']">
          Rejected
        </button>
      </div>
    </header>

    <div class="rounded-[24px] border border-dashed border-white/20 bg-muted/5 p-8 relative min-h-[500px]">
      <h2 class="text-sm font-black tracking-tighter italic text-white mb-8">SUBMISSION QUEUE</h2>
      
      <div v-if="loading" class="text-emerald-500/50 text-[10px] font-mono animate-pulse">SYNCING_DATA...</div>
      <div v-else-if="!items || items.length === 0" class="text-white/50 text-[10px] font-mono">NO_RECORDS_FOUND</div>
      
      <div v-else class="grid grid-cols-3 gap-8">
        <div v-for="item in items" :key="item.id" class="rounded-[24px] overflow-hidden border border-dashed border-white/10 bg-black/40 group relative flex flex-col">
          <div class="absolute inset-0 bg-gradient-to-br from-white/5 via-transparent pointer-events-none"></div>
          
          <!-- Image -->
          <div class="aspect-video bg-black relative">
            <img v-if="item.photos && item.photos.length > 0" :src="item.photos[0].url" class="w-full h-full object-cover opacity-80" />
            <div v-else class="absolute inset-0 flex items-center justify-center text-[8px] font-mono text-white/30">NO_MEDIA</div>
            
            <!-- Status Badge -->
            <div class="absolute top-4 right-4">
              <span v-if="item.status === 'pending'" class="inline-flex items-center px-3 h-5 rounded-full bg-amber-500/10 text-amber-500 text-[8px] font-mono uppercase border border-amber-500/20">PENDING_REVIEW</span>
              <span v-else-if="item.status === 'approved'" class="inline-flex items-center px-3 h-5 rounded-full bg-emerald-500/10 text-emerald-500 text-[8px] font-mono uppercase border border-emerald-500/20">APPROVED</span>
              <span v-else class="inline-flex items-center px-3 h-5 rounded-full bg-rose-500/10 text-rose-500 text-[8px] font-mono uppercase border border-rose-500/20 animate-pulse">REJECTED</span>
            </div>
          </div>
          
          <!-- Content -->
          <div class="p-6 flex flex-col gap-4 flex-grow relative z-10">
            <div class="flex justify-between items-start">
              <div>
                <h3 class="text-[10px] font-black uppercase tracking-widest text-emerald-400">{{ item.user_id ? `USER #${item.user_id}` : 'GUEST' }}</h3>
                <p class="text-[8px] font-mono text-white/40 mt-1">ID: {{ item.id }} | {{ new Date(item.created_at).toLocaleString() }}</p>
              </div>
            </div>
            
            <div class="grid grid-cols-2 gap-4 mt-2">
              <div class="bg-white/5 p-3 rounded-2xl border-none">
                <div class="text-[10px] font-black uppercase tracking-widest text-white/50 mb-1">BIKE MODEL</div>
                <div class="text-[10px] font-black tracking-widest uppercase text-white truncate">{{ item.bike_model || '—' }}</div>
              </div>
              <div class="bg-white/5 p-3 rounded-2xl border-none">
                <div class="text-[10px] font-black uppercase tracking-widest text-white/50 mb-1">LOCATION</div>
                <div class="text-[10px] font-black tracking-widest uppercase text-white truncate">{{ item.location || '—' }}</div>
              </div>
            </div>
            
            <div v-if="item.notes" class="bg-white/5 p-4 rounded-2xl border-none flex-grow mt-2">
               <div class="text-[10px] font-black uppercase tracking-widest text-white/50 mb-2">NOTES</div>
               <p class="text-xs text-white/80 leading-relaxed">{{ item.notes }}</p>
            </div>
            
            <div v-if="item.status === 'rejected' && item.admin_notes" class="bg-rose-500/10 p-4 rounded-2xl border border-rose-500/20 mt-2">
               <div class="text-[10px] font-black uppercase tracking-widest text-rose-500 mb-1">REJECTION REASON</div>
               <p class="text-xs text-rose-400">{{ item.admin_notes }}</p>
            </div>
          </div>
          
          <!-- Actions -->
          <div v-if="item.status === 'pending'" class="p-4 border-t border-dashed border-white/10 flex gap-4 bg-black/20 relative z-10">
            <button @click="approveItem(item.id)" class="flex-1 rounded-full h-11 bg-emerald-500 hover:bg-emerald-400 text-black font-black text-[10px] uppercase tracking-widest transition-colors">
              APPROVE
            </button>
            <button @click="rejectItem(item.id)" class="flex-1 rounded-full h-11 bg-rose-500/20 hover:bg-rose-500/30 text-rose-500 font-black text-[10px] uppercase tracking-widest transition-colors border border-rose-500/30">
              REJECT
            </button>
          </div>
          
          <div v-if="item.status === 'approved' || item.status === 'rejected'" class="p-4 border-t border-dashed border-white/10 flex justify-center bg-black/20 relative z-10">
             <button v-if="item.status === 'rejected'" @click="approveItem(item.id)" class="text-[10px] font-black uppercase tracking-widest text-white/50 hover:text-emerald-500 transition-colors">
               RE-APPROVE
             </button>
             <button v-if="item.status === 'approved'" @click="rejectItem(item.id)" class="text-[10px] font-black uppercase tracking-widest text-white/50 hover:text-rose-500 transition-colors">
               REVOKE APPROVAL
             </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
