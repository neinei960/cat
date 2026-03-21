<template>
  <view class="page">
    <view class="profile-card">
      <view class="avatar">{{ nickname.charAt(0) || '?' }}</view>
      <view class="info">
        <text class="name">{{ nickname || '未登录' }}</text>
        <text class="phone">{{ phone || '点击登录' }}</text>
      </view>
    </view>

    <view class="menu">
      <view class="menu-item" @click="go('/pages/appointment/list/index')">
        <text>我的预约</text><text class="arrow">></text>
      </view>
      <view class="menu-item" @click="go('/pages/pet/list/index')">
        <text>我的宠物</text><text class="arrow">></text>
      </view>
      <view class="menu-item" @click="go('/pages/order/list/index')">
        <text>我的订单</text><text class="arrow">></text>
      </view>
      <view class="menu-item" @click="go('/pages/profile/info/index')">
        <text>个人信息</text><text class="arrow">></text>
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useAuthStore } from '../../../store/auth'
const auth = useAuthStore()
const nickname = computed(() => auth.customerInfo?.nickname || '')
const phone = computed(() => auth.customerInfo?.phone || '')
function go(url: string) { uni.navigateTo({ url }) }
</script>

<style scoped>
.page{padding:24rpx;}
.profile-card{display:flex;align-items:center;gap:24rpx;background:#fff;border-radius:16rpx;padding:32rpx;margin-bottom:24rpx;}
.avatar{width:100rpx;height:100rpx;border-radius:50%;background:#4F46E5;color:#fff;display:flex;align-items:center;justify-content:center;font-size:40rpx;font-weight:bold;}
.name{font-size:32rpx;font-weight:600;color:#1F2937;display:block;}.phone{font-size:26rpx;color:#6B7280;display:block;margin-top:4rpx;}
.menu{background:#fff;border-radius:16rpx;overflow:hidden;}
.menu-item{display:flex;justify-content:space-between;padding:28rpx 24rpx;border-bottom:1rpx solid #F3F4F6;font-size:28rpx;color:#1F2937;}
.menu-item:last-child{border-bottom:none;}.arrow{color:#9CA3AF;}
</style>
