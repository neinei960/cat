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
        <view class="stat-card">
          <text class="stat-value">12</text>
          <text class="stat-label">今日预约数</text>
        </view>
        <view class="stat-card">
          <text class="stat-value">¥2,480</text>
          <text class="stat-label">今日营收</text>
        </view>
        <view class="stat-card">
          <text class="stat-value">3</text>
          <text class="stat-label">待确认预约</text>
        </view>
      </view>
    </view>
  </SideLayout>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useAuthStore } from '@/store/auth'
import SideLayout from '@/components/SideLayout.vue'

const authStore = useAuthStore()
const staffName = computed(() => authStore.staffInfo?.name || '员工')
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
