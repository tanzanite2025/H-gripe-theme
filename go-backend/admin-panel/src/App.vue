<template>
  <div class="min-h-screen bg-black text-white font-sans overflow-hidden">
    <!-- Navbar -->
    <nav class="border-b border-dashed border-white/10 bg-black/50 backdrop-blur-md sticky top-0 z-50">
      <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div class="flex items-center justify-between h-16">
          <div class="flex items-center">
            <div class="flex-shrink-0">
              <span class="text-xl font-black tracking-tighter italic uppercase bg-gradient-to-r from-emerald-400 to-cyan-400 bg-clip-text text-transparent">Tanzanite ERP</span>
            </div>
            <div class="hidden md:block ml-10">
              <div class="flex items-baseline space-x-4">
                <router-link to="/" class="nav-btn" active-class="nav-btn-active">Dashboard</router-link>
                <router-link to="/orders" class="nav-btn" active-class="nav-btn-active">Orders</router-link>
                <router-link to="/products" class="nav-btn" active-class="nav-btn-active">Products</router-link>
                <router-link to="/users" class="nav-btn" active-class="nav-btn-active">Users</router-link>
                <router-link to="/loyalty" class="nav-btn" active-class="nav-btn-active">Loyalty</router-link>
                <router-link to="/coupons" class="nav-btn" active-class="nav-btn-active">Coupons</router-link>
                <router-link to="/showcase" class="nav-btn" active-class="nav-btn-active">Showcase</router-link>
                <router-link to="/faq" class="nav-btn" active-class="nav-btn-active">FAQ</router-link>
              </div>
            </div>
          </div>
        </div>
      </div>
    </nav>

    <!-- Main Content -->
    <main class="max-w-7xl mx-auto py-6 sm:px-6 lg:px-8">
      <router-view v-slot="{ Component }">
        <transition name="fade" mode="out-in">
          <component :is="Component" />
        </transition>
      </router-view>
    </main>
  </div>
</template>

<script setup>
// Navigation is now handled by vue-router
</script>

<style scoped>
.nav-btn {
  @apply px-3 py-2 rounded-md text-sm font-black uppercase tracking-widest text-gray-300 hover:bg-gray-700 hover:text-white transition-all duration-200;
}
.nav-btn-active {
  @apply bg-gray-900 text-white italic border-b-2 border-emerald-500;
}

/* Base styles for the fake tailwind classes since real tailwind isn't fully set up yet */
.nav-btn {
  padding: 0.5rem 0.75rem;
  border-radius: 0.375rem;
  font-size: 0.875rem;
  font-weight: 900;
  text-transform: uppercase;
  letter-spacing: 0.1em;
  color: #d1d5db;
  transition: all 0.2s;
}
.nav-btn:hover {
  background-color: #374151;
  color: #fff;
}
.nav-btn-active {
  background-color: #111827;
  color: #fff;
  font-style: italic;
  border-bottom: 2px solid #10b981;
}

.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.3s ease;
}
.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
</style>
