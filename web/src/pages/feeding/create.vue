<template>
  <SideLayout>
    <view class="page">
      <view class="section-card">
        <text class="section-title">客户与猫咪</text>
        <view class="search-row">
          <input v-model="customerKeyword" class="input" placeholder="输入客户昵称或手机号搜索" @confirm="searchCustomers" />
          <view class="mini-btn" @click="searchCustomers">搜索</view>
        </view>
        <view class="select-list" v-if="customerKeyword.trim() && customers.length">
          <view v-for="customer in customers" :key="customer.ID">
            <view
              :class="['select-row', form.customer_id === customer.ID ? 'active' : '']"
              @click="selectCustomer(customer)"
            >
              <text class="select-name">{{ customer.nickname || '未命名客户' }}</text>
              <text class="select-phone">{{ customer.phone || '' }}</text>
              <text class="select-mark">{{ form.customer_id === customer.ID ? '已选' : '选择' }}</text>
            </view>
            <view v-if="form.customer_id === customer.ID && customerPets.length" class="pet-inline">
              <view v-for="pet in customerPets" :key="pet.ID" :class="['pet-chip', selectedPetIds.includes(pet.ID) ? 'active' : '']" @click="togglePet(pet.ID)">
                {{ pet.name }}
              </view>
            </view>
          </view>
        </view>
      </view>

      <view class="section-card">
        <text class="section-title">地址与联系人</text>
        <view class="compact-fields">
          <view class="compact-row">
            <text class="compact-label">地址</text>
            <input v-model="form.address_snapshot.address" class="compact-input flex" placeholder="上门地址" />
          </view>
          <view class="compact-row">
            <text class="compact-label">入户</text>
            <input v-model="form.address_snapshot.door_code" class="compact-input flex" placeholder="门锁密码/门卡" />
          </view>
          <view class="compact-row">
            <text class="compact-label">补充</text>
            <input v-model="form.address_snapshot.detail" class="compact-input flex" placeholder="停车、楼栋等" />
          </view>
          <view class="compact-row">
            <text class="compact-label">联系人</text>
            <input v-model="form.contact_name" class="compact-input" placeholder="可乐妈妈" />
            <text class="compact-label" style="margin-left: 16rpx;">电话</text>
            <input v-model="form.contact_phone" class="compact-input" placeholder="13800000000" type="number" />
          </view>
        </view>
      </view>

      <view class="section-card">
        <text class="section-title">服务日期</text>
        <view class="field-grid">
          <view class="date-range-row">
            <picker mode="date" :value="form.start_date" @change="form.start_date = $event.detail.value" class="date-range-picker">
              <view class="picker-card">
                <text class="field-label">开始日期</text>
                <text class="picker-value">{{ form.start_date || '请选择' }}</text>
              </view>
            </picker>
            <text class="date-range-sep">~</text>
            <picker mode="date" :value="form.end_date" @change="form.end_date = $event.detail.value" class="date-range-picker">
              <view class="picker-card">
                <text class="field-label">结束日期</text>
                <text class="picker-value">{{ form.end_date || '请选择' }}</text>
              </view>
            </picker>
          </view>
        </view>
        <view class="date-picker-block" v-if="rangeDateOptions.length">
          <view class="date-picker-head">
            <text class="field-label no-gap">具体上门日期</text>
            <view class="date-mode-switch">
              <view :class="['date-mode-chip', dateMode === 'daily' ? 'active' : '']" @click="setDateMode('daily')">每天</view>
              <view :class="['date-mode-chip', dateMode === 'custom' ? 'active' : '']" @click="setDateMode('custom')">指定日期</view>
            </view>
          </view>
          <view class="date-picker-head secondary" v-if="dateMode === 'custom'">
            <text class="helper-text inline">点击需要上门的具体日期。</text>
            <view class="date-actions">
              <text class="date-action" @click="selectAllDates">全选</text>
              <text class="date-action" @click="clearSelectedDates">清空</text>
            </view>
          </view>
          <text v-if="dateMode === 'daily'" class="helper-text">已按起止时间设置为每天上门，共 {{ form.selected_dates.length }} 天。</text>
          <view v-else class="date-chip-list">
            <view
              v-for="item in rangeDateOptions"
              :key="item.date"
              :class="['date-chip', form.selected_dates.includes(item.date) ? 'active' : '']"
              @click="toggleSelectedDate(item.date)"
            >
              {{ item.label }}
            </view>
          </view>
        </view>
      </view>

      <view class="section-card">
        <text class="section-title">服务内容</text>
        <view class="item-list">
          <view :class="['item-chip sm', playEnabled ? 'active' : '']" @click="playEnabled = !playEnabled">陪玩 ¥20/次</view>
          <view
            v-for="item in customItems"
            :key="item.code"
            :class="['item-chip sm', form.item_codes.includes(item.code) ? 'active' : '']"
            @click="toggleItem(item.code)"
          >
            <text>{{ item.name }}</text>
            <text v-if="item.extra_price > 0" class="item-price">+¥{{ item.extra_price }}</text>
          </view>
          <view :class="['item-chip sm', otherEnabled ? 'active' : '']" @click="otherEnabled = !otherEnabled">其他</view>
        </view>
        <view v-if="playEnabled" class="service-detail">
          <view class="play-mode-switch">
            <view :class="['date-mode-chip', playMode === 'daily' ? 'active' : '']" @click="playMode = 'daily'">每天</view>
            <view :class="['date-mode-chip', playMode === 'count' ? 'active' : '']" @click="playMode = 'count'">选择次数</view>
          </view>
          <view v-if="playMode === 'count'" class="play-count-row">
            <text class="play-count-label">次数</text>
            <view class="stepper">
              <view class="stepper-btn" @click="playCount = Math.max(1, playCount - 1)">-</view>
              <input class="stepper-input" type="number" v-model="playCount" />
              <view class="stepper-btn" @click="playCount++">+</view>
            </view>
            <text class="play-count-total">= ¥{{ playCount * 20 }}</text>
          </view>
          <text v-else class="helper-text">每天 {{ form.selected_dates.length }} 次 = ¥{{ form.selected_dates.length * 20 }}</text>
        </view>
        <view v-if="otherEnabled" class="other-price-row" style="margin-top: 10rpx;">
          <text class="other-price-label">其他金额 ¥</text>
          <input class="other-price-input" type="digit" v-model="otherPrice" placeholder="0" />
        </view>
      </view>

      <view class="section-card">
        <text class="section-title">备注</text>
        <textarea v-model="form.remark" class="textarea short" placeholder="猫咪应激点、喂药要求、禁区和紧急说明" />
      </view>

      <view class="summary-card">
        <text class="summary-title">预估信息</text>
        <text class="summary-text">已选 {{ form.selected_dates.length }} 天 · 预估总价 ¥{{ estimatedAmount.toFixed(2) }}</text>
        <text class="summary-sub">日常 {{ pricingPreview.regularDays }} 天{{ pricingPreview.regularDays >= settings.pricing.discount_start_day ? '(优惠)' : '' }}，节假日 {{ pricingPreview.holidayDays }} 天{{ pricingPreview.holidayDays >= settings.pricing.discount_start_day ? '(优惠)' : '' }}</text>
        <text class="summary-sub">基础金额 ¥{{ pricingPreview.baseAmount.toFixed(2) }}，附加服务 ¥{{ pricingPreview.extraAmount.toFixed(2) }}</text>
      </view>

      <view class="footer-bar">
        <button class="submit-btn" :loading="submitting" @click="submit">{{ isEdit ? '保存修改' : '创建计划' }}</button>
      </view>
    </view>
  </SideLayout>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { onLoad } from '@dcloudio/uni-app'
import SideLayout from '@/components/SideLayout.vue'
import { getCustomerList, getCustomerPets, getCustomer } from '@/api/customer'
import { createFeedingPlan, getFeedingPlan, getFeedingSettings, updateFeedingPlan } from '@/api/feeding'
import { getBoardingHolidays } from '@/api/boarding'
import { getFeedingDateOptions, parseFeedingAddress, parseFeedingSelectedDates, parseFeedingSelectedItems } from '@/utils/feeding'

const id = ref(0)
const isEdit = computed(() => id.value > 0)
const submitting = ref(false)
const customerKeyword = ref('')
const customers = ref<Customer[]>([])
const customerPets = ref<Pet[]>([])
const selectedCustomer = ref<Customer | null>(null)
const selectedPetIds = ref<number[]>([])
const settings = ref<FeedingSettings>({ pricing: { base_day_price: 85, holiday_day_price: 95, discount_day_price: 68, discount_holiday_price: 90, discount_start_day: 3 }, items: [] })
const dateMode = ref<'daily' | 'custom'>('custom')
const playEnabled = ref(false)
const playMode = ref<'daily' | 'count'>('daily')
const playCount = ref(1)
const otherEnabled = ref(false)
const otherPrice = ref('')
const holidayMap = ref<Record<string, boolean>>({})
const form = ref({
  customer_id: 0,
  address_snapshot: { address: '', detail: '', door_code: '' } as FeedingAddressSnapshot,
  contact_name: '',
  contact_phone: '',
  start_date: '',
  end_date: '',
  selected_dates: [] as string[],
  remark: '',
  rules: [] as FeedingPlanRule[],
  item_codes: [] as string[],
})

const rangeDateOptions = computed(() => getFeedingDateOptions(form.value.start_date, form.value.end_date))
const selectedDatesSorted = computed(() => [...form.value.selected_dates].sort())

const pricingPreview = computed(() => {
  let regularDays = 0
  let holidayDays = 0
  let discountDays = 0
  let baseAmount = 0
  const discountStartDay = Math.max(Number(settings.value.pricing.discount_start_day || 0), 1)
  // 先统计普通天数和节假日天数
  selectedDatesSorted.value.forEach((date) => {
    if (holidayMap.value[date]) holidayDays += 1
    else regularDays += 1
  })
  const regularDiscount = regularDays >= discountStartDay
  const holidayDiscount = holidayDays >= discountStartDay
  selectedDatesSorted.value.forEach((date) => {
    const isHoliday = !!holidayMap.value[date]
    const hasDiscount = isHoliday ? holidayDiscount : regularDiscount
    if (hasDiscount) discountDays += 1
    baseAmount += isHoliday
      ? (hasDiscount ? settings.value.pricing.discount_holiday_price : settings.value.pricing.holiday_day_price)
      : (hasDiscount ? settings.value.pricing.discount_day_price : settings.value.pricing.base_day_price)
  })
  let extraAmount = 0
  if (playEnabled.value) {
    const playTimes = playMode.value === 'daily' ? selectedDatesSorted.value.length : Number(playCount.value) || 0
    extraAmount += 20 * playTimes
  }
  // 自定义服务项的额外费用（按天）
  const customExtra = customItems.value
    .filter(item => form.value.item_codes.includes(item.code))
    .reduce((sum, item) => sum + (item.extra_price || 0), 0)
  extraAmount += customExtra * selectedDatesSorted.value.length
  if (otherEnabled.value) {
    extraAmount += Number(otherPrice.value) || 0
  }
  return {
    regularDays,
    holidayDays,
    discountDays,
    baseAmount,
    extraAmount,
  }
})

const estimatedAmount = computed(() => {
  return pricingPreview.value.baseAmount + pricingPreview.value.extraAmount
})

const customItems = computed(() => settings.value.items.filter(item => item.code !== 'play' && item.code !== 'other'))

function toggleItem(code: string) {
  const index = form.value.item_codes.indexOf(code)
  if (index >= 0) {
    form.value.item_codes.splice(index, 1)
  } else {
    form.value.item_codes.push(code)
  }
}

function syncItemCodes() {
  const codes = form.value.item_codes.filter(c => c !== 'play' && c !== 'other')
  if (playEnabled.value) codes.push('play')
  if (otherEnabled.value) codes.push('other')
  form.value.item_codes = codes
}

async function searchCustomers() {
  const res = await getCustomerList({ page: 1, page_size: 20, keyword: customerKeyword.value.trim() || undefined })
  customers.value = res.data?.list || []
}

async function selectCustomer(customer: Customer) {
  selectedCustomer.value = customer
  form.value.customer_id = customer.ID
  if (!form.value.contact_name) {
    form.value.contact_name = customer.nickname || ''
  }
  if (!form.value.contact_phone) {
    form.value.contact_phone = customer.phone || ''
  }
  const [petsRes, customerRes] = await Promise.all([
    getCustomerPets(customer.ID),
    getCustomer(customer.ID),
  ])
  customerPets.value = petsRes.data || []
  const fullCustomer = customerRes.data || customer
  selectedCustomer.value = fullCustomer
  // 从客户记录回填地址（仅当地址为空时）
  if (!form.value.address_snapshot.address && fullCustomer.address) {
    form.value.address_snapshot.address = fullCustomer.address
    form.value.address_snapshot.detail = fullCustomer.address_detail || ''
    form.value.address_snapshot.door_code = fullCustomer.door_code || ''
  }
}

function togglePet(petId: number) {
  const index = selectedPetIds.value.indexOf(petId)
  if (index >= 0) {
    selectedPetIds.value.splice(index, 1)
  } else {
    selectedPetIds.value.push(petId)
  }
}

function toggleSelectedDate(date: string) {
  const index = form.value.selected_dates.indexOf(date)
  if (index >= 0) {
    form.value.selected_dates.splice(index, 1)
    return
  }
  form.value.selected_dates.push(date)
  form.value.selected_dates.sort()
}

function selectAllDates() {
  form.value.selected_dates = rangeDateOptions.value.map(item => item.date)
}

function clearSelectedDates() {
  form.value.selected_dates = []
}

function setDateMode(mode: 'daily' | 'custom') {
  dateMode.value = mode
  if (mode === 'daily') {
    selectAllDates()
  }
}

function updateSelectedDatesForMode() {
  const availableDates = rangeDateOptions.value.map(item => item.date)
  const available = new Set(availableDates)
  if (dateMode.value === 'daily') {
    form.value.selected_dates = availableDates
    return
  }
  form.value.selected_dates = form.value.selected_dates.filter(date => available.has(date))
}

function syncDateModeFromSelection() {
  const availableDates = rangeDateOptions.value.map(item => item.date)
  if (!availableDates.length || !form.value.selected_dates.length) {
    dateMode.value = 'custom'
    return
  }
  dateMode.value = form.value.selected_dates.length === availableDates.length ? 'daily' : 'custom'
}

async function loadSettings() {
  const res = await getFeedingSettings()
  settings.value = res.data || settings.value
}

async function loadHolidayData() {
  try {
    const res = await getBoardingHolidays()
    const list = res.data || []
    const nextMap: Record<string, boolean> = {}
    list.forEach((item: any) => {
      if (item.holiday_date) nextMap[item.holiday_date] = true
    })
    holidayMap.value = nextMap
  } catch {}
}

async function loadPlan(planId: number) {
  const res = await getFeedingPlan(planId)
  const plan = res.data
  if (!plan) return
  id.value = planId
  form.value.customer_id = plan.customer_id
  form.value.address_snapshot = parseFeedingAddress(plan.address_snapshot_json)
  form.value.contact_name = plan.contact_name || ''
  form.value.contact_phone = plan.contact_phone || ''
  form.value.start_date = plan.start_date || ''
  form.value.end_date = plan.end_date || ''
  const parsedSelectedDates = parseFeedingSelectedDates(plan.selected_dates_json)
  form.value.selected_dates = parsedSelectedDates.length
    ? parsedSelectedDates
    : Array.from(new Set((plan.visits || []).map(item => item.scheduled_date).filter(Boolean))).sort()
  form.value.remark = plan.remark || ''
  form.value.rules = []
  form.value.item_codes = parseFeedingSelectedItems(plan.selected_items_json).map(item => item.code)
  playEnabled.value = form.value.item_codes.includes('play')
  otherEnabled.value = form.value.item_codes.includes('other')
  if (plan.play_mode) playMode.value = plan.play_mode as 'daily' | 'count'
  if (plan.play_count) playCount.value = plan.play_count
  if (plan.other_price) otherPrice.value = String(plan.other_price)
  selectedPetIds.value = (plan.pets || []).map(item => item.pet_id)
  syncDateModeFromSelection()
  if (plan.customer) {
    await selectCustomer(plan.customer)
  }
}

async function submit() {
  if (!form.value.customer_id) {
    uni.showToast({ title: '请先选择客户', icon: 'none' })
    return
  }
  if (!selectedPetIds.value.length) {
    uni.showToast({ title: '请至少选择 1 只猫咪', icon: 'none' })
    return
  }
  if (!form.value.start_date || !form.value.end_date) {
    uni.showToast({ title: '请选择日期范围', icon: 'none' })
    return
  }
  if (!form.value.selected_dates.length) {
    uni.showToast({ title: '请至少选择 1 个上门日期', icon: 'none' })
    return
  }
  syncItemCodes()
  submitting.value = true
  try {
    const payload = {
      ...form.value,
      pets: selectedPetIds.value.map(petId => ({ pet_id: petId })),
      selected_dates: [...form.value.selected_dates],
      rules: [],
      play_mode: playEnabled.value ? playMode.value : '',
      play_count: playEnabled.value && playMode.value === 'count' ? Number(playCount.value) || 0 : 0,
      other_price: otherEnabled.value ? Number(otherPrice.value) || 0 : 0,
    }
    const res = isEdit.value
      ? await updateFeedingPlan(id.value, payload)
      : await createFeedingPlan(payload)
    const planId = res.data?.ID
    if (planId) {
      uni.redirectTo({ url: `/pages/feeding/detail?id=${planId}` })
    }
  } finally {
    submitting.value = false
  }
}

onLoad(async (options) => {
  await loadSettings()
  if (options?.id) {
    await loadPlan(Number(options.id))
  }
  await loadHolidayData()
})

watch(
  () => [form.value.start_date, form.value.end_date],
  async () => {
    updateSelectedDatesForMode()
    await loadHolidayData()
  }
)
</script>

<style scoped>
.page { padding: 24rpx 24rpx calc(260rpx + 50px + env(safe-area-inset-bottom)); }
.hero-card, .section-card, .summary-card { background: #fff; border-radius: 22rpx; box-shadow: 0 12rpx 28rpx rgba(15, 23, 42, 0.06); padding: 24rpx; margin-bottom: 18rpx; }
.hero-title { display: block; font-size: 36rpx; font-weight: 700; color: #111827; }
.hero-subtitle { display: block; margin-top: 10rpx; font-size: 24rpx; color: #6B7280; line-height: 1.6; }
.section-title, .summary-title { display: block; font-size: 28rpx; font-weight: 700; color: #111827; margin-bottom: 18rpx; }
.search-row { display: flex; gap: 12rpx; }
.input, .textarea, .picker-card { width: 100%; min-height: 88rpx; padding: 22rpx 24rpx; background: #F8FAFC; border-radius: 18rpx; font-size: 26rpx; color: #111827; box-sizing: border-box; }
.textarea { min-height: 180rpx; }
.textarea.short { min-height: 112rpx; }
.picker-card { display: flex; flex-direction: column; justify-content: center; }
.date-range-row { display: flex; align-items: center; gap: 16rpx; }
.date-range-picker { flex: 1; }
.date-range-sep { font-size: 28rpx; color: #9CA3AF; flex-shrink: 0; }
.field-grid { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 16rpx; }
.compact-fields { display: flex; flex-direction: column; gap: 0; }
.compact-row { display: flex; align-items: center; gap: 10rpx; padding: 12rpx 0; border-bottom: 1rpx solid #F3F4F6; }
.compact-row:last-child { border-bottom: none; }
.compact-label { font-size: 24rpx; color: #111827; font-weight: 600; flex-shrink: 0; min-width: 64rpx; }
.compact-input { height: 56rpx; padding: 0 14rpx; background: #F8FAFC; border-radius: 10rpx; font-size: 24rpx; color: #111827; min-width: 0; }
.compact-input.flex { flex: 1; }
.field-card.full { grid-column: 1 / -1; }
.field-label { display: block; margin-bottom: 8rpx; font-size: 22rpx; color: #6B7280; }
.picker-value { font-size: 28rpx; color: #111827; }
.helper-text { display: block; margin-top: 10rpx; font-size: 22rpx; color: #6B7280; line-height: 1.6; }
.mini-btn { flex: 0 0 auto; min-width: 120rpx; padding: 0 24rpx; height: 88rpx; line-height: 88rpx; text-align: center; border-radius: 18rpx; background: #4F46E5; color: #fff; font-size: 24rpx; }
.select-list { display: flex; flex-direction: column; gap: 8rpx; margin-top: 12rpx; }
.select-row { display: flex; align-items: center; gap: 12rpx; padding: 14rpx 16rpx; border-radius: 12rpx; background: #F8FAFC; }
.select-row.active { background: #EEF2FF; border: 1rpx solid #C7D2FE; }
.select-name { font-size: 26rpx; font-weight: 600; color: #111827; flex-shrink: 0; }
.select-phone { font-size: 22rpx; color: #6B7280; flex: 1; }
.select-mark { font-size: 22rpx; color: #4F46E5; flex-shrink: 0; }
.sub-block { margin-top: 18rpx; }
.sub-title { display: block; font-size: 24rpx; color: #374151; margin-bottom: 12rpx; }
.pet-list, .item-list { display: flex; flex-wrap: wrap; gap: 10rpx; }
.pet-inline { display: flex; flex-wrap: wrap; gap: 10rpx; padding: 10rpx 16rpx 14rpx; background: #F5F3FF; border-radius: 0 0 12rpx 12rpx; }
.pet-chip, .item-chip { padding: 14rpx 22rpx; border-radius: 999rpx; background: #F3F4F6; color: #374151; font-size: 24rpx; }
.item-chip.sm { padding: 10rpx 20rpx; font-size: 24rpx; }
.pet-chip.active, .item-chip.active { background: #4F46E5; color: #fff; }
.item-price { margin-left: 8rpx; font-size: 20rpx; opacity: 0.9; }
.service-detail { margin-top: 12rpx; padding-left: 8rpx; }
.play-mode-switch { display: flex; gap: 12rpx; }
.play-count-row { display: flex; align-items: center; gap: 16rpx; margin-top: 16rpx; }
.play-count-label { font-size: 24rpx; color: #374151; }
.play-count-total { font-size: 24rpx; color: #4F46E5; font-weight: 600; }
.stepper { display: flex; align-items: center; border: 1rpx solid #E5E7EB; border-radius: 12rpx; overflow: hidden; }
.stepper-btn { width: 60rpx; height: 56rpx; display: flex; align-items: center; justify-content: center; font-size: 32rpx; color: #374151; background: #F9FAFB; }
.stepper-input { width: 80rpx; height: 56rpx; text-align: center; font-size: 26rpx; border: none; background: #fff; }
.other-price-row { display: flex; align-items: center; gap: 4rpx; }
.other-price-label { font-size: 26rpx; color: #374151; }
.other-price-input { width: 120rpx; height: 56rpx; padding: 0 12rpx; background: #F8FAFC; border-radius: 12rpx; font-size: 26rpx; text-align: center; }
.date-picker-block { margin-top: 18rpx; }
.date-picker-head { display: flex; align-items: center; justify-content: space-between; gap: 12rpx; }
.date-picker-head.secondary { margin-top: 14rpx; }
.date-actions { display: flex; gap: 16rpx; }
.date-action { font-size: 22rpx; color: #4F46E5; }
.no-gap { margin-bottom: 0; }
.inline { margin-top: 0; }
.date-mode-switch { display: flex; gap: 12rpx; }
.date-mode-chip { padding: 10rpx 22rpx; border-radius: 999rpx; background: #F3F4F6; color: #4B5563; font-size: 22rpx; border: 1rpx solid #E5E7EB; }
.date-mode-chip.active { background: #EEF2FF; color: #4338CA; border-color: #C7D2FE; }
.date-chip-list { display: flex; flex-wrap: wrap; gap: 12rpx; margin-top: 14rpx; }
.date-chip { padding: 14rpx 18rpx; border-radius: 16rpx; background: #F3F4F6; color: #374151; font-size: 22rpx; border: 1rpx solid #E5E7EB; }
.date-chip.active { background: #EEF2FF; color: #4338CA; border-color: #C7D2FE; }
.summary-text { display: block; font-size: 24rpx; color: #6B7280; line-height: 1.7; }
.summary-sub { display: block; margin-top: 8rpx; font-size: 22rpx; color: #6B7280; line-height: 1.7; }
.footer-bar { position: fixed; left: 24rpx; right: 24rpx; bottom: calc(32rpx + 50px + env(safe-area-inset-bottom)); z-index: 20; }
.submit-btn { width: 100%; height: 92rpx; border-radius: 999rpx; background: linear-gradient(135deg, #4F46E5, #6366F1); color: #fff; font-size: 30rpx; display: flex; align-items: center; justify-content: center; }
@media (max-width: 768px) {
  .field-grid { grid-template-columns: 1fr; }
}
</style>
