<template>
  <SideLayout>
  <view class="page">
    <!-- Date navigation -->
    <view class="date-nav">
      <view class="nav-btn" @click="prevDay">&lt;</view>
      <picker mode="date" :value="currentDate" @change="onDatePick">
        <view class="date-display">{{ currentDate }}</view>
      </picker>
      <view class="nav-btn" @click="nextDay">&gt;</view>
      <view class="quick-date-group">
        <view
          v-for="item in quickDateOptions"
          :key="item.value"
          :class="['quick-date-btn', currentDate === item.value ? 'active' : '']"
          @click="setCurrentDate(item.value)"
        >{{ item.label }}</view>
        <view class="pending-btn" @click="togglePendingPanel">
          <text>待处理</text>
          <text v-if="pendingCount > 0" class="pending-badge">{{ pendingCount }}</text>
        </view>
      </view>
    </view>

    <!-- 待处理面板：筛选 -->
    <FilterPanel
      :visible="showPendingFilter"
      :filter="pendingFilter"
      :status-options="pendingStatusOptions"
      status-label="预约状态"
      :staff-list="pendingStaffList"
      :categories="pendingCategories"
      @close="showPendingFilter = false"
      @confirm="onPendingFilterConfirm"
    />

    <!-- 待处理面板：列表 -->
    <view v-if="showPendingPanel" class="pending-overlay" @click.self="showPendingPanel = false">
      <view class="pending-panel">
        <view class="pending-header">
          <text class="pending-title">待处理预约</text>
          <view class="pending-header-right">
            <view class="pending-filter-btn" @click="showPendingFilter = true">
              <text>筛选</text>
              <text v-if="pendingActiveFilterCount > 0" class="pending-filter-badge">{{ pendingActiveFilterCount }}</text>
            </view>
            <text class="pending-close" @click="showPendingPanel = false">✕</text>
          </view>
        </view>

        <!-- 快捷状态标签 -->
        <view class="pending-quick-tabs">
          <view :class="['p-tab', pendingFilter.status === -1 ? 'active' : '']" @click="pendingFilter.status = -1; loadPending()">全部</view>
          <view :class="['p-tab', pendingFilter.status === 0 ? 'active' : '']" @click="pendingFilter.status = 0; loadPending()">待确认</view>
          <view :class="['p-tab', pendingFilter.status === 1 ? 'active' : '']" @click="pendingFilter.status = 1; loadPending()">已确认</view>
          <view :class="['p-tab', pendingFilter.status === 2 ? 'active' : '']" @click="pendingFilter.status = 2; loadPending()">服务中</view>
          <view :class="['p-tab', pendingFilter.status === 3 ? 'active' : '']" @click="pendingFilter.status = 3; loadPending()">待结算</view>
        </view>

        <!-- 活跃筛选提示 -->
        <view class="pending-active-filters" v-if="pendingActiveFilterCount > 0">
          <text v-if="pendingFilter.dateFrom || pendingFilter.dateTo" class="pf-tag">{{ pendingFilter.dateFrom || '...' }} ~ {{ pendingFilter.dateTo || '...' }} <text @click="pendingFilter.dateFrom = ''; pendingFilter.dateTo = ''; loadPending()">✕</text></text>
          <text v-if="pendingFilter.staffId > 0" class="pf-tag">{{ getPendingStaffName(pendingFilter.staffId) }} <text @click="pendingFilter.staffId = 0; loadPending()">✕</text></text>
        </view>

        <!-- 列表 -->
        <scroll-view scroll-y class="pending-list">
          <view v-if="pendingLoading" class="pending-empty">加载中...</view>
          <view v-else-if="pendingList.length === 0" class="pending-empty">暂无待处理预约</view>
          <view
            v-for="item in pendingList" :key="item.ID"
            class="pending-card"
            :style="{ borderLeftColor: getAppointmentStatusBlockStyle(item.status).borderLeftColor }"
            @click="showPendingPanel = false; goDetail(item.ID)"
          >
            <view class="pending-card-top">
              <text class="pending-card-date">{{ item.date }} {{ item.start_time }}-{{ item.end_time }}</text>
              <view class="pending-card-status" :style="getAppointmentStatusBadgeStyle(item.status)">{{ getAppointmentStatusLabel(item.status) }}</view>
            </view>
            <view class="pending-card-body">
              <text class="pending-card-pet">{{ getPendingPetSummary(item) }}</text>
              <text class="pending-card-customer">{{ item.customer?.nickname || item.customer?.phone || '-' }}</text>
            </view>
            <text class="pending-card-staff" v-if="item.staff">洗护师: {{ item.staff.name }}</text>
            <text class="pending-card-staff" v-else style="color: #F59E0B;">待分配</text>
          </view>
        </scroll-view>
      </view>
    </view>

    <!-- 状态图例 -->
    <view class="legend-bar">
      <view class="legend-item" v-for="s in statusLegend" :key="s.status">
        <view class="legend-dot" :style="{ background: s.color }"></view>
        <text class="legend-label">{{ s.label }}</text>
      </view>
    </view>

    <!-- Stats bar -->
    <view class="stats-bar" v-if="appointments.length > 0">
      <text class="stats-text">共 {{ appointments.length }} 个预约</text>
      <text class="stats-text" v-if="unassignedCount > 0" style="color: #F59E0B;">{{ unassignedCount }} 个待分配</text>
    </view>

    <!-- Unassigned appointments -->
    <view class="unassigned-section" v-if="unassignedAppts.length > 0">
      <text class="section-label">待分配洗护师</text>
      <view class="unassigned-list">
        <view
          v-for="appt in unassignedAppts"
          :key="appt.ID"
          class="appt-card unassigned"
          :style="getUnassignedCardStyle(appt.status)"
          @click="goDetail(appt.ID)"
        >
          <view class="appt-card-top">
            <text class="appt-time-text">{{ appt.start_time }} - {{ appt.end_time }}</text>
            <view class="status-tag" :style="getAppointmentStatusBadgeStyle(appt.status)">{{ getAppointmentStatusLabel(appt.status) }}</view>
          </view>
          <view class="appt-pet-stack">
            <view
              v-for="(petInfo, index) in getPetDisplayItems(appt)"
              :key="`${appt.ID}-${petInfo.id || index}`"
              class="appt-pet-entry"
            >
              <view class="appt-pet-row">
                <text class="appt-pet-name">{{ petInfo.name }}</text>
                <text class="appt-new-tag-lg" v-if="index === 0 && isNewCustomer(appt)">新客</text>
              </view>
              <text class="appt-pet-detail">{{ petInfo.meta }}</text>
              <view class="appt-tag-row" v-if="petInfo.tags.length > 0">
                <text
                  v-for="tag in petInfo.tags"
                  :key="`${appt.ID}-${petInfo.id}-${tag.text}`"
                  :class="['appt-tag', tag.className]"
                  :style="tag.style"
                >{{ tag.text }}</text>
              </view>
              <view class="appt-pet-services" v-if="petInfo.services.length > 0">
                <text class="appt-pet-service-item" v-for="svc in petInfo.services" :key="`${appt.ID}-${petInfo.id}-svc-${svc}`">{{ svc }}</text>
              </view>
            </view>
          </view>
          <view class="appt-owner-row" v-if="getCustomerName(appt) !== '-' || getCustomerPhone(appt)">
            <text class="appt-owner-label">主人</text>
            <text class="appt-owner-name">{{ getCustomerName(appt) }}</text>
            <text class="appt-owner-phone" v-if="getCustomerPhone(appt) && getCustomerPhone(appt) !== getCustomerName(appt)">{{ getCustomerPhone(appt) }}</text>
          </view>
        </view>
      </view>
    </view>

    <!-- Staff calendar grid -->
    <scroll-view scroll-x class="calendar-scroll" v-if="staffList.length > 0">
      <view class="calendar-grid" :style="{ width: gridWidth }">
        <!-- Header row: staff names -->
        <view class="grid-header">
          <view class="time-col header-cell">时间</view>
          <view class="staff-col header-cell" v-for="staff in staffList" :key="staff.ID">
            {{ staff.name }}
            <text class="staff-count">({{ getStaffApptCount(staff.ID) }})</text>
            <text class="staff-schedule-label">{{ getScheduleLabel(staff.ID) }}</text>
          </view>
        </view>

        <!-- Time rows -->
        <view class="grid-body">
          <view :class="['time-row', hasApptAtSlot(slot) ? 'time-row-has-appt' : '']" :style="getRowMinHeight(slot) ? { minHeight: getRowMinHeight(slot) } : {}" v-for="slot in timeSlots" :key="slot">
            <view class="time-col time-label">{{ slot }}</view>
            <view
              :class="['staff-col', 'time-cell', isOffDuty(staff.ID, slot) ? 'off-duty' : '']"
              v-for="staff in staffList"
              :key="staff.ID"
              @click="!isOffDuty(staff.ID, slot) && onCellClick(staff.ID, slot)"
            >
              <view
                v-for="appt in getAppointmentsAt(staff.ID, slot)"
                :key="appt.ID"
                class="appt-block"
                :style="getAppointmentBlockStyle(appt)"
                @click.stop="goDetail(appt.ID)"
              >
                <text class="appt-time">{{ appt.start_time }}</text>
                <view class="appt-pet-stack compact">
                  <view
                    v-for="(petInfo, index) in getPetDisplayItems(appt)"
                    :key="`${appt.ID}-grid-${petInfo.id || index}`"
                    class="appt-pet-entry compact"
                  >
                    <text class="appt-pet">{{ petInfo.name }} <text class="appt-new-tag" v-if="index === 0 && isNewCustomer(appt)">新客</text></text>
                    <text class="appt-pet-info">{{ petInfo.meta }}</text>
                    <view class="appt-tag-row compact" v-if="petInfo.tags.length > 0">
                      <text
                        v-for="tag in petInfo.tags"
                        :key="`${appt.ID}-grid-${petInfo.id}-${tag.text}`"
                        :class="['appt-tag', 'compact', tag.className]"
                        :style="tag.style"
                      >{{ tag.text }}</text>
                    </view>
                    <view class="appt-pet-services compact" v-if="petInfo.services.length > 0">
                      <text class="appt-pet-service-item compact" v-for="svc in petInfo.services" :key="`${appt.ID}-grid-${petInfo.id}-svc-${svc}`">{{ svc }}</text>
                    </view>
                  </view>
                </view>
                <view class="appt-owner-row compact" v-if="getCustomerName(appt) !== '-' || getCustomerPhone(appt)">
                  <text class="appt-owner-label compact">主人</text>
                  <text class="appt-owner-name compact">{{ getCustomerName(appt) }}</text>
                  <text class="appt-owner-phone compact" v-if="getCustomerPhone(appt) && getCustomerPhone(appt) !== getCustomerName(appt)">{{ getCustomerPhone(appt) }}</text>
                </view>
              </view>
            </view>
          </view>
        </view>
      </view>
    </scroll-view>

    <view v-if="staffList.length === 0 && appointments.length === 0" class="empty">
      <text>暂无预约和排班</text>
    </view>

    <!-- FAB -->
    <view class="fab" @click="goCreate">
      <text class="fab-icon">+</text>
    </view>
  </view>
  </SideLayout>
</template>

<script setup lang="ts">
import SideLayout from '@/components/SideLayout.vue'
import { ref, reactive, computed, onMounted } from 'vue'
import { onShow } from '@dcloudio/uni-app'
import FilterPanel from '@/components/FilterPanel.vue'
import { getAppointmentCalendar, getAppointmentList } from '@/api/appointment'
import { getDashboardOverview } from '@/api/dashboard'
import { getStaffList, getStaffSchedule } from '@/api/staff'
import { getCategoryTree } from '@/api/service-category'
import { getShop } from '@/api/shop'
import { getPersonalityBg, getPersonalityColor } from '@/utils/personality'
import {
  APPOINTMENT_STATUS_META,
  getAppointmentStatusBadgeStyle,
  getAppointmentStatusBlockStyle,
  getAppointmentStatusLabel,
} from '@/utils/appointment-status'

const statusLegend = [
  { status: 0, label: '待确认', color: APPOINTMENT_STATUS_META[0].blockAccent },
  { status: 1, label: '已确认', color: APPOINTMENT_STATUS_META[1].blockAccent },
  { status: 2, label: '服务中', color: APPOINTMENT_STATUS_META[2].blockAccent },
  { status: 3, label: '待结算', color: APPOINTMENT_STATUS_META[3].blockAccent },
  { status: 6, label: '已到店', color: APPOINTMENT_STATUS_META[6].blockAccent },
  { status: 5, label: '未到店', color: APPOINTMENT_STATUS_META[5].blockAccent },
]

function localDateStr(d: Date = new Date()): string {
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`
}

function offsetDateStr(offsetDays: number): string {
  const d = new Date()
  d.setDate(d.getDate() + offsetDays)
  return localDateStr(d)
}

const currentDate = ref(localDateStr())
const staffList = ref<Staff[]>([])
const appointments = ref<any[]>([])

// 待处理面板
const showPendingPanel = ref(false)
const showPendingFilter = ref(false)
const pendingCount = ref(0)
const pendingList = ref<any[]>([])
const pendingLoading = ref(false)
const pendingStaffList = ref<any[]>([])
const pendingCategories = ref<any[]>([])
const pendingFilter = reactive({
  status: -1, dateFrom: '', dateTo: '', staffId: 0, payMethod: '', categoryId: 0,
})
const pendingStatusOptions = [
  { value: 0, label: '待确认' },
  { value: 1, label: '已确认' },
  { value: 2, label: '服务中' },
  { value: 3, label: '待结算' },
  { value: 4, label: '已取消' },
  { value: 5, label: '未到店' },
]
const pendingActiveFilterCount = computed(() => {
  let c = 0
  if (pendingFilter.dateFrom || pendingFilter.dateTo) c++
  if (pendingFilter.staffId > 0) c++
  if (pendingFilter.categoryId > 0) c++
  return c
})
function getPendingStaffName(id: number) {
  return pendingStaffList.value.find((s: any) => s.ID === id)?.name || '未知'
}
function onPendingFilterConfirm(f: any) {
  Object.assign(pendingFilter, f)
  loadPending()
}

function togglePendingPanel() {
  showPendingPanel.value = !showPendingPanel.value
  if (showPendingPanel.value) {
    loadPending()
    loadPendingFilterOptions()
  }
}
async function loadPending() {
  pendingLoading.value = true
  try {
    const params: any = { page: 1, page_size: 100 }
    if (pendingFilter.status >= 0) params.status = pendingFilter.status
    if (pendingFilter.dateFrom) params.date_from = pendingFilter.dateFrom
    if (pendingFilter.dateTo) params.date_to = pendingFilter.dateTo
    if (pendingFilter.staffId > 0) params.staff_id = pendingFilter.staffId
    const res = await getAppointmentList(params)
    pendingList.value = res.data.list || []
  } finally { pendingLoading.value = false }
}
async function loadPendingCount() {
  try {
    const res = await getDashboardOverview()
    pendingCount.value = res.data?.pending_appointments || 0
  } catch {}
}
async function loadPendingFilterOptions() {
  if (pendingStaffList.value.length > 0) return // already loaded
  try {
    const [stRes, catRes] = await Promise.all([
      getStaffList({ page: 1, page_size: 50 }),
      getCategoryTree(),
    ])
    pendingStaffList.value = (stRes.data?.list || []).filter((s: any) => s.status === 1)
    pendingCategories.value = (catRes.data || []).filter((c: any) => !c.parent_id && c.status === 1)
  } catch {}
}
function getPendingPetSummary(item: any) {
  if (Array.isArray(item?.pets) && item.pets.length > 0) {
    return item.pets.map((p: any) => p.pet?.name).filter(Boolean).join('、') || '-'
  }
  return item?.pet?.name || '-'
}
const staffScheduleMap = ref<Record<number, { start: string; end: string; dayOff: boolean }>>({})
const shopOpenTime = ref('10:00')
const shopCloseTime = ref('22:00')
const showDatePicker = ref(false)
const quickDateOptions = computed(() => [
  { label: '今天', value: offsetDateStr(0) },
  { label: '明天', value: offsetDateStr(1) },
  { label: '后天', value: offsetDateStr(2) },
])

const timeSlots = computed(() => {
  // 以店铺营业时间为基础范围
  const shopStart = floorToHalfHour(parseTime(shopOpenTime.value))
  const shopEnd = ceilToHalfHour(parseTime(shopCloseTime.value))

  // 取员工排班和营业时间的交集：最早开始 = max(营业开始, 最早排班)，最晚结束 = min(营业结束, 最晚排班)
  const activeSchedules = Object.values(staffScheduleMap.value).filter(
    (schedule) => !schedule.dayOff && !!schedule.start && !!schedule.end
  )

  let startMin = shopStart
  let endMin = shopEnd

  if (activeSchedules.length > 0) {
    const scheduleStart = Math.min(...activeSchedules.map(s => floorToHalfHour(parseTime(s.start))))
    const scheduleEnd = Math.max(...activeSchedules.map(s => ceilToHalfHour(parseTime(s.end))))
    // 交集：取较晚的开始、较早的结束
    startMin = Math.max(shopStart, scheduleStart)
    endMin = Math.min(shopEnd, scheduleEnd)
  }

  // 确保预约不会超出范围
  const activeAppointments = appointments.value.filter((appt) => appt.status !== 4)
  activeAppointments.forEach((appt) => {
    startMin = Math.min(startMin, floorToHalfHour(parseTime(appt.start_time)))
    endMin = Math.max(endMin, ceilToHalfHour(parseTime(appt.end_time)))
  })

  if (startMin >= endMin) return []

  const slots: string[] = []
  for (let minute = startMin; minute <= endMin; minute += 30) {
    slots.push(minutesToTime(minute))
  }
  return slots
})

const gridWidth = computed(() => {
  const cols = staffList.value.length || 1
  return (160 + cols * 340) + 'rpx'
})

// Unassigned appointments (no staff_id)
const unassignedAppts = computed(() =>
  appointments.value.filter(a => !a.staff_id && a.status !== 4)
)
const unassignedCount = computed(() => unassignedAppts.value.length)

function sortStaffList(list: Staff[]) {
  return [...list].sort((a, b) => {
    const roleDiff = Number(a.role === 'admin') - Number(b.role === 'admin')
    if (roleDiff !== 0) return roleDiff
    return a.ID - b.ID
  })
}

function prevDay() {
  const d = new Date(currentDate.value + 'T12:00:00')
  d.setDate(d.getDate() - 1)
  setCurrentDate(localDateStr(d))
}

function nextDay() {
  const d = new Date(currentDate.value + 'T12:00:00')
  d.setDate(d.getDate() + 1)
  setCurrentDate(localDateStr(d))
}

function setCurrentDate(date: string) {
  if (!date || currentDate.value === date) return
  currentDate.value = date
  loadData()
}

function onDatePick(e: any) {
  setCurrentDate(e.detail.value)
}

function getAppointmentsAt(staffId: number, timeSlot: string): any[] {
  return appointments.value.filter(a =>
    a.staff_id === staffId && a.start_time === timeSlot && a.status !== 4
  )
}

function hasApptAtSlot(slot: string): boolean {
  return appointments.value.some(a => a.start_time === slot && a.status !== 4)
}

function getStaffApptCount(staffId: number): number {
  return appointments.value.filter(a => a.staff_id === staffId && a.status !== 4).length
}

function getApptSlots(appt: any): number {
  const startMin = parseTime(appt.start_time)
  const endMin = parseTime(appt.end_time)
  const duration = endMin - startMin
  return Math.max(Math.ceil(duration / 30), 1)
}

// 估算单个预约卡片的内容高度（rpx）
function estimateApptContentRpx(appt: any): number {
  const pets = getPetDisplayItems(appt)
  let h = 50
  for (const pet of pets) {
    h += 42 + 30
    if (pet.tags.length > 0) h += 40
    if (pet.services.length > 0) h += 44
    h += 16
  }
  if (getCustomerName(appt) !== '-') h += 42
  h += 24
  return h
}

// 计算每个时间槽的行高（rpx），基于该槽涉及的所有预约
const slotRowHeights = computed(() => {
  const heights: Record<string, number> = {}
  for (const slot of timeSlots.value) {
    heights[slot] = 80 // 默认
  }
  // 遍历所有预约，将内容高度平均分配到它所跨越的每个槽
  for (const appt of appointments.value) {
    if (appt.status === 4) continue
    const slots = getApptSlots(appt)
    const contentH = estimateApptContentRpx(appt)
    const perSlot = Math.ceil(contentH / slots)
    // 应用到该预约覆盖的每个时间槽
    const startMin = parseTime(appt.start_time)
    for (let i = 0; i < slots; i++) {
      const slotMin = startMin + i * 30
      const slotStr = minutesToTime(slotMin)
      if (heights[slotStr] !== undefined) {
        heights[slotStr] = Math.max(heights[slotStr], perSlot)
      }
    }
  }
  return heights
})

function getAppointmentBlockStyle(appt: any) {
  // 计算该卡片所跨行的总高度（rpx）
  const slots = getApptSlots(appt)
  const startMin = parseTime(appt.start_time)
  let totalH = 0
  for (let i = 0; i < slots; i++) {
    const slotStr = minutesToTime(startMin + i * 30)
    totalH += slotRowHeights.value[slotStr] || 80
  }
  const heightVw = (totalH - 2) / 750 * 100
  return {
    ...getAppointmentStatusBlockStyle(appt?.status),
    height: `${heightVw.toFixed(2)}vw`,
  }
}

function getRowMinHeight(slot: string): string {
  const h = slotRowHeights.value[slot] || 80
  if (h <= 80) return ''
  return `${h}rpx`
}

function getUnassignedCardStyle(status: number) {
  const blockStyle = getAppointmentStatusBlockStyle(status)
  return {
    borderLeftColor: blockStyle.borderLeftColor,
  }
}

function parseTime(t: string): number {
  const [h, m] = t.split(':').map(Number)
  return h * 60 + m
}

function minutesToTime(totalMinutes: number): string {
  const hours = Math.floor(totalMinutes / 60)
  const minutes = totalMinutes % 60
  return `${String(hours).padStart(2, '0')}:${String(minutes).padStart(2, '0')}`
}

function floorToHalfHour(totalMinutes: number) {
  return Math.floor(totalMinutes / 30) * 30
}

function ceilToHalfHour(totalMinutes: number) {
  return Math.ceil(totalMinutes / 30) * 30
}

async function loadData() {
  const [staffRes, apptRes, shopRes] = await Promise.all([
    getStaffList({ page: 1, page_size: 100 }),
    getAppointmentCalendar(currentDate.value, currentDate.value),
    getShop(),
  ])
  if (shopRes.data?.open_time) shopOpenTime.value = shopRes.data.open_time
  if (shopRes.data?.close_time) shopCloseTime.value = shopRes.data.close_time
  staffList.value = sortStaffList((staffRes.data.list || []).filter((s: Staff) => s.status === 1))
  appointments.value = apptRes.data || []

  // 加载每个员工当天的排班
  const scheduleMap: Record<number, { start: string; end: string; dayOff: boolean }> = {}
  const date = currentDate.value
  await Promise.all(staffList.value.map(async (staff) => {
    try {
      const res = await getStaffSchedule(staff.ID, date, date)
      const s = (res.data || [])[0]
      if (s) {
        scheduleMap[staff.ID] = { start: s.start_time || '12:00', end: s.end_time || '22:00', dayOff: s.is_day_off }
      }
    } catch { /* no schedule */ }
  }))
  staffScheduleMap.value = scheduleMap
}

function isOffDuty(staffId: number, timeSlot: string): boolean {
  const schedule = staffScheduleMap.value[staffId]
  if (!schedule) return false // 无排班数据不灰显
  if (schedule.dayOff) return true
  return timeSlot < schedule.start || timeSlot >= schedule.end
}

function getScheduleLabel(staffId: number): string {
  const s = staffScheduleMap.value[staffId]
  if (!s) return ''
  if (s.dayOff) return '休息'
  return `${s.start}-${s.end}`
}

function getAppointmentPets(appt: any) {
  if (Array.isArray(appt?.pets) && appt.pets.length > 0) {
    return appt.pets
  }
  if (appt?.pet) {
    return [{
      pet_id: appt.pet.ID,
      pet: appt.pet,
      services: appt.services || [],
    }]
  }
  return []
}

function getPetSummary(appt: any) {
  if (appt?.pet_summary) return appt.pet_summary
  const names = getAppointmentPets(appt)
    .map((petItem: any) => petItem.pet?.name)
    .filter(Boolean)
  if (names.length === 0) return '-'
  if (names.length === 1) return names[0]
  return `${names[0]}等${names.length}只`
}

function getPrimaryPet(appt: any) {
  return getAppointmentPets(appt)[0]?.pet || null
}

function getPetCount(appt: any) {
  return appt?.pet_count || getAppointmentPets(appt).length
}

function formatPetMeta(pet: any) {
  const parts: string[] = []
  if (pet?.breed) parts.push(pet.breed)
  if (pet?.gender === 1) parts.push('弟弟')
  if (pet?.gender === 2) parts.push('妹妹')
  return parts.join(' · ') || '未填写宠物信息'
}

function getPetTagItems(pet: any) {
  const tags: Array<{ text: string; className: string; style?: string }> = []
  const age = getPetAge(pet?.birth_date)
  if (age) tags.push({ text: age, className: 'tag-age' })
  if (pet?.fur_level) tags.push({ text: pet.fur_level, className: 'tag-fur' })
  if (pet?.neutered) tags.push({ text: '已绝育', className: 'tag-neutered' })
  if (pet?.personality) {
    tags.push({
      text: pet.personality,
      className: 'tag-personality',
      style: `background:${getPersonalityBg(pet.personality)};color:${getPersonalityColor(pet.personality)};`,
    })
  }
  if (pet?.aggression && pet.aggression !== '无') {
    tags.push({ text: `⚡ ${pet.aggression}`, className: 'tag-aggression' })
  }
  return tags
}

function getPetDisplayItems(appt: any) {
  const petItems = getAppointmentPets(appt)
  if (petItems.length === 0) {
    return [{ id: 'empty', name: '-', meta: '未填写宠物信息', tags: [] as Array<{ text: string; className: string; style?: string }>, services: [] as string[] }]
  }
  return petItems.map((petItem: any, index: number) => {
    const pet = petItem?.pet
    const serviceNames = (petItem.services || [])
      .map((s: any) => s.service_name)
      .filter(Boolean)
    return {
      id: pet?.ID || index,
      name: pet?.name || `猫咪${index + 1}`,
      meta: pet ? formatPetMeta(pet) : '未填写宠物信息',
      tags: pet ? getPetTagItems(pet) : [],
      services: serviceNames,
    }
  })
}

function getPetAge(birthDate?: string) {
  if (!birthDate) return ''
  const birth = new Date(birthDate)
  if (Number.isNaN(birth.getTime())) return ''
  const now = new Date()
  const months = (now.getFullYear() - birth.getFullYear()) * 12 + (now.getMonth() - birth.getMonth())
  if (months < 1) return '不到1个月'
  if (months < 12) return `${months}个月`
  const years = Math.floor(months / 12)
  const rem = months % 12
  return rem > 0 ? `${years}岁${rem}个月` : `${years}岁`
}

function getPetDetail(appt: any) {
  const pet = getPrimaryPet(appt)
  const parts = pet ? [formatPetMeta(pet)] : []
  if (getPetCount(appt) > 1) {
    parts.push(`共${getPetCount(appt)}只`)
  }
  return parts.join(' · ') || (getPetCount(appt) > 1 ? `共${getPetCount(appt)}只猫` : '未填写宠物信息')
}

function getCompactPetMeta(appt: any) {
  const pet = getPrimaryPet(appt)
  const parts: string[] = []
  if (pet) parts.push(formatPetMeta(pet))
  if (getPetCount(appt) > 1) parts.push(`${getPetCount(appt)}只`)
  return parts.join(' · ') || '多猫预约'
}

function getPetTagSummary(appt: any) {
  const pet = getPrimaryPet(appt)
  return formatPetTags(pet)
}

function getServiceSummary(appt: any) {
  const petGroups = getAppointmentPets(appt)
    .map((petItem: any) => {
      const serviceNames = (petItem.services || [])
        .map((service: any) => service.service_name)
        .filter(Boolean)
      if (serviceNames.length === 0) return ''
      const petName = petItem.pet?.name || '宠物'
      return `${petName}: ${serviceNames.join(' + ')}`
    })
    .filter(Boolean)

  if (petGroups.length > 0) {
    return petGroups.join('；')
  }

  return (appt.services || []).map((service: any) => service.service_name).join(' + ') || '-'
}

function isNewCustomer(appt: any): boolean {
  const name = appt.customer?.nickname || ''
  return name.startsWith('散客') || !appt.customer?.phone
}

function getCustomerDisplay(appt: any): string {
  const name = appt.customer?.nickname || appt.customer?.phone || '-'
  if (name.startsWith('散客')) return name.replace('散客', '新客')
  return name
}

function getCustomerName(appt: any): string {
  const nickname = appt.customer?.nickname || ''
  if (nickname.startsWith('散客')) return nickname.replace('散客', '新客')
  if (nickname) return nickname
  return appt.customer?.phone || '-'
}

function getCustomerPhone(appt: any): string {
  return appt.customer?.phone || ''
}

function goCreate() {
  uni.navigateTo({ url: `/pages/appointment/create?date=${currentDate.value}` })
}

function goDetail(id: number) {
  uni.navigateTo({ url: `/pages/appointment/detail?id=${id}` })
}

function onCellClick(staffId: number, time: string) {
  uni.navigateTo({ url: `/pages/appointment/create?date=${currentDate.value}&staff_id=${staffId}&time=${time}` })
}

onMounted(() => { loadData(); loadPendingCount() })
onShow(() => { loadData(); loadPendingCount() })
</script>

<style scoped>
.page { display: flex; flex-direction: column; height: 100vh; background: #F2F5F9; }
.date-nav { display: flex; align-items: center; gap: 12rpx; padding: 16rpx 24rpx; background: #fff; box-shadow: 0 8rpx 24rpx rgba(15, 23, 42, 0.05); flex-wrap: wrap; }
.nav-btn { width: 80rpx; height: 80rpx; min-width: 80rpx; min-height: 80rpx; display: flex; align-items: center; justify-content: center; background: #EEF2F7; border-radius: 50%; font-size: 28rpx; color: #334155; transition: background-color 0.15s, transform 0.1s; }
.nav-btn:active { background: #C7D2FE; transform: scale(0.9); }
.date-display { flex: 1; min-width: 220rpx; text-align: center; font-size: 30rpx; font-weight: 600; color: #0F172A; }
.quick-date-group { display: flex; gap: 10rpx; margin-left: auto; flex-wrap: wrap; justify-content: flex-end; }
.quick-date-btn { min-width: 84rpx; height: 60rpx; padding: 0 18rpx; display: flex; align-items: center; justify-content: center; border-radius: 999rpx; font-size: 24rpx; color: #475569; background: #F1F5F9; border: 1rpx solid transparent; }
.quick-date-btn.active { color: #4338CA; background: #E0E7FF; border-color: rgba(67, 56, 202, 0.18); font-weight: 600; }

.legend-bar { display: flex; flex-wrap: wrap; gap: 16rpx; padding: 14rpx 24rpx; background: #fff; border-top: 1rpx solid #E2E8F0; }
.legend-item { display: flex; align-items: center; gap: 6rpx; }
.legend-dot { width: 16rpx; height: 16rpx; border-radius: 4rpx; flex-shrink: 0; }
.legend-label { font-size: 22rpx; color: #6B7280; }
.stats-bar { display: flex; gap: 24rpx; padding: 12rpx 24rpx; background: #fff; border-top: 1rpx solid #E2E8F0; }
.stats-text { font-size: 24rpx; color: #64748B; }

.unassigned-section { padding: 16rpx 24rpx; }
.section-label { font-size: 26rpx; font-weight: 600; color: #D97706; display: block; margin-bottom: 12rpx; }
.unassigned-list { display: flex; flex-wrap: wrap; gap: 12rpx; }
.unassigned-list .appt-card { flex: 1; min-width: 420rpx; max-width: 100%; }
.appt-card { background: #fff; border-radius: 16rpx; padding: 14rpx 18rpx; border: 1rpx solid #E2E8F0; box-shadow: 0 10rpx 24rpx rgba(15, 23, 42, 0.06); }
.appt-card.unassigned { border-left: 8rpx solid transparent; }
.appt-card-top { display: flex; justify-content: space-between; align-items: center; margin-bottom: 8rpx; }
.appt-time-text { font-size: 26rpx; font-weight: 600; color: #0F172A; }
.status-tag { font-size: 20rpx; padding: 4rpx 12rpx; border-radius: 999rpx; }
.appt-pet-name { font-size: 28rpx; font-weight: 600; color: #0F172A; display: block; }
.appt-pet-services { display: flex; flex-wrap: wrap; gap: 6rpx; margin-top: 6rpx; }
.appt-pet-services.compact { gap: 4rpx; margin-top: 4rpx; }
.appt-pet-service-item { font-size: 22rpx; color: #4338CA; background: #EEF2FF; padding: 4rpx 14rpx; border-radius: 8rpx; line-height: 1.4; }
.appt-pet-service-item.compact { font-size: 18rpx; padding: 2rpx 10rpx; }
.appt-pet-stack { display: flex; flex-direction: column; gap: 8rpx; }
.appt-pet-stack.compact { gap: 6rpx; }
.appt-pet-entry { padding: 8rpx 0; border-bottom: 1rpx dashed rgba(148, 163, 184, 0.35); }
.appt-pet-entry:first-child { padding-top: 0; }
.appt-pet-entry:last-child { padding-bottom: 0; border-bottom: 0; }
.appt-pet-entry.compact { padding: 6rpx 0; }
.appt-owner-row { display: flex; align-items: center; flex-wrap: wrap; gap: 8rpx; margin-top: 10rpx; }
.appt-owner-row.compact { gap: 6rpx; margin-top: 8rpx; }
.appt-owner-label { font-size: 18rpx; color: #64748B; background: #F1F5F9; padding: 2rpx 10rpx; border-radius: 999rpx; }
.appt-owner-label.compact { font-size: 16rpx; padding: 2rpx 8rpx; }
.appt-owner-name { font-size: 22rpx; color: #475569; font-weight: 500; }
.appt-owner-name.compact { font-size: 18rpx; }
.appt-owner-phone { font-size: 18rpx; color: #64748B; background: rgba(241, 245, 249, 0.9); padding: 2rpx 10rpx; border-radius: 999rpx; }
.appt-owner-phone.compact { font-size: 16rpx; padding: 2rpx 8rpx; }
.appt-tag-row { display: flex; flex-wrap: wrap; gap: 8rpx; margin-top: 6rpx; }
.appt-tag-row.compact { gap: 6rpx; margin-top: 4rpx; }

.calendar-scroll { flex: 1; }
.calendar-grid { min-width: 100%; padding-bottom: 32rpx; }
.grid-header { display: flex; position: sticky; top: 0; z-index: 10; background: rgba(255, 255, 255, 0.96); border-bottom: 2rpx solid #CBD5E1; box-shadow: 0 8rpx 18rpx rgba(15, 23, 42, 0.05); }
.header-cell { padding: 18rpx 8rpx; font-size: 26rpx; font-weight: 600; color: #334155; text-align: center; background: rgba(248, 250, 252, 0.92); }
.staff-count { font-size: 20rpx; color: #64748B; font-weight: 500; }
.staff-schedule-label { font-size: 18rpx; color: #94A3B8; display: block; font-weight: 400; }
.time-col { width: 160rpx; min-width: 160rpx; background: #F8FAFC; }
.staff-col { width: 340rpx; min-width: 340rpx; border-left: 1rpx solid #E2E8F0; }

.grid-body { position: relative; background: linear-gradient(180deg, #FFFFFF 0%, #F8FAFC 100%); overflow: visible; }
.time-row { display: flex; min-height: 80rpx; border-bottom: 1rpx solid #E2E8F0; background: #fff; position: relative; }
.time-row:nth-child(odd) { background: #F6F8FA; }
.time-label { font-size: 22rpx; color: #64748B; padding: 8rpx 12rpx; line-height: 64rpx; font-weight: 600; background: rgba(248, 250, 252, 0.92); }
.time-cell { position: relative; min-height: 80rpx; padding: 8rpx; background: rgba(255, 255, 255, 0.78); overflow: visible; }
.time-cell.off-duty { background: repeating-linear-gradient(135deg, #F1F5F9, #F1F5F9 8rpx, #E8ECF1 8rpx, #E8ECF1 16rpx); opacity: 0.5; }

.appt-block {
  position: absolute;
  top: 0;
  left: 6rpx;
  right: 6rpx;
  background: linear-gradient(180deg, #F8FAFF 0%, #E8EEFF 100%);
  border: 1rpx solid rgba(67, 56, 202, 0.18);
  border-left: 8rpx solid #4338CA;
  border-radius: 14rpx;
  padding: 10rpx 12rpx;
  overflow: visible;
  cursor: pointer;
  box-shadow: 0 12rpx 24rpx rgba(67, 56, 202, 0.12);
  z-index: 5;
}
.appt-block:hover {
  z-index: 20;
}
.appt-time { font-size: 20rpx; color: #475569; display: block; font-weight: 600; margin-bottom: 2rpx; }
.appt-pet { font-size: 24rpx; font-weight: 700; color: #0F172A; display: block; }
.appt-pet-info { font-size: 18rpx; color: #64748B; display: block; line-height: 1.4; }
.appt-tag { display: inline-flex; align-items: center; padding: 4rpx 12rpx; border-radius: 999rpx; font-size: 18rpx; line-height: 1.2; background: #F3F4F6; color: #4B5563; }
.appt-tag.compact { padding: 2rpx 10rpx; font-size: 16rpx; }
.appt-tag.tag-fur { background: #EEF2FF; color: #4F46E5; }
.appt-tag.tag-neutered { background: #ECFDF5; color: #047857; }
.appt-tag.tag-aggression { background: #FEF2F2; color: #DC2626; }
.appt-new-tag { font-size: 18rpx; color: #F59E0B; font-weight: 600; }
.appt-new-tag-lg { font-size: 20rpx; color: #fff; background: #F59E0B; padding: 2rpx 12rpx; border-radius: 8rpx; margin-left: 8rpx; }
.appt-pet-row { display: flex; align-items: center; }
.appt-pet-detail { font-size: 22rpx; color: #64748B; display: block; margin-top: 4rpx; }

.empty { text-align: center; padding: 100rpx 0; color: #9CA3AF; font-size: 28rpx; }

/* Pending button */
.pending-btn { display: flex; align-items: center; gap: 6rpx; padding: 0 20rpx; height: 60rpx; border-radius: 999rpx; background: #FEF3C7; color: #92400E; font-size: 24rpx; font-weight: 600; border: 1rpx solid #FDE68A; }
.pending-badge { background: #EF4444; color: #fff; font-size: 20rpx; min-width: 32rpx; height: 32rpx; border-radius: 999rpx; display: flex; align-items: center; justify-content: center; padding: 0 8rpx; font-weight: 700; }

/* Pending overlay & panel */
.pending-overlay { position: fixed; top: 0; left: 0; right: 0; bottom: 0; background: rgba(0,0,0,0.45); z-index: 1500; display: flex; align-items: flex-end; justify-content: center; }
.pending-panel { background: #fff; border-radius: 24rpx 24rpx 0 0; width: 100%; max-height: 85vh; display: flex; flex-direction: column; padding-bottom: calc(50px + env(safe-area-inset-bottom)); }
.pending-header { display: flex; justify-content: space-between; align-items: center; padding: 24rpx 28rpx; border-bottom: 1rpx solid #F3F4F6; }
.pending-header-right { display: flex; align-items: center; gap: 16rpx; }
.pending-title { font-size: 32rpx; font-weight: 700; color: #1F2937; }
.pending-close { font-size: 36rpx; color: #9CA3AF; padding: 8rpx; }
.pending-filter-btn { display: flex; align-items: center; gap: 6rpx; font-size: 24rpx; color: #374151; background: #F3F4F6; padding: 10rpx 20rpx; border-radius: 10rpx; border: 1rpx solid #E5E7EB; }
.pending-filter-badge { background: #EF4444; color: #fff; font-size: 18rpx; min-width: 26rpx; height: 26rpx; border-radius: 999rpx; display: flex; align-items: center; justify-content: center; padding: 0 6rpx; }

/* Pending quick tabs & active filters */
.pending-quick-tabs { display: flex; gap: 8rpx; padding: 16rpx 28rpx; flex-wrap: wrap; border-bottom: 1rpx solid #F3F4F6; }
.p-tab { font-size: 22rpx; padding: 8rpx 20rpx; border-radius: 20rpx; background: #F3F4F6; color: #6B7280; }
.p-tab.active { background: #4F46E5; color: #fff; }
.pending-active-filters { display: flex; flex-wrap: wrap; gap: 8rpx; padding: 12rpx 28rpx; }
.pf-tag { font-size: 22rpx; color: #4F46E5; background: #EEF2FF; padding: 6rpx 16rpx; border-radius: 20rpx; }

/* Pending list */
.pending-list { flex: 1; overflow-y: auto; padding: 16rpx 28rpx; max-height: 55vh; }
.pending-empty { text-align: center; padding: 60rpx 0; color: #9CA3AF; font-size: 26rpx; }
.pending-card { background: #fff; border-radius: 12rpx; padding: 20rpx; margin-bottom: 16rpx; border: 1rpx solid #E5E7EB; border-left: 8rpx solid #4F46E5; }
.pending-card-top { display: flex; justify-content: space-between; align-items: center; margin-bottom: 8rpx; }
.pending-card-date { font-size: 24rpx; color: #6B7280; }
.pending-card-status { font-size: 20rpx; padding: 4rpx 12rpx; border-radius: 16rpx; }
.pending-card-body { margin-bottom: 6rpx; }
.pending-card-pet { font-size: 28rpx; font-weight: 600; color: #1F2937; }
.pending-card-customer { font-size: 24rpx; color: #6B7280; margin-left: 12rpx; }
.pending-card-staff { font-size: 22rpx; color: #6B7280; }

.fab {
  position: fixed;
  right: 32rpx;
  bottom: calc(50px + env(safe-area-inset-bottom) + 24rpx);
  width: 100rpx;
  height: 100rpx;
  border-radius: 50%;
  background: #4F46E5;
  color: #fff;
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: 0 8rpx 24rpx rgba(79,70,229,0.4);
  z-index: 100;
}
.fab-icon { font-size: 56rpx; line-height: 1; font-weight: 300; transform: translateY(-2rpx); }
</style>
