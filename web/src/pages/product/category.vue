<template>
  <SideLayout>
  <view class="page">
    <view class="header">
      <text class="title">商品分类管理</text>
    </view>

    <view class="list">
      <view class="item" v-for="(cat, idx) in categories" :key="cat.ID">
        <view class="item-main" v-if="editingId !== cat.ID">
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
import { ref, onMounted } from 'vue'
import { getProductCategories, createProductCategory, updateProductCategory, deleteProductCategory } from '@/api/product'

const categories = ref<any[]>([])
const loading = ref(true)
const newName = ref('')
const creating = ref(false)
const editingId = ref(0)
const editName = ref('')

onMounted(loadList)

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
  await updateProductCategory(id, { name: editName.value.trim() })
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
</script>

<style scoped>
.page { padding: 24rpx; }
.header { margin-bottom: 24rpx; }
.title { font-size: 36rpx; font-weight: bold; color: #1F2937; display: block; }
.list { background: #fff; border-radius: 16rpx; padding: 8rpx 24rpx; margin-bottom: 24rpx; box-shadow: 0 2rpx 8rpx rgba(0,0,0,0.04); }
.item { padding: 20rpx 0; border-bottom: 1rpx solid #F3F4F6; }
.item:last-child { border-bottom: none; }
.item-main { display: flex; align-items: center; }
.item-order { width: 48rpx; font-size: 26rpx; color: #9CA3AF; font-weight: 600; }
.item-name { flex: 1; font-size: 30rpx; color: #1F2937; font-weight: 500; }
.item-actions { display: flex; gap: 16rpx; }
.action-btn { font-size: 24rpx; padding: 6rpx 16rpx; border-radius: 8rpx; }
.action-btn.edit { color: #4F46E5; background: #EEF2FF; }
.action-btn.delete { color: #DC2626; background: #FEE2E2; }
.action-btn.save { color: #059669; background: #D1FAE5; }
.action-btn.cancel { color: #6B7280; background: #F3F4F6; }
.item-edit { display: flex; align-items: center; gap: 12rpx; }
.edit-input { background: #F3F4F6; border-radius: 8rpx; padding: 12rpx; font-size: 26rpx; flex: 1; }
.empty { text-align: center; padding: 40rpx; color: #9CA3AF; font-size: 26rpx; }
.add-section { background: #fff; border-radius: 16rpx; padding: 24rpx; margin-bottom: 24rpx; box-shadow: 0 2rpx 8rpx rgba(0,0,0,0.04); }
.section-title { font-size: 28rpx; font-weight: 600; color: #1F2937; display: block; margin-bottom: 16rpx; }
.add-form { display: flex; align-items: center; gap: 12rpx; }
.add-input { background: #F3F4F6; border-radius: 8rpx; padding: 16rpx; font-size: 28rpx; flex: 1; }
.add-btn { background: #4F46E5 !important; color: #fff !important; border-radius: 8rpx !important; font-size: 26rpx !important; }
</style>
