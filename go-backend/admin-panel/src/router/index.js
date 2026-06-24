import { createRouter, createWebHistory } from 'vue-router'
import Dashboard from '../views/Dashboard.vue'
import LoyaltyManagement from '../views/LoyaltyManagement.vue'
import CouponManagement from '../views/CouponManagement.vue'
import PictureWarehouseApproval from '../views/PictureWarehouseApproval.vue'
import FaqManagement from '../views/FaqManagement.vue'
import OrderManagement from '../views/OrderManagement.vue'
import UserManagement from '../views/UserManagement.vue'
import ProductManagement from '../views/ProductManagement.vue'

const routes = [
  {
    path: '/',
    name: 'Dashboard',
    component: Dashboard
  },
  {
    path: '/loyalty',
    name: 'Loyalty',
    component: LoyaltyManagement
  },
  {
    path: '/coupons',
    name: 'Coupons',
    component: CouponManagement
  },
  {
    path: '/showcase',
    name: 'Showcase',
    component: PictureWarehouseApproval
  },
  {
    path: '/faq',
    name: 'Faq',
    component: FaqManagement
  },
  {
    path: '/orders',
    name: 'Orders',
    component: OrderManagement
  },
  {
    path: '/users',
    name: 'Users',
    component: UserManagement
  },
  {
    path: '/products',
    name: 'Products',
    component: ProductManagement
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

export default router
