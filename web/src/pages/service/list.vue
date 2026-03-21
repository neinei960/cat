<template>
  <SideLayout>
  <view class="page">
    <view class="header">
      <text class="title">服务管理</text>
      <view class="header-actions">
        <view class="btn-secondary" @click="goCategory">分类管理</view>
        <view class="btn-add" @click="goAdd">+ 新增服务</view>
      </view>
    </view>

    <view class="main-layout">
      <!-- 分类侧栏（桌面端始终展示，手机端可折叠） -->
      <view class="category-sidebar-wrap">
        <!-- 手机端折叠触发条 -->
        <view class="mobile-cat-toggle" @click="showCategoryPanel = !showCategoryPanel">
          <view class="mobile-cat-toggle-left">
            <text class="mobile-cat-toggle-icon">🏷</text>
            <text class="mobile-cat-toggle-label">分类筛选</text>
            <text v-if="activeCatId !== 0" class="mobile-cat-badge">已选</text>
          </view>
          <text :class="['mobile-cat-arrow', showCategoryPanel ? 'open' : '']">▼</text>
        </view>

        <!-- 分类列表主体 -->
        <view :class="['category-sidebar', showCategoryPanel ? 'panel-open' : 'panel-closed']">
          <!-- 全部 / 未分类 -->
          <view
            :class="['cat-option', !activeCatId ? 'active' : '']"
            @click="activeCatId = 0; showCategoryPanel = false"
          >
            <view class="cat-active-bar"></view>
            <text class="cat-label">全部分类</text>
          </view>
          <view
            :class="['cat-option', activeCatId === -1 ? 'active' : '']"
            @click="activeCatId = -1; showCategoryPanel = false"
          >
            <view class="cat-active-bar"></view>
            <text class="cat-label">未分类</text>
          </view>

          <!-- 一级分类 + 二级分类 -->
          <view v-for="cat in categories" :key="cat.ID" class="cat-section">
            <!-- 一级分类行：左侧点击筛选，右侧箭头展开收起 -->
            <view :class="['cat-option parent', activeCatId === cat.ID ? 'active' : '']">
              <view class="cat-active-bar"></view>
              <text class="cat-label" @click="activeCatId = cat.ID; showCategoryPanel = false">{{ cat.name }}</text>
              <text
                v-if="cat.children?.length"
                :class="['cat-expand-arrow', expandedCats.has(cat.ID) ? 'open' : '']"
                @click.stop="toggleCat(cat.ID)"
              >▶</text>
            </view>

            <!-- 二级分类，受展开状态控制 -->
            <view
              v-if="cat.children?.length && expandedCats.has(cat.ID)"
              class="cat-children"
            >
              <view
                v-for="child in cat.children" :key="child.ID"
                :class="['cat-option child', activeCatId === child.ID ? 'active' : '']"
                @click="activeCatId = child.ID; showCategoryPanel = false"
              >
                <view class="cat-active-bar"></view>
                <text class="cat-connector">└</text>
                <text class="cat-label">{{ child.name }}</text>
              </view>
            </view>
          </view>
        </view>
      </view>

      <!-- 服务列表 -->
      <view class="service-list">
        <view v-if="loading" class="loading">加载中...</view>
        <view v-else-if="filteredList.length === 0" class="empty">暂无服务</view>
        <view v-else class="list">
          <view class="card" v-for="item in filteredList" :key="item.ID" @click="goEdit(item.ID)">
            <view class="card-top">
              <view class="svc-info">
                <text class="name">{{ item.name }}</text>
                <text class="category-tag">{{ getCategoryName(item) }}</text>
              </view>
              <view :class="['status', item.status === 1 ? 'active' : 'inactive']">
                {{ item.status === 1 ? '启用' : '停用' }}
              </view>
            </view>
            <view class="card-meta">
              <text class="price">¥{{ item.base_price }}</text>
              <text class="duration">{{ item.duration }}分钟</text>
            </view>
            <view class="desc" v-if="item.description">{{ item.description }}</view>
          </view>
        </view>
      </view>
    </view>
  </view>
  </SideLayout>
</template>

<script setup lang="ts">
import SideLayout from '@/components/SideLayout.vue'
import { ref, computed } from 'vue'
import { onShow } from '@dcloudio/uni-app'
import { getServiceList } from '@/api/service'
import { getCategoryTree } from '@/api/service-category'

const list = ref<ServiceItem[]>([])
const categories = ref<ServiceCategory[]>([])
const loading = ref(true)
const activeCatId = ref(0) // 0=全部, -1=未分类, >0=具体分类ID

// 手机端分类面板折叠状态（默认收起，节省空间）
const showCategoryPanel = ref(false)
// 记录哪些一级分类处于展开状态（Set 存 cat.ID）
const expandedCats = ref<Set<number>>(new Set())

function toggleCat(catId: number) {
  if (expandedCats.value.has(catId)) {
    expandedCats.value.delete(catId)
  } else {
    expandedCats.value.add(catId)
  }
  // 触发响应式更新
  expandedCats.value = new Set(expandedCats.value)
}

// Build a flat set of all category IDs under a parent
function getCategoryIds(catId: number): number[] {
  if (catId <= 0) return []
  const ids = [catId]
  for (const cat of categories.value) {
    if (cat.ID === catId && cat.children) {
      cat.children.forEach(c => ids.push(c.ID))
    }
  }
  return ids
}

const filteredList = computed(() => {
  if (activeCatId.value === 0) return list.value
  if (activeCatId.value === -1) return list.value.filter(s => !s.category_id)
  const ids = getCategoryIds(activeCatId.value)
  if (ids.length > 0) {
    return list.value.filter(s => s.category_id && ids.includes(s.category_id))
  }
  // It's a child category
  return list.value.filter(s => s.category_id === activeCatId.value)
})

function getCategoryName(item: ServiceItem): string {
  if (!item.category_id) return item.category || '未分类'
  // Search in tree
  for (const cat of categories.value) {
    if (cat.ID === item.category_id) return cat.name
    if (cat.children) {
      for (const child of cat.children) {
        if (child.ID === item.category_id) return `${cat.name} / ${child.name}`
      }
    }
  }
  return item.category || '未分类'
}

async function loadData() {
  loading.value = true
  try {
    const [sRes, cRes] = await Promise.all([
      getServiceList({ page: 1, page_size: 200 }),
      getCategoryTree(),
    ])
    list.value = sRes.data.list || []
    categories.value = cRes.data || []
  } finally {
    loading.value = false
  }
}

function goAdd() { uni.navigateTo({ url: '/pages/service/edit' }) }
function goEdit(id: number) { uni.navigateTo({ url: `/pages/service/edit?id=${id}` }) }
function goCategory() { uni.navigateTo({ url: '/pages/service/category' }) }

onShow(loadData)
</script>

<style scoped>
/* ===================== 基础 ===================== */
.page { padding: 24rpx; }
.header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 20rpx; }
.title { font-size: 36rpx; font-weight: bold; color: #1F2937; }
.header-actions { display: flex; gap: 16rpx; }
.btn-secondary {
  font-size: 26rpx; color: #4F46E5; border: 2rpx solid #C7D2FE;
  background: #EEF2FF; padding: 12rpx 24rpx; border-radius: 12rpx;
  transition: opacity 0.15s;
}
.btn-secondary:active { opacity: 0.7; }
.btn-add {
  font-size: 28rpx; color: #fff;
  background: linear-gradient(135deg, #6366F1, #4F46E5);
  padding: 12rpx 24rpx; border-radius: 12rpx;
  transition: opacity 0.15s;
}
.btn-add:active { opacity: 0.8; }

/* ===================== 主布局 ===================== */
.main-layout { display: flex; gap: 24rpx; align-items: flex-start; }

/* ===================== 分类侧栏包裹层 ===================== */
.category-sidebar-wrap {
  width: 240rpx;
  min-width: 240rpx;
  flex-shrink: 0;
}

/* 手机端折叠触发条：桌面端隐藏 */
.mobile-cat-toggle { display: none; }

/* ===================== 分类侧栏主体（桌面端） ===================== */
.category-sidebar {
  background: #fff;
  border-radius: 16rpx;
  padding: 12rpx 0;
  box-shadow: 0 2rpx 12rpx rgba(0, 0, 0, 0.06);
  max-height: calc(100vh - 200rpx);
  overflow-y: auto;
}

/* ===================== 分类选项通用 ===================== */
.cat-option {
  display: flex;
  align-items: center;
  padding: 16rpx 20rpx 16rpx 0;
  font-size: 26rpx;
  color: #6B7280;
  cursor: pointer;
  position: relative;
  transition: background 0.15s;
}
.cat-option:active { background: #F9FAFB; }

/* 左侧高亮色条（选中时显示） */
.cat-active-bar {
  width: 6rpx;
  height: 36rpx;
  border-radius: 0 4rpx 4rpx 0;
  background: transparent;
  margin-right: 18rpx;
  flex-shrink: 0;
  transition: background 0.2s;
}
.cat-option.active .cat-active-bar {
  background: #4F46E5;
}

/* 选中态背景 + 文字色 */
.cat-option.active {
  background: #EEF2FF;
  color: #4F46E5;
  font-weight: 600;
}

/* 一级分类 */
.cat-option.parent .cat-label {
  font-weight: 600;
  font-size: 27rpx;
  color: #374151;
  flex: 1;
}
.cat-option.parent.active .cat-label { color: #4F46E5; }

/* 分类名称 */
.cat-label {
  flex: 1;
  line-height: 1.4;
}

/* 展开/收起箭头 */
.cat-expand-arrow {
  font-size: 20rpx;
  color: #9CA3AF;
  padding: 4rpx 20rpx 4rpx 8rpx;
  transition: transform 0.2s;
  display: inline-block;
}
.cat-expand-arrow.open {
  transform: rotate(90deg);
  color: #4F46E5;
}

/* 二级分类缩进 */
.cat-option.child {
  padding-left: 0;
  font-size: 24rpx;
}
.cat-option.child .cat-active-bar {
  margin-left: 6rpx;
}
.cat-connector {
  font-size: 22rpx;
  color: #D1D5DB;
  margin-right: 10rpx;
  flex-shrink: 0;
  padding-left: 28rpx;
}

/* ===================== 服务列表 ===================== */
.service-list { flex: 1; min-width: 0; }
.loading, .empty { text-align: center; padding: 100rpx 0; color: #9CA3AF; font-size: 28rpx; }
.card {
  background: #fff; border-radius: 16rpx; padding: 24rpx;
  margin-bottom: 16rpx; box-shadow: 0 2rpx 12rpx rgba(0, 0, 0, 0.06);
  cursor: pointer; transition: box-shadow 0.2s, transform 0.15s;
}
.card:active { box-shadow: 0 4rpx 20rpx rgba(79, 70, 229, 0.12); transform: translateY(-2rpx); }
.card-top { display: flex; justify-content: space-between; align-items: center; }
.svc-info { flex: 1; min-width: 0; display: flex; align-items: center; flex-wrap: wrap; gap: 8rpx; }
.name { font-size: 30rpx; font-weight: 600; color: #1F2937; }
.category-tag {
  font-size: 22rpx; color: #4F46E5; background: #EEF2FF;
  padding: 4rpx 12rpx; border-radius: 8rpx;
}
.status { font-size: 24rpx; padding: 6rpx 16rpx; border-radius: 20rpx; flex-shrink: 0; }
.status.active { color: #059669; background: #D1FAE5; }
.status.inactive { color: #DC2626; background: #FEE2E2; }
.card-meta { display: flex; gap: 24rpx; margin-top: 12rpx; }
.price { font-size: 32rpx; font-weight: bold; color: #4F46E5; }
.duration { font-size: 26rpx; color: #6B7280; line-height: 44rpx; }
.desc { font-size: 24rpx; color: #9CA3AF; margin-top: 8rpx; }

/* ===================== 手机端响应式 ===================== */
@media (max-width: 600px) {
  /* 主布局改为纵向 */
  .main-layout {
    flex-direction: column;
  }

  /* 分类包裹层占满宽度 */
  .category-sidebar-wrap {
    width: 100%;
    min-width: unset;
  }

  /* 折叠触发条显示 */
  .mobile-cat-toggle {
    display: flex;
    align-items: center;
    justify-content: space-between;
    background: #fff;
    border-radius: 16rpx;
    padding: 20rpx 24rpx;
    box-shadow: 0 2rpx 12rpx rgba(0, 0, 0, 0.06);
    cursor: pointer;
  }
  .mobile-cat-toggle:active { background: #F9FAFB; }
  .mobile-cat-toggle-left {
    display: flex;
    align-items: center;
    gap: 12rpx;
  }
  .mobile-cat-toggle-icon { font-size: 30rpx; }
  .mobile-cat-toggle-label {
    font-size: 28rpx;
    font-weight: 600;
    color: #374151;
  }
  .mobile-cat-badge {
    font-size: 20rpx;
    color: #fff;
    background: #4F46E5;
    padding: 4rpx 12rpx;
    border-radius: 20rpx;
  }
  .mobile-cat-arrow {
    font-size: 22rpx;
    color: #9CA3AF;
    transition: transform 0.25s;
    display: inline-block;
  }
  .mobile-cat-arrow.open {
    transform: rotate(180deg);
  }

  /* 分类面板展开/收起动画 */
  .category-sidebar {
    /* 桌面端固定宽高的属性手机端全部重置 */
    width: 100%;
    min-width: unset;
    max-height: 0;
    overflow: hidden;
    padding: 0;
    margin-top: 0;
    box-shadow: none;
    border-radius: 0 0 16rpx 16rpx;
    transition: max-height 0.3s ease, padding 0.3s ease;
    /* 展开时从折叠条下方紧接 */
    background: #fff;
  }
  .category-sidebar.panel-open {
    max-height: 800rpx;
    padding: 12rpx 0;
    margin-top: 4rpx;
    overflow-y: auto;
    box-shadow: 0 4rpx 16rpx rgba(0, 0, 0, 0.08);
  }

  /* 手机端分类选项保持纵向列表，不改变结构 */
  .cat-option {
    padding: 18rpx 20rpx 18rpx 0;
    font-size: 28rpx;
  }
  .cat-active-bar {
    height: 40rpx;
  }

  /* 一级分类在手机端字号更大一些，更易点击 */
  .cat-option.parent .cat-label {
    font-size: 29rpx;
  }

  /* 二级分类保持缩进 */
  .cat-option.child {
    font-size: 26rpx;
  }
  .cat-option.child .cat-active-bar {
    margin-left: 6rpx;
  }
  .cat-connector {
    padding-left: 32rpx;
  }
}
</style>
