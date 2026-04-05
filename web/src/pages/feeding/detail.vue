<template>
  <SideLayout>
    <view class="page" v-if="plan">
      <view class="hero-card">
        <view>
          <text class="hero-title">喂养计划 #{{ plan.ID }}</text>
          <text class="hero-subtitle">{{ feedingStatusLabel(plan.status) }} · {{ formatFeedingDateRange(plan.start_date, plan.end_date) }}</text>
        </view>
        <view class="hero-actions">
          <view v-if="canManage" class="btn" @click="goEdit">修改</view>
          <view v-if="canManage && plan.status === 'active'" class="btn" @click="pausePlan">暂停</view>
          <view v-if="canManage && plan.status === 'paused'" class="btn" @click="resumePlan">恢复</view>
          <view v-if="canManage && plan.status !== 'cancelled' && plan.status !== 'completed'" class="btn danger" @click="cancelPlanAction">取消</view>
          <view v-if="canManage && !plan.order_id" class="btn btn-primary" @click="generateOrderAction">生成订单</view>
          <view v-else-if="plan.order_id" class="btn btn-primary" @click="goOrder">查看订单</view>
        </view>
      </view>

      <view class="info-grid">
        <view class="info-card">
          <text class="info-label">客户</text>
          <text class="info-value">{{ plan.customer?.nickname || plan.customer?.phone || '-' }}</text>
        </view>
        <view class="info-card">
          <text class="info-label">联系人</text>
          <text class="info-value">{{ plan.contact_name || '-' }}</text>
        </view>
        <view class="info-card">
          <text class="info-label">联系电话</text>
          <text class="info-value">{{ plan.contact_phone || '-' }}</text>
        </view>
        <view class="info-card">
          <text class="info-label">预计金额</text>
          <text class="info-value amount">¥{{ plan.total_amount.toFixed(2) }}</text>
        </view>
      </view>

      <view class="section-card">
        <text class="section-title">服务地址</text>
        <text class="multi-line">{{ address.address || '-' }}</text>
        <text v-if="address.detail" class="helper">补充：{{ address.detail }}</text>
        <text v-if="address.door_code" class="helper">入户：{{ address.door_code }}</text>
      </view>

      <view class="section-card">
        <text class="section-title">本次猫咪</text>
        <view class="chip-list">
          <view class="chip" v-for="item in plan.pets || []" :key="item.ID || item.pet_id">{{ item.pet?.name || item.pet_name_snapshot }}</view>
        </view>
      </view>

      <view class="section-card">
        <text class="section-title">服务内容</text>
        <view class="chip-list">
          <view class="chip soft" v-for="item in selectedItems" :key="item.code">{{ item.name }}</view>
        </view>
      </view>

      <view class="section-card">
        <text class="section-title">时间窗规则</text>
        <view class="rule-list">
          <view class="rule-row" v-for="rule in sortedRules" :key="`${rule.weekday}-${rule.window_code}`">
            <text>{{ weekdayLabel(rule.weekday) }}</text>
            <text>{{ feedingWindowLabel(rule.window_code) }} × {{ rule.visit_count }}</text>
          </view>
        </view>
      </view>

      <view class="section-card">
        <view class="section-head">
          <text class="section-title">上门记录</text>
          <text class="section-desc">共 {{ plan.visits?.length || 0 }} 条</text>
        </view>
        <view class="visit-list">
          <view class="visit-card" v-for="visit in plan.visits || []" :key="visit.ID">
            <view class="visit-head">
              <view>
                <text class="visit-title">{{ visit.scheduled_date }} · {{ feedingWindowLabel(visit.window_code) }}</text>
                <text class="visit-meta">状态：{{ feedingStatusLabel(visit.status) }} · 执行人：{{ visit.staff?.name || '待分配' }}</text>
              </view>
              <text class="visit-price">¥{{ visit.visit_price.toFixed(2) }}</text>
            </view>
            <view class="chip-list compact">
              <view class="chip soft" v-for="item in visit.items || []" :key="item.ID">{{ item.item_name_snapshot }}</view>
            </view>
            <text v-if="visit.exception_type" class="exception-text">异常：{{ visit.exception_type }}</text>
            <view class="visit-actions">
              <view v-if="canManage" class="mini-btn" @click="openAssign(visit)">改派</view>
              <view class="mini-btn primary" @click="openExecute(visit)">执行</view>
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
import { assignFeedingVisit, cancelFeedingPlan, generateFeedingOrder, getFeedingPlan, pauseFeedingPlan, resumeFeedingPlan } from '@/api/feeding'
import { getStaffList } from '@/api/staff'
import { useAuthStore } from '@/store/auth'
import { hasStaffRoleAtLeast } from '@/utils/staff-role'
import { feedingStatusLabel, feedingWindowLabel, feedingWeekdays, formatFeedingDateRange, parseFeedingAddress, parseFeedingSelectedItems } from '@/utils/feeding'

const authStore = useAuthStore()
const canManage = computed(() => hasStaffRoleAtLeast(authStore.staffInfo?.role, 'manager'))
const id = ref(0)
const plan = ref<FeedingPlan | null>(null)
const showAssign = ref(false)
const assignTarget = ref<FeedingVisit | null>(null)
const staffs = ref<Staff[]>([])

const address = computed(() => parseFeedingAddress(plan.value?.address_snapshot_json))
const selectedItems = computed(() => parseFeedingSelectedItems(plan.value?.selected_items_json))
const sortedRules = computed(() => [...(plan.value?.rules || [])].sort((a, b) => a.weekday - b.weekday || a.window_code.localeCompare(b.window_code)))

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
  if (!canManage.value) return
  const res = await getStaffList({ page: 1, page_size: 100 })
  staffs.value = res.data?.list || []
}

async function loadData() {
  if (!id.value) return
  const res = await getFeedingPlan(id.value)
  plan.value = res.data || null
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
.info-grid { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 16rpx; margin-bottom: 18rpx; }
.info-label { display: block; font-size: 22rpx; color: #6B7280; }
.info-value { display: block; margin-top: 10rpx; font-size: 28rpx; color: #111827; font-weight: 600; }
.amount { color: #4F46E5; }
.section-head { display: flex; justify-content: space-between; gap: 12rpx; align-items: center; margin-bottom: 16rpx; }
.section-title { font-size: 28rpx; font-weight: 700; color: #111827; }
.section-desc { font-size: 22rpx; color: #6B7280; }
.multi-line, .helper { display: block; font-size: 24rpx; color: #374151; line-height: 1.8; }
.helper { color: #6B7280; margin-top: 8rpx; }
.chip-list { display: flex; flex-wrap: wrap; gap: 12rpx; }
.chip { padding: 12rpx 20rpx; border-radius: 999rpx; background: #EEF2FF; color: #4F46E5; font-size: 24rpx; }
.chip.soft { background: #F3F4F6; color: #374151; }
.chip-list.compact { margin-top: 12rpx; }
.rule-list, .visit-list { display: flex; flex-direction: column; gap: 12rpx; }
.rule-row, .visit-card { padding: 18rpx; border-radius: 18rpx; background: #F8FAFC; }
.rule-row { display: flex; justify-content: space-between; gap: 12rpx; font-size: 24rpx; color: #374151; }
.visit-head { display: flex; justify-content: space-between; gap: 14rpx; align-items: flex-start; }
.visit-title { display: block; font-size: 26rpx; font-weight: 700; color: #111827; }
.visit-meta { display: block; margin-top: 6rpx; font-size: 22rpx; color: #6B7280; line-height: 1.6; }
.visit-price { font-size: 24rpx; color: #4F46E5; font-weight: 700; }
.exception-text { display: block; margin-top: 10rpx; font-size: 22rpx; color: #DC2626; }
.visit-actions { display: flex; gap: 12rpx; margin-top: 14rpx; }
.mini-btn { padding: 10rpx 20rpx; border-radius: 999rpx; background: #fff; color: #4B5563; font-size: 22rpx; border: 1rpx solid #E5E7EB; }
.mini-btn.primary { background: #4F46E5; color: #fff; border-color: #4F46E5; }
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
