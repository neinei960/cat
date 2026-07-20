<template>
  <SideLayout>
    <view class="page">
      <view class="hero">
        <view class="hero-copy">
          <text class="hero-eyebrow">经营分析</text>
          <text class="hero-title">数据看板</text>
          <text class="hero-subtitle">把营收趋势、会员经营、服务表现和员工业绩放进同一张视图里，先看趋势，再看原因。</text>
        </view>
        <view class="hero-tools">
          <view class="date-range">
            <picker mode="date" :value="startDate" @change="(e:any) => { startDate = e.detail.value; quickActive = ''; loadAll() }">
              <text class="date-btn">{{ startDate }}</text>
            </picker>
            <text class="sep">至</text>
            <picker mode="date" :value="endDate" @change="(e:any) => { endDate = e.detail.value; quickActive = ''; loadAll() }">
              <text class="date-btn">{{ endDate }}</text>
            </picker>
          </view>
          <view class="quick-dates">
            <text :class="['qd-btn', quickActive === 'today' ? 'active' : '']" @click="setQuickDate('today')">今日</text>
            <text :class="['qd-btn', quickActive === 'week' ? 'active' : '']" @click="setQuickDate('week')">本周</text>
            <text :class="['qd-btn', quickActive === 'month' ? 'active' : '']" @click="setQuickDate('month')">本月</text>
            <text :class="['qd-btn', quickActive === 'lastMonth' ? 'active' : '']" @click="setQuickDate('lastMonth')">上月</text>
          </view>
        </view>
      </view>

      <view class="kpi-grid">
        <view class="kpi-card kpi-primary">
          <text class="kpi-label">区间实收总额</text>
          <text class="kpi-value">¥{{ overview.month_collection.toFixed(2) }}</text>
          <text class="kpi-foot">订单收款 + 会员充值</text>
        </view>
        <view class="kpi-card">
          <text class="kpi-label">订单营业额</text>
          <text class="kpi-value">¥{{ overview.month_revenue.toFixed(2) }}</text>
          <text class="kpi-foot">当前时间区间已支付订单</text>
        </view>
        <view class="kpi-card">
          <text class="kpi-label">会员充值</text>
          <text class="kpi-value">¥{{ overview.month_recharge.toFixed(2) }}</text>
          <text class="kpi-foot">当前时间区间预充值</text>
        </view>
        <view class="kpi-card">
          <text class="kpi-label">客单价</text>
          <text class="kpi-value">¥{{ (overview.avg_order_value || 0).toFixed(0) }}</text>
          <text class="kpi-foot">平均每笔订单收入</text>
        </view>
        <view class="kpi-card">
          <text class="kpi-label">已支付订单</text>
          <text class="kpi-value">{{ overview.today_order_count }}</text>
          <text class="kpi-foot">当前时间区间已支付订单数</text>
        </view>
        <view class="kpi-card" :class="{ warn: (overview.no_show_rate || 0) > 0.1 }">
          <text class="kpi-label">爽约率</text>
          <text class="kpi-value">{{ ((overview.no_show_rate || 0) * 100).toFixed(1) }}%</text>
          <text class="kpi-foot">{{ overview.no_show_count || 0 }}/{{ overview.total_appointments || 0 }} 次预约未到店</text>
        </view>
      </view>

      <view class="section">
        <view class="section-header">
          <view>
            <text class="section-title">付款渠道金额</text>
            <text class="section-subtitle">按当前时间范围内已支付订单统计各收款渠道金额和占比。</text>
          </view>
        </view>
        <view class="payment-channel-list" v-if="paymentChannelRows.length > 0">
          <view class="payment-channel-row header">
            <text class="payment-channel-name">渠道</text>
            <text class="payment-channel-amount">金额</text>
            <text class="payment-channel-percent">占比</text>
          </view>
          <view class="payment-channel-row" v-for="item in paymentChannelRows" :key="item.key">
            <view class="payment-channel-name">
              <text class="payment-dot" :class="`payment-dot-${item.key}`"></text>
              <text>{{ item.label }}</text>
            </view>
            <text class="payment-channel-amount">¥{{ item.amount.toFixed(2) }}</text>
            <view class="payment-channel-percent">
              <text>{{ item.percent.toFixed(1) }}%</text>
              <view class="payment-bar">
                <view class="payment-bar-fill" :style="{ width: `${item.percent}%` }"></view>
              </view>
            </view>
          </view>
          <view class="payment-channel-total">
            <text>合计</text>
            <text>¥{{ paymentChannelTotal.toFixed(2) }}</text>
            <text>100%</text>
          </view>
        </view>
        <view v-else class="empty-sm">暂无付款渠道数据</view>
      </view>

      <view class="main-grid">
        <view class="section section-chart">
          <view class="section-header">
            <view>
              <text class="section-title">营收趋势</text>
              <text class="section-subtitle">先看趋势，再回头看项目、员工和会员结构。</text>
            </view>
          </view>

          <view class="chart-placeholder" v-if="filledData.length === 0">暂无数据</view>
          <view v-else class="line-chart-container">
            <view class="chart-summary">
              <view>
                <text class="chart-summary-label">{{ selectedRevenueItem.date }}</text>
                <text class="chart-summary-value">¥{{ selectedRevenueItem.revenue.toFixed(2) }}</text>
              </view>
              <text class="chart-summary-tip">点选柱子查看单日金额</text>
            </view>
            <view class="y-labels">
              <text class="y-label" v-for="(v, i) in yLabels" :key="i">{{ v }}</text>
            </view>
            <view class="chart-scroll">
              <view class="chart-area" :style="{ minWidth: chartMinWidth }">
                <view class="chart-canvas" :style="{ height: '320rpx' }">
                  <view class="grid-line" v-for="i in 4" :key="'g'+i" :style="{ bottom: ((i-1) * 25) + '%' }"></view>
                  <view class="chart-bars">
                    <view
                      v-for="(d, i) in filledData"
                      :key="d.date"
                      :class="['chart-col', activeRevenueIndex === i ? 'active' : '', maxRevenueIndex === i ? 'max' : '']"
                      @click="selectRevenueIndex(i)"
                      @tap="selectRevenueIndex(i)"
                    >
                      <view class="chart-bar-bg" :style="{ height: getBarPct(d.revenue) + '%' }"></view>
                      <view class="chart-dot" :style="{ bottom: getBarPct(d.revenue) + '%' }">
                        <view class="dot-inner"></view>
                        <text class="dot-tooltip" v-if="shouldShowRevenueLabel(i)">¥{{ d.revenue.toFixed(2) }}</text>
                      </view>
                    </view>
                  </view>
                </view>
                <view class="x-labels">
                  <text class="x-label" v-for="(d, i) in filledData" :key="d.date">{{ showXAxisLabel(i) ? d.date.substring(5) : '' }}</text>
                </view>
              </view>
            </view>
          </view>
        </view>

        <view class="section section-side">
          <view class="section-header">
            <view>
              <text class="section-title">经营摘要</text>
              <text class="section-subtitle">按当前月份统计客户和预约结构。</text>
            </view>
          </view>
          <view class="summary-card">
            <view class="summary-row">
              <text class="summary-label">预约数</text>
              <text class="summary-value">{{ overview.today_appointment_count }}</text>
            </view>
            <view class="summary-row">
              <text class="summary-label">新客</text>
              <text class="summary-value">{{ overview.today_new_customers }}</text>
            </view>
            <view class="summary-row">
              <text class="summary-label">老客</text>
              <text class="summary-value">{{ overview.regular_customer_count }}</text>
            </view>
            <view class="summary-row">
              <text class="summary-label">总客户</text>
              <text class="summary-value">{{ overview.total_customers }}</text>
            </view>
          </view>
        </view>
      </view>

      <view class="section">
        <view class="section-header">
          <view>
            <text class="section-title">会员概览</text>
            <text class="section-subtitle">看会员结构和余额体量，判断复购和充值表现。</text>
          </view>
        </view>
        <view class="member-grid">
          <view class="member-card">
            <text class="member-label">有效会员</text>
            <text class="member-val">{{ memberStats.active_members }}</text>
          </view>
          <view class="member-card">
            <text class="member-label">会员总余额</text>
            <text class="member-val">¥{{ memberStats.total_balance.toFixed(2) }}</text>
          </view>
          <view class="member-card">
            <text class="member-label">累计会员消费</text>
            <text class="member-val">¥{{ memberStats.total_member_spent.toFixed(2) }}</text>
          </view>
          <view class="member-card">
            <text class="member-label">区间充值</text>
            <text class="member-val">¥{{ memberStats.range_recharge.toFixed(2) }}</text>
          </view>
          <view class="member-card">
            <text class="member-label">区间余额消费</text>
            <text class="member-val">¥{{ memberStats.range_consumption.toFixed(2) }}</text>
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

      <view class="dual-grid">
        <view class="section">
          <view class="section-header">
            <view>
              <text class="section-title">项目营业额</text>
              <text class="section-subtitle">按服务、商品、上门喂养和寄养汇总，点击大类逐级展开。</text>
            </view>
          </view>
          <view class="rank-list">
            <view class="rank-item header" v-if="visibleProjectRevenueRows.length > 0">
              <text class="rank-no"></text>
              <text class="rank-name">项目名称</text>
              <text class="rank-count">次数</text>
              <text class="rank-revenue">营业额</text>
            </view>
            <view
              v-for="row in visibleProjectRevenueRows"
              :key="row.node.key"
              :class="['rank-item', 'tree-row', row.hasChildren ? 'clickable' : 'leaf']"
              @click="toggleProjectRevenueNode(row.node)"
            >
              <text class="rank-no">{{ row.hasChildren ? (isProjectRevenueExpanded(row.node) ? '⌄' : '›') : '' }}</text>
              <text class="rank-name" :style="{ paddingLeft: `${row.depth * 24}rpx` }">{{ row.node.name }}</text>
              <text class="rank-count">{{ row.node.count }}次</text>
              <text class="rank-revenue">¥{{ row.node.revenue.toFixed(2) }}</text>
            </view>
            <view class="rank-total" v-if="projectRevenueTree.length > 0">
              <text class="rank-name">合计</text>
              <text class="rank-count">{{ projectRevenueTotal.count }}次</text>
              <text class="rank-revenue">¥{{ projectRevenueTotal.revenue.toFixed(2) }}</text>
            </view>
            <view v-if="projectRevenueTree.length === 0" class="empty-sm">暂无数据</view>
          </view>
        </view>

        <view class="section">
          <view class="section-header">
            <view>
              <text class="section-title">员工业绩与提成</text>
              <text class="section-subtitle">按已支付订单统计业绩和提成，避免和待结算混在一起。</text>
            </view>
          </view>
          <view class="rank-list">
            <view class="rank-item header" v-if="staffPerformance.length > 0">
              <text class="rank-no">#</text>
              <text class="rank-name">洗护师</text>
              <text class="rank-count">接单</text>
              <text class="rank-revenue">营业额</text>
              <text class="rank-product">商品额</text>
              <text class="rank-commission">提成</text>
            </view>
            <template v-for="(s, i) in staffPerformance" :key="s.staff_id || s.staff_name">
              <view class="rank-item staff-row clickable" @click="toggleStaffCommissionDetails(s)">
                <text class="rank-no">{{ i + 1 }}</text>
                <view class="rank-name staff-name-cell">
                  <text>{{ s.staff_name }}</text>
                  <text class="expand-hint">{{ isStaffExpanded(s.staff_id) ? '收起' : '展开' }}</text>
                </view>
                <text class="rank-count">{{ s.appointment_count }}单</text>
                <text class="rank-revenue">¥{{ s.revenue.toFixed(2) }}</text>
                <text class="rank-product">¥{{ Number(s.product_revenue || 0).toFixed(2) }}</text>
                <text class="rank-commission">¥{{ (s.commission || 0).toFixed(2) }}</text>
              </view>
              <view v-if="isStaffExpanded(s.staff_id)" class="commission-detail-panel">
                <view v-if="staffDetailLoading[s.staff_id]" class="detail-state">加载中...</view>
                <view v-else-if="(staffCommissionDetails[s.staff_id] || []).length === 0" class="detail-state">暂无提成明细</view>
                <view v-else class="commission-detail-list">
                  <view class="commission-detail-row header">
                    <text class="detail-date">日</text>
                    <text class="detail-method">收款</text>
                    <text class="detail-formula">计算公式</text>
                    <text class="detail-commission">提成</text>
                  </view>
                  <view class="commission-detail-row clickable" v-for="item in staffCommissionDetails[s.staff_id]" :key="item.order_id" @click="goCommissionOrderDetail(item)">
                    <text class="detail-date">{{ formatCommissionDay(item.date) }}</text>
                    <text class="detail-method">{{ item.pay_method_label }}</text>
                    <text class="detail-formula">{{ item.formula }}</text>
                    <text class="detail-commission">¥{{ Number(item.commission || 0).toFixed(2) }}</text>
                  </view>
                </view>
              </view>
            </template>
            <view class="rank-total" v-if="staffPerformance.length > 0">
              <text class="rank-name">合计</text>
              <text class="rank-count">{{ staffPerformance.reduce((s, r) => s + r.appointment_count, 0) }}单</text>
              <text class="rank-revenue">¥{{ staffPerformance.reduce((s, r) => s + r.revenue, 0).toFixed(2) }}</text>
              <text class="rank-product">¥{{ staffPerformance.reduce((s, r) => s + Number(r.product_revenue || 0), 0).toFixed(2) }}</text>
              <text class="rank-commission">¥{{ staffPerformance.reduce((s, r) => s + (r.commission || 0), 0).toFixed(2) }}</text>
            </view>
            <view v-if="staffPerformance.length === 0" class="empty-sm">暂无数据</view>
          </view>
        </view>
      </view>

      <view class="section">
        <view class="section-header">
          <view>
            <text class="section-title">服务分类统计</text>
            <text class="section-subtitle">这里统计的是服务项目在当前时间范围内的汇总数据，不是单只猫的订单明细。</text>
          </view>
        </view>
        <view class="rank-list">
        <view class="rank-item header" v-if="categoryStats.length > 0">
          <text class="rank-name">服务</text>
          <text class="rank-fur">规格</text>
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
import {
  getDashboardOverview,
  getRevenueChart,
  getProjectRevenueTree,
  getStaffPerformance,
  getStaffCommissionDetails,
  getCategoryStats,
  getMemberStats,
  type ProjectRevenueNode,
  type StaffCommissionDetail,
} from '@/api/dashboard'

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
  today_revenue: 0, month_revenue: 0, month_recharge: 0, month_collection: 0,
  today_order_count: 0, today_appointment_count: 0,
  today_service_completed_count: 0, today_pending_settlement_count: 0, today_refunded_order_count: 0,
  today_new_customers: 0, regular_customer_count: 0, pending_appointments: 0, total_customers: 0,
  avg_order_value: 0, no_show_rate: 0, no_show_count: 0, total_appointments: 0,
  payment_breakdown: [] as { key: string; label: string; amount: number }[],
})
const revenueData = ref<any[]>([])
const projectRevenueTree = ref<ProjectRevenueNode[]>([])
const expandedProjectRevenue = ref<Record<string, boolean>>({})
const staffPerformance = ref<any[]>([])
const expandedStaffId = ref<number | null>(null)
const staffDetailLoading = ref<Record<number, boolean>>({})
const staffCommissionDetails = ref<Record<number, StaffCommissionDetail[]>>({})
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
let loadSeq = 0

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
const activeRevenueIndex = ref(-1)
const maxRevenueIndex = computed(() => {
  let index = 0
  let max = -1
  filledData.value.forEach((item, i) => {
    if (Number(item.revenue || 0) > max) {
      max = Number(item.revenue || 0)
      index = i
    }
  })
  return index
})
const selectedRevenueItem = computed(() => {
  const active = filledData.value[activeRevenueIndex.value]
  return active || filledData.value[maxRevenueIndex.value] || { date: '-', revenue: 0, order_count: 0 }
})
const chartMinWidth = computed(() => {
  if (filledData.value.length <= 12) return '100%'
  return `${filledData.value.length * 54}rpx`
})

function getBarPct(revenue: number): number {
  return Math.max((revenue / maxRevenue.value) * 100, 2)
}

function shouldShowRevenueLabel(index: number): boolean {
  const item = filledData.value[index]
  if (!item || Number(item.revenue || 0) <= 0) return false
  if (activeRevenueIndex.value >= 0) return activeRevenueIndex.value === index
  return maxRevenueIndex.value === index
}

function showXAxisLabel(index: number): boolean {
  const length = filledData.value.length
  if (length <= 12) return true
  const interval = length <= 20 ? 2 : 5
  return index === 0 || index === length - 1 || index === activeRevenueIndex.value || index % interval === 0
}

function selectRevenueIndex(index: number) {
  activeRevenueIndex.value = index
}

const yLabels = computed(() => {
  const max = maxRevenue.value
  return [max, Math.round(max * 0.66), Math.round(max * 0.33), 0].map(v => v > 0 ? '¥' + v : '0')
})

function hasProjectRevenueChildren(node: ProjectRevenueNode) {
  return (node.children || []).length > 0
}

function isProjectRevenueExpanded(node: ProjectRevenueNode) {
  return !!expandedProjectRevenue.value[node.key]
}

function toggleProjectRevenueNode(node: ProjectRevenueNode) {
  if (!hasProjectRevenueChildren(node)) return
  expandedProjectRevenue.value = {
    ...expandedProjectRevenue.value,
    [node.key]: !expandedProjectRevenue.value[node.key],
  }
}

const visibleProjectRevenueRows = computed(() => {
  const rows: Array<{ node: ProjectRevenueNode; depth: number; hasChildren: boolean }> = []
  const walk = (nodes: ProjectRevenueNode[], depth: number) => {
    nodes.forEach((node) => {
      const hasChildren = hasProjectRevenueChildren(node)
      rows.push({ node, depth, hasChildren })
      if (hasChildren && isProjectRevenueExpanded(node)) {
        walk(node.children || [], depth + 1)
      }
    })
  }
  walk(projectRevenueTree.value, 0)
  return rows
})

const projectRevenueTotal = computed(() => projectRevenueTree.value.reduce((total, node) => ({
  count: total.count + Number(node.count || 0),
  revenue: total.revenue + Number(node.revenue || 0),
}), { count: 0, revenue: 0 }))

const paymentChannelTotal = computed(() => (overview.value.payment_breakdown || []).reduce((sum, item) => sum + Number(item.amount || 0), 0))
const paymentChannelRows = computed(() => {
  const total = paymentChannelTotal.value
  return (overview.value.payment_breakdown || [])
    .filter(item => Number(item.amount || 0) > 0)
    .map(item => ({
      ...item,
      amount: Number(item.amount || 0),
      percent: total > 0 ? Number(item.amount || 0) / total * 100 : 0,
    }))
})

function isStaffExpanded(staffId: number) {
  return expandedStaffId.value === staffId
}

function formatCommissionDay(date: string) {
  const parts = String(date || '').split('-')
  return parts.length >= 3 ? parts[2] : date
}

function goCommissionOrderDetail(item: StaffCommissionDetail) {
  const orderId = Number(item?.order_id || 0)
  if (!orderId) {
    uni.showToast({ title: '订单信息缺失', icon: 'none' })
    return
  }
  uni.navigateTo({ url: `/pages/order/detail?id=${orderId}` })
}

async function toggleStaffCommissionDetails(staff: any) {
  const staffId = Number(staff?.staff_id || 0)
  if (!staffId) return
  if (expandedStaffId.value === staffId) {
    expandedStaffId.value = null
    return
  }
  expandedStaffId.value = staffId
  if (staffCommissionDetails.value[staffId]) {
    return
  }

  staffDetailLoading.value = { ...staffDetailLoading.value, [staffId]: true }
  try {
    const res = await getStaffCommissionDetails(staffId, startDate.value, endDate.value)
    staffCommissionDetails.value = {
      ...staffCommissionDetails.value,
      [staffId]: res.data || [],
    }
  } finally {
    staffDetailLoading.value = { ...staffDetailLoading.value, [staffId]: false }
  }
}

async function loadAll() {
  const seq = ++loadSeq
  const sd = startDate.value
  const ed = endDate.value
  const [oRes, rRes, sRes, pRes, cRes, mRes] = await Promise.all([
    getDashboardOverview(sd, ed),
    getRevenueChart(sd, ed),
    getProjectRevenueTree(sd, ed),
    getStaffPerformance(sd, ed),
    getCategoryStats(sd, ed),
    getMemberStats(sd, ed),
  ])
  if (seq !== loadSeq || sd !== startDate.value || ed !== endDate.value) {
    return
  }
  overview.value = oRes.data
  revenueData.value = rRes.data || []
  activeRevenueIndex.value = -1
  projectRevenueTree.value = sRes.data || []
  expandedProjectRevenue.value = {}
  staffPerformance.value = pRes.data || []
  expandedStaffId.value = null
  staffCommissionDetails.value = {}
  staffDetailLoading.value = {}
  categoryStats.value = cRes.data || []
  memberStats.value = mRes.data
}

onMounted(() => {
  setQuickDate('month')
})
</script>

<style scoped>
.page {
  padding: 24rpx;
  background:
    radial-gradient(circle at top right, rgba(249, 115, 22, 0.08), transparent 18%),
    linear-gradient(180deg, #FFF8F1 0%, #F7F8FC 18%, #F5F6FA 100%);
}

.hero {
  display: grid;
  grid-template-columns: 1.1fr 0.9fr;
  gap: 18rpx;
  margin-bottom: 22rpx;
}

.hero-copy,
.hero-tools {
  border-radius: 28rpx;
  padding: 28rpx;
}

.hero-copy {
  background: linear-gradient(145deg, #111827, #1F2937);
  color: #FFFFFF;
  box-shadow: 0 16rpx 42rpx rgba(15, 23, 42, 0.16);
}

.hero-eyebrow {
  display: block;
  color: #FDBA74;
  font-size: 22rpx;
  letter-spacing: 2rpx;
}

.hero-title {
  display: block;
  margin-top: 14rpx;
  font-size: 46rpx;
  font-weight: 800;
}

.hero-subtitle {
  display: block;
  margin-top: 14rpx;
  color: rgba(255, 255, 255, 0.72);
  font-size: 24rpx;
  line-height: 1.7;
}

.hero-tools {
  background: rgba(255, 255, 255, 0.92);
  box-shadow: 0 10rpx 30rpx rgba(15, 23, 42, 0.05);
}

.date-range {
  display: flex;
  align-items: center;
  gap: 10rpx;
  margin-bottom: 16rpx;
}

.date-btn {
  font-size: 24rpx;
  padding: 10rpx 18rpx;
  background: #FFF7ED;
  border-radius: 999rpx;
  color: #9A3412;
  border: 1rpx solid #FED7AA;
}

.sep {
  font-size: 24rpx;
  color: #6B7280;
}

.quick-dates {
  display: flex;
  flex-wrap: wrap;
  gap: 12rpx;
}

.qd-btn {
  font-size: 24rpx;
  padding: 10rpx 24rpx;
  border-radius: 999rpx;
  background: #F3F4F6;
  color: #6B7280;
}

.qd-btn.active {
  background: linear-gradient(135deg, #F97316, #EA580C);
  color: #fff;
}

.kpi-grid {
  display: grid;
  grid-template-columns: 1.15fr repeat(3, minmax(0, 1fr));
  gap: 16rpx;
  margin-bottom: 22rpx;
}

.kpi-card {
  background: rgba(255, 255, 255, 0.92);
  border-radius: 24rpx;
  padding: 24rpx;
  box-shadow: 0 10rpx 30rpx rgba(15, 23, 42, 0.05);
}

.kpi-primary {
  background: linear-gradient(145deg, #F97316, #EA580C);
  color: #FFFFFF;
}

.kpi-primary .kpi-label,
.kpi-primary .kpi-foot,
.kpi-primary .kpi-value {
  color: #FFFFFF;
}

.kpi-card.warn {
  background: linear-gradient(180deg, #FFFBEB, #FFFFFF);
  border: 1rpx solid #FDE68A;
}

.kpi-label {
  display: block;
  font-size: 22rpx;
  color: #6B7280;
}

.kpi-value {
  display: block;
  margin-top: 10rpx;
  font-size: 42rpx;
  font-weight: 800;
  color: #111827;
}

.kpi-foot {
  display: block;
  margin-top: 10rpx;
  font-size: 22rpx;
  line-height: 1.6;
  color: #94A3B8;
}

.main-grid,
.dual-grid {
  display: grid;
  grid-template-columns: 1.25fr 0.75fr;
  gap: 18rpx;
  margin-bottom: 22rpx;
}

.dual-grid {
  grid-template-columns: 1fr 1fr;
}

.section {
  background: rgba(255, 255, 255, 0.92);
  border-radius: 24rpx;
  padding: 26rpx;
  margin-bottom: 18rpx;
  box-shadow: 0 10rpx 30rpx rgba(15, 23, 42, 0.05);
}

.section-chart {
  background: linear-gradient(180deg, #FFFFFF, #FFF8F1);
}

.section-side {
  background: linear-gradient(180deg, #FFF7ED, #FFFFFF);
}

.section-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 18rpx;
  margin-bottom: 18rpx;
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

.member-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 12rpx;
  margin-bottom: 20rpx;
}

.member-card {
  background: linear-gradient(180deg, #F8FAFC 0%, #EEF2FF 100%);
  border: 1rpx solid #E5E7EB;
  border-radius: 18rpx;
  padding: 18rpx 10rpx;
  min-width: 0;
  text-align: center;
}

.member-label {
  display: block;
  font-size: 22rpx;
  color: #6B7280;
  text-align: center;
}

.member-val {
  display: block;
  margin-top: 12rpx;
  font-size: 28rpx;
  font-weight: 700;
  color: #3730A3;
  text-align: center;
  word-break: break-all;
}

.member-rank {
  margin-top: 8rpx;
}

.summary-card {
  display: flex;
  flex-direction: column;
  gap: 12rpx;
}

.summary-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 18rpx;
  border-radius: 18rpx;
  background: rgba(255, 255, 255, 0.84);
  border: 1rpx solid #F1F5F9;
}

.summary-label {
  color: #6B7280;
  font-size: 24rpx;
}

.summary-value {
  color: #111827;
  font-size: 32rpx;
  font-weight: 800;
}

.payment-channel-list {
  display: flex;
  flex-direction: column;
}

.payment-channel-row,
.payment-channel-total {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 180rpx 180rpx;
  align-items: center;
  gap: 16rpx;
  padding: 16rpx 0;
  border-bottom: 1rpx solid #F3F4F6;
  font-size: 26rpx;
}

.payment-channel-row.header {
  color: #374151;
  font-size: 24rpx;
  font-weight: 700;
  border-bottom: 2rpx solid #E5E7EB;
}

.payment-channel-name {
  display: flex;
  align-items: center;
  gap: 12rpx;
  min-width: 0;
  color: #111827;
}

.payment-channel-amount {
  color: #C2410C;
  font-weight: 700;
  text-align: right;
}

.payment-channel-percent {
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  gap: 8rpx;
  color: #6B7280;
  font-size: 24rpx;
}

.payment-dot {
  width: 16rpx;
  height: 16rpx;
  border-radius: 50%;
  background: #9CA3AF;
  flex-shrink: 0;
}

.payment-dot-wechat {
  background: #22C55E;
}

.payment-dot-qrcode {
  background: #4F46E5;
}

.payment-dot-meituan {
  background: #F97316;
}

.payment-dot-balance {
  background: #0EA5E9;
}

.payment-bar {
  width: 120rpx;
  height: 8rpx;
  border-radius: 999rpx;
  overflow: hidden;
  background: #E5E7EB;
}

.payment-bar-fill {
  height: 100%;
  min-width: 4rpx;
  border-radius: 999rpx;
  background: linear-gradient(90deg, #F97316, #4F46E5);
}

.payment-channel-total {
  border-bottom: none;
  margin-top: 4rpx;
  color: #EA580C;
  font-weight: 700;
}

.payment-channel-total text:nth-child(2),
.payment-channel-total text:nth-child(3) {
  text-align: right;
}

.chart-placeholder {
  text-align: center;
  padding: 40rpx;
  color: #9CA3AF;
  font-size: 26rpx;
}

.line-chart-container {
  display: grid;
  grid-template-columns: 84rpx minmax(0, 1fr);
  gap: 8rpx;
}

.chart-summary {
  grid-column: 1 / -1;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16rpx;
  padding: 14rpx 18rpx;
  border-radius: 18rpx;
  background: rgba(255, 247, 237, 0.82);
  border: 1rpx solid #FED7AA;
}

.chart-summary-label {
  display: block;
  font-size: 22rpx;
  color: #9A3412;
}

.chart-summary-value {
  display: block;
  margin-top: 4rpx;
  font-size: 34rpx;
  font-weight: 800;
  color: #C2410C;
}

.chart-summary-tip {
  flex-shrink: 0;
  font-size: 22rpx;
  color: #9CA3AF;
}

.y-labels {
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  width: 84rpx;
  min-width: 84rpx;
  padding: 0 0 44rpx;
}

.y-label {
  font-size: 18rpx;
  color: #9CA3AF;
  text-align: right;
}

.chart-area {
  min-width: 100%;
}

.chart-scroll {
  width: 100%;
  min-width: 0;
  overflow-x: auto;
  overflow-y: hidden;
}

.chart-canvas {
  position: relative;
  background: rgba(255, 255, 255, 0.78);
  border-radius: 18rpx;
  border: 1rpx solid #F3F4F6;
  overflow: visible;
  padding-top: 38rpx;
}

.grid-line {
  position: absolute;
  left: 0;
  right: 0;
  height: 1rpx;
  background: #F3F4F6;
}

.chart-bars {
  display: flex;
  height: 100%;
  align-items: flex-end;
}

.chart-col {
  flex: 1 0 48rpx;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: flex-end;
  height: 100%;
  position: relative;
}

.chart-bar-bg {
  width: 44%;
  background: linear-gradient(180deg, rgba(249, 115, 22, 0.3), rgba(249, 115, 22, 0.06));
  border-radius: 10rpx 10rpx 0 0;
  min-height: 4rpx;
  transition: height 0.3s;
}

.chart-col.active .chart-bar-bg,
.chart-col.max .chart-bar-bg {
  background: linear-gradient(180deg, rgba(234, 88, 12, 0.72), rgba(249, 115, 22, 0.14));
}

.chart-dot {
  position: absolute;
  left: 50%;
  transform: translateX(-50%);
}

.dot-inner {
  width: 14rpx;
  height: 14rpx;
  border-radius: 50%;
  background: #EA580C;
  border: 4rpx solid #fff;
  box-shadow: 0 2rpx 6rpx rgba(234, 88, 12, 0.24);
}

.dot-tooltip {
  position: absolute;
  bottom: 26rpx;
  left: 50%;
  transform: translateX(-50%);
  padding: 6rpx 10rpx;
  border-radius: 999rpx;
  background: #FFFFFF;
  border: 1rpx solid #FED7AA;
  box-shadow: 0 8rpx 20rpx rgba(124, 45, 18, 0.12);
  font-size: 20rpx;
  color: #9A3412;
  font-weight: 700;
  white-space: nowrap;
  z-index: 3;
}

.x-labels {
  display: flex;
  padding-top: 10rpx;
}

.x-label {
  flex: 1;
  text-align: center;
  font-size: 18rpx;
  color: #9CA3AF;
}

.rank-item {
  display: flex;
  align-items: center;
  padding: 14rpx 0;
  border-bottom: 1rpx solid #F3F4F6;
  font-size: 26rpx;
}

.rank-item:last-child {
  border-bottom: none;
}

.rank-item.header {
  font-weight: 600;
  color: #374151;
  border-bottom: 2rpx solid #E5E7EB;
  font-size: 24rpx;
}

.tree-row.clickable {
  cursor: pointer;
}

.staff-row.clickable {
  cursor: pointer;
}

.staff-row.clickable:active {
  background: #F9FAFB;
}

.tree-row.clickable .rank-name {
  font-weight: 600;
}

.tree-row.leaf .rank-name {
  color: #374151;
}

.rank-total {
  display: flex;
  align-items: center;
  padding: 14rpx 0;
  font-size: 26rpx;
  font-weight: 600;
  color: #EA580C;
  border-top: 2rpx solid #E5E7EB;
  margin-top: 4rpx;
}

.rank-no {
  width: 48rpx;
  font-weight: bold;
  color: #EA580C;
}

.rank-name {
  flex: 1;
  color: #111827;
}

.staff-name-cell {
  display: flex;
  align-items: center;
  gap: 12rpx;
  min-width: 0;
}

.expand-hint {
  padding: 2rpx 10rpx;
  border-radius: 999rpx;
  background: #EEF2FF;
  color: #4F46E5;
  font-size: 20rpx;
  font-weight: 600;
}

.rank-count {
  width: 100rpx;
  color: #6B7280;
  text-align: right;
}

.rank-revenue {
  width: 140rpx;
  color: #C2410C;
  font-weight: 600;
  text-align: right;
}

.rank-product {
  width: 140rpx;
  color: #2563EB;
  font-weight: 600;
  text-align: right;
}

.rank-commission {
  width: 140rpx;
  color: #059669;
  font-weight: 600;
  text-align: right;
}

.rank-fur {
  width: 100rpx;
  color: #6B7280;
  text-align: center;
  font-size: 24rpx;
}

.commission-detail-panel {
  margin: 0 0 10rpx 48rpx;
  padding: 12rpx 14rpx;
  border-radius: 16rpx;
  background: #F8FAFC;
  border: 1rpx solid #E5E7EB;
  overflow: hidden;
}

.detail-state {
  padding: 18rpx;
  color: #6B7280;
  font-size: 24rpx;
  text-align: center;
}

.commission-detail-list {
  width: 100%;
}

.commission-detail-row {
  display: grid;
  grid-template-columns: 44rpx 76rpx minmax(0, 1fr) 112rpx;
  align-items: center;
  column-gap: 10rpx;
  padding: 10rpx 0;
  border-bottom: 1rpx solid #E5E7EB;
  font-size: 21rpx;
  color: #374151;
}

.commission-detail-row:last-child {
  border-bottom: none;
}

.commission-detail-row.clickable {
  cursor: pointer;
}

.commission-detail-row.clickable:active {
  background: #EEF2FF;
}

.commission-detail-row.header {
  color: #6B7280;
  font-weight: 700;
}

.detail-date {
  color: #6B7280;
  font-weight: 600;
}

.detail-method {
  color: #6B7280;
  white-space: nowrap;
}

.detail-formula {
  color: #4B5563;
  line-height: 1.35;
  min-width: 0;
  word-break: break-all;
}

.detail-commission {
  text-align: right;
  color: #059669;
  font-weight: 700;
  white-space: nowrap;
}

.empty-sm {
  text-align: center;
  padding: 24rpx;
  color: #9CA3AF;
  font-size: 26rpx;
}

@media (max-width: 900px) {
  .hero,
  .main-grid,
  .dual-grid {
    grid-template-columns: 1fr;
  }

  .kpi-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: 12rpx;
  }

  .kpi-card {
    padding: 20rpx;
    min-width: 0;
  }

  .kpi-value {
    font-size: 34rpx;
    word-break: break-all;
  }

  .kpi-foot {
    font-size: 20rpx;
  }
}
</style>
