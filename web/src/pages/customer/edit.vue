<template>
  <view class="page">
    <view class="form">
      <view class="form-item">
        <text class="label">昵称 *</text>
        <input v-model="form.nickname" placeholder="请输入客户昵称" class="input" />
      </view>
      <view class="form-item">
        <text class="label">手机号</text>
        <input v-model="form.phone" type="number" placeholder="请输入手机号" class="input" />
      </view>
      <view class="form-item">
        <text class="label">性别</text>
        <picker :range="genders" :value="form.gender" @change="(e: any) => form.gender = e.detail.value">
          <view class="picker">{{ genders[form.gender] }}</view>
        </picker>
      </view>
      <view class="form-item">
        <text class="label">备注</text>
        <textarea v-model="form.remark" placeholder="客户备注" class="textarea" />
      </view>
      <view class="form-item">
        <text class="label">标签</text>
        <input v-model="form.tags" placeholder="逗号分隔，如：VIP,敏感犬" class="input" />
      </view>
    </view>

    <button class="btn-submit" @click="onSubmit" :loading="submitting">{{ id ? '保存' : '新增' }}</button>
  </view>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { onLoad } from '@dcloudio/uni-app'
import { getCustomer, createCustomer, updateCustomer } from '@/api/customer'
import { safeBack } from '@/utils/navigate'

const id = ref(0)
const submitting = ref(false)
const genders = ['未知', '男', '女']
const form = ref({ nickname: '', phone: '', gender: 0, remark: '', tags: '' })

onLoad((query) => {
  if (query?.id) {
    id.value = parseInt(query.id)
    loadData()
  }
})

async function loadData() {
  const res = await getCustomer(id.value)
  form.value = {
    nickname: res.data.nickname, phone: res.data.phone,
    gender: res.data.gender, remark: res.data.remark, tags: res.data.tags,
  }
}

async function onSubmit() {
  if (!form.value.nickname) { uni.showToast({ title: '请填写昵称', icon: 'none' }); return }
  submitting.value = true
  try {
    if (id.value) { await updateCustomer(id.value, form.value) }
    else { await createCustomer(form.value) }
    uni.showToast({ title: '保存成功', icon: 'success' })
    setTimeout(() => safeBack(), 500)
  } finally { submitting.value = false }
}
</script>

<style scoped>
.page { padding: 24rpx; }
.form { background: #fff; border-radius: 16rpx; padding: 8rpx 24rpx; margin-bottom: 32rpx; }
.form-item { padding: 24rpx 0; border-bottom: 1rpx solid #F3F4F6; }
.form-item:last-child { border-bottom: none; }
.label { font-size: 28rpx; color: #374151; display: block; margin-bottom: 12rpx; }
.input { font-size: 28rpx; color: #1F2937; height: 60rpx; }
.textarea { font-size: 28rpx; color: #1F2937; width: 100%; height: 160rpx; }
.picker { font-size: 28rpx; color: #1F2937; height: 60rpx; line-height: 60rpx; }
.btn-submit { background: #4F46E5; color: #fff; border-radius: 12rpx; font-size: 30rpx; margin-top: 16rpx; }
</style>
