<template>
  <view class="page">
    <view v-if="customer" class="profile">
      <view class="avatar">{{ (customer.nickname || '客').charAt(0) }}</view>
      <text class="name">{{ customer.nickname || '未命名' }}</text>
      <text class="phone">{{ customer.phone || '未绑定手机' }}</text>
    </view>

    <view class="stats" v-if="customer">
      <view class="stat-item">
        <text class="stat-val">{{ customer.visit_count }}</text>
        <text class="stat-label">到店次数</text>
      </view>
      <view class="stat-item">
        <text class="stat-val">¥{{ customer.total_spent.toFixed(2) }}</text>
        <text class="stat-label">累计消费</text>
      </view>
      <view class="stat-item">
        <text class="stat-val">{{ customer.last_visit_at ? customer.last_visit_at.substring(0,10) : '-' }}</text>
        <text class="stat-label">最近到店</text>
      </view>
    </view>

    <!-- 会员卡区块 -->
    <view class="section" v-if="customer">
      <view class="section-header">
        <text class="section-title">会员卡</text>
      </view>

      <!-- 已有会员卡 -->
      <view v-if="memberCard" class="member-card-info">
        <view class="mc-header" :style="{ background: memberCard.template?.color || 'linear-gradient(135deg, #4F46E5, #7C3AED)' }">
          <view class="mc-top">
            <text class="mc-name">{{ memberCard.card_name }}</text>
            <text class="mc-discount">{{ (memberCard.discount_rate * 10).toFixed(1) }}折</text>
          </view>
          <text class="mc-balance">¥{{ memberCard.balance.toFixed(2) }}</text>
          <text class="mc-expire">{{ memberCard.expire_at ? '有效期至 ' + memberCard.expire_at.substring(0,10) : '永久有效' }}</text>
        </view>
        <view class="mc-stats">
          <view class="mc-stat">
            <text class="mc-stat-val">¥{{ memberCard.total_recharge.toFixed(2) }}</text>
            <text class="mc-stat-label">累计充值</text>
          </view>
          <view class="mc-stat">
            <text class="mc-stat-val">¥{{ memberCard.total_spent.toFixed(2) }}</text>
            <text class="mc-stat-label">累计消费</text>
          </view>
        </view>
        <view class="mc-btns">
          <button class="btn-recharge" @click="showRechargeModal = true">充值</button>
          <button class="btn-adjust" v-if="isAdmin" @click="showAdjustModal = true">调整余额</button>
        </view>
      </view>

      <!-- 无会员卡 -->
      <view v-else class="no-card">
        <text class="no-card-text">该客户暂未开通会员卡</text>
        <button class="btn-open-card" @click="showOpenCardModal = true">开通会员卡</button>
      </view>

      <!-- 流水记录 -->
      <view v-if="records.length > 0" class="records">
        <text class="records-title">充值/消费记录</text>
        <view class="record-item" v-for="r in records" :key="r.ID">
          <view class="record-left">
            <text :class="['record-type', r.type === 1 || (r.type === 4 && r.amount > 0) ? 'type-in' : 'type-out']">
              {{ r.type === 1 ? '充值' : r.type === 2 ? '消费' : r.type === 4 ? '调整' : '退款' }}
            </text>
            <text class="record-remark">{{ r.remark }}</text>
          </view>
          <view class="record-right">
            <text :class="['record-amount', r.type === 1 || (r.type === 4 && r.amount > 0) ? 'amt-in' : 'amt-out']">
              {{ r.amount > 0 ? '+' : '' }}¥{{ r.amount.toFixed(2) }}
            </text>
            <text class="record-balance">余额:¥{{ r.balance_after.toFixed(2) }}</text>
            <text class="record-time">{{ r.CreatedAt.substring(5,16) }}</text>
          </view>
        </view>
      </view>
    </view>

    <view class="section" v-if="customer">
      <view class="section-header">
        <text class="section-title">备注</text>
        <text class="link" @click="goEdit">编辑</text>
      </view>
      <text class="remark">{{ customer.remark || '暂无备注' }}</text>
      <view class="tag-list" v-if="customer.tags">
        <text class="tag" v-for="t in customer.tags.split(',')" :key="t">{{ t }}</text>
      </view>
    </view>

    <view class="section">
      <view class="section-header">
        <text class="section-title">宠物</text>
        <text class="link" @click="goAddPet">+ 添加</text>
      </view>
      <view v-if="pets.length === 0" class="empty-sm">暂无宠物</view>
      <view class="pet-card" v-for="pet in pets" :key="pet.ID" @click="goEditPet(pet.ID)">
        <view class="pet-main">
          <view class="pet-top">
            <text class="pet-name">{{ pet.name }}</text>
            <text class="pet-detail">{{ pet.breed || '未知品种' }} · {{ pet.gender === 1 ? '♂弟弟' : pet.gender === 2 ? '♀妹妹' : '' }}<text v-if="pet.birth_date"> · {{ calcAge(pet.birth_date) }}</text></text>
            <text class="pet-weight" v-if="pet.weight">{{ pet.weight }}kg</text>
          </view>
          <view class="pet-tags" v-if="pet.fur_level || pet.neutered || pet.personality || (pet.aggression && pet.aggression !== '无')">
            <text class="pet-tag" v-if="pet.fur_level">{{ pet.fur_level }}</text>
            <text class="pet-tag" v-if="pet.neutered">已绝育</text>
            <text class="pet-tag" v-if="pet.personality" :style="{ background: getPersonalityBg(pet.personality), color: getPersonalityColor(pet.personality) }">{{ pet.personality }}</text>
            <text class="pet-tag" v-if="pet.aggression && pet.aggression !== '无'" style="background:#FEE2E2;color:#EF4444;">攻击性:{{ pet.aggression }}</text>
          </view>
        </view>
        <text class="pet-arrow">›</text>
      </view>
    </view>

    <!-- 开卡弹窗 -->
    <view class="modal-mask" v-if="showOpenCardModal" @click="showOpenCardModal = false">
      <view class="modal-body" @click.stop>
        <text class="modal-title">开通会员卡</text>
        <view class="form-item">
          <text class="label">选择会员卡</text>
          <view class="tpl-list">
            <view
              v-for="tpl in templates" :key="tpl.ID"
              :class="['tpl-option', selectedTplId === tpl.ID ? 'tpl-active' : '']"
              @click="selectTpl(tpl)"
            >
              <text class="tpl-name">{{ tpl.name }}</text>
              <text class="tpl-meta">门槛¥{{ tpl.min_recharge }} · {{ (tpl.discount_rate * 10).toFixed(1) }}折</text>
            </view>
          </view>
        </view>
        <view class="form-item">
          <text class="label">充值金额 *</text>
          <input v-model="rechargeAmount" type="digit" :placeholder="'最低' + minRecharge" class="input" />
        </view>
        <view class="modal-btns">
          <view class="modal-btn cancel" @click="showOpenCardModal = false">取消</view>
          <view class="modal-btn confirm" @click="doOpenCard">确认开卡</view>
        </view>
      </view>
    </view>

    <!-- 充值弹窗 -->
    <view class="modal-mask" v-if="showRechargeModal" @click="showRechargeModal = false">
      <view class="modal-body" @click.stop>
        <text class="modal-title">充值</text>
        <view class="form-item">
          <text class="label">充值金额 *</text>
          <input v-model="rechargeAmount" type="digit" placeholder="请输入金额" class="input" />
        </view>
        <view class="modal-btns">
          <view class="modal-btn cancel" @click="showRechargeModal = false">取消</view>
          <view class="modal-btn confirm" @click="doRecharge">确认充值</view>
        </view>
      </view>
    </view>

    <!-- 调整余额弹窗（仅店长/管理员） -->
    <view class="modal-mask" v-if="showAdjustModal" @click="showAdjustModal = false">
      <view class="modal-body" @click.stop>
        <text class="modal-title">调整余额</text>
        <view class="form-item">
          <text class="label">调整金额 *</text>
          <input v-model="adjustAmount" type="digit" placeholder="正数增加，负数减少" class="input" />
          <text class="field-hint">当前余额: ¥{{ memberCard?.balance.toFixed(2) }}</text>
        </view>
        <view class="form-item">
          <text class="label">调整原因 *</text>
          <input v-model="adjustRemark" placeholder="请填写调整原因" class="input" />
        </view>
        <view class="modal-btns">
          <view class="modal-btn cancel" @click="showAdjustModal = false">取消</view>
          <view class="modal-btn confirm" @click="doAdjust">确认调整</view>
        </view>
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { onLoad, onShow } from '@dcloudio/uni-app'
import { getCustomer, getCustomerPets } from '@/api/customer'
import { getCustomerCard, getCardTemplates, openCard, recharge, adjustBalance, getRechargeRecords } from '@/api/member-card'
import { useAuthStore } from '@/store/auth'
import { getPersonalityColor, getPersonalityBg } from '@/utils/personality'

function calcAge(birthDate: string): string {
  if (!birthDate) return ''
  const birth = new Date(birthDate)
  const now = new Date()
  const months = (now.getFullYear() - birth.getFullYear()) * 12 + (now.getMonth() - birth.getMonth())
  if (months < 1) return '不到1个月'
  if (months < 12) return `${months}个月`
  const years = Math.floor(months / 12)
  const rem = months % 12
  return rem > 0 ? `${years}岁${rem}个月` : `${years}岁`
}

const id = ref(0)
const customer = ref<Customer | null>(null)
const pets = ref<Pet[]>([])
const memberCard = ref<MemberCard | null>(null)
const records = ref<RechargeRecord[]>([])
const templates = ref<MemberCardTemplate[]>([])

const authStore = useAuthStore()
const isAdmin = computed(() => authStore.staffInfo?.role === 'admin')

const showOpenCardModal = ref(false)
const showRechargeModal = ref(false)
const showAdjustModal = ref(false)
const selectedTplId = ref(0)
const rechargeAmount = ref('')
const adjustAmount = ref('')
const adjustRemark = ref('')

const minRecharge = computed(() => {
  const tpl = templates.value.find(t => t.ID === selectedTplId.value)
  return tpl ? tpl.min_recharge : 0
})

function selectTpl(tpl: MemberCardTemplate) {
  selectedTplId.value = tpl.ID
  if (!rechargeAmount.value || parseFloat(rechargeAmount.value) < tpl.min_recharge) {
    rechargeAmount.value = String(tpl.min_recharge)
  }
}

onLoad(async (query) => {
  if (query?.id) {
    id.value = parseInt(query.id)
  }
})

onShow(async () => {
  if (!id.value) return
  await loadAll()
})

async function loadAll() {
  const [cRes, pRes, cardRes, recordRes, tplRes] = await Promise.all([
    getCustomer(id.value),
    getCustomerPets(id.value),
    getCustomerCard(id.value),
    getRechargeRecords(id.value),
    getCardTemplates(),
  ])
  customer.value = cRes.data
  pets.value = pRes.data || []
  memberCard.value = cardRes.data || null
  records.value = recordRes.data || []
  templates.value = (tplRes.data || []).filter((t: MemberCardTemplate) => t.status === 1)
}

async function doOpenCard() {
  if (!selectedTplId.value) {
    uni.showToast({ title: '请选择会员卡', icon: 'none' }); return
  }
  const amount = parseFloat(rechargeAmount.value)
  if (!amount || amount < minRecharge.value) {
    uni.showToast({ title: `充值金额不能低于${minRecharge.value}元`, icon: 'none' }); return
  }
  try {
    await openCard(id.value, { template_id: selectedTplId.value, recharge_amount: amount })
    uni.showToast({ title: '开卡成功', icon: 'success' })
    showOpenCardModal.value = false
    rechargeAmount.value = ''
    await loadAll()
  } catch (e: any) {
    uni.showToast({ title: e.message || '开卡失败', icon: 'none' })
  }
}

async function doRecharge() {
  const amount = parseFloat(rechargeAmount.value)
  if (!amount || amount <= 0) {
    uni.showToast({ title: '请输入充值金额', icon: 'none' }); return
  }
  try {
    await recharge(id.value, { amount })
    uni.showToast({ title: '充值成功', icon: 'success' })
    showRechargeModal.value = false
    rechargeAmount.value = ''
    await loadAll()
  } catch (e: any) {
    uni.showToast({ title: e.message || '充值失败', icon: 'none' })
  }
}

async function doAdjust() {
  const amount = parseFloat(adjustAmount.value)
  if (!amount) {
    uni.showToast({ title: '请输入调整金额', icon: 'none' }); return
  }
  if (!adjustRemark.value.trim()) {
    uni.showToast({ title: '请填写调整原因', icon: 'none' }); return
  }
  try {
    await adjustBalance(id.value, { amount, remark: adjustRemark.value.trim() })
    uni.showToast({ title: '调整成功', icon: 'success' })
    showAdjustModal.value = false
    adjustAmount.value = ''
    adjustRemark.value = ''
    await loadAll()
  } catch (e: any) {
    uni.showToast({ title: e.message || '调整失败', icon: 'none' })
  }
}

function goEdit() { uni.navigateTo({ url: `/pages/customer/edit?id=${id.value}` }) }
function goAddPet() { uni.navigateTo({ url: `/pages/pet/edit?owner_phone=${customer.value?.phone || ''}` }) }
function goEditPet(petId: number) { uni.navigateTo({ url: `/pages/pet/edit?id=${petId}` }) }
</script>

<style scoped>
.page { padding: 24rpx; }
.profile { background: #fff; border-radius: 16rpx; padding: 40rpx; text-align: center; margin-bottom: 16rpx; }
.avatar { width: 120rpx; height: 120rpx; border-radius: 50%; background: #6366F1; color: #fff; display: inline-flex; align-items: center; justify-content: center; font-size: 48rpx; font-weight: bold; margin-bottom: 16rpx; }
.name { display: block; font-size: 34rpx; font-weight: 600; color: #1F2937; }
.phone { display: block; font-size: 26rpx; color: #6B7280; margin-top: 8rpx; }
.stats { display: flex; background: #fff; border-radius: 16rpx; padding: 24rpx; margin-bottom: 16rpx; }
.stat-item { flex: 1; text-align: center; }
.stat-val { display: block; font-size: 30rpx; font-weight: bold; color: #4F46E5; }
.stat-label { display: block; font-size: 22rpx; color: #9CA3AF; margin-top: 8rpx; }
.section { background: #fff; border-radius: 16rpx; padding: 24rpx; margin-bottom: 16rpx; }
.section-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 16rpx; }
.section-title { font-size: 30rpx; font-weight: 600; color: #1F2937; }
.link { font-size: 26rpx; color: #4F46E5; }
.remark { font-size: 26rpx; color: #6B7280; }
.tag-list { display: flex; flex-wrap: wrap; gap: 12rpx; margin-top: 12rpx; }
.tag { font-size: 22rpx; padding: 6rpx 16rpx; background: #EEF2FF; color: #4F46E5; border-radius: 16rpx; }
.empty-sm { font-size: 26rpx; color: #9CA3AF; text-align: center; padding: 24rpx; }
.pet-card { padding: 16rpx 0; border-bottom: 1rpx solid #F3F4F6; display: flex; align-items: center; }
.pet-card:last-child { border-bottom: none; }
.pet-main { flex: 1; }
.pet-top { display: flex; align-items: baseline; gap: 8rpx; }
.pet-name { font-size: 28rpx; font-weight: 600; color: #1F2937; }
.pet-detail { font-size: 22rpx; color: #6B7280; flex: 1; }
.pet-weight { font-size: 24rpx; color: #4F46E5; font-weight: 600; }
.pet-tags { display: flex; gap: 8rpx; flex-wrap: wrap; margin-top: 8rpx; }
.pet-tag { font-size: 20rpx; padding: 2rpx 10rpx; background: #EEF2FF; color: #4F46E5; border-radius: 10rpx; }
.pet-arrow { font-size: 32rpx; color: #C0C4CC; margin-left: 8rpx; }

/* Member card */
.member-card-info { margin-bottom: 20rpx; }
.mc-header { border-radius: 16rpx; padding: 28rpx; color: #fff; margin-bottom: 16rpx; }
.mc-top { display: flex; justify-content: space-between; align-items: center; margin-bottom: 16rpx; }
.mc-name { font-size: 28rpx; font-weight: 600; }
.mc-discount { font-size: 24rpx; background: rgba(255,255,255,0.2); padding: 4rpx 16rpx; border-radius: 20rpx; }
.mc-balance { font-size: 52rpx; font-weight: 800; display: block; }
.mc-expire { font-size: 22rpx; opacity: 0.8; display: block; margin-top: 8rpx; }
.mc-stats { display: flex; gap: 16rpx; margin-bottom: 16rpx; }
.mc-stat { flex: 1; text-align: center; background: #F9FAFB; padding: 16rpx; border-radius: 12rpx; }
.mc-stat-val { display: block; font-size: 28rpx; font-weight: 600; color: #4F46E5; }
.mc-stat-label { display: block; font-size: 22rpx; color: #9CA3AF; margin-top: 4rpx; }
.mc-btns { display: flex; gap: 16rpx; }
.btn-recharge { flex: 1; background: #4F46E5; color: #fff; border-radius: 12rpx; font-size: 28rpx; }
.btn-adjust { flex: 1; background: #fff; color: #F59E0B; border: 1rpx solid #F59E0B; border-radius: 12rpx; font-size: 28rpx; }
.field-hint { font-size: 22rpx; color: #9CA3AF; display: block; margin-top: 4rpx; }
.no-card { text-align: center; padding: 24rpx 0; }
.no-card-text { font-size: 26rpx; color: #9CA3AF; display: block; margin-bottom: 16rpx; }
.btn-open-card { background: linear-gradient(135deg, #4F46E5, #7C3AED); color: #fff; border-radius: 12rpx; font-size: 28rpx; }

/* Records */
.records { margin-top: 20rpx; padding-top: 20rpx; border-top: 1rpx solid #F3F4F6; }
.records-title { font-size: 26rpx; font-weight: 600; color: #6B7280; display: block; margin-bottom: 12rpx; }
.record-item { display: flex; justify-content: space-between; align-items: center; padding: 14rpx 0; border-bottom: 1rpx solid #F9FAFB; }
.record-item:last-child { border-bottom: none; }
.record-left { display: flex; align-items: center; gap: 12rpx; }
.record-type { font-size: 22rpx; padding: 4rpx 12rpx; border-radius: 8rpx; }
.type-in { background: #D1FAE5; color: #059669; }
.type-out { background: #FEE2E2; color: #DC2626; }
.record-remark { font-size: 24rpx; color: #6B7280; }
.record-right { text-align: right; }
.record-amount { display: block; font-size: 26rpx; font-weight: 600; }
.amt-in { color: #059669; }
.amt-out { color: #DC2626; }
.record-balance { display: block; font-size: 20rpx; color: #9CA3AF; }
.record-time { display: block; font-size: 20rpx; color: #D1D5DB; }

/* Modal */
.modal-mask { position: fixed; top: 0; left: 0; right: 0; bottom: 0; background: rgba(0,0,0,0.5); display: flex; align-items: center; justify-content: center; z-index: 999; }
.modal-body { background: #fff; border-radius: 20rpx; padding: 40rpx; width: 600rpx; max-height: 80vh; overflow-y: auto; }
.modal-title { font-size: 32rpx; font-weight: 700; color: #1F2937; display: block; text-align: center; margin-bottom: 24rpx; }
.form-item { margin-bottom: 20rpx; }
.label { font-size: 26rpx; color: #374151; display: block; margin-bottom: 8rpx; }
.input { font-size: 28rpx; color: #1F2937; height: 60rpx; background: #F9FAFB; border-radius: 8rpx; padding: 0 16rpx; }
.tpl-list { display: flex; flex-direction: column; gap: 12rpx; }
.tpl-option { padding: 16rpx 20rpx; border: 2rpx solid #E5E7EB; border-radius: 12rpx; }
.tpl-active { border-color: #4F46E5; background: #EEF2FF; }
.tpl-name { font-size: 28rpx; font-weight: 600; color: #1F2937; display: block; }
.tpl-active .tpl-name { color: #4F46E5; }
.tpl-meta { font-size: 24rpx; color: #6B7280; display: block; margin-top: 4rpx; }
.modal-btns { display: flex; gap: 16rpx; margin-top: 24rpx; }
.modal-btn { flex: 1; text-align: center; padding: 18rpx; border-radius: 12rpx; font-size: 28rpx; }
.cancel { background: #F3F4F6; color: #6B7280; }
.confirm { background: #4F46E5; color: #fff; }
</style>
