<template>
  <SideLayout>
    <view class="page">
      <view class="hero-card">
        <view>
          <text class="hero-title">新建寄养</text>
          <text class="hero-subtitle">先选客户和猫咪，再确定寄养日期与房型，最后核价开单。</text>
        </view>
        <view class="hero-badge">{{ customerMode === 'regular' ? '老客开单' : '新客开单' }}</view>
      </view>

      <view class="section-card">
        <view class="section-head">
          <view class="section-index">1</view>
          <view>
            <text class="section-title">客户与猫咪</text>
            <text class="section-desc">先确认是谁来住，再选择要寄养的猫咪。</text>
          </view>
        </view>

        <view class="mode-tabs">
          <view :class="['mode-tab', customerMode === 'regular' ? 'active' : '']" @click="customerMode = 'regular'">老客</view>
          <view :class="['mode-tab', customerMode === 'new' ? 'active' : '']" @click="customerMode = 'new'">新客</view>
        </view>

        <template v-if="customerMode === 'regular'">
          <view class="search-row">
            <input v-model="customerKeyword" class="input search-input" placeholder="输入客户昵称或手机号搜索" @confirm="searchCustomers" />
            <view class="mini-btn" @click="searchCustomers">搜索</view>
          </view>

          <view class="customer-list" v-if="customers.length > 0">
            <view
              v-for="customer in customers"
              :key="customer.ID"
              :class="['select-card', selectedCustomer?.ID === customer.ID ? 'active' : '']"
              @click="selectCustomer(customer)"
            >
              <view>
                <text class="select-title">{{ customer.nickname || '未命名客户' }}</text>
                <text class="select-meta">{{ customer.phone || '未绑定手机号' }}</text>
              </view>
              <text class="select-mark">{{ selectedCustomer?.ID === customer.ID ? '已选' : '选择' }}</text>
            </view>
          </view>

          <view v-if="selectedCustomer" class="sub-block">
            <view class="sub-head">
              <text class="sub-title">已选客户</text>
              <text class="sub-value">{{ selectedCustomer.nickname || selectedCustomer.phone }}</text>
            </view>
            <view class="pet-check-list">
              <view class="pet-chip" v-for="pet in customerPets" :key="pet.ID" @click="togglePet(pet.ID)">
                <view :class="['pet-chip-inner', selectedPetIds.includes(pet.ID) ? 'active' : '']">
                  <text class="pet-chip-name">{{ pet.name }}</text>
                  <text class="pet-chip-mark">{{ selectedPetIds.includes(pet.ID) ? '已选' : '点选' }}</text>
                </view>
              </view>
            </view>
          </view>
        </template>

        <template v-else>
          <view class="field-grid single">
            <view class="field-card">
              <text class="field-label">客户昵称</text>
              <input v-model="newCustomer.nickname" class="input no-gap" placeholder="例如：可乐妈妈" />
            </view>
            <view class="field-card">
              <text class="field-label">手机号</text>
              <text class="field-tip">可选，不填也能先开单。</text>
              <input v-model="newCustomer.phone" class="input no-gap" placeholder="例如：13800000000" type="number" />
            </view>
          </view>

          <view class="sub-block">
            <view class="sub-head">
              <text class="sub-title">猫咪名单</text>
              <view class="mini-btn ghost" @click="addPetDraft">+ 添加猫咪</view>
            </view>

            <view class="draft-list">
              <view class="draft-card" v-for="(pet, index) in newPets" :key="pet.id">
                <view class="draft-top">
                  <text class="draft-title">猫咪 {{ index + 1 }}</text>
                  <view class="draft-del" v-if="newPets.length > 1" @click="removePetDraft(pet.id)">删除</view>
                </view>
                <view class="field-grid">
                  <view class="field-card">
                    <text class="field-label">猫咪名称</text>
                    <input v-model="pet.name" class="input no-gap" :placeholder="`例如：汤圆`" />
                  </view>
                  <view class="field-card">
                    <text class="field-label">品种</text>
                    <input v-model="pet.breed" class="input no-gap" placeholder="可选，例如：英短蓝猫" />
                  </view>
                </view>
              </view>
            </view>
          </view>
        </template>
      </view>

      <view class="section-card">
        <view class="section-head">
          <view class="section-index">2</view>
          <view>
            <text class="section-title">寄养安排</text>
            <text class="section-desc">确定日期后再查可用房型，系统会自动按库存扣减。</text>
          </view>
        </view>

        <view class="field-grid">
          <picker mode="date" :value="form.checkInAt" @change="form.checkInAt = $event.detail.value">
            <view class="picker-card">
              <text class="field-label">入住日期</text>
              <text class="picker-value">{{ form.checkInAt || '请选择日期' }}</text>
            </view>
          </picker>
          <picker mode="date" :value="form.checkOutAt" @change="form.checkOutAt = $event.detail.value">
            <view class="picker-card">
              <text class="field-label">离店日期</text>
              <text class="picker-value">{{ form.checkOutAt || '请选择日期' }}</text>
            </view>
          </picker>
        </view>

        <view class="summary-row">
          <view class="summary-pill">
            <text class="summary-label">入住猫咪</text>
            <text class="summary-value">{{ petCount }} 只</text>
          </view>
          <view class="summary-pill" v-if="stayText">
            <text class="summary-label">寄养时长</text>
            <text class="summary-value">{{ stayText }}</text>
          </view>
        </view>

        <view class="field-card remark-card">
          <text class="field-label">备注</text>
          <text class="field-tip">写饮食习惯、禁忌、应激情况、注意事项。</text>
          <textarea v-model="form.remark" class="textarea no-gap" placeholder="例如：晚上 9 点喂冻干，胆小，怕吹风机，不和陌生猫同放。" />
        </view>

        <view class="primary-action" @click="loadAvailableCabinets">查询可用房型</view>

        <view class="cabinet-list" v-if="availableCabinets.length > 0">
          <view
            v-for="cabinet in availableCabinets"
            :key="cabinet.ID"
            :class="['room-card', form.cabinetId === cabinet.ID ? 'active' : '']"
            @click="form.cabinetId = cabinet.ID"
          >
            <view class="room-head">
              <text class="room-title">{{ cabinet.cabinet_type }}</text>
              <text class="room-mark">{{ form.cabinetId === cabinet.ID ? '已选择' : '可选' }}</text>
            </view>
            <view class="room-tags">
              <text class="room-tag stock">剩 {{ cabinet.remaining_rooms || 0 }}/{{ cabinet.room_count || 1 }} 间</text>
              <text class="room-tag">每间 {{ cabinet.capacity }} 只</text>
              <text class="room-tag price">¥{{ cabinet.base_price }}/晚</text>
            </view>
            <text v-if="cabinet.remark" class="room-remark">{{ cabinet.remark }}</text>
          </view>
        </view>
      </view>

      <view class="section-card" v-if="policies.length > 0">
        <view class="section-head">
          <view class="section-index">3</view>
          <view>
            <text class="section-title">优惠策略</text>
            <text class="section-desc">默认全选可用优惠，你也可以手动取消。</text>
          </view>
        </view>

        <view class="policy-list">
          <view class="policy-card" v-for="policy in policies" :key="policy.ID" @click="togglePolicy(policy.ID)">
            <view>
              <text class="policy-name">{{ policy.name }}</text>
              <text class="policy-meta">{{ describePolicy(policy) }}</text>
            </view>
            <text :class="['policy-toggle', selectedPolicyIds.includes(policy.ID) ? 'active' : '']">
              {{ selectedPolicyIds.includes(policy.ID) ? '已用' : '不用' }}
            </text>
          </view>
        </view>
      </view>

      <view class="section-card preview-card">
        <view class="section-head">
          <view class="section-index accent">4</view>
          <view>
            <text class="section-title">金额预览</text>
            <text class="section-desc">先核价，再生成寄养单。</text>
          </view>
        </view>

        <view class="preview-trigger" @click="loadPreview">
          <text class="preview-trigger-text">{{ previewLoading ? '正在计算...' : '生成金额预览' }}</text>
        </view>

        <view v-if="preview" class="preview-box">
          <view class="preview-top">
            <view>
              <text class="preview-main">¥{{ preview.pay_amount.toFixed(2) }}</text>
              <text class="preview-sub">共 {{ preview.nights }} 晚，普通日 {{ preview.regular_nights }} 晚，节假日 {{ preview.holiday_nights }} 晚</text>
            </view>
            <view class="preview-chip">{{ selectedCabinetName || '未选房型' }}</view>
          </view>

          <view class="line-list">
            <view class="line-row" v-for="line in preview.lines" :key="`${line.type}-${line.label}`">
              <text class="line-name">{{ line.label }}</text>
              <text class="line-amount">¥{{ line.amount.toFixed(2) }}</text>
            </view>
          </view>
        </view>
      </view>

      <view class="footer-bar">
        <view class="footer-meta">
          <text class="footer-title">{{ selectedCabinetName || '还没选房型' }}</text>
          <text class="footer-sub">{{ petCount }} 只猫{{ stayText ? ` · ${stayText}` : '' }}</text>
        </view>
        <button class="submit-btn" :loading="submitting" @click="submit">确认创建</button>
      </view>
    </view>
  </SideLayout>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import SideLayout from '@/components/SideLayout.vue'
import { getCustomerList, createCustomer, getCustomerPets } from '@/api/customer'
import { createPet } from '@/api/pet'
import {
  createBoardingOrder,
  getAvailableBoardingCabinets,
  getBoardingPolicies,
  previewBoardingOrder,
} from '@/api/boarding'

type CustomerMode = 'regular' | 'new'
type DraftPet = { id: number; name: string; breed: string }

const customerMode = ref<CustomerMode>('regular')
const customerKeyword = ref('')
const customers = ref<Customer[]>([])
const selectedCustomer = ref<Customer | null>(null)
const customerPets = ref<Pet[]>([])
const selectedPetIds = ref<number[]>([])
const newCustomer = ref({ nickname: '', phone: '' })
const newPets = ref<DraftPet[]>([{ id: Date.now(), name: '', breed: '' }])

const form = ref({
  checkInAt: '',
  checkOutAt: '',
  cabinetId: 0,
  remark: '',
})

const preview = ref<BoardingPricePreview | null>(null)
const previewLoading = ref(false)
const submitting = ref(false)
const availableCabinets = ref<BoardingCabinet[]>([])
const policies = ref<BoardingDiscountPolicy[]>([])
const selectedPolicyIds = ref<number[]>([])

const petCount = computed(() => (
  customerMode.value === 'regular'
    ? selectedPetIds.value.length
    : newPets.value.filter((pet) => pet.name.trim()).length
))

const selectedCabinetName = computed(() => (
  availableCabinets.value.find((item) => item.ID === form.value.cabinetId)?.cabinet_type || ''
))

const stayText = computed(() => {
  if (!form.value.checkInAt || !form.value.checkOutAt) return ''
  const start = new Date(form.value.checkInAt)
  const end = new Date(form.value.checkOutAt)
  const diff = end.getTime() - start.getTime()
  if (Number.isNaN(diff) || diff <= 0) return ''
  return `${Math.round(diff / (24 * 60 * 60 * 1000))} 晚`
})

async function searchCustomers() {
  if (!customerKeyword.value.trim()) {
    uni.showToast({ title: '请输入关键词', icon: 'none' })
    return
  }
  const res = await getCustomerList({ page: 1, page_size: 20, keyword: customerKeyword.value.trim() })
  customers.value = res.data.list || []
}

async function selectCustomer(customer: Customer) {
  selectedCustomer.value = customer
  selectedPetIds.value = []
  const res = await getCustomerPets(customer.ID)
  customerPets.value = res.data || []
}

function togglePet(petId: number) {
  const idx = selectedPetIds.value.indexOf(petId)
  if (idx >= 0) selectedPetIds.value.splice(idx, 1)
  else selectedPetIds.value.push(petId)
}

function addPetDraft() {
  newPets.value.push({ id: Date.now() + Math.floor(Math.random() * 1000), name: '', breed: '' })
}

function removePetDraft(id: number) {
  newPets.value = newPets.value.filter((pet) => pet.id !== id)
}

function togglePolicy(id: number) {
  const idx = selectedPolicyIds.value.indexOf(id)
  if (idx >= 0) selectedPolicyIds.value.splice(idx, 1)
  else selectedPolicyIds.value.push(id)
}

function describePolicy(policy: BoardingDiscountPolicy) {
  try {
    const rule = JSON.parse(policy.rule_json || '{}')
    if (policy.policy_type === 'stay_n_free_m') {
      return `住 ${rule.stay || 0} 免 ${rule.free || 0}`
    }
    if (policy.policy_type === 'holiday_surcharge') {
      return `节假日每晚 +¥${rule.surcharge || 0}`
    }
  } catch {}
  return policy.remark || '寄养优惠'
}

function validateBeforePrice() {
  if (!form.value.checkInAt || !form.value.checkOutAt) {
    uni.showToast({ title: '请选择入住和离店日期', icon: 'none' })
    return false
  }
  if (petCount.value < 1) {
    uni.showToast({ title: '请至少选择一只猫咪', icon: 'none' })
    return false
  }
  return true
}

async function ensurePoliciesLoaded() {
  if (policies.value.length > 0) return
  const res = await getBoardingPolicies()
  policies.value = (res.data || []).filter((item) => item.status === 1)
  selectedPolicyIds.value = policies.value.map((item) => item.ID)
}

async function loadAvailableCabinets() {
  if (!validateBeforePrice()) return
  await ensurePoliciesLoaded()
  const res = await getAvailableBoardingCabinets({
    check_in_at: form.value.checkInAt,
    check_out_at: form.value.checkOutAt,
    pet_count: petCount.value,
  })
  availableCabinets.value = res.data || []
  if (!availableCabinets.value.find((item) => item.ID === form.value.cabinetId)) {
    form.value.cabinetId = availableCabinets.value[0]?.ID || 0
  }
}

async function loadPreview() {
  if (!validateBeforePrice()) return
  if (!form.value.cabinetId) {
    uni.showToast({ title: '请选择寄养房型', icon: 'none' })
    return
  }
  previewLoading.value = true
  try {
    const payload: any = {
      cabinet_id: form.value.cabinetId,
      check_in_at: form.value.checkInAt,
      check_out_at: form.value.checkOutAt,
      policy_ids: selectedPolicyIds.value,
    }
    if (customerMode.value === 'regular') {
      if (!selectedCustomer.value || selectedPetIds.value.length === 0) {
        uni.showToast({ title: '请选择客户和猫咪', icon: 'none' })
        return
      }
      payload.customer_id = selectedCustomer.value.ID
      payload.pet_ids = selectedPetIds.value
    } else {
      payload.pet_count = petCount.value
    }
    const res = await previewBoardingOrder(payload)
    preview.value = res.data
  } finally {
    previewLoading.value = false
  }
}

async function submit() {
  if (!preview.value) {
    await loadPreview()
    if (!preview.value) return
  }
  submitting.value = true
  try {
    let customerId = selectedCustomer.value?.ID || 0
    let petIds = [...selectedPetIds.value]
    if (customerMode.value === 'new') {
      if (!newCustomer.value.nickname.trim()) {
        uni.showToast({ title: '请填写客户昵称', icon: 'none' })
        return
      }
      const draftPets = newPets.value.filter((pet) => pet.name.trim())
      if (draftPets.length === 0) {
        uni.showToast({ title: '请至少填写一只猫咪', icon: 'none' })
        return
      }
      const customerRes = await createCustomer({
        nickname: newCustomer.value.nickname.trim(),
        phone: newCustomer.value.phone.trim(),
      })
      customerId = customerRes.data.ID
      petIds = []
      for (const pet of draftPets) {
        const petRes = await createPet({
          customer_id: customerId,
          name: pet.name.trim(),
          breed: pet.breed.trim(),
          species: '猫',
        })
        petIds.push(petRes.data.ID)
      }
    }
    const res = await createBoardingOrder({
      customer_id: customerId,
      pet_ids: petIds,
      cabinet_id: form.value.cabinetId,
      check_in_at: form.value.checkInAt,
      check_out_at: form.value.checkOutAt,
      policy_ids: selectedPolicyIds.value,
      remark: form.value.remark,
    })
    uni.showToast({ title: '寄养单已创建', icon: 'success' })
    setTimeout(() => {
      uni.redirectTo({ url: `/pages/boarding/detail?id=${res.data.ID}` })
    }, 400)
  } finally {
    submitting.value = false
  }
}

ensurePoliciesLoaded()
</script>

<style scoped>
.page {
  padding: 24rpx 24rpx 190rpx;
  display: flex;
  flex-direction: column;
  gap: 20rpx;
  background:
    radial-gradient(circle at top left, rgba(253, 230, 138, 0.38), transparent 34%),
    linear-gradient(180deg, #fffdf8 0%, #f7f8ff 54%, #f9fafb 100%);
}
.hero-card,
.section-card {
  background: rgba(255, 255, 255, 0.94);
  border: 1rpx solid rgba(226, 232, 240, 0.9);
  border-radius: 28rpx;
  padding: 24rpx;
  box-shadow: 0 18rpx 44rpx rgba(15, 23, 42, 0.06);
}
.hero-card {
  display: flex;
  justify-content: space-between;
  gap: 20rpx;
  align-items: flex-start;
}
.hero-title {
  display: block;
  font-size: 38rpx;
  font-weight: 700;
  color: #111827;
}
.hero-subtitle {
  display: block;
  margin-top: 10rpx;
  font-size: 23rpx;
  line-height: 1.65;
  color: #6b7280;
}
.hero-badge {
  padding: 12rpx 20rpx;
  border-radius: 999rpx;
  background: linear-gradient(135deg, #f59e0b, #ef4444);
  color: #fff;
  font-size: 22rpx;
  white-space: nowrap;
}
.section-head {
  display: flex;
  gap: 16rpx;
  align-items: flex-start;
  margin-bottom: 20rpx;
}
.section-index {
  width: 42rpx;
  height: 42rpx;
  border-radius: 50%;
  background: #eef2ff;
  color: #4f46e5;
  font-size: 24rpx;
  font-weight: 700;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  margin-top: 2rpx;
}
.section-index.accent {
  background: #fef3c7;
  color: #d97706;
}
.section-title {
  display: block;
  font-size: 30rpx;
  font-weight: 700;
  color: #111827;
}
.section-desc {
  display: block;
  margin-top: 6rpx;
  font-size: 22rpx;
  line-height: 1.6;
  color: #6b7280;
}
.mode-tabs {
  display: inline-flex;
  gap: 10rpx;
  padding: 8rpx;
  border-radius: 999rpx;
  background: #f3f4f6;
  margin-bottom: 20rpx;
}
.mode-tab {
  min-width: 116rpx;
  padding: 14rpx 22rpx;
  border-radius: 999rpx;
  text-align: center;
  font-size: 24rpx;
  color: #6b7280;
}
.mode-tab.active {
  background: linear-gradient(135deg, #4f46e5, #6366f1);
  color: #fff;
  box-shadow: 0 8rpx 18rpx rgba(79, 70, 229, 0.24);
}
.search-row {
  display: flex;
  gap: 12rpx;
  align-items: stretch;
}
.field-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 14rpx;
}
.field-grid.single {
  grid-template-columns: repeat(1, minmax(0, 1fr));
}
.field-card,
.picker-card {
  padding: 24rpx;
  background: #fbfcff;
  border: 1rpx solid #e5e7eb;
  border-radius: 20rpx;
}
.field-label {
  display: block;
  font-size: 24rpx;
  font-weight: 600;
  color: #374151;
}
.field-tip {
  display: block;
  margin-top: 6rpx;
  font-size: 22rpx;
  line-height: 1.5;
  color: #9ca3af;
}
.input,
.textarea {
  display: block;
  width: 100%;
  box-sizing: border-box;
  margin-top: 16rpx;
  padding: 18rpx 20rpx;
  border: 2rpx solid #dbe3f0;
  border-radius: 18rpx;
  background: #fff;
  font-size: 29rpx;
  line-height: 1.4;
  min-height: 88rpx;
  height: 88rpx;
  color: #111827;
}
.no-gap {
  margin-top: 12rpx;
}
.textarea {
  min-height: 196rpx;
  height: 196rpx;
  padding: 18rpx 20rpx;
  line-height: 1.7;
}
.search-input {
  margin-top: 0;
  flex: 1;
}
.mini-btn {
  min-width: 108rpx;
  padding: 0 22rpx;
  border-radius: 18rpx;
  background: linear-gradient(135deg, #4f46e5, #6366f1);
  color: #fff;
  font-size: 24rpx;
  display: flex;
  align-items: center;
  justify-content: center;
}
.mini-btn.ghost {
  min-width: auto;
  padding: 12rpx 18rpx;
  background: #eef2ff;
  color: #4f46e5;
}
.sub-block {
  margin-top: 18rpx;
  padding: 18rpx;
  border-radius: 22rpx;
  background: linear-gradient(180deg, #fcfcff, #f8fafc);
  border: 1rpx solid #eef2f7;
}
.sub-head {
  display: flex;
  justify-content: space-between;
  gap: 16rpx;
  align-items: center;
  margin-bottom: 14rpx;
}
.sub-title {
  font-size: 24rpx;
  font-weight: 600;
  color: #111827;
}
.sub-value {
  font-size: 22rpx;
  color: #6b7280;
}
.customer-list,
.cabinet-list,
.policy-list,
.draft-list,
.line-list,
.pet-check-list {
  display: flex;
  flex-direction: column;
  gap: 12rpx;
}
.select-card,
.policy-card {
  display: flex;
  justify-content: space-between;
  gap: 16rpx;
  align-items: center;
  padding: 18rpx 20rpx;
  border-radius: 20rpx;
  border: 1rpx solid #e5e7eb;
  background: #fafbff;
}
.select-card.active {
  border-color: #818cf8;
  background: linear-gradient(135deg, #eef2ff, #f8faff);
}
.select-title,
.policy-name {
  display: block;
  font-size: 26rpx;
  font-weight: 600;
  color: #111827;
}
.select-meta,
.policy-meta {
  display: block;
  margin-top: 6rpx;
  font-size: 22rpx;
  color: #6b7280;
}
.select-mark,
.policy-toggle {
  padding: 10rpx 16rpx;
  border-radius: 999rpx;
  background: #eef2ff;
  color: #4f46e5;
  font-size: 22rpx;
  white-space: nowrap;
}
.policy-toggle.active {
  background: #4f46e5;
  color: #fff;
}
.pet-chip {
  width: 100%;
}
.pet-chip-inner {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 18rpx 20rpx;
  border-radius: 18rpx;
  background: #fff;
  border: 1rpx solid #e5e7eb;
}
.pet-chip-inner.active {
  background: linear-gradient(135deg, #fef3c7, #fff7ed);
  border-color: #f59e0b;
}
.pet-chip-name {
  font-size: 25rpx;
  font-weight: 600;
  color: #111827;
}
.pet-chip-mark {
  font-size: 22rpx;
  color: #92400e;
}
.draft-card {
  padding: 18rpx;
  border-radius: 22rpx;
  background: #fffdf8;
  border: 1rpx solid #f3e8d2;
}
.draft-top {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 14rpx;
}
.draft-title {
  font-size: 24rpx;
  font-weight: 600;
  color: #111827;
}
.draft-del {
  font-size: 22rpx;
  color: #dc2626;
}
.picker-card {
  min-height: 138rpx;
}
.picker-value {
  display: block;
  margin-top: 14rpx;
  font-size: 30rpx;
  font-weight: 700;
  color: #111827;
}
.summary-row {
  display: flex;
  flex-wrap: wrap;
  gap: 12rpx;
  margin: 16rpx 0 18rpx;
}
.summary-pill {
  padding: 14rpx 18rpx;
  border-radius: 18rpx;
  background: #f8fafc;
  border: 1rpx solid #e5e7eb;
}
.summary-label {
  font-size: 20rpx;
  color: #6b7280;
}
.summary-value {
  margin-left: 8rpx;
  font-size: 24rpx;
  font-weight: 600;
  color: #111827;
}
.remark-card {
  margin-top: 4rpx;
}
.primary-action,
.preview-trigger {
  margin-top: 18rpx;
  width: 100%;
  min-height: 88rpx;
  border-radius: 20rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 28rpx;
  font-weight: 700;
  color: #fff;
  background: linear-gradient(135deg, #4f46e5, #6366f1);
  box-shadow: 0 14rpx 28rpx rgba(79, 70, 229, 0.22);
}
.room-card {
  padding: 20rpx;
  border-radius: 22rpx;
  border: 1rpx solid #e5e7eb;
  background: #fff;
}
.room-card.active {
  border-color: #6366f1;
  background: linear-gradient(180deg, #f5f7ff, #ffffff);
  box-shadow: 0 12rpx 24rpx rgba(99, 102, 241, 0.12);
}
.room-head {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 16rpx;
}
.room-title {
  font-size: 28rpx;
  font-weight: 700;
  color: #111827;
}
.room-mark {
  font-size: 22rpx;
  color: #4f46e5;
}
.room-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 10rpx;
  margin-top: 14rpx;
}
.room-tag {
  padding: 8rpx 14rpx;
  border-radius: 999rpx;
  background: #f3f4f6;
  font-size: 21rpx;
  color: #4b5563;
}
.room-tag.stock {
  background: #ecfdf5;
  color: #047857;
}
.room-tag.price {
  background: #fff7ed;
  color: #c2410c;
}
.room-remark {
  display: block;
  margin-top: 14rpx;
  font-size: 22rpx;
  line-height: 1.55;
  color: #6b7280;
}
.preview-card {
  background: linear-gradient(180deg, rgba(255, 255, 255, 0.96), rgba(248, 250, 255, 0.98));
}
.preview-trigger-text {
  color: inherit;
}
.preview-box {
  margin-top: 18rpx;
  padding: 22rpx;
  border-radius: 24rpx;
  background: linear-gradient(135deg, #312e81, #4338ca 58%, #4f46e5);
  color: #fff;
}
.preview-top {
  display: flex;
  justify-content: space-between;
  gap: 16rpx;
  align-items: flex-start;
}
.preview-main {
  display: block;
  font-size: 42rpx;
  font-weight: 700;
  color: #fff;
}
.preview-sub {
  display: block;
  margin-top: 10rpx;
  font-size: 22rpx;
  line-height: 1.6;
  color: rgba(255, 255, 255, 0.82);
}
.preview-chip {
  padding: 12rpx 18rpx;
  border-radius: 999rpx;
  background: rgba(255, 255, 255, 0.16);
  font-size: 22rpx;
  white-space: nowrap;
}
.line-list {
  margin-top: 18rpx;
}
.line-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 16rpx;
  padding: 14rpx 0;
  border-bottom: 1rpx solid rgba(255, 255, 255, 0.14);
}
.line-row:last-child {
  border-bottom: none;
}
.line-name,
.line-amount {
  font-size: 23rpx;
  color: #fff;
}
.footer-bar {
  position: fixed;
  left: 20rpx;
  right: 20rpx;
  bottom: calc(env(safe-area-inset-bottom) + 18rpx);
  z-index: 20;
  display: flex;
  align-items: center;
  gap: 16rpx;
  padding: 16rpx 18rpx;
  border-radius: 24rpx;
  background: rgba(17, 24, 39, 0.92);
  box-shadow: 0 18rpx 38rpx rgba(15, 23, 42, 0.28);
}
.footer-meta {
  flex: 1;
  min-width: 0;
}
.footer-title {
  display: block;
  font-size: 26rpx;
  font-weight: 700;
  color: #fff;
}
.footer-sub {
  display: block;
  margin-top: 6rpx;
  font-size: 22rpx;
  color: rgba(255, 255, 255, 0.72);
}
.submit-btn {
  margin: 0;
  min-width: 220rpx;
  height: 82rpx;
  line-height: 82rpx;
  border-radius: 18rpx;
  font-size: 28rpx;
  font-weight: 700;
  color: #fff;
  background: linear-gradient(135deg, #f59e0b, #ef4444);
}
@media (max-width: 768px) {
  .page {
    padding: 20rpx 20rpx 190rpx;
    gap: 16rpx;
  }
  .hero-card {
    flex-direction: column;
  }
  .hero-title {
    font-size: 34rpx;
  }
  .hero-subtitle,
  .section-desc {
    font-size: 24rpx;
  }
  .field-grid,
  .field-grid.single {
    grid-template-columns: repeat(1, minmax(0, 1fr));
  }
  .field-card,
  .picker-card,
  .draft-card,
  .sub-block {
    padding: 22rpx;
  }
  .search-row {
    flex-direction: column;
  }
  .mini-btn {
    min-height: 76rpx;
  }
  .input,
  .textarea {
    font-size: 30rpx;
  }
}
</style>
