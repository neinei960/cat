<template>
  <SideLayout>
  <view class="page">
    <view class="hero">
      <view>
        <text class="hero-title">客户标签管理</text>
        <text class="hero-subtitle">给熟客、回访客、高净值客户建立统一标签库</text>
      </view>
      <view class="hero-btn" @click="openCreate">+ 新增标签</view>
    </view>

    <view v-if="loading" class="state">加载中...</view>
    <view v-else-if="tags.length === 0" class="state">还没有客户标签</view>

    <view v-else class="list">
      <view class="tag-card" v-for="tag in tags" :key="tag.ID">
        <view class="tag-card-main">
          <view class="tag-card-head">
            <view class="tag-badge" :style="{ background: withAlpha(tag.color, 0.14), color: tag.color, borderColor: withAlpha(tag.color, 0.28) }">
              {{ tag.name }}
            </view>
            <text class="tag-status" v-if="tag.status === 0">已停用</text>
          </view>
          <view class="tag-metrics">
            <text class="tag-metric-value">{{ tag.relation_count || 0 }}</text>
            <text class="tag-metric-label">关联客户</text>
          </view>
        </view>
        <text class="tag-description">{{ tag.description || '暂无描述' }}</text>
        <view class="tag-actions">
          <text class="action edit" @click="openEdit(tag)">编辑</text>
          <text class="action stop" @click="toggleStatus(tag)">{{ tag.status === 1 ? '停用' : '启用' }}</text>
          <text class="action delete" @click="removeTag(tag)">删除</text>
        </view>
      </view>
    </view>

    <view class="modal-mask" v-if="showModal" @click="closeModal">
      <view class="modal-body" @click.stop>
        <text class="modal-title">{{ editingId ? '编辑客户标签' : '新增客户标签' }}</text>
        <view class="field">
          <text class="field-label">标签名</text>
          <input v-model="draft.name" class="field-input" placeholder="例如：需要回访" />
        </view>
        <view class="field">
          <text class="field-label">标签描述</text>
          <textarea v-model="draft.description" class="field-textarea" placeholder="写清楚这个标签什么时候用" />
        </view>
        <view class="field">
          <text class="field-label">标签颜色</text>
          <view class="color-list">
            <view
              v-for="color in colorOptions"
              :key="color"
              :class="['color-dot', draft.color === color ? 'color-dot-active' : '']"
              :style="{ background: color }"
              @click="draft.color = color"
            />
          </view>
        </view>
        <view class="modal-actions">
          <view class="modal-btn cancel" @click="closeModal">取消</view>
          <view class="modal-btn confirm" @click="submitTag">{{ submitting ? '保存中...' : '保存' }}</view>
        </view>
      </view>
    </view>
  </view>
  </SideLayout>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { onShow } from '@dcloudio/uni-app'
import SideLayout from '@/components/SideLayout.vue'
import { createCustomerTag, deleteCustomerTag, getCustomerTags, updateCustomerTag } from '@/api/customer-tag'

const loading = ref(true)
const submitting = ref(false)
const showModal = ref(false)
const editingId = ref(0)
const tags = ref<CustomerTag[]>([])
const colorOptions = ['#4F46E5', '#0EA5E9', '#10B981', '#F59E0B', '#EC4899', '#EF4444', '#8B5CF6']
const draft = ref({
  name: '',
  description: '',
  color: '#4F46E5',
  sort_order: 0,
  status: 1,
})

onShow(loadTags)

async function loadTags() {
  loading.value = true
  try {
    const res = await getCustomerTags()
    tags.value = res.data || []
  } finally {
    loading.value = false
  }
}

function openCreate() {
  editingId.value = 0
  draft.value = { name: '', description: '', color: '#4F46E5', sort_order: tags.value.length + 1, status: 1 }
  showModal.value = true
}

function openEdit(tag: CustomerTag) {
  editingId.value = tag.ID
  draft.value = {
    name: tag.name,
    description: tag.description || '',
    color: tag.color || '#4F46E5',
    sort_order: tag.sort_order || 0,
    status: tag.status ?? 1,
  }
  showModal.value = true
}

function closeModal() {
  showModal.value = false
}

async function submitTag() {
  if (!draft.value.name.trim()) {
    uni.showToast({ title: '请填写标签名', icon: 'none' })
    return
  }
  submitting.value = true
  try {
    if (editingId.value) await updateCustomerTag(editingId.value, draft.value)
    else await createCustomerTag(draft.value)
    uni.showToast({ title: '保存成功', icon: 'success' })
    closeModal()
    await loadTags()
  } finally {
    submitting.value = false
  }
}

async function toggleStatus(tag: CustomerTag) {
  await updateCustomerTag(tag.ID, {
    name: tag.name,
    description: tag.description,
    color: tag.color,
    sort_order: tag.sort_order,
    status: tag.status === 1 ? 0 : 1,
  })
  uni.showToast({ title: '已更新', icon: 'success' })
  await loadTags()
}

function removeTag(tag: CustomerTag) {
  uni.showModal({
    title: '删除标签',
    content: `确认删除「${tag.name}」？已关联客户会同步取消标签。`,
    confirmColor: '#EF4444',
    success: async (res) => {
      if (!res.confirm) return
      await deleteCustomerTag(tag.ID)
      uni.showToast({ title: '已删除', icon: 'success' })
      await loadTags()
    },
  })
}

function withAlpha(color: string, alpha: number) {
  const hex = color.replace('#', '')
  if (hex.length !== 6) return color
  const value = Math.round(alpha * 255).toString(16).padStart(2, '0')
  return `#${hex}${value}`
}
</script>

<style scoped>
.page { padding: 24rpx; }
.hero { background: linear-gradient(135deg, #EEF2FF, #F8FAFC); border-radius: 20rpx; padding: 24rpx; display: flex; justify-content: space-between; gap: 20rpx; align-items: center; margin-bottom: 20rpx; }
.hero-title { font-size: 34rpx; font-weight: 700; color: #111827; display: block; }
.hero-subtitle { font-size: 24rpx; color: #6B7280; display: block; margin-top: 8rpx; line-height: 1.5; }
.hero-btn { background: #4F46E5; color: #fff; padding: 14rpx 20rpx; border-radius: 999rpx; font-size: 24rpx; white-space: nowrap; }
.state { text-align: center; color: #9CA3AF; padding: 120rpx 0; font-size: 28rpx; }
.list { display: flex; flex-direction: column; gap: 16rpx; }
.tag-card { background: #fff; border-radius: 18rpx; padding: 22rpx 24rpx; box-shadow: 0 4rpx 16rpx rgba(15, 23, 42, 0.05); }
.tag-card-main { display: flex; align-items: flex-start; justify-content: space-between; gap: 16rpx; }
.tag-card-head { display: flex; align-items: center; gap: 16rpx; min-width: 0; }
.tag-badge { display: inline-flex; align-items: center; justify-content: center; padding: 10rpx 18rpx; border-radius: 999rpx; font-size: 24rpx; font-weight: 600; border: 1rpx solid transparent; }
.tag-status { font-size: 22rpx; color: #9CA3AF; }
.tag-metrics { flex-shrink: 0; text-align: right; }
.tag-metric-value { display: block; font-size: 30rpx; font-weight: 700; color: #111827; }
.tag-metric-label { display: block; font-size: 20rpx; color: #9CA3AF; margin-top: 4rpx; }
.tag-description { display: block; margin-top: 14rpx; font-size: 24rpx; color: #6B7280; line-height: 1.6; }
.tag-actions { display: flex; gap: 20rpx; margin-top: 18rpx; }
.action { font-size: 24rpx; }
.action.edit { color: #4F46E5; }
.action.stop { color: #D97706; }
.action.delete { color: #EF4444; }
.modal-mask { position: fixed; inset: 0; background: rgba(15, 23, 42, 0.35); display: flex; align-items: center; justify-content: center; padding: 24rpx; z-index: 99; }
.modal-body { width: 100%; background: #fff; border-radius: 20rpx; padding: 28rpx 24rpx; }
.modal-title { font-size: 30rpx; font-weight: 700; color: #111827; display: block; text-align: center; margin-bottom: 24rpx; }
.field { margin-bottom: 20rpx; }
.field-label { font-size: 24rpx; color: #374151; display: block; margin-bottom: 12rpx; }
.field-input, .field-textarea { width: 100%; background: #F9FAFB; border-radius: 12rpx; font-size: 28rpx; color: #111827; padding: 0 18rpx; }
.field-input { height: 76rpx; }
.field-textarea { height: 180rpx; padding-top: 16rpx; }
.color-list { display: flex; flex-wrap: wrap; gap: 16rpx; }
.color-dot { width: 56rpx; height: 56rpx; border-radius: 50%; border: 4rpx solid transparent; }
.color-dot-active { border-color: #111827; }
.modal-actions { display: flex; gap: 16rpx; margin-top: 28rpx; }
.modal-btn { flex: 1; text-align: center; padding: 18rpx 0; border-radius: 14rpx; font-size: 28rpx; }
.modal-btn.cancel { background: #F3F4F6; color: #6B7280; }
.modal-btn.confirm { background: #4F46E5; color: #fff; }
</style>
