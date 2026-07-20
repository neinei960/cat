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
      <view class="row">
        <text class="label">客户</text>
        <text
          :class="['row-value', order.customer_id ? 'pet-link' : '']"
          @click="goCustomerDetail(order.customer_id)"
        >{{ order.customer?.nickname || '-' }}</text>
      </view>
      <view class="row pet-row" v-if="headerPets.length || order.order_kind !== 'product' || canEditCustomerPet">
        <text class="label">猫咪</text>
        <view class="pet-name-list">
          <template v-if="headerPets.length">
            <template v-for="(pet, index) in headerPets" :key="`${pet.pet_id || pet.pet_name}-${index}`">
              <text
                class="pet-link pet-name-link"
                @click="goPetDetail(pet.pet_id, pet.pet_name)"
              >{{ pet.pet_name }}</text>
              <text v-if="index < headerPets.length - 1" class="pet-name-separator">、</text>
            </template>
          </template>
          <text v-else class="row-value">-</text>
        </view>
      </view>
      <view class="row" v-if="showOrderKindRow">
        <text class="label">订单类型</text>
        <text class="row-value">{{ orderKindLabel }}</text>
      </view>
      <view class="row" v-if="order.appointment_id">
        <text class="label">到店状态</text>
        <text :class="['row-value', appointmentIsLateValue ? 'late-flag' : '']">{{ appointmentIsLateValue ? '迟到' : '正常' }}</text>
      </view>
      <view class="row"><text class="label">经手员工</text><text>{{ order.staff?.name || '-' }}</text></view>
      <view class="row" v-if="order.pay_method"><text class="label">支付方式</text><text>{{ displayPayMethod }}</text></view>
      <view class="row" v-if="order.pay_method === 'mixed_balance'"><text class="label">余额扣款</text><text>¥{{ Number(order.member_balance_used || 0).toFixed(2) }}</text></view>
      <view class="row" v-if="order.pay_method === 'mixed_balance'"><text class="label">补差金额</text><text>¥{{ Number(order.cash_pay_amount || 0).toFixed(2) }}</text></view>
      <view class="row" v-if="order.pay_time"><text class="label">支付时间</text><text>{{ formatDateTime(order.pay_time) }}</text></view>
    </view>

    <view class="card">
      <text class="card-title">明细</text>
      <view v-for="(group, groupIndex) in petGroups" :key="`${group.pet_name}-${groupIndex}`" class="pet-group">
        <view v-if="shouldShowPetGroupHead(group)" class="pet-group-head">
          <text
            :class="[
              'pet-group-name',
              group.pet_name !== '零售商品' ? 'pet-link' : '',
              group.pet_name === '零售商品' ? 'group-retail' : ''
            ]"
            @click="group.pet_name !== '零售商品' ? goPetDetail(group.pet_id, group.pet_name) : undefined"
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
        <view class="total-row" v-if="showDetailBreakdown && serviceTotalValue > 0"><text>服务小计</text><text>¥{{ serviceTotalValue.toFixed(2) }}</text></view>
        <view class="total-row" v-if="showDetailBreakdown && serviceDiscountValue > 0"><text>服务优惠</text><text class="discount-text">-¥{{ serviceDiscountValue.toFixed(2) }}</text></view>
        <view class="total-row" v-if="showDetailBreakdown && productTotalValue > 0"><text>商品小计</text><text>¥{{ productTotalValue.toFixed(2) }}</text></view>
        <view class="total-row" v-if="showDetailBreakdown && productDiscountValue > 0"><text>商品优惠</text><text class="discount-text">-¥{{ productDiscountValue.toFixed(2) }}</text></view>
        <view class="total-row" v-if="showDetailBreakdown && addonTotalValue > 0"><text>附加费</text><text>¥{{ addonTotalValue.toFixed(2) }}</text></view>
        <view class="total-row" v-if="showDetailBreakdown && order.discount_amount"><text>优惠</text><text class="discount-text">-¥{{ order.discount_amount }}</text></view>
        <view class="total-row" v-if="appointmentDepositDeductionValue > 0"><text>{{ appointmentDepositLabel }}</text><text class="discount-text">-¥{{ appointmentDepositDeductionValue.toFixed(2) }}</text></view>
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
            placeholder="备注收款说明、客户要求或补充信息"
          />
        </view>
      </view>
    </view>

    <view class="actions">
      <button v-if="canEditCustomerPet" class="btn link" @click="openCustomerPetModal">修改客户/猫咪</button>
      <button v-if="canEditPrice" class="btn edit" @click="goEditOrder">修改订单</button>
      <button v-if="order.status === 0 && !isDeletedView" class="btn pay" @click="openPayModal">收款</button>
      <button v-if="order.status === 0 && isAdmin && !isDeletedView" class="btn cancel" @click="doCancel">取消订单</button>
      <button v-if="order.status === 1 && isAdmin && !isDeletedView" class="btn refund" @click="doRefund">退款</button>
      <button v-if="(order.status === 1 || order.status === 2 || order.status === 3) && isAdmin && !isDeletedView" class="btn delete" @click="doDelete">删除订单</button>
      <button v-if="canGenerateCareReport" class="btn report" @click="showCareReport = true">生成报告</button>
      <button class="btn receipt" @click="showReceipt = true">生成小票</button>
    </view>

    <view class="modal-mask" v-if="showCustomerPetModal" @click="closeCustomerPetModal">
      <view class="link-modal" @click.stop>
        <view class="modal-head">
          <view>
            <text class="modal-title">修改客户/猫咪</text>
            <text class="modal-subtitle">只修改订单归属，不改金额和明细</text>
          </view>
          <text class="modal-close" @click="closeCustomerPetModal">×</text>
        </view>

        <view class="link-section">
          <view class="link-section-head">
            <text class="link-section-title">客户</text>
            <text v-if="customerPetSelectedCustomer" class="link-clear" @click="clearCustomerPetCustomer">清空</text>
          </view>
          <view v-if="customerPetSelectedCustomer" class="selected-link-card">
            <text class="selected-link-name">{{ customerPetSelectedCustomer.nickname || '未命名客户' }}</text>
            <text class="selected-link-meta">{{ customerPetSelectedCustomer.phone || '未留手机号' }}</text>
          </view>
          <view class="link-search">
            <input
              v-model="customerPetCustomerKeyword"
              class="link-search-input"
              confirm-type="search"
              @input="searchCustomerPetCustomers"
              @confirm="searchCustomerPetCustomers"
            />
            <text v-if="!customerPetCustomerKeyword" class="link-search-placeholder">搜索客户昵称或手机号</text>
          </view>
          <view v-if="customerPetCustomerOptions.length" class="link-options">
            <view
              v-for="customer in customerPetCustomerOptions"
              :key="customer.ID"
              class="link-option"
              @click="selectCustomerPetCustomer(customer)"
            >
              <text class="link-option-name">{{ customer.nickname || '未命名客户' }}</text>
              <text class="link-option-meta">{{ customer.phone || '未留手机号' }}</text>
            </view>
          </view>
        </view>

        <view class="link-section">
          <view class="link-section-head">
            <text class="link-section-title">猫咪</text>
            <text v-if="customerPetSelectedPet" class="link-clear" @click="customerPetSelectedPet = null">清空</text>
          </view>
          <view v-if="!customerPetSelectedCustomer" class="link-empty">先选择客户；纯商品单也可以只绑定客户。</view>
          <view v-else-if="customerPetPets.length === 0" class="link-empty">该客户暂无猫咪档案，可以只保存客户。</view>
          <view v-else class="pet-choice-list">
            <view
              v-for="pet in customerPetPets"
              :key="pet.ID"
              :class="['pet-choice', customerPetSelectedPet?.ID === pet.ID ? 'active' : '']"
              @click="customerPetSelectedPet = pet"
            >
              <text class="pet-choice-name">{{ pet.name }}</text>
              <text class="pet-choice-meta">{{ pet.breed || '未知品种' }}</text>
            </view>
          </view>
        </view>

        <view class="link-actions">
          <button class="link-btn ghost" @click="closeCustomerPetModal">取消</button>
          <button class="link-btn primary" :disabled="savingCustomerPet" @click="saveCustomerPet">{{ savingCustomerPet ? '保存中...' : '保存' }}</button>
        </view>
      </view>
    </view>

    <!-- 小票弹窗 -->
    <view class="modal-mask" v-if="showReceipt" @click="closeReceipt">
      <view class="receipt-outer" @click.stop>
      <view class="receipt-wrap" v-if="!receiptImageUrl">
        <view class="receipt" id="receiptContent" :class="{ 'receipt-lite-render': receiptLiteRender }">
          <view class="receipt-card receipt-brand-card">
            <view class="receipt-brand-top">
              <view class="receipt-logo">
                <image
                  v-if="!logoError && !receiptLiteRender"
                  class="receipt-logo-img"
                  :src="logoSrc"
                  mode="aspectFill"
                  @error="logoError = true"
                />
                <text v-else class="receipt-logo-emoji">猫</text>
              </view>
              <view class="receipt-brand-main">
                <text class="receipt-shop">{{ shopName }}</text>
                <view class="receipt-sub">
                  <text
                    v-for="(line, index) in receiptSubtitleLines"
                    :key="`receipt-sub-${index}`"
                    class="receipt-sub-line"
                  >{{ line }}</text>
                </view>
              </view>
              <view v-if="!hasMemberCard" class="receipt-brand-tag muted">{{ statusMap[Number(order.status)] || '账单' }}</view>
            </view>

            <view class="receipt-meta-list">
              <view class="receipt-meta-row">
                <text class="receipt-meta-label">消费时间</text>
                <text class="receipt-meta-value">{{ formatReceiptDateTime(order.pay_time || order.CreatedAt) }}</text>
              </view>
              <view class="receipt-meta-row">
                <text class="receipt-meta-label">手机号码</text>
                <text class="receipt-meta-value">{{ maskPhone(order.customer?.phone) }}</text>
              </view>
              <view v-if="hasMemberCard" class="receipt-meta-row">
                <text class="receipt-meta-label">会员余额</text>
                <view class="receipt-member-badge">
                  <text class="receipt-member-level">VIP {{ memberCardLevel }}</text>
                  <text class="receipt-member-amount">¥{{ receiptBalanceBeforePay.toFixed(2) }}</text>
                </view>
              </view>
            </view>
          </view>

          <view
            v-for="(group, groupIndex) in receiptGroups"
            :key="`receipt-${group.pet_name}-${groupIndex}`"
            class="receipt-card receipt-group-card"
          >
            <text class="receipt-group-title">{{ group.pet_name }}</text>

            <view
              v-for="item in group.items"
              :key="`receipt-item-${groupIndex}-${item.ID}`"
              class="receipt-item"
            >
              <view class="receipt-item-main">
                <text class="receipt-item-name">{{ item.name }}</text>
                <view class="receipt-item-meta">
                  <text>售价 ¥{{ Number(item.unit_price || 0).toFixed(2) }}</text>
                  <template v-if="hasReceiptDiscount(item)">
                    <text class="receipt-item-dot">·</text>
                    <text>{{ getReceiptDiscountTag(item) }}</text>
                  </template>
                  <text class="receipt-item-dot">·</text>
                  <text>×{{ item.quantity }}</text>
                </view>
              </view>
              <text class="receipt-item-price">¥{{ calcReceiptAmount(item) }}</text>
            </view>
          </view>

          <view class="receipt-card receipt-summary-card">
            <view v-if="showReceiptBreakdown && serviceTotalValue > 0" class="receipt-summary-row">
              <text>服务小计</text>
              <text>¥{{ serviceTotalValue.toFixed(2) }}</text>
            </view>
            <view v-if="showReceiptBreakdown && productTotalValue > 0" class="receipt-summary-row">
              <text>商品小计</text>
              <text>¥{{ productTotalValue.toFixed(2) }}</text>
            </view>
            <view v-if="showReceiptBreakdown && addonTotalValue > 0" class="receipt-summary-row">
              <text>附加费</text>
              <text>¥{{ addonTotalValue.toFixed(2) }}</text>
            </view>
            <view v-if="showReceiptBreakdown && productDiscountValue > 0" class="receipt-summary-row saving">
              <text>商品优惠</text>
              <text>-¥{{ productDiscountValue.toFixed(2) }}</text>
            </view>
            <view v-if="receiptTotalSaving > 0" class="receipt-summary-row saving">
              <text>总优惠</text>
              <text>-¥{{ receiptTotalSaving.toFixed(2) }}</text>
            </view>
            <view v-if="appointmentDepositDeductionValue > 0" class="receipt-summary-row saving">
              <text>{{ appointmentDepositLabel }}</text>
              <text>-¥{{ appointmentDepositDeductionValue.toFixed(2) }}</text>
            </view>

            <view class="receipt-summary-divider"></view>

            <view class="receipt-total-row">
              <view class="receipt-total-copy">
                <text class="receipt-total-label">实付总计</text>
                <text class="receipt-total-tip">{{ receiptTotalSaving > 0 ? `本次已节省 ¥${receiptTotalSaving.toFixed(2)}` : '期待下次见 👋🏻' }}</text>
              </view>
              <text class="receipt-total-price">¥{{ Number(order.pay_amount || 0).toFixed(2) }}</text>
            </view>

            <view v-if="hasMemberCard" class="receipt-balance-row">
              <text class="receipt-meta-label">消费后余额</text>
              <text class="receipt-meta-value receipt-meta-value-strong">¥{{ receiptBalanceAfterPay.toFixed(2) }}</text>
            </view>
            <view v-if="receiptBalanceUsedAmount > 0" class="receipt-balance-row">
              <text class="receipt-meta-label">余额抵扣</text>
              <text class="receipt-meta-value receipt-meta-value-strong">¥{{ receiptBalanceUsedAmount.toFixed(2) }}</text>
            </view>
            <view v-if="receiptCashPayAmount > 0" class="receipt-balance-row">
              <text class="receipt-meta-label">用户需补</text>
              <text class="receipt-meta-value receipt-meta-value-strong">¥{{ receiptCashPayAmount.toFixed(2) }}</text>
            </view>
          </view>

          <view class="receipt-card receipt-footer-card">
            <view class="receipt-footer-row">
              <text class="receipt-footer-label">经手员工</text>
              <text class="receipt-footer-value">{{ order.staff?.name || '-' }}</text>
            </view>
            <view class="receipt-footer-row">
              <text class="receipt-footer-label">支付方式</text>
              <text class="receipt-footer-value">{{ displayPayMethod }}</text>
            </view>
            <view v-if="order.pay_method === 'mixed_balance'" class="receipt-footer-row">
              <text class="receipt-footer-label">余额扣款</text>
              <text class="receipt-footer-value">¥{{ Number(order.member_balance_used || 0).toFixed(2) }}</text>
            </view>
            <view v-if="order.pay_method === 'mixed_balance'" class="receipt-footer-row">
              <text class="receipt-footer-label">补差金额</text>
              <text class="receipt-footer-value">¥{{ Number(order.cash_pay_amount || 0).toFixed(2) }}</text>
            </view>
          </view>

          <view class="receipt-thanks">
            <text class="receipt-thanks-cn">赞美生命 创造健康美好的人宠生活</text>
            <text class="receipt-thanks-en">Praise life, Create a healthy and beautiful pet life.</text>
          </view>
        </view>
      </view>
        <view class="receipt-actions" v-if="!receiptImageUrl">
          <view class="btn-receipt-save" @click="saveReceiptImage">{{ generatingImage ? '生成中...' : '生成图片' }}</view>
          <view class="btn-receipt-close" @click="closeReceipt">关闭</view>
        </view>
        <!-- 生成后显示图片 -->
        <view v-if="receiptImageUrl" class="receipt-image-wrap">
          <text class="receipt-image-hint">{{ receiptImageHint }}</text>
          <img
            v-if="showNativeReceiptImage"
            :src="receiptPreviewSrc"
            class="receipt-image receipt-image-native"
            alt="小票图片"
            @click.stop
            @touchstart.stop
          />
          <image
            v-else
            :src="receiptPreviewSrc"
            mode="widthFix"
            class="receipt-image"
            show-menu-by-longpress
          />
          <view class="receipt-actions">
            <view class="btn-receipt-save" @click="downloadReceiptImage">保存图片</view>
            <view class="btn-receipt-close" @click="closeReceipt">关闭</view>
          </view>
        </view>
      </view>
    </view>

    <OrderCareReportModal
      v-if="order"
      :visible="showCareReport"
      :order="order"
      @close="showCareReport = false"
    />

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

        <view v-if="canChooseBalancePayment && customerCard" class="pay-choice-tip">
          <text class="pay-choice-title">会员余额</text>
          <text class="pay-choice-desc">该客户会员余额 ¥{{ memberBalance.toFixed(2) }}，店长可选择扣余额，也可以选择其它收款方式。</text>
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
          <view :class="['pay-card', 'qrcode', paying ? 'pay-card-disabled' : '']" @click="confirmPayMethod('qrcode')">
            <view class="pay-card-badge">码</view>
            <text class="pay-card-label">扫码</text>
            <text class="pay-card-sub">聚合码 / 扫码枪</text>
          </view>
          <view :class="['pay-card', 'wechat', paying ? 'pay-card-disabled' : '']" @click="confirmPayMethod('wechat')">
            <view class="pay-card-badge">微</view>
            <text class="pay-card-label">微信</text>
            <text class="pay-card-sub">微信转账 / 收款</text>
          </view>
          <view :class="['pay-card', 'meituan', paying ? 'pay-card-disabled' : '']" @click="confirmPayMethod('meituan')">
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
import { ref, computed, watch, onUnmounted, nextTick } from 'vue'
import { onLoad } from '@dcloudio/uni-app'
import SideLayout from '@/components/SideLayout.vue'
import OrderCareReportModal from '@/components/order/OrderCareReportModal.vue'
import { getOrder, payOrder, cancelOrder, refundOrder, updateOrderRemark, updateOrderCustomerPet, deleteOrder } from '@/api/order'
import { getCustomerList, getCustomerPets } from '@/api/customer'
import { getPetList } from '@/api/pet'
import { getShop } from '@/api/shop'
import { getCustomerCard } from '@/api/member-card'
import { useAuthStore } from '@/store/auth'
import html2canvas from 'html2canvas'
import { hasStaffRoleAtLeast } from '@/utils/staff-role'
import { getReceiptGroupName, getReceiptItemDisplayName, splitOrderItemName } from '@/utils/order-item-display'
import { canGenerateOrderCareReport } from '@/utils/order-care-report'
import { getReceiptCanvasScale } from '@/utils/receipt-image'
import { buildReceiptFileName, dataUrlToBlob, isAppleSafariBrowser, saveImageByUrl } from '@/utils/web-image-save'

const authStore = useAuthStore()
const isAdmin = computed(() => hasStaffRoleAtLeast(authStore.staffInfo?.role, 'manager'))
const canManageOpenedOrder = computed(() => hasStaffRoleAtLeast(authStore.staffInfo?.role, 'manager'))
const currentStaffRole = computed(() => {
  if (authStore.staffInfo?.role) return authStore.staffInfo.role
  try {
    const raw = uni.getStorageSync('staffInfo')
    if (!raw) return ''
    return JSON.parse(raw)?.role || ''
  } catch {
    return ''
  }
})
const isStoreOwner = computed(() => currentStaffRole.value === 'admin')
const order = ref<any>(null)
const isDeletedView = ref(false)
const showPayModal = ref(false)
const showReceipt = ref(false)
const showCareReport = ref(false)
const showCustomerPetModal = ref(false)
const remarkDraft = ref('')
const savingRemark = ref(false)
const savingCustomerPet = ref(false)
const paying = ref(false)
const memberBalance = ref(0)
const customerCard = ref<any>(null)
const logoError = ref(false)
const logoSrc = '/uploads/brand/logo.png'
const shopName = ref('猫咪洗护')
const shopSubtitle = ref('')
const lockedScrollY = ref(0)
const pageScrollLockApplied = ref(false)
const resolvingPetName = ref('')
const customerPetCustomerKeyword = ref('')
const customerPetCustomerOptions = ref<any[]>([])
const customerPetSelectedCustomer = ref<any | null>(null)
const customerPetSelectedPet = ref<any | null>(null)
const customerPetPets = ref<any[]>([])
let customerPetSearchTimer: ReturnType<typeof setTimeout> | null = null
const SERVICE_LIKE_ITEM_TYPES = [1, 4, 5, 6]

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
function hasSnapshotAmount(value: unknown) {
  return value !== null && value !== undefined && Number.isFinite(Number(value))
}

const hasMemberBalanceSnapshot = computed(() => {
  return hasSnapshotAmount(order.value?.member_balance_before) || hasSnapshotAmount(order.value?.member_balance_after)
})
const hasMemberCard = computed(() => !!customerCard.value?.ID || hasMemberBalanceSnapshot.value)
const memberCardLevel = computed(() => {
  return customerCard.value?.template?.name || customerCard.value?.card_name || '会员'
})
const receiptBalanceBeforePay = computed(() => {
  if (hasSnapshotAmount(order.value?.member_balance_before)) {
    return Math.max(Number(order.value?.member_balance_before), 0)
  }
  const balance = Number(memberBalance.value || 0)
  if (!order.value || order.value.pay_method !== 'balance') return Math.max(balance, 0)
  return Math.max(balance + Number(order.value.pay_amount || 0), 0)
})
const isMixedBalancePreview = computed(() => {
  return !!customerCard.value
    && Number(order.value?.pay_status || 0) === 0
    && memberBalance.value > 0
    && memberBalance.value < Number(order.value?.pay_amount || 0)
})
const receiptBalanceAfterPay = computed(() => {
  if (hasSnapshotAmount(order.value?.member_balance_after)) {
    return Math.max(Number(order.value?.member_balance_after), 0)
  }
  if (isMixedBalancePreview.value) {
    return 0
  }
  return Math.max(Number(memberBalance.value || 0), 0)
})
const receiptBalanceUsedAmount = computed(() => {
  const stored = Number(order.value?.member_balance_used || 0)
  if (stored > 0) return Math.max(stored, 0)
  if (!isMixedBalancePreview.value) return 0
  return Math.min(Math.max(Number(memberBalance.value || 0), 0), Math.max(Number(order.value?.pay_amount || 0), 0))
})
const receiptCashPayAmount = computed(() => {
  const stored = Number(order.value?.cash_pay_amount || 0)
  if (stored > 0) return Math.max(stored, 0)
  if (!isMixedBalancePreview.value) return 0
  return Math.max(Number(order.value?.pay_amount || 0) - receiptBalanceUsedAmount.value, 0)
})
const receiptSubtitleLines = computed(() => {
  const fallback = '专注猫咪科学与健康的可持续人宠美护生活'
  const raw = String(shopSubtitle.value || fallback)
  const normalized = raw
    .replace(/[\u200B-\u200F\u202A-\u202E\u2060-\u206F\uFEFF]/g, '')
    .replace(/[\u00A0\u1680\u2000-\u200A\u202F\u205F\u3000]+/g, ' ')
    .replace(/\s+/g, ' ')
    .trim()
  if (!normalized) return [fallback]
  return normalized.split(' ').filter(Boolean)
})
const serviceTotalValue = computed(() => {
  const stored = Number(order.value?.service_total || 0)
  if (stored > 0) return stored
  return getItemSubtotalByTypes(SERVICE_LIKE_ITEM_TYPES)
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
const appointmentIsLateValue = computed(() => !!order.value?.appointment_is_late)
const appointmentDepositLabel = computed(() => {
  if (order.value?.order_kind === 'boarding') return '定金抵扣'
  return appointmentIsLateValue.value ? '⚠️ 预约金抵扣' : '预约金抵扣'
})
const appointmentDepositDeductionValue = computed(() => Number(order.value?.appointment_deposit_deduction_amount || 0))
const chargeBucketCount = computed(() => {
  return [
    serviceTotalValue.value > 0,
    productTotalValue.value > 0,
    addonTotalValue.value > 0,
  ].filter(Boolean).length
})
const showReceiptBreakdown = computed(() => chargeBucketCount.value > 1)
const showDetailBreakdown = computed(() => chargeBucketCount.value > 1)
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
const hasAnyDiscount = computed(() => Number(order.value?.discount_amount || 0) > 0 || serviceDiscountValue.value > 0 || productDiscountValue.value > 0)
const showDiscountSummary = computed(() => hasAnyDiscount.value)
const showBillTotal = computed(() => hasAnyDiscount.value)
const receiptTotalSaving = computed(() => {
  const stored = Number(order.value?.discount_amount || 0)
  if (stored > 0) return stored
  return Number((serviceDiscountValue.value + productDiscountValue.value).toFixed(2))
})
const serviceDiscountRate = computed(() => {
  if (serviceTotalValue.value <= 0) return 1
  return (serviceTotalValue.value - serviceDiscountValue.value) / serviceTotalValue.value
})
const productDiscountRate = computed(() => {
  if (productTotalValue.value <= 0) return 1
  return (productTotalValue.value - productDiscountValue.value) / productTotalValue.value
})
function getMemberCardDiscountRate(value: unknown) {
  const rate = Number(value || 1)
  if (rate > 0 && rate < 1) return rate
  return 1
}
const receiptServiceDiscountRate = computed(() => {
  const customerRate = getMemberCardDiscountRate(order.value?.customer?.discount_rate)
  if (customerRate < 1) return customerRate
  return getMemberCardDiscountRate(customerCard.value?.discount_rate)
})
const receiptProductDiscountRate = computed(() => getMemberCardDiscountRate(customerCard.value?.product_discount_rate))

const canEditPrice = computed(() => {
  if (!order.value) return false
  if (isDeletedView.value) return false
  if (
    order.value.order_kind === 'feeding' ||
    order.value.order_kind === 'boarding' ||
    Number(order.value.feeding_plan_id || 0) > 0
  ) return false
  const status = Number(order.value.status || 0)
  const payStatus = Number(order.value.pay_status || 0)
  if ([2, 3].includes(status)) return false
  if (payStatus === 0) return true
  return payStatus === 1 && status === 1 && canManageOpenedOrder.value
})
const canEditCustomerPet = computed(() => {
  if (!order.value || isDeletedView.value) return false
  const status = Number(order.value.status || 0)
  return status === 0 || status === 1
})
const canGenerateCareReport = computed(() => canGenerateOrderCareReport(order.value))
const canChooseBalancePayment = computed(() => isStoreOwner.value)

const orderKindLabel = computed(() => {
  switch (order.value?.order_kind) {
    case 'product':
      return '商品零售'
    case 'mixed':
      return '服务 + 商品'
    case 'feeding':
      return '上门喂养'
    case 'boarding':
      return '寄养'
    default:
      return '服务订单'
  }
})

const showOrderKindRow = computed(() => {
  return !!order.value?.order_kind && order.value.order_kind !== 'service'
})

const headerPets = computed(() => {
  const result: Array<{ pet_id?: number; pet_name: string }> = []
  const appendPet = (petName?: string, petId?: number) => {
    const name = String(petName || '').trim()
    if (!name || name === '零售商品') return
    const normalizedPetId = Number(petId || 0)
    const existing = result.find((item) => item.pet_name === name)
    if (existing) {
      if (!existing.pet_id && normalizedPetId > 0) existing.pet_id = normalizedPetId
      return
    }
    result.push({
      pet_id: normalizedPetId > 0 ? normalizedPetId : undefined,
      pet_name: name,
    })
  }

  appendPet(order.value?.pet?.name, Number(order.value?.pet_id || order.value?.pet?.ID || 0))

  if (order.value?.order_kind !== 'boarding') {
    for (const group of petGroups.value) {
      appendPet(group?.pet_name, Number(group?.pet_id || 0))
    }
  }

  if (result.length === 0) {
    const summary = String(order.value?.pet_summary || '').trim()
    if (summary) {
      summary
        .split(/[、,，/]/)
        .map((item) => item.trim())
        .filter(Boolean)
        .forEach((name) => appendPet(name))
    }
  }

  return result
})

const retailNamePrefixes = computed(() => {
  const prefixes = new Set<string>()
  const appendPrefix = (value?: string) => {
    const nextValue = String(value || '').trim()
    if (!nextValue || nextValue === '零售商品' || nextValue === '未分组') return
    prefixes.add(nextValue)
  }

  appendPrefix(order.value?.pet?.name)

  const rawGroups = Array.isArray(order.value?.pet_groups) ? order.value.pet_groups : []
  for (const group of rawGroups) {
    appendPrefix(group?.pet_name)
  }

  const items = Array.isArray(order.value?.items) ? order.value.items : []
  for (const item of items) {
    if (Number(item?.item_type) === 2) continue
    const [petName] = splitOrderItemName(item?.name)
    appendPrefix(petName)
  }

  const summary = String(order.value?.pet_summary || '').trim()
  if (summary) {
    summary
      .split(/[、,，/]/)
      .map((item) => item.trim())
      .filter(Boolean)
      .forEach((name) => appendPrefix(name))
  }

  return Array.from(prefixes)
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
    return groups.map((group: any) => ({
      ...group,
      pet_name: getReceiptGroupName(group.pet_name, order.value?.order_kind),
      items: Array.isArray(group.items)
        ? group.items.map((item: any) => ({
            ...item,
            name: getReceiptItemDisplayName(item.name, group.pet_name === '零售商品', retailNamePrefixes.value, order.value?.order_kind),
          }))
        : [],
    }))
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
      groupMap.get(key)!.items.push({
        ...item,
        name: getReceiptItemDisplayName(item.name, true, retailNamePrefixes.value),
      })
      continue
    }
    const [petName, itemName] = splitOrderItemName(item.name)
    const key = petName || order.value?.pet?.name || '未分组'
    if (!groupMap.has(key)) {
      const nextGroup = { pet_name: getReceiptGroupName(key, order.value?.order_kind), items: [] as any[] }
      groupMap.set(key, nextGroup)
      grouped.push(nextGroup)
    }
    groupMap.get(key)!.items.push({
      ...item,
      name: getReceiptItemDisplayName(item.name, false, retailNamePrefixes.value, order.value?.order_kind) || itemName || item.name,
    })
  }
  return grouped
})

const receiptGroups = computed(() => {
  return petGroups.value.map((group: any) => {
    const isRetailGroup = group.pet_name === '零售商品'
    return {
      ...group,
      pet_name: getReceiptGroupName(group.pet_name, order.value?.order_kind),
      items: Array.isArray(group.items)
        ? group.items.map((item: any) => {
            return {
              ...item,
              name: getReceiptItemDisplayName(item.name, isRetailGroup, retailNamePrefixes.value, order.value?.order_kind),
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

function formatReceiptDateTime(val: string | undefined): string {
  if (!val) return '-'
  const str = val.replace('T', ' ').substring(0, 16)
  const match = str.match(/^(\d{4})-(\d{2})-(\d{2}) (\d{2}):(\d{2})/)
  if (!match) return val
  return `${match[1]}.${match[2]}.${match[3]} ${match[4]}:${match[5]}`
}

function maskPhone(phone: string | undefined): string {
  if (!phone || phone.length < 7) return phone || '-'
  return phone.substring(0, 3) + '****' + phone.substring(phone.length - 4)
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
  return getItemSubtotalByTypes([itemType])
}

function getItemSubtotalByTypes(itemTypes: number[]) {
  const items = Array.isArray(order.value?.items) ? order.value.items : []
  return items
    .filter((item: any) => itemTypes.includes(Number(item.item_type)))
    .reduce((sum: number, item: any) => sum + Number(item.amount || 0), 0)
}

function getReceiptDiscountTag(item: any) {
  if (item?.item_type === 1) {
    if (receiptServiceDiscountRate.value >= 1) return '-'
    return `${(receiptServiceDiscountRate.value * 10).toFixed(1)}折`
  }
  if (item?.item_type === 2) {
    if (receiptProductDiscountRate.value >= 1) return '-'
    return `${(receiptProductDiscountRate.value * 10).toFixed(1)}折`
  }
  return '-'
}

function hasReceiptDiscount(item: any) {
  return getReceiptDiscountTag(item) !== '-'
}

const receiptImageUrl = ref('')
const receiptBlobUrl = ref('')
const generatingImage = ref(false)
const receiptRenderToken = ref(0)
const receiptLiteRender = ref(false)
const isAppleSafari = computed(() => isAppleSafariBrowser())
const showNativeReceiptImage = computed(() => true)
const receiptPreviewSrc = computed(() => {
  if (!receiptImageUrl.value) return ''
  return isAppleSafari.value ? receiptImageUrl.value : receiptBlobUrl.value || receiptImageUrl.value
})
const receiptImageHint = computed(() => {
  if (isAppleSafari.value) {
    return '长按图片保存到相册；如未出现菜单，点击「保存图片」后在新页面长按'
  }
  return '点击「保存图片」或长按图片保存'
})

function isConstrainedReceiptCanvas() {
  if (typeof window === 'undefined' || typeof navigator === 'undefined') return false
  const userAgent = navigator.userAgent || ''
  return /iP(hone|od|ad)/i.test(userAgent) || window.innerWidth <= 480
}

async function renderReceiptCanvas(el: HTMLElement, scale: number) {
  await nextTick()
  return html2canvas(el, {
    backgroundColor: '#F6F2EA',
    scale,
    useCORS: true,
    allowTaint: false,
    logging: false,
    scrollX: 0,
    scrollY: 0,
    windowWidth: Math.ceil(el.scrollWidth),
    windowHeight: Math.ceil(el.scrollHeight),
    foreignObjectRendering: false,
    imageTimeout: 8000,
    removeContainer: true,
  })
}

async function renderReceiptDataUrl(el: HTMLElement, scale: number) {
  const canvas = await renderReceiptCanvas(el, scale)
  return canvas.toDataURL('image/png')
}

async function renderReceiptDataUrlWithFallback(el: HTMLElement, scale: number) {
  try {
    return await renderReceiptDataUrl(el, scale)
  } catch (primaryError) {
    console.error('html2canvas primary render error:', primaryError)
  }

  if (scale > 1) {
    try {
      return await renderReceiptDataUrl(el, 1)
    } catch (scaleError) {
      console.error('html2canvas scale-1 render error:', scaleError)
    }
  }

  receiptLiteRender.value = true
  try {
    return await renderReceiptDataUrl(el, 1)
  } finally {
    receiptLiteRender.value = false
    await nextTick()
  }
}

async function saveReceiptImage() {
  const el = document.getElementById('receiptContent')
  if (!el) {
    uni.showToast({ title: '找不到小票内容', icon: 'none' })
    return
  }
  const renderToken = receiptRenderToken.value + 1
  receiptRenderToken.value = renderToken
  generatingImage.value = true
  try {
    const scale = getReceiptCanvasScale(el.scrollWidth, el.scrollHeight, isConstrainedReceiptCanvas())
    const dataUrl = await renderReceiptDataUrlWithFallback(el, scale)
    if (!showReceipt.value || receiptRenderToken.value !== renderToken) return
    // Keep a blob URL for non-Safari download fallback.
    const blob = dataUrlToBlob(dataUrl)
    if (!showReceipt.value || receiptRenderToken.value !== renderToken) return
    receiptImageUrl.value = dataUrl
    if (receiptBlobUrl.value) URL.revokeObjectURL(receiptBlobUrl.value)
    receiptBlobUrl.value = URL.createObjectURL(blob)
  } catch (e) {
    if (!showReceipt.value || receiptRenderToken.value !== renderToken) return
    console.error('html2canvas error:', e)
    uni.showToast({ title: '生成失败，请截屏保存', icon: 'none' })
  } finally {
    if (receiptRenderToken.value === renderToken) {
      generatingImage.value = false
    }
  }
}

async function downloadReceiptImage() {
  if (!receiptImageUrl.value) return
  const result = await saveImageByUrl(
    receiptImageUrl.value,
    buildReceiptFileName(order.value?.order_no || 'receipt'),
    {
      title: '小票图片',
      blobUrl: receiptBlobUrl.value || undefined,
    }
  )
  if (result === 'preview') {
    uni.showToast({ title: '新页面已打开，请长按图片保存', icon: 'none' })
  }
}

function closeReceipt() {
  receiptRenderToken.value += 1
  generatingImage.value = false
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
  mixed_balance: '会员余额 + 补差',
  other: '其他',
  alipay: '扫码',
  cash: '其他',
  card: '会员余额',
}
const payMethodConfirmMap: Record<string, string> = {
  qrcode: '扫码',
  wechat: '微信',
  meituan: '美团',
  balance: '会员余额',
  other: '其他',
}
const mixedBalanceCashPayMethods = [
  { label: '扫码补差', method: 'qrcode' },
  { label: '微信补差', method: 'wechat' },
  { label: '美团补差', method: 'meituan' },
  { label: '其他补差', method: 'other' },
]
const displayPayMethod = computed(() => {
  if (!order.value?.pay_method) return '待付款'
  if (order.value.pay_method === 'mixed_balance') {
    const cashLabel = payMethodMap[order.value.cash_pay_method] || order.value.cash_pay_method || '补差'
    return `会员余额 + ${cashLabel}`
  }
  return payMethodMap[order.value.pay_method] || order.value.pay_method
})

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
  () => showReceipt.value || showPayModal.value || showCareReport.value || showCustomerPetModal.value,
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

async function openCustomerPetModal() {
  if (!order.value) return
  customerPetCustomerKeyword.value = ''
  customerPetCustomerOptions.value = []
  customerPetSelectedCustomer.value = order.value.customer || null
  customerPetSelectedPet.value = order.value.pet || null
  customerPetPets.value = []
  showCustomerPetModal.value = true
  const customerId = Number(order.value.customer_id || order.value.customer?.ID || 0)
  if (customerId > 0) {
    await loadCustomerPetPets(customerId)
  }
}

function closeCustomerPetModal() {
  showCustomerPetModal.value = false
}

async function searchCustomerPetCustomers() {
  if (customerPetSearchTimer) clearTimeout(customerPetSearchTimer)
  customerPetSearchTimer = setTimeout(async () => {
    const keyword = customerPetCustomerKeyword.value.trim()
    if (!keyword) {
      customerPetCustomerOptions.value = []
      return
    }
    try {
      const res = await getCustomerList({ page: 1, page_size: 20, keyword } as any)
      customerPetCustomerOptions.value = res.data?.list || []
    } catch {
      customerPetCustomerOptions.value = []
    }
  }, 250)
}

async function selectCustomerPetCustomer(customer: any) {
  customerPetSelectedCustomer.value = customer
  customerPetSelectedPet.value = null
  customerPetCustomerKeyword.value = ''
  customerPetCustomerOptions.value = []
  await loadCustomerPetPets(Number(customer.ID || 0))
}

async function loadCustomerPetPets(customerId: number) {
  if (!customerId) {
    customerPetPets.value = []
    return
  }
  try {
    const res = await getCustomerPets(customerId)
    customerPetPets.value = Array.isArray(res.data) ? res.data : []
  } catch {
    customerPetPets.value = []
  }
}

function clearCustomerPetCustomer() {
  customerPetSelectedCustomer.value = null
  customerPetSelectedPet.value = null
  customerPetPets.value = []
  customerPetCustomerKeyword.value = ''
  customerPetCustomerOptions.value = []
}

async function saveCustomerPet() {
  if (!order.value || savingCustomerPet.value) return
  savingCustomerPet.value = true
  try {
    const res = await updateOrderCustomerPet(order.value.ID, {
      customer_id: customerPetSelectedCustomer.value?.ID || null,
      pet_id: customerPetSelectedPet.value?.ID || null,
    })
    order.value = res.data
    remarkDraft.value = resolveEditableRemark(order.value)
    closeCustomerPetModal()
    uni.showToast({ title: '归属已保存', icon: 'success' })
  } catch (e: any) {
    uni.showToast({ title: e?.msg || e?.message || '保存失败', icon: 'none' })
  } finally {
    savingCustomerPet.value = false
  }
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
  customerCard.value = null
  if (order.value?.customer_id) {
    try {
      const cardRes = await getCustomerCard(order.value.customer_id)
      if (cardRes.data) {
        customerCard.value = cardRes.data
        memberBalance.value = Number(cardRes.data.balance || 0)
      }
    } catch {}
  }

  if (
    customerCard.value &&
    memberBalance.value >= Number(order.value?.pay_amount || 0) &&
    !canChooseBalancePayment.value
  ) {
    await payWithBalance()
    return
  }

  showPayModal.value = true
}

function confirmPayMethod(method: string) {
  if (paying.value) return
  const label = payMethodConfirmMap[method] || method
  uni.showModal({
    title: '确认收款方式',
    content: `确认使用「${label}」收款 ¥${Number(order.value?.pay_amount || 0).toFixed(2)}？`,
    confirmText: label,
    success: async (res) => {
      if (res.confirm) {
        await doPay(method, true)
      }
    },
  })
}

async function doPay(method: string, confirmed = false, cashPayMethod?: string) {
  if (paying.value) return
  if (!confirmed) {
    confirmPayMethod(method)
    return
  }
  paying.value = true
  try {
    await payOrder(order.value.ID, method, undefined, remarkDraft.value.trim(), cashPayMethod)
    showPayModal.value = false
    uni.showToast({ title: '收款成功', icon: 'success' })
    await reload()
  } catch (e: any) {
    uni.showToast({ title: e.message || '收款失败', icon: 'none' })
  } finally {
    paying.value = false
  }
}

async function payWithBalance() {
  if (memberBalance.value <= 0) {
    uni.showToast({ title: '该客户未开通会员卡', icon: 'none' })
    return
  }
  if (memberBalance.value < order.value.pay_amount) {
    payWithMixedBalance()
    return
  }
  uni.showModal({
    title: '确认扣款',
    content: `从会员余额中扣除¥${order.value.pay_amount.toFixed(2)}？\n扣后余额：¥${(memberBalance.value - order.value.pay_amount).toFixed(2)}`,
    success: async (res) => {
      if (res.confirm) {
        await doPay('balance', true)
      }
    }
  })
}

function payWithMixedBalance() {
  const itemList = mixedBalanceCashPayMethods.map((item) => item.label)
  uni.showActionSheet({
    itemList,
    success: (res) => {
      const selected = mixedBalanceCashPayMethods[res.tapIndex]
      if (!selected) return
      const label = payMethodMap[selected.method] || selected.method
      uni.showModal({
        title: '余额补差收款',
        content: `会员余额 ¥${memberBalance.value.toFixed(2)} 将全部扣除，未覆盖部分不再享受会员折扣，由后端重新计算后使用「${label}」补差。`,
        confirmText: label,
        success: async (modalRes) => {
          if (modalRes.confirm) {
            await doPay('mixed_balance', true, selected.method)
          }
        },
      })
    },
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

async function goPetDetail(id?: number, petName?: string) {
  const petId = Number(id || 0)
  if (petId > 0) {
    uni.navigateTo({ url: `/pages/pet/edit?id=${petId}` })
    return
  }
  const fallbackPetId = await resolvePetIdByName(petName)
  if (fallbackPetId > 0) {
    uni.navigateTo({ url: `/pages/pet/edit?id=${fallbackPetId}` })
    return
  }
  uni.showToast({ title: '未找到这只猫咪档案', icon: 'none' })
}

function shouldShowPetGroupHead(group: any) {
  if (!group) return false
  if (group.pet_name === '零售商品') return true
  return petGroups.value.length > 1
}

function goCustomerDetail(id?: number) {
  const customerId = Number(id || 0)
  if (customerId <= 0) return
  uni.navigateTo({ url: `/pages/customer/detail?id=${customerId}` })
}

async function resolvePetIdByName(petName?: string) {
  const keyword = String(petName || '').trim()
  if (!keyword || resolvingPetName.value === keyword) return 0
  resolvingPetName.value = keyword
  uni.showLoading({ title: '查找猫咪中', mask: true })
  try {
    const res = await getPetList({ keyword, page: 1, page_size: 20 })
    const list = Array.isArray(res.data?.list) ? res.data.list : []
    const exactMatches = list.filter((pet) => String(pet.name || '').trim() === keyword)
    if (!exactMatches.length) return 0

    const orderCustomerId = Number(order.value?.customer_id || 0)
    const orderCustomerPhone = String(order.value?.customer?.phone || '').trim()

    const sameCustomer = exactMatches.find((pet) => Number(pet.customer_id || 0) === orderCustomerId)
    if (sameCustomer?.ID) return sameCustomer.ID

    const samePhone = exactMatches.find((pet) => String(pet.customer?.phone || '').trim() === orderCustomerPhone)
    if (samePhone?.ID) return samePhone.ID

    if (exactMatches.length === 1) return exactMatches[0].ID
    return 0
  } catch {
    return 0
  } finally {
    resolvingPetName.value = ''
    uni.hideLoading()
  }
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
.late-flag {
  color: #C2410C;
  font-weight: 600;
}
.pet-row {
  align-items: flex-start;
}
.pet-name-list {
  max-width: 70%;
  display: flex;
  flex-wrap: wrap;
  justify-content: flex-end;
  align-items: center;
  text-align: right;
  gap: 0;
}
.pet-name-link {
  display: inline-block;
}
.pet-name-separator {
  color: #94A3B8;
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
  cursor: pointer;
  text-decoration: none;
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
  height: 112rpx;
  padding: 14rpx 20rpx;
  border-radius: 18rpx;
  background: #F8FAFC;
  border: 2rpx solid #E2E8F0;
  box-sizing: border-box;
  font-size: 26rpx;
  color: #111827;
  line-height: 42rpx;
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
.btn.link { background: #F5F3FF; color: #4F46E5; border-color: #DDD6FE; }

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
  z-index: 900;
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
.link-modal {
  width: min(92vw, 720rpx);
  max-height: 84vh;
  overflow-y: auto;
  background: #fff;
  border-radius: 28rpx;
  padding: 28rpx;
  box-sizing: border-box;
  box-shadow: 0 24rpx 60rpx rgba(15, 23, 42, 0.24);
}
.modal-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16rpx;
  margin-bottom: 24rpx;
}
.modal-subtitle {
  display: block;
  margin-top: 8rpx;
  font-size: 22rpx;
  color: #94A3B8;
}
.modal-close {
  width: 56rpx;
  height: 56rpx;
  line-height: 52rpx;
  text-align: center;
  border-radius: 50%;
  background: #F3F4F6;
  color: #64748B;
  font-size: 34rpx;
  flex-shrink: 0;
}
.link-section {
  padding: 18rpx 0;
  border-top: 1rpx solid #F1F5F9;
}
.link-section-head {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 14rpx;
}
.link-section-title {
  font-size: 26rpx;
  font-weight: 800;
  color: #111827;
}
.link-clear {
  font-size: 24rpx;
  color: #4F46E5;
}
.selected-link-card {
  display: flex;
  justify-content: space-between;
  gap: 16rpx;
  padding: 18rpx;
  border-radius: 16rpx;
  background: #F8FAFC;
  margin-bottom: 14rpx;
}
.selected-link-name {
  font-size: 28rpx;
  font-weight: 800;
  color: #111827;
}
.selected-link-meta,
.link-option-meta,
.pet-choice-meta {
  font-size: 23rpx;
  color: #64748B;
}
.link-search {
  position: relative;
  height: 72rpx;
  border-radius: 16rpx;
  border: 1rpx solid #E2E8F0;
  background: #F8FAFC;
}
.link-search-input {
  width: 100%;
  height: 72rpx;
  padding: 0 22rpx;
  box-sizing: border-box;
  font-size: 26rpx;
  color: #111827;
}
.link-search-placeholder {
  position: absolute;
  left: 22rpx;
  top: 50%;
  transform: translateY(-50%);
  font-size: 24rpx;
  color: #94A3B8;
  pointer-events: none;
}
.link-options,
.pet-choice-list {
  display: grid;
  gap: 12rpx;
  margin-top: 14rpx;
}
.link-option,
.pet-choice {
  padding: 18rpx;
  border-radius: 16rpx;
  border: 1rpx solid #E2E8F0;
  background: #fff;
}
.link-option-name,
.pet-choice-name {
  display: block;
  margin-bottom: 6rpx;
  font-size: 27rpx;
  font-weight: 800;
  color: #111827;
}
.pet-choice.active {
  border-color: #4F46E5;
  background: #F5F3FF;
}
.link-empty {
  padding: 18rpx;
  border-radius: 16rpx;
  background: #F8FAFC;
  color: #64748B;
  font-size: 24rpx;
}
.link-actions {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 16rpx;
  margin-top: 22rpx;
}
.link-btn {
  margin: 0;
  height: 84rpx;
  border-radius: 18rpx;
  font-size: 28rpx;
  font-weight: 800;
}
.link-btn.ghost {
  background: #fff;
  color: #64748B;
  border: 1rpx solid #CBD5E1;
}
.link-btn.primary {
  background: #4F46E5;
  color: #fff;
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
.pay-choice-tip {
  margin: -10rpx 0 22rpx;
  padding: 18rpx 20rpx;
  border-radius: 18rpx;
  background: #F8FAFC;
  border: 1rpx solid #E2E8F0;
}
.pay-choice-title {
  display: block;
  font-size: 24rpx;
  font-weight: 800;
  color: #334155;
}
.pay-choice-desc {
  display: block;
  margin-top: 8rpx;
  font-size: 22rpx;
  color: #64748B;
  line-height: 1.5;
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
.btn.receipt { background: #FBF5E6; color: #8A6B2F; border-color: #E8D5A0; }
.btn.report { background: #EEF9F1; color: #2E7A48; border-color: #CFE7D7; }

/* ===== Receipt Modal ===== */
.receipt-outer {
  width: min(97vw, 840rpx);
  max-width: 840rpx;
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

.receipt {
  background: #FDF8ED;
  border-radius: 32rpx;
  box-shadow: 0 18rpx 48rpx rgba(116, 86, 36, 0.14);
  padding: 16rpx;
  font-family: -apple-system, 'PingFang SC', 'Helvetica Neue', sans-serif;
  display: flex;
  flex-direction: column;
  gap: 12rpx;
}
.receipt-lite-render {
  background: #FDF8ED !important;
  box-shadow: none !important;
}
.receipt-lite-render .receipt-card,
.receipt-lite-render .receipt-logo,
.receipt-lite-render .receipt-member-badge {
  background: #FFFDF8 !important;
  box-shadow: none !important;
}
.receipt-lite-render .receipt-logo-img {
  display: none !important;
}
.receipt-card {
  background: linear-gradient(180deg, #FFFDF8, #FFF9EE);
  border-radius: 24rpx;
  padding: 18rpx;
  border: 1rpx solid #F0E1BD;
  box-shadow: 0 10rpx 26rpx rgba(117, 89, 42, 0.08);
}
.receipt-brand-card {
  padding: 20rpx 18rpx 18rpx;
}
.receipt-brand-top {
  display: flex;
  align-items: flex-start;
  gap: 12rpx;
  margin-bottom: 30rpx;
}
.receipt-logo {
  width: 84rpx;
  height: 84rpx;
  min-width: 84rpx;
  border-radius: 999rpx;
  overflow: hidden;
  background: linear-gradient(135deg, #F4DE9A, #E8C46B);
  color: #6B5531;
  display: flex;
  align-items: center;
  justify-content: center;
}
.receipt-logo-emoji {
  font-size: 36rpx;
  line-height: 1;
  text-align: center;
  display: block;
  font-weight: 700;
}
.receipt-logo-img {
  width: 100%;
  height: 100%;
  border-radius: 999rpx;
}
.receipt-brand-main {
  flex: 1;
  min-width: 0;
  margin-top: -2rpx;
}
.receipt-shop {
  margin: 0;
  font-size: 32rpx;
  line-height: 1.12;
  font-weight: 700;
  color: #2F2A26;
  display: block;
}
.receipt-sub {
  margin-top: 2rpx;
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 0;
}
.receipt-sub-line {
  font-size: 20rpx;
  line-height: 1.24;
  color: #8F7950;
  display: block;
}
.receipt-brand-tag {
  padding: 6rpx 12rpx;
  border-radius: 999rpx;
  background: #F7E5B5;
  color: #8A6B2F;
  font-size: 18rpx;
  font-weight: 700;
  flex-shrink: 0;
}
.receipt-brand-tag.muted {
  background: #F7EEDB;
  color: #9B8760;
}
.receipt-meta-list {
  display: flex;
  flex-direction: column;
  gap: 10rpx;
}
.receipt-meta-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 12rpx;
}
.receipt-meta-label {
  font-size: 20rpx;
  color: #8F7950;
}
.receipt-meta-value {
  font-size: 22rpx;
  color: #2F2A26;
  text-align: right;
}
.receipt-meta-value-strong {
  font-weight: 700;
  color: #7A602E;
}
.receipt-member-badge {
  display: inline-flex;
  align-items: center;
  gap: 8rpx;
  padding: 6rpx 14rpx;
  border-radius: 999rpx;
  background: #FCF2D9;
  border: 1rpx solid #E7CB82;
}
.receipt-member-level {
  font-size: 18rpx;
  font-weight: 700;
  color: #9A6A21;
  letter-spacing: 0.5rpx;
}
.receipt-member-amount {
  font-size: 20rpx;
  font-weight: 600;
  color: #9C7426;
}
.receipt-group-card {
  padding-top: 12rpx;
}
.receipt-group-title {
  font-size: 22rpx;
  font-weight: 600;
  color: #6E5631;
  margin-bottom: 6rpx;
  display: block;
}
.receipt-item {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 12rpx;
  padding: 10rpx 0;
  border-bottom: 1rpx solid #EEE0BF;
}
.receipt-item:last-child {
  border-bottom: none;
  padding-bottom: 0;
}
.receipt-item-main {
  flex: 1;
  min-width: 0;
}
.receipt-item-name {
  font-size: 22rpx;
  line-height: 1.28;
  color: #2F2A26;
  margin-bottom: 4rpx;
  display: block;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
.receipt-item-meta {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  font-size: 18rpx;
  line-height: 1.3;
  color: #8F7950;
}
.receipt-item-dot {
  margin: 0 6rpx;
  color: #D2BE94;
}
.receipt-item-price {
  font-size: 24rpx;
  line-height: 1.2;
  color: #2F2A26;
  font-weight: 600;
  white-space: nowrap;
}
.receipt-summary-card {
  padding-top: 14rpx;
}
.receipt-summary-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 22rpx;
  color: #3B342F;
  padding: 5rpx 0;
}
.receipt-summary-row.saving {
  color: #9A7121;
}
.receipt-summary-divider {
  height: 1rpx;
  background: #E9D8B2;
  margin: 8rpx 0 12rpx;
}
.receipt-total-row {
  display: flex;
  align-items: flex-end;
  justify-content: space-between;
  gap: 12rpx;
}
.receipt-total-copy {
  display: flex;
  flex-direction: column;
  gap: 4rpx;
}
.receipt-total-label {
  font-size: 20rpx;
  color: #8F7950;
}
.receipt-total-tip {
  font-size: 18rpx;
  color: #AF812A;
}
.receipt-total-price {
  font-size: 52rpx;
  line-height: 1;
  font-weight: 300;
  color: #C4A35A;
  letter-spacing: -1rpx;
}
.receipt-balance-row {
  margin-top: 12rpx;
  padding-top: 12rpx;
  border-top: 1rpx solid #E9D8B2;
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.receipt-footer-card {
  padding-top: 12rpx;
  padding-bottom: 12rpx;
}
.receipt-footer-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 12rpx;
  padding: 5rpx 0;
}
.receipt-footer-label {
  font-size: 20rpx;
  color: #8F7950;
}
.receipt-footer-value {
  font-size: 22rpx;
  color: #2F2A26;
  text-align: right;
}
.receipt-thanks {
  padding: 4rpx 8rpx 0;
  text-align: center;
}
.receipt-thanks-cn {
  font-size: 18rpx;
  color: #4A3F2F;
  letter-spacing: 1rpx;
  display: block;
}
.receipt-thanks-en {
  font-size: 15rpx;
  color: #B18939;
  font-weight: 300;
  display: block;
  margin-top: 2rpx;
}

/* ---- 操作按钮区 ---- */
.receipt-actions {
  display: flex;
  gap: 12rpx;
  padding: 12rpx 0 calc(4rpx + env(safe-area-inset-bottom));
  flex-shrink: 0;
}
.btn-receipt-save,
.btn-receipt-close {
  flex: 1;
  min-height: 76rpx;
  border-radius: 16rpx;
  font-size: 24rpx;
  font-weight: 700;
  text-align: center;
  display: flex;
  align-items: center;
  justify-content: center;
  box-sizing: border-box;
}
.btn-receipt-save { background: linear-gradient(135deg, #E8C86E, #D7A843); color: #5E4617; box-shadow: 0 14rpx 28rpx rgba(164, 122, 32, 0.22); }
.btn-receipt-close { background: #FFF6E4; color: #8A6B2F; border: 2rpx solid #E7CB82; }

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
.receipt-image-hint { font-size: 26rpx; color: #B18939; font-weight: 500; display: block; margin-bottom: 16rpx; }
.receipt-image { width: 100%; border-radius: 16rpx; box-shadow: 0 4rpx 20rpx rgba(0,0,0,0.1); }
.receipt-image-native { display: block; height: auto; -webkit-touch-callout: default; -webkit-user-select: auto; user-select: auto; object-fit: contain; }
</style>
