<template>
  <SideLayout>
  <view class="page">
    <!-- Date navigation -->
    <view class="date-nav">
      <view class="nav-btn" @click="prevDay">&lt;</view>
      <view class="date-display" @click="showDatePicker = true">{{ currentDate }}</view>
      <view class="nav-btn" @click="nextDay">&gt;</view>
      <view class="nav-btn today-btn" @click="goToday">今天</view>
    </view>

    <!-- Stats bar -->
    <view class="stats-bar" v-if="appointments.length > 0">
      <text class="stats-text">共 {{ appointments.length }} 个预约</text>
      <text class="stats-text" v-if="unassignedCount > 0" style="color: #F59E0B;">{{ unassignedCount }} 个待分配</text>
    </view>

    <!-- Unassigned appointments -->
    <view class="unassigned-section" v-if="unassignedAppts.length > 0">
      <text class="section-label">待分配技师</text>
      <view class="unassigned-list">
        <view
          v-for="appt in unassignedAppts"
          :key="appt.ID"
          class="appt-card unassigned"
          @click="goDetail(appt.ID)"
        >
          <view class="appt-card-top">
            <text class="appt-time-text">{{ appt.start_time }} - {{ appt.end_time }}</text>
            <view :class="['status-tag', `s${appt.status}`]">{{ statusMap[appt.status] }}</view>
          </view>
          <text class="appt-pet-name">{{ appt.pet?.name || '-' }}</text>
          <text class="appt-customer-name">{{ appt.customer?.nickname || '-' }}</text>
          <text class="appt-services">{{ (appt.services || []).map((s: any) => s.service_name).join(' + ') }}</text>
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
          </view>
        </view>

        <!-- Time rows -->
        <view class="grid-body">
          <view class="time-row" v-for="slot in timeSlots" :key="slot">
            <view class="time-col time-label">{{ slot }}</view>
            <view
              class="staff-col time-cell"
              v-for="staff in staffList"
              :key="staff.ID"
              @click="onCellClick(staff.ID, slot)"
            >
              <view
                v-for="appt in getAppointmentsAt(staff.ID, slot)"
                :key="appt.ID"
                :class="['appt-block', `status-${appt.status}`]"
                :style="{ height: getApptHeight(appt) + 'rpx' }"
                @click.stop="goDetail(appt.ID)"
              >
                <text class="appt-time">{{ appt.start_time }}</text>
                <text class="appt-pet">{{ appt.pet?.name }}</text>
                <text class="appt-customer">{{ appt.customer?.nickname }}</text>
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
    <view class="fab" @click="goCreate">+</view>
  </view>
  </SideLayout>
</template>

<script setup lang="ts">
import SideLayout from '@/components/SideLayout.vue'
import { ref, computed, onMounted } from 'vue'
import { onShow } from '@dcloudio/uni-app'
import { getAppointmentCalendar } from '@/api/appointment'
import { getStaffList } from '@/api/staff'

function localDateStr(d: Date = new Date()): string {
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`
}

const currentDate = ref(localDateStr())
const staffList = ref<Staff[]>([])
const appointments = ref<any[]>([])
const showDatePicker = ref(false)
const statusMap: Record<number, string> = { 0: '待确认', 1: '已确认', 2: '进行中', 3: '已完成', 4: '已取消', 5: '未到店' }

// Generate time slots from 08:00 to 21:00 in 30-min increments
const timeSlots = computed(() => {
  const slots: string[] = []
  for (let h = 8; h <= 21; h++) {
    slots.push(`${String(h).padStart(2, '0')}:00`)
    if (h < 21) slots.push(`${String(h).padStart(2, '0')}:30`)
  }
  return slots
})

const gridWidth = computed(() => {
  const cols = staffList.value.length || 1
  return (160 + cols * 240) + 'rpx'
})

// Unassigned appointments (no staff_id)
const unassignedAppts = computed(() =>
  appointments.value.filter(a => !a.staff_id && a.status !== 4)
)
const unassignedCount = computed(() => unassignedAppts.value.length)

function prevDay() {
  const d = new Date(currentDate.value + 'T12:00:00')
  d.setDate(d.getDate() - 1)
  currentDate.value = localDateStr(d)
  loadData()
}

function nextDay() {
  const d = new Date(currentDate.value + 'T12:00:00')
  d.setDate(d.getDate() + 1)
  currentDate.value = localDateStr(d)
  loadData()
}

function goToday() {
  currentDate.value = localDateStr()
  loadData()
}

function getAppointmentsAt(staffId: number, timeSlot: string): any[] {
  return appointments.value.filter(a =>
    a.staff_id === staffId && a.start_time === timeSlot && a.status !== 4
  )
}

function getStaffApptCount(staffId: number): number {
  return appointments.value.filter(a => a.staff_id === staffId && a.status !== 4).length
}

function getApptHeight(appt: any): number {
  const startMin = parseTime(appt.start_time)
  const endMin = parseTime(appt.end_time)
  const duration = endMin - startMin
  return Math.max((duration / 30) * 80, 80)
}

function parseTime(t: string): number {
  const [h, m] = t.split(':').map(Number)
  return h * 60 + m
}

async function loadData() {
  const [staffRes, apptRes] = await Promise.all([
    getStaffList({ page: 1, page_size: 100 }),
    getAppointmentCalendar(currentDate.value, currentDate.value),
  ])
  staffList.value = (staffRes.data.list || []).filter((s: Staff) => s.status === 1 && s.role !== 'admin')
  appointments.value = apptRes.data || []
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

onMounted(loadData)
onShow(loadData)
</script>

<style scoped>
.page { display: flex; flex-direction: column; height: 100vh; background: #F2F5F9; }
.date-nav { display: flex; align-items: center; gap: 16rpx; padding: 16rpx 24rpx; background: #fff; box-shadow: 0 8rpx 24rpx rgba(15, 23, 42, 0.05); }
.nav-btn { width: 60rpx; height: 60rpx; display: flex; align-items: center; justify-content: center; background: #EEF2F7; border-radius: 12rpx; font-size: 28rpx; color: #334155; }
.today-btn { width: auto; padding: 0 20rpx; font-size: 24rpx; color: #4338CA; background: #E0E7FF; }
.date-display { flex: 1; text-align: center; font-size: 30rpx; font-weight: 600; color: #0F172A; }

.stats-bar { display: flex; gap: 24rpx; padding: 12rpx 24rpx; background: #fff; border-top: 1rpx solid #E2E8F0; }
.stats-text { font-size: 24rpx; color: #64748B; }

.unassigned-section { padding: 16rpx 24rpx; }
.section-label { font-size: 26rpx; font-weight: 600; color: #D97706; display: block; margin-bottom: 12rpx; }
.unassigned-list { display: flex; flex-direction: column; gap: 12rpx; }
.appt-card { background: #fff; border-radius: 16rpx; padding: 16rpx 20rpx; border: 1rpx solid #E2E8F0; box-shadow: 0 10rpx 24rpx rgba(15, 23, 42, 0.06); }
.appt-card.unassigned { border-left: 8rpx solid #F59E0B; }
.appt-card-top { display: flex; justify-content: space-between; align-items: center; margin-bottom: 8rpx; }
.appt-time-text { font-size: 26rpx; font-weight: 600; color: #0F172A; }
.status-tag { font-size: 20rpx; padding: 4rpx 12rpx; border-radius: 999rpx; }
.s0 { background: #FEF3C7; color: #92400E; }
.s1 { background: #E0E7FF; color: #3730A3; }
.s2 { background: #DCFCE7; color: #166534; }
.s3 { background: #E2E8F0; color: #475569; }
.appt-pet-name { font-size: 28rpx; font-weight: 600; color: #0F172A; display: block; }
.appt-customer-name { font-size: 24rpx; color: #475569; display: block; }
.appt-services { font-size: 22rpx; color: #64748B; display: block; margin-top: 4rpx; }

.calendar-scroll { flex: 1; }
.calendar-grid { min-width: 100%; padding-bottom: 32rpx; }
.grid-header { display: flex; position: sticky; top: 0; z-index: 10; background: rgba(255, 255, 255, 0.96); border-bottom: 2rpx solid #CBD5E1; box-shadow: 0 8rpx 18rpx rgba(15, 23, 42, 0.05); }
.header-cell { padding: 18rpx 8rpx; font-size: 26rpx; font-weight: 600; color: #334155; text-align: center; background: rgba(248, 250, 252, 0.92); }
.staff-count { font-size: 20rpx; color: #64748B; font-weight: 500; }
.time-col { width: 160rpx; min-width: 160rpx; background: #F8FAFC; }
.staff-col { width: 240rpx; min-width: 240rpx; border-left: 1rpx solid #E2E8F0; }

.grid-body { position: relative; background: linear-gradient(180deg, #FFFFFF 0%, #F8FAFC 100%); }
.time-row { display: flex; min-height: 80rpx; border-bottom: 1rpx solid #E2E8F0; background: rgba(255, 255, 255, 0.8); }
.time-row:nth-child(odd) { background: rgba(241, 245, 249, 0.82); }
.time-label { font-size: 22rpx; color: #64748B; padding: 8rpx 12rpx; line-height: 64rpx; font-weight: 600; background: rgba(248, 250, 252, 0.92); }
.time-cell { position: relative; padding: 8rpx; background: rgba(255, 255, 255, 0.78); }

.appt-block {
  width: calc(100% - 12rpx);
  margin: 0 6rpx 8rpx;
  background: linear-gradient(180deg, #F8FAFF 0%, #E8EEFF 100%);
  border: 1rpx solid rgba(67, 56, 202, 0.18);
  border-left: 8rpx solid #4338CA;
  border-radius: 14rpx;
  padding: 10rpx 12rpx;
  overflow: hidden;
  cursor: pointer;
  box-shadow: 0 12rpx 24rpx rgba(67, 56, 202, 0.12);
}
.appt-block.status-0 { border-left-color: #D97706; background: linear-gradient(180deg, #FFF8E7 0%, #FDE7AF 100%); border-color: rgba(217, 119, 6, 0.2); box-shadow: 0 12rpx 24rpx rgba(217, 119, 6, 0.12); }
.appt-block.status-1 { border-left-color: #4338CA; background: linear-gradient(180deg, #F5F7FF 0%, #DEE8FF 100%); border-color: rgba(67, 56, 202, 0.2); }
.appt-block.status-2 { border-left-color: #059669; background: linear-gradient(180deg, #F0FDF4 0%, #CFFCE2 100%); border-color: rgba(5, 150, 105, 0.18); box-shadow: 0 12rpx 24rpx rgba(5, 150, 105, 0.12); }
.appt-block.status-3 { border-left-color: #475569; background: linear-gradient(180deg, #F8FAFC 0%, #E2E8F0 100%); border-color: rgba(71, 85, 105, 0.16); box-shadow: 0 10rpx 18rpx rgba(71, 85, 105, 0.1); }
.appt-block.status-4 { border-left-color: #DC2626; background: linear-gradient(180deg, #FEF2F2 0%, #FECACA 100%); border-color: rgba(220, 38, 38, 0.18); opacity: 0.72; box-shadow: none; }
.appt-time { font-size: 20rpx; color: #475569; display: block; font-weight: 600; margin-bottom: 2rpx; }
.appt-pet { font-size: 24rpx; font-weight: 700; color: #0F172A; display: block; }
.appt-customer { font-size: 20rpx; color: #475569; display: block; margin-top: 2rpx; }

.empty { text-align: center; padding: 100rpx 0; color: #9CA3AF; font-size: 28rpx; }

.fab { position: fixed; right: 32rpx; bottom: 60rpx; width: 100rpx; height: 100rpx; border-radius: 50%; background: #4F46E5; color: #fff; font-size: 48rpx; display: flex; align-items: center; justify-content: center; box-shadow: 0 8rpx 24rpx rgba(79,70,229,0.4); z-index: 100; }
</style>
