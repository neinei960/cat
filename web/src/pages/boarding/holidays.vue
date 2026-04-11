<template>
  <SideLayout>
    <view class="page">
      <view class="header">
        <text class="title">寄养节假日</text>
      </view>
      <view class="card">
        <text class="section-title">新增节假日</text>
        <picker mode="date" :value="form.holiday_date" @change="form.holiday_date = $event.detail.value">
          <view class="picker">{{ form.holiday_date || '选择日期' }}</view>
        </picker>
        <input v-model="form.name" class="input" placeholder="节假日名称，例如 五一" />
        <view class="btn btn-primary" @click="save">添加</view>
      </view>
      <view class="card">
        <text class="section-title">已配置日期</text>
        <view v-if="holidays.length === 0" class="empty">暂无节假日配置</view>
        <view class="holiday-row" v-for="item in holidays" :key="item.ID">
          <view>
            <text class="holiday-date">{{ item.holiday_date }}</text>
            <text class="holiday-name">{{ item.name || '节假日' }}</text>
          </view>
          <view class="holiday-del" @click="remove(item.ID)">删除</view>
        </view>
      </view>
    </view>
  </SideLayout>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { onShow } from '@dcloudio/uni-app'
import SideLayout from '@/components/SideLayout.vue'
import { createBoardingHoliday, deleteBoardingHoliday, getBoardingHolidays } from '@/api/boarding'

const holidays = ref<BoardingHoliday[]>([])
const form = ref({ holiday_date: '', name: '' })

async function loadData() {
  const res = await getBoardingHolidays()
  holidays.value = res.data || []
}

async function save() {
  if (!form.value.holiday_date) {
    uni.showToast({ title: '请选择日期', icon: 'none' })
    return
  }
  await createBoardingHoliday(form.value)
  uni.showToast({ title: '添加成功', icon: 'success' })
  form.value = { holiday_date: '', name: '' }
  await loadData()
}

async function remove(id: number) {
  await deleteBoardingHoliday(id)
  uni.showToast({ title: '已删除', icon: 'success' })
  await loadData()
}

onShow(loadData)
</script>

<style scoped>
.page { padding: 24rpx; display: flex; flex-direction: column; gap: 20rpx; }
.title { font-size: 34rpx; font-weight: 700; color: #111827; }
.card { background: #fff; border-radius: 18rpx; padding: 24rpx; box-shadow: 0 12rpx 28rpx rgba(15, 23, 42, 0.04); }
.section-title { display: block; font-size: 28rpx; font-weight: 700; color: #111827; margin-bottom: 14rpx; }
.input, .picker { width: 100%; box-sizing: border-box; margin-bottom: 14rpx; background: #F9FAFB; border: 1rpx solid #E5E7EB; border-radius: 12rpx; padding: 0 20rpx; font-size: 26rpx; color: #111827; min-height: 76rpx; display: flex; align-items: center; }
.input :deep(.uni-input-wrapper) {
  width: 100%;
  min-height: 76rpx;
  display: flex;
  align-items: center;
}
.input :deep(.uni-input-input) {
  width: 100%;
  min-height: 40rpx;
  font-size: 26rpx;
  line-height: 40rpx;
  color: #111827;
  text-align: left !important;
}
.input :deep(.uni-input-placeholder) {
  width: 100%;
  font-size: 26rpx;
  color: #9CA3AF;
  text-align: left !important;
}
.btn { display: inline-flex; align-items: center; justify-content: center; padding: 14rpx 24rpx; border-radius: 12rpx; background: #F3F4F6; color: #374151; font-size: 24rpx; }
.btn-primary { background: #4F46E5; color: #fff; }
.holiday-row { display: flex; justify-content: space-between; align-items: center; padding: 16rpx 0; border-bottom: 1rpx solid #F3F4F6; }
.holiday-row:last-child { border-bottom: none; }
.holiday-date { display: block; font-size: 26rpx; font-weight: 600; color: #111827; }
.holiday-name { display: block; margin-top: 6rpx; font-size: 22rpx; color: #6B7280; }
.holiday-del { font-size: 24rpx; color: #DC2626; }
.empty { color: #9CA3AF; font-size: 24rpx; }
</style>
