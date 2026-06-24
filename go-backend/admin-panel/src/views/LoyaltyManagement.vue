<template>
  <div class="erp-root flex flex-col gap-8 animate-in fade-in duration-700">
    <header class="erp-header rounded-[32px] border-dashed bg-muted/5 p-8 flex justify-between items-center">
      <div>
        <h1 class="text-lg font-black tracking-tighter italic uppercase">LOYALTY & TIERS MANAGEMENT</h1>
        <p class="text-[9px] font-black uppercase tracking-widest opacity-60 mt-1">Configure Membership Levels and User Points</p>
      </div>
      <button class="erp-btn-primary rounded-full h-11 px-6 text-[10px] font-black uppercase tracking-widest" @click="fetchData">
        Refresh Data
      </button>
    </header>

    <div class="grid grid-cols-1 md:grid-cols-2 gap-8">
      <!-- Member Levels Section -->
      <section class="erp-card rounded-[24px] border-dashed p-6 relative overflow-hidden">
        <div class="absolute inset-0 bg-gradient-to-br from-primary/5 via-transparent pointer-events-none"></div>
        <h2 class="text-sm font-black tracking-tighter italic mb-4">Member Levels</h2>
        
        <div class="flex flex-col gap-4">
          <div v-for="level in levels" :key="level.id" class="flex items-center justify-between p-4 bg-muted/10 rounded-2xl">
            <div>
              <div class="text-[10px] font-black uppercase tracking-widest">{{ level.name }}</div>
              <div class="text-[8px] font-mono mt-1 text-muted-foreground/50">{{ level.min_points }} - {{ level.max_points }} PTS</div>
            </div>
            <div class="erp-badge status-healthy rounded-full h-5 px-3 flex items-center justify-center text-[8px] font-mono">
              ACTIVE
            </div>
          </div>
          
          <div class="mt-4 border-t border-dashed pt-4">
            <h3 class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/50 mb-3">Add New Level</h3>
            <div class="flex gap-2">
              <input v-model="newLevel.name" placeholder="Name" class="erp-input flex-1 h-12 rounded-2xl border-none bg-muted/50 px-4" />
              <input v-model.number="newLevel.min_points" type="number" placeholder="Min Pts" class="erp-input w-24 h-12 rounded-2xl border-none bg-muted/50 px-4" />
              <input v-model.number="newLevel.max_points" type="number" placeholder="Max Pts" class="erp-input w-24 h-12 rounded-2xl border-none bg-muted/50 px-4" />
              <button @click="createLevel" class="erp-btn-secondary h-12 rounded-2xl px-4 text-[10px] font-black uppercase">Add</button>
            </div>
          </div>
        </div>
      </section>

      <!-- Point Adjustment Section -->
      <section class="erp-card rounded-[24px] border-dashed p-6 relative overflow-hidden">
        <div class="absolute inset-0 bg-gradient-to-br from-primary/5 via-transparent pointer-events-none"></div>
        <h2 class="text-sm font-black tracking-tighter italic mb-4">Manual Point Adjustment</h2>
        
        <div class="flex flex-col gap-4">
          <div>
            <label class="block text-[10px] font-black uppercase tracking-widest text-muted-foreground/50 mb-2">User ID</label>
            <input v-model.number="adjustForm.userId" type="number" class="erp-input w-full h-12 rounded-2xl border-none bg-muted/50 px-4" placeholder="Enter User ID" />
          </div>
          
          <div>
            <label class="block text-[10px] font-black uppercase tracking-widest text-muted-foreground/50 mb-2">Points (Negative to deduct)</label>
            <input v-model.number="adjustForm.points" type="number" class="erp-input w-full h-12 rounded-2xl border-none bg-muted/50 px-4" placeholder="+100 or -50" />
          </div>

          <div>
            <label class="block text-[10px] font-black uppercase tracking-widest text-muted-foreground/50 mb-2">Reason</label>
            <input v-model="adjustForm.reason" class="erp-input w-full h-12 rounded-2xl border-none bg-muted/50 px-4" placeholder="e.g., Customer Service Compensation" />
          </div>

          <button @click="adjustPoints" class="erp-btn-primary rounded-full h-11 w-full text-[10px] font-black uppercase tracking-widest mt-2">
            Execute Adjustment
          </button>
        </div>
      </section>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import http from '@/api/http'

const levels = ref([])
const newLevel = ref({ name: '', min_points: 0, max_points: 0 })

const adjustForm = ref({ userId: null, points: 0, reason: '' })

// Fetch levels
const fetchLevels = async () => {
  try {
    const data = await http('/marketing/levels')
    if (!data || !Array.isArray(data.levels)) {
      throw new Error("[CRITICAL] Invalid levels data structure from server")
    }
    levels.value = data.levels
  } catch (err) {
    alert("Failed to load levels: " + err.message)
  }
}

const createLevel = async () => {
  if (!newLevel.value.name) return
  try {
    await http('/marketing/levels', {
      method: 'POST',
      body: JSON.stringify(newLevel.value)
    })
    alert('Level Created!')
    newLevel.value = { name: '', min_points: 0, max_points: 0 }
    fetchLevels()
  } catch (err) {
    alert("Failed to create level: " + err.message)
  }
}

const adjustPoints = async () => {
  if (!adjustForm.value.userId || !adjustForm.value.reason) {
    alert('User ID and Reason are required')
    return
  }
  try {
    await http('/marketing/loyalty/transactions', {
      method: 'POST',
      body: JSON.stringify({
        user_id: Number(adjustForm.value.userId),
        points: Number(adjustForm.value.points),
        description: adjustForm.value.reason
      })
    })
    alert('Points adjusted successfully!')
    adjustForm.value = { userId: null, points: 0, reason: '' }
  } catch (err) {
    alert("Failed to adjust points: " + err.message)
  }
}

const fetchData = () => {
  fetchLevels()
}

onMounted(() => {
  fetchData()
})
</script>

<style scoped>
/* ERP UDS vanilla CSS implementations mimicking Tailwind utility classes requested in rules */
.erp-root {
  display: flex;
  flex-direction: column;
  gap: 2rem;
  animation: fadeIn 0.7s ease-in-out;
  padding: 2rem;
  font-family: 'Inter', system-ui, sans-serif;
  background-color: #0a0a0a;
  color: #ededed;
  min-height: 100vh;
}

.erp-header {
  border-radius: 32px;
  border: 1px dashed rgba(255, 255, 255, 0.2);
  background-color: rgba(255, 255, 255, 0.05);
  padding: 2rem;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.erp-card {
  border-radius: 24px;
  border: 1px dashed rgba(255, 255, 255, 0.2);
  padding: 1.5rem;
  background-color: #121212;
}

.text-lg { font-size: 1.5rem; }
.text-sm { font-size: 1.1rem; }
.font-black { font-weight: 900; }
.tracking-tighter { letter-spacing: -0.05em; }
.italic { font-style: italic; }
.uppercase { text-transform: uppercase; }

.text-\[10px\] { font-size: 10px; }
.text-\[9px\] { font-size: 9px; }
.text-\[8px\] { font-size: 8px; }
.tracking-widest { letter-spacing: 0.1em; }
.opacity-60 { opacity: 0.6; }
.font-mono { font-family: monospace; }
.text-muted-foreground\/50 { color: rgba(255, 255, 255, 0.5); }

.bg-muted\/10 { background-color: rgba(255, 255, 255, 0.1); }
.bg-muted\/50 { background-color: rgba(255, 255, 255, 0.08); color: white; }

.erp-btn-primary {
  background-color: #10b981;
  color: #000;
  border: none;
  cursor: pointer;
  transition: all 0.2s;
}
.erp-btn-primary:hover { background-color: #34d399; }

.erp-btn-secondary {
  background-color: rgba(255, 255, 255, 0.1);
  color: white;
  border: none;
  cursor: pointer;
}

.status-healthy {
  background-color: rgba(16, 185, 129, 0.1);
  color: #10b981;
}

.flex { display: flex; }
.flex-col { flex-direction: column; }
.items-center { align-items: center; }
.justify-between { justify-content: space-between; }
.justify-center { justify-content: center; }
.gap-8 { gap: 2rem; }
.gap-4 { gap: 1rem; }
.gap-2 { gap: 0.5rem; }

.grid { display: grid; }
.grid-cols-1 { grid-template-columns: repeat(1, minmax(0, 1fr)); }
@media (min-width: 768px) {
  .md\:grid-cols-2 { grid-template-columns: repeat(2, minmax(0, 1fr)); }
}

.rounded-full { border-radius: 9999px; }
.rounded-2xl { border-radius: 1rem; }
.h-11 { height: 2.75rem; }
.h-12 { height: 3rem; }
.h-5 { height: 1.25rem; }
.px-6 { padding-left: 1.5rem; padding-right: 1.5rem; }
.px-4 { padding-left: 1rem; padding-right: 1rem; }
.px-3 { padding-left: 0.75rem; padding-right: 0.75rem; }
.p-8 { padding: 2rem; }
.p-4 { padding: 1rem; }
.mt-1 { margin-top: 0.25rem; }
.mt-2 { margin-top: 0.5rem; }
.mt-4 { margin-top: 1rem; }
.mb-2 { margin-bottom: 0.5rem; }
.mb-3 { margin-bottom: 0.75rem; }
.mb-4 { margin-bottom: 1rem; }
.pt-4 { padding-top: 1rem; }

.w-full { width: 100%; }
.w-24 { width: 6rem; }
.flex-1 { flex: 1 1 0%; }
.block { display: block; }
.border-t { border-top-width: 1px; border-top-style: dashed; border-top-color: rgba(255, 255, 255, 0.2); }

@keyframes fadeIn {
  from { opacity: 0; transform: translateY(10px); }
  to { opacity: 1; transform: translateY(0); }
}
</style>
