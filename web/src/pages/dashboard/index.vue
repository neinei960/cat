<template>
  <SideLayout>
  <view class="page">
    <view class="header">
      <text class="title">数据看板</text>
      <view class="date-range">
        <picker mode="date" :value="startDate" @change="(e:any) => { startDate = e.detail.value; loadAll() }">
          <text class="date-btn">{{ startDate }}</text>
        </picker>
        <text class="sep">至</text>
        <picker mode="date" :value="endDate" @change="(e:any) => { endDate = e.detail.value; loadAll() }">
          <text class="date-btn">{{ endDate }}</text>
        </picker>
      </view>
    </view>

    <!-- Quick date buttons -->
    <view class="quick-dates">
      <text :class="['qd-btn', quickActive === 'today' ? 'active' : '']" @click="setQuickDate('today')">今日</text>
      <text :class="['qd-btn', quickActive === 'week' ? 'active' : '']" @click="setQuickDate('week')">本周</text>
      <text :class="['qd-btn', quickActive === 'month' ? 'active' : '']" @click="setQuickDate('month')">本月</text>
      <text :class="['qd-btn', quickActive === 'lastMonth' ? 'active' : '']" @click="setQuickDate('lastMonth')">上月</text>
    </view>

    <!-- Overview cards -->
    <view class="overview-grid">
      <view class="stat-card primary">
        <text class="stat-val">¥{{ overview.today_revenue.toFixed(2) }}</text>
        <text class="stat-label">营业额</text>
      </view>
      <view class="stat-card">
        <text class="stat-val">{{ overview.today_order_count }}</text>
        <text class="stat-label">订单数</text>
      </view>
      <view class="stat-card">
        <text class="stat-val">{{ overview.today_appointment_count }}</text>
        <text class="stat-label">预约数</text>
      </view>
      <view class="stat-card">
        <text class="stat-val">{{ overview.pending_appointments }}</text>
        <text class="stat-label">待处理预约</text>
      </view>
      <view class="stat-card">
        <text class="stat-val">{{ overview.today_new_customers }}</text>
        <text class="stat-label">新客户</text>
      </view>
      <view class="stat-card">
        <text class="stat-val">{{ overview.total_customers }}</text>
        <text class="stat-label">总客户</text>
      </view>
    </view>

    <!-- Revenue trend -->
    <view class="section">
      <text class="section-title">营收趋势</text>
      <view class="chart-placeholder" v-if="revenueData.length === 0">暂无数据</view>
      <view v-else class="bar-chart">
        <view class="bar-item" v-for="d in revenueData" :key="d.date">
          <view class="bar" :style="{ height: getBarHeight(d.revenue) + 'rpx' }"></view>
          <text class="bar-label">{{ d.date.substring(5) }}</text>
          <text class="bar-val">¥{{ d.revenue }}</text>
        </view>
      </view>
    </view>

    <!-- Service ranking (按项目筛选营业额) -->
    <view class="section">
      <text class="section-title">服务项目营业额</text>
      <view class="rank-list">
        <view class="rank-item header" v-if="serviceRanking.length > 0">
          <text class="rank-no">#</text>
          <text class="rank-name">项目名称</text>
          <text class="rank-count">次数</text>
          <text class="rank-revenue">营业额</text>
        </view>
        <view class="rank-item" v-for="(s, i) in serviceRanking" :key="s.service_name">
          <text class="rank-no">{{ i + 1 }}</text>
          <text class="rank-name">{{ s.service_name }}</text>
          <text class="rank-count">{{ s.count }}次</text>
          <text class="rank-revenue">¥{{ s.revenue.toFixed(2) }}</text>
        </view>
        <view class="rank-total" v-if="serviceRanking.length > 0">
          <text class="rank-name">合计</text>
          <text class="rank-count">{{ serviceRanking.reduce((s, r) => s + r.count, 0) }}次</text>
          <text class="rank-revenue">¥{{ serviceRanking.reduce((s, r) => s + r.revenue, 0).toFixed(2) }}</text>
        </view>
        <view v-if="serviceRanking.length === 0" class="empty-sm">暂无数据</view>
      </view>
    </view>

    <!-- Staff performance + commission (员工提成) -->
    <view class="section">
      <text class="section-title">员工业绩与提成</text>
      <view class="rank-list">
        <view class="rank-item header" v-if="staffPerformance.length > 0">
          <text class="rank-no">#</text>
          <text class="rank-name">技师</text>
          <text class="rank-count">接单</text>
          <text class="rank-revenue">营业额</text>
          <text class="rank-commission">提成</text>
        </view>
        <view class="rank-item" v-for="(s, i) in staffPerformance" :key="s.staff_name">
          <text class="rank-no">{{ i + 1 }}</text>
          <text class="rank-name">{{ s.staff_name }}</text>
          <text class="rank-count">{{ s.appointment_count }}单</text>
          <text class="rank-revenue">¥{{ s.revenue.toFixed(2) }}</text>
          <text class="rank-commission">¥{{ (s.commission || 0).toFixed(2) }}</text>
        </view>
        <view class="rank-total" v-if="staffPerformance.length > 0">
          <text class="rank-name">合计</text>
          <text class="rank-count">{{ staffPerformance.reduce((s, r) => s + r.appointment_count, 0) }}单</text>
          <text class="rank-revenue">¥{{ staffPerformance.reduce((s, r) => s + r.revenue, 0).toFixed(2) }}</text>
          <text class="rank-commission">¥{{ staffPerformance.reduce((s, r) => s + (r.commission || 0), 0).toFixed(2) }}</text>
        </view>
        <view v-if="staffPerformance.length === 0" class="empty-sm">暂无数据</view>
      </view>
    </view>

    <!-- Category stats -->
    <view class="section">
      <text class="section-title">分类明细</text>
      <view class="rank-list">
        <view class="rank-item header" v-if="categoryStats.length > 0">
          <text class="rank-name">项目</text>
          <text class="rank-fur">毛发等级</text>
          <text class="rank-count">次数</text>
          <text class="rank-revenue">营收</text>
        </view>
        <view class="rank-item" v-for="(s, i) in categoryStats" :key="i">
          <text class="rank-name">{{ s.service_name }}</text>
          <text class="rank-fur">{{ s.fur_level || '-' }}</text>
          <text class="rank-count">{{ s.count }}次</text>
          <text class="rank-revenue">¥{{ s.revenue.toFixed(2) }}</text>
        </view>
        <view v-if="categoryStats.length === 0" class="empty-sm">暂无数据</view>
      </view>
    </view>
  </view>
  </SideLayout>
</template>

<script setup lang="ts">
import SideLayout from '@/components/SideLayout.vue'
import { ref, onMounted, computed } from 'vue'
import { getDashboardOverview, getRevenueChart, getServiceRanking, getStaffPerformance, getCategoryStats } from '@/api/dashboard'

function localDateStr(d: Date = new Date()): string {
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`
}

const today = localDateStr()
const startDate = ref(today)
const endDate = ref(today)
const quickActive = ref('today')

function setQuickDate(type: string) {
  quickActive.value = type
  const now = new Date()
  if (type === 'today') {
    startDate.value = localDateStr()
    endDate.value = localDateStr()
  } else if (type === 'week') {
    const day = now.getDay() || 7
    const monday = new Date(now)
    monday.setDate(now.getDate() - day + 1)
    startDate.value = localDateStr(monday)
    endDate.value = localDateStr()
  } else if (type === 'month') {
    startDate.value = `${now.getFullYear()}-${String(now.getMonth() + 1).padStart(2, '0')}-01`
    endDate.value = localDateStr()
  } else if (type === 'lastMonth') {
    const last = new Date(now.getFullYear(), now.getMonth(), 0)
    startDate.value = `${last.getFullYear()}-${String(last.getMonth() + 1).padStart(2, '0')}-01`
    endDate.value = localDateStr(last)
  }
  loadAll()
}

const overview = ref({
  today_revenue: 0, today_order_count: 0, today_appointment_count: 0,
  today_new_customers: 0, pending_appointments: 0, total_customers: 0,
})
const revenueData = ref<any[]>([])
const serviceRanking = ref<any[]>([])
const staffPerformance = ref<any[]>([])
const categoryStats = ref<any[]>([])

const maxRevenue = computed(() => Math.max(...revenueData.value.map(d => d.revenue), 1))
function getBarHeight(revenue: number) { return Math.max((revenue / maxRevenue.value) * 200, 8) }

async function loadAll() {
  quickActive.value = '' // clear quick active when manually picking dates
  const sd = startDate.value
  const ed = endDate.value
  const [oRes, rRes, sRes, pRes, cRes] = await Promise.all([
    getDashboardOverview(sd, ed),
    getRevenueChart(sd, ed),
    getServiceRanking(sd, ed),
    getStaffPerformance(sd, ed),
    getCategoryStats(sd, ed),
  ])
  overview.value = oRes.data
  revenueData.value = rRes.data || []
  serviceRanking.value = sRes.data || []
  staffPerformance.value = pRes.data || []
  categoryStats.value = cRes.data || []
}

onMounted(() => {
  quickActive.value = 'today'
  loadAll()
})
</script>

<style scoped>
.page { padding: 24rpx; }
.header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 16rpx; }
.title { font-size: 36rpx; font-weight: bold; color: #1F2937; }
.date-range { display: flex; align-items: center; gap: 8rpx; }
.date-btn { font-size: 24rpx; padding: 8rpx 16rpx; background: #fff; border-radius: 8rpx; color: #374151; border: 1rpx solid #E5E7EB; }
.sep { font-size: 24rpx; color: #6B7280; }

.quick-dates { display: flex; gap: 12rpx; margin-bottom: 20rpx; }
.qd-btn { font-size: 24rpx; padding: 10rpx 24rpx; border-radius: 20rpx; background: #F3F4F6; color: #6B7280; }
.qd-btn.active { background: #4F46E5; color: #fff; }

.overview-grid { display: flex; flex-wrap: wrap; gap: 16rpx; margin-bottom: 24rpx; }
.stat-card { width: calc(33.33% - 12rpx); background: #fff; border-radius: 16rpx; padding: 24rpx 16rpx; text-align: center; box-shadow: 0 2rpx 8rpx rgba(0,0,0,0.04); }
.stat-card.primary { background: linear-gradient(135deg, #4F46E5, #7C3AED); }
.stat-card.primary .stat-val, .stat-card.primary .stat-label { color: #fff; }
.stat-val { font-size: 32rpx; font-weight: bold; color: #4F46E5; display: block; }
.stat-label { font-size: 22rpx; color: #6B7280; display: block; margin-top: 4rpx; }
.section { background: #fff; border-radius: 16rpx; padding: 24rpx; margin-bottom: 16rpx; }
.section-title { font-size: 28rpx; font-weight: 600; color: #1F2937; display: block; margin-bottom: 16rpx; }
.chart-placeholder { text-align: center; padding: 40rpx; color: #9CA3AF; font-size: 26rpx; }
.bar-chart { display: flex; align-items: flex-end; gap: 12rpx; height: 280rpx; padding-top: 40rpx; overflow-x: auto; }
.bar-item { flex: 1; min-width: 60rpx; display: flex; flex-direction: column; align-items: center; }
.bar { width: 100%; background: #4F46E5; border-radius: 8rpx 8rpx 0 0; min-height: 8rpx; }
.bar-label { font-size: 20rpx; color: #6B7280; margin-top: 8rpx; white-space: nowrap; }
.bar-val { font-size: 18rpx; color: #4F46E5; white-space: nowrap; }
.rank-list { }
.rank-item { display: flex; align-items: center; padding: 14rpx 0; border-bottom: 1rpx solid #F3F4F6; font-size: 26rpx; }
.rank-item:last-child { border-bottom: none; }
.rank-item.header { font-weight: 600; color: #374151; border-bottom: 2rpx solid #E5E7EB; font-size: 24rpx; }
.rank-total { display: flex; align-items: center; padding: 14rpx 0; font-size: 26rpx; font-weight: 600; color: #4F46E5; border-top: 2rpx solid #E5E7EB; margin-top: 4rpx; }
.rank-no { width: 48rpx; font-weight: bold; color: #4F46E5; }
.rank-name { flex: 1; color: #1F2937; }
.rank-count { width: 100rpx; color: #6B7280; text-align: right; }
.rank-revenue { width: 140rpx; color: #4F46E5; font-weight: 600; text-align: right; }
.rank-commission { width: 140rpx; color: #059669; font-weight: 600; text-align: right; }
.rank-fur { width: 100rpx; color: #6B7280; text-align: center; font-size: 24rpx; }
.empty-sm { text-align: center; padding: 24rpx; color: #9CA3AF; font-size: 26rpx; }
</style>
