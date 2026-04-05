<template>
  <SideLayout>
    <view class="page">
      <view class="section">
        <view class="section-head">
          <view>
            <text class="section-title">客户与猫咪</text>
            <text class="section-desc">商品可直接零售，服务需要先选猫咪。</text>
          </view>
          <view v-if="isEditing" class="section-badge">修改订单</view>
        </view>

        <view class="selector-block">
          <view class="selector-head">
            <text class="selector-title">客户</text>
            <text v-if="selectedCustomer" class="selector-link" @click="clearCustomer">清空</text>
          </view>
          <view v-if="selectedCustomer" class="selected-card">
            <view class="selected-top">
              <text class="selected-name">{{ selectedCustomer.nickname || '未命名客户' }}</text>
              <text class="selected-sub">{{ selectedCustomer.phone || '未留手机号' }}</text>
            </view>
            <view class="selected-tags">
              <text class="selected-tag" v-if="memberBalance > 0">余额 ¥{{ memberBalance.toFixed(2) }}</text>
              <text class="selected-tag" v-if="serviceDiscountRate < 1">服务 {{ (serviceDiscountRate * 10).toFixed(1) }} 折</text>
              <text class="selected-tag" v-if="productDiscountRate < 1">商品 {{ (productDiscountRate * 10).toFixed(1) }} 折</text>
            </view>
          </view>
          <view v-else>
            <view class="search-shell">
              <view class="search-accent">搜客户</view>
              <input
                v-model="customerKeyword"
                class="search-input"
                placeholder="输入客户昵称或手机号搜索"
                confirm-type="search"
                @input="searchCustomers"
                @confirm="searchCustomers"
              />
            </view>
            <view v-if="customerOptions.length > 0" class="option-list">
              <view class="option-card" v-for="customer in customerOptions" :key="customer.ID" @click="selectCustomer(customer)">
                <view class="option-row">
                  <text class="option-name">{{ customer.nickname || '未命名客户' }}</text>
                  <text class="option-meta">{{ customer.phone || '未留手机号' }}</text>
                </view>
                <view class="option-tags">
                  <text class="option-tag" v-if="customer.member_balance > 0">余额 ¥{{ customer.member_balance.toFixed(2) }}</text>
                  <text class="option-tag" v-if="customer.discount_rate > 0 && customer.discount_rate < 1">服务 {{ (customer.discount_rate * 10).toFixed(1) }} 折</text>
                </view>
              </view>
            </view>
          </view>
        </view>

        <view class="selector-block">
          <view class="selector-head">
            <text class="selector-title">猫咪</text>
            <text v-if="selectedPet" class="selector-link" @click="clearPet">清空</text>
          </view>
          <view v-if="selectedPet" class="selected-card selected-pet-card">
            <view class="selected-top">
              <text class="selected-name">{{ selectedPet.name }}</text>
              <text class="selected-sub">{{ selectedPet.breed || '未知品种' }}</text>
            </view>
            <view class="selected-tags">
              <text class="selected-tag" v-if="selectedPet.fur_level">{{ selectedPet.fur_level }}</text>
              <text class="selected-tag" v-if="selectedPet.personality">{{ selectedPet.personality }}</text>
              <text class="selected-tag warn" v-if="selectedPet.aggression && selectedPet.aggression !== '无'">{{ selectedPet.aggression }}</text>
            </view>
            <view class="alert-box" v-if="selectedPet.care_notes">
              <text class="alert-icon">!</text>
              <text class="alert-text">{{ selectedPet.care_notes }}</text>
            </view>
          </view>
          <view v-else>
            <view class="search-shell">
              <view class="search-accent">{{ selectedCustomer ? '当前客户' : '搜猫咪' }}</view>
              <input
                v-model="petKeyword"
                class="search-input"
                :placeholder="selectedCustomer ? '搜索该客户的猫咪' : '输入猫咪名搜索'"
                confirm-type="search"
                @input="searchPets"
                @confirm="searchPets"
              />
            </view>
            <view v-if="selectedCustomer && petOptions.length === 0" class="inline-hint">该客户暂无猫咪档案，可直接开商品单。</view>
            <view v-if="petOptions.length > 0" class="option-list">
              <view class="option-card" v-for="pet in petOptions" :key="pet.ID" @click="selectPet(pet)">
                <view class="option-row">
                  <text class="option-name">{{ pet.name }}</text>
                  <text class="option-meta">{{ pet.customer?.nickname || selectedCustomer?.nickname || '散客' }}</text>
                </view>
                <view class="option-row subtle">
                  <text class="option-meta">{{ pet.breed || '未知品种' }}</text>
                  <text class="option-meta">{{ pet.gender === 1 ? '公' : pet.gender === 2 ? '母' : '未知性别' }}</text>
                  <text class="option-meta" v-if="pet.birth_date">{{ calcAge(pet.birth_date) }}</text>
                </view>
                <view class="option-tags">
                  <text class="option-tag" v-if="pet.fur_level">{{ pet.fur_level }}</text>
                  <text class="option-tag" v-if="pet.personality">{{ pet.personality }}</text>
                </view>
              </view>
            </view>
          </view>
        </view>
      </view>

      <view class="section">
        <view class="section-head">
          <view>
            <text class="section-title">商品开单</text>
            <text class="section-desc">可直接零售，也可和服务一起结算。</text>
          </view>
        </view>

        <view class="search-shell">
          <view class="search-accent">搜商品</view>
          <input
            v-model="productKeyword"
            class="search-input"
            placeholder="搜索商品名称 / 品牌"
          />
        </view>

        <scroll-view scroll-x class="product-tab-bar">
          <view class="product-tab-list">
            <view :class="['product-tab', activeProductCategoryId === 0 ? 'active' : '']" @click="activeProductCategoryId = 0">全部</view>
            <view
              v-for="category in productCategories"
              :key="category.ID"
              :class="['product-tab', activeProductCategoryId === category.ID ? 'active' : '']"
              @click="activeProductCategoryId = category.ID"
            >
              {{ category.name }}
            </view>
          </view>
        </scroll-view>

        <view v-if="filteredProducts.length === 0" class="empty-block">暂无可售商品</view>
        <view v-else class="product-grid">
          <view class="product-card" v-for="product in filteredProducts" :key="product.ID" @click="openProductPicker(product)">
            <view class="product-card-top">
              <text class="product-name">{{ product.name }}</text>
              <text class="product-price">{{ formatProductPrice(product) }}</text>
            </view>
            <view class="product-card-bottom">
              <text class="product-meta">{{ product.brand || product.category?.name || '零售商品' }}</text>
              <text class="product-meta">{{ getSellableSkus(product).length }}个可售规格</text>
            </view>
          </view>
        </view>

        <view v-if="cartItems.length > 0" class="cart-box">
          <view class="cart-head">
            <text class="cart-title">商品购物车</text>
            <text class="cart-total">¥{{ productSubtotal.toFixed(2) }}</text>
          </view>
          <view class="cart-item" v-for="item in cartItems" :key="item.sku_id || item.display_name">
            <view class="cart-info">
              <text class="cart-name">{{ item.display_name }}</text>
              <text class="cart-meta">¥{{ item.unit_price.toFixed(2) }} / 件</text>
            </view>
            <view class="cart-actions">
              <view class="step-btn" @click="changeCartQuantity(item, -1)">-</view>
              <text class="step-value">{{ item.quantity }}</text>
              <view class="step-btn" @click="changeCartQuantity(item, 1)">+</view>
              <text class="cart-amount">¥{{ (item.unit_price * item.quantity).toFixed(2) }}</text>
              <text class="cart-remove" @click="removeCartItem(item)">删除</text>
            </view>
          </view>
        </view>
      </view>

      <view class="section service-fold-section">
        <view class="service-fold-trigger" @click="servicePanelExpanded = !servicePanelExpanded">
          <view class="service-fold-copy">
            <text class="section-title no-margin">服务项目</text>
            <text class="section-desc service-fold-desc">
              {{ servicePanelExpanded
                ? (selectedPet ? '已展开，可继续为当前猫咪选择服务。' : '展开后先选猫咪，再添加服务项目。')
                : (hasServiceItem ? '已选服务，点击继续查看或修改。' : '点击下方箭头展开后再添加服务。') }}
            </text>
          </view>
          <view class="service-fold-arrow-wrap">
            <text :class="['service-fold-arrow', servicePanelExpanded ? 'open' : '']">⌄</text>
          </view>
        </view>

        <view v-if="servicePanelExpanded" :class="['service-fold-body', !selectedPet ? 'service-fold-body-disabled' : '']">
          <view v-if="!selectedPet" class="disabled-panel">
            <text class="disabled-icon">🐱</text>
            <text class="disabled-text">选择猫咪后可添加服务</text>
          </view>

          <template v-else>
            <view class="svc-picker" v-if="categoryTree.length > 0">
              <view class="svc-picker-sidebar">
                <view
                  v-for="cat in categoryTree"
                  :key="cat.ID"
                  :class="['sidebar-item', activeCategoryId === cat.ID ? 'active' : '']"
                  @click="selectCategory(cat.ID)"
                >
                  <text>{{ cat.name }}</text>
                </view>
              </view>
              <view class="svc-picker-main">
                <scroll-view scroll-x class="sub-tab-bar">
                  <view class="sub-tab-list">
                    <view :class="['sub-tab', activeSubCategoryId === 0 ? 'active' : '']" @click="selectSubCategory(0)">全部</view>
                    <view
                      v-for="sub in subCategories"
                      :key="sub.ID"
                      :class="['sub-tab', activeSubCategoryId === sub.ID ? 'active' : '']"
                      @click="selectSubCategory(sub.ID)"
                    >
                      {{ sub.name }}
                    </view>
                  </view>
                </scroll-view>
                <scroll-view scroll-y class="svc-item-list">
                  <view v-if="filteredServices.length === 0" class="svc-empty">暂无服务</view>
                  <view
                    v-for="service in filteredServices"
                    :key="service.ID"
                    :class="['svc-item', selectedServiceId === service.ID ? 'checked' : '']"
                    @click="selectService(service)"
                  >
                    <view class="svc-item-info">
                      <text class="svc-item-name">{{ service.name }}</text>
                      <text class="svc-item-cat">{{ service.duration }}分钟</text>
                    </view>
                    <view class="svc-item-right">
                      <text class="svc-item-price">¥{{ service.base_price }}</text>
                      <view :class="['svc-item-check', selectedServiceId === service.ID ? 'on' : '']"></view>
                    </view>
                  </view>
                </scroll-view>
              </view>
            </view>

            <view class="spec-section" v-if="selectedServiceId && specList.length > 0">
              <text class="spec-title">选择规格</text>
              <view class="spec-picker">
                <view
                  v-for="spec in specList"
                  :key="spec.ID"
                  :class="['spec-opt', selectedSpecId === spec.ID && !useCustomPrice ? 'active' : '']"
                  @click="selectSpec(spec)"
                >
                  <text class="spec-name">{{ spec.name }}</text>
                  <text class="spec-meta">¥{{ spec.price }}<text v-if="spec.duration" class="spec-dur"> · {{ spec.duration }}分钟</text></text>
                </view>
                <view :class="['spec-opt', 'custom', useCustomPrice ? 'active' : '']" @click="enableCustomPrice">
                  <text class="spec-name">自定义</text>
                  <text class="spec-meta">手动输入</text>
                </view>
              </view>
              <view class="custom-price-row" v-if="useCustomPrice">
                <text class="custom-price-label">输入金额</text>
                <input v-model="customPriceInput" type="digit" placeholder="0.00" class="custom-price-input" @input="onCustomPriceInput" />
                <text class="custom-price-hint">适用于团购或特殊价格</text>
              </view>
            </view>

            <view class="spec-section" v-else-if="selectedServiceId">
              <view class="custom-price-toggle" @click="useCustomPrice = !useCustomPrice">
                <text>{{ useCustomPrice ? '使用基础价格 ¥' + getBasePrice() : '自定义金额（美团团购等）' }}</text>
              </view>
              <view class="custom-price-row" v-if="useCustomPrice">
                <text class="custom-price-label">输入金额</text>
                <input v-model="customPriceInput" type="digit" placeholder="0.00" class="custom-price-input" @input="onCustomPriceInput" />
              </view>
            </view>
          </template>
        </view>
      </view>

      <view class="section" v-if="selectedPet || cartItems.length > 0 || selectedServiceId">
        <view class="section-title-row">
          <text class="section-title no-margin">附加费用</text>
          <view class="addon-add-btn" @click="onAddAddon">
            <text class="addon-add-icon">+</text>
          </view>
        </view>
        <view class="addon-grid">
          <view
            class="addon-item"
            :class="{ 'addon-item--deleting': longPressId === addon.id }"
            v-for="addon in addonInputs"
            :key="addon.id"
            @longpress="onLongPress(addon.id)"
          >
            <view v-if="longPressId === addon.id && !isDesktopInteraction" class="addon-delete-badge" @click.stop="onDeleteAddon(addon)">
              <text class="addon-delete-icon">−</text>
            </view>
            <view v-if="isDesktopInteraction" class="addon-delete-inline" @click.stop="onDeleteAddon(addon)">×</view>
            <text class="addon-label">{{ addon.name }}</text>
            <input v-model="addon.amount" type="digit" placeholder="0" class="addon-input" :disabled="longPressId === addon.id" />
          </view>
        </view>
        <view v-if="longPressId !== null && !isDesktopInteraction" class="addon-cancel-hint" @click="longPressId = null">
          <text class="addon-cancel-text">点击此处取消删除模式</text>
        </view>
      </view>

      <view class="section summary" v-if="hasChargeItems">
        <view class="section-head compact">
          <view>
            <text class="section-title">金额汇总</text>
            <text class="section-desc">服务、商品和附加费分开统计。</text>
          </view>
        </view>
        <view class="summary-row" v-if="serviceSubtotal > 0">
          <text>服务小计</text>
          <text>¥{{ serviceSubtotal.toFixed(2) }}</text>
        </view>
        <view class="summary-row" v-if="serviceDiscountAmount > 0">
          <text>服务优惠</text>
          <text class="discount">-¥{{ serviceDiscountAmount.toFixed(2) }}</text>
        </view>
        <view class="summary-row" v-if="productSubtotal > 0">
          <text>商品小计</text>
          <text>¥{{ productSubtotal.toFixed(2) }}</text>
        </view>
        <view class="summary-row" v-if="productDiscountAmount > 0">
          <text>商品优惠</text>
          <text class="discount">-¥{{ productDiscountAmount.toFixed(2) }}</text>
        </view>
        <view class="summary-row" v-if="addonTotal > 0">
          <text>附加费</text>
          <text>¥{{ addonTotal.toFixed(2) }}</text>
        </view>
        <view class="summary-row">
          <text>订单总价</text>
          <text>¥{{ totalAmount.toFixed(2) }}</text>
        </view>
        <view class="summary-row total">
          <text>应付</text>
          <text class="pay-amount">¥{{ payAmount.toFixed(2) }}</text>
        </view>
      </view>

      <view class="section" v-if="hasChargeItems">
        <view class="form-row">
          <text class="label">经手员工 <text v-if="hasServiceItem && !selectedStaff" class="required">*必选</text></text>
          <picker :range="staffNames" :value="selectedStaffIdx" @change="(e: any) => selectedStaffIdx = Number(e.detail.value)">
            <view :class="['picker', hasServiceItem && !selectedStaff ? 'picker-warn' : '']">{{ staffNames[selectedStaffIdx] || '请选择' }}</view>
          </picker>
        </view>
        <view class="staff-commission" v-if="selectedStaff">
          <text>服务提成 {{ selectedStaff.commission_rate }}% · 商品提成 {{ selectedStaff.product_commission_rate || 0 }}%</text>
        </view>
        <view class="form-row">
          <text class="label">备注</text>
          <input v-model="remark" placeholder="备注" class="input" />
        </view>
      </view>

      <button v-if="hasChargeItems" class="btn-submit" @click="onSubmit" :loading="submitting">
        {{ isEditing ? '保存修改' : '确认开单' }}
      </button>
    </view>
  </SideLayout>
</template>

<script setup lang="ts">
import SideLayout from '@/components/SideLayout.vue'
import { ref, computed, onMounted } from 'vue'
import { onLoad } from '@dcloudio/uni-app'
import { getCustomer, getCustomerList, getCustomerPets } from '@/api/customer'
import { getPetList } from '@/api/pet'
import { getProductCategories, getProductList } from '@/api/product'
import { createOrder, getOrder, updateOrder } from '@/api/order'
import { getAddonList, createAddon, deleteAddon, priceLookup } from '@/api/addon'
import { getServiceList, getPriceRules } from '@/api/service'
import { getCategoryTree } from '@/api/service-category'
import { getStaffList } from '@/api/staff'
import { getCustomerCard } from '@/api/member-card'
import { getAppointment } from '@/api/appointment'
import { useDesktopInteraction } from '@/utils/interaction'

type CustomerOption = Customer
type PetOption = Pet
type ProductOption = Product
type ProductSkuOption = ProductSKU

interface CartItem {
  product_id: number
  sku_id: number
  product_name: string
  spec_name: string
  display_name: string
  quantity: number
  unit_price: number
}

const customerKeyword = ref('')
const customerOptions = ref<CustomerOption[]>([])
const customerPets = ref<PetOption[]>([])
const selectedCustomer = ref<CustomerOption | null>(null)

const petKeyword = ref('')
const petSearchResults = ref<PetOption[]>([])
const selectedPet = ref<PetOption | null>(null)

const productKeyword = ref('')
const productList = ref<ProductOption[]>([])
const productCategories = ref<any[]>([])
const activeProductCategoryId = ref(0)
const cartItems = ref<CartItem[]>([])

const serviceList = ref<any[]>([])
const categoryTree = ref<any[]>([])
const activeCategoryId = ref(0)
const activeSubCategoryId = ref(0)
const selectedServiceId = ref(0)
const servicePanelExpanded = ref(false)
const servicePrice = ref(0)
const specList = ref<any[]>([])
const selectedSpecId = ref(0)
const useCustomPrice = ref(false)
const customPriceInput = ref('')

const staffList = ref<any[]>([])
const selectedStaffIdx = ref(0)
const addonInputs = ref<{ id: number; name: string; amount: string }[]>([])
const longPressId = ref<number | null>(null)
const remark = ref('')
const submitting = ref(false)
const memberBalance = ref(0)
const customerCard = ref<MemberCard | null>(null)
const prefillAppointmentId = ref(0)
const editingOrderId = ref(0)

let customerSearchTimer: ReturnType<typeof setTimeout> | null = null
let petSearchTimer: ReturnType<typeof setTimeout> | null = null

const { isDesktopInteraction } = useDesktopInteraction()

const isEditing = computed(() => editingOrderId.value > 0)
const selectedStaff = computed(() => selectedStaffIdx.value > 0 ? staffList.value[selectedStaffIdx.value - 1] : null)
const staffNames = computed(() => ['未选择', ...staffList.value.map((staff: any) => staff.name)])

const subCategories = computed(() => {
  const category = categoryTree.value.find((item: any) => item.ID === activeCategoryId.value)
  return category?.children || []
})

const filteredServices = computed(() => {
  let list = serviceList.value.filter((item: any) => item.status === 1)
  if (activeCategoryId.value > 0) {
    const subIds = subCategories.value.map((item: any) => item.ID)
    if (activeSubCategoryId.value > 0) {
      list = list.filter((item: any) => item.category_id === activeSubCategoryId.value)
    } else {
      list = list.filter((item: any) => item.category_id && subIds.includes(item.category_id))
    }
  }
  return list
})

const petOptions = computed(() => {
  if (selectedCustomer.value) {
    const keyword = petKeyword.value.trim().toLowerCase()
    if (!keyword) return customerPets.value
    return customerPets.value.filter((pet) => {
      const name = String(pet.name || '').toLowerCase()
      const breed = String(pet.breed || '').toLowerCase()
      return name.includes(keyword) || breed.includes(keyword)
    })
  }
  return petSearchResults.value
})

const filteredProducts = computed(() => {
  const keyword = productKeyword.value.trim().toLowerCase()
  return productList.value.filter((product) => {
    if (product.status !== 1) return false
    if (activeProductCategoryId.value > 0 && product.category_id !== activeProductCategoryId.value) return false
    if (getSellableSkus(product).length === 0) return false
    if (!keyword) return true
    const haystack = `${product.name} ${product.brand || ''}`.toLowerCase()
    return haystack.includes(keyword)
  })
})

const addonTotal = computed(() => addonInputs.value.reduce((sum, addon) => sum + (parseFloat(addon.amount) || 0), 0))
const serviceSubtotal = computed(() => selectedServiceId.value > 0 ? roundCurrency(servicePrice.value) : 0)
const productSubtotal = computed(() => roundCurrency(cartItems.value.reduce((sum, item) => sum + item.unit_price * item.quantity, 0)))
const totalAmount = computed(() => roundCurrency(serviceSubtotal.value + productSubtotal.value + addonTotal.value))

const serviceDiscountRate = computed(() => {
  const rate = Number(selectedCustomer.value?.discount_rate || 1)
  return rate > 0 && rate < 1 ? rate : 1
})
const productDiscountRate = computed(() => {
  const rate = Number(customerCard.value?.product_discount_rate || 1)
  return rate > 0 && rate < 1 ? rate : 1
})
const serviceDiscountAmount = computed(() => roundCurrency(serviceSubtotal.value - roundCurrency(serviceSubtotal.value * serviceDiscountRate.value)))
const productDiscountAmount = computed(() => roundCurrency(productSubtotal.value - roundCurrency(productSubtotal.value * productDiscountRate.value)))
const payAmount = computed(() => roundCurrency(totalAmount.value - serviceDiscountAmount.value - productDiscountAmount.value))
const hasServiceItem = computed(() => selectedServiceId.value > 0)
const hasChargeItems = computed(() => hasServiceItem.value || cartItems.value.length > 0)

onLoad((query) => {
  if (query?.appointment_id) {
    prefillAppointmentId.value = parseInt(String(query.appointment_id)) || 0
  }
  if (query?.order_id) {
    editingOrderId.value = parseInt(String(query.order_id)) || 0
  }
})

onMounted(async () => {
  await Promise.all([
    loadServiceData(),
    loadProductData(),
    loadStaffs(),
    loadAddons(),
  ])

  if (editingOrderId.value) {
    await prefillFromOrder(editingOrderId.value)
  } else if (prefillAppointmentId.value) {
    await prefillFromAppointment(prefillAppointmentId.value)
  }
})

async function loadServiceData() {
  try {
    const [serviceRes, categoryRes] = await Promise.all([
      getServiceList({ page: 1, page_size: 200 } as any),
      getCategoryTree(),
    ])
    serviceList.value = serviceRes.data?.list || []
    categoryTree.value = (categoryRes.data || []).filter((item: any) => item.status === 1)
    if (categoryTree.value.length > 0) {
      activeCategoryId.value = categoryTree.value[0].ID
    }
  } catch {}
}

async function loadProductData() {
  try {
    const [categoryRes, productRes] = await Promise.all([
      getProductCategories(),
      getProductList({ page: 1, page_size: 500 } as any),
    ])
    productCategories.value = Array.isArray(categoryRes.data) ? categoryRes.data.filter((item: any) => item.status === 1) : []
    productList.value = productRes.data?.list || []
  } catch {}
}

async function loadStaffs() {
  try {
    const res = await getStaffList({ page: 1, page_size: 50 } as any)
    staffList.value = (res.data?.list || []).filter((staff: any) => staff.status === 1)
  } catch {}
}

async function loadAddons() {
  try {
    const res = await getAddonList()
    if (Array.isArray(res.data)) {
      const amountMap = new Map(addonInputs.value.map((item) => [item.id, item.amount]))
      addonInputs.value = res.data.map((item: any) => ({
        id: item.ID,
        name: item.name,
        amount: amountMap.get(item.ID) || '',
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

function selectCategory(categoryId: number) {
  activeCategoryId.value = categoryId
  activeSubCategoryId.value = 0
}

function selectSubCategory(categoryId: number) {
  activeSubCategoryId.value = categoryId
}

async function searchCustomers() {
  if (customerSearchTimer) clearTimeout(customerSearchTimer)
  customerSearchTimer = setTimeout(async () => {
    const keyword = customerKeyword.value.trim()
    if (!keyword) {
      customerOptions.value = []
      return
    }
    try {
      const res = await getCustomerList({ page: 1, page_size: 20, keyword } as any)
      customerOptions.value = res.data?.list || []
    } catch {
      customerOptions.value = []
    }
  }, 250)
}

async function selectCustomer(customer: CustomerOption) {
  customerOptions.value = []
  customerKeyword.value = ''
  await applyCustomerSelection(customer)
}

async function applyCustomerSelection(customer: CustomerOption) {
  const nextId = Number(customer.ID || 0)
  const prevPetCustomerId = Number(selectedPet.value?.customer_id || 0)
  selectedCustomer.value = customer
  await Promise.all([
    loadCustomerPets(nextId),
    loadCustomerCard(nextId),
  ])
  if (selectedPet.value && prevPetCustomerId > 0 && prevPetCustomerId !== nextId) {
    clearPet()
  }
}

async function loadCustomerPets(customerId: number) {
  if (!customerId) {
    customerPets.value = []
    return
  }
  try {
    const res = await getCustomerPets(customerId)
    customerPets.value = Array.isArray(res.data) ? res.data : []
  } catch {
    customerPets.value = []
  }
}

async function loadCustomerCard(customerId: number) {
  memberBalance.value = 0
  customerCard.value = null
  if (!customerId) return
  try {
    const res = await getCustomerCard(customerId)
    customerCard.value = res.data || null
    memberBalance.value = Number(res.data?.balance || 0)
  } catch {}
}

function clearCustomer() {
  selectedCustomer.value = null
  customerKeyword.value = ''
  customerOptions.value = []
  customerPets.value = []
  customerCard.value = null
  memberBalance.value = 0
  if (selectedPet.value) {
    clearPet()
  }
}

async function searchPets() {
  if (selectedCustomer.value) {
    return
  }
  if (petSearchTimer) clearTimeout(petSearchTimer)
  petSearchTimer = setTimeout(async () => {
    const keyword = petKeyword.value.trim()
    if (!keyword) {
      petSearchResults.value = []
      return
    }
    try {
      const res = await getPetList({ page: 1, page_size: 20, keyword } as any)
      petSearchResults.value = res.data?.list || []
    } catch {
      petSearchResults.value = []
    }
  }, 250)
}

async function selectPet(pet: PetOption) {
  if (pet.customer_id && (!selectedCustomer.value || selectedCustomer.value.ID !== pet.customer_id)) {
    if (pet.customer) {
      await applyCustomerSelection(pet.customer as CustomerOption)
    } else {
      try {
        const res = await getCustomer(pet.customer_id)
        if (res.data) {
          await applyCustomerSelection(res.data as CustomerOption)
        }
      } catch {}
    }
  }

  selectedPet.value = pet
  petKeyword.value = ''
  petSearchResults.value = []

  if (selectedServiceId.value > 0 && pet.fur_level) {
    await lookupPrice(selectedServiceId.value, pet.fur_level)
  }
}

function clearPet() {
  selectedPet.value = null
  petKeyword.value = ''
  petSearchResults.value = []
  selectedServiceId.value = 0
  selectedSpecId.value = 0
  specList.value = []
  servicePrice.value = 0
  useCustomPrice.value = false
  customPriceInput.value = ''
}

function calcAge(birthDate: string): string {
  if (!birthDate) return ''
  const birth = new Date(birthDate)
  const now = new Date()
  const months = (now.getFullYear() - birth.getFullYear()) * 12 + (now.getMonth() - birth.getMonth())
  if (months < 1) return '不到1个月'
  if (months < 12) return `${months}个月`
  const years = Math.floor(months / 12)
  const remain = months % 12
  return remain > 0 ? `${years}岁${remain}个月` : `${years}岁`
}

async function selectService(service: any) {
  servicePanelExpanded.value = true
  selectedServiceId.value = service.ID
  selectedSpecId.value = 0
  servicePrice.value = Number(service.base_price || 0)
  useCustomPrice.value = false
  customPriceInput.value = ''
  try {
    const res = await getPriceRules(service.ID)
    const rules = res.data || []
    specList.value = rules.map((item: any) => ({ ...item, name: item.name || item.fur_level || item.pet_size || '规格' }))
    if (selectedPet.value?.fur_level && rules.length > 0) {
      const match = rules.find((item: any) => item.fur_level === selectedPet.value?.fur_level)
      if (match) {
        servicePrice.value = Number(match.price || 0)
        selectedSpecId.value = match.ID
      }
    }
  } catch {
    specList.value = []
  }
}

function selectSpec(spec: any) {
  selectedSpecId.value = spec.ID
  servicePrice.value = Number(spec.price || 0)
  useCustomPrice.value = false
  customPriceInput.value = ''
}

function enableCustomPrice() {
  useCustomPrice.value = true
  selectedSpecId.value = 0
  if (customPriceInput.value) {
    servicePrice.value = parseFloat(customPriceInput.value) || 0
  }
}

function onCustomPriceInput() {
  servicePrice.value = parseFloat(customPriceInput.value) || 0
}

function getBasePrice() {
  const service = serviceList.value.find((item: any) => item.ID === selectedServiceId.value)
  return service ? Number(service.base_price || 0).toFixed(2) : '0.00'
}

async function lookupPrice(serviceId: number, furLevel: string) {
  try {
    const res = await priceLookup(serviceId, furLevel)
    servicePrice.value = Number(res.data?.price || 0)
  } catch {
    const service = serviceList.value.find((item: any) => item.ID === serviceId)
    servicePrice.value = Number(service?.base_price || 0)
  }
}

function getSellableSkus(product: ProductOption) {
  return (product.skus || []).filter((sku) => sku.sellable)
}

function formatProductPrice(product: ProductOption) {
  const prices = getSellableSkus(product).map((sku) => Number(sku.price || 0))
  if (prices.length === 0) return '不可售'
  const min = Math.min(...prices)
  const max = Math.max(...prices)
  return min === max ? `¥${min.toFixed(2)}` : `¥${min.toFixed(2)} - ¥${max.toFixed(2)}`
}

function formatCartDisplayName(productName: string, specName: string) {
  return specName ? `${productName} · ${specName}` : productName
}

function openProductPicker(product: ProductOption) {
  const skus = getSellableSkus(product)
  if (skus.length === 0) {
    uni.showToast({ title: '该商品暂无可售规格', icon: 'none' })
    return
  }
  if (skus.length === 1) {
    addProductToCart(product, skus[0])
    return
  }
  uni.showActionSheet({
    itemList: skus.map((sku) => `${sku.spec_name || '默认规格'} · ¥${Number(sku.price || 0).toFixed(2)}`),
    success: ({ tapIndex }) => {
      const selectedSku = skus[tapIndex]
      if (selectedSku) {
        addProductToCart(product, selectedSku)
      }
    },
  })
}

function addProductToCart(product: ProductOption, sku: ProductSkuOption) {
  const existing = cartItems.value.find((item) => item.sku_id === sku.ID)
  if (existing) {
    existing.quantity += 1
  } else {
    cartItems.value.push({
      product_id: product.ID,
      sku_id: sku.ID,
      product_name: product.name,
      spec_name: sku.spec_name || '',
      display_name: formatCartDisplayName(product.name, sku.spec_name || ''),
      quantity: 1,
      unit_price: Number(sku.price || 0),
    })
  }
  uni.showToast({ title: '已加入商品', icon: 'success' })
}

function changeCartQuantity(item: CartItem, delta: number) {
  const nextQuantity = item.quantity + delta
  if (nextQuantity <= 0) {
    removeCartItem(item)
    return
  }
  item.quantity = nextQuantity
}

function removeCartItem(target: CartItem) {
  cartItems.value = cartItems.value.filter((item) => item !== target)
}

function onAddAddon() {
  uni.showModal({
    title: '添加附加费类型',
    editable: true,
    placeholderText: '输入名称',
    success: async (res) => {
      if (res.confirm && res.content?.trim()) {
        try {
          await createAddon({ name: res.content.trim(), default_price: 0, is_variable: true } as any)
          await loadAddons()
        } catch {
          addonInputs.value.push({ id: Date.now(), name: res.content.trim(), amount: '' })
        }
      }
    },
  })
}

function onLongPress(id: number) {
  longPressId.value = id
}

function onDeleteAddon(addon: { id: number; name: string }) {
  uni.showModal({
    title: '删除附加费',
    content: `删除「${addon.name}」？`,
    confirmColor: '#EF4444',
    success: async (res) => {
      if (res.confirm) {
        if (addon.id > 0) {
          try {
            await deleteAddon(addon.id)
            await loadAddons()
          } catch {}
        } else {
          addonInputs.value = addonInputs.value.filter((item) => item.id !== addon.id)
        }
      }
      longPressId.value = null
    },
  })
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
        servicePanelExpanded.value = true
        if (typeof firstService.price === 'number' && firstService.price > 0) {
          servicePrice.value = firstService.price
        }
      }
    }

    if (appt?.staff_id) {
      const idx = staffList.value.findIndex((staff: any) => staff.ID === appt.staff_id)
      if (idx >= 0) {
        selectedStaffIdx.value = idx + 1
      }
    }

    if (appt?.notes) {
      remark.value = appt.notes
    }
  } catch {
    uni.showToast({ title: '预约信息带入失败', icon: 'none' })
  }
}

async function prefillFromOrder(orderId: number) {
  try {
    const res = await getOrder(orderId)
    const order = res.data
    if (!order) {
      uni.showToast({ title: '订单不存在', icon: 'none' })
      return
    }
    if (order.status !== 0) {
      uni.showToast({ title: '仅待付款订单可修改', icon: 'none' })
      return
    }

    if (order.customer) {
      await applyCustomerSelection(order.customer)
    }

    if (order.pet) {
      await selectPet(order.pet)
    }

    const serviceItem = (order.items || []).find((item: any) => item.item_type === 1)
    if (serviceItem?.item_id) {
      const service = serviceList.value.find((item: any) => item.ID === serviceItem.item_id)
      if (service) {
        await selectService(service)
        servicePanelExpanded.value = true
        const nextPrice = Number(serviceItem.unit_price || 0)
        if (nextPrice > 0 && nextPrice !== Number(servicePrice.value || 0)) {
          useCustomPrice.value = true
          customPriceInput.value = String(nextPrice)
          servicePrice.value = nextPrice
        }
      }
    }

    cartItems.value = []
    for (const item of order.items || []) {
      if (item.item_type !== 2) continue
      const found = findProductBySku(Number(item.item_id || 0))
      if (found) {
        cartItems.value.push({
          product_id: found.product.ID,
          sku_id: found.sku.ID,
          product_name: found.product.name,
          spec_name: found.sku.spec_name || '',
          display_name: formatCartDisplayName(found.product.name, found.sku.spec_name || ''),
          quantity: Number(item.quantity || 1),
          unit_price: Number(item.unit_price || 0),
        })
        continue
      }
      const [productName, specName] = splitDisplayName(item.name)
      cartItems.value.push({
        product_id: 0,
        sku_id: Number(item.item_id || 0),
        product_name: productName,
        spec_name: specName,
        display_name: item.name,
        quantity: Number(item.quantity || 1),
        unit_price: Number(item.unit_price || 0),
      })
    }

    const addonAmountMap = new Map<string, string>()
    for (const item of order.items || []) {
      if (item.item_type === 3) {
        addonAmountMap.set(item.name, String(item.amount || item.unit_price || 0))
      }
    }
    for (const addon of addonInputs.value) {
      addon.amount = addonAmountMap.get(addon.name) || ''
      addonAmountMap.delete(addon.name)
    }
    for (const [name, amount] of addonAmountMap.entries()) {
      addonInputs.value.push({ id: Date.now() + addonInputs.value.length, name, amount })
    }

    if (order.staff_id) {
      const idx = staffList.value.findIndex((staff: any) => staff.ID === order.staff_id)
      if (idx >= 0) {
        selectedStaffIdx.value = idx + 1
      }
    }
    remark.value = order.remark || ''
  } catch {
    uni.showToast({ title: '订单信息带入失败', icon: 'none' })
  }
}

function findProductBySku(skuId: number) {
  for (const product of productList.value) {
    const sku = (product.skus || []).find((item) => item.ID === skuId)
    if (sku) {
      return { product, sku }
    }
  }
  return null
}

function splitDisplayName(name: string) {
  const parts = String(name || '').split(' · ')
  if (parts.length < 2) return [name, '']
  return [parts[0], parts.slice(1).join(' · ')]
}

async function onSubmit() {
  if (hasServiceItem.value && !selectedPet.value) {
    uni.showToast({ title: '请选择猫咪后再添加服务', icon: 'none' })
    return
  }
  if (hasServiceItem.value && !selectedStaff.value) {
    uni.showToast({ title: '请选择洗护师', icon: 'none' })
    return
  }
  if (!hasServiceItem.value && cartItems.value.length === 0) {
    uni.showToast({ title: '请添加商品或服务', icon: 'none' })
    return
  }

  submitting.value = true
  try {
    const addons = addonInputs.value
      .filter((addon) => parseFloat(addon.amount) > 0)
      .map((addon) => ({ name: addon.name, amount: parseFloat(addon.amount) }))

    const items: any[] = []
    if (hasServiceItem.value) {
      items.push({
        item_type: 1,
        item_id: selectedServiceId.value,
        name: serviceList.value.find((item: any) => item.ID === selectedServiceId.value)?.name || '服务项目',
        quantity: 1,
        unit_price: Number(servicePrice.value || 0),
      })
    }
    for (const item of cartItems.value) {
      items.push({
        item_type: 2,
        item_id: item.sku_id,
        name: item.display_name,
        quantity: item.quantity,
        unit_price: Number(item.unit_price || 0),
      })
    }

    const payload = {
      pet_id: selectedPet.value?.ID || undefined,
      customer_id: selectedCustomer.value?.ID || selectedPet.value?.customer_id || undefined,
      staff_id: selectedStaff.value?.ID || undefined,
      remark: remark.value,
      addons,
      items,
    } as any

    const orderRes = isEditing.value
      ? await updateOrder(editingOrderId.value, payload)
      : await createOrder(payload)

    uni.showToast({ title: isEditing.value ? '已保存修改' : '开单成功', icon: 'success' })
    const orderId = orderRes.data?.ID
    setTimeout(() => {
      uni.redirectTo({ url: `/pages/order/detail?id=${orderId}` })
    }, 500)
  } catch (error: any) {
    uni.showToast({ title: error?.msg || error?.message || (isEditing.value ? '保存失败' : '开单失败'), icon: 'none' })
  } finally {
    submitting.value = false
  }
}

function roundCurrency(value: number) {
  return Math.round(Number(value || 0) * 100) / 100
}
</script>

<style scoped>
.page { padding: 24rpx; }
.section {
  background: #fff;
  border-radius: 20rpx;
  padding: 24rpx;
  margin-bottom: 16rpx;
  box-shadow: 0 8rpx 24rpx rgba(15, 23, 42, 0.04);
}
.section-disabled { opacity: 0.76; }
.service-fold-section { padding-top: 20rpx; }
.section-head {
  display: flex;
  justify-content: space-between;
  gap: 16rpx;
  align-items: flex-start;
  margin-bottom: 18rpx;
}
.section-head.compact { margin-bottom: 10rpx; }
.section-title {
  display: block;
  font-size: 30rpx;
  font-weight: 700;
  color: #111827;
}
.section-title.no-margin { margin-bottom: 0; }
.section-desc {
  display: block;
  margin-top: 6rpx;
  font-size: 24rpx;
  color: #6B7280;
  line-height: 1.5;
}
.section-badge {
  flex-shrink: 0;
  padding: 8rpx 18rpx;
  border-radius: 999rpx;
  background: #EEF2FF;
  color: #4F46E5;
  font-size: 22rpx;
  font-weight: 700;
}
.selector-block + .selector-block {
  margin-top: 22rpx;
  padding-top: 22rpx;
  border-top: 1rpx solid #F1F5F9;
}
.selector-head {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12rpx;
}
.selector-title {
  font-size: 26rpx;
  font-weight: 600;
  color: #374151;
}
.selector-link {
  min-width: 112rpx;
  height: 52rpx;
  padding: 0 18rpx;
  border-radius: 999rpx;
  background: #EEF2FF;
  color: #4F46E5;
  font-size: 22rpx;
  font-weight: 700;
  display: inline-flex;
  align-items: center;
  justify-content: center;
}
.search-shell {
  display: flex;
  align-items: center;
  gap: 16rpx;
  width: 100%;
  box-sizing: border-box;
  min-height: 96rpx;
  padding: 0 22rpx;
  background: linear-gradient(180deg, #FFFFFF 0%, #F8FAFF 100%);
  border: 2rpx solid #D9E3F4;
  border-radius: 22rpx;
  box-shadow: 0 10rpx 24rpx rgba(79, 70, 229, 0.06);
}
.search-shell:focus-within {
  border-color: #A5B4FC;
  box-shadow: 0 12rpx 28rpx rgba(79, 70, 229, 0.12);
}
.search-accent {
  flex-shrink: 0;
  min-width: 104rpx;
  height: 48rpx;
  padding: 0 16rpx;
  border-radius: 999rpx;
  background: #EEF2FF;
  color: #4F46E5;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  font-size: 22rpx;
  font-weight: 700;
  letter-spacing: 1rpx;
}
.search-input {
  flex: 1;
  width: auto;
  box-sizing: border-box;
  background: transparent;
  border: 0;
  border-radius: 0;
  padding: 0;
  font-size: 28rpx;
  color: #111827;
  overflow: visible;
}
.search-input :deep(.uni-input-wrapper) {
  min-height: 88rpx;
  padding: 0;
  box-sizing: border-box;
  display: flex;
  align-items: center;
  background: transparent;
}
.search-input :deep(.uni-input-form) {
  flex: 1;
  display: flex;
  align-items: center;
}
.search-input :deep(.uni-input-input) {
  width: 100%;
  min-height: 44rpx;
  line-height: 44rpx;
  font-size: 28rpx;
  color: #111827;
}
.search-input :deep(.uni-input-placeholder) {
  font-size: 26rpx;
  color: #9CA3AF;
}
.option-list {
  margin-top: 12rpx;
  display: flex;
  flex-direction: column;
  gap: 10rpx;
}
.option-card {
  border: 2rpx solid #E5E7EB;
  border-radius: 20rpx;
  padding: 20rpx 22rpx;
  background: #fff;
  box-shadow: 0 8rpx 20rpx rgba(15, 23, 42, 0.04);
}
.option-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 12rpx;
}
.option-row.subtle {
  justify-content: flex-start;
  margin-top: 6rpx;
  flex-wrap: wrap;
}
.option-name {
  font-size: 28rpx;
  font-weight: 600;
  color: #111827;
}
.option-meta {
  font-size: 24rpx;
  color: #6B7280;
}
.option-tags,
.selected-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 8rpx;
  margin-top: 10rpx;
}
.option-tag,
.selected-tag {
  font-size: 22rpx;
  color: #4F46E5;
  background: #EEF2FF;
  border-radius: 999rpx;
  padding: 6rpx 14rpx;
}
.selected-tag.warn { background: #FEF2F2; color: #DC2626; }
.selected-card {
  background: linear-gradient(180deg, #F8FAFF, #FFFFFF);
  border: 2rpx solid #DCE6FF;
  border-radius: 18rpx;
  padding: 20rpx;
}
.selected-pet-card {
  background: linear-gradient(180deg, #EEF2FF, #FFFFFF);
}
.selected-top {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 12rpx;
}
.selected-name {
  font-size: 30rpx;
  font-weight: 700;
  color: #1F2937;
}
.selected-sub {
  font-size: 24rpx;
  color: #6B7280;
}
.alert-box {
  margin-top: 12rpx;
  display: flex;
  gap: 8rpx;
  padding: 12rpx;
  background: #FEF3C7;
  border-radius: 12rpx;
}
.alert-icon {
  color: #D97706;
  font-weight: 700;
}
.alert-text {
  font-size: 24rpx;
  color: #92400E;
  line-height: 1.5;
  flex: 1;
}
.inline-hint,
.disabled-panel {
  text-align: center;
  color: #6B7280;
  background: #F8FAFC;
  border: 2rpx dashed #CBD5E1;
  border-radius: 16rpx;
  padding: 24rpx;
}
.disabled-panel {
  display: flex;
  flex-direction: column;
  gap: 10rpx;
  align-items: center;
}
.service-fold-trigger {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 18rpx;
}
.service-fold-copy {
  flex: 1;
  min-width: 0;
}
.service-fold-desc {
  margin-top: 8rpx;
}
.service-fold-arrow-wrap {
  width: 72rpx;
  height: 72rpx;
  flex-shrink: 0;
  border-radius: 50%;
  background: linear-gradient(180deg, #EEF2FF 0%, #E0E7FF 100%);
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: inset 0 0 0 2rpx rgba(99, 102, 241, 0.08);
}
.service-fold-arrow {
  font-size: 34rpx;
  line-height: 1;
  color: #4F46E5;
  font-weight: 700;
  transform: rotate(0deg);
  transition: transform 0.2s ease;
}
.service-fold-arrow.open {
  transform: rotate(180deg);
}
.service-fold-body {
  margin-top: 20rpx;
  padding-top: 20rpx;
  border-top: 1rpx solid #EEF2F7;
}
.service-fold-body-disabled .disabled-panel {
  margin-top: 4rpx;
}
.disabled-icon { font-size: 44rpx; }
.disabled-text { font-size: 26rpx; }
.svc-picker {
  display: flex;
  border: 2rpx solid #E5E7EB;
  border-radius: 18rpx;
  overflow: hidden;
  height: 600rpx;
}
.svc-picker-sidebar {
  width: 160rpx;
  min-width: 160rpx;
  background: #F8FAFC;
  border-right: 2rpx solid #E5E7EB;
  overflow-y: auto;
}
.sidebar-item {
  padding: 28rpx 16rpx;
  font-size: 26rpx;
  color: #6B7280;
  text-align: center;
  border-left: 6rpx solid transparent;
}
.sidebar-item.active {
  background: #fff;
  color: #111827;
  font-weight: 700;
  border-left-color: #4F46E5;
}
.svc-picker-main {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-width: 0;
  background: #fff;
}
.sub-tab-bar {
  white-space: nowrap;
  border-bottom: 2rpx solid #F1F5F9;
}
.sub-tab-list {
  display: inline-flex;
  gap: 12rpx;
  padding: 16rpx 16rpx 0;
}
.sub-tab {
  display: inline-block;
  padding: 12rpx 24rpx;
  border-radius: 999rpx;
  font-size: 24rpx;
  color: #6B7280;
  margin-bottom: 12rpx;
}
.sub-tab.active {
  background: #EEF2FF;
  color: #4F46E5;
  font-weight: 700;
}
.svc-item-list {
  flex: 1;
  overflow-y: auto;
}
.svc-empty,
.empty-block {
  text-align: center;
  color: #94A3B8;
  padding: 48rpx 0;
  font-size: 26rpx;
}
.svc-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 20rpx 24rpx;
  border-bottom: 1rpx solid #F3F4F6;
}
.svc-item.checked { background: #EEF2FF; }
.svc-item-info { flex: 1; min-width: 0; }
.svc-item-name {
  display: block;
  font-size: 28rpx;
  font-weight: 600;
  color: #111827;
}
.svc-item-cat {
  display: block;
  font-size: 22rpx;
  color: #94A3B8;
  margin-top: 4rpx;
}
.svc-item-right {
  display: flex;
  align-items: center;
  gap: 16rpx;
  margin-left: 16rpx;
}
.svc-item-price {
  font-size: 30rpx;
  font-weight: 700;
  color: #4F46E5;
}
.svc-item-check {
  width: 40rpx;
  height: 40rpx;
  border-radius: 50%;
  border: 3rpx solid #CBD5E1;
  box-sizing: border-box;
}
.svc-item-check.on {
  position: relative;
  border-color: #4F46E5;
  background: #4F46E5;
}
.svc-item-check.on::after {
  content: '';
  position: absolute;
  left: 50%;
  top: 45%;
  width: 12rpx;
  height: 20rpx;
  border: solid #fff;
  border-width: 0 3rpx 3rpx 0;
  transform: translate(-50%, -50%) rotate(45deg);
}
.spec-section {
  margin-top: 18rpx;
  padding-top: 18rpx;
  border-top: 1rpx solid #F3F4F6;
}
.spec-title {
  display: block;
  font-size: 26rpx;
  color: #6B7280;
  margin-bottom: 12rpx;
}
.spec-picker {
  display: flex;
  flex-wrap: wrap;
  gap: 12rpx;
}
.spec-opt {
  min-width: 180rpx;
  padding: 16rpx 22rpx;
  border-radius: 14rpx;
  border: 2rpx solid #E5E7EB;
  background: #F8FAFC;
}
.spec-opt.active {
  border-color: #4F46E5;
  background: #EEF2FF;
}
.spec-opt.custom { border-style: dashed; }
.spec-opt.custom.active { border-style: solid; }
.spec-name {
  display: block;
  font-size: 26rpx;
  font-weight: 600;
  color: #374151;
}
.spec-meta {
  display: block;
  margin-top: 6rpx;
  font-size: 24rpx;
  color: #4F46E5;
}
.spec-dur { color: #94A3B8; }
.custom-price-toggle {
  text-align: center;
  padding: 16rpx;
  border-radius: 12rpx;
  background: #EEF2FF;
  border: 2rpx dashed #C7D2FE;
  color: #4F46E5;
  font-size: 26rpx;
}
.custom-price-row {
  display: flex;
  align-items: center;
  gap: 12rpx;
  flex-wrap: wrap;
  margin-top: 14rpx;
}
.custom-price-label {
  font-size: 26rpx;
  color: #374151;
  font-weight: 600;
}
.custom-price-input {
  width: 220rpx;
  padding: 0 16rpx;
  height: 64rpx;
  border-radius: 12rpx;
  border: 2rpx solid #C7D2FE;
  background: #F8FAFC;
  font-size: 30rpx;
  color: #4F46E5;
  text-align: center;
  box-sizing: border-box;
}
.custom-price-hint {
  font-size: 22rpx;
  color: #94A3B8;
}
.product-tab-bar {
  white-space: nowrap;
  margin-top: 14rpx;
}
.product-tab-list {
  display: inline-flex;
  gap: 10rpx;
}
.product-tab {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  padding: 10rpx 22rpx;
  border-radius: 999rpx;
  background: #F3F4F6;
  color: #6B7280;
  font-size: 24rpx;
}
.product-tab.active {
  background: #4F46E5;
  color: #fff;
}
.product-grid {
  display: flex;
  flex-direction: column;
  gap: 10rpx;
  margin-top: 14rpx;
}
.product-card {
  border: 2rpx solid #E5E7EB;
  border-radius: 16rpx;
  padding: 18rpx 20rpx;
}
.product-card-top,
.product-card-bottom {
  display: flex;
  justify-content: space-between;
  gap: 12rpx;
  align-items: center;
}
.product-card-bottom { margin-top: 8rpx; }
.product-name {
  font-size: 28rpx;
  font-weight: 600;
  color: #111827;
}
.product-price {
  font-size: 28rpx;
  font-weight: 700;
  color: #4F46E5;
}
.product-meta {
  font-size: 22rpx;
  color: #94A3B8;
}
.cart-box {
  margin-top: 16rpx;
  padding: 18rpx;
  border-radius: 16rpx;
  background: #F8FAFC;
  border: 2rpx solid #E2E8F0;
}
.cart-head {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 10rpx;
}
.cart-title {
  font-size: 26rpx;
  font-weight: 700;
  color: #111827;
}
.cart-total {
  font-size: 30rpx;
  font-weight: 800;
  color: #4F46E5;
}
.cart-item + .cart-item {
  margin-top: 12rpx;
  padding-top: 12rpx;
  border-top: 1rpx solid #E2E8F0;
}
.cart-info {
  display: flex;
  justify-content: space-between;
  gap: 12rpx;
  align-items: center;
}
.cart-name {
  font-size: 26rpx;
  font-weight: 600;
  color: #1F2937;
  flex: 1;
}
.cart-meta {
  font-size: 22rpx;
  color: #94A3B8;
}
.cart-actions {
  margin-top: 10rpx;
  display: flex;
  align-items: center;
  gap: 14rpx;
  flex-wrap: wrap;
}
.step-btn {
  width: 56rpx;
  height: 56rpx;
  border-radius: 16rpx;
  background: linear-gradient(135deg, #EEF2FF, #E0E7FF);
  color: #4338CA;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 30rpx;
  font-weight: 700;
  box-shadow: inset 0 0 0 2rpx rgba(99, 102, 241, 0.08);
}
.step-value {
  min-width: 44rpx;
  text-align: center;
  font-size: 26rpx;
  font-weight: 600;
  color: #1F2937;
}
.cart-amount {
  margin-left: auto;
  font-size: 28rpx;
  font-weight: 700;
  color: #4F46E5;
}
.cart-remove {
  font-size: 24rpx;
  color: #EF4444;
}
.section-title-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16rpx;
}
.addon-grid {
  display: flex;
  flex-wrap: wrap;
  gap: 16rpx;
}
.addon-item {
  display: flex;
  align-items: center;
  gap: 8rpx;
  width: calc(50% - 8rpx);
  position: relative;
  background: #F8FAFC;
  border: 1rpx solid #E5E7EB;
  border-radius: 18rpx;
  padding: 16rpx;
  box-sizing: border-box;
}
.addon-item--deleting { background: #FFF1F2; border-color: #FECDD3; }
.addon-delete-inline {
  position: absolute;
  top: -12rpx;
  right: -12rpx;
  width: 40rpx;
  height: 40rpx;
  border-radius: 50%;
  background: #111827;
  color: #fff;
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 2;
  font-size: 24rpx;
  font-weight: 700;
}
.addon-delete-badge {
  position: absolute;
  top: -12rpx;
  left: -12rpx;
  width: 40rpx;
  height: 40rpx;
  border-radius: 50%;
  background: #EF4444;
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 2;
}
.addon-delete-icon { color: #fff; font-size: 28rpx; font-weight: 700; }
.addon-label { font-size: 24rpx; color: #374151; width: 100rpx; }
.addon-input {
  flex: 1;
  background: #fff;
  border-radius: 12rpx;
  border: 2rpx solid #E5E7EB;
  min-height: 60rpx;
  padding: 0 14rpx;
  text-align: right;
  font-size: 26rpx;
}
.addon-add-btn {
  width: 56rpx;
  height: 56rpx;
  border-radius: 50%;
  background: linear-gradient(135deg, #6366F1, #4F46E5);
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: 0 10rpx 22rpx rgba(79, 70, 229, 0.22);
}
.addon-add-icon {
  color: #fff;
  font-size: 32rpx;
  font-weight: 700;
}
.addon-cancel-hint {
  margin-top: 12rpx;
  padding: 12rpx;
  text-align: center;
  border: 1rpx dashed #FECDD3;
  border-radius: 12rpx;
}
.addon-cancel-text { font-size: 24rpx; color: #EF4444; }
.summary-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 10rpx 0;
  font-size: 26rpx;
  color: #374151;
}
.summary-row.total {
  margin-top: 8rpx;
  padding-top: 16rpx;
  border-top: 1rpx solid #E5E7EB;
  font-weight: 700;
}
.discount { color: #059669; }
.pay-amount {
  font-size: 36rpx;
  font-weight: 800;
  color: #4F46E5;
}
.form-row {
  display: flex;
  align-items: center;
  gap: 16rpx;
  padding: 16rpx 0;
  border-bottom: 1rpx solid #F3F4F6;
}
.form-row:last-child { border-bottom: none; }
.label {
  width: 160rpx;
  font-size: 28rpx;
  color: #374151;
}
.required {
  font-size: 22rpx;
  color: #EF4444;
  font-weight: 700;
}
.picker,
.input {
  flex: 1;
  font-size: 28rpx;
  color: #111827;
  min-height: 88rpx;
  border-radius: 16rpx;
  border: 2rpx solid #E5E7EB;
  background: #F8FAFC;
  padding: 0 22rpx;
  box-sizing: border-box;
  display: flex;
  align-items: center;
  justify-content: center;
}
.picker-warn { color: #EF4444; }
.staff-commission {
  padding: 8rpx 0;
  font-size: 24rpx;
  color: #6B7280;
}
.btn-submit {
  margin: 20rpx 0 0;
  min-height: 96rpx;
  background: linear-gradient(135deg, #4F46E5, #6366F1);
  color: #fff;
  border-radius: 22rpx;
  font-size: 29rpx;
  font-weight: 800;
  line-height: 1.2;
  letter-spacing: 2rpx;
  box-shadow: 0 14rpx 28rpx rgba(79, 70, 229, 0.24);
}
</style>
