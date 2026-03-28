<template>
  <SideLayout>
  <view class="page">

    <!-- 顶部标题区 -->
    <view class="header">
      <view class="header-left">
        <text class="title">宠物档案</text>
        <text class="subtitle">猫咪洗护健康记录 · 共 {{ total }} 只</text>
      </view>
      <view class="btn-add" @click="goAdd">
        <text class="btn-add-icon">+</text>
        <text class="btn-add-text">新增档案</text>
      </view>
    </view>

    <!-- 搜索栏 -->
    <view class="search-bar">
      <text class="search-icon">🔍</text>
      <input
        v-model="keyword"
        placeholder="搜索猫咪名 / 主人昵称 / 手机号"
        class="search-input"
        confirm-type="search"
        @confirm="onSearch"
        @input="onSearchInput"
      />
      <view v-if="keyword" class="search-clear" @click="clearSearch">✕</view>
    </view>

    <!-- 加载中 -->
    <view v-if="loading" class="loading">
      <text class="loading-icon">🐾</text>
      <text class="loading-text">正在加载猫咪档案...</text>
    </view>

    <!-- 空状态 -->
    <view v-else-if="groupedList.length === 0" class="empty">
      <text class="empty-icon">🐱</text>
      <text class="empty-title">还没有宠物档案</text>
      <text class="empty-desc">{{ keyword ? '没有找到匹配的猫咪' : '点击右上角新增第一只猫咪吧' }}</text>
    </view>

    <!-- 分组列表 -->
    <view v-else class="list">
      <view v-for="group in groupedList" :key="group.key" class="owner-group">

        <!-- 主人分组头部 -->
        <view class="group-header" v-if="group.ownerName" @click.stop="goCustomer(group.key)">
          <view class="group-accent"></view>
          <view class="group-avatar">
            <text class="group-avatar-text">{{ group.ownerName.charAt(0) }}</text>
          </view>
          <view class="group-info">
            <view class="group-name-row">
              <text class="group-name">{{ group.ownerName }}</text>
              <text class="group-count-badge">{{ group.pets.length }} 只</text>
            </view>
            <view class="owner-tags" v-if="group.customer">
              <text class="owner-tag owner-tag-card" v-if="group.customer.member_card">
                🎫 {{ group.customer.member_card.card_name }}
              </text>
              <text class="owner-tag owner-tag-balance" v-if="group.customer.member_balance">
                💰 余额 ¥{{ group.customer.member_balance }}
              </text>
              <text class="owner-tag owner-tag-visit" v-if="group.customer.last_visit_at">
                🕐 上次 {{ formatDate(group.customer.last_visit_at) }}
              </text>
            </view>
          </view>
          <text class="group-arrow">›</text>
        </view>

        <!-- 无主人分组头 -->
        <view class="group-header group-header-none" v-else>
          <view class="group-accent group-accent-none"></view>
          <text class="group-name-none">暂无主人信息</text>
        </view>

        <!-- 猫咪卡片 -->
        <view
          class="card"
          v-for="pet in group.pets"
          :key="pet.ID"
          @click="goEdit(pet.ID)"
        >
          <!-- 左侧 Avatar 区域 -->
          <view class="card-avatar-col">
            <img v-if="getPetAvatarUrl(pet.avatar)" :src="getPetAvatarUrl(pet.avatar)" class="avatar-image" />
            <view v-else class="avatar-circle" :class="pet.species === '犬' ? 'avatar-dog' : 'avatar-cat'">
              <text class="avatar-emoji">{{ pet.species === '犬' ? '🐕' : '🐱' }}</text>
            </view>
            <text class="avatar-weight" v-if="pet.weight">{{ pet.weight }}kg</text>
          </view>

          <!-- 右侧信息区域 -->
          <view class="card-body">
            <!-- 名字行 -->
            <view class="card-name-row">
              <text class="pet-name">{{ pet.name }}</text>
              <text class="pet-gender" :class="pet.gender === 1 ? 'gender-male' : pet.gender === 2 ? 'gender-female' : ''">
                {{ pet.gender === 1 ? '♂' : pet.gender === 2 ? '♀' : '' }}
              </text>
            </view>

            <!-- 品种与年龄 -->
            <text class="pet-breed">
              {{ pet.breed || '未知品种' }}<text v-if="pet.birth_date"> · {{ calcAge(pet.birth_date) }}</text>
            </text>

            <!-- 标签行 -->
            <view class="tag-row" v-if="pet.fur_level || pet.neutered || pet.personality || (pet.aggression && pet.aggression !== '无')">
              <text class="tag tag-fur" v-if="pet.fur_level">{{ pet.fur_level }}</text>
              <text class="tag tag-neutered" v-if="pet.neutered">已绝育</text>
              <text
                class="tag tag-personality"
                v-if="pet.personality"
                :style="{ background: getPersonalityBg(pet.personality), color: getPersonalityColor(pet.personality) }"
              >{{ pet.personality }}</text>
              <text
                class="tag tag-aggression"
                v-if="pet.aggression && pet.aggression !== '无'"
              >⚡ {{ pet.aggression }}</text>
            </view>

            <!-- 注意事项 -->
            <view class="care-notes" v-if="pet.care_notes">
              <view class="care-notes-bar"></view>
              <text class="care-notes-text">{{ pet.care_notes }}</text>
            </view>
          </view>
        </view>

      </view>

      <!-- 加载更多 -->
      <view v-if="loadingMore" class="load-more">
        <text class="loading-text">加载中...</text>
      </view>
      <view v-else-if="hasMore" class="load-more" @click="loadMore">
        <text class="load-more-text">上滑加载更多</text>
      </view>
      <view v-else class="load-more">
        <text class="load-more-done">已加载全部 {{ total }} 只宠物</text>
      </view>
    </view>

  </view>
  </SideLayout>
</template>

<script setup lang="ts">
import SideLayout from '@/components/SideLayout.vue'
import { ref, computed, onMounted } from 'vue'
import { onShow, onReachBottom } from '@dcloudio/uni-app'
import { getPetList } from '@/api/pet'
import { getPersonalityColor, getPersonalityBg } from '@/utils/personality'

function calcAge(birthDate: string): string {
  if (!birthDate) return ''
  const birth = new Date(birthDate)
  const now = new Date()
  const months = (now.getFullYear() - birth.getFullYear()) * 12 + (now.getMonth() - birth.getMonth())
  if (months < 1) return '不到1个月'
  if (months < 12) return `${months}个月`
  const years = Math.floor(months / 12)
  const rem = months % 12
  return rem > 0 ? `${years}岁${rem}个月` : `${years}岁`
}

function getPetAvatarUrl(avatar?: string): string {
  if (!avatar) return ''
  if (avatar.startsWith('http')) return avatar
  if (typeof window === 'undefined') return avatar
  return `${window.location.origin}${avatar}`
}

interface OwnerGroup {
  key: string
  ownerName: string
  customer?: Customer
  pets: Pet[]
}

const PAGE_SIZE = 50
const list = ref<Pet[]>([])
const total = ref(0)
const currentPage = ref(1)
const loading = ref(true)
const loadingMore = ref(false)
const hasMore = ref(true)
const keyword = ref('')
let searchTimer: ReturnType<typeof setTimeout> | null = null

const groupedList = computed<OwnerGroup[]>(() => {
  const groups: Record<string, OwnerGroup> = {}
  const noOwnerKey = '__no_owner__'

  for (const pet of list.value) {
    const cid = pet.customer_id || 0
    const key = cid ? String(cid) : noOwnerKey
    if (!groups[key]) {
      const ownerName = pet.customer
        ? (pet.customer.nickname || pet.customer.phone || `客户#${cid}`)
        : ''
      groups[key] = { key, ownerName, customer: pet.customer, pets: [] }
    }
    groups[key].pets.push(pet)
  }

  // 有主人的排前面，无主人的排后面
  const sorted = Object.values(groups).sort((a, b) => {
    if (a.key === noOwnerKey) return 1
    if (b.key === noOwnerKey) return -1
    return 0
  })
  return sorted
})

async function loadData() {
  loading.value = true
  currentPage.value = 1
  try {
    const params: any = { page: 1, page_size: PAGE_SIZE }
    if (keyword.value.trim()) params.keyword = keyword.value.trim()
    const res = await getPetList(params)
    list.value = res.data.list || []
    total.value = res.data.total || list.value.length
    hasMore.value = list.value.length < total.value
  } finally { loading.value = false }
}

async function loadMore() {
  if (loadingMore.value || !hasMore.value) return
  loadingMore.value = true
  try {
    currentPage.value++
    const params: any = { page: currentPage.value, page_size: PAGE_SIZE }
    if (keyword.value.trim()) params.keyword = keyword.value.trim()
    const res = await getPetList(params)
    const newItems = res.data.list || []
    list.value = [...list.value, ...newItems]
    total.value = res.data.total || list.value.length
    hasMore.value = list.value.length < total.value
  } finally { loadingMore.value = false }
}

function onSearch() { loadData() }
function onSearchInput() {
  if (searchTimer) clearTimeout(searchTimer)
  searchTimer = setTimeout(() => loadData(), 400)
}
function clearSearch() {
  keyword.value = ''
  loadData()
}

function goAdd() { uni.navigateTo({ url: '/pages/pet/edit' }) }
function goEdit(id: number) { uni.navigateTo({ url: `/pages/pet/edit?id=${id}` }) }
function goCustomer(id: string) { uni.navigateTo({ url: `/pages/customer/detail?id=${id}` }) }

function formatDate(dateStr: string): string {
  if (!dateStr) return ''
  const d = new Date(dateStr)
  const m = d.getMonth() + 1
  const day = d.getDate()
  return `${m}/${day}`
}

onMounted(loadData)
onShow(loadData)
onReachBottom(loadMore)
</script>

<style scoped>
/* =====================
   页面基础
   ===================== */
.page {
  padding: 32rpx 32rpx 48rpx;
  background: #F5F6FA;
  min-height: 100vh;
}

/* =====================
   顶部标题
   ===================== */
.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 28rpx;
}

.header-left {
  display: flex;
  flex-direction: column;
}

.title {
  font-size: 40rpx;
  font-weight: 600;
  color: #1F2937;
  line-height: 1.2;
  letter-spacing: -0.5rpx;
}

.subtitle {
  font-size: 22rpx;
  color: #9CA3AF;
  margin-top: 4rpx;
}

.btn-add {
  display: flex;
  align-items: center;
  gap: 6rpx;
  background: linear-gradient(135deg, #6366F1, #4F46E5);
  padding: 16rpx 28rpx;
  border-radius: 999rpx;
  box-shadow: 0 4rpx 16rpx rgba(99, 102, 241, 0.35);
  transition: opacity 0.2s, transform 0.15s;
}

.btn-add:active {
  opacity: 0.85;
  transform: scale(0.96);
}

.btn-add-icon {
  font-size: 32rpx;
  color: rgba(255, 255, 255, 0.95);
  line-height: 1;
  margin-top: -2rpx;
}

.btn-add-text {
  font-size: 26rpx;
  font-weight: 500;
  color: #fff;
}

/* =====================
   搜索栏
   ===================== */
.search-bar {
  position: relative;
  display: flex;
  align-items: center;
  background: #F3F4F6;
  border-radius: 20rpx;
  padding: 0 24rpx;
  margin-bottom: 28rpx;
  box-shadow: 0 2rpx 10rpx rgba(99, 102, 241, 0.06);
  border: 1.5rpx solid #E5E7EB;
}

.search-icon {
  font-size: 28rpx;
  margin-right: 12rpx;
  flex-shrink: 0;
}

.search-input {
  flex: 1;
  height: 80rpx;
  font-size: 26rpx;
  color: #1F2937;
  background: transparent;
}

.search-clear {
  font-size: 24rpx;
  color: #9CA3AF;
  padding: 8rpx;
  flex-shrink: 0;
}

/* =====================
   加载 / 空状态
   ===================== */
.loading {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 120rpx 0;
  gap: 16rpx;
}

.loading-icon {
  font-size: 60rpx;
  animation: bounce 1s infinite alternate;
}

.loading-text {
  font-size: 26rpx;
  color: #9CA3AF;
}

@keyframes bounce {
  from { transform: translateY(0); }
  to   { transform: translateY(-12rpx); }
}

.empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 120rpx 0;
  gap: 16rpx;
}

.empty-icon {
  font-size: 80rpx;
}

.empty-title {
  font-size: 30rpx;
  font-weight: 500;
  color: #1F2937;
}

.empty-desc {
  font-size: 24rpx;
  color: #9CA3AF;
  text-align: center;
}

/* =====================
   主人分组
   ===================== */
.owner-group {
  margin-bottom: 32rpx;
}

/* 分组头部 — 有主人 */
.group-header {
  display: flex;
  align-items: center;
  background: #F9FAFB;
  border-radius: 20rpx 20rpx 0 0;
  padding: 20rpx 24rpx 20rpx 0;
  margin-bottom: 3rpx;
  box-shadow: 0 2rpx 8rpx rgba(99, 102, 241, 0.06);
  transition: background 0.15s;
  overflow: hidden;
  border: 1.5rpx solid #E5E7EB;
  border-bottom: none;
}

.group-header:active {
  background: #EEF2FF;
}

.group-accent {
  width: 8rpx;
  height: 60rpx;
  background: linear-gradient(180deg, #6366F1, #4F46E5);
  border-radius: 0 4rpx 4rpx 0;
  margin-right: 20rpx;
  flex-shrink: 0;
}

.group-accent-none {
  background: #C7D2FE;
}

.group-avatar {
  width: 64rpx;
  height: 64rpx;
  border-radius: 50%;
  background: linear-gradient(135deg, #EEF2FF, #C7D2FE);
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  margin-right: 18rpx;
}

.group-avatar-text {
  font-size: 28rpx;
  font-weight: 600;
  color: #4F46E5;
}

.group-info {
  flex: 1;
  min-width: 0;
}

.group-name-row {
  display: flex;
  align-items: center;
  gap: 12rpx;
  margin-bottom: 8rpx;
}

.group-name {
  font-size: 28rpx;
  font-weight: 600;
  color: #1F2937;
}

.group-count-badge {
  font-size: 20rpx;
  font-weight: 500;
  color: #4F46E5;
  background: #EEF2FF;
  padding: 2rpx 12rpx;
  border-radius: 999rpx;
  border: 1rpx solid #C7D2FE;
}

.owner-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 8rpx;
}

.owner-tag {
  font-size: 20rpx;
  padding: 3rpx 12rpx;
  border-radius: 999rpx;
  font-weight: 400;
}

.owner-tag-card {
  background: #EEF2FF;
  color: #4338CA;
  border: 1rpx solid #C7D2FE;
}

.owner-tag-balance {
  background: #F0FDF4;
  color: #065F46;
  border: 1rpx solid #BBF7D0;
}

.owner-tag-visit {
  background: #F9FAFB;
  color: #6B7280;
  border: 1rpx solid #E5E7EB;
}

.group-arrow {
  font-size: 40rpx;
  color: #C7D2FE;
  font-weight: 300;
  margin-left: 8rpx;
  flex-shrink: 0;
}

/* 无主人分组头 */
.group-header-none {
  border-radius: 20rpx 20rpx 0 0;
}

.group-name-none {
  font-size: 24rpx;
  color: #9CA3AF;
  font-weight: 400;
}

/* =====================
   猫咪卡片
   ===================== */
.card {
  background: #fff;
  padding: 24rpx 24rpx;
  margin-bottom: 3rpx;
  display: flex;
  align-items: flex-start;
  gap: 20rpx;
  border-left: 1.5rpx solid #E5E7EB;
  border-right: 1.5rpx solid #E5E7EB;
  box-shadow: 0 1rpx 4rpx rgba(99, 102, 241, 0.04);
  transition: background 0.15s, transform 0.15s;
}

.card:last-child {
  border-radius: 0 0 20rpx 20rpx;
  margin-bottom: 0;
  border-bottom: 1.5rpx solid #E5E7EB;
  box-shadow: 0 4rpx 16rpx rgba(99, 102, 241, 0.08);
}

.card:only-child {
  border-radius: 0 0 20rpx 20rpx;
}

.card:active {
  background: #F5F6FA;
  transform: scale(0.99);
}

/* 左侧 Avatar 列 */
.card-avatar-col {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8rpx;
  flex-shrink: 0;
}

.avatar-circle {
  width: 88rpx;
  height: 88rpx;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
}

.avatar-image {
  width: 88rpx;
  height: 88rpx;
  border-radius: 50%;
  object-fit: cover;
  background: #F5F5F5;
  border: 1.5rpx solid #E5E7EB;
  box-shadow: 0 4rpx 12rpx rgba(99, 102, 241, 0.12);
}

.avatar-cat {
  background: linear-gradient(135deg, #FDE68A, #FCD34D);
}

.avatar-dog {
  background: linear-gradient(135deg, #BBF7D0, #6EE7B7);
}

.avatar-emoji {
  font-size: 48rpx;
}

.avatar-weight {
  font-size: 20rpx;
  color: #6B7280;
  background: #F3F4F6;
  padding: 2rpx 10rpx;
  border-radius: 999rpx;
  font-weight: 500;
  white-space: nowrap;
  border: 1rpx solid #E5E7EB;
}

/* 右侧信息列 */
.card-body {
  flex: 1;
  min-width: 0;
  padding-top: 4rpx;
}

.card-name-row {
  display: flex;
  align-items: center;
  gap: 8rpx;
  margin-bottom: 6rpx;
}

.pet-name {
  font-size: 32rpx;
  font-weight: 600;
  color: #1F2937;
  letter-spacing: -0.3rpx;
}

.pet-gender {
  font-size: 26rpx;
  font-weight: 500;
}

.gender-male {
  color: #3B82F6;
}

.gender-female {
  color: #EC4899;
}

.pet-breed {
  font-size: 24rpx;
  color: #6B7280;
  display: block;
  margin-bottom: 14rpx;
}

/* 标签行 */
.tag-row {
  display: flex;
  flex-wrap: wrap;
  gap: 8rpx;
  margin-bottom: 14rpx;
}

.tag {
  font-size: 21rpx;
  padding: 4rpx 14rpx;
  border-radius: 999rpx;
  font-weight: 400;
}

.tag-fur {
  background: #F3F4F6;
  color: #6B7280;
  border: 1rpx solid #E5E7EB;
}

.tag-neutered {
  background: #EFF6FF;
  color: #3B7DD8;
  border: 1rpx solid #BFDBFE;
}

.tag-personality {
  /* color/bg 由 inline style 注入 */
}

.tag-aggression {
  background: #FEE2E2;
  color: #DC2626;
}

/* 注意事项 */
.care-notes {
  display: flex;
  align-items: flex-start;
  gap: 10rpx;
  background: #F0F4FF;
  border-radius: 12rpx;
  padding: 12rpx 16rpx;
  margin-top: 4rpx;
  border: 1rpx solid #C7D2FE;
}

.care-notes-bar {
  width: 6rpx;
  min-height: 32rpx;
  align-self: stretch;
  background: linear-gradient(180deg, #6366F1, #4F46E5);
  border-radius: 999rpx;
  flex-shrink: 0;
}

.care-notes-text {
  font-size: 24rpx;
  color: #4338CA;
  line-height: 1.6;
  flex: 1;
}

.load-more {
  text-align: center;
  padding: 32rpx 0 48rpx;
}

.load-more-text {
  font-size: 24rpx;
  color: #9CA3AF;
}

.load-more-done {
  font-size: 22rpx;
  color: #D1D5DB;
}
</style>
