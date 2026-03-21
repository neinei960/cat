<template>
  <view class="page">
    <view class="title">确认预约</view>

    <view class="card">
      <view class="row"><text class="label">服务</text><text>{{ booking.selectedService?.name }}</text></view>
      <view class="row"><text class="label">技师</text><text>{{ booking.selectedStaff?.name || '系统分配' }}</text></view>
      <view class="row"><text class="label">日期</text><text>{{ booking.selectedDate }}</text></view>
      <view class="row"><text class="label">时间</text><text>{{ booking.selectedTime }}</text></view>
      <view class="row"><text class="label">价格</text><text class="price">¥{{ booking.selectedService?.base_price }}</text></view>
    </view>

    <view class="section-title">选择宠物</view>
    <view class="pet-list">
      <view
        v-for="p in pets" :key="p.ID"
        :class="['pet-card', booking.selectedPet?.ID === p.ID ? 'selected' : '']"
        @click="booking.selectedPet = p"
      >
        <text>{{ p.name }} ({{ p.species }})</text>
      </view>
    </view>

    <view class="form-item">
      <text class="label">备注</text>
      <textarea v-model="booking.notes" placeholder="特殊需求或注意事项" class="textarea" />
    </view>

    <button class="btn-submit" :disabled="!booking.selectedPet" :loading="submitting" @click="onSubmit">确认预约</button>
  </view>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { getPets, createAppointment } from '../../../api'
import { useBookingStore } from '../../../store/booking'

const booking = useBookingStore()
const pets = ref<any[]>([])
const submitting = ref(false)

onMounted(async () => {
  const res = await getPets()
  pets.value = res.data || []
})

async function onSubmit() {
  submitting.value = true
  try {
    await createAppointment({
      pet_id: booking.selectedPet.ID,
      staff_id: booking.selectedStaff?.ID,
      date: booking.selectedDate,
      start_time: booking.selectedTime,
      service_ids: [booking.selectedService.ID],
      notes: booking.notes,
    })
    uni.redirectTo({ url: '/pages/booking/result/index?status=success' })
  } catch (e) {
    uni.redirectTo({ url: '/pages/booking/result/index?status=fail' })
  } finally {
    submitting.value = false
  }
}
</script>

<style scoped>
.page { padding: 24rpx; }
.title { font-size: 32rpx; font-weight: bold; color: #1F2937; margin-bottom: 24rpx; }
.card { background: #fff; border-radius: 16rpx; padding: 24rpx; margin-bottom: 24rpx; }
.row { display: flex; justify-content: space-between; padding: 14rpx 0; border-bottom: 1rpx solid #F3F4F6; font-size: 28rpx; }
.row:last-child { border-bottom: none; }
.label { color: #6B7280; }
.price { color: #4F46E5; font-weight: bold; }
.section-title { font-size: 28rpx; font-weight: 600; color: #374151; margin-bottom: 16rpx; }
.pet-list { display: flex; flex-direction: column; gap: 12rpx; margin-bottom: 24rpx; }
.pet-card { background: #fff; border: 2rpx solid #E5E7EB; border-radius: 12rpx; padding: 20rpx; font-size: 28rpx; }
.pet-card.selected { border-color: #4F46E5; background: #EEF2FF; }
.form-item { margin-bottom: 24rpx; }
.form-item .label { font-size: 28rpx; color: #374151; display: block; margin-bottom: 8rpx; }
.textarea { background: #fff; border-radius: 12rpx; padding: 16rpx; font-size: 28rpx; width: 100%; height: 120rpx; }
.btn-submit { background: #4F46E5; color: #fff; border-radius: 12rpx; font-size: 30rpx; }
.btn-submit[disabled] { opacity: 0.5; }
</style>
