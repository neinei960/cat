<template>
  <SideLayout>
  <view class="page">
    <!-- 猫咪选择 -->
    <view class="section">
      <text class="section-title">选择猫咪</text>
      <view class="search-bar" v-if="!selectedPet">
        <input v-model="petKeyword" placeholder="输入猫咪名搜索" class="search-input" @confirm="searchPets" @input="searchPets" />
      </view>
      <view class="option-list" v-if="!selectedPet && petList.length > 0">
        <view class="pet-card" v-for="p in petList" :key="p.ID" @click="selectPet(p)">
          <view class="pet-card-row1">
            <text class="pet-card-name">{{ p.name }}</text>
            <text class="pet-card-owner" v-if="p.customer?.nickname">{{ p.customer.nickname }}</text>
            <text class="pet-card-owner" v-else>散客</text>
          </view>
          <view class="pet-card-row2">
            <text class="pet-card-info">{{ p.breed || '未知品种' }}</text>
            <text class="pet-card-dot">·</text>
            <text class="pet-card-info">{{ p.gender === 1 ? '公' : p.gender === 2 ? '母' : '未知' }}</text>
            <text class="pet-card-dot" v-if="p.birth_date">·</text>
            <text class="pet-card-info" v-if="p.birth_date">{{ calcAge(p.birth_date) }}</text>
          </view>
          <view class="pet-card-row3" v-if="p.fur_level || p.personality || (p.aggression && p.aggression !== '无')">
            <text class="pet-label" v-if="p.fur_level">{{ p.fur_level }}</text>
            <text class="pet-label" v-if="p.personality" :style="{ background: getPersonalityBg(p.personality), color: getPersonalityColor(p.personality) }">{{ p.personality }}</text>
            <text class="pet-label warn" v-if="p.aggression && p.aggression !== '无'">{{ p.aggression }}</text>
            <text class="pet-label" v-if="p.customer?.nickname">主人:{{ p.customer.nickname }}</text>
          </view>
        </view>
      </view>
      <view v-if="selectedPet" class="selected-pet">
        <view class="pet-header">
          <text class="pet-name-lg">{{ selectedPet.name }}</text>
          <text class="change-btn" @click="clearPet">更换</text>
        </view>
        <view class="pet-tags">
          <text class="tag" v-if="selectedPet.fur_level">{{ selectedPet.fur_level }}</text>
          <text class="tag" v-if="selectedPet.personality">{{ selectedPet.personality }}</text>
          <text class="tag warn" v-if="selectedPet.aggression && selectedPet.aggression !== '无'">攻击性:{{ selectedPet.aggression }}</text>
        </view>
        <view class="pet-alert" v-if="selectedPet.care_notes">
          <text class="alert-icon">!</text>
          <text class="alert-text">{{ selectedPet.care_notes }}</text>
        </view>
        <view class="pet-alert" v-if="selectedPet.forbidden_zones">
          <text class="alert-icon">x</text>
          <text class="alert-text">禁区: {{ selectedPet.forbidden_zones }}</text>
        </view>
      </view>
    </view>

    <!-- 洗浴项目 -->
    <view class="section" v-if="selectedPet">
      <text class="section-title">洗浴项目</text>
      <view class="service-picker" v-if="serviceList.length > 0">
        <view
          class="service-opt"
          :class="{ active: selectedServiceId === s.ID }"
          v-for="s in serviceList"
          :key="s.ID"
          @click="selectService(s)"
        >
          <text>{{ s.name }}</text>
          <text class="svc-price" v-if="selectedServiceId === s.ID && !selectedSpecId">¥{{ servicePrice.toFixed(2) }}</text>
        </view>
      </view>

      <!-- 规格选择 -->
      <view v-if="selectedServiceId && specList.length > 0" class="spec-section">
        <text class="spec-title">选择规格</text>
        <view class="spec-picker">
          <view
            v-for="spec in specList" :key="spec.ID"
            :class="['spec-opt', selectedSpecId === spec.ID && !useCustomPrice ? 'active' : '']"
            @click="selectSpec(spec)"
          >
            <text class="spec-name">{{ spec.name }}</text>
            <text class="spec-meta">¥{{ spec.price }} <text v-if="spec.duration" class="spec-dur">· {{ spec.duration }}分钟</text></text>
          </view>
          <view :class="['spec-opt custom', useCustomPrice ? 'active' : '']" @click="enableCustomPrice">
            <text class="spec-name">自定义</text>
            <text class="spec-meta">手动输入</text>
          </view>
        </view>
        <view v-if="useCustomPrice" class="custom-price-row">
          <text class="custom-price-label">输入金额</text>
          <input v-model="customPriceInput" type="digit" placeholder="0.00" class="custom-price-input" @input="onCustomPriceInput" />
          <text class="custom-price-hint">适用于美团团购等特殊价格</text>
        </view>
      </view>

      <view v-if="selectedServiceId && specList.length === 0" class="spec-section">
        <view class="custom-price-toggle" @click="useCustomPrice = !useCustomPrice">
          <text>{{ useCustomPrice ? '使用基础价格 ¥' + getBasePrice() : '自定义金额（美团团购等）' }}</text>
        </view>
        <view v-if="useCustomPrice" class="custom-price-row">
          <text class="custom-price-label">输入金额</text>
          <input v-model="customPriceInput" type="digit" placeholder="0.00" class="custom-price-input" @input="onCustomPriceInput" />
        </view>
      </view>
    </view>

    <!-- 附加费用 -->
    <view class="section" v-if="selectedPet">
      <view class="section-title-row">
        <text class="section-title" style="margin-bottom:0">附加费用</text>
        <view class="addon-add-btn" @click="onAddAddon">
          <text class="addon-add-icon">+</text>
        </view>
      </view>
      <view class="addon-grid">
        <view
          class="addon-item"
          :class="{ 'addon-item--deleting': longPressId === a.id }"
          v-for="a in addonInputs"
          :key="a.id"
          @longpress="onLongPress(a.id)"
        >
          <view v-if="longPressId === a.id" class="addon-delete-badge" @click.stop="onDeleteAddon(a)">
            <text class="addon-delete-icon">−</text>
          </view>
          <text class="addon-label">{{ a.name }}</text>
          <input v-model="a.amount" type="digit" placeholder="0" class="addon-input" :disabled="longPressId === a.id" />
        </view>
      </view>
      <view v-if="longPressId !== null" class="addon-cancel-hint" @click="longPressId = null">
        <text class="addon-cancel-text">点击此处取消删除模式</text>
      </view>
    </view>

    <!-- 合计 -->
    <view class="section summary" v-if="selectedPet && selectedServiceId">
      <view class="summary-row">
        <text>基础价格</text>
        <text>¥{{ servicePrice.toFixed(2) }}</text>
      </view>
      <view class="summary-row" v-if="addonTotal > 0">
        <text>附加费合计</text>
        <text>¥{{ addonTotal.toFixed(2) }}</text>
      </view>
      <view class="summary-row">
        <text>合计</text>
        <text class="bold">¥{{ totalAmount.toFixed(2) }}</text>
      </view>
      <view class="summary-row" v-if="discountRate < 1">
        <text>会员折扣 ({{ (discountRate * 10).toFixed(1) }}折)</text>
        <text class="discount">-¥{{ discountAmount.toFixed(2) }}</text>
      </view>
      <view class="summary-row total">
        <text>实付</text>
        <text class="pay-amount">¥{{ payAmount.toFixed(2) }}</text>
      </view>
    </view>

    <!-- 洗护师 + 备注 -->
    <view class="section" v-if="selectedPet">
      <view class="form-row">
        <text class="label">洗护师 <text v-if="!selectedStaff" class="required">*必选</text></text>
        <picker :range="staffNames" :value="selectedStaffIdx" @change="(e: any) => selectedStaffIdx = Number(e.detail.value)">
          <view :class="['picker', !selectedStaff ? 'picker-warn' : '']">{{ staffNames[selectedStaffIdx] || '请选择' }}</view>
        </picker>
      </view>
      <view class="staff-commission" v-if="selectedStaff">
        <text>提成比例: {{ selectedStaff.commission_rate }}%</text>
      </view>
      <view class="form-row">
        <text class="label">备注</text>
        <input v-model="remark" placeholder="备注" class="input" />
      </view>
    </view>

    <button v-if="selectedPet && selectedServiceId" class="btn-submit" @click="onSubmit" :loading="submitting">确认开单</button>
  </view>
  </SideLayout>
</template>

<script setup lang="ts">
import SideLayout from '@/components/SideLayout.vue'
import { ref, computed, onMounted } from 'vue'
import { onLoad } from '@dcloudio/uni-app'
import { getPetList } from '@/api/pet'
import { createOrder } from '@/api/order'
import { getAddonList, createAddon, deleteAddon, priceLookup } from '@/api/addon'
import { getServiceList, getPriceRules } from '@/api/service'
import { getStaffList } from '@/api/staff'
import { getCustomerCard } from '@/api/member-card'
import { safeBack } from '@/utils/navigate'
import { getPersonalityColor, getPersonalityBg } from '@/utils/personality'
import { getAppointment } from '@/api/appointment'

const petKeyword = ref('')
const petList = ref<any[]>([])
const selectedPet = ref<any>(null)
const serviceList = ref<any[]>([])
const selectedServiceId = ref(0)
const servicePrice = ref(0)
const specList = ref<any[]>([])
const selectedSpecId = ref(0)
const staffList = ref<any[]>([])
const selectedStaffIdx = ref(0)
const addonInputs = ref<{ id: number; name: string; amount: string }[]>([])
const longPressId = ref<number | null>(null)
const remark = ref('')
const submitting = ref(false)
const useCustomPrice = ref(false)
const customPriceInput = ref('')
const memberBalance = ref(0)
const prefillAppointmentId = ref(0)

const staffNames = computed(() => ['未选择', ...staffList.value.map(s => s.name)])
const selectedStaff = computed(() => {
  if (selectedStaffIdx.value === 0) return null
  return staffList.value[selectedStaffIdx.value - 1]
})

const addonTotal = computed(() =>
  addonInputs.value.reduce((s, a) => s + (parseFloat(a.amount) || 0), 0)
)
const totalAmount = computed(() => servicePrice.value + addonTotal.value)

const discountRate = computed(() => {
  if (!selectedPet.value?.customer?.discount_rate) return 1
  const r = selectedPet.value.customer.discount_rate
  return r > 0 && r < 1 ? r : 1
})
const payAmount = computed(() => Math.round(totalAmount.value * discountRate.value * 100) / 100)
const discountAmount = computed(() => totalAmount.value - payAmount.value)

onLoad((query) => {
  if (query?.appointment_id) {
    prefillAppointmentId.value = parseInt(String(query.appointment_id)) || 0
  }
})

onMounted(async () => {
  try {
    const sRes = await getServiceList({ page_size: 50 } as any)
    if (sRes.data?.list) serviceList.value = sRes.data.list
  } catch {}
  try {
    const stRes = await getStaffList({ page_size: 50 } as any)
    if (stRes.data?.list) staffList.value = stRes.data.list.filter((s: any) => s.status === 1)
  } catch {}
  await loadAddons()
  if (prefillAppointmentId.value) {
    await prefillFromAppointment(prefillAppointmentId.value)
  }
})

async function loadAddons() {
  try {
    const aRes = await getAddonList()
    if (Array.isArray(aRes.data)) {
      const amountMap = new Map(addonInputs.value.map(a => [a.id, a.amount]))
      addonInputs.value = aRes.data.map(a => ({
        id: a.ID, name: a.name, amount: amountMap.get(a.ID) || ''
      }))
    }
  } catch {
    if (addonInputs.value.length === 0) {
      addonInputs.value = [
        { id: -1, name: '超重费', amount: '' },
        { id: -2, name: '去油费', amount: '' },
        { id: -3, name: '药浴', amount: '' },
        { id: -4, name: '刷牙', amount: '' },
        { id: -5, name: '开结', amount: '' },
        { id: -6, name: '攻击费', amount: '' },
      ]
    }
  }
}

function onAddAddon() {
  uni.showModal({
    title: '添加附加费类型', editable: true, placeholderText: '输入名称',
    success: async (res) => {
      if (res.confirm && res.content?.trim()) {
        try {
          await createAddon({ name: res.content.trim(), default_price: 0, is_variable: true } as any)
          await loadAddons()
        } catch { addonInputs.value.push({ id: Date.now(), name: res.content.trim(), amount: '' }) }
      }
    }
  })
}

function onLongPress(id: number) { longPressId.value = id }

function onDeleteAddon(a: { id: number; name: string }) {
  uni.showModal({
    title: '删除附加费', content: `删除「${a.name}」？`, confirmColor: '#EF4444',
    success: async (res) => {
      if (res.confirm) {
        if (a.id > 0) { try { await deleteAddon(a.id); await loadAddons() } catch {} }
        else { addonInputs.value = addonInputs.value.filter(i => i.id !== a.id) }
      }
      longPressId.value = null
    }
  })
}

function calcAge(birthDate: string): string {
  if (!birthDate) return ''
  const birth = new Date(birthDate)
  const now = new Date()
  const months = (now.getFullYear() - birth.getFullYear()) * 12 + (now.getMonth() - birth.getMonth())
  if (months < 1) return '不到1个月'
  if (months < 12) return `${months}个月`
  const years = Math.floor(months / 12)
  const rem = months % 12
  return rem > 0 ? `${years}岁${rem}个月` : `${years}岁`
}

let searchTimer: any = null
function searchPets() {
  clearTimeout(searchTimer)
  searchTimer = setTimeout(async () => {
    if (!petKeyword.value) { petList.value = []; return }
    const res = await getPetList({ page: 1, page_size: 20, keyword: petKeyword.value } as any)
    petList.value = (res.data as any)?.list || []
  }, 300)
}

async function selectPet(p: any) {
  selectedPet.value = p
  petList.value = []
  petKeyword.value = ''
  if (selectedServiceId.value > 0 && p.fur_level) {
    lookupPrice(selectedServiceId.value, p.fur_level)
  }
  memberBalance.value = 0
  if (p.customer_id) {
    try {
      const cardRes = await getCustomerCard(p.customer_id)
      if (cardRes.data && cardRes.data.balance > 0) memberBalance.value = cardRes.data.balance
    } catch {}
  }
}

function clearPet() {
  selectedPet.value = null
  servicePrice.value = 0
  selectedServiceId.value = 0
  selectedSpecId.value = 0
  specList.value = []
  memberBalance.value = 0
  useCustomPrice.value = false
  customPriceInput.value = ''
}

async function selectService(s: any) {
  selectedServiceId.value = s.ID
  selectedSpecId.value = 0
  servicePrice.value = s.base_price
  useCustomPrice.value = false
  customPriceInput.value = ''
  try {
    const res = await getPriceRules(s.ID)
    const rules = res.data || []
    specList.value = rules.map((r: any) => ({ ...r, name: r.name || r.fur_level || r.pet_size || '规格' }))
  } catch { specList.value = [] }
}

async function prefillFromAppointment(appointmentId: number) {
  try {
    const res = await getAppointment(appointmentId)
    const appt = res.data
    const firstPetGroup = Array.isArray(appt?.pets) && appt.pets.length > 0 ? appt.pets[0] : null
    const pet = firstPetGroup?.pet || appt?.pet
    if (pet) {
      await selectPet(pet)
    }

    const firstService = firstPetGroup?.services?.[0] || appt?.services?.[0]
    if (firstService?.service_id) {
      const service = serviceList.value.find((item) => item.ID === firstService.service_id)
      if (service) {
        await selectService(service)
        if (typeof firstService.price === 'number' && firstService.price > 0) {
          servicePrice.value = firstService.price
        }
      }
    }

    if (appt?.staff_id) {
      const idx = staffList.value.findIndex((staff) => staff.ID === appt.staff_id)
      if (idx >= 0) {
        selectedStaffIdx.value = idx + 1
      }
    }

    if (appt?.notes) {
      remark.value = appt.notes
    }

    const petCount = Array.isArray(appt?.pets) ? appt.pets.length : 0
    const serviceCount = firstPetGroup?.services?.length || appt?.services?.length || 0
    if (petCount > 1 || serviceCount > 1) {
      uni.showToast({ title: '已带入预约首项信息，请核对后开单', icon: 'none' })
    }
  } catch {
    uni.showToast({ title: '预约信息带入失败', icon: 'none' })
  }
}

function selectSpec(spec: any) {
  selectedSpecId.value = spec.ID
  servicePrice.value = spec.price
  useCustomPrice.value = false
  customPriceInput.value = ''
}

function enableCustomPrice() {
  useCustomPrice.value = true
  selectedSpecId.value = 0
  if (customPriceInput.value) servicePrice.value = parseFloat(customPriceInput.value) || 0
}

function onCustomPriceInput() { servicePrice.value = parseFloat(customPriceInput.value) || 0 }

function getBasePrice(): string {
  const svc = serviceList.value.find(s => s.ID === selectedServiceId.value)
  return svc ? svc.base_price.toFixed(2) : '0'
}

async function lookupPrice(serviceId: number, furLevel: string) {
  try {
    const res = await priceLookup(serviceId, furLevel)
    servicePrice.value = res.data.price
  } catch {
    const svc = serviceList.value.find(s => s.ID === serviceId)
    servicePrice.value = svc?.base_price || 0
  }
}

async function onSubmit() {
  if (!selectedPet.value) { uni.showToast({ title: '请选择猫咪', icon: 'none' }); return }
  if (!selectedServiceId.value) { uni.showToast({ title: '请选择洗浴项目', icon: 'none' }); return }
  if (!selectedStaff.value) { uni.showToast({ title: '请选择洗护师', icon: 'none' }); return }

  submitting.value = true
  try {
    const addons = addonInputs.value
      .filter(a => parseFloat(a.amount) > 0)
      .map(a => ({ name: a.name, amount: parseFloat(a.amount) }))

    const orderRes = await createOrder({
      pet_id: selectedPet.value.ID,
      customer_id: selectedPet.value.customer_id || undefined,
      service_id: selectedServiceId.value,
      staff_id: selectedStaff.value?.ID || undefined,
      addons,
      remark: remark.value,
    } as any)

    uni.showToast({ title: '开单成功', icon: 'success' })
    const orderId = orderRes.data?.ID
    setTimeout(() => uni.redirectTo({ url: `/pages/order/detail?id=${orderId}` }), 500)
  } catch (e: any) {
    uni.showToast({ title: e.message || '开单失败', icon: 'none' })
  } finally { submitting.value = false }
}
</script>

<style scoped>
.page { padding: 24rpx; }
.section { background: #fff; border-radius: 16rpx; padding: 24rpx; margin-bottom: 16rpx; }
.section-title { font-size: 28rpx; font-weight: 600; color: #1F2937; display: block; margin-bottom: 16rpx; }
.section-title-row { display: flex; justify-content: space-between; align-items: center; margin-bottom: 16rpx; }
.search-bar { margin-bottom: 12rpx; }
.search-input { background: #F3F4F6; border-radius: 12rpx; padding: 16rpx 24rpx; font-size: 28rpx; }
.option-list { display: flex; flex-direction: column; gap: 8rpx; max-height: 400rpx; overflow-y: auto; }
.pet-name { font-size: 28rpx; color: #1F2937; font-weight: 500; }
.pet-info { font-size: 24rpx; color: #6B7280; margin-left: 16rpx; }
.pet-card { background: #fff; border-radius: 12rpx; padding: 20rpx 24rpx; border: 2rpx solid #E5E7EB; margin-bottom: 8rpx; }
.pet-card:active { border-color: #4F46E5; background: #FAFAFF; }
.pet-card-row1 { display: flex; justify-content: space-between; align-items: center; margin-bottom: 8rpx; }
.pet-card-name { font-size: 32rpx; font-weight: 700; color: #1F2937; }
.pet-card-owner { font-size: 24rpx; color: #6B7280; background: #F3F4F6; padding: 4rpx 16rpx; border-radius: 20rpx; }
.pet-card-row2 { display: flex; align-items: center; gap: 6rpx; margin-bottom: 8rpx; }
.pet-card-info { font-size: 26rpx; color: #6B7280; }
.pet-card-dot { font-size: 26rpx; color: #D1D5DB; }
.pet-card-row3 { display: flex; gap: 8rpx; flex-wrap: wrap; }
.pet-label { font-size: 22rpx; padding: 4rpx 14rpx; border-radius: 8rpx; background: #EEF2FF; color: #4F46E5; }
.pet-label.warn { background: #FEE2E2; color: #DC2626; }
.selected-pet { background: #EEF2FF; border-radius: 12rpx; padding: 20rpx; }
.pet-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 12rpx; }
.pet-name-lg { font-size: 32rpx; font-weight: bold; color: #4F46E5; }
.change-btn { color: #6B7280; font-size: 24rpx; }
.pet-tags { display: flex; gap: 12rpx; flex-wrap: wrap; margin-bottom: 12rpx; }
.tag { font-size: 22rpx; padding: 4rpx 16rpx; border-radius: 20rpx; background: #E0E7FF; color: #4338CA; }
.tag.warn { background: #FEE2E2; color: #DC2626; }
.pet-alert { display: flex; align-items: flex-start; gap: 8rpx; padding: 12rpx; background: #FEF3C7; border-radius: 8rpx; margin-bottom: 8rpx; }
.alert-icon { font-size: 24rpx; font-weight: bold; color: #D97706; width: 32rpx; }
.alert-text { font-size: 24rpx; color: #92400E; flex: 1; }
.service-picker { display: flex; flex-wrap: wrap; gap: 12rpx; }
.service-opt { padding: 16rpx 24rpx; border-radius: 12rpx; background: #F3F4F6; font-size: 26rpx; color: #374151; display: flex; flex-direction: column; align-items: center; min-width: 160rpx; }
.service-opt.active { background: #4F46E5; color: #fff; }
.svc-price { font-size: 28rpx; font-weight: bold; margin-top: 4rpx; }
.spec-section { margin-top: 20rpx; padding-top: 20rpx; border-top: 1rpx solid #F3F4F6; }
.spec-title { font-size: 26rpx; color: #6B7280; display: block; margin-bottom: 12rpx; }
.spec-picker { display: flex; flex-wrap: wrap; gap: 12rpx; }
.spec-opt { padding: 16rpx 24rpx; border-radius: 12rpx; background: #F9FAFB; border: 2rpx solid #E5E7EB; display: flex; flex-direction: column; align-items: center; min-width: 160rpx; }
.spec-opt.active { background: #EEF2FF; border-color: #4F46E5; }
.spec-opt.custom { border-style: dashed; }
.spec-opt.custom.active { border-style: solid; }
.spec-name { font-size: 26rpx; color: #374151; font-weight: 500; }
.spec-opt.active .spec-name { color: #4F46E5; }
.spec-meta { font-size: 24rpx; color: #4F46E5; font-weight: 600; margin-top: 4rpx; }
.spec-dur { font-size: 22rpx; color: #9CA3AF; font-weight: 400; }
.custom-price-row { display: flex; align-items: center; gap: 12rpx; margin-top: 16rpx; flex-wrap: wrap; }
.custom-price-label { font-size: 26rpx; color: #374151; font-weight: 500; }
.custom-price-input { width: 200rpx; font-size: 32rpx; font-weight: 600; color: #4F46E5; height: 60rpx; background: #F9FAFB; border: 2rpx solid #C7D2FE; border-radius: 10rpx; padding: 0 16rpx; text-align: center; }
.custom-price-hint { font-size: 22rpx; color: #9CA3AF; }
.custom-price-toggle { text-align: center; padding: 14rpx; font-size: 26rpx; color: #4F46E5; background: #EEF2FF; border-radius: 10rpx; border: 1rpx dashed #C7D2FE; }
.addon-grid { display: flex; flex-wrap: wrap; gap: 16rpx; }
.addon-item { display: flex; align-items: center; gap: 8rpx; width: calc(50% - 8rpx); position: relative; background: #F9FAFB; border: 1rpx solid #E5E7EB; border-radius: 12rpx; padding: 12rpx; }
.addon-item--deleting { background: #FFF1F2; border-color: #FECDD3; animation: addonShake 0.3s ease-in-out; }
@keyframes addonShake { 0%,100%{transform:translateX(0)} 25%{transform:translateX(-4rpx)} 75%{transform:translateX(4rpx)} }
.addon-delete-badge { position: absolute; top: -12rpx; left: -12rpx; width: 40rpx; height: 40rpx; border-radius: 50%; background: #EF4444; display: flex; align-items: center; justify-content: center; z-index: 2; }
.addon-delete-icon { color: #fff; font-size: 28rpx; font-weight: 700; }
.addon-label { font-size: 24rpx; color: #374151; width: 100rpx; }
.addon-input { background: #fff; border-radius: 8rpx; padding: 8rpx 12rpx; font-size: 26rpx; flex: 1; text-align: right; border: 1rpx solid #E5E7EB; }
.addon-add-btn { width: 48rpx; height: 48rpx; border-radius: 50%; background: linear-gradient(135deg, #6366F1, #4F46E5); display: flex; align-items: center; justify-content: center; }
.addon-add-icon { color: #fff; font-size: 32rpx; font-weight: 600; }
.addon-cancel-hint { margin-top: 12rpx; padding: 12rpx; border: 1rpx dashed #FECDD3; border-radius: 8rpx; text-align: center; }
.addon-cancel-text { font-size: 24rpx; color: #EF4444; }
.summary { }
.summary-row { display: flex; justify-content: space-between; padding: 8rpx 0; font-size: 26rpx; color: #374151; }
.summary-row.total { border-top: 1rpx solid #E5E7EB; padding-top: 16rpx; margin-top: 8rpx; }
.bold { font-weight: 600; }
.discount { color: #059669; }
.pay-amount { font-size: 36rpx; font-weight: bold; color: #4F46E5; }
.form-row { display: flex; align-items: center; gap: 16rpx; padding: 16rpx 0; border-bottom: 1rpx solid #F3F4F6; }
.form-row:last-child { border-bottom: none; }
.label { font-size: 28rpx; color: #374151; width: 140rpx; }
.required { color: #EF4444; font-size: 22rpx; font-weight: 600; }
.picker { font-size: 28rpx; color: #1F2937; flex: 1; }
.picker-warn { color: #EF4444; }
.input { font-size: 28rpx; color: #1F2937; flex: 1; }
.staff-commission { font-size: 24rpx; color: #6B7280; padding: 8rpx 0; }
.btn-submit { background: #4F46E5; color: #fff; border-radius: 12rpx; font-size: 30rpx; margin-top: 16rpx; }
</style>
