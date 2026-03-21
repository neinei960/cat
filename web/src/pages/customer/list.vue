<template>
  <SideLayout>
  <view class="page">
    <view class="header">
      <text class="title">客户管理</text>
      <view class="header-actions">
        <view class="btn-trash" @click="goTrash">回收站</view>
        <view class="btn-add" @click="goAdd">+ 新增客户</view>
      </view>
    </view>

    <view class="search-bar">
      <input v-model="keyword" placeholder="搜索客户姓名/手机号" class="search-input" @confirm="loadData" />
    </view>

    <view v-if="loading" class="loading">加载中...</view>
    <view v-else-if="list.length === 0" class="empty">暂无客户</view>

    <view v-else class="list">
      <view class="card" v-for="item in list" :key="item.ID" @click="goDetail(item.ID)" @longpress="onLongPress(item)">
        <view class="card-top">
          <view class="avatar">{{ (item.nickname || '客').charAt(0) }}</view>
          <view class="info">
            <text class="name">{{ item.nickname || '未命名' }}</text>
            <text class="phone">{{ item.phone || '未绑定手机' }}</text>
          </view>
          <view class="visit-count">到店 {{ item.visit_count }} 次</view>
        </view>
        <view class="card-bottom">
          <text class="spent">累计消费 ¥{{ item.total_spent.toFixed(2) }}</text>
          <text class="member-badge" v-if="item.member_card_id">会员 ¥{{ (item.member_balance || 0).toFixed(2) }}</text>
          <text class="tags" v-if="item.tags">{{ item.tags }}</text>
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
import { getCustomerList, deleteCustomer } from '@/api/customer'

const list = ref<Customer[]>([])
const loading = ref(true)
const keyword = ref('')

async function loadData() {
  loading.value = true
  try {
    const res = await getCustomerList({ page: 1, page_size: 50, keyword: keyword.value || undefined })
    list.value = res.data.list || []
  } finally { loading.value = false }
}

function goAdd() { uni.navigateTo({ url: '/pages/customer/edit' }) }
function goDetail(id: number) { uni.navigateTo({ url: `/pages/customer/detail?id=${id}` }) }
function goTrash() { uni.navigateTo({ url: '/pages/customer/trash' }) }

function onLongPress(item: any) {
  try { uni.vibrateShort({}) } catch (_) {}
  uni.showActionSheet({
    itemList: ['删除该客户'],
    success: (res) => {
      if (res.tapIndex === 0) {
        uni.showModal({
          title: '确认删除',
          content: `确定要删除客户「${item.nickname || '未命名'}」吗？\n可在回收站中1天内恢复。`,
          confirmColor: '#EF4444',
          success: async (modalRes) => {
            if (modalRes.confirm) {
              try {
                await deleteCustomer(item.ID)
                uni.showToast({ title: '已移入回收站', icon: 'success' })
                await loadData()
              } catch {
                uni.showToast({ title: '删除失败', icon: 'none' })
              }
            }
          }
        })
      }
    }
  })
}

onMounted(loadData)
onShow(loadData)
</script>

<style scoped>
.page { padding: 24rpx; }
.header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 20rpx; }
.title { font-size: 36rpx; font-weight: bold; color: #1F2937; }
.header-actions { display: flex; gap: 12rpx; align-items: center; }
.btn-trash { font-size: 26rpx; color: #6B7280; background: #F3F4F6; padding: 12rpx 20rpx; border-radius: 12rpx; }
.btn-add { font-size: 28rpx; color: #fff; background: #4F46E5; padding: 12rpx 24rpx; border-radius: 12rpx; }
.search-bar { margin-bottom: 24rpx; }
.search-input { background: #fff; border-radius: 12rpx; padding: 16rpx 24rpx; font-size: 28rpx; }
.loading, .empty { text-align: center; padding: 100rpx 0; color: #9CA3AF; font-size: 28rpx; }
.card { background: #fff; border-radius: 16rpx; padding: 24rpx; margin-bottom: 16rpx; box-shadow: 0 2rpx 8rpx rgba(0,0,0,0.04); }
.card-top { display: flex; align-items: center; }
.avatar { width: 80rpx; height: 80rpx; border-radius: 50%; background: #6366F1; color: #fff; display: flex; align-items: center; justify-content: center; font-size: 32rpx; font-weight: bold; }
.info { flex: 1; margin-left: 20rpx; }
.name { font-size: 30rpx; font-weight: 600; color: #1F2937; display: block; }
.phone { font-size: 24rpx; color: #6B7280; display: block; margin-top: 4rpx; }
.visit-count { font-size: 24rpx; color: #4F46E5; }
.card-bottom { display: flex; justify-content: space-between; margin-top: 16rpx; padding-top: 16rpx; border-top: 1rpx solid #F3F4F6; }
.spent { font-size: 26rpx; color: #374151; }
.member-badge { font-size: 22rpx; color: #4F46E5; background: #EEF2FF; padding: 4rpx 14rpx; border-radius: 12rpx; margin-left: 12rpx; }
.tags { font-size: 24rpx; color: #6B7280; }
</style>
