<template>
  <view class="page">
    <view class="title">我的订单</view>
    <view v-if="list.length===0" class="empty">暂无订单</view>
    <view class="list">
      <view class="card" v-for="o in list" :key="o.ID" @click="goDetail(o.ID)">
        <view class="row"><text class="no">{{ o.order_no }}</text><text :class="['status',`s${o.status}`]">{{ statusMap[o.status] }}</text></view>
        <text class="items">{{ (o.items||[]).map((i:any)=>i.name).join(', ') }}</text>
        <text class="amount">¥{{ o.pay_amount }}</text>
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
// Orders are accessed via B-end for now; C-end order list would need a separate endpoint
// For MVP, show empty
const list = ref<any[]>([])
const statusMap: Record<number,string> = {0:'待付款',1:'已完成',2:'已取消',3:'已退款'}
function goDetail(id:number) { uni.navigateTo({url:`/pages/order/detail/index?id=${id}`}) }
onMounted(()=>{})
</script>

<style scoped>
.page{padding:24rpx;}.title{font-size:36rpx;font-weight:bold;color:#1F2937;margin-bottom:24rpx;}.empty{text-align:center;padding:100rpx 0;color:#9CA3AF;font-size:28rpx;}
.card{background:#fff;border-radius:16rpx;padding:24rpx;margin-bottom:16rpx;}.row{display:flex;justify-content:space-between;margin-bottom:8rpx;}
.no{font-size:24rpx;color:#9CA3AF;}.status{font-size:22rpx;padding:4rpx 12rpx;border-radius:12rpx;}
.s0{background:#FEF3C7;color:#92400E;}.s1{background:#D1FAE5;color:#059669;}.s2,.s3{background:#F3F4F6;color:#6B7280;}
.items{font-size:26rpx;color:#374151;display:block;}.amount{font-size:30rpx;font-weight:bold;color:#4F46E5;display:block;margin-top:8rpx;}
</style>
