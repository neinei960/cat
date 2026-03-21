<template>
  <SideLayout>
  <view class="page">

    <!-- 区块1：已添加的猫咪列表 -->
    <view v-if="petOrders.length > 0" class="section">
      <text class="section-title">已添加猫咪（{{ petOrders.length }}只）</text>
      <view class="pet-order-card" v-for="(po, idx) in petOrders" :key="idx">
        <view class="poc-header">
          <view class="poc-title-row">
            <text class="poc-pet-name">{{ po.pet.name }}</text>
            <text class="poc-svc-name" v-if="getServiceName(po.serviceId)">· {{ getServiceName(po.serviceId) }}</text>
          </view>
          <view class="poc-actions">
            <text class="poc-edit-btn" @click="editPetOrder(idx)">编辑</text>
            <text class="poc-del-btn" @click="removePetOrder(idx)">✕</text>
          </view>
        </view>
        <view class="poc-price-row">
          <text class="poc-spec" v-if="getSpecName(po)">{{ getSpecName(po) }}</text>
          <text class="poc-price">
            ¥{{ (po.useCustomPrice ? (parseFloat(po.customPrice) || 0) : po.servicePrice).toFixed(2) }}
          </text>
        </view>
        <view class="poc-addons" v-if="getActiveAddons(po).length > 0">
          <text class="poc-addon-item" v-for="a in getActiveAddons(po)" :key="a.id">
            {{ a.name }} ¥{{ a.amount }}
          </text>
        </view>
      </view>
    </view>

    <!-- 区块2：添加/编辑面板 -->
    <view v-if="showAddPanel" class="section">
      <text class="section-title">{{ editingIndex >= 0 ? '编辑猫咪' : '添加猫咪' }}</text>

      <!-- 猫咪选择 -->
      <view class="sub-section">
        <text class="sub-title">选择猫咪</text>
        <view class="search-bar" v-if="!panel.selectedPet">
          <input
            :value="petKeyword"
            placeholder="输入猫咪名搜索"
            class="search-input"
            @input="(e: any) => { petKeyword = e.detail.value; searchPets() }"
            @confirm="searchPets"
          />
        </view>
        <view class="option-list" v-if="!panel.selectedPet && petList.length > 0">
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
            </view>
          </view>
        </view>
        <view v-if="panel.selectedPet" class="selected-pet">
          <view class="pet-header">
            <text class="pet-name-lg">{{ panel.selectedPet.name }}</text>
            <text class="change-btn" @click="clearPanelPet">更换</text>
          </view>
          <view class="pet-tags">
            <text class="tag" v-if="panel.selectedPet.fur_level">{{ panel.selectedPet.fur_level }}</text>
            <text class="tag" v-if="panel.selectedPet.personality">{{ panel.selectedPet.personality }}</text>
            <text class="tag warn" v-if="panel.selectedPet.aggression && panel.selectedPet.aggression !== '无'">攻击性:{{ panel.selectedPet.aggression }}</text>
          </view>
          <view class="pet-alert" v-if="panel.selectedPet.care_notes">
            <text class="alert-icon">!</text>
            <text class="alert-text">{{ panel.selectedPet.care_notes }}</text>
          </view>
          <view class="pet-alert" v-if="panel.selectedPet.forbidden_zones">
            <text class="alert-icon">x</text>
            <text class="alert-text">禁区: {{ panel.selectedPet.forbidden_zones }}</text>
          </view>
        </view>
      </view>

      <!-- 洗浴项目 -->
      <view class="sub-section" v-if="panel.selectedPet">
        <text class="sub-title">洗浴项目</text>
        <view class="service-picker" v-if="serviceList.length > 0">
          <view
            class="service-opt"
            :class="{ active: panel.serviceId === s.ID }"
            v-for="s in serviceList"
            :key="s.ID"
            @click="selectService(s)"
          >
            <text>{{ s.name }}</text>
            <text class="svc-price" v-if="panel.serviceId === s.ID && !panel.selectedSpecId">¥{{ panel.servicePrice.toFixed(2) }}</text>
          </view>
        </view>

        <!-- 规格选择 -->
        <view v-if="panel.serviceId && panel.specList.length > 0" class="spec-section">
          <text class="spec-title">选择规格</text>
          <view class="spec-picker">
            <view
              v-for="spec in panel.specList" :key="spec.ID"
              :class="['spec-opt', panel.selectedSpecId === spec.ID && !panel.useCustomPrice ? 'active' : '']"
              @click="selectSpec(spec)"
            >
              <text class="spec-name">{{ spec.name }}</text>
              <text class="spec-meta">¥{{ spec.price }} <text v-if="spec.duration" class="spec-dur">· {{ spec.duration }}分钟</text></text>
            </view>
            <view :class="['spec-opt custom', panel.useCustomPrice ? 'active' : '']" @click="enableCustomPrice">
              <text class="spec-name">自定义</text>
              <text class="spec-meta">手动输入</text>
            </view>
          </view>
          <view v-if="panel.useCustomPrice" class="custom-price-row">
            <text class="custom-price-label">输入金额</text>
            <input
              :value="panel.customPrice"
              type="digit"
              placeholder="0.00"
              class="custom-price-input"
              @input="(e: any) => { panel.customPrice = e.detail.value; panel.servicePrice = parseFloat(e.detail.value) || 0 }"
            />
            <text class="custom-price-hint">适用于美团团购等特殊价格</text>
          </view>
        </view>

        <!-- 无规格时自定义价格 -->
        <view v-if="panel.serviceId && panel.specList.length === 0" class="spec-section">
          <view class="custom-price-toggle" @click="panel.useCustomPrice = !panel.useCustomPrice">
            <text>{{ panel.useCustomPrice ? '使用基础价格 ¥' + getPanelBasePrice() : '自定义金额（美团团购等）' }}</text>
          </view>
          <view v-if="panel.useCustomPrice" class="custom-price-row">
            <text class="custom-price-label">输入金额</text>
            <input
              :value="panel.customPrice"
              type="digit"
              placeholder="0.00"
              class="custom-price-input"
              @input="(e: any) => { panel.customPrice = e.detail.value; panel.servicePrice = parseFloat(e.detail.value) || 0 }"
            />
          </view>
        </view>
      </view>

      <!-- 附加费用 -->
      <view class="sub-section" v-if="panel.selectedPet">
        <view class="section-title-row">
          <text class="sub-title" style="margin-bottom:0">附加费用</text>
          <view class="addon-add-btn" @click="onAddAddon">
            <text class="addon-add-icon">+</text>
          </view>
        </view>
        <view class="addon-grid">
          <view
            class="addon-item"
            :class="{ 'addon-item--deleting': longPressId === a.id }"
            v-for="a in panel.addons"
            :key="a.id"
            @longpress="onLongPress(a.id)"
          >
            <view
              v-if="longPressId === a.id"
              class="addon-delete-badge"
              @click.stop="onDeleteAddon(a)"
            >
              <text class="addon-delete-icon">−</text>
            </view>
            <text class="addon-label">{{ a.name }}</text>
            <input
              :value="a.amount"
              type="digit"
              placeholder="0"
              class="addon-input"
              :disabled="longPressId === a.id"
              @input="(e: any) => { a.amount = e.detail.value }"
            />
          </view>
        </view>
        <view v-if="longPressId !== null" class="addon-cancel-hint" @click="longPressId = null">
          <text class="addon-cancel-text">点击此处取消删除模式</text>
        </view>
      </view>

      <!-- 添加下一只 / 确认修改 -->
      <view v-if="panel.selectedPet && panel.serviceId" class="panel-confirm-row">
        <view v-if="editingIndex >= 0">
          <button class="btn-confirm-add" @click="confirmAddPetOrder">确认修改</button>
        </view>
        <view v-else>
          <button class="btn-add-next-inline" @click="addNextPet">+ 添加下一只猫咪</button>
        </view>
      </view>
    </view>

    <!-- 区块3：合计 + 洗护师 + 提交（面板有猫且选了服务就显示） -->
    <view v-if="allPetOrders.length > 0" class="section summary">
      <view class="summary-row" v-for="(po, idx) in allPetOrders" :key="idx">
        <text>{{ po.pet.name }}</text>
        <text>¥{{ getPetOrderTotal(po).toFixed(2) }}</text>
      </view>
      <view class="summary-row total">
        <text>合计</text>
        <text class="pay-amount">¥{{ grandTotal.toFixed(2) }}</text>
      </view>
    </view>

    <view v-if="allPetOrders.length > 0" class="section">
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

    <button
      v-if="allPetOrders.length > 0"
      class="btn-submit"
      @click="onSubmit"
      :loading="submitting"
    >确认开单{{ allPetOrders.length > 1 ? '（' + allPetOrders.length + '只猫）' : '' }}</button>

  </view>
  </SideLayout>
</template>

<script setup lang="ts">
import SideLayout from '@/components/SideLayout.vue'
import { ref, computed, reactive, onMounted } from 'vue'
import { getPetList } from '@/api/pet'
import { createOrder, payOrder } from '@/api/order'
import { getAddonList, createAddon, deleteAddon, priceLookup } from '@/api/addon'
import { getServiceList, getPriceRules } from '@/api/service'
import { getStaffList } from '@/api/staff'
import { getCustomerCard } from '@/api/member-card'
import { safeBack } from '@/utils/navigate'
import { getPersonalityColor, getPersonalityBg } from '@/utils/personality'

// ---- 类型 ----
interface AddonInput {
  id: number
  name: string
  amount: string
}

interface PetOrder {
  pet: any
  serviceId: number
  servicePrice: number
  specId: number
  specList: any[]
  useCustomPrice: boolean
  customPrice: string
  addons: AddonInput[]
}

// ---- 全局数据 ----
const serviceList = ref<any[]>([])
const staffList = ref<any[]>([])
const addonTemplate = ref<AddonInput[]>([])  // 全局 addon 模板（名称+id），不含金额

const selectedStaffIdx = ref(0)
const remark = ref('')
const submitting = ref(false)

const staffNames = computed(() => ['未选择', ...staffList.value.map(s => s.name)])
const selectedStaff = computed(() => {
  if (selectedStaffIdx.value === 0) return null
  return staffList.value[selectedStaffIdx.value - 1]
})

// ---- 多猫列表 ----
const petOrders = ref<PetOrder[]>([])

// ---- 面板状态 ----
const showAddPanel = ref(true)
const editingIndex = ref(-1)  // -1 = 新增，>=0 = 编辑第 idx 只

// 当前编辑面板的状态
const panel = reactive({
  selectedPet: null as any,
  serviceId: 0,
  servicePrice: 0,
  specList: [] as any[],
  selectedSpecId: 0,
  useCustomPrice: false,
  customPrice: '',
  addons: [] as AddonInput[],
})

// ---- 搜索猫咪 ----
const petKeyword = ref('')
const petList = ref<any[]>([])
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
  panel.selectedPet = p
  petList.value = []
  petKeyword.value = ''
  // 如果已选了服务，重新 lookup 价格
  if (panel.serviceId > 0 && p.fur_level) {
    await lookupPrice(panel.serviceId, p.fur_level)
  }
}

function clearPanelPet() {
  panel.selectedPet = null
  panel.serviceId = 0
  panel.servicePrice = 0
  panel.specList = []
  panel.selectedSpecId = 0
  panel.useCustomPrice = false
  panel.customPrice = ''
}

// ---- 服务/规格选择 ----
async function selectService(s: any) {
  panel.serviceId = s.ID
  panel.selectedSpecId = 0
  panel.servicePrice = s.base_price
  panel.useCustomPrice = false
  panel.customPrice = ''
  try {
    const res = await getPriceRules(s.ID)
    const rules = res.data || []
    panel.specList = rules.map((r: any) => ({
      ...r,
      name: r.name || r.fur_level || r.pet_size || '规格',
    }))
    // 如果猫有毛量信息，自动 lookup 价格
    if (panel.selectedPet?.fur_level) {
      await lookupPrice(s.ID, panel.selectedPet.fur_level)
    }
  } catch {
    panel.specList = []
  }
}

function selectSpec(spec: any) {
  panel.selectedSpecId = spec.ID
  panel.servicePrice = spec.price
  panel.useCustomPrice = false
  panel.customPrice = ''
}

function enableCustomPrice() {
  panel.useCustomPrice = true
  panel.selectedSpecId = 0
  if (panel.customPrice) {
    panel.servicePrice = parseFloat(panel.customPrice) || 0
  }
}

function getPanelBasePrice(): string {
  const svc = serviceList.value.find(s => s.ID === panel.serviceId)
  return svc ? svc.base_price.toFixed(2) : '0'
}

async function lookupPrice(serviceId: number, furLevel: string) {
  try {
    const res = await priceLookup(serviceId, furLevel)
    panel.servicePrice = res.data.price
  } catch {
    const svc = serviceList.value.find(s => s.ID === serviceId)
    panel.servicePrice = svc?.base_price || 0
  }
}

// ---- 附加费 ----
const longPressId = ref<number | null>(null)

function onLongPress(id: number) {
  longPressId.value = id
  try { uni.vibrateShort({}) } catch (_) {}
}

function onAddAddon() {
  uni.showModal({
    title: '添加附加费类型',
    editable: true,
    placeholderText: '请输入费用名称，如：开结费',
    success: async (res) => {
      if (!res.confirm || !res.content?.trim()) return
      const name = res.content.trim()
      try {
        uni.showLoading({ title: '添加中...' })
        await createAddon({ name, default_price: 0, is_variable: true })
        await loadAddons()
        uni.hideLoading()
        uni.showToast({ title: `已添加「${name}」`, icon: 'success' })
      } catch (e) {
        uni.hideLoading()
        uni.showToast({ title: '添加失败，请重试', icon: 'none' })
      }
    },
  })
}

function onDeleteAddon(a: AddonInput) {
  uni.showModal({
    title: '删除附加费类型',
    content: `确定要删除「${a.name}」吗？\n这将从全局费用列表中移除。`,
    confirmColor: '#EF4444',
    success: async (res) => {
      if (!res.confirm) {
        longPressId.value = null
        return
      }
      if (a.id < 0) {
        panel.addons = panel.addons.filter(item => item.id !== a.id)
        longPressId.value = null
        return
      }
      try {
        uni.showLoading({ title: '删除中...' })
        await deleteAddon(a.id)
        await loadAddons()
        uni.hideLoading()
        uni.showToast({ title: `已删除「${a.name}」`, icon: 'success' })
      } catch (e) {
        uni.hideLoading()
        uni.showToast({ title: '删除失败，请重试', icon: 'none' })
      } finally {
        longPressId.value = null
      }
    },
  })
}

// 加载全局附加费模板（只有名称，金额清空）
async function loadAddons() {
  try {
    const aRes = await getAddonList()
    if (Array.isArray(aRes.data)) {
      addonTemplate.value = aRes.data.map(a => ({ id: a.ID, name: a.name, amount: '' }))
    }
  } catch (e) {
    addonTemplate.value = [
      { id: -1, name: '超重费', amount: '' },
      { id: -2, name: '去油费', amount: '' },
      { id: -3, name: '药浴', amount: '' },
      { id: -4, name: '刷牙', amount: '' },
      { id: -5, name: '开结', amount: '' },
      { id: -6, name: '攻击费', amount: '' },
    ]
  }
  // 同步更新面板的 addons（保留已填金额）
  syncPanelAddons()
}

// 将最新模板同步到面板（保留已有金额）
function syncPanelAddons() {
  const amountMap = new Map(panel.addons.map(a => [a.id, a.amount]))
  panel.addons = addonTemplate.value.map(a => ({
    id: a.id,
    name: a.name,
    amount: amountMap.get(a.id) || '',
  }))
}

// 初始化面板 addons（从模板复制，金额全清）
function initPanelAddons() {
  panel.addons = addonTemplate.value.map(a => ({ id: a.id, name: a.name, amount: '' }))
}

// ---- 面板操作 ----
function openAddPanel() {
  editingIndex.value = -1
  panel.selectedPet = null
  panel.serviceId = 0
  panel.servicePrice = 0
  panel.specList = []
  panel.selectedSpecId = 0
  panel.useCustomPrice = false
  panel.customPrice = ''
  initPanelAddons()
  petKeyword.value = ''
  petList.value = []
  showAddPanel.value = true
}

function editPetOrder(idx: number) {
  const po = petOrders.value[idx]
  editingIndex.value = idx
  panel.selectedPet = po.pet
  panel.serviceId = po.serviceId
  panel.servicePrice = po.servicePrice
  panel.specList = po.specList
  panel.selectedSpecId = po.specId
  panel.useCustomPrice = po.useCustomPrice
  panel.customPrice = po.customPrice
  // 深拷贝 addons
  panel.addons = po.addons.map(a => ({ ...a }))
  petKeyword.value = ''
  petList.value = []
  showAddPanel.value = true
}

function confirmAddPetOrder() {
  if (!panel.selectedPet) {
    uni.showToast({ title: '请选择猫咪', icon: 'none' }); return
  }
  if (!panel.serviceId) {
    uni.showToast({ title: '请选择洗浴项目', icon: 'none' }); return
  }

  const po: PetOrder = {
    pet: panel.selectedPet,
    serviceId: panel.serviceId,
    servicePrice: panel.servicePrice,
    specId: panel.selectedSpecId,
    specList: [...panel.specList],
    useCustomPrice: panel.useCustomPrice,
    customPrice: panel.customPrice,
    addons: panel.addons.map(a => ({ ...a })),
  }

  if (editingIndex.value >= 0) {
    petOrders.value[editingIndex.value] = po
    uni.showToast({ title: '已更新', icon: 'success' })
  } else {
    petOrders.value.push(po)
    uni.showToast({ title: '已添加', icon: 'success' })
  }

  // 收起面板
  showAddPanel.value = false
  editingIndex.value = -1
}

function removePetOrder(idx: number) {
  uni.showModal({
    title: '确认删除',
    content: `移除「${petOrders.value[idx].pet.name}」的开单信息？`,
    confirmColor: '#EF4444',
    success: (res) => {
      if (!res.confirm) return
      petOrders.value.splice(idx, 1)
      if (petOrders.value.length === 0) {
        showAddPanel.value = true
      }
    },
  })
}

// ---- 价格计算 ----
const panelTotalPrice = computed(() => {
  const base = panel.useCustomPrice ? (parseFloat(panel.customPrice) || 0) : panel.servicePrice
  const addons = panel.addons.reduce((s, a) => s + (parseFloat(a.amount) || 0), 0)
  return base + addons
})

function getPetOrderTotal(po: PetOrder): number {
  const base = po.useCustomPrice ? (parseFloat(po.customPrice) || 0) : po.servicePrice
  const addons = po.addons.reduce((s, a) => s + (parseFloat(a.amount) || 0), 0)
  return base + addons
}

// 当前面板如果有有效数据，也算一只猫（用于合计和提交）
function currentPanelAsPetOrder(): PetOrder | null {
  if (!panel.selectedPet || !panel.serviceId) return null
  if (editingIndex.value >= 0) return null // 编辑模式不重复计算
  return {
    pet: panel.selectedPet,
    serviceId: panel.serviceId,
    servicePrice: panel.servicePrice,
    specId: panel.selectedSpecId,
    specList: panel.specList,
    useCustomPrice: panel.useCustomPrice,
    customPrice: panel.customPrice,
    addons: panel.addons,
  }
}

const allPetOrders = computed<PetOrder[]>(() => {
  const list = [...petOrders.value]
  const current = currentPanelAsPetOrder()
  if (current) list.push(current)
  return list
})

const grandTotal = computed(() =>
  allPetOrders.value.reduce((s, po) => s + getPetOrderTotal(po), 0)
)

// 添加下一只：把当前面板保存到列表，清空面板
function addNextPet() {
  const current = currentPanelAsPetOrder()
  if (current) {
    petOrders.value.push(current)
  }
  resetPanel()
  showAddPanel.value = true
}

const commission = computed(() => {
  if (!selectedStaff.value) return 0
  return Math.round(grandTotal.value * selectedStaff.value.commission_rate) / 100
})

// ---- 辅助函数 ----
function getServiceName(serviceId: number): string {
  return serviceList.value.find(s => s.ID === serviceId)?.name || ''
}

function getSpecName(po: PetOrder): string {
  if (po.useCustomPrice) return '自定义价格'
  if (!po.specId) return ''
  const spec = po.specList.find(s => s.ID === po.specId)
  return spec?.name || ''
}

function getActiveAddons(po: PetOrder) {
  return po.addons.filter(a => parseFloat(a.amount) > 0)
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

// ---- 提交 ----
async function onSubmit() {
  const orders = allPetOrders.value
  if (orders.length === 0) {
    uni.showToast({ title: '请至少添加一只猫咪并选择服务', icon: 'none' }); return
  }
  if (!selectedStaff.value) {
    uni.showToast({ title: '请选择洗护师', icon: 'none' }); return
  }

  submitting.value = true
  try {
    for (const po of orders) {
      await createOrder({
        pet_id: po.pet.ID,
        customer_id: po.pet.customer_id || undefined,
        service_id: po.serviceId,
        staff_id: selectedStaff.value?.ID || undefined,
        addons: po.addons
          .filter(a => parseFloat(a.amount) > 0)
          .map(a => ({ name: a.name, amount: parseFloat(a.amount) })),
        remark: remark.value,
      } as any)
    }
    uni.showToast({ title: '开单成功', icon: 'success' })
    setTimeout(() => uni.redirectTo({ url: '/pages/order/list' }), 600)
  } catch (e: any) {
    uni.showToast({ title: e.message || '开单失败', icon: 'none' })
  } finally {
    submitting.value = false
  }
}

// ---- 初始化 ----
onMounted(async () => {
  try {
    const sRes = await getServiceList({ page_size: 50 } as any)
    if (sRes.data?.list) serviceList.value = sRes.data.list
  } catch (e) { /* ignore */ }

  try {
    const stRes = await getStaffList({ page_size: 50 } as any)
    if (stRes.data?.list) staffList.value = stRes.data.list.filter((s: any) => s.role === 'staff')
  } catch (e) { /* ignore */ }

  await loadAddons()
})
</script>

<style scoped>
.page { padding: 24rpx; }
.section { background: #fff; border-radius: 16rpx; padding: 24rpx; margin-bottom: 16rpx; }
.section-title { font-size: 28rpx; font-weight: 600; color: #1F2937; display: block; margin-bottom: 16rpx; }
.sub-section { margin-bottom: 24rpx; }
.sub-section:last-child { margin-bottom: 0; }
.sub-title { font-size: 26rpx; font-weight: 600; color: #374151; display: block; margin-bottom: 12rpx; }

/* 已添加猫咪卡片 */
.pet-order-card {
  border: 2rpx solid #E5E7EB;
  border-radius: 12rpx;
  padding: 16rpx 20rpx;
  margin-bottom: 12rpx;
}
.pet-order-card:last-child { margin-bottom: 0; }
.poc-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 8rpx; }
.poc-title-row { display: flex; align-items: center; gap: 4rpx; }
.poc-pet-name { font-size: 30rpx; font-weight: 700; color: #1F2937; }
.poc-svc-name { font-size: 26rpx; color: #6B7280; }
.poc-actions { display: flex; align-items: center; gap: 16rpx; }
.poc-edit-btn { font-size: 24rpx; color: #4F46E5; padding: 4rpx 12rpx; }
.poc-del-btn { font-size: 24rpx; color: #9CA3AF; padding: 4rpx 12rpx; }
.poc-price-row { display: flex; align-items: center; gap: 12rpx; margin-bottom: 6rpx; }
.poc-spec { font-size: 24rpx; color: #6B7280; background: #F3F4F6; padding: 2rpx 12rpx; border-radius: 20rpx; }
.poc-price { font-size: 28rpx; font-weight: 600; color: #4F46E5; }
.poc-addons { display: flex; flex-wrap: wrap; gap: 8rpx; }
.poc-addon-item { font-size: 22rpx; color: #6B7280; background: #F9FAFB; padding: 4rpx 12rpx; border-radius: 8rpx; }

/* 搜索/猫咪选择 */
.search-bar { margin-bottom: 12rpx; }
.search-input { background: #F3F4F6; border-radius: 12rpx; padding: 16rpx 24rpx; font-size: 28rpx; width: 100%; box-sizing: border-box; }
.option-list { display: flex; flex-direction: column; gap: 8rpx; max-height: 400rpx; overflow-y: auto; }
.pet-card {
  background: #fff; border-radius: 12rpx; padding: 20rpx 24rpx;
  border: 2rpx solid #E5E7EB; margin-bottom: 8rpx;
}
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

/* 服务/规格 */
.service-picker { display: flex; flex-wrap: wrap; gap: 12rpx; }
.service-opt { padding: 16rpx 24rpx; border-radius: 12rpx; background: #F3F4F6; font-size: 26rpx; color: #374151; display: flex; flex-direction: column; align-items: center; min-width: 160rpx; }
.service-opt.active { background: #4F46E5; color: #fff; }
.svc-price { font-size: 28rpx; font-weight: bold; margin-top: 4rpx; }
.spec-section { margin-top: 20rpx; padding-top: 20rpx; border-top: 1rpx solid #F3F4F6; }
.spec-title { font-size: 26rpx; color: #6B7280; display: block; margin-bottom: 12rpx; }
.spec-picker { display: flex; flex-wrap: wrap; gap: 12rpx; }
.spec-opt {
  padding: 16rpx 24rpx; border-radius: 12rpx;
  background: #F9FAFB; border: 2rpx solid #E5E7EB;
  display: flex; flex-direction: column; align-items: center;
  min-width: 160rpx;
}
.spec-opt.active { background: #EEF2FF; border-color: #4F46E5; }
.spec-name { font-size: 26rpx; color: #374151; font-weight: 500; }
.spec-opt.active .spec-name { color: #4F46E5; }
.spec-meta { font-size: 24rpx; color: #4F46E5; font-weight: 600; margin-top: 4rpx; }
.spec-dur { font-size: 22rpx; color: #9CA3AF; font-weight: 400; }
.spec-opt.custom { border-style: dashed; }
.spec-opt.custom.active { border-style: solid; }
.custom-price-row { display: flex; align-items: center; gap: 12rpx; margin-top: 16rpx; flex-wrap: wrap; }
.custom-price-label { font-size: 26rpx; color: #374151; font-weight: 500; }
.custom-price-input { width: 200rpx; font-size: 32rpx; font-weight: 600; color: #4F46E5; height: 60rpx; background: #F9FAFB; border: 2rpx solid #C7D2FE; border-radius: 10rpx; padding: 0 16rpx; text-align: center; }
.custom-price-hint { font-size: 22rpx; color: #9CA3AF; }
.custom-price-toggle { text-align: center; padding: 14rpx; font-size: 26rpx; color: #4F46E5; background: #EEF2FF; border-radius: 10rpx; border: 1rpx dashed #C7D2FE; }

/* 附加费 */
.section-title-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 16rpx;
}
.addon-add-btn {
  width: 48rpx;
  height: 48rpx;
  border-radius: 999rpx;
  background: linear-gradient(135deg, #6366F1, #4F46E5);
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: 0 4rpx 12rpx rgba(79, 70, 229, 0.35);
  transition: transform 0.15s, box-shadow 0.15s;
}
.addon-add-btn:active {
  transform: scale(0.88);
  box-shadow: 0 2rpx 6rpx rgba(79, 70, 229, 0.2);
}
.addon-add-icon {
  font-size: 36rpx;
  color: #fff;
  font-weight: 300;
  line-height: 1;
  margin-top: -2rpx;
}
.addon-grid { display: flex; flex-wrap: wrap; gap: 16rpx; }
.addon-item {
  position: relative;
  display: flex;
  align-items: center;
  gap: 8rpx;
  width: calc(50% - 8rpx);
  background: #F9FAFB;
  border-radius: 12rpx;
  padding: 12rpx 16rpx;
  border: 2rpx solid transparent;
  transition: background 0.2s, border-color 0.2s, transform 0.2s;
}
.addon-item--deleting {
  background: #FEE2E2;
  border-color: #FCA5A5;
  animation: addonShake 0.4s ease;
}
@keyframes addonShake {
  0%   { transform: translateX(0); }
  20%  { transform: translateX(-6rpx) rotate(-1deg); }
  40%  { transform: translateX(6rpx) rotate(1deg); }
  60%  { transform: translateX(-4rpx) rotate(-0.5deg); }
  80%  { transform: translateX(4rpx) rotate(0.5deg); }
  100% { transform: translateX(0); }
}
.addon-delete-badge {
  position: absolute;
  top: -14rpx;
  left: -14rpx;
  width: 40rpx;
  height: 40rpx;
  border-radius: 999rpx;
  background: #EF4444;
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: 0 2rpx 8rpx rgba(239, 68, 68, 0.45);
  z-index: 10;
  transition: transform 0.1s;
}
.addon-delete-badge:active { transform: scale(0.85); }
.addon-delete-icon {
  font-size: 32rpx;
  color: #fff;
  font-weight: 700;
  line-height: 1;
  margin-top: -2rpx;
}
.addon-label { font-size: 24rpx; color: #374151; width: 100rpx; flex-shrink: 0; }
.addon-item--deleting .addon-label { color: #B91C1C; font-weight: 600; }
.addon-input { background: #fff; border-radius: 8rpx; padding: 10rpx 12rpx; font-size: 26rpx; flex: 1; text-align: right; border: 1rpx solid #E5E7EB; }
.addon-item--deleting .addon-input { background: #FEF2F2; border-color: #FCA5A5; color: #B91C1C; }
.addon-cancel-hint {
  margin-top: 16rpx;
  padding: 16rpx;
  background: #FEE2E2;
  border-radius: 10rpx;
  text-align: center;
  border: 1rpx dashed #FCA5A5;
}
.addon-cancel-text { font-size: 24rpx; color: #DC2626; }

/* 面板确认行 */
.panel-confirm-row {
  display: flex;
  align-items: center;
  gap: 20rpx;
  margin-top: 24rpx;
  padding-top: 20rpx;
  border-top: 1rpx solid #F3F4F6;
}
.btn-add-next-inline {
  background: #EEF2FF; color: #4F46E5; border: 2rpx dashed #C7D2FE;
  border-radius: 12rpx; font-size: 28rpx; padding: 16rpx 0; text-align: center; width: 100%;
}
.btn-confirm-add {
  flex: 1;
  background: #4F46E5;
  color: #fff;
  border-radius: 12rpx;
  font-size: 28rpx;
  height: 80rpx;
  line-height: 80rpx;
  padding: 0;
}
.panel-price-hint {
  font-size: 32rpx;
  font-weight: 700;
  color: #4F46E5;
  min-width: 120rpx;
  text-align: right;
}

/* 添加下一只 */
.add-next-section { padding: 0; }
.btn-add-next {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 12rpx;
  padding: 28rpx;
  border: 2rpx dashed #C7D2FE;
  border-radius: 16rpx;
  background: #F5F3FF;
}
.btn-add-next:active { background: #EDE9FE; }
.add-next-icon { font-size: 36rpx; color: #4F46E5; font-weight: 300; }
.add-next-text { font-size: 28rpx; color: #4F46E5; font-weight: 500; }

/* 合计 */
.summary { }
.summary-row { display: flex; justify-content: space-between; padding: 8rpx 0; font-size: 26rpx; color: #374151; }
.summary-row.total { border-top: 1rpx solid #E5E7EB; padding-top: 16rpx; margin-top: 8rpx; }
.pay-amount { font-size: 36rpx; font-weight: bold; color: #4F46E5; }

/* 洗护师/备注 */
.form-row { display: flex; align-items: center; gap: 16rpx; padding: 16rpx 0; border-bottom: 1rpx solid #F3F4F6; }
.form-row:last-child { border-bottom: none; }
.label { font-size: 28rpx; color: #374151; width: 140rpx; }
.picker { font-size: 28rpx; color: #1F2937; flex: 1; }
.input { font-size: 28rpx; color: #1F2937; flex: 1; }
.required { color: #EF4444; font-size: 22rpx; font-weight: 600; }
.picker-warn { color: #EF4444; }
.staff-commission { font-size: 24rpx; color: #6B7280; padding: 8rpx 0; }

/* 提交按钮 */
.btn-submit { background: #4F46E5; color: #fff; border-radius: 12rpx; font-size: 30rpx; margin-top: 16rpx; }
</style>
