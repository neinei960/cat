<template>
  <SideLayout>
  <view class="page">
    <view class="section-title">基础信息</view>
    <view class="form">
      <view class="form-item">
        <text class="label">服务名称 *</text>
        <input v-model="form.name" placeholder="如：标准洗澡" class="input" />
      </view>
      <view class="form-item">
        <text class="label">服务分类</text>
        <view class="cat-selector">
          <view class="cat-select-card" @click="chooseParentCategory">
            <view class="cat-select-text">
              <text class="cat-select-label">一级分类</text>
              <text class="cat-select-value">{{ currentParentLabel }}</text>
            </view>
            <text class="cat-select-arrow">›</text>
          </view>
          <view v-if="selectedParentIdx > 0 && childOptions.length" class="cat-select-card" @click="chooseChildCategory">
            <view class="cat-select-text">
              <text class="cat-select-label">二级分类</text>
              <text class="cat-select-value">{{ currentChildLabel }}</text>
            </view>
            <text class="cat-select-arrow">›</text>
          </view>
          <view v-if="selectedCategoryLabel" class="cat-selected-row">
            <text class="cat-selected-label">当前分类</text>
            <text class="cat-selected-value">{{ selectedCategoryLabel }}</text>
            <text class="cat-selected-clear" @click.stop="clearCategory">清空</text>
          </view>
        </view>
      </view>

      <!-- 计费方式 -->
      <view class="form-item">
        <text class="label">计费方式 *</text>
        <view class="pricing-type-toggle">
          <view :class="['toggle-btn', form.pricing_type === 1 ? 'active' : '']" @click="form.pricing_type = 1">按次</view>
          <view :class="['toggle-btn', form.pricing_type === 2 ? 'active' : '']" @click="form.pricing_type = 2">按天</view>
        </view>
      </view>

      <!-- 按次模式 -->
      <template v-if="form.pricing_type === 1">
        <view class="form-item">
          <text class="label">基础价格 (元) *</text>
          <input v-model="form.base_price" type="digit" placeholder="0.00" class="input input-amount" />
        </view>
        <view class="form-item">
          <text class="label">时长 (分钟) *</text>
          <input v-model="form.duration" type="number" placeholder="60" class="input" />
        </view>
      </template>

      <!-- 按天模式 -->
      <template v-if="form.pricing_type === 2">
        <view class="form-item">
          <text class="label">平时单价 (元/天) *</text>
          <input v-model="form.base_price" type="digit" placeholder="80" class="input input-amount" />
        </view>
        <view class="form-item">
          <text class="label">节假日单价 (元/天)</text>
          <input v-model="form.holiday_price" type="digit" placeholder="95" class="input input-amount" />
        </view>
      </template>

      <view class="form-item">
        <text class="label">描述</text>
        <textarea v-model="form.description" placeholder="服务描述" class="textarea" />
      </view>
      <view class="form-item">
        <text class="label">排序</text>
        <input v-model="form.sort_order" type="number" placeholder="0" class="input" />
      </view>
    </view>

    <!-- 服务规格/项目 (按次模式) -->
    <template v-if="form.pricing_type === 1">
      <view class="section-title">
        <text>服务规格</text>
        <view class="btn-add-spec" @click="addSpec">+ 添加规格</view>
      </view>
      <view class="specs-hint" v-if="specs.length === 0">
        <text>可添加多个规格，如：短毛/长毛/A/B/C/D，每个规格有独立的价格和时长</text>
      </view>

      <view class="spec-list" v-if="specs.length > 0">
        <view class="spec-header">
          <text class="spec-col-name">规格名称 *</text>
          <text class="spec-col-price">价格 (元) *</text>
          <text class="spec-col-dur">时长 (分钟)</text>
          <text class="spec-col-act">操作</text>
        </view>
        <view class="spec-row" v-for="(spec, idx) in specs" :key="idx">
          <view class="spec-col-name">
            <input v-model="spec.name" placeholder="如：短毛猫" class="spec-input" />
          </view>
          <view class="spec-col-price">
            <input v-model="spec.price" type="digit" placeholder="0" class="spec-input spec-input-center input-amount" />
          </view>
          <view class="spec-col-dur">
            <input v-model="spec.duration" type="number" placeholder="60" class="spec-input spec-input-center" />
          </view>
          <view class="spec-col-act">
            <text class="spec-del" @click="removeSpec(idx)">删除</text>
          </view>
        </view>
      </view>
    </template>

    <!-- 优惠策略 (按天模式) -->
    <template v-if="form.pricing_type === 2">
      <view class="section-title">
        <text>优惠策略</text>
        <view class="btn-add-spec" @click="addDiscount">+ 添加策略</view>
      </view>
      <view class="specs-hint" v-if="discountItems.length === 0">
        <text>可添加多种优惠，如：满3天享优惠价、住7免1等</text>
      </view>

      <view class="discount-list" v-if="discountItems.length > 0">
        <view class="discount-card" v-for="(d, idx) in discountItems" :key="idx">
          <view class="discount-row">
            <view class="discount-field">
              <text class="discount-label">类型</text>
              <view class="pricing-type-toggle small">
                <view :class="['toggle-btn', d.type === 1 ? 'active' : '']" @click="d.type = 1">满天折扣</view>
                <view :class="['toggle-btn', d.type === 2 ? 'active' : '']" @click="d.type = 2">住N免M</view>
              </view>
            </view>
            <view class="discount-field">
              <text class="discount-label">适用</text>
              <view class="pricing-type-toggle small">
                <view :class="['toggle-btn', !d.is_holiday ? 'active' : '']" @click="d.is_holiday = false">平时</view>
                <view :class="['toggle-btn', d.is_holiday ? 'active' : '']" @click="d.is_holiday = true">节假日</view>
              </view>
            </view>
          </view>
          <view class="discount-row">
            <view class="discount-field">
              <text class="discount-label">满 N 天</text>
              <input v-model="d.min_days" type="number" placeholder="3" class="input" />
            </view>
            <view class="discount-field" v-if="d.type === 1">
              <text class="discount-label">优惠单价 (元/天)</text>
              <input v-model="d.discount_price" type="digit" placeholder="68" class="input input-amount" />
            </view>
            <view class="discount-field" v-if="d.type === 2">
              <text class="discount-label">免 M 天</text>
              <input v-model="d.free_days" type="number" placeholder="1" class="input" />
            </view>
            <view class="discount-act">
              <text class="spec-del" @click="removeDiscount(idx)">删除</text>
            </view>
          </view>
        </view>
      </view>
    </template>

    <button class="btn-submit" @click="onSubmit" :loading="submitting">{{ id ? '保存' : '新增' }}</button>
    <button class="btn-delete" v-if="id" @click="onDelete">删除服务</button>

    <view v-if="showCategoryPicker" class="cat-modal-mask" @click="closeCategoryPicker">
      <view class="cat-modal" @click.stop>
        <view class="cat-modal-header">
          <text class="cat-modal-title">{{ categoryPickerTitle }}</text>
          <text class="cat-modal-close" @click="closeCategoryPicker">取消</text>
        </view>
        <scroll-view scroll-y class="cat-modal-list">
          <view
            v-for="option in categoryPickerOptions"
            :key="`${categoryPickerMode}-${option.index}`"
            :class="['cat-modal-item', option.active ? 'active' : '']"
            @click="onCategoryOptionSelect(option.index)"
          >
            <text class="cat-modal-item-label">{{ option.label }}</text>
            <text v-if="option.active" class="cat-modal-item-check">✓</text>
          </view>
        </scroll-view>
      </view>
    </view>
  </view>
  </SideLayout>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { onLoad } from '@dcloudio/uni-app'
import SideLayout from '@/components/SideLayout.vue'
import { getService, createService, updateService, deleteService, getPriceRules, createPriceRule, deletePriceRule, getDiscounts, createDiscount, deleteDiscount } from '@/api/service'
import { getCategoryTree } from '@/api/service-category'
import { safeBack } from '@/utils/navigate'

const id = ref(0)
const submitting = ref(false)
const categories = ref<ServiceCategory[]>([])

const form = ref({
  name: '', category: '', category_id: undefined as number | undefined,
  base_price: '', duration: '', holiday_price: '',
  pricing_type: 1,
  description: '', applicable_species: '', applicable_sizes: '', sort_order: 0
})

// Specs (price rules) - for pricing_type=1
interface SpecItem {
  id?: number
  name: string
  price: string
  duration: string
}
const specs = ref<SpecItem[]>([])
const existingRuleIds = ref<number[]>([])

function addSpec() {
  specs.value.push({ name: '', price: '', duration: '' })
}

function removeSpec(idx: number) {
  specs.value.splice(idx, 1)
}

// Discounts - for pricing_type=2
interface DiscountItem {
  id?: number
  type: number
  min_days: string
  discount_price: string
  free_days: string
  is_holiday: boolean
}
const discountItems = ref<DiscountItem[]>([])
const existingDiscountIds = ref<number[]>([])

function addDiscount() {
  discountItems.value.push({ type: 1, min_days: '', discount_price: '', free_days: '', is_holiday: false })
}

function removeDiscount(idx: number) {
  discountItems.value.splice(idx, 1)
}

// Category picker state
const selectedParentIdx = ref(0)
const selectedChildIdx = ref(0)

const parentOptions = computed(() => {
  return [{ ID: 0, name: '未分类' } as any, ...categories.value]
})
const parentNames = computed(() => parentOptions.value.map(c => c.name))

const childOptions = computed(() => {
  const parent = parentOptions.value[selectedParentIdx.value]
  if (!parent || parent.ID === 0) return []
  return parent.children || []
})
const childNames = computed(() => {
  if (!childOptions.value.length) return []
  return ['（不选二级）', ...childOptions.value.map((c: ServiceCategory) => c.name)]
})

const currentParentLabel = computed(() => parentNames.value[selectedParentIdx.value] || '选择一级分类')
const currentChildLabel = computed(() => {
  if (!childOptions.value.length) return '当前一级分类没有二级'
  return childNames.value[selectedChildIdx.value] || '选择二级分类'
})
const selectedCategoryLabel = computed(() => form.value.category || '')
const showCategoryPicker = ref(false)
const categoryPickerMode = ref<'parent' | 'child'>('parent')
const categoryPickerTitle = computed(() => categoryPickerMode.value === 'parent' ? '选择一级分类' : '选择二级分类')
const categoryPickerOptions = computed(() => {
  const source = categoryPickerMode.value === 'parent' ? parentNames.value : childNames.value
  const selected = categoryPickerMode.value === 'parent' ? selectedParentIdx.value : selectedChildIdx.value
  return source.map((label, index) => ({
    label,
    index,
    active: index === selected,
  }))
})

function selectParentByIndex(index: number) {
  selectedParentIdx.value = index
  selectedChildIdx.value = 0
  updateCategoryId()
}

function selectChildByIndex(index: number) {
  selectedChildIdx.value = index
  updateCategoryId()
}

function updateCategoryId() {
  const parent = parentOptions.value[selectedParentIdx.value]
  if (!parent || parent.ID === 0) {
    form.value.category_id = undefined
    form.value.category = ''
    return
  }
  if (selectedChildIdx.value > 0 && childOptions.value.length > 0) {
    const child = childOptions.value[selectedChildIdx.value - 1]
    form.value.category_id = child.ID
    form.value.category = `${parent.name}/${child.name}`
  } else {
    form.value.category_id = parent.ID
    form.value.category = parent.name
  }
}

function clearCategory() {
  selectedParentIdx.value = 0
  selectedChildIdx.value = 0
  updateCategoryId()
}

function chooseParentCategory() {
  categoryPickerMode.value = 'parent'
  showCategoryPicker.value = true
}

function chooseChildCategory() {
  if (!childOptions.value.length) return
  categoryPickerMode.value = 'child'
  showCategoryPicker.value = true
}

function closeCategoryPicker() {
  showCategoryPicker.value = false
}

function onCategoryOptionSelect(index: number) {
  if (categoryPickerMode.value === 'parent') {
    selectParentByIndex(index)
  } else {
    selectChildByIndex(index)
  }
  closeCategoryPicker()
}

function restorePickerFromId(catId: number | undefined) {
  if (!catId) { selectedParentIdx.value = 0; selectedChildIdx.value = 0; return }
  for (let i = 0; i < categories.value.length; i++) {
    const cat = categories.value[i]
    if (cat.ID === catId) {
      selectedParentIdx.value = i + 1
      selectedChildIdx.value = 0
      return
    }
    if (cat.children) {
      for (let j = 0; j < cat.children.length; j++) {
        if (cat.children[j].ID === catId) {
          selectedParentIdx.value = i + 1
          selectedChildIdx.value = j + 1
          return
        }
      }
    }
  }
}

onLoad(async (query) => {
  const res = await getCategoryTree()
  categories.value = res.data || []

  if (query?.id) {
    id.value = parseInt(query.id)
    await loadData()
  }
})

async function loadData() {
  const res = await getService(id.value)
  const d = res.data
  form.value = {
    name: d.name, category: d.category || '', category_id: d.category_id,
    base_price: String(d.base_price), duration: String(d.duration),
    holiday_price: String(d.holiday_price || ''),
    pricing_type: d.pricing_type || 1,
    description: d.description, applicable_species: d.applicable_species || '',
    applicable_sizes: d.applicable_sizes || '', sort_order: d.sort_order,
  }
  restorePickerFromId(d.category_id)

  // Load existing price rules as specs
  try {
    const rulesRes = await getPriceRules(id.value)
    const rules = rulesRes.data || []
    existingRuleIds.value = rules.map((r: any) => r.ID)
    specs.value = rules.map((r: any) => ({
      id: r.ID,
      name: r.name || r.fur_level || '',
      price: String(r.price),
      duration: r.duration != null ? String(r.duration) : '',
    }))
  } catch {}

  // Load existing discounts
  try {
    const discRes = await getDiscounts(id.value)
    const discs = discRes.data || []
    existingDiscountIds.value = discs.map((d: any) => d.ID)
    discountItems.value = discs.map((d: any) => ({
      id: d.ID,
      type: d.type,
      min_days: String(d.min_days),
      discount_price: String(d.discount_price || ''),
      free_days: String(d.free_days || ''),
      is_holiday: d.is_holiday,
    }))
  } catch {}
}

async function onSubmit() {
  if (!form.value.name || !form.value.base_price) {
    uni.showToast({ title: '请填写必填项', icon: 'none' }); return
  }
  if (form.value.pricing_type === 1 && !form.value.duration) {
    uni.showToast({ title: '请填写时长', icon: 'none' }); return
  }

  // Validate specs (pricing_type=1)
  if (form.value.pricing_type === 1) {
    for (const spec of specs.value) {
      if (!spec.name.trim() || !spec.price) {
        uni.showToast({ title: '规格名称和价格为必填', icon: 'none' }); return
      }
    }
  }

  submitting.value = true
  try {
    updateCategoryId()
    const data = {
      ...form.value,
      base_price: parseFloat(form.value.base_price),
      duration: parseInt(form.value.duration) || 0,
      holiday_price: parseFloat(form.value.holiday_price) || 0,
      pricing_type: form.value.pricing_type,
    }

    let serviceId = id.value
    if (id.value) {
      await updateService(id.value, data)
    } else {
      const res = await createService(data)
      serviceId = res.data.ID
    }

    // Sync specs (pricing_type=1)
    for (const oldId of existingRuleIds.value) {
      await deletePriceRule(serviceId, oldId)
    }
    if (form.value.pricing_type === 1) {
      for (const spec of specs.value) {
        if (spec.name.trim()) {
          await createPriceRule(serviceId, {
            name: spec.name.trim(),
            price: parseFloat(spec.price),
            duration: parseInt(spec.duration) || 0,
          })
        }
      }
    }

    // Sync discounts (pricing_type=2)
    for (const oldId of existingDiscountIds.value) {
      await deleteDiscount(serviceId, oldId)
    }
    if (form.value.pricing_type === 2) {
      for (const d of discountItems.value) {
        if (d.min_days) {
          await createDiscount(serviceId, {
            type: d.type,
            min_days: parseInt(d.min_days),
            discount_price: parseFloat(d.discount_price) || 0,
            free_days: parseInt(d.free_days) || 0,
            is_holiday: d.is_holiday,
          })
        }
      }
    }

    uni.showToast({ title: '保存成功', icon: 'success' })
    setTimeout(() => safeBack(), 500)
  } finally { submitting.value = false }
}

async function onDelete() {
  uni.showModal({
    title: '确认删除', content: '确认删除该服务？',
    success: async (res) => {
      if (res.confirm) {
        await deleteService(id.value)
        uni.showToast({ title: '已删除', icon: 'success' })
        setTimeout(() => safeBack(), 500)
      }
    }
  })
}
</script>

<style scoped>
.page { padding: 24rpx; max-width: 800px; }
.section-title {
  display: flex; justify-content: space-between; align-items: center;
  font-size: 32rpx; font-weight: 700; color: #1F2937;
  margin-bottom: 16rpx; margin-top: 8rpx;
}
.form { background: #fff; border-radius: 16rpx; padding: 8rpx 24rpx; margin-bottom: 32rpx; }
.form-item { padding: 24rpx 0; border-bottom: 1rpx solid #F3F4F6; }
.form-item:last-child { border-bottom: none; }
.label { font-size: 28rpx; color: #374151; display: block; margin-bottom: 12rpx; }
.input { font-size: 28rpx; color: #1F2937; height: 60rpx; text-align: left; }
.textarea { font-size: 28rpx; color: #1F2937; width: 100%; height: 160rpx; text-align: left; }
.input :deep(.uni-input-wrapper) {
  width: 100%;
}
.input :deep(.uni-input-input),
.input :deep(.uni-input-placeholder) {
  text-align: left !important;
}
.textarea :deep(.uni-textarea-textarea),
.textarea :deep(.uni-textarea-placeholder) {
  text-align: left !important;
}
.cat-selector { display: flex; flex-direction: column; gap: 16rpx; }
.cat-select-card {
  display: flex; align-items: center; justify-content: space-between;
  background: #F9FAFB; border: 2rpx solid #E5E7EB; border-radius: 16rpx; padding: 20rpx 24rpx;
}
.cat-select-text { display: flex; flex-direction: column; gap: 6rpx; }
.cat-select-label { font-size: 24rpx; color: #6B7280; }
.cat-select-value { font-size: 28rpx; color: #111827; font-weight: 600; }
.cat-select-arrow { font-size: 36rpx; color: #9CA3AF; line-height: 1; }
.cat-selected-row {
  display: flex; align-items: center; gap: 12rpx; flex-wrap: wrap;
  background: #EEF2FF; color: #4338CA; border-radius: 14rpx; padding: 16rpx 20rpx;
}
.cat-selected-label { font-size: 24rpx; color: #6366F1; }
.cat-selected-value { flex: 1; font-size: 26rpx; font-weight: 600; }
.cat-selected-clear { font-size: 24rpx; color: #4F46E5; }
.cat-modal-mask {
  position: fixed; inset: 0; z-index: 7000;
  background: rgba(17, 24, 39, 0.45);
  display: flex; align-items: flex-end; justify-content: center;
}
.cat-modal {
  width: 100%; max-width: 800px; background: #FFFFFF;
  border-radius: 28rpx 28rpx 0 0; overflow: hidden;
}
.cat-modal-header {
  display: flex; align-items: center; justify-content: space-between;
  padding: 24rpx 28rpx; border-bottom: 1rpx solid #E5E7EB;
}
.cat-modal-title { font-size: 30rpx; font-weight: 700; color: #111827; }
.cat-modal-close { font-size: 26rpx; color: #6B7280; }
.cat-modal-list { max-height: 60vh; }
.cat-modal-item {
  display: flex; align-items: center; justify-content: space-between;
  padding: 26rpx 28rpx; border-bottom: 1rpx solid #F3F4F6;
  font-size: 28rpx; color: #111827; background: #FFFFFF;
}
.cat-modal-item.active {
  background: #EEF2FF; color: #4338CA; font-weight: 600;
}
.cat-modal-item-label { flex: 1; }
.cat-modal-item-check { font-size: 28rpx; color: #4F46E5; }

/* Pricing type toggle */
.pricing-type-toggle {
  display: flex; gap: 0; border-radius: 12rpx; overflow: hidden; border: 2rpx solid #E5E7EB;
}
.pricing-type-toggle .toggle-btn {
  flex: 1; text-align: center; padding: 16rpx 24rpx; font-size: 28rpx;
  color: #6B7280; background: #F9FAFB; cursor: pointer; transition: all 0.2s;
}
.pricing-type-toggle .toggle-btn.active {
  color: #fff; background: #4F46E5; font-weight: 600;
}
.pricing-type-toggle.small .toggle-btn {
  padding: 10rpx 16rpx; font-size: 24rpx;
}

/* Specs */
.btn-add-spec {
  font-size: 26rpx; color: #4F46E5; background: #EEF2FF;
  padding: 10rpx 24rpx; border-radius: 12rpx; border: 2rpx solid #C7D2FE;
}
.specs-hint {
  background: #F9FAFB; border-radius: 12rpx; padding: 24rpx;
  margin-bottom: 24rpx; font-size: 24rpx; color: #9CA3AF;
}
.spec-list {
  background: #fff; border-radius: 16rpx; padding: 0 24rpx;
  margin-bottom: 32rpx; box-shadow: 0 2rpx 8rpx rgba(0,0,0,0.04);
}
.spec-header {
  display: flex; gap: 12rpx; padding: 20rpx 0;
  border-bottom: 2rpx solid #E5E7EB; font-size: 24rpx; color: #6B7280; font-weight: 600;
}
.spec-row {
  display: flex; gap: 12rpx; padding: 16rpx 0;
  border-bottom: 1rpx solid #F3F4F6; align-items: center;
}
.spec-row:last-child { border-bottom: none; }
.spec-col-name { flex: 3; }
.spec-col-price { flex: 2; text-align: center; }
.spec-col-dur { flex: 2; text-align: center; }
.spec-col-act { flex: 1; text-align: center; }
.spec-input {
  font-size: 28rpx; color: #1F2937; height: 56rpx;
  background: #F9FAFB; border-radius: 8rpx; padding: 0 12rpx; text-align: left;
}
.spec-input :deep(.uni-input-wrapper) {
  width: 100%;
}
.spec-input :deep(.uni-input-input),
.spec-input :deep(.uni-input-placeholder) {
  text-align: left !important;
}
.spec-input-center,
.spec-input-center :deep(.uni-input-input),
.spec-input-center :deep(.uni-input-placeholder) {
  text-align: center !important;
}
.spec-del { font-size: 24rpx; color: #EF4444; }

/* Discount cards */
.discount-list { margin-bottom: 32rpx; }
.discount-card {
  background: #fff; border-radius: 16rpx; padding: 24rpx;
  margin-bottom: 16rpx; box-shadow: 0 2rpx 8rpx rgba(0,0,0,0.04);
}
.discount-row {
  display: flex; gap: 16rpx; align-items: flex-end; margin-bottom: 16rpx;
}
.discount-row:last-child { margin-bottom: 0; }
.discount-field { flex: 1; }
.discount-label { font-size: 24rpx; color: #6B7280; display: block; margin-bottom: 8rpx; }
.discount-act { width: 80rpx; text-align: center; padding-bottom: 8rpx; }

.btn-submit { background: #4F46E5; color: #fff; border-radius: 12rpx; font-size: 30rpx; margin-top: 16rpx; }
.btn-delete { background: #fff; color: #DC2626; border: 1rpx solid #DC2626; border-radius: 12rpx; font-size: 30rpx; margin-top: 16rpx; }
</style>
