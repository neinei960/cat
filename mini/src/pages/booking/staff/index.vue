<template>
  <view class="page">
    <view class="title">选择技师</view>
    <view class="list">
      <view
        v-for="s in staffs" :key="s.ID"
        :class="['card', booking.selectedStaff?.ID === s.ID ? 'selected' : '']"
        @click="booking.selectedStaff = s"
      >
        <view class="avatar">{{ s.name.charAt(0) }}</view>
        <text class="name">{{ s.name }}</text>
      </view>
    </view>
    <view :class="['card', !booking.selectedStaff ? 'selected' : '']" @click="booking.selectedStaff = null">
      <text class="name">不指定（系统分配）</text>
    </view>
    <button class="btn-next" @click="next" style="margin-top: 32rpx;">下一步 - 选时间</button>
  </view>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { getStaffs } from '../../../api'
import { useBookingStore } from '../../../store/booking'

const staffs = ref<any[]>([])
const booking = useBookingStore()

onMounted(async () => {
  if (booking.selectedService) {
    const res = await getStaffs(booking.selectedService.ID)
    staffs.value = res.data || []
  }
})

function next() { uni.navigateTo({ url: '/pages/booking/time/index' }) }
</script>

<style scoped>
.page { padding: 24rpx; }
.title { font-size: 32rpx; font-weight: bold; color: #1F2937; margin-bottom: 24rpx; }
.list { display: flex; flex-direction: column; gap: 16rpx; margin-bottom: 16rpx; }
.card { background: #fff; border: 2rpx solid #E5E7EB; border-radius: 16rpx; padding: 24rpx; display: flex; align-items: center; gap: 16rpx; }
.card.selected { border-color: #4F46E5; background: #EEF2FF; }
.avatar { width: 72rpx; height: 72rpx; border-radius: 50%; background: #4F46E5; color: #fff; display: flex; align-items: center; justify-content: center; font-size: 28rpx; font-weight: bold; }
.name { font-size: 28rpx; font-weight: 600; color: #1F2937; }
.btn-next { background: #4F46E5; color: #fff; border-radius: 12rpx; font-size: 30rpx; }
</style>
