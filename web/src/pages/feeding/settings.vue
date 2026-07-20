<template>
  <SideLayout>
    <view class="page">
      <view class="section-card">
        <text class="section-title">价格规则</text>
        <view class="field-grid">
          <view class="field-card">
            <text class="field-label">日常价格 / 天</text>
            <input v-model="pricing.base_day_price" class="input" type="digit" />
          </view>
          <view class="field-card">
            <text class="field-label">法定节假日价格 / 天</text>
            <input v-model="pricing.holiday_day_price" class="input" type="digit" />
          </view>
          <view class="field-card full">
            <text class="field-label">第 N 天起（日常） / 天</text>
            <input v-model="pricing.discount_day_price" class="input" type="digit" />
          </view>
          <view class="field-card full">
            <text class="field-label">第 N 天起（节假日） / 天</text>
            <input v-model="pricing.discount_holiday_price" class="input" type="digit" />
          </view>
          <view class="field-card full">
            <text class="field-label">优惠开始天数</text>
            <input v-model="pricing.discount_start_day" class="input" type="number" />
          </view>
        </view>
        <text class="helper-text">附加服务仍通过下方“服务内容模板”的加收金额配置，例如“超长侍玩喂玩服务 +20/天”。</text>
        <view class="submit-row">
          <view class="btn btn-primary" @click="savePricing">保存价格规则</view>
        </view>
      </view>

      <view class="section-card">
        <view class="section-head">
          <text class="section-title">节假日日期</text>
          <text v-if="editingHolidayRange" class="link-btn" @click="resetHolidayForm">取消修改</text>
        </view>
        <text class="helper-text no-top">上门喂养和寄养共用这组节假日日期；上门计价会按这里拆分日常价和节假日价。</text>
        <view class="holiday-form">
          <view class="date-grid">
            <picker mode="date" :value="holidayForm.start_date" @change="onHolidayStartChange($event.detail.value)">
              <view class="picker">{{ holidayForm.start_date || '开始日期' }}</view>
            </picker>
            <picker mode="date" :value="holidayForm.end_date" @change="holidayForm.end_date = $event.detail.value">
              <view class="picker">{{ holidayForm.end_date || '结束日期' }}</view>
            </picker>
          </view>
          <input v-model="holidayForm.name" class="input" placeholder="节假日名称，例如 端午" />
          <input v-model="holidayForm.surcharge_amount" class="input" type="digit" placeholder="寄养每晚加收，可填 0" />
          <view class="btn btn-primary full-btn" @click="saveHolidayRange">{{ editingHolidayRange ? '保存节假日' : '添加节假日' }}</view>
        </view>
        <view v-if="holidayRanges.length === 0" class="empty">暂无节假日配置</view>
        <view v-else class="holiday-list">
          <view class="holiday-row" v-for="range in holidayRanges" :key="range.key">
            <view class="holiday-main">
              <text class="holiday-date">{{ range.dateLabel }}</text>
              <text class="holiday-name">{{ range.name || '节假日' }} · {{ range.count }}天</text>
            </view>
            <view class="holiday-actions">
              <text class="holiday-edit" @click="editHolidayRange(range)">编辑</text>
              <text class="holiday-del" @click="removeHolidayRange(range)">删除</text>
            </view>
          </view>
        </view>
      </view>

      <view class="section-card">
        <view class="section-head">
          <text class="section-title">服务内容模板</text>
          <view class="btn" @click="addItem">新增模板</view>
        </view>
        <view class="item-table">
          <view class="item-tr item-header">
            <view class="item-td name">名称</view>
            <view class="item-td price">加收</view>
            <view class="item-td action"></view>
          </view>
          <view class="item-tr" v-for="(item, index) in items" :key="`${item.code}-${index}`">
            <view class="item-td name"><input v-model="item.name" class="cell-input" placeholder="名称" /></view>
            <view class="item-td price"><input v-model="item.extra_price" class="cell-input" type="digit" placeholder="0" /></view>
            <view class="item-td action"><text class="del-btn" @click="items.splice(index, 1)">删除</text></view>
          </view>
        </view>
        <view class="submit-row">
          <view class="btn btn-primary" @click="saveItems">保存模板</view>
        </view>
      </view>
    </view>
  </SideLayout>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { onShow } from '@dcloudio/uni-app'
import SideLayout from '@/components/SideLayout.vue'
import { getFeedingSettings, updateFeedingItems, updateFeedingPricing } from '@/api/feeding'
import { createBoardingHoliday, deleteBoardingHoliday, getBoardingHolidays, updateBoardingHolidayRange } from '@/api/boarding'

const pricing = ref<FeedingPricingSetting>({
  base_day_price: 85,
  holiday_day_price: 95,
  discount_day_price: 68,
  discount_holiday_price: 90,
  discount_start_day: 3,
})
const items = ref<FeedingItemTemplate[]>([])
const holidays = ref<BoardingHoliday[]>([])
const holidayForm = ref({ start_date: '', end_date: '', name: '', surcharge_amount: '' })
const editingHolidayRange = ref<HolidayRange | null>(null)

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
    if (last && last.name === name && last.surcharge_amount === surchargeAmount && isNextDate(last.end_date, date)) {
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

function addItem() {
  items.value.push({ code: `item_${Date.now()}`, name: '', extra_price: 0 })
}

async function loadData() {
  const [settingsRes, holidaysRes] = await Promise.all([
    getFeedingSettings(),
    getBoardingHolidays(),
  ])
  pricing.value = { ...settingsRes.data.pricing }
  items.value = (settingsRes.data.items || []).map(item => ({ ...item }))
  holidays.value = holidaysRes.data || []
}

async function savePricing() {
  await updateFeedingPricing({
    base_day_price: Number(pricing.value.base_day_price || 0),
    holiday_day_price: Number(pricing.value.holiday_day_price || 0),
    discount_day_price: Number(pricing.value.discount_day_price || 0),
    discount_holiday_price: Number(pricing.value.discount_holiday_price || 0),
    discount_start_day: Number(pricing.value.discount_start_day || 0),
  })
  uni.showToast({ title: '已保存', icon: 'success' })
}

async function saveItems() {
  await updateFeedingItems(items.value.map(item => ({
    code: item.code.trim(),
    name: item.name.trim(),
    extra_price: Number(item.extra_price || 0),
  })))
  uni.showToast({ title: '已保存', icon: 'success' })
  await loadData()
}

function onHolidayStartChange(value: string) {
  holidayForm.value.start_date = value
  if (!holidayForm.value.end_date || holidayForm.value.end_date < value) {
    holidayForm.value.end_date = value
  }
}

async function saveHolidayRange() {
  if (!holidayForm.value.start_date || !holidayForm.value.end_date) {
    uni.showToast({ title: '请选择节假日日期', icon: 'none' })
    return
  }
  if (holidayForm.value.end_date < holidayForm.value.start_date) {
    uni.showToast({ title: '结束日期不能早于开始日期', icon: 'none' })
    return
  }
  const surchargeAmount = Number(holidayForm.value.surcharge_amount || 0)
  if (!Number.isFinite(surchargeAmount) || surchargeAmount < 0) {
    uni.showToast({ title: '请填写有效加收金额', icon: 'none' })
    return
  }
  const payload = {
    start_date: holidayForm.value.start_date,
    end_date: holidayForm.value.end_date,
    name: holidayForm.value.name.trim(),
    surcharge_amount: surchargeAmount,
  }
  const res = editingHolidayRange.value
    ? await updateBoardingHolidayRange({ ...payload, ids: editingHolidayRange.value.ids })
    : await createBoardingHoliday(payload)
  const changed = res.data || []
  uni.showToast({ title: changed.length > 1 ? `已保存${changed.length}天` : '已保存', icon: 'success' })
  resetHolidayForm()
  await loadData()
}

function editHolidayRange(range: HolidayRange) {
  editingHolidayRange.value = range
  holidayForm.value = {
    start_date: range.start_date,
    end_date: range.end_date,
    name: range.name,
    surcharge_amount: range.surcharge_amount > 0 ? String(range.surcharge_amount) : '',
  }
}

async function removeHolidayRange(range: HolidayRange) {
  await Promise.all(range.ids.map(id => deleteBoardingHoliday(id)))
  if (editingHolidayRange.value?.key === range.key) {
    resetHolidayForm()
  }
  uni.showToast({ title: range.ids.length > 1 ? '已删除该范围' : '已删除', icon: 'success' })
  await loadData()
}

function resetHolidayForm() {
  editingHolidayRange.value = null
  holidayForm.value = { start_date: '', end_date: '', name: '', surcharge_amount: '' }
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
.page { padding: 24rpx; }
.section-card { background: #fff; border-radius: 22rpx; box-shadow: 0 12rpx 28rpx rgba(15, 23, 42, 0.06); padding: 24rpx; margin-bottom: 18rpx; }
.section-head { display: flex; justify-content: space-between; gap: 12rpx; align-items: center; margin-bottom: 18rpx; }
.section-title { font-size: 28rpx; font-weight: 700; color: #111827; }
.field-grid { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 12rpx; }
.field-card { display: flex; align-items: center; gap: 12rpx; }
.field-card.full { grid-column: 1 / -1; }
.field-label { font-size: 22rpx; color: #6B7280; white-space: nowrap; flex-shrink: 0; }
.helper-text { display: block; margin-top: 14rpx; font-size: 22rpx; color: #6B7280; line-height: 1.7; }
.helper-text.no-top { margin-top: 0; margin-bottom: 14rpx; }
.input { width: 100%; min-height: 68rpx; padding: 14rpx 18rpx; background: #F8FAFC; border-radius: 14rpx; font-size: 26rpx; color: #111827; box-sizing: border-box; flex: 1; min-width: 0; }
.submit-row { display: flex; justify-content: flex-end; margin-top: 18rpx; }
.btn { padding: 14rpx 22rpx; border-radius: 16rpx; background: #F8FAFC; color: #374151; font-size: 24rpx; border: 1rpx solid #E5E7EB; }
.btn-primary { background: linear-gradient(135deg, #4F46E5, #6366F1); color: #fff; border-color: transparent; }
.btn.danger { color: #DC2626; background: #FEF2F2; border-color: #FECACA; text-align: center; }
.link-btn { color: #4F46E5; font-size: 24rpx; white-space: nowrap; }
.date-grid { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 12rpx; margin-bottom: 12rpx; }
.picker { min-height: 68rpx; padding: 0 18rpx; background: #F8FAFC; border-radius: 14rpx; font-size: 26rpx; color: #111827; display: flex; align-items: center; box-sizing: border-box; }
.holiday-form { display: flex; flex-direction: column; gap: 12rpx; margin-bottom: 18rpx; }
.full-btn { text-align: center; }
.holiday-list { display: flex; flex-direction: column; border-top: 1rpx solid #F3F4F6; }
.holiday-row { display: flex; align-items: center; justify-content: space-between; gap: 16rpx; padding: 16rpx 0; border-bottom: 1rpx solid #F3F4F6; }
.holiday-main { min-width: 0; }
.holiday-date { display: block; font-size: 26rpx; font-weight: 700; color: #111827; }
.holiday-name { display: block; margin-top: 6rpx; font-size: 22rpx; color: #6B7280; }
.holiday-actions { display: flex; align-items: center; gap: 18rpx; flex-shrink: 0; }
.holiday-edit { color: #4F46E5; font-size: 24rpx; }
.holiday-del { color: #DC2626; font-size: 24rpx; }
.empty { color: #9CA3AF; font-size: 24rpx; }
.item-table { border: 1rpx solid #E5E7EB; border-radius: 14rpx; overflow: hidden; }
.item-tr { display: flex; align-items: center; border-bottom: 1rpx solid #F3F4F6; }
.item-tr:last-child { border-bottom: none; }
.item-header { background: #F9FAFB; }
.item-header .item-td { font-size: 22rpx; color: #6B7280; font-weight: 600; padding: 14rpx 12rpx; }
.item-td { padding: 8rpx 10rpx; font-size: 24rpx; color: #374151; }
.item-td.code { flex: 2; }
.item-td.name { flex: 2; }
.item-td.price { flex: 1.5; }
.item-td.action { flex: 1; text-align: center; }
.cell-input { width: 100%; height: 60rpx; padding: 0 12rpx; background: #F8FAFC; border-radius: 10rpx; font-size: 24rpx; color: #111827; box-sizing: border-box; }
.del-btn { font-size: 22rpx; color: #DC2626; }
</style>
