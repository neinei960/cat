<template>
  <SideLayout>
    <view class="page">
      <view v-if="loading" class="state">加载中...</view>
      <view v-else-if="!order" class="state">寄养订单不存在</view>
      <template v-else>
        <view class="card">
          <view class="head">
            <view>
              <text class="title">{{ order.cabinet?.cabinet_type || '寄养房型' }} · {{ statusLabel(order.status) }}</text>
              <text class="sub">{{ order.customer?.nickname || order.customer?.phone || '-' }} · {{ petNames }}</text>
            </view>
            <view class="btn btn-primary" @click="goOrderDetail" v-if="order.order_id">去收款</view>
          </view>
          <view class="info-list">
            <view class="info-row"><text class="label">入住日期</text><text class="value">{{ order.check_in_at }}</text></view>
            <view class="info-row"><text class="label">预计离店</text><text class="value">{{ order.check_out_at }}</text></view>
            <view class="info-row" v-if="order.actual_check_out_at"><text class="label">实际离店</text><text class="value">{{ order.actual_check_out_at }}</text></view>
            <view class="info-row"><text class="label">寄养晚数</text><text class="value">{{ order.nights }} 晚</text></view>
            <view class="info-row"><text class="label">应收金额</text><text class="value price">¥{{ order.pay_amount.toFixed(2) }}</text></view>
            <view class="info-row" v-if="order.remark"><text class="label">备注</text><text class="value">{{ order.remark }}</text></view>
          </view>
        </view>

        <view class="card">
          <text class="section-title">金额明细</text>
          <view class="line-row" v-for="line in previewLines" :key="`${line.type}-${line.label}`">
            <text class="line-name">{{ line.label }}</text>
            <text class="line-amount">¥{{ line.amount.toFixed(2) }}</text>
          </view>
        </view>

        <view class="card" v-if="logs.length > 0">
          <text class="section-title">操作记录</text>
          <view class="log-row" v-for="log in logs" :key="log.ID">
            <text class="log-title">{{ actionLabel(log.action) }}</text>
            <text class="log-meta">{{ log.operator?.name || '-' }} · {{ formatTime(log.CreatedAt) }}</text>
            <text class="log-content">{{ log.content }}</text>
          </view>
        </view>

        <view class="actions">
          <view v-if="order.status === 'pending_checkin'" class="action-btn primary" @click="handleCheckIn">办理入住</view>
          <view v-if="order.status === 'pending_checkin'" class="action-btn danger" @click="handleCancel">取消</view>
          <view v-if="order.status === 'checked_in'" class="action-btn" @click="handleExtend">续住</view>
          <view v-if="order.status === 'checked_in'" class="action-btn" @click="handleChangeCabinet">换房型</view>
          <view v-if="order.status === 'checked_in'" class="action-btn primary" @click="handleCheckOut">办理离店</view>
        </view>
      </template>
    </view>
  </SideLayout>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { onLoad, onShow } from '@dcloudio/uni-app'
import SideLayout from '@/components/SideLayout.vue'
import {
  cancelBoardingOrder,
  changeBoardingCabinet,
  checkInBoardingOrder,
  checkOutBoardingOrder,
  extendBoardingOrder,
  getAvailableBoardingCabinets,
  getBoardingOrder,
} from '@/api/boarding'

const id = ref(0)
const loading = ref(false)
const order = ref<BoardingOrder | null>(null)

const previewLines = computed<BoardingPriceLine[]>(() => {
  try {
    return JSON.parse(order.value?.price_snapshot_json || '{}')?.lines || []
  } catch {
    return []
  }
})

const logs = computed(() => order.value?.logs || [])
const petNames = computed(() => order.value?.pets?.map((item) => item.pet?.name || item.pet_name_snapshot).filter(Boolean).join('、') || '-')

function statusLabel(status: string) {
  return {
    pending_checkin: '待入住',
    checked_in: '在住',
    checked_out: '已离店',
    cancelled: '已取消',
  }[status] || status
}

function actionLabel(action: string) {
  return {
    create: '创建',
    check_in: '入住',
    check_out: '离店',
    extend: '续住',
    change_cabinet: '换房型',
    cancel: '取消',
  }[action] || action
}

function formatTime(value?: string) {
  if (!value) return '-'
  return value.replace('T', ' ').slice(0, 16)
}

function goOrderDetail() {
  if (!order.value?.order_id) return
  uni.navigateTo({ url: `/pages/order/detail?id=${order.value.order_id}` })
}

async function loadData() {
  if (!id.value) return
  loading.value = true
  try {
    const res = await getBoardingOrder(id.value)
    order.value = res.data
  } finally {
    loading.value = false
  }
}

async function handleCheckIn() {
  await checkInBoardingOrder(id.value)
  uni.showToast({ title: '已办理入住', icon: 'success' })
  await loadData()
}

async function handleCancel() {
  uni.showModal({
    title: '确认取消',
    content: '取消后寄养单和关联未支付订单会一起取消。',
    success: async (res) => {
      if (!res.confirm) return
      await cancelBoardingOrder(id.value)
      uni.showToast({ title: '已取消', icon: 'success' })
      await loadData()
    },
  })
}

async function handleExtend() {
  uni.showModal({
    title: '续住到',
    editable: true,
    placeholderText: 'YYYY-MM-DD',
    content: order.value?.check_out_at || '',
    success: async (res) => {
      if (!res.confirm || !res.content?.trim()) return
      await extendBoardingOrder(id.value, res.content.trim())
      uni.showToast({ title: '续住成功', icon: 'success' })
      await loadData()
    },
  })
}

async function handleChangeCabinet() {
  if (!order.value) return
  const res = await getAvailableBoardingCabinets({
    check_in_at: order.value.check_in_at,
    check_out_at: order.value.check_out_at,
    pet_count: order.value.pets?.length || 1,
    exclude_order_id: order.value.ID,
  })
  const cabinets = (res.data || []).filter((item) => item.ID !== order.value?.cabinet_id)
  if (cabinets.length === 0) {
    uni.showToast({ title: '当前没有可更换的寄养房型', icon: 'none' })
    return
  }
  uni.showActionSheet({
    itemList: cabinets.map((item) => `${item.cabinet_type} · 剩${item.remaining_rooms || 0}/${item.room_count || 1}间 · ¥${item.base_price}/晚`),
    success: async ({ tapIndex }) => {
      await changeBoardingCabinet(id.value, cabinets[tapIndex].ID)
      uni.showToast({ title: '换房型成功', icon: 'success' })
      await loadData()
    },
  })
}

async function handleCheckOut() {
  uni.showModal({
    title: '实际离店日期',
    editable: true,
    placeholderText: 'YYYY-MM-DD',
    content: order.value?.check_out_at || '',
    success: async (res) => {
      if (!res.confirm || !res.content?.trim()) return
      await checkOutBoardingOrder(id.value, res.content.trim())
      uni.showToast({ title: '已重算离店金额', icon: 'success' })
      await loadData()
    },
  })
}

onLoad((query) => {
  id.value = Number(query?.id || 0)
})

onShow(loadData)
</script>

<style scoped>
.page { padding: 24rpx; display: flex; flex-direction: column; gap: 20rpx; }
.state { text-align: center; padding: 120rpx 24rpx; color: #9CA3AF; }
.card { background: #fff; border-radius: 18rpx; padding: 24rpx; box-shadow: 0 12rpx 28rpx rgba(15, 23, 42, 0.04); }
.head { display: flex; justify-content: space-between; gap: 12rpx; align-items: center; margin-bottom: 20rpx; }
.title { display: block; font-size: 32rpx; font-weight: 700; color: #111827; }
.sub { display: block; margin-top: 8rpx; font-size: 24rpx; color: #6B7280; }
.btn { padding: 14rpx 24rpx; border-radius: 12rpx; background: #EEF2FF; color: #4F46E5; font-size: 24rpx; }
.btn-primary { background: #4F46E5; color: #fff; }
.info-list, .line-row, .log-row { display: flex; flex-direction: column; gap: 8rpx; }
.info-row, .line-row { padding: 14rpx 0; border-bottom: 1rpx solid #F3F4F6; display: flex; justify-content: space-between; gap: 16rpx; }
.info-row:last-child, .line-row:last-child { border-bottom: none; }
.label, .line-name { font-size: 24rpx; color: #6B7280; }
.value, .line-amount { font-size: 24rpx; color: #111827; text-align: right; }
.price { color: #DC2626; font-weight: 700; }
.section-title { display: block; font-size: 28rpx; font-weight: 700; color: #111827; margin-bottom: 14rpx; }
.log-row { padding: 16rpx 0; border-bottom: 1rpx solid #F3F4F6; }
.log-row:last-child { border-bottom: none; }
.log-title { font-size: 24rpx; font-weight: 600; color: #111827; }
.log-meta, .log-content { font-size: 22rpx; color: #6B7280; }
.actions { display: flex; flex-wrap: wrap; gap: 12rpx; }
.action-btn { padding: 16rpx 24rpx; border-radius: 12rpx; background: #EEF2FF; color: #4F46E5; font-size: 24rpx; }
.action-btn.primary { background: #4F46E5; color: #fff; }
.action-btn.danger { background: #FEE2E2; color: #DC2626; }
</style>
