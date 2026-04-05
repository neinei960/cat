<template>
  <SideLayout>
    <view class="page">
      <view class="hero-card">
        <view>
          <text class="hero-title">{{ isEdit ? '修改上门喂养计划' : '新建上门喂养计划' }}</text>
          <text class="hero-subtitle">先选客户和猫咪，再配置地址、日期、时间窗和服务内容。</text>
        </view>
      </view>

      <view class="section-card">
        <text class="section-title">客户与猫咪</text>
        <view class="search-row">
          <input v-model="customerKeyword" class="input" placeholder="输入客户昵称或手机号搜索" @confirm="searchCustomers" />
          <view class="mini-btn" @click="searchCustomers">搜索</view>
        </view>
        <view class="select-list" v-if="customers.length">
          <view
            v-for="customer in customers"
            :key="customer.ID"
            :class="['select-card', form.customer_id === customer.ID ? 'active' : '']"
            @click="selectCustomer(customer)"
          >
            <view>
              <text class="select-title">{{ customer.nickname || '未命名客户' }}</text>
              <text class="select-meta">{{ customer.phone || '未绑定手机号' }}</text>
            </view>
            <text class="select-mark">{{ form.customer_id === customer.ID ? '已选' : '选择' }}</text>
          </view>
        </view>
        <view v-if="selectedCustomer" class="sub-block">
          <text class="sub-title">勾选本次服务猫咪</text>
          <view class="pet-list">
            <view v-for="pet in customerPets" :key="pet.ID" :class="['pet-chip', selectedPetIds.includes(pet.ID) ? 'active' : '']" @click="togglePet(pet.ID)">
              {{ pet.name }}
            </view>
          </view>
        </view>
      </view>

      <view class="section-card">
        <text class="section-title">地址与联系人</text>
        <view class="field-grid">
          <view class="field-card full">
            <text class="field-label">服务地址</text>
            <textarea v-model="form.address_snapshot.address" class="textarea short" placeholder="填写上门地址" />
          </view>
          <view class="field-card">
            <text class="field-label">门禁/入户方式</text>
            <input v-model="form.address_snapshot.door_code" class="input" placeholder="例如：门锁密码/门卡位置" />
          </view>
          <view class="field-card">
            <text class="field-label">补充说明</text>
            <input v-model="form.address_snapshot.detail" class="input" placeholder="例如：停车、楼栋、注意事项" />
          </view>
          <view class="field-card">
            <text class="field-label">联系人</text>
            <input v-model="form.contact_name" class="input" placeholder="例如：可乐妈妈" />
          </view>
          <view class="field-card">
            <text class="field-label">联系电话</text>
            <input v-model="form.contact_phone" class="input" placeholder="例如：13800000000" type="number" />
          </view>
        </view>
      </view>

      <view class="section-card">
        <text class="section-title">服务日期</text>
        <view class="field-grid">
          <picker mode="date" :value="form.start_date" @change="form.start_date = $event.detail.value">
            <view class="picker-card">
              <text class="field-label">开始日期</text>
              <text class="picker-value">{{ form.start_date || '请选择' }}</text>
            </view>
          </picker>
          <picker mode="date" :value="form.end_date" @change="form.end_date = $event.detail.value">
            <view class="picker-card">
              <text class="field-label">结束日期</text>
              <text class="picker-value">{{ form.end_date || '请选择' }}</text>
            </view>
          </picker>
        </view>
      </view>

      <view class="section-card">
        <text class="section-title">每周时间窗</text>
        <view class="rule-list">
          <view class="rule-row" v-for="weekday in feedingWeekdays" :key="weekday.value">
            <text class="weekday">{{ weekday.label }}</text>
            <view class="window-counter-list">
              <view
                v-for="windowItem in feedingWindows"
                :key="windowItem.value"
                :class="['window-counter', getRuleCount(weekday.value, windowItem.value) > 0 ? 'active' : '']"
                @click="cycleRule(weekday.value, windowItem.value)"
              >
                <text>{{ windowItem.label }}</text>
                <text class="counter">{{ getRuleCount(weekday.value, windowItem.value) }}</text>
              </view>
            </view>
          </view>
        </view>
      </view>

      <view class="section-card">
        <text class="section-title">服务内容</text>
        <view class="item-list">
          <view
            v-for="item in settings.items"
            :key="item.code"
            :class="['item-chip', form.item_codes.includes(item.code) ? 'active' : '']"
            @click="toggleItem(item.code)"
          >
            <text>{{ item.name }}</text>
            <text v-if="item.extra_price > 0" class="item-price">+¥{{ item.extra_price }}</text>
          </view>
        </view>
      </view>

      <view class="section-card">
        <text class="section-title">备注</text>
        <textarea v-model="form.remark" class="textarea" placeholder="填写猫咪应激点、喂药要求、禁区和紧急说明" />
      </view>

      <view class="summary-card">
        <text class="summary-title">预估信息</text>
        <text class="summary-text">已选猫咪 {{ selectedPetIds.length }} 只 · 已选服务 {{ form.item_codes.length }} 项 · 预计基础价 ¥{{ estimatedAmount.toFixed(2) }}</text>
      </view>

      <view class="footer-bar">
        <button class="submit-btn" :loading="submitting" @click="submit">{{ isEdit ? '保存修改' : '创建计划' }}</button>
      </view>
    </view>
  </SideLayout>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { onLoad } from '@dcloudio/uni-app'
import SideLayout from '@/components/SideLayout.vue'
import { getCustomerList, getCustomerPets, getCustomer } from '@/api/customer'
import { createFeedingPlan, getFeedingPlan, getFeedingSettings, updateFeedingPlan } from '@/api/feeding'
import { feedingWeekdays, feedingWindows, parseFeedingAddress, parseFeedingSelectedItems } from '@/utils/feeding'

const id = ref(0)
const isEdit = computed(() => id.value > 0)
const submitting = ref(false)
const customerKeyword = ref('')
const customers = ref<Customer[]>([])
const customerPets = ref<Pet[]>([])
const selectedCustomer = ref<Customer | null>(null)
const selectedPetIds = ref<number[]>([])
const settings = ref<FeedingSettings>({ pricing: { base_visit_price: 50, extra_pet_price: 20, holiday_surcharge: 20 }, items: [] })
const form = ref({
  customer_id: 0,
  address_snapshot: { address: '', detail: '', door_code: '' } as FeedingAddressSnapshot,
  contact_name: '',
  contact_phone: '',
  start_date: '',
  end_date: '',
  remark: '',
  rules: [] as FeedingPlanRule[],
  item_codes: [] as string[],
})

const estimatedAmount = computed(() => {
  const extraCats = Math.max(selectedPetIds.value.length - 1, 0)
  const itemExtra = settings.value.items
    .filter(item => form.value.item_codes.includes(item.code))
    .reduce((sum, item) => sum + (item.extra_price || 0), 0)
  const visitCount = form.value.rules.reduce((sum, rule) => sum + (rule.visit_count || 0), 0)
  const single = settings.value.pricing.base_visit_price + settings.value.pricing.extra_pet_price * extraCats + itemExtra
  return visitCount * single
})

function getRuleCount(weekday: number, windowCode: string) {
  return form.value.rules.find(item => item.weekday === weekday && item.window_code === windowCode)?.visit_count || 0
}

function cycleRule(weekday: number, windowCode: string) {
  const next = (getRuleCount(weekday, windowCode) + 1) % 3
  const index = form.value.rules.findIndex(item => item.weekday === weekday && item.window_code === windowCode)
  if (index >= 0 && next === 0) {
    form.value.rules.splice(index, 1)
    return
  }
  if (index >= 0) {
    form.value.rules[index].visit_count = next
    return
  }
  form.value.rules.push({ weekday, window_code: windowCode, visit_count: 1 })
}

function toggleItem(code: string) {
  const index = form.value.item_codes.indexOf(code)
  if (index >= 0) {
    form.value.item_codes.splice(index, 1)
  } else {
    form.value.item_codes.push(code)
  }
}

async function searchCustomers() {
  const res = await getCustomerList({ page: 1, page_size: 20, keyword: customerKeyword.value.trim() || undefined })
  customers.value = res.data?.list || []
}

async function selectCustomer(customer: Customer) {
  selectedCustomer.value = customer
  form.value.customer_id = customer.ID
  if (!form.value.contact_name) {
    form.value.contact_name = customer.nickname || ''
  }
  if (!form.value.contact_phone) {
    form.value.contact_phone = customer.phone || ''
  }
  const [petsRes, customerRes] = await Promise.all([
    getCustomerPets(customer.ID),
    getCustomer(customer.ID),
  ])
  customerPets.value = petsRes.data || []
  selectedCustomer.value = customerRes.data || customer
}

function togglePet(petId: number) {
  const index = selectedPetIds.value.indexOf(petId)
  if (index >= 0) {
    selectedPetIds.value.splice(index, 1)
  } else {
    selectedPetIds.value.push(petId)
  }
}

async function loadSettings() {
  const res = await getFeedingSettings()
  settings.value = res.data || settings.value
  if (!form.value.item_codes.length) {
    form.value.item_codes = settings.value.items.slice(0, 3).map(item => item.code)
  }
}

async function loadPlan(planId: number) {
  const res = await getFeedingPlan(planId)
  const plan = res.data
  if (!plan) return
  id.value = planId
  form.value.customer_id = plan.customer_id
  form.value.address_snapshot = parseFeedingAddress(plan.address_snapshot_json)
  form.value.contact_name = plan.contact_name || ''
  form.value.contact_phone = plan.contact_phone || ''
  form.value.start_date = plan.start_date || ''
  form.value.end_date = plan.end_date || ''
  form.value.remark = plan.remark || ''
  form.value.rules = (plan.rules || []).map(item => ({
    weekday: item.weekday,
    window_code: item.window_code,
    visit_count: item.visit_count,
  }))
  form.value.item_codes = parseFeedingSelectedItems(plan.selected_items_json).map(item => item.code)
  selectedPetIds.value = (plan.pets || []).map(item => item.pet_id)
  if (plan.customer) {
    await selectCustomer(plan.customer)
  }
}

async function submit() {
  if (!form.value.customer_id) {
    uni.showToast({ title: '请先选择客户', icon: 'none' })
    return
  }
  if (!selectedPetIds.value.length) {
    uni.showToast({ title: '请至少选择 1 只猫咪', icon: 'none' })
    return
  }
  if (!form.value.start_date || !form.value.end_date) {
    uni.showToast({ title: '请选择日期范围', icon: 'none' })
    return
  }
  if (!form.value.rules.length) {
    uni.showToast({ title: '请配置时间窗', icon: 'none' })
    return
  }
  if (!form.value.item_codes.length) {
    uni.showToast({ title: '请至少选择 1 项服务内容', icon: 'none' })
    return
  }
  submitting.value = true
  try {
    const payload = {
      ...form.value,
      pets: selectedPetIds.value.map(petId => ({ pet_id: petId })),
      rules: form.value.rules.map(item => ({
        weekday: item.weekday,
        window_code: item.window_code,
        visit_count: item.visit_count,
      })),
    }
    const res = isEdit.value
      ? await updateFeedingPlan(id.value, payload)
      : await createFeedingPlan(payload)
    const planId = res.data?.ID
    if (planId) {
      uni.redirectTo({ url: `/pages/feeding/detail?id=${planId}` })
    }
  } finally {
    submitting.value = false
  }
}

onLoad(async (options) => {
  await loadSettings()
  if (options?.id) {
    await loadPlan(Number(options.id))
  } else {
    await searchCustomers()
  }
})
</script>

<style scoped>
.page { padding: 24rpx 24rpx 180rpx; }
.hero-card, .section-card, .summary-card { background: #fff; border-radius: 22rpx; box-shadow: 0 12rpx 28rpx rgba(15, 23, 42, 0.06); padding: 24rpx; margin-bottom: 18rpx; }
.hero-title { display: block; font-size: 36rpx; font-weight: 700; color: #111827; }
.hero-subtitle { display: block; margin-top: 10rpx; font-size: 24rpx; color: #6B7280; line-height: 1.6; }
.section-title, .summary-title { display: block; font-size: 28rpx; font-weight: 700; color: #111827; margin-bottom: 18rpx; }
.search-row { display: flex; gap: 12rpx; }
.input, .textarea, .picker-card { width: 100%; min-height: 88rpx; padding: 22rpx 24rpx; background: #F8FAFC; border-radius: 18rpx; font-size: 26rpx; color: #111827; box-sizing: border-box; }
.textarea { min-height: 180rpx; }
.textarea.short { min-height: 112rpx; }
.picker-card { display: flex; flex-direction: column; justify-content: center; }
.field-grid { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 16rpx; }
.field-card.full { grid-column: 1 / -1; }
.field-label { display: block; margin-bottom: 8rpx; font-size: 22rpx; color: #6B7280; }
.picker-value { font-size: 28rpx; color: #111827; }
.mini-btn { flex: 0 0 auto; min-width: 120rpx; padding: 0 24rpx; height: 88rpx; line-height: 88rpx; text-align: center; border-radius: 18rpx; background: #4F46E5; color: #fff; font-size: 24rpx; }
.select-list { display: flex; flex-direction: column; gap: 12rpx; margin-top: 18rpx; }
.select-card { padding: 20rpx; border-radius: 18rpx; background: #F8FAFC; display: flex; justify-content: space-between; align-items: center; gap: 14rpx; }
.select-card.active { background: #EEF2FF; border: 1rpx solid #C7D2FE; }
.select-title { display: block; font-size: 28rpx; font-weight: 600; color: #111827; }
.select-meta { display: block; margin-top: 6rpx; font-size: 22rpx; color: #6B7280; }
.select-mark { font-size: 24rpx; color: #4F46E5; }
.sub-block { margin-top: 18rpx; }
.sub-title { display: block; font-size: 24rpx; color: #374151; margin-bottom: 12rpx; }
.pet-list, .item-list { display: flex; flex-wrap: wrap; gap: 12rpx; }
.pet-chip, .item-chip { padding: 14rpx 22rpx; border-radius: 999rpx; background: #F3F4F6; color: #374151; font-size: 24rpx; }
.pet-chip.active, .item-chip.active { background: #4F46E5; color: #fff; }
.item-price { margin-left: 8rpx; font-size: 20rpx; opacity: 0.9; }
.rule-list { display: flex; flex-direction: column; gap: 14rpx; }
.rule-row { padding: 18rpx; border-radius: 18rpx; background: #F8FAFC; }
.weekday { display: block; margin-bottom: 12rpx; font-size: 24rpx; color: #111827; font-weight: 600; }
.window-counter-list { display: flex; gap: 12rpx; flex-wrap: wrap; }
.window-counter { min-width: 152rpx; padding: 12rpx 18rpx; border-radius: 16rpx; background: #fff; display: flex; justify-content: space-between; align-items: center; font-size: 22rpx; color: #4B5563; border: 1rpx solid #E5E7EB; }
.window-counter.active { border-color: #4F46E5; color: #4F46E5; background: #EEF2FF; }
.counter { min-width: 32rpx; text-align: center; font-weight: 700; }
.summary-text { display: block; font-size: 24rpx; color: #6B7280; line-height: 1.7; }
.footer-bar { position: fixed; left: 24rpx; right: 24rpx; bottom: calc(32rpx + env(safe-area-inset-bottom)); z-index: 20; }
.submit-btn { width: 100%; height: 92rpx; border-radius: 999rpx; background: linear-gradient(135deg, #4F46E5, #6366F1); color: #fff; font-size: 30rpx; display: flex; align-items: center; justify-content: center; }
@media (max-width: 768px) {
  .field-grid { grid-template-columns: 1fr; }
}
</style>
