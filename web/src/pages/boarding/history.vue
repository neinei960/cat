<template>
  <SideLayout>
    <view class="page">
      <view class="hero">
        <view>
          <text class="title">历史寄养记录</text>
          <text class="subtitle">查看已离店猫咪的寄养记录、家长、房型和备注摘要。</text>
        </view>
      </view>

      <view class="filter-card">
        <view class="filter-row">
          <picker mode="date" :value="filter.dateFrom" @change="onDateFromChange">
            <view class="filter-chip">
              <text class="filter-label">开始</text>
              <text class="filter-value single-line">{{ filter.dateFrom || '日期' }}</text>
            </view>
          </picker>
          <picker mode="date" :value="filter.dateTo" @change="onDateToChange">
            <view class="filter-chip">
              <text class="filter-label">结束</text>
              <text class="filter-value single-line">{{ filter.dateTo || '日期' }}</text>
            </view>
          </picker>
          <picker :range="cabinetNames" :value="selectedCabinetIndex" @change="onCabinetChange">
            <view class="filter-chip">
              <text class="filter-label">房型</text>
              <text class="filter-value single-line">{{ selectedCabinetLabel }}</text>
            </view>
          </picker>
        </view>
        <view class="filter-foot">
          <text class="filter-summary">{{ activeFilterSummary }}</text>
          <view v-if="hasActiveFilter" class="filter-reset" @click="resetFilter">重置</view>
        </view>
      </view>

      <view v-if="loading && list.length === 0" class="state">加载中...</view>
      <view v-else-if="list.length === 0" class="state">{{ hasActiveFilter ? '没有符合筛选条件的历史记录' : '还没有已结束的寄养记录' }}</view>

      <view v-else class="history-list">
        <view
          v-for="item in list"
          :key="item.ID"
          class="history-card"
          @click="goDetail(item.ID)"
        >
          <view class="card-head">
            <view class="head-copy">
              <text class="card-title">{{ petNames(item) }}</text>
              <text class="card-sub">{{ customerLabel(item) }}</text>
            </view>
            <text class="status-pill">已离店</text>
          </view>

          <view class="meta-list">
            <view class="meta-row">
              <text class="meta-label">日期</text>
              <text class="meta-value">{{ item.check_in_at }} → {{ displayCheckOut(item) }}</text>
            </view>
            <view class="meta-row">
              <text class="meta-label">房型</text>
              <text class="meta-value">{{ roomSummary(item) }}</text>
            </view>
          </view>

          <view v-if="remarkSummary(item.remark)" class="remark-row">
            <text class="remark-label">备注</text>
            <text class="remark-value">{{ remarkSummary(item.remark) }}</text>
          </view>

          <view class="card-foot">
            <text class="foot-copy">{{ roomCountCopy(item) }}</text>
            <text class="foot-link">查看详情</text>
          </view>
        </view>
      </view>

      <view v-if="loadingMore" class="load-more">
        <text class="load-more-text">加载中...</text>
      </view>
      <view v-else-if="hasMore" class="load-more" @click="loadMore">
        <text class="load-more-text">上滑加载更多</text>
      </view>
      <view v-else-if="list.length > 0" class="load-more">
        <text class="load-more-done">已加载全部 {{ total }} 条历史记录</text>
      </view>
    </view>
  </SideLayout>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { onReachBottom, onShow } from '@dcloudio/uni-app'
import SideLayout from '@/components/SideLayout.vue'
import { getBoardingCabinets, getBoardingOrders } from '@/api/boarding'
import {
  getBoardingHistoryCustomerLabel,
  getBoardingHistoryPetNames,
  getBoardingHistoryRemarkSummary,
  getBoardingHistoryRoomSummary,
} from '@/utils/boarding-history'

const PAGE_SIZE = 20

const list = ref<BoardingOrder[]>([])
const total = ref(0)
const currentPage = ref(1)
const loading = ref(false)
const loadingMore = ref(false)
const hasMore = ref(true)
const cabinets = ref<BoardingCabinet[]>([])
const filter = ref({
  dateFrom: '',
  dateTo: '',
  cabinetId: 0,
})

const cabinetNames = computed(() => ['全部房型', ...cabinets.value.map((item) => item.cabinet_type)])
const selectedCabinetIndex = computed(() => {
  const index = cabinets.value.findIndex((item) => item.ID === filter.value.cabinetId)
  return index >= 0 ? index + 1 : 0
})
const selectedCabinetLabel = computed(() => {
  if (!filter.value.cabinetId) return '全部房型'
  return cabinets.value.find((item) => item.ID === filter.value.cabinetId)?.cabinet_type || '全部房型'
})
const hasActiveFilter = computed(() => !!(filter.value.dateFrom || filter.value.dateTo || filter.value.cabinetId))
const activeFilterSummary = computed(() => {
  const segments = []
  if (filter.value.dateFrom || filter.value.dateTo) {
    segments.push(`${filter.value.dateFrom || '开始'} ~ ${filter.value.dateTo || '结束'}`)
  }
  if (filter.value.cabinetId) {
    segments.push(selectedCabinetLabel.value)
  }
  return segments.length ? `已筛选：${segments.join(' · ')}` : '按日期和房型筛选'
})

function customerLabel(order: BoardingOrder) {
  return getBoardingHistoryCustomerLabel(order)
}

function petNames(order: BoardingOrder) {
  return getBoardingHistoryPetNames(order)
}

function roomSummary(order: BoardingOrder) {
  return getBoardingHistoryRoomSummary(order)
}

function remarkSummary(remark?: string) {
  return getBoardingHistoryRemarkSummary(remark, 30)
}

function displayCheckOut(order: BoardingOrder) {
  return order.actual_check_out_at || order.check_out_at
}

function roomCountCopy(order: BoardingOrder) {
  const roomCount = order.rooms?.length || 0
  if (roomCount > 1) return `${roomCount} 个房间`
  return `${roomSummary(order)} · ${displayCheckOut(order)}`
}

function goDetail(id: number) {
  uni.navigateTo({ url: `/pages/boarding/detail?id=${id}` })
}

function normalizeDateRange() {
  if (filter.value.dateFrom && filter.value.dateTo && filter.value.dateFrom > filter.value.dateTo) {
    filter.value.dateTo = filter.value.dateFrom
  }
}

function onDateFromChange(e: any) {
  filter.value.dateFrom = e.detail.value
  normalizeDateRange()
  loadData()
}

function onDateToChange(e: any) {
  filter.value.dateTo = e.detail.value
  if (filter.value.dateFrom && filter.value.dateTo < filter.value.dateFrom) {
    filter.value.dateFrom = filter.value.dateTo
  }
  loadData()
}

function onCabinetChange(e: any) {
  const index = Number(e.detail.value || 0)
  filter.value.cabinetId = index > 0 ? cabinets.value[index - 1]?.ID || 0 : 0
  loadData()
}

function resetFilter() {
  filter.value = {
    dateFrom: '',
    dateTo: '',
    cabinetId: 0,
  }
  loadData()
}

async function loadCabinets() {
  const res = await getBoardingCabinets()
  cabinets.value = (res.data || []).filter((item) => item.status === 'enabled')
}

async function loadData() {
  loading.value = true
  currentPage.value = 1
  try {
    const res = await getBoardingOrders({
      page: 1,
      page_size: PAGE_SIZE,
      status: 'checked_out',
      date_from: filter.value.dateFrom || undefined,
      date_to: filter.value.dateTo || undefined,
      cabinet_id: filter.value.cabinetId || undefined,
    })
    list.value = res.data.list || []
    total.value = res.data.total || list.value.length
    hasMore.value = list.value.length < total.value
  } finally {
    loading.value = false
  }
}

async function loadMore() {
  if (loadingMore.value || !hasMore.value) return
  loadingMore.value = true
  try {
    currentPage.value += 1
    const res = await getBoardingOrders({
      page: currentPage.value,
      page_size: PAGE_SIZE,
      status: 'checked_out',
      date_from: filter.value.dateFrom || undefined,
      date_to: filter.value.dateTo || undefined,
      cabinet_id: filter.value.cabinetId || undefined,
    })
    const nextList = res.data.list || []
    list.value = [...list.value, ...nextList]
    total.value = res.data.total || list.value.length
    hasMore.value = list.value.length < total.value
  } finally {
    loadingMore.value = false
  }
}

onShow(async () => {
  await loadCabinets()
  await loadData()
})
onReachBottom(loadMore)
</script>

<style scoped>
.page {
  padding: 24rpx;
  display: flex;
  flex-direction: column;
  gap: 20rpx;
}
.hero {
  display: flex;
  justify-content: space-between;
  gap: 16rpx;
  align-items: flex-start;
}
.title {
  display: block;
  font-size: 36rpx;
  font-weight: 700;
  color: #111827;
}
.subtitle {
  display: block;
  margin-top: 10rpx;
  font-size: 24rpx;
  color: #6b7280;
  line-height: 1.6;
}
.filter-card {
  border-radius: 22rpx;
  padding: 16rpx;
  background: rgba(255, 255, 255, 0.96);
  border: 1rpx solid rgba(226, 232, 240, 0.95);
  box-shadow: 0 10rpx 24rpx rgba(15, 23, 42, 0.05);
}
.filter-row {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 12rpx;
}
.filter-chip {
  min-height: 68rpx;
  border-radius: 16rpx;
  padding: 12rpx 14rpx;
  background: #f8fafc;
  border: 1rpx solid #e5e7eb;
  box-sizing: border-box;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 10rpx;
}
.filter-label,
.meta-label,
.remark-label {
  font-size: 20rpx;
  color: #94a3b8;
  flex-shrink: 0;
}
.filter-value,
.meta-value {
  font-size: 22rpx;
  color: #111827;
  font-weight: 700;
  line-height: 1.4;
}
.filter-value.single-line {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  text-align: right;
}
.filter-foot {
  margin-top: 10rpx;
  display: flex;
  justify-content: space-between;
  gap: 16rpx;
  align-items: center;
}
.filter-summary {
  font-size: 20rpx;
  color: #64748b;
  line-height: 1.5;
}
.filter-reset {
  padding: 8rpx 16rpx;
  border-radius: 999rpx;
  background: #eef2ff;
  color: #4f46e5;
  font-size: 20rpx;
  font-weight: 700;
  white-space: nowrap;
}
.state {
  text-align: center;
  padding: 120rpx 24rpx;
  color: #9ca3af;
  font-size: 28rpx;
}
.history-list {
  display: flex;
  flex-direction: column;
  gap: 14rpx;
}
.history-card {
  background: rgba(255, 255, 255, 0.96);
  border-radius: 18rpx;
  padding: 16rpx 16rpx 14rpx;
  border: 1rpx solid rgba(226, 232, 240, 0.95);
  box-shadow: 0 8rpx 18rpx rgba(15, 23, 42, 0.045);
}
.card-head,
.card-foot {
  display: flex;
  justify-content: space-between;
  gap: 12rpx;
  align-items: center;
}
.head-copy {
  min-width: 0;
}
.card-title {
  display: block;
  font-size: 27rpx;
  font-weight: 800;
  color: #111827;
  line-height: 1.4;
}
.card-sub {
  display: block;
  margin-top: 4rpx;
  font-size: 21rpx;
  color: #64748b;
  line-height: 1.5;
}
.status-pill {
  padding: 6rpx 12rpx;
  border-radius: 999rpx;
  background: #f3f4f6;
  color: #6b7280;
  font-size: 20rpx;
  font-weight: 700;
  white-space: nowrap;
}
.meta-list {
  display: flex;
  flex-direction: column;
  gap: 6rpx;
  margin-top: 10rpx;
}
.meta-row,
.remark-row {
  display: flex;
  justify-content: space-between;
  gap: 14rpx;
  align-items: center;
  min-width: 0;
}
.meta-value,
.remark-value {
  min-width: 0;
  text-align: right;
}
.remark-row {
  margin-top: 8rpx;
}
.remark-value {
  font-size: 21rpx;
  line-height: 1.5;
  color: #475569;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.card-foot {
  margin-top: 8rpx;
}
.foot-copy {
  font-size: 20rpx;
  color: #94a3b8;
  line-height: 1.4;
}
.foot-link {
  font-size: 20rpx;
  font-weight: 700;
  color: #4f46e5;
}
.load-more {
  padding: 18rpx 0 12rpx;
  text-align: center;
}
.load-more-text,
.load-more-done {
  font-size: 22rpx;
  color: #9ca3af;
}
@media (max-width: 768px) {
  .filter-row {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
  .filter-row > :last-child {
    grid-column: 1 / -1;
  }
  .filter-foot {
    flex-direction: column;
    align-items: flex-start;
  }
  .meta-row,
  .remark-row {
    align-items: flex-start;
    gap: 10rpx;
  }
  .meta-value,
  .remark-value {
    flex: 1;
  }
}
</style>
