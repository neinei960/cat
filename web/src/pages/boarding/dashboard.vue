<template>
  <SideLayout>
    <view class="page">
      <view class="hero">
        <view>
          <text class="title">寄养看板</text>
          <text class="subtitle">按房型看总间数、已住、待入住和剩余可售间数。</text>
        </view>
        <view class="hero-actions">
          <view class="btn btn-primary" @click="go('/pages/boarding/create')">新建寄养</view>
          <view v-if="isAdmin" class="btn" @click="go('/pages/boarding/cabinets')">房型设置</view>
          <view v-if="isAdmin" class="btn" @click="go('/pages/boarding/policies')">优惠</view>
          <view v-if="isAdmin" class="btn" @click="go('/pages/boarding/holidays')">节假日</view>
        </view>
      </view>

      <view v-if="loading" class="state">加载中...</view>
      <view v-else-if="groups.length === 0" class="state">还没有寄养房型，先去添加房型。</view>

      <view v-else class="group-list">
        <view class="group-card" v-for="group in groups" :key="group.cabinet_id">
          <view class="group-head">
            <view>
              <text class="group-title">{{ group.cabinet_type }}</text>
              <text class="group-meta">每间可住 {{ group.capacity }} 只 · ¥{{ group.base_price }}/晚</text>
            </view>
            <text class="group-status">{{ statusLabel(group.status) }}</text>
          </view>

          <view class="stats-row">
            <view class="stat-card">
              <text class="stat-value">{{ group.room_count }}</text>
              <text class="stat-label">总间数</text>
            </view>
            <view class="stat-card occupied">
              <text class="stat-value">{{ group.occupied_rooms }}</text>
              <text class="stat-label">在住</text>
            </view>
            <view class="stat-card reserved">
              <text class="stat-value">{{ group.reserved_rooms }}</text>
              <text class="stat-label">待入住</text>
            </view>
            <view class="stat-card available">
              <text class="stat-value">{{ group.remaining_rooms }}</text>
              <text class="stat-label">可售</text>
            </view>
          </view>

          <text v-if="group.remark" class="remark">{{ group.remark }}</text>

          <view v-if="group.orders?.length" class="order-list">
            <view class="order-row" v-for="item in group.orders" :key="item.ID" @click="go(`/pages/boarding/detail?id=${item.ID}`)">
              <view>
                <text class="order-pets">{{ petNames(item) }}</text>
                <text class="order-meta">{{ item.customer?.nickname || item.customer?.phone || '-' }} · {{ statusLabel(item.status) }}</text>
              </view>
              <text class="order-date">{{ item.check_in_at }} → {{ item.check_out_at }}</text>
            </view>
          </view>
        </view>
      </view>
    </view>
  </SideLayout>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { onShow } from '@dcloudio/uni-app'
import SideLayout from '@/components/SideLayout.vue'
import { getBoardingDashboard } from '@/api/boarding'
import { useAuthStore } from '@/store/auth'
import { hasStaffRoleAtLeast } from '@/utils/staff-role'

const authStore = useAuthStore()
const isAdmin = computed(() => hasStaffRoleAtLeast(authStore.staffInfo?.role, 'admin'))
const loading = ref(false)
const groups = ref<BoardingDashboardGroup[]>([])

function go(url: string) {
  uni.navigateTo({ url })
}

function statusLabel(status: string) {
  return {
    idle: '空闲',
    reserved: '待入住',
    occupied: '在住',
    cleaning: '清洁中',
    disabled: '停用',
    enabled: '启用',
    pending_checkin: '待入住',
    checked_in: '在住',
    checked_out: '已离店',
    cancelled: '已取消',
  }[status] || status
}

function petNames(order?: BoardingOrder) {
  return order?.pets?.map((item) => item.pet?.name || item.pet_name_snapshot).filter(Boolean).join('、') || '未选猫咪'
}

async function loadData() {
  loading.value = true
  try {
    const res = await getBoardingDashboard()
    groups.value = res.data || []
  } finally {
    loading.value = false
  }
}

onShow(loadData)
</script>

<style scoped>
.page { padding: 24rpx; }
.hero { display: flex; justify-content: space-between; gap: 16rpx; align-items: flex-start; margin-bottom: 24rpx; }
.title { display: block; font-size: 36rpx; font-weight: 700; color: #111827; }
.subtitle { display: block; margin-top: 10rpx; font-size: 24rpx; color: #6B7280; line-height: 1.6; }
.hero-actions { display: flex; flex-wrap: wrap; gap: 12rpx; justify-content: flex-end; }
.btn { padding: 14rpx 22rpx; background: #fff; border: 1rpx solid #E5E7EB; border-radius: 12rpx; font-size: 24rpx; color: #374151; }
.btn-primary { background: #4F46E5; color: #fff; border-color: #4F46E5; }
.state { text-align: center; padding: 120rpx 24rpx; color: #9CA3AF; font-size: 28rpx; }
.group-list { display: flex; flex-direction: column; gap: 20rpx; }
.group-card { background: #fff; border-radius: 20rpx; padding: 24rpx; box-shadow: 0 12rpx 28rpx rgba(15, 23, 42, 0.05); }
.group-head { display: flex; justify-content: space-between; align-items: flex-start; gap: 16rpx; }
.group-title { font-size: 30rpx; font-weight: 700; color: #1F2937; }
.group-meta { display: block; margin-top: 6rpx; font-size: 22rpx; color: #6B7280; }
.group-status { font-size: 22rpx; color: #4F46E5; }
.stats-row { display: grid; grid-template-columns: repeat(4, minmax(0, 1fr)); gap: 12rpx; margin-top: 18rpx; }
.stat-card { padding: 18rpx 12rpx; border-radius: 16rpx; background: #F8FAFC; text-align: center; }
.stat-card.occupied { background: #EEF2FF; }
.stat-card.reserved { background: #FFF7ED; }
.stat-card.available { background: #ECFDF5; }
.stat-value { display: block; font-size: 30rpx; font-weight: 700; color: #111827; }
.stat-label { display: block; margin-top: 6rpx; font-size: 22rpx; color: #6B7280; }
.remark { display: block; margin-top: 14rpx; font-size: 22rpx; color: #6B7280; line-height: 1.6; }
.order-list { margin-top: 18rpx; display: flex; flex-direction: column; gap: 12rpx; }
.order-row { display: flex; justify-content: space-between; gap: 16rpx; align-items: center; padding: 18rpx 20rpx; border-radius: 16rpx; background: #F9FAFB; }
.order-pets { display: block; font-size: 26rpx; font-weight: 600; color: #111827; }
.order-meta, .order-date { display: block; margin-top: 6rpx; font-size: 22rpx; color: #6B7280; text-align: right; }
@media (max-width: 768px) {
  .hero { flex-direction: column; }
  .hero-actions { width: 100%; justify-content: flex-start; }
}
</style>
