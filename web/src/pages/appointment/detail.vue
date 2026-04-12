<template>
  <SideLayout>
  <view class="page" v-if="appt">
    <view class="status-bar" :style="getAppointmentStatusBarStyle(appt.status)">
      <text class="status-text">{{ getAppointmentStatusLabel(appt.status) }}</text>
    </view>

    <view class="card">
      <view class="row"><text class="label">日期</text><text>{{ appt.date }}</text></view>
      <view class="row"><text class="label">时间</text><text>{{ appt.start_time }} - {{ appt.end_time }}</text></view>
      <view class="row"><text class="label">客户</text><text>{{ appt.customer?.nickname || appt.customer?.phone || '-' }}</text></view>
      <view class="row"><text class="label">宠物</text><text>{{ getPetSummary(appt) }}</text></view>
      <view class="row"><text class="label">洗护师</text><text>{{ appt.staff?.name || '待分配' }}</text></view>
      <view class="row"><text class="label">来源</text><text>{{ sourceMap[appt.source] }}</text></view>
      <view class="row"><text class="label">金额</text><text class="amount">¥{{ appt.total_amount }}</text></view>
    </view>

    <view class="card" v-if="appointmentPets.length">
      <text class="card-title">宠物与服务</text>
      <view class="pet-block" v-for="petItem in appointmentPets" :key="petItem.ID || petItem.pet_id">
        <view class="pet-header">
          <text class="pet-name">{{ getPetName(petItem) }}</text>
          <text v-if="petItem.pet?.aggression && petItem.pet.aggression !== '无'" class="aggression-warn">⚡ {{ petItem.pet.aggression }}</text>
          <text class="pet-meta">{{ getPetMeta(petItem) }}</text>
        </view>
        <view class="svc-item" v-for="s in petItem.services || []" :key="s.ID || `${petItem.pet_id}-${s.service_id}`">
          <text>{{ s.service_name }}</text>
          <text class="svc-meta">¥{{ s.price }} · {{ s.duration }}分钟</text>
        </view>
      </view>
    </view>

    <view class="card" v-if="displayCustomerNotes || appt.staff_notes">
      <text class="card-title">备注</text>
      <text class="notes" v-if="displayCustomerNotes">客户: {{ displayCustomerNotes }}</text>
      <text class="notes" v-if="appt.staff_notes">洗护师: {{ appt.staff_notes }}</text>
    </view>

    <view class="card" v-if="statusLogs.length">
      <text class="card-title">状态变更记录</text>
      <view class="log-item" v-for="log in statusLogs" :key="log.ID">
        <view class="log-head">
          <text class="log-status">{{ getStatusLabel(log.from_status) }} → {{ getStatusLabel(log.to_status) }}</text>
          <text class="log-time">{{ formatLogTime(log.CreatedAt) }}</text>
        </view>
        <text class="log-operator">操作人：{{ log.operator?.name || '系统/客户' }}</text>
        <text class="log-note" v-if="log.note">备注：{{ log.note }}</text>
      </view>
    </view>

    <!-- 服务记录 -->
    <view class="card" v-if="appt.status >= 2">
      <view class="card-title-row">
        <text class="card-title" style="margin-bottom:0">服务记录</text>
        <view class="add-record-btn" @click="showRecordForm = true" v-if="appt.status === 2 || appt.status === 3 || appt.status === 7">+ 添加记录</view>
      </view>
      <view v-if="serviceRecords.length === 0" class="empty-records">暂无服务记录</view>
      <view class="record-item" v-for="rec in serviceRecords" :key="rec.ID">
        <view class="record-header">
          <text class="record-staff">{{ rec.staff?.name || '技师' }}</text>
          <text class="record-time">{{ rec.CreatedAt?.substring(0, 16) }}</text>
        </view>
        <text class="record-notes" v-if="rec.notes">{{ rec.notes }}</text>
        <view class="record-tags" v-if="rec.skin_issues || rec.fur_condition || rec.weight">
          <text class="record-tag" v-if="rec.weight">体重: {{ rec.weight }}</text>
          <text class="record-tag" v-if="rec.fur_condition">毛况: {{ rec.fur_condition }}</text>
          <text class="record-tag warn" v-if="rec.skin_issues">皮肤: {{ rec.skin_issues }}</text>
        </view>
        <view class="record-photos" v-if="rec.photos">
          <image v-for="(url, idx) in rec.photos.split(',')" :key="idx" :src="url" class="record-photo" mode="aspectFill" @click="previewPhoto(rec.photos.split(','), idx)" />
        </view>
      </view>
    </view>

    <!-- 添加服务记录弹窗 -->
    <view class="modal-mask" v-if="showRecordForm" @click="showRecordForm = false">
      <view class="modal modal-lg" @click.stop>
        <text class="modal-title">添加服务记录</text>
        <view class="form-group">
          <text class="form-label">服务记录</text>
          <textarea v-model="recordForm.notes" placeholder="使用了什么浴液、剃了哪个部位、发现什么问题..." class="form-textarea" />
        </view>
        <view class="form-row-inline">
          <view class="form-group half">
            <text class="form-label">体重</text>
            <input v-model="recordForm.weight" placeholder="如 4.5kg" class="form-input-sm" />
          </view>
          <view class="form-group half">
            <text class="form-label">毛发状况</text>
            <input v-model="recordForm.fur_condition" placeholder="如 轻微打结" class="form-input-sm" />
          </view>
        </view>
        <view class="form-group">
          <text class="form-label">皮肤问题</text>
          <input v-model="recordForm.skin_issues" placeholder="如 耳朵发红、背部掉毛" class="form-input-sm" />
        </view>
        <view class="form-group">
          <text class="form-label">照片（最多3张）</text>
          <view class="photo-upload-row">
            <image v-for="(url, idx) in recordPhotos" :key="idx" :src="url" class="photo-thumb" mode="aspectFill">
              <text class="photo-del" @click.stop="recordPhotos.splice(idx, 1)">✕</text>
            </image>
            <view class="photo-add" v-if="recordPhotos.length < 3" @click="uploadPhoto">+</view>
          </view>
        </view>
        <view class="modal-btns">
          <view class="modal-btn cancel" @click="showRecordForm = false">取消</view>
          <view class="modal-btn confirm" @click="submitRecord">保存</view>
        </view>
      </view>
    </view>

    <!-- Action buttons based on status -->
    <view class="actions-panel">
      <view class="actions-row">
        <view v-if="appt.status === 0 || appt.status === 1 || appt.status === 2 || appt.status === 6" class="action-btn confirm" @click="goBatchBilling">完成服务</view>
        <view v-if="appt.status === 3" class="action-btn billing" @click="goBatchBilling">去开单</view>
        <view v-if="appt.status <= 1 || appt.status === 6" class="action-btn edit" @click="goEdit">修改预约</view>
        <view v-if="appt.status !== 4 && appt.status !== 5 && appt.status !== 7" class="action-btn noshow" @click="doNoShow">未到店</view>
        <view v-if="appt.status <= 1 || appt.status === 6" class="action-btn cancel" @click="doCancel">取消预约</view>
      </view>
      <view v-if="appt.status <= 1 && !appt.staff" class="assign-inline" @click="showAssign = true">分配洗护师</view>
    </view>

    <!-- Assign staff modal -->
    <view class="modal-mask" v-if="showAssign" @click="showAssign = false">
      <view class="modal" @click.stop>
        <text class="modal-title">分配洗护师</text>
        <view class="option-list">
          <view class="option" v-for="s in staffList" :key="s.ID" @click="doAssign(s.ID)">
            {{ s.name }}
          </view>
        </view>
      </view>
    </view>
  </view>
  </SideLayout>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { onLoad, onShow } from '@dcloudio/uni-app'
import SideLayout from '@/components/SideLayout.vue'
import { getAppointment, getAppointmentStatusLogs, updateAppointmentStatus, assignStaff } from '@/api/appointment'
import { getStaffList } from '@/api/staff'
import { APPOINTMENT_STATUS_META, getAppointmentStatusBarStyle, getAppointmentStatusLabel } from '@/utils/appointment-status'
import { sanitizeAppointmentNotes } from '@/utils/appointment-notes'
import { request } from '@/api/request'
import { compareStaffRole } from '@/utils/staff-role'

const appt = ref<any>(null)
const apptId = ref(0)
const staffList = ref<Staff[]>([])
const showAssign = ref(false)
const statusLogs = ref<any[]>([])

// 服务记录
const serviceRecords = ref<any[]>([])
const showRecordForm = ref(false)
const recordForm = ref({ notes: '', weight: '', fur_condition: '', skin_issues: '' })
const recordPhotos = ref<string[]>([])

async function loadRecords() {
  if (!apptId.value) return
  try {
    const res = await request<any[]>({ url: `/b/service-records?appointment_id=${apptId.value}` })
    serviceRecords.value = res.data || []
  } catch {}
}

async function submitRecord() {
  const petId = appointmentPets.value[0]?.pet_id || appt.value?.pet_id || 0
  try {
    await request({ url: '/b/service-records', method: 'POST', data: {
      appointment_id: apptId.value,
      pet_id: petId,
      notes: recordForm.value.notes,
      photos: recordPhotos.value.join(','),
      skin_issues: recordForm.value.skin_issues,
      fur_condition: recordForm.value.fur_condition,
      weight: recordForm.value.weight,
    }})
    showRecordForm.value = false
    recordForm.value = { notes: '', weight: '', fur_condition: '', skin_issues: '' }
    recordPhotos.value = []
    uni.showToast({ title: '记录已保存', icon: 'success' })
    loadRecords()
  } catch {
    uni.showToast({ title: '保存失败', icon: 'none' })
  }
}

function uploadPhoto() {
  uni.chooseImage({
    count: 3 - recordPhotos.value.length,
    sizeType: ['compressed'],
    success: async (res) => {
      for (const path of res.tempFilePaths) {
        try {
          const uploadRes = await new Promise<string>((resolve, reject) => {
            uni.uploadFile({
              url: '/api/v1/b/upload',
              filePath: path,
              name: 'file',
              header: { Authorization: `Bearer ${uni.getStorageSync('token')}` },
              success: (r) => {
                const data = JSON.parse(r.data)
                resolve(data.data?.url || data.url || '')
              },
              fail: reject,
            })
          })
          if (uploadRes) recordPhotos.value.push(uploadRes)
        } catch {}
      }
    }
  })
}

function previewPhoto(urls: string[], idx: number) {
  uni.previewImage({ urls, current: urls[idx] })
}
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

const displayCustomerNotes = computed(() => sanitizeAppointmentNotes(appt.value?.notes))

onLoad(async (query) => {
  if (query?.id) {
    apptId.value = parseInt(query.id)
    await Promise.all([loadAppointmentDetail(), loadStaffOptions(), loadRecords(), loadStatusLogs()])
  }
})

onShow(async () => {
  if (!apptId.value) return
  await Promise.all([loadAppointmentDetail(), loadRecords(), loadStatusLogs()])
})

async function loadAppointmentDetail() {
  if (!apptId.value) return
  const res = await getAppointment(apptId.value)
  appt.value = res.data
}

async function loadStatusLogs() {
  if (!apptId.value) return
  try {
    const res = await getAppointmentStatusLogs(apptId.value)
    statusLogs.value = res.data || []
  } catch {
    statusLogs.value = []
  }
}

function sortStaffList(list: Staff[]) {
  return [...list].sort((a, b) => {
    const roleDiff = compareStaffRole(a.role, b.role)
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
        await Promise.all([loadAppointmentDetail(), loadStatusLogs()])
      }
    }
  })
}

async function doNoShow() {
  uni.showModal({
    title: '确认未到店',
    content: '确认将该预约标记为未到店？这表示客户本次爽约。',
    confirmColor: '#EA580C',
    success: async (res) => {
      if (res.confirm) {
        await updateAppointmentStatus(appt.value.ID, { status: 5 })
        uni.showToast({ title: '已标记未到店', icon: 'success' })
        await Promise.all([loadAppointmentDetail(), loadStatusLogs()])
      }
    },
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

function getStatusLabel(status: number) {
  return APPOINTMENT_STATUS_META[status]?.label || `状态${status}`
}

function formatLogTime(value?: string) {
  if (!value) return '-'
  return String(value).slice(0, 16).replace('T', ' ')
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
.aggression-warn { font-size: 22rpx; color: #DC2626; background: #FEE2E2; padding: 4rpx 14rpx; border-radius: 8rpx; font-weight: 600; margin-left: 8rpx; }
.pet-meta { font-size: 22rpx; color: #6B7280; text-align: right; }
.svc-item { display: flex; justify-content: space-between; padding: 12rpx 0; border-bottom: 1rpx solid #F3F4F6; font-size: 26rpx; }
.pet-block .svc-item:last-child { border-bottom: none; padding-bottom: 0; }
.svc-meta { color: #6B7280; }
.notes { font-size: 26rpx; color: #6B7280; display: block; margin-bottom: 8rpx; }
.log-item { padding: 18rpx 0; border-bottom: 1rpx solid #F3F4F6; }
.log-item:last-child { border-bottom: none; padding-bottom: 0; }
.log-head { display: flex; justify-content: space-between; gap: 16rpx; margin-bottom: 6rpx; }
.log-status { font-size: 24rpx; font-weight: 700; color: #1F2937; }
.log-time { font-size: 22rpx; color: #9CA3AF; }
.log-operator, .log-note { display: block; font-size: 24rpx; color: #6B7280; line-height: 1.5; }
.actions-panel { background: #fff; border-radius: 20rpx; padding: 20rpx; margin-top: 16rpx; box-shadow: 0 10rpx 28rpx rgba(15, 23, 42, 0.06); }
.actions-row { display: grid; grid-template-columns: repeat(4, 1fr); gap: 12rpx; align-items: stretch; }
.btn { border-radius: 16rpx; font-size: 30rpx; min-height: 96rpx; display: flex; align-items: center; justify-content: center; line-height: 1.2; box-sizing: border-box; }
.action-btn { min-height: 84rpx; padding: 0 8rpx; border-radius: 16rpx; display: flex; align-items: center; justify-content: center; text-align: center; font-size: 24rpx; font-weight: 700; line-height: 1.2; white-space: nowrap; box-sizing: border-box; }
.confirm { background: #4F46E5; color: #fff; }
.arrived { background: #A855F7; color: #fff; }
.edit { background: #EEF2FF; color: #4338CA; border: 1rpx solid rgba(79, 70, 229, 0.16); }
.complete { background: #059669; color: #fff; }
.billing { background: #4F46E5; color: #fff; }
.noshow { background: #FFF7ED; color: #C2410C; border: 1rpx solid #FDBA74; }
.cancel { background: #FEF2F2; color: #DC2626; border: 1rpx solid rgba(239, 68, 68, 0.28); }
.assign-inline { margin-top: 14rpx; min-height: 78rpx; font-size: 24rpx; font-weight: 600; background: #EEF2FF; color: #4F46E5; border: 1rpx solid rgba(79, 70, 229, 0.16); border-radius: 16rpx; display: flex; align-items: center; justify-content: center; }
.modal-mask { position: fixed; top: 0; left: 0; right: 0; bottom: 0; background: rgba(0,0,0,0.5); display: flex; align-items: center; justify-content: center; padding: 32rpx; z-index: 3200; box-sizing: border-box; }
.modal { background: #fff; border-radius: 16rpx; padding: 32rpx; width: 80%; max-height: 60vh; overflow-y: auto; box-sizing: border-box; }
.modal-title { font-size: 32rpx; font-weight: bold; margin-bottom: 24rpx; display: block; }
.option-list { display: flex; flex-direction: column; gap: 12rpx; }
.option { background: #F9FAFB; border-radius: 12rpx; padding: 20rpx 24rpx; font-size: 28rpx; }

/* Service Records */
.card-title-row { display: flex; justify-content: space-between; align-items: center; margin-bottom: 16rpx; }
.add-record-btn { font-size: 24rpx; color: #4F46E5; background: #EEF2FF; padding: 8rpx 20rpx; border-radius: 20rpx; font-weight: 600; }
.empty-records { font-size: 24rpx; color: #9CA3AF; text-align: center; padding: 24rpx; }
.record-item { padding: 16rpx 0; border-bottom: 1rpx solid #F3F4F6; }
.record-item:last-child { border-bottom: none; }
.record-header { display: flex; justify-content: space-between; margin-bottom: 8rpx; }
.record-staff { font-size: 24rpx; font-weight: 600; color: #4F46E5; }
.record-time { font-size: 22rpx; color: #9CA3AF; }
.record-notes { font-size: 26rpx; color: #374151; display: block; margin-bottom: 8rpx; line-height: 1.5; }
.record-tags { display: flex; flex-wrap: wrap; gap: 8rpx; margin-bottom: 8rpx; }
.record-tag { font-size: 22rpx; padding: 4rpx 14rpx; border-radius: 8rpx; background: #F3F4F6; color: #374151; }
.record-tag.warn { background: #FEE2E2; color: #DC2626; }
.record-photos { display: flex; gap: 12rpx; flex-wrap: wrap; }
.record-photo { width: 160rpx; height: 160rpx; border-radius: 12rpx; }

/* Record form */
.modal-lg { width: 90%; max-height: 80vh; }
.form-group { margin-bottom: 20rpx; }
.form-group.half { flex: 1; }
.form-label { font-size: 26rpx; color: #374151; display: block; margin-bottom: 8rpx; font-weight: 500; }
.form-textarea { width: 100%; height: 160rpx; font-size: 26rpx; padding: 16rpx; border: 1rpx solid #E5E7EB; border-radius: 12rpx; box-sizing: border-box; }
.form-input-sm { width: 100%; font-size: 26rpx; padding: 12rpx 16rpx; border: 1rpx solid #E5E7EB; border-radius: 10rpx; box-sizing: border-box; }
.form-row-inline { display: flex; gap: 16rpx; }
.photo-upload-row { display: flex; gap: 12rpx; flex-wrap: wrap; }
.photo-thumb { width: 140rpx; height: 140rpx; border-radius: 12rpx; position: relative; }
.photo-del { position: absolute; top: -8rpx; right: -8rpx; width: 36rpx; height: 36rpx; background: #EF4444; color: #fff; border-radius: 50%; font-size: 20rpx; display: flex; align-items: center; justify-content: center; }
.photo-add { width: 140rpx; height: 140rpx; border: 2rpx dashed #D1D5DB; border-radius: 12rpx; display: flex; align-items: center; justify-content: center; font-size: 48rpx; color: #9CA3AF; }
.modal-btns { display: flex; gap: 16rpx; margin-top: 24rpx; }
.modal-btn { flex: 1; text-align: center; padding: 18rpx; border-radius: 12rpx; font-size: 28rpx; font-weight: 600; }
.modal-btn.cancel { background: #F3F4F6; color: #374151; }
.modal-btn.confirm { background: #4F46E5; color: #fff; }

@media (max-width: 768px) {
  .modal-mask {
    align-items: flex-end;
    padding: 24rpx 20rpx calc(88rpx + env(safe-area-inset-bottom));
  }

  .actions-panel {
    padding: 16rpx;
    border-radius: 18rpx;
  }

  .btn {
    min-height: 80rpx;
    font-size: 24rpx;
  }

  .actions-row {
    gap: 10rpx;
  }

  .action-btn {
    min-height: 76rpx;
    font-size: 22rpx;
    border-radius: 14rpx;
  }

  .assign-inline {
    min-height: 72rpx;
    font-size: 22rpx;
  }

  .modal,
  .modal-lg {
    width: 100%;
    max-height: calc(100vh - 184rpx - env(safe-area-inset-bottom));
    padding: 28rpx 24rpx calc(28rpx + env(safe-area-inset-bottom));
    border-radius: 28rpx 28rpx 20rpx 20rpx;
  }

  .modal-title {
    margin-bottom: 20rpx;
    line-height: 1.3;
  }

  .form-row-inline {
    flex-direction: column;
    gap: 0;
  }

  .form-group {
    margin-bottom: 16rpx;
  }

  .form-label {
    font-size: 24rpx;
    line-height: 1.4;
  }

  .form-textarea {
    height: 140rpx;
    font-size: 24rpx;
    line-height: 1.45;
  }

  .form-input-sm {
    min-height: 72rpx;
    font-size: 24rpx;
    line-height: 1.4;
  }

  .photo-thumb,
  .photo-add {
    width: 120rpx;
    height: 120rpx;
  }

  .modal-btns {
    position: sticky;
    bottom: calc(env(safe-area-inset-bottom) * -1);
    margin: 20rpx -24rpx calc(-28rpx - env(safe-area-inset-bottom));
    padding: 18rpx 24rpx calc(18rpx + env(safe-area-inset-bottom));
    background: #fff;
    border-top: 1rpx solid #F3F4F6;
  }

  .modal-btn {
    padding: 20rpx 16rpx;
    font-size: 26rpx;
    line-height: 1.2;
  }
}
</style>
