<template>
  <div class="flex flex-col gap-8 animate-in fade-in duration-700 min-h-screen p-8 bg-[#0a0a0a] text-white font-sans">
    
    <!-- Header Module -->
    <header class="rounded-[32px] border border-dashed border-white/20 bg-muted/5 p-8 flex justify-between items-center relative overflow-hidden">
      <div class="absolute inset-0 bg-gradient-to-br from-emerald-500/5 via-transparent pointer-events-none"></div>
      <div>
        <h1 class="text-lg font-black tracking-tighter italic uppercase text-white">ORDER MANAGEMENT</h1>
        <p class="text-[9px] font-black uppercase tracking-widest opacity-60 mt-1">Centralized Processing & Fulfillment</p>
      </div>
      <div class="flex gap-4">
        <input v-model="filters.search" @keyup.enter="fetchOrders" placeholder="Search Order # or Name" class="h-11 rounded-full border-none bg-white/5 px-6 text-white text-[10px] font-mono uppercase focus:ring-1 focus:ring-emerald-500 outline-none w-64" />
        <button @click="fetchOrders" class="rounded-full h-11 px-8 bg-emerald-500 hover:bg-emerald-400 text-black font-black text-[10px] uppercase tracking-widest transition-colors shadow-[0_0_15px_rgba(16,185,129,0.3)]">
          SEARCH
        </button>
      </div>
    </header>

    <!-- Filters & List -->
    <div class="rounded-[24px] border border-dashed border-white/20 bg-muted/5 p-8 relative flex-grow flex flex-col">
      <div class="flex justify-between items-center mb-8">
        <h2 class="text-sm font-black tracking-tighter italic text-white">ALL TRANSACTIONS</h2>
        
        <div class="flex gap-2">
          <select v-model="filters.status" @change="fetchOrders" class="h-9 rounded-full border-none bg-white/5 px-4 text-white text-[10px] font-bold uppercase tracking-widest outline-none">
            <option value="" class="bg-[#111]">ALL STATUS</option>
            <option value="pending" class="bg-[#111]">PENDING</option>
            <option value="processing" class="bg-[#111]">PROCESSING</option>
            <option value="shipped" class="bg-[#111]">SHIPPED</option>
            <option value="completed" class="bg-[#111]">COMPLETED</option>
            <option value="cancelled" class="bg-[#111]">CANCELLED</option>
          </select>
          <select v-model="filters.payment_status" @change="fetchOrders" class="h-9 rounded-full border-none bg-white/5 px-4 text-white text-[10px] font-bold uppercase tracking-widest outline-none">
            <option value="" class="bg-[#111]">ALL PAYMENTS</option>
            <option value="unpaid" class="bg-[#111]">UNPAID</option>
            <option value="paid" class="bg-[#111]">PAID</option>
            <option value="refunded" class="bg-[#111]">REFUNDED</option>
          </select>
          <select v-model="filters.shipping_status" @change="fetchOrders" class="h-9 rounded-full border-none bg-white/5 px-4 text-white text-[10px] font-bold uppercase tracking-widest outline-none">
            <option value="" class="bg-[#111]">ALL SHIPPING</option>
            <option value="pending" class="bg-[#111]">UNSHIPPED</option>
            <option value="shipped" class="bg-[#111]">SHIPPED</option>
            <option value="delivered" class="bg-[#111]">DELIVERED</option>
          </select>
        </div>
      </div>
      
      <div v-if="loading" class="text-emerald-500/50 text-[10px] font-mono animate-pulse flex-grow flex items-center justify-center">SYNCING_DATA...</div>
      <div v-else-if="orders.length === 0" class="text-white/50 text-[10px] font-mono flex-grow flex items-center justify-center">NO_ORDERS_FOUND</div>
      
      <div v-else class="flex flex-col gap-2">
        <!-- Table Header -->
        <div class="grid grid-cols-12 gap-4 px-6 py-3 border-b border-dashed border-white/10 text-[10px] font-black uppercase tracking-widest text-white/50">
          <div class="col-span-2">Order #</div>
          <div class="col-span-2">Customer</div>
          <div class="col-span-2">Date</div>
          <div class="col-span-1">Items</div>
          <div class="col-span-1 text-right">Total</div>
          <div class="col-span-3 text-center">Status Matrix</div>
          <div class="col-span-1 text-right">Actions</div>
        </div>
        
        <!-- Rows -->
        <div v-for="order in orders" :key="order.id" class="grid grid-cols-12 gap-4 px-6 py-4 items-center rounded-2xl border border-dashed border-transparent hover:border-white/10 hover:bg-white/5 transition-colors group cursor-pointer" @click="viewDetails(order.id)">
          <div class="col-span-2 text-sm font-black tracking-tighter italic text-white">{{ order.order_number }}</div>
          <div class="col-span-2 text-xs text-white/80 truncate">{{ order.customer_name || `User #${order.user_id}` }}</div>
          <div class="col-span-2 text-[10px] font-mono text-white/50">{{ new Date(order.created_at).toLocaleString() }}</div>
          <div class="col-span-1 text-[10px] font-mono text-white/80 text-center">{{ order.item_count }}</div>
          <div class="col-span-1 text-[10px] font-mono text-emerald-400 text-right">${{ order.total_amount.toFixed(2) }}</div>
          
          <div class="col-span-3 flex justify-center gap-2">
            <!-- Overall Status -->
            <span :class="getStatusClass(order.status)" class="px-2 py-1 rounded-md text-[8px] font-mono uppercase border w-20 text-center truncate">
              {{ order.status }}
            </span>
            <!-- Payment -->
            <span :class="getPaymentStatusClass(order.payment_status)" class="px-2 py-1 rounded-md text-[8px] font-mono uppercase border w-16 text-center truncate">
              {{ order.payment_status }}
            </span>
            <!-- Shipping -->
            <span :class="getShippingStatusClass(order.shipping_status)" class="px-2 py-1 rounded-md text-[8px] font-mono uppercase border w-20 text-center truncate">
              {{ order.shipping_status }}
            </span>
          </div>
          
          <div class="col-span-1 text-right opacity-0 group-hover:opacity-100 transition-opacity">
            <button class="text-[10px] font-black uppercase tracking-widest text-emerald-500 hover:text-white transition-colors" @click.stop="viewDetails(order.id)">
              OPEN
            </button>
          </div>
        </div>
      </div>
      
      <!-- Pagination -->
      <div v-if="totalPages > 1" class="mt-8 flex justify-center gap-2 border-t border-dashed border-white/10 pt-6">
        <button v-for="p in totalPages" :key="p" @click="changePage(p)" :class="['w-8 h-8 rounded-full text-[10px] font-mono flex items-center justify-center transition-colors', p === filters.page ? 'bg-emerald-500 text-black font-bold' : 'bg-white/5 text-white/50 hover:bg-white/10']">
          {{ p }}
        </button>
      </div>
    </div>

    <!-- Order Details Modal -->
    <div v-if="selectedOrder" class="fixed inset-0 z-50 flex items-center justify-center bg-black/90 backdrop-blur-md p-4 animate-in fade-in">
      <div class="w-full max-w-5xl h-[85vh] rounded-[32px] bg-[#0f0f0f] shadow-2xl relative overflow-hidden flex flex-col border border-white/10">
        <!-- Modal Header -->
        <div class="p-8 border-b border-dashed border-white/10 flex justify-between items-center bg-gradient-to-r from-white/5 to-transparent">
          <div>
            <div class="flex items-center gap-3">
              <h2 class="text-2xl font-black tracking-tighter italic uppercase text-white">{{ selectedOrder.order_number }}</h2>
              <span :class="getStatusClass(selectedOrder.status)" class="px-3 py-1 rounded-full text-[10px] font-mono uppercase border">{{ selectedOrder.status }}</span>
            </div>
            <p class="text-[10px] font-mono text-white/50 mt-2">{{ new Date(selectedOrder.created_at).toLocaleString() }}</p>
          </div>
          <div class="flex gap-3">
             <select v-model="editStatus.status" class="h-11 rounded-full border border-dashed border-white/20 bg-black px-4 text-white text-[10px] font-bold uppercase tracking-widest outline-none">
              <option value="pending">PENDING</option>
              <option value="processing">PROCESSING</option>
              <option value="shipped">SHIPPED</option>
              <option value="completed">COMPLETED</option>
              <option value="cancelled">CANCELLED</option>
            </select>
            <button @click="updateOrderStatus" class="rounded-full h-11 px-6 bg-white/10 hover:bg-white/20 text-white font-black text-[10px] uppercase tracking-widest transition-colors">
              UPDATE STATUS
            </button>
            <button @click="closeDetails" class="rounded-full w-11 h-11 bg-rose-500/10 text-rose-500 flex items-center justify-center hover:bg-rose-500/20 transition-colors ml-4">
              <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" viewBox="0 0 20 20" fill="currentColor"><path fill-rule="evenodd" d="M4.293 4.293a1 1 0 011.414 0L10 8.586l4.293-4.293a1 1 0 111.414 1.414L11.414 10l4.293 4.293a1 1 0 01-1.414 1.414L10 11.414l-4.293 4.293a1 1 0 01-1.414-1.414L8.586 10 4.293 5.707a1 1 0 010-1.414z" clip-rule="evenodd" /></svg>
            </button>
          </div>
        </div>

        <!-- Modal Body -->
        <div class="flex-grow overflow-y-auto p-8 grid grid-cols-3 gap-8">
          
          <!-- Left Column: Items -->
          <div class="col-span-2 flex flex-col gap-6">
            <div class="rounded-[24px] border border-dashed border-white/10 bg-black/40 p-6">
              <h3 class="text-[10px] font-black uppercase tracking-widest text-white/50 mb-4">ORDER ITEMS</h3>
              <div class="flex flex-col gap-4">
                <div v-for="item in selectedOrder.items" :key="item.id" class="flex justify-between items-center py-3 border-b border-dashed border-white/5 last:border-0">
                  <div class="flex items-center gap-4">
                    <div class="w-12 h-12 rounded-xl bg-white/5 flex items-center justify-center text-[8px] font-mono text-white/30 border border-white/10">IMG</div>
                    <div>
                      <div class="text-sm font-bold text-white">{{ item.product_name }}</div>
                      <div class="text-[9px] font-mono text-white/40 mt-1">SKU: {{ item.sku }} | QTY: {{ item.quantity }}</div>
                    </div>
                  </div>
                  <div class="text-right">
                    <div class="text-sm font-mono text-white">${{ item.total.toFixed(2) }}</div>
                    <div class="text-[9px] font-mono text-white/40 mt-1">${{ item.price.toFixed(2) }} each</div>
                  </div>
                </div>
              </div>
            </div>

            <div class="rounded-[24px] border border-dashed border-white/10 bg-black/40 p-6">
              <h3 class="text-[10px] font-black uppercase tracking-widest text-white/50 mb-4">FINANCIAL SUMMARY</h3>
              <div class="space-y-2 font-mono text-xs">
                <div class="flex justify-between text-white/60"><span>Subtotal</span><span>${{ selectedOrder.subtotal_amount.toFixed(2) }}</span></div>
                <div class="flex justify-between text-white/60"><span>Shipping Fee</span><span>${{ selectedOrder.shipping_fee.toFixed(2) }}</span></div>
                <div class="flex justify-between text-white/60"><span>Tax</span><span>${{ selectedOrder.tax_amount.toFixed(2) }}</span></div>
                <div class="flex justify-between text-emerald-400/80"><span>Discount</span><span>-${{ selectedOrder.discount_amount.toFixed(2) }}</span></div>
                <div class="flex justify-between text-emerald-400/80"><span>Points Value</span><span>-${{ selectedOrder.points_value.toFixed(2) }}</span></div>
                <div class="flex justify-between text-white font-bold text-lg pt-4 border-t border-dashed border-white/10 mt-2">
                  <span>TOTAL</span><span class="text-emerald-400">${{ selectedOrder.total_amount.toFixed(2) }}</span>
                </div>
              </div>
            </div>
          </div>
          
          <!-- Right Column: Meta & Addressing -->
          <div class="col-span-1 flex flex-col gap-6">
            
            <div class="rounded-[24px] border border-dashed border-white/10 bg-black/40 p-6 relative">
              <h3 class="text-[10px] font-black uppercase tracking-widest text-white/50 mb-4">PAYMENT & SHIPPING</h3>
              
              <div class="mb-4">
                <div class="text-[8px] font-black uppercase tracking-widest text-white/30 mb-1">PAYMENT STATUS</div>
                <div class="flex items-center justify-between">
                  <span :class="getPaymentStatusClass(selectedOrder.payment_status)" class="px-2 py-1 rounded-md text-[9px] font-mono uppercase border">
                    {{ selectedOrder.payment_status }}
                  </span>
                  <button @click="markPaid" v-if="selectedOrder.payment_status === 'unpaid'" class="text-[9px] font-black uppercase tracking-widest text-emerald-500 hover:text-emerald-400">MARK PAID</button>
                  <button @click="issueRefund" v-if="selectedOrder.payment_status === 'paid'" class="text-[9px] font-black uppercase tracking-widest text-amber-500 hover:text-amber-400">REFUND</button>
                </div>
              </div>

              <div>
                <div class="text-[8px] font-black uppercase tracking-widest text-white/30 mb-1">SHIPPING STATUS</div>
                <div class="flex items-center justify-between">
                  <span :class="getShippingStatusClass(selectedOrder.shipping_status)" class="px-2 py-1 rounded-md text-[9px] font-mono uppercase border">
                    {{ selectedOrder.shipping_status }}
                  </span>
                  <button @click="shipOrder" v-if="selectedOrder.shipping_status === 'pending'" class="text-[9px] font-black uppercase tracking-widest text-emerald-500 hover:text-emerald-400">SHIP OUT</button>
                </div>
                <div v-if="selectedOrder.tracking_number" class="mt-2 text-[10px] font-mono text-white/70 bg-white/5 p-2 rounded-lg break-all">
                  TRK: {{ selectedOrder.tracking_number }} ({{ selectedOrder.carrier_code }})
                </div>
              </div>
            </div>

            <div class="rounded-[24px] border border-dashed border-white/10 bg-black/40 p-6">
              <h3 class="text-[10px] font-black uppercase tracking-widest text-white/50 mb-4">SHIPPING ADDRESS</h3>
              <div v-if="selectedOrder.shipping_address" class="text-xs text-white/80 leading-relaxed font-mono">
                {{ selectedOrder.shipping_address.first_name }} {{ selectedOrder.shipping_address.last_name }}<br>
                <template v-if="selectedOrder.shipping_address.company">{{ selectedOrder.shipping_address.company }}<br></template>
                {{ selectedOrder.shipping_address.address_1 }}<br>
                <template v-if="selectedOrder.shipping_address.address_2">{{ selectedOrder.shipping_address.address_2 }}<br></template>
                {{ selectedOrder.shipping_address.city }}, {{ selectedOrder.shipping_address.state }} {{ selectedOrder.shipping_address.postal_code }}<br>
                {{ selectedOrder.shipping_address.country }}<br>
                <span class="text-white/50 mt-2 block">Phone: {{ selectedOrder.shipping_address.phone }}</span>
                <span class="text-white/50 block">Email: {{ selectedOrder.shipping_address.email }}</span>
              </div>
              <div v-else class="text-[10px] font-mono text-white/30">NO ADDRESS PROVIDED</div>
            </div>
            
            <div class="rounded-[24px] border border-dashed border-white/10 bg-black/40 p-6">
              <h3 class="text-[10px] font-black uppercase tracking-widest text-white/50 mb-4">ADMIN NOTE</h3>
              <textarea v-model="editStatus.admin_note" rows="3" class="w-full rounded-xl border border-dashed border-white/20 bg-black p-3 text-xs text-white focus:ring-1 focus:ring-emerald-500 outline-none placeholder-white/20" placeholder="Internal notes..."></textarea>
              <div class="flex justify-end mt-2">
                <button @click="updateAdminNote" class="text-[9px] font-black uppercase tracking-widest text-emerald-500 hover:text-emerald-400">SAVE NOTE</button>
              </div>
            </div>

          </div>
        </div>
      </div>
    </div>

  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import http from '@/api/http'

const orders = ref([])
const totalPages = ref(1)
const loading = ref(true)

const filters = ref({
  page: 1,
  pageSize: 20,
  status: '',
  payment_status: '',
  shipping_status: '',
  search: ''
})

const selectedOrder = ref(null)
const editStatus = ref({
  status: '',
  admin_note: ''
})

const fetchOrders = async () => {
  loading.value = true
  try {
    const query = new URLSearchParams()
    query.append('page', filters.value.page)
    query.append('page_size', filters.value.pageSize)
    if (filters.value.status) query.append('status', filters.value.status)
    if (filters.value.payment_status) query.append('payment_status', filters.value.payment_status)
    if (filters.value.shipping_status) query.append('shipping_status', filters.value.shipping_status)
    if (filters.value.search) query.append('search', filters.value.search)

    const res = await http(`/orders?${query.toString()}`)
    if (!res || !Array.isArray(res.orders)) {
      throw new Error("[CRITICAL] Orders list returned from server is empty or malformed")
    }
    orders.value = res.orders
    totalPages.value = res.total_pages || 1
  } catch (err) {
    alert("Failed to load orders: " + err.message)
  } finally {
    loading.value = false
  }
}

const changePage = (p) => {
  filters.value.page = p
  fetchOrders()
}

const viewDetails = async (id) => {
  try {
    const data = await http(`/orders/${id}`)
    if (!data || !data.order) {
      throw new Error(`[CRITICAL] Order ID ${id} not found on server`)
    }
    selectedOrder.value = data.order
    editStatus.value.status = data.order.status
    editStatus.value.admin_note = data.order.admin_note
  } catch (err) {
    alert("Failed to load order details: " + err.message)
  }
}

const closeDetails = () => {
  selectedOrder.value = null
}

const updateOrderStatus = async () => {
  if (!selectedOrder.value) return
  try {
    await http(`/orders/${selectedOrder.value.id}/status`, {
      method: 'PATCH',
      body: JSON.stringify({ status: editStatus.value.status })
    })
    selectedOrder.value.status = editStatus.value.status
    // update in list too
    const listed = orders.value.find(o => o.id === selectedOrder.value.id)
    if (listed) listed.status = editStatus.value.status
    alert("Order status updated successfully")
  } catch (err) {
    alert("Update failed: " + err.message)
  }
}

const updateAdminNote = async () => {
  if (!selectedOrder.value) return
  try {
    await http(`/orders/${selectedOrder.value.id}/admin-note`, {
      method: 'PATCH',
      body: JSON.stringify({ admin_note: editStatus.value.admin_note })
    })
    selectedOrder.value.admin_note = editStatus.value.admin_note
    alert("Note saved")
  } catch (err) {
    alert("Failed to save note: " + err.message)
  }
}

const markPaid = async () => {
  if (!confirm("Confirm marking this order as PAID?")) return
  try {
    await http(`/orders/${selectedOrder.value.id}/payment-status`, {
      method: 'PATCH',
      body: JSON.stringify({ payment_status: 'paid' })
    })
    selectedOrder.value.payment_status = 'paid'
    const listed = orders.value.find(o => o.id === selectedOrder.value.id)
    if (listed) listed.payment_status = 'paid'
    alert("Order marked as PAID")
  } catch (err) {
    alert("Failed to update payment status: " + err.message)
  }
}

const issueRefund = async () => {
  if (!confirm("Are you sure you want to mark this as REFUNDED?")) return
  try {
    await http(`/orders/${selectedOrder.value.id}/payment-status`, {
      method: 'PATCH',
      body: JSON.stringify({ payment_status: 'refunded' })
    })
    selectedOrder.value.payment_status = 'refunded'
    const listed = orders.value.find(o => o.id === selectedOrder.value.id)
    if (listed) listed.payment_status = 'refunded'
    alert("Order marked as REFUNDED")
  } catch (err) {
    alert("Failed to refund order: " + err.message)
  }
}

const shipOrder = async () => {
  const trackingNumber = prompt("Enter Tracking Number (Optional):")
  if (trackingNumber === null) return // user cancelled
  
  try {
    await http(`/orders/${selectedOrder.value.id}/shipping-status`, {
      method: 'PATCH',
      body: JSON.stringify({ shipping_status: 'shipped' })
    })
    selectedOrder.value.shipping_status = 'shipped'
    
    if (trackingNumber) {
      await http(`/orders/${selectedOrder.value.id}/tracking`, {
        method: 'PATCH',
        body: JSON.stringify({ tracking_number: trackingNumber, carrier_code: 'AUTO' })
      })
      selectedOrder.value.tracking_number = trackingNumber
    }
    
    const listed = orders.value.find(o => o.id === selectedOrder.value.id)
    if (listed) listed.shipping_status = 'shipped'
    alert("Order marked as SHIPPED")
  } catch (err) {
    alert("Failed to update shipping: " + err.message)
  }
}

// Helpers for badges
const getStatusClass = (status) => {
  switch (status) {
    case 'completed': return 'bg-emerald-500/10 text-emerald-500 border-emerald-500/20';
    case 'processing': return 'bg-sky-500/10 text-sky-500 border-sky-500/20';
    case 'pending': return 'bg-amber-500/10 text-amber-500 border-amber-500/20';
    case 'cancelled': case 'refunded': return 'bg-rose-500/10 text-rose-500 border-rose-500/20';
    default: return 'bg-white/10 text-white/70 border-white/20';
  }
}

const getPaymentStatusClass = (status) => {
  if (status === 'paid') return 'bg-emerald-500/10 text-emerald-500 border-emerald-500/20';
  if (status === 'refunded') return 'bg-rose-500/10 text-rose-500 border-rose-500/20';
  return 'bg-amber-500/10 text-amber-500 border-amber-500/20';
}

const getShippingStatusClass = (status) => {
  if (status === 'delivered') return 'bg-emerald-500/10 text-emerald-500 border-emerald-500/20';
  if (status === 'shipped') return 'bg-sky-500/10 text-sky-500 border-sky-500/20';
  return 'bg-amber-500/10 text-amber-500 border-amber-500/20';
}

onMounted(() => {
  fetchOrders()
})
</script>
