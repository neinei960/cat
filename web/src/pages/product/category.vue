<template>
  <SideLayout>
  <view class="page">
    <view class="header">
      <text class="title">商品分类管理</text>
      <text class="subtitle">拖拽左侧手柄可调整分类顺序，松手后自动保存</text>
    </view>

    <view class="list">
      <view
        class="item"
        :class="{
          dragging: draggingId === cat.ID,
          'drag-over': dragOverId === cat.ID && draggingId !== cat.ID,
          saving: savingOrder,
        }"
        v-for="(cat, idx) in categories"
        :key="cat.ID"
        :data-category-id="cat.ID"
        :draggable="editingId !== cat.ID && !savingOrder"
        @dragstart="onDragStart(cat.ID)"
        @dragover.prevent="onDragOver(cat.ID)"
        @dragleave="onDragLeave(cat.ID)"
        @drop.prevent="onDrop(cat.ID)"
        @dragend="onDragEnd"
      >
        <view class="item-main" v-if="editingId !== cat.ID">
          <view
            class="drag-handle"
            @touchstart.stop.prevent="beginTouchDrag(idx, $event)"
            @mousedown.stop
          >
            <text class="drag-icon">⋮⋮</text>
          </view>
          <text class="item-order">{{ idx + 1 }}</text>
          <text class="item-name">{{ cat.name }}</text>
          <view class="item-actions">
            <text class="action-btn edit" @click="startEdit(cat)">编辑</text>
            <text class="action-btn delete" @click="onDelete(cat)">删除</text>
          </view>
        </view>
        <view class="item-edit" v-else>
          <input v-model="editName" class="edit-input" placeholder="分类名称" />
          <text class="action-btn save" @click="onSaveEdit(cat.ID)">保存</text>
          <text class="action-btn cancel" @click="editingId = 0">取消</text>
        </view>
      </view>
      <view v-if="categories.length === 0 && !loading" class="empty">暂无商品分类，请添加</view>
      <view v-if="loading" class="empty">加载中...</view>
    </view>

    <view class="add-section">
      <text class="section-title">添加新分类</text>
      <view class="add-form">
        <input v-model="newName" class="add-input" placeholder="分类名称" />
        <button class="add-btn" @click="onCreate" :loading="creating" size="mini">添加</button>
      </view>
    </view>
  </view>
  </SideLayout>
</template>

<script setup lang="ts">
import SideLayout from '@/components/SideLayout.vue'
import { ref, onMounted, onBeforeUnmount } from 'vue'
import { getProductCategories, createProductCategory, updateProductCategory, deleteProductCategory } from '@/api/product'

const categories = ref<any[]>([])
const loading = ref(true)
const newName = ref('')
const creating = ref(false)
const editingId = ref(0)
const editName = ref('')
const draggingId = ref(0)
const dragOverId = ref(0)
const savingOrder = ref(false)
let dragSnapshot: any[] = []
let dragMoved = false

onMounted(loadList)
onBeforeUnmount(removeTouchDragListeners)

async function loadList() {
  loading.value = true
  try {
    const res = await getProductCategories()
    categories.value = Array.isArray(res.data) ? res.data : []
  } finally { loading.value = false }
}

async function onCreate() {
  if (!newName.value.trim()) {
    uni.showToast({ title: '请输入分类名称', icon: 'none' }); return
  }
  creating.value = true
  try {
    await createProductCategory({
      name: newName.value.trim(),
      sort_order: categories.value.length + 1,
    })
    newName.value = ''
    uni.showToast({ title: '添加成功', icon: 'success' })
    await loadList()
  } finally { creating.value = false }
}

function startEdit(cat: any) {
  editingId.value = cat.ID
  editName.value = cat.name
}

async function onSaveEdit(id: number) {
  if (!editName.value.trim()) {
    uni.showToast({ title: '名称不能为空', icon: 'none' }); return
  }
  const current = categories.value.find(cat => cat.ID === id)
  await updateProductCategory(id, {
    name: editName.value.trim(),
    sort_order: current?.sort_order ?? categories.value.findIndex(cat => cat.ID === id) + 1,
    status: current?.status,
  })
  editingId.value = 0
  uni.showToast({ title: '已保存', icon: 'success' })
  await loadList()
}

function onDelete(cat: any) {
  uni.showModal({
    title: '确认删除',
    content: `确认删除分类「${cat.name}」？`,
    success: async (res) => {
      if (res.confirm) {
        await deleteProductCategory(cat.ID)
        uni.showToast({ title: '已删除', icon: 'success' })
        await loadList()
      }
    }
  })
}

function onDragStart(id: number) {
  draggingId.value = id
  dragOverId.value = id
}

function onDragOver(id: number) {
  if (!draggingId.value || draggingId.value === id) return
  dragOverId.value = id
}

function onDragLeave(id: number) {
  if (dragOverId.value === id) {
    dragOverId.value = 0
  }
}

async function onDrop(targetId: number) {
  const sourceId = draggingId.value
  resetDragState()
  if (!sourceId || sourceId === targetId || savingOrder.value) return

  const sourceIndex = categories.value.findIndex(cat => cat.ID === sourceId)
  const targetIndex = categories.value.findIndex(cat => cat.ID === targetId)
  if (sourceIndex < 0 || targetIndex < 0) return

  const nextList = [...categories.value]
  const [moved] = nextList.splice(sourceIndex, 1)
  nextList.splice(targetIndex, 0, moved)
  await persistOrder(nextList, categories.value)
}

function onDragEnd() {
  resetDragState()
}

function resetDragState() {
  draggingId.value = 0
  dragOverId.value = 0
}

function getEventPoint(event: any) {
  const touch = event?.touches?.[0] || event?.changedTouches?.[0]
  if (touch) return { x: touch.clientX, y: touch.clientY }
  if (typeof event?.clientX === 'number' && typeof event?.clientY === 'number') {
    return { x: event.clientX, y: event.clientY }
  }
  return null
}

function removeTouchDragListeners() {
  if (typeof window === 'undefined') return
  window.removeEventListener('touchmove', handleTouchDragMove as EventListener)
  window.removeEventListener('touchend', handleTouchDragEnd as EventListener)
  window.removeEventListener('touchcancel', handleTouchDragEnd as EventListener)
  document.body.style.userSelect = ''
}

function beginTouchDrag(index: number, event: any) {
  if (typeof window === 'undefined' || savingOrder.value || editingId.value !== 0) return
  const item = categories.value[index]
  if (!item || categories.value.length < 2) return
  draggingId.value = item.ID
  dragOverId.value = item.ID
  dragSnapshot = [...categories.value]
  dragMoved = false
  document.body.style.userSelect = 'none'
  window.addEventListener('touchmove', handleTouchDragMove as EventListener, { passive: false })
  window.addEventListener('touchend', handleTouchDragEnd as EventListener)
  window.addEventListener('touchcancel', handleTouchDragEnd as EventListener)
  handleTouchDragMove(event)
}

function handleTouchDragMove(event: Event) {
  if (!draggingId.value || typeof document === 'undefined') return
  const point = getEventPoint(event)
  if (!point) return
  if ('preventDefault' in event) event.preventDefault()
  const element = document.elementFromPoint(point.x, point.y) as HTMLElement | null
  const item = element?.closest('.item') as HTMLElement | null
  const targetId = Number(item?.dataset?.categoryId || 0)
  if (!targetId || targetId === draggingId.value) return

  const fromIndex = categories.value.findIndex(cat => cat.ID === draggingId.value)
  const targetIndex = categories.value.findIndex(cat => cat.ID === targetId)
  if (fromIndex < 0 || targetIndex < 0 || fromIndex === targetIndex) return

  const nextList = [...categories.value]
  const [moved] = nextList.splice(fromIndex, 1)
  nextList.splice(targetIndex, 0, moved)
  categories.value = nextList
  dragOverId.value = targetId
  dragMoved = true
}

async function handleTouchDragEnd() {
  const activeId = draggingId.value
  removeTouchDragListeners()
  resetDragState()
  if (!activeId || !dragMoved) return
  await persistOrder(categories.value, dragSnapshot)
  dragSnapshot = []
  dragMoved = false
}

async function persistOrder(list: any[], previousList: any[]) {
  const normalized = list.map((cat, index) => ({
    ...cat,
    sort_order: index + 1,
  }))
  categories.value = normalized

  savingOrder.value = true
  try {
    await Promise.all(normalized.map((cat, index) =>
      updateProductCategory(cat.ID, {
        name: cat.name,
        sort_order: index + 1,
        status: cat.status,
      })
    ))
    uni.showToast({ title: '排序已保存', icon: 'success' })
  } catch {
    categories.value = previousList
    uni.showToast({ title: '排序保存失败', icon: 'none' })
    await loadList()
  } finally {
    savingOrder.value = false
  }
}
</script>

<style scoped>
.page { padding: 24rpx; }
.header { margin-bottom: 24rpx; }
.title { font-size: 36rpx; font-weight: bold; color: #1F2937; display: block; }
.subtitle { display: block; margin-top: 8rpx; font-size: 24rpx; color: #6B7280; }
.list { background: #fff; border-radius: 16rpx; padding: 8rpx 24rpx; margin-bottom: 24rpx; box-shadow: 0 2rpx 8rpx rgba(0,0,0,0.04); }
.item { padding: 20rpx 0; border-bottom: 1rpx solid #F3F4F6; transition: background 0.2s ease, transform 0.2s ease; }
.item:last-child { border-bottom: none; }
.item.dragging { opacity: 0.55; }
.item.drag-over { background: #EEF2FF; }
.item.saving { pointer-events: none; }
.item-main { display: flex; align-items: center; }
.drag-handle { width: 44rpx; display: flex; justify-content: center; align-items: center; margin-right: 8rpx; cursor: grab; }
.drag-icon { color: #C4B5FD; font-size: 30rpx; font-weight: 700; letter-spacing: -4rpx; }
.item-order { width: 48rpx; font-size: 26rpx; color: #9CA3AF; font-weight: 600; }
.item-name { flex: 1; font-size: 30rpx; color: #1F2937; font-weight: 500; }
.item-actions { display: flex; gap: 16rpx; }
.action-btn { font-size: 24rpx; padding: 6rpx 16rpx; border-radius: 8rpx; }
.action-btn.edit { color: #4F46E5; background: #EEF2FF; }
.action-btn.delete { color: #DC2626; background: #FEE2E2; }
.action-btn.save { color: #059669; background: #D1FAE5; }
.action-btn.cancel { color: #6B7280; background: #F3F4F6; }
.item-edit { display: flex; align-items: center; gap: 12rpx; }
.edit-input,
.add-input {
  flex: 1;
  min-height: 76rpx;
  background: #F3F4F6;
  border-radius: 12rpx;
  padding: 0 20rpx;
  box-sizing: border-box;
  display: flex;
  align-items: center;
}
.edit-input :deep(.uni-input-wrapper),
.add-input :deep(.uni-input-wrapper) {
  width: 100%;
  min-height: 76rpx;
  display: flex;
  align-items: center;
}
.edit-input :deep(.uni-input-input),
.add-input :deep(.uni-input-input) {
  width: 100%;
  min-height: 40rpx;
  font-size: 28rpx;
  line-height: 40rpx;
  color: #111827;
  text-align: left !important;
}
.edit-input :deep(.uni-input-placeholder),
.add-input :deep(.uni-input-placeholder) {
  width: 100%;
  font-size: 28rpx;
  color: #9CA3AF;
  text-align: left !important;
}
.empty { text-align: center; padding: 40rpx; color: #9CA3AF; font-size: 26rpx; }
.add-section { background: #fff; border-radius: 16rpx; padding: 24rpx; margin-bottom: 24rpx; box-shadow: 0 2rpx 8rpx rgba(0,0,0,0.04); }
.section-title { font-size: 28rpx; font-weight: 600; color: #1F2937; display: block; margin-bottom: 16rpx; }
.add-form { display: flex; align-items: center; gap: 12rpx; }
.add-btn { background: #4F46E5 !important; color: #fff !important; border-radius: 8rpx !important; font-size: 26rpx !important; }
</style>
