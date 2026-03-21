<template>
  <SideLayout>
  <view class="page">
    <view class="header">
      <text class="title">订单管理</text>
      <view class="btn-add" @click="goCreate">+ 开单</view>
    </view>

    <view class="search-bar">
      <input
        v-model="keyword"
        placeholder="搜索订单号 / 客户 / 猫咪名"
        class="search-input"
        confirm-type="search"
        @confirm="onSearch"
        @input="onSearchInput"
      />
      <view v-if="keyword" class="search-clear" @click="clearSearch">✕</view>
    </view>

    <view class="tabs">
      <view :class="['tab', activeStatus === -1 ? 'active' : '']" @click="switchTab(-1)">全部</view>
      <view :class="['tab', activeStatus === 0 ? 'active' : '']" @click="switchTab(0)">待付款</view>
      <view :class="['tab', activeStatus === 1 ? 'active' : '']" @click="switchTab(1)">已完成</view>
      <view :class="['tab', activeStatus === 3 ? 'active' : '']" @click="switchTab(3)">已退款</view>
    </view>

    <view v-if="loading" class="loading">加载中...</view>
    <view v-else-if="list.length === 0" class="empty">暂无订单</view>

    <view v-else class="list">
      <view class="card" v-for="item in list" :key="item.ID" @click="goDetail(item.ID)">
        <view class="card-top">
          <text class="order-no">{{ item.order_no }}</text>
          <view :class="['status', `s${item.status}`]">{{ statusMap[item.status] }}</view>
        </view>
        <view class="card-body">
          <text class="customer">{{ item.customer?.nickname || '-' }}</text>
          <text class="pet-name" v-if="item.pet">🐱 {{ item.pet.name }}</text>
          <text class="items-summary">{{ (item.items || []).map((i: any) => i.name).join(', ') }}</text>
        </view>
        <view class="card-footer">
          <text class="pay-method" v-if="item.pay_method">{{ payMethodMap[item.pay_method] || item.pay_method }}</text>
          <text class="amount">¥{{ item.pay_amount }}</text>
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
import { getOrderList } from '@/api/order'

const list = ref<any[]>([])
const loading = ref(true)
const activeStatus = ref(-1)
const keyword = ref('')
let searchTimer: ReturnType<typeof setTimeout> | null = null
const statusMap: Record<number, string> = { 0: '待付款', 1: '已完成', 2: '已取消', 3: '已退款' }
const payMethodMap: Record<string, string> = { wechat: '微信', alipay: '支付宝', cash: '现金', meituan: '美团' }

async function loadData() {
  loading.value = true
  try {
    const params: any = { page: 1, page_size: 50 }
    if (activeStatus.value >= 0) params.status = activeStatus.value
    if (keyword.value.trim()) params.keyword = keyword.value.trim()
    const res = await getOrderList(params)
    list.value = res.data.list || []
  } finally { loading.value = false }
}

function switchTab(s: number) { activeStatus.value = s; loadData() }
function onSearch() { loadData() }
function onSearchInput() {
  if (searchTimer) clearTimeout(searchTimer)
  searchTimer = setTimeout(() => loadData(), 400)
}
function clearSearch() { keyword.value = ''; loadData() }
function goCreate() { uni.navigateTo({ url: '/pages/order/create' }) }
function goDetail(id: number) { uni.navigateTo({ url: `/pages/order/detail?id=${id}` }) }

onMounted(loadData)
onShow(loadData)
</script>

<style scoped>
.page { padding: 24rpx; }
.header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 20rpx; }
.title { font-size: 36rpx; font-weight: bold; color: #1F2937; }
.btn-add { font-size: 28rpx; color: #fff; background: #4F46E5; padding: 12rpx 24rpx; border-radius: 12rpx; }
.search-bar { position: relative; margin-bottom: 16rpx; }
.search-input { background: #fff; border-radius: 12rpx; padding: 16rpx 60rpx 16rpx 24rpx; font-size: 26rpx; color: #1F2937; box-shadow: 0 2rpx 8rpx rgba(0,0,0,0.04); }
.search-clear { position: absolute; right: 20rpx; top: 50%; transform: translateY(-50%); font-size: 28rpx; color: #9CA3AF; padding: 8rpx; }
.tabs { display: flex; gap: 12rpx; margin-bottom: 24rpx; }
.tab { font-size: 24rpx; padding: 10rpx 20rpx; border-radius: 20rpx; background: #F3F4F6; color: #6B7280; }
.tab.active { background: #4F46E5; color: #fff; }
.loading, .empty { text-align: center; padding: 100rpx 0; color: #9CA3AF; font-size: 28rpx; }
.card { background: #fff; border-radius: 16rpx; padding: 24rpx; margin-bottom: 16rpx; box-shadow: 0 2rpx 8rpx rgba(0,0,0,0.04); }
.card-top { display: flex; justify-content: space-between; align-items: center; margin-bottom: 12rpx; }
.order-no { font-size: 24rpx; color: #9CA3AF; }
.status { font-size: 22rpx; padding: 6rpx 16rpx; border-radius: 16rpx; }
.s0 { background: #FEF3C7; color: #92400E; }
.s1 { background: #D1FAE5; color: #059669; }
.s2 { background: #F3F4F6; color: #6B7280; }
.s3 { background: #FEE2E2; color: #DC2626; }
.card-body { margin-bottom: 12rpx; }
.customer { font-size: 28rpx; font-weight: 600; color: #1F2937; }
.pet-name { font-size: 24rpx; color: #6B7280; margin-left: 12rpx; }
.items-summary { font-size: 24rpx; color: #6B7280; display: block; margin-top: 4rpx; }
.card-footer { display: flex; justify-content: space-between; padding-top: 12rpx; border-top: 1rpx solid #F3F4F6; }
.pay-method { font-size: 24rpx; color: #6B7280; }
.amount { font-size: 32rpx; font-weight: bold; color: #4F46E5; }
</style>
