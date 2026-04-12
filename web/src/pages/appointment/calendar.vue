<template>
  <SideLayout>
  <view class="page">
    <view class="calendar-hero">
      <view class="calendar-hero-head">
        <view class="calendar-hero-copy">
          <text class="calendar-hero-title">预约看板</text>
          <text class="calendar-hero-subtitle">{{ calendarHeroSubtitle }}</text>
        </view>
        <view class="pending-btn" @click="togglePendingPanel">
          <text>待处理</text>
          <text v-if="pendingCount > 0" class="pending-badge">{{ pendingCount }}</text>
        </view>
      </view>

      <!-- Date navigation -->
      <view class="date-nav">
        <view class="nav-btn" @click="prevDay">&lt;</view>
        <view class="date-display" @click="showCalPicker = true">{{ formatDateDisplay(currentDate) }}</view>
        <view class="nav-btn" @click="nextDay">&gt;</view>
        <view class="quick-date-group">
          <view
            v-for="item in quickDateOptions"
            :key="item.value"
            :class="['quick-date-btn', currentDate === item.value ? 'active' : '']"
            @click="setCurrentDate(item.value)"
          >{{ item.label }}</view>
        </view>
      </view>

      <view class="summary-strip">
        <view
          v-for="card in calendarSummaryCards"
          :key="card.key"
          :class="['summary-card', `summary-card-${card.tone}`]"
        >
          <text class="summary-card-label">{{ card.label }}</text>
          <text class="summary-card-value">{{ card.value }}</text>
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
            <view
              v-if="isDesktopInteraction && !pendingSelectionMode"
              class="pending-select-btn"
              @click="enterPendingSelectionMode()"
            >批量选择</view>
            <view
              v-if="pendingSelectionMode"
              class="pending-delete-btn"
              :class="{ disabled: pendingSelectedIds.length === 0 }"
              @click="confirmDeletePendingSelected"
            >
              <text>删除</text>
              <text v-if="pendingSelectedIds.length > 0" class="pending-delete-count">{{ pendingSelectedIds.length }}</text>
            </view>
            <view
              v-if="pendingSelectionMode"
              class="pending-cancel-btn"
              @click="exitPendingSelectionMode"
            >取消</view>
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
          <view :class="['p-tab', pendingFilter.status === 7 ? 'active' : '']" @click="pendingFilter.status = 7; loadPending()">已开单</view>
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
            :class="['pending-card', pendingSelectionMode ? 'selection-mode' : '', isPendingSelected(item.ID) ? 'selected' : '']"
            :style="{ borderLeftColor: getAppointmentStatusBlockStyle(item.status).borderLeftColor }"
            @touchstart="startPendingLongPress(item.ID)"
            @touchend="clearPendingLongPress"
            @touchcancel="clearPendingLongPress"
            @touchmove="clearPendingLongPress"
            @click="handlePendingCardClick(item.ID)"
          >
            <view v-if="pendingSelectionMode" class="pending-select-dot">
              <view :class="['pending-select-inner', isPendingSelected(item.ID) ? 'selected' : '']"></view>
            </view>
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

    <view v-if="pendingConfirmDelete" class="pending-confirm-mask" @click.self="cancelPendingDelete">
      <view class="pending-confirm-dialog" @click.stop>
        <view class="pending-confirm-head">
          <text class="pending-confirm-title">确认删除</text>
          <text class="pending-confirm-close" @click="cancelPendingDelete">✕</text>
        </view>
        <text class="pending-confirm-text">确认删除 {{ pendingSelectedIds.length }} 条预约？删除后不可恢复。</text>
        <view class="pending-confirm-actions">
          <view class="pending-confirm-btn secondary" @click="cancelPendingDelete">取消</view>
          <view class="pending-confirm-btn danger" @click="executePendingDelete">确认删除</view>
        </view>
      </view>
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
              <view class="appt-pet-note" v-if="petInfo.noteText">{{ petInfo.noteText }}</view>
            </view>
          </view>
          <view class="appt-notes" v-if="getAppointmentGeneralNoteSummary(appt)">{{ getAppointmentGeneralNoteSummary(appt) }}</view>
        </view>
      </view>
    </view>

    <!-- Staff calendar grid -->
    <scroll-view scroll-x class="calendar-scroll" v-if="staffList.length > 0">
      <view class="calendar-grid" :style="{ width: gridWidth }">
        <!-- Header row: staff names -->
        <view class="grid-header">
          <view class="time-col header-cell">时间</view>
          <view
            :class="[
              'staff-col',
              'header-cell',
              'staff-header-cell',
              canReorderStaff ? 'reorderable' : '',
              draggingStaffId === staff.ID ? 'dragging' : '',
              savingStaffOrder ? 'saving' : '',
            ]"
            v-for="(staff, index) in staffList"
            :key="staff.ID"
            :data-staff-id="staff.ID"
            @longpress.stop="beginStaffDrag(index, $event)"
            @mousedown.stop.prevent="isDesktopInteraction ? beginStaffDrag(index, $event) : undefined"
          >
            <view class="staff-name-row">
              <text
                :class="['staff-move-arrow', index === 0 || savingStaffOrder ? 'disabled' : '']"
                @click.stop="moveStaffByArrow(index, -1)"
              >‹</text>
              <text class="staff-name">{{ staff.name }}</text>
              <text
                :class="['staff-move-arrow', index === staffList.length - 1 || savingStaffOrder ? 'disabled' : '']"
                @click.stop="moveStaffByArrow(index, 1)"
              >›</text>
            </view>
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
                <view class="appt-head">
                  <text class="appt-time">{{ appt.start_time }}-{{ appt.end_time }}</text>
                  <view class="appt-status-mini" :style="getAppointmentStatusBadgeStyle(appt.status)">{{ getAppointmentStatusLabel(appt.status) }}</view>
                </view>
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
                    <view class="appt-pet-note compact" v-if="petInfo.noteText">{{ petInfo.noteText }}</view>
                  </view>
                </view>
                <view class="appt-notes compact" v-if="getAppointmentGeneralNoteSummary(appt)">{{ getAppointmentGeneralNoteSummary(appt) }}</view>
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

    <!-- 日历选择器 -->
    <CalendarPicker
      :visible="showCalPicker"
      :value="currentDate"
      :appointment-dates="calendarMarkedDates"
      :full-dates="calendarFullDates"
      @select="onCalSelect"
      @toggle-full="onCalendarToggleFull"
      @month-change="onCalendarMonthChange"
      @close="showCalPicker = false"
    />
  </view>
  </SideLayout>
</template>

<script setup lang="ts">
import SideLayout from '@/components/SideLayout.vue'
import CalendarPicker from '@/components/CalendarPicker.vue'
import { ref, reactive, computed, onMounted, onBeforeUnmount } from 'vue'
import { onLoad, onShow } from '@dcloudio/uni-app'
import FilterPanel from '@/components/FilterPanel.vue'
import { deleteAppointment, getAppointmentCalendar, getAppointmentCalendarSummary, getAppointmentList, setAppointmentCalendarMark } from '@/api/appointment'
import { getDashboardOverview } from '@/api/dashboard'
import { getStaffList, getStaffSchedule, updateStaffOrder } from '@/api/staff'
import { getCategoryTree } from '@/api/service-category'
import { getShop } from '@/api/shop'
import { getPersonalityBg, getPersonalityColor } from '@/utils/personality'
import { compareStaffRole, hasStaffRoleAtLeast } from '@/utils/staff-role'
import { useDesktopInteraction } from '@/utils/interaction'
import { useAuthStore } from '@/store/auth'
import {
  getAppointmentStatusMeta,
  APPOINTMENT_STATUS_META,
  getAppointmentStatusBadgeStyle,
  getAppointmentStatusBlockStyle,
  getAppointmentStatusLabel,
} from '@/utils/appointment-status'
import { buildAppointmentCalendarSummary } from '@/utils/appointment-calendar-summary'
import { sanitizeAppointmentNotes } from '@/utils/appointment-notes'

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
const { isDesktopInteraction } = useDesktopInteraction()
const authStore = useAuthStore()
const canReorderStaff = computed(() => hasStaffRoleAtLeast(authStore.staffInfo?.role, 'manager'))
const draggingStaffId = ref<number | null>(null)
const savingStaffOrder = ref(false)
let staffDragMoved = false
let staffDragSnapshot: Staff[] = []

// 待处理面板
const showPendingPanel = ref(false)
const showPendingFilter = ref(false)
const pendingCount = ref(0)
const pendingList = ref<any[]>([])
const pendingLoading = ref(false)
const pendingStaffList = ref<any[]>([])
const pendingCategories = ref<any[]>([])
const pendingSelectionMode = ref(false)
const pendingSelectedIds = ref<number[]>([])
const pendingConfirmDelete = ref(false)
let pendingLongPressTimer: ReturnType<typeof setTimeout> | null = null
let pendingLongPressTriggered = false
const pendingFilter = reactive({
  status: -1, dateFrom: '', dateTo: '', staffId: 0, payMethod: '', categoryId: 0, productKeyword: '',
})
const pendingStatusOptions = [
  { value: 0, label: '待确认' },
  { value: 1, label: '已确认' },
  { value: 2, label: '服务中' },
  { value: 3, label: '待结算' },
  { value: 7, label: '已开单' },
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
  } else {
    exitPendingSelectionMode()
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
function isPendingSelected(id: number) {
  return pendingSelectedIds.value.includes(id)
}
function togglePendingSelected(id: number) {
  if (isPendingSelected(id)) {
    pendingSelectedIds.value = pendingSelectedIds.value.filter((item) => item !== id)
    return
  }
  pendingSelectedIds.value = [...pendingSelectedIds.value, id]
}
function enterPendingSelectionMode(id?: number) {
  pendingSelectionMode.value = true
  if (typeof id === 'number' && !isPendingSelected(id)) {
    pendingSelectedIds.value = [...pendingSelectedIds.value, id]
  }
}
function exitPendingSelectionMode() {
  pendingSelectionMode.value = false
  pendingSelectedIds.value = []
  pendingConfirmDelete.value = false
}
function startPendingLongPress(id: number) {
  clearPendingLongPress()
  pendingLongPressTriggered = false
  pendingLongPressTimer = setTimeout(() => {
    pendingLongPressTriggered = true
    enterPendingSelectionMode(id)
  }, 450)
}
function clearPendingLongPress() {
  if (pendingLongPressTimer) {
    clearTimeout(pendingLongPressTimer)
    pendingLongPressTimer = null
  }
}
function handlePendingCardClick(id: number) {
  if (pendingLongPressTriggered) {
    pendingLongPressTriggered = false
    return
  }
  clearPendingLongPress()
  if (pendingSelectionMode.value) {
    togglePendingSelected(id)
    if (pendingSelectedIds.value.length === 0) {
      exitPendingSelectionMode()
    }
    return
  }
  showPendingPanel.value = false
  goDetail(id)
}
function confirmDeletePendingSelected() {
  if (pendingSelectedIds.value.length === 0) return
  pendingConfirmDelete.value = true
}

function cancelPendingDelete() {
  pendingConfirmDelete.value = false
}

async function executePendingDelete() {
  pendingConfirmDelete.value = false
  let successCount = 0
  const failedIds: number[] = []
  const failedMessages: string[] = []
  const ids = [...pendingSelectedIds.value]
  for (const id of ids) {
    try {
      await deleteAppointment(id)
      successCount++
    } catch (error: any) {
      failedIds.push(id)
      const msg = error?.msg || error?.message || '删除失败'
      failedMessages.push(`#${id} ${msg}`)
    }
  }

  if (successCount > 0) {
    uni.showToast({ title: `已删除 ${successCount} 条`, icon: 'none' })
  }
  if (failedMessages.length > 0) {
    uni.showToast({ title: failedMessages.slice(0, 3).join('; '), icon: 'none', duration: 3000 })
  }

  if (successCount > 0) {
    await Promise.all([loadPending(), loadPendingCount(), loadData()])
  }

  if (failedIds.length > 0) {
    pendingSelectionMode.value = true
    pendingSelectedIds.value = failedIds
    return
  }

  exitPendingSelectionMode()
}
const staffScheduleMap = ref<Record<number, { start: string; end: string; dayOff: boolean }>>({})
const shopOpenTime = ref('10:00')
const shopCloseTime = ref('22:00')
const showDatePicker = ref(false)
const showCalPicker = ref(false)
const calendarMarkedDates = ref<string[]>([])
const calendarFullDates = ref<string[]>([])
const quickDateOptions = computed(() => [
  { label: '今天', value: offsetDateStr(0) },
  { label: '明天', value: offsetDateStr(1) },
  { label: '后天', value: offsetDateStr(2) },
])
const calendarSummary = computed(() => buildAppointmentCalendarSummary(appointments.value))
const calendarHeroSubtitle = computed(() => {
  const parts: string[] = []
  if (calendarSummary.value.pendingConfirm > 0) parts.push(`${calendarSummary.value.pendingConfirm}个待确认`)
  if (calendarSummary.value.unassigned > 0) parts.push(`${calendarSummary.value.unassigned}个待分配`)
  if (calendarSummary.value.waitingCheckout > 0) parts.push(`${calendarSummary.value.waitingCheckout}个待结算`)
  if (parts.length > 0) return parts.join(' · ')
  if (calendarSummary.value.total > 0) return '今天预约节奏平稳'
  return '今天先从新增预约开始'
})
const calendarSummaryCards = computed(() => [
  { key: 'total', label: '今日预约', value: calendarSummary.value.total, tone: 'primary' },
  { key: 'pending-confirm', label: '待确认', value: calendarSummary.value.pendingConfirm, tone: calendarSummary.value.pendingConfirm > 0 ? 'warning' : 'neutral' },
  { key: 'unassigned', label: '待分配', value: calendarSummary.value.unassigned, tone: calendarSummary.value.unassigned > 0 ? 'warning' : 'neutral' },
  { key: 'waiting-checkout', label: '待结算', value: calendarSummary.value.waitingCheckout, tone: calendarSummary.value.waitingCheckout > 0 ? 'warning' : 'neutral' },
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
  return (116 + cols * 340) + 'rpx'
})

// Unassigned appointments (no staff_id)
const unassignedAppts = computed(() =>
  appointments.value.filter(a => !a.staff_id && a.status !== 4)
)
const unassignedCount = computed(() => unassignedAppts.value.length)

onLoad((query) => {
  const date = typeof query?.date === 'string' ? query.date.trim() : ''
  if (date) currentDate.value = date
})

onBeforeUnmount(() => {
  removeStaffDragListeners()
})

function sortStaffList(list: Staff[]) {
  return [...list].sort((a, b) => {
    const aSort = Number(a.sort_order || 0)
    const bSort = Number(b.sort_order || 0)
    const aHasSort = aSort > 0
    const bHasSort = bSort > 0
    if (aHasSort && bHasSort && aSort !== bSort) return aSort - bSort
    if (aHasSort !== bHasSort) return aHasSort ? -1 : 1
    const roleDiff = compareStaffRole(a.role, b.role)
    if (roleDiff !== 0) return roleDiff
    return a.ID - b.ID
  })
}

function moveStaffColumn(list: Staff[], fromIndex: number, toIndex: number) {
  const next = [...list]
  const [moved] = next.splice(fromIndex, 1)
  next.splice(toIndex, 0, moved)
  return next
}

function getEventPoint(event: any) {
  const touch = event?.touches?.[0] || event?.changedTouches?.[0]
  if (touch) {
    return { x: touch.clientX, y: touch.clientY }
  }
  if (typeof event?.clientX === 'number' && typeof event?.clientY === 'number') {
    return { x: event.clientX, y: event.clientY }
  }
  return null
}

function removeStaffDragListeners() {
  if (typeof window === 'undefined') return
  window.removeEventListener('touchmove', handleStaffDragMove as EventListener)
  window.removeEventListener('touchend', handleStaffDragEnd as EventListener)
  window.removeEventListener('touchcancel', handleStaffDragEnd as EventListener)
  window.removeEventListener('mousemove', handleStaffDragMove as EventListener)
  window.removeEventListener('mouseup', handleStaffDragEnd as EventListener)
  document.body.style.userSelect = ''
}

function beginStaffDrag(index: number, event: any) {
  if (!canReorderStaff.value || savingStaffOrder.value || typeof window === 'undefined') return
  if (!staffList.value[index] || staffList.value.length < 2) return
  draggingStaffId.value = staffList.value[index].ID
  staffDragSnapshot = [...staffList.value]
  staffDragMoved = false
  document.body.style.userSelect = 'none'
  window.addEventListener('touchmove', handleStaffDragMove as EventListener, { passive: false })
  window.addEventListener('touchend', handleStaffDragEnd as EventListener)
  window.addEventListener('touchcancel', handleStaffDragEnd as EventListener)
  window.addEventListener('mousemove', handleStaffDragMove as EventListener)
  window.addEventListener('mouseup', handleStaffDragEnd as EventListener)
  handleStaffDragMove(event)
}

function handleStaffDragMove(event: Event) {
  if (draggingStaffId.value == null || typeof document === 'undefined') return
  const point = getEventPoint(event)
  if (!point) return
  if ('preventDefault' in event) {
    event.preventDefault()
  }
  const element = document.elementFromPoint(point.x, point.y) as HTMLElement | null
  const headerCell = element?.closest('.staff-header-cell') as HTMLElement | null
  const targetId = Number(headerCell?.dataset?.staffId || 0)
  if (!targetId || targetId === draggingStaffId.value) return
  const fromIndex = staffList.value.findIndex((staff) => staff.ID === draggingStaffId.value)
  const targetIndex = staffList.value.findIndex((staff) => staff.ID === targetId)
  if (fromIndex < 0 || targetIndex < 0 || fromIndex === targetIndex) return
  staffList.value = moveStaffColumn(staffList.value, fromIndex, targetIndex)
  staffDragMoved = true
}

async function handleStaffDragEnd() {
  const activeId = draggingStaffId.value
  removeStaffDragListeners()
  draggingStaffId.value = null
  if (!activeId || !staffDragMoved) return
  await persistStaffOrder(staffDragSnapshot)
}

async function persistStaffOrder(snapshot: Staff[]) {
  savingStaffOrder.value = true
  try {
    await updateStaffOrder(staffList.value.map((staff) => staff.ID))
    staffList.value = staffList.value.map((staff, index) => ({
      ...staff,
      sort_order: index + 1,
    }))
    uni.showToast({ title: '员工顺序已更新', icon: 'success' })
  } catch {
    staffList.value = [...snapshot]
    uni.showToast({ title: '保存顺序失败', icon: 'none' })
  } finally {
    savingStaffOrder.value = false
    staffDragSnapshot = []
    staffDragMoved = false
  }
}

async function moveStaffByArrow(index: number, step: -1 | 1) {
  if (!canReorderStaff.value || savingStaffOrder.value) return
  const targetIndex = index + step
  if (targetIndex < 0 || targetIndex >= staffList.value.length) return
  const snapshot = [...staffList.value]
  staffList.value = moveStaffColumn(staffList.value, index, targetIndex)
  staffDragSnapshot = snapshot
  staffDragMoved = true
  await persistStaffOrder(snapshot)
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

function onCalSelect(date: string) {
  showCalPicker.value = false
  setCurrentDate(date)
}

async function onCalendarMonthChange(payload: { startDate: string; endDate: string }) {
  try {
    const res = await getAppointmentCalendarSummary(payload.startDate, payload.endDate)
    const items = Array.isArray(res.data) ? res.data : []
    calendarMarkedDates.value = items.filter((item: any) => item?.has_appointments && item?.date).map((item: any) => item.date)
    calendarFullDates.value = items.filter((item: any) => item?.is_full && item?.date).map((item: any) => item.date)
  } catch {
    calendarMarkedDates.value = []
    calendarFullDates.value = []
  }
}

async function onCalendarToggleFull(payload: { date: string; marked: boolean }) {
  try {
    await setAppointmentCalendarMark(payload.date, payload.marked)
    const next = new Set(calendarFullDates.value)
    if (payload.marked) {
      next.add(payload.date)
    } else {
      next.delete(payload.date)
    }
    calendarFullDates.value = Array.from(next).sort()
    uni.showToast({ title: payload.marked ? '已标红' : '已取消标红', icon: 'none' })
  } catch (error: any) {
    uni.showToast({ title: error?.msg || error?.message || '操作失败', icon: 'none' })
  }
}

const weekDayNames = ['周日', '周一', '周二', '周三', '周四', '周五', '周六']
function formatDateDisplay(dateStr: string) {
  const d = new Date(dateStr + 'T12:00:00')
  const m = d.getMonth() + 1
  const day = d.getDate()
  const wd = weekDayNames[d.getDay()]
  return `${m}月${day}日 ${wd}`
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
    if (pet.noteText) h += 18 + estimateWrappedLineCount(pet.noteText) * 28
    h += 16
  }
  const noteSummary = getAppointmentGeneralNoteSummary(appt)
  if (noteSummary) {
    h += 18 + estimateNoteLineCount(noteSummary) * 28
  }
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
  // 补偿跨行的 border-bottom（每行 1rpx）+ 2rpx 冗余确保覆盖完整
  totalH += slots
  const heightVw = totalH / 750 * 100
  const statusMeta = getAppointmentStatusMeta(appt?.status)
  return {
    ...getAppointmentStatusBlockStyle(appt?.status),
    background: 'linear-gradient(180deg, rgba(255,255,255,0.98) 0%, rgba(255,255,255,0.96) 100%)',
    '--appt-accent': statusMeta.blockAccent,
    '--appt-accent-soft': statusMeta.badgeBg,
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
    return [{ id: 'empty', name: '-', meta: '未填写宠物信息', tags: [] as Array<{ text: string; className: string; style?: string }>, services: [] as string[], noteText: '' }]
  }
  const petNoteMap = getAppointmentPetNoteMap(appt)
  return petItems.map((petItem: any, index: number) => {
    const pet = petItem?.pet
    const serviceNames = (petItem.services || [])
      .map((s: any) => s.service_name)
      .filter(Boolean)
    const petName = pet?.name || `猫咪${index + 1}`
    return {
      id: pet?.ID || index,
      name: petName,
      meta: pet ? formatPetMeta(pet) : '未填写宠物信息',
      tags: pet ? getPetTagItems(pet) : [],
      services: serviceNames,
      noteText: petNoteMap[petName] || '',
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

function normalizeNoteText(text?: string): string {
  return sanitizeAppointmentNotes(text)
    .replace(/\r\n/g, '\n')
    .split('\n')
    .map(line => line.trim())
    .filter(Boolean)
    .join('\n')
}

function getNoteLines(text?: string): string[] {
  return normalizeNoteText(text)
    .split('\n')
    .map(line => line.trim())
    .filter(Boolean)
}

function splitNoteSegments(text: string): string[] {
  return String(text || '')
    .split(/[；;]+/)
    .map(segment => segment.trim())
    .filter(Boolean)
}

function getAppointmentPetNoteMap(appt: any): Record<string, string> {
  const petNames = getAppointmentPets(appt)
    .map((petItem: any, index: number) => petItem?.pet?.name || `猫咪${index + 1}`)
    .filter(Boolean)
    .sort((a: string, b: string) => b.length - a.length)

  const noteMap: Record<string, string[]> = {}
  for (const line of getNoteLines(appt?.notes)) {
    const matchedName = petNames.find((name: string) => line.startsWith(`${name}：`) || line.startsWith(`${name}:`))
    if (!matchedName) continue
    const content = line.slice(matchedName.length + 1).trim()
    if (!content) continue
    if (!noteMap[matchedName]) noteMap[matchedName] = []
    noteMap[matchedName].push(...splitNoteSegments(content))
  }

  return Object.fromEntries(Object.entries(noteMap).map(([name, lines]) => [name, lines.join('\n')]))
}

function estimateWrappedLineCount(line: string, maxCharsPerLine = 14): number {
  const text = line.trim()
  if (!text) return 0
  return Math.max(1, Math.ceil(Array.from(text).length / maxCharsPerLine))
}

function estimateNoteLineCount(text: string): number {
  return text
    .split('\n')
    .reduce((total, line) => total + estimateWrappedLineCount(line), 0)
}

function getAppointmentNoteSummary(appt: any): string {
  const parts: string[] = []
  const customerNotes = normalizeNoteText(appt?.notes)
  const staffNotes = normalizeNoteText(appt?.staff_notes)
  if (customerNotes) parts.push(customerNotes)
  if (staffNotes) parts.push(`洗护师：${staffNotes}`)
  return parts.join('\n')
}

function getAppointmentGeneralNoteSummary(appt: any): string {
  const petNames = getAppointmentPets(appt)
    .map((petItem: any, index: number) => petItem?.pet?.name || `猫咪${index + 1}`)
    .filter(Boolean)
    .sort((a: string, b: string) => b.length - a.length)

  const customerNotes = getNoteLines(appt?.notes)
    .filter(line => !petNames.some((name: string) => line.startsWith(`${name}：`) || line.startsWith(`${name}:`)))
    .join('\n')

  const parts: string[] = []
  if (customerNotes) parts.push(customerNotes)
  const staffNotes = normalizeNoteText(appt?.staff_notes)
  if (staffNotes) parts.push(`洗护师：${staffNotes}`)
  return parts.join('\n')
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
.page { display: flex; flex-direction: column; height: 100vh; background: var(--cat-color-page-bg); }
.calendar-hero { padding: 20rpx 24rpx 16rpx; background: linear-gradient(180deg, #fff7ea 0%, var(--cat-color-page-bg) 100%); }
.calendar-hero-head { display: flex; align-items: flex-start; justify-content: space-between; gap: 16rpx; margin-bottom: 16rpx; }
.calendar-hero-copy { display: flex; flex-direction: column; gap: 6rpx; min-width: 0; }
.calendar-hero-title { font-size: 38rpx; font-weight: 700; color: var(--cat-color-text-main); }
.calendar-hero-subtitle { font-size: 24rpx; color: var(--cat-color-text-muted); line-height: 1.5; }
.date-nav { display: flex; align-items: center; gap: 12rpx; padding: 20rpx; background: var(--cat-color-card-bg); border: 1rpx solid var(--cat-color-border); border-radius: 24rpx; box-shadow: var(--cat-shadow-soft); flex-wrap: wrap; }
.nav-btn { width: 80rpx; height: 80rpx; min-width: 80rpx; min-height: 80rpx; display: flex; align-items: center; justify-content: center; background: var(--cat-color-primary-soft); border-radius: 50%; font-size: 28rpx; color: var(--cat-color-primary-deep); transition: background-color 0.15s, transform 0.1s; }
.nav-btn:active { background: var(--cat-color-primary-soft-strong); transform: scale(0.9); }
.date-display { flex: 1; min-width: 220rpx; text-align: center; font-size: 30rpx; font-weight: 600; color: var(--cat-color-text-main); }
.quick-date-group { display: flex; gap: 10rpx; margin-left: auto; flex-wrap: wrap; justify-content: flex-end; }
.quick-date-btn { min-width: 84rpx; height: 60rpx; padding: 0 18rpx; display: flex; align-items: center; justify-content: center; border-radius: 999rpx; font-size: 24rpx; color: var(--cat-color-text-muted); background: #f7efe0; border: 1rpx solid transparent; }
.quick-date-btn.active { color: var(--cat-color-primary-deep); background: var(--cat-color-primary-soft); border-color: rgba(160, 120, 48, 0.18); font-weight: 600; }

.summary-strip { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 12rpx; margin-top: 14rpx; }
.summary-card { background: var(--cat-color-card-bg); border: 1rpx solid var(--cat-color-border); border-radius: 20rpx; padding: 18rpx 20rpx; box-shadow: var(--cat-shadow-soft); display: flex; flex-direction: column; gap: 8rpx; }
.summary-card-primary { background: linear-gradient(135deg, #fff7ea 0%, var(--cat-color-card-bg) 100%); }
.summary-card-warning { background: linear-gradient(135deg, var(--cat-color-primary-soft) 0%, #fff6dd 100%); }
.summary-card-neutral { background: #fffdf8; }
.summary-card-label { font-size: 22rpx; color: var(--cat-color-text-muted); }
.summary-card-value { font-size: 40rpx; font-weight: 700; color: var(--cat-color-text-main); line-height: 1; }

.unassigned-section { padding: 16rpx 24rpx; }
.section-label { font-size: 26rpx; font-weight: 600; color: var(--cat-color-primary-deep); display: block; margin-bottom: 12rpx; }
.unassigned-list { display: flex; flex-wrap: wrap; gap: 12rpx; }
.unassigned-list .appt-card { flex: 1; min-width: 420rpx; max-width: 100%; }
.appt-card { background: var(--cat-color-card-bg); border-radius: 20rpx; padding: 14rpx 18rpx; border: 1rpx solid var(--cat-color-border); box-shadow: var(--cat-shadow-soft); }
.appt-card.unassigned { border-left: 8rpx solid transparent; }
.appt-card-top { display: flex; justify-content: space-between; align-items: center; margin-bottom: 8rpx; }
.appt-time-text { font-size: 26rpx; font-weight: 600; color: var(--cat-color-text-main); }
.status-tag { font-size: 20rpx; padding: 4rpx 12rpx; border-radius: 999rpx; }
.appt-pet-name { font-size: 28rpx; font-weight: 600; color: var(--cat-color-text-main); display: block; }
.appt-pet-services { display: flex; flex-wrap: wrap; gap: 6rpx; margin-top: 6rpx; }
.appt-pet-services.compact { gap: 4rpx; margin-top: 4rpx; }
.appt-pet-service-item { font-size: 22rpx; color: var(--cat-color-primary-deep); background: var(--cat-color-primary-soft); padding: 4rpx 14rpx; border-radius: 8rpx; line-height: 1.4; }
.appt-pet-service-item.compact { font-size: 18rpx; padding: 2rpx 10rpx; }
.appt-pet-note {
  display: block;
  width: 100%;
  margin-top: 6rpx;
  font-size: 20rpx;
  line-height: 1.45;
  color: #475569;
  white-space: pre-wrap;
  word-break: break-word;
  overflow-wrap: anywhere;
}
.appt-pet-note.compact {
  margin-top: 4rpx;
  font-size: 18rpx;
  line-height: 1.4;
}
.appt-pet-stack { display: flex; flex-direction: column; gap: 8rpx; }
.appt-pet-stack.compact { gap: 6rpx; }
.appt-pet-entry { padding: 8rpx 0; border-bottom: 1rpx dashed rgba(148, 163, 184, 0.35); }
.appt-pet-entry:first-child { padding-top: 0; }
.appt-pet-entry:last-child { padding-bottom: 0; border-bottom: 0; }
.appt-pet-entry.compact { padding: 6rpx 0; }
.appt-notes {
  display: block;
  width: 100%;
  margin-top: 10rpx;
  color: #475569;
  font-size: 20rpx;
  line-height: 1.45;
  white-space: pre-wrap;
  word-break: break-word;
  overflow-wrap: anywhere;
}
.appt-notes.compact {
  margin-top: 8rpx;
  font-size: 18rpx;
  line-height: 1.4;
}
.appt-tag-row { display: flex; flex-wrap: wrap; gap: 8rpx; margin-top: 6rpx; }
.appt-tag-row.compact { gap: 6rpx; margin-top: 4rpx; }

.calendar-scroll { flex: 1; }
.calendar-grid { min-width: 100%; padding-bottom: 32rpx; }
.grid-header { display: flex; position: sticky; top: 0; z-index: 10; background: rgba(255, 250, 243, 0.96); border-bottom: 2rpx solid var(--cat-color-border); box-shadow: 0 8rpx 18rpx rgba(116, 88, 38, 0.08); }
.header-cell { padding: 18rpx 8rpx; font-size: 26rpx; font-weight: 600; color: var(--cat-color-text-main); text-align: center; background: rgba(255, 247, 234, 0.92); }
.staff-header-cell {
  transition: background 0.2s ease, transform 0.2s ease, box-shadow 0.2s ease, opacity 0.2s ease;
}
.staff-header-cell.reorderable {
  cursor: grab;
}
.staff-header-cell.dragging {
  background: var(--cat-color-primary-soft);
  box-shadow: inset 0 0 0 2rpx var(--cat-color-primary);
  opacity: 0.86;
  transform: scale(0.98);
}
.staff-header-cell.saving {
  pointer-events: none;
}
.staff-name-row {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 10rpx;
  margin-bottom: 4rpx;
}
.staff-name {
  display: block;
  max-width: 180rpx;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.staff-move-arrow {
  width: 34rpx;
  height: 34rpx;
  border-radius: 999rpx;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  background: var(--cat-color-primary-soft);
  color: var(--cat-color-primary-deep);
  font-size: 22rpx;
  line-height: 1;
  flex-shrink: 0;
}
.staff-move-arrow.disabled {
  opacity: 0.35;
}
.staff-count { font-size: 20rpx; color: var(--cat-color-text-muted); font-weight: 500; }
.staff-schedule-label { font-size: 18rpx; color: var(--cat-color-text-light); display: block; font-weight: 400; }
.time-col {
  width: 116rpx;
  min-width: 116rpx;
  background: #fbf4e8;
  display: flex;
  align-items: center;
  justify-content: center;
}
.staff-col { width: 260rpx; min-width: 260rpx; border-left: 1rpx solid var(--cat-color-border); }

.grid-body { position: relative; background: linear-gradient(180deg, #fffdf8 0%, #f8f1e5 100%); overflow: visible; }
.time-row { display: flex; min-height: 80rpx; border-bottom: 1rpx solid var(--cat-color-border); background: #fffdf8; position: relative; }
.time-row:nth-child(odd) { background: #fff8ef; }
.time-label {
  font-size: 22rpx;
  color: var(--cat-color-text-muted);
  padding: 0 8rpx;
  min-height: 80rpx;
  line-height: 1;
  font-weight: 600;
  background: rgba(255, 247, 234, 0.92);
  text-align: center;
  box-sizing: border-box;
}
.time-cell { position: relative; min-height: 80rpx; padding: 8rpx; background: rgba(255, 253, 248, 0.78); overflow: visible; box-sizing: border-box; }
.time-cell.off-duty { background: repeating-linear-gradient(135deg, #F1F5F9, #F1F5F9 8rpx, #E8ECF1 8rpx, #E8ECF1 16rpx); opacity: 0.5; }

.appt-block {
  position: absolute;
  top: 0;
  left: 6rpx;
  right: 6rpx;
  border: 1rpx solid rgba(160, 120, 48, 0.14);
  border-left: 10rpx solid var(--appt-accent, var(--cat-color-primary-deep));
  border-radius: 16rpx;
  padding: 12rpx 12rpx 10rpx;
  overflow: visible;
  cursor: pointer;
  box-shadow: 0 14rpx 28rpx rgba(116, 88, 38, 0.14);
  z-index: 5;
  box-sizing: border-box;
}
.appt-block::before {
  content: '';
  position: absolute;
  left: 0;
  right: 0;
  top: 0;
  height: 8rpx;
  border-radius: 16rpx 16rpx 0 0;
  background: linear-gradient(90deg, var(--appt-accent, #4338CA) 0%, var(--appt-accent-soft, #EEF2FF) 100%);
  opacity: 0.95;
}
.appt-block:hover {
  z-index: 20;
}
.appt-head { display: flex; align-items: center; justify-content: space-between; gap: 8rpx; margin-bottom: 6rpx; }
.appt-status-mini { flex-shrink: 0; font-size: 16rpx; line-height: 1.1; padding: 3rpx 10rpx; border-radius: 999rpx; font-weight: 700; box-shadow: inset 0 0 0 1rpx rgba(255,255,255,0.45); }
.appt-time { font-size: 20rpx; color: var(--cat-color-text-main); display: block; font-weight: 800; margin-bottom: 0; letter-spacing: 0.2rpx; }
.appt-pet { font-size: 24rpx; font-weight: 800; color: var(--cat-color-text-main); display: block; }
.appt-pet-info { font-size: 18rpx; color: #5f4f3e; display: block; line-height: 1.45; font-weight: 500; }
.appt-tag { display: inline-flex; align-items: center; padding: 4rpx 12rpx; border-radius: 999rpx; font-size: 18rpx; line-height: 1.2; background: #F3F4F6; color: #4B5563; }
.appt-tag.compact { padding: 2rpx 10rpx; font-size: 16rpx; }
.appt-tag.tag-fur { background: var(--cat-color-primary-soft); color: var(--cat-color-primary-deep); }
.appt-tag.tag-neutered { background: #ECFDF5; color: #047857; }
.appt-tag.tag-aggression { background: #FEF2F2; color: #DC2626; }
.appt-new-tag { font-size: 18rpx; color: var(--cat-color-primary-deep); font-weight: 600; }
.appt-new-tag-lg { font-size: 20rpx; color: var(--cat-color-text-main); background: var(--cat-color-primary); padding: 2rpx 12rpx; border-radius: 8rpx; margin-left: 8rpx; }
.appt-pet-row { display: flex; align-items: center; }
.appt-pet-detail { font-size: 22rpx; color: #64748B; display: block; margin-top: 4rpx; }
.appt-pet-service-item.compact {
  font-size: 18rpx;
  padding: 4rpx 10rpx;
  color: #0F172A;
  background: var(--appt-accent-soft, #EEF2FF);
  border: 1rpx solid rgba(15, 23, 42, 0.08);
  font-weight: 700;
  box-shadow: inset 0 1rpx 0 rgba(255, 255, 255, 0.8);
}
.appt-notes.compact {
  margin-top: 8rpx;
  font-size: 18rpx;
  line-height: 1.4;
  color: #334155;
  font-weight: 500;
}

.empty { text-align: center; padding: 100rpx 0; color: #9CA3AF; font-size: 28rpx; }

/* Pending button */
.pending-btn { display: flex; align-items: center; gap: 6rpx; padding: 0 20rpx; height: 60rpx; border-radius: 999rpx; background: linear-gradient(135deg, var(--cat-color-primary) 0%, #efc97c 100%); color: var(--cat-color-text-main); font-size: 24rpx; font-weight: 600; border: 1rpx solid var(--cat-color-border-strong); box-shadow: 0 10rpx 24rpx rgba(160, 120, 48, 0.18); }
.pending-badge { background: #EF4444; color: #fff; font-size: 20rpx; min-width: 32rpx; height: 32rpx; border-radius: 999rpx; display: flex; align-items: center; justify-content: center; padding: 0 8rpx; font-weight: 700; }

/* Pending overlay & panel */
.pending-overlay { position: fixed; top: 0; left: 0; right: 0; bottom: 0; background: rgba(0,0,0,0.45); z-index: 1500; display: flex; align-items: flex-end; justify-content: center; }
.pending-panel { background: var(--cat-color-card-bg); border-radius: 24rpx 24rpx 0 0; width: 100%; max-height: 85vh; display: flex; flex-direction: column; padding-bottom: calc(50px + env(safe-area-inset-bottom)); }
.pending-confirm-mask { position: fixed; inset: 0; background: rgba(15, 23, 42, 0.38); z-index: 1700; display: flex; align-items: center; justify-content: center; padding: 32rpx; box-sizing: border-box; }
.pending-confirm-dialog { width: 100%; max-width: 520rpx; background: var(--cat-color-card-bg); border-radius: 24rpx; box-shadow: 0 24rpx 60rpx rgba(116, 88, 38, 0.22); padding: 28rpx; box-sizing: border-box; }
.pending-confirm-head { display: flex; align-items: center; justify-content: space-between; gap: 16rpx; margin-bottom: 18rpx; }
.pending-confirm-title { font-size: 32rpx; font-weight: 700; color: var(--cat-color-text-main); }
.pending-confirm-close { font-size: 32rpx; color: var(--cat-color-text-light); padding: 4rpx 8rpx; }
.pending-confirm-text { display: block; font-size: 26rpx; line-height: 1.7; color: var(--cat-color-text-muted); }
.pending-confirm-actions { display: flex; gap: 16rpx; margin-top: 28rpx; }
.pending-confirm-btn { flex: 1; min-height: 84rpx; display: flex; align-items: center; justify-content: center; border-radius: 18rpx; font-size: 26rpx; font-weight: 700; }
.pending-confirm-btn.secondary { background: #f7efe0; color: var(--cat-color-text-muted); }
.pending-confirm-btn.danger { background: #DC2626; color: #fff; box-shadow: 0 12rpx 24rpx rgba(220, 38, 38, 0.24); }
.pending-header { display: flex; justify-content: space-between; align-items: center; padding: 24rpx 28rpx; border-bottom: 1rpx solid var(--cat-color-border); }
.pending-header-right { display: flex; align-items: center; gap: 16rpx; }
.pending-title { font-size: 32rpx; font-weight: 700; color: var(--cat-color-text-main); }
.pending-close { font-size: 36rpx; color: var(--cat-color-text-light); padding: 8rpx; }
.pending-select-btn { font-size: 24rpx; color: var(--cat-color-primary-deep); background: var(--cat-color-primary-soft); padding: 10rpx 20rpx; border-radius: 10rpx; border: 1rpx solid var(--cat-color-border-strong); }
.pending-filter-btn { display: flex; align-items: center; gap: 6rpx; font-size: 24rpx; color: var(--cat-color-text-main); background: #f7efe0; padding: 10rpx 20rpx; border-radius: 10rpx; border: 1rpx solid var(--cat-color-border); }
.pending-delete-btn { display: flex; align-items: center; gap: 6rpx; font-size: 24rpx; color: #fff; background: #DC2626; padding: 10rpx 20rpx; border-radius: 10rpx; border: 1rpx solid #DC2626; }
.pending-delete-btn.disabled { opacity: 0.45; }
.pending-delete-count { background: rgba(255,255,255,0.2); color: #fff; font-size: 18rpx; min-width: 26rpx; height: 26rpx; border-radius: 999rpx; display: flex; align-items: center; justify-content: center; padding: 0 6rpx; }
.pending-cancel-btn { font-size: 24rpx; color: var(--cat-color-text-muted); background: #fffdf8; padding: 10rpx 20rpx; border-radius: 10rpx; border: 1rpx solid var(--cat-color-border); }
.pending-filter-badge { background: #EF4444; color: #fff; font-size: 18rpx; min-width: 26rpx; height: 26rpx; border-radius: 999rpx; display: flex; align-items: center; justify-content: center; padding: 0 6rpx; }

/* Pending quick tabs & active filters */
.pending-quick-tabs { display: flex; gap: 8rpx; padding: 16rpx 28rpx; flex-wrap: wrap; border-bottom: 1rpx solid var(--cat-color-border); }
.p-tab { font-size: 22rpx; padding: 8rpx 20rpx; border-radius: 20rpx; background: #f7efe0; color: var(--cat-color-text-muted); }
.p-tab.active { background: var(--cat-color-primary); color: var(--cat-color-text-main); }
.pending-active-filters { display: flex; flex-wrap: wrap; gap: 8rpx; padding: 12rpx 28rpx; }
.pf-tag { font-size: 22rpx; color: var(--cat-color-primary-deep); background: var(--cat-color-primary-soft); padding: 6rpx 16rpx; border-radius: 20rpx; }

/* Pending list */
.pending-list { flex: 1; overflow-y: auto; padding: 16rpx 28rpx; max-height: 55vh; }
.pending-empty { text-align: center; padding: 60rpx 0; color: #9CA3AF; font-size: 26rpx; }
.pending-card { position: relative; background: #fffdf8; border-radius: 16rpx; padding: 20rpx; margin-bottom: 16rpx; border: 1rpx solid var(--cat-color-border); border-left: 8rpx solid var(--cat-color-primary); box-shadow: var(--cat-shadow-soft); }
.pending-card.selection-mode { padding-right: 72rpx; }
.pending-card.selected { border-color: var(--cat-color-primary-deep); box-shadow: 0 0 0 2rpx rgba(160, 120, 48, 0.15); }
.pending-select-dot { position: absolute; top: 20rpx; right: 20rpx; width: 32rpx; height: 32rpx; border-radius: 999rpx; border: 2rpx solid var(--cat-color-border-strong); display: flex; align-items: center; justify-content: center; background: #fff; }
.pending-select-inner { width: 18rpx; height: 18rpx; border-radius: 999rpx; background: transparent; }
.pending-select-inner.selected { background: var(--cat-color-primary-deep); }
.pending-card-top { display: flex; justify-content: space-between; align-items: center; margin-bottom: 8rpx; }
.pending-card-date { font-size: 24rpx; color: var(--cat-color-text-muted); }
.pending-card-status { font-size: 20rpx; padding: 4rpx 12rpx; border-radius: 16rpx; }
.pending-card-body { margin-bottom: 6rpx; }
.pending-card-pet { font-size: 28rpx; font-weight: 600; color: var(--cat-color-text-main); }
.pending-card-customer { font-size: 24rpx; color: var(--cat-color-text-muted); margin-left: 12rpx; }
.pending-card-staff { font-size: 22rpx; color: var(--cat-color-text-muted); }

.fab {
  position: fixed;
  right: 32rpx;
  bottom: calc(50px + env(safe-area-inset-bottom) + 24rpx);
  width: 100rpx;
  height: 100rpx;
  border-radius: 50%;
  background: linear-gradient(135deg, var(--cat-color-primary) 0%, #efc97c 100%);
  color: #fff;
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: 0 8rpx 24rpx rgba(160, 120, 48, 0.34);
  z-index: 100;
}
.fab-icon { font-size: 56rpx; line-height: 1; font-weight: 300; transform: translateY(-2rpx); }

@media (min-width: 768px) {
  .summary-strip {
    grid-template-columns: repeat(4, minmax(0, 1fr));
  }
}
</style>
