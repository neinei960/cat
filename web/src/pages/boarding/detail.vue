<template>
  <SideLayout>
    <view class="page">
      <view v-if="loading" class="state">加载中...</view>
      <view v-else-if="!order" class="state">寄养订单不存在</view>
      <template v-else>
        <view class="card summary-card">
          <view class="summary-head">
            <view class="summary-copy">
              <text class="summary-caption">{{ isHistoryOrder ? '历史寄养详情' : '寄养详情' }}</text>
              <text class="summary-title">{{ customerLabel }}</text>
              <text class="summary-sub">{{ allPetNames }} · {{ roomSummary }}</text>
            </view>
            <view class="summary-side">
              <text :class="['status-pill', order.status, 'summary-status']">{{ statusLabel(order.status) }}</text>
              <view v-if="canAddProducts" class="btn summary-btn" @click="goAddProduct">添加商品</view>
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
            <view :class="['summary-fact', 'action', 'switch-action', updatingDeworming ? 'disabled' : '']">
              <text class="summary-fact-label">驱虫</text>
              <view class="summary-fact-main">
                <text class="summary-fact-value">{{ dewormingLabel(order.has_deworming) }}</text>
                <switch
                  class="summary-switch"
                  color="#4f46e5"
                  :disabled="updatingDeworming"
                  :checked="order.has_deworming === true"
                  @change="handleDewormingSwitchChange"
                />
              </view>
            </view>
            <view class="summary-fact accent">
              <text class="summary-fact-label">应收</text>
              <text class="summary-fact-value price">¥{{ displayPayAmount.toFixed(2) }}</text>
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
              <text class="section-title">家长与猫咪</text>
              <text class="section-subtitle">回看本次寄养对应的家长信息和猫咪基础档案。</text>
            </view>
          </view>

          <view class="profile-grid">
            <view class="profile-panel">
              <text class="profile-title">家长信息</text>
              <view class="profile-row">
                <text class="profile-label">家长</text>
                <text class="profile-value">{{ customerLabel }}</text>
              </view>
              <view class="profile-row">
                <text class="profile-label">手机号</text>
                <text class="profile-value">{{ customerPhone }}</text>
              </view>
              <view v-if="customerRemark" class="profile-row block">
                <text class="profile-label">家长备注</text>
                <text class="profile-value multiline">{{ customerRemark }}</text>
              </view>
            </view>

            <view class="profile-panel">
              <text class="profile-title">猫咪信息</text>
              <view class="pet-profile-list">
                <view v-for="pet in petProfiles" :key="pet.name" class="pet-profile-row">
                  <text class="pet-profile-name">{{ pet.name }}</text>
                  <text class="pet-profile-meta">{{ pet.meta || '未补充档案' }}</text>
                </view>
              </view>
            </view>
          </view>
        </view>

        <view class="card section-card">
          <view class="section-headline">
            <view>
              <text class="section-title">房间安排</text>
              <text class="section-subtitle">{{ isHistoryOrder ? '回看每个房间的猫咪、日期和金额，明细按需展开。' : '先看每个房间的猫咪、日期和金额，明细按需展开。' }}</text>
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
                <view v-if="roomPreview(room)?.special_item_name || room.special_item_name" class="fact-pill">
                  <text class="fact-label">特殊寄养项目</text>
                  <text class="fact-value">{{ roomPreview(room)?.special_item_name || room.special_item_name }}</text>
                </view>
                <view v-if="(roomPreview(room)?.special_item_days || room.special_item_days) > 0" class="fact-pill">
                  <text class="fact-label">特殊天数</text>
                  <text class="fact-value">{{ roomPreview(room)?.special_item_days || room.special_item_days }} 天 · ¥{{ Number(roomPreview(room)?.special_item_daily_price || room.special_item_daily_price || 0).toFixed(2) }}/天</text>
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
                <view v-if="room.status === 'pending_checkin'" class="action-btn" @click="openAdjustPrice(room)">调整价格</view>
                <view v-if="room.status === 'pending_checkin'" class="action-btn danger" @click="handleCancel(room)">取消</view>
                <view v-if="room.status === 'checked_in'" class="action-btn" @click="openAdjustPrice(room)">调整价格</view>
                <view v-if="canExtendRoom(room)" class="action-btn" @click="handleExtend(room)">续住</view>
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
            <view v-if="displayLogs.length > 0" class="tab-switch">
              <view :class="['tab-pill', detailTab === 'amount' ? 'active' : '']" @click="detailTab = 'amount'">金额</view>
              <view :class="['tab-pill', detailTab === 'logs' ? 'active' : '']" @click="detailTab = 'logs'">日志</view>
            </view>
          </view>

          <template v-if="detailTab === 'amount' || displayLogs.length === 0">
            <view class="total-strip">
              <view>
                <text class="total-label">整单应收</text>
                <text class="total-value">¥{{ displayPayAmount.toFixed(2) }}</text>
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
            <view v-for="log in displayLogs" :key="log.id" class="log-row">
              <text class="log-title">{{ log.title }}</text>
              <text class="log-meta">{{ log.meta }}</text>
              <text class="log-content">{{ log.content }}</text>
            </view>
          </view>
        </view>

        <view v-if="canCancelWholeOrder" class="footer-actions">
          <view class="action-btn danger large" @click="handleCancelWholeOrder">整单取消</view>
        </view>

        <view v-if="showCheckInSheet" class="sheet-mask" @click="closeCheckInSheet" @touchmove.stop.prevent></view>
        <view v-if="showCheckInSheet" class="sheet-card" @click.stop @touchmove.stop>
          <text class="sheet-title">{{ activeRoom ? `${roomLabel(activeRoom)} ${checkInSheetMode === 'adjust_price' ? '调整价格' : '办理入住'}` : (checkInSheetMode === 'adjust_price' ? '调整价格' : '办理入住') }}</text>
          <text class="sheet-desc">{{ checkInSheetMode === 'adjust_price' ? '调整这个房间的优惠金额，整单收款会自动同步。' : '入住优惠按房间单独录入，整单收款会自动同步。' }}</text>
          <view class="sheet-scroll" @touchstart="onSheetScrollTouchStart" @touchmove="onSheetScrollTouchMove">
            <view class="sheet-scroll-inner">
              <view class="field-card">
                <text class="field-label">特殊寄养项目</text>
                <text class="field-tip">可同时选择多个加收项目，分别填写日价和天数。</text>
                <view class="special-option-list">
                  <view
                    v-for="item in editableSpecialItems"
                    :key="item.ID"
                    class="special-option-wrap"
                  >
                    <view
                      :class="['special-option', isSheetSpecialItemSelected(item.ID) ? 'active' : '']"
                      @click="toggleSheetSpecialItem(item)"
                    >
                      <view>
                        <text class="special-option-name">{{ item.name }}</text>
                        <text class="special-option-meta">默认 ¥{{ Number(item.default_daily_price || 0).toFixed(2) }}/天</text>
                      </view>
                      <text class="special-option-mark">{{ isSheetSpecialItemSelected(item.ID) ? '已选' : '选择' }}</text>
                    </view>
                    <view
                      v-if="sheetSpecialItemSelection(item.ID)"
                      class="special-input-grid"
                    >
                      <view class="field-card compact">
                        <text class="field-label">特殊日价</text>
                        <input
                          :value="sheetSpecialItemSelection(item.ID)?.dailyPrice"
                          class="sheet-input"
                          type="digit"
                          placeholder="例如：30"
                          @input="updateSheetSpecialItem(item.ID, 'dailyPrice', $event)"
                        />
                      </view>
                      <view class="field-card compact">
                        <text class="field-label">特殊天数</text>
                        <input
                          :value="sheetSpecialItemSelection(item.ID)?.days"
                          class="sheet-input"
                          type="number"
                          placeholder="例如：2"
                          @input="updateSheetSpecialItem(item.ID, 'days', $event)"
                        />
                      </view>
                      <view class="special-day-actions">
                        <view class="special-day-fill" @click.stop="fillSheetSpecialItemEveryDay(item.ID)">每天</view>
                        <text class="field-tip">按全部寄养晚数自动填写，也可以手动改天数。</text>
                      </view>
                    </view>
                  </view>
                </view>
              </view>
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
            </view>
          </view>
          <view class="sheet-actions" @touchmove.stop.prevent>
            <view class="sheet-btn" @click="closeCheckInSheet">取消</view>
            <view class="sheet-btn primary" @click="submitCheckIn">{{ checkInSheetMode === 'adjust_price' ? '保存调整' : '确认入住' }}</view>
          </view>
        </view>
      </template>
    </view>
  </SideLayout>
</template>

<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, ref, watch } from 'vue'
import { onLoad, onShow } from '@dcloudio/uni-app'
import SideLayout from '@/components/SideLayout.vue'
import {
  adjustBoardingOrderPrice,
  adjustBoardingRoomPrice,
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
  getBoardingSpecialItems,
  updateBoardingOrderDeworming,
} from '@/api/boarding'
import {
  buildBoardingHistoryPetProfiles,
  getBoardingHistoryCustomerLabel,
  getBoardingHistoryRoomSummary,
} from '@/utils/boarding-history'

const id = ref(0)
const loading = ref(false)
const order = ref<BoardingOrder | null>(null)
const showCheckInSheet = ref(false)
const useDiscount = ref(false)
const discountAmountInput = ref('')
const specialItems = ref<BoardingSpecialItem[]>([])
type SpecialItemSelectionState = { id: number; dailyPrice: string; days: string }
const specialItemSelectionsInput = ref<SpecialItemSelectionState[]>([])
const activeRoom = ref<BoardingOrderRoom | null>(null)
const checkInSheetMode = ref<'check_in' | 'adjust_price'>('check_in')
const roomLineOpen = ref<Record<string, boolean>>({})
const detailTab = ref<'amount' | 'logs'>('amount')
const updatingDeworming = ref(false)
let sheetTouchStartY = 0
let pageLockState: {
  scrollTop: number
  bodyPosition: string
  bodyTop: string
  bodyLeft: string
  bodyRight: string
  bodyWidth: string
  bodyOverflow: string
  htmlOverflow: string
} | null = null

function lockPageScroll() {
  if (typeof window === 'undefined' || typeof document === 'undefined' || pageLockState) return
  const body = document.body
  const html = document.documentElement
  const scrollTop = window.scrollY || html.scrollTop || body.scrollTop || 0
  pageLockState = {
    scrollTop,
    bodyPosition: body.style.position,
    bodyTop: body.style.top,
    bodyLeft: body.style.left,
    bodyRight: body.style.right,
    bodyWidth: body.style.width,
    bodyOverflow: body.style.overflow,
    htmlOverflow: html.style.overflow,
  }
  body.style.position = 'fixed'
  body.style.top = `-${scrollTop}px`
  body.style.left = '0'
  body.style.right = '0'
  body.style.width = '100%'
  body.style.overflow = 'hidden'
  html.style.overflow = 'hidden'
}

function unlockPageScroll() {
  if (typeof window === 'undefined' || typeof document === 'undefined' || !pageLockState) return
  const body = document.body
  const html = document.documentElement
  const restoreTop = pageLockState.scrollTop
  body.style.position = pageLockState.bodyPosition
  body.style.top = pageLockState.bodyTop
  body.style.left = pageLockState.bodyLeft
  body.style.right = pageLockState.bodyRight
  body.style.width = pageLockState.bodyWidth
  body.style.overflow = pageLockState.bodyOverflow
  html.style.overflow = pageLockState.htmlOverflow
  pageLockState = null
  window.scrollTo(0, restoreTop)
}

function onSheetScrollTouchStart(event: TouchEvent) {
  sheetTouchStartY = event.touches?.[0]?.clientY || 0
}

function onSheetScrollTouchMove(event: TouchEvent) {
  const scrollEl = event.currentTarget as HTMLElement | null
  if (!scrollEl) return
  const currentY = event.touches?.[0]?.clientY || 0
  const deltaY = currentY - sheetTouchStartY
  const canScroll = scrollEl.scrollHeight > scrollEl.clientHeight
  const atTop = scrollEl.scrollTop <= 0
  const atBottom = scrollEl.scrollTop + scrollEl.clientHeight >= scrollEl.scrollHeight - 1

  if (!canScroll || (atTop && deltaY > 0) || (atBottom && deltaY < 0)) {
    event.preventDefault()
  }
  event.stopPropagation()
}

watch(showCheckInSheet, async (visible) => {
  if (visible) {
    await nextTick()
    lockPageScroll()
    return
  }
  unlockPageScroll()
})

onBeforeUnmount(unlockPageScroll)

const aggregatePreview = computed<BoardingPricePreview | null>(() => {
  try {
    return JSON.parse(order.value?.price_snapshot_json || '{}') || null
  } catch {
    return null
  }
})

const displayRooms = computed(() => order.value?.rooms || [])
const linkedOrderRetailLines = computed(() => {
  const linkedOrder = order.value?.order
  if (!linkedOrder) return [] as Array<{ type: string; label: string; amount: number }>

  const lines: Array<{ type: string; label: string; amount: number }> = []
  for (const item of linkedOrder.items || []) {
    if (item.item_type !== 2) continue
    lines.push({
      type: 'product',
      label: item.name,
      amount: Number(item.amount || Number(item.unit_price || 0) * Number(item.quantity || 1)),
    })
  }
  const productDiscountAmount = Number(linkedOrder.product_discount_amount || 0)
  if (productDiscountAmount > 0) {
    lines.push({
      type: 'product_discount',
      label: '商品优惠',
      amount: -productDiscountAmount,
    })
  }
  return lines
})
const aggregateLines = computed(() => [...(aggregatePreview.value?.lines || []), ...linkedOrderRetailLines.value])
const logs = computed(() => order.value?.logs || [])
const displayLogs = computed(() => {
  const rows = logs.value
    .filter((log) => !['check_in', 'check_out'].includes(log.action))
    .map((log) => ({
      id: `log-${log.ID}`,
      title: actionLabel(log.action),
      meta: `${log.operator?.name || '-'} · ${formatTime(log.CreatedAt)}`,
      content: log.content,
      sortTime: log.CreatedAt || '',
    }))

  const paymentLogs = order.value?.payment_logs || []
  if (paymentLogs.length > 0) {
    paymentLogs.forEach((payment) => {
      rows.push({
        id: `payment-${payment.order_id}`,
        title: '支付',
        meta: `收款 · ${formatTime(payment.pay_time)}`,
        content: `支付金额 ¥${Number(payment.pay_amount || 0).toFixed(2)}`,
        sortTime: payment.pay_time || '',
      })
    })
  } else {
    const linkedOrder = order.value?.order
    if (!linkedOrder || Number(linkedOrder.pay_status || 0) !== 1) {
      return rows.sort((a, b) => (a.sortTime || '').localeCompare(b.sortTime || ''))
    }
    const payTime = linkedOrder.pay_time || linkedOrder.CreatedAt || ''
    rows.push({
      id: `payment-${linkedOrder.ID}`,
      title: '支付',
      meta: `收款 · ${formatTime(payTime)}`,
      content: `支付金额 ¥${Number(linkedOrder.pay_amount || 0).toFixed(2)}`,
      sortTime: payTime,
    })
  }

  return rows.sort((a, b) => (a.sortTime || '').localeCompare(b.sortTime || ''))
})
const allPetNames = computed(() => order.value?.pets?.map((item) => item.pet?.name || item.pet_name_snapshot).filter(Boolean).join('、') || '-')
const canCancelWholeOrder = computed(() => displayRooms.value.length > 1 && displayRooms.value.every((room) => room.status === 'pending_checkin'))
const isHistoryOrder = computed(() => order.value?.status === 'checked_out')
const displayPayAmount = computed(() => {
  const linkedOrderPayAmount = Number(order.value?.order?.pay_amount || 0)
  if (order.value?.order) return linkedOrderPayAmount
  return Number(order.value?.pay_amount || 0)
})
const customerLabel = computed(() => getBoardingHistoryCustomerLabel(order.value))
const customerPhone = computed(() => order.value?.customer?.phone || '-')
const customerRemark = computed(() => (order.value?.customer?.remark || '').trim())
const roomSummary = computed(() => getBoardingHistoryRoomSummary(order.value))
const petProfiles = computed(() => buildBoardingHistoryPetProfiles(order.value))
const canAddProducts = computed(() => {
  const boarding = order.value
  if (!boarding?.order_id) return false
  const linkedOrder = boarding.order
  if (!linkedOrder) return true
  if ([2, 3].includes(Number(linkedOrder.status || 0))) return false
  return Number(linkedOrder.pay_status || 0) === 0
})
const editableSpecialItems = computed(() => {
  const list = [...specialItems.value]
  for (const selection of specialItemSelectionsInput.value) {
    if (!selection.id || list.some((item) => item.ID === selection.id)) continue
    list.unshift({
      ID: selection.id,
      shop_id: order.value?.shop_id || 0,
      name: activeRoom.value?.special_item_name || roomPreview(activeRoom.value || undefined)?.special_item_name || '特殊寄养项目',
      default_daily_price: Number(selection.dailyPrice || 0),
      sort_order: -1,
      status: 0,
      remark: '',
    })
  }
  return dedupeSpecialItems(list)
})

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
    adjust_price: '调价',
    update_deworming: '驱虫',
    check_out: '离店',
    extend: '续住',
    change_cabinet: '换房型',
    cancel: '取消',
  }[action] || action
}

function dewormingLabel(value?: boolean | null) {
  if (value === true) return '已驱虫'
  return '未驱虫'
}

function roomLabel(room: BoardingOrderRoom) {
  return `房间${room.room_index || 1}`
}

function canExtendRoom(room: BoardingOrderRoom) {
  return room.status === 'checked_in' || room.status === 'checked_out'
}

function roomPetNames(room: BoardingOrderRoom) {
  return room.pets?.map((item) => item.pet?.name || item.pet_name_snapshot).filter(Boolean).join('、') || '未选猫咪'
}

function roomPreview(room?: BoardingOrderRoom | null) {
  if (!room) return null
  const aggregateRoom = aggregatePreview.value?.rooms?.find((item) => item.room_index === room.room_index)
  if (aggregateRoom) return aggregateRoom
  try {
    const raw = JSON.parse(room.price_snapshot_json || '{}') as BoardingPricePreview
    return {
      room_index: room.room_index,
      cabinet_id: room.cabinet_id,
      cabinet_type: room.cabinet?.cabinet_type || '',
      pet_count: room.pets?.length || 1,
      special_item_id: raw.special_item_id || room.special_item_id,
      special_item_name: raw.special_item_name || room.special_item_name || '',
      special_item_daily_price: raw.special_item_daily_price || room.special_item_daily_price || 0,
      special_item_days: raw.special_item_days || room.special_item_days || 0,
      check_in_at: room.check_in_at,
      check_out_at: room.check_out_at,
      nights: raw.nights || room.nights,
      regular_nights: raw.regular_nights || 0,
      holiday_nights: raw.holiday_nights || 0,
      base_amount: raw.base_amount || room.base_amount,
      extra_pet_amount: raw.extra_pet_amount || 0,
      holiday_surcharge_amount: raw.holiday_surcharge_amount || room.holiday_surcharge_amount,
      special_item_amount: raw.special_item_amount || room.special_item_amount || 0,
      discount_amount: raw.discount_amount || room.discount_amount,
      manual_discount_amount: room.manual_discount_amount,
      pay_amount: Math.max((raw.pay_amount || room.pay_amount) - (room.manual_discount_amount || 0), 0),
      lines: raw.lines || [],
    } as BoardingRoomPreview
  } catch {
    return null
  }
}

function specialSelectionsFromRoom(room: BoardingOrderRoom): SpecialItemSelectionState[] {
  const preview = roomPreview(room)
  const lineSelections = (preview?.lines || [])
    .filter((line) => line.type === 'special_item')
    .map((line) => ({
      id: Number(line.special_item_id || 0),
      dailyPrice: String(line.unit_price || ''),
      days: String(line.quantity || ''),
      label: line.label,
    }))
  if (lineSelections.length > 0 && lineSelections.every((item) => item.id > 0)) {
    return lineSelections.map(({ id, dailyPrice, days }) => ({ id, dailyPrice, days }))
  }
  const legacyID = Number(preview?.special_item_id || room.special_item_id || 0)
  if (!legacyID) return []
  const days = String(preview?.special_item_days || room.special_item_days || '')
  return [{
    id: legacyID,
    dailyPrice: String(preview?.special_item_daily_price || room.special_item_daily_price || ''),
    days,
  }]
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

function specialItemLabel(id: number) {
  return editableSpecialItems.value.find((item) => item.ID === id)?.name || '特殊寄养项目'
}

function normalizeSpecialItemName(name?: string) {
  return String(name || '').trim()
}

function dedupeSpecialItems(items: BoardingSpecialItem[]) {
  const byName = new Map<string, BoardingSpecialItem>()
  for (const item of items) {
    const key = normalizeSpecialItemName(item.name) || String(item.ID)
    const existing = byName.get(key)
    if (!existing || Number(item.ID || 0) > Number(existing.ID || 0)) {
      byName.set(key, item)
    }
  }
  return Array.from(byName.values())
}

function isSheetSpecialItemSelected(itemId: number) {
  return specialItemSelectionsInput.value.some((item) => item.id === itemId)
}

function sheetSpecialItemSelection(itemId: number) {
  return specialItemSelectionsInput.value.find((item) => item.id === itemId)
}

function toggleSheetSpecialItem(item: BoardingSpecialItem) {
  if (isSheetSpecialItemSelected(item.ID)) {
    specialItemSelectionsInput.value = specialItemSelectionsInput.value.filter((selection) => selection.id !== item.ID)
    return
  }
  const itemName = normalizeSpecialItemName(item.name)
  specialItemSelectionsInput.value = [
    ...specialItemSelectionsInput.value.filter((selection) => normalizeSpecialItemName(specialItemLabel(selection.id)) !== itemName),
    { id: item.ID, dailyPrice: String(item.default_daily_price || ''), days: '' },
  ]
}

function updateSheetSpecialItem(itemId: number, field: 'dailyPrice' | 'days', event: any) {
  const value = event?.detail?.value ?? ''
  specialItemSelectionsInput.value = specialItemSelectionsInput.value.map((item) => (
    item.id === itemId ? { ...item, [field]: value } : item
  ))
}

function fillSheetSpecialItemEveryDay(itemId: number) {
  const nights = Number(roomPreview(activeRoom.value)?.nights || activeRoom.value?.nights || 0)
  if (nights <= 0) {
    uni.showToast({ title: '当前房间没有可用寄养晚数', icon: 'none' })
    return
  }
  specialItemSelectionsInput.value = specialItemSelectionsInput.value.map((item) => (
    item.id === itemId ? { ...item, days: String(nights) } : item
  ))
}

async function loadSpecialItems() {
  const res = await getBoardingSpecialItems({ active_only: 1 })
  specialItems.value = (res.data || []).filter((item) => item.status === 1)
}

function closeCheckInSheet() {
  showCheckInSheet.value = false
  activeRoom.value = null
  checkInSheetMode.value = 'check_in'
  specialItemSelectionsInput.value = []
}

function toggleDiscount() {
  useDiscount.value = !useDiscount.value
  if (!useDiscount.value) discountAmountInput.value = ''
}

function goOrderDetail() {
  if (!order.value?.order_id) return
  uni.navigateTo({ url: `/pages/order/detail?id=${order.value.order_id}` })
}

function goAddProduct() {
  if (!order.value?.order_id) return
  uni.navigateTo({ url: `/pages/order/create?order_id=${order.value.order_id}` })
}

async function handleDewormingSwitchChange(e: any) {
  if (!order.value || updatingDeworming.value) return
  const target = Boolean(e?.detail?.value)
  const current = order.value.has_deworming === true
  if (current === target) return
  updatingDeworming.value = true
  uni.showLoading({ title: '保存中', mask: true })
  let updated = false
  try {
    await updateBoardingOrderDeworming(order.value.ID, target)
    await loadData()
    updated = true
  } finally {
    updatingDeworming.value = false
    uni.hideLoading()
  }
  if (updated) {
    uni.showToast({ title: '驱虫状态已更新', icon: 'success' })
  }
}

async function loadData() {
  if (!id.value) return
  loading.value = true
  try {
    const [orderRes] = await Promise.all([
      getBoardingOrder(id.value),
      loadSpecialItems(),
    ])
    const res = orderRes
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
  checkInSheetMode.value = 'check_in'
  useDiscount.value = Number(room.manual_discount_amount || 0) > 0
  discountAmountInput.value = useDiscount.value ? Number(room.manual_discount_amount || 0).toFixed(2) : ''
  specialItemSelectionsInput.value = specialSelectionsFromRoom(room)
  showCheckInSheet.value = true
}

function openAdjustPrice(room: BoardingOrderRoom) {
  activeRoom.value = room
  checkInSheetMode.value = 'adjust_price'
  useDiscount.value = Number(room.manual_discount_amount || 0) > 0
  discountAmountInput.value = useDiscount.value ? Number(room.manual_discount_amount || 0).toFixed(2) : ''
  specialItemSelectionsInput.value = specialSelectionsFromRoom(room)
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
  const payload: { discount_amount?: number; special_item_id?: number; special_item_daily_price?: number; special_item_days?: number; special_items?: Array<{ id: number; daily_price: number; days: number }> } = {
    discount_amount: discountAmount,
    special_items: [],
  }
  if (specialItemSelectionsInput.value.length > 0) {
    const nights = Number(roomPreview(activeRoom.value)?.nights || activeRoom.value.nights || 0)
    for (const item of specialItemSelectionsInput.value) {
      const specialItemDailyPrice = Number(item.dailyPrice || 0)
      const specialItemDays = Number(item.days || 0)
      const itemName = specialItemLabel(item.id)
      if (!Number.isFinite(specialItemDailyPrice) || specialItemDailyPrice <= 0) {
        uni.showToast({ title: `${itemName} 请填写有效日价`, icon: 'none' })
        return
      }
      if (!Number.isInteger(specialItemDays) || specialItemDays < 1) {
        uni.showToast({ title: `${itemName} 请填写有效天数`, icon: 'none' })
        return
      }
      if (nights > 0 && specialItemDays > nights) {
        uni.showToast({ title: `${itemName} 天数不能超过寄养晚数`, icon: 'none' })
        return
      }
    }
    payload.special_items = specialItemSelectionsInput.value.map((item) => ({
      id: item.id,
      daily_price: Number(item.dailyPrice || 0),
      days: Number(item.days || 0),
    }))
  } else {
    payload.special_item_id = 0
    payload.special_item_daily_price = 0
    payload.special_item_days = 0
  }
  if (checkInSheetMode.value === 'adjust_price') {
    if (activeRoom.value.ID > 0) {
      await adjustBoardingRoomPrice(order.value.ID, activeRoom.value.ID, payload)
    } else {
      await adjustBoardingOrderPrice(order.value.ID, payload)
    }
    closeCheckInSheet()
    uni.showToast({ title: '价格已调整', icon: 'success' })
    await loadData()
    return
  }
  if (activeRoom.value.ID > 0) {
    await checkInBoardingRoom(order.value.ID, activeRoom.value.ID, payload)
  } else {
    await checkInBoardingOrder(order.value.ID, payload)
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
.summary-fact.action {
  justify-content: space-between;
  min-width: 260rpx;
}
.summary-fact.action.disabled {
  opacity: 0.72;
}
.summary-fact.switch-action {
  padding-right: 14rpx;
}
.summary-fact-main {
  display: inline-flex;
  align-items: center;
  gap: 10rpx;
}
.summary-switch {
  flex: 0 0 auto;
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
.summary-fact-hint {
  font-size: 22rpx;
  color: #4f46e5;
  font-weight: 600;
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
.profile-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 16rpx;
  margin-top: 18rpx;
}
.profile-panel {
  border-radius: 22rpx;
  background: #f8fafc;
  border: 1rpx solid #e5e7eb;
  padding: 20rpx;
}
.profile-title {
  display: block;
  font-size: 26rpx;
  font-weight: 800;
  color: #111827;
}
.profile-row {
  display: flex;
  justify-content: space-between;
  gap: 16rpx;
  align-items: flex-start;
  padding-top: 16rpx;
}
.profile-row.block {
  display: block;
}
.profile-label {
  font-size: 22rpx;
  color: #94a3b8;
  white-space: nowrap;
}
.profile-value {
  font-size: 24rpx;
  color: #111827;
  font-weight: 700;
  line-height: 1.5;
  text-align: right;
}
.profile-value.multiline {
  display: block;
  margin-top: 8rpx;
  text-align: left;
  color: #475569;
  font-weight: 500;
}
.pet-profile-list {
  display: flex;
  flex-direction: column;
  gap: 12rpx;
  margin-top: 16rpx;
}
.pet-profile-row {
  padding: 16rpx 18rpx;
  border-radius: 18rpx;
  background: rgba(255, 255, 255, 0.88);
  border: 1rpx solid #e5e7eb;
}
.pet-profile-name {
  display: block;
  font-size: 24rpx;
  font-weight: 700;
  color: #111827;
}
.pet-profile-meta {
  display: block;
  margin-top: 8rpx;
  font-size: 22rpx;
  line-height: 1.5;
  color: #64748b;
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
  max-height: calc(100vh - 112px - env(safe-area-inset-bottom));
  background: #fff;
  border-radius: 24rpx;
  padding: 24rpx;
  box-shadow: 0 24rpx 48rpx rgba(15, 23, 42, 0.24);
  z-index: 51;
  box-sizing: border-box;
  display: flex;
  flex-direction: column;
  overflow: hidden;
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
  flex-shrink: 0;
}
.sheet-scroll {
  flex: 1;
  min-height: 0;
  margin-top: 12rpx;
  overflow-x: hidden;
  overflow-y: auto;
  -webkit-overflow-scrolling: touch;
  overscroll-behavior: contain;
  touch-action: pan-y;
}
.sheet-scroll-inner {
  padding-bottom: 8rpx;
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
.sheet-scroll-inner > .field-card:first-child {
  margin-top: 0;
}
.field-tip {
  display: block;
  margin-top: 6rpx;
  font-size: 22rpx;
  color: #6b7280;
  line-height: 1.5;
}
.special-option-list {
  display: flex;
  flex-direction: column;
  gap: 12rpx;
  margin-top: 14rpx;
}
.special-option-wrap {
  display: flex;
  flex-direction: column;
  gap: 10rpx;
}
.special-option {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16rpx;
  padding: 18rpx;
  border-radius: 18rpx;
  border: 1rpx solid #e5e7eb;
  background: #fff;
}
.special-option.active {
  border-color: #4f46e5;
  background: #eef2ff;
}
.special-option-name,
.special-selected-name {
  display: block;
  font-size: 25rpx;
  font-weight: 600;
  color: #111827;
}
.special-option-meta {
  display: block;
  margin-top: 6rpx;
  font-size: 22rpx;
  color: #6b7280;
}
.special-option-mark {
  flex-shrink: 0;
  padding: 8rpx 14rpx;
  border-radius: 999rpx;
  background: #eef2ff;
  color: #4f46e5;
  font-size: 22rpx;
}
.special-option.active .special-option-mark {
  background: #4f46e5;
  color: #fff;
}
.special-input-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 12rpx;
  padding: 12rpx;
  border-radius: 18rpx;
  background: #f8faff;
  border: 1rpx solid #e0e7ff;
}
.special-input-grid .field-card {
  margin-top: 0;
}
.special-day-actions {
  grid-column: 1 / -1;
  display: flex;
  align-items: center;
  gap: 14rpx;
  padding: 4rpx 2rpx 2rpx;
}
.special-day-fill {
  flex-shrink: 0;
  padding: 12rpx 22rpx;
  border-radius: 999rpx;
  background: #4f46e5;
  color: #fff;
  font-size: 23rpx;
  font-weight: 600;
}
.sheet-grid {
  display: grid;
  grid-template-columns: 1.2fr repeat(2, minmax(0, 1fr));
  gap: 12rpx;
  margin-top: 12rpx;
}
.sheet-grid .field-card {
  margin-top: 0;
}
.sheet-picker {
  min-height: 84rpx;
  padding: 0 18rpx;
  border-radius: 16rpx;
  background: #fff;
  border: 2rpx solid #e5e7eb;
  display: flex;
  align-items: center;
}
.sheet-picker-value {
  font-size: 28rpx;
  color: #111827;
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
  flex-shrink: 0;
  padding-top: 16rpx;
  border-top: 1rpx solid #eef2f7;
  background: #fff;
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
  .sheet-grid {
    grid-template-columns: repeat(1, minmax(0, 1fr));
  }
  .profile-grid {
    grid-template-columns: minmax(0, 1fr);
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
