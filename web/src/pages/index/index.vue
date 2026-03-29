<template>
  <SideLayout>
    <view class="workstation">
      <view class="hero-panel">
        <view class="hero-copy">
          <text class="hero-eyebrow">{{ todayStr }} {{ weekDay }}</text>
          <text class="hero-title">{{ greetingText }}，{{ staffName }}</text>
          <text class="hero-subtitle">今天先盯紧到店节奏、营收转化和待处理预约，把门店最重要的动作放在第一屏。</text>
        </view>
        <view class="hero-main-card" @click="go('/pages/dashboard/index')">
          <text class="hero-main-label">今日营收</text>
          <text class="hero-main-value">¥{{ overview.today_revenue.toFixed(0) }}</text>
          <text class="hero-main-foot">点击查看完整经营看板</text>
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
          <text class="ops-label">今日订单</text>
          <text class="ops-value">{{ overview.today_order_count }}</text>
          <text class="ops-desc">结算笔数与开单进度</text>
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
        <view class="quick-actions">
          <view class="action-item action-primary" @click="go('/pages/order/create')">
            <view class="action-icon">🧾</view>
            <view class="action-copy">
              <text class="action-title">立即开单</text>
              <text class="action-desc">到店后快速记录服务与商品</text>
            </view>
          </view>
          <view class="action-item" @click="go('/pages/appointment/create')">
            <view class="action-icon">📅</view>
            <view class="action-copy">
              <text class="action-title">新建预约</text>
              <text class="action-desc">补录电话预约或现场预约</text>
            </view>
          </view>
          <view class="action-item" @click="go('/pages/customer/list')">
            <view class="action-icon">👥</view>
            <view class="action-copy">
              <text class="action-title">客户管理</text>
              <text class="action-desc">查看会员卡、余额和标签信息</text>
            </view>
          </view>
          <view class="action-item" @click="go('/pages/pet/list')">
            <view class="action-icon">🐱</view>
            <view class="action-copy">
              <text class="action-title">猫咪档案</text>
              <text class="action-desc">快速查找宠物资料与偏好</text>
            </view>
          </view>
        </view>
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
              <text class="summary-label">今日订单</text>
              <text class="summary-value">{{ overview.today_order_count }}</text>
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
import { getDashboardOverview } from '@/api/dashboard'
import { getAppointmentCalendar } from '@/api/appointment'

const authStore = useAuthStore()
const staffName = computed(() => authStore.staffInfo?.name || '员工')

const now = new Date()
const todayStr = `${now.getMonth() + 1}月${now.getDate()}日`
const weekDays = ['周日', '周一', '周二', '周三', '周四', '周五', '周六']
const weekDay = weekDays[now.getDay()]
const hour = now.getHours()
const greetingText = hour < 12 ? '早上好' : hour < 18 ? '下午好' : '晚上好'

const statusColors: Record<number, string> = {
  0: '#D97706', 1: '#4338CA', 2: '#059669', 3: '#0284C7', 4: '#6B7280', 5: '#DC2626',
}

const statusText: Record<number, string> = {
  0: '待确认', 1: '已确认', 2: '服务中', 3: '待结算', 4: '已取消', 5: '未到店',
}

const overview = ref({
  today_revenue: 0, today_order_count: 0, today_appointment_count: 0,
  today_new_customers: 0, pending_appointments: 0, total_customers: 0,
})
const todayAppts = ref<any[]>([])

function go(url: string) {
  uni.navigateTo({ url })
}

function localDateStr(d: Date = new Date()) {
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`
}

function getPetName(appt: any) {
  if (appt.pets?.length) return appt.pets.map((p: any) => p.pet?.name || '').filter(Boolean).join('、') || '-'
  return appt.pet?.name || '-'
}

function getCustomerName(appt: any) {
  return appt.customer?.nickname || appt.customer?.phone || '-'
}

async function loadData() {
  try {
    const [ovRes, apptRes] = await Promise.all([
      getDashboardOverview(),
      getAppointmentCalendar(localDateStr(), localDateStr()),
    ])
    overview.value = ovRes.data
    todayAppts.value = (apptRes.data || [])
      .filter((a: any) => a.status !== 4)
      .sort((a: any, b: any) => a.start_time.localeCompare(b.start_time))
  } catch {}
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
.hero-main-card {
  border-radius: 28rpx;
  padding: 32rpx 30rpx;
  box-sizing: border-box;
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
  justify-content: space-between;
  box-shadow: 0 14rpx 34rpx rgba(234, 88, 12, 0.2);
}

.hero-main-label {
  font-size: 24rpx;
  opacity: 0.9;
}

.hero-main-value {
  display: block;
  margin-top: 14rpx;
  font-size: 56rpx;
  font-weight: 800;
}

.hero-main-foot {
  display: block;
  margin-top: 18rpx;
  font-size: 22rpx;
  opacity: 0.86;
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

.quick-actions {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 16rpx;
}

.action-item {
  background: #FFFFFF;
  border: 1rpx solid #E5E7EB;
  border-radius: 22rpx;
  padding: 22rpx;
}

.action-primary {
  background: linear-gradient(150deg, #FFF7ED, #FFFFFF);
  border-color: #FDBA74;
}

.action-icon {
  width: 76rpx;
  height: 76rpx;
  border-radius: 20rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #111827, #374151);
  font-size: 34rpx;
}

.action-copy {
  margin-top: 18rpx;
}

.action-title {
  display: block;
  color: #111827;
  font-size: 26rpx;
  font-weight: 700;
}

.action-desc {
  display: block;
  margin-top: 8rpx;
  color: #94A3B8;
  font-size: 22rpx;
  line-height: 1.5;
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
  .content-grid,
  .quick-actions,
  .ops-grid {
    grid-template-columns: 1fr;
  }
}
</style>
