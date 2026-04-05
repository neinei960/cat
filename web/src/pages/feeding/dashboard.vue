<template>
  <SideLayout>
    <view class="page">
      <view class="hero-card">
        <view>
          <text class="hero-title">上门喂养看板</text>
          <text class="hero-subtitle">按日期查看待上门、已分配、进行中、异常和已完成任务。</text>
        </view>
        <view class="hero-actions">
          <view class="btn btn-primary" @click="go('/pages/feeding/create')">新建计划</view>
          <view v-if="canManage" class="btn" @click="go('/pages/feeding/settings')">设置</view>
        </view>
      </view>

      <view class="toolbar">
        <picker mode="date" :value="date" @change="onDateChange">
          <view class="toolbar-card">
            <text class="toolbar-label">日期</text>
            <text class="toolbar-value">{{ date }}</text>
          </view>
        </picker>
        <picker v-if="canManage" :range="staffOptions" range-key="name" :value="staffIndex" @change="onStaffChange">
          <view class="toolbar-card">
            <text class="toolbar-label">员工</text>
            <text class="toolbar-value">{{ staffOptions[staffIndex]?.name || '全部员工' }}</text>
          </view>
        </picker>
      </view>

      <view class="window-tabs">
        <view
          v-for="item in windowOptions"
          :key="item.value"
          :class="['window-tab', activeWindow === item.value ? 'active' : '']"
          @click="setWindow(item.value)"
        >
          {{ item.label }}
        </view>
      </view>

      <view v-if="loading" class="state-card">加载中...</view>
      <view v-else class="group-list">
        <view class="group-card" v-for="group in groups" :key="group.status">
          <view class="group-head">
            <text class="group-title">{{ group.label }}</text>
            <text class="group-count">{{ group.count }}</text>
          </view>
          <view v-if="group.visits?.length" class="visit-list">
            <view class="visit-card" v-for="visit in group.visits" :key="visit.ID" @click="openPlan(visit)">
              <view class="visit-head">
                <text class="visit-time">{{ feedingWindowLabel(visit.window_code) }}</text>
                <text class="visit-status">{{ feedingStatusLabel(visit.status) }}</text>
              </view>
              <text class="visit-main">{{ visit.plan?.customer?.nickname || visit.plan?.customer?.phone || '未命名客户' }}</text>
              <text class="visit-sub">{{ petSummary(visit.plan) }}</text>
              <text class="visit-sub">执行人：{{ visit.staff?.name || '待分配' }}</text>
              <view class="visit-actions">
                <view class="mini-btn" @click.stop="openPlan(visit)">详情</view>
                <view class="mini-btn primary" @click.stop="openExecute(visit)">执行</view>
              </view>
            </view>
          </view>
          <view v-else class="empty-card">当前分组暂无任务</view>
        </view>
      </view>
    </view>
  </SideLayout>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { onShow } from '@dcloudio/uni-app'
import SideLayout from '@/components/SideLayout.vue'
import { getFeedingDashboard } from '@/api/feeding'
import { getStaffList } from '@/api/staff'
import { useAuthStore } from '@/store/auth'
import { hasStaffRoleAtLeast } from '@/utils/staff-role'
import { feedingStatusLabel, feedingWindowLabel } from '@/utils/feeding'

const authStore = useAuthStore()
const canManage = computed(() => hasStaffRoleAtLeast(authStore.staffInfo?.role, 'manager'))
const date = ref(formatLocalDate(new Date()))
const loading = ref(false)
const groups = ref<FeedingDashboardGroup[]>([])
const activeWindow = ref('')
const staffOptions = ref<Array<{ ID: number; name: string }>>([{ ID: 0, name: '全部员工' }])
const staffIndex = ref(0)

const windowOptions = [
  { value: '', label: '全部时段' },
  { value: 'morning', label: '早间' },
  { value: 'afternoon', label: '午后' },
  { value: 'evening', label: '晚间' },
]

function go(url: string) {
  uni.navigateTo({ url })
}

function onDateChange(event: any) {
  date.value = event.detail.value
  loadData()
}

function onStaffChange(event: any) {
  staffIndex.value = Number(event.detail.value || 0)
  loadData()
}

function setWindow(value: string) {
  activeWindow.value = value
  loadData()
}

function petSummary(plan?: FeedingPlan) {
  return plan?.pets?.map(item => item.pet?.name || item.pet_name_snapshot).filter(Boolean).join('、') || '未选猫咪'
}

function openPlan(visit: FeedingVisit) {
  if (!visit.plan?.ID) return
  uni.navigateTo({ url: `/pages/feeding/detail?id=${visit.plan.ID}` })
}

function openExecute(visit: FeedingVisit) {
  if (!visit.plan?.ID) return
  uni.navigateTo({ url: `/pages/feeding/visit-execute?plan_id=${visit.plan.ID}&visit_id=${visit.ID}` })
}

async function loadStaffs() {
  if (!canManage.value) return
  const res = await getStaffList({ page: 1, page_size: 100 })
  const list = (res.data?.list || []).map(item => ({ ID: item.ID, name: item.name }))
  staffOptions.value = [{ ID: 0, name: '全部员工' }, ...list]
}

async function loadData() {
  loading.value = true
  try {
    const selectedStaff = staffOptions.value[staffIndex.value]?.ID || 0
    const res = await getFeedingDashboard({
      date: date.value,
      staff_id: selectedStaff || undefined,
      window_code: activeWindow.value || undefined,
    })
    groups.value = res.data?.groups || []
  } finally {
    loading.value = false
  }
}

onShow(async () => {
  await loadStaffs()
  await loadData()
})

function formatLocalDate(value: Date) {
  const year = value.getFullYear()
  const month = String(value.getMonth() + 1).padStart(2, '0')
  const day = String(value.getDate()).padStart(2, '0')
  return `${year}-${month}-${day}`
}
</script>

<style scoped>
.page { padding: 24rpx; }
.hero-card, .toolbar-card, .group-card, .state-card { background: #fff; border-radius: 22rpx; box-shadow: 0 12rpx 28rpx rgba(15, 23, 42, 0.06); }
.hero-card { padding: 24rpx; display: flex; justify-content: space-between; gap: 20rpx; margin-bottom: 20rpx; }
.hero-title { display: block; font-size: 36rpx; font-weight: 700; color: #111827; }
.hero-subtitle { display: block; margin-top: 10rpx; font-size: 24rpx; color: #6B7280; line-height: 1.6; }
.hero-actions { display: flex; gap: 12rpx; flex-wrap: wrap; justify-content: flex-end; }
.btn { padding: 16rpx 24rpx; border-radius: 16rpx; background: #F8FAFC; color: #374151; font-size: 24rpx; border: 1rpx solid #E5E7EB; }
.btn-primary { background: linear-gradient(135deg, #4F46E5, #6366F1); color: #fff; border-color: transparent; }
.toolbar { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 16rpx; margin-bottom: 16rpx; }
.toolbar-card { padding: 20rpx; }
.toolbar-label { display: block; font-size: 22rpx; color: #6B7280; }
.toolbar-value { display: block; margin-top: 8rpx; font-size: 28rpx; color: #111827; font-weight: 600; }
.window-tabs { display: flex; gap: 12rpx; margin-bottom: 20rpx; overflow-x: auto; }
.window-tab { flex: 0 0 auto; padding: 14rpx 24rpx; border-radius: 999rpx; background: #EEF2FF; color: #4F46E5; font-size: 24rpx; }
.window-tab.active { background: #4F46E5; color: #fff; }
.state-card { padding: 120rpx 24rpx; text-align: center; color: #9CA3AF; font-size: 28rpx; }
.group-list { display: flex; flex-direction: column; gap: 18rpx; }
.group-card { padding: 22rpx; }
.group-head { display: flex; justify-content: space-between; align-items: center; margin-bottom: 14rpx; }
.group-title { font-size: 28rpx; font-weight: 700; color: #111827; }
.group-count { min-width: 48rpx; height: 48rpx; line-height: 48rpx; text-align: center; border-radius: 999rpx; background: #EEF2FF; color: #4F46E5; font-size: 24rpx; font-weight: 700; }
.visit-list { display: flex; flex-direction: column; gap: 12rpx; }
.visit-card { padding: 18rpx; border-radius: 18rpx; background: #F8FAFC; }
.visit-head { display: flex; justify-content: space-between; align-items: center; gap: 12rpx; }
.visit-time, .visit-status { font-size: 22rpx; color: #6B7280; }
.visit-main { display: block; margin-top: 10rpx; font-size: 28rpx; font-weight: 700; color: #111827; }
.visit-sub { display: block; margin-top: 6rpx; font-size: 22rpx; color: #6B7280; line-height: 1.6; }
.visit-actions { display: flex; gap: 12rpx; margin-top: 14rpx; }
.mini-btn { padding: 10rpx 20rpx; border-radius: 999rpx; background: #fff; color: #4B5563; font-size: 22rpx; border: 1rpx solid #E5E7EB; }
.mini-btn.primary { background: #4F46E5; color: #fff; border-color: #4F46E5; }
.empty-card { padding: 26rpx 18rpx; border-radius: 18rpx; background: #F9FAFB; text-align: center; color: #9CA3AF; font-size: 24rpx; }
@media (max-width: 768px) {
  .hero-card { flex-direction: column; }
  .hero-actions { justify-content: flex-start; }
}
</style>
