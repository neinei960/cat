<template>
  <SideLayout>
    <view class="page" v-if="plan">
      <view class="hero-card">
      <view>
          <text class="hero-title">喂养计划 #{{ plan.ID }}</text>
          <text class="hero-subtitle">{{ feedingStatusLabel(plan.status) }} · {{ formatFeedingDateRange(plan.start_date, plan.end_date) }}</text>
        </view>
        <view class="hero-actions">
          <view v-if="canOperate" class="btn" @click="goEdit">修改</view>
          <view v-if="canOperate && plan.status === 'active'" class="btn" @click="pausePlan">暂停</view>
          <view v-if="canOperate && plan.status === 'paused'" class="btn" @click="resumePlan">恢复</view>
          <view v-if="canOperate && plan.status !== 'cancelled' && plan.status !== 'completed'" class="btn danger" @click="cancelPlanAction">取消</view>
          <view v-if="canOperate && !plan.order_id" class="btn btn-primary" @click="generateOrderAction">生成订单</view>
          <view v-else-if="plan.order_id" class="btn btn-primary" @click="goOrder">查看订单</view>
        </view>
      </view>

      <view class="section-card compact">
        <view class="info-row">
          <text class="info-label">客户</text>
          <text class="info-val">{{ plan.customer?.nickname || plan.customer?.phone || '-' }}</text>
        </view>
        <view class="info-row">
          <text class="info-label">联系人</text>
          <text class="info-val">{{ plan.contact_name || '-' }}</text>
          <text class="info-label" style="margin-left: 24rpx;">电话</text>
          <text class="info-val">{{ plan.contact_phone || '-' }}</text>
        </view>
        <view class="info-row">
          <text class="info-label">地址</text>
          <text class="info-val">{{ address.address || '-' }}</text>
        </view>
        <view v-if="address.door_code" class="info-row">
          <text class="info-label">入户</text>
          <text class="info-val">{{ address.door_code }}</text>
        </view>
        <view class="info-row">
          <text class="info-label">猫咪</text>
          <view class="chip-list inline">
            <view class="chip sm" v-for="item in plan.pets || []" :key="item.ID || item.pet_id">{{ item.pet?.name || item.pet_name_snapshot }}</view>
          </view>
        </view>
        <view class="info-row">
          <text class="info-label">服务</text>
          <view class="chip-list inline">
            <view class="chip sm soft" v-for="item in selectedItems" :key="item.code">{{ item.name }}</view>
          </view>
        </view>
        <view class="info-row">
          <text class="info-label">预计</text>
          <text class="info-val amount">¥{{ plan.total_amount.toFixed(2) }}</text>
        </view>
        <view class="info-row">
          <text class="info-label">定金</text>
          <input class="inline-input" type="digit" v-model="depositInput" placeholder="0" @blur="saveDeposit" />
          <text class="info-label" style="margin-left: 24rpx;">尾款</text>
          <text class="info-val">¥{{ balanceDisplay }}</text>
        </view>
      </view>

      <view class="section-card">
        <view class="section-head">
          <text class="section-title">上门安排</text>
          <view v-if="playDates.size > 0" class="play-badge">
            <view class="play-dot" />
            <text class="play-badge-text">× {{ playDates.size }}</text>
          </view>
        </view>
        <view v-if="selectedDates.length" class="schedule-summary">
          <text class="helper">{{ scheduleSummary }}　长按日期标记陪玩</text>
          <view class="chip-list compact">
            <view
              :class="['chip', 'soft', 'date-chip', playDates.has(date) ? 'date-chip-play' : '']"
              v-for="date in selectedDates"
              :key="date"
              @longpress="togglePlayDate(date)"
            >
              {{ date }}
              <view v-if="playDates.has(date)" class="play-dot in-chip" />
            </view>
          </view>
        </view>
        <view v-else-if="sortedRules.length" class="rule-list">
          <view class="rule-row" v-for="rule in sortedRules" :key="`${rule.weekday}-${rule.window_code}`">
            <text>{{ weekdayLabel(rule.weekday) }}</text>
            <text>{{ feedingWindowLabel(rule.window_code) }} × {{ rule.visit_count }}</text>
          </view>
        </view>
        <text v-else class="helper">未配置上门安排</text>
      </view>

      <view class="section-card" v-if="historyPlans.length" id="history-plans">
        <view class="section-head">
          <text class="section-title">历史计划</text>
          <text class="section-desc">同客户的上门喂养计划</text>
        </view>
        <view class="history-list">
          <view class="history-order" v-for="item in historyPlans" :key="item.ID" @click="goHistoryPlan(item.ID)">
            <view class="history-main">
              <text class="history-no">计划 #{{ item.ID }}</text>
              <text class="history-amount">¥{{ Number(item.total_amount || 0).toFixed(2) }}</text>
            </view>
            <view class="history-sub">
              <text>{{ planDateText(item) }}</text>
              <text>{{ feedingStatusLabel(item.status) }}</text>
            </view>
          </view>
        </view>
      </view>

      <view class="modal-mask" v-if="showAssign" @click="showAssign = false">
        <view class="modal" @click.stop>
          <text class="modal-title">选择执行员工</text>
          <view class="option-list">
            <view class="option" v-for="staff in staffs" :key="staff.ID" @click="assignVisit(staff.ID)">{{ staff.name }}</view>
          </view>
        </view>
      </view>
    </view>
  </SideLayout>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { onLoad, onShow } from '@dcloudio/uni-app'
import SideLayout from '@/components/SideLayout.vue'
import { assignFeedingVisit, cancelFeedingPlan, generateFeedingOrder, getFeedingPlan, getFeedingPlans, pauseFeedingPlan, resumeFeedingPlan, updateFeedingDeposit, updateFeedingPlayDates, updateFeedingVisitNote } from '@/api/feeding'
import { getStaffList } from '@/api/staff'
import { useAuthStore } from '@/store/auth'
import { hasStaffRoleAtLeast } from '@/utils/staff-role'
import { feedingStatusLabel, feedingWindowLabel, feedingWeekdays, formatFeedingDateRange, getFeedingDateOptions, parseFeedingAddress, parseFeedingSelectedDates, parseFeedingSelectedItems } from '@/utils/feeding'

const authStore = useAuthStore()
const canOperate = computed(() => hasStaffRoleAtLeast(authStore.staffInfo?.role, 'staff'))
const id = ref(0)
const plan = ref<FeedingPlan | null>(null)
const showAssign = ref(false)
const assignTarget = ref<FeedingVisit | null>(null)
const staffs = ref<Staff[]>([])

const planHasPlay = computed(() => {
  const items = parseFeedingSelectedItems(plan.value?.selected_items_json)
  return items.some(item => item.code === 'play')
})
const editingNoteId = ref<number | null>(null)
const editingNoteText = ref('')
const playDates = ref<Set<string>>(new Set())
const savingPlayDates = ref(false)
const depositInput = ref('')
const historyPlans = ref<FeedingPlan[]>([])
const balanceDisplay = computed(() => {
  const total = plan.value?.total_amount || 0
  const deposit = Number(depositInput.value) || 0
  return (total - deposit).toFixed(2)
})

async function saveDeposit() {
  if (!plan.value) return
  const val = Number(depositInput.value) || 0
  await updateFeedingDeposit(plan.value.ID, val)
  await loadData()
}

async function togglePlayDate(date: string) {
  if (!plan.value?.ID || savingPlayDates.value) return
  const next = new Set(playDates.value)
  if (next.has(date)) {
    next.delete(date)
  } else {
    next.add(date)
  }
  const previous = new Set(playDates.value)
  playDates.value = next
  uni.vibrateShort({})
  savingPlayDates.value = true
  try {
    const res = await updateFeedingPlayDates(plan.value.ID, Array.from(next).sort())
    if (res.data) {
      playDates.value = new Set(parseFeedingSelectedDates(res.data.play_dates_json))
      plan.value = { ...(plan.value as FeedingPlan), play_dates_json: res.data.play_dates_json }
    }
    uni.showToast({ title: next.has(date) ? '已标记陪玩' : '已取消陪玩', icon: 'none' })
  } catch (err: any) {
    playDates.value = previous
    uni.showToast({ title: err?.message || '保存失败', icon: 'none' })
  } finally {
    savingPlayDates.value = false
  }
}

const address = computed(() => parseFeedingAddress(plan.value?.address_snapshot_json))
const selectedItems = computed(() => parseFeedingSelectedItems(plan.value?.selected_items_json))
const sortedRules = computed(() => [...(plan.value?.rules || [])].sort((a, b) => a.weekday - b.weekday || a.window_code.localeCompare(b.window_code)))
const selectedDates = computed(() => {
  const parsed = parseFeedingSelectedDates(plan.value?.selected_dates_json)
  if (parsed.length) return parsed
  return Array.from(new Set((plan.value?.visits || []).map(item => item.scheduled_date).filter(Boolean))).sort()
})
const scheduleSummary = computed(() => {
  if (!plan.value || !selectedDates.value.length) return '未配置上门安排'
  const totalDates = getFeedingDateOptions(plan.value.start_date, plan.value.end_date).length
  return selectedDates.value.length === totalDates
    ? `按日期范围每天上门，共 ${selectedDates.value.length} 天`
    : `按指定日期上门，共 ${selectedDates.value.length} 天`
})

function weekdayLabel(weekday: number) {
  return feedingWeekdays.find(item => item.value === weekday)?.label || `周${weekday}`
}

function goEdit() {
  if (!plan.value?.ID) return
  uni.navigateTo({ url: `/pages/feeding/create?id=${plan.value.ID}` })
}

function openExecute(visit: FeedingVisit) {
  if (!plan.value?.ID) return
  uni.navigateTo({ url: `/pages/feeding/visit-execute?plan_id=${plan.value.ID}&visit_id=${visit.ID}` })
}

function openAssign(visit: FeedingVisit) {
  assignTarget.value = visit
  showAssign.value = true
}

async function assignVisit(staffId: number) {
  if (!assignTarget.value) return
  await assignFeedingVisit(assignTarget.value.ID, staffId)
  showAssign.value = false
  await loadData()
}

async function loadStaffs() {
  if (!canOperate.value) return
  const res = await getStaffList({ page: 1, page_size: 100 })
  staffs.value = res.data?.list || []
}

async function loadData() {
  if (!id.value) return
  const res = await getFeedingPlan(id.value)
  plan.value = res.data || null
  if (plan.value) {
    depositInput.value = plan.value.deposit ? String(plan.value.deposit) : ''
    playDates.value = new Set(parseFeedingSelectedDates(plan.value.play_dates_json))
    await loadHistoryPlans()
  }
}

async function loadHistoryPlans() {
  if (!plan.value?.customer_id) {
    historyPlans.value = []
    return
  }
  try {
    const res = await getFeedingPlans({
      page: 1,
      page_size: 20,
      customer_id: plan.value.customer_id,
    })
    const list = res.data?.list || []
    historyPlans.value = list.filter((item: FeedingPlan) => Number(item.ID) !== Number(plan.value?.ID))
  } catch {
    historyPlans.value = []
  }
}

function startEditNote(visit: FeedingVisit) {
  editingNoteId.value = visit.ID
  editingNoteText.value = visit.internal_note || ''
}

async function saveNote(visit: FeedingVisit) {
  await updateFeedingVisitNote(visit.ID, { internal_note: editingNoteText.value })
  visit.internal_note = editingNoteText.value
  editingNoteId.value = null
}

function cancelEditNote() {
  editingNoteId.value = null
}

function pausePlan() {
  if (!plan.value) return
  uni.showModal({
    title: '暂停计划',
    content: '确认暂停当前上门喂养计划？',
    success: async (res) => {
      if (!res.confirm || !plan.value) return
      await pauseFeedingPlan(plan.value.ID)
      await loadData()
    },
  })
}

function resumePlan() {
  if (!plan.value) return
  uni.showModal({
    title: '恢复计划',
    content: '确认恢复当前上门喂养计划？',
    success: async (res) => {
      if (!res.confirm || !plan.value) return
      await resumeFeedingPlan(plan.value.ID)
      await loadData()
    },
  })
}

function cancelPlanAction() {
  if (!plan.value) return
  uni.showModal({
    title: '取消计划',
    content: '取消后，未执行任务会全部作废。确认继续？',
    success: async (res) => {
      if (!res.confirm || !plan.value) return
      await cancelFeedingPlan(plan.value.ID)
      await loadData()
    },
  })
}

function generateOrderAction() {
  if (!plan.value) return
  uni.showModal({
    title: '生成订单',
    content: '确认按已完成任务生成订单？',
    success: async (res) => {
      if (!res.confirm || !plan.value) return
      const result = await generateFeedingOrder(plan.value.ID)
      const orderId = result.data?.ID
      if (orderId) {
        uni.navigateTo({ url: `/pages/order/detail?id=${orderId}` })
      } else {
        await loadData()
      }
    },
  })
}

function goOrder() {
  if (!plan.value?.order_id) return
  uni.navigateTo({ url: `/pages/order/detail?id=${plan.value.order_id}` })
}

function goHistoryPlan(planId: number) {
  uni.navigateTo({ url: `/pages/feeding/detail?id=${planId}` })
}

function planDateText(item: FeedingPlan) {
  return formatFeedingDateRange(item.start_date, item.end_date)
}

onLoad((options) => {
  id.value = Number(options?.id || 0)
})

onShow(async () => {
  await loadStaffs()
  await loadData()
})
</script>

<style scoped>
.page { padding: 24rpx 24rpx 120rpx; }
.hero-card, .info-card, .section-card { background: #fff; border-radius: 22rpx; box-shadow: 0 12rpx 28rpx rgba(15, 23, 42, 0.06); padding: 24rpx; margin-bottom: 18rpx; }
.hero-card { display: flex; justify-content: space-between; gap: 18rpx; }
.hero-title { display: block; font-size: 34rpx; font-weight: 700; color: #111827; }
.hero-subtitle { display: block; margin-top: 10rpx; font-size: 24rpx; color: #6B7280; }
.hero-actions { display: flex; flex-wrap: wrap; gap: 12rpx; justify-content: flex-end; }
.btn { padding: 14rpx 22rpx; border-radius: 16rpx; background: #F8FAFC; color: #374151; font-size: 24rpx; border: 1rpx solid #E5E7EB; }
.btn-primary { background: linear-gradient(135deg, #4F46E5, #6366F1); color: #fff; border-color: transparent; }
.btn.danger { color: #DC2626; }
.compact { padding: 18rpx 20rpx; }
.info-row { display: flex; align-items: center; gap: 8rpx; padding: 10rpx 0; border-bottom: 1rpx solid #F3F4F6; }
.info-row:last-child { border-bottom: none; }
.info-label { font-size: 24rpx; color: #111827; font-weight: 700; flex-shrink: 0; min-width: 64rpx; }
.info-val { font-size: 24rpx; color: #6B7280; flex: 1; }
.info-val.amount { color: #4F46E5; font-weight: 700; font-size: 28rpx; }
.chip-list.inline { flex: 1; }
.chip.sm { padding: 6rpx 14rpx; font-size: 22rpx; }
.inline-input { width: 140rpx; height: 52rpx; padding: 0 12rpx; background: #F8FAFC; border-radius: 10rpx; font-size: 26rpx; color: #111827; font-weight: 600; text-align: center; }
.section-head { display: flex; justify-content: space-between; gap: 12rpx; align-items: center; margin-bottom: 16rpx; }
.section-title { font-size: 28rpx; font-weight: 700; color: #111827; }
.section-desc { font-size: 22rpx; color: #6B7280; }
.helper { display: block; font-size: 22rpx; color: #6B7280; }
.chip-list { display: flex; flex-wrap: wrap; gap: 12rpx; }
.chip { padding: 12rpx 20rpx; border-radius: 999rpx; background: #EEF2FF; color: #4F46E5; font-size: 24rpx; }
.chip.soft { background: #F3F4F6; color: #374151; }
.chip-list.compact { margin-top: 12rpx; }
.schedule-summary { display: flex; flex-direction: column; gap: 12rpx; }
.rule-list, .visit-list { display: flex; flex-direction: column; gap: 8rpx; }
.rule-row, .visit-card { padding: 14rpx 18rpx; border-radius: 14rpx; background: #F8FAFC; }
.history-list { display: flex; flex-direction: column; gap: 12rpx; }
.history-order { padding: 18rpx; border-radius: 16rpx; background: #F8FAFC; }
.history-main { display: flex; justify-content: space-between; align-items: center; gap: 12rpx; }
.history-no { font-size: 24rpx; color: #111827; font-weight: 600; }
.history-amount { font-size: 24rpx; color: #4F46E5; font-weight: 700; }
.history-sub { display: flex; justify-content: space-between; gap: 12rpx; margin-top: 8rpx; font-size: 22rpx; color: #6B7280; }
.rule-row { display: flex; justify-content: space-between; gap: 12rpx; font-size: 24rpx; color: #374151; }
.visit-head { display: flex; justify-content: space-between; gap: 10rpx; align-items: center; }
.visit-head-left { flex: 1; min-width: 0; }
.visit-title { display: block; font-size: 26rpx; font-weight: 700; color: #111827; }
.visit-meta { display: block; font-size: 22rpx; color: #6B7280; }
.visit-price { font-size: 24rpx; color: #4F46E5; font-weight: 700; white-space: nowrap; }
.visit-row { display: flex; justify-content: space-between; align-items: center; margin-top: 8rpx; }
.exception-text { display: block; margin-top: 6rpx; font-size: 22rpx; color: #DC2626; }
.visit-actions { display: flex; gap: 8rpx; flex-shrink: 0; }
.mini-btn { padding: 8rpx 18rpx; border-radius: 999rpx; background: #fff; color: #4B5563; font-size: 22rpx; border: 1rpx solid #E5E7EB; }
.mini-btn.primary { background: #4F46E5; color: #fff; border-color: #4F46E5; }
.chip.muted { background: #E5E7EB; color: #9CA3AF; }
.date-chip { position: relative; -webkit-user-select: none; user-select: none; -webkit-touch-callout: none; }
.date-chip-play { background: #FEF3C7; color: #92400E; border: 1rpx solid #F59E0B; }
.play-dot { width: 14rpx; height: 14rpx; border-radius: 50%; background: #F97316; flex-shrink: 0; }
.play-dot.in-chip { position: absolute; top: 6rpx; right: 6rpx; }
.play-badge { display: flex; align-items: center; gap: 6rpx; }
.play-badge-text { font-size: 22rpx; color: #F97316; font-weight: 600; }
.visit-note { margin-top: 6rpx; }
.note-display { cursor: pointer; }
.note-text { font-size: 22rpx; color: #9CA3AF; }
.note-edit { display: flex; align-items: center; gap: 8rpx; }
.note-input { flex: 1; font-size: 22rpx; padding: 8rpx 14rpx; border: 1rpx solid #E5E7EB; border-radius: 10rpx; background: #fff; }
.modal-mask { position: fixed; inset: 0; background: rgba(15, 23, 42, 0.45); display: flex; align-items: center; justify-content: center; padding: 32rpx; z-index: 5000; }
.modal { width: 100%; max-width: 620rpx; background: #fff; border-radius: 22rpx; padding: 28rpx; }
.modal-title { display: block; font-size: 30rpx; font-weight: 700; color: #111827; margin-bottom: 18rpx; }
.option-list { display: flex; flex-direction: column; gap: 12rpx; }
.option { padding: 20rpx; border-radius: 16rpx; background: #F8FAFC; font-size: 24rpx; color: #374151; }
@media (max-width: 768px) {
  .hero-card { flex-direction: column; }
  .hero-actions { justify-content: flex-start; }
  .info-grid { grid-template-columns: 1fr; }
}
</style>
