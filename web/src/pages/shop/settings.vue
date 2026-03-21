<template>
  <SideLayout>
  <view class="page">
    <view v-if="!shop" class="loading">加载中...</view>
    <view v-else class="form">
      <view class="form-item">
        <text class="label">店铺名称</text>
        <input v-model="shop.name" placeholder="请输入店铺名称" class="input" />
      </view>
      <view class="form-item">
        <text class="label">联系电话</text>
        <input v-model="shop.phone" type="number" placeholder="请输入电话" class="input" />
      </view>
      <view class="form-item">
        <text class="label">地址</text>
        <input v-model="shop.address" placeholder="请输入地址" class="input" />
      </view>
    </view>

    <button class="btn-submit" v-if="shop" @click="onSave" :loading="saving">保存</button>
  </view>
  </SideLayout>
</template>

<script setup lang="ts">
import SideLayout from '@/components/SideLayout.vue'
import { ref, onMounted } from 'vue'
import { getShop, updateShop } from '@/api/shop'

const shop = ref<Shop | null>(null)
const saving = ref(false)

onMounted(async () => {
  const res = await getShop()
  shop.value = res.data
})

async function onSave() {
  if (!shop.value) return
  saving.value = true
  try {
    await updateShop({ name: shop.value.name, phone: shop.value.phone, address: shop.value.address })
    uni.showToast({ title: '保存成功', icon: 'success' })
  } finally { saving.value = false }
}
</script>

<style scoped>
.page { padding: 24rpx; }
.loading { text-align: center; padding: 100rpx 0; color: #9CA3AF; }
.form { background: #fff; border-radius: 16rpx; padding: 8rpx 24rpx; margin-bottom: 32rpx; }
.form-item { padding: 24rpx 0; border-bottom: 1rpx solid #F3F4F6; }
.form-item:last-child { border-bottom: none; }
.label { font-size: 28rpx; color: #374151; display: block; margin-bottom: 12rpx; }
.input { font-size: 28rpx; color: #1F2937; height: 60rpx; }
.btn-submit { background: #4F46E5; color: #fff; border-radius: 12rpx; font-size: 30rpx; margin-top: 16rpx; }
</style>
