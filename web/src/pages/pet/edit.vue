<template>
  <view class="page">
    <!-- 基本身份 -->
    <view class="section">
      <text class="section-title">基本信息</text>
      <view class="form-item">
        <text class="label">猫咪名 *</text>
        <input v-model="form.name" placeholder="请输入猫咪名" class="input" />
      </view>
      <view class="row">
        <view class="col">
          <text class="label">品种</text>
          <input v-model="form.breed" placeholder="英短蓝猫" class="input" />
        </view>
        <view class="col">
          <text class="label">主人手机号</text>
          <input v-model="form.owner_phone" type="tel" placeholder="可选" class="input" maxlength="11" />
        </view>
      </view>
    </view>

    <!-- 生理特征 -->
    <view class="section">
      <text class="section-title">生理特征</text>
      <view class="form-item">
        <text class="label">性别</text>
        <view class="gender-btns">
          <view :class="['gender-btn', form.gender === 1 ? 'active-male' : '']" @click="form.gender = 1">弟弟 ♂</view>
          <view :class="['gender-btn', form.gender === 2 ? 'active-female' : '']" @click="form.gender = 2">妹妹 ♀</view>
        </view>
      </view>
      <view class="row">
        <view class="col">
          <text class="label">年龄</text>
          <view class="input-unit">
            <input v-model="ageInput" type="digit" placeholder="0" class="input flex1" @input="onAgeInput" />
            <text class="unit">岁</text>
          </view>
        </view>
        <view class="col">
          <text class="label">体重</text>
          <view class="input-unit">
            <input v-model="form.weight" type="digit" placeholder="0" class="input flex1" />
            <text class="unit">kg</text>
          </view>
        </view>
      </view>
      <view class="form-item">
        <text class="label">出生日期</text>
        <picker mode="date" :value="form.birth_date" @change="onBirthDateChange">
          <view class="picker">{{ form.birth_date || '由年龄自动计算，或手动选择' }} ›</view>
        </picker>
      </view>
      <view class="form-item inline-switch">
        <text class="label">是否绝育</text>
        <switch :checked="form.neutered" @change="(e: any) => form.neutered = e.detail.value" />
      </view>
    </view>

    <!-- 外观特征 -->
    <view class="section">
      <text class="section-title">外观特征</text>
      <view class="row">
        <view class="col">
          <text class="label">毛色</text>
          <input v-model="form.coat_color" placeholder="白色、蓝色" class="input" />
        </view>
        <view class="col">
          <text class="label">毛发等级 *</text>
          <view class="fur-tags">
            <view
              v-for="fl in furLevels"
              :key="fl"
              :class="['fur-tag', form.fur_level === fl ? 'fur-active' : '']"
              @click="form.fur_level = fl"
              @longpress="onFurLongPress(fl)"
            >{{ fl }}</view>
            <view class="fur-tag fur-add" @click="showFurAdd = true">+</view>
          </view>
        </view>
      </view>
    </view>

    <!-- 性格与护理 -->
    <view class="section">
      <text class="section-title">性格与护理</text>
      <view class="form-item">
        <text class="label">性格特征（点击选择，长按排序/删除）</text>
        <view class="personality-tags">
          <view
            v-for="(p, idx) in personalities" :key="p"
            :class="['p-tag', form.personality === p ? 'p-selected' : '', draggingIdx === idx ? 'p-dragging' : '']"
            :style="{ background: form.personality === p ? getPersonalityColor(p) : getPersonalityBg(p), color: form.personality === p ? '#fff' : getPersonalityColor(p), borderColor: getPersonalityColor(p) }"
            @click="form.personality = form.personality === p ? '' : p"
            @longpress.prevent="onPersonalityLongPress(p, idx)"
            @touchstart="onPTagTouchStart($event, idx)"
            @touchmove.prevent="onPTagTouchMove($event)"
            @touchend="onPTagTouchEnd"
          >{{ p }}</view>
          <view class="p-tag p-add" @click="showPersonalityAdd = true">+</view>
        </view>
      </view>
      <view class="form-item">
        <text class="label">攻击性</text>
        <view class="aggression-tags">
          <view
            v-for="a in aggressions" :key="a"
            :class="['a-tag', form.aggression === a ? 'a-selected' : '']"
            :style="form.aggression === a ? { background: aggressionColorMap[a], color: '#fff' } : {}"
            @click="form.aggression = form.aggression === a ? '' : a"
          >{{ a }}</view>
        </view>
      </view>
      <view class="row">
        <view class="col">
          <text class="label">洗澡频率</text>
          <picker :range="bathFreqs" :value="bathFreqIndex" @change="(e: any) => form.bath_frequency = bathFreqs[e.detail.value]">
            <view class="picker">{{ form.bath_frequency || '请选择' }} ›</view>
          </picker>
        </view>
        <view class="col">
          <text class="label">禁区</text>
          <input v-model="form.forbidden_zones" placeholder="肚子、尾巴" class="input" />
        </view>
      </view>
    </view>

    <!-- 备注 -->
    <view class="section">
      <text class="section-title">备注信息</text>
      <view class="form-item">
        <text class="label">洗护注意事项</text>
        <textarea v-model="form.care_notes" placeholder="特别注意的洗护事项" class="textarea" />
      </view>
      <view class="form-item">
        <text class="label">行为备注</text>
        <textarea v-model="form.behavior_notes" placeholder="其他行为特点" class="textarea" />
      </view>
    </view>

    <button class="btn-submit" @click="onSubmit" :loading="submitting">{{ id ? '保存' : '新增' }}</button>
    <button class="btn-delete" v-if="id" @click="onDelete">删除猫咪</button>

    <!-- 毛发等级新增弹窗 -->
    <view class="modal-mask" v-if="showFurAdd" @click="showFurAdd = false">
      <view class="modal-body" @click.stop>
        <text class="modal-title">新增毛发等级</text>
        <input v-model="newFurName" placeholder="输入名称" class="input" style="margin: 16rpx 0;" />
        <view class="modal-btns">
          <view class="modal-btn cancel" @click="showFurAdd = false">取消</view>
          <view class="modal-btn confirm" @click="addFurLevel">确定</view>
        </view>
      </view>
    </view>

    <!-- 新增性格弹窗 -->
    <view class="modal-mask" v-if="showPersonalityAdd" @click="showPersonalityAdd = false">
      <view class="modal-body" @click.stop>
        <text class="modal-title">新增性格标签</text>
        <input v-model="newPersonalityName" placeholder="输入性格名称" class="input" style="margin: 16rpx 0;" />
        <text class="label">选择颜色（风险等级）</text>
        <view class="color-options">
          <view
            v-for="co in colorOptions" :key="co.value"
            :class="['color-opt', newPersonalityColor === co.value ? 'color-selected' : '']"
            :style="{ background: co.value }"
            @click="newPersonalityColor = co.value"
          >
            <text class="color-opt-text">{{ co.name }}</text>
          </view>
        </view>
        <view class="modal-btns">
          <view class="modal-btn cancel" @click="showPersonalityAdd = false">取消</view>
          <view class="modal-btn confirm" @click="addPersonality">确定</view>
        </view>
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { onLoad } from '@dcloudio/uni-app'
import { getPet, createPet, updatePet, deletePet } from '@/api/pet'
import { safeBack } from '@/utils/navigate'
import { getPersonalityColor, getPersonalityBg, personalityColors, colorOptions } from '@/utils/personality'

const id = ref(0)
const submitting = ref(false)
const personalities = ref(['神仙宝贝', '胆大开放', '胆小敏感', '过度活跃', '笑里藏刀', '绝世凶兽'])
const aggressions = ['无', '可能', '有']
const aggressionColorMap: Record<string, string> = { '无': '#10B981', '可能': '#F59E0B', '有': '#EF4444' }
const showPersonalityAdd = ref(false)
const newPersonalityName = ref('')
const newPersonalityColor = ref('#F59E0B')

// Drag & longpress for personality tags
const draggingIdx = ref(-1)
let dragStartX = 0
let dragMoved = false

function onPersonalityLongPress(name: string, idx: number) {
  uni.showActionSheet({
    itemList: ['上移', '下移', `删除「${name}」`],
    success: (res) => {
      if (res.tapIndex === 0 && idx > 0) {
        // Move up
        const arr = [...personalities.value]
        ;[arr[idx - 1], arr[idx]] = [arr[idx], arr[idx - 1]]
        personalities.value = arr
      } else if (res.tapIndex === 1 && idx < personalities.value.length - 1) {
        // Move down
        const arr = [...personalities.value]
        ;[arr[idx], arr[idx + 1]] = [arr[idx + 1], arr[idx]]
        personalities.value = arr
      } else if (res.tapIndex === 2) {
        personalities.value = personalities.value.filter(p => p !== name)
        if (form.value.personality === name) form.value.personality = ''
      }
    }
  })
}

function onPTagTouchStart(e: any, idx: number) {
  dragStartX = e.touches[0].clientX
  dragMoved = false
  draggingIdx.value = idx
}

function onPTagTouchMove(e: any) {
  if (draggingIdx.value < 0) return
  const dx = e.touches[0].clientX - dragStartX
  if (Math.abs(dx) > 40) {
    dragMoved = true
    const arr = [...personalities.value]
    const fromIdx = draggingIdx.value
    if (dx > 0 && fromIdx < arr.length - 1) {
      // Swipe right = move right
      ;[arr[fromIdx], arr[fromIdx + 1]] = [arr[fromIdx + 1], arr[fromIdx]]
      personalities.value = arr
      draggingIdx.value = fromIdx + 1
      dragStartX = e.touches[0].clientX
    } else if (dx < 0 && fromIdx > 0) {
      // Swipe left = move left
      ;[arr[fromIdx - 1], arr[fromIdx]] = [arr[fromIdx], arr[fromIdx - 1]]
      personalities.value = arr
      draggingIdx.value = fromIdx - 1
      dragStartX = e.touches[0].clientX
    }
  }
}

function onPTagTouchEnd() {
  draggingIdx.value = -1
}

function addPersonality() {
  const name = newPersonalityName.value.trim()
  if (!name) return
  if (personalities.value.includes(name)) {
    uni.showToast({ title: '已存在', icon: 'none' }); return
  }
  personalities.value.push(name)
  personalityColors[name] = newPersonalityColor.value
  form.value.personality = name
  newPersonalityName.value = ''
  showPersonalityAdd.value = false
}
const bathFreqs = ['每月', '两月', '三月', '半年', '一年']

const furLevels = ref<string[]>(['短毛猫', '长毛猫', 'A', 'B', 'C', 'D'])
const showFurAdd = ref(false)
const newFurName = ref('')

const ageInput = ref('')

const form = ref({
  name: '', owner_phone: '', breed: '', gender: 0, birth_date: '',
  weight: '', coat_color: '', fur_level: '', personality: '',
  aggression: '', forbidden_zones: '', bath_frequency: '',
  neutered: false, care_notes: '', behavior_notes: '',
})

const aggressionIndex = computed(() => Math.max(aggressions.indexOf(form.value.aggression), 0))
const bathFreqIndex = computed(() => Math.max(bathFreqs.indexOf(form.value.bath_frequency), 0))

let syncing = false

// 年龄 -> 出生日期（实时）
function onAgeInput() {
  if (syncing) return
  const age = parseFloat(ageInput.value)
  if (!age || age <= 0) { form.value.birth_date = ''; return }
  syncing = true
  const now = new Date()
  const totalMonths = Math.round(age * 12)
  const birthDate = new Date(now.getFullYear(), now.getMonth() - totalMonths, 1)
  form.value.birth_date = birthDate.toISOString().split('T')[0]
  syncing = false
}

// 出生日期 -> 年龄
function onBirthDateChange(e: any) {
  if (syncing) return
  syncing = true
  form.value.birth_date = e.detail.value
  const birth = new Date(e.detail.value)
  const now = new Date()
  const months = (now.getFullYear() - birth.getFullYear()) * 12 + (now.getMonth() - birth.getMonth())
  ageInput.value = months > 0 ? (months / 12).toFixed(1) : ''
  syncing = false
}

// 毛发等级长按删除
function onFurLongPress(fl: string) {
  uni.showActionSheet({
    itemList: [`删除「${fl}」`],
    success: (res) => {
      if (res.tapIndex === 0) {
        furLevels.value = furLevels.value.filter(f => f !== fl)
        if (form.value.fur_level === fl) form.value.fur_level = ''
      }
    }
  })
}

// 新增毛发等级
function addFurLevel() {
  const name = newFurName.value.trim()
  if (!name) return
  if (furLevels.value.includes(name)) {
    uni.showToast({ title: '已存在', icon: 'none' })
    return
  }
  furLevels.value.push(name)
  form.value.fur_level = name
  newFurName.value = ''
  showFurAdd.value = false
}

onLoad((query) => {
  if (query?.id) {
    id.value = parseInt(query.id)
    loadData()
  }
  if (query?.owner_phone) {
    form.value.owner_phone = query.owner_phone
  }
})

async function loadData() {
  const res = await getPet(id.value)
  const d = res.data
  form.value = {
    name: d.name, owner_phone: d.customer?.phone || '',
    breed: d.breed, gender: d.gender, birth_date: d.birth_date ? d.birth_date.split('T')[0] : '',
    weight: String(d.weight || ''),
    coat_color: d.coat_color, fur_level: d.fur_level || '',
    personality: d.personality || '', aggression: d.aggression || '',
    forbidden_zones: d.forbidden_zones || '', bath_frequency: d.bath_frequency || '',
    neutered: d.neutered, care_notes: d.care_notes || '',
    behavior_notes: d.behavior_notes || '',
  }
  // 反算年龄
  if (form.value.birth_date) {
    const birth = new Date(form.value.birth_date)
    const now = new Date()
    const months = (now.getFullYear() - birth.getFullYear()) * 12 + (now.getMonth() - birth.getMonth())
    ageInput.value = months > 0 ? (months / 12).toFixed(1) : ''
  }
}

async function onSubmit() {
  if (!form.value.name) {
    uni.showToast({ title: '请填写猫咪名', icon: 'none' }); return
  }
  submitting.value = true
  try {
    const data: any = {
      ...form.value,
      species: '猫',
      weight: form.value.weight ? parseFloat(form.value.weight) : 0,
    }
    if (id.value) { await updatePet(id.value, data) }
    else { await createPet(data) }
    uni.showToast({ title: '保存成功', icon: 'success' })
    setTimeout(() => safeBack(), 500)
  } finally { submitting.value = false }
}

async function onDelete() {
  uni.showModal({
    title: '确认删除', content: '确认删除该猫咪档案？',
    success: async (res) => {
      if (res.confirm) {
        await deletePet(id.value)
        uni.showToast({ title: '已删除', icon: 'success' })
        setTimeout(() => safeBack(), 500)
      }
    }
  })
}
</script>

<style scoped>
.page { padding: 24rpx; }

/* 分区 */
.section { background: #fff; border-radius: 16rpx; padding: 20rpx 24rpx; margin-bottom: 20rpx; }
.section-title { font-size: 26rpx; font-weight: 600; color: #6B7280; display: block; margin-bottom: 16rpx; }

/* 表单项 */
.form-item { margin-bottom: 16rpx; }
.form-item:last-child { margin-bottom: 0; }
.label { font-size: 26rpx; color: #374151; display: block; margin-bottom: 8rpx; }
.input { font-size: 28rpx; color: #1F2937; height: 60rpx; background: #F9FAFB; border-radius: 8rpx; padding: 0 16rpx; }
.textarea { font-size: 28rpx; color: #1F2937; width: 100%; height: 120rpx; background: #F9FAFB; border-radius: 8rpx; padding: 12rpx 16rpx; }
.picker { font-size: 28rpx; color: #1F2937; height: 60rpx; line-height: 60rpx; background: #F9FAFB; border-radius: 8rpx; padding: 0 16rpx; }

/* 两列布局 */
.row { display: flex; gap: 16rpx; margin-bottom: 16rpx; }
.row:last-child { margin-bottom: 0; }
.col { flex: 1; min-width: 0; }

/* 带单位的输入框 */
.input-unit { display: flex; align-items: center; background: #F9FAFB; border-radius: 8rpx; padding-right: 12rpx; }
.input-unit .input { background: transparent; flex: 1; }
.unit { font-size: 26rpx; color: #6B7280; white-space: nowrap; }
.flex1 { flex: 1; }

/* 性别按钮 */
.gender-btns { display: flex; gap: 16rpx; }
.gender-btn { flex: 1; text-align: center; padding: 16rpx 0; font-size: 28rpx; border-radius: 12rpx; background: #F3F4F6; color: #6B7280; transition: all 0.2s; }
.active-male { background: #DBEAFE; color: #2563EB; font-weight: 600; }
.active-female { background: #FCE7F3; color: #DB2777; font-weight: 600; }

/* 毛发等级标签 */
.fur-tags { display: flex; flex-wrap: wrap; gap: 10rpx; }
.fur-tag { font-size: 24rpx; padding: 8rpx 18rpx; border-radius: 20rpx; background: #F3F4F6; color: #6B7280; }
.fur-active { background: #4F46E5; color: #fff; font-weight: 600; }
.fur-add { background: #EEF2FF; color: #4F46E5; font-weight: 600; }

/* 绝育 inline */
.inline-switch { display: flex; justify-content: space-between; align-items: center; }
.inline-switch .label { margin-bottom: 0; }

/* 性格标签 */
.personality-tags { display: flex; flex-wrap: wrap; gap: 12rpx; }
.p-tag {
  font-size: 26rpx; padding: 10rpx 22rpx; border-radius: 24rpx;
  border: 2rpx solid transparent; font-weight: 500;
  transition: all 0.2s;
}
.p-selected { font-weight: 700; box-shadow: 0 2rpx 8rpx rgba(0,0,0,0.15); }
.p-dragging { transform: scale(1.1); box-shadow: 0 4rpx 16rpx rgba(0,0,0,0.25); z-index: 10; position: relative; }
.p-add { background: #EEF2FF !important; color: #4F46E5 !important; border-color: #C7D2FE !important; font-weight: 600; }

/* 攻击性标签 */
.aggression-tags { display: flex; gap: 12rpx; }
.a-tag {
  flex: 1; text-align: center; padding: 14rpx 0; font-size: 26rpx;
  border-radius: 12rpx; background: #F3F4F6; color: #6B7280;
  border: 2rpx solid transparent; font-weight: 500; transition: all 0.2s;
}
.a-selected { font-weight: 700; border-color: transparent; }

/* 颜色选项 */
.color-options { display: flex; gap: 12rpx; margin: 12rpx 0; }
.color-opt {
  flex: 1; padding: 14rpx 8rpx; border-radius: 12rpx; text-align: center;
  border: 3rpx solid transparent; transition: all 0.2s;
}
.color-selected { border-color: #1F2937; box-shadow: 0 2rpx 8rpx rgba(0,0,0,0.2); }
.color-opt-text { font-size: 22rpx; color: #fff; font-weight: 600; }

/* 按钮 */
.btn-submit { background: #4F46E5; color: #fff; border-radius: 12rpx; font-size: 30rpx; margin-top: 8rpx; }
.btn-delete { background: #fff; color: #DC2626; border: 1rpx solid #DC2626; border-radius: 12rpx; font-size: 30rpx; margin-top: 16rpx; }

/* 弹窗 */
.modal-mask { position: fixed; top: 0; left: 0; right: 0; bottom: 0; background: rgba(0,0,0,0.5); display: flex; align-items: center; justify-content: center; z-index: 999; }
.modal-body { background: #fff; border-radius: 16rpx; padding: 32rpx; width: 560rpx; }
.modal-title { font-size: 30rpx; font-weight: 600; color: #1F2937; display: block; text-align: center; }
.modal-btns { display: flex; gap: 16rpx; margin-top: 16rpx; }
.modal-btn { flex: 1; text-align: center; padding: 16rpx; border-radius: 8rpx; font-size: 28rpx; }
.cancel { background: #F3F4F6; color: #6B7280; }
.confirm { background: #4F46E5; color: #fff; }
</style>
