<template>
  <SideLayout>
    <view class="page">
      <view class="header">
        <view class="header-copy">
          <text class="title">特殊寄养项目</text>
          <text class="subtitle">维护寄养期间按天额外加收的项目，建单时可改日价和填写特殊天数。</text>
        </view>
        <view class="btn btn-primary create-btn" @click="startCreate">新建项目</view>
      </view>

      <view v-if="editing" class="card">
        <text class="section-title">{{ form.id ? '编辑项目' : '新建项目' }}</text>
        <view class="field">
          <text class="field-label">项目名称</text>
          <input v-model="form.name" class="input" placeholder="例如：用药护理" />
        </view>
        <view class="field">
          <text class="field-label">默认日价（元/天）</text>
          <input v-model="form.default_daily_price" class="input" type="digit" placeholder="例如 10" />
        </view>
        <view class="field">
          <text class="field-label">状态</text>
          <picker :range="statusOptions" :value="statusIndex" @change="onStatusChange">
            <view class="picker">{{ statusOptions[statusIndex] }}</view>
          </picker>
        </view>
        <view class="field">
          <text class="field-label">备注</text>
          <textarea v-model="form.remark" class="textarea" placeholder="例如：绝育恢复、单独观察、喂药" />
        </view>
        <view class="actions">
          <view class="btn" @click="cancelEdit">取消</view>
          <view class="btn btn-primary" @click="save">保存</view>
        </view>
      </view>

      <view v-if="items.length === 0" class="empty-card">还没有特殊寄养项目，先创建一个。</view>

      <view
        v-for="item in items"
        :key="item.ID"
        class="card"
        @longpress="confirmDelete(item)"
        @touchstart="startDeletePress(item)"
        @touchmove="cancelDeletePress"
        @touchend="cancelDeletePress"
        @touchcancel="cancelDeletePress"
      >
        <view class="item-head">
          <view>
            <text class="item-title">{{ item.name }}</text>
            <text class="item-meta">默认 ¥{{ Number(item.default_daily_price || 0).toFixed(2) }}/天 · {{ item.status === 1 ? '启用' : '停用' }}</text>
          </view>
          <view class="item-edit" @click="edit(item)">编辑</view>
        </view>
        <text v-if="item.remark" class="item-remark">{{ item.remark }}</text>
      </view>
    </view>
  </SideLayout>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { onShow } from '@dcloudio/uni-app'
import SideLayout from '@/components/SideLayout.vue'
import { createBoardingSpecialItem, deleteBoardingSpecialItem, getBoardingSpecialItems, updateBoardingSpecialItem } from '@/api/boarding'

const items = ref<BoardingSpecialItem[]>([])
const editing = ref(false)
const deleteConfirming = ref(false)
let deletePressTimer: ReturnType<typeof setTimeout> | null = null
const statusOptions = ['启用', '停用']
const form = ref({
  id: 0,
  name: '',
  default_daily_price: '30',
  sort_order: '10',
  status: 1,
  remark: '',
})

const statusIndex = computed(() => (form.value.status === 1 ? 0 : 1))

function onStatusChange(e: any) {
  form.value.status = e.detail.value === 0 ? 1 : 0
}

function resetForm() {
  form.value = {
    id: 0,
    name: '',
    default_daily_price: '30',
    sort_order: '10',
    status: 1,
    remark: '',
  }
}

function startCreate() {
  resetForm()
  editing.value = true
}

function cancelEdit() {
  editing.value = false
}

function edit(item: BoardingSpecialItem) {
  editing.value = true
  form.value = {
    id: item.ID,
    name: item.name,
    default_daily_price: String(item.default_daily_price || 0),
    sort_order: String(item.sort_order || 0),
    status: item.status || 1,
    remark: item.remark || '',
  }
}

async function loadData() {
  const res = await getBoardingSpecialItems()
  items.value = res.data || []
}

async function save() {
  const name = form.value.name.trim()
  const defaultDailyPrice = Number(form.value.default_daily_price || 0)
  if (!name) {
    uni.showToast({ title: '请填写项目名称', icon: 'none' })
    return
  }
  if (!Number.isFinite(defaultDailyPrice) || defaultDailyPrice <= 0) {
    uni.showToast({ title: '请填写有效默认日价', icon: 'none' })
    return
  }
  const payload = {
    name,
    default_daily_price: defaultDailyPrice,
    sort_order: Number(form.value.sort_order || 0),
    status: form.value.status,
    remark: form.value.remark.trim(),
  }
  if (form.value.id) await updateBoardingSpecialItem(form.value.id, payload)
  else await createBoardingSpecialItem(payload)
  uni.showToast({ title: '保存成功', icon: 'success' })
  editing.value = false
  await loadData()
}

function startDeletePress(item: BoardingSpecialItem) {
  cancelDeletePress()
  deletePressTimer = setTimeout(() => {
    deletePressTimer = null
    confirmDelete(item)
  }, 650)
}

function cancelDeletePress() {
  if (!deletePressTimer) return
  clearTimeout(deletePressTimer)
  deletePressTimer = null
}

function confirmDelete(item: BoardingSpecialItem) {
  if (deleteConfirming.value) return
  deleteConfirming.value = true
  cancelDeletePress()
  uni.showModal({
    title: '删除项目',
    content: `确认删除「${item.name}」吗？`,
    confirmText: '删除',
    confirmColor: '#dc2626',
    success: async (res) => {
      if (!res.confirm) return
      await deleteBoardingSpecialItem(item.ID)
      if (form.value.id === item.ID) {
        editing.value = false
      }
      uni.showToast({ title: '已删除', icon: 'success' })
      await loadData()
    },
    complete: () => {
      deleteConfirming.value = false
    },
  })
}

onShow(loadData)
</script>

<style scoped>
.page { padding: 20rpx 24rpx; display: flex; flex-direction: column; gap: 16rpx; }
.header { display: flex; justify-content: space-between; align-items: flex-start; gap: 20rpx; }
.header-copy { flex: 1; min-width: 0; }
.title { display: block; font-size: 34rpx; font-weight: 700; color: #111827; }
.subtitle { display: block; margin-top: 8rpx; font-size: 22rpx; line-height: 1.6; color: #6b7280; }
.card, .empty-card { background: #fff; border-radius: 18rpx; padding: 20rpx 24rpx; box-shadow: 0 12rpx 28rpx rgba(15, 23, 42, 0.04); }
.empty-card { text-align: center; color: #9ca3af; }
.section-title { display: block; margin-bottom: 12rpx; font-size: 28rpx; font-weight: 700; color: #111827; }
.field { margin-bottom: 12rpx; }
.field-label { display: block; margin-bottom: 6rpx; font-size: 22rpx; font-weight: 600; line-height: 1.4; color: #6b7280; }
.input, .picker { width: 100%; box-sizing: border-box; background: #f9fafb; border: 1rpx solid #e5e7eb; border-radius: 12rpx; padding: 0 20rpx; font-size: 26rpx; color: #111827; min-height: 68rpx; display: flex; align-items: center; }
.textarea { width: 100%; box-sizing: border-box; background: #f9fafb; border: 1rpx solid #e5e7eb; border-radius: 12rpx; padding: 16rpx 20rpx; font-size: 26rpx; color: #111827; min-height: 96rpx; }
.input :deep(.uni-input-wrapper) { width: 100%; min-height: 68rpx; display: flex; align-items: center; }
.input :deep(.uni-input-input) { width: 100%; min-height: 40rpx; font-size: 26rpx; line-height: 40rpx; color: #111827; text-align: left !important; }
.input :deep(.uni-input-placeholder) { width: 100%; font-size: 26rpx; color: #9ca3af; text-align: left !important; }
.btn { padding: 14rpx 24rpx; border-radius: 12rpx; background: #f3f4f6; color: #374151; font-size: 24rpx; }
.btn-primary { background: #4f46e5; color: #fff; }
.create-btn {
  flex: 0 0 auto;
  min-width: 136rpx;
  min-height: 68rpx;
  padding: 0 24rpx;
  border-radius: 16rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  box-sizing: border-box;
  white-space: nowrap;
  font-size: 24rpx;
  font-weight: 600;
  line-height: 1;
  box-shadow: 0 10rpx 20rpx rgba(79, 70, 229, 0.22);
}
.actions, .item-head { display: flex; justify-content: space-between; gap: 12rpx; align-items: center; }
.item-title { display: block; font-size: 28rpx; font-weight: 700; color: #111827; }
.item-meta, .item-remark { display: block; margin-top: 8rpx; font-size: 22rpx; color: #6b7280; line-height: 1.6; }
.item-edit { font-size: 24rpx; color: #4f46e5; }
</style>
