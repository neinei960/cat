<template>
  <SideLayout>
  <view class="page">
    <view v-if="loading" class="state">加载中...</view>
    <view v-else-if="!appt" class="state">预约不存在</view>
    <template v-else>
      <view class="summary-card">
        <text class="summary-title">{{ isEditing ? '修改合单' : '预约合单确认' }}</text>
        <text class="summary-line">{{ appt.date }} {{ appt.start_time }} - {{ appt.end_time }}</text>
        <text class="summary-line">客户：{{ appt.customer?.nickname || appt.customer?.phone || '-' }}</text>
        <text class="summary-line">洗护师：{{ appt.staff?.name || '待分配' }}</text>
        <text class="summary-line">宠物：{{ petSummary }}</text>
        <text class="summary-amount">{{ isEditing ? '修改后应付' : '预计生成 1 单' }} · ¥{{ totalAmount.toFixed(2) }}</text>
      </view>

      <view class="draft-list">
        <view class="draft-card" v-for="(draft, di) in drafts" :key="draft.petId">
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

          <!-- 服务列表（可编辑） -->
          <view class="service-list">
            <view class="service-row" v-for="(svc, si) in draft.services" :key="`${draft.petId}-${svc.service_id}-${si}`">
              <text class="service-name">{{ svc.service_name }}</text>
              <view class="service-edit">
                <input
                  v-if="isEditingPrice(di, si)"
                  v-model="editPriceValue"
                  type="digit"
                  class="service-price-input"
                  :focus="true"
                  confirm-type="done"
                  @blur="saveInlinePrice"
                  @confirm="saveInlinePrice"
                />
                <text v-else class="service-price" @click="editPrice(di, si)">¥{{ svc.price }}</text>
                <text class="service-dur">{{ svc.duration }}分钟</text>
                <text class="service-del" @click="removeService(di, si)">✕</text>
              </view>
            </view>
            <view class="add-service-row" @click="openAddService(di)">
              <text class="add-service-text">+ 添加服务</text>
            </view>
          </view>

          <!-- 商品列表 -->
          <view class="product-list" v-if="draft.products.length > 0">
            <view class="product-row" v-for="(prod, pi) in draft.products" :key="`${draft.petId}-prod-${pi}`">
              <text class="product-name">{{ prod.name }}</text>
              <view class="product-edit">
                <text class="product-price">¥{{ prod.price }}</text>
                <view class="qty-ctrl">
                  <text class="qty-btn" @click="changeProductQty(di, pi, -1)">−</text>
                  <text class="qty-val">{{ prod.quantity }}</text>
                  <text class="qty-btn" @click="changeProductQty(di, pi, 1)">+</text>
                </view>
                <text class="product-del" @click="removeProduct(di, pi)">✕</text>
              </view>
            </view>
          </view>
          <view class="add-service-row" @click="openAddProduct(di)">
            <text class="add-service-text">+ 添加商品</text>
          </view>
        </view>
      </view>

      <view class="notes-card">
        <view class="notes-head">
          <text class="notes-title">预约备注</text>
          <text class="notes-tip">保存合单时同步更新</text>
        </view>
        <textarea
          v-model="noteDraft"
          class="notes-input"
          auto-height
          maxlength="300"
          placeholder="填写预约备注，如送达时间、注意事项"
        />
      </view>

      <view class="submit-bar">
        <button class="submit-btn" :loading="submitting" @click="submitBatch">{{ isEditing ? '保存修改' : '确认生成订单' }}</button>
      </view>
      <!-- 添加服务弹窗（按分类） -->
      <view v-if="showAddService" class="modal-mask" @click="showAddService = false">
        <view class="modal-body modal-body-tall" @click.stop>
          <text class="modal-title">添加服务</text>
          <!-- 一级分类 tabs -->
          <scroll-view scroll-x class="cat-tabs" show-scrollbar="false">
            <view class="cat-tabs-inner">
              <view
                v-for="cat in topCategories"
                :key="cat.ID"
                :class="['cat-tab', activeCat1 === cat.ID ? 'cat-tab-active' : '']"
                @click="activeCat1 = cat.ID; activeCat2 = 0"
              >{{ cat.name }}</view>
            </view>
          </scroll-view>
          <!-- 二级分类 tabs -->
          <scroll-view scroll-x class="cat-tabs cat-tabs-sub" show-scrollbar="false" v-if="subCategories.length">
            <view class="cat-tabs-inner">
              <view
                :class="['cat-tab cat-tab-sub', activeCat2 === 0 ? 'cat-tab-active' : '']"
                @click="activeCat2 = 0"
              >全部</view>
              <view
                v-for="cat in subCategories"
                :key="cat.ID"
                :class="['cat-tab cat-tab-sub', activeCat2 === cat.ID ? 'cat-tab-active' : '']"
                @click="activeCat2 = cat.ID"
              >{{ cat.name }}</view>
            </view>
          </scroll-view>
          <!-- 服务列表 -->
          <view v-if="filteredServices.length === 0" class="modal-empty">该分类下暂无服务</view>
          <scroll-view scroll-y class="service-pick-list">
            <view v-for="svc in filteredServices" :key="svc.ID" class="service-pick-group">
              <view class="service-pick-item" @click="toggleServiceExpand(svc)">
                <text class="service-pick-name">{{ svc.name }}</text>
                <text class="service-pick-arrow" v-if="svc.price_rules?.length">{{ expandedServiceId === svc.ID ? '▾' : '▸' }}</text>
                <text class="service-pick-price" v-else @click.stop="addService(svc)">¥{{ svc.base_price }} · {{ svc.duration }}分钟</text>
              </view>
              <!-- 价格规格展开 -->
              <view v-if="expandedServiceId === svc.ID && svc.price_rules?.length" class="price-rules">
                <view
                  class="price-rule-item"
                  v-for="rule in svc.price_rules"
                  :key="rule.ID"
                  @click="addServiceWithRule(svc, rule)"
                >
                  <text class="rule-level">{{ rule.name || rule.fur_level || '标准' }}</text>
                  <text class="rule-price">¥{{ rule.price }}{{ (rule.duration || svc.duration) ? ` · ${rule.duration || svc.duration}分钟` : '' }}</text>
                </view>
              </view>
            </view>
          </scroll-view>
        </view>
      </view>
      <!-- 添加商品弹窗 -->
      <view v-if="showAddProduct" class="modal-mask" @click="showAddProduct = false">
        <view class="modal-body modal-body-tall" @click.stop>
          <text class="modal-title">添加商品</text>
          <input
            :value="productKeyword"
            class="modal-search"
            placeholder="搜索商品名称 / 品牌 / 分类 / 规格"
            confirm-type="search"
            @input="onProductKeywordInput"
            @confirm="onProductKeywordConfirm"
          />
          <scroll-view scroll-x class="cat-tabs" show-scrollbar="false">
            <view class="cat-tabs-inner">
              <view
                :class="['cat-tab', activeProductCat === 0 ? 'cat-tab-active' : '']"
                @click="setProductCategory(0)"
              >全部</view>
              <view
                v-for="cat in productCategories"
                :key="cat.ID"
                :class="['cat-tab', activeProductCat === cat.ID ? 'cat-tab-active' : '']"
                @click="setProductCategory(cat.ID)"
              >{{ cat.name }}</view>
            </view>
          </scroll-view>
          <view v-if="productLoading" class="modal-empty">搜索中...</view>
          <view v-else-if="filteredProducts.length === 0" class="modal-empty">暂无商品</view>
          <scroll-view scroll-y class="service-pick-list">
            <view v-for="prod in filteredProducts" :key="prod.ID" class="service-pick-group">
              <view class="service-pick-item" @click="toggleProductExpand(prod)">
                <text class="service-pick-name">{{ prod.name }}</text>
                <text class="service-pick-arrow" v-if="(prod.skus || prod.SKUs || []).filter(s => s.sellable !== false).length > 1">{{ expandedProductId === prod.ID ? '▾' : '▸' }}</text>
                <text class="service-pick-price" v-else @click.stop="toggleProductExpand(prod)">¥{{ (prod.skus || prod.SKUs || [])[0]?.price ?? '-' }}</text>
              </view>
              <view v-if="expandedProductId === prod.ID" class="price-rules">
                <view
                  class="price-rule-item"
                  v-for="sku in (prod.skus || prod.SKUs || []).filter(s => s.sellable !== false)"
                  :key="sku.ID"
                  @click="addProductSKU(prod, sku)"
                >
                  <text class="rule-level">{{ sku.spec_name || '默认' }}</text>
                  <text class="rule-price">¥{{ sku.price }}</text>
                </view>
              </view>
            </view>
          </scroll-view>
        </view>
      </view>
    </template>
  </view>
  </SideLayout>
</template>

<script setup lang="ts">
import SideLayout from '@/components/SideLayout.vue'
import { computed, ref, reactive } from 'vue'
import { onLoad } from '@dcloudio/uni-app'
import { getAppointment, updateAppointmentNotes } from '@/api/appointment'
import { createBatchOrdersFromAppointment, getOrder, updateOrder } from '@/api/order'
import { getServiceList } from '@/api/service'
import { getCategoryTree } from '@/api/service-category'
import { getProductList, getProductCategories } from '@/api/product'
import { getPersonalityBg, getPersonalityColor } from '@/utils/personality'

const appointmentId = ref(0)
const editingOrderId = ref(0)
const existingOrder = ref<any>(null)
const appt = ref<any>(null)
const loading = ref(true)
const submitting = ref(false)
const allServices = ref<any[]>([])
const categoryTree = ref<any[]>([]) // 树形：顶级分类，子分类在 children 里
const activeCat1 = ref(0)
const activeCat2 = ref(0)

const topCategories = computed(() => categoryTree.value)
const subCategories = computed(() => {
  const top = categoryTree.value.find(c => c.ID === activeCat1.value)
  return top?.children || []
})
const filteredServices = computed(() => {
  const existing = new Set(drafts[addingDraftIndex.value]?.services.map(s => s.service_id) || [])
  // 有价格规则的服务不过滤（可以选不同规格），无规则的按 service_id 去重
  let list = allServices.value.filter(s => (s.price_rules?.length > 0) || !existing.has(s.ID))
  if (activeCat2.value) {
    list = list.filter(s => s.category_id === activeCat2.value)
  } else if (activeCat1.value) {
    const subIds = new Set(subCategories.value.map((c: any) => c.ID))
    list = list.filter(s => subIds.has(s.category_id))
  }
  return list
})

interface DraftService {
  service_id: number
  service_name: string
  price: number
  duration: number
}

interface DraftProduct {
  product_id: number
  sku_id: number
  name: string
  price: number
  quantity: number
}

interface Draft {
  petId: number
  petName: string
  meta: string
  tags: Array<{ text: string; className: string; style?: string }>
  services: DraftService[]
  products: DraftProduct[]
  amount: number
}

const drafts = reactive<Draft[]>([])
const noteDraft = ref('')

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
  return rem > 0 ? `${years}年${rem}个月` : `${years}年`
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

function recalcAmount(draft: Draft) {
  const svcTotal = draft.services.reduce((s, svc) => s + svc.price, 0)
  const prodTotal = draft.products.reduce((s, p) => s + p.price * p.quantity, 0)
  draft.amount = svcTotal + prodTotal
}

const totalAmount = computed(() => drafts.reduce((sum, d) => sum + d.amount, 0))
const petSummary = computed(() => drafts.map((draft) => draft.petName).filter(Boolean).join(' / ') || '-')
const isEditing = computed(() => editingOrderId.value > 0)

// === 修改价格 ===
const editingPrice = ref<{ di: number; si: number } | null>(null)
const editPriceValue = ref('')

function isEditingPrice(di: number, si: number) {
  return editingPrice.value?.di === di && editingPrice.value?.si === si
}

function editPrice(di: number, si: number) {
  const svc = drafts[di].services[si]
  editingPrice.value = { di, si }
  editPriceValue.value = String(svc.price)
}

function saveInlinePrice() {
  if (!editingPrice.value) return
  const { di, si } = editingPrice.value
  const val = parseFloat(editPriceValue.value)
  if (isNaN(val) || val < 0) {
    uni.showToast({ title: '请输入有效价格', icon: 'none' })
    return
  }
  drafts[di].services[si].price = val
  recalcAmount(drafts[di])
  editingPrice.value = null
}

// === 删除服务 ===
function removeService(di: number, si: number) {
  const svc = drafts[di].services[si]
  uni.showModal({
    title: '删除服务',
    content: `确定删除「${svc.service_name}」？`,
    confirmColor: '#EF4444',
    success: (res) => {
      if (res.confirm) {
        drafts[di].services.splice(si, 1)
        recalcAmount(drafts[di])
      }
    }
  })
}

// === 添加服务 ===
const showAddService = ref(false)
const addingDraftIndex = ref(0)
const expandedServiceId = ref(0)

function toggleServiceExpand(svc: any) {
  if (svc.price_rules?.length) {
    expandedServiceId.value = expandedServiceId.value === svc.ID ? 0 : svc.ID
  } else {
    addService(svc)
  }
}

function addServiceWithRule(svc: any, rule: any) {
  const ruleName = rule.name || rule.fur_level || ''
  const name = ruleName ? `${svc.name}(${ruleName})` : svc.name
  drafts[addingDraftIndex.value].services.push({
    service_id: svc.ID,
    service_name: name,
    price: rule.price,
    duration: rule.duration || svc.duration,
  })
  recalcAmount(drafts[addingDraftIndex.value])
  showAddService.value = false
  expandedServiceId.value = 0
}

function openAddService(di: number) {
  addingDraftIndex.value = di
  expandedServiceId.value = 0
  if (topCategories.value.length && !activeCat1.value) {
    activeCat1.value = topCategories.value[0].ID
  }
  activeCat2.value = 0
  showAddService.value = true
}

function addService(svc: any) {
  drafts[addingDraftIndex.value].services.push({
    service_id: svc.ID,
    service_name: svc.name,
    price: svc.base_price,
    duration: svc.duration,
  })
  recalcAmount(drafts[addingDraftIndex.value])
  showAddService.value = false
}

// === 添加商品 ===
const showAddProduct = ref(false)
const addingProductDraftIndex = ref(0)
const productCategories = ref<any[]>([])
const activeProductCat = ref(0)
const expandedProductId = ref(0)
const productKeyword = ref('')
const filteredProducts = ref<any[]>([])
const productLoading = ref(false)
let productSearchTimer: ReturnType<typeof setTimeout> | null = null

async function loadProductOptions() {
  productLoading.value = true
  try {
    const params: any = { page: 1, page_size: 100 }
    if (activeProductCat.value) params.category_id = activeProductCat.value
    if (productKeyword.value.trim()) params.keyword = productKeyword.value.trim()
    const res = await getProductList(params)
    filteredProducts.value = (res.data?.list || []).filter((p: any) => p.status === 1)
  } catch {
    filteredProducts.value = []
  } finally {
    productLoading.value = false
  }
}

function onProductKeywordInput(e: any) {
  productKeyword.value = e.detail?.value || ''
  if (productSearchTimer) clearTimeout(productSearchTimer)
  productSearchTimer = setTimeout(() => {
    loadProductOptions()
  }, 250)
}

function onProductKeywordConfirm() {
  if (productSearchTimer) clearTimeout(productSearchTimer)
  loadProductOptions()
}

function setProductCategory(categoryID: number) {
  activeProductCat.value = categoryID
  loadProductOptions()
}

function openAddProduct(di: number) {
  addingProductDraftIndex.value = di
  expandedProductId.value = 0
  productKeyword.value = ''
  activeProductCat.value = 0
  showAddProduct.value = true
  loadProductOptions()
}

function toggleProductExpand(prod: any) {
  const skus = (prod.skus || prod.SKUs || []).filter((s: any) => s.sellable !== false)
  if (skus.length > 1) {
    expandedProductId.value = expandedProductId.value === prod.ID ? 0 : prod.ID
  } else if (skus.length === 1) {
    addProductSKU(prod, skus[0])
  } else {
    addProductDirect(prod)
  }
}

function addProductDirect(prod: any) {
  const skus = (prod.skus || prod.SKUs || []).filter((s: any) => s.sellable !== false)
  const sku = skus[0]
  const price = sku?.price ?? prod.base_price ?? 0
  drafts[addingProductDraftIndex.value].products.push({
    product_id: prod.ID,
    sku_id: sku?.ID || 0,
    name: sku?.spec_name ? `${prod.name}(${sku.spec_name})` : prod.name,
    price,
    quantity: 1,
  })
  recalcAmount(drafts[addingProductDraftIndex.value])
  uni.showToast({ title: '已添加', icon: 'success', duration: 800 })
}

function addProductSKU(prod: any, sku: any) {
  drafts[addingProductDraftIndex.value].products.push({
    product_id: prod.ID,
    sku_id: sku.ID,
    name: sku.spec_name ? `${prod.name}(${sku.spec_name})` : prod.name,
    price: sku.price,
    quantity: 1,
  })
  recalcAmount(drafts[addingProductDraftIndex.value])
  expandedProductId.value = 0
  uni.showToast({ title: '已添加', icon: 'success', duration: 800 })
}

function changeProductQty(di: number, pi: number, delta: number) {
  const prod = drafts[di].products[pi]
  prod.quantity = Math.max(1, prod.quantity + delta)
  recalcAmount(drafts[di])
}

function removeProduct(di: number, pi: number) {
  const prod = drafts[di].products[pi]
  uni.showModal({
    title: '删除商品',
    content: `确定删除「${prod.name}」？`,
    confirmColor: '#EF4444',
    success: (res) => {
      if (res.confirm) {
        drafts[di].products.splice(pi, 1)
        recalcAmount(drafts[di])
      }
    }
  })
}

// === 加载 ===
async function loadData() {
  if (!appointmentId.value) return
  loading.value = true
  try {
    const [apptRes, orderRes] = await Promise.all([
      getAppointment(appointmentId.value),
      editingOrderId.value ? getOrder(editingOrderId.value) : Promise.resolve(null as any),
    ])
    appt.value = apptRes.data
    existingOrder.value = orderRes?.data || null
    noteDraft.value = appt.value?.notes || ''

    // 加载服务列表、分类、商品（用于添加服务/商品）
    try {
      const [svcRes, catRes, prodCatRes] = await Promise.all([
        getServiceList({ page: 1, page_size: 200 } as any),
        getCategoryTree(),
        getProductCategories(),
      ])
      allServices.value = (svcRes.data?.list || []).filter((s: any) => s.status === 1)
      categoryTree.value = catRes.data || []
      productCategories.value = (prodCatRes.data || []).filter((c: any) => c.status === 1)
    } catch { /* ignore */ }

    const petMap = new Map<string, any>()
    const pets = Array.isArray(appt.value?.pets) ? appt.value.pets : []
    for (const petItem of pets) {
      const key = petItem?.pet?.name || `宠物#${petItem.pet_id}`
      if (!petMap.has(key)) {
        petMap.set(key, petItem)
      }
    }

    drafts.length = 0
    if (existingOrder.value?.pet_groups?.length) {
      for (const group of existingOrder.value.pet_groups) {
        const petItem = petMap.get(group.pet_name)
        const svcs: DraftService[] = []
        const prods: DraftProduct[] = []
        for (const item of (group.items || [])) {
          if (item.item_type === 2) {
            prods.push({
              product_id: item.item_id,
              sku_id: 0,
              name: item.name,
              price: Number(item.unit_price || 0),
              quantity: item.quantity || 1,
            })
          } else {
            const service = allServices.value.find((svc: any) => svc.ID === item.item_id)
            svcs.push({
              service_id: item.item_id,
              service_name: item.name,
              price: Number(item.unit_price || item.amount || 0),
              duration: Number(service?.duration || 0),
            })
          }
        }
        const svcTotal = svcs.reduce((s: number, svc: DraftService) => s + svc.price, 0)
        const prodTotal = prods.reduce((s: number, p: DraftProduct) => s + p.price * p.quantity, 0)
        drafts.push({
          petId: petItem?.pet_id || 0,
          petName: group.pet_name,
          meta: getPetMeta(petItem?.pet),
          tags: getPetTags(petItem?.pet),
          services: svcs,
          products: prods,
          amount: svcTotal + prodTotal,
        })
      }
    } else {
      for (const petItem of pets) {
        const svcs = (petItem.services || []).map((s: any) => ({
          service_id: s.service_id,
          service_name: s.service_name,
          price: Number(s.price || 0),
          duration: Number(s.duration || 0),
        }))
        drafts.push({
          petId: petItem.pet_id,
          petName: petItem.pet?.name || `宠物#${petItem.pet_id}`,
          meta: getPetMeta(petItem.pet),
          tags: getPetTags(petItem.pet),
          services: svcs,
          products: [],
          amount: svcs.reduce((s: number, svc: DraftService) => s + svc.price, 0),
        })
      }
    }
  } finally {
    loading.value = false
  }
}

async function submitBatch() {
  if (!appointmentId.value) return
  submitting.value = true
  try {
    await syncAppointmentNotes()

    let res: any
    if (isEditing.value) {
      const items = drafts.flatMap((draft) => [
        ...draft.services.map((svc) => ({
          item_type: 1,
          item_id: svc.service_id,
          name: `${draft.petName} · ${svc.service_name}`,
          quantity: 1,
          unit_price: svc.price,
        })),
        ...draft.products.map((p) => ({
          item_type: 2,
          item_id: p.product_id,
          name: `${draft.petName} · ${p.name}`,
          quantity: p.quantity,
          unit_price: p.price,
        })),
      ])
      res = await updateOrder(editingOrderId.value, {
        customer_id: existingOrder.value?.customer_id || appt.value?.customer_id,
        staff_id: existingOrder.value?.staff_id || appt.value?.staff_id,
        remark: existingOrder.value?.remark || appt.value?.notes || '',
        items,
      } as any)
    } else {
      res = await createBatchOrdersFromAppointment(appointmentId.value, {
        overrides: drafts.map(d => ({
          pet_id: d.petId,
          services: d.services.map(s => ({
            service_id: s.service_id,
            service_name: s.service_name,
            price: s.price,
            duration: s.duration,
          })),
          products: d.products.map(p => ({
            product_id: p.product_id,
            sku_id: p.sku_id,
            name: p.name,
            price: p.price,
            quantity: p.quantity,
          })),
        })),
      })
    }
    const order = res.data
    uni.showToast({ title: isEditing.value ? '已保存修改' : '已生成1张订单', icon: 'success' })
    setTimeout(() => {
      if (order?.ID) {
        uni.redirectTo({ url: `/pages/order/detail?id=${order.ID}` })
        return
      }
      uni.redirectTo({ url: '/pages/order/list' })
    }, 500)
  } catch (e: any) {
    uni.showToast({ title: e?.msg || e?.message || '批量开单失败', icon: 'none' })
  } finally {
    submitting.value = false
  }
}

async function syncAppointmentNotes() {
  const currentNotes = appt.value?.notes || ''
  const nextNotes = noteDraft.value.trim()
  if (!appt.value || nextNotes === currentNotes) return
  const res = await updateAppointmentNotes(appointmentId.value, nextNotes)
  appt.value = res.data || { ...appt.value, notes: nextNotes }
}

onLoad((query) => {
  appointmentId.value = parseInt(String(query?.appointment_id || 0)) || 0
  editingOrderId.value = parseInt(String(query?.order_id || 0)) || 0
  loadData()
})
</script>

<style scoped>
.page { padding: 24rpx; padding-bottom: calc(220rpx + env(safe-area-inset-bottom)); }
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
.service-row { display: flex; justify-content: space-between; align-items: center; gap: 16rpx; padding: 12rpx 0; border-bottom: 1rpx solid #F3F4F6; }
.service-row:last-child { border-bottom: none; }
.service-name { font-size: 24rpx; color: #1F2937; flex: 1; min-width: 0; }
.service-edit { display: flex; align-items: center; gap: 12rpx; }
.service-price { font-size: 24rpx; color: #4F46E5; font-weight: 600; padding: 4rpx 12rpx; background: #EEF2FF; border-radius: 8rpx; }
.service-price-input {
  width: 108rpx;
  min-height: 52rpx;
  padding: 0 12rpx;
  background: #EEF2FF;
  border: 1rpx solid rgba(79, 70, 229, 0.18);
  border-radius: 10rpx;
  box-sizing: border-box;
  font-size: 24rpx;
  color: #4338CA;
  font-weight: 700;
  text-align: center;
}
.service-dur { font-size: 20rpx; color: #9CA3AF; }
.service-del { font-size: 24rpx; color: #EF4444; padding: 4rpx 8rpx; }

.add-service-row { padding: 16rpx 0 4rpx; text-align: center; }
.add-service-text { font-size: 24rpx; color: #4F46E5; }

.product-list { margin-top: 12rpx; border-top: 1rpx solid #EEF2F7; padding-top: 8rpx; }
.product-row { display: flex; justify-content: space-between; align-items: center; gap: 12rpx; padding: 10rpx 0; border-bottom: 1rpx solid #F3F4F6; }
.product-row:last-child { border-bottom: none; }
.product-name { font-size: 24rpx; color: #1F2937; flex: 1; min-width: 0; }
.product-edit { display: flex; align-items: center; gap: 10rpx; }
.product-price { font-size: 24rpx; color: #059669; font-weight: 600; padding: 4rpx 12rpx; background: #ECFDF5; border-radius: 8rpx; }
.product-del { font-size: 24rpx; color: #EF4444; padding: 4rpx 8rpx; }
.qty-ctrl { display: flex; align-items: center; gap: 0; background: #F3F4F6; border-radius: 10rpx; overflow: hidden; }
.qty-btn { width: 48rpx; height: 44rpx; display: flex; align-items: center; justify-content: center; font-size: 28rpx; color: #374151; }
.qty-btn:active { background: #E5E7EB; }
.qty-val { width: 48rpx; text-align: center; font-size: 24rpx; font-weight: 600; color: #111827; }

.notes-head { display: flex; align-items: center; justify-content: space-between; gap: 12rpx; margin-bottom: 12rpx; }
.notes-title { font-size: 26rpx; font-weight: 600; color: #1F2937; display: block; }
.notes-tip { font-size: 22rpx; color: #9CA3AF; flex-shrink: 0; }
.notes-input {
  width: 100%;
  min-height: 132rpx;
  box-sizing: border-box;
  padding: 20rpx;
  border-radius: 14rpx;
  background: #F8FAFC;
  border: 1rpx solid #E5E7EB;
  font-size: 24rpx;
  line-height: 1.6;
  color: #1F2937;
}
.submit-bar {
  position: fixed;
  left: 0;
  right: 0;
  bottom: calc(50px + env(safe-area-inset-bottom));
  padding: 20rpx 24rpx 28rpx;
  background: linear-gradient(180deg, rgba(245, 246, 250, 0), rgba(245, 246, 250, 0.92) 26%, #F5F6FA 100%);
  z-index: 100;
}

.submit-btn {
  margin: 0;
  min-height: 92rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #4F46E5, #6366F1);
  color: #fff;
  border-radius: 18rpx;
  font-size: 30rpx;
  line-height: 1.2;
  text-align: center;
  white-space: nowrap;
  font-weight: 800;
  padding: 0 32rpx;
  box-shadow: 0 14rpx 28rpx rgba(79, 70, 229, 0.24);
}

/* 弹窗 */
.modal-mask { position: fixed; top: 0; left: 0; right: 0; bottom: 0; background: rgba(0,0,0,0.4); z-index: 999; display: flex; align-items: center; justify-content: center; }
.modal-body { background: #fff; border-radius: 20rpx; padding: 32rpx; width: 80%; max-width: 600rpx; }
.modal-body-tall { max-height: 70vh; }
.modal-title { font-size: 30rpx; font-weight: 700; color: #111827; display: block; margin-bottom: 20rpx; }
.modal-svc-name { font-size: 26rpx; color: #6B7280; display: block; margin-bottom: 16rpx; }
.modal-input { background: #F3F4F6; border-radius: 12rpx; padding: 16rpx; font-size: 28rpx; width: 100%; }
.modal-btns { display: flex; gap: 16rpx; margin-top: 24rpx; }
.modal-btn {
  flex: 1;
  min-height: 84rpx;
  text-align: center;
  border-radius: 16rpx;
  font-size: 28rpx;
  font-weight: 700;
  display: flex;
  align-items: center;
  justify-content: center;
  box-sizing: border-box;
}
.modal-btn.cancel { background: #F8FAFC; color: #64748B; border: 2rpx solid #CBD5E1; }
.modal-btn.confirm { background: linear-gradient(135deg, #4F46E5, #6366F1); color: #fff; box-shadow: 0 12rpx 24rpx rgba(79, 70, 229, 0.2); }
.modal-search { width: 100%; min-height: 64rpx; padding: 12rpx 18rpx; background: #F3F4F6; border-radius: 12rpx; font-size: 26rpx; color: #1F2937; box-sizing: border-box; margin-bottom: 12rpx; }
.modal-empty { text-align: center; padding: 40rpx 0; color: #9CA3AF; font-size: 26rpx; }

.cat-tabs { margin-bottom: 12rpx; white-space: nowrap; }
.cat-tabs-sub { margin-bottom: 16rpx; }
.cat-tabs-inner { display: inline-flex; gap: 12rpx; }
.cat-tab { font-size: 26rpx; padding: 10rpx 20rpx; border-radius: 999rpx; background: #F3F4F6; color: #6B7280; white-space: nowrap; }
.cat-tab-active { background: #4F46E5; color: #fff; }
.cat-tab-sub { font-size: 24rpx; padding: 8rpx 16rpx; }
.service-pick-list { max-height: 40vh; }
.service-pick-group { border-bottom: 1rpx solid #F3F4F6; }
.service-pick-group:last-child { border-bottom: none; }
.service-pick-item { display: flex; justify-content: space-between; align-items: center; padding: 20rpx 0; }
.service-pick-arrow { font-size: 24rpx; color: #9CA3AF; margin-left: 8rpx; }
.price-rules { padding: 0 0 12rpx 24rpx; }
.price-rule-item { display: flex; justify-content: space-between; align-items: center; padding: 14rpx 16rpx; margin-bottom: 8rpx; background: #F9FAFB; border-radius: 12rpx; }
.price-rule-item:active { background: #EEF2FF; }
.rule-level { font-size: 26rpx; color: #1F2937; font-weight: 500; }
.rule-price { font-size: 24rpx; color: #6B7280; }
.service-pick-name { font-size: 28rpx; color: #1F2937; }
.service-pick-price { font-size: 24rpx; color: #6B7280; }
</style>
