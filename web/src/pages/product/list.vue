<template>
  <SideLayout>
  <view class="page">
    <view class="header">
      <text class="title">商品管理</text>
      <view class="header-btns">
        <view class="btn-secondary" @click="goCategory">分类管理</view>
        <view class="btn-add" @click="goCreate">+ 新增商品</view>
      </view>
    </view>

    <view class="search-bar">
      <input
        v-model="keyword"
        placeholder="搜索商品名 / 品牌"
        class="search-input"
        confirm-type="search"
        @confirm="onSearch"
        @input="onSearchInput"
      />
      <view v-if="keyword" class="search-clear" @click="clearSearch">✕</view>
    </view>

    <view class="tabs">
      <view :class="['tab', activeCategoryId === 0 ? 'active' : '']" @click="switchCategory(0)">全部</view>
      <view
        v-for="cat in categories"
        :key="cat.ID"
        :class="['tab', activeCategoryId === cat.ID ? 'active' : '']"
        @click="switchCategory(cat.ID)"
      >{{ cat.name }}</view>
    </view>

    <view v-if="loading" class="loading">加载中...</view>
    <view v-else-if="list.length === 0" class="empty">暂无商品</view>

    <view v-else class="list">
      <view
        class="card"
        v-for="item in list"
        :key="item.ID"
        @click="goEdit(item.ID)"
        @longpress="onLongPress(item)"
      >
        <view class="card-top">
          <text class="product-name">{{ item.name }}</text>
          <view :class="['badge', item.status === 1 ? 'badge-on' : 'badge-off']">
            {{ item.status === 1 ? '可售' : '已下架' }}
          </view>
        </view>
        <view class="card-bottom">
          <view class="card-bottom-left">
            <view v-if="item.category_name" class="cat-tag">{{ item.category_name }}</view>
            <text v-if="item.brand" class="brand-text">{{ item.brand }}</text>
          </view>
          <text class="price">{{ formatPrice(item) }}</text>
        </view>
      </view>
    </view>
  </view>
  </SideLayout>
</template>

<script setup lang="ts">
import SideLayout from '@/components/SideLayout.vue'
import { ref } from 'vue'
import { onShow } from '@dcloudio/uni-app'
import { getProductList, getProductCategories, deleteProduct } from '@/api/product'

const list = ref<any[]>([])
const categories = ref<any[]>([])
const loading = ref(true)
const activeCategoryId = ref(0)
const keyword = ref('')
let searchTimer: ReturnType<typeof setTimeout> | null = null

function formatPrice(item: any): string {
  const skus = item.skus || []
  if (!skus.length) return item.price ? `¥${item.price}` : '-'
  if (skus.length === 1) return `¥${skus[0].price}`
  const prices = skus.map((s: any) => s.price).filter((p: any) => p != null)
  if (!prices.length) return '-'
  const min = Math.min(...prices)
  const max = Math.max(...prices)
  return min === max ? `¥${min}` : `¥${min}-${max}`
}

async function loadCategories() {
  try {
    const res = await getProductCategories()
    categories.value = Array.isArray(res.data) ? res.data : []
  } catch {}
}

async function loadData() {
  loading.value = true
  try {
    const params: any = { page: 1, page_size: 100 }
    if (activeCategoryId.value) params.category_id = activeCategoryId.value
    if (keyword.value.trim()) params.keyword = keyword.value.trim()
    const res = await getProductList(params)
    list.value = res.data.list || []
  } finally { loading.value = false }
}

function switchCategory(id: number) { activeCategoryId.value = id; loadData() }
function onSearch() { loadData() }
function onSearchInput() {
  if (searchTimer) clearTimeout(searchTimer)
  searchTimer = setTimeout(() => loadData(), 400)
}
function clearSearch() { keyword.value = ''; loadData() }
function goCreate() { uni.navigateTo({ url: '/pages/product/edit' }) }
function goEdit(id: number) { uni.navigateTo({ url: `/pages/product/edit?id=${id}` }) }
function goCategory() { uni.navigateTo({ url: '/pages/product/category' }) }

function onLongPress(item: any) {
  uni.showActionSheet({
    itemList: ['删除'],
    success: (res) => {
      if (res.tapIndex === 0) {
        uni.showModal({
          title: '确认删除',
          content: `确认删除商品「${item.name}」？`,
          success: async (r) => {
            if (r.confirm) {
              await deleteProduct(item.ID)
              uni.showToast({ title: '已删除', icon: 'success' })
              loadData()
            }
          }
        })
      }
    }
  })
}

onShow(async () => {
  await loadCategories()
  await loadData()
})
</script>

<style scoped>
.page { padding: 24rpx; }
.header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 20rpx; }
.title { font-size: 36rpx; font-weight: bold; color: #1F2937; }
.header-btns { display: flex; gap: 16rpx; align-items: center; }
.btn-secondary { font-size: 28rpx; color: #6B7280; background: #F3F4F6; padding: 12rpx 24rpx; border-radius: 12rpx; }
.btn-add { font-size: 28rpx; color: #fff; background: #4F46E5; padding: 12rpx 24rpx; border-radius: 12rpx; }
.search-bar { position: relative; margin-bottom: 16rpx; }
.search-input { background: #fff; border-radius: 12rpx; padding: 16rpx 60rpx 16rpx 24rpx; font-size: 26rpx; color: #1F2937; box-shadow: 0 2rpx 8rpx rgba(0,0,0,0.04); }
.search-clear { position: absolute; right: 20rpx; top: 50%; transform: translateY(-50%); font-size: 28rpx; color: #9CA3AF; padding: 8rpx; }
.tabs { display: flex; gap: 12rpx; margin-bottom: 24rpx; flex-wrap: wrap; }
.tab { font-size: 24rpx; padding: 10rpx 20rpx; border-radius: 20rpx; background: #F3F4F6; color: #6B7280; }
.tab.active { background: #4F46E5; color: #fff; }
.loading, .empty { text-align: center; padding: 100rpx 0; color: #9CA3AF; font-size: 28rpx; }
.card { background: #fff; border-radius: 16rpx; padding: 24rpx; margin-bottom: 16rpx; box-shadow: 0 2rpx 8rpx rgba(0,0,0,0.04); }
.card-top { display: flex; justify-content: space-between; align-items: center; margin-bottom: 16rpx; }
.product-name { font-size: 30rpx; font-weight: 600; color: #1F2937; flex: 1; margin-right: 16rpx; }
.badge { font-size: 22rpx; padding: 6rpx 16rpx; border-radius: 16rpx; }
.badge-on { background: #D1FAE5; color: #059669; }
.badge-off { background: #F3F4F6; color: #6B7280; }
.card-bottom { display: flex; justify-content: space-between; align-items: center; }
.card-bottom-left { display: flex; align-items: center; gap: 12rpx; flex-wrap: wrap; }
.cat-tag { font-size: 22rpx; color: #4F46E5; background: #EEF2FF; padding: 4rpx 16rpx; border-radius: 8rpx; }
.brand-text { font-size: 24rpx; color: #9CA3AF; }
.price { font-size: 32rpx; font-weight: bold; color: #4F46E5; }
</style>
