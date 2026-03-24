<template>
  <SideLayout>
  <view class="page">
    <view class="header">
      <text class="title">员工管理</text>
      <view class="btn-add" v-if="isAdmin" @click="goAdd">+ 新增员工</view>
    </view>

    <view v-if="loading" class="loading">加载中...</view>

    <view v-else-if="list.length === 0" class="empty">暂无员工</view>

    <view v-else class="list">
      <view class="card" v-for="item in list" :key="item.ID" @click="goEdit(item.ID)">
        <view class="card-top">
          <view class="avatar">{{ item.name.charAt(0) }}</view>
          <view class="info">
            <text class="name">{{ item.name }}</text>
            <text class="role">{{ roleMap[item.role] || item.role }}</text>
          </view>
          <view :class="['status', item.status === 1 ? 'active' : 'inactive']">
            {{ item.status === 1 ? '在职' : '离职' }}
          </view>
        </view>
        <view class="card-bottom">
          <text class="phone">{{ item.phone }}</text>
          <view class="commissions" v-if="item.commission_rate || item.product_commission_rate || item.feeding_commission_rate">
            <text class="commission" v-if="item.commission_rate">洗{{ item.commission_rate }}%</text>
            <text class="commission" v-if="item.product_commission_rate">品{{ item.product_commission_rate }}%</text>
            <text class="commission" v-if="item.feeding_commission_rate">喂{{ item.feeding_commission_rate }}%</text>
          </view>
        </view>
      </view>
    </view>
  </view>
  </SideLayout>
</template>

<script setup lang="ts">
import SideLayout from '@/components/SideLayout.vue'
import { ref, computed, onMounted } from 'vue'
import { getStaffList } from '@/api/staff'
import { useAuthStore } from '@/store/auth'

const authStore = useAuthStore()
const isAdmin = computed(() => authStore.staffInfo?.role === 'admin')

const list = ref<Staff[]>([])
const loading = ref(true)
const roleMap: Record<string, string> = { admin: '店主', manager: '经理', staff: '技师' }

async function loadData() {
  loading.value = true
  try {
    const res = await getStaffList({ page: 1, page_size: 100 })
    list.value = res.data.list || []
  } finally {
    loading.value = false
  }
}

function goAdd() {
  uni.navigateTo({ url: '/pages/staff/edit' })
}

function goEdit(id: number) {
  uni.navigateTo({ url: `/pages/staff/edit?id=${id}` })
}

onMounted(loadData)
</script>

<style scoped>
.page { padding: 24rpx; }
.header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 24rpx; }
.title { font-size: 36rpx; font-weight: bold; color: #1F2937; }
.btn-add { font-size: 28rpx; color: #fff; background: #4F46E5; padding: 12rpx 24rpx; border-radius: 12rpx; }
.loading, .empty { text-align: center; padding: 100rpx 0; color: #9CA3AF; font-size: 28rpx; }
.card { background: #fff; border-radius: 16rpx; padding: 24rpx; margin-bottom: 16rpx; box-shadow: 0 2rpx 8rpx rgba(0,0,0,0.04); }
.card-top { display: flex; align-items: center; }
.avatar { width: 80rpx; height: 80rpx; border-radius: 50%; background: #4F46E5; color: #fff; display: flex; align-items: center; justify-content: center; font-size: 32rpx; font-weight: bold; }
.info { flex: 1; margin-left: 20rpx; }
.name { font-size: 30rpx; font-weight: 600; color: #1F2937; display: block; }
.role { font-size: 24rpx; color: #6B7280; display: block; margin-top: 4rpx; }
.status { font-size: 24rpx; padding: 6rpx 16rpx; border-radius: 20rpx; }
.status.active { color: #059669; background: #D1FAE5; }
.status.inactive { color: #DC2626; background: #FEE2E2; }
.card-bottom { display: flex; justify-content: space-between; margin-top: 16rpx; padding-top: 16rpx; border-top: 1rpx solid #F3F4F6; }
.phone { font-size: 26rpx; color: #6B7280; }
.commissions { display: flex; gap: 12rpx; }
.commission { font-size: 24rpx; color: #4F46E5; }
</style>
