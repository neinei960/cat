<template>
  <view v-if="visible" class="cal-mask" @click="$emit('close')">
    <view class="cal-panel" @click.stop>
      <!-- 月份导航 -->
      <view class="cal-header">
        <view class="cal-nav" @click="prevMonth">‹</view>
        <text class="cal-month">{{ viewYear }}年{{ viewMonth }}月</text>
        <view class="cal-nav" @click="nextMonth">›</view>
      </view>

      <!-- 星期头 -->
      <view class="cal-weekdays">
        <text v-for="w in weekLabels" :key="w" :class="['cal-wd', w === '六' || w === '日' ? 'cal-wd-weekend' : '']">{{ w }}</text>
      </view>

      <!-- 日期网格 -->
      <view class="cal-grid">
        <view
          v-for="(day, idx) in calendarDays"
          :key="idx"
          :class="getDayClass(day)"
          @click="day.date && selectDay(day)"
        >
          <text class="cal-day-num">{{ day.day || '' }}</text>
          <text v-if="day.holiday" class="cal-holiday">{{ day.holiday }}</text>
          <text v-else-if="day.isWorkday" class="cal-workday">班</text>
          <view v-if="day.dot" class="cal-dot"></view>
        </view>
      </view>

      <!-- 底部 -->
      <view class="cal-footer">
        <view class="cal-today-btn" @click="goToday">今天</view>
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'

const props = defineProps<{
  visible: boolean
  value: string // YYYY-MM-DD
  appointmentDates?: string[] // 有预约的日期列表
}>()

const emit = defineEmits<{
  (e: 'select', date: string): void
  (e: 'close'): void
  (e: 'month-change', payload: { year: number; month: number; startDate: string; endDate: string }): void
}>()

const weekLabels = ['一', '二', '三', '四', '五', '六', '日']

const viewYear = ref(2026)
const viewMonth = ref(1)

// 节假日数据：{ "2026-01-01": { name: "元旦", isOffDay: true } }
const holidayMap = ref<Record<string, { name: string; isOffDay: boolean }>>({})
const loadedYears = new Set<number>()

function getMonthRange(year: number, month: number) {
  const startDate = `${year}-${String(month).padStart(2, '0')}-01`
  const endDate = `${year}-${String(month).padStart(2, '0')}-${String(new Date(year, month, 0).getDate()).padStart(2, '0')}`
  return { startDate, endDate }
}

function emitMonthChange(year: number, month: number) {
  const { startDate, endDate } = getMonthRange(year, month)
  emit('month-change', { year, month, startDate, endDate })
}

async function loadHolidayData(year: number) {
  if (loadedYears.has(year)) return
  loadedYears.add(year)
  try {
    const url = `https://cdn.jsdelivr.net/gh/NateScarlet/holiday-cn@master/${year}.json`
    const resp = await fetch(url)
    const data = await resp.json()
    const days = data.days || []
    const map = { ...holidayMap.value }
    for (const d of days) {
      // 同名节日只保留第一天的名字，后续天标记"休"或"班"
      const existing = Object.values(map).filter(v => v.name === d.name)
      const label = existing.length === 0 ? d.name : (d.isOffDay ? '休' : '班')
      map[d.date] = { name: label, isOffDay: d.isOffDay }
    }
    holidayMap.value = map
  } catch { /* CDN 不可用时静默降级 */ }
}

watch(() => props.visible, (v) => {
  if (v && props.value) {
    const d = new Date(props.value + 'T12:00:00')
    viewYear.value = d.getFullYear()
    viewMonth.value = d.getMonth() + 1
    loadHolidayData(viewYear.value)
    emitMonthChange(viewYear.value, viewMonth.value)
  }
})

watch(viewYear, (y) => loadHolidayData(y))

watch(() => [viewYear.value, viewMonth.value], ([year, month], [prevYear, prevMonth]) => {
  if (year === prevYear && month === prevMonth) return
  loadHolidayData(year)
  emitMonthChange(year, month)
})

function getHoliday(dateStr: string): string {
  return holidayMap.value[dateStr]?.name || ''
}

function isOfficialOffDay(dateStr: string): boolean | null {
  const h = holidayMap.value[dateStr]
  if (!h) return null // 无数据
  return h.isOffDay
}

interface CalDay {
  day: number
  date: string
  inMonth: boolean
  isToday: boolean
  isSelected: boolean
  isWeekend: boolean
  holiday: string
  isWorkday: boolean // 调休补班
  dot: boolean
}

const calendarDays = computed<CalDay[]>(() => {
  const y = viewYear.value
  const m = viewMonth.value
  const firstDay = new Date(y, m - 1, 1)
  const lastDay = new Date(y, m, 0)
  const daysInMonth = lastDay.getDate()

  // 周一=0, 周日=6
  let startWeekday = firstDay.getDay() - 1
  if (startWeekday < 0) startWeekday = 6

  const today = new Date()
  const todayStr = `${today.getFullYear()}-${String(today.getMonth() + 1).padStart(2, '0')}-${String(today.getDate()).padStart(2, '0')}`
  const apptSet = new Set(props.appointmentDates || [])

  const days: CalDay[] = []

  // 上月填充
  const prevMonthLast = new Date(y, m - 1, 0).getDate()
  for (let i = startWeekday - 1; i >= 0; i--) {
    const d = prevMonthLast - i
    const pm = m === 1 ? 12 : m - 1
    const py = m === 1 ? y - 1 : y
    const dateStr = `${py}-${String(pm).padStart(2, '0')}-${String(d).padStart(2, '0')}`
    days.push({ day: d, date: dateStr, inMonth: false, isToday: false, isSelected: false, isWeekend: false, holiday: '', isWorkday: false, dot: false })
  }

  // 当月
  for (let d = 1; d <= daysInMonth; d++) {
    const dateStr = `${y}-${String(m).padStart(2, '0')}-${String(d).padStart(2, '0')}`
    const weekday = new Date(y, m - 1, d).getDay()
    const isWeekend = weekday === 0 || weekday === 6
    const offDay = isOfficialOffDay(dateStr)
    days.push({
      day: d,
      date: dateStr,
      inMonth: true,
      isToday: dateStr === todayStr,
      isSelected: dateStr === props.value,
      isWeekend,
      holiday: getHoliday(dateStr),
      isWorkday: offDay === false, // 调休补班
      dot: apptSet.has(dateStr),
    })
  }

  // 下月填充到 42 格（6行）
  const remaining = 42 - days.length
  for (let d = 1; d <= remaining; d++) {
    const nm = m === 12 ? 1 : m + 1
    const ny = m === 12 ? y + 1 : y
    const dateStr = `${ny}-${String(nm).padStart(2, '0')}-${String(d).padStart(2, '0')}`
    days.push({ day: d, date: dateStr, inMonth: false, isToday: false, isSelected: false, isWeekend: false, holiday: '', isWorkday: false, dot: false })
  }

  return days
})

function getDayClass(day: CalDay) {
  return [
    'cal-day',
    !day.date ? 'cal-day-empty' : '',
    !day.inMonth ? 'cal-day-outside' : '',
    day.isToday ? 'cal-day-today' : '',
    day.isSelected ? 'cal-day-selected' : '',
    day.isWeekend && day.inMonth ? 'cal-day-weekend' : '',
    day.holiday ? 'cal-day-holiday' : '',
  ]
}

function selectDay(day: CalDay) {
  if (day.date) emit('select', day.date)
}

function prevMonth() {
  if (viewMonth.value === 1) { viewYear.value--; viewMonth.value = 12 }
  else viewMonth.value--
}

function nextMonth() {
  if (viewMonth.value === 12) { viewYear.value++; viewMonth.value = 1 }
  else viewMonth.value++
}

function goToday() {
  const now = new Date()
  viewYear.value = now.getFullYear()
  viewMonth.value = now.getMonth() + 1
  const todayStr = `${now.getFullYear()}-${String(now.getMonth() + 1).padStart(2, '0')}-${String(now.getDate()).padStart(2, '0')}`
  emit('select', todayStr)
}
</script>

<style scoped>
.cal-mask { position: fixed; top: 0; left: 0; right: 0; bottom: 0; background: rgba(0,0,0,0.4); z-index: 999; display: flex; align-items: flex-start; justify-content: center; padding-top: 120rpx; }
.cal-panel { background: #fff; border-radius: 24rpx; width: 92%; max-width: 700rpx; padding: 28rpx 20rpx 20rpx; box-shadow: 0 20rpx 60rpx rgba(0,0,0,0.15); }

.cal-header { display: flex; align-items: center; justify-content: space-between; margin-bottom: 20rpx; padding: 0 12rpx; }
.cal-nav { width: 64rpx; height: 64rpx; border-radius: 50%; background: #F3F4F6; display: flex; align-items: center; justify-content: center; font-size: 36rpx; color: #4B5563; }
.cal-nav:active { background: #E5E7EB; }
.cal-month { font-size: 32rpx; font-weight: 700; color: #1F2937; }

.cal-weekdays { display: grid; grid-template-columns: repeat(7, 1fr); margin-bottom: 8rpx; }
.cal-wd { text-align: center; font-size: 24rpx; color: #9CA3AF; padding: 8rpx 0; }
.cal-wd-weekend { color: #EF4444; }

.cal-grid { display: grid; grid-template-columns: repeat(7, 1fr); gap: 4rpx; }

.cal-day { position: relative; display: flex; flex-direction: column; align-items: center; justify-content: center; height: 80rpx; border-radius: 12rpx; cursor: pointer; }
.cal-day:active { background: #F3F4F6; }
.cal-day-empty { pointer-events: none; }
.cal-day-outside { opacity: 0.3; }
.cal-day-num { font-size: 28rpx; color: #1F2937; line-height: 1.2; }
.cal-day-weekend .cal-day-num { color: #EF4444; }
.cal-day-holiday .cal-day-num { color: #EF4444; }

.cal-day-today { background: #EEF2FF; }
.cal-day-today .cal-day-num { color: #4F46E5; font-weight: 700; }

.cal-day-selected { background: #4F46E5; }
.cal-day-selected .cal-day-num { color: #fff; font-weight: 700; }
.cal-day-selected .cal-holiday { color: rgba(255,255,255,0.8); }
.cal-day-selected .cal-workday { color: rgba(255,255,255,0.8); }

.cal-holiday { font-size: 16rpx; color: #EF4444; line-height: 1; margin-top: 2rpx; max-width: 100%; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.cal-workday { font-size: 16rpx; color: #0284C7; line-height: 1; margin-top: 2rpx; }

.cal-dot { position: absolute; bottom: 6rpx; width: 8rpx; height: 8rpx; border-radius: 50%; background: #4F46E5; }
.cal-day-selected .cal-dot { background: #fff; }

.cal-footer { display: flex; justify-content: center; padding-top: 16rpx; border-top: 1rpx solid #F3F4F6; margin-top: 12rpx; }
.cal-today-btn { font-size: 28rpx; color: #4F46E5; padding: 12rpx 40rpx; }
</style>
