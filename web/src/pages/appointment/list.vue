<template>
  <SideLayout>
  <view class="page">
    <view class="header">
      <text class="title">预约列表</text>
      <view class="btn-add" @click="goCreate">+ 新建预约</view>
    </view>

    <view class="tabs">
      <view :class="['tab', activeStatus === -1 ? 'active' : '']" @click="switchTab(-1)">全部</view>
      <view :class="['tab', activeStatus === 0 ? 'active' : '']" @click="switchTab(0)">待确认</view>
      <view :class="['tab', activeStatus === 1 ? 'active' : '']" @click="switchTab(1)">已确认</view>
      <view :class="['tab', activeStatus === 2 ? 'active' : '']" @click="switchTab(2)">进行中</view>
      <view :class="['tab', activeStatus === 3 ? 'active' : '']" @click="switchTab(3)">已完成</view>
    </view>

    <view v-if="loading" class="loading">加载中...</view>
    <view v-else-if="list.length === 0" class="empty">暂无预约</view>

    <view v-else class="list">
      <view class="card" v-for="item in list" :key="item.ID" @click="goDetail(item.ID)">
        <view class="card-top">
          <view class="date-time">
            <text class="date">{{ item.date }}</text>
            <text class="time">{{ item.start_time }} - {{ item.end_time }}</text>
          </view>
          <view :class="['status', `s${item.status}`]">{{ statusMap[item.status] }}</view>
        </view>
        <view class="card-body">
          <text class="pet-name">{{ item.pet?.name || '-' }}</text>
          <text class="customer-name">{{ item.customer?.nickname || '-' }}</text>
          <text class="staff-name" v-if="item.staff">技师: {{ item.staff.name }}</text>
          <text class="staff-name" v-else>待分配技师</text>
        </view>
        <view class="card-footer">
          <text class="services">{{ (item.services || []).map((s: any) => s.service_name).join(' + ') }}</text>
          <text class="amount">¥{{ item.total_amount }}</text>
        </view>
      </view>
    </view>
  </view>
  </SideLayout>
</template>

<script setup lang="ts">
import SideLayout from '@/components/SideLayout.vue'
import { ref, onMounted } from 'vue'
import { onShow } from '@dcloudio/uni-app'
import { getAppointmentList } from '@/api/appointment'

const list = ref<any[]>([])
const loading = ref(true)
const activeStatus = ref(-1)
const statusMap: Record<number, string> = { 0: '待确认', 1: '已确认', 2: '进行中', 3: '已完成', 4: '已取消', 5: '未到店' }

async function loadData() {
  loading.value = true
  try {
    const params: any = { page: 1, page_size: 50 }
    if (activeStatus.value >= 0) params.status = activeStatus.value
    const res = await getAppointmentList(params)
    list.value = res.data.list || []
  } finally { loading.value = false }
}

function switchTab(status: number) {
  activeStatus.value = status
  loadData()
}

function goCreate() { uni.navigateTo({ url: '/pages/appointment/create' }) }
function goDetail(id: number) { uni.navigateTo({ url: `/pages/appointment/detail?id=${id}` }) }

onMounted(loadData)
onShow(loadData)
</script>

<style scoped>
.page { padding: 24rpx; }
.header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 20rpx; }
.title { font-size: 36rpx; font-weight: bold; color: #1F2937; }
.btn-add { font-size: 28rpx; color: #fff; background: #4F46E5; padding: 12rpx 24rpx; border-radius: 12rpx; }
.tabs { display: flex; gap: 12rpx; margin-bottom: 24rpx; overflow-x: auto; }
.tab { font-size: 24rpx; padding: 10rpx 20rpx; border-radius: 20rpx; background: #F3F4F6; color: #6B7280; white-space: nowrap; }
.tab.active { background: #4F46E5; color: #fff; }
.loading, .empty { text-align: center; padding: 100rpx 0; color: #9CA3AF; font-size: 28rpx; }
.card { background: #fff; border-radius: 16rpx; padding: 24rpx; margin-bottom: 16rpx; box-shadow: 0 2rpx 8rpx rgba(0,0,0,0.04); }
.card-top { display: flex; justify-content: space-between; align-items: center; margin-bottom: 12rpx; }
.date { font-size: 28rpx; font-weight: 600; color: #1F2937; }
.time { font-size: 24rpx; color: #6B7280; margin-left: 12rpx; }
.status { font-size: 22rpx; padding: 6rpx 16rpx; border-radius: 16rpx; }
.s0 { background: #FEF3C7; color: #92400E; }
.s1 { background: #EEF2FF; color: #4F46E5; }
.s2 { background: #D1FAE5; color: #059669; }
.s3 { background: #F3F4F6; color: #6B7280; }
.s4 { background: #FEE2E2; color: #DC2626; }
.s5 { background: #FEE2E2; color: #DC2626; }
.card-body { display: flex; gap: 16rpx; align-items: center; margin-bottom: 12rpx; }
.pet-name { font-size: 28rpx; font-weight: 600; color: #1F2937; }
.customer-name { font-size: 26rpx; color: #6B7280; }
.staff-name { font-size: 24rpx; color: #4F46E5; }
.card-footer { display: flex; justify-content: space-between; padding-top: 12rpx; border-top: 1rpx solid #F3F4F6; }
.services { font-size: 24rpx; color: #6B7280; flex: 1; }
.amount { font-size: 28rpx; font-weight: bold; color: #4F46E5; }
</style>
