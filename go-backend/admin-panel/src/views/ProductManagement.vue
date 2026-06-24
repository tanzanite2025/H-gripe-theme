<template>
  <div class="flex flex-col gap-8 animate-in fade-in duration-700 min-h-screen bg-[#0a0a0a] text-white font-sans p-8">
    
    <!-- Header Module -->
    <header class="rounded-[32px] border border-dashed border-white/20 bg-muted/5 p-8 flex justify-between items-center relative overflow-hidden">
      <div class="absolute inset-0 bg-gradient-to-br from-indigo-500/5 via-transparent pointer-events-none"></div>
      <div>
        <h1 class="text-lg font-black tracking-tighter italic uppercase text-white">PRODUCT MANAGEMENT</h1>
        <p class="text-[9px] font-black uppercase tracking-widest opacity-60 mt-1">Catalog & Inventory Control</p>
      </div>
      <div class="flex gap-4">
        <input v-model="filters.search" @keyup.enter="fetchProducts" placeholder="Search SKU or Name" class="h-11 rounded-full border-none bg-white/5 px-6 text-white text-[10px] font-mono uppercase focus:ring-1 focus:ring-indigo-500 outline-none w-64" />
        <button @click="fetchProducts" class="rounded-full h-11 px-8 bg-indigo-500 hover:bg-indigo-400 text-black font-black text-[10px] uppercase tracking-widest transition-colors shadow-[0_0_15px_rgba(99,102,241,0.3)]">
          SEARCH
        </button>
      </div>
    </header>

    <!-- Filters & List -->
    <div class="rounded-[24px] border border-dashed border-white/20 bg-muted/5 p-8 relative flex-grow flex flex-col">
      <div class="flex justify-between items-center mb-8">
        <h2 class="text-sm font-black tracking-tighter italic text-white">CATALOG DATABASE</h2>
        
        <div class="flex gap-2">
          <select v-model="filters.status" @change="fetchProducts" class="h-9 rounded-full border-none bg-white/5 px-4 text-white text-[10px] font-bold uppercase tracking-widest outline-none">
            <option value="" class="bg-[#111]">ALL STATUS</option>
            <option value="active" class="bg-[#111]">ACTIVE</option>
            <option value="inactive" class="bg-[#111]">INACTIVE</option>
            <option value="out_of_stock" class="bg-[#111]">OUT OF STOCK</option>
          </select>
        </div>
      </div>
      
      <div v-if="loading" class="text-indigo-500/50 text-[10px] font-mono animate-pulse flex-grow flex items-center justify-center">SYNCING_CATALOG...</div>
      <div v-else-if="products.length === 0" class="text-white/50 text-[10px] font-mono flex-grow flex items-center justify-center">NO_PRODUCTS_FOUND</div>
      
      <div v-else class="flex flex-col gap-2">
        <!-- Table Header -->
        <div class="grid grid-cols-12 gap-4 px-6 py-3 border-b border-dashed border-white/10 text-[10px] font-black uppercase tracking-widest text-white/50">
          <div class="col-span-1">ID</div>
          <div class="col-span-4">Product Name</div>
          <div class="col-span-2">SKU</div>
          <div class="col-span-1 text-right">Price</div>
          <div class="col-span-1 text-center">Stock</div>
          <div class="col-span-1 text-center">Status</div>
          <div class="col-span-2 text-right">Actions</div>
        </div>
        
        <!-- Rows -->
        <div v-for="product in products" :key="product.id" class="grid grid-cols-12 gap-4 px-6 py-4 items-center rounded-2xl border border-dashed border-transparent hover:border-white/10 hover:bg-white/5 transition-colors group">
          <div class="col-span-1 text-xs font-mono text-white/50">{{ product.id }}</div>
          <div class="col-span-4">
            <div class="text-sm font-black tracking-tighter italic text-white">{{ product.name || product.title }}</div>
            <div class="text-[9px] font-mono text-white/40 mt-1 uppercase truncate">{{ product.category_name || 'UNCATEGORIZED' }}</div>
          </div>
          <div class="col-span-2 text-xs text-white/80 font-mono truncate">{{ product.sku }}</div>
          <div class="col-span-1 text-[10px] font-mono text-emerald-400 text-right">${{ formatPrice(product.price) }}</div>
          <div class="col-span-1 text-[10px] font-mono text-white/80 text-center">{{ product.stock || 0 }}</div>
          
          <div class="col-span-1 flex justify-center">
             <span :class="getStatusClass(product.status)" class="px-2 py-1 rounded-md text-[8px] font-mono uppercase border w-24 text-center truncate">
              {{ product.status }}
            </span>
          </div>
          
          <div class="col-span-2 text-right flex gap-3 justify-end opacity-0 group-hover:opacity-100 transition-opacity">
            <button v-if="product.status !== 'active'" class="text-[10px] font-black uppercase tracking-widest text-emerald-500 hover:text-emerald-400 transition-colors" @click="updateStatus(product.id, 'active')">
              PUBLISH
            </button>
            <button v-if="product.status === 'active'" class="text-[10px] font-black uppercase tracking-widest text-amber-500 hover:text-amber-400 transition-colors" @click="updateStatus(product.id, 'inactive')">
              UNPUBLISH
            </button>
          </div>
        </div>
      </div>
      
      <!-- Pagination -->
      <div v-if="totalPages > 1" class="mt-8 flex justify-center gap-2 border-t border-dashed border-white/10 pt-6">
        <button v-for="p in totalPages" :key="p" @click="changePage(p)" :class="['w-8 h-8 rounded-full text-[10px] font-mono flex items-center justify-center transition-colors', p === filters.page ? 'bg-indigo-500 text-black font-bold' : 'bg-white/5 text-white/50 hover:bg-white/10']">
          {{ p }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import http from '@/api/http'

const products = ref([])
const totalPages = ref(1)
const loading = ref(true)

const filters = ref({
  page: 1,
  pageSize: 20,
  status: '',
  search: ''
})

const fetchProducts = async () => {
  loading.value = true
  try {
    const query = new URLSearchParams()
    query.append('page', filters.value.page)
    query.append('page_size', filters.value.pageSize)
    if (filters.value.status) query.append('status', filters.value.status)
    if (filters.value.search) query.append('search', filters.value.search)

    const res = await http(`/products?${query.toString()}`)
    if (!res || !Array.isArray(res.products)) {
      throw new Error("[CRITICAL] Malformed products response from server")
    }
    products.value = res.products
    totalPages.value = res.total_pages || 1
  } catch (err) {
    alert("Failed to load products: " + err.message)
  } finally {
    loading.value = false
  }
}

const changePage = (p) => {
  filters.value.page = p
  fetchProducts()
}

const updateStatus = async (id, status) => {
  if (status === 'inactive' && !confirm("Are you sure you want to UNPUBLISH this product? It will be hidden from the store.")) return
  
  try {
    await http(`/products/${id}/status`, {
      method: 'PATCH',
      body: JSON.stringify({ status })
    })
    
    // Optimistic Update
    const listed = products.value.find(p => p.id === id)
    if (listed) listed.status = status
    alert("Product status updated successfully")
  } catch (err) {
    alert("Failed: " + err.message)
  }
}

const formatPrice = (price) => {
  if (price == null) return "0.00"
  return Number(price).toFixed(2)
}

// Helpers for badges
const getStatusClass = (status) => {
  switch (status) {
    case 'active': return 'bg-emerald-500/10 text-emerald-500 border-emerald-500/20';
    case 'inactive': return 'bg-white/10 text-white/50 border-white/20';
    case 'out_of_stock': return 'bg-rose-500/10 text-rose-500 border-rose-500/20';
    default: return 'bg-white/10 text-white/70 border-white/20';
  }
}

onMounted(() => {
  fetchProducts()
})
</script>
