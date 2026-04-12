<template>
  <SideLayout>
    <view class="page">
      <view v-if="loading" class="state">加载中...</view>
      <view v-else-if="!order" class="state">寄养订单不存在</view>
      <template v-else>
        <view class="card summary-card">
          <view class="summary-head">
            <view class="summary-copy">
              <text class="summary-caption">寄养详情</text>
              <text class="summary-title">{{ order.customer?.nickname || order.customer?.phone || '-' }}</text>
              <text class="summary-sub">{{ allPetNames }} · {{ displayRooms.length }} 个房间分组</text>
            </view>
            <view class="summary-side">
              <text :class="['status-pill', order.status, 'summary-status']">{{ statusLabel(order.status) }}</text>
              <view v-if="order.order_id" class="btn btn-primary summary-btn" @click="goOrderDetail">去收款</view>
            </view>
          </view>

          <view class="summary-facts">
            <view class="summary-fact">
              <text class="summary-fact-label">入住</text>
              <text class="summary-fact-value">{{ order.check_in_at }}</text>
            </view>
            <view class="summary-fact">
              <text class="summary-fact-label">离店</text>
              <text class="summary-fact-value">{{ order.check_out_at }}</text>
            </view>
            <view class="summary-fact">
              <text class="summary-fact-label">驱虫</text>
              <text class="summary-fact-value">{{ dewormingLabel(order.has_deworming) }}</text>
            </view>
            <view class="summary-fact accent">
              <text class="summary-fact-label">应收</text>
              <text class="summary-fact-value price">¥{{ order.pay_amount.toFixed(2) }}</text>
            </view>
          </view>

          <view v-if="order.remark" class="note-card">
            <text class="note-label">入住备注</text>
            <text class="note-value">{{ order.remark }}</text>
          </view>
        </view>

        <view class="card section-card">
          <view class="section-headline">
            <view>
              <text class="section-title">房间安排</text>
              <text class="section-subtitle">先看每个房间的猫咪、日期和金额，明细按需展开。</text>
            </view>
          </view>
          <view class="room-list">
            <view v-for="room in displayRooms" :key="room.ID || `legacy-${room.room_index}`" class="room-card">
              <view class="room-head">
                <view>
                  <text class="room-title">{{ roomLabel(room) }} · {{ room.cabinet?.cabinet_type || '未选房型' }}</text>
                  <text class="room-sub">{{ roomPetNames(room) }}</text>
                </view>
                <text :class="['status-pill', room.status]">{{ statusLabel(room.status) }}</text>
              </view>

              <view class="room-facts">
                <view class="fact-pill">
                  <text class="fact-label">入住</text>
                  <text class="fact-value">{{ room.check_in_at }}</text>
                </view>
                <view class="fact-pill">
                  <text class="fact-label">离店</text>
                  <text class="fact-value">{{ room.check_out_at }}</text>
                </view>
                <view v-if="room.actual_check_out_at" class="fact-pill">
                  <text class="fact-label">实际离店</text>
                  <text class="fact-value">{{ room.actual_check_out_at }}</text>
                </view>
                <view class="fact-pill">
                  <text class="fact-label">晚数</text>
                  <text class="fact-value">{{ room.nights }} 晚</text>
                </view>
                <view class="fact-pill accent">
                  <text class="fact-label">房间应收</text>
                  <text class="fact-value price">¥{{ roomDisplayPay(room).toFixed(2) }}</text>
                </view>
                <view v-if="room.manual_discount_amount > 0" class="fact-pill discount">
                  <text class="fact-label">入住优惠</text>
                  <text class="fact-value discount">-¥{{ room.manual_discount_amount.toFixed(2) }}</text>
                </view>
              </view>

              <view v-if="roomPreview(room)?.lines?.length" class="detail-toggle" @click="toggleRoomLines(room)">
                <text class="detail-toggle-text">{{ isRoomLinesOpen(room) ? '收起费用明细' : '展开费用明细' }}</text>
              </view>

              <view v-if="isRoomLinesOpen(room) && roomPreview(room)?.lines?.length" class="line-list compact-line-list">
                <view v-for="line in roomPreview(room)?.lines || []" :key="`${room.ID}-${line.type}-${line.label}`" class="line-row">
                  <text class="line-name">{{ line.label }}</text>
                  <text class="line-amount" :class="{ discount: line.amount < 0 }">{{ signedMoney(line.amount) }}</text>
                </view>
              </view>

              <view class="room-actions">
                <view v-if="room.status === 'pending_checkin'" class="action-btn primary" @click="openCheckIn(room)">办理入住</view>
                <view v-if="room.status === 'pending_checkin'" class="action-btn danger" @click="handleCancel(room)">取消</view>
                <view v-if="room.status === 'checked_in'" class="action-btn" @click="handleExtend(room)">续住</view>
                <view v-if="room.status === 'checked_in'" class="action-btn" @click="handleChangeCabinet(room)">换房型</view>
                <view v-if="room.status === 'checked_in'" class="action-btn primary" @click="handleCheckOut(room)">办理离店</view>
              </view>
            </view>
          </view>
        </view>

        <view class="card section-card detail-card">
          <view class="section-headline between">
            <view>
              <text class="section-title">更多信息</text>
              <text class="section-subtitle">金额和日志放在同一个区域里切换查看。</text>
            </view>
            <view v-if="logs.length > 0" class="tab-switch">
              <view :class="['tab-pill', detailTab === 'amount' ? 'active' : '']" @click="detailTab = 'amount'">金额</view>
              <view :class="['tab-pill', detailTab === 'logs' ? 'active' : '']" @click="detailTab = 'logs'">日志</view>
            </view>
          </view>

          <template v-if="detailTab === 'amount' || logs.length === 0">
            <view class="total-strip">
              <view>
                <text class="total-label">整单应收</text>
                <text class="total-value">¥{{ order.pay_amount.toFixed(2) }}</text>
              </view>
              <text class="total-meta">{{ displayRooms.length }} 个房间 · {{ allPetNames }}</text>
            </view>

            <view class="line-list compact-line-list">
              <view v-for="line in aggregateLines" :key="`${line.type}-${line.label}`" class="line-row">
                <text class="line-name">{{ line.label }}</text>
                <text class="line-amount" :class="{ discount: line.amount < 0 }">{{ signedMoney(line.amount) }}</text>
              </view>
            </view>
          </template>

          <view v-else class="log-list">
            <view v-for="log in logs" :key="log.ID" class="log-row">
              <text class="log-title">{{ actionLabel(log.action) }}</text>
              <text class="log-meta">{{ log.operator?.name || '-' }} · {{ formatTime(log.CreatedAt) }}</text>
              <text class="log-content">{{ log.content }}</text>
            </view>
          </view>
        </view>

        <view v-if="canCancelWholeOrder" class="footer-actions">
          <view class="action-btn danger large" @click="handleCancelWholeOrder">整单取消</view>
        </view>

        <view v-if="showCheckInSheet" class="sheet-mask" @click="closeCheckInSheet"></view>
        <view v-if="showCheckInSheet" class="sheet-card">
          <text class="sheet-title">{{ activeRoom ? `${roomLabel(activeRoom)} 办理入住` : '办理入住' }}</text>
          <text class="sheet-desc">入住优惠按房间单独录入，整单收款会自动同步。</text>
          <view class="check-row" @click="toggleDiscount">
            <view :class="['check-box', useDiscount ? 'active' : '']">
              <text v-if="useDiscount" class="check-mark">✓</text>
            </view>
            <view class="check-copy">
              <text class="check-title">享受入住优惠</text>
              <text class="check-sub">勾选后录入这个房间的优惠金额</text>
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
  cancelBoardingRoom,
  changeBoardingCabinet,
  changeBoardingRoomCabinet,
  checkInBoardingOrder,
  checkInBoardingRoom,
  checkOutBoardingOrder,
  checkOutBoardingRoom,
  extendBoardingOrder,
  extendBoardingRoom,
  getAvailableBoardingCabinets,
  getBoardingOrder,
} from '@/api/boarding'

const id = ref(0)
const loading = ref(false)
const order = ref<BoardingOrder | null>(null)
const showCheckInSheet = ref(false)
const useDiscount = ref(false)
const discountAmountInput = ref('')
const activeRoom = ref<BoardingOrderRoom | null>(null)
const roomLineOpen = ref<Record<string, boolean>>({})
const detailTab = ref<'amount' | 'logs'>('amount')

const aggregatePreview = computed<BoardingPricePreview | null>(() => {
  try {
    return JSON.parse(order.value?.price_snapshot_json || '{}') || null
  } catch {
    return null
  }
})

const displayRooms = computed(() => order.value?.rooms || [])
const aggregateLines = computed(() => aggregatePreview.value?.lines || [])
const logs = computed(() => order.value?.logs || [])
const allPetNames = computed(() => order.value?.pets?.map((item) => item.pet?.name || item.pet_name_snapshot).filter(Boolean).join('、') || '-')
const canCancelWholeOrder = computed(() => displayRooms.value.length > 1 && displayRooms.value.every((room) => room.status === 'pending_checkin'))

function roomKey(room: BoardingOrderRoom) {
  return String(room.ID || `legacy-${room.room_index || 1}`)
}

function statusLabel(status: string) {
  return {
    pending_checkin: '待入住',
    checked_in: '在住',
    checked_out: '已离店',
    cancelled: '已取消',
    mixed: '混合状态',
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

function dewormingLabel(value?: boolean | null) {
  if (value === true) return '已驱虫'
  if (value === false) return '未驱虫'
  return '未填写'
}

function roomLabel(room: BoardingOrderRoom) {
  return `房间${room.room_index || 1}`
}

function roomPetNames(room: BoardingOrderRoom) {
  return room.pets?.map((item) => item.pet?.name || item.pet_name_snapshot).filter(Boolean).join('、') || '未选猫咪'
}

function roomPreview(room: BoardingOrderRoom) {
  const aggregateRoom = aggregatePreview.value?.rooms?.find((item) => item.room_index === room.room_index)
  if (aggregateRoom) return aggregateRoom
  try {
    const raw = JSON.parse(room.price_snapshot_json || '{}') as BoardingPricePreview
    return {
      room_index: room.room_index,
      cabinet_id: room.cabinet_id,
      cabinet_type: room.cabinet?.cabinet_type || '',
      pet_count: room.pets?.length || 1,
      check_in_at: room.check_in_at,
      check_out_at: room.check_out_at,
      nights: raw.nights || room.nights,
      regular_nights: raw.regular_nights || 0,
      holiday_nights: raw.holiday_nights || 0,
      base_amount: raw.base_amount || room.base_amount,
      extra_pet_amount: raw.extra_pet_amount || 0,
      holiday_surcharge_amount: raw.holiday_surcharge_amount || room.holiday_surcharge_amount,
      discount_amount: raw.discount_amount || room.discount_amount,
      manual_discount_amount: room.manual_discount_amount,
      pay_amount: Math.max((raw.pay_amount || room.pay_amount) - (room.manual_discount_amount || 0), 0),
      lines: raw.lines || [],
    } as BoardingRoomPreview
  } catch {
    return null
  }
}

function roomDisplayPay(room: BoardingOrderRoom) {
  return roomPreview(room)?.pay_amount ?? Math.max((room.pay_amount || 0) - (room.manual_discount_amount || 0), 0)
}

function isRoomLinesOpen(room: BoardingOrderRoom) {
  return !!roomLineOpen.value[roomKey(room)]
}

function toggleRoomLines(room: BoardingOrderRoom) {
  const key = roomKey(room)
  roomLineOpen.value = {
    ...roomLineOpen.value,
    [key]: !roomLineOpen.value[key],
  }
}

function formatTime(value?: string) {
  if (!value) return '-'
  return value.replace('T', ' ').slice(0, 16)
}

function signedMoney(value?: number | null) {
  const amount = Number(value || 0)
  return `${amount < 0 ? '-' : ''}¥${Math.abs(amount).toFixed(2)}`
}

function closeCheckInSheet() {
  showCheckInSheet.value = false
  activeRoom.value = null
}

function toggleDiscount() {
  useDiscount.value = !useDiscount.value
  if (!useDiscount.value) discountAmountInput.value = ''
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
    if (!Object.keys(roomLineOpen.value).length && res.data.rooms?.length === 1) {
      const onlyRoom = res.data.rooms[0]
      roomLineOpen.value = { [roomKey(onlyRoom)]: true }
    }
  } finally {
    loading.value = false
  }
}

function openCheckIn(room: BoardingOrderRoom) {
  activeRoom.value = room
  useDiscount.value = Number(room.manual_discount_amount || 0) > 0
  discountAmountInput.value = useDiscount.value ? Number(room.manual_discount_amount || 0).toFixed(2) : ''
  showCheckInSheet.value = true
}

async function submitCheckIn() {
  if (!activeRoom.value || !order.value) return
  let discountAmount = 0
  if (useDiscount.value) {
    discountAmount = Number(discountAmountInput.value || 0)
    if (!Number.isFinite(discountAmount) || discountAmount <= 0) {
      uni.showToast({ title: '请输入有效优惠金额', icon: 'none' })
      return
    }
  }
  if (activeRoom.value.ID > 0) {
    await checkInBoardingRoom(order.value.ID, activeRoom.value.ID, { discount_amount: discountAmount })
  } else {
    await checkInBoardingOrder(order.value.ID, { discount_amount: discountAmount })
  }
  closeCheckInSheet()
  uni.showToast({ title: '已办理入住', icon: 'success' })
  await loadData()
}

async function handleCancel(room: BoardingOrderRoom) {
  if (!order.value) return
  uni.showModal({
    title: '确认取消',
    content: `${roomLabel(room)} 取消后会从整单中移除，并同步更新收款金额。`,
    success: async (res) => {
      if (!res.confirm) return
      if (room.ID > 0) await cancelBoardingRoom(order.value!.ID, room.ID)
      else await cancelBoardingOrder(order.value!.ID)
      uni.showToast({ title: '已取消', icon: 'success' })
      await loadData()
    },
  })
}

async function handleCancelWholeOrder() {
  if (!order.value) return
  uni.showModal({
    title: '确认整单取消',
    content: '整张寄养单下的所有待入住房间都会一起取消。',
    success: async (res) => {
      if (!res.confirm) return
      await cancelBoardingOrder(order.value!.ID)
      uni.showToast({ title: '已整单取消', icon: 'success' })
      await loadData()
    },
  })
}

async function handleExtend(room: BoardingOrderRoom) {
  if (!order.value) return
  uni.showModal({
    title: `${roomLabel(room)} 续住到`,
    editable: true,
    placeholderText: 'YYYY-MM-DD',
    content: room.check_out_at || '',
    success: async (res) => {
      if (!res.confirm || !res.content?.trim()) return
      if (room.ID > 0) await extendBoardingRoom(order.value!.ID, room.ID, res.content.trim())
      else await extendBoardingOrder(order.value!.ID, res.content.trim())
      uni.showToast({ title: '续住成功', icon: 'success' })
      await loadData()
    },
  })
}

async function handleChangeCabinet(room: BoardingOrderRoom) {
  if (!order.value) return
  const res = await getAvailableBoardingCabinets({
    check_in_at: room.check_in_at,
    check_out_at: room.check_out_at,
    pet_count: room.pets?.length || 1,
    exclude_order_id: order.value.ID,
    exclude_room_id: room.ID || undefined,
  })
  const cabinets = (res.data || []).filter((item) => item.ID !== room.cabinet_id)
  if (cabinets.length === 0) {
    uni.showToast({ title: '当前没有可更换的寄养房型', icon: 'none' })
    return
  }
  uni.showActionSheet({
    itemList: cabinets.map((item) => `${item.cabinet_type} · 剩${item.remaining_rooms || 0}/${item.room_count || 1}间 · ¥${item.base_price}/晚${item.extra_pet_price > 0 ? ` · 第二只+¥${item.extra_pet_price}` : ''}`),
    success: async ({ tapIndex }) => {
      if (room.ID > 0) await changeBoardingRoomCabinet(order.value!.ID, room.ID, cabinets[tapIndex].ID)
      else await changeBoardingCabinet(order.value!.ID, cabinets[tapIndex].ID)
      uni.showToast({ title: '换房型成功', icon: 'success' })
      await loadData()
    },
  })
}

async function handleCheckOut(room: BoardingOrderRoom) {
  if (!order.value) return
  uni.showModal({
    title: `${roomLabel(room)} 实际离店日期`,
    editable: true,
    placeholderText: 'YYYY-MM-DD',
    content: room.check_out_at || '',
    success: async (res) => {
      if (!res.confirm || !res.content?.trim()) return
      if (room.ID > 0) await checkOutBoardingRoom(order.value!.ID, room.ID, res.content.trim())
      else await checkOutBoardingOrder(order.value!.ID, res.content.trim())
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
.page {
  padding: 24rpx 24rpx calc(180rpx + env(safe-area-inset-bottom));
  display: flex;
  flex-direction: column;
  gap: 20rpx;
  background:
    radial-gradient(circle at top left, rgba(224, 231, 255, 0.55), transparent 28%),
    linear-gradient(180deg, #f8faff 0%, #f5f7fb 40%, #f8fafc 100%);
}
.state {
  text-align: center;
  padding: 120rpx 24rpx;
  color: #9ca3af;
}
.card {
  background: rgba(255, 255, 255, 0.96);
  border-radius: 26rpx;
  padding: 24rpx;
  border: 1rpx solid rgba(226, 232, 240, 0.95);
  box-shadow: 0 16rpx 36rpx rgba(15, 23, 42, 0.06);
}
.summary-card {
  background: linear-gradient(180deg, rgba(255, 255, 255, 0.98), rgba(248, 250, 255, 0.98));
}
.summary-head,
.room-head,
.section-headline,
.section-headline.between,
.total-strip,
.sheet-actions {
  display: flex;
  justify-content: space-between;
  gap: 16rpx;
  align-items: flex-start;
}
.summary-copy {
  min-width: 0;
}
.summary-caption {
  display: block;
  font-size: 22rpx;
  color: #818cf8;
  font-weight: 700;
  letter-spacing: 1rpx;
}
.summary-title {
  display: block;
  margin-top: 8rpx;
  font-size: 38rpx;
  font-weight: 800;
  color: #111827;
}
.summary-sub,
.section-subtitle {
  display: block;
  margin-top: 8rpx;
  font-size: 23rpx;
  line-height: 1.6;
  color: #6b7280;
}
.summary-side {
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  gap: 12rpx;
}
.btn {
  padding: 14rpx 22rpx;
  border-radius: 16rpx;
  border: 1rpx solid #e5e7eb;
  background: #fff;
  font-size: 24rpx;
  color: #374151;
}
.btn-primary {
  background: linear-gradient(135deg, #4f46e5, #6366f1);
  border-color: transparent;
  color: #fff;
}
.summary-btn {
  white-space: nowrap;
}
.status-pill {
  padding: 10rpx 16rpx;
  border-radius: 999rpx;
  background: #eef2ff;
  color: #4f46e5;
  font-size: 22rpx;
  white-space: nowrap;
  font-weight: 700;
}
.summary-status {
  font-size: 23rpx;
}
.status-pill.checked_in {
  background: #ecfdf5;
  color: #059669;
}
.status-pill.checked_out {
  background: #f3f4f6;
  color: #6b7280;
}
.status-pill.cancelled {
  background: #fef2f2;
  color: #dc2626;
}
.status-pill.mixed {
  background: #fff7ed;
  color: #c2410c;
}
.summary-facts {
  display: flex;
  flex-wrap: wrap;
  gap: 12rpx;
  margin-top: 20rpx;
}
.summary-fact,
.note-card,
.room-card,
.log-row,
.field-card {
  border-radius: 20rpx;
  background: #f8fafc;
  border: 1rpx solid #e5e7eb;
}
.summary-fact {
  display: inline-flex;
  align-items: center;
  gap: 10rpx;
  padding: 14rpx 18rpx;
  background: #f8fafc;
  border: 1rpx solid #e5e7eb;
}
.summary-fact.accent {
  background: linear-gradient(135deg, #eef2ff, #f8faff);
  border-color: #c7d2fe;
}
.summary-fact-label,
.note-label,
.fact-label,
.field-label {
  font-size: 22rpx;
  color: #94a3b8;
}
.summary-fact-value,
.note-value,
.fact-value {
  font-size: 24rpx;
  color: #111827;
  font-weight: 700;
  line-height: 1.4;
}
.summary-fact-value.price,
.fact-value.price,
.total-value {
  color: #4f46e5;
}
.note-card {
  margin-top: 18rpx;
  padding: 18rpx 20rpx;
}
.note-label,
.field-label {
  display: block;
}
.note-value {
  display: block;
  margin-top: 8rpx;
  font-size: 24rpx;
  font-weight: 500;
  color: #475569;
}
.section-card + .section-card {
  margin-top: 0;
}
.section-title {
  display: block;
  font-size: 30rpx;
  font-weight: 800;
  color: #111827;
}
.mini-link {
  padding: 10rpx 14rpx;
  border-radius: 999rpx;
  background: #eef2ff;
  color: #4f46e5;
  font-size: 22rpx;
  white-space: nowrap;
}
.room-list,
.line-list,
.log-list {
  display: flex;
  flex-direction: column;
  gap: 16rpx;
}
.room-list,
.log-list {
  margin-top: 18rpx;
}
.room-card {
  padding: 20rpx;
}
.room-title {
  display: block;
  font-size: 28rpx;
  font-weight: 800;
  color: #111827;
}
.room-sub {
  display: block;
  margin-top: 8rpx;
  font-size: 22rpx;
  line-height: 1.6;
  color: #6b7280;
}
.room-facts {
  display: flex;
  flex-wrap: wrap;
  gap: 12rpx;
  margin-top: 16rpx;
}
.fact-pill {
  display: inline-flex;
  align-items: center;
  gap: 10rpx;
  min-width: 180rpx;
  padding: 14rpx 16rpx;
  border-radius: 18rpx;
  background: #fff;
  border: 1rpx solid #e5e7eb;
}
.fact-pill.accent {
  background: linear-gradient(135deg, #eef2ff, #f8faff);
  border-color: #c7d2fe;
}
.fact-pill.discount {
  background: #fef2f2;
  border-color: #fecaca;
}
.fact-value.discount,
.line-amount.discount {
  color: #dc2626;
}
.detail-card {
  overflow: hidden;
}
.tab-switch {
  display: inline-flex;
  gap: 8rpx;
  padding: 6rpx;
  border-radius: 999rpx;
  background: #f1f5f9;
}
.tab-pill {
  padding: 10rpx 18rpx;
  border-radius: 999rpx;
  font-size: 22rpx;
  color: #64748b;
  white-space: nowrap;
}
.tab-pill.active {
  background: linear-gradient(135deg, #4f46e5, #6366f1);
  color: #fff;
  box-shadow: 0 8rpx 18rpx rgba(79, 70, 229, 0.18);
}
.detail-toggle {
  margin-top: 16rpx;
}
.detail-toggle-text {
  display: inline-flex;
  padding: 10rpx 14rpx;
  border-radius: 999rpx;
  background: #f1f5f9;
  color: #475569;
  font-size: 22rpx;
}
.compact-line-list {
  margin-top: 12rpx;
  padding: 14rpx 18rpx;
  border-radius: 18rpx;
  background: #fff;
  border: 1rpx dashed #dbe3f0;
}
.line-row {
  display: flex;
  justify-content: space-between;
  gap: 16rpx;
  align-items: center;
  padding: 12rpx 0;
  border-bottom: 1rpx dashed #e5e7eb;
}
.line-row:last-child {
  border-bottom: none;
}
.line-name,
.line-amount {
  font-size: 24rpx;
  color: #374151;
}
.room-actions,
.footer-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 12rpx;
  margin-top: 18rpx;
}
.action-btn {
  flex: 1;
  min-width: 200rpx;
  padding: 18rpx 16rpx;
  border-radius: 18rpx;
  background: #eef2ff;
  color: #4f46e5;
  text-align: center;
  font-size: 24rpx;
  font-weight: 700;
}
.action-btn.primary {
  background: linear-gradient(135deg, #4f46e5, #6366f1);
  color: #fff;
}
.action-btn.danger {
  background: #fef2f2;
  color: #dc2626;
}
.action-btn.large {
  flex: none;
  width: 100%;
}
.total-strip {
  margin-top: 18rpx;
  padding: 18rpx 20rpx;
  border-radius: 20rpx;
  background: linear-gradient(135deg, #eef2ff, #f8faff);
  border: 1rpx solid #c7d2fe;
  align-items: center;
}
.total-label {
  display: block;
  font-size: 22rpx;
  color: #818cf8;
}
.total-value {
  display: block;
  margin-top: 6rpx;
  font-size: 38rpx;
  font-weight: 800;
}
.total-meta {
  font-size: 22rpx;
  line-height: 1.5;
  color: #6b7280;
  text-align: right;
}
.log-row {
  padding: 18rpx;
}
.log-title {
  display: block;
  font-size: 24rpx;
  font-weight: 700;
  color: #111827;
}
.log-meta {
  display: block;
  margin-top: 6rpx;
  font-size: 22rpx;
  color: #9ca3af;
}
.log-content {
  display: block;
  margin-top: 8rpx;
  font-size: 24rpx;
  color: #4b5563;
  line-height: 1.6;
}
.sheet-mask {
  position: fixed;
  inset: 0;
  background: rgba(15, 23, 42, 0.35);
  z-index: 50;
}
.sheet-card {
  position: fixed;
  left: 24rpx;
  right: 24rpx;
  bottom: 58px;
  bottom: calc(58px + env(safe-area-inset-bottom));
  background: #fff;
  border-radius: 24rpx;
  padding: 24rpx;
  box-shadow: 0 24rpx 48rpx rgba(15, 23, 42, 0.24);
  z-index: 51;
}
.sheet-title {
  display: block;
  font-size: 30rpx;
  font-weight: 700;
  color: #111827;
}
.sheet-desc {
  display: block;
  margin-top: 10rpx;
  font-size: 22rpx;
  line-height: 1.6;
  color: #6b7280;
}
.check-row {
  display: flex;
  gap: 16rpx;
  align-items: center;
  padding: 18rpx 0;
}
.check-box {
  width: 36rpx;
  height: 36rpx;
  border-radius: 10rpx;
  border: 2rpx solid #cbd5e1;
  display: flex;
  align-items: center;
  justify-content: center;
}
.check-box.active {
  background: #4f46e5;
  border-color: #4f46e5;
}
.check-mark {
  color: #fff;
  font-size: 22rpx;
}
.check-title {
  display: block;
  font-size: 24rpx;
  font-weight: 600;
  color: #111827;
}
.check-sub {
  display: block;
  margin-top: 6rpx;
  font-size: 22rpx;
  color: #6b7280;
}
.field-card {
  margin-top: 12rpx;
  padding: 18rpx;
}
.sheet-input {
  height: 84rpx;
  padding: 0 18rpx;
  border-radius: 16rpx;
  background: #fff;
  border: 2rpx solid #e5e7eb;
  font-size: 28rpx;
  color: #111827;
}
.sheet-actions {
  margin-top: 20rpx;
}
.sheet-btn {
  flex: 1;
  padding: 20rpx 16rpx;
  border-radius: 18rpx;
  background: #f3f4f6;
  text-align: center;
  color: #374151;
  font-size: 24rpx;
  font-weight: 600;
}
.sheet-btn.primary {
  background: linear-gradient(135deg, #4f46e5, #6366f1);
  color: #fff;
}
@media (max-width: 768px) {
  .page {
    padding: 20rpx 20rpx calc(170rpx + env(safe-area-inset-bottom));
    gap: 16rpx;
  }
  .summary-head,
  .section-headline,
  .total-strip {
    flex-direction: column;
  }
  .summary-side {
    width: 100%;
    flex-direction: row;
    justify-content: space-between;
    align-items: center;
  }
  .summary-btn {
    padding: 14rpx 18rpx;
  }
  .fact-pill {
    min-width: calc(50% - 6rpx);
    box-sizing: border-box;
  }
  .action-btn {
    min-width: calc(50% - 6rpx);
  }
  .total-meta {
    text-align: left;
  }
  .summary-fact {
    width: calc(50% - 6rpx);
    box-sizing: border-box;
    justify-content: space-between;
  }
  .tab-switch {
    width: 100%;
    box-sizing: border-box;
  }
  .tab-pill {
    flex: 1;
    text-align: center;
  }
}
</style>
