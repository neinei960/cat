<template>
  <view class="page" v-if="appt">
    <view :class="['status-bar', `s${appt.status}`]">
      <text class="status-text">{{ statusMap[appt.status] }}</text>
    </view>

    <view class="card">
      <view class="row"><text class="label">日期</text><text>{{ appt.date }}</text></view>
      <view class="row"><text class="label">时间</text><text>{{ appt.start_time }} - {{ appt.end_time }}</text></view>
      <view class="row"><text class="label">客户</text><text>{{ appt.customer?.nickname || '-' }}</text></view>
      <view class="row"><text class="label">宠物</text><text>{{ appt.pet?.name || '-' }} ({{ appt.pet?.species }} {{ appt.pet?.breed }})</text></view>
      <view class="row"><text class="label">技师</text><text>{{ appt.staff?.name || '待分配' }}</text></view>
      <view class="row"><text class="label">来源</text><text>{{ sourceMap[appt.source] }}</text></view>
      <view class="row"><text class="label">金额</text><text class="amount">¥{{ appt.total_amount }}</text></view>
    </view>

    <view class="card" v-if="appt.services && appt.services.length">
      <text class="card-title">服务项目</text>
      <view class="svc-item" v-for="s in appt.services" :key="s.ID">
        <text>{{ s.service_name }}</text>
        <text class="svc-meta">¥{{ s.price }} · {{ s.duration }}分钟</text>
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
      <button v-if="appt.status === 2" class="btn complete" @click="doAction(3)">完成服务</button>
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
import { ref } from 'vue'
import { onLoad } from '@dcloudio/uni-app'
import { getAppointment, updateAppointmentStatus, assignStaff } from '@/api/appointment'
import { getStaffList } from '@/api/staff'

const appt = ref<any>(null)
const staffList = ref<Staff[]>([])
const showAssign = ref(false)
const statusMap: Record<number, string> = { 0: '待确认', 1: '已确认', 2: '进行中', 3: '已完成', 4: '已取消', 5: '未到店' }
const sourceMap: Record<number, string> = { 1: '小程序', 2: '商家创建', 3: '电话' }

onLoad(async (query) => {
  if (query?.id) {
    const res = await getAppointment(parseInt(query.id))
    appt.value = res.data
    const sRes = await getStaffList({ page: 1, page_size: 100 })
    staffList.value = (sRes.data.list || []).filter((s: Staff) => s.status === 1)
  }
})

async function doAction(status: number) {
  await updateAppointmentStatus(appt.value.ID, { status })
  uni.showToast({ title: '操作成功', icon: 'success' })
  const res = await getAppointment(appt.value.ID)
  appt.value = res.data
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
</script>

<style scoped>
.page { padding: 24rpx; }
.status-bar { padding: 24rpx; border-radius: 16rpx; margin-bottom: 16rpx; text-align: center; }
.status-text { font-size: 32rpx; font-weight: bold; }
.s0 { background: #FEF3C7; color: #92400E; }
.s1 { background: #EEF2FF; color: #4F46E5; }
.s2 { background: #D1FAE5; color: #059669; }
.s3 { background: #F3F4F6; color: #6B7280; }
.s4 { background: #FEE2E2; color: #DC2626; }
.s5 { background: #FEE2E2; color: #DC2626; }
.card { background: #fff; border-radius: 16rpx; padding: 24rpx; margin-bottom: 16rpx; }
.card-title { font-size: 28rpx; font-weight: 600; color: #1F2937; display: block; margin-bottom: 16rpx; }
.row { display: flex; justify-content: space-between; padding: 12rpx 0; border-bottom: 1rpx solid #F3F4F6; font-size: 28rpx; }
.row:last-child { border-bottom: none; }
.label { color: #6B7280; }
.amount { color: #4F46E5; font-weight: bold; }
.svc-item { display: flex; justify-content: space-between; padding: 12rpx 0; border-bottom: 1rpx solid #F3F4F6; font-size: 26rpx; }
.svc-meta { color: #6B7280; }
.notes { font-size: 26rpx; color: #6B7280; display: block; margin-bottom: 8rpx; }
.actions { display: flex; flex-direction: column; gap: 16rpx; margin-top: 16rpx; }
.btn { border-radius: 12rpx; font-size: 30rpx; }
.confirm { background: #4F46E5; color: #fff; }
.start { background: #10B981; color: #fff; }
.complete { background: #059669; color: #fff; }
.cancel { background: #fff; color: #DC2626; border: 1rpx solid #DC2626; }
.assign { background: #EEF2FF; color: #4F46E5; }
.modal-mask { position: fixed; top: 0; left: 0; right: 0; bottom: 0; background: rgba(0,0,0,0.5); display: flex; align-items: center; justify-content: center; z-index: 999; }
.modal { background: #fff; border-radius: 16rpx; padding: 32rpx; width: 80%; max-height: 60vh; overflow-y: auto; }
.modal-title { font-size: 32rpx; font-weight: bold; margin-bottom: 24rpx; display: block; }
.option-list { display: flex; flex-direction: column; gap: 12rpx; }
.option { background: #F9FAFB; border-radius: 12rpx; padding: 20rpx 24rpx; font-size: 28rpx; }
</style>
