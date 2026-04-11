<template>
  <SideLayout>
    <view class="page">
      <view class="hint">已删除订单会在 2 天后自动清除，期间可恢复</view>

      <view v-if="loading" class="loading">加载中...</view>
      <view v-else-if="list.length === 0" class="empty">回收站为空</view>

      <view v-else class="list">
        <view class="card" v-for="item in list" :key="item.ID" @click="goDetail(item.ID)">
          <view class="card-top">
            <view class="info">
              <text class="order-no">{{ item.order_no }}</text>
              <text class="customer">{{ getOrderTitle(item) }}</text>
            </view>
            <view class="btn-restore" @click.stop="onRestore(item)">恢复</view>
          </view>

          <view class="meta-row">
            <text class="meta">{{ formatTime(item.DeletedAt) }} 删除</text>
            <text class="meta">实付 ¥{{ item.pay_amount || 0 }}</text>
          </view>

          <view class="meta-row">
            <text class="status">{{ statusMap[item.status] || '未知状态' }}</text>
            <text v-if="item.appointment?.date" class="meta">预约 {{ item.appointment.date }}</text>
          </view>
        </view>
      </view>
    </view>
  </SideLayout>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import SideLayout from '@/components/SideLayout.vue'
import { getDeletedOrders, restoreOrder } from '@/api/order'

const list = ref<any[]>([])
const loading = ref(true)
const statusMap: Record<number, string> = { 0: '待付款', 1: '已支付', 2: '已取消', 3: '已退款' }

async function loadData() {
  loading.value = true
  try {
    const res = await getDeletedOrders({ page: 1, page_size: 50 })
    list.value = res.data.list || []
  } finally {
    loading.value = false
  }
}

function formatTime(t: string) {
  if (!t) return ''
  const d = new Date(t)
  return `${d.getMonth() + 1}/${d.getDate()} ${d.getHours().toString().padStart(2, '0')}:${d.getMinutes().toString().padStart(2, '0')}`
}

function getOrderTitle(item: any) {
  const customerName = item.customer?.nickname || '散客'
  if (item.pet_summary) return `${customerName} · 🐱${item.pet_summary}`
  if (item.pet?.name) return `${customerName} · 🐱${item.pet.name}`
  if (item.order_kind === 'product') return `${customerName} · 商品零售`
  if (item.order_kind === 'feeding') return `${customerName} · 上门喂养`
  return customerName
}

function goDetail(id: number) {
  uni.navigateTo({ url: `/pages/order/detail?id=${id}&include_deleted=1` })
}

async function onRestore(item: any) {
  uni.showModal({
    title: '确认恢复',
    content: `确定要恢复订单 ${item.order_no} 吗？`,
    success: async (res) => {
      if (!res.confirm) return
      try {
        await restoreOrder(item.ID)
        uni.showToast({ title: '已恢复', icon: 'success' })
        await loadData()
      } catch (error: any) {
        uni.showToast({ title: error?.message || '恢复失败', icon: 'none' })
      }
    },
  })
}

onMounted(loadData)
</script>

<style scoped>
.page { padding: 24rpx; }
.hint { font-size: 24rpx; color: #B45309; background: #FFFBEB; padding: 16rpx 24rpx; border-radius: 12rpx; margin-bottom: 24rpx; text-align: center; }
.loading, .empty { text-align: center; padding: 100rpx 0; color: #9CA3AF; font-size: 28rpx; }
.card { background: #fff; border-radius: 16rpx; padding: 24rpx; margin-bottom: 16rpx; box-shadow: 0 2rpx 8rpx rgba(0, 0, 0, 0.04); }
.card-top { display: flex; align-items: center; justify-content: space-between; gap: 16rpx; }
.info { flex: 1; min-width: 0; }
.order-no { display: block; font-size: 24rpx; color: #9CA3AF; }
.customer { display: block; font-size: 28rpx; font-weight: 600; color: #1F2937; margin-top: 6rpx; }
.btn-restore { font-size: 26rpx; color: #4F46E5; background: #EEF2FF; padding: 10rpx 24rpx; border-radius: 12rpx; flex-shrink: 0; }
.meta-row { display: flex; justify-content: space-between; align-items: center; gap: 12rpx; margin-top: 14rpx; }
.meta { font-size: 24rpx; color: #6B7280; }
.status { font-size: 22rpx; color: #92400E; background: #FEF3C7; padding: 6rpx 14rpx; border-radius: 999rpx; }
</style>
