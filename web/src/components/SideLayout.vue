<template>
  <!-- 桌面端：侧边栏 -->
  <view class="side-layout" v-if="isDesktop">
    <view class="sidebar">
      <view class="sidebar-logo">
        <text class="logo-text">猫咪洗护</text>
      </view>
      <view class="sidebar-menu">
        <view
          v-for="item in menuItems"
          :key="item.path"
          class="menu-item"
          :class="{ active: isActive(item.path) }"
          @click="navigate(item.path)"
        >
          <text class="menu-icon">{{ item.icon }}</text>
          <text class="menu-label">{{ item.label }}</text>
        </view>
      </view>
      <view class="sidebar-footer">
        <text class="user-name">{{ staffName }}</text>
        <view class="logout-btn" @click="handleLogout">
          <text class="logout-text">退出登录</text>
        </view>
      </view>
    </view>
    <view class="side-layout-content">
      <slot />
    </view>
  </view>

  <!-- 手机端：底部导航栏 -->
  <view class="app-layout" v-else>
    <view class="app-content">
      <slot />
    </view>
    <view class="bottom-tabbar">
      <view
        v-for="item in tabItems"
        :key="item.path"
        class="tab-item"
        :class="{ active: isActive(item.path), 'tab-item-highlight': item.highlight }"
        @click="navigate(item.path)"
      >
        <view v-if="item.highlight" class="tab-icon-highlight">
          <text class="tab-icon">{{ item.icon }}</text>
        </view>
        <text v-else class="tab-icon">{{ item.icon }}</text>
        <text class="tab-label" :style="item.highlight ? 'color: #4F46E5;' : ''">{{ item.label }}</text>
      </view>
      <view class="tab-item" @click="showMoreMenu = true">
        <text class="tab-icon">⚙️</text>
        <text class="tab-label">更多</text>
      </view>
    </view>

    <!-- 更多菜单弹出层 -->
    <view class="more-mask" v-if="showMoreMenu" @click="showMoreMenu = false">
      <view class="more-panel" @click.stop>
        <view class="more-header">
          <text class="more-title">更多功能</text>
          <text class="more-close" @click="showMoreMenu = false">✕</text>
        </view>
        <view class="more-grid">
          <view
            v-for="item in moreItems"
            :key="item.path"
            class="more-item"
            @click="navigateMore(item.path)"
          >
            <text class="more-icon">{{ item.icon }}</text>
            <text class="more-label">{{ item.label }}</text>
          </view>
        </view>
        <view class="more-footer">
          <text class="user-name-app">{{ staffName }}</text>
          <view class="logout-btn-app" @click="handleLogout">
            <text class="logout-text-app">退出登录</text>
          </view>
        </view>
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useAuthStore } from '@/store/auth'

const authStore = useAuthStore()
const staffName = computed(() => authStore.staffInfo?.name || '员工')
const isAdmin = computed(() => authStore.staffInfo?.role === 'admin')
const showMoreMenu = ref(false)

const screenWidth = ref(800)
const isDesktop = computed(() => screenWidth.value >= 768)

function updateScreenWidth() {
  try {
    const info = uni.getSystemInfoSync()
    screenWidth.value = info.windowWidth
  } catch {
    screenWidth.value = 800
  }
}

// #ifdef H5
function onResize() {
  screenWidth.value = window.innerWidth
}
// #endif

onMounted(() => {
  updateScreenWidth()
  // #ifdef H5
  window.addEventListener('resize', onResize)
  // #endif
})

onUnmounted(() => {
  // #ifdef H5
  window.removeEventListener('resize', onResize)
  // #endif
})

const allMenuItems = [
  { icon: '🏠', label: '工作台', path: '/pages/index/index' },
  { icon: '🧾', label: '开单', path: '/pages/order/create' },
  { icon: '📅', label: '预约日历', path: '/pages/appointment/calendar' },
  { icon: '📋', label: '预约列表', path: '/pages/appointment/list' },
  { icon: '🐱', label: '猫咪管理', path: '/pages/pet/list' },
  { icon: '👥', label: '客户管理', path: '/pages/customer/list' },
  { icon: '📋', label: '订单管理', path: '/pages/order/list' },
  { icon: '✂️', label: '服务管理', path: '/pages/service/list', adminOnly: true },
  { icon: '📦', label: '商品管理', path: '/pages/product/list' },
  { icon: '🧑‍💼', label: '员工管理', path: '/pages/staff/list', adminOnly: true },
  { icon: '📊', label: '数据看板', path: '/pages/dashboard/index', adminOnly: true },
  { icon: '💳', label: '会员卡', path: '/pages/setting/member-card', adminOnly: true },
  { icon: '⚙️', label: '店铺设置', path: '/pages/shop/settings', adminOnly: true },
]
const menuItems = computed(() => allMenuItems.filter(m => !m.adminOnly || isAdmin.value))

const tabItems = [
  { icon: '🏠', label: '工作台', path: '/pages/index/index' },
  { icon: '🧾', label: '开单', path: '/pages/order/create', highlight: true },
  { icon: '📅', label: '预约', path: '/pages/appointment/calendar' },
  { icon: '🐱', label: '猫咪', path: '/pages/pet/list' },
  { icon: '📋', label: '订单', path: '/pages/order/list' },
]

const allMoreItems = [
  { icon: '📋', label: '预约列表', path: '/pages/appointment/list' },
  { icon: '👥', label: '客户管理', path: '/pages/customer/list' },
  { icon: '✂️', label: '服务管理', path: '/pages/service/list', adminOnly: true },
  { icon: '📦', label: '商品管理', path: '/pages/product/list' },
  { icon: '🧑‍💼', label: '员工管理', path: '/pages/staff/list', adminOnly: true },
  { icon: '📊', label: '数据看板', path: '/pages/dashboard/index', adminOnly: true },
  { icon: '💳', label: '会员卡', path: '/pages/setting/member-card', adminOnly: true },
  { icon: '⚙️', label: '店铺设置', path: '/pages/shop/settings', adminOnly: true },
]
const moreItems = computed(() => allMoreItems.filter(m => !m.adminOnly || isAdmin.value))

function isActive(path: string): boolean {
  const pages = getCurrentPages()
  if (!pages.length) return false
  const currentRoute = '/' + pages[pages.length - 1].route
  return currentRoute === path
}

function navigate(path: string) {
  if (isActive(path)) return
  uni.reLaunch({ url: path })
}

function navigateMore(path: string) {
  showMoreMenu.value = false
  if (isActive(path)) return
  uni.reLaunch({ url: path })
}

function handleLogout() {
  uni.showModal({
    title: '提示',
    content: '确定要退出登录吗？',
    success: (res) => {
      if (res.confirm) {
        authStore.logout()
      }
    },
  })
}
</script>

<style scoped>
/* ========== 桌面端侧边栏 ========== */
.side-layout {
  display: flex;
  min-height: 100vh;
}

.sidebar {
  width: 200px;
  min-width: 200px;
  background-color: #1F2937;
  display: flex;
  flex-direction: column;
  position: fixed;
  top: 0;
  left: 0;
  bottom: 0;
  z-index: 1000;
  overflow-y: auto;
}

.sidebar-logo {
  padding: 24px 16px;
  border-bottom: 1px solid #374151;
}

.logo-text {
  font-size: 18px;
  font-weight: 700;
  color: #FFFFFF;
}

.sidebar-menu {
  flex: 1;
  padding: 8px 0;
}

.menu-item {
  display: flex;
  align-items: center;
  padding: 12px 16px;
  cursor: pointer;
  transition: background-color 0.15s;
}

.menu-item:hover {
  background-color: #374151;
}

.menu-item.active {
  background-color: #4F46E5;
}

.menu-icon {
  font-size: 16px;
  margin-right: 10px;
  width: 24px;
  text-align: center;
}

.menu-label {
  font-size: 14px;
  color: #D1D5DB;
}

.menu-item.active .menu-label {
  color: #FFFFFF;
  font-weight: 500;
}

.sidebar-footer {
  padding: 16px;
  border-top: 1px solid #374151;
}

.user-name {
  font-size: 14px;
  color: #D1D5DB;
  display: block;
  margin-bottom: 8px;
}

.logout-btn {
  cursor: pointer;
}

.logout-text {
  font-size: 13px;
  color: #9CA3AF;
}

.logout-btn:hover .logout-text {
  color: #F87171;
}

.side-layout-content {
  flex: 1;
  margin-left: 200px;
  min-height: 100vh;
  background-color: #F5F6FA;
}

/* ========== 手机端底部导航 ========== */
.app-layout {
  display: flex;
  flex-direction: column;
  min-height: 100vh;
}

.app-content {
  flex: 1;
  padding-bottom: calc(50px + env(safe-area-inset-bottom));
  background-color: #F5F6FA;
}

.bottom-tabbar {
  position: fixed;
  left: 0;
  right: 0;
  bottom: 0;
  height: calc(50px + env(safe-area-inset-bottom));
  padding-bottom: env(safe-area-inset-bottom);
  background: rgba(255, 255, 255, 0.95);
  backdrop-filter: blur(10px);
  -webkit-backdrop-filter: blur(10px);
  display: flex;
  align-items: center;
  justify-content: space-around;
  border-top: 1px solid #E5E7EB;
  z-index: 1000;
}

.tab-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  flex: 1;
  height: 100%;
}

.tab-icon {
  font-size: 18px;
  line-height: 1;
}

.tab-label {
  font-size: 10px;
  color: #6B7280;
  margin-top: 2px;
}

.tab-item.active .tab-label {
  color: #4F46E5;
  font-weight: 500;
}

.tab-icon-highlight {
  background: linear-gradient(135deg, #6366F1, #4F46E5);
  width: 52px;
  height: 52px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-top: -15px;
  box-shadow: 0 4px 12px rgba(79, 70, 229, 0.3);
}

.tab-item-highlight .tab-label {
  color: #4F46E5;
  font-weight: 500;
}

/* ========== 更多弹出面板 ========== */
.more-mask {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: rgba(0, 0, 0, 0.5);
  z-index: 2000;
  display: flex;
  align-items: flex-end;
}

.more-panel {
  width: 100%;
  background-color: #FFFFFF;
  border-radius: 12px 12px 0 0;
  padding: 16px;
  max-height: 70vh;
  overflow-y: auto;
}

.more-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}

.more-title {
  font-size: 16px;
  font-weight: 600;
  color: #1F2937;
}

.more-close {
  font-size: 16px;
  color: #9CA3AF;
  padding: 4px;
}

.more-grid {
  display: flex;
  flex-wrap: wrap;
}

.more-item {
  width: 25%;
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 12px 0;
  border-radius: 12px;
  transition: background-color 0.15s, transform 0.1s;
  cursor: pointer;
}

.more-item:active {
  background-color: #EEF2FF;
  transform: scale(0.94);
}

.more-icon {
  font-size: 24px;
  margin-bottom: 4px;
}

.more-label {
  font-size: 12px;
  color: #374151;
}

.more-footer {
  margin-top: 16px;
  padding-top: 12px;
  border-top: 1px solid #E5E7EB;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.user-name-app {
  font-size: 14px;
  color: #6B7280;
}

.logout-btn-app {
  padding: 6px 12px;
}

.logout-text-app {
  font-size: 13px;
  color: #EF4444;
}
</style>
