<template>
  <view class="page" v-if="appt">
    <view class="status-bar" :style="getAppointmentStatusBarStyle(appt.status)">
      <text class="status-text">{{ getAppointmentStatusLabel(appt.status) }}</text>
    </view>

    <view class="card">
      <view class="row"><text class="label">日期</text><text>{{ appt.date }}</text></view>
      <view class="row"><text class="label">时间</text><text>{{ appt.start_time }} - {{ appt.end_time }}</text></view>
      <view class="row"><text class="label">客户</text><text>{{ appt.customer?.nickname || appt.customer?.phone || '-' }}</text></view>
      <view class="row"><text class="label">宠物</text><text>{{ getPetSummary(appt) }}</text></view>
      <view class="row"><text class="label">技师</text><text>{{ appt.staff?.name || '待分配' }}</text></view>
      <view class="row"><text class="label">来源</text><text>{{ sourceMap[appt.source] }}</text></view>
      <view class="row"><text class="label">金额</text><text class="amount">¥{{ appt.total_amount }}</text></view>
    </view>

    <view class="card" v-if="appointmentPets.length">
      <text class="card-title">宠物与服务</text>
      <view class="pet-block" v-for="petItem in appointmentPets" :key="petItem.ID || petItem.pet_id">
        <view class="pet-header">
          <text class="pet-name">{{ getPetName(petItem) }}</text>
          <text class="pet-meta">{{ getPetMeta(petItem) }}</text>
        </view>
        <view class="svc-item" v-for="s in petItem.services || []" :key="s.ID || `${petItem.pet_id}-${s.service_id}`">
          <text>{{ s.service_name }}</text>
          <text class="svc-meta">¥{{ s.price }} · {{ s.duration }}分钟</text>
        </view>
      </view>
    </view>

    <view class="card" v-if="appt.notes || appt.staff_notes">
      <text class="card-title">备注</text>
      <text class="notes" v-if="appt.notes">客户: {{ appt.notes }}</text>
      <text class="notes" v-if="appt.staff_notes">技师: {{ appt.staff_notes }}</text>
    </view>

    <!-- Action buttons based on status -->
    <view class="actions">
      <button v-if="appt.status === 0" class="btn confirm" @click="doAction(1)">确认预约</button>
      <button v-if="appt.status === 1" class="btn start" @click="doAction(2)">开始服务</button>
      <button v-if="appt.status <= 1" class="btn edit" @click="goEdit">修改当前预约</button>
      <button v-if="appt.status === 2" class="btn complete" @click="doAction(3)">完成服务</button>
      <button v-if="appt.status === 3" class="btn billing" @click="goBatchBilling">去开单</button>
      <button v-if="appt.status <= 1" class="btn cancel" @click="doCancel">取消预约</button>
      <button v-if="appt.status <= 1 && !appt.staff" class="btn assign" @click="showAssign = true">分配技师</button>
    </view>

    <!-- Assign staff modal -->
    <view class="modal-mask" v-if="showAssign" @click="showAssign = false">
      <view class="modal" @click.stop>
        <text class="modal-title">分配技师</text>
        <view class="option-list">
          <view class="option" v-for="s in staffList" :key="s.ID" @click="doAssign(s.ID)">
            {{ s.name }}
          </view>
        </view>
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { onLoad, onShow } from '@dcloudio/uni-app'
import { getAppointment, updateAppointmentStatus, assignStaff } from '@/api/appointment'
import { getStaffList } from '@/api/staff'
import { getAppointmentStatusBarStyle, getAppointmentStatusLabel } from '@/utils/appointment-status'

const appt = ref<any>(null)
const apptId = ref(0)
const staffList = ref<Staff[]>([])
const showAssign = ref(false)
const sourceMap: Record<number, string> = { 1: '小程序', 2: '商家创建', 3: '电话' }
const appointmentPets = computed(() => {
  if (Array.isArray(appt.value?.pets) && appt.value.pets.length > 0) {
    return appt.value.pets
  }
  if (appt.value?.pet) {
    return [{
      ID: appt.value.pet.ID,
      pet_id: appt.value.pet.ID,
      pet: appt.value.pet,
      services: appt.value.services || [],
    }]
  }
  return []
})

onLoad(async (query) => {
  if (query?.id) {
    apptId.value = parseInt(query.id)
    await Promise.all([loadAppointmentDetail(), loadStaffOptions()])
  }
})

onShow(async () => {
  if (!apptId.value) return
  await loadAppointmentDetail()
})

async function loadAppointmentDetail() {
  if (!apptId.value) return
  const res = await getAppointment(apptId.value)
  appt.value = res.data
}

function sortStaffList(list: Staff[]) {
  return [...list].sort((a, b) => {
    const roleDiff = Number(a.role === 'admin') - Number(b.role === 'admin')
    if (roleDiff !== 0) return roleDiff
    return a.ID - b.ID
  })
}

async function loadStaffOptions() {
  const sRes = await getStaffList({ page: 1, page_size: 100 })
  staffList.value = sortStaffList((sRes.data.list || []).filter((s: Staff) => s.status === 1))
}

async function doAction(status: number) {
  await updateAppointmentStatus(appt.value.ID, { status })
  uni.showToast({ title: '操作成功', icon: 'success' })
  const res = await getAppointment(appt.value.ID)
  appt.value = res.data
  if (status === 3) {
    setTimeout(() => {
      uni.navigateTo({ url: `/pages/order/batch-create?appointment_id=${appt.value.ID}` })
    }, 400)
  }
}

async function doCancel() {
  uni.showModal({
    title: '确认取消', content: '确认取消该预约？',
    success: async (res) => {
      if (res.confirm) {
        await updateAppointmentStatus(appt.value.ID, { status: 4, cancelled_by: 'staff' })
        uni.showToast({ title: '已取消', icon: 'success' })
        const r = await getAppointment(appt.value.ID)
        appt.value = r.data
      }
    }
  })
}

async function doAssign(staffId: number) {
  await assignStaff(appt.value.ID, staffId)
  showAssign.value = false
  uni.showToast({ title: '已分配', icon: 'success' })
  const res = await getAppointment(appt.value.ID)
  appt.value = res.data
}

function goEdit() {
  uni.navigateTo({ url: `/pages/appointment/create?id=${appt.value.ID}` })
}

function goBatchBilling() {
  uni.navigateTo({ url: `/pages/order/batch-create?appointment_id=${appt.value.ID}` })
}

function getPetSummary(appointment: any) {
  if (appointment?.pet_summary) return appointment.pet_summary
  const names = appointmentPets.value
    .map((petItem: any) => petItem.pet?.name)
    .filter(Boolean)
  if (names.length === 0) return '-'
  if (names.length === 1) return names[0]
  return `${names[0]}等${names.length}只`
}

function getPetName(petItem: any) {
  return petItem?.pet?.name || `宠物#${petItem?.pet_id || ''}`
}

function getPetMeta(petItem: any) {
  const parts = [petItem?.pet?.species, petItem?.pet?.breed].filter(Boolean)
  return parts.length > 0 ? parts.join(' · ') : '未填写品种'
}
</script>

<style scoped>
.page { padding: 24rpx; }
.status-bar { padding: 24rpx; border-radius: 16rpx; margin-bottom: 16rpx; text-align: center; }
.status-text { font-size: 32rpx; font-weight: bold; }
.card { background: #fff; border-radius: 16rpx; padding: 24rpx; margin-bottom: 16rpx; }
.card-title { font-size: 28rpx; font-weight: 600; color: #1F2937; display: block; margin-bottom: 16rpx; }
.row { display: flex; justify-content: space-between; padding: 12rpx 0; border-bottom: 1rpx solid #F3F4F6; font-size: 28rpx; }
.row:last-child { border-bottom: none; }
.label { color: #6B7280; }
.amount { color: #4F46E5; font-weight: bold; }
.pet-block { padding: 20rpx 0; border-bottom: 1rpx solid #F3F4F6; }
.pet-block:first-of-type { padding-top: 0; }
.pet-block:last-of-type { border-bottom: none; padding-bottom: 0; }
.pet-header { display: flex; justify-content: space-between; gap: 16rpx; margin-bottom: 12rpx; }
.pet-name { font-size: 28rpx; font-weight: 700; color: #1F2937; }
.pet-meta { font-size: 22rpx; color: #6B7280; text-align: right; }
.svc-item { display: flex; justify-content: space-between; padding: 12rpx 0; border-bottom: 1rpx solid #F3F4F6; font-size: 26rpx; }
.pet-block .svc-item:last-child { border-bottom: none; padding-bottom: 0; }
.svc-meta { color: #6B7280; }
.notes { font-size: 26rpx; color: #6B7280; display: block; margin-bottom: 8rpx; }
.actions { display: flex; flex-direction: column; gap: 16rpx; margin-top: 16rpx; }
.btn { border-radius: 12rpx; font-size: 30rpx; }
.confirm { background: #4F46E5; color: #fff; }
.start { background: #10B981; color: #fff; }
.edit { background: #EEF2FF; color: #4F46E5; }
.complete { background: #059669; color: #fff; }
.billing { background: #4F46E5; color: #fff; }
.cancel { background: #fff; color: #DC2626; border: 1rpx solid #DC2626; }
.assign { background: #EEF2FF; color: #4F46E5; }
.modal-mask { position: fixed; top: 0; left: 0; right: 0; bottom: 0; background: rgba(0,0,0,0.5); display: flex; align-items: center; justify-content: center; z-index: 999; }
.modal { background: #fff; border-radius: 16rpx; padding: 32rpx; width: 80%; max-height: 60vh; overflow-y: auto; }
.modal-title { font-size: 32rpx; font-weight: bold; margin-bottom: 24rpx; display: block; }
.option-list { display: flex; flex-direction: column; gap: 12rpx; }
.option { background: #F9FAFB; border-radius: 12rpx; padding: 20rpx 24rpx; font-size: 28rpx; }
</style>
