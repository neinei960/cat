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
          <view class="menu-icon" :class="{ 'menu-icon-cat': item.catIcon }">
            <image v-if="item.catIcon" class="cat-sticker" :src="catSticker" mode="aspectFit" />
            <text v-else>{{ item.icon }}</text>
          </view>
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
          <view class="tab-icon" :class="{ 'tab-icon-cat': item.catIcon }">
            <image v-if="item.catIcon" class="cat-sticker" :src="catSticker" mode="aspectFit" />
            <text v-else>{{ item.icon }}</text>
          </view>
        </view>
        <view v-else class="tab-icon" :class="{ 'tab-icon-cat': item.catIcon }">
          <image v-if="item.catIcon" class="cat-sticker" :src="catSticker" mode="aspectFit" />
          <text v-else>{{ item.icon }}</text>
        </view>
        <text class="tab-label" :style="item.highlight ? 'color: var(--cat-color-primary-deep);' : ''">{{ item.label }}</text>
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
import { hasStaffRoleAtLeast } from '@/utils/staff-role'
import catSticker from '@/assets/cat-sticker.jpg'

const authStore = useAuthStore()
const staffName = computed(() => authStore.staffInfo?.name || '员工')
const currentRole = computed(() => authStore.staffInfo?.role || 'staff')
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
  { icon: '🐱', label: '猫咪管理', path: '/pages/pet/list', catIcon: true },
  { icon: '👥', label: '客户管理', path: '/pages/customer/list' },
  { icon: '📋', label: '订单管理', path: '/pages/order/list' },
  { icon: '✂️', label: '服务管理', path: '/pages/service/list', minRole: 'admin' },
  { icon: '📦', label: '商品管理', path: '/pages/product/list' },
  { icon: '🛵', label: '上门喂养', path: '/pages/feeding/dashboard' },
  { icon: '🏨', label: '寄养看板', path: '/pages/boarding/dashboard' },
  { icon: '🧑‍💼', label: '员工管理', path: '/pages/staff/list', minRole: 'manager' },
  { icon: '📊', label: '数据看板', path: '/pages/dashboard/index', minRole: 'manager' },
  { icon: '💳', label: '会员卡', path: '/pages/setting/member-card', minRole: 'manager' },
  { icon: '⚙️', label: '店铺设置', path: '/pages/shop/settings', minRole: 'admin' },
]
const menuItems = computed(() => allMenuItems.filter(m => !m.minRole || hasStaffRoleAtLeast(currentRole.value, m.minRole as any)))

const tabItems = [
  { icon: '🏠', label: '工作台', path: '/pages/index/index' },
  { icon: '🧾', label: '开单', path: '/pages/order/create', highlight: true },
  { icon: '📅', label: '预约', path: '/pages/appointment/calendar' },
  { icon: '🐱', label: '猫咪', path: '/pages/pet/list', catIcon: true },
  { icon: '📋', label: '订单', path: '/pages/order/list' },
]

const allMoreItems = [
  { icon: '📋', label: '预约列表', path: '/pages/appointment/list' },
  { icon: '👥', label: '客户管理', path: '/pages/customer/list' },
  { icon: '✂️', label: '服务管理', path: '/pages/service/list', minRole: 'admin' },
  { icon: '📦', label: '商品管理', path: '/pages/product/list' },
  { icon: '🛵', label: '上门喂养', path: '/pages/feeding/dashboard' },
  { icon: '🏨', label: '寄养看板', path: '/pages/boarding/dashboard' },
  { icon: '🧑‍💼', label: '员工管理', path: '/pages/staff/list', minRole: 'manager' },
  { icon: '📊', label: '数据看板', path: '/pages/dashboard/index', minRole: 'manager' },
  { icon: '💳', label: '会员卡', path: '/pages/setting/member-card', minRole: 'manager' },
  { icon: '⚙️', label: '店铺设置', path: '/pages/shop/settings', minRole: 'admin' },
]
const moreItems = computed(() => allMoreItems.filter(m => !m.minRole || hasStaffRoleAtLeast(currentRole.value, m.minRole as any)))

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
  background: linear-gradient(180deg, var(--cat-color-sidebar-bg) 0%, #3d2d1b 100%);
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
  border-bottom: 1px solid rgba(255, 244, 223, 0.12);
}

.logo-text {
  font-size: 18px;
  font-weight: 700;
  color: #fff9ef;
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
  background-color: var(--cat-color-sidebar-hover);
}

.menu-item.active {
  background: linear-gradient(135deg, var(--cat-color-primary-soft) 0%, #fff1ca 100%);
  border-radius: 12px;
  margin: 0 8px;
  position: relative;
  box-shadow: inset 0 0 0 1px rgba(160, 120, 48, 0.08);
}
.menu-item.active::before {
  content: '';
  position: absolute;
  left: 0;
  top: 50%;
  transform: translateY(-50%);
  width: 4px;
  height: 60%;
  background-color: var(--cat-color-primary);
  border-radius: 2px;
}

.menu-icon {
  font-size: 16px;
  width: 32px;
  height: 32px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 6px;
  background-color: rgba(255, 250, 243, 0.1);
  margin-right: 10px;
}

.menu-icon-cat {
  padding: 2px;
}

.menu-item.active .menu-icon {
  background: linear-gradient(135deg, var(--cat-color-primary) 0%, #efc97c 100%);
}

.menu-label {
  font-size: 14px;
  color: var(--cat-color-sidebar-text);
}

.menu-item.active .menu-label {
  color: var(--cat-color-text-main);
  font-weight: 600;
}

.sidebar-footer {
  padding: 16px;
  border-top: 1px solid rgba(255, 244, 223, 0.12);
}

.user-name {
  font-size: 14px;
  color: var(--cat-color-sidebar-text);
  display: block;
  margin-bottom: 8px;
}

.logout-btn {
  cursor: pointer;
}

.logout-text {
  font-size: 13px;
  color: #d8c09c;
}

.logout-btn:hover .logout-text {
  color: #F87171;
}

.side-layout-content {
  flex: 1;
  margin-left: 200px;
  min-height: 100vh;
  background-color: var(--cat-color-page-bg);
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
  background-color: var(--cat-color-page-bg);
}

.bottom-tabbar {
  position: fixed;
  left: 0;
  right: 0;
  bottom: 0;
  height: calc(50px + env(safe-area-inset-bottom));
  padding-bottom: env(safe-area-inset-bottom);
  background: rgba(255, 250, 243, 0.96);
  backdrop-filter: blur(10px);
  -webkit-backdrop-filter: blur(10px);
  display: flex;
  align-items: center;
  justify-content: space-around;
  border-top: 1px solid var(--cat-color-border);
  z-index: 1000;
  box-shadow: 0 -8px 24px rgba(116, 88, 38, 0.08);
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
  width: 38rpx;
  height: 38rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 8rpx;
  background-color: var(--cat-color-primary-soft);
}

.tab-icon-cat {
  padding: 4rpx;
}

.tab-item.active .tab-icon {
  background-color: var(--cat-color-card-bg-strong);
}

.tab-icon-highlight .tab-icon {
  background: none;
}

.tab-label {
  font-size: 10px;
  color: var(--cat-color-text-muted);
  margin-top: 2px;
}

.tab-item.active .tab-label {
  color: var(--cat-color-primary-deep);
  font-weight: 500;
}

.tab-icon-highlight {
  background: linear-gradient(135deg, var(--cat-color-primary) 0%, #efc97c 100%);
  width: 52px;
  height: 52px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-top: -15px;
  box-shadow: 0 6px 18px rgba(160, 120, 48, 0.32);
}

.tab-item-highlight .tab-label {
  color: var(--cat-color-primary-deep);
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
  background-color: var(--cat-color-card-bg);
  border-radius: 20px 20px 0 0;
  padding: 16px;
  max-height: 70vh;
  overflow-y: auto;
  border-top: 1px solid var(--cat-color-border);
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
  color: var(--cat-color-text-main);
}

.more-close {
  font-size: 16px;
  color: var(--cat-color-text-light);
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
  padding: 10px 0;
  border-radius: 12px;
  transition: background-color 0.15s, transform 0.1s;
  cursor: pointer;
}

.more-item:active {
  background-color: var(--cat-color-primary-soft);
  transform: scale(0.94);
}

.more-icon {
  font-size: 24px;
  margin-bottom: 4px;
}

.cat-sticker {
  width: 100%;
  height: 100%;
  display: block;
}

.more-label {
  font-size: 12px;
  color: var(--cat-color-text-main);
}

.more-footer {
  margin-top: 16px;
  padding-top: 12px;
  border-top: 1px solid var(--cat-color-border);
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.user-name-app {
  font-size: 14px;
  color: var(--cat-color-text-muted);
}

.logout-btn-app {
  padding: 6px 12px;
}

.logout-text-app {
  font-size: 13px;
  color: var(--cat-color-primary-deep);
}
</style>
