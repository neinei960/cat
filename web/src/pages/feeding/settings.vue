<template>
  <SideLayout>
    <view class="page">
      <view class="section-card">
        <text class="section-title">价格规则</text>
        <view class="field-grid">
          <view class="field-card">
            <text class="field-label">日常价格 / 天</text>
            <input v-model="pricing.base_day_price" class="input" type="digit" />
          </view>
          <view class="field-card">
            <text class="field-label">法定节假日价格 / 天</text>
            <input v-model="pricing.holiday_day_price" class="input" type="digit" />
          </view>
          <view class="field-card full">
            <text class="field-label">第 N 天起（日常） / 天</text>
            <input v-model="pricing.discount_day_price" class="input" type="digit" />
          </view>
          <view class="field-card full">
            <text class="field-label">第 N 天起（节假日） / 天</text>
            <input v-model="pricing.discount_holiday_price" class="input" type="digit" />
          </view>
          <view class="field-card full">
            <text class="field-label">优惠开始天数</text>
            <input v-model="pricing.discount_start_day" class="input" type="number" />
          </view>
        </view>
        <text class="helper-text">附加服务仍通过下方“服务内容模板”的加收金额配置，例如“超长侍玩喂玩服务 +20/天”。</text>
        <view class="submit-row">
          <view class="btn btn-primary" @click="savePricing">保存价格规则</view>
        </view>
      </view>

      <view class="section-card">
        <view class="section-head">
          <text class="section-title">服务内容模板</text>
          <view class="btn" @click="addItem">新增模板</view>
        </view>
        <view class="item-table">
          <view class="item-tr item-header">
            <view class="item-td name">名称</view>
            <view class="item-td price">加收</view>
            <view class="item-td action"></view>
          </view>
          <view class="item-tr" v-for="(item, index) in items" :key="`${item.code}-${index}`">
            <view class="item-td name"><input v-model="item.name" class="cell-input" placeholder="名称" /></view>
            <view class="item-td price"><input v-model="item.extra_price" class="cell-input" type="digit" placeholder="0" /></view>
            <view class="item-td action"><text class="del-btn" @click="items.splice(index, 1)">删除</text></view>
          </view>
        </view>
        <view class="submit-row">
          <view class="btn btn-primary" @click="saveItems">保存模板</view>
        </view>
      </view>
    </view>
  </SideLayout>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { onShow } from '@dcloudio/uni-app'
import SideLayout from '@/components/SideLayout.vue'
import { getFeedingSettings, updateFeedingItems, updateFeedingPricing } from '@/api/feeding'

const pricing = ref<FeedingPricingSetting>({
  base_day_price: 85,
  holiday_day_price: 95,
  discount_day_price: 68,
  discount_holiday_price: 90,
  discount_start_day: 3,
})
const items = ref<FeedingItemTemplate[]>([])

function addItem() {
  items.value.push({ code: `item_${Date.now()}`, name: '', extra_price: 0 })
}

async function loadData() {
  const res = await getFeedingSettings()
  pricing.value = { ...res.data.pricing }
  items.value = (res.data.items || []).map(item => ({ ...item }))
}

async function savePricing() {
  await updateFeedingPricing({
    base_day_price: Number(pricing.value.base_day_price || 0),
    holiday_day_price: Number(pricing.value.holiday_day_price || 0),
    discount_day_price: Number(pricing.value.discount_day_price || 0),
    discount_holiday_price: Number(pricing.value.discount_holiday_price || 0),
    discount_start_day: Number(pricing.value.discount_start_day || 0),
  })
  uni.showToast({ title: '已保存', icon: 'success' })
}

async function saveItems() {
  await updateFeedingItems(items.value.map(item => ({
    code: item.code.trim(),
    name: item.name.trim(),
    extra_price: Number(item.extra_price || 0),
  })))
  uni.showToast({ title: '已保存', icon: 'success' })
  await loadData()
}

onShow(loadData)
</script>

<style scoped>
.page { padding: 24rpx; }
.section-card { background: #fff; border-radius: 22rpx; box-shadow: 0 12rpx 28rpx rgba(15, 23, 42, 0.06); padding: 24rpx; margin-bottom: 18rpx; }
.section-head { display: flex; justify-content: space-between; gap: 12rpx; align-items: center; margin-bottom: 18rpx; }
.section-title { font-size: 28rpx; font-weight: 700; color: #111827; }
.field-grid { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 12rpx; }
.field-card { display: flex; align-items: center; gap: 12rpx; }
.field-card.full { grid-column: 1 / -1; }
.field-label { font-size: 22rpx; color: #6B7280; white-space: nowrap; flex-shrink: 0; }
.helper-text { display: block; margin-top: 14rpx; font-size: 22rpx; color: #6B7280; line-height: 1.7; }
.input { width: 100%; min-height: 68rpx; padding: 14rpx 18rpx; background: #F8FAFC; border-radius: 14rpx; font-size: 26rpx; color: #111827; box-sizing: border-box; flex: 1; min-width: 0; }
.submit-row { display: flex; justify-content: flex-end; margin-top: 18rpx; }
.btn { padding: 14rpx 22rpx; border-radius: 16rpx; background: #F8FAFC; color: #374151; font-size: 24rpx; border: 1rpx solid #E5E7EB; }
.btn-primary { background: linear-gradient(135deg, #4F46E5, #6366F1); color: #fff; border-color: transparent; }
.btn.danger { color: #DC2626; background: #FEF2F2; border-color: #FECACA; text-align: center; }
.item-table { border: 1rpx solid #E5E7EB; border-radius: 14rpx; overflow: hidden; }
.item-tr { display: flex; align-items: center; border-bottom: 1rpx solid #F3F4F6; }
.item-tr:last-child { border-bottom: none; }
.item-header { background: #F9FAFB; }
.item-header .item-td { font-size: 22rpx; color: #6B7280; font-weight: 600; padding: 14rpx 12rpx; }
.item-td { padding: 8rpx 10rpx; font-size: 24rpx; color: #374151; }
.item-td.code { flex: 2; }
.item-td.name { flex: 2; }
.item-td.price { flex: 1.5; }
.item-td.action { flex: 1; text-align: center; }
.cell-input { width: 100%; height: 60rpx; padding: 0 12rpx; background: #F8FAFC; border-radius: 10rpx; font-size: 24rpx; color: #111827; box-sizing: border-box; }
.del-btn { font-size: 22rpx; color: #DC2626; }
</style>
