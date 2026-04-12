<template>
  <SideLayout>
    <view class="page">
      <view class="hero-card">
        <view>
          <text class="hero-title">新建寄养</text>
          <text class="hero-subtitle">先选客户和猫咪，再按房间分组安排寄养，最后核价创建整单。</text>
        </view>
      </view>

      <view class="section-card">
        <view class="section-head">
          <view class="section-index">1</view>
          <view>
            <text class="section-title">客户与猫咪</text>
            <text class="section-desc">先确认是谁来住，再选择本次要寄养的猫咪。</text>
          </view>
        </view>

        <view class="mode-tabs">
          <view :class="['mode-tab', customerMode === 'regular' ? 'active' : '']" @click="customerMode = 'regular'">老客</view>
          <view :class="['mode-tab', customerMode === 'new' ? 'active' : '']" @click="customerMode = 'new'">新客</view>
        </view>

        <template v-if="customerMode === 'regular'">
          <view v-if="!lockCustomerSearch" class="search-row">
            <input v-model="customerKeyword" class="input search-input" placeholder="输入客户昵称或手机号搜索" @confirm="searchCustomers" />
            <view class="mini-btn" @click="searchCustomers">搜索</view>
          </view>

          <view v-if="customers.length > 0 && !lockCustomerSearch" class="stack-list">
            <view
              v-for="customer in customers"
              :key="customer.ID"
              :class="['select-card', selectedCustomer?.ID === customer.ID ? 'active' : '']"
              @click="selectCustomer(customer)"
            >
              <view>
                <text class="select-title">{{ customer.nickname || '未命名客户' }}</text>
                <text class="select-meta">{{ customer.phone || '未绑定手机号' }}</text>
              </view>
              <text class="select-mark">{{ selectedCustomer?.ID === customer.ID ? '已选' : '选择' }}</text>
            </view>
          </view>

          <view v-if="selectedCustomer" class="sub-block">
            <view class="sub-head">
              <view>
                <text class="sub-title">已选客户</text>
                <text class="sub-value">{{ selectedCustomer.nickname || selectedCustomer.phone }}</text>
              </view>
              <view v-if="lockCustomerSearch" class="mini-btn ghost" @click="resetSelectedCustomer">更换客户</view>
            </view>
            <view class="stack-list">
              <view class="pet-chip" v-for="pet in customerPets" :key="pet.ID" @click="togglePet(pet.ID)">
                <view :class="['pet-chip-inner', selectedPetIds.includes(pet.ID) ? 'active' : '']">
                  <text class="pet-chip-name">{{ pet.name }}</text>
                  <text class="pet-chip-mark">{{ selectedPetIds.includes(pet.ID) ? '已选' : '点选' }}</text>
                </view>
              </view>
            </view>
          </view>
        </template>

        <template v-else>
          <view class="field-grid single">
            <view class="field-card">
              <text class="field-label">客户昵称</text>
              <input v-model="newCustomer.nickname" class="input no-gap" placeholder="例如：可乐妈妈" />
            </view>
            <view class="field-card">
              <text class="field-label">手机号</text>
              <text class="field-tip">可选，不填也能先创建寄养单。</text>
              <input v-model="newCustomer.phone" class="input no-gap" type="number" placeholder="例如：13800000000" />
            </view>
          </view>

          <view class="sub-block">
            <view class="sub-head">
              <view>
                <text class="sub-title">猫咪名单</text>
                <text class="sub-value">{{ newPets.length }} 只猫，切换 tab 编辑</text>
              </view>
              <view class="mini-btn ghost" @click="addPetDraft">+ 添加猫咪</view>
            </view>

            <view class="draft-tab-list">
              <view
                v-for="(pet, index) in newPets"
                :key="pet.id"
                :class="['draft-tab', activeDraftPetId === pet.id ? 'active' : '']"
                @click="selectDraftPet(pet.id)"
              >
                <view class="draft-tab-copy">
                  <text class="draft-tab-title">{{ pet.name.trim() || `猫咪${index + 1}` }}</text>
                  <text class="draft-tab-meta">{{ pet.breed.trim() || '未填品种' }}</text>
                </view>
                <view v-if="newPets.length > 1" class="draft-tab-del" @click.stop="removePetDraft(pet.id)">删除</view>
              </view>
            </view>

            <view v-if="activeDraftPet" class="draft-card active-draft-card">
              <view class="draft-top">
                <text class="draft-title">{{ activeDraftPet.name.trim() || `编辑 ${activeDraftPetLabel}` }}</text>
                <text class="draft-hint">填写当前 tab 这只猫的信息</text>
              </view>
              <view class="field-grid">
                <view class="field-card compact">
                  <text class="field-label">猫咪名称</text>
                  <input v-model="activeDraftPet.name" class="input no-gap" placeholder="例如：汤圆" />
                </view>
                <view class="field-card compact">
                  <text class="field-label">品种</text>
                  <input v-model="activeDraftPet.breed" class="input no-gap" placeholder="可选，例如：英短蓝猫" />
                </view>
              </view>
            </view>
          </view>
        </template>
      </view>

      <view class="section-card">
        <view class="section-head">
          <view class="section-index">2</view>
          <view>
            <text class="section-title">入住日期与备注</text>
            <text class="section-desc">房间分组默认共用同一入住和离店日期，后续可单独续住或离店。</text>
          </view>
        </view>

        <view class="field-grid date-grid">
          <picker mode="date" :value="form.checkInAt" @change="handleDateChange('checkInAt', $event.detail.value)">
            <view class="picker-card">
              <text class="field-label">入住日期</text>
              <text class="picker-value">{{ form.checkInAt || '请选择日期' }}</text>
            </view>
          </picker>
          <picker mode="date" :value="form.checkOutAt" @change="handleDateChange('checkOutAt', $event.detail.value)">
            <view class="picker-card">
              <text class="field-label">离店日期</text>
              <text class="picker-value">{{ form.checkOutAt || '请选择日期' }}</text>
            </view>
          </picker>
        </view>

        <view class="summary-row">
          <view class="summary-pill">
            <text class="summary-label">入住猫咪</text>
            <text class="summary-value">{{ selectedPetOptions.length }} 只</text>
          </view>
          <view v-if="stayText" class="summary-pill">
            <text class="summary-label">寄养时长</text>
            <text class="summary-value">{{ stayText }}</text>
          </view>
        </view>

        <view class="field-card settings-card">
          <view class="settings-row">
            <view class="settings-copy">
              <text class="field-label">是否合住</text>
              <text class="field-tip">合住表示多只猫共享一个房间。</text>
            </view>
            <view class="option-row compact">
              <view :class="['option-pill compact', roomMode === 'shared' ? 'active' : '']" @click="setRoomMode('shared')">合住</view>
              <view :class="['option-pill compact', roomMode === 'split' ? 'active' : '']" @click="setRoomMode('split')">分房</view>
            </view>
          </view>
          <view class="settings-divider"></view>
          <view class="settings-row">
            <view class="settings-copy">
              <text class="field-label">是否已驱虫</text>
              <text class="field-tip">入住前确认一次，详情页会保留记录。</text>
            </view>
            <view class="option-row compact">
              <view :class="['option-pill compact', form.hasDeworming === true ? 'active' : '']" @click="form.hasDeworming = true">已驱虫</view>
              <view :class="['option-pill compact', form.hasDeworming === false ? 'active negative' : '']" @click="form.hasDeworming = false">未驱虫</view>
            </view>
          </view>
        </view>

        <view class="field-card remark-card">
          <text class="field-label">备注</text>
          <text class="field-tip">写饮食习惯、禁忌、应激情况、注意事项。</text>
          <textarea v-model="form.remark" class="textarea no-gap" placeholder="例如：晚上 9 点喂冻干，胆小，怕吹风机，不和陌生猫同放。" />
        </view>
      </view>

      <view class="section-card">
        <view class="section-head">
          <view class="section-index">3</view>
          <view>
            <text class="section-title">房间分组</text>
            <text class="section-desc">先把猫咪分到不同房间，再为每个房间选择寄养房型。</text>
          </view>
        </view>

        <view v-if="selectedPetOptions.length === 0" class="empty-box">先在上面选择猫咪，才能继续分房。</view>

        <template v-else>
          <view class="assign-block">
            <view class="sub-head">
              <text class="sub-title">猫咪分房</text>
              <view v-if="roomMode === 'split'" class="mini-btn ghost" @click="addRoomGroup">+ 新增房间</view>
            </view>
            <view class="assign-list">
              <view class="assign-chip" v-for="pet in selectedPetOptions" :key="pet.key" @click="openPetAssignment(pet)">
                <text class="assign-name">{{ pet.label }}</text>
                <text class="assign-room">{{ roomLabelById(petRoomAssignments[pet.key]) }}</text>
              </view>
            </view>
          </view>

          <view class="room-group-tabs">
            <view
              v-for="(group, index) in roomGroups"
              :key="group.id"
              :class="['room-group-tab', activeRoomGroupId === group.id ? 'active' : '']"
              @click="selectRoomGroup(group.id)"
            >
              <text class="room-group-tab-title">{{ roomLabel(index) }}</text>
              <text class="room-group-tab-meta">{{ roomPets(group.id).length }} 只猫</text>
            </view>
          </view>

          <view v-if="activeRoomGroup" class="room-group-card active-room-group-card">
            <view class="room-group-head">
              <view>
                <text class="room-group-title">{{ roomLabel(activeRoomGroupIndex - 1) }}</text>
                <text class="room-group-sub">{{ roomPets(activeRoomGroup.id).map((item) => item.label).join('、') || '先分配猫咪到这个房间' }}</text>
              </view>
              <view v-if="roomMode === 'split' && roomGroups.length > 1" class="draft-del" @click="removeRoomGroup(activeRoomGroup.id)">删除</view>
            </view>

            <view class="room-tags">
              <text class="room-tag">{{ roomPets(activeRoomGroup.id).length }} 只猫</text>
              <text v-if="activeRoomGroup.preview" class="room-tag price">¥{{ activeRoomGroup.preview.pay_amount.toFixed(2) }}</text>
            </view>

            <view v-if="roomPets(activeRoomGroup.id).length === 0" class="empty-mini">这个房间还没有分配猫咪。</view>

            <template v-else>
              <view class="stack-list">
                <view
                  v-for="cabinet in activeRoomGroup.availableCabinets"
                  :key="cabinet.ID"
                  :class="['room-card', activeRoomGroup.cabinetId === cabinet.ID ? 'active' : '']"
                  @click="selectCabinet(activeRoomGroup.id, cabinet.ID)"
                >
                  <view class="room-head">
                    <text class="room-title">{{ cabinet.cabinet_type }}</text>
                    <text class="room-mark">{{ activeRoomGroup.cabinetId === cabinet.ID ? '已选择' : '可选' }}</text>
                  </view>
                  <view class="room-tags">
                    <text class="room-tag stock">剩 {{ cabinet.remaining_rooms || 0 }}/{{ cabinet.room_count || 1 }} 间</text>
                    <text class="room-tag">每间 {{ cabinet.capacity }} 只</text>
                    <text class="room-tag price">¥{{ cabinet.base_price }}/晚</text>
                    <text v-if="cabinet.extra_pet_price > 0" class="room-tag extra">第二只 +¥{{ cabinet.extra_pet_price }}/晚</text>
                  </view>
                </view>
              </view>

              <view v-if="activeRoomGroup.preview" class="room-preview-box">
                <view class="room-preview-head">
                  <text class="room-preview-title">{{ roomLabel(activeRoomGroupIndex - 1) }} 预览</text>
                  <text class="room-preview-value">¥{{ activeRoomGroup.preview.pay_amount.toFixed(2) }}</text>
                </view>
                <view class="line-list compact-list">
                  <view v-for="line in activeRoomGroup.preview.lines" :key="`${activeRoomGroup.id}-${line.type}-${line.label}`" class="line-row">
                    <text class="line-name">{{ line.label }}</text>
                    <text class="line-amount">¥{{ line.amount.toFixed(2) }}</text>
                  </view>
                </view>
              </view>
            </template>
          </view>

          <view class="primary-action" @click="loadAvailableCabinets">查询各房间可用房型</view>
        </template>
      </view>

      <view v-if="policies.length > 0" class="section-card">
        <view class="section-head">
          <view class="section-index accent">4</view>
          <view>
            <text class="section-title">优惠策略</text>
            <text class="section-desc">默认全选可用优惠，你也可以手动取消。</text>
          </view>
        </view>

        <view class="stack-list">
          <view class="policy-card" v-for="policy in policies" :key="policy.ID" @click="togglePolicy(policy.ID)">
            <view>
              <text class="policy-name">{{ policy.name }}</text>
              <text class="policy-meta">{{ describePolicy(policy) }}</text>
            </view>
            <text :class="['policy-toggle', selectedPolicyIds.includes(policy.ID) ? 'active' : '']">
              {{ selectedPolicyIds.includes(policy.ID) ? '已用' : '不用' }}
            </text>
          </view>
        </view>
      </view>

      <view class="section-card preview-card">
        <view class="section-head">
          <view class="section-index accent">5</view>
          <view>
            <text class="section-title">金额预览</text>
            <text class="section-desc">整单统一收款，但会按房间拆分明细。</text>
          </view>
        </view>

        <view class="preview-trigger" @click="loadPreview">
          <text class="preview-trigger-text">{{ previewLoading ? '正在计算...' : '生成金额预览' }}</text>
        </view>

        <view v-if="preview" class="preview-box">
          <view class="preview-top">
            <view>
              <text class="preview-main">¥{{ preview.pay_amount.toFixed(2) }}</text>
              <text class="preview-sub">共 {{ preview.rooms?.length || roomGroups.length }} 个房间 · {{ preview.nights }} 晚</text>
            </view>
            <view class="preview-chip">{{ selectedPetOptions.length }} 只猫</view>
          </view>

          <view class="line-list">
            <view v-for="line in preview.lines" :key="`${line.type}-${line.label}`" class="line-row">
              <text class="line-name">{{ line.label }}</text>
              <text class="line-amount">¥{{ line.amount.toFixed(2) }}</text>
            </view>
          </view>
        </view>
      </view>

      <view class="section-card action-card">
        <view class="action-summary">
          <view>
            <text class="action-title">{{ roomSummaryText }}</text>
            <text class="action-sub">{{ selectedPetOptions.length }} 只猫{{ stayText ? ` · ${stayText}` : '' }}</text>
          </view>
          <text v-if="preview" class="action-price">¥{{ preview.pay_amount.toFixed(2) }}</text>
        </view>
        <button class="submit-btn block" :loading="submitting" @click="submit">确认创建</button>
      </view>

      <view v-if="assignmentModalPet" class="assign-modal-mask" @click="closePetAssignment">
        <view class="assign-modal" @click.stop>
          <view class="assign-modal-head">
            <view>
              <text class="assign-modal-title">{{ assignmentModalPet.label }}</text>
              <text class="assign-modal-subtitle">选择要分配到哪个房间</text>
            </view>
            <view class="assign-modal-close" @click="closePetAssignment">关闭</view>
          </view>

          <view class="stack-list">
            <view
              v-for="group in roomGroups"
              :key="group.id"
              :class="['assign-option', petRoomAssignments[assignmentModalPet.key] === group.id ? 'active' : '']"
              @click="assignPetToRoom(group.id)"
            >
              <view>
                <text class="assign-option-title">{{ roomLabelById(group.id) }}</text>
                <text class="assign-option-meta">{{ roomPets(group.id).map((item) => item.label).join('、') || '这个房间还没有猫咪' }}</text>
              </view>
              <text class="assign-option-mark">{{ petRoomAssignments[assignmentModalPet.key] === group.id ? '已选' : '选择' }}</text>
            </view>

            <view class="assign-option create" @click="assignPetToNewRoom">
              <view>
                <text class="assign-option-title">新增房间</text>
                <text class="assign-option-meta">新建一个房间，并把这只猫放进去</text>
              </view>
              <text class="assign-option-mark">新增</text>
            </view>
          </view>
        </view>
      </view>
    </view>
  </SideLayout>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import SideLayout from '@/components/SideLayout.vue'
import { getCustomerList, createCustomer, getCustomerPets } from '@/api/customer'
import { createPet } from '@/api/pet'
import {
  createBoardingOrder,
  getAvailableBoardingCabinets,
  getBoardingPolicies,
  previewBoardingOrder,
} from '@/api/boarding'

type CustomerMode = 'regular' | 'new'
type RoomMode = 'shared' | 'split'
type DraftPet = { id: number; name: string; breed: string }
type RoomGroupState = {
  id: number
  cabinetId: number
  availableCabinets: BoardingCabinet[]
  preview: BoardingRoomPreview | null
}
type SelectedPetOption = {
  key: string
  label: string
  petId?: number
  breed?: string
}

const customerMode = ref<CustomerMode>('regular')
const roomMode = ref<RoomMode>('shared')
const customerKeyword = ref('')
const customers = ref<Customer[]>([])
const selectedCustomer = ref<Customer | null>(null)
const customerPets = ref<Pet[]>([])
const selectedPetIds = ref<number[]>([])
const newCustomer = ref({ nickname: '', phone: '' })
const newPets = ref<DraftPet[]>([{ id: Date.now(), name: '', breed: '' }])
const activeDraftPetId = ref<number | null>(newPets.value[0]?.id || null)

const form = ref({
  checkInAt: '',
  checkOutAt: '',
  hasDeworming: null as boolean | null,
  remark: '',
})

const preview = ref<BoardingPricePreview | null>(null)
const previewLoading = ref(false)
const submitting = ref(false)
const policies = ref<BoardingDiscountPolicy[]>([])
const selectedPolicyIds = ref<number[]>([])
const roomGroups = ref<RoomGroupState[]>([])
const activeRoomGroupId = ref<number | null>(null)
const petRoomAssignments = ref<Record<string, number>>({})
const assignmentModalPet = ref<SelectedPetOption | null>(null)

let roomGroupSeed = 1

const selectedPetOptions = computed<SelectedPetOption[]>(() => {
  if (customerMode.value === 'regular') {
    return customerPets.value
      .filter((pet) => selectedPetIds.value.includes(pet.ID))
      .map((pet) => ({
        key: `pet-${pet.ID}`,
        label: pet.name || `猫咪${pet.ID}`,
        petId: pet.ID,
        breed: pet.breed,
      }))
  }
  return newPets.value
    .filter((pet) => pet.name.trim())
    .map((pet) => ({
      key: `draft-${pet.id}`,
      label: pet.name.trim(),
      breed: pet.breed.trim(),
    }))
})

const stayText = computed(() => {
  if (!form.value.checkInAt || !form.value.checkOutAt) return ''
  const start = new Date(form.value.checkInAt)
  const end = new Date(form.value.checkOutAt)
  const diff = end.getTime() - start.getTime()
  if (Number.isNaN(diff) || diff <= 0) return ''
  return `${Math.round(diff / (24 * 60 * 60 * 1000))} 晚`
})

const activeDraftPet = computed(() => newPets.value.find((pet) => pet.id === activeDraftPetId.value) || newPets.value[0] || null)

const activeDraftPetLabel = computed(() => {
  const activePet = activeDraftPet.value
  if (!activePet) return '猫咪'
  const index = newPets.value.findIndex((pet) => pet.id === activePet.id)
  return index >= 0 ? `猫咪${index + 1}` : '猫咪'
})

const lockCustomerSearch = computed(() => customerMode.value === 'regular'
  && !!selectedCustomer.value
  && selectedPetIds.value.length > 0)

const activeRoomGroup = computed(() => roomGroups.value.find((group) => group.id === activeRoomGroupId.value) || roomGroups.value[0] || null)

const activeRoomGroupIndex = computed(() => {
  const group = activeRoomGroup.value
  if (!group) return 1
  const index = roomGroups.value.findIndex((item) => item.id === group.id)
  return index >= 0 ? index + 1 : 1
})

const roomSummaryText = computed(() => {
  if (roomGroups.value.length === 0) return '还没分房'
  return roomGroups.value
    .map((group, index) => `${roomLabel(index)}${group.cabinetId ? ` · ${cabinetName(group.cabinetId, group)}` : ''}`)
    .join(' / ')
})

watch(
  () => selectedPetOptions.value.map((item) => item.key).join('|'),
  () => {
    syncRoomGroups()
    clearRoomSelections()
  },
  { immediate: true }
)

watch(
  () => `${form.value.checkInAt}|${form.value.checkOutAt}`,
  () => {
    clearRoomSelections()
  }
)

watch(
  () => newPets.value.map((pet) => pet.id).join('|'),
  () => {
    if (!newPets.value.some((pet) => pet.id === activeDraftPetId.value)) {
      activeDraftPetId.value = newPets.value[0]?.id || null
    }
  },
  { immediate: true }
)

watch(
  () => roomGroups.value.map((group) => group.id).join('|'),
  () => {
    if (!roomGroups.value.some((group) => group.id === activeRoomGroupId.value)) {
      activeRoomGroupId.value = roomGroups.value[0]?.id || null
    }
  },
  { immediate: true }
)

function createRoomGroupState(): RoomGroupState {
  roomGroupSeed += 1
  return {
    id: roomGroupSeed,
    cabinetId: 0,
    availableCabinets: [],
    preview: null,
  }
}

function roomLabel(index: number) {
  return `房间${index + 1}`
}

function roomLabelById(groupId?: number) {
  const index = roomGroups.value.findIndex((item) => item.id === groupId)
  return index >= 0 ? roomLabel(index) : '未分房'
}

function cabinetName(cabinetId: number, group: RoomGroupState) {
  return group.availableCabinets.find((item) => item.ID === cabinetId)?.cabinet_type || '未选房型'
}

function roomPets(groupId: number) {
  return selectedPetOptions.value.filter((pet) => petRoomAssignments.value[pet.key] === groupId)
}

function syncRoomGroups() {
  if (selectedPetOptions.value.length === 0) {
    roomGroups.value = []
    activeRoomGroupId.value = null
    petRoomAssignments.value = {}
    closePetAssignment()
    return
  }

  if (roomGroups.value.length === 0) {
    roomGroups.value = [{ id: 1, cabinetId: 0, availableCabinets: [], preview: null }]
    roomGroupSeed = 1
    activeRoomGroupId.value = roomGroups.value[0].id
  }

  const validRoomIds = new Set(roomGroups.value.map((item) => item.id))
  const nextAssignments: Record<string, number> = {}
  const firstRoomId = roomGroups.value[0].id

  if (roomMode.value === 'shared') {
    const first = roomGroups.value[0]
    roomGroups.value = [{ ...first, preview: null }]
    activeRoomGroupId.value = roomGroups.value[0].id
    for (const pet of selectedPetOptions.value) {
      nextAssignments[pet.key] = roomGroups.value[0].id
    }
    petRoomAssignments.value = nextAssignments
    return
  }

  const nextGroups = [...roomGroups.value]
  const usedRoomIDs = new Set<number>()
  for (const pet of selectedPetOptions.value) {
    const assignedRoomId = petRoomAssignments.value[pet.key]
    if (validRoomIds.has(assignedRoomId)) {
      nextAssignments[pet.key] = assignedRoomId
      usedRoomIDs.add(assignedRoomId)
      continue
    }

    const unusedGroup = nextGroups.find((group) => !usedRoomIDs.has(group.id))
    if (unusedGroup) {
      nextAssignments[pet.key] = unusedGroup.id
      usedRoomIDs.add(unusedGroup.id)
      continue
    }

    const createdGroup = createRoomGroupState()
    nextGroups.push(createdGroup)
    nextAssignments[pet.key] = createdGroup.id
    usedRoomIDs.add(createdGroup.id)
  }
  petRoomAssignments.value = nextAssignments

  const usedRoomIds = new Set(Object.values(nextAssignments))
  roomGroups.value = nextGroups.filter((group) => usedRoomIds.has(group.id))
  if (!roomGroups.value.some((group) => group.id === activeRoomGroupId.value)) {
    activeRoomGroupId.value = roomGroups.value[0]?.id || null
  }
}

function clearRoomSelections() {
  preview.value = null
  roomGroups.value = roomGroups.value.map((group) => ({
    ...group,
    cabinetId: 0,
    availableCabinets: [],
    preview: null,
  }))
}

function setRoomMode(mode: RoomMode) {
  roomMode.value = mode
  if (mode === 'split') {
    splitPetsIntoSeparateRooms()
    return
  }
  syncRoomGroups()
  activeRoomGroupId.value = roomGroups.value[0]?.id || null
}

function splitPetsIntoSeparateRooms() {
  if (selectedPetOptions.value.length === 0) {
    roomGroups.value = []
    petRoomAssignments.value = {}
    closePetAssignment()
    return
  }

  const nextGroups = selectedPetOptions.value.map((_, index) => {
    const existing = roomGroups.value[index]
    if (existing) {
      return {
        ...existing,
        cabinetId: 0,
        availableCabinets: [],
        preview: null,
      }
    }
    return createRoomGroupState()
  })

  const nextAssignments: Record<string, number> = {}
  selectedPetOptions.value.forEach((pet, index) => {
    nextAssignments[pet.key] = nextGroups[index].id
  })

  roomGroups.value = nextGroups
  activeRoomGroupId.value = nextGroups[0]?.id || null
  petRoomAssignments.value = nextAssignments
  preview.value = null
  closePetAssignment()
}

async function searchCustomers() {
  if (!customerKeyword.value.trim()) {
    uni.showToast({ title: '请输入关键词', icon: 'none' })
    return
  }
  const res = await getCustomerList({ page: 1, page_size: 20, keyword: customerKeyword.value.trim() })
  customers.value = res.data.list || []
}

async function selectCustomer(customer: Customer) {
  selectedCustomer.value = customer
  selectedPetIds.value = []
  const res = await getCustomerPets(customer.ID)
  customerPets.value = res.data || []
}

function resetSelectedCustomer() {
  selectedCustomer.value = null
  customerPets.value = []
  selectedPetIds.value = []
  preview.value = null
}

function togglePet(petId: number) {
  const idx = selectedPetIds.value.indexOf(petId)
  if (idx >= 0) selectedPetIds.value.splice(idx, 1)
  else selectedPetIds.value.push(petId)
}

function addPetDraft() {
  const id = Date.now() + Math.floor(Math.random() * 1000)
  newPets.value.push({ id, name: '', breed: '' })
  activeDraftPetId.value = id
}

function removePetDraft(id: number) {
  if (newPets.value.length <= 1) return
  newPets.value = newPets.value.filter((pet) => pet.id !== id)
  if (activeDraftPetId.value === id) {
    activeDraftPetId.value = newPets.value[0]?.id || null
  }
}

function selectDraftPet(id: number) {
  activeDraftPetId.value = id
}

function addRoomGroup() {
  if (roomMode.value !== 'split') return
  const group = createRoomGroupState()
  roomGroups.value.push(group)
  activeRoomGroupId.value = group.id
}

function removeRoomGroup(groupId: number) {
  if (roomGroups.value.length <= 1) return
  const target = roomGroups.value.find((item) => item.id === groupId)
  if (!target) return
  const fallbackRoomId = roomGroups.value.find((item) => item.id !== groupId)?.id || roomGroups.value[0].id
  for (const pet of selectedPetOptions.value) {
    if (petRoomAssignments.value[pet.key] === groupId) {
      petRoomAssignments.value[pet.key] = fallbackRoomId
    }
  }
  roomGroups.value = roomGroups.value.filter((item) => item.id !== groupId)
  activeRoomGroupId.value = fallbackRoomId
  syncRoomGroups()
}

function selectRoomGroup(groupId: number) {
  activeRoomGroupId.value = groupId
}

function openPetAssignment(pet: SelectedPetOption) {
  if (roomMode.value === 'shared') return
  assignmentModalPet.value = pet
}

function closePetAssignment() {
  assignmentModalPet.value = null
}

function assignPetToRoom(groupId: number) {
  const pet = assignmentModalPet.value
  if (!pet) return
  const previousGroupId = petRoomAssignments.value[pet.key]
  petRoomAssignments.value[pet.key] = groupId
  activeRoomGroupId.value = groupId
  clearGroupPreviewByRoomIds([previousGroupId, groupId])
  closePetAssignment()
}

function assignPetToNewRoom() {
  const pet = assignmentModalPet.value
  if (!pet) return
  const group = createRoomGroupState()
  const previousGroupId = petRoomAssignments.value[pet.key]
  roomGroups.value.push(group)
  petRoomAssignments.value[pet.key] = group.id
  activeRoomGroupId.value = group.id
  clearGroupPreviewByRoomIds([previousGroupId, group.id])
  closePetAssignment()
}

function clearGroupPreviewByRoomIds(groupIds: Array<number | undefined>) {
  const idSet = new Set(groupIds.filter(Boolean) as number[])
  roomGroups.value = roomGroups.value.map((group) => (
    idSet.has(group.id) ? { ...group, preview: null, availableCabinets: [], cabinetId: 0 } : group
  ))
  preview.value = null
}

function handleDateChange(field: 'checkInAt' | 'checkOutAt', value: string) {
  form.value[field] = value
}

function togglePolicy(id: number) {
  const idx = selectedPolicyIds.value.indexOf(id)
  if (idx >= 0) selectedPolicyIds.value.splice(idx, 1)
  else selectedPolicyIds.value.push(id)
}

function describePolicy(policy: BoardingDiscountPolicy) {
  try {
    const rule = JSON.parse(policy.rule_json || '{}')
    if (policy.policy_type === 'stay_n_free_m') {
      return `住 ${rule.stay || 0} 免 ${rule.free || 0}`
    }
    if (policy.policy_type === 'holiday_surcharge') {
      return `节假日每晚 +¥${rule.surcharge || 0}`
    }
  } catch {}
  return policy.remark || '寄养优惠'
}

function validateCoreFields() {
  if (!form.value.checkInAt || !form.value.checkOutAt) {
    uni.showToast({ title: '请选择入住和离店日期', icon: 'none' })
    return false
  }
  if (selectedPetOptions.value.length < 1) {
    uni.showToast({ title: '请至少选择一只猫咪', icon: 'none' })
    return false
  }
  if (customerMode.value === 'regular' && !selectedCustomer.value) {
    uni.showToast({ title: '请选择客户', icon: 'none' })
    return false
  }
  if (customerMode.value === 'new' && !newCustomer.value.nickname.trim()) {
    uni.showToast({ title: '请填写客户昵称', icon: 'none' })
    return false
  }
  return true
}

function validateRoomGroups(requireCabinet = true) {
  if (!validateCoreFields()) return false
  if (roomGroups.value.length === 0) {
    uni.showToast({ title: '请先分配房间', icon: 'none' })
    return false
  }
  for (let index = 0; index < roomGroups.value.length; index += 1) {
    const group = roomGroups.value[index]
    if (roomPets(group.id).length === 0) {
      uni.showToast({ title: `${roomLabel(index)} 还没分配猫咪`, icon: 'none' })
      return false
    }
    if (requireCabinet && !group.cabinetId) {
      uni.showToast({ title: `${roomLabel(index)} 还没选房型`, icon: 'none' })
      return false
    }
  }
  return true
}

function buildRoomGroupsPayload(mode: 'preview' | 'submit', petIdMap?: Record<string, number>) {
  return roomGroups.value.map((group) => {
    const pets = roomPets(group.id)
    const payload: {
      pet_ids?: number[]
      pet_count?: number
      cabinet_id: number
      check_in_at: string
      check_out_at: string
    } = {
      cabinet_id: group.cabinetId,
      check_in_at: form.value.checkInAt,
      check_out_at: form.value.checkOutAt,
    }
    if (mode === 'submit') {
      payload.pet_ids = pets.map((pet) => petIdMap?.[pet.key] || pet.petId || 0).filter(Boolean)
    } else if (customerMode.value === 'regular') {
      payload.pet_ids = pets.map((pet) => pet.petId || 0).filter(Boolean)
    } else {
      payload.pet_count = pets.length
    }
    return payload
  })
}

async function ensurePoliciesLoaded() {
  if (policies.value.length > 0) return
  const res = await getBoardingPolicies()
  policies.value = (res.data || []).filter((item) => item.status === 1)
  selectedPolicyIds.value = policies.value.map((item) => item.ID)
}

async function loadAvailableCabinets() {
  if (!validateRoomGroups(false)) return
  await ensurePoliciesLoaded()
  preview.value = null
  const nextGroups: RoomGroupState[] = []
  for (const group of roomGroups.value) {
    const pets = roomPets(group.id)
    const res = await getAvailableBoardingCabinets({
      check_in_at: form.value.checkInAt,
      check_out_at: form.value.checkOutAt,
      pet_count: pets.length,
    })
    const list = res.data || []
    nextGroups.push({
      ...group,
      availableCabinets: list,
      cabinetId: list.find((item) => item.ID === group.cabinetId) ? group.cabinetId : (list[0]?.ID || 0),
      preview: null,
    })
  }
  roomGroups.value = nextGroups
}

function selectCabinet(groupId: number, cabinetId: number) {
  roomGroups.value = roomGroups.value.map((group) => (
    group.id === groupId ? { ...group, cabinetId, preview: null } : group
  ))
  preview.value = null
}

async function loadPreview() {
  if (!validateRoomGroups(true)) return
  previewLoading.value = true
  try {
    const payload = {
      customer_id: customerMode.value === 'regular' ? selectedCustomer.value?.ID : undefined,
      policy_ids: selectedPolicyIds.value,
      room_groups: buildRoomGroupsPayload('preview'),
    }
    const res = await previewBoardingOrder(payload)
    preview.value = res.data
    roomGroups.value = roomGroups.value.map((group, index) => ({
      ...group,
      preview: res.data.rooms?.find((item) => item.room_index === index + 1) || null,
    }))
  } finally {
    previewLoading.value = false
  }
}

async function submit() {
  if (!preview.value) {
    await loadPreview()
    if (!preview.value) return
  }
  if (form.value.hasDeworming === null) {
    uni.showToast({ title: '请选择是否已驱虫', icon: 'none' })
    return
  }

  submitting.value = true
  try {
    let customerId = selectedCustomer.value?.ID || 0
    const petIdMap: Record<string, number> = {}

    if (customerMode.value === 'new') {
      const customerRes = await createCustomer({
        nickname: newCustomer.value.nickname.trim(),
        phone: newCustomer.value.phone.trim(),
      })
      customerId = customerRes.data.ID
      for (const pet of newPets.value.filter((item) => item.name.trim())) {
        const petRes = await createPet({
          customer_id: customerId,
          name: pet.name.trim(),
          breed: pet.breed.trim(),
          species: '猫',
        })
        petIdMap[`draft-${pet.id}`] = petRes.data.ID
      }
    } else {
      for (const pet of selectedPetOptions.value) {
        if (pet.petId) petIdMap[pet.key] = pet.petId
      }
    }

    const res = await createBoardingOrder({
      customer_id: customerId,
      policy_ids: selectedPolicyIds.value,
      room_groups: buildRoomGroupsPayload('submit', petIdMap),
      has_deworming: form.value.hasDeworming,
      remark: form.value.remark,
    })
    uni.showToast({ title: '寄养单已创建', icon: 'success' })
    setTimeout(() => {
      uni.redirectTo({ url: `/pages/boarding/detail?id=${res.data.ID}` })
    }, 400)
  } finally {
    submitting.value = false
  }
}

ensurePoliciesLoaded()
</script>

<style scoped>
.page {
  padding: 24rpx 24rpx calc(32rpx + env(safe-area-inset-bottom));
  display: flex;
  flex-direction: column;
  gap: 20rpx;
  background:
    radial-gradient(circle at top left, rgba(253, 230, 138, 0.38), transparent 34%),
    linear-gradient(180deg, #fffdf8 0%, #f7f8ff 54%, #f9fafb 100%);
}
.hero-card,
.section-card {
  background: rgba(255, 255, 255, 0.94);
  border: 1rpx solid rgba(226, 232, 240, 0.9);
  border-radius: 28rpx;
  padding: 24rpx;
  box-shadow: 0 18rpx 44rpx rgba(15, 23, 42, 0.06);
}
.hero-title { display: block; font-size: 38rpx; font-weight: 700; color: #111827; }
.hero-subtitle { display: block; margin-top: 10rpx; font-size: 23rpx; line-height: 1.65; color: #6b7280; }
.section-head { display: flex; gap: 16rpx; align-items: flex-start; margin-bottom: 20rpx; }
.section-index { width: 42rpx; height: 42rpx; border-radius: 50%; background: #eef2ff; color: #4f46e5; font-size: 24rpx; font-weight: 700; display: flex; align-items: center; justify-content: center; flex-shrink: 0; margin-top: 2rpx; }
.section-index.accent { background: #fef3c7; color: #d97706; }
.section-title { display: block; font-size: 30rpx; font-weight: 700; color: #111827; }
.section-desc { display: block; margin-top: 6rpx; font-size: 22rpx; line-height: 1.6; color: #6b7280; }
.mode-tabs { display: inline-flex; gap: 10rpx; padding: 8rpx; border-radius: 999rpx; background: #f3f4f6; margin-bottom: 20rpx; }
.mode-tab { min-width: 116rpx; padding: 14rpx 22rpx; border-radius: 999rpx; text-align: center; font-size: 24rpx; color: #6b7280; }
.mode-tab.active { background: linear-gradient(135deg, #4f46e5, #6366f1); color: #fff; box-shadow: 0 8rpx 18rpx rgba(79, 70, 229, 0.24); }
.search-row { display: flex; gap: 12rpx; align-items: stretch; }
.field-grid { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 14rpx; }
.field-grid.single { grid-template-columns: repeat(1, minmax(0, 1fr)); }
.date-grid { grid-template-columns: repeat(2, minmax(0, 1fr)); }
.field-card, .picker-card { padding: 24rpx; background: #fbfcff; border: 1rpx solid #e5e7eb; border-radius: 20rpx; }
.field-card.compact { padding: 18rpx; }
.field-label { display: block; font-size: 24rpx; font-weight: 600; color: #374151; }
.field-tip { display: block; margin-top: 6rpx; font-size: 22rpx; line-height: 1.5; color: #9ca3af; }
.input, .textarea { display: block; width: 100%; box-sizing: border-box; margin-top: 16rpx; padding: 18rpx 20rpx; border: 2rpx solid #dbe3f0; border-radius: 18rpx; background: #fff; font-size: 29rpx; line-height: 1.4; min-height: 88rpx; height: 88rpx; color: #111827; }
.no-gap { margin-top: 12rpx; }
.textarea { min-height: 168rpx; height: 168rpx; line-height: 1.7; }
.search-input { margin-top: 0; flex: 1; }
.mini-btn { min-width: 108rpx; padding: 0 22rpx; border-radius: 18rpx; background: linear-gradient(135deg, #4f46e5, #6366f1); color: #fff; font-size: 24rpx; display: flex; align-items: center; justify-content: center; }
.mini-btn.ghost { min-width: auto; padding: 12rpx 18rpx; background: #eef2ff; color: #4f46e5; }
.sub-block, .assign-block { margin-top: 18rpx; padding: 18rpx; border-radius: 22rpx; background: linear-gradient(180deg, #fcfcff, #f8fafc); border: 1rpx solid #eef2f7; }
.sub-head { display: flex; justify-content: space-between; gap: 16rpx; align-items: center; margin-bottom: 14rpx; }
.sub-title { font-size: 24rpx; font-weight: 600; color: #111827; }
.sub-value { font-size: 22rpx; color: #6b7280; }
.stack-list, .line-list { display: flex; flex-direction: column; gap: 12rpx; }
.select-card, .policy-card, .room-card { display: flex; justify-content: space-between; gap: 16rpx; align-items: center; padding: 18rpx 20rpx; border-radius: 20rpx; border: 1rpx solid #e5e7eb; background: #fafbff; }
.select-card.active, .room-card.active { border-color: #818cf8; background: linear-gradient(135deg, #eef2ff, #f8faff); }
.select-title, .policy-name, .room-title { display: block; font-size: 26rpx; font-weight: 600; color: #111827; }
.select-meta, .policy-meta, .room-remark { display: block; margin-top: 6rpx; font-size: 22rpx; color: #6b7280; }
.select-mark, .policy-toggle, .room-mark { padding: 10rpx 16rpx; border-radius: 999rpx; background: #eef2ff; color: #4f46e5; font-size: 22rpx; white-space: nowrap; }
.policy-toggle.active { background: #4f46e5; color: #fff; }
.pet-chip-inner { display: flex; justify-content: space-between; align-items: center; padding: 18rpx 20rpx; border-radius: 18rpx; background: #fff; border: 1rpx solid #e5e7eb; }
.pet-chip-inner.active { background: linear-gradient(135deg, #fef3c7, #fff7ed); border-color: #f59e0b; }
.pet-chip-name { font-size: 25rpx; font-weight: 600; color: #111827; }
.pet-chip-mark { font-size: 22rpx; color: #92400e; }
.draft-card, .room-group-card { padding: 18rpx; border-radius: 22rpx; background: #fffdf8; border: 1rpx solid #f3e8d2; }
.active-draft-card,
.active-room-group-card { margin-top: 16rpx; }
.draft-top, .room-group-head { display: flex; justify-content: space-between; align-items: center; gap: 16rpx; margin-bottom: 14rpx; }
.draft-title, .room-group-title { font-size: 26rpx; font-weight: 700; color: #111827; }
.draft-del { font-size: 22rpx; color: #ef4444; }
.draft-hint { display: block; margin-top: 6rpx; font-size: 21rpx; color: #9ca3af; }
.draft-tab-list,
.room-group-tabs {
  display: flex;
  gap: 12rpx;
  overflow-x: auto;
  padding-bottom: 6rpx;
}
.draft-tab,
.room-group-tab {
  flex: 0 0 auto;
  min-width: 176rpx;
  padding: 16rpx 18rpx;
  border-radius: 20rpx;
  border: 1rpx solid #e5e7eb;
  background: #fff;
}
.draft-tab.active,
.room-group-tab.active {
  border-color: #818cf8;
  background: linear-gradient(135deg, #eef2ff, #f8faff);
  box-shadow: 0 10rpx 24rpx rgba(99, 102, 241, 0.12);
}
.draft-tab-copy {
  display: flex;
  flex-direction: column;
  gap: 6rpx;
}
.draft-tab-title,
.room-group-tab-title {
  font-size: 24rpx;
  font-weight: 700;
  color: #111827;
}
.draft-tab-meta,
.room-group-tab-meta {
  font-size: 21rpx;
  color: #6b7280;
}
.draft-tab-del {
  margin-top: 10rpx;
  font-size: 21rpx;
  color: #ef4444;
}
.room-group-sub { display: block; margin-top: 6rpx; font-size: 22rpx; color: #6b7280; line-height: 1.5; }
.summary-row, .option-row, .room-tags, .assign-list { display: flex; flex-wrap: wrap; gap: 12rpx; margin-top: 18rpx; }
.summary-pill, .room-tag, .assign-chip { padding: 12rpx 18rpx; border-radius: 999rpx; background: #f8fafc; font-size: 22rpx; color: #475569; display: inline-flex; gap: 10rpx; align-items: center; }
.summary-label { color: #94a3b8; }
.summary-value, .assign-name { color: #111827; font-weight: 600; }
.assign-room { color: #4f46e5; font-weight: 600; }
.option-pill { flex: 1; min-width: 180rpx; padding: 18rpx 16rpx; border-radius: 18rpx; text-align: center; background: #f3f4f6; color: #6b7280; font-size: 24rpx; }
.option-pill.active { background: linear-gradient(135deg, #4f46e5, #6366f1); color: #fff; }
.option-pill.active.negative { background: linear-gradient(135deg, #ef4444, #f97316); }
.option-row.compact {
  margin-top: 0;
  flex-wrap: nowrap;
}
.option-pill.compact {
  min-width: 116rpx;
  padding: 14rpx 16rpx;
  font-size: 22rpx;
}
.settings-card {
  display: flex;
  flex-direction: column;
  gap: 16rpx;
}
.settings-row {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 18rpx;
}
.settings-copy {
  min-width: 0;
  flex: 1;
}
.settings-divider {
  height: 1rpx;
  background: #e5e7eb;
}
.remark-card {
  margin-top: 18rpx;
}
.empty-box, .empty-mini { padding: 28rpx 24rpx; border-radius: 20rpx; background: #f8fafc; color: #94a3b8; font-size: 24rpx; text-align: center; }
.empty-mini { margin-top: 12rpx; }
.primary-action, .preview-trigger { margin-top: 18rpx; padding: 24rpx; border-radius: 22rpx; background: linear-gradient(135deg, #4f46e5, #6366f1); text-align: center; box-shadow: 0 18rpx 30rpx rgba(79, 70, 229, 0.18); }
.preview-trigger-text { font-size: 28rpx; font-weight: 700; color: #fff; }
.preview-box, .room-preview-box { margin-top: 18rpx; padding: 20rpx; border-radius: 22rpx; background: #fff; border: 1rpx solid #e5e7eb; }
.preview-top, .room-preview-head { display: flex; justify-content: space-between; gap: 16rpx; align-items: flex-start; }
.preview-main { display: block; font-size: 40rpx; font-weight: 800; color: #111827; }
.preview-sub { display: block; margin-top: 8rpx; font-size: 22rpx; color: #6b7280; }
.preview-chip, .room-preview-value { padding: 10rpx 16rpx; border-radius: 999rpx; background: #eef2ff; color: #4f46e5; font-size: 22rpx; font-weight: 600; }
.room-preview-title { font-size: 24rpx; font-weight: 700; color: #111827; }
.compact-list { margin-top: 12rpx; }
.line-row { display: flex; justify-content: space-between; gap: 16rpx; align-items: center; padding: 12rpx 0; }
.line-name, .line-amount { font-size: 24rpx; color: #374151; }
.assign-modal-mask {
  position: fixed;
  inset: 0;
  background: rgba(15, 23, 42, 0.48);
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 24rpx;
  z-index: 2200;
}
.assign-modal {
  width: min(100%, 640rpx);
  max-height: calc(100vh - 120rpx);
  overflow: auto;
  border-radius: 28rpx;
  background: #fff;
  padding: 24rpx;
  box-shadow: 0 28rpx 60rpx rgba(15, 23, 42, 0.28);
}
.assign-modal-head {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 16rpx;
  margin-bottom: 18rpx;
}
.assign-modal-title {
  display: block;
  font-size: 30rpx;
  font-weight: 700;
  color: #111827;
}
.assign-modal-subtitle {
  display: block;
  margin-top: 8rpx;
  font-size: 22rpx;
  color: #6b7280;
}
.assign-modal-close {
  padding: 10rpx 16rpx;
  border-radius: 999rpx;
  background: #f3f4f6;
  color: #6b7280;
  font-size: 22rpx;
  white-space: nowrap;
}
.assign-option {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 16rpx;
  padding: 20rpx;
  border-radius: 22rpx;
  border: 1rpx solid #e5e7eb;
  background: #fbfcff;
}
.assign-option.active {
  border-color: #818cf8;
  background: linear-gradient(135deg, #eef2ff, #f8faff);
}
.assign-option.create {
  border-style: dashed;
}
.assign-option-title {
  display: block;
  font-size: 26rpx;
  font-weight: 600;
  color: #111827;
}
.assign-option-meta {
  display: block;
  margin-top: 8rpx;
  font-size: 22rpx;
  line-height: 1.5;
  color: #6b7280;
}
.assign-option-mark {
  padding: 10rpx 16rpx;
  border-radius: 999rpx;
  background: #eef2ff;
  color: #4f46e5;
  font-size: 22rpx;
  white-space: nowrap;
}
.action-card {
  gap: 0;
}
.action-summary {
  display: flex;
  justify-content: space-between;
  gap: 16rpx;
  align-items: flex-start;
}
.action-title {
  display: block;
  font-size: 27rpx;
  font-weight: 700;
  color: #111827;
}
.action-sub {
  display: block;
  margin-top: 6rpx;
  font-size: 22rpx;
  color: #6b7280;
}
.action-price {
  font-size: 34rpx;
  font-weight: 800;
  color: #f59e0b;
}
.submit-btn { min-width: 220rpx; height: 88rpx; margin: 0; border: none; border-radius: 22rpx; background: linear-gradient(135deg, #f59e0b, #fb923c); color: #fff; font-size: 28rpx; font-weight: 700; display: flex; align-items: center; justify-content: center; box-shadow: none; }
.submit-btn::after {
  border: none;
}
.submit-btn.block {
  width: 100%;
  margin-top: 18rpx;
}
@media (max-width: 768px) {
  .field-grid { grid-template-columns: repeat(1, minmax(0, 1fr)); }
  .date-grid { grid-template-columns: repeat(2, minmax(0, 1fr)); }
  .settings-row {
    flex-direction: column;
  }
  .option-row.compact {
    width: 100%;
  }
  .option-pill.compact {
    flex: 1;
  }
  .assign-modal {
    width: 100%;
    max-height: calc(100vh - 96rpx);
    padding: 22rpx;
  }
}
</style>
