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
            <view class="info-row"><text class="label">驱虫情况</text><text class="value">{{ dewormingLabel(order.has_deworming) }}</text></view>
            <view class="info-row"><text class="label">应收金额</text><text class="value price">¥{{ order.pay_amount.toFixed(2) }}</text></view>
            <view class="info-row"><text class="label">节假日加收</text><text class="value">{{ amountLabel(order.holiday_surcharge_amount) }}</text></view>
            <view class="info-row"><text class="label">是否享受优惠</text><text class="value">{{ order.manual_discount_amount > 0 ? '是' : '否' }}</text></view>
            <view class="info-row" v-if="order.manual_discount_amount > 0"><text class="label">入住优惠</text><text class="value price-discount">-¥{{ order.manual_discount_amount.toFixed(2) }}</text></view>
            <view class="info-row" v-if="order.remark"><text class="label">备注</text><text class="value">{{ order.remark }}</text></view>
          </view>
        </view>

        <view class="card">
          <text class="section-title">金额明细</text>
          <view class="price-summary">
            <view class="summary-item">
              <text class="summary-name">基础住宿</text>
              <text class="summary-value">¥{{ amountValue(lineAmount('base')).toFixed(2) }}</text>
            </view>
            <view class="summary-item">
              <text class="summary-name">第二只加价</text>
              <text class="summary-value">{{ amountLabel(lineAmount('extra_pet')) }}</text>
            </view>
            <view class="summary-item">
              <text class="summary-name">节假日加收</text>
              <text class="summary-value">{{ amountLabel(order.holiday_surcharge_amount) }}</text>
            </view>
            <view class="summary-item">
              <text class="summary-name">是否享受优惠</text>
              <text class="summary-value">{{ order.manual_discount_amount > 0 ? '是' : '否' }}</text>
            </view>
            <view v-if="order.manual_discount_amount > 0" class="summary-item">
              <text class="summary-name">优惠金额</text>
              <text class="summary-value discount">-¥{{ order.manual_discount_amount.toFixed(2) }}</text>
            </view>
          </view>
          <view class="line-row" v-for="line in previewLines" :key="`${line.type}-${line.label}`">
            <view class="line-copy">
              <text class="line-name">{{ line.label }}</text>
              <text class="line-meta">{{ line.quantity }} × {{ signedMoney(line.unit_price) }}</text>
            </view>
            <text class="line-amount" :class="{ 'line-amount-discount': line.amount < 0 }">{{ signedMoney(line.amount) }}</text>
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

        <view v-if="showCheckInSheet" class="sheet-mask" @click="closeCheckInSheet"></view>
        <view v-if="showCheckInSheet" class="sheet-card">
          <text class="sheet-title">办理入住</text>
          <text class="sheet-desc">可在入住时登记是否享受优惠，未勾选则按原金额办理入住。</text>
          <view class="check-row" @click="toggleDiscount">
            <view class="check-box" :class="{ active: useDiscount }">
              <text v-if="useDiscount" class="check-mark">✓</text>
            </view>
            <view class="check-copy">
              <text class="check-title">享受优惠</text>
              <text class="check-sub">勾选后录入本次入住优惠金额</text>
            </view>
          </view>
          <view v-if="useDiscount" class="field-card">
            <text class="field-label">优惠金额</text>
            <input v-model="discountAmountInput" class="sheet-input" type="digit" placeholder="请输入优惠金额，例如 100" />
          </view>
          <view class="sheet-actions">
            <view class="sheet-btn" @click="closeCheckInSheet">取消</view>
            <view class="sheet-btn primary" @click="submitCheckIn">确认入住</view>
          </view>
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
const showCheckInSheet = ref(false)
const useDiscount = ref(false)
const discountAmountInput = ref('')

const pricePreview = computed<BoardingPricePreview | null>(() => {
  try {
    return JSON.parse(order.value?.price_snapshot_json || '{}') || null
  } catch {
    return null
  }
})

const previewLines = computed<BoardingPriceLine[]>(() => {
  return pricePreview.value?.lines || []
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

function dewormingLabel(value?: boolean | null) {
  if (value === true) return '已驱虫'
  if (value === false) return '未驱虫'
  return '未填写'
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

function amountValue(value?: number | null) {
  return Number(value || 0)
}

function amountLabel(value?: number | null) {
  const amount = amountValue(value)
  if (amount <= 0) return '无'
  return `¥${amount.toFixed(2)}`
}

function signedMoney(value?: number | null) {
  const amount = amountValue(value)
  return `${amount < 0 ? '-' : ''}¥${Math.abs(amount).toFixed(2)}`
}

function lineAmount(type: string) {
  return previewLines.value
    .filter((line) => line.type === type)
    .reduce((sum, line) => sum + amountValue(line.amount), 0)
}

function closeCheckInSheet() {
  showCheckInSheet.value = false
}

function toggleDiscount() {
  useDiscount.value = !useDiscount.value
  if (!useDiscount.value) {
    discountAmountInput.value = ''
  }
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
  useDiscount.value = amountValue(order.value?.manual_discount_amount) > 0
  discountAmountInput.value = useDiscount.value ? amountValue(order.value?.manual_discount_amount).toFixed(2) : ''
  showCheckInSheet.value = true
}

async function submitCheckIn() {
  let discountAmount = 0
  if (useDiscount.value) {
    discountAmount = Number(discountAmountInput.value || 0)
    if (!Number.isFinite(discountAmount) || discountAmount <= 0) {
      uni.showToast({ title: '请输入有效优惠金额', icon: 'none' })
      return
    }
  }
  await checkInBoardingOrder(id.value, { discount_amount: discountAmount })
  closeCheckInSheet()
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
    itemList: cabinets.map((item) => `${item.cabinet_type} · 剩${item.remaining_rooms || 0}/${item.room_count || 1}间 · ¥${item.base_price}/晚${item.extra_pet_price > 0 ? ` · 第二只+¥${item.extra_pet_price}` : ''}`),
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
.line-row { flex-direction: row; align-items: flex-start; }
.info-row:last-child, .line-row:last-child { border-bottom: none; }
.label, .line-name { font-size: 24rpx; color: #6B7280; }
.value, .line-amount { font-size: 24rpx; color: #111827; text-align: right; }
.price { color: #DC2626; font-weight: 700; }
.price-discount { color: #16A34A; font-weight: 700; }
.section-title { display: block; font-size: 28rpx; font-weight: 700; color: #111827; margin-bottom: 14rpx; }
.log-row { padding: 16rpx 0; border-bottom: 1rpx solid #F3F4F6; }
.log-row:last-child { border-bottom: none; }
.log-title { font-size: 24rpx; font-weight: 600; color: #111827; }
.log-meta, .log-content { font-size: 22rpx; color: #6B7280; }
.actions { display: flex; flex-wrap: wrap; gap: 12rpx; }
.action-btn { padding: 16rpx 24rpx; border-radius: 12rpx; background: #EEF2FF; color: #4F46E5; font-size: 24rpx; }
.action-btn.primary { background: #4F46E5; color: #fff; }
.action-btn.danger { background: #FEE2E2; color: #DC2626; }
.price-summary { display: flex; flex-direction: column; gap: 10rpx; margin-bottom: 12rpx; padding-bottom: 12rpx; border-bottom: 1rpx solid #F3F4F6; }
.summary-item { display: flex; justify-content: space-between; gap: 16rpx; }
.summary-name, .summary-value { font-size: 23rpx; color: #374151; }
.summary-value.discount { color: #16A34A; font-weight: 700; }
.line-copy { display: flex; flex-direction: column; gap: 6rpx; }
.line-meta { font-size: 21rpx; color: #9CA3AF; }
.line-amount-discount { color: #16A34A; font-weight: 700; }
.sheet-mask { position: fixed; inset: 0; background: rgba(15, 23, 42, 0.38); z-index: 60; }
.sheet-card {
  position: fixed;
  left: 24rpx;
  right: 24rpx;
  bottom: calc(32rpx + env(safe-area-inset-bottom));
  z-index: 61;
  background: #fff;
  border-radius: 24rpx;
  padding: 28rpx 24rpx;
  box-shadow: 0 20rpx 40rpx rgba(15, 23, 42, 0.18);
}
.sheet-title { display: block; font-size: 30rpx; font-weight: 700; color: #111827; }
.sheet-desc { display: block; margin-top: 10rpx; font-size: 22rpx; line-height: 1.6; color: #6B7280; }
.check-row { display: flex; gap: 16rpx; align-items: center; margin-top: 24rpx; padding: 20rpx; border-radius: 18rpx; background: #F8FAFC; }
.check-box { width: 36rpx; height: 36rpx; border-radius: 10rpx; border: 2rpx solid #CBD5E1; display: flex; align-items: center; justify-content: center; background: #fff; flex-shrink: 0; }
.check-box.active { background: #4F46E5; border-color: #4F46E5; }
.check-mark { font-size: 22rpx; color: #fff; font-weight: 700; }
.check-copy { display: flex; flex-direction: column; gap: 6rpx; }
.check-title { font-size: 25rpx; color: #111827; font-weight: 600; }
.check-sub { font-size: 21rpx; color: #6B7280; }
.field-card { margin-top: 20rpx; display: flex; flex-direction: column; gap: 10rpx; }
.field-label { font-size: 23rpx; color: #6B7280; }
.sheet-input { min-height: 84rpx; border-radius: 16rpx; background: #F8FAFC; padding: 0 22rpx; font-size: 26rpx; color: #111827; }
.sheet-actions { display: flex; gap: 16rpx; margin-top: 24rpx; }
.sheet-btn { flex: 1; text-align: center; padding: 18rpx 20rpx; border-radius: 16rpx; background: #EEF2FF; color: #4F46E5; font-size: 26rpx; font-weight: 600; }
.sheet-btn.primary { background: #4F46E5; color: #fff; }
</style>
