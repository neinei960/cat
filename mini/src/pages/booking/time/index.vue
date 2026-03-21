<template>
  <view class="page">
    <view class="title">选择时间</view>

    <view class="section-title">选择日期</view>
    <scroll-view scroll-x class="date-scroll">
      <view class="date-list">
        <view
          v-for="d in dates" :key="d.value"
          :class="['date-item', booking.selectedDate === d.value ? 'selected' : '']"
          @click="selectDate(d.value)"
        >
          <text class="day">{{ d.label }}</text>
          <text class="date">{{ d.value.substring(5) }}</text>
        </view>
      </view>
    </scroll-view>

    <view v-if="loading" class="loading">查询中...</view>
    <view v-else-if="slots.length === 0 && booking.selectedDate" class="empty">该日期无可用时段</view>

    <view class="section-title" v-if="slots.length">选择时间</view>
    <view class="slots-grid">
      <view
        v-for="s in slots" :key="s"
        :class="['slot', booking.selectedTime === s ? 'selected' : '']"
        @click="booking.selectedTime = s"
      >{{ s }}</view>
    </view>

    <button class="btn-next" :disabled="!booking.selectedDate || !booking.selectedTime" @click="next">下一步 - 确认</button>
  </view>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { getSlots } from '../../../api'
import { useBookingStore } from '../../../store/booking'

const booking = useBookingStore()
const loading = ref(false)
const staffSlots = ref<any[]>([])

const dayNames = ['周日', '周一', '周二', '周三', '周四', '周五', '周六']
const dates = computed(() => {
  const result = []
  for (let i = 0; i < 7; i++) {
    const d = new Date()
    d.setDate(d.getDate() + i)
    result.push({
      value: d.toISOString().substring(0, 10),
      label: i === 0 ? '今天' : i === 1 ? '明天' : dayNames[d.getDay()],
    })
  }
  return result
})

const slots = computed(() => {
  const all = new Set<string>()
  for (const ss of staffSlots.value) {
    if (booking.selectedStaff && ss.staff.ID !== booking.selectedStaff.ID) continue
    for (const slot of ss.slots) {
      all.add(slot.start_time)
    }
  }
  return Array.from(all).sort()
})

async function selectDate(date: string) {
  booking.selectedDate = date
  booking.selectedTime = ''
  if (!booking.selectedService) return
  loading.value = true
  try {
    const res = await getSlots(date, booking.selectedService.ID)
    staffSlots.value = res.data || []
  } finally { loading.value = false }
}

function next() { uni.navigateTo({ url: '/pages/booking/confirm/index' }) }

onMounted(() => {
  if (!booking.selectedDate && dates.value.length) {
    selectDate(dates.value[0].value)
  }
})
</script>

<style scoped>
.page { padding: 24rpx; }
.title { font-size: 32rpx; font-weight: bold; color: #1F2937; margin-bottom: 24rpx; }
.section-title { font-size: 28rpx; font-weight: 600; color: #374151; margin: 16rpx 0; }
.date-scroll { white-space: nowrap; margin-bottom: 16rpx; }
.date-list { display: inline-flex; gap: 16rpx; }
.date-item { display: inline-flex; flex-direction: column; align-items: center; padding: 16rpx 24rpx; background: #fff; border: 2rpx solid #E5E7EB; border-radius: 12rpx; min-width: 120rpx; }
.date-item.selected { border-color: #4F46E5; background: #4F46E5; color: #fff; }
.date-item.selected .day, .date-item.selected .date { color: #fff; }
.day { font-size: 24rpx; color: #6B7280; }
.date { font-size: 26rpx; color: #1F2937; margin-top: 4rpx; }
.loading, .empty { text-align: center; padding: 40rpx; color: #9CA3AF; font-size: 26rpx; }
.slots-grid { display: flex; flex-wrap: wrap; gap: 16rpx; margin-bottom: 32rpx; }
.slot { padding: 16rpx 28rpx; background: #fff; border: 2rpx solid #E5E7EB; border-radius: 12rpx; font-size: 28rpx; color: #374151; }
.slot.selected { border-color: #4F46E5; background: #4F46E5; color: #fff; }
.btn-next { background: #4F46E5; color: #fff; border-radius: 12rpx; font-size: 30rpx; }
.btn-next[disabled] { opacity: 0.5; }
</style>
