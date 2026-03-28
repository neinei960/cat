<template>
  <SideLayout>
  <view class="page">
    <!-- Header -->
    <view class="page-header">
      <text class="page-title">{{ isEditMode ? '🛠️ 修改预约' : '🐱 新建预约' }}</text>
      <text class="page-subtitle">{{ isEditMode ? '调整当前预约的客户、服务和时间安排' : '为毛孩子安排一次舒适的洗护体验' }}</text>
    </view>

    <!-- Step indicator -->
    <view class="steps-wrapper">
      <view class="steps">
        <view v-for="(s, i) in [{n:1, icon:'👤', label:'客户'}, {n:2, icon:'✂️', label:'服务'}, {n:3, icon:'📅', label:'时间'}, {n:4, icon:'✅', label:'确认'}]" :key="s.n"
          :class="['step-item', step >= s.n ? 'active' : '', step === s.n ? 'current' : '']">
          <view class="step-circle">
            <text v-if="step > s.n" class="step-check">✓</text>
            <text v-else class="step-icon">{{ s.icon }}</text>
          </view>
          <text class="step-label">{{ s.label }}</text>
          <view v-if="i < 3" :class="['step-line', step > s.n ? 'line-active' : '']"></view>
        </view>
      </view>
    </view>

    <!-- Step 1: Customer & Pet -->
    <view v-if="step === 1" class="step-content">
      <!-- Tab: 熟客 / 新客 -->
      <view v-if="!isEditMode" class="tab-bar">
        <view :class="['tab', customerMode === 'regular' ? 'tab-active' : '']" @click="customerMode = 'regular'">
          <text class="tab-icon">💛</text>
          <text>熟客</text>
        </view>
        <view :class="['tab', customerMode === 'new' ? 'tab-active' : '']" @click="customerMode = 'new'">
          <text class="tab-icon">🌟</text>
          <text>新客</text>
        </view>
      </view>

      <!-- 熟客 Tab -->
      <view v-if="customerMode === 'regular'">
        <view class="card">
          <view class="section-title">
            <text class="section-icon">👤</text>
            <text>选择客户</text>
          </view>
          <view class="search-wrap">
            <view class="search-bar">
              <text class="search-icon">🔍</text>
              <input
                v-model="customerKeyword"
                placeholder="输入客户名 / 手机号 / 猫咪名搜索"
                class="search-input"
                @input="onCustomerKeywordInput"
                @confirm="searchCustomers"
              />
            </view>
            <view v-if="showCustomerSuggestions" class="search-suggestions">
              <view
                v-for="item in customerSuggestions"
                :key="item.key"
                class="suggestion-item"
                @click="selectCustomerSuggestion(item)"
              >
                <view class="suggestion-main">
                  <text class="suggestion-title">{{ item.title }}</text>
                  <text v-if="item.subtitle" class="suggestion-subtitle">{{ item.subtitle }}</text>
                  <view v-if="item.petMeta" class="suggestion-pet-meta">
                    <text class="suggestion-pet-line">{{ item.petMeta }}</text>
                    <view v-if="item.petTags.length > 0" class="suggestion-tag-row">
                      <text
                        v-for="tag in item.petTags"
                        :key="`${item.key}-${tag.text}`"
                        :class="['suggestion-tag', tag.className]"
                        :style="tag.style"
                      >{{ tag.text }}</text>
                    </view>
                  </view>
                </view>
                <text class="suggestion-arrow">›</text>
              </view>
            </view>
            <view v-else-if="showCustomerSearchEmpty" class="search-empty">
              没有匹配的客户或猫咪
            </view>
          </view>
          <view v-if="!customerKeyword.trim()" class="option-list">
            <view
              v-for="c in customerList" :key="c.ID"
              :class="['option', form.customer_id === c.ID ? 'selected' : '']"
              @click="selectCustomer(c)"
            >
              <text class="option-icon">{{ form.customer_id === c.ID ? '💜' : '🤍' }}</text>
              <text>{{ c.nickname || c.phone }}</text>
            </view>
          </view>
        </view>

        <view v-if="form.customer_id" class="card">
          <view class="section-title">
            <text class="section-icon">🐾</text>
            <text>选择宠物（可多选）</text>
          </view>
          <view class="option-list">
            <view
              v-for="p in petList" :key="p.ID"
              :class="['option', isPetSelected(p.ID) ? 'selected' : '']"
              @click="togglePetSelection(p.ID)"
            >
              <text class="option-icon">{{ isPetSelected(p.ID) ? '🐱' : '🐾' }}</text>
              <text>{{ p.name }} ({{ p.species }} {{ p.breed }})</text>
            </view>
          </view>
        </view>

        <button class="btn-primary" @click="nextStep" :disabled="!form.customer_id || form.pets.length === 0">
          下一步 →
        </button>
      </view>

      <!-- 新客 Tab -->
      <view v-if="customerMode === 'new'">
        <view class="card">
          <view class="section-title">
            <text class="section-icon">📝</text>
            <text>客户信息（选填）</text>
          </view>
          <view class="form-row">
            <text class="form-label">👤 姓名</text>
            <input v-model="newCustomer.nickname" placeholder="客户姓名" class="form-input-direct" />
          </view>
          <view class="form-row">
            <text class="form-label">📱 手机号</text>
            <input v-model="newCustomer.phone" placeholder="手机号码" class="form-input-direct" type="number" />
          </view>
        </view>

        <view class="card">
          <view class="section-title">
            <text class="section-icon">📋</text>
            <text>粘贴宝贝信息</text>
          </view>
          <view class="textarea-wrapper">
            <textarea v-model="templateText" placeholder="将宝贝洗护小调查的内容粘贴到这里，系统会自动解析..." class="template-textarea" />
          </view>
          <button class="btn-secondary" @click="parseTemplate">
            ✨ 智能解析
          </button>
        </view>

        <!-- 解析结果预览 -->
        <view v-if="showParsed" class="card card-highlight">
          <view class="section-title">
            <text class="section-icon">🎉</text>
            <text>解析结果（点击可编辑）</text>
          </view>
          <view class="parsed-grid">
            <view class="parsed-item" v-if="parsed.name" @click="editField('name', '大名')">
              <text class="parsed-label">🐱 大名</text>
              <text class="parsed-value">{{ parsed.name }}</text>
            </view>
            <view class="parsed-item" v-if="parsed.breed" @click="editField('breed', '品种')">
              <text class="parsed-label">🏷️ 品种</text>
              <text class="parsed-value">{{ parsed.breed }}</text>
            </view>
            <view class="parsed-item" @click="editField('gender', '性别')">
              <text class="parsed-label">⚧ 性别</text>
              <text class="parsed-value">{{ ['未知','公','母'][parsed.gender] }}</text>
            </view>
            <view class="parsed-item">
              <text class="parsed-label">💉 绝育</text>
              <text class="parsed-value">{{ parsed.neutered ? '是' : '否' }}</text>
            </view>
            <view class="parsed-item" v-if="parsed.age" @click="editField('age', '年龄')">
              <text class="parsed-label">🎂 年龄</text>
              <text class="parsed-value">{{ parsed.age }}</text>
            </view>
            <view class="parsed-item" v-if="parsed.furMatted" @click="editField('furMatted', '毛发打结')">
              <text class="parsed-label">🧶 打结</text>
              <text class="parsed-value">{{ parsed.furMatted }}</text>
            </view>
            <view class="parsed-item" v-if="parsed.lastBathTime" @click="editField('lastBathTime', '上次洗澡')">
              <text class="parsed-label">🛁 上次洗澡</text>
              <text class="parsed-value">{{ parsed.lastBathTime }}</text>
            </view>
            <view class="parsed-item" v-if="parsed.vaccination" @click="editField('vaccination', '疫苗')">
              <text class="parsed-label">💉 疫苗</text>
              <text class="parsed-value">{{ parsed.vaccination }}</text>
            </view>
            <view class="parsed-item" v-if="parsed.healthHistory" @click="editField('healthHistory', '疾病史')">
              <text class="parsed-label">🏥 疾病史</text>
              <text class="parsed-value">{{ parsed.healthHistory }}</text>
            </view>
            <view class="parsed-item" v-if="parsed.personality" @click="editField('personality', '性格')">
              <text class="parsed-label">😸 性格</text>
              <text class="parsed-value">{{ parsed.personality }}</text>
            </view>
            <view class="parsed-item full-width" v-if="parsed.reactions" @click="editField('reactions', '特殊反应')">
              <text class="parsed-label">⚡ 特殊反应</text>
              <text class="parsed-value">{{ parsed.reactions }}</text>
            </view>
            <view class="parsed-item full-width" v-if="parsed.source">
              <text class="parsed-label">📍 来源</text>
              <text class="parsed-value">{{ parsed.source }}</text>
            </view>
          </view>
        </view>

        <button class="btn-primary" @click="submitNewCustomer" :disabled="!canSubmitNew" :loading="newSubmitting">
          下一步 →
        </button>
      </view>
    </view>

    <!-- Step 2: Services (3-level category selector) -->
    <view v-if="step === 2" class="step-content">
      <view class="card" v-for="item in selectedPetConfigs" :key="item.pet.ID">
        <view class="section-title">
          <text class="section-icon">✂️</text>
          <text>{{ item.pet.name }} · 选择服务</text>
        </view>

        <view class="svc-picker">
          <!-- Level 1: Left sidebar -->
          <view class="svc-picker-sidebar">
            <view
              v-for="cat in categoryTree" :key="cat.ID"
              :class="['sidebar-item', activeCategoryId === cat.ID ? 'active' : '']"
              @click="selectCategory(cat.ID)"
            >
              <text>{{ cat.name }}</text>
            </view>
          </view>

          <!-- Right content area -->
          <view class="svc-picker-main">
            <!-- Level 2: Sub-category tabs -->
            <scroll-view scroll-x class="sub-tab-bar">
              <view class="sub-tab-list">
                <view
                  :class="['sub-tab', activeSubCategoryId === 0 ? 'active' : '']"
                  @click="selectSubCategory(0)"
                >全部</view>
                <view
                  v-for="sub in subCategories" :key="sub.ID"
                  :class="['sub-tab', activeSubCategoryId === sub.ID ? 'active' : '']"
                  @click="selectSubCategory(sub.ID)"
                >{{ sub.name }}</view>
              </view>
            </scroll-view>

            <!-- Level 3: Service items -->
            <scroll-view scroll-y class="svc-item-list">
              <view v-if="filteredServices.length === 0" class="svc-empty">暂无服务</view>
              <view
                v-for="s in filteredServices" :key="s.ID"
                :class="['svc-item', item.selection.service_ids.includes(s.ID) ? 'checked' : '']"
                @click="toggleService(item.pet.ID, s.ID)"
              >
                <view class="svc-item-info">
                  <text class="svc-item-name">{{ s.name }}</text>
                  <text class="svc-item-cat">{{ getSubCategoryName(s.category_id) }} · {{ s.duration }}分钟</text>
                </view>
                <view class="svc-item-right">
                  <text class="svc-item-price">¥{{ s.base_price }}</text>
                  <view :class="['svc-item-check', item.selection.service_ids.includes(s.ID) ? 'on' : '']"></view>
                </view>
              </view>
            </scroll-view>
          </view>
        </view>
      </view>

      <view class="summary-card" v-if="form.pets.length > 0">
        <view class="summary-row">
          <text class="summary-label">💰 合计费用</text>
          <text class="summary-amount">¥{{ totalAmount }}</text>
        </view>
        <view class="summary-row">
          <text class="summary-label">⏱️ 预计时长</text>
          <text class="summary-duration">{{ totalDuration }}分钟</text>
        </view>
      </view>

      <!-- Selected services summary bar -->
      <view class="selected-bar" v-if="selectedServiceIds.length > 0">
        <text class="selected-bar-text">已选择 {{ selectedServiceIds.length }} 项</text>
      </view>

      <view class="btn-row">
        <button class="btn-ghost" @click="step = 1">← 上一步</button>
        <button class="btn-primary" @click="nextStep" :disabled="!canProceedServices">下一步 →</button>
      </view>
    </view>

    <!-- Step 3: Date & Time -->
    <view v-if="step === 3" class="step-content">
      <view class="card">
        <view class="section-title">
          <text class="section-icon">📅</text>
          <text>选择日期</text>
        </view>
        <picker mode="date" :value="form.date" @change="onDateChange">
          <view class="date-picker">
            <text class="date-icon">📆</text>
            <text :class="form.date ? 'date-text' : 'date-placeholder'">{{ form.date || '点击选择日期' }}</text>
            <text class="date-arrow">›</text>
          </view>
        </picker>
      </view>

      <view class="card occupancy-card">
        <view class="section-title">
          <text class="section-icon">🕒</text>
          <text>预约占位</text>
        </view>
        <view class="occupancy-row">
          <text class="occupancy-label">服务预计时长</text>
          <text class="occupancy-value">{{ totalDuration }}分钟</text>
        </view>
        <view class="occupancy-row" v-if="occupiedDuration > 0">
          <text class="occupancy-label">当前预约占位</text>
          <text class="occupancy-value">{{ occupiedDuration }}分钟</text>
        </view>
        <text class="occupancy-tip">开始和结束时间由工作人员手动选择，日历占位以该时段为准。</text>
        <text v-if="occupiedDuration > 0 && occupiedDuration !== totalDuration" class="occupancy-warning">
          当前占位与服务预计时长不一致，提交后将按手动选择的时段覆盖日历。
        </text>
      </view>

      <view v-if="slotsLoading" class="loading-card">
        <text class="loading-text">⏳ 查询可用起始时间...</text>
      </view>
      <view v-else-if="staffSlots.length === 0 && form.date" class="empty-card">
        <text class="empty-icon">😿</text>
        <text class="empty-text">该日期暂无可用起始时间</text>
      </view>

      <view v-for="ss in staffSlots" :key="ss.staff.ID" class="card">
        <view class="staff-name">
          <text class="staff-icon">💇</text>
          <text>{{ ss.staff.name }}</text>
        </view>
        <text class="time-section-label">开始时间</text>
        <view class="slots-grid">
          <view
            v-for="slot in ss.slots" :key="slot.start_time"
            :class="['slot', form.staff_id === ss.staff.ID && form.start_time === slot.start_time ? 'selected' : '']"
            @click="selectStartSlot(ss.staff.ID, slot.start_time)"
          >
            {{ slot.start_time }}
          </view>
        </view>

        <view v-if="form.staff_id === ss.staff.ID && form.start_time" class="end-time-panel">
          <text class="time-section-label">结束时间</text>
          <view class="slots-grid" v-if="getEndTimeOptions(ss.staff.ID).length > 0">
            <view
              v-for="endTime in getEndTimeOptions(ss.staff.ID)"
              :key="endTime"
              :class="['slot', form.end_time === endTime ? 'selected' : '']"
              @click="selectEndTime(endTime)"
            >
              {{ endTime }}
            </view>
          </view>
          <text v-else class="end-time-empty">当前开始时间之后没有连续可用区间，请重新选择开始时间。</text>
        </view>
      </view>

      <view class="btn-row">
        <button class="btn-ghost" @click="step = 2">← 上一步</button>
        <button class="btn-primary" @click="nextStep" :disabled="!form.date || !hasSelectedSlot">下一步 →</button>
      </view>
    </view>

    <!-- Step 4: Confirm -->
    <view v-if="step === 4" class="step-content">
      <!-- Detail card -->
      <view class="card confirm-card">
        <view class="confirm-header">
          <text class="confirm-title">预约详情</text>
        </view>
        <view class="confirm-row">
          <text class="label">👤 客户</text>
          <text class="value">{{ selectedCustomer?.nickname || selectedCustomer?.phone || newCustomer.nickname || newCustomer.phone || '散客' }}</text>
        </view>
        <view class="confirm-row">
          <text class="label">🐱 宠物</text>
          <text class="value">{{ confirmPetNames || parsed.name }}</text>
        </view>
        <view class="confirm-row">
          <text class="label">📅 日期</text>
          <text class="value">{{ form.date }}</text>
        </view>
        <view class="confirm-row">
          <text class="label">⏰ 时间</text>
          <text class="value">{{ form.start_time }} - {{ form.end_time }}</text>
        </view>
        <view class="confirm-row">
          <text class="label">💇 洗护师</text>
          <text class="value">{{ selectedStaffName }}</text>
        </view>
        <view class="confirm-row">
          <text class="label">⏱️ 服务预计</text>
          <text class="value">{{ totalDuration }}分钟</text>
        </view>
        <view class="confirm-row">
          <text class="label">🕒 占位时长</text>
          <text class="value">{{ occupiedDuration }}分钟</text>
        </view>
        <view class="confirm-row confirm-row-amount">
          <text class="label">💰 金额</text>
          <text class="amount">¥{{ totalAmount }}</text>
        </view>
      </view>

      <!-- Pet & service cards -->
      <view class="pet-card" v-for="item in confirmPetSummaries" :key="item.pet_id">
        <view class="pet-card-top">
          <view class="pet-card-avatar">🐱</view>
          <view class="pet-card-info">
            <text class="pet-card-name">{{ item.name }}</text>
            <text class="pet-card-breed" v-if="item.breed">{{ item.breed }}</text>
          </view>
        </view>

        <view class="pet-svc-row pet-svc-row-label">
          <text class="pet-svc-title">预约项目</text>
          <view class="pet-svc-add" @click="openServicePicker(item.pet_id)">添加 (已选{{ item.services.length }}项)</view>
        </view>

        <view class="pet-svc-table" v-if="item.services.length > 0">
          <view class="pet-svc-table-head">
            <text class="pet-svc-th">预约项目</text>
            <text class="pet-svc-th-op">操作</text>
          </view>
          <view class="pet-svc-table-row" v-for="(svcName, idx) in item.services" :key="idx">
            <text class="pet-svc-td">{{ svcName }}</text>
            <view class="pet-svc-td-op" @click="removeServiceFromPet(item.pet_id, idx)">
              <text class="pet-svc-del-btn">删除</text>
            </view>
          </view>
        </view>
        <view v-else class="pet-svc-empty">暂未选择服务项目</view>
      </view>

      <!-- Service picker popup -->
      <view v-if="showServicePicker" class="svc-picker-overlay" @click.self="showServicePicker = false">
        <view class="svc-picker-popup">
          <view class="svc-picker-popup-header">
            <text class="svc-picker-popup-title">选择服务</text>
            <text class="svc-picker-popup-close" @click="showServicePicker = false">✕</text>
          </view>
          <view class="svc-picker">
            <view class="svc-picker-sidebar">
              <view
                v-for="cat in categoryTree" :key="cat.ID"
                :class="['sidebar-item', activeCategoryId === cat.ID ? 'active' : '']"
                @click="selectCategory(cat.ID)"
              ><text>{{ cat.name }}</text></view>
            </view>
            <view class="svc-picker-main">
              <scroll-view scroll-x class="sub-tab-bar">
                <view class="sub-tab-list">
                  <view :class="['sub-tab', activeSubCategoryId === 0 ? 'active' : '']" @click="selectSubCategory(0)">全部</view>
                  <view v-for="sub in subCategories" :key="sub.ID" :class="['sub-tab', activeSubCategoryId === sub.ID ? 'active' : '']" @click="selectSubCategory(sub.ID)">{{ sub.name }}</view>
                </view>
              </scroll-view>
              <scroll-view scroll-y class="svc-item-list">
                <view v-if="filteredServices.length === 0" class="svc-empty">暂无服务</view>
                <view
                  v-for="s in filteredServices" :key="s.ID"
                  :class="['svc-item', pickerPetSelection?.service_ids.includes(s.ID) ? 'checked' : '']"
                  @click="togglePickerService(s.ID)"
                >
                  <view class="svc-item-info">
                    <text class="svc-item-name">{{ s.name }}</text>
                    <text class="svc-item-cat">{{ getSubCategoryName(s.category_id) }} · {{ s.duration }}分钟</text>
                  </view>
                  <view class="svc-item-right">
                    <text class="svc-item-price">¥{{ s.base_price }}</text>
                    <view :class="['svc-item-check', pickerPetSelection?.service_ids.includes(s.ID) ? 'on' : '']"></view>
                  </view>
                </view>
              </scroll-view>
            </view>
          </view>
          <view class="svc-picker-popup-footer">
            <button class="btn-primary" @click="showServicePicker = false">确定</button>
          </view>
        </view>
      </view>

      <view class="card">
        <view class="form-row" style="margin-bottom: 0;">
          <text class="form-label">📝 备注</text>
          <view class="input-wrapper">
            <textarea v-model="form.notes" placeholder="有什么需要特别注意的吗？" class="textarea" />
          </view>
        </view>
      </view>

      <view class="btn-row">
        <button class="btn-ghost" @click="step = 3">← 上一步</button>
        <button class="btn-submit" @click="onSubmit" :loading="submitting">{{ isEditMode ? '💾 保存修改' : '🎉 确认预约' }}</button>
      </view>
    </view>
  </view>
  </SideLayout>
</template>

<script setup lang="ts">
import { ref, reactive, computed, nextTick } from 'vue'
import { onLoad } from '@dcloudio/uni-app'
import SideLayout from '@/components/SideLayout.vue'
import { getCustomerList, getCustomerPets, createCustomer } from '@/api/customer'
import { createPet } from '@/api/pet'
import { getServiceList } from '@/api/service'
import { getCategoryTree } from '@/api/service-category'
import { getAvailableSlots, createAppointment, getAppointment, updateAppointment } from '@/api/appointment'
import { safeBack } from '@/utils/navigate'
import { getPersonalityBg, getPersonalityColor } from '@/utils/personality'

interface CustomerSuggestion {
  key: string
  customer: Customer
  pet?: Pet
  title: string
  subtitle: string
  petMeta?: string
  petTags: Array<{ text: string; className: string; style?: string }>
}

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
  return rem > 0 ? `${years}岁${rem}个月` : `${years}岁`
}

function getSuggestionPetMeta(pet: Pet): string {
  const parts: string[] = []
  if (pet.gender === 1) parts.push('弟弟')
  if (pet.gender === 2) parts.push('妹妹')
  const age = calcAge(pet.birth_date)
  if (age) parts.push(age)
  return parts.join(' · ')
}

function getSuggestionPetTags(pet: Pet) {
  const tags: Array<{ text: string; className: string; style?: string }> = []
  if (pet.fur_level) tags.push({ text: pet.fur_level, className: 'tag-fur' })
  if (pet.neutered) tags.push({ text: '已绝育', className: 'tag-neutered' })
  if (pet.personality) {
    tags.push({
      text: pet.personality,
      className: 'tag-personality',
      style: `background:${getPersonalityBg(pet.personality)};color:${getPersonalityColor(pet.personality)};`,
    })
  }
  if (pet.aggression && pet.aggression !== '无') {
    tags.push({ text: `⚡ ${pet.aggression}`, className: 'tag-aggression' })
  }
  return tags
}

interface AppointmentPetFormItem {
  pet_id: number
  service_ids: number[]
}

const step = ref(1)
const submitting = ref(false)
const slotsLoading = ref(false)
const customerMode = ref<'regular' | 'new'>('regular')
const editAppointmentId = ref(0)
const editingAppointment = ref<any>(null)

const form = ref({
  customer_id: 0, staff_id: 0,
  date: '', start_time: '', end_time: '', pets: [] as AppointmentPetFormItem[], notes: '',
})

const customerKeyword = ref('')
const customerList = ref<Customer[]>([])
const petList = ref<Pet[]>([])
const serviceList = ref<ServiceItem[]>([])
const categoryTree = ref<any[]>([])
const activeCategoryId = ref<number>(0)
const activeSubCategoryId = ref<number>(0)
const showServicePicker = ref(false)
const pickerPetId = ref(0)
const staffSlots = ref<any[]>([])
const searchingCustomers = ref(false)
const customerSuggestionOpen = ref(false)
let customerSearchTimer: ReturnType<typeof setTimeout> | null = null

// New customer fields
const newCustomer = ref({ nickname: '', phone: '' })
const templateText = ref('')
const newSubmitting = ref(false)

interface ParsedPetInfo {
  name: string
  age: string
  breed: string
  gender: number
  neutered: boolean
  furMatted: string
  lastBathTime: string
  vaccination: string
  healthHistory: string
  personality: string
  reactions: string
  source: string  // 来源，如"树街の猫"
}

const showParsed = ref(false)
const parseKey = ref(0)
const parsed = reactive<ParsedPetInfo>({
  name: '', age: '', breed: '', gender: 0, neutered: false,
  furMatted: '', lastBathTime: '', vaccination: '',
  healthHistory: '', personality: '', reactions: '', source: '',
})

const canSubmitNew = computed(() => showParsed.value && parsed.name.trim() !== '')
const isEditMode = computed(() => editAppointmentId.value > 0)

const selectedCustomer = computed(() => customerList.value.find(c => c.ID === form.value.customer_id))
const showCustomerSuggestions = computed(() =>
  customerSuggestionOpen.value &&
  customerMode.value === 'regular' &&
  customerKeyword.value.trim() !== '' &&
  customerSuggestions.value.length > 0
)
const showCustomerSearchEmpty = computed(() =>
  customerSuggestionOpen.value &&
  customerKeyword.value.trim() !== '' &&
  !searchingCustomers.value &&
  customerSuggestions.value.length === 0
)
const customerSuggestions = computed<CustomerSuggestion[]>(() => {
  const keyword = customerKeyword.value.trim().toLowerCase()
  if (!keyword) return []

  return customerList.value.flatMap((customer) => {
    const pets = Array.isArray(customer.pets) ? customer.pets : []
    const matchedPets = pets.filter((pet) => (pet.name || '').toLowerCase().includes(keyword))

      if (matchedPets.length > 0) {
        return matchedPets.map((pet) => ({
          key: `pet-${customer.ID}-${pet.ID}`,
          customer,
          pet,
          title: `${pet.name} · ${customer.nickname || customer.phone || `客户#${customer.ID}`}`,
          subtitle: [customer.phone, pet.breed].filter(Boolean).join(' · '),
          petMeta: getSuggestionPetMeta(pet),
          petTags: getSuggestionPetTags(pet),
        }))
      }

      return [{
        key: `customer-${customer.ID}`,
        customer,
        title: customer.nickname || customer.phone || `客户#${customer.ID}`,
        subtitle: [customer.phone, pets.slice(0, 3).map((pet) => pet.name).join(' / ')].filter(Boolean).join(' · '),
        petTags: [],
      }]
  }).slice(0, 20)
})
const selectedPetConfigs = computed(() =>
  form.value.pets
    .map(selection => {
      const pet = petList.value.find(item => item.ID === selection.pet_id)
      if (!pet) return null
      return { pet, selection }
    })
    .filter(Boolean) as { pet: Pet; selection: AppointmentPetFormItem }[]
)
const confirmPetSummaries = computed(() =>
  selectedPetConfigs.value.map(({ pet, selection }) => ({
    pet_id: pet.ID,
    name: pet.name,
    breed: pet.breed,
    services: selection.service_ids
      .map(id => serviceList.value.find(s => s.ID === id)?.name)
      .filter(Boolean) as string[],
  }))
)
const confirmPetNames = computed(() => confirmPetSummaries.value.map(item => item.name).join('、'))
const selectedStaffName = computed(() => {
  for (const ss of staffSlots.value) {
    if (ss.staff.ID === form.value.staff_id) return ss.staff.name
  }
  return '待分配'
})

const totalAmount = computed(() =>
  form.value.pets.reduce((sum, petItem) => (
    sum + petItem.service_ids.reduce((petSum, id) => {
      const s = serviceList.value.find(sv => sv.ID === id)
      return petSum + (s?.base_price || 0)
    }, 0)
  ), 0)
)
const totalDuration = computed(() =>
  form.value.pets.reduce((sum, petItem) => (
    sum + petItem.service_ids.reduce((petSum, id) => {
      const s = serviceList.value.find(sv => sv.ID === id)
      return petSum + (s?.duration || 0)
    }, 0)
  ), 0)
)
const selectedServiceIds = computed(() => {
  const ids = new Set<number>()
  form.value.pets.forEach((petItem) => {
    petItem.service_ids.forEach((id) => ids.add(id))
  })
  return Array.from(ids)
})
const subCategories = computed(() => {
  if (activeCategoryId.value === 0) return []
  const cat = categoryTree.value.find(c => c.ID === activeCategoryId.value)
  return cat?.children || []
})
const filteredServices = computed(() => {
  let list = serviceList.value
  if (activeCategoryId.value > 0) {
    const subIds = subCategories.value.map((c: any) => c.ID)
    if (activeSubCategoryId.value > 0) {
      list = list.filter(s => s.category_id === activeSubCategoryId.value)
    } else {
      list = list.filter(s => s.category_id && subIds.includes(s.category_id))
    }
  }
  return list
})
const canProceedServices = computed(() =>
  form.value.pets.length > 0 && form.value.pets.every(petItem => petItem.service_ids.length > 0)
)
const hasSelectedSlot = computed(() =>
  !!form.value.staff_id && !!form.value.start_time && !!form.value.end_time
)
const occupiedDuration = computed(() => {
  if (!form.value.start_time || !form.value.end_time) return 0
  return Math.max(parseTime(form.value.end_time) - parseTime(form.value.start_time), 0)
})

onLoad(async (query) => {
  if (query?.id) {
    editAppointmentId.value = parseInt(query.id)
    customerMode.value = 'regular'
  }
  if (query?.date) form.value.date = query.date
  if (query?.staff_id) form.value.staff_id = parseInt(query.staff_id)
  if (query?.time) form.value.start_time = query.time

  const [cRes, sRes, catRes] = await Promise.all([
    getCustomerList({ page: 1, page_size: 100 }),
    getServiceList({ page: 1, page_size: 100 }),
    getCategoryTree(),
  ])
  customerList.value = cRes.data.list || []
  serviceList.value = (sRes.data.list || []).filter((s: ServiceItem) => s.status === 1)
  categoryTree.value = (catRes.data || []).filter((c: any) => c.status === 1)
  if (categoryTree.value.length > 0) {
    activeCategoryId.value = categoryTree.value[0].ID
  }

  if (editAppointmentId.value) {
    await loadAppointmentForEdit(editAppointmentId.value)
  }
})

function upsertCustomerOption(customer?: Customer | null) {
  if (!customer?.ID) return
  const idx = customerList.value.findIndex(item => item.ID === customer.ID)
  if (idx >= 0) {
    customerList.value[idx] = { ...customerList.value[idx], ...customer }
  } else {
    customerList.value.unshift(customer)
  }
}

function normalizeAppointmentPets(appointment: any): AppointmentPetFormItem[] {
  if (Array.isArray(appointment?.pets) && appointment.pets.length > 0) {
    return appointment.pets.map((petItem: any) => ({
      pet_id: petItem.pet_id || petItem.pet?.ID,
      service_ids: (petItem.services || [])
        .map((serviceItem: any) => serviceItem.service_id || serviceItem.ServiceID)
        .filter(Boolean),
    })).filter((petItem: AppointmentPetFormItem) => petItem.pet_id > 0)
  }

  if (appointment?.pet?.ID) {
    return [{
      pet_id: appointment.pet.ID,
      service_ids: (appointment.services || [])
        .map((serviceItem: any) => serviceItem.service_id || serviceItem.ServiceID)
        .filter(Boolean),
    }]
  }

  return []
}

async function loadAppointmentForEdit(id: number) {
  const res = await getAppointment(id)
  const appointment = res.data
  editingAppointment.value = appointment
  upsertCustomerOption(appointment.customer as Customer)

  const customerId = appointment.customer_id || appointment.customer?.ID || 0
  if (customerId) {
    const customer = (appointment.customer || customerList.value.find(item => item.ID === customerId) || { ID: customerId }) as Customer
    await selectCustomer(customer)
  }

  form.value = {
    customer_id: customerId,
    staff_id: appointment.staff_id || appointment.staff?.ID || 0,
    date: appointment.date || '',
    start_time: appointment.start_time || '',
    end_time: appointment.end_time || '',
    pets: normalizeAppointmentPets(appointment),
    notes: appointment.notes || '',
  }
  customerKeyword.value = appointment.customer?.nickname || appointment.customer?.phone || ''

  if (form.value.date && selectedServiceIds.value.length > 0) {
    await loadSlots()
  }
}

async function searchCustomers() {
  const keyword = customerKeyword.value.trim()
  searchingCustomers.value = true
  try {
    const res = await getCustomerList({ page: 1, page_size: keyword ? 50 : 100, keyword: keyword || undefined })
    customerList.value = res.data.list || []
  } finally {
    searchingCustomers.value = false
  }
}

function onCustomerKeywordInput() {
  customerSuggestionOpen.value = true
  if (customerSearchTimer) clearTimeout(customerSearchTimer)
  customerSearchTimer = setTimeout(() => {
    searchCustomers()
  }, 250)
}

async function selectCustomer(c: Customer, preferredPetId = 0) {
  form.value.customer_id = c.ID
  form.value.pets = []
  const res = await getCustomerPets(c.ID)
  petList.value = res.data || []
  if (preferredPetId) {
    ensurePetSelected(preferredPetId)
  }
}

async function selectCustomerSuggestion(item: CustomerSuggestion) {
  customerSuggestionOpen.value = false
  customerKeyword.value = item.pet?.name || item.customer.nickname || item.customer.phone || ''
  await selectCustomer(item.customer, item.pet?.ID || 0)
}

function isPetSelected(petId: number) {
  return form.value.pets.some(item => item.pet_id === petId)
}

function getPetSelection(petId: number) {
  return form.value.pets.find(item => item.pet_id === petId)
}

const pickerPetSelection = computed(() =>
  form.value.pets.find(p => p.pet_id === pickerPetId.value)
)

function openServicePicker(petId: number) {
  pickerPetId.value = petId
  showServicePicker.value = true
}
function togglePickerService(serviceId: number) {
  const selection = form.value.pets.find(p => p.pet_id === pickerPetId.value)
  if (!selection) return
  const idx = selection.service_ids.indexOf(serviceId)
  if (idx >= 0) selection.service_ids.splice(idx, 1)
  else selection.service_ids.push(serviceId)
}
function removeServiceFromPet(petId: number, serviceIndex: number) {
  const selection = form.value.pets.find(p => p.pet_id === petId)
  if (!selection) return
  // Get the actual service ID from confirmPetSummaries
  const pet = confirmPetSummaries.value.find(p => p.pet_id === petId)
  if (!pet) return
  const svcName = pet.services[serviceIndex]
  const svc = serviceList.value.find(s => s.name === svcName)
  if (svc) {
    const i = selection.service_ids.indexOf(svc.ID)
    if (i >= 0) selection.service_ids.splice(i, 1)
  }
}

function selectCategory(catId: number) {
  activeCategoryId.value = catId
  activeSubCategoryId.value = 0
}
function selectSubCategory(subId: number) {
  activeSubCategoryId.value = subId
}
function getSubCategoryName(categoryId?: number): string {
  if (!categoryId) return ''
  for (const cat of categoryTree.value) {
    const sub = (cat.children || []).find((c: any) => c.ID === categoryId)
    if (sub) return sub.name
  }
  return ''
}

function ensurePetSelected(petId: number) {
  if (!isPetSelected(petId)) {
    form.value.pets.push({ pet_id: petId, service_ids: [] })
  }
}

function togglePetSelection(petId: number) {
  const idx = form.value.pets.findIndex(item => item.pet_id === petId)
  if (idx >= 0) {
    form.value.pets.splice(idx, 1)
  } else {
    form.value.pets.push({ pet_id: petId, service_ids: [] })
  }
}

function toggleService(petId: number, serviceId: number) {
  const selection = getPetSelection(petId)
  if (!selection) return
  const idx = selection.service_ids.indexOf(serviceId)
  if (idx >= 0) selection.service_ids.splice(idx, 1)
  else selection.service_ids.push(serviceId)
}

async function onDateChange(e: any) {
  form.value.date = e.detail.value
  form.value.start_time = ''
  form.value.end_time = ''
  form.value.staff_id = 0
  if (totalDuration.value > 0) {
    await loadSlots()
  }
}

async function loadSlots() {
  if (selectedServiceIds.value.length === 0 || totalDuration.value <= 0) {
    staffSlots.value = []
    form.value.staff_id = 0
    form.value.start_time = ''
    form.value.end_time = ''
    return
  }
  slotsLoading.value = true
  try {
    const res = await getAvailableSlots(form.value.date, {
      service_ids: selectedServiceIds.value,
      duration: 30,
      exclude_id: editAppointmentId.value || undefined,
    })
    staffSlots.value = sortStaffSlots(res.data || [])
    mergeCurrentEditSlotIntoStaffSlots()
    if (!isCurrentSlotAvailable()) {
      form.value.staff_id = 0
      form.value.start_time = ''
      form.value.end_time = ''
    } else if (!isCurrentEndTimeValid()) {
      form.value.end_time = ''
    }
  } finally { slotsLoading.value = false }
}

function buildManualSlotRange(startTime: string, endTime: string) {
  const slots: { start_time: string; end_time: string }[] = []
  let currentMinute = parseTime(startTime)
  const endMinute = parseTime(endTime)
  while (currentMinute < endMinute) {
    const nextMinute = currentMinute + 30
    slots.push({
      start_time: minutesToTime(currentMinute),
      end_time: minutesToTime(nextMinute),
    })
    currentMinute = nextMinute
  }
  return slots
}

function mergeCurrentEditSlotIntoStaffSlots() {
  if (!isEditMode.value || !editingAppointment.value) return
  const originalStaffId = editingAppointment.value.staff_id || editingAppointment.value.staff?.ID || 0
  if (
    !originalStaffId ||
    form.value.date !== editingAppointment.value.date ||
    form.value.staff_id !== originalStaffId ||
    !form.value.start_time ||
    !form.value.end_time
  ) {
    return
  }

  const manualSlots = buildManualSlotRange(form.value.start_time, form.value.end_time)
  if (manualSlots.length === 0) return

  let staffSlot = staffSlots.value.find(item => item.staff?.ID === originalStaffId)
  if (!staffSlot) {
    staffSlot = {
      staff: editingAppointment.value.staff || { ID: originalStaffId, name: `洗护师#${originalStaffId}` },
      slots: [],
    }
    staffSlots.value = [...staffSlots.value, staffSlot]
  }

  const slotMap = new Map<string, { start_time: string; end_time: string }>()
  ;(staffSlot.slots || []).forEach((slot: { start_time: string; end_time: string }) => {
    slotMap.set(slot.start_time, slot)
  })
  manualSlots.forEach((slot) => {
    slotMap.set(slot.start_time, slot)
  })

  staffSlot.slots = Array.from(slotMap.values()).sort((a, b) => parseTime(a.start_time) - parseTime(b.start_time))
  staffSlots.value = sortStaffSlots(staffSlots.value)
}

function isCurrentSlotAvailable() {
  if (!form.value.staff_id || !form.value.start_time) {
    return false
  }
  return staffSlots.value.some((staffSlot) =>
    staffSlot.staff?.ID === form.value.staff_id &&
    Array.isArray(staffSlot.slots) &&
    staffSlot.slots.some((slot: { start_time: string }) => slot.start_time === form.value.start_time)
  )
}

function isCurrentEndTimeValid() {
  if (!form.value.staff_id || !form.value.start_time || !form.value.end_time) {
    return false
  }
  return getEndTimeOptions(form.value.staff_id).includes(form.value.end_time)
}

function getEndTimeOptions(staffId: number) {
  if (!form.value.start_time || form.value.staff_id !== staffId) {
    return []
  }
  const staffSlot = staffSlots.value.find(item => item.staff?.ID === staffId)
  const slotSet = new Set<string>((staffSlot?.slots || []).map((slot: { start_time: string }) => slot.start_time))
  const options: string[] = []
  let currentMinute = parseTime(form.value.start_time)

  while (slotSet.has(minutesToTime(currentMinute))) {
    currentMinute += 30
    options.push(minutesToTime(currentMinute))
  }

  return options
}

function selectStartSlot(staffId: number, startTime: string) {
  const staffChanged = form.value.staff_id !== staffId
  const startChanged = form.value.start_time !== startTime
  form.value.staff_id = staffId
  form.value.start_time = startTime
  if (staffChanged || startChanged || !isCurrentEndTimeValid()) {
    form.value.end_time = ''
  }
}

function selectEndTime(endTime: string) {
  form.value.end_time = endTime
}

function parseTime(time: string) {
  const [hour, minute] = time.split(':').map(Number)
  return hour * 60 + minute
}

function minutesToTime(totalMinutes: number) {
  const hour = Math.floor(totalMinutes / 60)
  const minute = totalMinutes % 60
  return `${String(hour).padStart(2, '0')}:${String(minute).padStart(2, '0')}`
}

function sortStaffSlots(list: any[]) {
  return [...list].sort((a, b) => {
    const roleDiff = Number(a.staff?.role === 'admin') - Number(b.staff?.role === 'admin')
    if (roleDiff !== 0) return roleDiff
    return (a.staff?.ID || 0) - (b.staff?.ID || 0)
  })
}

async function nextStep() {
  if (step.value === 1) {
    if (!form.value.customer_id || form.value.pets.length === 0) {
      uni.showToast({ title: '请先选择客户和至少一只宠物', icon: 'none' })
      return
    }
  }
  if (step.value === 2) {
    if (!canProceedServices.value) {
      uni.showToast({ title: '每只宠物都需要选择至少一个服务', icon: 'none' })
      return
    }
    if (form.value.date) {
      await loadSlots()
    }
  }
  if (step.value === 3) {
    if (!hasSelectedSlot.value) {
      uni.showToast({ title: '请先选择完整预约时段', icon: 'none' })
      return
    }
    if (!isCurrentSlotAvailable() || !isCurrentEndTimeValid()) {
      await loadSlots()
      uni.showToast({ title: '可用时段已变更，请重新选择', icon: 'none' })
      return
    }
  }
  step.value++
}

// --- Template parsing ---

function parsePetTemplate(text: string): ParsedPetInfo {
  const result: ParsedPetInfo = {
    name: '', age: '', breed: '', gender: 0, neutered: false,
    furMatted: '', lastBathTime: '', vaccination: '',
    healthHistory: '', personality: '', reactions: '', source: '',
  }

  const lines = text.split(/\n/).map(l => l.trim()).filter(Boolean)

  // Extract source from first line (e.g. "树街の猫|宝贝洗护小调查" → "树街の猫")
  if (lines.length > 0 && lines[0].includes('|')) {
    result.source = lines[0].split('|')[0].trim()
  }
  const sectionHeaders = ['基本信息', '健康情况', '性格小秘密']

  for (const line of lines) {
    if (sectionHeaders.includes(line)) continue

    // Split by ：or :
    const match = line.match(/^(.+?)[：:](.+)$/)
    if (!match) continue

    const key = match[1].trim()
    const val = match[2].trim()

    if (key.includes('大名') || key === '名字' || key === '姓名') {
      result.name = val
    } else if (key.includes('年龄')) {
      result.age = val
    } else if (key.includes('品种')) {
      result.breed = val
    } else if (key.includes('性别')) {
      if (val.includes('公')) result.gender = 1
      else if (val.includes('母')) result.gender = 2
      else result.gender = 0
    } else if (key.includes('绝育')) {
      result.neutered = val.includes('是') || val === '已绝育'
    } else if (key.includes('打结') || key.includes('毛发')) {
      result.furMatted = val
    } else if (key.includes('洗澡')) {
      result.lastBathTime = val
    } else if (key.includes('疫苗')) {
      result.vaccination = val
    } else if (key.includes('疾病') || key.includes('健康')) {
      result.healthHistory = val
    } else if (key.includes('i猫') || key.includes('e猫') || key.includes('性格')) {
      result.personality = val
    } else if (key.includes('反应') || key.includes('对水')) {
      result.reactions = val
    }
  }

  return result
}

function parseTemplate() {
  if (!templateText.value.trim()) {
    uni.showToast({ title: '请先粘贴模版内容', icon: 'none' })
    return
  }
  const result = parsePetTemplate(templateText.value)
  // Hide, assign new values, bump key to force DOM recreation
  showParsed.value = false
  Object.assign(parsed, result)
  parseKey.value++
  nextTick(() => {
    showParsed.value = true
  })
  if (!result.name) {
    uni.showToast({ title: '未识别到猫咪名字，请手动填写', icon: 'none' })
  }
}

function ageToBirthDate(age: string): string | undefined {
  if (!age) return undefined
  const now = new Date()
  const monthMatch = age.match(/(\d+)\s*个?月/)
  const yearMatch = age.match(/(\d+)\s*岁/)
  if (monthMatch) {
    now.setMonth(now.getMonth() - parseInt(monthMatch[1]))
  } else if (yearMatch) {
    now.setFullYear(now.getFullYear() - parseInt(yearMatch[1]))
  } else {
    return undefined
  }
  return now.toISOString().split('T')[0]
}

function editField(field: string, label: string) {
  const current = (parsed as any)[field] || ''
  uni.showModal({
    title: `修改${label}`,
    editable: true,
    placeholderText: `请输入${label}`,
    content: String(current),
    success: (res) => {
      if (res.confirm && res.content !== undefined) {
        (parsed as any)[field] = res.content
      }
    }
  })
}

function buildCareNotes(p: ParsedPetInfo): string {
  const parts: string[] = []
  if (p.furMatted) parts.push(`毛发打结：${p.furMatted}`)
  if (p.lastBathTime) parts.push(`上次洗澡：${p.lastBathTime}`)
  if (p.vaccination) parts.push(`疫苗：${p.vaccination}`)
  if (p.healthHistory) parts.push(`疾病史：${p.healthHistory}`)
  return parts.join('\n')
}

async function submitNewCustomer() {
  const phone = newCustomer.value.phone.trim()
  const nickname = newCustomer.value.nickname.trim()
  const hasPetInfo = showParsed.value && parsed.name.trim() !== ''

  if (!phone && !nickname && !hasPetInfo) {
    uni.showToast({ title: '请至少填写手机号或姓名', icon: 'none' })
    return
  }

  newSubmitting.value = true
  try {
    let customerId = 0

    // Step 1: Find or create customer
    if (phone) {
      const searchRes = await getCustomerList({ page: 1, page_size: 1, keyword: phone })
      const existing = (searchRes.data.list || []).find((c: Customer) => c.phone === phone)
      if (existing) {
        customerId = existing.ID
        uni.showToast({ title: '该客户已存在，已自动关联', icon: 'none' })
      } else {
        const cRes = await createCustomer({ phone, nickname: nickname || phone })
        customerId = cRes.data.ID
      }
    } else {
      const cRes = await createCustomer({ nickname: nickname || `散客-${parsed.name || '未命名'}` })
      customerId = cRes.data.ID
    }

    form.value.customer_id = customerId

    // Step 2: Create pet (only if we have pet info from parsing)
    if (hasPetInfo) {
      const petData: Partial<Pet> = {
        customer_id: customerId,
        name: parsed.name,
        species: '猫',
        breed: parsed.breed || '',
        gender: parsed.gender,
        neutered: parsed.neutered,
        personality: parsed.personality || '',
        behavior_notes: parsed.reactions || '',
        care_notes: buildCareNotes(parsed),
      }

      const birthDate = ageToBirthDate(parsed.age)
      if (birthDate) {
        petData.birth_date = birthDate
      }

      const petRes = await createPet(petData)
      form.value.pets = [{ pet_id: petRes.data.ID, service_ids: [] }]
      petList.value = [petRes.data]
    }

    // Auto-fill appointment notes with key info for the groomer
    const notesParts: string[] = []
    if (parsed.source) notesParts.push(`来源：${parsed.source}`)
    if (parsed.furMatted && parsed.furMatted !== '否' && parsed.furMatted !== '无') notesParts.push(`毛发打结：${parsed.furMatted}`)
    if (parsed.lastBathTime) notesParts.push(`上次洗澡：${parsed.lastBathTime}`)
    if (parsed.reactions) notesParts.push(`注意：${parsed.reactions}`)
    if (parsed.personality) notesParts.push(`性格：${parsed.personality}`)
    if (notesParts.length) form.value.notes = notesParts.join('\n')

    step.value = 2
  } catch (e: any) {
    uni.showToast({ title: e.message || '创建失败', icon: 'none' })
  } finally {
    newSubmitting.value = false
  }
}

async function onSubmit() {
  submitting.value = true
  try {
    // 提交前校验：重新获取服务列表，确保选中的服务仍然有效（防止服务被删除/下架后仍提交旧ID）
    const sRes = await getServiceList({ page: 1, page_size: 100 })
    const freshServices = (sRes.data.list || []).filter((s: ServiceItem) => s.status === 1)
    const freshIds = new Set(freshServices.map(s => s.ID))
    let hasInvalid = false
    for (const petItem of form.value.pets) {
      const invalid = petItem.service_ids.filter(id => !freshIds.has(id))
      if (invalid.length > 0) {
        hasInvalid = true
        petItem.service_ids = petItem.service_ids.filter(id => freshIds.has(id))
      }
    }
    serviceList.value = freshServices
    if (hasInvalid) {
      uni.showToast({ title: '部分服务已下架，已自动移除，请确认后重新提交', icon: 'none', duration: 3000 })
      submitting.value = false
      return
    }

    const payload = {
      customer_id: form.value.customer_id,
      pet_id: form.value.pets[0]?.pet_id,
      pets: form.value.pets.map(item => ({
        pet_id: item.pet_id,
        service_ids: item.service_ids,
      })),
      staff_id: form.value.staff_id || undefined,
      date: form.value.date,
      start_time: form.value.start_time,
      end_time: form.value.end_time,
      source: 2,
      notes: form.value.notes,
    }

    if (isEditMode.value) {
      await updateAppointment(editAppointmentId.value, payload)
      uni.showToast({ title: '修改成功', icon: 'success' })
      setTimeout(() => safeBack(), 500)
    } else {
      await createAppointment(payload)
      uni.showToast({ title: '预约成功', icon: 'success' })
      setTimeout(() => safeBack(), 500)
    }
  } finally { submitting.value = false }
}
</script>

<style scoped>
/* ========== Page Base ========== */
.page {
  padding: 24rpx 28rpx 48rpx;
  background: linear-gradient(180deg, #EEF2FF 0%, #F9FAFB 40%, #F3F4F6 100%);
  min-height: 100vh;
}

/* ========== Header ========== */
.page-header {
  text-align: center;
  padding: 20rpx 0 8rpx;
  margin-bottom: 16rpx;
}
.page-title {
  font-size: 40rpx;
  font-weight: 700;
  color: #1E1B4B;
  display: block;
  letter-spacing: 2rpx;
}
.page-subtitle {
  font-size: 24rpx;
  color: #6B7280;
  display: block;
  margin-top: 8rpx;
}

/* ========== Step Indicator ========== */
.steps-wrapper {
  background: #fff;
  border-radius: 20rpx;
  padding: 28rpx 20rpx;
  margin-bottom: 28rpx;
  box-shadow: 0 4rpx 20rpx rgba(79, 70, 229, 0.08);
}
.steps {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  position: relative;
}
.step-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  flex: 1;
  position: relative;
}
.step-circle {
  width: 64rpx;
  height: 64rpx;
  border-radius: 50%;
  background: #F3F4F6;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 10rpx;
  transition: all 0.3s;
  border: 3rpx solid #E5E7EB;
}
.step-item.active .step-circle {
  background: #EEF2FF;
  border-color: #C7D2FE;
}
.step-item.current .step-circle {
  background: #4F46E5;
  border-color: #4F46E5;
  box-shadow: 0 4rpx 16rpx rgba(79, 70, 229, 0.35);
}
.step-icon {
  font-size: 28rpx;
}
.step-item.current .step-icon {
  font-size: 26rpx;
}
.step-check {
  font-size: 28rpx;
  color: #4F46E5;
  font-weight: 700;
}
.step-label {
  font-size: 22rpx;
  color: #9CA3AF;
  font-weight: 500;
}
.step-item.active .step-label {
  color: #6366F1;
}
.step-item.current .step-label {
  color: #4F46E5;
  font-weight: 700;
}
.step-line {
  position: absolute;
  top: 32rpx;
  left: calc(50% + 36rpx);
  width: calc(100% - 72rpx);
  height: 4rpx;
  background: #E5E7EB;
  border-radius: 2rpx;
}
.step-line.line-active {
  background: linear-gradient(90deg, #4F46E5, #818CF8);
}

/* ========== Content Area ========== */
.step-content {
  min-height: 400rpx;
}

/* ========== Card ========== */
.card {
  background: #fff;
  border-radius: 20rpx;
  padding: 28rpx;
  margin-bottom: 24rpx;
  box-shadow: 0 2rpx 16rpx rgba(0, 0, 0, 0.04);
  border: 2rpx solid rgba(229, 231, 235, 0.6);
}
.card-highlight {
  border-color: #C7D2FE;
  background: linear-gradient(135deg, #fff 0%, #F5F3FF 100%);
  box-shadow: 0 4rpx 24rpx rgba(79, 70, 229, 0.1);
}

/* ========== Tab Bar ========== */
.tab-bar {
  display: flex;
  margin-bottom: 24rpx;
  background: #fff;
  border-radius: 20rpx;
  padding: 8rpx;
  box-shadow: 0 2rpx 16rpx rgba(0, 0, 0, 0.04);
  border: 2rpx solid rgba(229, 231, 235, 0.6);
}
.tab {
  flex: 1;
  text-align: center;
  padding: 20rpx 0;
  font-size: 30rpx;
  color: #9CA3AF;
  border-radius: 16rpx;
  font-weight: 500;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8rpx;
}
.tab-active {
  background: linear-gradient(135deg, #4F46E5, #6366F1);
  color: #fff;
  font-weight: 700;
  box-shadow: 0 4rpx 16rpx rgba(79, 70, 229, 0.3);
}
.tab-icon {
  font-size: 28rpx;
}

/* ========== Section Title ========== */
.section-title {
  font-size: 30rpx;
  font-weight: 700;
  color: #1E1B4B;
  margin-bottom: 20rpx;
  display: flex;
  align-items: center;
  gap: 10rpx;
}
.section-icon {
  font-size: 28rpx;
}

/* ========== Search Bar ========== */
.search-bar {
  display: flex;
  align-items: center;
  background: #F9FAFB;
  border: 2rpx solid #E5E7EB;
  border-radius: 16rpx;
  padding: 4rpx 20rpx;
  transition: all 0.2s;
}
.search-wrap {
  position: relative;
  margin-bottom: 20rpx;
}
.search-icon {
  font-size: 28rpx;
  margin-right: 12rpx;
}
.search-input {
  flex: 1;
  font-size: 28rpx;
  padding: 18rpx 0;
  background: transparent;
  color: #374151;
}
.search-suggestions {
  position: absolute;
  left: 0;
  right: 0;
  top: calc(100% - 12rpx);
  background: #fff;
  border: 2rpx solid #E5E7EB;
  border-radius: 18rpx;
  box-shadow: 0 16rpx 40rpx rgba(15, 23, 42, 0.1);
  overflow: hidden;
  z-index: 20;
}
.suggestion-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16rpx;
  padding: 20rpx 24rpx;
  border-bottom: 1rpx solid #EEF2FF;
  background: #fff;
}
.suggestion-item:last-child {
  border-bottom: none;
}
.suggestion-main {
  display: flex;
  flex-direction: column;
  gap: 6rpx;
  min-width: 0;
}
.suggestion-title {
  font-size: 27rpx;
  color: #1F2937;
  font-weight: 600;
}
.suggestion-subtitle {
  font-size: 22rpx;
  color: #6B7280;
}
.suggestion-pet-meta {
  display: flex;
  flex-direction: column;
  gap: 8rpx;
  margin-top: 4rpx;
}
.suggestion-pet-line {
  font-size: 22rpx;
  color: #4B5563;
}
.suggestion-tag-row {
  display: flex;
  flex-wrap: wrap;
  gap: 8rpx;
}
.suggestion-tag {
  display: inline-flex;
  align-items: center;
  padding: 4rpx 12rpx;
  border-radius: 999rpx;
  font-size: 18rpx;
  line-height: 1.2;
  background: #F3F4F6;
  color: #4B5563;
}
.suggestion-tag.tag-fur {
  background: #EEF2FF;
  color: #4F46E5;
}
.suggestion-tag.tag-neutered {
  background: #ECFDF5;
  color: #047857;
}
.suggestion-tag.tag-aggression {
  background: #FEF2F2;
  color: #DC2626;
}
.suggestion-arrow {
  font-size: 32rpx;
  color: #C7D2FE;
  flex-shrink: 0;
}
.search-empty {
  margin: -8rpx 0 20rpx;
  font-size: 24rpx;
  color: #9CA3AF;
  padding-left: 8rpx;
}

/* ========== Option List ========== */
.option-list {
  display: flex;
  flex-direction: column;
  gap: 16rpx;
}
.option {
  background: #F9FAFB;
  border: 2rpx solid #E5E7EB;
  border-radius: 16rpx;
  padding: 24rpx;
  font-size: 28rpx;
  color: #374151;
  display: flex;
  align-items: center;
  gap: 12rpx;
  transition: all 0.2s;
}
.option:active {
  transform: scale(0.98);
}
.option.selected {
  border-color: #4F46E5;
  background: linear-gradient(135deg, #EEF2FF, #F5F3FF);
  color: #4338CA;
  box-shadow: 0 2rpx 12rpx rgba(79, 70, 229, 0.12);
}
.option-icon {
  font-size: 28rpx;
}

/* ========== Service Options ========== */
.service-option {
  padding: 20rpx 24rpx;
}
.svc-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  width: 100%;
}
.svc-left {
  display: flex;
  align-items: center;
  gap: 12rpx;
}
.svc-check {
  font-size: 28rpx;
}
.svc-name {
  font-size: 28rpx;
  font-weight: 600;
  color: #1F2937;
}
.service-option.selected .svc-name {
  color: #4338CA;
}
.svc-right {
  display: flex;
  flex-direction: column;
  align-items: flex-end;
}
.svc-price {
  font-size: 28rpx;
  font-weight: 700;
  color: #4F46E5;
}
.svc-duration {
  font-size: 22rpx;
  color: #9CA3AF;
  margin-top: 4rpx;
}

/* ========== Summary Card ========== */
.summary-card {
  background: linear-gradient(135deg, #4F46E5, #6366F1);
  border-radius: 20rpx;
  padding: 28rpx;
  margin-bottom: 8rpx;
  box-shadow: 0 6rpx 24rpx rgba(79, 70, 229, 0.3);
}
.summary-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8rpx 0;
}
.summary-label {
  font-size: 26rpx;
  color: rgba(255, 255, 255, 0.85);
}
.summary-amount {
  font-size: 36rpx;
  font-weight: 800;
  color: #fff;
}
.summary-duration {
  font-size: 28rpx;
  font-weight: 600;
  color: rgba(255, 255, 255, 0.9);
}
.occupancy-card {
  border-color: #C7D2FE;
}
.occupancy-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8rpx 0;
}
.occupancy-label {
  font-size: 26rpx;
  color: #6B7280;
}
.occupancy-value {
  font-size: 28rpx;
  font-weight: 700;
  color: #1F2937;
}
.occupancy-tip {
  display: block;
  margin-top: 10rpx;
  font-size: 24rpx;
  color: #6366F1;
  line-height: 1.6;
}
.occupancy-warning {
  display: block;
  margin-top: 8rpx;
  font-size: 24rpx;
  color: #D97706;
  line-height: 1.6;
}

/* ========== Date Picker ========== */
.date-picker {
  display: flex;
  align-items: center;
  background: #F9FAFB;
  border: 2rpx solid #E5E7EB;
  border-radius: 16rpx;
  padding: 22rpx 24rpx;
  font-size: 28rpx;
  transition: all 0.2s;
}
.date-icon {
  font-size: 30rpx;
  margin-right: 14rpx;
}
.date-text {
  flex: 1;
  color: #1F2937;
  font-weight: 600;
}
.date-placeholder {
  flex: 1;
  color: #9CA3AF;
}
.date-arrow {
  font-size: 36rpx;
  color: #C7D2FE;
  font-weight: 300;
}

/* ========== Loading & Empty ========== */
.loading-card {
  background: #fff;
  border-radius: 20rpx;
  padding: 48rpx;
  text-align: center;
  margin-bottom: 24rpx;
  box-shadow: 0 2rpx 16rpx rgba(0, 0, 0, 0.04);
}
.loading-text {
  font-size: 28rpx;
  color: #6B7280;
}
.empty-card {
  background: #fff;
  border-radius: 20rpx;
  padding: 64rpx 48rpx;
  text-align: center;
  margin-bottom: 24rpx;
  box-shadow: 0 2rpx 16rpx rgba(0, 0, 0, 0.04);
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 16rpx;
}
.empty-icon {
  font-size: 56rpx;
}
.empty-text {
  font-size: 28rpx;
  color: #9CA3AF;
}

/* ========== Staff & Slots ========== */
.staff-name {
  font-size: 30rpx;
  font-weight: 700;
  color: #1E1B4B;
  margin-bottom: 16rpx;
  display: flex;
  align-items: center;
  gap: 10rpx;
}
.staff-icon {
  font-size: 28rpx;
}
.slots-grid {
  display: flex;
  flex-wrap: wrap;
  gap: 16rpx;
}
.time-section-label {
  display: block;
  font-size: 24rpx;
  color: #6B7280;
  margin-bottom: 12rpx;
}
.end-time-panel {
  margin-top: 24rpx;
  padding-top: 24rpx;
  border-top: 2rpx dashed #E5E7EB;
}
.end-time-empty {
  display: block;
  font-size: 24rpx;
  color: #D97706;
  line-height: 1.6;
}
.slot {
  padding: 18rpx 28rpx;
  background: #F9FAFB;
  border: 2rpx solid #E5E7EB;
  border-radius: 12rpx;
  font-size: 26rpx;
  color: #374151;
  font-weight: 500;
  transition: all 0.2s;
}
.slot:active {
  transform: scale(0.95);
}
.slot.selected {
  border-color: #4F46E5;
  background: linear-gradient(135deg, #4F46E5, #6366F1);
  color: #fff;
  font-weight: 600;
  box-shadow: 0 4rpx 16rpx rgba(79, 70, 229, 0.3);
}

/* ========== Confirm ========== */
.confirm-card {
  border: none;
  overflow: hidden;
}
.confirm-header {
  margin: -28rpx -28rpx 20rpx -28rpx;
  padding: 28rpx 28rpx;
  background: linear-gradient(135deg, #1E1B4B, #312E81);
}
.confirm-title {
  font-size: 30rpx;
  font-weight: 700;
  color: #fff;
  letter-spacing: 2rpx;
}
.confirm-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 22rpx 0;
  border-bottom: 1rpx solid #F3F4F6;
  font-size: 28rpx;
}
.confirm-row:last-child {
  border-bottom: none;
}
.confirm-row .label {
  color: #6B7280;
  font-size: 26rpx;
}
.confirm-row .value {
  color: #1F2937;
  font-weight: 600;
  text-align: right;
  max-width: 60%;
}
.confirm-row-amount {
  margin-top: 8rpx;
  padding-top: 24rpx;
  border-top: 2rpx dashed #E5E7EB;
  border-bottom: none;
}
.amount {
  font-size: 40rpx;
  color: #DC2626;
  font-weight: 800;
}

/* Pet card in confirm */
.pet-card {
  background: #fff;
  border-radius: 20rpx;
  padding: 28rpx;
  margin-bottom: 20rpx;
  box-shadow: 0 2rpx 12rpx rgba(0, 0, 0, 0.04);
}
.pet-card-top {
  display: flex;
  align-items: center;
  gap: 20rpx;
  padding-bottom: 24rpx;
  margin-bottom: 20rpx;
  border-bottom: 1rpx solid #F3F4F6;
}
.pet-card-avatar {
  width: 80rpx;
  height: 80rpx;
  border-radius: 50%;
  background: linear-gradient(135deg, #FEF3C7, #FDE68A);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 40rpx;
  flex-shrink: 0;
}
.pet-card-info {
  display: flex;
  flex-direction: column;
}
.pet-card-name {
  font-size: 32rpx;
  font-weight: 700;
  color: #1F2937;
}
.pet-card-breed {
  font-size: 24rpx;
  color: #9CA3AF;
  margin-top: 4rpx;
}

.card-title {
  font-size: 28rpx;
  font-weight: 700;
  color: #1F2937;
  display: block;
  margin-bottom: 18rpx;
}
.confirm-pet-block {
  padding: 20rpx 0;
  border-bottom: 2rpx solid #F3F4F6;
}
.confirm-pet-block:first-of-type {
  padding-top: 0;
}
.confirm-pet-block:last-of-type {
  border-bottom: none;
  padding-bottom: 0;
}
.confirm-pet-name {
  display: block;
  font-size: 28rpx;
  font-weight: 700;
  color: #1F2937;
}
.confirm-pet-services {
  display: block;
  margin-top: 8rpx;
  font-size: 24rpx;
  color: #6B7280;
  line-height: 1.6;
}

/* ========== Forms ========== */
.form-row {
  margin-bottom: 20rpx;
}
.form-row-half {
  display: flex;
  gap: 20rpx;
  margin-bottom: 20rpx;
}
.half {
  flex: 1;
}
.form-label {
  font-size: 24rpx;
  color: #6B7280;
  margin-bottom: 10rpx;
  display: block;
  font-weight: 500;
}
.form-input-direct {
  background: #FFFFFF;
  border: 2rpx solid #C7D2FE;
  border-radius: 14rpx;
  padding: 0 24rpx;
  font-size: 28rpx;
  width: 100%;
  box-sizing: border-box;
  color: #1F2937;
  height: 80rpx;
  line-height: 80rpx;
}
.input-wrapper {
  background: #F9FAFB;
  border: 2rpx solid #E5E7EB;
  border-radius: 14rpx;
  transition: all 0.2s;
}
.form-input {
  background: transparent;
  border: none;
  padding: 20rpx 24rpx;
  font-size: 28rpx;
  width: 100%;
  box-sizing: border-box;
  color: #1F2937;
  min-height: 60rpx;
}
.picker-value {
  color: #374151;
  font-weight: 500;
}

/* ========== Textarea ========== */
.textarea-wrapper {
  background: #F9FAFB;
  border: 2rpx dashed #C7D2FE;
  border-radius: 16rpx;
  overflow: hidden;
  margin-bottom: 4rpx;
}
.template-textarea {
  background: transparent;
  border: none;
  padding: 24rpx;
  font-size: 26rpx;
  width: 100%;
  height: 280rpx;
  box-sizing: border-box;
  color: #374151;
  line-height: 1.6;
}
.textarea-sm {
  height: 120rpx;
}

/* Parsed results grid */
.parsed-grid {
  display: flex;
  flex-wrap: wrap;
  gap: 16rpx;
}
.parsed-item {
  width: calc(50% - 8rpx);
  background: #fff;
  border: 2rpx solid #E5E7EB;
  border-radius: 16rpx;
  padding: 16rpx 20rpx;
  box-sizing: border-box;
}
.parsed-item.full-width {
  width: 100%;
}
.parsed-label {
  font-size: 22rpx;
  color: #9CA3AF;
  display: block;
  margin-bottom: 4rpx;
}
.parsed-value {
  font-size: 28rpx;
  color: #1F2937;
  font-weight: 500;
  display: block;
  word-break: break-all;
}
.textarea {
  background: transparent;
  border: none;
  border-radius: 0;
  padding: 20rpx 24rpx;
  font-size: 28rpx;
  width: 100%;
  height: 140rpx;
  color: #374151;
  box-sizing: border-box;
}

/* ========== Buttons ========== */
.btn-row {
  display: flex;
  gap: 20rpx;
  margin-top: 36rpx;
  padding-bottom: 20rpx;
}

.btn-primary {
  flex: 1;
  background: linear-gradient(135deg, #4F46E5, #6366F1);
  color: #fff;
  border: none;
  border-radius: 16rpx;
  font-size: 30rpx;
  font-weight: 700;
  padding: 24rpx 0;
  letter-spacing: 2rpx;
  box-shadow: 0 6rpx 20rpx rgba(79, 70, 229, 0.3);
  margin-top: 28rpx;
}
.btn-primary:active {
  transform: scale(0.98);
  box-shadow: 0 2rpx 10rpx rgba(79, 70, 229, 0.3);
}
.btn-primary[disabled] {
  opacity: 0.45;
  box-shadow: none;
}

.btn-secondary {
  background: #EEF2FF;
  color: #4F46E5;
  border: 2rpx solid #C7D2FE;
  border-radius: 14rpx;
  font-size: 28rpx;
  font-weight: 600;
  padding: 20rpx 0;
  margin-top: 16rpx;
  letter-spacing: 2rpx;
}
.btn-secondary:active {
  background: #E0E7FF;
}

.btn-ghost {
  flex: 0.6;
  background: #fff;
  color: #6B7280;
  border: 2rpx solid #E5E7EB;
  border-radius: 16rpx;
  font-size: 28rpx;
  font-weight: 600;
  padding: 24rpx 0;
}
.btn-ghost:active {
  background: #F9FAFB;
}

.btn-submit {
  flex: 1;
  background: linear-gradient(135deg, #059669, #10B981);
  color: #fff;
  border: none;
  border-radius: 16rpx;
  font-size: 32rpx;
  font-weight: 700;
  padding: 24rpx 0;
  letter-spacing: 4rpx;
  box-shadow: 0 6rpx 20rpx rgba(5, 150, 105, 0.3);
}
.btn-submit:active {
  transform: scale(0.98);
}

/* ========== Pet Service Table (confirm step) ========== */
.pet-svc-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12rpx;
}
.pet-svc-row-label {
  margin-bottom: 16rpx;
}
.pet-svc-title {
  font-size: 28rpx;
  font-weight: 700;
  color: #1F2937;
}
.pet-svc-add {
  font-size: 24rpx;
  color: #1E1B4B;
  background: #fff;
  border: 2rpx solid #1E1B4B;
  border-radius: 32rpx;
  padding: 10rpx 28rpx;
  font-weight: 600;
}
.pet-svc-table {
  border: 1rpx solid #E5E7EB;
  border-radius: 12rpx;
  overflow: hidden;
}
.pet-svc-table-head {
  display: flex;
  background: #F9FAFB;
  border-bottom: 1rpx solid #E5E7EB;
  padding: 16rpx 24rpx;
}
.pet-svc-th {
  flex: 1;
  font-size: 24rpx;
  color: #9CA3AF;
  font-weight: 500;
}
.pet-svc-th-op {
  width: 80rpx;
  text-align: center;
  font-size: 24rpx;
  color: #9CA3AF;
  font-weight: 500;
}
.pet-svc-table-row {
  display: flex;
  align-items: center;
  padding: 20rpx 24rpx;
  border-bottom: 1rpx solid #F3F4F6;
}
.pet-svc-table-row:last-child {
  border-bottom: none;
}
.pet-svc-td {
  flex: 1;
  font-size: 26rpx;
  color: #374151;
}
.pet-svc-td-op {
  width: 80rpx;
  text-align: center;
}
.pet-svc-del-btn {
  font-size: 24rpx;
  color: #EF4444;
  font-weight: 500;
}
.pet-svc-empty {
  font-size: 26rpx;
  color: #D1D5DB;
  text-align: center;
  padding: 40rpx 0;
  background: #FAFAFA;
  border-radius: 12rpx;
}

/* Service picker popup overlay */
.svc-picker-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  z-index: 1000;
  display: flex;
  align-items: center;
  justify-content: center;
}
.svc-picker-popup {
  background: #fff;
  border-radius: 24rpx;
  width: 90%;
  max-width: 700rpx;
  max-height: 80vh;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}
.svc-picker-popup-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 24rpx 28rpx;
  border-bottom: 2rpx solid #F3F4F6;
}
.svc-picker-popup-title {
  font-size: 32rpx;
  font-weight: 700;
  color: #1F2937;
}
.svc-picker-popup-close {
  font-size: 36rpx;
  color: #9CA3AF;
  padding: 8rpx;
}
.svc-picker-popup .svc-picker {
  height: 600rpx;
  border: none;
  border-radius: 0;
}
.svc-picker-popup-footer {
  padding: 16rpx 28rpx 24rpx;
  border-top: 2rpx solid #F3F4F6;
}
.svc-picker-popup-footer .btn-primary {
  width: 100%;
}

/* ========== Service Picker (3-level) ========== */
.svc-picker {
  display: flex;
  border: 2rpx solid #E5E7EB;
  border-radius: 16rpx;
  overflow: hidden;
  height: 700rpx;
}
.svc-picker-sidebar {
  width: 160rpx;
  min-width: 160rpx;
  background: #F9FAFB;
  border-right: 2rpx solid #E5E7EB;
  overflow-y: auto;
}
.sidebar-item {
  padding: 28rpx 16rpx;
  font-size: 26rpx;
  color: #6B7280;
  text-align: center;
  border-left: 6rpx solid transparent;
  position: relative;
}
.sidebar-item.active {
  background: #fff;
  color: #1F2937;
  font-weight: 600;
  border-left-color: #F59E0B;
}
.svc-picker-main {
  flex: 1;
  display: flex;
  flex-direction: column;
  background: #fff;
  min-width: 0;
}
.sub-tab-bar {
  white-space: nowrap;
  border-bottom: 2rpx solid #F3F4F6;
  flex-shrink: 0;
}
.sub-tab-list {
  display: inline-flex;
  padding: 16rpx 16rpx 0;
  gap: 12rpx;
}
.sub-tab {
  display: inline-block;
  padding: 12rpx 24rpx;
  font-size: 24rpx;
  color: #6B7280;
  border-radius: 32rpx;
  margin-bottom: 12rpx;
  white-space: nowrap;
}
.sub-tab.active {
  background: #FEF3C7;
  color: #92400E;
  font-weight: 600;
}
.svc-item-list {
  flex: 1;
  overflow-y: auto;
  padding: 8rpx 0;
}
.svc-empty {
  text-align: center;
  color: #9CA3AF;
  font-size: 26rpx;
  padding: 80rpx 0;
}
.svc-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 20rpx 24rpx;
  border-bottom: 1rpx solid #F3F4F6;
}
.svc-item.checked {
  background: #FFFBEB;
}
.svc-item-info {
  flex: 1;
  min-width: 0;
}
.svc-item-name {
  font-size: 28rpx;
  font-weight: 500;
  color: #1F2937;
  display: block;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.svc-item-cat {
  font-size: 22rpx;
  color: #9CA3AF;
  display: block;
  margin-top: 4rpx;
}
.svc-item-right {
  display: flex;
  align-items: center;
  gap: 16rpx;
  margin-left: 16rpx;
  flex-shrink: 0;
}
.svc-item-price {
  font-size: 30rpx;
  font-weight: 700;
  color: #DC2626;
}
.svc-item-check {
  width: 40rpx;
  height: 40rpx;
  border-radius: 50%;
  border: 3rpx solid #D1D5DB;
  box-sizing: border-box;
}
.svc-item-check.on {
  background: #F59E0B;
  border-color: #F59E0B;
  position: relative;
}
.svc-item-check.on::after {
  content: '';
  position: absolute;
  left: 50%;
  top: 45%;
  width: 12rpx;
  height: 20rpx;
  border: solid #fff;
  border-width: 0 3rpx 3rpx 0;
  transform: translate(-50%, -50%) rotate(45deg);
}

/* Selected bar */
.selected-bar {
  background: #FFFBEB;
  border: 2rpx solid #FDE68A;
  border-radius: 12rpx;
  padding: 16rpx 24rpx;
  margin-bottom: 16rpx;
  text-align: center;
}
.selected-bar-text {
  font-size: 26rpx;
  color: #92400E;
  font-weight: 600;
}
</style>
