<template>
  <SideLayout>
    <view class="page">
      <view class="header">
        <text class="title">寄养节假日</text>
      </view>
      <view class="card">
        <text class="section-title">新增节假日范围</text>
        <picker mode="date" :value="form.start_date" @change="onStartDateChange($event.detail.value)">
          <view class="picker">{{ form.start_date || '选择开始日期' }}</view>
        </picker>
        <picker mode="date" :value="form.end_date" @change="form.end_date = $event.detail.value">
          <view class="picker">{{ form.end_date || '选择结束日期' }}</view>
        </picker>
        <text class="range-tip">开始和结束日期均包含在节假日范围内</text>
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
const form = ref({ start_date: '', end_date: '', name: '' })

async function loadData() {
  const res = await getBoardingHolidays()
  holidays.value = res.data || []
}

function onStartDateChange(value: string) {
  form.value.start_date = value
  if (!form.value.end_date || form.value.end_date < value) {
    form.value.end_date = value
  }
}

async function save() {
  if (!form.value.start_date || !form.value.end_date) {
    uni.showToast({ title: '请选择开始和结束日期', icon: 'none' })
    return
  }
  if (form.value.end_date < form.value.start_date) {
    uni.showToast({ title: '结束日期不能早于开始日期', icon: 'none' })
    return
  }
  const res = await createBoardingHoliday(form.value)
  const created = res.data || []
  if (created.length === 0) {
    uni.showToast({ title: '所选日期已存在', icon: 'none' })
  } else {
    uni.showToast({ title: created.length > 1 ? `已添加${created.length}天` : '添加成功', icon: 'success' })
  }
  form.value = { start_date: '', end_date: '', name: '' }
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
.range-tip { display: block; margin: -2rpx 0 14rpx; font-size: 22rpx; color: #6B7280; }
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
