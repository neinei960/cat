<template>
  <view class="page">
    <view class="section-title">基础信息</view>
    <view class="form">
      <view class="form-item">
        <text class="label">服务名称 *</text>
        <input v-model="form.name" placeholder="如：标准洗澡" class="input" />
      </view>
      <view class="form-item">
        <text class="label">服务分类</text>
        <view class="cat-picker">
          <picker :range="parentNames" :value="selectedParentIdx" @change="onParentChange">
            <view class="picker">{{ parentNames[selectedParentIdx] || '选择一级分类' }}</view>
          </picker>
          <picker v-if="childOptions.length" :range="childNames" :value="selectedChildIdx" @change="onChildChange">
            <view class="picker">{{ childNames[selectedChildIdx] || '选择二级分类' }}</view>
          </picker>
        </view>
      </view>
      <view class="form-item">
        <text class="label">基础价格 (元) *</text>
        <input v-model="form.base_price" type="digit" placeholder="0.00" class="input" />
      </view>
      <view class="form-item">
        <text class="label">时长 (分钟) *</text>
        <input v-model="form.duration" type="number" placeholder="60" class="input" />
      </view>
      <view class="form-item">
        <text class="label">描述</text>
        <textarea v-model="form.description" placeholder="服务描述" class="textarea" />
      </view>
      <view class="form-item">
        <text class="label">排序</text>
        <input v-model="form.sort_order" type="number" placeholder="0" class="input" />
      </view>
    </view>

    <!-- 服务规格/项目 -->
    <view class="section-title">
      <text>服务规格</text>
      <view class="btn-add-spec" @click="addSpec">+ 添加规格</view>
    </view>
    <view class="specs-hint" v-if="specs.length === 0">
      <text>可添加多个规格，如：短毛/长毛/A/B/C/D，每个规格有独立的价格和时长</text>
    </view>

    <view class="spec-list" v-if="specs.length > 0">
      <!-- 表头 -->
      <view class="spec-header">
        <text class="spec-col-name">规格名称 *</text>
        <text class="spec-col-price">价格 (元) *</text>
        <text class="spec-col-dur">时长 (分钟)</text>
        <text class="spec-col-act">操作</text>
      </view>

      <!-- 规格行 -->
      <view class="spec-row" v-for="(spec, idx) in specs" :key="idx">
        <view class="spec-col-name">
          <input v-model="spec.name" placeholder="如：短毛猫" class="spec-input" />
        </view>
        <view class="spec-col-price">
          <input v-model="spec.price" type="digit" placeholder="0" class="spec-input" />
        </view>
        <view class="spec-col-dur">
          <input v-model="spec.duration" type="number" placeholder="60" class="spec-input" />
        </view>
        <view class="spec-col-act">
          <text class="spec-del" @click="removeSpec(idx)">删除</text>
        </view>
      </view>
    </view>

    <button class="btn-submit" @click="onSubmit" :loading="submitting">{{ id ? '保存' : '新增' }}</button>
    <button class="btn-delete" v-if="id" @click="onDelete">删除服务</button>
  </view>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { onLoad } from '@dcloudio/uni-app'
import { getService, createService, updateService, deleteService, getPriceRules, createPriceRule, deletePriceRule } from '@/api/service'
import { getCategoryTree } from '@/api/service-category'
import { safeBack } from '@/utils/navigate'

const id = ref(0)
const submitting = ref(false)
const categories = ref<ServiceCategory[]>([])

const form = ref({
  name: '', category: '', category_id: undefined as number | undefined,
  base_price: '', duration: '',
  description: '', applicable_species: '', applicable_sizes: '', sort_order: 0
})

// Specs (price rules)
interface SpecItem {
  id?: number  // existing rule ID
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

function onParentChange(e: any) {
  selectedParentIdx.value = +e.detail.value
  selectedChildIdx.value = 0
  updateCategoryId()
}
function onChildChange(e: any) {
  selectedChildIdx.value = +e.detail.value
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
      duration: String(r.duration || ''),
    }))
  } catch {}
}

async function onSubmit() {
  if (!form.value.name || !form.value.base_price || !form.value.duration) {
    uni.showToast({ title: '请填写必填项', icon: 'none' }); return
  }

  // Validate specs
  for (const spec of specs.value) {
    if (!spec.name.trim() || !spec.price) {
      uni.showToast({ title: '规格名称和价格为必填', icon: 'none' }); return
    }
  }

  submitting.value = true
  try {
    const data = {
      ...form.value,
      base_price: parseFloat(form.value.base_price),
      duration: parseInt(form.value.duration),
    }

    let serviceId = id.value
    if (id.value) {
      await updateService(id.value, data)
    } else {
      const res = await createService(data)
      serviceId = res.data.ID
    }

    // Sync specs: delete old rules, create new ones
    for (const oldId of existingRuleIds.value) {
      await deletePriceRule(serviceId, oldId)
    }
    for (const spec of specs.value) {
      if (spec.name.trim()) {
        await createPriceRule(serviceId, {
          name: spec.name.trim(),
          price: parseFloat(spec.price),
          duration: parseInt(spec.duration) || 0,
        })
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
.input { font-size: 28rpx; color: #1F2937; height: 60rpx; }
.textarea { font-size: 28rpx; color: #1F2937; width: 100%; height: 160rpx; }
.picker { font-size: 28rpx; color: #1F2937; height: 60rpx; line-height: 60rpx; background: #F9FAFB; padding: 0 16rpx; border-radius: 8rpx; }
.cat-picker { display: flex; gap: 24rpx; }
.cat-picker picker { flex: 1; }

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
.spec-col-price { flex: 2; }
.spec-col-dur { flex: 2; }
.spec-col-act { flex: 1; text-align: center; }
.spec-input {
  font-size: 28rpx; color: #1F2937; height: 56rpx;
  background: #F9FAFB; border-radius: 8rpx; padding: 0 12rpx;
}
.spec-del { font-size: 24rpx; color: #EF4444; }

.btn-submit { background: #4F46E5; color: #fff; border-radius: 12rpx; font-size: 30rpx; margin-top: 16rpx; }
.btn-delete { background: #fff; color: #DC2626; border: 1rpx solid #DC2626; border-radius: 12rpx; font-size: 30rpx; margin-top: 16rpx; }
</style>
