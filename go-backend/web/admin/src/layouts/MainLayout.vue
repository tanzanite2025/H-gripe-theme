<template>
  <el-container class="layout-container">
    <!-- 侧边栏 -->
    <el-aside :width="isCollapse ? '64px' : '200px'" class="sidebar">
      <div class="logo">
        <h3 v-if="!isCollapse">Tanzanite</h3>
        <span v-else>T</span>
      </div>

      <el-menu
        :default-active="activeMenu"
        :collapse="isCollapse"
        :router="true"
        background-color="#304156"
        text-color="#bfcbd9"
        active-text-color="#409eff"
      >
        <el-menu-item index="/" :route="{ name: 'Dashboard' }">
          <el-icon><DataAnalysis /></el-icon>
          <template #title>仪表板</template>
        </el-menu-item>

        <el-menu-item
          v-if="hasPermission('product:view')"
          index="/products"
          :route="{ name: 'Products' }"
        >
          <el-icon><Goods /></el-icon>
          <template #title>商品管理</template>
        </el-menu-item>

        <el-menu-item
          v-if="hasPermission('order:view')"
          index="/orders"
          :route="{ name: 'Orders' }"
        >
          <el-icon><ShoppingCart /></el-icon>
          <template #title>订单管理</template>
        </el-menu-item>

        <el-menu-item
          v-if="hasPermission('user:view')"
          index="/users"
          :route="{ name: 'Users' }"
        >
          <el-icon><User /></el-icon>
          <template #title>用户管理</template>
        </el-menu-item>

        <el-menu-item
          v-if="hasPermission('content:view')"
          index="/content"
          :route="{ name: 'Content' }"
        >
          <el-icon><Document /></el-icon>
          <template #title>内容管理</template>
        </el-menu-item>

        <el-menu-item
          v-if="hasPermission('faq:view')"
          index="/faqs"
          :route="{ name: 'FAQs' }"
        >
          <el-icon><QuestionFilled /></el-icon>
          <template #title>FAQ管理</template>
        </el-menu-item>

        <el-menu-item
          v-if="hasPermission('gallery:view')"
          index="/galleries"
          :route="{ name: 'Galleries' }"
        >
          <el-icon><Picture /></el-icon>
          <template #title>图库管理</template>
        </el-menu-item>

        <el-menu-item
          v-if="hasPermission('subscription:view')"
          index="/subscriptions"
          :route="{ name: 'Subscriptions' }"
        >
          <el-icon><Message /></el-icon>
          <template #title>订阅管理</template>
        </el-menu-item>

        <el-menu-item
          v-if="hasPermission('ticket:view')"
          index="/tickets"
          :route="{ name: 'Tickets' }"
        >
          <el-icon><ChatDotRound /></el-icon>
          <template #title>工单管理</template>
        </el-menu-item>

        <el-menu-item
          v-if="hasPermission('marketing:view')"
          index="/marketing"
          :route="{ name: 'Marketing' }"
        >
          <el-icon><Promotion /></el-icon>
          <template #title>营销管理</template>
        </el-menu-item>

        <el-menu-item
          v-if="hasPermission('settings:view')"
          index="/settings"
          :route="{ name: 'Settings' }"
        >
          <el-icon><Setting /></el-icon>
          <template #title>系统设置</template>
        </el-menu-item>
      </el-menu>
    </el-aside>

    <!-- 主内容区 -->
    <el-container>
      <!-- 顶部导航栏 -->
      <el-header class="header">
        <div class="header-left">
          <el-icon class="collapse-icon" @click="toggleCollapse">
            <Fold v-if="!isCollapse" />
            <Expand v-else />
          </el-icon>
        </div>

        <div class="header-right">
          <el-dropdown @command="handleCommand">
            <span class="user-dropdown">
              <el-avatar :size="32" :src="userAvatar" />
              <span class="username">{{ user?.username || '管理员' }}</span>
              <el-icon><ArrowDown /></el-icon>
            </span>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item disabled>
                  {{ user?.email }}
                </el-dropdown-item>
                <el-dropdown-item divided command="logout">
                  <el-icon><SwitchButton /></el-icon>
                  退出登录
                </el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </div>
      </el-header>

      <!-- 内容区域 -->
      <el-main class="main-content">
        <router-view />
      </el-main>
    </el-container>
  </el-container>
</template>

<script setup>
import { ref, computed } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { ElMessageBox } from 'element-plus'
import {
  DataAnalysis,
  Goods,
  ShoppingCart,
  User,
  Document,
  QuestionFilled,
  Picture,
  Message,
  ChatDotRound,
  Promotion,
  Setting,
  Fold,
  Expand,
  ArrowDown,
  SwitchButton
} from '@element-plus/icons-vue'
import { useAuthStore } from '@/stores/auth'

const router = useRouter()
const route = useRoute()
const authStore = useAuthStore()

const isCollapse = ref(false)
const user = computed(() => authStore.user)
const userAvatar = computed(() => `https://ui-avatars.com/api/?name=${user.value?.username || 'Admin'}&background=409eff&color=fff`)

const activeMenu = computed(() => route.path)

const hasPermission = (permission) => {
  return authStore.hasPermission(permission)
}

const toggleCollapse = () => {
  isCollapse.value = !isCollapse.value
}

const handleCommand = async (command) => {
  if (command === 'logout') {
    try {
      await ElMessageBox.confirm('确定要退出登录吗？', '提示', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      })

      authStore.logout()
      router.push('/login')
    } catch {
      // 用户取消
    }
  }
}
</script>

<style scoped>
.layout-container {
  height: 100vh;
}

.sidebar {
  background-color: #304156;
  transition: width 0.3s;
  overflow-x: hidden;
}

.logo {
  height: 60px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #fff;
  font-size: 20px;
  font-weight: bold;
  border-bottom: 1px solid #1f2d3d;
}

.logo h3 {
  margin: 0;
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  background-color: #fff;
  border-bottom: 1px solid #e4e7ed;
  padding: 0 20px;
}

.header-left {
  display: flex;
  align-items: center;
}

.collapse-icon {
  font-size: 20px;
  cursor: pointer;
  transition: color 0.3s;
}

.collapse-icon:hover {
  color: #409eff;
}

.header-right {
  display: flex;
  align-items: center;
}

.user-dropdown {
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
  padding: 8px;
  border-radius: 4px;
  transition: background-color 0.3s;
}

.user-dropdown:hover {
  background-color: #f5f7fa;
}

.username {
  font-size: 14px;
  color: #303133;
}

.main-content {
  background-color: #f0f2f5;
  padding: 20px;
}
</style>
