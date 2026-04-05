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
        placeholder="搜索订单号 / 客户 / 猫咪 / 商品 / 服务"
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
      <view :class="['tab', filter.status === 1 ? 'active' : '']" @click="filter.status = 1; loadData()">已支付</view>
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
      <text v-if="filter.productKeyword.trim()" class="filter-tag">商品: {{ filter.productKeyword }} <text @click="filter.productKeyword = ''; loadData()">✕</text></text>
    </view>

    <FilterPanel
      :visible="showFilter"
      :filter="filter"
      :status-options="orderStatusOptions"
      status-label="订单状态"
      :pay-methods="orderPayMethods"
      :show-product-keyword="true"
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
      <view
        class="card"
        v-for="item in list"
        :key="item.ID"
        @click="handleCardClick(item.ID)"
        @longpress="openCardActions(item)"
        @touchstart="startCardLongPress(item)"
        @touchend="clearCardLongPress"
        @touchcancel="clearCardLongPress"
        @touchmove="clearCardLongPress"
      >
        <view class="card-top">
          <text class="order-no">{{ item.order_no }}</text>
          <view class="card-top-right">
            <view :class="['status', `s${item.status}`]">{{ statusMap[item.status] }}</view>
            <view
              v-if="isDesktopInteraction && item.status === 0"
              class="card-action-btn primary"
              @click.stop="editOrder(item)"
            >修改</view>
            <view
              v-else-if="isDesktopInteraction && (item.status === 2 || item.status === 3)"
              class="card-action-btn danger"
              @click.stop="confirmDeleteOrder(item)"
            >删除</view>
          </view>
        </view>
        <view class="card-body">
          <text class="customer">{{ getOrderTitle(item) }}</text>
          <view class="order-meta">
            <text class="order-time">{{ item.CreatedAt?.substring(0, 16).replace('T', ' ') }}</text>
            <text v-if="item.appointment?.date" class="appointment-date">预约 {{ item.appointment.date }}</text>
          </view>
          <text class="items-summary">{{ (item.items || []).map((i: any) => i.name).join(', ') }}</text>
        </view>
        <view class="card-footer">
          <view class="footer-left">
            <text v-if="item.order_kind" class="order-kind">{{ getOrderKindLabel(item.order_kind) }}</text>
            <text
              v-if="item.pay_method"
              :class="['pay-method', `pay-method-${resolvePayMethodKey(item.pay_method)}`]"
            >{{ payMethodMap[item.pay_method] || item.pay_method }}</text>
            <text
              v-if="item.pay_method && payMethodBadgeMap[resolvePayMethodKey(item.pay_method)]"
              :class="['pay-method-badge', `pay-method-badge-${resolvePayMethodKey(item.pay_method)}`]"
            >{{ payMethodBadgeMap[resolvePayMethodKey(item.pay_method)] }}</text>
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
import { onLoad, onShow } from '@dcloudio/uni-app'
import { deleteOrder, getOrderList } from '@/api/order'
import { getStaffList } from '@/api/staff'
import { getCategoryTree } from '@/api/service-category'
import { useDesktopInteraction } from '@/utils/interaction'

const list = ref<any[]>([])
const loading = ref(true)
const keyword = ref('')
const showFilter = ref(false)
const staffList = ref<any[]>([])
const categories = ref<any[]>([])
let searchTimer: ReturnType<typeof setTimeout> | null = null
let suppressCardClickUntil = 0
let cardLongPressTimer: ReturnType<typeof setTimeout> | null = null
let cardLongPressTriggered = false
const { isDesktopInteraction } = useDesktopInteraction()
const statusMap: Record<number, string> = { 0: '待付款', 1: '已支付', 2: '已取消', 3: '已退款' }
const payMethodMap: Record<string, string> = {
  qrcode: '扫码',
  wechat: '微信',
  meituan: '美团',
  balance: '会员余额',
  other: '其他',
  alipay: '扫码',
  cash: '其他',
  card: '会员余额',
}
const payMethodBadgeMap: Record<string, string> = {
  qrcode: '扫码收款',
  wechat: '微信收款',
  meituan: '美团订单',
  balance: '会员扣款',
  other: '其他收款',
}

const orderStatusOptions = [
  { value: 0, label: '待付款' },
  { value: 1, label: '已支付' },
  { value: 2, label: '已取消' },
  { value: 3, label: '已退款' },
]
const orderPayMethods = [
  { value: 'qrcode', label: '扫码' },
  { value: 'wechat', label: '微信' },
  { value: 'meituan', label: '美团' },
  { value: 'balance', label: '会员余额' },
  { value: 'other', label: '其他' },
]

const filter = reactive({
  dateFrom: '',
  dateTo: '',
  status: -1,
  staffId: 0,
  payMethod: '',
  categoryId: 0,
  productKeyword: '',
})

const activeFilterCount = computed(() => {
  let c = 0
  if (filter.dateFrom || filter.dateTo) c++
  if (filter.staffId > 0) c++
  if (filter.payMethod) c++
  if (filter.categoryId > 0) c++
  if (filter.productKeyword.trim()) c++
  return c
})

function getStaffName(id: number) {
  return staffList.value.find((s: any) => s.ID === id)?.name || '未知'
}

function toggleMeituan() {
  filter.payMethod = filter.payMethod === 'meituan' ? '' : 'meituan'
  loadData()
}

function getOrderTitle(item: any) {
  const customerName = item.customer?.nickname || '散客'
  if (item.pet_summary) {
    return `${customerName} · 🐱${item.pet_summary}`
  }
  if (item.pet?.name) {
    return `${customerName} · 🐱${item.pet.name}`
  }
  if (item.order_kind === 'product') {
    return `${customerName} · 商品零售`
  }
  return customerName
}

function getOrderKindLabel(kind: string) {
  if (kind === 'mixed') return '服务 + 商品'
  if (kind === 'product') return '商品零售'
  if (kind === 'feeding') return '上门喂养'
  return '服务订单'
}

function resolvePayMethodKey(method: string) {
  if (method === 'alipay') return 'qrcode'
  if (method === 'cash') return 'other'
  if (method === 'card') return 'balance'
  return method
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
    if (filter.productKeyword.trim()) params.product_keyword = filter.productKeyword.trim()
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

function getEditUrl(item: any) {
  const isBatchOrder = !!item.appointment_id && (!item.pet_id || (item.pet_groups?.length || 0) > 1)
  return isBatchOrder
    ? `/pages/order/batch-create?appointment_id=${item.appointment_id}&order_id=${item.ID}`
    : `/pages/order/create?order_id=${item.ID}`
}

function editOrder(item: any) {
  uni.navigateTo({ url: getEditUrl(item) })
}

function confirmDeleteOrder(item: any) {
  uni.showModal({
    title: '删除订单',
    content: `确认删除订单 ${item.order_no} 吗？`,
    success: async (res) => {
      if (!res.confirm) return
      try {
        await deleteOrder(item.ID)
        uni.showToast({ title: '已删除', icon: 'success' })
        await loadData()
      } catch (error: any) {
        uni.showToast({ title: error?.message || '删除失败', icon: 'none' })
      }
    },
  })
}

function handleCardClick(id: number) {
  if (cardLongPressTriggered) {
    cardLongPressTriggered = false
    return
  }
  clearCardLongPress()
  if (Date.now() < suppressCardClickUntil) return
  goDetail(id)
}

function startCardLongPress(item: any) {
  clearCardLongPress()
  cardLongPressTriggered = false
  cardLongPressTimer = setTimeout(() => {
    cardLongPressTriggered = true
    openCardActions(item)
  }, 450)
}

function clearCardLongPress() {
  if (cardLongPressTimer) {
    clearTimeout(cardLongPressTimer)
    cardLongPressTimer = null
  }
}

function openCardActions(item: any) {
  clearCardLongPress()
  suppressCardClickUntil = Date.now() + 800
  if (item.status === 0) {
    uni.showActionSheet({
      itemList: ['修改订单'],
      success: ({ tapIndex }) => {
        if (tapIndex !== 0) return
        editOrder(item)
      },
    })
    return
  }
  if (item.status !== 2 && item.status !== 3) {
    uni.showToast({ title: '当前状态不可操作', icon: 'none' })
    return
  }
  uni.showActionSheet({
    itemList: ['删除订单'],
    success: ({ tapIndex }) => {
      if (tapIndex !== 0) return
      confirmDeleteOrder(item)
    },
  })
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

onLoad((query) => {
  const date = typeof query?.date === 'string' ? query.date.trim() : ''
  if (date) {
    filter.dateFrom = date
    filter.dateTo = date
  }
})

onMounted(() => { loadData(); loadFilterOptions() })
onShow(loadData)
</script>

<style scoped>
.page { padding: 24rpx; }
.header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 20rpx; gap: 16rpx; }
.header-right { display: flex; gap: 12rpx; align-items: center; flex-shrink: 0; }
.title { font-size: 36rpx; font-weight: bold; color: #1F2937; }
.btn-add,
.btn-filter {
  min-height: 72rpx;
  padding: 0 24rpx;
  border-radius: 16rpx;
  font-size: 26rpx;
  font-weight: 700;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  box-sizing: border-box;
}
.btn-add { color: #fff; background: linear-gradient(135deg, #4F46E5, #6366F1); box-shadow: 0 10rpx 24rpx rgba(79, 70, 229, 0.2); }
.btn-filter { gap: 6rpx; color: #374151; background: #F8FAFC; border: 2rpx solid #E2E8F0; }
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
.card-top { display: flex; justify-content: space-between; align-items: center; margin-bottom: 12rpx; gap: 12rpx; }
.card-top-right { display: flex; align-items: center; gap: 10rpx; flex-shrink: 0; }
.order-no { font-size: 24rpx; color: #9CA3AF; }
.status { font-size: 22rpx; padding: 6rpx 16rpx; border-radius: 16rpx; }
.s0 { background: #FEF3C7; color: #92400E; }
.s1 { background: #D1FAE5; color: #059669; }
.s2 { background: #F3F4F6; color: #6B7280; }
.s3 { background: #FEE2E2; color: #DC2626; }
.card-action-btn { padding: 8rpx 16rpx; border-radius: 999rpx; font-size: 22rpx; font-weight: 700; }
.card-action-btn.primary { color: #4F46E5; background: #EEF2FF; }
.card-action-btn.danger { color: #DC2626; background: #FEF2F2; }
.card-body { margin-bottom: 12rpx; }
.customer { font-size: 28rpx; font-weight: 600; color: #1F2937; display: block; }
.customer-pet { font-size: 24rpx; font-weight: 400; color: #6B7280; }
.order-meta { display: flex; align-items: center; justify-content: space-between; gap: 12rpx; margin-top: 4rpx; }
.order-time { font-size: 22rpx; color: #9CA3AF; display: block; }
.appointment-date { font-size: 22rpx; color: #4F46E5; background: #EEF2FF; border-radius: 999rpx; padding: 4rpx 12rpx; flex-shrink: 0; }
.items-summary { font-size: 24rpx; color: #6B7280; display: block; margin-top: 4rpx; }
.card-footer { display: flex; justify-content: space-between; align-items: center; padding-top: 12rpx; border-top: 1rpx solid #F3F4F6; }
.footer-left { display: flex; align-items: center; gap: 10rpx; flex-wrap: wrap; }
.order-kind { font-size: 20rpx; color: #4F46E5; background: #EEF2FF; padding: 4rpx 10rpx; border-radius: 8rpx; }
.pay-method {
  font-size: 24rpx;
  font-weight: 700;
  padding: 6rpx 14rpx;
  border-radius: 999rpx;
  border: 1rpx solid transparent;
}
.pay-method-qrcode {
  color: #2563EB;
  background: #EFF6FF;
  border-color: #BFDBFE;
}
.pay-method-wechat {
  color: #15803D;
  background: #F0FDF4;
  border-color: #BBF7D0;
}
.pay-method-meituan {
  color: #EA580C;
  background: #FFF7ED;
  border-color: #FED7AA;
}
.pay-method-balance {
  color: #7C3AED;
  background: #F5F3FF;
  border-color: #DDD6FE;
}
.pay-method-other {
  color: #475569;
  background: #F8FAFC;
  border-color: #CBD5E1;
}
.pay-method-badge {
  font-size: 20rpx;
  font-weight: 600;
  padding: 4rpx 10rpx;
  border-radius: 8rpx;
  border: 1rpx solid transparent;
}
.pay-method-badge-qrcode {
  color: #2563EB;
  background: #EFF6FF;
  border-color: #BFDBFE;
}
.pay-method-badge-wechat {
  color: #15803D;
  background: #F0FDF4;
  border-color: #BBF7D0;
}
.pay-method-badge-meituan {
  color: #EA580C;
  background: #FFF7ED;
  border-color: #FED7AA;
}
.pay-method-badge-balance {
  color: #7C3AED;
  background: #F5F3FF;
  border-color: #DDD6FE;
}
.pay-method-badge-other {
  color: #475569;
  background: #F8FAFC;
  border-color: #CBD5E1;
}
.amount { font-size: 36rpx; font-weight: 700; color: #4F46E5; }

@media (max-width: 768px) {
  .header {
    align-items: flex-start;
  }
  .header-right {
    width: auto;
  }
}
</style>
