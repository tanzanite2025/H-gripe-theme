<script setup>
import { ref, onMounted, computed } from 'vue'

const items = ref([])
const loading = ref(true)
const selectedPageId = ref('support-payment')

const showModal = ref(false)
const modalMode = ref('create') // 'create' or 'edit'
const currentFaq = ref({
  id: null,
  question: '',
  answer: '',
  page_id: '',
  category: '',
  order: 0,
  status: 'published'
})

// Hardcoded common page IDs for convenience, though they can type any
const commonPageIds = [
  'support-payment', 'support-shipping', 'support-warranty', 
  'support-warranty-check', 'support-product-feedback', 'support-test-report',
  'company-membership', 'guides-wheelset-buyers', 'guides-tireguides',
  'products-spoke-calculator', 'company-oem-odm', 'company-certificates',
  'company-contact', 'company-global-partners', 'company-ourstory'
]

const fetchFaqs = async () => {
  loading.value = true
  try {
    const res = await fetch(`http://localhost:8080/api/admin/faqs?page_size=100`, {
      headers: {
        'Authorization': `Bearer ${localStorage.getItem('token') || ''}`
      }
    })
    
    const text = await res.text()
    let json
    try {
      json = JSON.parse(text)
    } catch(e) {
      // Endpoint might not exist on admin yet. Wait, admin FAQ is GET /api/v1/content/faqs?
      // Ah, GET /api/v1/admin/faqs doesn't exist, we use /api/v1/content/faqs for GET
      const fallbackRes = await fetch(`http://localhost:8080/api/v1/content/faqs?page_id=${selectedPageId.value}&page_size=100`, {
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('token') || ''}`
        }
      })
      json = await fallbackRes.json()
      if (!fallbackRes.ok) throw new Error(json.error || "Failed to fetch")
    }

    // In case the response structure has data array
    items.value = json.data || []
  } catch (err) {
    console.error(err)
    alert("[CRITICAL] Failed to load FAQs: " + err.message)
  } finally {
    loading.value = false
  }
}

const changePageId = (pid) => {
  selectedPageId.value = pid
  fetchFaqs()
}

const openCreateModal = () => {
  modalMode.value = 'create'
  currentFaq.value = {
    id: null,
    question: '',
    answer: '',
    page_id: selectedPageId.value,
    category: '',
    order: 0,
    status: 'published'
  }
  showModal.value = true
}

const openEditModal = (faq) => {
  modalMode.value = 'edit'
  currentFaq.value = { ...faq }
  showModal.value = true
}

const saveFaq = async () => {
  if (!currentFaq.value.question || !currentFaq.value.answer || !currentFaq.value.page_id || !currentFaq.value.category) {
    alert("[CRITICAL] Missing required fields")
    return
  }

  const isCreate = modalMode.value === 'create'
  const url = isCreate 
    ? 'http://localhost:8080/api/admin/faqs' 
    : `http://localhost:8080/api/admin/faqs/${currentFaq.value.id}`
  const method = isCreate ? 'POST' : 'PUT'

  try {
    const res = await fetch(url, {
      method,
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${localStorage.getItem('token') || ''}`
      },
      body: JSON.stringify(currentFaq.value)
    })
    
    if (!res.ok) {
      const err = await res.json()
      throw new Error(err.error || "Save failed")
    }
    
    showModal.value = false
    fetchFaqs()
  } catch (err) {
    console.error(err)
    alert("[CRITICAL] Save failed: " + err.message)
  }
}

const deleteFaq = async (id) => {
  if (!confirm("Delete this FAQ permanently?")) return
  
  // Optimistic
  if (!items.value) throw new Error("[CRITICAL] Missing items data")
  const original = items.value.find(i => i.id === id)
  items.value = items.value.filter(i => i.id !== id)
  
  try {
    const res = await fetch(`http://localhost:8080/api/admin/faqs/${id}`, {
      method: 'DELETE',
      headers: {
        'Authorization': `Bearer ${localStorage.getItem('token') || ''}`
      }
    })
    if (!res.ok) throw new Error("Delete failed")
  } catch (err) {
    console.error(err)
    alert("[CRITICAL] " + err.message)
    items.value.push(original) // Revert
  }
}

onMounted(() => {
  fetchFaqs()
})
</script>

<template>
  <div class="flex flex-col gap-8 animate-in fade-in duration-700 min-h-screen bg-[#0a0a0a] text-white font-sans p-8">
    
    <!-- Header Module -->
    <header class="rounded-[32px] border border-dashed border-white/20 bg-muted/5 p-8 flex justify-between items-center relative overflow-hidden">
      <div class="absolute inset-0 bg-gradient-to-br from-emerald-500/5 via-transparent pointer-events-none"></div>
      <div>
        <h1 class="text-lg font-black tracking-tighter italic uppercase text-white">FAQ MANAGEMENT</h1>
        <p class="text-[9px] font-black uppercase tracking-widest opacity-60 mt-1">Manage Page-Specific Questions and Answers</p>
      </div>
      <div class="flex gap-4 items-center">
        <select v-model="selectedPageId" @change="fetchFaqs" class="h-11 rounded-2xl border-none bg-white/5 px-4 text-white text-[10px] font-bold uppercase tracking-widest focus:ring-1 focus:ring-emerald-500 outline-none">
          <option v-for="pid in commonPageIds" :key="pid" :value="pid" class="bg-[#111]">{{ pid }}</option>
        </select>
        <button @click="openCreateModal" class="rounded-full h-11 px-6 bg-emerald-500 hover:bg-emerald-400 text-black font-black text-[10px] uppercase tracking-widest transition-colors">
          + ADD FAQ
        </button>
      </div>
    </header>

    <div class="rounded-[24px] border border-dashed border-white/20 bg-muted/5 p-8 relative min-h-[500px]">
      <h2 class="text-sm font-black tracking-tighter italic text-white mb-8">
        FAQS FOR <span class="text-emerald-500">{{ selectedPageId }}</span>
      </h2>
      
      <div v-if="loading" class="text-emerald-500/50 text-[10px] font-mono animate-pulse">SYNCING_DATA...</div>
      <div v-else-if="items.length === 0" class="text-white/50 text-[10px] font-mono">NO_FAQS_FOR_THIS_PAGE</div>
      
      <div v-else class="flex flex-col gap-4">
        <div v-for="item in items" :key="item.id" class="rounded-2xl border border-dashed border-white/10 bg-black/40 p-5 group flex justify-between items-start transition-colors hover:bg-white/5 hover:border-white/20">
          <div class="flex-1 pr-8">
            <div class="flex items-center gap-3 mb-2">
              <span class="inline-flex px-2 py-1 rounded-md bg-emerald-500/10 text-emerald-500 text-[8px] font-mono uppercase border border-emerald-500/20">
                {{ item.category }}
              </span>
              <span v-if="item.status === 'draft'" class="inline-flex px-2 py-1 rounded-md bg-amber-500/10 text-amber-500 text-[8px] font-mono uppercase border border-amber-500/20">
                DRAFT
              </span>
              <span class="text-[8px] font-mono text-white/40">ORDER: {{ item.order }} | VIEWS: {{ item.view_count }}</span>
            </div>
            <h3 class="text-sm font-bold text-white mb-2">{{ item.question }}</h3>
            <div class="text-xs text-white/50 line-clamp-2" v-html="item.answer"></div>
          </div>
          <div class="flex flex-col gap-2 opacity-0 group-hover:opacity-100 transition-opacity">
            <button @click="openEditModal(item)" class="text-[10px] font-black uppercase tracking-widest px-4 py-2 rounded-full bg-white/10 hover:bg-white/20 text-white transition-colors">
              EDIT
            </button>
            <button @click="deleteFaq(item.id)" class="text-[10px] font-black uppercase tracking-widest px-4 py-2 rounded-full bg-rose-500/10 hover:bg-rose-500/20 text-rose-500 transition-colors">
              DEL
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- Modal -->
    <div v-if="showModal" class="fixed inset-0 z-50 flex items-center justify-center bg-black/80 backdrop-blur-sm p-4 animate-in fade-in">
      <div class="w-full max-w-2xl rounded-[32px] bg-[#111] shadow-2xl relative overflow-hidden flex flex-col border border-white/10">
        <div class="p-8 border-b border-dashed border-white/10">
          <h2 class="text-lg font-black tracking-tighter italic uppercase text-emerald-500">
            {{ modalMode === 'create' ? 'CREATE FAQ' : 'EDIT FAQ' }}
          </h2>
        </div>

        <div class="p-8 flex flex-col gap-6 max-h-[60vh] overflow-y-auto">
          <div class="grid grid-cols-2 gap-6">
            <div>
              <label class="block text-[10px] font-black uppercase tracking-widest text-white/50 mb-2">PAGE ID *</label>
              <input v-model="currentFaq.page_id" type="text" placeholder="e.g. support-payment" class="w-full h-12 rounded-2xl border-none bg-white/5 px-4 text-white text-sm font-mono focus:ring-1 focus:ring-emerald-500 outline-none">
            </div>
            <div>
              <label class="block text-[10px] font-black uppercase tracking-widest text-white/50 mb-2">CATEGORY *</label>
              <input v-model="currentFaq.category" type="text" placeholder="e.g. Payment Methods" class="w-full h-12 rounded-2xl border-none bg-white/5 px-4 text-white text-sm focus:ring-1 focus:ring-emerald-500 outline-none">
            </div>
          </div>
          
          <div>
            <label class="block text-[10px] font-black uppercase tracking-widest text-white/50 mb-2">QUESTION *</label>
            <input v-model="currentFaq.question" type="text" placeholder="Enter question..." class="w-full h-12 rounded-2xl border-none bg-white/5 px-4 text-white text-sm focus:ring-1 focus:ring-emerald-500 outline-none">
          </div>

          <div>
            <label class="block text-[10px] font-black uppercase tracking-widest text-white/50 mb-2">ANSWER (HTML SUPPORTED) *</label>
            <textarea v-model="currentFaq.answer" rows="6" placeholder="Enter answer..." class="w-full rounded-2xl border-none bg-white/5 p-4 text-white text-sm focus:ring-1 focus:ring-emerald-500 outline-none"></textarea>
          </div>

          <div class="grid grid-cols-2 gap-6">
            <div>
              <label class="block text-[10px] font-black uppercase tracking-widest text-white/50 mb-2">DISPLAY ORDER</label>
              <input v-model.number="currentFaq.order" type="number" class="w-full h-12 rounded-2xl border-none bg-white/5 px-4 text-white text-sm font-mono focus:ring-1 focus:ring-emerald-500 outline-none">
            </div>
            <div>
              <label class="block text-[10px] font-black uppercase tracking-widest text-white/50 mb-2">STATUS</label>
              <select v-model="currentFaq.status" class="w-full h-12 rounded-2xl border-none bg-white/5 px-4 text-white text-sm focus:ring-1 focus:ring-emerald-500 outline-none">
                <option value="published" class="bg-[#111]">Published</option>
                <option value="draft" class="bg-[#111]">Draft</option>
              </select>
            </div>
          </div>
        </div>

        <div class="p-8 border-t border-dashed border-white/10 flex justify-end gap-4 bg-black/20">
          <button @click="showModal = false" class="text-[10px] font-black uppercase tracking-widest px-6 h-11 rounded-full border border-dashed border-white/20 hover:bg-white/5 transition-colors">
            CANCEL
          </button>
          <button @click="saveFaq" class="rounded-full h-11 px-8 bg-emerald-500 hover:bg-emerald-400 text-black font-black text-[10px] uppercase tracking-widest transition-colors">
            SAVE FAQ
          </button>
        </div>
      </div>
    </div>

  </div>
</template>
