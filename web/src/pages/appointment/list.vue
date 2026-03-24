<template>
  <SideLayout>
  <view class="page">
    <view class="header">
      <text class="title">预约列表</text>
      <view class="btn-add" @click="goCreate">+ 新建预约</view>
    </view>

    <view class="tabs">
      <view :class="['tab', activeStatus === -1 ? 'tab-all-active' : '']" @click="switchTab(-1)">全部</view>
      <view
        v-for="status in primaryStatuses"
        :key="status"
        class="tab"
        :style="getAppointmentStatusTabStyle(status, activeStatus === status)"
        @click="switchTab(status)"
      >
        {{ getAppointmentStatusLabel(status) }}
      </view>
    </view>

    <view v-if="loading" class="loading">加载中...</view>
    <view v-else-if="list.length === 0" class="empty">暂无预约</view>

    <view v-else class="list">
      <view class="card" v-for="item in list" :key="item.ID" :style="getCardStyle(item.status)" @click="goDetail(item.ID)">
        <view class="card-top">
          <view class="date-time">
            <text class="date">{{ item.date }}</text>
            <text class="time">{{ item.start_time }} - {{ item.end_time }}</text>
          </view>
          <view class="status" :style="getAppointmentStatusBadgeStyle(item.status)">{{ getAppointmentStatusLabel(item.status) }}</view>
        </view>
        <view class="card-body">
          <text class="pet-name">{{ getPetSummary(item) }}</text>
          <text class="customer-name">{{ item.customer?.nickname || item.customer?.phone || '-' }}</text>
          <text class="staff-name" v-if="item.staff">技师: {{ item.staff.name }}</text>
          <text class="staff-name" v-else>待分配技师</text>
        </view>
        <view class="card-footer">
          <text class="services">{{ getServiceSummary(item) }}</text>
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
import {
  getAppointmentStatusBadgeStyle,
  getAppointmentStatusBlockStyle,
  getAppointmentStatusLabel,
  getAppointmentStatusTabStyle,
} from '@/utils/appointment-status'

const list = ref<any[]>([])
const loading = ref(true)
const activeStatus = ref(-1)
const primaryStatuses = [0, 1, 2, 3]

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

function getCardStyle(status: number) {
  const blockStyle = getAppointmentStatusBlockStyle(status)
  return {
    borderLeft: `8rpx solid ${blockStyle.borderLeftColor}`,
  }
}

function getAppointmentPets(item: any) {
  if (Array.isArray(item?.pets) && item.pets.length > 0) {
    return item.pets
  }
  if (item?.pet) {
    return [{
      pet_id: item.pet.ID,
      pet: item.pet,
      services: item.services || [],
    }]
  }
  return []
}

function getPetSummary(item: any) {
  if (item?.pet_summary) return item.pet_summary
  const names = getAppointmentPets(item)
    .map((petItem: any) => petItem.pet?.name)
    .filter(Boolean)
  if (names.length === 0) return '-'
  if (names.length === 1) return names[0]
  return `${names[0]}等${names.length}只`
}

function getServiceSummary(item: any) {
  const petGroups = getAppointmentPets(item)
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

  return (item.services || []).map((service: any) => service.service_name).join(' + ') || '-'
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
.tab { font-size: 24rpx; padding: 10rpx 20rpx; border-radius: 20rpx; background: #F8FAFC; color: #6B7280; white-space: nowrap; border: 1rpx solid #E5E7EB; }
.tab-all-active { background: #111827; color: #fff; border-color: #111827; }
.loading, .empty { text-align: center; padding: 100rpx 0; color: #9CA3AF; font-size: 28rpx; }
.card { background: #fff; border-radius: 16rpx; padding: 24rpx; margin-bottom: 16rpx; box-shadow: 0 2rpx 8rpx rgba(0,0,0,0.04); }
.card-top { display: flex; justify-content: space-between; align-items: center; margin-bottom: 12rpx; }
.date { font-size: 28rpx; font-weight: 600; color: #1F2937; }
.time { font-size: 24rpx; color: #6B7280; margin-left: 12rpx; }
.status { font-size: 22rpx; padding: 6rpx 16rpx; border-radius: 16rpx; }
.card-body { display: flex; gap: 16rpx; align-items: center; margin-bottom: 12rpx; flex-wrap: wrap; }
.pet-name { font-size: 28rpx; font-weight: 600; color: #1F2937; }
.customer-name { font-size: 26rpx; color: #6B7280; }
.staff-name { font-size: 24rpx; color: #4F46E5; }
.card-footer { display: flex; justify-content: space-between; gap: 16rpx; padding-top: 12rpx; border-top: 1rpx solid #F3F4F6; align-items: flex-start; }
.services { font-size: 24rpx; color: #6B7280; flex: 1; line-height: 1.5; }
.amount { font-size: 28rpx; font-weight: bold; color: #4F46E5; }
</style>
