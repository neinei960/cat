<template>
  <view class="page">
    <view class="hint">已删除的客户将在 1 天后自动清除</view>

    <view v-if="loading" class="loading">加载中...</view>
    <view v-else-if="list.length === 0" class="empty">回收站为空</view>

    <view v-else class="list">
      <view class="card" v-for="item in list" :key="item.ID">
        <view class="card-top">
          <view class="avatar">{{ (item.nickname || '客').charAt(0) }}</view>
          <view class="info">
            <text class="name">{{ item.nickname || '未命名' }}</text>
            <text class="phone">{{ item.phone || '未绑定手机' }}</text>
          </view>
          <view class="btn-restore" @click="onRestore(item)">恢复</view>
        </view>
        <view class="card-bottom">
          <text class="deleted-at">删除时间：{{ formatTime(item.DeletedAt) }}</text>
        </view>
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { getDeletedCustomers, restoreCustomer } from '@/api/customer'

const list = ref<any[]>([])
const loading = ref(true)

async function loadData() {
  loading.value = true
  try {
    const res = await getDeletedCustomers({ page: 1, page_size: 50 })
    list.value = res.data.list || []
  } finally { loading.value = false }
}

function formatTime(t: string) {
  if (!t) return ''
  const d = new Date(t)
  return `${d.getMonth() + 1}/${d.getDate()} ${d.getHours().toString().padStart(2, '0')}:${d.getMinutes().toString().padStart(2, '0')}`
}

async function onRestore(item: any) {
  uni.showModal({
    title: '确认恢复',
    content: `确定要恢复客户「${item.nickname || '未命名'}」吗？`,
    success: async (res) => {
      if (res.confirm) {
        try {
          await restoreCustomer(item.ID)
          uni.showToast({ title: '已恢复', icon: 'success' })
          await loadData()
        } catch {
          uni.showToast({ title: '恢复失败', icon: 'none' })
        }
      }
    }
  })
}

onMounted(loadData)
</script>

<style scoped>
.page { padding: 24rpx; }
.hint { font-size: 24rpx; color: #F59E0B; background: #FFFBEB; padding: 16rpx 24rpx; border-radius: 12rpx; margin-bottom: 24rpx; text-align: center; }
.loading, .empty { text-align: center; padding: 100rpx 0; color: #9CA3AF; font-size: 28rpx; }
.card { background: #fff; border-radius: 16rpx; padding: 24rpx; margin-bottom: 16rpx; box-shadow: 0 2rpx 8rpx rgba(0,0,0,0.04); }
.card-top { display: flex; align-items: center; }
.avatar { width: 80rpx; height: 80rpx; border-radius: 50%; background: #D1D5DB; color: #fff; display: flex; align-items: center; justify-content: center; font-size: 32rpx; font-weight: bold; }
.info { flex: 1; margin-left: 20rpx; }
.name { font-size: 30rpx; font-weight: 600; color: #6B7280; display: block; }
.phone { font-size: 24rpx; color: #9CA3AF; display: block; margin-top: 4rpx; }
.btn-restore { font-size: 26rpx; color: #4F46E5; background: #EEF2FF; padding: 10rpx 24rpx; border-radius: 12rpx; }
.card-bottom { margin-top: 12rpx; padding-top: 12rpx; border-top: 1rpx solid #F3F4F6; }
.deleted-at { font-size: 24rpx; color: #9CA3AF; }
</style>
