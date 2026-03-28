<template>
  <SideLayout>
  <view class="page">
    <view class="header">
      <text class="title">订单管理</text>
      <view class="header-right">
        <view class="btn-filter" @click="showFilter = true">
          <text>筛选</text>
          <text v-if="activeFilterCount > 0" class="filter-badge">{{ activeFilterCount }}</text>
        </view>
        <view class="btn-add" @click="goCreate">+ 开单</view>
      </view>
    </view>

    <view class="search-bar">
      <input
        v-model="keyword"
        placeholder="搜索订单号 / 客户 / 猫咪名 / 服务项目"
        class="search-input"
        confirm-type="search"
        @confirm="onSearch"
        @input="onSearchInput"
      />
      <view v-if="keyword" class="search-clear" @click="clearSearch">✕</view>
    </view>

    <!-- 快捷标签 -->
    <view class="tabs">
      <view :class="['tab', filter.status === -1 ? 'active' : '']" @click="filter.status = -1; loadData()">全部</view>
      <view :class="['tab', filter.status === 0 ? 'active' : '']" @click="filter.status = 0; loadData()">待付款</view>
      <view :class="['tab', filter.status === 1 ? 'active' : '']" @click="filter.status = 1; loadData()">已完成</view>
      <view :class="['tab', filter.status === 3 ? 'active' : '']" @click="filter.status = 3; loadData()">已退款</view>
      <view class="tab-sep"></view>
      <view :class="['tab tab-meituan', filter.payMethod === 'meituan' ? 'active' : '']" @click="toggleMeituan">美团</view>
    </view>

    <!-- 活跃筛选条件提示 -->
    <view class="active-filters" v-if="activeFilterCount > 0">
      <text class="active-filters-text">已筛选: </text>
      <text v-if="filter.dateFrom || filter.dateTo" class="filter-tag">{{ filter.dateFrom || '...' }} ~ {{ filter.dateTo || '...' }} <text @click="filter.dateFrom = ''; filter.dateTo = ''; loadData()">✕</text></text>
      <text v-if="filter.staffId > 0" class="filter-tag">{{ getStaffName(filter.staffId) }} <text @click="filter.staffId = 0; loadData()">✕</text></text>
      <text v-if="filter.payMethod" class="filter-tag">{{ payMethodMap[filter.payMethod] || filter.payMethod }} <text @click="filter.payMethod = ''; loadData()">✕</text></text>
    </view>

    <FilterPanel
      :visible="showFilter"
      :filter="filter"
      :status-options="orderStatusOptions"
      status-label="订单状态"
      :pay-methods="orderPayMethods"
      :staff-list="staffList"
      :categories="categories"
      @close="showFilter = false"
      @confirm="onFilterConfirm"
    />

    <view v-if="loading" class="loading">
      <text class="loading-icon">🧾</text>
      <text class="loading-text">正在加载订单...</text>
    </view>
    <view v-else-if="list.length === 0" class="empty">
      <text class="empty-icon">🧾</text>
      <text class="empty-title">{{ keyword || filter.status >= 0 ? '没有找到匹配的订单' : '还没有订单' }}</text>
      <text class="empty-desc">{{ keyword || filter.status >= 0 ? '试试调整筛选条件' : '点击右上角开单吧' }}</text>
    </view>

    <view v-else class="list">
      <view class="card" v-for="item in list" :key="item.ID" @click="goDetail(item.ID)">
        <view class="card-top">
          <text class="order-no">{{ item.order_no }}</text>
          <view :class="['status', `s${item.status}`]">{{ statusMap[item.status] }}</view>
        </view>
        <view class="card-body">
          <text class="customer">{{ item.customer?.nickname || '-' }}<text v-if="item.pet" class="customer-pet"> · 🐱{{ item.pet.name }}</text></text>
          <text class="order-time">{{ item.CreatedAt?.substring(0, 16).replace('T', ' ') }}</text>
          <text class="items-summary">{{ (item.items || []).map((i: any) => i.name).join(', ') }}</text>
        </view>
        <view class="card-footer">
          <view class="footer-left">
            <text class="pay-method" v-if="item.pay_method">{{ payMethodMap[item.pay_method] || item.pay_method }}</text>
            <text class="meituan-badge" v-if="item.pay_method === 'meituan'">美团订单</text>
          </view>
          <text class="amount">¥{{ item.pay_amount }}</text>
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
import { getOrderList } from '@/api/order'
import { getStaffList } from '@/api/staff'
import { getCategoryTree } from '@/api/service-category'

const list = ref<any[]>([])
const loading = ref(true)
const keyword = ref('')
const showFilter = ref(false)
const staffList = ref<any[]>([])
const categories = ref<any[]>([])
let searchTimer: ReturnType<typeof setTimeout> | null = null
const statusMap: Record<number, string> = { 0: '待付款', 1: '已完成', 2: '已取消', 3: '已退款' }
const payMethodMap: Record<string, string> = { wechat: '扫码', alipay: '扫码', cash: '现金', meituan: '美团', card: '会员卡', balance: '余额' }

const orderStatusOptions = [
  { value: 0, label: '待付款' },
  { value: 1, label: '已完成' },
  { value: 2, label: '已取消' },
  { value: 3, label: '已退款' },
]
const orderPayMethods = [
  { value: 'wechat', label: '扫码' },
  { value: 'cash', label: '现金' },
  { value: 'meituan', label: '美团' },
  { value: 'balance', label: '余额' },
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
  if (filter.payMethod) c++
  if (filter.categoryId > 0) c++
  return c
})

function getStaffName(id: number) {
  return staffList.value.find((s: any) => s.ID === id)?.name || '未知'
}

function toggleMeituan() {
  filter.payMethod = filter.payMethod === 'meituan' ? '' : 'meituan'
  loadData()
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
    if (keyword.value.trim()) params.keyword = keyword.value.trim()
    if (filter.dateFrom) params.date_from = filter.dateFrom
    if (filter.dateTo) params.date_to = filter.dateTo
    if (filter.staffId > 0) params.staff_id = filter.staffId
    if (filter.payMethod) params.pay_method = filter.payMethod
    const res = await getOrderList(params)
    list.value = res.data.list || []
  } finally { loading.value = false }
}

function onSearch() { loadData() }
function onSearchInput() {
  if (searchTimer) clearTimeout(searchTimer)
  searchTimer = setTimeout(() => loadData(), 400)
}
function clearSearch() { keyword.value = ''; loadData() }
function goCreate() { uni.navigateTo({ url: '/pages/order/create' }) }
function goDetail(id: number) { uni.navigateTo({ url: `/pages/order/detail?id=${id}` }) }

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
.search-bar { position: relative; margin-bottom: 16rpx; }
.search-input { background: #fff; border-radius: 12rpx; padding: 16rpx 60rpx 16rpx 24rpx; font-size: 26rpx; color: #1F2937; box-shadow: 0 2rpx 8rpx rgba(0,0,0,0.04); }
.search-clear { position: absolute; right: 20rpx; top: 50%; transform: translateY(-50%); font-size: 28rpx; color: #9CA3AF; padding: 8rpx; }
.tabs { display: flex; gap: 12rpx; margin-bottom: 24rpx; align-items: center; flex-wrap: wrap; }
.tab { font-size: 24rpx; padding: 10rpx 20rpx; border-radius: 20rpx; background: #F3F4F6; color: #6B7280; }
.tab.active { background: #4F46E5; color: #fff; }
.tab-sep { width: 1rpx; height: 28rpx; background: #E5E7EB; }
.tab-meituan { background: #FFF7ED; color: #EA580C; border: 1rpx solid #FED7AA; }
.tab-meituan.active { background: #EA580C; color: #fff; border-color: #EA580C; }
.loading, .empty { display: flex; flex-direction: column; align-items: center; padding: 120rpx 0; gap: 16rpx; }
.loading-icon, .empty-icon { font-size: 64rpx; }
.loading-text { font-size: 26rpx; color: #9CA3AF; }
.loading-icon { animation: bounce 1s infinite alternate; }
.empty-title { font-size: 30rpx; font-weight: 500; color: #4B5563; }
.empty-desc { font-size: 24rpx; color: #9CA3AF; }
@keyframes bounce { from { transform: translateY(0); } to { transform: translateY(-12rpx); } }
.card { background: #fff; border-radius: 16rpx; padding: 24rpx; margin-bottom: 16rpx; box-shadow: 0 2rpx 8rpx rgba(0,0,0,0.04); }
.card-top { display: flex; justify-content: space-between; align-items: center; margin-bottom: 12rpx; }
.order-no { font-size: 24rpx; color: #9CA3AF; }
.status { font-size: 22rpx; padding: 6rpx 16rpx; border-radius: 16rpx; }
.s0 { background: #FEF3C7; color: #92400E; }
.s1 { background: #D1FAE5; color: #059669; }
.s2 { background: #F3F4F6; color: #6B7280; }
.s3 { background: #FEE2E2; color: #DC2626; }
.card-body { margin-bottom: 12rpx; }
.customer { font-size: 28rpx; font-weight: 600; color: #1F2937; display: block; }
.customer-pet { font-size: 24rpx; font-weight: 400; color: #6B7280; }
.order-time { font-size: 22rpx; color: #9CA3AF; display: block; margin-top: 4rpx; }
.items-summary { font-size: 24rpx; color: #6B7280; display: block; margin-top: 4rpx; }
.card-footer { display: flex; justify-content: space-between; align-items: center; padding-top: 12rpx; border-top: 1rpx solid #F3F4F6; }
.footer-left { display: flex; align-items: center; gap: 8rpx; }
.pay-method { font-size: 24rpx; color: #6B7280; }
.meituan-badge { font-size: 20rpx; color: #EA580C; background: #FFF7ED; padding: 4rpx 10rpx; border-radius: 8rpx; border: 1rpx solid #FED7AA; }
.amount { font-size: 36rpx; font-weight: 700; color: #4F46E5; }
</style>
