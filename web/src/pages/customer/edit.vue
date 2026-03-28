<template>
  <SideLayout>
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
      <view class="form-item form-item-tags">
        <view class="label-row">
          <text class="label">客户标签</text>
          <text class="manage-link" @click="goTagManage">管理标签</text>
        </view>
        <view v-if="selectedTags.length" class="selected-tags">
          <text
            v-for="tag in selectedTags"
            :key="tag.ID"
            class="tag-chip tag-chip-selected"
            :style="{ background: withAlpha(tag.color, 0.14), color: tag.color, borderColor: withAlpha(tag.color, 0.28) }"
            @click="toggleTag(tag.ID)"
          >
            {{ tag.name }}
          </text>
        </view>
        <text v-else class="tag-empty">还没选标签，点下面快速关联</text>
        <view v-if="availableTags.length" class="tag-picker">
          <text
            v-for="tag in availableTags"
            :key="tag.ID"
            :class="['tag-chip', form.customer_tag_ids.includes(tag.ID) ? 'tag-chip-selected' : '']"
            :style="form.customer_tag_ids.includes(tag.ID)
              ? { background: withAlpha(tag.color, 0.14), color: tag.color, borderColor: withAlpha(tag.color, 0.28) }
              : { color: '#6B7280', background: '#F9FAFB', borderColor: '#E5E7EB' }"
            @click="toggleTag(tag.ID)"
          >
            {{ tag.name }}
          </text>
        </view>
        <view v-else class="tag-empty-block">
          <text class="tag-empty">还没有客户标签，先建几个再给客户打标</text>
          <view class="tag-empty-btn" @click="goTagManage">+ 新建标签</view>
        </view>
        <text class="tag-desc" v-if="selectedTags.length">{{ selectedTags.map(item => item.description).filter(Boolean).join(' · ') }}</text>
      </view>
    </view>

    <button class="btn-submit" @click="onSubmit" :loading="submitting">{{ id ? '保存' : '新增' }}</button>
  </view>
  </SideLayout>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { onLoad, onShow } from '@dcloudio/uni-app'
import SideLayout from '@/components/SideLayout.vue'
import { getCustomer, createCustomer, updateCustomer } from '@/api/customer'
import { getCustomerTags } from '@/api/customer-tag'
import { safeBack } from '@/utils/navigate'

const id = ref(0)
const submitting = ref(false)
const genders = ['未知', '男', '女']
const availableTags = ref<CustomerTag[]>([])
const form = ref({ nickname: '', phone: '', gender: 0, remark: '', customer_tag_ids: [] as number[] })
const selectedTags = computed(() => availableTags.value.filter(tag => form.value.customer_tag_ids.includes(tag.ID)))

onLoad((query) => {
  if (query?.id) {
    id.value = parseInt(query.id)
    loadData()
  }
})

onShow(() => {
  loadTags()
})

async function loadData() {
  const res = await getCustomer(id.value)
  form.value = {
    nickname: res.data.nickname, phone: res.data.phone,
    gender: res.data.gender, remark: res.data.remark,
    customer_tag_ids: (res.data.customer_tags || []).map(tag => tag.ID),
  }
}

async function loadTags() {
  try {
    const res = await getCustomerTags()
    availableTags.value = (res.data || []).filter(tag => tag.status === 1)
  } catch {
    availableTags.value = []
  }
}

function toggleTag(tagID: number) {
  const current = form.value.customer_tag_ids
  form.value.customer_tag_ids = current.includes(tagID)
    ? current.filter(id => id !== tagID)
    : [...current, tagID]
}

function withAlpha(color: string, alpha: number) {
  const hex = color.replace('#', '')
  if (hex.length !== 6) return color
  const value = Math.round(alpha * 255).toString(16).padStart(2, '0')
  return `#${hex}${value}`
}

function goTagManage() {
  uni.navigateTo({ url: '/pages/customer/tag-manage' })
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
.form-item-tags { padding-bottom: 12rpx; }
.label-row { display: flex; justify-content: space-between; align-items: center; margin-bottom: 12rpx; }
.label { font-size: 28rpx; color: #374151; display: block; margin-bottom: 12rpx; }
.manage-link { font-size: 24rpx; color: #4F46E5; }
.input { font-size: 28rpx; color: #1F2937; height: 60rpx; }
.textarea { font-size: 28rpx; color: #1F2937; width: 100%; height: 160rpx; }
.picker { font-size: 28rpx; color: #1F2937; height: 60rpx; line-height: 60rpx; }
.selected-tags, .tag-picker { display: flex; flex-wrap: wrap; gap: 12rpx; }
.tag-picker { margin-top: 12rpx; }
.tag-chip { padding: 10rpx 18rpx; border-radius: 999rpx; font-size: 24rpx; border: 1rpx solid transparent; }
.tag-chip-selected { font-weight: 600; }
.tag-empty-block { display: flex; align-items: center; justify-content: space-between; gap: 16rpx; background: #F9FAFB; border-radius: 14rpx; padding: 16rpx 18rpx; }
.tag-empty { font-size: 24rpx; color: #9CA3AF; display: block; }
.tag-empty-btn { flex-shrink: 0; font-size: 22rpx; color: #4F46E5; background: #EEF2FF; padding: 10rpx 16rpx; border-radius: 999rpx; }
.tag-desc { font-size: 22rpx; color: #6B7280; display: block; line-height: 1.6; margin-top: 14rpx; }
.btn-submit { background: #4F46E5; color: #fff; border-radius: 12rpx; font-size: 30rpx; margin-top: 16rpx; }
</style>
