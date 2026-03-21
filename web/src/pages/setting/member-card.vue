<template>
  <SideLayout>
    <view class="page">
      <view class="header">
        <text class="title">会员卡管理</text>
        <view class="btn-add" @click="showAdd = true">+ 新增模板</view>
      </view>

      <view v-if="loading" class="loading">加载中...</view>
      <view v-else-if="list.length === 0" class="empty">暂无会员卡模板，点击上方按钮创建</view>

      <view v-else class="card-list">
        <view class="card" v-for="tpl in list" :key="tpl.ID">
          <view class="card-header" :style="{ background: getCardColor(tpl) }">
            <text class="card-name">{{ tpl.name }}</text>
            <text class="card-amount">¥{{ tpl.min_recharge.toFixed(0) }}</text>
          </view>
          <view class="card-body">
            <view class="card-info-row">
              <text class="info-label">储值门槛</text>
              <text class="info-value">¥{{ tpl.min_recharge }}</text>
            </view>
            <view class="card-info-row">
              <text class="info-label">服务折扣</text>
              <text class="info-value discount">{{ (tpl.discount_rate * 10).toFixed(1) }}折</text>
            </view>
            <view class="card-info-row">
              <text class="info-label">商品折扣</text>
              <text class="info-value discount">{{ tpl.product_discount_rate && tpl.product_discount_rate < 1 ? (tpl.product_discount_rate * 10).toFixed(1) + '折' : '无折扣' }}</text>
            </view>
            <view class="card-info-row">
              <text class="info-label">有效期</text>
              <text class="info-value">{{ tpl.valid_days > 0 ? tpl.valid_days + '天' : '永久' }}</text>
            </view>
            <view class="card-info-row">
              <text class="info-label">状态</text>
              <text :class="['info-value', tpl.status === 1 ? 'status-on' : 'status-off']">
                {{ tpl.status === 1 ? '在售' : '停售' }}
              </text>
            </view>
          </view>
          <!-- Per-category discounts -->
          <view class="card-discounts" v-if="tpl.discounts && tpl.discounts.length > 0">
            <view class="discount-row" v-for="d in tpl.discounts" :key="d.ID">
              <text class="discount-cat">{{ d.category_name }}</text>
              <text class="discount-val">{{ (d.discount_rate * 10).toFixed(1) }}折</text>
            </view>
          </view>
          <view class="card-discounts" v-else>
            <text class="no-discount">统一折扣，未设分类折扣</text>
          </view>
          <view class="card-actions">
            <text class="action-btn" @click="editTpl(tpl)">编辑</text>
            <text class="action-btn" @click="editDiscounts(tpl)">设置折扣</text>
            <text class="action-btn danger" @click="deleteTpl(tpl)">删除</text>
          </view>
        </view>
      </view>

      <!-- Add/Edit Modal -->
      <view class="modal-mask" v-if="showAdd || showEdit" @click="closeModal">
        <view class="modal-body" @click.stop>
          <text class="modal-title">{{ showEdit ? '编辑模板' : '新增模板' }}</text>
          <view class="form-item">
            <text class="label">卡名称 *</text>
            <input v-model="modalForm.name" placeholder="如：I / II / III / 金卡" class="input" />
          </view>
          <view class="form-item">
            <text class="label">储值门槛 (元) *</text>
            <input v-model="modalForm.min_recharge" type="digit" placeholder="1000" class="input" />
          </view>
          <view class="form-item">
            <text class="label">服务折扣 * (如0.8=八折)</text>
            <input v-model="modalForm.discount_rate" type="digit" placeholder="0.9" class="input" />
          </view>
          <view class="form-item">
            <text class="label">商品折扣 (如0.9=九折，1=无折扣)</text>
            <input v-model="modalForm.product_discount_rate" type="digit" placeholder="1" class="input" />
          </view>
          <view class="form-item">
            <text class="label">有效天数 (0=永久)</text>
            <input v-model="modalForm.valid_days" type="number" placeholder="0" class="input" />
          </view>
          <view class="form-item">
            <text class="label">卡片颜色</text>
            <view class="color-picker">
              <view
                v-for="c in colorPresets" :key="c.value"
                :class="['color-swatch', modalForm.color === c.value ? 'swatch-active' : '']"
                :style="{ background: c.value }"
                @click="modalForm.color = c.value"
              >
                <text class="swatch-label">{{ c.name }}</text>
              </view>
            </view>
          </view>
          <view class="modal-btns">
            <view class="modal-btn cancel" @click="closeModal">取消</view>
            <view class="modal-btn confirm" @click="onSubmit">确定</view>
          </view>
        </view>
      </view>
      <!-- Discount edit modal -->
      <view class="modal-mask" v-if="showDiscountModal" @click="showDiscountModal = false">
        <view class="modal-body" @click.stop>
          <text class="modal-title">设置分类折扣 - {{ discountTplName }}</text>
          <text class="discount-hint">为每个一级分类设置不同折扣（留空或1表示不打折）</text>
          <view class="discount-form" v-for="(d, idx) in discountForm" :key="d.category_id">
            <text class="discount-form-cat">{{ d.category_name }}</text>
            <view class="discount-input-wrap">
              <input v-model="d.discount_rate" type="digit" placeholder="如0.8" class="input discount-input" />
              <text class="discount-unit">{{ d.discount_rate && parseFloat(d.discount_rate) < 1 ? (parseFloat(d.discount_rate) * 10).toFixed(1) + '折' : '无折扣' }}</text>
            </view>
          </view>
          <view v-if="discountForm.length === 0" class="empty-sm">请先在"服务管理"中创建一级分类</view>
          <view class="modal-btns">
            <view class="modal-btn cancel" @click="showDiscountModal = false">取消</view>
            <view class="modal-btn confirm" @click="saveDiscounts">保存</view>
          </view>
        </view>
      </view>
    </view>
  </SideLayout>
</template>

<script setup lang="ts">
import SideLayout from '@/components/SideLayout.vue'
import { ref } from 'vue'
import { onShow } from '@dcloudio/uni-app'
import { getCardTemplates, createCardTemplate, updateCardTemplate, deleteCardTemplate, setTemplateDiscounts } from '@/api/member-card'
import { getCategoryTree } from '@/api/service-category'

const list = ref<MemberCardTemplate[]>([])
const loading = ref(true)
const showAdd = ref(false)
const showEdit = ref(false)
const editId = ref(0)

const modalForm = ref({ name: '', min_recharge: '', discount_rate: '', product_discount_rate: '1', valid_days: '0', color: 'linear-gradient(135deg, #4F46E5, #7C3AED)' })

const colorPresets = [
  { name: '紫', value: 'linear-gradient(135deg, #4F46E5, #7C3AED)' },
  { name: '金', value: 'linear-gradient(135deg, #F59E0B, #D97706)' },
  { name: '绿', value: 'linear-gradient(135deg, #10B981, #059669)' },
  { name: '蓝', value: 'linear-gradient(135deg, #3B82F6, #2563EB)' },
  { name: '红', value: 'linear-gradient(135deg, #EF4444, #DC2626)' },
  { name: '粉', value: 'linear-gradient(135deg, #EC4899, #DB2777)' },
  { name: '黑金', value: 'linear-gradient(135deg, #1F2937, #92400E)' },
]

function getCardColor(tpl: MemberCardTemplate) {
  return tpl.color || 'linear-gradient(135deg, #4F46E5, #7C3AED)'
}

onShow(loadData)

async function loadData() {
  loading.value = true
  try {
    const res = await getCardTemplates()
    list.value = res.data || []
  } finally { loading.value = false }
}

function editTpl(tpl: MemberCardTemplate) {
  editId.value = tpl.ID
  modalForm.value = {
    name: tpl.name,
    min_recharge: String(tpl.min_recharge),
    discount_rate: String(tpl.discount_rate),
    product_discount_rate: String(tpl.product_discount_rate || 1),
    valid_days: String(tpl.valid_days),
    color: tpl.color || 'linear-gradient(135deg, #4F46E5, #7C3AED)',
  }
  showEdit.value = true
}

// Discount editing
const showDiscountModal = ref(false)
const discountTplId = ref(0)
const discountTplName = ref('')
const discountForm = ref<{ category_id: number; category_name: string; discount_rate: string }[]>([])
const serviceCategories = ref<ServiceCategory[]>([])

async function editDiscounts(tpl: MemberCardTemplate) {
  discountTplId.value = tpl.ID
  discountTplName.value = tpl.name

  // Load service categories if not loaded
  if (serviceCategories.value.length === 0) {
    try {
      const res = await getCategoryTree()
      serviceCategories.value = res.data || []
    } catch {}
  }

  // Build form: one row per 1st-level category, pre-fill existing discounts
  discountForm.value = serviceCategories.value.map(cat => {
    const existing = (tpl.discounts || []).find(d => d.category_id === cat.ID)
    return {
      category_id: cat.ID,
      category_name: cat.name,
      discount_rate: existing ? String(existing.discount_rate) : '',
    }
  })

  showDiscountModal.value = true
}

async function saveDiscounts() {
  const discounts = discountForm.value
    .filter(d => d.discount_rate && parseFloat(d.discount_rate) > 0 && parseFloat(d.discount_rate) < 1)
    .map(d => ({
      category_id: d.category_id,
      category_name: d.category_name,
      discount_rate: parseFloat(d.discount_rate),
    }))

  try {
    await setTemplateDiscounts(discountTplId.value, discounts)
    uni.showToast({ title: '保存成功', icon: 'success' })
    showDiscountModal.value = false
    await loadData()
  } catch (e: any) {
    uni.showToast({ title: e.message || '保存失败', icon: 'none' })
  }
}

function closeModal() {
  showAdd.value = false
  showEdit.value = false
  modalForm.value = { name: '', min_recharge: '', discount_rate: '', product_discount_rate: '1', valid_days: '0', color: 'linear-gradient(135deg, #4F46E5, #7C3AED)' }
}

async function onSubmit() {
  const f = modalForm.value
  if (!f.name || !f.min_recharge || !f.discount_rate) {
    uni.showToast({ title: '请填写必填项', icon: 'none' }); return
  }
  const data = {
    name: f.name,
    min_recharge: parseFloat(f.min_recharge),
    discount_rate: parseFloat(f.discount_rate),
    product_discount_rate: parseFloat(f.product_discount_rate) || 1,
    valid_days: parseInt(f.valid_days) || 0,
    color: f.color,
  }
  try {
    if (showEdit.value) {
      await updateCardTemplate(editId.value, data)
    } else {
      await createCardTemplate(data)
    }
    uni.showToast({ title: '保存成功', icon: 'success' })
    closeModal()
    await loadData()
  } catch (e: any) {
    uni.showToast({ title: e.message || '操作失败', icon: 'none' })
  }
}

function deleteTpl(tpl: MemberCardTemplate) {
  uni.showModal({
    title: '确认删除', content: `删除会员卡模板"${tpl.name}"？`,
    success: async (res) => {
      if (res.confirm) {
        try {
          await deleteCardTemplate(tpl.ID)
          uni.showToast({ title: '已删除', icon: 'success' })
          await loadData()
        } catch (e: any) {
          uni.showToast({ title: e.message || '删除失败', icon: 'none' })
        }
      }
    }
  })
}
</script>

<style scoped>
.page { padding: 24rpx; max-width: 900px; }
.header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 24rpx; }
.title { font-size: 36rpx; font-weight: bold; color: #1F2937; }
.btn-add { font-size: 28rpx; color: #fff; background: #4F46E5; padding: 12rpx 24rpx; border-radius: 12rpx; }
.loading, .empty { text-align: center; padding: 100rpx 0; color: #9CA3AF; font-size: 28rpx; }

.card-list { display: flex; flex-wrap: wrap; gap: 24rpx; }
.card { width: calc(50% - 12rpx); border-radius: 20rpx; overflow: hidden; box-shadow: 0 4rpx 16rpx rgba(0,0,0,0.08); background: #fff; }
.card-header { padding: 32rpx 24rpx; color: #fff; }
.card-name { font-size: 32rpx; font-weight: 700; display: block; }
.card-amount { font-size: 48rpx; font-weight: 800; display: block; margin-top: 12rpx; }
.card-body { padding: 20rpx 24rpx; }
.card-info-row { display: flex; justify-content: space-between; padding: 8rpx 0; font-size: 26rpx; }
.info-label { color: #6B7280; }
.info-value { color: #1F2937; font-weight: 500; }
.info-value.discount { color: #4F46E5; font-weight: 700; }
.status-on { color: #059669; }
.status-off { color: #DC2626; }
.card-actions { display: flex; justify-content: flex-end; gap: 24rpx; padding: 16rpx 24rpx; border-top: 1rpx solid #F3F4F6; }
.action-btn { font-size: 26rpx; color: #4F46E5; }
.action-btn.danger { color: #EF4444; }

/* Modal */
.modal-mask { position: fixed; top: 0; left: 0; right: 0; bottom: 0; background: rgba(0,0,0,0.5); display: flex; align-items: center; justify-content: center; z-index: 999; }
.modal-body { background: #fff; border-radius: 20rpx; padding: 40rpx; width: 600rpx; }
.modal-title { font-size: 32rpx; font-weight: 700; color: #1F2937; display: block; text-align: center; margin-bottom: 24rpx; }
.form-item { margin-bottom: 20rpx; }
.label { font-size: 26rpx; color: #374151; display: block; margin-bottom: 8rpx; }
.input { font-size: 28rpx; color: #1F2937; height: 60rpx; background: #F9FAFB; border-radius: 8rpx; padding: 0 16rpx; }
.color-picker { display: flex; flex-wrap: wrap; gap: 12rpx; }
.color-swatch { width: 72rpx; height: 72rpx; border-radius: 12rpx; display: flex; align-items: center; justify-content: center; border: 3rpx solid transparent; transition: all 0.2s; }
.swatch-active { border-color: #1F2937; box-shadow: 0 2rpx 12rpx rgba(0,0,0,0.3); transform: scale(1.1); }
.swatch-label { font-size: 20rpx; color: #fff; font-weight: 600; }
.modal-btns { display: flex; gap: 16rpx; margin-top: 24rpx; }
.modal-btn { flex: 1; text-align: center; padding: 18rpx; border-radius: 12rpx; font-size: 28rpx; }
.cancel { background: #F3F4F6; color: #6B7280; }
.confirm { background: #4F46E5; color: #fff; }

/* Discounts in card */
.card-discounts { padding: 12rpx 24rpx; border-top: 1rpx solid #F3F4F6; }
.discount-row { display: flex; justify-content: space-between; padding: 6rpx 0; font-size: 24rpx; }
.discount-cat { color: #6B7280; }
.discount-val { color: #4F46E5; font-weight: 600; }
.no-discount { font-size: 22rpx; color: #9CA3AF; }

/* Discount modal */
.discount-hint { font-size: 22rpx; color: #9CA3AF; display: block; margin-bottom: 16rpx; text-align: center; }
.discount-form { display: flex; align-items: center; gap: 16rpx; margin-bottom: 12rpx; }
.discount-form-cat { font-size: 28rpx; color: #374151; font-weight: 500; width: 160rpx; }
.discount-input-wrap { display: flex; align-items: center; gap: 8rpx; flex: 1; }
.discount-input { width: 160rpx !important; text-align: center; }
.discount-unit { font-size: 24rpx; color: #4F46E5; white-space: nowrap; }
.empty-sm { font-size: 26rpx; color: #9CA3AF; text-align: center; padding: 24rpx; }

@media (max-width: 600px) {
  .card { width: 100%; }
}
</style>
