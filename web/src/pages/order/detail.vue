<template>
  <SideLayout>
  <view class="page" v-if="order">
    <view :class="['status-bar', `s${order.status}`]">
      <text class="status-text">{{ statusMap[order.status] }}</text>
    </view>

    <view class="card">
      <view class="row"><text class="label">订单号</text><text>{{ order.order_no }}</text></view>
      <view class="row"><text class="label">客户</text><text>{{ order.customer?.nickname || '-' }}</text></view>
      <view class="row"><text class="label">洗护师</text><text>{{ order.staff?.name || '-' }}</text></view>
      <view class="row" v-if="order.pay_method"><text class="label">支付方式</text><text>{{ payMethodMap[order.pay_method] || order.pay_method }}</text></view>
      <view class="row" v-if="order.pay_time"><text class="label">支付时间</text><text>{{ order.pay_time }}</text></view>
    </view>

    <view class="card">
      <text class="card-title">明细</text>
      <view class="item-row" v-for="item in (order.items || [])" :key="item.ID">
        <text class="item-name">{{ item.name }}</text>
        <text class="item-qty">x{{ item.quantity }}</text>
        <text class="item-amount">¥{{ item.amount }}</text>
      </view>
      <view class="total-section">
        <view class="total-row"><text>总计</text><text>¥{{ order.total_amount }}</text></view>
        <view class="total-row" v-if="order.discount_amount"><text>优惠</text><text class="discount-text">-¥{{ order.discount_amount }}</text></view>
        <view class="total-row final"><text>应付</text><text class="pay-amount">¥{{ order.pay_amount }}</text></view>
      </view>
    </view>

    <view class="actions">
      <button v-if="order.status === 0" class="btn pay" @click="openPayModal">收款</button>
      <button v-if="order.status === 0 && isAdmin" class="btn cancel" @click="doCancel">取消订单</button>
      <button v-if="order.status === 1 && isAdmin" class="btn refund" @click="doRefund">退款</button>
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
              <text class="receipt-shop">树街の猫</text>
              <text class="receipt-sub">赛级皮毛调理体验店</text>
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
              <text class="receipt-info-label">客户姓名</text>
              <text class="receipt-info-value">{{ order.customer?.nickname || '-' }}</text>
            </view>
            <view class="receipt-info-row">
              <text class="receipt-info-label">手机号码</text>
              <text class="receipt-info-value">{{ maskPhone(order.customer?.phone) }}</text>
            </view>
            <view class="receipt-info-row" v-if="customerCard">
              <text class="receipt-info-label">当前余额</text>
              <view class="receipt-balance-badge">
                <text>💳 {{ customerCard.card_name }} · ¥{{ memberBalance.toFixed(2) }}</text>
              </view>
            </view>
          </view>

          <!-- 分隔线 -->
          <view class="receipt-dashed"></view>

          <!-- 服务明细表格 -->
          <view class="receipt-table">
            <view class="receipt-table-head">
              <text class="rt-name">服务项目</text>
              <text class="rt-price">零售价</text>
              <text class="rt-rate">折扣</text>
              <text class="rt-qty">数量</text>
              <text class="rt-amount">小计</text>
            </view>
            <view :class="['receipt-table-row', index % 2 === 1 ? 'receipt-table-row-alt' : '']"
                  v-for="(item, index) in (order.items || [])" :key="item.ID">
              <text class="rt-name">{{ item.name }}</text>
              <text class="rt-price">{{ item.unit_price }}</text>
              <text class="rt-rate">
                <text v-if="discountRateDisplay !== '-'" class="rt-rate-tag">{{ discountRateDisplay }}</text>
                <text v-else>-</text>
              </text>
              <text class="rt-qty">{{ item.quantity }}</text>
              <text class="rt-amount">{{ (item.unit_price * item.quantity * orderDiscountRate).toFixed(2) }}</text>
            </view>
          </view>

          <!-- 分隔线 -->
          <view class="receipt-dashed"></view>

          <!-- 金额汇总 -->
          <view class="receipt-summary">
            <view class="receipt-row">
              <text class="receipt-row-label">账单总价</text>
              <text class="receipt-row-value">¥{{ order.total_amount }}</text>
            </view>
            <view class="receipt-row" v-if="order.discount_amount > 0">
              <text class="receipt-row-label">优惠金额</text>
              <view class="receipt-discount-tag">
                <text>省 ¥{{ order.discount_amount }}</text>
              </view>
            </view>
            <view class="receipt-pay-block">
              <text class="receipt-pay-label">应付金额</text>
              <text class="receipt-pay-amount">¥{{ order.pay_amount }}</text>
            </view>
            <view class="receipt-row" v-if="order.pay_method === 'balance'">
              <text class="receipt-row-label">消费后余额</text>
              <view class="receipt-after-balance-badge">
                <text>¥{{ balanceAfterPay.toFixed(2) }}</text>
              </view>
            </view>
          </view>

          <!-- 分隔线 -->
          <view class="receipt-dashed"></view>

          <!-- 底部信息 -->
          <view class="receipt-footer">
            <view class="receipt-footer-row">
              <text class="receipt-footer-label">服务洗护师</text>
              <text class="receipt-footer-value">{{ order.staff?.name || '-' }}</text>
            </view>
            <view class="receipt-footer-row">
              <text class="receipt-footer-label">支付方式</text>
              <text class="receipt-footer-value">{{ payMethodMap[order.pay_method] || order.pay_method || '待付款' }}</text>
            </view>
            <view class="receipt-footer-row">
              <text class="receipt-footer-label">订单编号</text>
              <text class="receipt-footer-value receipt-order-no">{{ order.order_no }}</text>
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
      <view class="modal" @click.stop>
        <text class="modal-title">选择收款方式</text>
        <text class="modal-amount">应收：¥{{ order.pay_amount }}</text>

        <view class="pay-grid">
          <view class="pay-card" @click="doPay('cash')">
            <text class="pay-card-icon">💵</text>
            <text class="pay-card-label">现金</text>
          </view>
          <view class="pay-card" @click="doPay('wechat')">
            <text class="pay-card-icon">📱</text>
            <text class="pay-card-label">扫码</text>
          </view>
          <view class="pay-card" @click="doPay('meituan')">
            <text class="pay-card-icon">🟠</text>
            <text class="pay-card-label">美团</text>
          </view>
          <view :class="['pay-card', memberBalance <= 0 ? 'pay-card-disabled' : '']" @click="payWithBalance">
            <text class="pay-card-icon">💳</text>
            <text class="pay-card-label">会员余额</text>
            <text class="pay-card-sub" v-if="memberBalance > 0">余额¥{{ memberBalance.toFixed(2) }}</text>
            <text class="pay-card-sub warn" v-else>未开卡</text>
          </view>
        </view>
      </view>
    </view>
  </view>
  </SideLayout>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { onLoad } from '@dcloudio/uni-app'
import SideLayout from '@/components/SideLayout.vue'
import { getOrder, payOrder, cancelOrder, refundOrder } from '@/api/order'
import { getCustomerCard } from '@/api/member-card'
import { useAuthStore } from '@/store/auth'
import html2canvas from 'html2canvas'

const authStore = useAuthStore()
const isAdmin = computed(() => authStore.staffInfo?.role === 'admin')
const order = ref<any>(null)
const showPayModal = ref(false)
const showReceipt = ref(false)
const memberBalance = ref(0)
const customerCard = ref<any>(null)
const logoError = ref(false)
const logoSrc = '/uploads/brand/logo.png'
const balanceAfterPay = computed(() => {
  if (!order.value || order.value.pay_method !== 'balance') return 0
  return Math.max(memberBalance.value, 0)
})

const orderDiscountRate = computed(() => {
  if (!order.value || !order.value.total_amount || order.value.total_amount === 0) return 1
  return order.value.pay_amount / order.value.total_amount
})

const discountRateDisplay = computed(() => {
  const r = orderDiscountRate.value
  if (r >= 1) return '-'
  return (r * 10).toFixed(1) + '折'
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
const statusMap: Record<number, string> = { 0: '待付款', 1: '已完成', 2: '已取消', 3: '已退款' }
const payMethodMap: Record<string, string> = { wechat: '扫码', alipay: '扫码', cash: '现金', meituan: '美团', balance: '会员余额' }

onLoad(async (query) => {
  if (query?.id) {
    const res = await getOrder(parseInt(query.id))
    order.value = res.data
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

async function reload() {
  const res = await getOrder(order.value.ID)
  order.value = res.data
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
    await payOrder(order.value.ID, method)
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
</script>

<style scoped>
.page { padding: 24rpx; }
.status-bar { padding: 24rpx; border-radius: 16rpx; margin-bottom: 16rpx; text-align: center; }
.status-text { font-size: 32rpx; font-weight: bold; }
.s0 { background: #FEF3C7; color: #92400E; }
.s1 { background: #D1FAE5; color: #059669; }
.s2 { background: #F3F4F6; color: #6B7280; }
.s3 { background: #FEE2E2; color: #DC2626; }
.card { background: #fff; border-radius: 16rpx; padding: 24rpx; margin-bottom: 16rpx; }
.card-title { font-size: 28rpx; font-weight: 600; color: #1F2937; display: block; margin-bottom: 16rpx; }
.row { display: flex; justify-content: space-between; padding: 12rpx 0; border-bottom: 1rpx solid #F3F4F6; font-size: 28rpx; }
.row:last-child { border-bottom: none; }
.label { color: #6B7280; }
.item-row { display: flex; justify-content: space-between; padding: 12rpx 0; border-bottom: 1rpx solid #F3F4F6; font-size: 26rpx; }
.item-name { flex: 1; }
.item-qty { width: 80rpx; text-align: center; color: #6B7280; }
.item-amount { width: 120rpx; text-align: right; }
.total-section { margin-top: 16rpx; padding-top: 16rpx; border-top: 2rpx solid #E5E7EB; }
.total-row { display: flex; justify-content: space-between; font-size: 26rpx; padding: 8rpx 0; color: #6B7280; }
.total-row.final { font-size: 30rpx; font-weight: bold; color: #1F2937; }
.discount-text { color: #059669; }
.pay-amount { color: #4F46E5; }
.actions { display: flex; flex-direction: column; gap: 16rpx; margin-top: 16rpx; }
.btn { border-radius: 12rpx; font-size: 30rpx; }
.pay { background: #4F46E5; color: #fff; }
.cancel { background: #fff; color: #6B7280; border: 1rpx solid #D1D5DB; }
.refund { background: #fff; color: #DC2626; border: 1rpx solid #DC2626; }

/* Pay modal */
.modal-mask { position: fixed; top: 0; left: 0; right: 0; bottom: 0; background: rgba(0,0,0,0.5); display: flex; align-items: center; justify-content: center; z-index: 999; }
.modal { background: #fff; border-radius: 20rpx; padding: 40rpx; width: 80%; }
.modal-title { font-size: 32rpx; font-weight: bold; display: block; text-align: center; }
.modal-amount { font-size: 48rpx; font-weight: 800; color: #4F46E5; display: block; text-align: center; margin: 16rpx 0 28rpx; }

.pay-grid { display: flex; flex-wrap: wrap; gap: 20rpx; }
.pay-card {
  width: calc(50% - 10rpx); display: flex; flex-direction: column; align-items: center;
  padding: 28rpx 16rpx; border-radius: 16rpx; background: #F9FAFB;
  border: 2rpx solid #E5E7EB; gap: 6rpx;
}
.pay-card:active { border-color: #4F46E5; background: #EEF2FF; }
.pay-card-disabled { opacity: 0.5; }
.pay-card-icon { font-size: 40rpx; }
.pay-card-label { font-size: 28rpx; color: #374151; font-weight: 600; }
.pay-card-sub { font-size: 22rpx; color: #059669; }
.pay-card-sub.warn { color: #DC2626; }

/* Receipt button */
.btn.receipt { background: #fff; color: #374151; border: 1rpx solid #D1D5DB; }

/* ===== Receipt Modal ===== */
.receipt-outer {
  width: 88%;
  max-width: 640rpx;
  max-height: 90vh;
  display: flex;
  flex-direction: column;
}
.receipt-wrap {
  flex: 1;
  overflow-y: auto;
  min-height: 0;
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
  padding: 20rpx 24rpx;
  display: flex;
  align-items: center;
  gap: 20rpx;
}
.receipt-logo {
  width: 80rpx; height: 80rpx; min-width: 80rpx;
  border-radius: 50%; overflow: hidden;
  border: 2rpx solid #E8D9B5;
  background: #fff;
}
.receipt-logo-emoji { font-size: 40rpx; line-height: 80rpx; text-align: center; display: block; }
.receipt-logo-img { width: 80rpx; height: 80rpx; border-radius: 50%; }
.receipt-brand { flex: 1; }
.receipt-shop { font-size: 30rpx; font-weight: 600; color: #3D3D3D; letter-spacing: 1rpx; display: block; }
.receipt-sub { font-size: 22rpx; color: #B8A88A; font-weight: 300; display: block; margin-top: 2rpx; }
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

/* ---- 明细表格 ---- */
.receipt-table {
  padding: 0 36rpx;
  margin: 24rpx 0;
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
  font-size: 24rpx;
  color: #4A3F2F;
  padding: 12rpx 12rpx;
  border-radius: 8rpx;
  align-items: center;
}
/* 交替背景色 — 极淡奶油 */
.receipt-table-row-alt {
  background: #FAF5E8;
}
.rt-name { flex: 3; }
.rt-price { flex: 2; text-align: right; color: #8A7A62; }
.rt-rate { flex: 1.5; text-align: center; }
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
.rt-qty { flex: 1; text-align: center; color: #8A7A62; }
.rt-amount { flex: 2; text-align: right; font-weight: 600; color: #4A3F2F; }

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
.receipt-after-balance-badge {
  background: #FBF5E6;
  border: 1rpx solid #E8D5A0;
  color: #A07830;
  font-size: 22rpx;
  font-weight: 500;
  padding: 4rpx 18rpx;
  border-radius: 999rpx;
}

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
.receipt-order-no {
  font-size: 20rpx;
  color: #B8A88A;
  letter-spacing: 1rpx;
}

/* ---- 感谢语 ---- */
.receipt-thanks {
  padding: 14rpx 28rpx 16rpx;
  text-align: center;
}
.receipt-thanks-cn { font-size: 20rpx; color: #4A3F2F; letter-spacing: 1rpx; display: block; }
.receipt-thanks-en { font-size: 17rpx; color: #C4A35A; font-weight: 300; display: block; margin-top: 2rpx; }

/* ---- 操作按钮区 ---- */
.receipt-actions { display: flex; gap: 16rpx; padding: 20rpx 0 0; flex-shrink: 0; }
.btn-receipt-save { flex: 1; background: linear-gradient(135deg, #6366F1, #4F46E5); color: #fff; border-radius: 12rpx; font-size: 28rpx; text-align: center; padding: 18rpx 0; }
.btn-receipt-close { flex: 1; background: #F3F4F6; color: #6B7280; border-radius: 12rpx; font-size: 28rpx; text-align: center; padding: 18rpx 0; }

/* Receipt image preview */
.receipt-image-wrap { margin-top: 20rpx; text-align: center; }
.receipt-image-hint { font-size: 26rpx; color: #C4A35A; font-weight: 500; display: block; margin-bottom: 16rpx; }
.receipt-image { width: 100%; border-radius: 16rpx; box-shadow: 0 4rpx 20rpx rgba(0,0,0,0.1); }
</style>
