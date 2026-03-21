<template>
  <SideLayout>
    <view class="page">
      <view class="page-header">
        <text class="page-title">服务分类管理</text>
        <button class="btn-primary-sm" @click="addCategory(null)">+ 新建一级分类</button>
      </view>

      <view v-if="loading" class="loading">加载中...</view>

      <view v-else class="category-tree">
        <view v-for="cat in categories" :key="cat.ID" class="cat-group">
          <!-- 一级分类 -->
          <view class="cat-item level-1">
            <view class="cat-info" @click="toggleExpand(cat.ID)">
              <text class="expand-icon">{{ expanded[cat.ID] ? '▼' : '▶' }}</text>
              <text class="cat-name">{{ cat.name }}</text>
              <text class="cat-badge" v-if="cat.children?.length">{{ cat.children.length }}个子分类</text>
            </view>
            <view class="cat-actions">
              <text class="action-btn" @click="addCategory(cat.ID)">+子分类</text>
              <text class="action-btn" @click="editCategory(cat)">编辑</text>
              <text class="action-btn danger" @click="removeCategory(cat)">删除</text>
            </view>
          </view>

          <!-- 二级分类 -->
          <view v-if="expanded[cat.ID]" class="children">
            <view v-for="child in cat.children" :key="child.ID" class="cat-item level-2">
              <view class="cat-info">
                <text class="cat-name">{{ child.name }}</text>
              </view>
              <view class="cat-actions">
                <text class="action-btn" @click="editCategory(child)">编辑</text>
                <text class="action-btn danger" @click="removeCategory(child)">删除</text>
              </view>
            </view>
            <view v-if="!cat.children?.length" class="empty-children">
              <text class="empty-text">暂无子分类</text>
            </view>
          </view>
        </view>

        <view v-if="!loading && categories.length === 0" class="empty">
          <text>暂无分类，点击上方按钮创建</text>
        </view>
      </view>
    </view>
  </SideLayout>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { onShow } from '@dcloudio/uni-app'
import SideLayout from '@/components/SideLayout.vue'
import { getCategoryTree, createCategory, updateCategory, deleteCategory } from '@/api/service-category'

const loading = ref(false)
const categories = ref<ServiceCategory[]>([])
const expanded = reactive<Record<number, boolean>>({})

onShow(async () => {
  await loadTree()
})

async function loadTree() {
  loading.value = true
  try {
    const res = await getCategoryTree()
    categories.value = res.data || []
    // Auto-expand all
    categories.value.forEach(c => { expanded[c.ID] = true })
  } finally {
    loading.value = false
  }
}

function toggleExpand(id: number) {
  expanded[id] = !expanded[id]
}

function addCategory(parentId: number | null) {
  uni.showModal({
    title: parentId ? '新建二级分类' : '新建一级分类',
    editable: true,
    placeholderText: '输入分类名称',
    success: async (res) => {
      if (res.confirm && res.content?.trim()) {
        try {
          await createCategory({
            name: res.content.trim(),
            parent_id: parentId || undefined,
          })
          uni.showToast({ title: '创建成功', icon: 'success' })
          await loadTree()
        } catch (e: any) {
          uni.showToast({ title: e.message || '创建失败', icon: 'none' })
        }
      }
    }
  })
}

function editCategory(cat: ServiceCategory) {
  uni.showModal({
    title: '编辑分类',
    editable: true,
    placeholderText: '分类名称',
    content: cat.name,
    success: async (res) => {
      if (res.confirm && res.content?.trim()) {
        try {
          await updateCategory(cat.ID, { name: res.content.trim() })
          uni.showToast({ title: '修改成功', icon: 'success' })
          await loadTree()
        } catch (e: any) {
          uni.showToast({ title: e.message || '修改失败', icon: 'none' })
        }
      }
    }
  })
}

function removeCategory(cat: ServiceCategory) {
  uni.showModal({
    title: '确认删除',
    content: `确定要删除分类"${cat.name}"吗？`,
    success: async (res) => {
      if (res.confirm) {
        try {
          await deleteCategory(cat.ID)
          uni.showToast({ title: '删除成功', icon: 'success' })
          await loadTree()
        } catch (e: any) {
          uni.showToast({ title: e.message || '删除失败', icon: 'none' })
        }
      }
    }
  })
}
</script>

<style scoped>
.page { padding: 24rpx; max-width: 800px; }
.page-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 32rpx; }
.page-title { font-size: 36rpx; font-weight: 700; color: #1F2937; }
.btn-primary-sm { background: #4F46E5; color: #fff; font-size: 26rpx; padding: 12rpx 28rpx; border-radius: 12rpx; }
.loading { text-align: center; padding: 60rpx; color: #9CA3AF; }
.category-tree { display: flex; flex-direction: column; gap: 16rpx; }
.cat-group { background: #fff; border-radius: 16rpx; overflow: hidden; box-shadow: 0 2rpx 12rpx rgba(0,0,0,0.04); }
.cat-item { display: flex; justify-content: space-between; align-items: center; padding: 20rpx 24rpx; }
.level-1 { background: #F9FAFB; border-bottom: 2rpx solid #F3F4F6; }
.level-2 { border-bottom: 1rpx solid #F3F4F6; padding-left: 56rpx; }
.level-2:last-child { border-bottom: none; }
.cat-info { display: flex; align-items: center; gap: 12rpx; flex: 1; }
.expand-icon { font-size: 20rpx; color: #9CA3AF; width: 24rpx; }
.cat-name { font-size: 28rpx; color: #1F2937; font-weight: 500; }
.cat-badge { font-size: 22rpx; color: #6B7280; background: #E5E7EB; padding: 2rpx 12rpx; border-radius: 8rpx; }
.cat-actions { display: flex; gap: 16rpx; }
.action-btn { font-size: 24rpx; color: #4F46E5; }
.action-btn.danger { color: #EF4444; }
.children { }
.empty-children { padding: 20rpx 56rpx; }
.empty-text { font-size: 24rpx; color: #9CA3AF; }
.empty { text-align: center; padding: 60rpx; color: #9CA3AF; font-size: 28rpx; }
</style>
