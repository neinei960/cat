<template>
  <SideLayout>
    <view class="page">
      <view class="header">
        <text class="title">会员卡管理</text>
        <view class="btn-add" @click="onAdd">+ 新增模板</view>
      </view>

      <view v-if="loading" class="loading">加载中...</view>
      <view v-else-if="list.length === 0" class="empty">暂无会员卡模板，点击上方按钮创建</view>

      <view v-else class="card-list">
        <view class="card" v-for="tpl in list" :key="tpl.ID">
          <view class="card-header" :style="{ background: getCardColor(tpl) }">
            <view class="card-name-row">
              <text class="card-name">{{ tpl.name }}</text>
              <text class="card-type-badge" v-if="tpl.card_type === 2">次卡</text>
            </view>
            <text class="card-amount" v-if="tpl.card_type !== 2">¥{{ tpl.min_recharge.toFixed(0) }}</text>
            <text class="card-amount" v-else>{{ tpl.total_times }}次 / ¥{{ tpl.times_price }}</text>
          </view>
          <view class="card-body">
            <view class="card-info-row" v-if="tpl.card_type !== 2">
              <text class="info-label">储值门槛</text>
              <text class="info-value">¥{{ tpl.min_recharge }}</text>
            </view>
            <view class="card-info-row" v-if="tpl.card_type === 2">
              <text class="info-label">总次数</text>
              <text class="info-value">{{ tpl.total_times }}次</text>
            </view>
            <view class="card-info-row">
              <text class="info-label">服务折扣</text>
              <text class="info-value discount">{{ (tpl.discount_rate * 10).toFixed(1) }}折</text>
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
          <view class="card-discounts">
            <view class="discount-row" v-if="tpl.product_discount_rate && tpl.product_discount_rate < 1">
              <text class="discount-cat">商品折扣</text>
              <text class="discount-val">{{ (tpl.product_discount_rate * 10).toFixed(1) }}折</text>
            </view>
            <view class="discount-row" v-for="d in (tpl.discounts || [])" :key="d.ID">
              <text class="discount-cat">{{ d.category_name }}</text>
              <text class="discount-val">{{ (d.discount_rate * 10).toFixed(1) }}折</text>
            </view>
            <text class="no-discount" v-if="(!tpl.discounts || tpl.discounts.length === 0) && !(tpl.product_discount_rate && tpl.product_discount_rate < 1)">统一折扣，未设分类折扣</text>
          </view>
          <view class="card-actions">
            <text class="action-btn" @click="editTpl(tpl)">编辑</text>
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
            <text class="label">卡类型</text>
            <view class="type-tabs">
              <view :class="['type-tab', modalForm.card_type === '1' ? 'active' : '']" @click="modalForm.card_type = '1'">储值卡</view>
              <view :class="['type-tab', modalForm.card_type === '2' ? 'active' : '']" @click="modalForm.card_type = '2'">次卡</view>
            </view>
          </view>
          <view class="form-item" v-if="modalForm.card_type === '1'">
            <text class="label">储值门槛 (元) *</text>
            <input v-model="modalForm.min_recharge" type="digit" placeholder="1000" class="input input-amount" />
          </view>
          <view class="form-item" v-if="modalForm.card_type === '2'">
            <text class="label">总次数 *</text>
            <input v-model="modalForm.total_times" type="number" placeholder="10" class="input" />
          </view>
          <view class="form-item" v-if="modalForm.card_type === '2'">
            <text class="label">售价 (元) *</text>
            <input v-model="modalForm.times_price" type="digit" placeholder="800" class="input input-amount" />
          </view>
          <view class="form-item">
            <text class="label">服务折扣 * (如0.8=八折)</text>
            <input v-model="modalForm.discount_rate" type="digit" placeholder="0.9" class="input" />
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
          <!-- 折扣设置（内嵌） -->
          <view class="discount-section">
            <text class="discount-section-title">折扣设置</text>
            <text class="discount-hint">留空或1表示不打折</text>
            <view class="discount-form">
              <text class="discount-form-cat">商品统一折扣</text>
              <view class="discount-input-wrap">
                <input v-model="modalForm.product_discount_rate" type="digit" placeholder="如0.9" class="input discount-input" />
                <text class="discount-unit">{{ modalForm.product_discount_rate && parseFloat(modalForm.product_discount_rate) < 1 ? (parseFloat(modalForm.product_discount_rate) * 10).toFixed(1) + '折' : '无折扣' }}</text>
              </view>
            </view>
            <view class="discount-form" v-for="d in discountForm" :key="d.category_id">
              <text class="discount-form-cat">{{ d.category_name }}</text>
              <view class="discount-input-wrap">
                <input v-model="d.discount_rate" type="digit" placeholder="如0.8" class="input discount-input" />
                <text class="discount-unit">{{ d.discount_rate && parseFloat(d.discount_rate) < 1 ? (parseFloat(d.discount_rate) * 10).toFixed(1) + '折' : '无折扣' }}</text>
              </view>
            </view>
            <view v-if="discountForm.length === 0" class="empty-sm">暂无服务分类，商品折扣已可设置</view>
          </view>
          <view class="modal-btns">
            <view class="modal-btn cancel" @click="closeModal">取消</view>
            <view class="modal-btn confirm" @click="onSubmit">确定</view>
          </view>
        </view>
      </view>
    </view>
  </SideLayout>
</template>

<script setup lang="ts">
import SideLayout from '@/components/SideLayout.vue'
import { ref } from 'vue'
import { onLoad, onShow } from '@dcloudio/uni-app'
import { getCardTemplates, createCardTemplate, updateCardTemplate, deleteCardTemplate, setTemplateDiscounts } from '@/api/member-card'
import { getCategoryTree } from '@/api/service-category'

const list = ref<MemberCardTemplate[]>([])
const loading = ref(true)
const showAdd = ref(false)
const showEdit = ref(false)
const editId = ref(0)
const returnToOrderCreate = ref(false)
const returnCustomerId = ref(0)
const ORDER_CREATE_REFRESH_KEY = 'order_create_refresh_member_card'
const ORDER_CREATE_RETURN_CUSTOMER_KEY = 'order_create_return_customer_id'

const modalForm = ref({ name: '', card_type: '1', min_recharge: '', discount_rate: '', product_discount_rate: '1', valid_days: '0', total_times: '', times_price: '', color: 'linear-gradient(135deg, #4F46E5, #7C3AED)' })

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

onLoad((query) => {
  returnToOrderCreate.value = query?.return_to === 'order_create'
  returnCustomerId.value = parseInt(String(query?.customer_id || 0)) || 0
})

onShow(loadData)

async function loadData() {
  loading.value = true
  try {
    const res = await getCardTemplates()
    list.value = res.data || []
  } finally { loading.value = false }
}

async function editTpl(tpl: MemberCardTemplate) {
  editId.value = tpl.ID
  modalForm.value = {
    name: tpl.name,
    card_type: String(tpl.card_type || 1),
    min_recharge: String(tpl.min_recharge),
    discount_rate: String(tpl.discount_rate),
    product_discount_rate: String(tpl.product_discount_rate || 1),
    valid_days: String(tpl.valid_days),
    total_times: String(tpl.total_times || ''),
    times_price: String(tpl.times_price || ''),
    color: tpl.color || 'linear-gradient(135deg, #4F46E5, #7C3AED)',
  }
  await loadDiscountForm(tpl)
  showEdit.value = true
}

async function loadDiscountForm(tpl?: MemberCardTemplate) {
  if (serviceCategories.value.length === 0) {
    try {
      const res = await getCategoryTree()
      serviceCategories.value = res.data || []
    } catch {}
  }
  discountForm.value = serviceCategories.value.map(cat => {
    const existing = tpl?.discounts?.find(d => d.category_id === cat.ID)
    return {
      category_id: cat.ID,
      category_name: cat.name,
      discount_rate: existing ? String(existing.discount_rate) : '',
    }
  })
}

const discountForm = ref<{ category_id: number; category_name: string; discount_rate: string }[]>([])
const serviceCategories = ref<ServiceCategory[]>([])

function closeModal() {
  showAdd.value = false
  showEdit.value = false
  modalForm.value = { name: '', card_type: '1', min_recharge: '', discount_rate: '', product_discount_rate: '1', valid_days: '0', total_times: '', times_price: '', color: 'linear-gradient(135deg, #4F46E5, #7C3AED)' }
}

async function onAdd() {
  await loadDiscountForm()
  showAdd.value = true
}

async function onSubmit() {
  const f = modalForm.value
  const isTimesCard = f.card_type === '2'
  if (!f.name || !f.discount_rate) {
    uni.showToast({ title: '请填写必填项', icon: 'none' }); return
  }
  if (isTimesCard && (!f.total_times || !f.times_price)) {
    uni.showToast({ title: '次卡请填写总次数和售价', icon: 'none' }); return
  }
  if (!isTimesCard && !f.min_recharge) {
    uni.showToast({ title: '储值卡请填写储值门槛', icon: 'none' }); return
  }
  const data: any = {
    name: f.name,
    card_type: parseInt(f.card_type) || 1,
    min_recharge: parseFloat(f.min_recharge) || 0,
    discount_rate: parseFloat(f.discount_rate),
    product_discount_rate: parseFloat(f.product_discount_rate) || 1,
    valid_days: parseInt(f.valid_days) || 0,
    total_times: isTimesCard ? parseInt(f.total_times) || 0 : 0,
    times_price: isTimesCard ? parseFloat(f.times_price) || 0 : 0,
    color: f.color,
  }
  try {
    let tplId = editId.value
    if (showEdit.value) {
      await updateCardTemplate(tplId, data)
    } else {
      const res = await createCardTemplate(data)
      tplId = res.data?.ID
    }
    // 保存分类折扣
    if (tplId) {
      const discounts = discountForm.value
        .filter(d => d.discount_rate && parseFloat(d.discount_rate) > 0 && parseFloat(d.discount_rate) < 1)
        .map(d => ({
          category_id: d.category_id,
          category_name: d.category_name,
          discount_rate: parseFloat(d.discount_rate),
        }))
      await setTemplateDiscounts(tplId, discounts)
    }
    uni.showToast({ title: '保存成功', icon: 'success' })
    if (returnToOrderCreate.value) {
      uni.setStorageSync(ORDER_CREATE_REFRESH_KEY, '1')
      if (returnCustomerId.value > 0) {
        uni.setStorageSync(ORDER_CREATE_RETURN_CUSTOMER_KEY, String(returnCustomerId.value))
      } else {
        uni.removeStorageSync(ORDER_CREATE_RETURN_CUSTOMER_KEY)
      }
      setTimeout(() => {
        uni.reLaunch({ url: '/pages/order/create' })
      }, 250)
      return
    }
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
.modal-mask {
  position: fixed;
  inset: 0;
  background: rgba(0,0,0,0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 32rpx;
  box-sizing: border-box;
  z-index: 5000;
  overflow: hidden;
}
.modal-body {
  background: #fff;
  border-radius: 20rpx;
  padding: 40rpx 40rpx calc(40rpx + env(safe-area-inset-bottom));
  width: min(600rpx, 100%);
  max-width: calc(100vw - 64rpx);
  max-height: calc(100vh - 96rpx - env(safe-area-inset-bottom));
  overflow-y: auto;
  -webkit-overflow-scrolling: touch;
  box-sizing: border-box;
}
.modal-title { font-size: 32rpx; font-weight: 700; color: #1F2937; display: block; text-align: center; margin-bottom: 24rpx; }
.form-item { margin-bottom: 20rpx; }
.label { font-size: 26rpx; color: #374151; display: block; margin-bottom: 8rpx; }
.input { font-size: 28rpx; color: #1F2937; height: 60rpx; background: #F9FAFB; border-radius: 8rpx; padding: 0 16rpx; }
.color-picker { display: flex; flex-wrap: wrap; gap: 12rpx; }
.color-swatch { width: 72rpx; height: 72rpx; border-radius: 12rpx; display: flex; align-items: center; justify-content: center; border: 3rpx solid transparent; transition: all 0.2s; }
.swatch-active { border-color: #1F2937; box-shadow: 0 2rpx 12rpx rgba(0,0,0,0.3); transform: scale(1.1); }
.swatch-label { font-size: 20rpx; color: #fff; font-weight: 600; }
.modal-btns {
  display: flex;
  gap: 16rpx;
  margin-top: 24rpx;
  position: sticky;
  bottom: calc(-24rpx - env(safe-area-inset-bottom));
  padding-top: 16rpx;
  padding-bottom: env(safe-area-inset-bottom);
  background: linear-gradient(180deg, rgba(255,255,255,0.88), #fff 28rpx);
}
.modal-btn { flex: 1; text-align: center; padding: 18rpx; border-radius: 12rpx; font-size: 28rpx; }
.cancel { background: #F3F4F6; color: #6B7280; }
.confirm { background: #4F46E5; color: #fff; }

/* Discounts in card */
.card-discounts { padding: 12rpx 24rpx; border-top: 1rpx solid #F3F4F6; }
.discount-row { display: flex; justify-content: space-between; padding: 6rpx 0; font-size: 24rpx; }
.discount-cat { color: #6B7280; }
.discount-val { color: #4F46E5; font-weight: 600; }
.no-discount { font-size: 22rpx; color: #9CA3AF; }

/* Discount section in modal */
.discount-section { margin-top: 20rpx; padding-top: 20rpx; border-top: 1rpx solid #E5E7EB; }
.discount-section-title { font-size: 28rpx; font-weight: 600; color: #374151; display: block; margin-bottom: 8rpx; }
.discount-hint { font-size: 22rpx; color: #9CA3AF; display: block; margin-bottom: 16rpx; }
.discount-form { display: flex; align-items: center; gap: 16rpx; margin-bottom: 12rpx; }
.discount-form-cat { font-size: 28rpx; color: #374151; font-weight: 500; width: 160rpx; }
.discount-input-wrap { display: flex; align-items: center; gap: 8rpx; flex: 1; }
.discount-input { width: 160rpx !important; text-align: center; }
.discount-unit { font-size: 24rpx; color: #4F46E5; white-space: nowrap; }
.empty-sm { font-size: 26rpx; color: #9CA3AF; text-align: center; padding: 24rpx; }

.card-name-row { display: flex; align-items: center; gap: 8rpx; }
.card-type-badge { font-size: 18rpx; background: rgba(255,255,255,0.3); padding: 2rpx 12rpx; border-radius: 8rpx; color: #fff; }
.type-tabs { display: flex; gap: 12rpx; }
.type-tab { flex: 1; text-align: center; padding: 16rpx; border-radius: 12rpx; background: #F3F4F6; color: #374151; font-size: 26rpx; font-weight: 500; }
.type-tab.active { background: #4F46E5; color: #fff; }

@media (max-width: 600px) {
  .card { width: 100%; }
  .modal-mask {
    align-items: flex-end;
    padding: 24rpx 24rpx calc(24rpx + env(safe-area-inset-bottom));
  }
  .modal-body {
    width: 100%;
    max-width: none;
    max-height: calc(100vh - 112rpx - env(safe-area-inset-bottom));
    border-radius: 24rpx 24rpx 0 0;
  }
}
</style>
