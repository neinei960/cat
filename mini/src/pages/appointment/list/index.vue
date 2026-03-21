<template>
  <view class="page">
    <view class="tabs">
      <view :class="['tab', tab === 'upcoming' ? 'active' : '']" @click="tab = 'upcoming'">即将到来</view>
      <view :class="['tab', tab === 'history' ? 'active' : '']" @click="tab = 'history'">历史记录</view>
    </view>
    <view v-if="loading" class="loading">加载中...</view>
    <view v-else-if="filteredList.length === 0" class="empty">暂无预约</view>
    <view v-else class="list">
      <view class="card" v-for="a in filteredList" :key="a.ID" @click="goDetail(a.ID)">
        <view class="card-top">
          <text class="date">{{ a.date }} {{ a.start_time }}</text>
          <text :class="['status', `s${a.status}`]">{{ statusMap[a.status] }}</text>
        </view>
        <text class="pet">{{ a.pet?.name || '-' }}</text>
        <text class="services">{{ (a.services||[]).map((s:any)=>s.service_name).join('+') }}</text>
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { onShow } from '@dcloudio/uni-app'
import { getAppointments } from '../../../api'

const tab = ref('upcoming')
const list = ref<any[]>([])
const loading = ref(true)
const statusMap: Record<number,string> = {0:'待确认',1:'已确认',2:'进行中',3:'已完成',4:'已取消',5:'未到店'}

const filteredList = computed(() => {
  if (tab.value === 'upcoming') return list.value.filter(a => a.status <= 2)
  return list.value.filter(a => a.status >= 3)
})

async function load() {
  loading.value = true
  try { const res = await getAppointments(1, 50); list.value = res.data.list || [] }
  finally { loading.value = false }
}
function goDetail(id: number) { uni.navigateTo({ url: `/pages/appointment/detail/index?id=${id}` }) }
onMounted(load); onShow(load)
</script>

<style scoped>
.page { padding: 24rpx; }
.tabs { display: flex; gap: 16rpx; margin-bottom: 24rpx; }
.tab { flex: 1; text-align: center; padding: 16rpx; background: #F3F4F6; color: #6B7280; border-radius: 12rpx; font-size: 28rpx; }
.tab.active { background: #4F46E5; color: #fff; }
.loading,.empty { text-align: center; padding: 100rpx 0; color: #9CA3AF; font-size: 28rpx; }
.card { background: #fff; border-radius: 16rpx; padding: 24rpx; margin-bottom: 16rpx; }
.card-top { display: flex; justify-content: space-between; margin-bottom: 8rpx; }
.date { font-size: 28rpx; font-weight: 600; color: #1F2937; }
.status { font-size: 22rpx; padding: 4rpx 12rpx; border-radius: 12rpx; }
.s0{background:#FEF3C7;color:#92400E;}.s1{background:#EEF2FF;color:#4F46E5;}.s2{background:#D1FAE5;color:#059669;}.s3{background:#F3F4F6;color:#6B7280;}.s4,.s5{background:#FEE2E2;color:#DC2626;}
.pet { font-size: 28rpx; color: #374151; display: block; }
.services { font-size: 24rpx; color: #6B7280; display: block; margin-top: 4rpx; }
</style>
