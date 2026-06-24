<template>
  <div class="flex flex-col gap-8 animate-in fade-in duration-700 min-h-screen bg-[#0a0a0a] text-white font-sans p-8">
    
    <!-- Header Module -->
    <header class="rounded-[32px] border border-dashed border-white/20 bg-muted/5 p-8 flex justify-between items-center relative overflow-hidden">
      <div class="absolute inset-0 bg-gradient-to-br from-fuchsia-500/5 via-transparent pointer-events-none"></div>
      <div>
        <h1 class="text-lg font-black tracking-tighter italic uppercase text-white">USER MANAGEMENT</h1>
        <p class="text-[9px] font-black uppercase tracking-widest opacity-60 mt-1">Identity & Access Control</p>
      </div>
      <div class="flex gap-4">
        <input v-model="filters.search" @keyup.enter="fetchUsers" placeholder="Search Email or Username" class="h-11 rounded-full border-none bg-white/5 px-6 text-white text-[10px] font-mono uppercase focus:ring-1 focus:ring-fuchsia-500 outline-none w-64" />
        <button @click="fetchUsers" class="rounded-full h-11 px-8 bg-fuchsia-500 hover:bg-fuchsia-400 text-black font-black text-[10px] uppercase tracking-widest transition-colors shadow-[0_0_15px_rgba(217,70,239,0.3)]">
          SEARCH
        </button>
      </div>
    </header>

    <!-- Filters & List -->
    <div class="rounded-[24px] border border-dashed border-white/20 bg-muted/5 p-8 relative flex-grow flex flex-col">
      <div class="flex justify-between items-center mb-8">
        <h2 class="text-sm font-black tracking-tighter italic text-white">ALL ACCOUNTS</h2>
        
        <div class="flex gap-2">
          <select v-model="filters.role" @change="fetchUsers" class="h-9 rounded-full border-none bg-white/5 px-4 text-white text-[10px] font-bold uppercase tracking-widest outline-none">
            <option value="" class="bg-[#111]">ALL ROLES</option>
            <option value="user" class="bg-[#111]">USER</option>
            <option value="admin" class="bg-[#111]">ADMIN</option>
          </select>
          <select v-model="filters.status" @change="fetchUsers" class="h-9 rounded-full border-none bg-white/5 px-4 text-white text-[10px] font-bold uppercase tracking-widest outline-none">
            <option value="" class="bg-[#111]">ALL STATUS</option>
            <option value="active" class="bg-[#111]">ACTIVE</option>
            <option value="inactive" class="bg-[#111]">INACTIVE</option>
            <option value="suspended" class="bg-[#111]">SUSPENDED</option>
          </select>
        </div>
      </div>
      
      <div v-if="loading" class="text-fuchsia-500/50 text-[10px] font-mono animate-pulse flex-grow flex items-center justify-center">SYNCING_USERS...</div>
      <div v-else-if="users.length === 0" class="text-white/50 text-[10px] font-mono flex-grow flex items-center justify-center">NO_USERS_FOUND</div>
      
      <div v-else class="flex flex-col gap-2">
        <!-- Table Header -->
        <div class="grid grid-cols-12 gap-4 px-6 py-3 border-b border-dashed border-white/10 text-[10px] font-black uppercase tracking-widest text-white/50">
          <div class="col-span-1">ID</div>
          <div class="col-span-3">User</div>
          <div class="col-span-3">Email</div>
          <div class="col-span-2">Joined</div>
          <div class="col-span-1 text-center">Role</div>
          <div class="col-span-1 text-center">Status</div>
          <div class="col-span-1 text-right">Actions</div>
        </div>
        
        <!-- Rows -->
        <div v-for="user in users" :key="user.id" class="grid grid-cols-12 gap-4 px-6 py-4 items-center rounded-2xl border border-dashed border-transparent hover:border-white/10 hover:bg-white/5 transition-colors group">
          <div class="col-span-1 text-xs font-mono text-white/50">{{ user.id }}</div>
          <div class="col-span-3">
            <div class="text-sm font-black tracking-tighter italic text-white">{{ user.username }}</div>
            <div class="text-[9px] font-mono text-white/40 mt-1 uppercase">{{ user.first_name }} {{ user.last_name }}</div>
          </div>
          <div class="col-span-3 text-xs text-white/80 font-mono truncate">{{ user.email }}</div>
          <div class="col-span-2 text-[10px] font-mono text-white/50">{{ new Date(user.created_at).toLocaleString() }}</div>
          
          <div class="col-span-1 flex justify-center">
             <span :class="getRoleClass(user.role)" class="px-2 py-1 rounded-md text-[8px] font-mono uppercase border w-16 text-center truncate">
              {{ user.role }}
            </span>
          </div>
          
          <div class="col-span-1 flex justify-center">
            <span :class="getStatusClass(user.status)" class="px-2 py-1 rounded-md text-[8px] font-mono uppercase border w-20 text-center truncate">
              {{ user.status }}
            </span>
          </div>
          
          <div class="col-span-1 text-right flex gap-3 justify-end opacity-0 group-hover:opacity-100 transition-opacity">
            <button v-if="user.status !== 'suspended'" class="text-[10px] font-black uppercase tracking-widest text-rose-500 hover:text-rose-400 transition-colors" @click="updateStatus(user.id, 'suspended')">
              BAN
            </button>
            <button v-if="user.status === 'suspended'" class="text-[10px] font-black uppercase tracking-widest text-emerald-500 hover:text-emerald-400 transition-colors" @click="updateStatus(user.id, 'active')">
              UNBAN
            </button>
          </div>
        </div>
      </div>
      
      <!-- Pagination -->
      <div v-if="totalPages > 1" class="mt-8 flex justify-center gap-2 border-t border-dashed border-white/10 pt-6">
        <button v-for="p in totalPages" :key="p" @click="changePage(p)" :class="['w-8 h-8 rounded-full text-[10px] font-mono flex items-center justify-center transition-colors', p === filters.page ? 'bg-fuchsia-500 text-black font-bold' : 'bg-white/5 text-white/50 hover:bg-white/10']">
          {{ p }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import http from '@/api/http'

const users = ref([])
const totalPages = ref(1)
const loading = ref(true)

const filters = ref({
  page: 1,
  pageSize: 20,
  role: '',
  status: '',
  search: ''
})

const fetchUsers = async () => {
  loading.value = true
  try {
    const query = new URLSearchParams()
    query.append('page', filters.value.page)
    query.append('page_size', filters.value.pageSize)
    if (filters.value.role) query.append('role', filters.value.role)
    if (filters.value.status) query.append('status', filters.value.status)
    if (filters.value.search) query.append('search', filters.value.search)

    const res = await http(`/users?${query.toString()}`)
    if (!res || !Array.isArray(res.users)) {
      throw new Error("[CRITICAL] Malformed users response from server")
    }
    users.value = res.users
    totalPages.value = res.total_pages || 1
  } catch (err) {
    alert("Failed to load users: " + err.message)
  } finally {
    loading.value = false
  }
}

const changePage = (p) => {
  filters.value.page = p
  fetchUsers()
}

const updateStatus = async (id, status) => {
  if (status === 'suspended' && !confirm("Are you sure you want to BAN this user?")) return
  if (status === 'active' && !confirm("Are you sure you want to UNBAN this user?")) return
  
  try {
    await http(`/users/${id}/status`, {
      method: 'PATCH',
      body: JSON.stringify({ status })
    })
    
    // Optimistic Update
    const listed = users.value.find(u => u.id === id)
    if (listed) listed.status = status
    alert("User status updated successfully")
  } catch (err) {
    alert("Failed: " + err.message)
  }
}

// Helpers for badges
const getStatusClass = (status) => {
  switch (status) {
    case 'active': return 'bg-emerald-500/10 text-emerald-500 border-emerald-500/20';
    case 'inactive': return 'bg-white/10 text-white/50 border-white/20';
    case 'suspended': return 'bg-rose-500/10 text-rose-500 border-rose-500/20';
    default: return 'bg-white/10 text-white/70 border-white/20';
  }
}

const getRoleClass = (role) => {
  if (role === 'admin') return 'bg-fuchsia-500/10 text-fuchsia-500 border-fuchsia-500/20';
  return 'bg-white/5 text-white/50 border-white/10';
}

onMounted(() => {
  fetchUsers()
})
</script>
