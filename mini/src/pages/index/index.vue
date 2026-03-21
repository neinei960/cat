<template>
  <view class="page">
    <view class="shop-header">
      <text class="shop-name">宠物洗护中心</text>
      <text class="shop-desc">专业宠物洗护美容服务</text>
    </view>

    <view class="quick-book" @click="goBook">
      <text class="quick-text">立即预约</text>
      <text class="quick-arrow">→</text>
    </view>

    <view class="section-title">热门服务</view>
    <view class="service-list">
      <view class="service-card" v-for="s in services" :key="s.ID" @click="goBookService(s)">
        <text class="svc-name">{{ s.name }}</text>
        <view class="svc-meta">
          <text class="svc-price">¥{{ s.base_price }}</text>
          <text class="svc-duration">{{ s.duration }}分钟</text>
        </view>
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { getServices } from '../../api'
import { useBookingStore } from '../../store/booking'

const services = ref<any[]>([])
const booking = useBookingStore()

onMounted(async () => {
  try {
    const res = await getServices()
    services.value = res.data || []
  } catch (e) {}
})

function goBook() {
  booking.reset()
  uni.navigateTo({ url: '/pages/booking/service/index' })
}

function goBookService(s: any) {
  booking.reset()
  booking.selectedService = s
  uni.navigateTo({ url: '/pages/booking/staff/index' })
}
</script>

<style scoped>
.page { padding: 24rpx; }
.shop-header { background: linear-gradient(135deg, #4F46E5, #7C3AED); border-radius: 20rpx; padding: 48rpx 32rpx; margin-bottom: 24rpx; color: #fff; }
.shop-name { font-size: 40rpx; font-weight: bold; display: block; }
.shop-desc { font-size: 26rpx; opacity: 0.8; display: block; margin-top: 8rpx; }
.quick-book { display: flex; justify-content: space-between; align-items: center; background: #4F46E5; border-radius: 16rpx; padding: 32rpx; margin-bottom: 32rpx; }
.quick-text { font-size: 32rpx; font-weight: bold; color: #fff; }
.quick-arrow { font-size: 36rpx; color: #fff; }
.section-title { font-size: 30rpx; font-weight: 600; color: #1F2937; margin-bottom: 16rpx; }
.service-list { display: flex; flex-direction: column; gap: 16rpx; }
.service-card { background: #fff; border-radius: 16rpx; padding: 24rpx; box-shadow: 0 2rpx 8rpx rgba(0,0,0,0.04); }
.svc-name { font-size: 30rpx; font-weight: 600; color: #1F2937; display: block; }
.svc-meta { display: flex; gap: 16rpx; margin-top: 8rpx; }
.svc-price { font-size: 30rpx; font-weight: bold; color: #4F46E5; }
.svc-duration { font-size: 24rpx; color: #6B7280; line-height: 42rpx; }
</style>
