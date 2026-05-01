<template>
  <view v-if="visible" class="care-report-mask" @click="onMaskClick">
    <view class="care-report-panel" @click.stop>
      <view class="care-report-header">
        <view class="care-report-copy">
          <text class="care-report-title">生成护理报告</text>
          <text class="care-report-subtitle">{{ previewUrl ? '报告已生成，可直接保存图片' : '点击底图直接填写与勾选' }}</text>
        </view>
        <text class="care-report-close" @click="handleClose">✕</text>
      </view>

      <template v-if="previewUrl">
        <view class="care-report-preview">
          <text class="care-report-preview-hint">护理报告生成完成</text>
          <image :src="previewImageUrl" mode="widthFix" class="care-report-preview-image" show-menu-by-longpress />
        </view>
        <view class="care-report-actions">
          <view class="care-report-btn primary" @click="savePreview">{{ saving ? '处理中...' : '保存图片' }}</view>
          <view class="care-report-btn ghost" @click="resetPreview">重新填写</view>
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

          <view class="care-report-stage-card editable">
            <OrderCareReportStage
              ref="stageRef"
              :draft="draft"
              editable
              :active-editor-key="activeEditorMarker"
              @edit-target="openEditor"
              @toggle-body-shape="toggleBodyShape"
              @toggle-section-check="handleStageSectionToggle"
            />
          </view>
        </view>

        <view v-if="editorDescriptor" class="care-report-editor-dock">
          <view class="care-report-editor-sheet">
            <view class="care-report-editor-head">
              <text class="care-report-editor-title">{{ editorDescriptor.title }}</text>
              <text class="care-report-editor-close" @click="closeEditor">收起</text>
            </view>

            <template v-if="editorDescriptor.kind === 'choice'">
              <view class="care-report-editor-options">
                <view
                  v-for="option in editorDescriptor.options"
                  :key="option.value"
                  :class="['care-report-editor-chip', getDraftFieldValue(editorDescriptor.key) === option.value ? 'active' : '']"
                  @click="updateFieldValue(editorDescriptor.key, option.value)"
                >
                  {{ option.label }}
                </view>
              </view>
            </template>

            <template v-else-if="editorDescriptor.kind === 'date'">
              <picker mode="date" :value="formatPickerDate(getDraftFieldValue(editorDescriptor.key))" @change="onEditorDateChange(editorDescriptor.key, $event)">
                <view class="care-report-picker">{{ getDraftFieldValue(editorDescriptor.key) || '请选择日期' }}</view>
              </picker>
            </template>

            <template v-else-if="editorDescriptor.kind === 'note'">
              <textarea
                :value="getSectionNote(editorDescriptor.sectionKey)"
                class="care-report-textarea care-report-editor-textarea"
                maxlength="120"
                auto-height
                placeholder="备注（选填）"
                @input="updateSectionNote(editorDescriptor.sectionKey, $event)"
              />
            </template>

            <template v-else-if="editorDescriptor.kind === 'textarea'">
              <textarea
                :value="getDraftFieldValue(editorDescriptor.key)"
                class="care-report-textarea care-report-editor-textarea"
                :maxlength="editorDescriptor.maxLength"
                auto-height
                :placeholder="editorDescriptor.placeholder"
                @input="onEditorInput(editorDescriptor.key, $event)"
              />
            </template>

            <template v-else>
              <input
                :value="getDraftFieldValue(editorDescriptor.key)"
                class="care-report-input"
                :maxlength="editorDescriptor.maxLength"
                :placeholder="editorDescriptor.placeholder"
                @input="onEditorInput(editorDescriptor.key, $event)"
              />
            </template>
          </view>
        </view>

        <view class="care-report-actions">
          <view class="care-report-btn primary" @click="submit">{{ submitting ? '生成中...' : '生成报告' }}</view>
          <view class="care-report-btn ghost" @click="handleClose">关闭</view>
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
import { createPetBathReport } from '@/api/pet-bath-report'
import { uploadH5File } from '@/api/upload'
import {
  buildOrderCareReportDraft,
  buildOrderCareReportPetOptions,
  normalizeOrderCareReportDate,
  orderCareReportSectionDefinitions,
  type OrderCareReportDraft,
  type OrderCareReportPetOption,
  type OrderCareReportSectionDefinition,
  type OrderCareReportSectionKey,
} from '@/utils/order-care-report'
import { createCropperPreviewUrl } from '@/utils/image-cropper'
import { buildOrderCareReportFileName, saveImageByUrl } from '@/utils/web-image-save'

type CareReportStageExpose = {
  exportPngBlob: () => Promise<Blob>
}

type EditableFieldKey = 'petName' | 'breed' | 'gender' | 'age' | 'careContent' | 'careDate' | 'nextCareDate' | 'weight'
type StageEditTarget =
  | { type: 'field'; key: EditableFieldKey }
  | { type: 'note'; sectionKey: OrderCareReportSectionKey }
  | { type: 'portrait' }
type ActiveEditor =
  | { kind: 'field'; key: EditableFieldKey }
  | { kind: 'note'; sectionKey: OrderCareReportSectionKey }
  | null

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
const submitting = ref(false)
const saving = ref(false)
const uploadingPortrait = ref(false)
const cropperVisible = ref(false)
const cropperSrc = ref('')
const modalSessionToken = ref(0)
const stageRef = ref<CareReportStageExpose | null>(null)
const activeEditor = ref<ActiveEditor>(null)

const sectionDefinitions: OrderCareReportSectionDefinition[] = orderCareReportSectionDefinitions

const petOptions = computed<OrderCareReportPetOption[]>(() => {
  if (!props.order) return []
  return buildOrderCareReportPetOptions(props.order)
})

const previewImageUrl = computed(() => resolveAbsoluteUrl(previewUrl.value))

const activeEditorMarker = computed(() => {
  if (!activeEditor.value) return ''
  if (activeEditor.value.kind === 'field') {
    return `field:${activeEditor.value.key}`
  }
  return `note:${activeEditor.value.sectionKey}`
})

const editorDescriptor = computed(() => {
  if (!activeEditor.value || !draft.value) return null

  if (activeEditor.value.kind === 'note') {
    const section = sectionDefinitions.find((item) => item.key === activeEditor.value?.sectionKey)
    return {
      kind: 'note' as const,
      marker: `note:${activeEditor.value.sectionKey}`,
      title: `${section?.label || '当前项'}备注`,
      sectionKey: activeEditor.value.sectionKey,
    }
  }

  const fieldKey = activeEditor.value.key
  if (fieldKey === 'careDate' || fieldKey === 'nextCareDate') {
    return {
      kind: 'date' as const,
      marker: `field:${fieldKey}`,
      key: fieldKey,
      title: fieldKey === 'careDate' ? '护理日期' : '建议下次护理日期',
    }
  }

  if (fieldKey === 'gender') {
    return {
      kind: 'choice' as const,
      marker: 'field:gender',
      key: fieldKey,
      title: '性别',
      options: [
        { label: 'GG', value: 'GG' },
        { label: 'MM', value: 'MM' },
      ],
    }
  }

  if (fieldKey === 'careContent') {
    return {
      kind: 'textarea' as const,
      marker: 'field:careContent',
      key: fieldKey,
      title: '护理内容',
      placeholder: '例如 Harmurry精致皮毛调理',
      maxLength: 60,
    }
  }

  const fieldConfigs: Record<Exclude<EditableFieldKey, 'gender' | 'careDate' | 'nextCareDate' | 'careContent'>, { title: string; placeholder: string; maxLength: number }> = {
    petName: { title: '宝贝名字', placeholder: '填写宝贝名字', maxLength: 20 },
    breed: { title: '我的品种', placeholder: '填写品种', maxLength: 30 },
    age: { title: '年龄', placeholder: '例如 2岁1月', maxLength: 20 },
    weight: { title: '体重', placeholder: '例如 5.55 KG', maxLength: 20 },
  }

  return {
    kind: 'text' as const,
    marker: `field:${fieldKey}`,
    key: fieldKey as keyof typeof fieldConfigs,
    title: fieldConfigs[fieldKey as keyof typeof fieldConfigs].title,
    placeholder: fieldConfigs[fieldKey as keyof typeof fieldConfigs].placeholder,
    maxLength: fieldConfigs[fieldKey as keyof typeof fieldConfigs].maxLength,
  }
})

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
  cropperVisible.value = false
  clearCropperSrc()
  activeEditor.value = null
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
  cropperVisible.value = false
  clearCropperSrc()
  activeEditor.value = null
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
  activeEditor.value = { kind: 'field', key: 'careContent' }
}

function handleClose() {
  emit('close')
}

function onMaskClick() {
  if (cropperVisible.value) return
  handleClose()
}

function closeEditor() {
  activeEditor.value = null
}

function formatPickerDate(value: string) {
  return String(value || '').replace(/\./g, '-')
}

function getSectionValue(sectionKey: OrderCareReportSectionKey) {
  return draft.value?.[sectionKey]
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

function handleStageSectionToggle(payload: { sectionKey: OrderCareReportSectionKey; value: string }) {
  toggleSectionCheck(payload.sectionKey, payload.value)
}

function toggleBodyShape(value: string) {
  if (!draft.value) return
  draft.value.bodyShape = draft.value.bodyShape === value ? '' : value
}

function openEditor(target: StageEditTarget) {
  if (target.type === 'portrait') {
    activeEditor.value = null
    choosePortrait()
    return
  }
  if (target.type === 'field') {
    activeEditor.value = { kind: 'field', key: target.key }
    return
  }
  activeEditor.value = { kind: 'note', sectionKey: target.sectionKey }
}

function getDraftFieldValue(key: EditableFieldKey) {
  return draft.value?.[key] || ''
}

function updateFieldValue(key: EditableFieldKey, value: string) {
  if (!draft.value) return
  draft.value[key] = value
}

function onEditorInput(key: EditableFieldKey, event: any) {
  updateFieldValue(key, String(event?.detail?.value || ''))
}

function onEditorDateChange(field: 'careDate' | 'nextCareDate', event: any) {
  updateFieldValue(field, String(event?.detail?.value || '').replace(/-/g, '.'))
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

function validateDraft() {
  if (!draft.value) {
    uni.showToast({ title: '请先选择猫咪', icon: 'none' })
    return false
  }
  if (!draft.value.portraitUrl) {
    uni.showToast({ title: '请先上传护理照片', icon: 'none' })
    return false
  }
  if (!draft.value.careContent.trim()) {
    uni.showToast({ title: '请填写护理内容', icon: 'none' })
    return false
  }
  if (!draft.value.weight.trim()) {
    uni.showToast({ title: '请填写体重', icon: 'none' })
    return false
  }
  if (!draft.value.careDate) {
    uni.showToast({ title: '请填写护理日期', icon: 'none' })
    return false
  }
  if (!draft.value.nextCareDate) {
    uni.showToast({ title: '请填写建议下次护理日期', icon: 'none' })
    return false
  }
  if (!draft.value.bodyShape) {
    uni.showToast({ title: '请选择体型', icon: 'none' })
    return false
  }
  return true
}

async function submit() {
  if (!props.order || !draft.value || submitting.value) return
  if (!validateDraft()) return
  const currentDraft = draft.value
  const sessionToken = modalSessionToken.value
  if (typeof File === 'undefined') {
    uni.showToast({ title: '当前环境暂不支持导出图片', icon: 'none' })
    return
  }
  submitting.value = true
  uni.showLoading({ title: '生成中...' })
  try {
    await nextTick()
    const stageBlob = await stageRef.value?.exportPngBlob()
    if (!stageBlob) {
      throw new Error('护理报告导出失败')
    }
    const fileName = buildOrderCareReportFileName(props.order.order_no || `NO${props.order.ID}`, draft.value.petName)
    const file = new File([stageBlob], fileName, { type: 'image/png' })
    const imageUrl = await uploadH5File(file, { preserveOriginal: true })
    await createPetBathReport(draft.value.petId, imageUrl, normalizeOrderCareReportDate(draft.value.careDate) || undefined)
    if (modalSessionToken.value !== sessionToken || draft.value !== currentDraft || !props.visible) return
    previewUrl.value = imageUrl
    activeEditor.value = null
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
  if (draft.value) {
    activeEditor.value = { kind: 'field', key: 'careContent' }
  }
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
  background: rgba(17, 17, 17, 0.58);
  z-index: 1200;
}

.care-report-panel {
  width: 100vw;
  height: 100vh;
  background: #FFF9F0;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

.care-report-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 20rpx;
  padding: 28rpx 28rpx 22rpx;
  border-bottom: 2rpx solid rgba(210, 180, 120, 0.28);
}

.care-report-copy {
  display: flex;
  flex-direction: column;
  gap: 8rpx;
}

.care-report-title {
  font-size: 34rpx;
  font-weight: 700;
  color: #2F2417;
}

.care-report-subtitle {
  font-size: 24rpx;
  color: #7B6242;
}

.care-report-close {
  font-size: 36rpx;
  color: #8A6B2F;
  line-height: 1;
  padding: 4rpx;
}

.care-report-form,
.care-report-select,
.care-report-preview {
  flex: 1;
  overflow-y: auto;
  padding: 28rpx;
  box-sizing: border-box;
}

.care-report-select-title,
.care-report-preview-hint {
  font-size: 28rpx;
  font-weight: 600;
  color: #3D2D16;
  margin-bottom: 20rpx;
  display: block;
}

.care-report-pet-list {
  display: flex;
  flex-direction: column;
  gap: 16rpx;
}

.care-report-pet-card {
  background: #FFFFFF;
  border: 2rpx solid #E9D6AE;
  border-radius: 22rpx;
  padding: 24rpx;
  display: flex;
  flex-direction: column;
  gap: 8rpx;
}

.care-report-pet-name {
  font-size: 30rpx;
  font-weight: 700;
  color: #2F2417;
}

.care-report-pet-meta {
  font-size: 24rpx;
  color: #8A7452;
}

.care-report-switcher {
  display: flex;
  flex-wrap: wrap;
  gap: 12rpx;
  margin-bottom: 20rpx;
}

.care-report-switch-chip {
  padding: 12rpx 20rpx;
  border-radius: 999rpx;
  border: 2rpx solid #E3D0A4;
  background: #FFFDF8;
  color: #7A603A;
  font-size: 24rpx;
  line-height: 1.2;
}

.care-report-switch-chip.active {
  background: #F7E5B4;
  border-color: #D7AF51;
  color: #5E4617;
  font-weight: 600;
}

.care-report-stage-card,
.care-report-editor-sheet {
  background: #FFFFFF;
  border-radius: 24rpx;
  border: 2rpx solid rgba(220, 198, 151, 0.35);
  box-sizing: border-box;
}

.care-report-stage-card {
  padding: 16rpx;
}

.care-report-stage-card.editable {
  box-shadow: 0 10rpx 30rpx rgba(64, 42, 10, 0.06);
}

.care-report-editor-dock {
  padding: 0 28rpx 20rpx;
  border-top: 2rpx solid rgba(210, 180, 120, 0.14);
  background: #FFF9F0;
}

.care-report-editor-sheet {
  padding: 22rpx;
  display: flex;
  flex-direction: column;
  gap: 16rpx;
}

.care-report-editor-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 20rpx;
}

.care-report-editor-title {
  font-size: 28rpx;
  font-weight: 700;
  color: #2F2417;
}

.care-report-editor-close {
  font-size: 24rpx;
  color: #8A6B2F;
}

.care-report-editor-options {
  display: flex;
  flex-wrap: wrap;
  gap: 12rpx;
}

.care-report-editor-chip {
  padding: 14rpx 24rpx;
  border-radius: 999rpx;
  border: 2rpx solid #E3D0A4;
  background: #FFFDF8;
  color: #7A603A;
  font-size: 24rpx;
  line-height: 1.2;
}

.care-report-editor-chip.active {
  background: #F7E5B4;
  border-color: #D7AF51;
  color: #5E4617;
  font-weight: 600;
}

.care-report-input,
.care-report-picker,
.care-report-textarea {
  width: 100%;
  background: #FFFCF6;
  border-radius: 16rpx;
  border: 2rpx solid #F0E0BA;
  box-sizing: border-box;
  font-size: 26rpx;
  color: #2F2417;
}

.care-report-input {
  display: flex;
  align-items: center;
  min-height: 76rpx;
  padding: 0 20rpx;
  overflow: hidden;
}

.care-report-picker,
.care-report-textarea {
  padding: 18rpx 20rpx;
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
  height: 40rpx;
  line-height: 40rpx;
  font-size: 26rpx;
  color: #2F2417;
  text-align: left !important;
}

.care-report-input :deep(.uni-input-placeholder) {
  width: 100%;
  line-height: 40rpx;
  font-size: 26rpx;
  color: #B79C6B;
  text-align: left !important;
}

.care-report-picker {
  min-height: 64rpx;
  display: flex;
  align-items: center;
}

.care-report-textarea {
  min-height: 112rpx;
}

.care-report-editor-textarea {
  margin-top: 0;
}

.care-report-preview-image {
  width: 100%;
  border-radius: 20rpx;
  background: #FFFFFF;
  box-shadow: 0 8rpx 30rpx rgba(24, 16, 7, 0.14);
}

.care-report-actions {
  display: flex;
  gap: 16rpx;
  padding: 22rpx 28rpx 28rpx;
  border-top: 2rpx solid rgba(210, 180, 120, 0.24);
}

.care-report-btn {
  flex: 1;
  height: 84rpx;
  border-radius: 20rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 28rpx;
  font-weight: 600;
}

.care-report-btn.primary {
  background: linear-gradient(135deg, #E8C86E, #D7A843);
  color: #5E4617;
}

.care-report-btn.ghost {
  background: #FFF3D8;
  color: #8A6B2F;
  border: 2rpx solid #E7CB82;
}
</style>
