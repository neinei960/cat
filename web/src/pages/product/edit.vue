<template>
  <SideLayout>
  <view class="page">
    <!-- 基本信息 -->
    <view class="section-title">基本信息</view>
    <view class="section">
      <view class="form-item">
        <text class="label">商品名称 *</text>
        <input v-model="form.name" placeholder="请输入商品名称" class="input" />
      </view>
      <view class="form-item">
        <text class="label">商品分类</text>
        <picker :range="categoryNames" :value="selectedCategoryIdx" @change="onCategoryChange">
          <view class="picker">{{ categoryNames[selectedCategoryIdx] || '请选择分类' }}</view>
        </picker>
      </view>
      <view class="form-item">
        <text class="label">品牌</text>
        <input
          v-model="form.brand"
          placeholder="请输入品牌"
          class="input"
          @focus="showBrandSuggestions = true"
          @blur="hideBrandSuggestions"
        />
        <view v-if="showBrandSuggestions && filteredBrands.length" class="brand-suggestions">
          <view
            class="brand-suggest-item"
            v-for="b in filteredBrands"
            :key="b"
            @click="selectBrand(b)"
          >{{ b }}</view>
        </view>
      </view>
      <view class="form-item">
        <text class="label">商品描述</text>
        <textarea v-model="form.description" placeholder="商品描述（选填）" class="textarea" />
      </view>
    </view>

    <!-- 商品规格 -->
    <view class="section-title-row">
      <text class="section-title-text">商品规格</text>
      <view class="spec-toggle">
        <view
          :class="['toggle-btn', !form.multi_spec ? 'toggle-active' : '']"
          @click="setSpecMode(false)"
        >单规格</view>
        <view
          :class="['toggle-btn', form.multi_spec ? 'toggle-active' : '']"
          @click="setSpecMode(true)"
        >多规格</view>
      </view>
    </view>
    <view class="section">
      <!-- 单规格 -->
      <view v-if="!form.multi_spec" class="single-spec">
        <view class="two-col">
          <view class="col-item">
            <text class="label">零售价（元）*</text>
            <input
              :value="singleSku.price"
              type="digit"
              placeholder="0.00"
              class="input input-amount"
              @input="singleSku.price = $event.detail.value"
            />
          </view>
          <view class="col-item">
            <text class="label">重量（kg）</text>
            <input
              :value="singleSku.weight"
              type="digit"
              placeholder="0"
              class="input"
              @input="singleSku.weight = $event.detail.value"
            />
          </view>
        </view>
      </view>

      <!-- 多规格 -->
      <view v-else class="multi-spec">
        <text class="label">规格值</text>
        <view class="spec-tags">
          <view class="spec-tag" v-for="(spec, idx) in multiSkus" :key="idx">
            <text class="spec-tag-name">{{ spec.spec_name || '规格' + (idx + 1) }}</text>
            <text class="spec-tag-del" @click="removeSpecItem(idx)">×</text>
          </view>
          <view class="spec-tag-add" @click="addSpecItem">+ 添加</view>
        </view>

        <view v-for="(sku, idx) in multiSkus" :key="idx" class="sku-card">
          <view class="sku-card-top">
            <text class="sku-name">{{ sku.spec_name || '规格' + (idx + 1) }}</text>
            <view class="sku-sellable-row">
              <text class="sku-sellable-label">可售</text>
              <switch :checked="sku.sellable" @change="sku.sellable = $event.detail.value" color="#4F46E5" />
            </view>
          </view>
          <view class="two-col">
            <view class="col-item">
              <text class="label">零售价（元）*</text>
              <input
                :value="sku.price"
                type="digit"
                placeholder="0.00"
                class="input input-amount"
                @input="sku.price = $event.detail.value"
              />
            </view>
            <view class="col-item">
              <text class="label">重量（kg）</text>
              <input
                :value="sku.weight"
                type="digit"
                placeholder="0"
                class="input"
                @input="sku.weight = $event.detail.value"
              />
            </view>
          </view>
        </view>
      </view>
    </view>

    <button class="btn-submit" @click="onSubmit" :loading="submitting">{{ id ? '保存' : '新增' }}</button>
    <button class="btn-delete" v-if="id" @click="onDelete">删除商品</button>
  </view>
  </SideLayout>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { onLoad } from '@dcloudio/uni-app'
import SideLayout from '@/components/SideLayout.vue'
import { getProduct, createProduct, updateProduct, deleteProduct, getProductBrands, getProductCategories } from '@/api/product'
import { safeBack } from '@/utils/navigate'

const id = ref(0)
const submitting = ref(false)
const categories = ref<any[]>([])
const brands = ref<string[]>([])
const showBrandSuggestions = ref(false)

const form = ref({
  name: '',
  category_id: undefined as number | undefined,
  brand: '',
  description: '',
  multi_spec: false,
})

interface SkuItem {
  spec_name: string
  price: string
  weight: string
  sellable: boolean
}

const singleSku = ref<SkuItem>({ spec_name: '', price: '', weight: '', sellable: true })
const multiSkus = ref<SkuItem[]>([])

// Category picker
const categoryOptions = computed(() => [{ ID: 0, name: '未分类' }, ...categories.value])
const categoryNames = computed(() => categoryOptions.value.map(c => c.name))
const selectedCategoryIdx = ref(0)

function onCategoryChange(e: any) {
  selectedCategoryIdx.value = +e.detail.value
  const cat = categoryOptions.value[selectedCategoryIdx.value]
  form.value.category_id = cat?.ID || undefined
}

// Brand suggestions
const filteredBrands = computed(() => {
  if (!form.value.brand.trim()) return brands.value.slice(0, 8)
  return brands.value.filter(b => b.toLowerCase().includes(form.value.brand.toLowerCase()))
})

function selectBrand(b: string) {
  form.value.brand = b
  showBrandSuggestions.value = false
}

function hideBrandSuggestions() {
  setTimeout(() => { showBrandSuggestions.value = false }, 200)
}

// Spec mode switch
function setSpecMode(multi: boolean) {
  if (multi === form.value.multi_spec) return
  if (!multi && multiSkus.value.length > 1) {
    uni.showModal({
      title: '切换为单规格',
      content: '切换后将只保留第一个规格，确认继续？',
      success: (res) => {
        if (res.confirm) {
          form.value.multi_spec = false
          if (multiSkus.value.length > 0) {
            singleSku.value = { ...multiSkus.value[0] }
          }
        }
      }
    })
  } else {
    form.value.multi_spec = multi
    if (multi && multiSkus.value.length === 0) {
      multiSkus.value.push({ spec_name: '', price: singleSku.value.price, weight: singleSku.value.weight, sellable: true })
    }
  }
}

function addSpecItem() {
  uni.showModal({
    title: '添加规格',
    editable: true,
    placeholderText: '请输入规格名称',
    success: (res) => {
      if (res.confirm && res.content?.trim()) {
        multiSkus.value.push({ spec_name: res.content.trim(), price: '', weight: '', sellable: true })
      }
    }
  })
}

function removeSpecItem(idx: number) {
  multiSkus.value.splice(idx, 1)
}

onLoad(async (query) => {
  // Load categories and brands in parallel
  const [catRes, brandRes] = await Promise.allSettled([
    getProductCategories(),
    getProductBrands(),
  ])
  if (catRes.status === 'fulfilled') categories.value = Array.isArray(catRes.value.data) ? catRes.value.data : []
  if (brandRes.status === 'fulfilled') brands.value = Array.isArray(brandRes.value.data) ? brandRes.value.data : []

  if (query?.id) {
    id.value = parseInt(query.id)
    await loadData()
  }
})

async function loadData() {
  const res = await getProduct(id.value)
  const d = res.data
  form.value = {
    name: d.name || '',
    category_id: d.category_id,
    brand: d.brand || '',
    description: d.description || '',
    multi_spec: !!d.multi_spec,
  }

  // Restore category picker
  if (d.category_id) {
    const idx = categoryOptions.value.findIndex(c => c.ID === d.category_id)
    if (idx >= 0) selectedCategoryIdx.value = idx
  }

  // Restore SKUs
  const skus: any[] = d.skus || []
  if (!d.multi_spec) {
    const sku = skus[0] || {}
    singleSku.value = {
      spec_name: '',
      price: sku.price != null ? String(sku.price) : '',
      weight: sku.weight != null ? String(sku.weight) : '',
      sellable: sku.sellable !== false,
    }
  } else {
    multiSkus.value = skus.map((s: any) => ({
      spec_name: s.spec_name || '',
      price: s.price != null ? String(s.price) : '',
      weight: s.weight != null ? String(s.weight) : '',
      sellable: s.sellable !== false,
    }))
  }
}

async function onSubmit() {
  if (!form.value.name.trim()) {
    uni.showToast({ title: '请填写商品名称', icon: 'none' }); return
  }

  let skus: any[]
  if (!form.value.multi_spec) {
    if (!singleSku.value.price) {
      uni.showToast({ title: '请填写零售价', icon: 'none' }); return
    }
    skus = [{
      spec_name: '',
      price: parseFloat(singleSku.value.price),
      weight: parseFloat(singleSku.value.weight) || 0,
      sellable: singleSku.value.sellable,
    }]
  } else {
    if (!multiSkus.value.length) {
      uni.showToast({ title: '请至少添加一个规格', icon: 'none' }); return
    }
    for (const sku of multiSkus.value) {
      if (!sku.spec_name.trim()) {
        uni.showToast({ title: '规格名称不能为空', icon: 'none' }); return
      }
      if (!sku.price) {
        uni.showToast({ title: `规格「${sku.spec_name}」请填写价格`, icon: 'none' }); return
      }
    }
    skus = multiSkus.value.map(s => ({
      spec_name: s.spec_name,
      price: parseFloat(s.price),
      weight: parseFloat(s.weight) || 0,
      sellable: s.sellable,
    }))
  }

  const data = {
    name: form.value.name.trim(),
    category_id: form.value.category_id || 0,
    brand: form.value.brand.trim(),
    description: form.value.description.trim(),
    multi_spec: form.value.multi_spec,
    skus,
  }

  submitting.value = true
  try {
    if (id.value) {
      await updateProduct(id.value, data)
    } else {
      await createProduct(data)
    }
    uni.showToast({ title: '保存成功', icon: 'success' })
    setTimeout(() => safeBack(), 500)
  } finally { submitting.value = false }
}

async function onDelete() {
  uni.showModal({
    title: '确认删除',
    content: '确认删除该商品？',
    success: async (res) => {
      if (res.confirm) {
        await deleteProduct(id.value)
        uni.showToast({ title: '已删除', icon: 'success' })
        setTimeout(() => safeBack(), 500)
      }
    }
  })
}
</script>

<style scoped>
.page { padding: 24rpx; max-width: 800px; background: #F5F6FA; min-height: 100vh; }
.section-title { font-size: 32rpx; font-weight: 700; color: #1F2937; margin-bottom: 16rpx; margin-top: 8rpx; }
.section-title-row {
  display: flex; justify-content: space-between; align-items: center;
  margin-bottom: 16rpx; margin-top: 8rpx;
}
.section-title-text { font-size: 32rpx; font-weight: 700; color: #1F2937; }
.section { background: #fff; border-radius: 16rpx; padding: 8rpx 24rpx; margin-bottom: 32rpx; box-shadow: 0 2rpx 8rpx rgba(0,0,0,0.04); }
.form-item { padding: 24rpx 0; border-bottom: 1rpx solid #F3F4F6; position: relative; }
.form-item:last-child { border-bottom: none; }
.label { font-size: 28rpx; color: #374151; display: block; margin-bottom: 12rpx; }
.input { background: #F9FAFB; border-radius: 8rpx; height: 72rpx; padding: 0 16rpx; font-size: 28rpx; color: #1F2937; }
.textarea { background: #F9FAFB; border-radius: 8rpx; padding: 16rpx; font-size: 28rpx; color: #1F2937; width: 100%; height: 160rpx; }
.picker { background: #F9FAFB; border-radius: 8rpx; height: 72rpx; line-height: 72rpx; padding: 0 16rpx; font-size: 28rpx; color: #1F2937; }
.brand-suggestions {
  position: absolute; left: 0; right: 0; top: 100%; z-index: 100;
  background: #fff; border-radius: 12rpx; box-shadow: 0 4rpx 16rpx rgba(0,0,0,0.1);
  max-height: 300rpx; overflow-y: auto;
}
.brand-suggest-item { padding: 20rpx 24rpx; font-size: 28rpx; color: #374151; border-bottom: 1rpx solid #F3F4F6; }
.brand-suggest-item:last-child { border-bottom: none; }

/* Spec toggle */
.spec-toggle { display: flex; background: #F3F4F6; border-radius: 20rpx; padding: 4rpx; }
.toggle-btn { padding: 10rpx 28rpx; border-radius: 16rpx; font-size: 26rpx; color: #6B7280; transition: all 0.2s; }
.toggle-active { background: #4F46E5; color: #fff; }

/* Single spec */
.two-col { display: flex; gap: 24rpx; }
.col-item { flex: 1; }

/* Multi spec tags */
.spec-tags { display: flex; flex-wrap: wrap; gap: 16rpx; padding: 16rpx 0; }
.spec-tag {
  display: flex; align-items: center;
  background: #F3F4F6; border-radius: 20rpx; padding: 8rpx 20rpx;
  font-size: 26rpx; color: #374151;
}
.spec-tag-name { }
.spec-tag-del { color: #9CA3AF; margin-left: 8rpx; font-size: 28rpx; }
.spec-tag-add {
  background: #EEF2FF; color: #4F46E5; border-radius: 20rpx;
  padding: 8rpx 20rpx; font-size: 26rpx;
}

/* SKU cards */
.sku-card {
  background: #F9FAFB; border-radius: 12rpx; padding: 20rpx;
  margin-top: 16rpx; border-left: 6rpx solid #4F46E5;
}
.sku-card-top { display: flex; justify-content: space-between; align-items: center; margin-bottom: 16rpx; }
.sku-name { font-size: 28rpx; font-weight: 600; color: #1F2937; }
.sku-sellable-row { display: flex; align-items: center; gap: 8rpx; }
.sku-sellable-label { font-size: 24rpx; color: #6B7280; }

/* Buttons */
.btn-submit { background: #4F46E5; color: #fff; border-radius: 12rpx; font-size: 30rpx; margin-top: 16rpx; }
.btn-delete { background: #fff; color: #DC2626; border: 1rpx solid #DC2626; border-radius: 12rpx; font-size: 30rpx; margin-top: 16rpx; }
</style>
