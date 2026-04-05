<template>
  <SideLayout>
    <view class="page">
      <view class="header">
        <view>
          <text class="title">寄养房型设置</text>
          <text class="subtitle">按房型录入库存。开单时系统会按同时段已入住和待入住数量，自动扣减可售间数。</text>
        </view>
        <view class="btn btn-primary" @click="startCreate">+ 添加房型</view>
      </view>

      <view v-if="editing" class="form-card">
        <text class="section-title">{{ form.id ? '编辑寄养房型' : '新增寄养房型' }}</text>
        <text class="section-tip">例子：阳光单间 6 间 / 每间最多住 1 只 / 168 元每晚。</text>

        <view class="field">
          <text class="field-label">房型名称</text>
          <text class="field-tip">前台开寄养单时看到的选项名称。</text>
          <input v-model="form.cabinet_type" class="input" placeholder="例如：康娜温柔乡、阳光单间" />
        </view>

        <view class="field">
          <text class="field-label">总间数</text>
          <text class="field-tip">这个房型一共有多少间。有人住进来后，对应时段可售间数会自动减 1。</text>
          <input v-model="form.room_count" class="input" type="number" placeholder="例如：6" />
        </view>

        <view class="field">
          <text class="field-label">每间可住猫数</text>
          <text class="field-tip">单间最多能住几只猫，用来限制同柜多猫。</text>
          <input v-model="form.capacity" class="input" type="number" placeholder="例如：1、2" />
        </view>

        <view class="field">
          <text class="field-label">每晚价格</text>
          <text class="field-tip">普通日每间每晚的基础寄养价格。</text>
          <input v-model="form.base_price" class="input input-amount" type="digit" placeholder="例如：168" />
        </view>

        <view class="field">
          <text class="field-label">当前状态</text>
          <text class="field-tip">启用：可售；清洁中：暂不安排；停用：下架。</text>
          <picker :range="statusOptions" :value="statusIndex" @change="onStatusChange">
            <view class="picker">{{ statusOptions[statusIndex] }}</view>
          </picker>
        </view>

        <view class="field">
          <text class="field-label">备注</text>
          <text class="field-tip">可写房间位置、采光、适合什么猫等说明。</text>
          <textarea v-model="form.remark" class="textarea" placeholder="例如：靠窗、适合胆大猫咪、带监控" />
        </view>

        <view class="actions">
          <view class="btn" @click="cancelEdit">取消</view>
          <view class="btn btn-primary" @click="save">保存</view>
        </view>
      </view>

      <view class="group-list">
        <view class="group-card" v-for="item in cabinets" :key="item.ID">
          <view class="group-head">
            <view>
              <text class="group-title">{{ item.cabinet_type }}</text>
              <text class="group-meta">共 {{ item.room_count || 1 }} 间 · 每间可住 {{ item.capacity }} 只 · ¥{{ item.base_price }}/晚</text>
            </view>
            <view class="row-action" @click="edit(item)">编辑</view>
          </view>
          <text class="group-tip">{{ stateLabel(item.status) }}{{ item.remark ? ` · ${item.remark}` : '' }}</text>
        </view>
      </view>
    </view>
  </SideLayout>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { onShow } from '@dcloudio/uni-app'
import SideLayout from '@/components/SideLayout.vue'
import { createBoardingCabinet, getBoardingCabinets, updateBoardingCabinet } from '@/api/boarding'

const cabinets = ref<BoardingCabinet[]>([])
const editing = ref(false)
const form = ref<any>({ id: 0, cabinet_type: '', room_count: '1', capacity: '1', base_price: '0', status: 'enabled', remark: '' })
const statusOptions = ['启用', '清洁中', '停用']
const statusValues = ['enabled', 'cleaning', 'disabled']
const statusIndex = computed(() => Math.max(statusValues.indexOf(form.value.status), 0))

function stateLabel(status: string) {
  return { enabled: '启用', cleaning: '清洁中', disabled: '停用' }[status] || status
}

function onStatusChange(e: any) {
  form.value.status = statusValues[e.detail.value]
}

function startCreate() {
  editing.value = true
  form.value = { id: 0, cabinet_type: '', room_count: '1', capacity: '1', base_price: '0', status: 'enabled', remark: '' }
}

function edit(item: BoardingCabinet) {
  editing.value = true
  form.value = {
    id: item.ID,
    cabinet_type: item.cabinet_type,
    room_count: String(item.room_count || 1),
    capacity: String(item.capacity),
    base_price: String(item.base_price),
    status: item.status,
    remark: item.remark || '',
  }
}

function cancelEdit() {
  editing.value = false
}

async function loadData() {
  const res = await getBoardingCabinets()
  cabinets.value = res.data || []
}

async function save() {
  const payload = {
    cabinet_type: form.value.cabinet_type.trim(),
    room_count: Number(form.value.room_count) || 1,
    capacity: Number(form.value.capacity) || 1,
    base_price: Number(form.value.base_price) || 0,
    status: form.value.status,
    remark: form.value.remark,
  }
  if (!payload.cabinet_type) {
    uni.showToast({ title: '请填写房型名称', icon: 'none' })
    return
  }
  if (form.value.id) await updateBoardingCabinet(form.value.id, payload)
  else await createBoardingCabinet(payload)
  uni.showToast({ title: '保存成功', icon: 'success' })
  editing.value = false
  await loadData()
}

onShow(loadData)
</script>

<style scoped>
.page { padding: 24rpx; display: flex; flex-direction: column; gap: 20rpx; }
.header { display: flex; justify-content: space-between; align-items: flex-start; gap: 12rpx; }
.title { font-size: 34rpx; font-weight: 700; color: #111827; }
.subtitle { display: block; margin-top: 8rpx; font-size: 22rpx; color: #6B7280; line-height: 1.6; max-width: 520rpx; }
.btn { padding: 14rpx 24rpx; border-radius: 12rpx; background: #F3F4F6; color: #374151; font-size: 24rpx; }
.btn-primary { background: #4F46E5; color: #fff; }
.form-card, .group-card { background: #fff; border-radius: 18rpx; padding: 24rpx; box-shadow: 0 12rpx 28rpx rgba(15, 23, 42, 0.04); }
.section-title, .group-title { display: block; font-size: 28rpx; font-weight: 700; color: #111827; margin-bottom: 14rpx; }
.section-tip, .group-tip { display: block; font-size: 22rpx; color: #6B7280; line-height: 1.6; }
.section-tip { margin-bottom: 18rpx; }
.field { margin-bottom: 18rpx; }
.field-label { display: block; font-size: 24rpx; font-weight: 600; color: #1F2937; margin-bottom: 6rpx; }
.field-tip { display: block; font-size: 20rpx; color: #9CA3AF; margin-bottom: 10rpx; line-height: 1.5; }
.input, .textarea, .picker { width: 100%; box-sizing: border-box; margin-bottom: 14rpx; background: #F9FAFB; border: 1rpx solid #E5E7EB; border-radius: 12rpx; padding: 18rpx 20rpx; font-size: 26rpx; color: #111827; min-height: 60rpx; }
.textarea { min-height: 120rpx; }
.actions { display: flex; gap: 12rpx; }
.group-list { display: flex; flex-direction: column; gap: 18rpx; }
.group-head { display: flex; justify-content: space-between; gap: 12rpx; align-items: flex-start; margin-bottom: 8rpx; }
.group-meta { display: block; font-size: 22rpx; color: #6B7280; }
.row-action { font-size: 24rpx; color: #4F46E5; }
</style>
