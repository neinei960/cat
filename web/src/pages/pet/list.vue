<template>
  <SideLayout>
  <view class="page">
    <view class="header">
      <text class="title">宠物管理</text>
      <view class="btn-add" @click="goAdd">+ 新增</view>
    </view>

    <view class="search-bar">
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

    <view v-if="loading" class="loading">加载中...</view>
    <view v-else-if="groupedList.length === 0" class="empty">暂无宠物</view>

    <view v-else class="list">
      <view v-for="group in groupedList" :key="group.key" class="owner-group">
        <view class="group-header" v-if="group.ownerName">
          <view class="group-left">
            <text class="group-label group-label-link" @click.stop="goCustomer(group.key)">👤 {{ group.ownerName }}</text>
            <view class="owner-info" v-if="group.customer">
              <text class="owner-tag" v-if="group.customer.member_card">{{ group.customer.member_card.card_name }}</text>
              <text class="owner-tag owner-tag-balance" v-if="group.customer.member_balance">余额¥{{ group.customer.member_balance }}</text>
              <text class="owner-tag owner-tag-visit" v-if="group.customer.last_visit_at">上次:{{ formatDate(group.customer.last_visit_at) }}</text>
            </view>
          </view>
          <text class="group-count">{{ group.pets.length }}只</text>
        </view>
        <view class="group-header" v-else>
          <text class="group-label" style="color: #9CA3AF;">暂无主人</text>
        </view>
        <view class="card" v-for="pet in group.pets" :key="pet.ID" @click="goEdit(pet.ID)">
          <view class="card-top">
            <view class="avatar">{{ pet.species === '犬' ? '🐕' : '🐱' }}</view>
            <view class="info">
              <text class="name">{{ pet.name }}</text>
              <text class="breed">{{ pet.breed || '未知品种' }} · {{ pet.gender === 1 ? '♂弟弟' : pet.gender === 2 ? '♀妹妹' : '未知' }}<text v-if="pet.birth_date"> · {{ calcAge(pet.birth_date) }}</text></text>
            </view>
            <text class="weight" v-if="pet.weight">{{ pet.weight }}kg</text>
          </view>
          <view class="card-meta">
            <view class="tags">
              <text class="tag" v-if="pet.fur_level">{{ pet.fur_level }}</text>
              <text class="tag" v-if="pet.neutered">已绝育</text>
              <text class="tag" v-if="pet.personality" :style="{ background: getPersonalityBg(pet.personality), color: getPersonalityColor(pet.personality) }">{{ pet.personality }}</text>
              <text class="tag" v-if="pet.aggression && pet.aggression !== '无'" :style="{ background: '#FEE2E2', color: '#EF4444' }">攻击性:{{ pet.aggression }}</text>
            </view>
          </view>
          <view class="alerts" v-if="pet.care_notes">
            <text class="alert-text">⚠ {{ pet.care_notes }}</text>
          </view>
        </view>
      </view>
    </view>
  </view>
  </SideLayout>
</template>

<script setup lang="ts">
import SideLayout from '@/components/SideLayout.vue'
import { ref, computed, onMounted } from 'vue'
import { onShow } from '@dcloudio/uni-app'
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

interface OwnerGroup {
  key: string
  ownerName: string
  customer?: Customer
  pets: Pet[]
}

const list = ref<Pet[]>([])
const loading = ref(true)
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
  try {
    const params: any = { page: 1, page_size: 200 }
    if (keyword.value.trim()) params.keyword = keyword.value.trim()
    const res = await getPetList(params)
    list.value = res.data.list || []
  } finally { loading.value = false }
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
</script>

<style scoped>
.page { padding: 24rpx; }
.header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 20rpx; }
.title { font-size: 36rpx; font-weight: bold; color: #1F2937; }
.btn-add { font-size: 28rpx; color: #fff; background: #4F46E5; padding: 12rpx 24rpx; border-radius: 12rpx; }
.search-bar { position: relative; margin-bottom: 20rpx; }
.search-input { background: #fff; border-radius: 12rpx; padding: 16rpx 60rpx 16rpx 24rpx; font-size: 26rpx; color: #1F2937; box-shadow: 0 2rpx 8rpx rgba(0,0,0,0.04); }
.search-clear { position: absolute; right: 20rpx; top: 50%; transform: translateY(-50%); font-size: 28rpx; color: #9CA3AF; padding: 8rpx; }
.loading, .empty { text-align: center; padding: 100rpx 0; color: #9CA3AF; font-size: 28rpx; }
.owner-group { margin-bottom: 8rpx; }
.group-header { display: flex; justify-content: space-between; align-items: flex-start; padding: 16rpx 8rpx 8rpx; }
.group-left { flex: 1; }
.group-label { font-size: 24rpx; font-weight: 600; color: #6B7280; }
.group-label-link { color: #4F46E5; text-decoration: underline; }
.owner-info { display: flex; gap: 8rpx; flex-wrap: wrap; margin-top: 6rpx; margin-left: 4rpx; }
.owner-tag { font-size: 20rpx; padding: 2rpx 10rpx; border-radius: 8rpx; background: #EEF2FF; color: #4F46E5; }
.owner-tag-balance { background: #FEF3C7; color: #92400E; }
.owner-tag-visit { background: #F3F4F6; color: #6B7280; }
.group-count { font-size: 22rpx; color: #9CA3AF; margin-top: 4rpx; }
.card { background: #fff; border-radius: 16rpx; padding: 20rpx 24rpx; margin-bottom: 12rpx; box-shadow: 0 2rpx 8rpx rgba(0,0,0,0.04); }
.card-top { display: flex; align-items: center; }
.avatar { font-size: 44rpx; width: 64rpx; text-align: center; }
.info { flex: 1; margin-left: 12rpx; }
.name { font-size: 30rpx; font-weight: 600; color: #1F2937; display: block; }
.breed { font-size: 24rpx; color: #6B7280; display: block; margin-top: 2rpx; }
.weight { font-size: 26rpx; color: #4F46E5; font-weight: 600; }
.card-meta { display: flex; justify-content: space-between; align-items: center; margin-top: 12rpx; }
.tags { display: flex; gap: 8rpx; flex-wrap: wrap; }
.tag { font-size: 22rpx; padding: 4rpx 12rpx; background: #EEF2FF; color: #4F46E5; border-radius: 12rpx; }
.alerts { margin-top: 10rpx; padding: 10rpx; background: #FEF3C7; border-radius: 8rpx; }
.alert-text { font-size: 24rpx; color: #92400E; }
</style>
