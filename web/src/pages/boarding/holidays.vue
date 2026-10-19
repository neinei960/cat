<template>
  <SideLayout>
    <view class="page">
      <view class="header">
        <text class="title">寄养节假日</text>
      </view>
      <view class="card">
        <view class="form-title-row">
          <text class="section-title">{{ editingRange ? '修改节假日范围' : '新增节假日范围' }}</text>
          <view v-if="editingRange" class="btn btn-ghost" @click="resetForm">取消修改</view>
        </view>
        <view class="date-grid">
          <picker mode="date" :value="form.start_date" @change="onStartDateChange($event.detail.value)">
            <view class="picker">{{ form.start_date || '选择开始日期' }}</view>
          </picker>
          <picker mode="date" :value="form.end_date" @change="form.end_date = $event.detail.value">
            <view class="picker">{{ form.end_date || '选择结束日期' }}</view>
          </picker>
        </view>
        <text class="range-tip">开始和结束日期均包含在节假日范围内</text>
        <input v-model="form.name" class="input" placeholder="节假日名称，例如 五一" />
        <input v-model="form.surcharge_amount" class="input" type="digit" placeholder="每晚加收金额，例如 30" />
        <view class="btn btn-primary" @click="save">{{ editingRange ? '保存修改' : '添加' }}</view>
      </view>
      <view class="card">
        <text class="section-title">已配置范围</text>
        <view v-if="holidays.length === 0" class="empty">暂无节假日配置</view>
        <view class="holiday-row" v-for="item in holidayRanges" :key="item.key">
          <view>
            <text class="holiday-date">{{ item.dateLabel }}</text>
            <text class="holiday-name">{{ item.name || '节假日' }} · {{ holidaySurchargeLabel(item) }} · {{ item.count }}天</text>
          </view>
          <view class="holiday-actions">
            <view class="holiday-edit" @click="editRange(item)">编辑</view>
            <view class="holiday-del" @click="removeRange(item)">删除</view>
          </view>
        </view>
      </view>
    </view>
  </SideLayout>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { onShow } from '@dcloudio/uni-app'
import SideLayout from '@/components/SideLayout.vue'
import { createBoardingHoliday, deleteBoardingHoliday, getBoardingHolidays, updateBoardingHolidayRange } from '@/api/boarding'

const holidays = ref<BoardingHoliday[]>([])
const form = ref({ start_date: '', end_date: '', name: '', surcharge_amount: '' })
const editingRange = ref<HolidayRange | null>(null)

interface HolidayRange {
  key: string
  ids: number[]
  name: string
  surcharge_amount: number
  start_date: string
  end_date: string
  dateLabel: string
  count: number
}

const holidayRanges = computed<HolidayRange[]>(() => {
  const sorted = [...holidays.value].sort((a, b) => {
    const byDate = String(a.holiday_date || '').localeCompare(String(b.holiday_date || ''))
    if (byDate !== 0) return byDate
    return Number(a.ID || 0) - Number(b.ID || 0)
  })
  const ranges: HolidayRange[] = []
  for (const item of sorted) {
    const date = item.holiday_date
    const name = item.name || '节假日'
    const surchargeAmount = Number(item.surcharge_amount || 0)
    const last = ranges[ranges.length - 1]
    if (
      last &&
      last.name === name &&
      last.surcharge_amount === surchargeAmount &&
      isNextDate(last.end_date, date)
    ) {
      last.ids.push(item.ID)
      last.end_date = date
      last.dateLabel = formatDateRange(last.start_date, last.end_date)
      last.count += 1
      last.key = `${last.start_date}-${last.end_date}-${last.name}-${last.surcharge_amount}`
    } else {
      ranges.push({
        key: `${date}-${item.ID}`,
        ids: [item.ID],
        name,
        surcharge_amount: surchargeAmount,
        start_date: date,
        end_date: date,
        dateLabel: formatDateRange(date, date),
        count: 1,
      })
    }
  }
  return ranges
})

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
  const surchargeAmount = Number(form.value.surcharge_amount || 0)
  if (!Number.isFinite(surchargeAmount) || surchargeAmount < 0) {
    uni.showToast({ title: '请填写有效加收金额', icon: 'none' })
    return
  }
  const payload = {
    start_date: form.value.start_date,
    end_date: form.value.end_date,
    name: form.value.name,
    surcharge_amount: surchargeAmount,
  }
  const res = editingRange.value
    ? await updateBoardingHolidayRange({ ...payload, ids: editingRange.value.ids })
    : await createBoardingHoliday(payload)
  const created = res.data || []
  if (created.length === 0) {
    uni.showToast({ title: '所选日期已存在', icon: 'none' })
  } else {
    uni.showToast({
      title: editingRange.value ? '修改成功' : (created.length > 1 ? `已添加${created.length}天` : '添加成功'),
      icon: 'success',
    })
  }
  resetForm()
  await loadData()
}

function editRange(range: HolidayRange) {
  editingRange.value = range
  form.value = {
    start_date: range.start_date,
    end_date: range.end_date,
    name: range.name,
    surcharge_amount: range.surcharge_amount > 0 ? String(range.surcharge_amount) : '',
  }
}

function resetForm() {
  editingRange.value = null
  form.value = { start_date: '', end_date: '', name: '', surcharge_amount: '' }
}

async function removeRange(range: HolidayRange) {
  await Promise.all(range.ids.map((id) => deleteBoardingHoliday(id)))
  if (editingRange.value?.key === range.key) {
    resetForm()
  }
  uni.showToast({ title: range.ids.length > 1 ? '已删除该范围' : '已删除', icon: 'success' })
  await loadData()
}

function holidaySurchargeLabel(item: Pick<BoardingHoliday, 'surcharge_amount'>) {
  const amount = Number(item.surcharge_amount || 0)
  if (amount > 0) return `加收 ¥${amount.toFixed(2)}/晚`
  return '按默认节假日加收'
}

function parseDateText(dateText: string) {
  const [year, month, day] = String(dateText || '').split('-').map(Number)
  if (!year || !month || !day) return null
  return new Date(year, month - 1, day)
}

function isNextDate(current: string, next: string) {
  const currentDate = parseDateText(current)
  const nextDate = parseDateText(next)
  if (!currentDate || !nextDate) return false
  currentDate.setDate(currentDate.getDate() + 1)
  return formatDate(currentDate) === next
}

function formatDate(date: Date) {
  const year = date.getFullYear()
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  return `${year}-${month}-${day}`
}

function formatDateRange(startDate: string, endDate: string) {
  return startDate === endDate ? startDate : `${startDate} 至 ${endDate}`
}

onShow(loadData)
</script>

<style scoped>
.page { padding: 24rpx; display: flex; flex-direction: column; gap: 20rpx; }
.title { font-size: 34rpx; font-weight: 700; color: #111827; }
.card { background: #fff; border-radius: 18rpx; padding: 24rpx; box-shadow: 0 12rpx 28rpx rgba(15, 23, 42, 0.04); }
.form-title-row { display: flex; align-items: center; justify-content: space-between; gap: 16rpx; margin-bottom: 14rpx; }
.section-title { display: block; font-size: 28rpx; font-weight: 700; color: #111827; margin-bottom: 14rpx; }
.form-title-row .section-title { margin-bottom: 0; }
.date-grid { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 14rpx; }
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
.btn-ghost { padding: 10rpx 16rpx; color: #4F46E5; background: #EEF2FF; white-space: nowrap; }
.holiday-row { display: flex; justify-content: space-between; align-items: center; padding: 16rpx 0; border-bottom: 1rpx solid #F3F4F6; }
.holiday-row:last-child { border-bottom: none; }
.holiday-date { display: block; font-size: 26rpx; font-weight: 600; color: #111827; }
.holiday-name { display: block; margin-top: 6rpx; font-size: 22rpx; color: #6B7280; }
.holiday-actions { display: flex; align-items: center; gap: 18rpx; flex-shrink: 0; }
.holiday-edit { font-size: 24rpx; color: #4F46E5; }
.holiday-del { font-size: 24rpx; color: #DC2626; }
.empty { color: #9CA3AF; font-size: 24rpx; }
</style>
