<template>
  <SideLayout>
  <view class="page">
    <view class="header">
      <view class="header-main">
        <text class="title">客户管理</text>
        <text class="title-count">{{ total }} 人</text>
      </view>
      <view class="header-actions">
        <view class="btn-tag" @click="goTagManage">标签</view>
        <view class="btn-trash" @click="goTrash">回收站</view>
        <view class="btn-add" @click="goAdd">+ 新增</view>
      </view>
    </view>

    <view class="search-bar">
      <input v-model="keyword" placeholder="搜索客户姓名/手机号/猫咪名字" class="search-input" @confirm="loadData" />
    </view>

    <view class="tag-entry" @click="goTagManage">
      <view class="tag-entry-main">
        <text class="tag-entry-title">客户标签</text>
        <text class="tag-entry-desc">统一管理熟客、回访、高净值等标签</text>
      </view>
      <text class="tag-entry-action">去设置</text>
    </view>

    <view class="filter-section">
      <text class="filter-label">会员等级</text>
      <scroll-view scroll-x class="filter-scroll" show-scrollbar="false">
        <view class="filter-tabs">
          <view :class="['filter-chip', activeTemplateId === 0 ? 'filter-chip-active' : '']" @click="switchTemplate(0)">
            全部等级
          </view>
          <view
            v-for="tpl in cardTemplates"
            :key="tpl.ID"
            :class="['filter-chip', activeTemplateId === tpl.ID ? 'filter-chip-active' : '']"
            @click="switchTemplate(tpl.ID)"
          >
            {{ tpl.name }}
          </view>
        </view>
      </scroll-view>
    </view>

    <view class="filter-section" v-if="tagFilters.length">
      <text class="filter-label">客户标签</text>
      <scroll-view scroll-x class="filter-scroll" show-scrollbar="false">
        <view class="filter-tabs">
          <view :class="['filter-chip', activeTagId === 0 ? 'filter-chip-active' : '']" @click="switchTag(0)">
            全部标签
          </view>
          <view
            v-for="tag in tagFilters"
            :key="tag.ID"
            :class="['filter-chip', activeTagId === tag.ID ? 'filter-chip-active' : '']"
            :style="activeTagId === tag.ID ? { background: tag.color, color: '#fff' } : { background: withAlpha(tag.color, 0.12), color: tag.color }"
            @click="switchTag(tag.ID)"
          >
            {{ tag.name }}
          </view>
        </view>
      </scroll-view>
    </view>

    <view v-if="loading" class="loading">
      <text class="loading-icon">👥</text>
      <text class="loading-text">正在加载客户...</text>
    </view>
    <view v-else-if="list.length === 0" class="empty">
      <text class="empty-icon">👥</text>
      <text class="empty-title">还没有客户</text>
      <text class="empty-desc">点击右上角新增第一位客户吧</text>
    </view>

    <view v-else class="list">
      <view class="card" v-for="item in list" :key="item.ID" @click="goDetail(item.ID)" @longpress="onLongPress(item)">
        <view class="card-top">
          <view class="avatar">{{ (item.nickname || item.phone || '客').charAt(0) }}</view>
          <view class="info">
            <text class="name">{{ item.nickname || item.phone || '未命名' }}</text>
            <text class="phone" v-if="item.nickname">{{ item.phone || '未绑定手机' }}</text>
          </view>
          <view class="visit-count">到店 {{ item.visit_count }} 次</view>
        </view>
        <view class="card-middle">
          <view :class="['info-pill', hasMemberCard(item) ? 'info-pill-member' : 'info-pill-muted']">
            <text class="info-pill-label">会员卡</text>
            <text class="info-pill-value">{{ getMemberCardName(item) }}</text>
          </view>
          <view :class="['info-pill', hasMemberCard(item) ? 'info-pill-balance' : 'info-pill-muted']">
            <text class="info-pill-label">余额</text>
            <text class="info-pill-value">¥{{ formatMoney(getMemberBalance(item)) }}</text>
          </view>
        </view>
        <!-- 猫咪信息 -->
        <view class="pet-list" v-if="item.pets?.length">
          <view class="pet-item" v-for="pet in item.pets" :key="pet.ID">
            <view class="pet-item-top">
              <text class="pet-item-name">{{ pet.name }}</text>
              <text class="pet-item-meta">{{ formatPetMeta(pet) }}</text>
            </view>
            <view class="pet-item-tags" v-if="getPetTags(pet).length">
              <text
                v-for="tag in getPetTags(pet)"
                :key="tag.text"
                class="pet-item-tag"
                :style="tag.style"
              >{{ tag.text }}</text>
            </view>
          </view>
        </view>
        <view class="card-bottom">
          <text class="spent">累计消费 ¥{{ item.total_spent.toFixed(2) }}</text>
          <view class="customer-tags" v-if="item.customer_tags?.length">
            <text
              v-for="tag in item.customer_tags.slice(0, 3)"
              :key="tag.ID"
              class="customer-tag"
              :style="{ background: withAlpha(tag.color, 0.14), color: tag.color }"
            >
              {{ tag.name }}
            </text>
          </view>
        </view>
      </view>

      <!-- 加载更多 -->
      <view v-if="loadingMore" class="load-more">加载中...</view>
      <view v-else-if="hasMore" class="load-more" @click="loadMore">上滑加载更多</view>
      <view v-else class="load-more load-more-done">已加载全部 {{ total }} 位客户</view>
    </view>
  </view>
  </SideLayout>
</template>

<script setup lang="ts">
import SideLayout from '@/components/SideLayout.vue'
import { ref } from 'vue'
import { onShow, onReachBottom } from '@dcloudio/uni-app'
import { getCustomerList, deleteCustomer } from '@/api/customer'
import { getCustomerTags } from '@/api/customer-tag'
import { getCardTemplates } from '@/api/member-card'
import { getPersonalityBg, getPersonalityColor } from '@/utils/personality'

const PAGE_SIZE = 50
const list = ref<Customer[]>([])
const cardTemplates = ref<MemberCardTemplate[]>([])
const tagFilters = ref<CustomerTag[]>([])
const total = ref(0)
const currentPage = ref(1)
const loading = ref(true)
const loadingMore = ref(false)
const hasMore = ref(true)
const keyword = ref('')
const activeTemplateId = ref(0)
const activeTagId = ref(0)

function hasMemberCard(item: Customer) {
  return !!(item.member_card_id || item.member_card?.ID)
}

function getMemberCardName(item: Customer) {
  return item.member_card?.template?.name || item.member_card?.card_name || '未开通会员卡'
}

function getMemberBalance(item: Customer) {
  return item.member_balance || item.member_card?.balance || 0
}

function formatMoney(amount: number) {
  return Number(amount || 0).toFixed(2)
}

function withAlpha(color: string, alpha: number) {
  const hex = color.replace('#', '')
  if (hex.length !== 6) return color
  const value = Math.round(alpha * 255).toString(16).padStart(2, '0')
  return `#${hex}${value}`
}

function formatPetMeta(pet: any) {
  const parts: string[] = []
  if (pet.breed) parts.push(pet.breed)
  if (pet.gender === 1) parts.push('弟弟')
  else if (pet.gender === 2) parts.push('妹妹')
  return parts.join(' · ') || ''
}

function getPetAge(birthDate?: string) {
  if (!birthDate) return ''
  const birth = new Date(birthDate)
  if (Number.isNaN(birth.getTime())) return ''
  const now = new Date()
  const months = (now.getFullYear() - birth.getFullYear()) * 12 + (now.getMonth() - birth.getMonth())
  if (months < 1) return '不到1个月'
  if (months < 12) return `${months}个月`
  const years = Math.floor(months / 12)
  const rem = months % 12
  return rem > 0 ? `${years}岁${rem}个月` : `${years}岁`
}

function getPetTags(pet: any) {
  const tags: Array<{ text: string; style: string }> = []
  const age = getPetAge(pet.birth_date)
  if (age) tags.push({ text: age, style: 'background:#F3F4F6;color:#4B5563;' })
  if (pet.fur_level) tags.push({ text: pet.fur_level, style: 'background:#EEF2FF;color:#4F46E5;' })
  if (pet.neutered) tags.push({ text: '已绝育', style: 'background:#ECFDF5;color:#047857;' })
  if (pet.personality) {
    tags.push({
      text: pet.personality,
      style: `background:${getPersonalityBg(pet.personality)};color:${getPersonalityColor(pet.personality)};`,
    })
  }
  return tags
}

async function loadTemplates() {
  try {
    const res = await getCardTemplates()
    cardTemplates.value = res.data || []
  } catch {
    cardTemplates.value = []
  }
}

async function loadTagFilters() {
  try {
    const res = await getCustomerTags()
    tagFilters.value = (res.data || []).filter(tag => tag.status === 1)
  } catch {
    tagFilters.value = []
  }
}

async function loadData() {
  loading.value = true
  currentPage.value = 1
  try {
    const res = await getCustomerList({
      page: 1,
      page_size: PAGE_SIZE,
      keyword: keyword.value.trim() || undefined,
      member_card_template_id: activeTemplateId.value || undefined,
      customer_tag_id: activeTagId.value || undefined,
    })
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
    const res = await getCustomerList({
      page: currentPage.value,
      page_size: PAGE_SIZE,
      keyword: keyword.value.trim() || undefined,
      member_card_template_id: activeTemplateId.value || undefined,
      customer_tag_id: activeTagId.value || undefined,
    })
    const newItems = res.data.list || []
    list.value = [...list.value, ...newItems]
    total.value = res.data.total || list.value.length
    hasMore.value = list.value.length < total.value
  } finally { loadingMore.value = false }
}

function goAdd() { uni.navigateTo({ url: '/pages/customer/edit' }) }
function goDetail(id: number) { uni.navigateTo({ url: `/pages/customer/detail?id=${id}` }) }
function goTrash() { uni.navigateTo({ url: '/pages/customer/trash' }) }
function goTagManage() { uni.navigateTo({ url: '/pages/customer/tag-manage' }) }
function switchTemplate(id: number) {
  if (activeTemplateId.value === id) return
  activeTemplateId.value = id
  loadData()
}

function switchTag(id: number) {
  if (activeTagId.value === id) return
  activeTagId.value = id
  loadData()
}

function onLongPress(item: any) {
  try { uni.vibrateShort({}) } catch (_) {}
  uni.showActionSheet({
    itemList: ['删除该客户'],
    success: (res) => {
      if (res.tapIndex === 0) {
        uni.showModal({
          title: '确认删除',
          content: `确定要删除客户「${item.nickname || '未命名'}」吗？\n可在回收站中1天内恢复。`,
          confirmColor: '#EF4444',
          success: async (modalRes) => {
            if (modalRes.confirm) {
              try {
                await deleteCustomer(item.ID)
                uni.showToast({ title: '已移入回收站', icon: 'success' })
                await loadData()
              } catch {
                uni.showToast({ title: '删除失败', icon: 'none' })
              }
            }
          }
        })
      }
    }
  })
}

onShow(async () => {
  await loadTemplates()
  await loadTagFilters()
  await loadData()
})
onReachBottom(loadMore)
</script>

<style scoped>
.page { padding: 20rpx 24rpx 24rpx; }
.header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 14rpx; gap: 12rpx; }
.header-main { display: flex; align-items: baseline; gap: 10rpx; min-width: 0; }
.title { font-size: 32rpx; font-weight: 700; color: #1F2937; line-height: 1.1; }
.title-count { font-size: 22rpx; color: #6B7280; background: #F3F4F6; padding: 4rpx 12rpx; border-radius: 999rpx; white-space: nowrap; }
.header-actions { display: flex; gap: 8rpx; align-items: center; flex-shrink: 0; }
.btn-tag { font-size: 22rpx; color: #4F46E5; background: #EEF2FF; padding: 10rpx 16rpx; border-radius: 999rpx; line-height: 1; }
.btn-trash { font-size: 22rpx; color: #6B7280; background: #F3F4F6; padding: 10rpx 16rpx; border-radius: 999rpx; line-height: 1; }
.btn-add { font-size: 24rpx; color: #fff; background: #4F46E5; padding: 10rpx 18rpx; border-radius: 999rpx; line-height: 1; }
.search-bar { margin-bottom: 18rpx; }
.search-input { background: #fff; border-radius: 12rpx; padding: 14rpx 20rpx; font-size: 26rpx; }
.tag-entry { margin-bottom: 18rpx; background: linear-gradient(135deg, #EEF2FF, #F8FAFC); border-radius: 16rpx; padding: 18rpx 20rpx; display: flex; align-items: center; justify-content: space-between; gap: 16rpx; }
.tag-entry-main { min-width: 0; }
.tag-entry-title { display: block; font-size: 28rpx; font-weight: 700; color: #1F2937; }
.tag-entry-desc { display: block; margin-top: 6rpx; font-size: 22rpx; color: #6B7280; line-height: 1.5; }
.tag-entry-action { flex-shrink: 0; font-size: 22rpx; color: #4F46E5; background: rgba(255,255,255,0.92); padding: 10rpx 16rpx; border-radius: 999rpx; }
.filter-section { margin-bottom: 18rpx; }
.filter-label { display: block; margin-bottom: 10rpx; font-size: 22rpx; color: #6B7280; }
.filter-scroll { white-space: nowrap; }
.filter-tabs { display: inline-flex; gap: 12rpx; padding-bottom: 4rpx; }
.filter-chip { font-size: 24rpx; color: #6B7280; background: #F3F4F6; padding: 10rpx 18rpx; border-radius: 999rpx; white-space: nowrap; }
.filter-chip-active { color: #fff; background: #4F46E5; }
.filter-chip-count { font-size: 20rpx; opacity: 0.88; }
.loading, .empty { display: flex; flex-direction: column; align-items: center; padding: 120rpx 0; gap: 16rpx; }
.loading-icon, .empty-icon { font-size: 64rpx; }
.loading-text { font-size: 26rpx; color: #9CA3AF; }
.loading-icon { animation: bounce 1s infinite alternate; }
.empty-title { font-size: 30rpx; font-weight: 500; color: #4B5563; }
.empty-desc { font-size: 24rpx; color: #9CA3AF; }
@keyframes bounce { from { transform: translateY(0); } to { transform: translateY(-12rpx); } }
.card { background: #fff; border-radius: 16rpx; padding: 24rpx; margin-bottom: 16rpx; box-shadow: 0 2rpx 8rpx rgba(0,0,0,0.04); }
.card-top { display: flex; align-items: center; }
.avatar { width: 80rpx; height: 80rpx; border-radius: 50%; background: #6366F1; color: #fff; display: flex; align-items: center; justify-content: center; font-size: 32rpx; font-weight: bold; }
.info { flex: 1; margin-left: 20rpx; }
.name { font-size: 30rpx; font-weight: 600; color: #1F2937; display: block; }
.phone { font-size: 24rpx; color: #6B7280; display: block; margin-top: 4rpx; }
.visit-count { font-size: 24rpx; color: #4F46E5; }
.card-middle { display: flex; gap: 12rpx; margin-top: 16rpx; flex-wrap: wrap; }
.info-pill { display: flex; align-items: center; gap: 8rpx; padding: 10rpx 16rpx; border-radius: 14rpx; }
.info-pill-member { background: #EEF2FF; }
.info-pill-balance { background: #ECFDF5; }
.info-pill-muted { background: #F3F4F6; }
.info-pill-label { font-size: 22rpx; color: #6B7280; }
.info-pill-value { font-size: 24rpx; color: #1F2937; font-weight: 600; }
.card-bottom { display: flex; justify-content: space-between; align-items: center; gap: 12rpx; margin-top: 16rpx; padding-top: 16rpx; border-top: 1rpx solid #F3F4F6; }
.spent { font-size: 26rpx; color: #374151; }
.customer-tags { display: flex; flex-wrap: wrap; justify-content: flex-end; gap: 8rpx; }
.customer-tag { font-size: 20rpx; padding: 6rpx 12rpx; border-radius: 999rpx; }
.pet-list { margin-top: 14rpx; display: flex; flex-direction: column; gap: 10rpx; }
.pet-item { background: #F8FAFC; border-radius: 12rpx; padding: 12rpx 16rpx; border: 1rpx solid #E2E8F0; }
.pet-item-top { display: flex; align-items: baseline; gap: 10rpx; }
.pet-item-name { font-size: 26rpx; font-weight: 600; color: #1F2937; }
.pet-item-meta { font-size: 22rpx; color: #6B7280; }
.pet-item-tags { display: flex; flex-wrap: wrap; gap: 8rpx; margin-top: 8rpx; }
.pet-item-tag { font-size: 20rpx; padding: 4rpx 12rpx; border-radius: 999rpx; line-height: 1.3; }
.load-more { text-align: center; padding: 32rpx 0; font-size: 24rpx; color: #9CA3AF; }
.load-more-done { color: #D1D5DB; font-size: 22rpx; }
</style>
