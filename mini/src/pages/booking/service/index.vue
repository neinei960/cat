<template>
  <view class="page">
    <view class="title">选择服务</view>
    <view class="list">
      <view
        v-for="s in services" :key="s.ID"
        :class="['card', booking.selectedService?.ID === s.ID ? 'selected' : '']"
        @click="select(s)"
      >
        <view class="row">
          <text class="name">{{ s.name }}</text>
          <text class="price">¥{{ s.base_price }}</text>
        </view>
        <text class="desc">{{ s.category }} · {{ s.duration }}分钟</text>
      </view>
    </view>
    <button class="btn-next" :disabled="!booking.selectedService" @click="next">下一步 - 选技师</button>
  </view>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { getServices } from '../../../api'
import { useBookingStore } from '../../../store/booking'

const services = ref<any[]>([])
const booking = useBookingStore()

onMounted(async () => {
  const res = await getServices()
  services.value = res.data || []
})

function select(s: any) { booking.selectedService = s }
function next() { uni.navigateTo({ url: '/pages/booking/staff/index' }) }
</script>

<style scoped>
.page { padding: 24rpx; }
.title { font-size: 32rpx; font-weight: bold; color: #1F2937; margin-bottom: 24rpx; }
.list { display: flex; flex-direction: column; gap: 16rpx; margin-bottom: 32rpx; }
.card { background: #fff; border: 2rpx solid #E5E7EB; border-radius: 16rpx; padding: 24rpx; }
.card.selected { border-color: #4F46E5; background: #EEF2FF; }
.row { display: flex; justify-content: space-between; }
.name { font-size: 30rpx; font-weight: 600; color: #1F2937; }
.price { font-size: 30rpx; font-weight: bold; color: #4F46E5; }
.desc { font-size: 24rpx; color: #6B7280; margin-top: 8rpx; }
.btn-next { background: #4F46E5; color: #fff; border-radius: 12rpx; font-size: 30rpx; }
.btn-next[disabled] { opacity: 0.5; }
</style>
