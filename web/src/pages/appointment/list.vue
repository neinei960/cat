<template>
  <SideLayout>
  <view class="page">
    <view class="header">
      <text class="title">预约列表</text>
      <view class="header-right">
        <view class="btn-filter" @click="showFilter = true">
          <text>筛选</text>
          <text v-if="activeFilterCount > 0" class="filter-badge">{{ activeFilterCount }}</text>
        </view>
        <view class="btn-add" @click="goCreate">+ 新建预约</view>
      </view>
    </view>

    <view class="tabs">
      <view :class="['tab', filter.status === -1 ? 'tab-all-active' : '']" @click="filter.status = -1; loadData()">全部</view>
      <view
        v-for="status in primaryStatuses"
        :key="status"
        class="tab"
        :style="getAppointmentStatusTabStyle(status, filter.status === status)"
        @click="filter.status = status; loadData()"
      >
        {{ getAppointmentStatusLabel(status) }}
      </view>
    </view>

    <view class="active-filters" v-if="activeFilterCount > 0">
      <text class="active-filters-text">已筛选: </text>
      <text v-if="filter.dateFrom || filter.dateTo" class="filter-tag">{{ filter.dateFrom || '...' }} ~ {{ filter.dateTo || '...' }} <text @click="filter.dateFrom = ''; filter.dateTo = ''; loadData()">✕</text></text>
      <text v-if="filter.staffId > 0" class="filter-tag">{{ getStaffName(filter.staffId) }} <text @click="filter.staffId = 0; loadData()">✕</text></text>
    </view>

    <FilterPanel
      :visible="showFilter"
      :filter="filter"
      :status-options="apptStatusOptions"
      status-label="预约状态"
      :staff-list="staffList"
      :categories="categories"
      @close="showFilter = false"
      @confirm="onFilterConfirm"
    />

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
          <view class="pet-row">
            <text class="pet-name">{{ getPetSummary(item) }}</text>
            <text v-if="hasAggression(item)" class="aggression-warn">⚡ 攻击风险</text>
          </view>
          <text class="customer-name">{{ item.customer?.nickname || item.customer?.phone || '-' }}</text>
          <text class="staff-name" v-if="item.staff">洗护师: {{ item.staff.name }}</text>
          <text class="staff-name" v-else>待分配洗护师</text>
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
import FilterPanel from '@/components/FilterPanel.vue'
import { ref, reactive, computed, onMounted } from 'vue'
import { onShow } from '@dcloudio/uni-app'
import { getAppointmentList } from '@/api/appointment'
import { getStaffList } from '@/api/staff'
import { getCategoryTree } from '@/api/service-category'
import {
  getAppointmentStatusBadgeStyle,
  getAppointmentStatusBlockStyle,
  getAppointmentStatusLabel,
  getAppointmentStatusTabStyle,
} from '@/utils/appointment-status'

const list = ref<any[]>([])
const loading = ref(true)
const showFilter = ref(false)
const staffList = ref<any[]>([])
const categories = ref<any[]>([])
const primaryStatuses = [0, 1, 2, 3]

const apptStatusOptions = [
  { value: 0, label: '待确认' },
  { value: 1, label: '已确认' },
  { value: 2, label: '服务中' },
  { value: 3, label: '待结算' },
  { value: 7, label: '已开单' },
  { value: 4, label: '已取消' },
  { value: 5, label: '未到店' },
]

const filter = reactive({
  dateFrom: '',
  dateTo: '',
  status: -1,
  staffId: 0,
  payMethod: '',
  categoryId: 0,
})

const activeFilterCount = computed(() => {
  let c = 0
  if (filter.dateFrom || filter.dateTo) c++
  if (filter.staffId > 0) c++
  return c
})

function getStaffName(id: number) {
  return staffList.value.find((s: any) => s.ID === id)?.name || '未知'
}

function onFilterConfirm(f: any) {
  Object.assign(filter, f)
  loadData()
}

async function loadData() {
  loading.value = true
  try {
    const params: any = { page: 1, page_size: 50 }
    if (filter.status >= 0) params.status = filter.status
    if (filter.dateFrom) params.date_from = filter.dateFrom
    if (filter.dateTo) params.date_to = filter.dateTo
    if (filter.staffId > 0) params.staff_id = filter.staffId
    const res = await getAppointmentList(params)
    list.value = res.data.list || []
  } finally { loading.value = false }
}

async function loadFilterOptions() {
  try {
    const [stRes, catRes] = await Promise.all([
      getStaffList({ page: 1, page_size: 50 }),
      getCategoryTree(),
    ])
    staffList.value = (stRes.data?.list || []).filter((s: any) => s.status === 1)
    categories.value = (catRes.data || []).filter((c: any) => !c.parent_id && c.status === 1)
  } catch {}
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

function hasAggression(item: any): boolean {
  const pets = getAppointmentPets(item)
  return pets.some((p: any) => p.pet?.aggression && p.pet.aggression !== '无')
}

function goCreate() { uni.navigateTo({ url: '/pages/appointment/create' }) }
function goDetail(id: number) { uni.navigateTo({ url: `/pages/appointment/detail?id=${id}` }) }

onMounted(() => { loadData(); loadFilterOptions() })
onShow(loadData)
</script>

<style scoped>
.page { padding: 24rpx; }
.header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 20rpx; }
.header-right { display: flex; gap: 12rpx; align-items: center; }
.title { font-size: 36rpx; font-weight: bold; color: #1F2937; }
.btn-add { font-size: 28rpx; color: #fff; background: #4F46E5; padding: 12rpx 24rpx; border-radius: 12rpx; }
.btn-filter { display: flex; align-items: center; gap: 6rpx; font-size: 26rpx; color: #374151; background: #F3F4F6; padding: 12rpx 24rpx; border-radius: 12rpx; border: 1rpx solid #E5E7EB; }
.filter-badge { background: #EF4444; color: #fff; font-size: 20rpx; min-width: 28rpx; height: 28rpx; border-radius: 999rpx; display: flex; align-items: center; justify-content: center; padding: 0 6rpx; }
.active-filters { display: flex; flex-wrap: wrap; gap: 8rpx; align-items: center; margin-bottom: 16rpx; }
.active-filters-text { font-size: 22rpx; color: #6B7280; }
.filter-tag { font-size: 22rpx; color: #4F46E5; background: #EEF2FF; padding: 6rpx 16rpx; border-radius: 20rpx; }
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
.pet-row { display: flex; align-items: center; gap: 8rpx; }
.pet-name { font-size: 28rpx; font-weight: 600; color: #1F2937; }
.aggression-warn { font-size: 20rpx; color: #DC2626; background: #FEE2E2; padding: 4rpx 12rpx; border-radius: 8rpx; font-weight: 600; white-space: nowrap; }
.customer-name { font-size: 26rpx; color: #6B7280; }
.staff-name { font-size: 24rpx; color: #4F46E5; }
.card-footer { display: flex; justify-content: space-between; gap: 16rpx; padding-top: 12rpx; border-top: 1rpx solid #F3F4F6; align-items: flex-start; }
.services { font-size: 24rpx; color: #6B7280; flex: 1; line-height: 1.5; }
.amount { font-size: 28rpx; font-weight: bold; color: #4F46E5; }
</style>
