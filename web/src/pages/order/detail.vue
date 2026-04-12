<template>
  <SideLayout>
  <view class="page" v-if="order">
    <view :class="['status-bar', `s${order.status}`]">
      <text class="status-text">{{ statusMap[order.status] }}</text>
    </view>
    <view v-if="isDeletedView" class="deleted-banner">
      <text>该订单当前位于回收站中，2 天内可恢复。</text>
    </view>

    <view class="card">
      <view class="row"><text class="label">订单号</text><text>{{ order.order_no }}</text></view>
      <view class="row"><text class="label">客户</text><text>{{ order.customer?.nickname || '-' }}</text></view>
      <view class="row">
        <text class="label">{{ order.order_kind === 'product' ? '订单类型' : '猫咪' }}</text>
        <text
          :class="['row-value', primaryPetId ? 'pet-link' : '']"
          @click="goPetDetail(primaryPetId)"
        >{{ order.order_kind === 'product' ? '商品零售' : (order.pet_summary || order.pet?.name || '-') }}</text>
      </view>
      <view class="row"><text class="label">经手员工</text><text>{{ order.staff?.name || '-' }}</text></view>
      <view class="row" v-if="order.pay_method"><text class="label">支付方式</text><text>{{ payMethodMap[order.pay_method] || order.pay_method }}</text></view>
      <view class="row" v-if="order.pay_time"><text class="label">支付时间</text><text>{{ formatDateTime(order.pay_time) }}</text></view>
    </view>

    <view class="card">
      <text class="card-title">明细</text>
      <view v-for="(group, groupIndex) in petGroups" :key="`${group.pet_name}-${groupIndex}`" class="pet-group">
        <view class="pet-group-head">
          <text
            :class="['pet-group-name', group.pet_id ? 'pet-link' : '', group.pet_name === '零售商品' ? 'group-retail' : '']"
            @click="goPetDetail(group.pet_id)"
          >{{ group.pet_name === '零售商品' ? '📦' : '🐱' }} {{ group.pet_name }}</text>
          <text class="pet-group-count">{{ group.items.length }}项</text>
        </view>
        <view class="item-row" v-for="item in group.items" :key="item.ID">
          <text class="item-name">{{ item.name }}</text>
          <text class="item-qty">x{{ item.quantity }}</text>
          <text class="item-amount">¥{{ item.amount }}</text>
        </view>
      </view>
      <view class="total-section">
        <view class="total-row" v-if="serviceTotalValue > 0"><text>服务小计</text><text>¥{{ serviceTotalValue.toFixed(2) }}</text></view>
        <view class="total-row" v-if="serviceDiscountValue > 0"><text>服务优惠</text><text class="discount-text">-¥{{ serviceDiscountValue.toFixed(2) }}</text></view>
        <view class="total-row" v-if="productTotalValue > 0"><text>商品小计</text><text>¥{{ productTotalValue.toFixed(2) }}</text></view>
        <view class="total-row" v-if="productDiscountValue > 0"><text>商品优惠</text><text class="discount-text">-¥{{ productDiscountValue.toFixed(2) }}</text></view>
        <view class="total-row" v-if="addonTotalValue > 0"><text>附加费</text><text>¥{{ addonTotalValue.toFixed(2) }}</text></view>
        <view class="total-row"><text>总计</text><text>¥{{ order.total_amount }}</text></view>
        <view class="total-row" v-if="order.discount_amount"><text>优惠</text><text class="discount-text">-¥{{ order.discount_amount }}</text></view>
        <view class="total-row final"><text>应付</text><text class="pay-amount">¥{{ order.pay_amount }}</text></view>
        <view class="remark-block">
          <view class="remark-head">
            <text class="remark-title">备注</text>
            <text class="remark-save" @click="saveRemark">{{ savingRemark ? '保存中...' : '保存' }}</text>
          </view>
          <textarea
            v-model="remarkDraft"
            class="remark-input"
            maxlength="200"
            auto-height
            placeholder="备注收款说明、客户要求或补充信息"
          />
        </view>
      </view>
    </view>

    <view class="actions">
      <button v-if="canEditPrice" class="btn edit" @click="goEditOrder">修改订单</button>
      <button v-if="order.status === 0 && !isDeletedView" class="btn pay" @click="openPayModal">收款</button>
      <button v-if="order.status === 0 && isAdmin && !isDeletedView" class="btn cancel" @click="doCancel">取消订单</button>
      <button v-if="order.status === 1 && isAdmin && !isDeletedView" class="btn refund" @click="doRefund">退款</button>
      <button v-if="(order.status === 1 || order.status === 2 || order.status === 3) && isAdmin && !isDeletedView" class="btn delete" @click="doDelete">删除订单</button>
      <button class="btn receipt" @click="showReceipt = true">生成小票</button>
    </view>

    <!-- 小票弹窗 -->
    <view class="modal-mask" v-if="showReceipt" @click="showReceipt = false">
      <view class="receipt-outer" @click.stop>
      <view class="receipt-wrap">
        <view class="receipt" id="receiptContent">

          <!-- 头部 Logo 区域 -->
          <view class="receipt-header">
            <view class="receipt-logo">
              <image
                v-if="!logoError"
                class="receipt-logo-img"
                :src="logoSrc"
                mode="aspectFill"
                @error="logoError = true"
              />
              <text v-if="logoError" class="receipt-logo-emoji">🐱</text>
            </view>
            <view class="receipt-brand">
              <text class="receipt-shop">{{ shopName }}</text>
              <text class="receipt-sub" v-if="shopSubtitle">{{ shopSubtitle }}</text>
              <text class="receipt-sub2">— 消费小票 —</text>
            </view>
          </view>

          <!-- 分隔线 -->
          <view class="receipt-dashed"></view>

          <!-- 客户信息 -->
          <view class="receipt-info">
            <view class="receipt-info-row">
              <text class="receipt-info-label">消费时间</text>
              <text class="receipt-info-value">{{ formatDateTime(order.pay_time || order.CreatedAt) }}</text>
            </view>
            <view class="receipt-info-row">
              <text class="receipt-info-label">手机号码</text>
              <text class="receipt-info-value">{{ maskPhone(order.customer?.phone) }}</text>
            </view>
            <view class="receipt-info-row" v-if="hasMemberCard">
              <text class="receipt-info-label">会员余额</text>
              <view class="receipt-member-badge">
                <text class="receipt-member-level">VIP {{ memberCardLevel }}</text>
                <text class="receipt-member-amount">¥{{ balanceBeforePay.toFixed(2) }}</text>
              </view>
            </view>
          </view>

          <!-- 分隔线 -->
          <view class="receipt-dashed"></view>

          <!-- 服务明细表格 -->
          <view class="receipt-table">
          <view class="receipt-table-head">
            <text class="rt-name">项目</text>
            <text class="rt-price">零售价</text>
            <text class="rt-rate">折扣</text>
            <text class="rt-qty">数量</text>
            <text class="rt-amount">小计</text>
          </view>
            <view v-for="(group, groupIndex) in receiptGroups" :key="`receipt-${group.pet_name}-${groupIndex}`" class="receipt-group">
              <view class="receipt-group-head">
                <text class="receipt-group-name">{{ group.pet_name }}</text>
              </view>
              <view
                :class="['receipt-table-row', rowIndex % 2 === 1 ? 'receipt-table-row-alt' : '']"
                v-for="(item, rowIndex) in group.items"
                :key="`receipt-item-${groupIndex}-${item.ID}`"
              >
                <text class="rt-name">{{ item.name }}</text>
                <text class="rt-price">{{ item.unit_price }}</text>
                <text class="rt-rate">
                  <text v-if="getReceiptDiscountTag(item) !== '-'" class="rt-rate-tag">{{ getReceiptDiscountTag(item) }}</text>
                  <text v-else>-</text>
                </text>
                <text class="rt-qty">{{ item.quantity }}</text>
                <text class="rt-amount">{{ calcReceiptAmount(item) }}</text>
              </view>
            </view>
          </view>

          <!-- 分隔线 -->
          <view class="receipt-dashed"></view>

          <!-- 金额汇总 -->
          <view class="receipt-summary">
            <view class="receipt-row" v-if="showReceiptBreakdown && serviceTotalValue > 0">
              <text class="receipt-row-label">服务小计</text>
              <text class="receipt-row-value">¥{{ serviceTotalValue.toFixed(2) }}</text>
            </view>
            <view class="receipt-row" v-if="showReceiptBreakdown && productTotalValue > 0">
              <text class="receipt-row-label">商品小计</text>
              <text class="receipt-row-value">¥{{ productTotalValue.toFixed(2) }}</text>
            </view>
            <view class="receipt-row" v-if="showReceiptBreakdown && showDiscountSummary && productDiscountValue > 0">
              <text class="receipt-row-label">商品优惠</text>
              <view class="receipt-discount-tag">
                <text>省 ¥{{ productDiscountValue.toFixed(2) }}</text>
              </view>
            </view>
            <view class="receipt-row" v-if="showReceiptBreakdown && addonTotalValue > 0">
              <text class="receipt-row-label">附加费</text>
              <text class="receipt-row-value">¥{{ addonTotalValue.toFixed(2) }}</text>
            </view>
            <view class="receipt-row" v-if="showReceiptBreakdown && showBillTotal">
              <text class="receipt-row-label">账单总价</text>
              <text class="receipt-row-value">¥{{ order.total_amount }}</text>
            </view>
            <view class="receipt-row" v-if="showReceiptBreakdown && showDiscountSummary && order.discount_amount > 0">
              <text class="receipt-row-label">优惠金额</text>
              <view class="receipt-discount-tag">
                <text>省 ¥{{ order.discount_amount }}</text>
              </view>
              </view>
            <view class="receipt-pay-block">
              <text class="receipt-pay-label">应付金额</text>
              <text class="receipt-pay-amount">¥{{ order.pay_amount }}</text>
            </view>
            <view class="receipt-info-row" v-if="hasMemberCard && order.pay_method === 'balance'">
              <text class="receipt-info-label">消费后余额</text>
              <view class="receipt-member-badge">
                <text class="receipt-member-level">VIP {{ memberCardLevel }}</text>
                <text class="receipt-member-amount">¥{{ balanceAfterPay.toFixed(2) }}</text>
              </view>
            </view>
          </view>

          <!-- 分隔线 -->
          <view class="receipt-dashed"></view>

          <!-- 底部信息 -->
          <view class="receipt-footer">
            <view class="receipt-footer-row">
              <text class="receipt-footer-label">经手员工</text>
              <text class="receipt-footer-value">{{ order.staff?.name || '-' }}</text>
            </view>
            <view class="receipt-footer-row">
              <text class="receipt-footer-label">支付方式</text>
              <text class="receipt-footer-value">{{ payMethodMap[order.pay_method] || order.pay_method || '待付款' }}</text>
            </view>
          </view>

          <!-- 感谢语 -->
          <view class="receipt-thanks">
            <text class="receipt-thanks-cn">赞美生命 创造健康美好的人宠生活</text>
            <text class="receipt-thanks-en">Praise life, Create a healthy and beautiful pet life.</text>
          </view>

        </view>
      </view>
        <view class="receipt-actions" v-if="!receiptImageUrl">
          <view class="btn-receipt-save" @click="saveReceiptImage">{{ generatingImage ? '生成中...' : '生成图片' }}</view>
          <view class="btn-receipt-close" @click="showReceipt = false">关闭</view>
        </view>
        <!-- 生成后显示图片 -->
        <view v-if="receiptImageUrl" class="receipt-image-wrap">
          <text class="receipt-image-hint">点击「保存图片」或长按图片保存</text>
          <image :src="receiptBlobUrl || receiptImageUrl" mode="widthFix" class="receipt-image" show-menu-by-longpress />
          <view class="receipt-actions">
            <view class="btn-receipt-save" @click="downloadReceiptImage">保存图片</view>
            <view class="btn-receipt-close" @click="closeReceipt">关闭</view>
          </view>
        </view>
      </view>
    </view>

    <!-- Pay modal -->
    <view class="modal-mask" v-if="showPayModal" @click="showPayModal = false">
      <view class="pay-modal" @click.stop>
        <view class="pay-modal-header">
          <view>
            <text class="modal-title">选择收款方式</text>
            <text class="pay-modal-subtitle">确认金额后直接完成收款</text>
          </view>
          <text class="pay-modal-close" @click="showPayModal = false">✕</text>
        </view>

        <view class="pay-amount-panel">
          <text class="pay-amount-label">本单应收</text>
          <text class="modal-amount">¥{{ order.pay_amount }}</text>
        </view>

        <view class="pay-remark-panel">
          <text class="pay-remark-label">备注</text>
          <textarea
            v-model="remarkDraft"
            class="pay-remark-input"
            maxlength="200"
            auto-height
            placeholder="可填写收款备注"
          />
        </view>

        <view class="pay-grid">
          <view class="pay-card qrcode" @click="doPay('qrcode')">
            <view class="pay-card-badge">码</view>
            <text class="pay-card-label">扫码</text>
            <text class="pay-card-sub">聚合码 / 扫码枪</text>
          </view>
          <view class="pay-card wechat" @click="doPay('wechat')">
            <view class="pay-card-badge">微</view>
            <text class="pay-card-label">微信</text>
            <text class="pay-card-sub">微信转账 / 收款</text>
          </view>
          <view class="pay-card meituan" @click="doPay('meituan')">
            <view class="pay-card-badge">团</view>
            <text class="pay-card-label">美团</text>
            <text class="pay-card-sub">平台核销订单</text>
          </view>
          <view :class="['pay-card', 'balance', memberBalance <= 0 ? 'pay-card-disabled' : '']" @click="payWithBalance">
            <view class="pay-card-badge">卡</view>
            <text class="pay-card-label">会员余额</text>
            <text class="pay-card-sub" v-if="memberBalance > 0">可用 ¥{{ memberBalance.toFixed(2) }}</text>
            <text class="pay-card-sub warn" v-else>未开卡 / 无余额</text>
          </view>
          <view class="pay-card other" @click="doPay('other')">
            <view class="pay-card-badge">其</view>
            <text class="pay-card-label">其他</text>
            <text class="pay-card-sub">现金 / 转账 / 线下补录</text>
          </view>
        </view>
      </view>
    </view>
  </view>
  </SideLayout>
</template>

<script setup lang="ts">
import { ref, computed, watch, onUnmounted } from 'vue'
import { onLoad } from '@dcloudio/uni-app'
import SideLayout from '@/components/SideLayout.vue'
import { getOrder, payOrder, cancelOrder, refundOrder, updateOrderRemark, deleteOrder } from '@/api/order'
import { getShop } from '@/api/shop'
import { getCustomerCard } from '@/api/member-card'
import { useAuthStore } from '@/store/auth'
import html2canvas from 'html2canvas'
import { hasStaffRoleAtLeast } from '@/utils/staff-role'

const authStore = useAuthStore()
const isAdmin = computed(() => hasStaffRoleAtLeast(authStore.staffInfo?.role, 'manager'))
const order = ref<any>(null)
const isDeletedView = ref(false)
const showPayModal = ref(false)
const showReceipt = ref(false)
const remarkDraft = ref('')
const savingRemark = ref(false)
const memberBalance = ref(0)
const customerCard = ref<any>(null)
const logoError = ref(false)
const logoSrc = '/uploads/brand/logo.png'
const shopName = ref('猫咪洗护')
const shopSubtitle = ref('')
const lockedScrollY = ref(0)
const pageScrollLockApplied = ref(false)

function setPageScrollLock(locked: boolean) {
  if (typeof window === 'undefined' || typeof document === 'undefined') return
  const body = document.body
  const html = document.documentElement
  if (!body || !html) return

  if (locked) {
    if (pageScrollLockApplied.value) return
    lockedScrollY.value = window.scrollY || window.pageYOffset || 0
    body.style.position = 'fixed'
    body.style.top = `-${lockedScrollY.value}px`
    body.style.left = '0'
    body.style.right = '0'
    body.style.width = '100%'
    body.style.overflow = 'hidden'
    html.style.overflow = 'hidden'
    pageScrollLockApplied.value = true
    return
  }

  if (!pageScrollLockApplied.value) return
  body.style.position = ''
  body.style.top = ''
  body.style.left = ''
  body.style.right = ''
  body.style.width = ''
  body.style.overflow = ''
  html.style.overflow = ''
  window.scrollTo(0, lockedScrollY.value)
  pageScrollLockApplied.value = false
}
const hasMemberCard = computed(() => !!customerCard.value?.ID)
const memberCardLevel = computed(() => {
  return customerCard.value?.template?.name || customerCard.value?.card_name || '会员'
})
const balanceBeforePay = computed(() => {
  const balance = Number(memberBalance.value || 0)
  if (!order.value || order.value.pay_method !== 'balance') return Math.max(balance, 0)
  return Math.max(balance + Number(order.value.pay_amount || 0), 0)
})
const balanceAfterPay = computed(() => {
  if (!order.value || order.value.pay_method !== 'balance') return 0
  return Math.max(memberBalance.value, 0)
})
const showDiscountSummary = computed(() => hasMemberCard.value)
const showBillTotal = computed(() => hasMemberCard.value)

const serviceTotalValue = computed(() => {
  const stored = Number(order.value?.service_total || 0)
  if (stored > 0) return stored
  return getItemSubtotal(1)
})
const productTotalValue = computed(() => {
  const stored = Number(order.value?.product_total || 0)
  if (stored > 0) return stored
  return getItemSubtotal(2)
})
const addonTotalValue = computed(() => {
  const stored = Number(order.value?.addon_total || 0)
  if (stored > 0) return stored
  return getItemSubtotal(3)
})
const showReceiptBreakdown = computed(() => {
  const chargeBuckets = [
    serviceTotalValue.value > 0,
    productTotalValue.value > 0,
    addonTotalValue.value > 0,
  ].filter(Boolean).length
  return chargeBuckets > 1
})
const serviceDiscountValue = computed(() => {
  const stored = Number(order.value?.service_discount_amount || 0)
  if (stored > 0) return stored
  if (serviceTotalValue.value > 0 && productTotalValue.value === 0) {
    return Number(order.value?.discount_amount || 0)
  }
  return 0
})
const productDiscountValue = computed(() => {
  const stored = Number(order.value?.product_discount_amount || 0)
  if (stored > 0) return stored
  if (productTotalValue.value > 0 && serviceTotalValue.value === 0) {
    return Number(order.value?.discount_amount || 0)
  }
  return 0
})
const serviceDiscountRate = computed(() => {
  if (serviceTotalValue.value <= 0) return 1
  return (serviceTotalValue.value - serviceDiscountValue.value) / serviceTotalValue.value
})
const productDiscountRate = computed(() => {
  if (productTotalValue.value <= 0) return 1
  return (productTotalValue.value - productDiscountValue.value) / productTotalValue.value
})

const canEditPrice = computed(() => {
  if (!order.value) return false
  if (isDeletedView.value) return false
  if (order.value.order_kind === 'feeding' || Number(order.value.feeding_plan_id || 0) > 0) return false
  return Number(order.value.pay_status || 0) === 0 && ![2, 3].includes(Number(order.value.status || 0))
})

const primaryPetId = computed(() => {
  const directPetId = Number(order.value?.pet_id || order.value?.pet?.ID || 0)
  if (directPetId > 0) return directPetId
  if (petGroups.value.length === 1) {
    const groupedPetId = Number(petGroups.value[0]?.pet_id || 0)
    if (groupedPetId > 0) return groupedPetId
  }
  return 0
})

const petGroups = computed(() => {
  const groups = order.value?.pet_groups
  if (Array.isArray(groups) && groups.length > 0) {
    return groups
  }
  const items = Array.isArray(order.value?.items) ? order.value.items : []
  if (!items.length) return []

  const grouped: Array<{ pet_name: string; items: any[] }> = []
  const groupMap = new Map<string, { pet_name: string; items: any[] }>()
  for (const item of items) {
    if (item.item_type === 2) {
      const key = '零售商品'
      if (!groupMap.has(key)) {
        const nextGroup = { pet_name: key, items: [] as any[] }
        groupMap.set(key, nextGroup)
        grouped.push(nextGroup)
      }
      groupMap.get(key)!.items.push({ ...item })
      continue
    }
    const [petName, itemName] = splitOrderItemName(item.name)
    const key = petName || order.value?.pet?.name || '未分组'
    if (!groupMap.has(key)) {
      const nextGroup = { pet_name: key, items: [] as any[] }
      groupMap.set(key, nextGroup)
      grouped.push(nextGroup)
    }
    groupMap.get(key)!.items.push({
      ...item,
      name: itemName || item.name,
    })
  }
  return grouped
})

const receiptGroups = computed(() => {
  return petGroups.value.map((group: any) => {
    const isRetailGroup = group.pet_name === '零售商品'
    return {
      ...group,
      items: Array.isArray(group.items)
        ? group.items.map((item: any) => {
            if (!isRetailGroup) return item
            const [, itemName] = splitOrderItemName(item.name)
            return {
              ...item,
              name: itemName || item.name,
            }
          })
        : [],
    }
  })
})

function formatDateTime(val: string | undefined): string {
  if (!val) return '-'
  // 统一处理 ISO 格式和已有空格格式
  const str = val.replace('T', ' ').substring(0, 19)
  const match = str.match(/^(\d{4})-(\d{2})-(\d{2}) (\d{2}):(\d{2}):(\d{2})/)
  if (!match) return val
  return `${match[1]}年${match[2]}月${match[3]}日 ${match[4]}:${match[5]}:${match[6]}`
}

function maskPhone(phone: string | undefined): string {
  if (!phone || phone.length < 7) return phone || '-'
  return phone.substring(0, 3) + '****' + phone.substring(phone.length - 4)
}

function splitOrderItemName(name: string | undefined): [string, string] {
  if (!name) return ['', '']
  const parts = name.split(' · ')
  if (parts.length < 2) return ['', name]
  return [parts[0].trim(), parts.slice(1).join(' · ').trim()]
}

function calcReceiptAmount(item: any): string {
  const quantity = Number(item?.quantity || 0)
  const unitPrice = Number(item?.unit_price || 0)
  const amount = unitPrice * quantity
  if (item?.item_type === 1) {
    return (amount * serviceDiscountRate.value).toFixed(2)
  }
  if (item?.item_type === 2) {
    return (amount * productDiscountRate.value).toFixed(2)
  }
  return amount.toFixed(2)
}

function getItemSubtotal(itemType: number) {
  const items = Array.isArray(order.value?.items) ? order.value.items : []
  return items
    .filter((item: any) => Number(item.item_type) === itemType)
    .reduce((sum: number, item: any) => sum + Number(item.amount || 0), 0)
}

function getReceiptDiscountTag(item: any) {
  if (item?.item_type === 1) {
    if (serviceDiscountRate.value >= 1) return '-'
    return `${(serviceDiscountRate.value * 10).toFixed(1)}折`
  }
  if (item?.item_type === 2) {
    if (productDiscountRate.value >= 1) return '-'
    return `${(productDiscountRate.value * 10).toFixed(1)}折`
  }
  return '-'
}

const receiptImageUrl = ref('')
const receiptBlobUrl = ref('')
const generatingImage = ref(false)

function dataURLtoBlob(dataURL: string): Blob {
  const parts = dataURL.split(',')
  const mime = parts[0].match(/:(.*?);/)![1]
  const raw = atob(parts[1])
  const arr = new Uint8Array(raw.length)
  for (let i = 0; i < raw.length; i++) arr[i] = raw.charCodeAt(i)
  return new Blob([arr], { type: mime })
}

async function saveReceiptImage() {
  const el = document.getElementById('receiptContent')
  if (!el) {
    uni.showToast({ title: '找不到小票内容', icon: 'none' })
    return
  }
  generatingImage.value = true
  try {
    const canvas = await html2canvas(el, {
      backgroundColor: '#fff',
      scale: 2,
      useCORS: true,
      allowTaint: true,
      logging: false,
      scrollX: 0,
      scrollY: 0,
      windowWidth: el.scrollWidth,
      windowHeight: el.scrollHeight,
    })
    const dataUrl = canvas.toDataURL('image/png')
    receiptImageUrl.value = dataUrl
    // Create blob URL for iOS Safari compatibility
    const blob = dataURLtoBlob(dataUrl)
    if (receiptBlobUrl.value) URL.revokeObjectURL(receiptBlobUrl.value)
    receiptBlobUrl.value = URL.createObjectURL(blob)
  } catch (e) {
    console.error('html2canvas error:', e)
    uni.showToast({ title: '生成失败，请截屏保存', icon: 'none' })
  } finally {
    generatingImage.value = false
  }
}

function downloadReceiptImage() {
  if (!receiptBlobUrl.value) return
  const a = document.createElement('a')
  a.href = receiptBlobUrl.value
  a.download = `小票_${order.value?.order_no || 'receipt'}.png`
  document.body.appendChild(a)
  a.click()
  document.body.removeChild(a)
}

function closeReceipt() {
  if (receiptBlobUrl.value) {
    URL.revokeObjectURL(receiptBlobUrl.value)
    receiptBlobUrl.value = ''
  }
  receiptImageUrl.value = ''
  showReceipt.value = false
}
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

onLoad(async (query) => {
  if (query?.id) {
    isDeletedView.value = query.include_deleted === '1'
    const res = await getOrder(parseInt(query.id), isDeletedView.value)
    order.value = res.data
    remarkDraft.value = resolveEditableRemark(order.value)
    // Load shop info for receipt
    try {
      const shopRes = await getShop()
      if (shopRes.data) {
        shopName.value = shopRes.data.name || '猫咪洗护'
        shopSubtitle.value = shopRes.data.address || ''
      }
    } catch {}
    // Load member balance for receipt
    if (order.value?.customer_id) {
      try {
        const cardRes = await getCustomerCard(order.value.customer_id)
        if (cardRes.data) {
          memberBalance.value = cardRes.data.balance
          customerCard.value = cardRes.data
        }
      } catch {}
    }
  }
})

watch(
  () => showReceipt.value || showPayModal.value,
  (visible) => {
    setPageScrollLock(visible)
  }
)

onUnmounted(() => {
  setPageScrollLock(false)
})

async function reload() {
  const res = await getOrder(order.value.ID, isDeletedView.value)
  order.value = res.data
  remarkDraft.value = resolveEditableRemark(order.value)
}

function resolveEditableRemark(target?: Order | null) {
  if (!target) return ''
  const remark = (target.remark || '').trim()
  if (!remark) return ''
  const isBoardingSystemRemark = remark.startsWith('寄养订单 · ')
    && !!target.items?.length
    && target.items.every((item) => [4, 5, 6].includes(item.item_type))
  return isBoardingSystemRemark ? '' : remark
}

async function openPayModal() {
  // Load member balance
  memberBalance.value = 0
  if (order.value?.customer_id) {
    try {
      const cardRes = await getCustomerCard(order.value.customer_id)
      if (cardRes.data && cardRes.data.balance > 0) {
        memberBalance.value = cardRes.data.balance
      }
    } catch {}
  }
  showPayModal.value = true
}

async function doPay(method: string) {
  try {
    await payOrder(order.value.ID, method, undefined, remarkDraft.value.trim())
    showPayModal.value = false
    uni.showToast({ title: '收款成功', icon: 'success' })
    await reload()
  } catch (e: any) {
    uni.showToast({ title: e.message || '收款失败', icon: 'none' })
  }
}

async function payWithBalance() {
  if (memberBalance.value <= 0) {
    uni.showToast({ title: '该客户未开通会员卡', icon: 'none' })
    return
  }
  if (memberBalance.value < order.value.pay_amount) {
    uni.showModal({
      title: '余额不足',
      content: `会员余额¥${memberBalance.value.toFixed(2)}，应付¥${order.value.pay_amount.toFixed(2)}，余额不足。`,
      showCancel: false,
    })
    return
  }
  uni.showModal({
    title: '确认扣款',
    content: `从会员余额中扣除¥${order.value.pay_amount.toFixed(2)}？\n扣后余额：¥${(memberBalance.value - order.value.pay_amount).toFixed(2)}`,
    success: async (res) => {
      if (res.confirm) {
        await doPay('balance')
      }
    }
  })
}

async function doCancel() {
  uni.showModal({
    title: '确认取消', content: '确认取消该订单？',
    success: async (res) => {
      if (res.confirm) {
        await cancelOrder(order.value.ID)
        uni.showToast({ title: '已取消', icon: 'success' })
        await reload()
      }
    }
  })
}

async function doRefund() {
  uni.showModal({
    title: '确认退款', content: '确认退款该订单？',
    success: async (res) => {
      if (res.confirm) {
        await refundOrder(order.value.ID)
        uni.showToast({ title: '已退款', icon: 'success' })
        await reload()
      }
    }
  })
}

function goEditOrder() {
  if (!canEditPrice.value) return
  const isBatchOrder = !!order.value.appointment_id && (!order.value.pet_id || (order.value.pet_groups?.length || 0) > 1)
  const url = isBatchOrder
    ? `/pages/order/batch-create?appointment_id=${order.value.appointment_id}&order_id=${order.value.ID}`
    : `/pages/order/create?order_id=${order.value.ID}`
  uni.navigateTo({ url })
}

function goPetDetail(id?: number) {
  const petId = Number(id || 0)
  if (petId <= 0) return
  uni.navigateTo({ url: `/pages/pet/edit?id=${petId}` })
}

async function doDelete() {
  if (!order.value || ![1, 2, 3].includes(Number(order.value.status))) return
  uni.showModal({
    title: '删除订单',
    content: `确认删除订单 ${order.value.order_no} 吗？\n可在回收站中 2 天内恢复。`,
    success: async (res) => {
      if (!res.confirm) return
      try {
        await deleteOrder(order.value.ID)
        uni.showToast({ title: '已删除', icon: 'success' })
        setTimeout(() => {
          uni.redirectTo({ url: '/pages/order/list' })
        }, 400)
      } catch (e: any) {
        uni.showToast({ title: e?.msg || e?.message || '删除失败', icon: 'none' })
      }
    },
  })
}

async function saveRemark() {
  if (!order.value || savingRemark.value) return
  savingRemark.value = true
  const nextRemark = remarkDraft.value.trim()
  try {
    await updateOrderRemark(order.value.ID, nextRemark)
    order.value.remark = nextRemark
    uni.showToast({ title: '备注已保存', icon: 'success' })
  } catch (e: any) {
    uni.showToast({ title: e.message || '保存失败', icon: 'none' })
  } finally {
    savingRemark.value = false
  }
}
</script>

<style scoped>
.page { padding: 24rpx; }
.status-bar { padding: 24rpx; border-radius: 16rpx; margin-bottom: 16rpx; text-align: center; }
.status-text { font-size: 32rpx; font-weight: bold; }
.deleted-banner {
  margin-bottom: 16rpx;
  padding: 18rpx 22rpx;
  border-radius: 16rpx;
  background: #FFF7ED;
  color: #C2410C;
  font-size: 24rpx;
  text-align: center;
}
.s0 { background: #FEF3C7; color: #92400E; }
.s1 { background: #D1FAE5; color: #059669; }
.s2 { background: #F3F4F6; color: #6B7280; }
.s3 { background: #FEE2E2; color: #DC2626; }
.card { background: #fff; border-radius: 16rpx; padding: 24rpx; margin-bottom: 16rpx; }
.card-title { font-size: 28rpx; font-weight: 600; color: #1F2937; display: block; margin-bottom: 16rpx; }
.row { display: flex; justify-content: space-between; padding: 12rpx 0; border-bottom: 1rpx solid #F3F4F6; font-size: 28rpx; }
.row:last-child { border-bottom: none; }
.label { color: #6B7280; }
.row-value {
  max-width: 70%;
  text-align: right;
}
.pet-group + .pet-group {
  margin-top: 18rpx;
}
.pet-group-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12rpx;
  margin-bottom: 8rpx;
  padding: 14rpx 16rpx;
  border-radius: 16rpx;
  background: #F8FAFC;
}
.pet-group-name {
  font-size: 26rpx;
  font-weight: 700;
  color: #1E293B;
}
.group-retail { color: #7C3AED; text-decoration: none; }
.pet-link {
  color: #4F46E5;
  text-decoration: underline;
}
.pet-group-count {
  font-size: 22rpx;
  color: #64748B;
}
.item-row { display: flex; justify-content: space-between; padding: 12rpx 0; border-bottom: 1rpx solid #F3F4F6; font-size: 26rpx; }
.item-name { flex: 1; }
.item-qty { width: 80rpx; text-align: center; color: #6B7280; }
.item-amount { width: 120rpx; text-align: right; }
.total-section { margin-top: 16rpx; padding-top: 16rpx; border-top: 2rpx solid #E5E7EB; }
.total-row { display: flex; justify-content: space-between; font-size: 26rpx; padding: 8rpx 0; color: #6B7280; }
.total-row.final { font-size: 30rpx; font-weight: bold; color: #1F2937; }
.discount-text { color: #059669; }
.pay-amount { color: #4F46E5; }
.remark-block {
  margin-top: 18rpx;
  padding-top: 18rpx;
  border-top: 1rpx solid #E5E7EB;
}
.remark-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12rpx;
  margin-bottom: 12rpx;
}
.remark-title {
  font-size: 26rpx;
  font-weight: 700;
  color: #334155;
}
.remark-save {
  min-width: 104rpx;
  height: 52rpx;
  padding: 0 18rpx;
  border-radius: 999rpx;
  background: #EEF2FF;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  font-size: 22rpx;
  font-weight: 700;
  color: #4F46E5;
}
.remark-input {
  width: 100%;
  min-height: 108rpx;
  padding: 18rpx 20rpx;
  border-radius: 18rpx;
  background: #F8FAFC;
  border: 2rpx solid #E2E8F0;
  box-sizing: border-box;
  font-size: 26rpx;
  color: #111827;
  line-height: 1.6;
  box-shadow: 0 8rpx 20rpx rgba(15, 23, 42, 0.04);
}
.actions {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 16rpx;
  margin-top: 20rpx;
}
.btn {
  margin: 0;
  min-height: 94rpx;
  padding: 0 24rpx;
  border-radius: 20rpx;
  font-size: 27rpx;
  line-height: 1.2;
  font-weight: 700;
  display: flex;
  align-items: center;
  justify-content: center;
  box-sizing: border-box;
  border: 2rpx solid transparent;
  box-shadow: 0 8rpx 20rpx rgba(15, 23, 42, 0.06);
}
.edit { background: #EEF2FF; color: #4338CA; border-color: #C7D2FE; }
.pay { background: linear-gradient(135deg, #4F46E5, #6366F1); color: #fff; box-shadow: 0 14rpx 28rpx rgba(79, 70, 229, 0.24); }
.cancel { background: #fff; color: #64748B; border-color: #CBD5E1; }
.refund { background: #FFF1F2; color: #DC2626; border-color: #FECDD3; }
.delete { background: #FFF1F2; color: #DC2626; border-color: #FCA5A5; }

/* Pay modal */
.modal-mask {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0,0,0,0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 24rpx 24rpx calc(24rpx + env(safe-area-inset-bottom));
  box-sizing: border-box;
  z-index: 5000;
  overscroll-behavior: contain;
}
.pay-modal {
  width: 86%;
  max-width: 680rpx;
  background: linear-gradient(180deg, #FFFFFF, #FBFCFF);
  border-radius: 28rpx;
  padding: 28rpx;
  box-shadow: 0 24rpx 60rpx rgba(15, 23, 42, 0.24);
}
.pay-modal-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16rpx;
}
.modal-title {
  font-size: 32rpx;
  font-weight: 800;
  color: #111827;
  display: block;
}
.pay-modal-subtitle {
  display: block;
  margin-top: 8rpx;
  font-size: 22rpx;
  color: #94A3B8;
}
.pay-modal-close {
  width: 52rpx;
  height: 52rpx;
  line-height: 52rpx;
  text-align: center;
  border-radius: 50%;
  background: #F3F4F6;
  color: #6B7280;
  font-size: 24rpx;
  flex-shrink: 0;
}
.pay-amount-panel {
  margin: 24rpx 0 26rpx;
  padding: 22rpx 24rpx;
  border-radius: 22rpx;
  background: linear-gradient(135deg, #EEF2FF, #F8FAFF);
  border: 1rpx solid #C7D2FE;
}
.pay-amount-label {
  display: block;
  font-size: 22rpx;
  color: #6366F1;
  letter-spacing: 1rpx;
}
.modal-amount {
  display: block;
  margin-top: 10rpx;
  font-size: 56rpx;
  line-height: 1;
  font-weight: 900;
  color: #4338CA;
}
.pay-remark-panel {
  margin: 0 0 22rpx;
}
.pay-remark-label {
  display: block;
  margin-bottom: 10rpx;
  font-size: 24rpx;
  font-weight: 700;
  color: #475569;
}
.pay-remark-input {
  width: 100%;
  min-height: 96rpx;
  padding: 18rpx 20rpx;
  border-radius: 18rpx;
  background: #F8FAFC;
  border: 2rpx solid #E2E8F0;
  box-sizing: border-box;
  font-size: 26rpx;
  color: #111827;
  line-height: 1.6;
  box-shadow: 0 8rpx 20rpx rgba(15, 23, 42, 0.04);
}

.pay-grid { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 18rpx; }
.pay-card {
  min-height: 172rpx;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 22rpx 20rpx;
  border-radius: 22rpx;
  background: #FFFFFF;
  border: 2rpx solid #E5E7EB;
  box-sizing: border-box;
  box-shadow: 0 8rpx 20rpx rgba(15, 23, 42, 0.05);
  gap: 12rpx;
  text-align: center;
}
.pay-card:active { transform: scale(0.98); }
.pay-card.qrcode { border-color: #BFDBFE; background: linear-gradient(180deg, #F8FBFF, #FFFFFF); }
.pay-card.wechat { border-color: #BBF7D0; background: linear-gradient(180deg, #F7FFF8, #FFFFFF); }
.pay-card.meituan { border-color: #FED7AA; background: linear-gradient(180deg, #FFF7ED, #FFFFFF); }
.pay-card.balance { border-color: #C7D2FE; background: linear-gradient(180deg, #F8FAFF, #FFFFFF); }
.pay-card.other { border-color: #E5E7EB; background: linear-gradient(180deg, #FCFCFD, #FFFFFF); }
.pay-card-disabled { opacity: 0.55; }
.pay-card-badge {
  width: 56rpx;
  height: 56rpx;
  border-radius: 16rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 24rpx;
  font-weight: 800;
  color: #111827;
  background: rgba(255,255,255,0.85);
  border: 1rpx solid rgba(148, 163, 184, 0.22);
}
.pay-card-label { font-size: 30rpx; color: #111827; font-weight: 700; }
.pay-card-sub { font-size: 22rpx; color: #059669; line-height: 1.45; }
.pay-card-sub.warn { color: #DC2626; }

/* Receipt button */
.btn.receipt { background: #F8FAFC; color: #334155; border-color: #CBD5E1; }

/* ===== Receipt Modal ===== */
.receipt-outer {
  width: 94%;
  max-width: 760rpx;
  max-height: 90vh;
  display: flex;
  flex-direction: column;
}
.receipt-wrap {
  flex: 1;
  overflow-y: auto;
  min-height: 0;
  overscroll-behavior: contain;
  -webkit-overflow-scrolling: touch;
  touch-action: pan-y;
}

/* 小票卡片主体 — 奶油色背景，暖调阴影 */
.receipt {
  background: #FDFBF7;
  border-radius: 24rpx;
  box-shadow: 0 8rpx 40rpx rgba(139, 109, 56, 0.13);
  overflow: hidden;
  font-family: -apple-system, 'PingFang SC', 'Helvetica Neue', sans-serif;
}

/* ---- 头部 — 参照微信公众号排版 ---- */
.receipt-header {
  background: #FAF5E8;
  padding: 20rpx 24rpx 12rpx;
  display: flex;
  align-items: center;
  gap: 20rpx;
}
.receipt-logo {
  width: 96rpx; height: 96rpx; min-width: 96rpx;
  border-radius: 50%; overflow: hidden;
  border: 2rpx solid #E8D9B5;
  background: #fff;
}
.receipt-logo-emoji { font-size: 48rpx; line-height: 96rpx; text-align: center; display: block; }
.receipt-logo-img { width: 96rpx; height: 96rpx; border-radius: 50%; }
.receipt-brand {
  flex: 1;
  padding-left: 12rpx;
}
.receipt-shop { font-size: 30rpx; font-weight: 600; color: #3D3D3D; letter-spacing: 1rpx; display: block; }
.receipt-sub { font-size: 22rpx; color: #B8A88A; font-weight: 300; display: block; margin-top: 9rpx; }
.receipt-sub2 { display: none; }

/* ---- 虚线分隔 — 极浅金色 ---- */
.receipt-dashed {
  margin: 0 24rpx;
  border-top: 1rpx dashed #DED0AA;
}

/* ---- 客户信息区 ---- */
.receipt-info {
  padding: 24rpx 36rpx;
  display: flex;
  flex-direction: column;
  gap: 16rpx;
}
.receipt-info-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.receipt-info-label {
  font-size: 24rpx;
  color: #B8A88A;
  font-weight: 300;
}
.receipt-info-value {
  font-size: 24rpx;
  color: #4A3F2F;
  font-weight: 400;
}
/* 当前余额 — 金色标签 */
.receipt-balance-badge {
  background: #FBF5E6;
  border: 1rpx solid #E8D5A0;
  color: #A07830;
  font-size: 22rpx;
  font-weight: 500;
  padding: 4rpx 18rpx;
  border-radius: 999rpx;
}
.receipt-member-badge {
  display: inline-flex;
  align-items: center;
  gap: 12rpx;
  padding: 8rpx 18rpx;
  border-radius: 999rpx;
  background: #FBF5E6;
  border: 1rpx solid #E8D5A0;
}
.receipt-member-level {
  font-size: 20rpx;
  font-weight: 700;
  color: #9A6A21;
  letter-spacing: 0.5rpx;
}
.receipt-member-amount {
  font-size: 22rpx;
  font-weight: 600;
  color: #A07830;
}

/* ---- 明细表格 ---- */
.receipt-table {
  padding: 0 36rpx;
  margin: 24rpx 0;
}
.receipt-group + .receipt-group {
  margin-top: 14rpx;
}
.receipt-group-head {
  padding: 14rpx 12rpx 8rpx;
}
.receipt-group-name {
  font-size: 24rpx;
  font-weight: 700;
  color: #7C6242;
}
.receipt-table-head {
  display: flex;
  font-size: 20rpx;
  color: #B8A88A;
  font-weight: 500;
  padding: 10rpx 12rpx;
  background: #FAF5E8;
  border-radius: 8rpx;
  margin-bottom: 4rpx;
  letter-spacing: 1rpx;
}
.receipt-table-row {
  display: flex;
  font-size: 22rpx;
  color: #4A3F2F;
  padding: 12rpx 12rpx;
  border-radius: 8rpx;
  align-items: flex-start;
  line-height: 1.35;
}
/* 交替背景色 — 极淡奶油 */
.receipt-table-row-alt {
  background: #FAF5E8;
}
.rt-name {
  flex: 4.6;
  min-width: 0;
  font-size: 21rpx;
  word-break: break-word;
}
.rt-price { flex: 1.6; text-align: right; color: #8A7A62; font-size: 20rpx; }
.rt-rate { flex: 1.2; text-align: center; font-size: 20rpx; }
/* 折扣小标签 — 柔和绿色 */
.rt-rate-tag {
  background: #F0FAF5;
  color: #3A8A62;
  font-size: 20rpx;
  font-weight: 500;
  padding: 2rpx 8rpx;
  border-radius: 999rpx;
  border: 1rpx solid #B5DEC8;
}
.rt-qty { flex: 0.9; text-align: center; color: #8A7A62; font-size: 20rpx; }
.rt-amount { flex: 1.5; text-align: right; font-weight: 600; color: #4A3F2F; font-size: 20rpx; }

/* ---- 金额汇总 ---- */
.receipt-summary {
  padding: 24rpx 36rpx;
  display: flex;
  flex-direction: column;
  gap: 16rpx;
}
.receipt-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.receipt-row-label {
  font-size: 24rpx;
  color: #B8A88A;
  font-weight: 300;
}
.receipt-row-value {
  font-size: 24rpx;
  color: #4A3F2F;
}
/* 优惠金额 — 金色标签 */
.receipt-discount-tag {
  background: #FBF5E6;
  border: 1rpx solid #E8D5A0;
  color: #A07830;
  font-size: 22rpx;
  font-weight: 600;
  padding: 4rpx 18rpx;
  border-radius: 999rpx;
}
/* 应付金额 — 简约金色底线装饰，无渐变块 */
.receipt-pay-block {
  background: transparent;
  border-radius: 0;
  padding: 20rpx 0 16rpx;
  display: flex;
  justify-content: space-between;
  align-items: flex-end;
  margin-top: 8rpx;
  border-top: 1rpx solid #DED0AA;
  border-bottom: none;
}
.receipt-pay-label {
  font-size: 26rpx;
  color: #C4A35A;
  font-weight: 400;
  letter-spacing: 1rpx;
}
.receipt-pay-amount {
  font-size: 52rpx;
  font-weight: 300;
  color: #C4A35A;
  letter-spacing: -1rpx;
}
/* 消费后余额 — 金色标签 */
/* ---- 底部信息区 ---- */
.receipt-footer {
  padding: 14rpx 32rpx;
  background: #FAF5E8;
  display: flex;
  flex-direction: column;
  gap: 6rpx;
}
.receipt-footer-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.receipt-footer-label {
  font-size: 22rpx;
  color: #C4B08A;
  font-weight: 300;
}
.receipt-footer-value {
  font-size: 22rpx;
  color: #6B5C42;
}
/* ---- 感谢语 ---- */
.receipt-thanks {
  padding: 14rpx 28rpx 16rpx;
  text-align: center;
}
.receipt-thanks-cn { font-size: 20rpx; color: #4A3F2F; letter-spacing: 1rpx; display: block; }
.receipt-thanks-en { font-size: 17rpx; color: #C4A35A; font-weight: 300; display: block; margin-top: 2rpx; }

/* ---- 操作按钮区 ---- */
.receipt-actions {
  display: flex;
  gap: 16rpx;
  padding: 20rpx 0 calc(8rpx + env(safe-area-inset-bottom));
  flex-shrink: 0;
}
.btn-receipt-save,
.btn-receipt-close {
  flex: 1;
  min-height: 92rpx;
  border-radius: 20rpx;
  font-size: 27rpx;
  font-weight: 700;
  text-align: center;
  display: flex;
  align-items: center;
  justify-content: center;
  box-sizing: border-box;
}
.btn-receipt-save { background: linear-gradient(135deg, #6366F1, #4F46E5); color: #fff; box-shadow: 0 14rpx 28rpx rgba(79, 70, 229, 0.2); }
.btn-receipt-close { background: #F8FAFC; color: #64748B; border: 2rpx solid #CBD5E1; }

.btn:active,
.btn-receipt-save:active,
.btn-receipt-close:active { transform: scale(0.98); }

@media (max-width: 768px) {
  .actions {
    grid-template-columns: 1fr;
  }
}

/* Receipt image preview */
.receipt-image-wrap { margin-top: 20rpx; text-align: center; }
.receipt-image-hint { font-size: 26rpx; color: #C4A35A; font-weight: 500; display: block; margin-bottom: 16rpx; }
.receipt-image { width: 100%; border-radius: 16rpx; box-shadow: 0 4rpx 20rpx rgba(0,0,0,0.1); }
</style>
