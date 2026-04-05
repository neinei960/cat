<template>
  <SideLayout>
    <view class="page">
      <view class="section-card">
        <text class="section-title">价格规则</text>
        <view class="field-grid">
          <view class="field-card">
            <text class="field-label">基础上门费 / 次</text>
            <input v-model="pricing.base_visit_price" class="input" type="digit" />
          </view>
          <view class="field-card">
            <text class="field-label">加猫费 / 只 / 次</text>
            <input v-model="pricing.extra_pet_price" class="input" type="digit" />
          </view>
          <view class="field-card full">
            <text class="field-label">节假日加价 / 次</text>
            <input v-model="pricing.holiday_surcharge" class="input" type="digit" />
          </view>
        </view>
        <view class="submit-row">
          <view class="btn btn-primary" @click="savePricing">保存价格规则</view>
        </view>
      </view>

      <view class="section-card">
        <view class="section-head">
          <text class="section-title">服务内容模板</text>
          <view class="btn" @click="addItem">新增模板</view>
        </view>
        <view class="item-list">
          <view class="item-card" v-for="(item, index) in items" :key="`${item.code}-${index}`">
            <view class="field-grid">
              <view class="field-card">
                <text class="field-label">编码</text>
                <input v-model="item.code" class="input" placeholder="例如：feed" />
              </view>
              <view class="field-card">
                <text class="field-label">名称</text>
                <input v-model="item.name" class="input" placeholder="例如：添粮" />
              </view>
              <view class="field-card">
                <text class="field-label">加收金额</text>
                <input v-model="item.extra_price" class="input" type="digit" />
              </view>
              <view class="field-card">
                <text class="field-label">操作</text>
                <view class="btn danger" @click="items.splice(index, 1)">删除</view>
              </view>
            </view>
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
  base_visit_price: 50,
  extra_pet_price: 20,
  holiday_surcharge: 20,
})
const items = ref<FeedingItemTemplate[]>([])

function addItem() {
  items.value.push({ code: '', name: '', extra_price: 0 })
}

async function loadData() {
  const res = await getFeedingSettings()
  pricing.value = { ...res.data.pricing }
  items.value = (res.data.items || []).map(item => ({ ...item }))
}

async function savePricing() {
  await updateFeedingPricing({
    base_visit_price: Number(pricing.value.base_visit_price || 0),
    extra_pet_price: Number(pricing.value.extra_pet_price || 0),
    holiday_surcharge: Number(pricing.value.holiday_surcharge || 0),
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
.field-grid { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 16rpx; }
.field-card.full { grid-column: 1 / -1; }
.field-label { display: block; margin-bottom: 8rpx; font-size: 22rpx; color: #6B7280; }
.input { width: 100%; min-height: 88rpx; padding: 22rpx 24rpx; background: #F8FAFC; border-radius: 18rpx; font-size: 26rpx; color: #111827; box-sizing: border-box; }
.submit-row { display: flex; justify-content: flex-end; margin-top: 18rpx; }
.btn { padding: 14rpx 22rpx; border-radius: 16rpx; background: #F8FAFC; color: #374151; font-size: 24rpx; border: 1rpx solid #E5E7EB; }
.btn-primary { background: linear-gradient(135deg, #4F46E5, #6366F1); color: #fff; border-color: transparent; }
.btn.danger { color: #DC2626; background: #FEF2F2; border-color: #FECACA; text-align: center; }
.item-list { display: flex; flex-direction: column; gap: 16rpx; }
.item-card { padding: 18rpx; border-radius: 18rpx; background: #F8FAFC; }
@media (max-width: 768px) {
  .field-grid { grid-template-columns: 1fr; }
}
</style>
