<template>
  <view class="page">
    <view class="header">
      <text class="title">我的宠物</text>
      <view class="btn-add" @click="goAdd">+ 添加</view>
    </view>
    <view v-if="pets.length === 0" class="empty">还没有添加宠物</view>
    <view v-else class="list">
      <view class="card" v-for="p in pets" :key="p.ID" @click="goEdit(p.ID)">
        <view class="icon">{{ p.species === '犬' ? '🐕' : '🐱' }}</view>
        <view class="info">
          <text class="name">{{ p.name }}</text>
          <text class="breed">{{ p.species }} · {{ p.breed || '未知品种' }}</text>
        </view>
        <text class="weight" v-if="p.weight">{{ p.weight }}kg</text>
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { onShow } from '@dcloudio/uni-app'
import { getPets } from '../../../api'
const pets = ref<any[]>([])
async function load() { const r = await getPets(); pets.value = r.data || [] }
function goAdd() { uni.navigateTo({ url: '/pages/pet/edit/index' }) }
function goEdit(id: number) { uni.navigateTo({ url: `/pages/pet/edit/index?id=${id}` }) }
onMounted(load); onShow(load)
</script>

<style scoped>
.page{padding:24rpx;}.header{display:flex;justify-content:space-between;align-items:center;margin-bottom:24rpx;}.title{font-size:36rpx;font-weight:bold;color:#1F2937;}
.btn-add{font-size:28rpx;color:#4F46E5;}.empty{text-align:center;padding:100rpx 0;color:#9CA3AF;font-size:28rpx;}
.card{background:#fff;border-radius:16rpx;padding:24rpx;margin-bottom:16rpx;display:flex;align-items:center;gap:16rpx;}
.icon{font-size:48rpx;}.info{flex:1;}.name{font-size:30rpx;font-weight:600;color:#1F2937;display:block;}.breed{font-size:24rpx;color:#6B7280;display:block;margin-top:4rpx;}.weight{font-size:26rpx;color:#4F46E5;}
</style>
