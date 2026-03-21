<template>
  <view class="page">
    <view class="result-icon">{{ isSuccess ? '✓' : '✗' }}</view>
    <text class="result-title">{{ isSuccess ? '预约成功' : '预约失败' }}</text>
    <text class="result-desc">{{ isSuccess ? '商家确认后将通知您' : '请稍后重试' }}</text>
    <button class="btn-primary" @click="goAppointments">查看我的预约</button>
    <button class="btn-secondary" @click="goHome">返回首页</button>
  </view>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { onLoad } from '@dcloudio/uni-app'

const isSuccess = ref(true)

onLoad((query) => {
  isSuccess.value = query?.status !== 'fail'
})

function goAppointments() { uni.switchTab({ url: '/pages/appointment/list/index' }) }
function goHome() { uni.switchTab({ url: '/pages/index/index' }) }
</script>

<style scoped>
.page { padding: 80rpx 48rpx; text-align: center; }
.result-icon { font-size: 120rpx; margin-bottom: 24rpx; }
.result-title { font-size: 40rpx; font-weight: bold; color: #1F2937; display: block; margin-bottom: 16rpx; }
.result-desc { font-size: 28rpx; color: #6B7280; display: block; margin-bottom: 48rpx; }
.btn-primary { background: #4F46E5; color: #fff; border-radius: 12rpx; font-size: 30rpx; margin-bottom: 16rpx; }
.btn-secondary { background: #F3F4F6; color: #374151; border-radius: 12rpx; font-size: 30rpx; }
</style>
