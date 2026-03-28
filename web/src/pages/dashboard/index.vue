<template>
  <SideLayout>
  <view class="page">
    <view class="header">
      <text class="title">数据看板</text>
      <view class="date-range">
        <picker mode="date" :value="startDate" @change="(e:any) => { startDate = e.detail.value; quickActive = ''; loadAll() }">
          <text class="date-btn">{{ startDate }}</text>
        </picker>
        <text class="sep">至</text>
        <picker mode="date" :value="endDate" @change="(e:any) => { endDate = e.detail.value; quickActive = ''; loadAll() }">
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

    <!-- Member stats -->
    <view class="section">
      <text class="section-title">会员概览</text>
      <view class="member-grid">
        <view class="member-card">
          <text class="member-val">{{ memberStats.active_members }}</text>
          <text class="member-label">有效会员</text>
        </view>
        <view class="member-card">
          <text class="member-val">{{ memberStats.frozen_members }}</text>
          <text class="member-label">冻结会员</text>
        </view>
        <view class="member-card">
          <text class="member-val">¥{{ memberStats.total_balance.toFixed(2) }}</text>
          <text class="member-label">会员总余额</text>
        </view>
        <view class="member-card">
          <text class="member-val">¥{{ memberStats.total_member_spent.toFixed(2) }}</text>
          <text class="member-label">累计会员消费</text>
        </view>
        <view class="member-card">
          <text class="member-val">¥{{ memberStats.range_recharge.toFixed(2) }}</text>
          <text class="member-label">区间充值</text>
        </view>
        <view class="member-card">
          <text class="member-val">¥{{ memberStats.range_consumption.toFixed(2) }}</text>
          <text class="member-label">区间余额消费</text>
        </view>
      </view>

      <view class="rank-list member-rank">
        <view class="rank-item header" v-if="memberStats.template_breakdown.length > 0">
          <text class="rank-name">会员卡类型</text>
          <text class="rank-count">人数</text>
        </view>
        <view class="rank-item" v-for="item in memberStats.template_breakdown" :key="item.template_id">
          <text class="rank-name">{{ item.template_name }}</text>
          <text class="rank-count">{{ item.count }}人</text>
        </view>
        <view v-if="memberStats.template_breakdown.length === 0" class="empty-sm">暂无会员卡数据</view>
      </view>
    </view>

    <!-- Revenue trend - Line Chart -->
    <view class="section">
      <text class="section-title">营收趋势</text>
      <view class="chart-placeholder" v-if="filledData.length === 0">暂无数据</view>
      <view v-else class="line-chart-container">
        <!-- Y axis labels -->
        <view class="y-labels">
          <text class="y-label" v-for="(v, i) in yLabels" :key="i">{{ v }}</text>
        </view>
        <!-- Chart area -->
        <view class="chart-area">
          <view class="chart-canvas" :style="{ height: '300rpx' }">
            <!-- Grid lines -->
            <view class="grid-line" v-for="i in 4" :key="'g'+i" :style="{ bottom: ((i-1) * 25) + '%' }"></view>
            <!-- Bars (subtle background) + Line overlay -->
            <view class="chart-bars">
              <view class="chart-col" v-for="(d, idx) in filledData" :key="d.date">
                <view class="chart-bar-bg" :style="{ height: getBarPct(d.revenue) + '%' }"></view>
                <view class="chart-dot" :style="{ bottom: getBarPct(d.revenue) + '%' }">
                  <view class="dot-inner"></view>
                  <text class="dot-tooltip" v-if="d.revenue > 0">¥{{ d.revenue.toFixed(0) }}</text>
                </view>
              </view>
            </view>
          </view>
          <!-- X axis -->
          <view class="x-labels">
            <text class="x-label" v-for="d in filledData" :key="d.date">{{ d.date.substring(5) }}</text>
          </view>
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
          <text class="rank-name">洗护师</text>
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
import { getDashboardOverview, getRevenueChart, getServiceRanking, getStaffPerformance, getCategoryStats, getMemberStats } from '@/api/dashboard'

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
const memberStats = ref({
  active_members: 0,
  frozen_members: 0,
  total_balance: 0,
  total_member_spent: 0,
  range_recharge: 0,
  range_consumption: 0,
  template_breakdown: [] as { template_id: number; template_name: string; count: number }[],
})

// Fill date gaps so chart has continuous points
function fillDateGaps(data: any[], sd: string, ed: string): any[] {
  const map = new Map(data.map(d => [d.date, d]))
  const result: any[] = []
  const cur = new Date(sd)
  const end = new Date(ed)
  while (cur <= end) {
    const ds = localDateStr(cur)
    result.push(map.get(ds) || { date: ds, revenue: 0, order_count: 0 })
    cur.setDate(cur.getDate() + 1)
  }
  return result
}

const filledData = computed(() => fillDateGaps(revenueData.value, startDate.value, endDate.value))
const maxRevenue = computed(() => Math.max(...filledData.value.map(d => d.revenue), 1))

function getBarPct(revenue: number): number {
  return Math.max((revenue / maxRevenue.value) * 100, 2)
}

const yLabels = computed(() => {
  const max = maxRevenue.value
  return [max, Math.round(max * 0.66), Math.round(max * 0.33), 0].map(v => v > 0 ? '¥' + v : '0')
})

async function loadAll() {
  const sd = startDate.value
  const ed = endDate.value
  const [oRes, rRes, sRes, pRes, cRes, mRes] = await Promise.all([
    getDashboardOverview(sd, ed),
    getRevenueChart(sd, ed),
    getServiceRanking(sd, ed),
    getStaffPerformance(sd, ed),
    getCategoryStats(sd, ed),
    getMemberStats(sd, ed),
  ])
  overview.value = oRes.data
  revenueData.value = rRes.data || []
  serviceRanking.value = sRes.data || []
  staffPerformance.value = pRes.data || []
  categoryStats.value = cRes.data || []
  memberStats.value = mRes.data
}

onMounted(() => {
  setQuickDate('week')
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
.member-grid { display: flex; flex-wrap: wrap; gap: 16rpx; margin-bottom: 20rpx; }
.member-card { width: calc(33.33% - 12rpx); background: linear-gradient(180deg, #F8FAFC 0%, #EEF2FF 100%); border: 1rpx solid #E5E7EB; border-radius: 16rpx; padding: 20rpx 16rpx; text-align: center; }
.member-val { font-size: 30rpx; font-weight: 700; color: #3730A3; display: block; }
.member-label { font-size: 22rpx; color: #6B7280; display: block; margin-top: 6rpx; }
.member-rank { margin-top: 8rpx; }
.chart-placeholder { text-align: center; padding: 40rpx; color: #9CA3AF; font-size: 26rpx; }

/* Line chart */
.line-chart-container { display: flex; gap: 8rpx; overflow-x: auto; }
.y-labels { display: flex; flex-direction: column; justify-content: space-between; width: 80rpx; min-width: 80rpx; padding: 0 0 40rpx; }
.y-label { font-size: 18rpx; color: #9CA3AF; text-align: right; }
.chart-area { flex: 1; min-width: 0; }
.chart-canvas { position: relative; background: #FAFBFC; border-radius: 12rpx; border: 1rpx solid #F3F4F6; overflow: hidden; }
.grid-line { position: absolute; left: 0; right: 0; height: 1rpx; background: #F3F4F6; }
.chart-bars { display: flex; height: 100%; align-items: flex-end; }
.chart-col { flex: 1; display: flex; flex-direction: column; align-items: center; justify-content: flex-end; height: 100%; position: relative; }
.chart-bar-bg { width: 60%; background: linear-gradient(180deg, rgba(79,70,229,0.15), rgba(79,70,229,0.03)); border-radius: 6rpx 6rpx 0 0; min-height: 4rpx; transition: height 0.3s; }
.chart-dot { position: absolute; left: 50%; transform: translateX(-50%); }
.dot-inner { width: 12rpx; height: 12rpx; border-radius: 50%; background: #4F46E5; border: 3rpx solid #fff; box-shadow: 0 2rpx 6rpx rgba(79,70,229,0.3); }
.dot-tooltip { position: absolute; bottom: 20rpx; left: 50%; transform: translateX(-50%); font-size: 18rpx; color: #4F46E5; font-weight: 600; white-space: nowrap; }
.x-labels { display: flex; padding-top: 8rpx; }
.x-label { flex: 1; text-align: center; font-size: 18rpx; color: #9CA3AF; }
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
