<template>
  <view v-if="visible" class="care-report-mask" @click="onMaskClick">
    <view class="care-report-panel" @click.stop>
      <view class="care-report-header">
        <view class="care-report-copy">
          <text class="care-report-title">生成护理报告</text>
          <text class="care-report-subtitle">{{ previewUrl ? '报告已生成，可直接保存图片' : '填写护理信息后生成报告' }}</text>
        </view>
        <text class="care-report-close" @click="handleClose">✕</text>
      </view>

      <template v-if="previewUrl">
        <view class="care-report-preview">
          <text class="care-report-preview-hint">护理报告生成完成</text>
          <image :src="previewImageUrl" mode="widthFix" class="care-report-preview-image" show-menu-by-longpress />
        </view>
        <view class="care-report-actions generated">
          <view class="care-report-btn primary" @click="savePreview">{{ saving ? '处理中...' : '保存图片' }}</view>
          <view class="care-report-btn secondary" @click="resetPreview">重新填写</view>
          <view class="care-report-btn ghost" @click="handleClose">关闭</view>
        </view>
      </template>

      <template v-else-if="!draft">
        <view class="care-report-select">
          <text class="care-report-select-title">选择本次生成报告的猫咪</text>
          <view class="care-report-pet-list">
            <view
              v-for="pet in petOptions"
              :key="pet.petId"
              class="care-report-pet-card"
              @click="selectPet(pet.petId)"
            >
              <text class="care-report-pet-name">{{ pet.petName || '当前猫咪' }}</text>
              <text class="care-report-pet-meta">点击进入填写</text>
            </view>
          </view>
        </view>
      </template>

      <template v-else>
        <view class="care-report-form">
          <view v-if="petOptions.length > 1" class="care-report-switcher">
            <view
              v-for="pet in petOptions"
              :key="`switch-${pet.petId}`"
              :class="['care-report-switch-chip', selectedPetId === pet.petId ? 'active' : '']"
              @click="selectPet(pet.petId)"
            >
              {{ pet.petName || '当前猫咪' }}
            </view>
          </view>

          <view class="care-report-form-section care-report-basic-section">
            <text class="care-report-section-title">基本信息</text>

            <view id="care-report-anchor-portrait" class="care-report-photo-row">
              <image
                v-if="draft.portraitUrl"
                :src="resolveAbsoluteUrl(draft.portraitUrl)"
                class="care-report-photo-preview"
                mode="aspectFill"
              />
              <view v-else class="care-report-photo-placeholder">照片</view>
              <view class="care-report-photo-copy">
                <text class="care-report-field-label">护理照片</text>
                <text class="care-report-field-hint">用于报告右上角展示</text>
              </view>
              <view class="care-report-photo-button" @click="choosePortrait">
                {{ uploadingPortrait ? '上传中...' : (draft.portraitUrl ? '更换照片' : '上传照片') }}
              </view>
            </view>

            <view class="care-report-form-row">
              <view class="care-report-field">
                <text class="care-report-field-label">宝贝名字</text>
                <input
                  :value="draft.petName"
                  class="care-report-input"
                  maxlength="20"
                  placeholder="填写宝贝名字"
                  @input="updateDraftTextField('petName', $event)"
                />
              </view>

              <view class="care-report-field">
                <text class="care-report-field-label">品种</text>
                <input
                  :value="draft.breed"
                  class="care-report-input"
                  maxlength="30"
                  placeholder="填写品种"
                  @input="updateDraftTextField('breed', $event)"
                />
              </view>

              <view class="care-report-field">
                <text class="care-report-field-label">性别</text>
                <view class="care-report-segmented">
                  <view :class="['care-report-segment', draft.gender === 'GG' ? 'active' : '']" @click="setGender('GG')">GG</view>
                  <view :class="['care-report-segment', draft.gender === 'MM' ? 'active' : '']" @click="setGender('MM')">MM</view>
                </view>
              </view>

              <view class="care-report-field">
                <text class="care-report-field-label">年龄</text>
                <input
                  :value="draft.age"
                  class="care-report-input"
                  maxlength="20"
                  placeholder="例如 2岁1月"
                  @input="updateDraftTextField('age', $event)"
                />
              </view>

              <view id="care-report-anchor-care-content" class="care-report-field">
                <text class="care-report-field-label">护理内容</text>
                <textarea
                  :value="draft.careContent"
                  class="care-report-textarea"
                  maxlength="60"
                  auto-height
                  placeholder="填写本次护理项目"
                  @input="updateDraftTextField('careContent', $event)"
                />
              </view>

              <view id="care-report-anchor-weight" class="care-report-field">
                <text class="care-report-field-label">体重</text>
                <input
                  :value="draft.weight"
                  class="care-report-input"
                  maxlength="20"
                  inputmode="decimal"
                  placeholder="例如 5.58 KG"
                  @input="updateDraftTextField('weight', $event)"
                />
              </view>

              <view id="care-report-anchor-care-date" class="care-report-field">
                <text class="care-report-field-label">护理日期</text>
                <picker mode="date" :value="formatPickerDate(draft.careDate)" @change="updateDraftDate('careDate', $event)">
                  <view class="care-report-picker">{{ draft.careDate || '请选择护理日期' }}</view>
                </picker>
              </view>

              <view id="care-report-anchor-next-care-date" class="care-report-field">
                <text class="care-report-field-label">建议下次护理日期</text>
                <picker mode="date" :value="formatPickerDate(draft.nextCareDate)" @change="updateDraftDate('nextCareDate', $event)">
                  <view class="care-report-picker">{{ draft.nextCareDate || '请选择下次护理日期' }}</view>
                </picker>
              </view>
            </view>
          </view>

          <view id="care-report-anchor-body-shape" class="care-report-form-section">
            <text class="care-report-section-title">体型</text>
            <view class="care-report-choice-grid body-shape">
              <view
                v-for="option in bodyShapeOptions"
                :key="option.value"
                :class="['care-report-choice', draft.bodyShape === option.value ? 'active' : '']"
                @click="setBodyShape(option.value)"
              >
                {{ option.label }}
              </view>
            </view>
          </view>

          <view class="care-report-inspection-list">
            <view
              v-for="section in sectionDefinitions"
              :key="section.key"
              class="care-report-inspection-section"
            >
              <view class="care-report-section-head" @click="toggleSectionPanel(section.key)">
                <text class="care-report-section-title">{{ section.label }}</text>
                <view class="care-report-section-summary">
                  <text>{{ getSectionSelectedCount(section.key) }}项</text>
                  <text class="care-report-section-toggle">{{ expandedSectionKey === section.key ? '−' : '+' }}</text>
                </view>
              </view>
              <view v-if="expandedSectionKey === section.key" class="care-report-section-body">
                <view class="care-report-choice-grid">
                  <view
                    v-for="option in section.options"
                    :key="option.value"
                    :class="['care-report-choice', isSectionCheckSelected(section.key, option.value) ? 'active' : '']"
                    @click="toggleSectionCheck(section.key, option.value)"
                  >
                    {{ option.label }}
                  </view>
                </view>
                <textarea
                  :value="getSectionNote(section.key)"
                  class="care-report-textarea care-report-section-note"
                  maxlength="120"
                  auto-height
                  placeholder="备注（选填）"
                  @input="updateSectionNote(section.key, $event)"
                />
              </view>
            </view>
          </view>
        </view>

        <view class="care-report-actions">
          <view class="care-report-btn secondary" @click="togglePreview">预览报告</view>
          <view class="care-report-btn primary" @click="submit">{{ submitting ? '生成中...' : '生成报告' }}</view>
        </view>

        <view v-if="previewExpanded" class="care-report-draft-preview" @click.stop>
          <view class="care-report-draft-preview-head">
            <view class="care-report-copy">
              <text class="care-report-title">报告预览</text>
              <text class="care-report-subtitle">预览仅用于核对，返回后可继续填写</text>
            </view>
            <text class="care-report-preview-back" @click="togglePreview">返回填写</text>
          </view>
          <view class="care-report-draft-preview-scroll">
            <OrderCareReportStage :draft="draft" />
          </view>
          <view class="care-report-actions care-report-draft-preview-actions">
            <view class="care-report-btn secondary" @click="togglePreview">返回填写</view>
            <view class="care-report-btn primary" @click="submit">{{ submitting ? '生成中...' : '生成报告' }}</view>
          </view>
        </view>
      </template>
    </view>

    <ImageCropper
      :visible="cropperVisible"
      :src="cropperSrc"
      @cancel="onCropCancel"
      @confirm="onCropConfirm"
    />
  </view>
</template>

<script setup lang="ts">
import { computed, nextTick, ref, watch } from 'vue'
import ImageCropper from '@/components/ImageCropper.vue'
import OrderCareReportStage from '@/components/order/OrderCareReportStage.vue'
import { createOrderCareReport } from '@/api/order-care-report'
import { uploadH5File } from '@/api/upload'
import {
  buildOrderCareReportDraft,
  buildOrderCareReportPayload,
  buildOrderCareReportPetOptions,
  orderCareReportBodyShapeOptions,
  orderCareReportSectionDefinitions,
  type OrderCareReportDraft,
  type OrderCareReportPetOption,
  type OrderCareReportSectionDefinition,
  type OrderCareReportSectionKey,
} from '@/utils/order-care-report'
import { createCropperPreviewUrl } from '@/utils/image-cropper'
import { buildOrderCareReportFileName, saveImageByUrl } from '@/utils/web-image-save'

type TextDraftField = 'petName' | 'breed' | 'age' | 'careContent' | 'weight'

const props = defineProps<{
  visible: boolean
  order: any | null
}>()

const emit = defineEmits<{
  (e: 'close'): void
}>()

const selectedPetId = ref(0)
const draft = ref<OrderCareReportDraft | null>(null)
const previewUrl = ref('')
const previewExpanded = ref(false)
const expandedSectionKey = ref<OrderCareReportSectionKey | null>('skin')
const submitting = ref(false)
const saving = ref(false)
const uploadingPortrait = ref(false)
const cropperVisible = ref(false)
const cropperSrc = ref('')
const modalSessionToken = ref(0)

const bodyShapeOptions = orderCareReportBodyShapeOptions
const sectionDefinitions: OrderCareReportSectionDefinition[] = orderCareReportSectionDefinitions

const petOptions = computed<OrderCareReportPetOption[]>(() => {
  if (!props.order) return []
  return buildOrderCareReportPetOptions(props.order)
})

const previewImageUrl = computed(() => resolveAbsoluteUrl(previewUrl.value))

watch(
  () => [props.visible, props.order?.ID],
  ([visible]) => {
    if (visible) {
      initializeDraft()
      return
    }
    resetState()
  },
  { immediate: true }
)

function initializeDraft() {
  modalSessionToken.value += 1
  previewUrl.value = ''
  previewExpanded.value = false
  expandedSectionKey.value = 'skin'
  cropperVisible.value = false
  clearCropperSrc()
  const options = petOptions.value
  if (options.length === 1) {
    selectPet(options[0].petId)
    return
  }
  selectedPetId.value = 0
  draft.value = null
}

function resetState() {
  modalSessionToken.value += 1
  selectedPetId.value = 0
  draft.value = null
  previewUrl.value = ''
  previewExpanded.value = false
  expandedSectionKey.value = 'skin'
  cropperVisible.value = false
  clearCropperSrc()
  submitting.value = false
  saving.value = false
  uploadingPortrait.value = false
  uni.hideLoading()
}

function clearCropperSrc() {
  if (cropperSrc.value && cropperSrc.value.startsWith('blob:')) {
    URL.revokeObjectURL(cropperSrc.value)
  }
  cropperSrc.value = ''
}

function selectPet(petId: number) {
  if (!props.order) return
  selectedPetId.value = petId
  draft.value = buildOrderCareReportDraft(props.order, petId)
  previewUrl.value = ''
  previewExpanded.value = false
  expandedSectionKey.value = 'skin'
}

function handleClose() {
  emit('close')
}

function onMaskClick() {
  if (cropperVisible.value) return
  handleClose()
}

function formatPickerDate(value: string) {
  return String(value || '').replace(/\./g, '-')
}

function updateDraftTextField(key: TextDraftField, event: any) {
  if (!draft.value) return
  draft.value[key] = String(event?.detail?.value || '')
}

function updateDraftDate(field: 'careDate' | 'nextCareDate', event: any) {
  if (!draft.value) return
  draft.value[field] = String(event?.detail?.value || '').replace(/-/g, '.')
}

function setGender(value: 'GG' | 'MM') {
  if (!draft.value) return
  draft.value.gender = value
}

function setBodyShape(value: string) {
  if (!draft.value) return
  draft.value.bodyShape = value
}

function getSectionValue(sectionKey: OrderCareReportSectionKey) {
  return draft.value?.[sectionKey]
}

function getSectionSelectedCount(sectionKey: OrderCareReportSectionKey) {
  return getSectionValue(sectionKey)?.checks.length || 0
}

function isSectionCheckSelected(sectionKey: OrderCareReportSectionKey, value: string) {
  return Boolean(getSectionValue(sectionKey)?.checks.includes(value))
}

function getSectionNote(sectionKey: OrderCareReportSectionKey) {
  return getSectionValue(sectionKey)?.note || ''
}

function updateSectionNote(sectionKey: OrderCareReportSectionKey, event: any) {
  const section = getSectionValue(sectionKey)
  if (!section) return
  section.note = String(event?.detail?.value || '')
}

function toggleSectionCheck(sectionKey: OrderCareReportSectionKey, code: string) {
  const section = getSectionValue(sectionKey)
  if (!section) return
  const values = new Set(section.checks)
  if (values.has(code)) values.delete(code)
  else values.add(code)
  section.checks = Array.from(values)
}

function toggleSectionPanel(sectionKey: OrderCareReportSectionKey) {
  expandedSectionKey.value = expandedSectionKey.value === sectionKey ? null : sectionKey
}

function togglePreview() {
  previewExpanded.value = !previewExpanded.value
}

async function choosePortrait() {
  if (uploadingPortrait.value) return
  if (typeof document === 'undefined' || typeof window === 'undefined' || typeof File === 'undefined') {
    uni.showToast({ title: '当前环境暂不支持裁剪上传', icon: 'none' })
    return
  }

  const input = document.createElement('input')
  input.type = 'file'
  input.accept = 'image/*'
  input.style.display = 'none'
  document.body.appendChild(input)
  input.onchange = async () => {
    const file = input.files?.[0]
    if (!file) {
      input.remove()
      return
    }
    try {
      clearCropperSrc()
      cropperSrc.value = await createCropperPreviewUrl(file)
      cropperVisible.value = true
    } catch (error: any) {
      uni.showToast({ title: error?.message || '图片读取失败', icon: 'none' })
    } finally {
      input.value = ''
      input.remove()
    }
  }
  input.click()
}

function onCropCancel() {
  cropperVisible.value = false
  clearCropperSrc()
}

async function onCropConfirm(blob: Blob) {
  if (!draft.value) return
  const currentDraft = draft.value
  const sessionToken = modalSessionToken.value
  cropperVisible.value = false
  clearCropperSrc()
  uploadingPortrait.value = true
  uni.showLoading({ title: '上传中...' })
  try {
    const file = new File([blob], 'care-report.jpg', { type: 'image/jpeg' })
    const portraitUrl = await uploadH5File(file)
    if (modalSessionToken.value !== sessionToken || draft.value !== currentDraft || !props.visible) return
    currentDraft.portraitUrl = portraitUrl
    uni.showToast({ title: '照片上传成功', icon: 'success' })
  } catch (error: any) {
    if (modalSessionToken.value !== sessionToken || !props.visible) return
    uni.showToast({ title: error?.message || '上传失败', icon: 'none' })
  } finally {
    if (modalSessionToken.value === sessionToken) {
      uploadingPortrait.value = false
      uni.hideLoading()
    }
  }
}

function focusValidationTarget(anchorId: string) {
  previewExpanded.value = false
  nextTick(() => {
    if (typeof document === 'undefined') return
    document.getElementById(anchorId)?.scrollIntoView({ behavior: 'smooth', block: 'center' })
  })
}

function validateDraft() {
  if (!draft.value) {
    uni.showToast({ title: '请先选择猫咪', icon: 'none' })
    return false
  }

  const requiredChecks = [
    { valid: () => Boolean(draft.value?.portraitUrl), anchor: 'care-report-anchor-portrait', message: '请先上传护理照片' },
    { valid: () => Boolean(draft.value?.careContent.trim()), anchor: 'care-report-anchor-care-content', message: '请填写护理内容' },
    { valid: () => Boolean(draft.value?.weight.trim()), anchor: 'care-report-anchor-weight', message: '请填写体重' },
    { valid: () => Boolean(draft.value?.careDate), anchor: 'care-report-anchor-care-date', message: '请填写护理日期' },
    { valid: () => Boolean(draft.value?.nextCareDate), anchor: 'care-report-anchor-next-care-date', message: '请填写建议下次护理日期' },
    { valid: () => Boolean(draft.value?.bodyShape), anchor: 'care-report-anchor-body-shape', message: '请选择体型' },
  ]

  for (const check of requiredChecks) {
    if (check.valid()) continue
    focusValidationTarget(check.anchor)
    uni.showToast({ title: check.message, icon: 'none' })
    return false
  }
  return true
}

async function submit() {
  if (!props.order || !draft.value || submitting.value) return
  if (!validateDraft()) return
  const currentDraft = draft.value
  const sessionToken = modalSessionToken.value
  submitting.value = true
  previewExpanded.value = false
  uni.showLoading({ title: '生成中...' })
  try {
    const response = await createOrderCareReport(Number(props.order.ID), buildOrderCareReportPayload(currentDraft))
    const imageUrl = response.data?.image_url
    if (!imageUrl) throw new Error('护理报告生成失败')
    if (modalSessionToken.value !== sessionToken || draft.value !== currentDraft || !props.visible) return
    previewUrl.value = imageUrl
    uni.showToast({ title: '护理报告已生成', icon: 'success' })
  } catch (error: any) {
    if (modalSessionToken.value !== sessionToken || !props.visible) return
    uni.showToast({ title: error?.message || '生成失败', icon: 'none' })
  } finally {
    if (modalSessionToken.value === sessionToken) {
      submitting.value = false
      uni.hideLoading()
    }
  }
}

function resetPreview() {
  previewUrl.value = ''
  previewExpanded.value = false
}

async function savePreview() {
  if (!props.order || !draft.value || !previewUrl.value || saving.value) return
  saving.value = true
  try {
    const result = await saveImageByUrl(
      previewUrl.value,
      buildOrderCareReportFileName(props.order.order_no || `NO${props.order.ID}`, draft.value.petName),
      { title: '护理报告图片' }
    )
    if (result === 'preview') {
      uni.showToast({ title: '新页面已打开，请长按图片保存', icon: 'none' })
    }
  } finally {
    saving.value = false
  }
}

function resolveAbsoluteUrl(value: string) {
  if (!value) return ''
  if (/^(https?:)?\/\//i.test(value) || value.startsWith('data:')) return value
  if (typeof window === 'undefined') return value
  return new URL(value, window.location.origin).toString()
}
</script>

<style scoped>
.care-report-mask {
  position: fixed;
  inset: 0;
  background: rgba(17, 24, 39, 0.58);
  z-index: 1200;
}

.care-report-panel {
  position: relative;
  width: 100vw;
  height: 100vh;
  background: #F5F6F8;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

.care-report-header,
.care-report-draft-preview-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 20rpx;
  padding: 28rpx;
  border-bottom: 2rpx solid #E5E7EB;
  background: #FFFFFF;
}

.care-report-copy {
  display: flex;
  flex-direction: column;
  gap: 8rpx;
}

.care-report-title {
  font-size: 34rpx;
  font-weight: 700;
  color: #111827;
}

.care-report-subtitle {
  font-size: 24rpx;
  color: #6B7280;
}

.care-report-close,
.care-report-preview-back {
  font-size: 28rpx;
  color: #4F46E5;
  line-height: 1.2;
  padding: 8rpx;
}

.care-report-form,
.care-report-select,
.care-report-preview {
  flex: 1;
  overflow-y: auto;
  box-sizing: border-box;
  -webkit-overflow-scrolling: touch;
}

.care-report-form {
  padding-bottom: 24rpx;
}

.care-report-select,
.care-report-preview {
  padding: 28rpx;
}

.care-report-select-title,
.care-report-preview-hint {
  display: block;
  margin-bottom: 20rpx;
  font-size: 28rpx;
  font-weight: 600;
  color: #1F2937;
}

.care-report-pet-list {
  display: flex;
  flex-direction: column;
  gap: 16rpx;
}

.care-report-pet-card {
  background: #FFFFFF;
  border: 2rpx solid #E5E7EB;
  border-radius: 16rpx;
  padding: 24rpx;
  display: flex;
  flex-direction: column;
  gap: 8rpx;
}

.care-report-pet-name {
  font-size: 30rpx;
  font-weight: 700;
  color: #111827;
}

.care-report-pet-meta,
.care-report-field-hint {
  font-size: 24rpx;
  color: #6B7280;
}

.care-report-switcher {
  display: flex;
  flex-wrap: wrap;
  gap: 12rpx;
  padding: 20rpx 28rpx;
  background: #FFFFFF;
  border-bottom: 2rpx solid #E5E7EB;
}

.care-report-switch-chip {
  min-height: 64rpx;
  padding: 0 20rpx;
  border-radius: 12rpx;
  border: 2rpx solid #D1D5DB;
  background: #FFFFFF;
  color: #4B5563;
  font-size: 24rpx;
  display: flex;
  align-items: center;
}

.care-report-switch-chip.active {
  background: #EEF2FF;
  border-color: #6366F1;
  color: #4338CA;
  font-weight: 600;
}

.care-report-form-section,
.care-report-inspection-section {
  padding: 28rpx;
  background: #FFFFFF;
  border-bottom: 2rpx solid #E5E7EB;
}

.care-report-section-title {
  font-size: 30rpx;
  font-weight: 700;
  color: #111827;
}

.care-report-photo-row {
  display: grid;
  grid-template-columns: 112rpx minmax(0, 1fr) auto;
  align-items: center;
  gap: 18rpx;
  margin-top: 24rpx;
  padding-bottom: 24rpx;
  border-bottom: 2rpx solid #F0F1F3;
}

.care-report-photo-preview,
.care-report-photo-placeholder {
  width: 112rpx;
  height: 112rpx;
  border-radius: 56rpx;
}

.care-report-photo-placeholder {
  display: flex;
  align-items: center;
  justify-content: center;
  background: #F3F4F6;
  border: 2rpx dashed #C7CBD1;
  color: #9CA3AF;
  font-size: 24rpx;
}

.care-report-photo-copy {
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 8rpx;
}

.care-report-photo-button {
  min-height: 72rpx;
  padding: 0 20rpx;
  border-radius: 12rpx;
  border: 2rpx solid #C7D2FE;
  color: #4338CA;
  font-size: 24rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  white-space: nowrap;
}

.care-report-form-row {
  display: grid;
  grid-template-columns: 1fr;
  gap: 22rpx;
  margin-top: 24rpx;
}

.care-report-field {
  display: flex;
  flex-direction: column;
  gap: 12rpx;
}

.care-report-field-label {
  font-size: 26rpx;
  font-weight: 600;
  color: #374151;
}

.care-report-input,
.care-report-picker,
.care-report-textarea {
  width: 100%;
  background: #FFFFFF;
  border-radius: 12rpx;
  border: 2rpx solid #D1D5DB;
  box-sizing: border-box;
  font-size: 28rpx;
  color: #111827;
}

.care-report-input {
  display: flex;
  align-items: center;
  min-height: 88rpx;
  padding: 0 22rpx;
  overflow: hidden;
}

.care-report-picker {
  min-height: 88rpx;
  padding: 0 22rpx;
  display: flex;
  align-items: center;
}

.care-report-textarea {
  min-height: 128rpx;
  padding: 20rpx 22rpx;
  line-height: 40rpx;
}

.care-report-input :deep(.uni-input-wrapper) {
  width: 100%;
  min-height: 100%;
  height: 100%;
  display: flex;
  align-items: center;
}

.care-report-input :deep(.uni-input-input) {
  width: 100%;
  height: 44rpx;
  line-height: 44rpx;
  font-size: 28rpx;
  color: #111827;
  text-align: left !important;
}

.care-report-input :deep(.uni-input-placeholder) {
  width: 100%;
  line-height: 44rpx;
  font-size: 28rpx;
  color: #9CA3AF;
  text-align: left !important;
}

.care-report-segmented {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 12rpx;
}

.care-report-segment,
.care-report-choice {
  min-height: 88rpx;
  border-radius: 12rpx;
  border: 2rpx solid #D1D5DB;
  background: #FFFFFF;
  color: #4B5563;
  display: flex;
  align-items: center;
  justify-content: center;
  text-align: center;
  box-sizing: border-box;
}

.care-report-segment {
  font-size: 28rpx;
  font-weight: 600;
}

.care-report-segment.active,
.care-report-choice.active {
  border-color: #6366F1;
  background: #EEF2FF;
  color: #4338CA;
  font-weight: 700;
}

.care-report-choice-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 12rpx;
  margin-top: 22rpx;
}

.care-report-choice {
  padding: 12rpx;
  font-size: 25rpx;
  line-height: 34rpx;
  word-break: break-word;
}

.care-report-inspection-list {
  margin-top: 20rpx;
  border-top: 2rpx solid #E5E7EB;
}

.care-report-inspection-section {
  padding-top: 0;
  padding-bottom: 0;
}

.care-report-section-head {
  min-height: 100rpx;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 20rpx;
}

.care-report-section-summary {
  flex-shrink: 0;
  display: flex;
  align-items: center;
  gap: 16rpx;
  font-size: 24rpx;
  color: #6B7280;
}

.care-report-section-toggle {
  width: 48rpx;
  height: 48rpx;
  border-radius: 24rpx;
  background: #F3F4F6;
  color: #4B5563;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 32rpx;
  line-height: 1;
}

.care-report-section-body {
  padding: 0 0 28rpx;
}

.care-report-section-body .care-report-choice-grid {
  margin-top: 0;
}

.care-report-section-note {
  margin-top: 20rpx;
}

.care-report-preview-image {
  width: 100%;
  border-radius: 16rpx;
  background: #FFFFFF;
}

.care-report-actions {
  flex-shrink: 0;
  display: flex;
  gap: 16rpx;
  padding: 20rpx 28rpx calc(20rpx + env(safe-area-inset-bottom));
  border-top: 2rpx solid #E5E7EB;
  background: #FFFFFF;
}

.care-report-actions.generated {
  flex-wrap: wrap;
}

.care-report-btn {
  flex: 1;
  min-width: 0;
  height: 88rpx;
  border-radius: 12rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 28rpx;
  font-weight: 600;
  box-sizing: border-box;
}

.care-report-btn.primary {
  background: #4F46E5;
  color: #FFFFFF;
}

.care-report-btn.secondary {
  background: #EEF2FF;
  color: #4338CA;
  border: 2rpx solid #C7D2FE;
}

.care-report-btn.ghost {
  background: #FFFFFF;
  color: #4B5563;
  border: 2rpx solid #D1D5DB;
}

.care-report-draft-preview {
  position: absolute;
  inset: 0;
  z-index: 4;
  display: flex;
  flex-direction: column;
  background: #F5F6F8;
}

.care-report-draft-preview-scroll {
  flex: 1;
  overflow-y: auto;
  padding: 20rpx;
  -webkit-overflow-scrolling: touch;
}
</style>
