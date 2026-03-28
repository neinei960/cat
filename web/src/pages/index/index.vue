<template>
  <SideLayout>
    <view class="workstation">
      <!-- 顶部问候 -->
      <view class="header">
        <view class="greeting">
          <text class="greeting-text">{{ greetingText }}，{{ staffName }}</text>
          <text class="greeting-sub">{{ todayStr }} {{ weekDay }}</text>
        </view>
      </view>

      <!-- 数据卡片 -->
      <view class="stats-section">
        <view class="stat-card" @click="go('/pages/appointment/calendar')">
          <text class="stat-value">{{ overview.today_appointment_count }}</text>
          <text class="stat-label">今日预约</text>
        </view>
        <view class="stat-card primary" @click="go('/pages/dashboard/index')">
          <text class="stat-value">¥{{ overview.today_revenue.toFixed(0) }}</text>
          <text class="stat-label">今日营收</text>
        </view>
        <view class="stat-card" @click="go('/pages/appointment/list')">
          <text class="stat-value">{{ overview.pending_appointments }}</text>
          <text class="stat-label">待处理</text>
        </view>
      </view>

      <!-- 第二行数据 -->
      <view class="stats-section">
        <view class="stat-card small">
          <text class="stat-value-sm">{{ overview.today_new_customers }}</text>
          <text class="stat-label">今日新客</text>
        </view>
        <view class="stat-card small">
          <text class="stat-value-sm">{{ overview.today_order_count }}</text>
          <text class="stat-label">今日订单</text>
        </view>
        <view class="stat-card small">
          <text class="stat-value-sm">{{ overview.total_customers }}</text>
          <text class="stat-label">总客户数</text>
        </view>
      </view>

      <!-- 快捷操作 -->
      <view class="section">
        <text class="section-title">快捷操作</text>
        <view class="quick-actions">
          <view class="action-item" @click="go('/pages/order/create')">
            <view class="action-icon" style="background: linear-gradient(135deg, #6366F1, #4F46E5);">🧾</view>
            <text class="action-label">开单</text>
          </view>
          <view class="action-item" @click="go('/pages/appointment/create')">
            <view class="action-icon" style="background: linear-gradient(135deg, #F59E0B, #D97706);">📅</view>
            <text class="action-label">新建预约</text>
          </view>
          <view class="action-item" @click="go('/pages/pet/list')">
            <view class="action-icon" style="background: linear-gradient(135deg, #10B981, #059669);">🐱</view>
            <text class="action-label">查找猫咪</text>
          </view>
          <view class="action-item" @click="go('/pages/customer/list')">
            <view class="action-icon" style="background: linear-gradient(135deg, #EC4899, #DB2777);">👥</view>
            <text class="action-label">客户管理</text>
          </view>
        </view>
      </view>

      <!-- 今日预约预览 -->
      <view class="section">
        <view class="section-header">
          <text class="section-title">今日预约</text>
          <text class="section-more" @click="go('/pages/appointment/calendar')">查看全部 ›</text>
        </view>
        <view v-if="todayAppts.length === 0" class="no-data">
          <text class="no-data-text">今日暂无预约</text>
        </view>
        <view v-else class="appt-list">
          <view
            class="appt-item"
            v-for="appt in todayAppts.slice(0, 5)"
            :key="appt.ID"
            @click="go('/pages/appointment/detail?id=' + appt.ID)"
          >
            <view class="appt-time-col">
              <text class="appt-time">{{ appt.start_time }}</text>
              <text class="appt-end">{{ appt.end_time }}</text>
            </view>
            <view class="appt-info">
              <text class="appt-pet">{{ getPetName(appt) }}</text>
              <text class="appt-customer">{{ getCustomerName(appt) }}</text>
            </view>
            <view class="appt-status-dot" :style="{ background: statusColors[appt.status] || '#9CA3AF' }"></view>
          </view>
        </view>
      </view>
    </view>
  </SideLayout>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { onShow } from '@dcloudio/uni-app'
import { useAuthStore } from '@/store/auth'
import SideLayout from '@/components/SideLayout.vue'
import { getDashboardOverview } from '@/api/dashboard'
import { getAppointmentCalendar } from '@/api/appointment'

const authStore = useAuthStore()
const staffName = computed(() => authStore.staffInfo?.name || '员工')

const now = new Date()
const todayStr = `${now.getMonth() + 1}月${now.getDate()}日`
const weekDays = ['周日', '周一', '周二', '周三', '周四', '周五', '周六']
const weekDay = weekDays[now.getDay()]
const hour = now.getHours()
const greetingText = hour < 12 ? '早上好' : hour < 18 ? '下午好' : '晚上好'

const statusColors: Record<number, string> = {
  0: '#D97706', 1: '#4338CA', 2: '#059669', 3: '#0284C7', 4: '#6B7280', 5: '#DC2626',
}

const overview = ref({
  today_revenue: 0, today_order_count: 0, today_appointment_count: 0,
  today_new_customers: 0, pending_appointments: 0, total_customers: 0,
})
const todayAppts = ref<any[]>([])

function go(url: string) {
  uni.navigateTo({ url })
}

function localDateStr(d: Date = new Date()) {
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`
}

function getPetName(appt: any) {
  if (appt.pets?.length) return appt.pets.map((p: any) => p.pet?.name || '').filter(Boolean).join('、') || '-'
  return appt.pet?.name || '-'
}

function getCustomerName(appt: any) {
  return appt.customer?.nickname || appt.customer?.phone || '-'
}

async function loadData() {
  try {
    const [ovRes, apptRes] = await Promise.all([
      getDashboardOverview(),
      getAppointmentCalendar(localDateStr(), localDateStr()),
    ])
    overview.value = ovRes.data
    todayAppts.value = (apptRes.data || [])
      .filter((a: any) => a.status !== 4)
      .sort((a: any, b: any) => a.start_time.localeCompare(b.start_time))
  } catch {}
}

onShow(loadData)
</script>

<style scoped>
.workstation { min-height: 100vh; background-color: #F5F6FA; padding: 0 32rpx 40rpx; }
.header { padding: 40rpx 0 24rpx; }
.greeting-text { font-size: 36rpx; font-weight: 700; color: #1F2937; display: block; }
.greeting-sub { font-size: 24rpx; color: #9CA3AF; margin-top: 6rpx; display: block; }

.stats-section { display: flex; gap: 16rpx; margin-bottom: 20rpx; }
.stat-card { flex: 1; background: #fff; border-radius: 16rpx; padding: 24rpx 16rpx; display: flex; flex-direction: column; align-items: center; box-shadow: 0 4rpx 16rpx rgba(0,0,0,0.04); }
.stat-card.primary { background: linear-gradient(135deg, #4F46E5, #7C3AED); }
.stat-card.primary .stat-value, .stat-card.primary .stat-label { color: #fff; }
.stat-card.small { padding: 20rpx 16rpx; }
.stat-value { font-size: 36rpx; font-weight: 700; color: #4F46E5; margin-bottom: 6rpx; }
.stat-value-sm { font-size: 30rpx; font-weight: 700; color: #1F2937; margin-bottom: 4rpx; }
.stat-label { font-size: 22rpx; color: #9CA3AF; }

.section { background: #fff; border-radius: 16rpx; padding: 24rpx; margin-bottom: 20rpx; box-shadow: 0 4rpx 16rpx rgba(0,0,0,0.04); }
.section-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 20rpx; }
.section-title { font-size: 28rpx; font-weight: 600; color: #1F2937; }
.section-more { font-size: 24rpx; color: #6366F1; }

.quick-actions { display: flex; justify-content: space-around; gap: 16rpx; }
.action-item { display: flex; flex-direction: column; align-items: center; gap: 10rpx; }
.action-icon { width: 96rpx; height: 96rpx; border-radius: 24rpx; display: flex; align-items: center; justify-content: center; font-size: 40rpx; }
.action-label { font-size: 24rpx; color: #4B5563; }

.no-data { text-align: center; padding: 40rpx 0; }
.no-data-text { font-size: 26rpx; color: #9CA3AF; }

.appt-list { display: flex; flex-direction: column; gap: 2rpx; }
.appt-item { display: flex; align-items: center; gap: 20rpx; padding: 20rpx 0; border-bottom: 1rpx solid #F3F4F6; }
.appt-item:last-child { border-bottom: none; }
.appt-time-col { display: flex; flex-direction: column; align-items: center; min-width: 80rpx; }
.appt-time { font-size: 28rpx; font-weight: 600; color: #1F2937; }
.appt-end { font-size: 20rpx; color: #9CA3AF; }
.appt-info { flex: 1; min-width: 0; }
.appt-pet { font-size: 28rpx; font-weight: 500; color: #1F2937; display: block; }
.appt-customer { font-size: 22rpx; color: #9CA3AF; display: block; margin-top: 4rpx; }
.appt-status-dot { width: 16rpx; height: 16rpx; border-radius: 50%; flex-shrink: 0; }
</style>
