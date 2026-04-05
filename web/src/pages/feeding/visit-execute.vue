<template>
  <SideLayout>
    <view class="page" v-if="visit">
      <view class="hero-card">
        <view>
          <text class="hero-title">{{ visit.scheduled_date }} · {{ feedingWindowLabel(visit.window_code) }}</text>
          <text class="hero-subtitle">{{ plan?.customer?.nickname || plan?.customer?.phone || '-' }} · {{ petSummary }}</text>
        </view>
        <text class="hero-status">{{ feedingStatusLabel(visit.status) }}</text>
      </view>

      <view class="section-card">
        <text class="section-title">完成项</text>
        <view class="item-list">
          <view class="item-row" v-for="item in itemChecks" :key="item.ID" @click="item.checked = !item.checked">
            <text>{{ item.item_name_snapshot }}</text>
            <text :class="['item-check', item.checked ? 'active' : '']">{{ item.checked ? '已完成' : '待完成' }}</text>
          </view>
        </view>
      </view>

      <view class="section-card">
        <text class="section-title">履约图片</text>
        <view class="photo-list">
          <image v-for="item in mediaList" :key="item.ID" :src="item.url" class="photo" mode="aspectFill" />
          <view class="photo-add" @click="uploadMedia">+</view>
        </view>
      </view>

      <view class="section-card">
        <text class="section-title">客户可见备注</text>
        <textarea v-model="customerNote" class="textarea" placeholder="例如：猫咪状态正常，已完成添粮换水" />
      </view>

      <view class="section-card">
        <text class="section-title">内部备注 / 异常</text>
        <textarea v-model="internalNote" class="textarea" placeholder="例如：开门耗时较长，建议下次提前提供门禁密码" />
        <input v-model="exceptionType" class="input" placeholder="异常类型，例如：无法入户 / 猫咪躲藏 / 喂药失败" />
      </view>

      <view class="footer-bar">
        <view v-if="visit.status === 'pending' || visit.status === 'assigned'" class="action-btn ghost" @click="startAction">开始执行</view>
        <view class="action-btn warn" @click="exceptionAction">标记异常</view>
        <view class="action-btn primary" @click="completeAction">完成本次上门</view>
      </view>
    </view>
  </SideLayout>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { onLoad, onShow } from '@dcloudio/uni-app'
import SideLayout from '@/components/SideLayout.vue'
import { addFeedingVisitMedia, completeFeedingVisit, exceptionFeedingVisit, getFeedingPlan, startFeedingVisit } from '@/api/feeding'
import { uploadFile } from '@/api/upload'
import { feedingStatusLabel, feedingWindowLabel } from '@/utils/feeding'

const planId = ref(0)
const visitId = ref(0)
const plan = ref<FeedingPlan | null>(null)
const visit = ref<FeedingVisit | null>(null)
const itemChecks = ref<FeedingVisitItem[]>([])
const mediaList = ref<FeedingVisitMedia[]>([])
const customerNote = ref('')
const internalNote = ref('')
const exceptionType = ref('')

const petSummary = computed(() => plan.value?.pets?.map(item => item.pet?.name || item.pet_name_snapshot).filter(Boolean).join('、') || '未选猫咪')

async function loadData() {
  if (!planId.value || !visitId.value) return
  const res = await getFeedingPlan(planId.value)
  plan.value = res.data || null
  visit.value = (plan.value?.visits || []).find(item => item.ID === visitId.value) || null
  itemChecks.value = (visit.value?.items || []).map(item => ({ ...item }))
  mediaList.value = [...(visit.value?.media || [])]
  customerNote.value = visit.value?.customer_note || ''
  internalNote.value = visit.value?.internal_note || ''
  exceptionType.value = visit.value?.exception_type || ''
}

async function uploadMedia() {
  if (!visit.value) return
  const chooseRes = await new Promise<UniApp.ChooseImageSuccessCallbackResult>((resolve, reject) => {
    uni.chooseImage({ count: 1, sizeType: ['compressed'], success: resolve, fail: reject })
  })
  const filePath = chooseRes.tempFilePaths?.[0]
  if (!filePath) return
  const url = await uploadFile(filePath)
  const res = await addFeedingVisitMedia(visit.value.ID, { media_type: 'image', url })
  mediaList.value.push(res.data)
}

async function startAction() {
  if (!visit.value) return
  await startFeedingVisit(visit.value.ID)
  await loadData()
}

async function completeAction() {
  if (!visit.value) return
  await completeFeedingVisit(visit.value.ID, {
    item_checks: itemChecks.value.map(item => ({ id: item.ID, checked: item.checked })),
    customer_note: customerNote.value,
    internal_note: internalNote.value,
  })
  await loadData()
  uni.showToast({ title: '已完成', icon: 'success' })
}

async function exceptionAction() {
  if (!visit.value) return
  if (!exceptionType.value.trim()) {
    uni.showToast({ title: '请先填写异常类型', icon: 'none' })
    return
  }
  await exceptionFeedingVisit(visit.value.ID, {
    exception_type: exceptionType.value,
    customer_note: customerNote.value,
    internal_note: internalNote.value,
  })
  await loadData()
}

onLoad((options) => {
  planId.value = Number(options?.plan_id || 0)
  visitId.value = Number(options?.visit_id || 0)
})

onShow(loadData)
</script>

<style scoped>
.page { padding: 24rpx 24rpx 180rpx; }
.hero-card, .section-card { background: #fff; border-radius: 22rpx; box-shadow: 0 12rpx 28rpx rgba(15, 23, 42, 0.06); padding: 24rpx; margin-bottom: 18rpx; }
.hero-card { display: flex; justify-content: space-between; gap: 18rpx; }
.hero-title { display: block; font-size: 34rpx; font-weight: 700; color: #111827; }
.hero-subtitle { display: block; margin-top: 8rpx; font-size: 24rpx; color: #6B7280; }
.hero-status { font-size: 24rpx; color: #4F46E5; }
.section-title { display: block; font-size: 28rpx; font-weight: 700; color: #111827; margin-bottom: 18rpx; }
.item-list { display: flex; flex-direction: column; gap: 12rpx; }
.item-row { padding: 18rpx; border-radius: 18rpx; background: #F8FAFC; display: flex; justify-content: space-between; gap: 12rpx; font-size: 24rpx; color: #374151; }
.item-check { color: #9CA3AF; }
.item-check.active { color: #10B981; font-weight: 700; }
.photo-list { display: flex; flex-wrap: wrap; gap: 12rpx; }
.photo, .photo-add { width: 160rpx; height: 160rpx; border-radius: 18rpx; }
.photo-add { display: flex; align-items: center; justify-content: center; background: #EEF2FF; color: #4F46E5; font-size: 60rpx; }
.textarea, .input { width: 100%; box-sizing: border-box; padding: 22rpx 24rpx; background: #F8FAFC; border-radius: 18rpx; font-size: 26rpx; color: #111827; }
.textarea { min-height: 160rpx; }
.input { margin-top: 14rpx; min-height: 88rpx; }
.footer-bar { position: fixed; left: 24rpx; right: 24rpx; bottom: calc(32rpx + env(safe-area-inset-bottom)); display: flex; gap: 12rpx; z-index: 20; }
.action-btn { flex: 1; min-height: 92rpx; border-radius: 999rpx; display: flex; align-items: center; justify-content: center; font-size: 28rpx; }
.action-btn.ghost { background: #fff; color: #4B5563; border: 1rpx solid #E5E7EB; }
.action-btn.warn { background: #FEF2F2; color: #DC2626; }
.action-btn.primary { background: linear-gradient(135deg, #4F46E5, #6366F1); color: #fff; }
</style>
