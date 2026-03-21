<template>
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
      <!-- 创建时设置初始密码 -->
      <view class="form-item" v-if="!id && isAdmin">
        <text class="label">初始密码</text>
        <input v-model="form.password" type="text" placeholder="留空默认 123456" class="input" />
      </view>
      <!-- 仅 admin 可见的管理字段 -->
      <view class="form-item" v-if="isAdmin">
        <text class="label">提成比例 (%)</text>
        <input v-model="form.commission_rate" type="digit" placeholder="0" class="input" />
      </view>
      <view class="form-item" v-if="id && isAdmin">
        <text class="label">状态</text>
        <picker :range="statusList" :range-key="'label'" :value="statusIndex" @change="onStatusChange">
          <view class="picker">{{ statusList[statusIndex].label }}</view>
        </picker>
      </view>
    </view>

    <!-- 重置密码区域 (编辑模式, 仅 admin) -->
    <view class="section" v-if="id && isAdmin">
      <view class="section-header">
        <text class="section-title">密码管理</text>
      </view>
      <view class="reset-row">
        <input v-model="newPassword" type="text" placeholder="输入新密码（至少6位）" class="input flex1" />
        <view class="btn-reset" @click="onResetPassword">重置密码</view>
      </view>
    </view>

    <button class="btn-submit" @click="onSubmit" :loading="submitting">{{ id ? '保存' : '新增' }}</button>
    <button class="btn-delete" v-if="id && isAdmin" @click="onDelete">删除员工</button>
  </view>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { onLoad } from '@dcloudio/uni-app'
import { getStaff, createStaff, updateStaff, deleteStaff, resetStaffPassword } from '@/api/staff'
import { safeBack } from '@/utils/navigate'
import { useAuthStore } from '@/store/auth'

const authStore = useAuthStore()
const isAdmin = computed(() => authStore.staffInfo?.role === 'admin')

const id = ref(0)
const submitting = ref(false)
const newPassword = ref('')
const form = ref({ name: '', phone: '', password: '', commission_rate: 0, status: 1 })
const statusList = [
  { label: '在职', value: 1 },
  { label: '离职', value: 2 },
]

const statusIndex = computed(() => statusList.findIndex(s => s.value === form.value.status) || 0)
function onStatusChange(e: any) { form.value.status = statusList[e.detail.value].value }

onLoad((query) => {
  if (query?.id) {
    id.value = parseInt(query.id)
    loadData()
  }
})

async function loadData() {
  const res = await getStaff(id.value)
  form.value = {
    name: res.data.name,
    phone: res.data.phone,
    password: '',
    commission_rate: res.data.commission_rate,
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
    if (id.value) {
      await updateStaff(id.value, {
        name: form.value.name,
        phone: form.value.phone,
        commission_rate: form.value.commission_rate,
        status: form.value.status,
      })
    } else {
      await createStaff({
        phone: form.value.phone,
        name: form.value.name,
        password: form.value.password || undefined,
        commission_rate: form.value.commission_rate,
      })
    }
    uni.showToast({ title: '保存成功', icon: 'success' })
    setTimeout(() => safeBack(), 500)
  } finally {
    submitting.value = false
  }
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
.btn-submit { background: #4F46E5; color: #fff; border-radius: 12rpx; font-size: 30rpx; margin-top: 16rpx; }
.btn-delete { background: #fff; color: #DC2626; border: 1rpx solid #DC2626; border-radius: 12rpx; font-size: 30rpx; margin-top: 16rpx; }
</style>
