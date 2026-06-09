<script setup>
import { ref, onMounted } from 'vue'

const coupons = ref([])
const loading = ref(true)
const showCreateModal = ref(false)

const newCoupon = ref({
  code: '',
  type: 'percentage',
  value: 0,
  description: '',
  min_amount: 0,
  max_discount: 0,
  usage_limit: 0,
  start_date: '',
  end_date: '',
  enabled: true
})

// Load coupons from the backend
const fetchCoupons = async () => {
  try {
    const res = await fetch('http://localhost:8080/api/v1/admin/marketing/coupons/all', {
      headers: {
        'Authorization': `Bearer ${localStorage.getItem('token') || ''}`
      }
    })
    const json = await res.json()
    if (res.ok) {
      coupons.value = json.data || []
    } else {
      console.error("[CRITICAL] Failed to fetch coupons", json.error)
    }
  } catch (err) {
    console.error("[CRITICAL] API Error", err)
  } finally {
    loading.value = false
  }
}

// Create a new coupon
const createCoupon = async () => {
  try {
    // Format dates to RFC3339 if not empty
    const payload = { ...newCoupon.value }
    if (payload.start_date) payload.start_date = new Date(payload.start_date).toISOString()
    else payload.start_date = new Date().toISOString()
    
    if (payload.end_date) payload.end_date = new Date(payload.end_date).toISOString()
    else {
      let nextYear = new Date()
      nextYear.setFullYear(nextYear.getFullYear() + 1)
      payload.end_date = nextYear.toISOString()
    }

    payload.value = Number(payload.value)
    payload.min_amount = Number(payload.min_amount)
    payload.max_discount = Number(payload.max_discount)
    payload.usage_limit = Number(payload.usage_limit)

    const res = await fetch('http://localhost:8080/api/v1/admin/marketing/coupons', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${localStorage.getItem('token') || ''}`
      },
      body: JSON.stringify(payload)
    })
    if (res.ok) {
      showCreateModal.value = false
      fetchCoupons()
    } else {
      const err = await res.json()
      console.error("[CRITICAL] Failed to create coupon", err.error)
      alert("Error: " + err.error)
    }
  } catch (err) {
    console.error("[CRITICAL] API Error", err)
  }
}

// Toggle enabled status
const toggleCouponStatus = async (coupon) => {
  const updatedCoupon = { ...coupon, enabled: !coupon.enabled }
  try {
    const res = await fetch(`http://localhost:8080/api/v1/admin/marketing/coupons/${coupon.id}`, {
      method: 'PUT',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${localStorage.getItem('token') || ''}`
      },
      body: JSON.stringify(updatedCoupon)
    })
    if (res.ok) {
      coupon.enabled = updatedCoupon.enabled
    } else {
      console.error("[CRITICAL] Failed to update coupon status")
    }
  } catch (err) {
    console.error("[CRITICAL] API Error", err)
  }
}

// Delete coupon
const deleteCoupon = async (id) => {
  if (!confirm("Are you sure you want to delete this coupon?")) return
  try {
    const res = await fetch(`http://localhost:8080/api/v1/admin/marketing/coupons/${id}`, {
      method: 'DELETE',
      headers: {
        'Authorization': `Bearer ${localStorage.getItem('token') || ''}`
      }
    })
    if (res.ok) {
      coupons.value = coupons.value.filter(c => c.id !== id)
    } else {
      console.error("[CRITICAL] Failed to delete coupon")
    }
  } catch (err) {
    console.error("[CRITICAL] API Error", err)
  }
}

const formatDate = (dateStr) => {
  if (!dateStr) return '-'
  return new Date(dateStr).toLocaleDateString()
}

onMounted(() => {
  fetchCoupons()
})
</script>

<template>
  <div class="flex flex-col gap-8 animate-in fade-in duration-700 min-h-screen p-8 bg-[#0a0a0a] text-white font-sans">
    
    <!-- Header Module -->
    <header class="rounded-[32px] border border-dashed border-white/20 bg-muted/5 p-8 flex justify-between items-center relative overflow-hidden">
      <div class="absolute inset-0 bg-gradient-to-br from-primary/5 via-transparent pointer-events-none"></div>
      <div>
        <h1 class="text-lg font-black tracking-tighter italic uppercase text-white">COUPONS & PROMOTIONS</h1>
        <p class="text-[9px] font-black uppercase tracking-widest opacity-60 mt-1">Manage Discount Codes and Campaigns</p>
      </div>
      <button @click="showCreateModal = true" class="rounded-full h-11 px-6 bg-emerald-500 hover:bg-emerald-400 text-black font-black text-[10px] uppercase tracking-widest transition-colors">
        + Create Coupon
      </button>
    </header>

    <!-- Coupons List -->
    <div class="rounded-[24px] border border-dashed border-white/20 bg-muted/5 p-6 relative">
      <h2 class="text-sm font-black tracking-tighter italic text-white mb-6">ACTIVE & ARCHIVED COUPONS</h2>
      
      <div v-if="loading" class="text-white/50 text-[10px] font-mono animate-pulse">LOADING_DATA...</div>
      
      <div v-else-if="coupons.length === 0" class="text-white/50 text-[10px] font-mono">NO_COUPONS_FOUND</div>
      
      <div v-else class="flex flex-col gap-4">
        <div v-for="coupon in coupons" :key="coupon.id" 
             class="flex items-center justify-between p-5 rounded-2xl border border-dashed border-white/10 bg-black/20 hover:bg-white/5 transition-colors group">
          
          <div class="flex items-center gap-6">
            <!-- Status Indicator -->
            <div :class="[
              'w-2 h-2 rounded-full shadow-lg',
              coupon.enabled ? 'bg-emerald-500 shadow-emerald-500/50' : 'bg-rose-500 shadow-rose-500/50'
            ]"></div>
            
            <div>
              <div class="flex items-center gap-3">
                <span class="text-sm font-black tracking-tighter italic uppercase text-white">{{ coupon.code }}</span>
                <span class="px-2 py-0.5 rounded-full text-[8px] font-mono border border-white/10 text-white/70">
                  {{ coupon.type.toUpperCase() }}
                </span>
              </div>
              <div class="text-[9px] font-black uppercase tracking-widest opacity-60 mt-1">
                Value: <span class="text-emerald-400">{{ coupon.type === 'percentage' ? coupon.value + '%' : '$' + coupon.value }}</span> 
                | Min Spend: ${{ coupon.min_amount }}
                | Uses: {{ coupon.used_count }} / {{ coupon.usage_limit > 0 ? coupon.usage_limit : '∞' }}
              </div>
              <div class="text-[8px] font-mono text-white/40 mt-1">
                VALID: {{ formatDate(coupon.start_date) }} - {{ formatDate(coupon.end_date) }}
              </div>
            </div>
          </div>

          <div class="flex items-center gap-3 opacity-0 group-hover:opacity-100 transition-opacity">
            <button @click="toggleCouponStatus(coupon)" 
                    class="text-[10px] font-black uppercase tracking-widest px-4 py-2 rounded-full border border-dashed border-white/20 hover:bg-white/10 transition-colors">
              {{ coupon.enabled ? 'DISABLE' : 'ENABLE' }}
            </button>
            <button @click="deleteCoupon(coupon.id)" 
                    class="text-[10px] font-black uppercase tracking-widest px-4 py-2 rounded-full bg-rose-500/10 text-rose-500 hover:bg-rose-500/20 transition-colors">
              DELETE
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- Create Modal -->
    <div v-if="showCreateModal" class="fixed inset-0 z-50 flex items-center justify-center bg-black/80 backdrop-blur-sm p-4 animate-in fade-in">
      <div class="w-full max-w-xl rounded-[32px] bg-[#111] shadow-2xl relative overflow-hidden flex flex-col">
        <div class="absolute inset-0 bg-gradient-to-br from-primary/5 via-transparent pointer-events-none"></div>
        
        <div class="p-8 border-b border-dashed border-white/10">
          <h2 class="text-lg font-black tracking-tighter italic uppercase text-white">ISSUE NEW COUPON</h2>
          <p class="text-[9px] font-black uppercase tracking-widest opacity-60 mt-1">Configure limits and values</p>
        </div>

        <div class="p-8 flex flex-col gap-6 overflow-y-auto max-h-[60vh]">
          <!-- Form Group -->
          <div>
            <label class="block text-[10px] font-black uppercase tracking-widest text-white/50 mb-2">Coupon Code</label>
            <input v-model="newCoupon.code" type="text" placeholder="e.g. SUMMER20" 
                   class="w-full h-12 rounded-2xl border-none bg-white/5 px-4 text-white font-mono uppercase focus:ring-1 focus:ring-emerald-500 outline-none">
          </div>
          
          <div class="grid grid-cols-2 gap-6">
            <div>
              <label class="block text-[10px] font-black uppercase tracking-widest text-white/50 mb-2">Discount Type</label>
              <select v-model="newCoupon.type" class="w-full h-12 rounded-2xl border-none bg-white/5 px-4 text-white text-sm focus:ring-1 focus:ring-emerald-500 outline-none">
                <option value="percentage" class="bg-[#111]">Percentage (%)</option>
                <option value="fixed" class="bg-[#111]">Fixed Amount ($)</option>
              </select>
            </div>
            <div>
              <label class="block text-[10px] font-black uppercase tracking-widest text-white/50 mb-2">Discount Value</label>
              <input v-model="newCoupon.value" type="number" 
                     class="w-full h-12 rounded-2xl border-none bg-white/5 px-4 text-white text-sm focus:ring-1 focus:ring-emerald-500 outline-none">
            </div>
          </div>

          <div class="grid grid-cols-2 gap-6">
            <div>
              <label class="block text-[10px] font-black uppercase tracking-widest text-white/50 mb-2">Minimum Spend ($)</label>
              <input v-model="newCoupon.min_amount" type="number" 
                     class="w-full h-12 rounded-2xl border-none bg-white/5 px-4 text-white text-sm focus:ring-1 focus:ring-emerald-500 outline-none">
            </div>
            <div>
              <label class="block text-[10px] font-black uppercase tracking-widest text-white/50 mb-2">Usage Limit (0=∞)</label>
              <input v-model="newCoupon.usage_limit" type="number" 
                     class="w-full h-12 rounded-2xl border-none bg-white/5 px-4 text-white text-sm focus:ring-1 focus:ring-emerald-500 outline-none">
            </div>
          </div>
          
          <div class="grid grid-cols-2 gap-6">
            <div>
              <label class="block text-[10px] font-black uppercase tracking-widest text-white/50 mb-2">Start Date (Optional)</label>
              <input v-model="newCoupon.start_date" type="date" 
                     class="w-full h-12 rounded-2xl border-none bg-white/5 px-4 text-white/70 text-sm focus:ring-1 focus:ring-emerald-500 outline-none">
            </div>
            <div>
              <label class="block text-[10px] font-black uppercase tracking-widest text-white/50 mb-2">End Date (Optional)</label>
              <input v-model="newCoupon.end_date" type="date" 
                     class="w-full h-12 rounded-2xl border-none bg-white/5 px-4 text-white/70 text-sm focus:ring-1 focus:ring-emerald-500 outline-none">
            </div>
          </div>
        </div>

        <div class="p-8 border-t border-dashed border-white/10 flex justify-end gap-4">
          <button @click="showCreateModal = false" 
                  class="text-[10px] font-black uppercase tracking-widest px-6 h-11 rounded-full border border-dashed border-white/20 hover:bg-white/5 transition-colors">
            CANCEL
          </button>
          <button @click="createCoupon" 
                  class="rounded-full h-11 px-8 bg-emerald-500 hover:bg-emerald-400 text-black font-black text-[10px] uppercase tracking-widest transition-colors">
            CREATE & ISSUE
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
/* Chrome, Safari, Edge, Opera */
input::-webkit-outer-spin-button,
input::-webkit-inner-spin-button {
  -webkit-appearance: none;
  margin: 0;
}

/* Firefox */
input[type=number] {
  -moz-appearance: textfield;
}
</style>
