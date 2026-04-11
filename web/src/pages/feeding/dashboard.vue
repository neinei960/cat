<template>
  <SideLayout>
    <view class="page">
      <view class="top-bar">
        <text class="top-title">上门喂养</text>
        <view class="top-actions">
          <view class="btn btn-primary" @click="go('/pages/feeding/create')">新建</view>
          <view class="btn" @click="go('/pages/feeding/settings')">设置</view>
        </view>
      </view>

      <!-- 视图切换 -->
      <view class="view-tabs">
        <view :class="['view-tab', viewMode === 'calendar' ? 'active' : '']" @click="viewMode = 'calendar'">日历</view>
        <view :class="['view-tab', viewMode === 'cards' ? 'active' : '']" @click="viewMode = 'cards'">计划</view>
        <view :class="['view-tab', viewMode === 'history' ? 'active' : '']" @click="viewMode = 'history'">历史</view>
      </view>

      <view v-if="loading" class="state-card">加载中...</view>

      <!-- ========== 日历视图 ========== -->
      <view v-else-if="viewMode === 'calendar'" class="calendar-view">
        <view class="cal-header">
          <text class="cal-arrow" @click="calPrev">‹</text>
          <text class="cal-title">{{ calYear }}年{{ calMonth }}月</text>
          <text class="cal-arrow" @click="calNext">›</text>
          <text class="cal-today-btn" @click="calGoToday">今天</text>
        </view>
        <view class="cal-weekdays">
          <text class="cal-wd" v-for="w in ['一','二','三','四','五','六','日']" :key="w">{{ w }}</text>
        </view>
        <view class="cal-grid">
          <view
            v-for="(cell, idx) in calCells"
            :key="idx"
            :class="['cal-cell', {
              'out-month': !cell.inMonth,
              'is-today': cell.date === today,
              'is-selected': cell.date === selectedDate,
              'is-holiday-cell': cell.inMonth && holidaySet.has(cell.date),
            }]"
            @click="cell.inMonth && selectDate(cell.date)"
          >
            <text class="cal-day">{{ cell.day }}</text>
            <view v-if="cell.inMonth && feedingCountMap[cell.date]" class="cal-dot-row">
              <view :class="['cal-dot', holidaySet.has(cell.date) ? 'red' : 'green']" />
              <text class="cal-count">{{ feedingCountMap[cell.date] }}</text>
            </view>
          </view>
        </view>

        <!-- 选中日期详情 -->
        <view v-if="selectedDate" class="day-detail">
          <text class="day-title">{{ formatSelectedDate }} · {{ selectedDayPlans.length }}家上门</text>
          <view v-if="!selectedDayPlans.length" class="empty-hint">当天无上门安排</view>
          <view v-else class="plan-cards">
            <view class="plan-card" v-for="row in selectedDayPlans" :key="row.planId" @click="openPlan(row.planId)">
              <view class="pc-head">
                <text class="pc-name">{{ row.name }}</text>
                <text class="pc-meta">{{ row.catCount }}猫 · {{ row.totalDays }}天</text>
              </view>
              <text class="pc-addr">{{ row.address || '未填地址' }}</text>
              <view class="pc-tags" v-if="row.playDateSet.has(selectedDate) || row.extraTags.length">
                <text v-if="row.playDateSet.has(selectedDate)" class="pc-tag play">{{ row.playDayTag }}</text>
                <text v-for="tag in row.extraTags" :key="tag" class="pc-tag">{{ tag }}</text>
              </view>
              <view class="pc-money">
                <text>预计 ¥{{ (row.deposit + row.balance).toFixed(0) }}</text>
                <text v-if="row.deposit"> · 定金 ¥{{ row.deposit }}</text>
                <text> · 尾款 ¥{{ row.balance.toFixed(0) }}</text>
              </view>
            </view>
          </view>
        </view>
      </view>

      <!-- ========== 卡片视图 ========== -->
      <view v-else-if="viewMode === 'cards'" class="cards-view">
        <view v-if="!activeRows.length" class="state-card">暂无进行中的喂养计划</view>
        <view v-else class="plan-cards">
          <view class="plan-card" v-for="row in activeRows" :key="row.planId" @click="openPlan(row.planId)">
            <view class="pc-head">
              <text class="pc-name">{{ row.name }}</text>
              <text class="pc-meta">{{ row.catCount }}猫 · {{ row.totalDays }}天</text>
            </view>
            <!-- 迷你热力条 -->
            <view class="heatmap">
              <view
                v-for="d in row.allPlanDates"
                :key="d"
                :class="['heat-cell', {
                  'h-normal': row.dateSet.has(d) && !holidaySet.has(d),
                  'h-holiday': row.dateSet.has(d) && holidaySet.has(d),
                  'h-empty': !row.dateSet.has(d),
                }]"
              />
            </view>
            <text class="pc-addr">{{ row.address || '未填地址' }}</text>
            <view class="pc-money">
              <text>预计 ¥{{ (row.deposit + row.balance).toFixed(0) }}</text>
              <text v-if="row.deposit"> · 定金 ¥{{ row.deposit }}</text>
              <text> · 尾款 ¥{{ row.balance.toFixed(0) }}</text>
            </view>
            <view class="pc-tags" v-if="row.playSummaryTag || row.extraTags.length || row.remark">
              <text v-if="row.playSummaryTag" class="pc-tag play">{{ row.playSummaryTag }}</text>
              <text v-for="tag in row.extraTags" :key="tag" class="pc-tag">{{ tag }}</text>
              <text v-if="row.remark" class="pc-tag muted">{{ row.remark }}</text>
            </view>
          </view>
        </view>
      </view>

      <view v-else-if="viewMode === 'history'" class="cards-view">
        <view v-if="!historyRows.length" class="state-card">暂无历史喂养计划</view>
        <view v-else class="plan-cards">
          <view class="plan-card history-card" v-for="row in historyRows" :key="row.planId" @click="openPlan(row.planId)">
            <view class="pc-head">
              <text class="pc-name">{{ row.name }}</text>
              <text class="pc-meta">{{ row.catCount }}猫 · {{ row.totalDays }}天</text>
            </view>
            <text class="pc-addr">{{ row.address || '未填地址' }}</text>
            <view class="pc-tags">
              <text :class="['pc-tag', 'status-tag', `status-${row.status}`]">{{ feedingStatusLabel(row.status) }}</text>
              <text v-if="row.playSummaryTag" class="pc-tag play">{{ row.playSummaryTag }}</text>
              <text v-for="tag in row.extraTags" :key="tag" class="pc-tag">{{ tag }}</text>
            </view>
            <view class="pc-money">
              <text>{{ row.dateRange }}</text>
              <text> · ¥{{ (row.deposit + row.balance).toFixed(0) }}</text>
            </view>
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
import { getFeedingPlans } from '@/api/feeding'
import { getBoardingHolidays } from '@/api/boarding'
import { feedingStatusLabel, formatFeedingDateRange, parseFeedingAddress, parseFeedingSelectedDates } from '@/utils/feeding'

const loading = ref(false)
const plans = ref<FeedingPlan[]>([])
const today = formatLocalDate(new Date())
const holidaySet = ref<Set<string>>(new Set())
const viewMode = ref<'calendar' | 'cards' | 'history'>('calendar')

// ========== 日历逻辑 ==========
const calYear = ref(new Date().getFullYear())
const calMonth = ref(new Date().getMonth() + 1)
const selectedDate = ref(today)

function calPrev() {
  if (calMonth.value === 1) { calYear.value--; calMonth.value = 12 }
  else calMonth.value--
}
function calNext() {
  if (calMonth.value === 12) { calYear.value++; calMonth.value = 1 }
  else calMonth.value++
}
function calGoToday() {
  const now = new Date()
  calYear.value = now.getFullYear()
  calMonth.value = now.getMonth() + 1
  selectedDate.value = today
}
function selectDate(d: string) {
  selectedDate.value = d
}

interface CalCell { day: number; date: string; inMonth: boolean }
const calCells = computed<CalCell[]>(() => {
  const y = calYear.value, m = calMonth.value
  const firstDay = new Date(y, m - 1, 1)
  let startWeekday = firstDay.getDay()
  if (startWeekday === 0) startWeekday = 7 // Mon=1
  const daysInMonth = new Date(y, m, 0).getDate()
  const cells: CalCell[] = []
  // prev month fill
  const prevDays = new Date(y, m - 1, 0).getDate()
  for (let i = startWeekday - 1; i > 0; i--) {
    const d = prevDays - i + 1
    const pm = m === 1 ? 12 : m - 1
    const py = m === 1 ? y - 1 : y
    cells.push({ day: d, date: `${py}-${String(pm).padStart(2,'0')}-${String(d).padStart(2,'0')}`, inMonth: false })
  }
  // current month
  for (let d = 1; d <= daysInMonth; d++) {
    cells.push({ day: d, date: `${y}-${String(m).padStart(2,'0')}-${String(d).padStart(2,'0')}`, inMonth: true })
  }
  // next month fill
  while (cells.length < 42) {
    const d = cells.length - (startWeekday - 1) - daysInMonth + 1
    const nm = m === 12 ? 1 : m + 1
    const ny = m === 12 ? y + 1 : y
    cells.push({ day: d, date: `${ny}-${String(nm).padStart(2,'0')}-${String(d).padStart(2,'0')}`, inMonth: false })
  }
  return cells
})

// 每天有多少家上门
const feedingCountMap = computed(() => {
  const map: Record<string, number> = {}
  for (const row of activeRows.value) {
    row.dateSet.forEach(d => { map[d] = (map[d] || 0) + 1 })
  }
  return map
})

const formatSelectedDate = computed(() => {
  if (!selectedDate.value) return ''
  const parts = selectedDate.value.split('-')
  const weekdays = ['日', '一', '二', '三', '四', '五', '六']
  const wd = weekdays[new Date(selectedDate.value).getDay()]
  return `${parseInt(parts[1])}月${parseInt(parts[2])}日 周${wd}`
})

const selectedDayPlans = computed(() => {
  return activeRows.value.filter(r => r.dateSet.has(selectedDate.value))
})

// ========== 数据 ==========
async function loadHolidays() {
  try {
    const res = await getBoardingHolidays()
    const list = res.data || []
    holidaySet.value = new Set(list.map((item: any) => item.holiday_date).filter(Boolean))
  } catch {}
}

interface TableRow {
  planId: number
  name: string
  catCount: number
  totalDays: number
  dateSet: Set<string>
  allPlanDates: string[]
  deposit: number
  balance: number
  playDateSet: Set<string>
  playDayTag: string
  playSummaryTag: string
  extraTags: string[]
  remark: string
  address: string
  contactPhone: string
  status: string
  dateRange: string
}


function generateDateRange(start: string, end: string) {
  const dates: string[] = []
  const cur = new Date(start + 'T00:00:00')
  const endD = new Date(end + 'T00:00:00')
  while (cur <= endD) {
    dates.push(formatLocalDate(cur))
    cur.setDate(cur.getDate() + 1)
  }
  return dates
}

const rows = computed<TableRow[]>(() => {
  return plans.value.map(plan => {
    const dates = parseFeedingSelectedDates(plan.selected_dates_json)
    const dateSet = new Set(dates)
    const pets = plan.pets || []
    const petNames = pets.map(p => p.pet?.name || p.pet_name_snapshot).filter(Boolean).join('/')
    const customerName = plan.customer?.nickname || plan.customer?.phone || ''
    const addr = parseFeedingAddress(plan.address_snapshot_json)
    const playDateSet = new Set(parseFeedingSelectedDates(plan.play_dates_json))
    const playDayTag = '陪玩'
    let playSummaryTag = ''
    const extraTags: string[] = []
    if (playDateSet.size > 0) playSummaryTag = `陪玩×${playDateSet.size}`
    else if (plan.play_mode === 'daily') playSummaryTag = '陪玩'
    else if (plan.play_mode === 'count' && plan.play_count > 0) playSummaryTag = `陪玩×${plan.play_count}`
    if (plan.other_price > 0) extraTags.push(`其他+${plan.other_price}`)
    return {
      planId: plan.ID,
      name: petNames || customerName,
      catCount: pets.length,
      totalDays: dates.length,
      dateSet,
      allPlanDates: generateDateRange(plan.start_date, plan.end_date),
      deposit: plan.deposit || 0,
      balance: plan.total_amount - (plan.deposit || 0),
      playDateSet,
      playDayTag,
      playSummaryTag,
      extraTags,
      remark: plan.remark || '',
      address: addr.address || '',
      contactPhone: plan.contact_phone || '',
      status: plan.status || 'draft',
      dateRange: formatFeedingDateRange(plan.start_date, plan.end_date),
    }
  })
})

const activeRows = computed(() => rows.value.filter(row => row.status === 'active'))
const historyRows = computed(() => rows.value.filter(row => row.status !== 'active'))


function formatLocalDate(value: Date) {
  const y = value.getFullYear()
  const m = String(value.getMonth() + 1).padStart(2, '0')
  const d = String(value.getDate()).padStart(2, '0')
  return `${y}-${m}-${d}`
}

function go(url: string) { uni.navigateTo({ url }) }
function openPlan(planId: number) { uni.navigateTo({ url: `/pages/feeding/detail?id=${planId}` }) }

async function loadData() {
  loading.value = true
  try {
    const res = await getFeedingPlans({ page: 1, page_size: 200 })
    plans.value = res.data?.list || []
  } finally {
    loading.value = false
  }
}

onShow(() => {
  loadHolidays()
  loadData()
})
</script>

<style scoped>
.page { padding: 24rpx 24rpx calc(160rpx + env(safe-area-inset-bottom)); }
.top-bar { display: flex; justify-content: space-between; align-items: center; margin-bottom: 12rpx; }
.top-title { font-size: 32rpx; font-weight: 700; color: #111827; }
.top-actions { display: flex; gap: 10rpx; }
.btn { padding: 14rpx 22rpx; border-radius: 14rpx; background: #F8FAFC; color: #374151; font-size: 22rpx; border: 1rpx solid #E5E7EB; }
.btn-primary { background: linear-gradient(135deg, #4F46E5, #6366F1); color: #fff; border-color: transparent; }
.state-card { background: #fff; border-radius: 18rpx; padding: 100rpx 24rpx; text-align: center; color: #9CA3AF; font-size: 26rpx; }

/* 视图切换 */
.view-tabs { display: flex; gap: 8rpx; margin-bottom: 16rpx; }
.view-tab { flex: 1; text-align: center; padding: 14rpx 0; border-radius: 12rpx; font-size: 24rpx; color: #6B7280; background: #F3F4F6; }
.view-tab.active { background: #4F46E5; color: #fff; font-weight: 600; }

/* ========== 日历视图 ========== */
.calendar-view { background: #fff; border-radius: 18rpx; padding: 20rpx; box-shadow: 0 8rpx 24rpx rgba(15,23,42,0.06); }
.cal-header { display: flex; align-items: center; justify-content: center; gap: 16rpx; margin-bottom: 16rpx; }
.cal-title { font-size: 28rpx; font-weight: 700; color: #111827; }
.cal-arrow { font-size: 36rpx; font-weight: 700; color: #4F46E5; padding: 0 12rpx; -webkit-user-select: none; user-select: none; }
.cal-today-btn { font-size: 22rpx; color: #4F46E5; padding: 6rpx 16rpx; border: 1rpx solid #C7D2FE; border-radius: 8rpx; margin-left: 8rpx; }
.cal-weekdays { display: grid; grid-template-columns: repeat(7, 1fr); margin-bottom: 8rpx; }
.cal-wd { text-align: center; font-size: 20rpx; color: #9CA3AF; font-weight: 600; }
.cal-grid { display: grid; grid-template-columns: repeat(7, 1fr); gap: 4rpx; }
.cal-cell { display: flex; flex-direction: column; align-items: center; padding: 8rpx 0; border-radius: 10rpx; min-height: 72rpx; }
.cal-cell.out-month { opacity: 0.3; }
.cal-cell.is-today .cal-day { color: #4F46E5; font-weight: 700; }
.cal-cell.is-selected { background: #EEF2FF; }
.cal-cell.is-holiday-cell .cal-day { color: #DC2626; }
.cal-day { font-size: 24rpx; color: #374151; }
.cal-dot-row { display: flex; align-items: center; gap: 4rpx; margin-top: 4rpx; }
.cal-dot { width: 10rpx; height: 10rpx; border-radius: 50%; }
.cal-dot.green { background: #22C55E; }
.cal-dot.red { background: #EF4444; }
.cal-count { font-size: 16rpx; color: #6B7280; }

/* 日详情 */
.day-detail { margin-top: 20rpx; border-top: 1rpx solid #F3F4F6; padding-top: 16rpx; }
.day-title { display: block; font-size: 26rpx; font-weight: 700; color: #111827; margin-bottom: 12rpx; }
.empty-hint { font-size: 24rpx; color: #9CA3AF; padding: 20rpx 0; text-align: center; }

/* ========== 卡片视图 & 共享卡片样式 ========== */
.plan-cards { display: flex; flex-direction: column; gap: 14rpx; }
.plan-card { background: #fff; border-radius: 16rpx; padding: 18rpx 20rpx; box-shadow: 0 4rpx 16rpx rgba(15,23,42,0.06); border: 1rpx solid #F3F4F6; }
.plan-card:active { background: #FAFAFE; }
.pc-head { display: flex; justify-content: space-between; align-items: center; }
.pc-name { font-size: 26rpx; font-weight: 700; color: #111827; }
.pc-meta { font-size: 20rpx; color: #9CA3AF; }
.pc-addr { display: block; margin-top: 8rpx; font-size: 22rpx; color: #6B7280; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.pc-money { margin-top: 8rpx; font-size: 22rpx; color: #374151; }
.pc-tags { display: flex; flex-wrap: wrap; gap: 8rpx; margin-top: 8rpx; }
.pc-tag { font-size: 20rpx; padding: 4rpx 12rpx; background: #EEF2FF; color: #4F46E5; border-radius: 6rpx; }
.pc-tag.play { background: #FEF3C7; color: #B45309; font-weight: 700; }
.pc-tag.muted { background: #F3F4F6; color: #6B7280; }
.status-tag { font-weight: 600; }
.status-active { background: #DBEAFE; color: #1D4ED8; }
.status-completed { background: #DCFCE7; color: #166534; }
.status-cancelled { background: #FEE2E2; color: #B91C1C; }
.status-paused { background: #FEF3C7; color: #92400E; }
.status-draft { background: #E5E7EB; color: #4B5563; }

/* 迷你热力条 */
.heatmap { display: flex; gap: 3rpx; margin-top: 10rpx; flex-wrap: wrap; }
.heat-cell { width: 14rpx; height: 14rpx; border-radius: 3rpx; background: #F3F4F6; }
.heat-cell.h-normal { background: #BBF7D0; }
.heat-cell.h-holiday { background: #FECACA; }
.heat-cell.h-empty { background: #F3F4F6; }

</style>
