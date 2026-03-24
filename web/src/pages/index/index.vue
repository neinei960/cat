<template>
  <SideLayout>
    <view class="workstation">
      <view class="header">
        <view class="greeting">
          <text class="greeting-text">你好，{{ staffName }}</text>
          <text class="greeting-sub">欢迎回到工作台</text>
        </view>
      </view>

      <view class="stats-section">
        <view class="stat-card" @click="uni.reLaunch({ url: '/pages/appointment/calendar' })">
          <text class="stat-value">{{ overview.today_appointment_count }}</text>
          <text class="stat-label">今日预约</text>
        </view>
        <view class="stat-card primary" @click="uni.reLaunch({ url: '/pages/dashboard/index' })">
          <text class="stat-value">¥{{ overview.today_revenue.toFixed(0) }}</text>
          <text class="stat-label">今日营收</text>
        </view>
        <view class="stat-card" @click="uni.reLaunch({ url: '/pages/appointment/list' })">
          <text class="stat-value">{{ overview.pending_appointments }}</text>
          <text class="stat-label">待处理</text>
        </view>
      </view>
    </view>
  </SideLayout>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useAuthStore } from '@/store/auth'
import SideLayout from '@/components/SideLayout.vue'
import { getDashboardOverview } from '@/api/dashboard'

const authStore = useAuthStore()
const staffName = computed(() => authStore.staffInfo?.name || '员工')

const overview = ref({
  today_revenue: 0,
  today_order_count: 0,
  today_appointment_count: 0,
  today_new_customers: 0,
  pending_appointments: 0,
  total_customers: 0,
})

onMounted(async () => {
  try {
    const res = await getDashboardOverview()
    overview.value = res.data
  } catch {}
})
</script>

<style scoped>
.workstation {
  min-height: 100vh;
  background-color: #F5F6FA;
  padding: 0 32rpx 40rpx;
}
.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 40rpx 0 32rpx;
}
.greeting-text {
  font-size: 36rpx;
  font-weight: 700;
  color: #1F2937;
  display: block;
}
.greeting-sub {
  font-size: 26rpx;
  color: #6B7280;
  margin-top: 8rpx;
  display: block;
}
.stats-section {
  display: flex;
  gap: 20rpx;
  margin-bottom: 40rpx;
}
.stat-card {
  flex: 1;
  background-color: #FFFFFF;
  border-radius: 16rpx;
  padding: 28rpx 20rpx;
  display: flex;
  flex-direction: column;
  align-items: center;
  box-shadow: 0 4rpx 16rpx rgba(0, 0, 0, 0.04);
}
.stat-card.primary {
  background: linear-gradient(135deg, #4F46E5, #7C3AED);
}
.stat-card.primary .stat-value,
.stat-card.primary .stat-label {
  color: #fff;
}
.stat-value {
  font-size: 36rpx;
  font-weight: 700;
  color: #4F46E5;
  margin-bottom: 8rpx;
}
.stat-label {
  font-size: 24rpx;
  color: #6B7280;
}
</style>
