<template>
  <SideLayout>
    <view class="workstation">
      <view class="hero-panel">
        <view class="hero-copy">
          <text class="hero-eyebrow">{{ todayStr }} {{ weekDay }}</text>
          <text class="hero-title">{{ greetingText }}，{{ staffName }}</text>
          <text class="hero-subtitle">今天先盯紧到店节奏、营收转化和待处理预约，把门店最重要的动作放在第一屏。</text>
        </view>
        <view class="hero-metrics">
          <view class="hero-main-card" @click="go('/pages/dashboard/index')">
            <text class="hero-main-label">今日营收</text>
            <text class="hero-main-value">¥{{ formatRevenueAmount(overview.today_revenue) }}</text>
            <view v-if="todayRevenueBreakdown.length" class="hero-breakdown">
              <view v-for="item in todayRevenueBreakdown" :key="`today-${item.key}`" class="hero-breakdown-chip hero-breakdown-chip-light">
                <text class="hero-breakdown-key">{{ item.label }}</text>
                <text class="hero-breakdown-amount">¥{{ formatRevenueAmount(item.amount) }}</text>
              </view>
            </view>
            <text class="hero-main-foot">点击查看完整经营看板</text>
          </view>
          <picker
            class="hero-date-picker"
            mode="date"
            :value="recentRevenueDate"
            :start="recentRevenueMinDate"
            :end="recentRevenueMaxDate"
            @change="handleRecentRevenueDateChange"
          >
            <view class="hero-recent-card">
              <view class="hero-recent-head">
                <text class="hero-recent-label">近日营收</text>
                <text class="hero-recent-date">{{ recentRevenueDateLabel }}</text>
              </view>
              <text class="hero-recent-value">{{ recentRevenueValueText }}</text>
              <view v-if="recentRevenueBreakdown.length" class="hero-breakdown">
                <view v-for="item in recentRevenueBreakdown" :key="`recent-${item.key}`" class="hero-breakdown-chip">
                  <text class="hero-breakdown-key">{{ item.label }}</text>
                  <text class="hero-breakdown-amount">¥{{ formatRevenueAmount(item.amount) }}</text>
                </view>
              </view>
              <text class="hero-recent-foot">点击切换日期</text>
            </view>
          </picker>
        </view>
      </view>

      <view class="ops-grid">
        <view class="ops-card warn" @click="go('/pages/appointment/list')">
          <text class="ops-label">待处理预约</text>
          <text class="ops-value">{{ overview.pending_appointments }}</text>
          <text class="ops-desc">待确认、待分配、待跟进</text>
        </view>
        <view class="ops-card cool" @click="go('/pages/appointment/calendar')">
          <text class="ops-label">今日预约</text>
          <text class="ops-value">{{ overview.today_appointment_count }}</text>
          <text class="ops-desc">查看全天到店与排班节奏</text>
        </view>
        <view class="ops-card neutral">
          <text class="ops-label">今日已支付</text>
          <text class="ops-value">{{ overview.today_order_count }}</text>
          <text class="ops-desc">已完成收款的订单数</text>
        </view>
        <view class="ops-card neutral">
          <text class="ops-label">今日新客</text>
          <text class="ops-value">{{ overview.today_new_customers }}</text>
          <text class="ops-desc">今天新增进店客户</text>
        </view>
      </view>

      <view class="section section-actions">
        <view class="section-header">
          <view>
            <text class="section-title">快捷动作</text>
            <text class="section-subtitle">高频动作放在最前面，避免切页找入口。</text>
          </view>
        </view>
        <scroll-view scroll-x class="quick-actions-scroll" show-scrollbar="false">
          <view class="quick-actions">
            <view class="action-item action-primary" @click="go('/pages/order/create')">
              <view class="action-icon">🧾</view>
              <view class="action-copy">
                <text class="action-title">立即开单</text>
                <text class="action-desc">快速记服务</text>
              </view>
            </view>
            <view class="action-item" @click="go('/pages/appointment/create')">
              <view class="action-icon">📅</view>
              <view class="action-copy">
                <text class="action-title">新建预约</text>
                <text class="action-desc">补录或现场约</text>
              </view>
            </view>
            <view class="action-item" @click="go('/pages/customer/list')">
              <view class="action-icon">👥</view>
              <view class="action-copy">
                <text class="action-title">客户管理</text>
                <text class="action-desc">会员与余额</text>
              </view>
            </view>
            <view class="action-item" @click="go('/pages/boarding/create')">
              <view class="action-icon">🏨</view>
              <view class="action-copy">
                <text class="action-title">寄养开单</text>
                <text class="action-desc">选柜并登记</text>
              </view>
            </view>
            <view class="action-item" @click="go('/pages/pet/list')">
              <view class="action-icon action-icon-cat">
                <image class="action-icon-image" :src="catSticker" mode="aspectFit" />
              </view>
              <view class="action-copy">
                <text class="action-title">猫咪档案</text>
                <text class="action-desc">资料与偏好</text>
              </view>
            </view>
          </view>
        </scroll-view>
      </view>

      <view class="content-grid">
        <view class="section">
          <view class="section-header">
            <view>
              <text class="section-title">今日预约流</text>
              <text class="section-subtitle">从时间和状态快速扫出今天的工作节奏。</text>
            </view>
            <text class="section-more" @click="go('/pages/appointment/calendar')">查看全部 ›</text>
          </view>
          <view v-if="todayAppts.length === 0" class="no-data">
            <text class="no-data-title">今日暂无预约</text>
            <text class="no-data-text">可以先安排新预约，或者检查是否有临时加单。</text>
          </view>
          <view v-else class="appt-list">
            <view
              class="appt-item"
              v-for="appt in todayAppts.slice(0, 5)"
              :key="appt.ID"
              @click="go('/pages/appointment/detail?id=' + appt.ID)"
            >
              <view class="appt-time-col">
                <text class="appt-time">{{ appt.start_time }}</text>
                <text class="appt-end">{{ appt.end_time }}</text>
              </view>
              <view class="appt-info">
                <text class="appt-pet">{{ getPetName(appt) }}</text>
                <text class="appt-customer">{{ getCustomerName(appt) }}</text>
              </view>
              <view class="appt-status">
                <text class="appt-status-pill" :style="{ background: `${statusColors[appt.status] || '#9CA3AF'}22`, color: statusColors[appt.status] || '#6B7280' }">
                  {{ statusText[appt.status] || '未知状态' }}
                </text>
              </view>
            </view>
          </view>
        </view>

        <view class="section section-side">
          <view class="section-header">
            <view>
              <text class="section-title">客户经营概览</text>
              <text class="section-subtitle">帮助店员判断今天该关注老客还是新客。</text>
            </view>
          </view>
          <view class="summary-list">
            <view class="summary-row">
              <text class="summary-label">总客户数</text>
              <text class="summary-value">{{ overview.total_customers }}</text>
            </view>
            <view class="summary-row">
              <text class="summary-label">今日新客</text>
              <text class="summary-value">{{ overview.today_new_customers }}</text>
            </view>
            <view class="summary-row">
              <text class="summary-label">今日已支付</text>
              <text class="summary-value">{{ overview.today_order_count }}</text>
            </view>
            <view class="summary-row">
              <text class="summary-label">待结算</text>
              <text class="summary-value">{{ overview.today_pending_settlement_count }}</text>
            </view>
          </view>
          <view class="side-tip">
            <text class="side-tip-title">建议动作</text>
            <text class="side-tip-text">待处理预约较多时，优先清理预约列表；新客增多时，记得同步完善客户和会员信息。</text>
          </view>
        </view>
      </view>
    </view>
  </SideLayout>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { onShow } from '@dcloudio/uni-app'
import { useAuthStore } from '@/store/auth'
import SideLayout from '@/components/SideLayout.vue'
import type { DashboardOverview, DashboardPaymentBreakdownItem } from '@/api/dashboard'
import { getDashboardOverview } from '@/api/dashboard'
import { getAppointmentCalendar } from '@/api/appointment'
import catSticker from '@/assets/cat-sticker.jpg'

const authStore = useAuthStore()
const staffName = computed(() => authStore.staffInfo?.name || '员工')

const now = new Date()
const todayStr = `${now.getMonth() + 1}月${now.getDate()}日`
const weekDays = ['周日', '周一', '周二', '周三', '周四', '周五', '周六']
const weekDay = weekDays[now.getDay()]
const hour = now.getHours()
const greetingText = hour < 12 ? '早上好' : hour < 18 ? '下午好' : '晚上好'

const statusColors: Record<number, string> = {
  0: '#D97706', 1: '#4338CA', 2: '#059669', 3: '#0284C7', 4: '#6B7280', 5: '#DC2626', 7: '#7C3AED',
}

const statusText: Record<number, string> = {
  0: '待确认', 1: '已确认', 2: '服务中', 3: '待结算', 4: '已取消', 5: '未到店', 7: '已开单',
}

const emptyOverview = (): DashboardOverview => ({
  today_revenue: 0, today_order_count: 0, today_appointment_count: 0,
  today_service_completed_count: 0, today_pending_settlement_count: 0, today_refunded_order_count: 0,
  today_new_customers: 0, pending_appointments: 0, total_customers: 0,
  avg_order_value: 0, no_show_rate: 0, no_show_count: 0, total_appointments: 0,
  payment_breakdown: [],
})

const overview = ref<DashboardOverview>(emptyOverview())
const todayAppts = ref<any[]>([])
const recentOverview = ref<DashboardOverview | null>(null)
const recentRevenueLoading = ref(false)

const today = new Date()
const recentRevenueMaxDate = localDateStr(today)
const recentRevenueMinDate = localDateStr(addDays(today, -15))
const recentRevenueDate = ref(localDateStr(addDays(today, -1)))

const recentRevenueDateLabel = computed(() => formatShortDate(recentRevenueDate.value))
const todayRevenueBreakdown = computed(() => normalizeBreakdown(overview.value.payment_breakdown))
const recentRevenueBreakdown = computed(() => normalizeBreakdown(recentOverview.value?.payment_breakdown || []))
const recentRevenueValueText = computed(() => {
  if (recentRevenueLoading.value) return '...'
  if (!recentOverview.value) return '--'
  return `¥${formatRevenueAmount(recentOverview.value.today_revenue)}`
})

function go(url: string) {
  uni.navigateTo({ url })
}

function addDays(date: Date, days: number) {
  const next = new Date(date)
  next.setDate(next.getDate() + days)
  return next
}

function localDateStr(d: Date = new Date()) {
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`
}

function formatShortDate(dateStr: string) {
  const [year, month, day] = dateStr.split('-').map((item) => Number(item))
  if (!year || !month || !day) return dateStr
  return `${month}月${day}日`
}

function normalizeBreakdown(items: DashboardPaymentBreakdownItem[]) {
  return (items || []).filter((item) => Number(item?.amount || 0) > 0)
}

function formatRevenueAmount(amount: number) {
  return Number(amount || 0).toFixed(2)
}

function getPetName(appt: any) {
  if (appt.pets?.length) return appt.pets.map((p: any) => p.pet?.name || '').filter(Boolean).join('、') || '-'
  return appt.pet?.name || '-'
}

function getCustomerName(appt: any) {
  return appt.customer?.nickname || appt.customer?.phone || '-'
}

async function loadRecentRevenue(date = recentRevenueDate.value, showError = false) {
  recentRevenueLoading.value = true
  try {
    const res = await getDashboardOverview(date, date)
    recentOverview.value = res.data || emptyOverview()
    recentRevenueDate.value = date
  } catch {
    recentOverview.value = null
    if (showError) {
      uni.showToast({ title: '近日营收加载失败', icon: 'none' })
    }
  } finally {
    recentRevenueLoading.value = false
  }
}

function handleRecentRevenueDateChange(event: any) {
  const nextDate = String(event?.detail?.value || '').trim()
  if (!nextDate || nextDate === recentRevenueDate.value) return
  loadRecentRevenue(nextDate, true)
}

async function loadData() {
  const [ovRes, apptRes] = await Promise.allSettled([
    getDashboardOverview(),
    getAppointmentCalendar(localDateStr(), localDateStr()),
    loadRecentRevenue(recentRevenueDate.value),
  ])

  if (ovRes.status === 'fulfilled') {
    overview.value = ovRes.value.data
  }

  if (apptRes.status === 'fulfilled') {
    todayAppts.value = (apptRes.value.data || [])
      .filter((a: any) => a.status !== 4)
      .sort((a: any, b: any) => a.start_time.localeCompare(b.start_time))
  } else {
    todayAppts.value = []
  }
}

onShow(loadData)
</script>

<style scoped>
.workstation {
  min-height: 100vh;
  background:
    radial-gradient(circle at top left, rgba(249, 115, 22, 0.08), transparent 24%),
    linear-gradient(180deg, #FFF8F1 0%, #F7F8FC 22%, #F5F6FA 100%);
  padding: 28rpx 32rpx 48rpx;
}

.hero-panel {
  display: grid;
  grid-template-columns: 1.35fr 0.85fr;
  gap: 18rpx;
  margin-bottom: 22rpx;
}

.hero-copy,
.hero-main-card,
.hero-recent-card {
  border-radius: 28rpx;
  padding: 32rpx 30rpx;
  box-sizing: border-box;
}

.hero-metrics {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 18rpx;
}

.hero-copy {
  background: linear-gradient(145deg, #1F2937, #0F172A);
  color: #FFFFFF;
  box-shadow: 0 16rpx 36rpx rgba(15, 23, 42, 0.16);
}

.hero-eyebrow {
  display: block;
  color: #FDE68A;
  font-size: 22rpx;
  letter-spacing: 2rpx;
}

.hero-title {
  display: block;
  margin-top: 16rpx;
  font-size: 48rpx;
  font-weight: 800;
  line-height: 1.15;
}

.hero-subtitle {
  display: block;
  margin-top: 16rpx;
  color: rgba(255, 255, 255, 0.7);
  font-size: 26rpx;
  line-height: 1.7;
}

.hero-main-card {
  background: linear-gradient(160deg, #F97316, #EA580C);
  color: #FFFFFF;
  display: flex;
  flex-direction: column;
  box-shadow: 0 14rpx 34rpx rgba(234, 88, 12, 0.2);
}

.hero-date-picker {
  display: block;
}

.hero-recent-card {
  min-height: 100%;
  background: linear-gradient(160deg, #FFF6D8, #F5E4AA);
  color: #2F2A26;
  display: flex;
  flex-direction: column;
  border: 1rpx solid rgba(201, 165, 63, 0.22);
  box-shadow: 0 14rpx 30rpx rgba(168, 126, 32, 0.12);
}

.hero-main-label {
  font-size: 24rpx;
  opacity: 0.9;
}

.hero-recent-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 10rpx;
}

.hero-recent-label {
  font-size: 24rpx;
  color: rgba(47, 42, 38, 0.86);
}

.hero-recent-date {
  flex-shrink: 0;
  padding: 6rpx 12rpx;
  border-radius: 999rpx;
  font-size: 20rpx;
  color: #7C5F22;
  background: rgba(255, 255, 255, 0.58);
}

.hero-main-value {
  display: block;
  margin-top: 14rpx;
  font-size: 56rpx;
  font-weight: 800;
}

.hero-recent-value {
  display: block;
  margin-top: 14rpx;
  font-size: 48rpx;
  font-weight: 800;
  color: #2F2A26;
}

.hero-main-foot {
  display: block;
  margin-top: auto;
  font-size: 22rpx;
  opacity: 0.86;
}

.hero-recent-foot {
  display: block;
  margin-top: auto;
  font-size: 22rpx;
  color: #7C6F5C;
}

.hero-breakdown {
  display: flex;
  flex-wrap: wrap;
  gap: 10rpx;
  margin-top: 16rpx;
  margin-bottom: 18rpx;
}

.hero-breakdown-chip {
  display: inline-flex;
  align-items: center;
  gap: 8rpx;
  min-width: 0;
  padding: 8rpx 12rpx;
  border-radius: 999rpx;
  background: rgba(255, 255, 255, 0.48);
  color: #5C4A1F;
}

.hero-breakdown-chip-light {
  background: rgba(255, 255, 255, 0.18);
  color: rgba(255, 255, 255, 0.96);
}

.hero-breakdown-key,
.hero-breakdown-amount {
  font-size: 20rpx;
  line-height: 1.2;
  white-space: nowrap;
}

.hero-breakdown-key {
  opacity: 0.82;
}

.hero-breakdown-amount {
  font-weight: 700;
}

.ops-grid {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 16rpx;
  margin-bottom: 22rpx;
}

.ops-card {
  background: #FFFFFF;
  border-radius: 24rpx;
  padding: 24rpx 22rpx;
  box-shadow: 0 8rpx 24rpx rgba(15, 23, 42, 0.05);
}

.ops-card.warn {
  background: linear-gradient(180deg, #FFF7ED, #FFFFFF);
  border: 1rpx solid #FDBA74;
}

.ops-card.cool {
  background: linear-gradient(180deg, #EEF2FF, #FFFFFF);
  border: 1rpx solid #C7D2FE;
}

.ops-card.neutral {
  border: 1rpx solid #E5E7EB;
}

.ops-label {
  display: block;
  color: #6B7280;
  font-size: 22rpx;
}

.ops-value {
  display: block;
  margin-top: 10rpx;
  color: #111827;
  font-size: 42rpx;
  font-weight: 800;
}

.ops-desc {
  display: block;
  margin-top: 10rpx;
  color: #94A3B8;
  font-size: 22rpx;
  line-height: 1.5;
}

.section {
  background: rgba(255, 255, 255, 0.92);
  border-radius: 24rpx;
  padding: 26rpx;
  box-shadow: 0 10rpx 30rpx rgba(15, 23, 42, 0.05);
}

.section-actions {
  margin-bottom: 22rpx;
}

.section-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 18rpx;
  margin-bottom: 20rpx;
}

.section-title {
  display: block;
  font-size: 30rpx;
  font-weight: 700;
  color: #111827;
}

.section-subtitle {
  display: block;
  margin-top: 8rpx;
  font-size: 22rpx;
  color: #9CA3AF;
  line-height: 1.6;
}

.section-more {
  font-size: 24rpx;
  color: #F97316;
  white-space: nowrap;
}

.quick-actions-scroll {
  width: 100%;
}

.quick-actions {
  display: flex;
  gap: 16rpx;
  width: max-content;
}

.action-item {
  width: 180rpx;
  background: #FFFFFF;
  border: 1rpx solid #E5E7EB;
  border-radius: 22rpx;
  padding: 20rpx 18rpx;
  box-sizing: border-box;
}

.action-primary {
  background: linear-gradient(150deg, #FFF7ED, #FFFFFF);
  border-color: #FDBA74;
}

.action-icon {
  width: 64rpx;
  height: 64rpx;
  border-radius: 18rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #111827, #374151);
  font-size: 30rpx;
}

.action-icon-cat {
  padding: 6rpx;
  background: linear-gradient(135deg, #FFF7ED, #FDE68A);
}

.action-icon-image {
  width: 100%;
  height: 100%;
  display: block;
}

.action-copy {
  margin-top: 14rpx;
}

.action-title {
  display: block;
  color: #111827;
  font-size: 25rpx;
  font-weight: 700;
}

.action-desc {
  display: block;
  margin-top: 6rpx;
  color: #94A3B8;
  font-size: 20rpx;
  line-height: 1.4;
}

.content-grid {
  display: grid;
  grid-template-columns: 1.35fr 0.85fr;
  gap: 18rpx;
}

.section-side {
  background: linear-gradient(180deg, #FFFFFF, #FFF7ED);
}

.no-data {
  text-align: center;
  padding: 56rpx 0;
}

.no-data-title {
  display: block;
  color: #374151;
  font-size: 28rpx;
  font-weight: 700;
}

.no-data-text {
  display: block;
  margin-top: 10rpx;
  color: #9CA3AF;
  font-size: 24rpx;
}

.appt-list {
  display: flex;
  flex-direction: column;
  gap: 12rpx;
}

.appt-item {
  display: flex;
  align-items: center;
  gap: 18rpx;
  padding: 18rpx 0;
  border-bottom: 1rpx solid #F1F5F9;
}

.appt-item:last-child {
  border-bottom: none;
}

.appt-time-col {
  display: flex;
  flex-direction: column;
  align-items: center;
  min-width: 94rpx;
}

.appt-time {
  font-size: 30rpx;
  font-weight: 700;
  color: #111827;
}

.appt-end {
  font-size: 20rpx;
  color: #94A3B8;
  margin-top: 4rpx;
}

.appt-info {
  flex: 1;
  min-width: 0;
}

.appt-pet {
  display: block;
  font-size: 28rpx;
  font-weight: 600;
  color: #111827;
}

.appt-customer {
  display: block;
  margin-top: 6rpx;
  font-size: 22rpx;
  color: #6B7280;
}

.appt-status {
  flex-shrink: 0;
}

.appt-status-pill {
  font-size: 22rpx;
  padding: 8rpx 14rpx;
  border-radius: 999rpx;
}

.summary-list {
  display: flex;
  flex-direction: column;
  gap: 14rpx;
}

.summary-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16rpx 18rpx;
  border-radius: 18rpx;
  background: rgba(255, 255, 255, 0.8);
  border: 1rpx solid #F1F5F9;
}

.summary-label {
  color: #6B7280;
  font-size: 24rpx;
}

.summary-value {
  color: #111827;
  font-size: 30rpx;
  font-weight: 800;
}

.side-tip {
  margin-top: 18rpx;
  padding: 20rpx;
  border-radius: 20rpx;
  background: #FFF7ED;
  border: 1rpx solid #FED7AA;
}

.side-tip-title {
  display: block;
  color: #C2410C;
  font-size: 24rpx;
  font-weight: 700;
}

.side-tip-text {
  display: block;
  margin-top: 10rpx;
  color: #9A3412;
  font-size: 22rpx;
  line-height: 1.7;
}

@media (max-width: 900px) {
  .hero-panel,
  .content-grid {
    grid-template-columns: 1fr;
  }

  .hero-copy,
  .hero-main-card,
  .hero-recent-card,
  .section {
    padding: 24rpx;
  }

  .hero-metrics {
    gap: 14rpx;
  }

  .hero-subtitle {
    margin-top: 12rpx;
    font-size: 24rpx;
    line-height: 1.55;
  }

  .hero-main-value {
    margin-top: 10rpx;
    font-size: 52rpx;
  }

  .hero-recent-value {
    margin-top: 10rpx;
    font-size: 42rpx;
  }

  .hero-main-foot {
    margin-top: 12rpx;
    font-size: 20rpx;
  }

  .hero-recent-foot {
    margin-top: 12rpx;
    font-size: 20rpx;
  }

  .ops-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: 14rpx;
  }

  .ops-card {
    padding: 20rpx 18rpx;
    border-radius: 20rpx;
    min-height: 160rpx;
    box-sizing: border-box;
  }

  .ops-label {
    font-size: 21rpx;
  }

  .ops-value {
    margin-top: 8rpx;
    font-size: 38rpx;
  }

  .ops-desc {
    margin-top: 8rpx;
    font-size: 20rpx;
    line-height: 1.4;
  }

  .section-header {
    margin-bottom: 16rpx;
  }

  .summary-list {
    gap: 10rpx;
  }

  .summary-row {
    padding: 14rpx 16rpx;
  }

  .side-tip {
    margin-top: 14rpx;
    padding: 18rpx;
  }
}

@media (max-width: 520px) {
  .workstation {
    padding: 20rpx 24rpx 40rpx;
  }

  .hero-copy {
    padding-bottom: 22rpx;
  }

  .hero-title {
    font-size: 42rpx;
  }

  .hero-metrics {
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: 12rpx;
  }

  .hero-recent-label,
  .hero-main-label {
    font-size: 22rpx;
  }

  .hero-recent-date {
    padding: 4rpx 10rpx;
    font-size: 18rpx;
  }

  .hero-main-value {
    font-size: 46rpx;
  }

  .hero-recent-value {
    font-size: 38rpx;
  }

  .ops-card {
    min-height: 148rpx;
  }

  .ops-desc {
    font-size: 19rpx;
  }
}
</style>
