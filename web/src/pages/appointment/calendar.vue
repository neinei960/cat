<template>
  <SideLayout>
  <view class="page">
    <!-- Date navigation -->
    <view class="date-nav">
      <view class="nav-btn" @click="prevDay">&lt;</view>
      <view class="date-display" @click="showDatePicker = true">{{ currentDate }}</view>
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
            </view>
          </view>
          <view class="appt-owner-row" v-if="getCustomerName(appt) !== '-' || getCustomerPhone(appt)">
            <text class="appt-owner-label">主人</text>
            <text class="appt-owner-name">{{ getCustomerName(appt) }}</text>
            <text class="appt-owner-phone" v-if="getCustomerPhone(appt) && getCustomerPhone(appt) !== getCustomerName(appt)">{{ getCustomerPhone(appt) }}</text>
          </view>
          <text class="appt-services">{{ getServiceSummary(appt) }}</text>
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
          <view class="time-row" v-for="slot in timeSlots" :key="slot">
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
import { ref, computed, onMounted } from 'vue'
import { onShow } from '@dcloudio/uni-app'
import { getAppointmentCalendar } from '@/api/appointment'
import { getStaffList, getStaffSchedule } from '@/api/staff'
import { getPersonalityBg, getPersonalityColor } from '@/utils/personality'
import {
  getAppointmentStatusBadgeStyle,
  getAppointmentStatusBlockStyle,
  getAppointmentStatusLabel,
} from '@/utils/appointment-status'

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
const staffScheduleMap = ref<Record<number, { start: string; end: string; dayOff: boolean }>>({})
const showDatePicker = ref(false)
const quickDateOptions = computed(() => [
  { label: '今天', value: offsetDateStr(0) },
  { label: '明天', value: offsetDateStr(1) },
  { label: '后天', value: offsetDateStr(2) },
])

const timeSlots = computed(() => {
  const activeSchedules = Object.values(staffScheduleMap.value).filter(
    (schedule) => !schedule.dayOff && !!schedule.start && !!schedule.end
  )

  let startMin = Number.POSITIVE_INFINITY
  let endMin = Number.NEGATIVE_INFINITY

  activeSchedules.forEach((schedule) => {
    startMin = Math.min(startMin, floorToHalfHour(parseTime(schedule.start)))
    endMin = Math.max(endMin, ceilToHalfHour(parseTime(schedule.end)))
  })

  if (!Number.isFinite(startMin) || !Number.isFinite(endMin) || startMin >= endMin) {
    const activeAppointments = appointments.value.filter((appt) => appt.status !== 4)
    if (activeAppointments.length === 0) return []
    activeAppointments.forEach((appt) => {
      startMin = Math.min(startMin, floorToHalfHour(parseTime(appt.start_time)))
      endMin = Math.max(endMin, ceilToHalfHour(parseTime(appt.end_time)))
    })
  }

  const slots: string[] = []
  for (let minute = startMin; minute <= endMin; minute += 30) {
    slots.push(minutesToTime(minute))
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
  return Math.max((duration / 30) * 80 - 8, 72)
}

function getAppointmentBlockStyle(appt: any) {
  return {
    ...getAppointmentStatusBlockStyle(appt?.status),
    height: `${getApptHeight(appt)}rpx`,
  }
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
  const [staffRes, apptRes] = await Promise.all([
    getStaffList({ page: 1, page_size: 100 }),
    getAppointmentCalendar(currentDate.value, currentDate.value),
  ])
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
  const pets = getAppointmentPets(appt)
    .map((petItem: any) => petItem?.pet)
    .filter(Boolean)
  if (pets.length === 0) {
    return [{ id: 'empty', name: '-', meta: '未填写宠物信息', tags: [] as Array<{ text: string; className: string; style?: string }> }]
  }
  return pets.map((pet: any, index: number) => ({
    id: pet.ID || index,
    name: pet.name || `猫咪${index + 1}`,
    meta: formatPetMeta(pet),
    tags: getPetTagItems(pet),
  }))
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

onMounted(loadData)
onShow(loadData)
</script>

<style scoped>
.page { display: flex; flex-direction: column; height: 100vh; background: #F2F5F9; }
.date-nav { display: flex; align-items: center; gap: 12rpx; padding: 16rpx 24rpx; background: #fff; box-shadow: 0 8rpx 24rpx rgba(15, 23, 42, 0.05); flex-wrap: wrap; }
.nav-btn { width: 60rpx; height: 60rpx; display: flex; align-items: center; justify-content: center; background: #EEF2F7; border-radius: 12rpx; font-size: 28rpx; color: #334155; }
.date-display { flex: 1; min-width: 220rpx; text-align: center; font-size: 30rpx; font-weight: 600; color: #0F172A; }
.quick-date-group { display: flex; gap: 10rpx; margin-left: auto; flex-wrap: wrap; justify-content: flex-end; }
.quick-date-btn { min-width: 84rpx; height: 60rpx; padding: 0 18rpx; display: flex; align-items: center; justify-content: center; border-radius: 999rpx; font-size: 24rpx; color: #475569; background: #F1F5F9; border: 1rpx solid transparent; }
.quick-date-btn.active { color: #4338CA; background: #E0E7FF; border-color: rgba(67, 56, 202, 0.18); font-weight: 600; }

.stats-bar { display: flex; gap: 24rpx; padding: 12rpx 24rpx; background: #fff; border-top: 1rpx solid #E2E8F0; }
.stats-text { font-size: 24rpx; color: #64748B; }

.unassigned-section { padding: 16rpx 24rpx; }
.section-label { font-size: 26rpx; font-weight: 600; color: #D97706; display: block; margin-bottom: 12rpx; }
.unassigned-list { display: flex; flex-direction: column; gap: 12rpx; }
.appt-card { background: #fff; border-radius: 16rpx; padding: 16rpx 20rpx; border: 1rpx solid #E2E8F0; box-shadow: 0 10rpx 24rpx rgba(15, 23, 42, 0.06); }
.appt-card.unassigned { border-left: 8rpx solid transparent; }
.appt-card-top { display: flex; justify-content: space-between; align-items: center; margin-bottom: 8rpx; }
.appt-time-text { font-size: 26rpx; font-weight: 600; color: #0F172A; }
.status-tag { font-size: 20rpx; padding: 4rpx 12rpx; border-radius: 999rpx; }
.appt-pet-name { font-size: 28rpx; font-weight: 600; color: #0F172A; display: block; }
.appt-services { font-size: 22rpx; color: #64748B; display: block; margin-top: 4rpx; line-height: 1.5; }
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
.staff-col { width: 240rpx; min-width: 240rpx; border-left: 1rpx solid #E2E8F0; }

.grid-body { position: relative; background: linear-gradient(180deg, #FFFFFF 0%, #F8FAFC 100%); overflow: visible; }
.time-row { display: flex; height: 80rpx; min-height: 80rpx; border-bottom: 1rpx solid #E2E8F0; background: rgba(255, 255, 255, 0.8); }
.time-row:nth-child(odd) { background: rgba(241, 245, 249, 0.82); }
.time-label { font-size: 22rpx; color: #64748B; padding: 8rpx 12rpx; line-height: 64rpx; font-weight: 600; background: rgba(248, 250, 252, 0.92); }
.time-cell { position: relative; height: 80rpx; padding: 8rpx; background: rgba(255, 255, 255, 0.78); overflow: visible; }
.time-cell.off-duty { background: repeating-linear-gradient(135deg, #F1F5F9, #F1F5F9 8rpx, #E8ECF1 8rpx, #E8ECF1 16rpx); opacity: 0.5; }

.appt-block {
  position: absolute;
  top: 4rpx;
  left: 6rpx;
  right: 6rpx;
  background: linear-gradient(180deg, #F8FAFF 0%, #E8EEFF 100%);
  border: 1rpx solid rgba(67, 56, 202, 0.18);
  border-left: 8rpx solid #4338CA;
  border-radius: 14rpx;
  padding: 10rpx 12rpx;
  overflow: hidden;
  cursor: pointer;
  box-shadow: 0 12rpx 24rpx rgba(67, 56, 202, 0.12);
  z-index: 3;
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
