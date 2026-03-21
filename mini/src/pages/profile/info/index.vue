<template>
  <view class="page">
    <view class="card">
      <view class="row"><text class="label">昵称</text><text>{{ auth.customerInfo?.nickname || '-' }}</text></view>
      <view class="row"><text class="label">手机号</text><text>{{ auth.customerInfo?.phone || '未绑定' }}</text></view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { useAuthStore } from '../../../store/auth'
const auth = useAuthStore()
</script>

<style scoped>
.page{padding:24rpx;}
.card{background:#fff;border-radius:16rpx;padding:24rpx;}
.row{display:flex;justify-content:space-between;padding:16rpx 0;border-bottom:1rpx solid #F3F4F6;font-size:28rpx;}
.row:last-child{border-bottom:none;}.label{color:#6B7280;}
</style>
