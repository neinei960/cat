<template>
  <SideLayout>
  <view class="page">
    <view class="form">
      <view class="form-item">
        <text class="label">姓名 *</text>
        <input v-model="form.name" placeholder="请输入姓名" class="input" />
      </view>
      <view class="form-item">
        <text class="label">手机号 *</text>
        <input v-model="form.phone" type="number" placeholder="请输入手机号" class="input" :disabled="!!id && !isAdmin" />
      </view>
      <view class="form-item" v-if="isAdmin">
        <text class="label">角色</text>
        <picker :range="roleList" :range-key="'label'" :value="roleIndex" @change="onRoleChange">
          <view class="picker">{{ roleList[roleIndex].label }}</view>
        </picker>
      </view>
      <!-- 创建时设置初始密码 -->
      <view class="form-item" v-if="!id && isAdmin">
        <text class="label">初始密码</text>
        <input v-model="form.password" type="text" placeholder="留空默认 123456" class="input" />
      </view>
      <!-- 仅 admin 可见的管理字段 -->
      <view class="form-item" v-if="isAdmin">
        <text class="label">洗浴提成 (%)</text>
        <input v-model="form.commission_rate" type="digit" placeholder="0" class="input" />
      </view>
      <view class="form-item" v-if="isAdmin">
        <text class="label">商品提成 (%)</text>
        <input v-model="form.product_commission_rate" type="digit" placeholder="0" class="input" />
      </view>
      <view class="form-item" v-if="isAdmin">
        <text class="label">上门喂养提成 (%)</text>
        <input v-model="form.feeding_commission_rate" type="digit" placeholder="0" class="input" />
      </view>
      <view class="form-item" v-if="id && isAdmin">
        <text class="label">状态</text>
        <picker :range="statusList" :range-key="'label'" :value="statusIndex" @change="onStatusChange">
          <view class="picker">{{ statusList[statusIndex].label }}</view>
        </picker>
      </view>
    </view>

    <!-- 重置密码区域 (编辑模式, 仅 admin) -->
    <view class="section" v-if="id && canManageSchedule">
      <view class="section-header">
        <text class="section-title">密码管理</text>
      </view>
      <view class="reset-row">
        <input v-model="newPassword" type="text" placeholder="输入新密码（至少6位）" class="input flex1" />
        <view class="btn-reset" @click="onResetPassword">重置密码</view>
      </view>
    </view>

    <!-- 工作时间设置 (编辑模式, 仅 admin) -->
    <view class="section" v-if="id && isAdmin">
      <view class="section-header">
        <text class="section-title">工作时间</text>
        <text class="hint">设置本周排班，可逐日调整</text>
      </view>
      <view class="schedule-row" v-for="(day, idx) in weekSchedule" :key="day.date">
        <view class="day-label">
          <text class="day-name">{{ dayNames[idx] }}</text>
          <text class="day-date">{{ day.date.slice(5) }}</text>
        </view>
        <view v-if="day.is_day_off" class="day-off-tag" @click="day.is_day_off = false">休息</view>
        <view v-else class="time-inputs">
          <input v-model="day.start_time" placeholder="12:00" class="time-input" maxlength="5" />
          <text class="time-sep">-</text>
          <input v-model="day.end_time" placeholder="22:00" class="time-input" maxlength="5" />
          <input v-model="day.max_capacity" type="number" placeholder="并发" class="cap-input" maxlength="2" />
          <view class="btn-day-off" @click="day.is_day_off = true">休</view>
        </view>
      </view>
      <view class="schedule-actions">
        <view class="btn-copy" @click="copyToAll">复制第一天到全周</view>
        <view class="btn-save-schedule" @click="saveSchedule">保存排班</view>
      </view>
    </view>

    <button v-if="isAdmin" class="btn-submit" @click="onSubmit" :loading="submitting">{{ id ? '保存' : '新增' }}</button>
    <button class="btn-delete" v-if="id && isAdmin" @click="onDelete">删除员工</button>
  </view>
  </SideLayout>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { onLoad } from '@dcloudio/uni-app'
import SideLayout from '@/components/SideLayout.vue'
import { getStaff, createStaff, updateStaff, deleteStaff, resetStaffPassword, getStaffSchedule, batchSetSchedule } from '@/api/staff'
import { safeBack } from '@/utils/navigate'
import { useAuthStore } from '@/store/auth'
import { hasStaffRoleAtLeast, staffRoleLabel } from '@/utils/staff-role'

const authStore = useAuthStore()
const isAdmin = computed(() => hasStaffRoleAtLeast(authStore.staffInfo?.role, 'admin'))
const canManageSchedule = computed(() => hasStaffRoleAtLeast(authStore.staffInfo?.role, 'manager'))

const id = ref(0)
const submitting = ref(false)
const newPassword = ref('')
const form = ref({ name: '', phone: '', password: '', role: 'staff', commission_rate: 0, product_commission_rate: 0, feeding_commission_rate: 0, status: 1 })
const roleList = ['staff', 'manager', 'admin'].map((value) => ({ label: staffRoleLabel(value), value }))
const statusList = [
  { label: '在职', value: 1 },
  { label: '离职', value: 2 },
]

const statusIndex = computed(() => statusList.findIndex(s => s.value === form.value.status) || 0)
function onStatusChange(e: any) { form.value.status = statusList[e.detail.value].value }
const roleIndex = computed(() => Math.max(roleList.findIndex(s => s.value === form.value.role), 0))
function onRoleChange(e: any) { form.value.role = roleList[e.detail.value].value }

// 排班
interface DaySchedule {
  date: string
  start_time: string
  end_time: string
  is_day_off: boolean
  max_capacity: number
}
const dayNames = ['周一', '周二', '周三', '周四', '周五', '周六', '周日']
const weekSchedule = ref<DaySchedule[]>([])

function getWeekDates(): string[] {
  const now = new Date()
  const day = now.getDay() || 7 // 周日=7
  const monday = new Date(now)
  monday.setDate(now.getDate() - day + 1)
  return Array.from({ length: 7 }, (_, i) => {
    const d = new Date(monday)
    d.setDate(monday.getDate() + i)
    return d.toISOString().split('T')[0]
  })
}

function initWeekSchedule() {
  weekSchedule.value = getWeekDates().map(date => ({
    date, start_time: '12:00', end_time: '22:00', is_day_off: false, max_capacity: 1,
  }))
}

async function loadSchedule() {
  const dates = getWeekDates()
  try {
    const res = await getStaffSchedule(id.value, dates[0], dates[6])
    const existing = res.data || []
    weekSchedule.value = dates.map(date => {
      const found = existing.find((s: any) => s.date === date)
      return found
        ? { date, start_time: found.start_time, end_time: found.end_time, is_day_off: found.is_day_off, max_capacity: found.max_capacity || 1 }
        : { date, start_time: '12:00', end_time: '22:00', is_day_off: false, max_capacity: 1 }
    })
  } catch {
    initWeekSchedule()
  }
}

function copyToAll() {
  if (!weekSchedule.value.length) return
  const first = weekSchedule.value[0]
  weekSchedule.value.forEach((day, i) => {
    if (i > 0) {
      day.start_time = first.start_time
      day.end_time = first.end_time
      day.is_day_off = first.is_day_off
      day.max_capacity = first.max_capacity
    }
  })
}

async function saveSchedule() {
  try {
    await batchSetSchedule(id.value, weekSchedule.value.map(d => ({
      date: d.date,
      start_time: d.is_day_off ? '' : d.start_time,
      end_time: d.is_day_off ? '' : d.end_time,
      is_day_off: d.is_day_off,
      max_capacity: Math.max(1, Number(d.max_capacity) || 1),
    })))
    uni.showToast({ title: '排班已保存', icon: 'success' })
  } catch (e: any) {
    uni.showToast({ title: e.message || '保存失败', icon: 'none' })
  }
}

onLoad((query) => {
  if (query?.id) {
    id.value = parseInt(query.id)
    loadData()
    if (canManageSchedule.value) loadSchedule()
  }
})

async function loadData() {
  const res = await getStaff(id.value)
  form.value = {
    name: res.data.name,
    phone: res.data.phone,
    password: '',
    role: res.data.role || 'staff',
    commission_rate: res.data.commission_rate,
    product_commission_rate: res.data.product_commission_rate || 0,
    feeding_commission_rate: res.data.feeding_commission_rate || 0,
    status: res.data.status,
  }
}

async function onSubmit() {
  if (!form.value.name || !form.value.phone) {
    uni.showToast({ title: '请填写必填项', icon: 'none' })
    return
  }
  submitting.value = true
  try {
    const payload = {
      name: form.value.name,
      phone: form.value.phone,
      commission_rate: toNumber(form.value.commission_rate),
      product_commission_rate: toNumber(form.value.product_commission_rate),
      feeding_commission_rate: toNumber(form.value.feeding_commission_rate),
      role: form.value.role,
      status: form.value.status,
    }
    if (id.value) {
      await updateStaff(id.value, payload)
    } else {
      await createStaff({
        phone: payload.phone,
        name: payload.name,
        password: form.value.password || undefined,
        role: payload.role,
        commission_rate: payload.commission_rate,
        product_commission_rate: payload.product_commission_rate,
        feeding_commission_rate: payload.feeding_commission_rate,
      })
    }
    uni.showToast({ title: '保存成功', icon: 'success' })
    setTimeout(() => safeBack(), 500)
  } finally {
    submitting.value = false
  }
}

function toNumber(value: string | number) {
  const num = Number(value)
  return Number.isFinite(num) ? num : 0
}

async function onResetPassword() {
  if (!newPassword.value || newPassword.value.length < 6) {
    uni.showToast({ title: '密码至少6位', icon: 'none' })
    return
  }
  uni.showModal({
    title: '确认重置密码',
    content: `将员工密码重置为「${newPassword.value}」？`,
    success: async (res) => {
      if (res.confirm) {
        try {
          await resetStaffPassword(id.value, newPassword.value)
          uni.showToast({ title: '密码已重置', icon: 'success' })
          newPassword.value = ''
        } catch (e: any) {
          uni.showToast({ title: e.message || '重置失败', icon: 'none' })
        }
      }
    }
  })
}

async function onDelete() {
  uni.showModal({
    title: '确认删除',
    content: '删除后不可恢复，确认删除该员工？',
    success: async (res) => {
      if (res.confirm) {
        await deleteStaff(id.value)
        uni.showToast({ title: '已删除', icon: 'success' })
        setTimeout(() => safeBack(), 500)
      }
    }
  })
}
</script>

<style scoped>
.page { padding: 24rpx; }
.form { background: #fff; border-radius: 16rpx; padding: 8rpx 24rpx; margin-bottom: 24rpx; }
.form-item { padding: 24rpx 0; border-bottom: 1rpx solid #F3F4F6; }
.form-item:last-child { border-bottom: none; }
.label { font-size: 28rpx; color: #374151; display: block; margin-bottom: 12rpx; }
.input { font-size: 28rpx; color: #1F2937; height: 60rpx; }
.picker { font-size: 28rpx; color: #1F2937; height: 60rpx; line-height: 60rpx; }
.section { background: #fff; border-radius: 16rpx; padding: 20rpx 24rpx; margin-bottom: 24rpx; }
.section-header { margin-bottom: 16rpx; }
.section-title { font-size: 26rpx; font-weight: 600; color: #6B7280; }
.reset-row { display: flex; gap: 16rpx; align-items: center; }
.flex1 { flex: 1; }
.btn-reset { font-size: 26rpx; color: #fff; background: #F59E0B; padding: 12rpx 24rpx; border-radius: 12rpx; white-space: nowrap; }
/* 排班 */
.hint { font-size: 22rpx; color: #9CA3AF; }
.schedule-row { display: flex; align-items: center; padding: 14rpx 0; border-bottom: 1rpx solid #F3F4F6; }
.schedule-row:last-child { border-bottom: none; }
.day-label { width: 100rpx; }
.day-name { font-size: 26rpx; font-weight: 600; color: #1F2937; display: block; }
.day-date { font-size: 20rpx; color: #9CA3AF; }
.time-inputs { display: flex; align-items: center; gap: 8rpx; flex: 1; }
.time-input { width: 120rpx; text-align: center; font-size: 26rpx; color: #1F2937; background: #F9FAFB; border-radius: 8rpx; height: 56rpx; }
.cap-input { width: 96rpx; text-align: center; font-size: 24rpx; color: #1F2937; background: #EEF2FF; border-radius: 8rpx; height: 56rpx; }
.time-sep { font-size: 24rpx; color: #9CA3AF; }
.btn-day-off { font-size: 22rpx; color: #9CA3AF; padding: 8rpx 16rpx; background: #F3F4F6; border-radius: 8rpx; }
.day-off-tag { font-size: 24rpx; color: #9CA3AF; background: #F3F4F6; padding: 8rpx 24rpx; border-radius: 8rpx; flex: 1; text-align: center; }
.schedule-actions { display: flex; gap: 16rpx; margin-top: 16rpx; }
.btn-copy { flex: 1; text-align: center; font-size: 26rpx; color: #6B7280; background: #F3F4F6; padding: 14rpx; border-radius: 10rpx; }
.btn-save-schedule { flex: 1; text-align: center; font-size: 26rpx; color: #fff; background: #4F46E5; padding: 14rpx; border-radius: 10rpx; }

.btn-submit { background: #4F46E5; color: #fff; border-radius: 12rpx; font-size: 30rpx; margin-top: 16rpx; }
.btn-delete { background: #fff; color: #DC2626; border: 1rpx solid #DC2626; border-radius: 12rpx; font-size: 30rpx; margin-top: 16rpx; }
</style>
