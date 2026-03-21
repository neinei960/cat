<template>
  <SideLayout>
  <view class="page">
    <view class="header">
      <text class="title">毛发类别管理</text>
      <text class="desc">管理猫咪毛发等级分类，用于洗浴定价</text>
    </view>

    <!-- 现有类别列表 -->
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
          <input v-model="editName" class="edit-input" placeholder="类别名称" />
          <input v-model="editSort" type="number" class="edit-input sort" placeholder="排序" />
          <text class="action-btn save" @click="onSaveEdit(cat.ID)">保存</text>
          <text class="action-btn cancel" @click="editingId = 0">取消</text>
        </view>
      </view>
      <view v-if="categories.length === 0" class="empty">暂无毛发类别，请添加</view>
    </view>

    <!-- 新增 -->
    <view class="add-section">
      <text class="section-title">添加新类别</text>
      <view class="add-form">
        <input v-model="newName" class="add-input" placeholder="类别名称（如 E类）" />
        <input v-model="newSort" type="number" class="add-input sort" placeholder="排序号" />
        <button class="add-btn" @click="onCreate" :loading="creating" size="mini">添加</button>
      </view>
    </view>

    <view class="tips">
      <text class="tip-title">说明</text>
      <text class="tip-text">1. 毛发类别用于猫咪建档时选择，也是洗浴定价的关键维度</text>
      <text class="tip-text">2. 添加新类别后，需到「服务管理」中为对应服务配置该类别的定价规则</text>
      <text class="tip-text">3. 删除类别不会影响已有猫咪的毛发等级记录</text>
    </view>
  </view>
  </SideLayout>
</template>

<script setup lang="ts">
import SideLayout from '@/components/SideLayout.vue'
import { ref, onMounted } from 'vue'
import { getFurCategoryList, createFurCategory, updateFurCategory, deleteFurCategory } from '@/api/fur-category'
import type { FurCategory } from '@/api/fur-category'

const categories = ref<FurCategory[]>([])
const newName = ref('')
const newSort = ref('')
const creating = ref(false)
const editingId = ref(0)
const editName = ref('')
const editSort = ref('')

onMounted(loadList)

async function loadList() {
  const res = await getFurCategoryList()
  categories.value = Array.isArray(res.data) ? res.data : []
}

async function onCreate() {
  if (!newName.value.trim()) {
    uni.showToast({ title: '请输入类别名称', icon: 'none' })
    return
  }
  creating.value = true
  try {
    await createFurCategory({
      name: newName.value.trim(),
      sort_order: parseInt(newSort.value) || categories.value.length + 1,
    })
    newName.value = ''
    newSort.value = ''
    uni.showToast({ title: '添加成功', icon: 'success' })
    await loadList()
  } finally {
    creating.value = false
  }
}

function startEdit(cat: FurCategory) {
  editingId.value = cat.ID
  editName.value = cat.name
  editSort.value = String(cat.sort_order)
}

async function onSaveEdit(id: number) {
  if (!editName.value.trim()) {
    uni.showToast({ title: '名称不能为空', icon: 'none' })
    return
  }
  await updateFurCategory(id, {
    name: editName.value.trim(),
    sort_order: parseInt(editSort.value) || 0,
  })
  editingId.value = 0
  uni.showToast({ title: '已保存', icon: 'success' })
  await loadList()
}

function onDelete(cat: FurCategory) {
  uni.showModal({
    title: '确认删除',
    content: `确认删除毛发类别「${cat.name}」？已有猫咪的毛发等级不受影响。`,
    success: async (res) => {
      if (res.confirm) {
        await deleteFurCategory(cat.ID)
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
.desc { font-size: 24rpx; color: #6B7280; display: block; margin-top: 8rpx; }
.list { background: #fff; border-radius: 16rpx; padding: 8rpx 24rpx; margin-bottom: 24rpx; }
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
.edit-input.sort { width: 100rpx; flex: none; }
.empty { text-align: center; padding: 40rpx; color: #9CA3AF; font-size: 26rpx; }
.add-section { background: #fff; border-radius: 16rpx; padding: 24rpx; margin-bottom: 24rpx; }
.section-title { font-size: 28rpx; font-weight: 600; color: #1F2937; display: block; margin-bottom: 16rpx; }
.add-form { display: flex; align-items: center; gap: 12rpx; }
.add-input { background: #F3F4F6; border-radius: 8rpx; padding: 16rpx; font-size: 28rpx; flex: 1; }
.add-input.sort { width: 120rpx; flex: none; }
.add-btn { background: #4F46E5 !important; color: #fff !important; border-radius: 8rpx !important; font-size: 26rpx !important; }
.tips { background: #FFFBEB; border-radius: 16rpx; padding: 24rpx; }
.tip-title { font-size: 26rpx; font-weight: 600; color: #92400E; display: block; margin-bottom: 12rpx; }
.tip-text { font-size: 24rpx; color: #78716C; display: block; line-height: 1.6; }
</style>
