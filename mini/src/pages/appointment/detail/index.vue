<template>
  <view class="page" v-if="appt">
    <view :class="['status-bar', `s${appt.status}`]">{{ statusMap[appt.status] }}</view>
    <view class="card">
      <view class="row"><text class="label">日期</text><text>{{ appt.date }}</text></view>
      <view class="row"><text class="label">时间</text><text>{{ appt.start_time }}-{{ appt.end_time }}</text></view>
      <view class="row"><text class="label">宠物</text><text>{{ appt.pet?.name }}</text></view>
      <view class="row"><text class="label">洗护师</text><text>{{ appt.staff?.name || '待分配' }}</text></view>
      <view class="row"><text class="label">服务</text><text>{{ (appt.services||[]).map((s:any)=>s.service_name).join('+') }}</text></view>
      <view class="row"><text class="label">金额</text><text class="price">¥{{ appt.total_amount }}</text></view>
    </view>
    <button v-if="appt.status <= 1" class="btn-cancel" @click="doCancel">取消预约</button>
  </view>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { onLoad } from '@dcloudio/uni-app'
import { getAppointment, cancelAppointment } from '../../../api'
const appt = ref<any>(null)
const statusMap: Record<number,string> = {0:'待确认',1:'已确认',2:'进行中',3:'已完成',4:'已取消',5:'未到店'}
onLoad(async (q) => { if(q?.id) { const r = await getAppointment(parseInt(q.id)); appt.value = r.data }})
async function doCancel() {
  uni.showModal({ title:'确认取消', content:'确认取消预约？', success: async(r) => {
    if(r.confirm) { await cancelAppointment(appt.value.ID); uni.showToast({title:'已取消',icon:'success'}); const res = await getAppointment(appt.value.ID); appt.value = res.data }
  }})
}
</script>

<style scoped>
.page{padding:24rpx;}.status-bar{text-align:center;padding:24rpx;border-radius:16rpx;font-size:32rpx;font-weight:bold;margin-bottom:16rpx;}
.s0{background:#FEF3C7;color:#92400E;}.s1{background:#EEF2FF;color:#4F46E5;}.s2{background:#D1FAE5;color:#059669;}.s3{background:#F3F4F6;color:#6B7280;}.s4,.s5{background:#FEE2E2;color:#DC2626;}
.card{background:#fff;border-radius:16rpx;padding:24rpx;margin-bottom:24rpx;}.row{display:flex;justify-content:space-between;padding:12rpx 0;border-bottom:1rpx solid #F3F4F6;font-size:28rpx;}.row:last-child{border-bottom:none;}.label{color:#6B7280;}.price{color:#4F46E5;font-weight:bold;}
.btn-cancel{background:#fff;color:#DC2626;border:1rpx solid #DC2626;border-radius:12rpx;font-size:30rpx;}
</style>
