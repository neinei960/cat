<template>
  <SideLayout>
  <view class="page">
    <view class="editor-shell">
      <view class="action-bar">
        <text class="action-link" @click="goBack">取消</text>
        <view :class="['action-save', (!canSubmit || submitting) ? 'disabled' : '']" @click="onSubmit">
          {{ submitting ? '保存中' : '保存' }}
        </view>
      </view>

      <view class="page-head">
        <text class="page-title">{{ id ? '编辑客户资料' : '新增客户资料' }}</text>
      </view>

      <view class="field-group">
        <text class="group-label">备注名</text>
        <view class="field-card">
          <input v-model="form.nickname" placeholder="填写客户称呼" class="field-input" />
        </view>
      </view>

      <view class="field-group">
        <text class="group-label">电话</text>
        <view class="field-card field-card-phone">
          <text class="field-icon">⊖</text>
          <input v-model="form.phone" type="number" placeholder="填写手机号" class="field-input field-input-phone" />
        </view>
      </view>

      <view class="field-group">
        <text class="group-label">标签</text>
        <view class="field-card field-card-arrow" @click="goTagManage">
          <text :class="['field-summary', selectedTags.length ? 'filled' : '']">{{ selectedTagSummary }}</text>
          <text class="field-arrow">›</text>
        </view>
        <scroll-view v-if="availableTags.length" scroll-x class="tag-scroll" show-scrollbar="false">
          <view class="tag-picker tag-picker-inline">
            <text
              v-for="tag in availableTags"
              :key="tag.ID"
              :class="['tag-chip', form.customer_tag_ids.includes(tag.ID) ? 'tag-chip-selected' : '']"
              :style="form.customer_tag_ids.includes(tag.ID)
                ? { background: withAlpha(tag.color, 0.14), color: tag.color, borderColor: withAlpha(tag.color, 0.28) }
                : { color: '#6B7280', background: '#F7F8FA', borderColor: '#ECEEF2' }"
              @click="toggleTag(tag.ID)"
            >
              {{ tag.name }}
            </text>
          </view>
        </scroll-view>
        <view v-else class="tag-empty-block">
          <text class="tag-empty">还没有客户标签，先去新建几个</text>
          <view class="tag-empty-btn" @click="goTagManage">管理标签</view>
        </view>
      </view>

      <view class="field-group">
        <text class="group-label">备注</text>
        <view class="field-card field-card-textarea">
          <!-- #ifdef H5 -->
          <div
            ref="remarkEditorRef"
            class="remark-editor"
            contenteditable="plaintext-only"
            :data-placeholder="form.remark ? '' : '添加文字'"
            @input="onRemarkInput"
            @blur="syncRemarkEditor"
          />
          <!-- #endif -->
          <!-- #ifndef H5 -->
          <textarea v-model="form.remark" placeholder="添加文字" class="field-textarea" :auto-height="false" />
          <!-- #endif -->
        </view>
      </view>

      <view class="field-group">
        <text class="group-label">性别</text>
        <view class="field-card">
          <picker :range="genders" :value="form.gender" @change="(e: any) => form.gender = e.detail.value" class="field-picker">
            <view class="picker-line">
              <text :class="['field-summary', 'filled']">{{ genders[form.gender] }}</text>
              <text class="field-arrow">›</text>
            </view>
          </picker>
        </view>
      </view>

      <view class="field-group">
        <text class="group-label">上门地址</text>
        <view class="field-card">
          <input v-model="form.address" placeholder="填写地址（与喂养地址互通）" class="field-input" />
        </view>
      </view>

      <view class="field-group">
        <text class="group-label">入户密码</text>
        <view class="field-card">
          <input v-model="form.door_code" placeholder="门锁 / 门卡" class="field-input" />
        </view>
      </view>

      <view class="field-group">
        <text class="group-label">补充信息</text>
        <view class="field-card">
          <input v-model="form.address_detail" placeholder="楼栋 / 停车 / 备注" class="field-input" />
        </view>
      </view>
    </view>
  </view>
  </SideLayout>
</template>

<script setup lang="ts">
import { computed, nextTick, ref, watch } from 'vue'
import { onLoad, onShow } from '@dcloudio/uni-app'
import SideLayout from '@/components/SideLayout.vue'
import { getCustomer, createCustomer, updateCustomer } from '@/api/customer'
import { getCustomerTags } from '@/api/customer-tag'
import { safeBack } from '@/utils/navigate'

const id = ref(0)
const submitting = ref(false)
const genders = ['未知', '男', '女']
const availableTags = ref<CustomerTag[]>([])
const form = ref({ nickname: '', phone: '', gender: 0, remark: '', address: '', address_detail: '', door_code: '', customer_tag_ids: [] as number[] })
const selectedTags = computed(() => availableTags.value.filter(tag => form.value.customer_tag_ids.includes(tag.ID)))
const selectedTagSummary = computed(() => selectedTags.value.length ? selectedTags.value.map(tag => tag.name).join('、') : '添加标签')
const canSubmit = computed(() => form.value.nickname.trim().length > 0)
const remarkEditorRef = ref<HTMLElement | null>(null)

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
    address: res.data.address || '', address_detail: res.data.address_detail || '', door_code: res.data.door_code || '',
    customer_tag_ids: (res.data.customer_tags || []).map(tag => tag.ID),
  }
  syncRemarkEditor()
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

function goBack() {
  safeBack()
}

function normalizeRemarkText(value: string) {
  return value.replace(/\r/g, '').replace(/\n{3,}/g, '\n\n')
}

function onRemarkInput(event: Event) {
  const target = event.target as HTMLElement | null
  if (!target) return
  form.value.remark = normalizeRemarkText(target.innerText || '')
}

function syncRemarkEditor() {
  // #ifdef H5
  nextTick(() => {
    if (!remarkEditorRef.value) return
    const normalized = normalizeRemarkText(form.value.remark || '')
    if (remarkEditorRef.value.innerText !== normalized) {
      remarkEditorRef.value.innerText = normalized
    }
  })
  // #endif
}

watch(() => form.value.remark, () => {
  syncRemarkEditor()
})

async function onSubmit() {
  if (!canSubmit.value || submitting.value) return
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
.page {
  min-height: 100vh;
  background: #FFFFFF;
  padding: 24rpx 28rpx 80rpx;
}
.editor-shell {
  max-width: 820rpx;
  margin: 0 auto;
}
.action-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 34rpx;
}
.action-link {
  font-size: 28rpx;
  color: #2F3640;
}
.action-save {
  min-width: 100rpx;
  height: 56rpx;
  padding: 0 20rpx;
  border-radius: 16rpx;
  background: #4F46E5;
  color: #FFFFFF;
  font-size: 26rpx;
  font-weight: 600;
  display: flex;
  align-items: center;
  justify-content: center;
}
.action-save.disabled {
  background: #EEF1F5;
  color: #B5BCC8;
}
.page-head {
  text-align: center;
  margin-bottom: 34rpx;
}
.page-title {
  font-size: 42rpx;
  font-weight: 600;
  color: #1F2937;
}
.field-group {
  margin-bottom: 26rpx;
}
.group-label {
  display: block;
  font-size: 22rpx;
  color: #6B7280;
  margin-bottom: 10rpx;
}
.field-card {
  background: #F7F8FA;
  border-radius: 16rpx;
  padding: 0 20rpx;
  min-height: 82rpx;
  display: flex;
  align-items: center;
}
.field-card-phone {
  gap: 14rpx;
}
.field-card-arrow {
  justify-content: space-between;
}
.field-card-textarea {
  align-items: stretch;
  padding-top: 10rpx;
  padding-bottom: 10rpx;
}
.field-icon {
  font-size: 28rpx;
  color: #E86C79;
  line-height: 1;
}
.field-input,
.field-input-phone {
  flex: 1;
  min-width: 0;
  height: 82rpx;
  font-size: 28rpx;
  color: #1F2937;
}
.field-textarea {
  font-size: 28rpx;
  color: #1F2937;
  line-height: 1.5;
}
.remark-editor {
  width: 100%;
  min-height: 72rpx;
  max-height: 72rpx;
  overflow-y: auto;
  font-size: 28rpx;
  color: #1F2937;
  line-height: 36rpx;
  outline: none;
  white-space: pre-wrap;
  word-break: break-word;
}
.remark-editor:empty::before {
  content: attr(data-placeholder);
  color: #B3B9C5;
  pointer-events: none;
}
.field-summary {
  flex: 1;
  min-width: 0;
  font-size: 28rpx;
  color: #B3B9C5;
}
.field-summary.filled {
  color: #1F2937;
}
.field-arrow {
  margin-left: 16rpx;
  font-size: 30rpx;
  color: #B6BCC8;
  line-height: 1;
}
.picker-line {
  flex: 1;
  min-width: 0;
  height: 82rpx;
  display: flex;
  align-items: center;
  justify-content: space-between;
}
.field-picker {
  flex: 1;
}
.tag-scroll {
  margin-top: 10rpx;
  white-space: nowrap;
}
.tag-picker {
  display: inline-flex;
  gap: 10rpx;
  padding-right: 8rpx;
}
.tag-chip {
  padding: 8rpx 16rpx;
  border-radius: 999rpx;
  font-size: 22rpx;
  border: 1rpx solid transparent;
  white-space: nowrap;
}
.tag-chip-selected {
  font-weight: 600;
}
.tag-empty-block {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16rpx;
  background: #F7F8FA;
  border-radius: 16rpx;
  padding: 16rpx 18rpx;
  margin-top: 16rpx;
}
.tag-empty {
  font-size: 24rpx;
  color: #9CA3AF;
  display: block;
}
.tag-empty-btn {
  flex-shrink: 0;
  font-size: 22rpx;
  color: #4F46E5;
  background: #EEF2FF;
  padding: 10rpx 16rpx;
  border-radius: 999rpx;
}
</style>
