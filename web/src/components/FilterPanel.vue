<template>
  <view v-if="visible" class="fp-overlay" @click.self="$emit('close')">
    <view class="fp-panel">
      <scroll-view scroll-y class="fp-body">
        <!-- 日期范围 -->
        <view class="fp-section">
          <view class="fp-date-row">
            <picker mode="date" :value="localFilter.dateFrom" @change="(e: any) => localFilter.dateFrom = e.detail.value">
              <view class="fp-date-btn">{{ localFilter.dateFrom || '开始日期' }}</view>
            </picker>
            <text class="fp-date-sep">至</text>
            <picker mode="date" :value="localFilter.dateTo" @change="(e: any) => localFilter.dateTo = e.detail.value">
              <view class="fp-date-btn">{{ localFilter.dateTo || '结束日期' }}</view>
            </picker>
          </view>
        </view>

        <!-- 状态 -->
        <view class="fp-section" v-if="statusOptions.length > 0">
          <text class="fp-label">{{ statusLabel }}</text>
          <view class="fp-chips">
            <view
              v-for="opt in statusOptions" :key="opt.value"
              :class="['fp-chip', localFilter.status === opt.value ? 'active' : '']"
              @click="localFilter.status = localFilter.status === opt.value ? -1 : opt.value"
            >{{ opt.label }}</view>
          </view>
        </view>

        <!-- 支付方式 (订单专用) -->
        <view class="fp-section" v-if="payMethods.length > 0">
          <text class="fp-label">支付方式</text>
          <view class="fp-chips">
            <view
              v-for="pm in payMethods" :key="pm.value"
              :class="['fp-chip', localFilter.payMethod === pm.value ? 'active' : '']"
              @click="localFilter.payMethod = localFilter.payMethod === pm.value ? '' : pm.value"
            >{{ pm.label }}</view>
          </view>
        </view>

        <!-- 洗护师 -->
        <view class="fp-section" v-if="staffList.length > 0">
          <text class="fp-label">洗护师</text>
          <view class="fp-chips">
            <view
              v-for="s in staffList" :key="s.ID"
              :class="['fp-chip', localFilter.staffId === s.ID ? 'active' : '']"
              @click="localFilter.staffId = localFilter.staffId === s.ID ? 0 : s.ID"
            >{{ s.name }}</view>
          </view>
        </view>

        <!-- 服务分类 -->
        <view class="fp-section" v-if="categories.length > 0">
          <text class="fp-label">服务分类</text>
          <view class="fp-chips">
            <view
              v-for="cat in categories" :key="cat.ID"
              :class="['fp-chip', localFilter.categoryId === cat.ID ? 'active' : '']"
              @click="localFilter.categoryId = localFilter.categoryId === cat.ID ? 0 : cat.ID"
            >{{ cat.name }}</view>
          </view>
        </view>
      </scroll-view>

      <view class="fp-footer">
        <view class="fp-btn fp-btn-reset" @click="onReset">重置</view>
        <view class="fp-btn fp-btn-confirm" @click="onConfirm">确定</view>
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { reactive, watch } from 'vue'

interface FilterState {
  dateFrom: string
  dateTo: string
  status: number
  staffId: number
  payMethod: string
  categoryId: number
}

const props = withDefaults(defineProps<{
  visible: boolean
  filter: FilterState
  statusOptions?: { value: number; label: string }[]
  statusLabel?: string
  payMethods?: { value: string; label: string }[]
  staffList?: any[]
  categories?: any[]
}>(), {
  statusOptions: () => [],
  statusLabel: '状态',
  payMethods: () => [],
  staffList: () => [],
  categories: () => [],
})

const emit = defineEmits<{
  close: []
  confirm: [filter: FilterState]
}>()

const localFilter = reactive<FilterState>({
  dateFrom: '',
  dateTo: '',
  status: -1,
  staffId: 0,
  payMethod: '',
  categoryId: 0,
})

watch(() => props.visible, (v) => {
  if (v) Object.assign(localFilter, props.filter)
})

function onReset() {
  localFilter.dateFrom = ''
  localFilter.dateTo = ''
  localFilter.status = -1
  localFilter.staffId = 0
  localFilter.payMethod = ''
  localFilter.categoryId = 0
}
function onConfirm() {
  emit('confirm', { ...localFilter })
  emit('close')
}
</script>

<style scoped>
.fp-overlay { position: fixed; top: 0; left: 0; right: 0; bottom: 0; background: rgba(0,0,0,0.4); z-index: 3000; display: flex; justify-content: flex-end; }
.fp-panel { width: 80%; max-width: 620rpx; background: #fff; height: 100%; display: flex; flex-direction: column; }
.fp-body { flex: 1; padding: 32rpx 28rpx; overflow-y: auto; }
.fp-section { margin-bottom: 36rpx; }
.fp-label { font-size: 28rpx; font-weight: 600; color: #1F2937; display: block; margin-bottom: 16rpx; }
.fp-date-row { display: flex; align-items: center; gap: 12rpx; background: #F9FAFB; border: 1rpx solid #E5E7EB; border-radius: 12rpx; padding: 16rpx 20rpx; }
.fp-date-btn { font-size: 26rpx; color: #374151; flex: 1; text-align: center; min-width: 140rpx; }
.fp-date-sep { font-size: 26rpx; color: #9CA3AF; }
.fp-chips { display: flex; flex-wrap: wrap; gap: 12rpx; }
.fp-chip { font-size: 24rpx; padding: 12rpx 28rpx; border-radius: 999rpx; background: #F3F4F6; color: #374151; border: 1rpx solid #E5E7EB; }
.fp-chip.active { background: #1F2937; color: #fff; border-color: #1F2937; }
.fp-footer { display: flex; gap: 16rpx; padding: 20rpx 28rpx; border-top: 1rpx solid #F3F4F6; padding-bottom: calc(20rpx + env(safe-area-inset-bottom)); }
.fp-btn { flex: 1; text-align: center; padding: 20rpx 0; border-radius: 12rpx; font-size: 28rpx; font-weight: 600; }
.fp-btn-reset { background: #fff; color: #374151; border: 2rpx solid #E5E7EB; }
.fp-btn-confirm { background: #FBBF24; color: #1F2937; }
</style>

