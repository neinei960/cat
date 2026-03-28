<template>
  <SideLayout>
  <view class="page">
    <view v-if="loading" class="state">加载中...</view>
    <view v-else-if="!appt" class="state">预约不存在</view>
    <template v-else>
      <view class="summary-card">
        <text class="summary-title">预约开单确认</text>
        <text class="summary-line">{{ appt.date }} {{ appt.start_time }} - {{ appt.end_time }}</text>
        <text class="summary-line">客户：{{ appt.customer?.nickname || appt.customer?.phone || '-' }}</text>
        <text class="summary-line">洗护师：{{ appt.staff?.name || '待分配' }}</text>
        <text class="summary-amount">预计共 {{ drafts.length }} 单 · ¥{{ totalAmount.toFixed(2) }}</text>
      </view>

      <view class="draft-list">
        <view class="draft-card" v-for="(draft, di) in drafts" :key="draft.petId">
          <view class="draft-head">
            <text class="draft-name">{{ draft.petName }}</text>
            <text class="draft-price">¥{{ draft.amount.toFixed(2) }}</text>
          </view>
          <text class="draft-meta">{{ draft.meta }}</text>
          <view class="draft-tags" v-if="draft.tags.length > 0">
            <text
              v-for="tag in draft.tags"
              :key="`${draft.petId}-${tag.text}`"
              :class="['draft-tag', tag.className]"
              :style="tag.style"
            >{{ tag.text }}</text>
          </view>

          <!-- 服务列表（可编辑） -->
          <view class="service-list">
            <view class="service-row" v-for="(svc, si) in draft.services" :key="`${draft.petId}-${svc.service_id}-${si}`">
              <text class="service-name">{{ svc.service_name }}</text>
              <view class="service-edit">
                <text class="service-price" @click="editPrice(di, si)">¥{{ svc.price }}</text>
                <text class="service-dur">{{ svc.duration }}分钟</text>
                <text class="service-del" @click="removeService(di, si)">✕</text>
              </view>
            </view>
            <view class="add-service-row" @click="openAddService(di)">
              <text class="add-service-text">+ 添加服务</text>
            </view>
          </view>
        </view>
      </view>

      <view class="notes-card" v-if="appt.notes">
        <text class="notes-title">预约备注</text>
        <text class="notes-text">{{ appt.notes }}</text>
      </view>

      <button class="submit-btn" :loading="submitting" @click="submitBatch">确认生成 {{ drafts.length }} 张订单</button>

      <!-- 修改价格弹窗 -->
      <view v-if="editingPrice" class="modal-mask" @click="editingPrice = null">
        <view class="modal-body" @click.stop>
          <text class="modal-title">修改价格</text>
          <text class="modal-svc-name">{{ editingPrice.name }}</text>
          <input v-model="editPriceValue" type="digit" class="modal-input" placeholder="输入新价格" />
          <view class="modal-btns">
            <view class="modal-btn cancel" @click="editingPrice = null">取消</view>
            <view class="modal-btn confirm" @click="savePrice">确定</view>
          </view>
        </view>
      </view>

      <!-- 添加服务弹窗（按分类） -->
      <view v-if="showAddService" class="modal-mask" @click="showAddService = false">
        <view class="modal-body modal-body-tall" @click.stop>
          <text class="modal-title">添加服务</text>
          <!-- 一级分类 tabs -->
          <scroll-view scroll-x class="cat-tabs" show-scrollbar="false">
            <view class="cat-tabs-inner">
              <view
                v-for="cat in topCategories"
                :key="cat.ID"
                :class="['cat-tab', activeCat1 === cat.ID ? 'cat-tab-active' : '']"
                @click="activeCat1 = cat.ID; activeCat2 = 0"
              >{{ cat.name }}</view>
            </view>
          </scroll-view>
          <!-- 二级分类 tabs -->
          <scroll-view scroll-x class="cat-tabs cat-tabs-sub" show-scrollbar="false" v-if="subCategories.length">
            <view class="cat-tabs-inner">
              <view
                :class="['cat-tab cat-tab-sub', activeCat2 === 0 ? 'cat-tab-active' : '']"
                @click="activeCat2 = 0"
              >全部</view>
              <view
                v-for="cat in subCategories"
                :key="cat.ID"
                :class="['cat-tab cat-tab-sub', activeCat2 === cat.ID ? 'cat-tab-active' : '']"
                @click="activeCat2 = cat.ID"
              >{{ cat.name }}</view>
            </view>
          </scroll-view>
          <!-- 服务列表 -->
          <view v-if="filteredServices.length === 0" class="modal-empty">该分类下暂无服务</view>
          <scroll-view scroll-y class="service-pick-list">
            <view v-for="svc in filteredServices" :key="svc.ID" class="service-pick-group">
              <view class="service-pick-item" @click="toggleServiceExpand(svc)">
                <text class="service-pick-name">{{ svc.name }}</text>
                <text class="service-pick-arrow" v-if="svc.price_rules?.length">{{ expandedServiceId === svc.ID ? '▾' : '▸' }}</text>
                <text class="service-pick-price" v-else @click.stop="addService(svc)">¥{{ svc.base_price }} · {{ svc.duration }}分钟</text>
              </view>
              <!-- 价格规格展开 -->
              <view v-if="expandedServiceId === svc.ID && svc.price_rules?.length" class="price-rules">
                <view
                  class="price-rule-item"
                  v-for="rule in svc.price_rules"
                  :key="rule.ID"
                  @click="addServiceWithRule(svc, rule)"
                >
                  <text class="rule-level">{{ rule.name || rule.fur_level || '标准' }}</text>
                  <text class="rule-price">¥{{ rule.price }}{{ (rule.duration || svc.duration) ? ` · ${rule.duration || svc.duration}分钟` : '' }}</text>
                </view>
              </view>
            </view>
          </scroll-view>
        </view>
      </view>
    </template>
  </view>
  </SideLayout>
</template>

<script setup lang="ts">
import SideLayout from '@/components/SideLayout.vue'
import { computed, ref, reactive } from 'vue'
import { onLoad } from '@dcloudio/uni-app'
import { getAppointment } from '@/api/appointment'
import { createBatchOrdersFromAppointment } from '@/api/order'
import { getPersonalityBg, getPersonalityColor } from '@/utils/personality'

const appointmentId = ref(0)
const appt = ref<any>(null)
const loading = ref(true)
const submitting = ref(false)
const allServices = ref<any[]>([])
const categoryTree = ref<any[]>([]) // 树形：顶级分类，子分类在 children 里
const activeCat1 = ref(0)
const activeCat2 = ref(0)

const topCategories = computed(() => categoryTree.value)
const subCategories = computed(() => {
  const top = categoryTree.value.find(c => c.ID === activeCat1.value)
  return top?.children || []
})
const filteredServices = computed(() => {
  const existing = new Set(drafts[addingDraftIndex.value]?.services.map(s => s.service_id) || [])
  // 有价格规则的服务不过滤（可以选不同规格），无规则的按 service_id 去重
  let list = allServices.value.filter(s => (s.price_rules?.length > 0) || !existing.has(s.ID))
  if (activeCat2.value) {
    list = list.filter(s => s.category_id === activeCat2.value)
  } else if (activeCat1.value) {
    const subIds = new Set(subCategories.value.map((c: any) => c.ID))
    list = list.filter(s => subIds.has(s.category_id))
  }
  return list
})

interface DraftService {
  service_id: number
  service_name: string
  price: number
  duration: number
}

interface Draft {
  petId: number
  petName: string
  meta: string
  tags: Array<{ text: string; className: string; style?: string }>
  services: DraftService[]
  amount: number
}

const drafts = reactive<Draft[]>([])

function calcAge(birthDate?: string): string {
  if (!birthDate) return ''
  const birth = new Date(birthDate)
  if (Number.isNaN(birth.getTime())) return ''
  const now = new Date()
  const months = (now.getFullYear() - birth.getFullYear()) * 12 + (now.getMonth() - birth.getMonth())
  if (months < 1) return '不到1个月'
  if (months < 12) return `${months}个月`
  const years = Math.floor(months / 12)
  const rem = months % 12
  return rem > 0 ? `${years}年${rem}个月` : `${years}年`
}

function getPetMeta(pet: any) {
  const parts: string[] = []
  if (pet?.breed) parts.push(pet.breed)
  if (pet?.gender === 1) parts.push('弟弟')
  if (pet?.gender === 2) parts.push('妹妹')
  return parts.join(' · ') || '未填写宠物信息'
}

function getPetTags(pet: any) {
  const tags: Array<{ text: string; className: string; style?: string }> = []
  const age = calcAge(pet?.birth_date)
  if (age) tags.push({ text: age, className: 'tag-age' })
  if (pet?.fur_level) tags.push({ text: pet.fur_level, className: 'tag-fur' })
  if (pet?.neutered) tags.push({ text: '已绝育', className: 'tag-neutered' })
  if (pet?.personality) {
    tags.push({
      text: pet.personality,
      className: 'tag-personality',
      style: `background:${getPersonalityBg(pet.personality)};color:${getPersonalityColor(pet.personality)};`,
    })
  }
  if (pet?.aggression && pet.aggression !== '无') {
    tags.push({ text: `⚡ ${pet.aggression}`, className: 'tag-aggression' })
  }
  return tags
}

function recalcAmount(draft: Draft) {
  draft.amount = draft.services.reduce((s, svc) => s + svc.price, 0)
}

const totalAmount = computed(() => drafts.reduce((sum, d) => sum + d.amount, 0))

// === 修改价格 ===
const editingPrice = ref<{ di: number; si: number; name: string } | null>(null)
const editPriceValue = ref('')

function editPrice(di: number, si: number) {
  const svc = drafts[di].services[si]
  editingPrice.value = { di, si, name: svc.service_name }
  editPriceValue.value = String(svc.price)
}

function savePrice() {
  if (!editingPrice.value) return
  const { di, si } = editingPrice.value
  const val = parseFloat(editPriceValue.value)
  if (isNaN(val) || val < 0) {
    uni.showToast({ title: '请输入有效价格', icon: 'none' }); return
  }
  drafts[di].services[si].price = val
  recalcAmount(drafts[di])
  editingPrice.value = null
}

// === 删除服务 ===
function removeService(di: number, si: number) {
  const svc = drafts[di].services[si]
  uni.showModal({
    title: '删除服务',
    content: `确定删除「${svc.service_name}」？`,
    confirmColor: '#EF4444',
    success: (res) => {
      if (res.confirm) {
        drafts[di].services.splice(si, 1)
        recalcAmount(drafts[di])
      }
    }
  })
}

// === 添加服务 ===
const showAddService = ref(false)
const addingDraftIndex = ref(0)
const expandedServiceId = ref(0)

function toggleServiceExpand(svc: any) {
  if (svc.price_rules?.length) {
    expandedServiceId.value = expandedServiceId.value === svc.ID ? 0 : svc.ID
  } else {
    addService(svc)
  }
}

function addServiceWithRule(svc: any, rule: any) {
  const ruleName = rule.name || rule.fur_level || ''
  const name = ruleName ? `${svc.name}(${ruleName})` : svc.name
  drafts[addingDraftIndex.value].services.push({
    service_id: svc.ID,
    service_name: name,
    price: rule.price,
    duration: rule.duration || svc.duration,
  })
  recalcAmount(drafts[addingDraftIndex.value])
  showAddService.value = false
  expandedServiceId.value = 0
}

function openAddService(di: number) {
  addingDraftIndex.value = di
  expandedServiceId.value = 0
  if (topCategories.value.length && !activeCat1.value) {
    activeCat1.value = topCategories.value[0].ID
  }
  activeCat2.value = 0
  showAddService.value = true
}

function addService(svc: any) {
  drafts[addingDraftIndex.value].services.push({
    service_id: svc.ID,
    service_name: svc.name,
    price: svc.base_price,
    duration: svc.duration,
  })
  recalcAmount(drafts[addingDraftIndex.value])
  showAddService.value = false
}

// === 加载 ===
async function loadData() {
  if (!appointmentId.value) return
  loading.value = true
  try {
    const res = await getAppointment(appointmentId.value)
    appt.value = res.data

    // 构建可编辑 drafts
    const pets = Array.isArray(res.data?.pets) ? res.data.pets : []
    drafts.length = 0
    for (const petItem of pets) {
      const svcs = (petItem.services || []).map((s: any) => ({
        service_id: s.service_id,
        service_name: s.service_name,
        price: Number(s.price || 0),
        duration: Number(s.duration || 0),
      }))
      drafts.push({
        petId: petItem.pet_id,
        petName: petItem.pet?.name || `宠物#${petItem.pet_id}`,
        meta: getPetMeta(petItem.pet),
        tags: getPetTags(petItem.pet),
        services: svcs,
        amount: svcs.reduce((s: number, svc: DraftService) => s + svc.price, 0),
      })
    }

    // 加载服务列表和分类（用于添加服务）
    try {
      const { request } = await import('@/api/request')
      const [svcRes, catRes] = await Promise.all([
        request<any>({ url: '/b/services', params: { page: 1, page_size: 200 } }),
        request<any>({ url: '/b/service-categories' }),
      ])
      allServices.value = (svcRes.data?.list || svcRes.data || []).filter((s: any) => s.status === 1)
      categoryTree.value = catRes.data || []
    } catch { /* ignore */ }
  } finally {
    loading.value = false
  }
}

async function submitBatch() {
  if (!appointmentId.value) return
  // 先把修改后的价格同步回预约（通过更新预约的服务价格）
  submitting.value = true
  try {
    const res = await createBatchOrdersFromAppointment(appointmentId.value, {
      overrides: drafts.map(d => ({
        pet_id: d.petId,
        services: d.services.map(s => ({
          service_id: s.service_id,
          price: s.price,
          duration: s.duration,
        })),
      })),
    })
    const orders = res.data || []
    uni.showToast({ title: `已生成${orders.length}张订单`, icon: 'success' })
    setTimeout(() => {
      if (orders.length === 1 && orders[0]?.ID) {
        uni.redirectTo({ url: `/pages/order/detail?id=${orders[0].ID}` })
      } else {
        uni.redirectTo({ url: '/pages/order/list' })
      }
    }, 500)
  } catch (e: any) {
    uni.showToast({ title: e.message || '批量开单失败', icon: 'none' })
  } finally {
    submitting.value = false
  }
}

onLoad((query) => {
  appointmentId.value = parseInt(String(query?.appointment_id || 0)) || 0
  loadData()
})
</script>

<style scoped>
.page { padding: 24rpx; padding-bottom: calc(120rpx + env(safe-area-inset-bottom)); }
.state { text-align: center; padding: 120rpx 0; color: #9CA3AF; font-size: 28rpx; }
.summary-card, .draft-card, .notes-card { background: #fff; border-radius: 18rpx; padding: 24rpx; margin-bottom: 16rpx; box-shadow: 0 4rpx 18rpx rgba(15, 23, 42, 0.06); }
.summary-title { font-size: 32rpx; font-weight: 700; color: #111827; display: block; margin-bottom: 12rpx; }
.summary-line { font-size: 24rpx; color: #6B7280; display: block; margin-top: 6rpx; }
.summary-amount { font-size: 28rpx; color: #4F46E5; font-weight: 700; display: block; margin-top: 16rpx; }
.draft-list { display: flex; flex-direction: column; gap: 16rpx; }
.draft-head { display: flex; justify-content: space-between; align-items: center; gap: 16rpx; }
.draft-name { font-size: 30rpx; font-weight: 700; color: #111827; }
.draft-price { font-size: 28rpx; color: #4F46E5; font-weight: 700; }
.draft-meta { font-size: 24rpx; color: #6B7280; display: block; margin-top: 8rpx; }
.draft-tags { display: flex; flex-wrap: wrap; gap: 8rpx; margin-top: 10rpx; }
.draft-tag { display: inline-flex; align-items: center; padding: 4rpx 12rpx; border-radius: 999rpx; font-size: 18rpx; line-height: 1.2; background: #F3F4F6; color: #4B5563; }
.draft-tag.tag-age { background: #F8FAFC; color: #475569; }
.draft-tag.tag-fur { background: #EEF2FF; color: #4F46E5; }
.draft-tag.tag-neutered { background: #ECFDF5; color: #047857; }
.draft-tag.tag-aggression { background: #FEF2F2; color: #DC2626; }

.service-list { margin-top: 16rpx; border-top: 1rpx solid #EEF2F7; padding-top: 8rpx; }
.service-row { display: flex; justify-content: space-between; align-items: center; gap: 16rpx; padding: 12rpx 0; border-bottom: 1rpx solid #F3F4F6; }
.service-row:last-child { border-bottom: none; }
.service-name { font-size: 24rpx; color: #1F2937; flex: 1; }
.service-edit { display: flex; align-items: center; gap: 12rpx; }
.service-price { font-size: 24rpx; color: #4F46E5; font-weight: 600; padding: 4rpx 12rpx; background: #EEF2FF; border-radius: 8rpx; }
.service-dur { font-size: 20rpx; color: #9CA3AF; }
.service-del { font-size: 24rpx; color: #EF4444; padding: 4rpx 8rpx; }

.add-service-row { padding: 16rpx 0 4rpx; text-align: center; }
.add-service-text { font-size: 24rpx; color: #4F46E5; }

.notes-title { font-size: 26rpx; font-weight: 600; color: #1F2937; display: block; margin-bottom: 10rpx; }
.notes-text { font-size: 24rpx; color: #6B7280; line-height: 1.6; }
.submit-btn { position: fixed; bottom: 0; left: 0; right: 0; margin: 0; background: #4F46E5; color: #fff; border-radius: 0; font-size: 30rpx; padding: 20rpx 32rpx calc(20rpx + env(safe-area-inset-bottom)); z-index: 100; }

/* 弹窗 */
.modal-mask { position: fixed; top: 0; left: 0; right: 0; bottom: 0; background: rgba(0,0,0,0.4); z-index: 999; display: flex; align-items: center; justify-content: center; }
.modal-body { background: #fff; border-radius: 20rpx; padding: 32rpx; width: 80%; max-width: 600rpx; }
.modal-body-tall { max-height: 70vh; }
.modal-title { font-size: 30rpx; font-weight: 700; color: #111827; display: block; margin-bottom: 20rpx; }
.modal-svc-name { font-size: 26rpx; color: #6B7280; display: block; margin-bottom: 16rpx; }
.modal-input { background: #F3F4F6; border-radius: 12rpx; padding: 16rpx; font-size: 28rpx; width: 100%; }
.modal-btns { display: flex; gap: 16rpx; margin-top: 24rpx; }
.modal-btn { flex: 1; text-align: center; padding: 16rpx 0; border-radius: 12rpx; font-size: 28rpx; }
.modal-btn.cancel { background: #F3F4F6; color: #6B7280; }
.modal-btn.confirm { background: #4F46E5; color: #fff; }
.modal-empty { text-align: center; padding: 40rpx 0; color: #9CA3AF; font-size: 26rpx; }

.cat-tabs { margin-bottom: 12rpx; white-space: nowrap; }
.cat-tabs-sub { margin-bottom: 16rpx; }
.cat-tabs-inner { display: inline-flex; gap: 12rpx; }
.cat-tab { font-size: 26rpx; padding: 10rpx 20rpx; border-radius: 999rpx; background: #F3F4F6; color: #6B7280; white-space: nowrap; }
.cat-tab-active { background: #4F46E5; color: #fff; }
.cat-tab-sub { font-size: 24rpx; padding: 8rpx 16rpx; }
.service-pick-list { max-height: 40vh; }
.service-pick-group { border-bottom: 1rpx solid #F3F4F6; }
.service-pick-group:last-child { border-bottom: none; }
.service-pick-item { display: flex; justify-content: space-between; align-items: center; padding: 20rpx 0; }
.service-pick-arrow { font-size: 24rpx; color: #9CA3AF; margin-left: 8rpx; }
.price-rules { padding: 0 0 12rpx 24rpx; }
.price-rule-item { display: flex; justify-content: space-between; align-items: center; padding: 14rpx 16rpx; margin-bottom: 8rpx; background: #F9FAFB; border-radius: 12rpx; }
.price-rule-item:active { background: #EEF2FF; }
.rule-level { font-size: 26rpx; color: #1F2937; font-weight: 500; }
.rule-price { font-size: 24rpx; color: #6B7280; }
.service-pick-name { font-size: 28rpx; color: #1F2937; }
.service-pick-price { font-size: 24rpx; color: #6B7280; }
</style>
