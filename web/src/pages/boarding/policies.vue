<template>
  <SideLayout>
    <view class="page">
      <view class="header">
        <text class="title">寄养优惠</text>
        <view class="btn btn-primary" @click="startCreate">+ 新建优惠</view>
      </view>
      <view v-if="editing" class="card">
        <text class="section-title">{{ form.id ? '编辑优惠' : '新建优惠' }}</text>
        <input v-model="form.name" class="input" placeholder="优惠名称" />
        <picker :range="policyTypeOptions" :value="policyTypeIndex" @change="onTypeChange">
          <view class="picker">{{ policyTypeOptions[policyTypeIndex] }}</view>
        </picker>
        <view v-if="form.policy_type === 'stay_n_free_m'" class="inline-grid">
          <input v-model="form.ruleStay" class="input" type="number" placeholder="住 N 天" />
          <input v-model="form.ruleFree" class="input" type="number" placeholder="免 M 天" />
        </view>
        <input v-else v-model="form.ruleSurcharge" class="input input-amount" type="digit" placeholder="每晚加收金额" />
        <view class="inline-grid">
          <picker mode="date" :value="form.valid_from" @change="form.valid_from = $event.detail.value">
            <view class="picker">{{ form.valid_from || '开始日期（可选）' }}</view>
          </picker>
          <picker mode="date" :value="form.valid_to" @change="form.valid_to = $event.detail.value">
            <view class="picker">{{ form.valid_to || '结束日期（可选）' }}</view>
          </picker>
        </view>
        <input v-model="form.priority" class="input" type="number" placeholder="优先级，数字越大越优先" />
        <picker :range="statusOptions" :value="statusIndex" @change="onStatusChange">
          <view class="picker">{{ statusOptions[statusIndex] }}</view>
        </picker>
        <textarea v-model="form.remark" class="textarea" placeholder="备注" />
        <view class="actions">
          <view class="btn" @click="cancelEdit">取消</view>
          <view class="btn btn-primary" @click="save">保存</view>
        </view>
      </view>
      <view class="card" v-for="item in policies" :key="item.ID">
        <view class="policy-head">
          <view>
            <text class="policy-name">{{ item.name }}</text>
            <text class="policy-meta">{{ describe(item) }}</text>
          </view>
          <view class="policy-edit" @click="edit(item)">编辑</view>
        </view>
        <text class="policy-extra">优先级 {{ item.priority }} · {{ item.status === 1 ? '启用' : '停用' }}</text>
      </view>
    </view>
  </SideLayout>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { onShow } from '@dcloudio/uni-app'
import SideLayout from '@/components/SideLayout.vue'
import { createBoardingPolicy, getBoardingPolicies, updateBoardingPolicy } from '@/api/boarding'

const policies = ref<BoardingDiscountPolicy[]>([])
const editing = ref(false)
const form = ref<any>({
  id: 0,
  name: '',
  policy_type: 'stay_n_free_m',
  ruleStay: '5',
  ruleFree: '1',
  ruleSurcharge: '30',
  valid_from: '',
  valid_to: '',
  priority: '10',
  status: 1,
  remark: '',
})
const policyTypeOptions = ['住N免M', '节假日加收']
const policyTypeValues = ['stay_n_free_m', 'holiday_surcharge']
const statusOptions = ['启用', '停用']

const policyTypeIndex = computed(() => Math.max(policyTypeValues.indexOf(form.value.policy_type), 0))
const statusIndex = computed(() => (form.value.status === 1 ? 0 : 1))

function onTypeChange(e: any) {
  form.value.policy_type = policyTypeValues[e.detail.value]
}

function onStatusChange(e: any) {
  form.value.status = e.detail.value === 0 ? 1 : 0
}

function startCreate() {
  editing.value = true
  form.value = {
    id: 0,
    name: '',
    policy_type: 'stay_n_free_m',
    ruleStay: '5',
    ruleFree: '1',
    ruleSurcharge: '30',
    valid_from: '',
    valid_to: '',
    priority: '10',
    status: 1,
    remark: '',
  }
}

function cancelEdit() {
  editing.value = false
}

function describe(item: BoardingDiscountPolicy) {
  try {
    const rule = JSON.parse(item.rule_json || '{}')
    if (item.policy_type === 'stay_n_free_m') return `住 ${rule.stay || 0} 免 ${rule.free || 0}`
    if (item.policy_type === 'holiday_surcharge') return `节假日每晚 +¥${rule.surcharge || 0}`
  } catch {}
  return item.remark || '寄养优惠'
}

function edit(item: BoardingDiscountPolicy) {
  let ruleStay = '5'
  let ruleFree = '1'
  let ruleSurcharge = '30'
  try {
    const rule = JSON.parse(item.rule_json || '{}')
    ruleStay = String(rule.stay || 5)
    ruleFree = String(rule.free || 1)
    ruleSurcharge = String(rule.surcharge || 30)
  } catch {}
  editing.value = true
  form.value = {
    id: item.ID,
    name: item.name,
    policy_type: item.policy_type,
    ruleStay,
    ruleFree,
    ruleSurcharge,
    valid_from: item.valid_from || '',
    valid_to: item.valid_to || '',
    priority: String(item.priority || 0),
    status: item.status,
    remark: item.remark || '',
  }
}

async function loadData() {
  const res = await getBoardingPolicies()
  policies.value = res.data || []
}

async function save() {
  if (!form.value.name.trim()) {
    uni.showToast({ title: '请填写优惠名称', icon: 'none' })
    return
  }
  const rule = form.value.policy_type === 'stay_n_free_m'
    ? { stay: Number(form.value.ruleStay) || 0, free: Number(form.value.ruleFree) || 0 }
    : { surcharge: Number(form.value.ruleSurcharge) || 0 }
  const payload = {
    name: form.value.name.trim(),
    policy_type: form.value.policy_type,
    rule,
    valid_from: form.value.valid_from,
    valid_to: form.value.valid_to,
    priority: Number(form.value.priority) || 0,
    stackable: true,
    status: form.value.status,
    remark: form.value.remark,
  }
  if (form.value.id) await updateBoardingPolicy(form.value.id, payload)
  else await createBoardingPolicy(payload)
  uni.showToast({ title: '保存成功', icon: 'success' })
  editing.value = false
  await loadData()
}

onShow(loadData)
</script>

<style scoped>
.page { padding: 24rpx; display: flex; flex-direction: column; gap: 20rpx; }
.header { display: flex; justify-content: space-between; align-items: center; gap: 12rpx; }
.title { font-size: 34rpx; font-weight: 700; color: #111827; }
.card { background: #fff; border-radius: 18rpx; padding: 24rpx; box-shadow: 0 12rpx 28rpx rgba(15, 23, 42, 0.04); }
.section-title { display: block; font-size: 28rpx; font-weight: 700; color: #111827; margin-bottom: 14rpx; }
.input, .picker, .textarea { width: 100%; box-sizing: border-box; margin-bottom: 14rpx; background: #F9FAFB; border: 1rpx solid #E5E7EB; border-radius: 12rpx; padding: 18rpx 20rpx; font-size: 26rpx; color: #111827; min-height: 60rpx; }
.textarea { min-height: 120rpx; }
.inline-grid { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 12rpx; }
.btn { padding: 14rpx 24rpx; border-radius: 12rpx; background: #F3F4F6; color: #374151; font-size: 24rpx; }
.btn-primary { background: #4F46E5; color: #fff; }
.actions, .policy-head { display: flex; justify-content: space-between; gap: 12rpx; align-items: center; }
.policy-name { display: block; font-size: 28rpx; font-weight: 700; color: #111827; }
.policy-meta, .policy-extra { display: block; margin-top: 8rpx; font-size: 22rpx; color: #6B7280; }
.policy-edit { font-size: 24rpx; color: #4F46E5; }
</style>
