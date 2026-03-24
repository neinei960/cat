<template>
  <SideLayout>
  <view class="page">
    <view v-if="loading" class="state">加载中...</view>
    <view v-else-if="!appt" class="state">预约不存在</view>
    <template v-else>
      <view class="summary-card">
        <text class="summary-title">预约开单确认</text>
        <text class="summary-line">{{ appt.date }} {{ appt.start_time }} - {{ appt.end_time }}</text>
        <text class="summary-line">客户：{{ appt.customer?.nickname || appt.customer?.phone || '-' }}</text>
        <text class="summary-line">技师：{{ appt.staff?.name || '待分配' }}</text>
        <text class="summary-amount">预计共 {{ orderDrafts.length }} 单 · ¥{{ totalAmount.toFixed(2) }}</text>
      </view>

      <view class="draft-list">
        <view class="draft-card" v-for="draft in orderDrafts" :key="draft.petId">
          <view class="draft-head">
            <text class="draft-name">{{ draft.petName }}</text>
            <text class="draft-price">¥{{ draft.amount.toFixed(2) }}</text>
          </view>
          <text class="draft-meta">{{ draft.meta }}</text>
          <view class="draft-tags" v-if="draft.tags.length > 0">
            <text
              v-for="tag in draft.tags"
              :key="`${draft.petId}-${tag.text}`"
              :class="['draft-tag', tag.className]"
              :style="tag.style"
            >{{ tag.text }}</text>
          </view>
          <view class="service-list">
            <view class="service-row" v-for="svc in draft.services" :key="`${draft.petId}-${svc.service_id}`">
              <text class="service-name">{{ svc.service_name }}</text>
              <text class="service-meta">¥{{ svc.price }} · {{ svc.duration }}分钟</text>
            </view>
          </view>
        </view>
      </view>

      <view class="notes-card" v-if="appt.notes">
        <text class="notes-title">预约备注</text>
        <text class="notes-text">{{ appt.notes }}</text>
      </view>

      <button class="submit-btn" :loading="submitting" @click="submitBatch">确认生成 {{ orderDrafts.length }} 张订单</button>
    </template>
  </view>
  </SideLayout>
</template>

<script setup lang="ts">
import SideLayout from '@/components/SideLayout.vue'
import { computed, ref } from 'vue'
import { onLoad } from '@dcloudio/uni-app'
import { getAppointment } from '@/api/appointment'
import { createBatchOrdersFromAppointment } from '@/api/order'
import { getPersonalityBg, getPersonalityColor } from '@/utils/personality'

const appointmentId = ref(0)
const appt = ref<any>(null)
const loading = ref(true)
const submitting = ref(false)

function calcAge(birthDate?: string): string {
  if (!birthDate) return ''
  const birth = new Date(birthDate)
  if (Number.isNaN(birth.getTime())) return ''
  const now = new Date()
  const months = (now.getFullYear() - birth.getFullYear()) * 12 + (now.getMonth() - birth.getMonth())
  if (months < 1) return '不到1个月'
  if (months < 12) return `${months}个月`
  const years = Math.floor(months / 12)
  const rem = months % 12
  return rem > 0 ? `${years}岁${rem}个月` : `${years}岁`
}

function getPetMeta(pet: any) {
  const parts: string[] = []
  if (pet?.breed) parts.push(pet.breed)
  if (pet?.gender === 1) parts.push('弟弟')
  if (pet?.gender === 2) parts.push('妹妹')
  return parts.join(' · ') || '未填写宠物信息'
}

function getPetTags(pet: any) {
  const tags: Array<{ text: string; className: string; style?: string }> = []
  const age = calcAge(pet?.birth_date)
  if (age) tags.push({ text: age, className: 'tag-age' })
  if (pet?.fur_level) tags.push({ text: pet.fur_level, className: 'tag-fur' })
  if (pet?.neutered) tags.push({ text: '已绝育', className: 'tag-neutered' })
  if (pet?.personality) {
    tags.push({
      text: pet.personality,
      className: 'tag-personality',
      style: `background:${getPersonalityBg(pet.personality)};color:${getPersonalityColor(pet.personality)};`,
    })
  }
  if (pet?.aggression && pet.aggression !== '无') {
    tags.push({ text: `⚡ ${pet.aggression}`, className: 'tag-aggression' })
  }
  return tags
}

const orderDrafts = computed(() => {
  const pets = Array.isArray(appt.value?.pets) ? appt.value.pets : []
  return pets.map((petItem: any) => ({
    petId: petItem.pet_id,
    petName: petItem.pet?.name || `宠物#${petItem.pet_id}`,
    meta: getPetMeta(petItem.pet),
    tags: getPetTags(petItem.pet),
    amount: Number(petItem.total_amount || 0),
    services: petItem.services || [],
  }))
})

const totalAmount = computed(() => orderDrafts.value.reduce((sum, draft) => sum + draft.amount, 0))

async function loadData() {
  if (!appointmentId.value) return
  loading.value = true
  try {
    const res = await getAppointment(appointmentId.value)
    appt.value = res.data
  } finally {
    loading.value = false
  }
}

async function submitBatch() {
  if (!appointmentId.value) return
  submitting.value = true
  try {
    const res = await createBatchOrdersFromAppointment(appointmentId.value)
    const orders = res.data || []
    uni.showToast({ title: `已生成${orders.length}张订单`, icon: 'success' })
    setTimeout(() => {
      if (orders.length === 1 && orders[0]?.ID) {
        uni.redirectTo({ url: `/pages/order/detail?id=${orders[0].ID}` })
      } else {
        uni.redirectTo({ url: '/pages/order/list' })
      }
    }, 500)
  } catch (e: any) {
    uni.showToast({ title: e.message || '批量开单失败', icon: 'none' })
  } finally {
    submitting.value = false
  }
}

onLoad((query) => {
  appointmentId.value = parseInt(String(query?.appointment_id || 0)) || 0
  loadData()
})
</script>

<style scoped>
.page { padding: 24rpx; }
.state { text-align: center; padding: 120rpx 0; color: #9CA3AF; font-size: 28rpx; }
.summary-card, .draft-card, .notes-card { background: #fff; border-radius: 18rpx; padding: 24rpx; margin-bottom: 16rpx; box-shadow: 0 4rpx 18rpx rgba(15, 23, 42, 0.06); }
.summary-title { font-size: 32rpx; font-weight: 700; color: #111827; display: block; margin-bottom: 12rpx; }
.summary-line { font-size: 24rpx; color: #6B7280; display: block; margin-top: 6rpx; }
.summary-amount { font-size: 28rpx; color: #4F46E5; font-weight: 700; display: block; margin-top: 16rpx; }
.draft-list { display: flex; flex-direction: column; gap: 16rpx; }
.draft-head { display: flex; justify-content: space-between; align-items: center; gap: 16rpx; }
.draft-name { font-size: 30rpx; font-weight: 700; color: #111827; }
.draft-price { font-size: 28rpx; color: #4F46E5; font-weight: 700; }
.draft-meta { font-size: 24rpx; color: #6B7280; display: block; margin-top: 8rpx; }
.draft-tags { display: flex; flex-wrap: wrap; gap: 8rpx; margin-top: 10rpx; }
.draft-tag { display: inline-flex; align-items: center; padding: 4rpx 12rpx; border-radius: 999rpx; font-size: 18rpx; line-height: 1.2; background: #F3F4F6; color: #4B5563; }
.draft-tag.tag-age { background: #F8FAFC; color: #475569; }
.draft-tag.tag-fur { background: #EEF2FF; color: #4F46E5; }
.draft-tag.tag-neutered { background: #ECFDF5; color: #047857; }
.draft-tag.tag-aggression { background: #FEF2F2; color: #DC2626; }
.service-list { margin-top: 16rpx; border-top: 1rpx solid #EEF2F7; padding-top: 8rpx; }
.service-row { display: flex; justify-content: space-between; gap: 16rpx; padding: 12rpx 0; border-bottom: 1rpx solid #F3F4F6; }
.service-row:last-child { border-bottom: none; padding-bottom: 0; }
.service-name { font-size: 24rpx; color: #1F2937; }
.service-meta { font-size: 22rpx; color: #6B7280; }
.notes-title { font-size: 26rpx; font-weight: 600; color: #1F2937; display: block; margin-bottom: 10rpx; }
.notes-text { font-size: 24rpx; color: #6B7280; line-height: 1.6; }
.submit-btn { margin-top: 20rpx; background: #4F46E5; color: #fff; border-radius: 14rpx; font-size: 30rpx; }
</style>
