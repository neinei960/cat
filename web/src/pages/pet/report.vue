<template>
  <SideLayout>
    <view class="page">
      <view class="hero">
        <view>
          <text class="title">洗浴报告记录</text>
          <text class="subtitle">{{ petName || '当前猫咪' }}的历史洗浴报告与上传记录</text>
        </view>
        <view class="upload-button" @click="chooseReportImage">上传报告</view>
      </view>

      <view v-if="loading" class="empty-state">加载中...</view>
      <view v-else-if="reports.length === 0" class="empty-state">
        <text class="empty-title">还没有洗浴报告</text>
        <text class="empty-desc">上传后会记录洗浴日期，点击日期可编辑，点击图片可放大查看</text>
      </view>

      <view v-else class="report-grid">
        <view
          v-for="(report, index) in reports"
          :key="report.ID"
          class="report-card"
          :class="{ 'report-card-active': draggingReportId === report.ID }"
          :data-report-id="report.ID"
          @click="onReportClick(index)"
          @longpress="openReportActions(report)"
          @touchstart="startCardLongPress(report)"
          @touchend="clearCardLongPress"
          @touchcancel="clearCardLongPress"
          @touchmove="clearCardLongPress"
        >
          <view class="report-toolbar">
            <picker mode="date" :value="getEditableBathDate(report)" @change="onBathDateChange(report, $event)" @click.stop>
              <view class="report-date editable-date">
                <view class="report-date-copy">
                  <text class="report-date-label">洗浴日期</text>
                  <text class="report-date-main">{{ formatBathDate(report) }}</text>
                </view>
                <text class="report-date-arrow">编辑</text>
              </view>
            </picker>
            <view class="report-tools">
              <view
                v-if="canDragSort"
                class="drag-handle"
                @click.stop
                @touchstart.stop.prevent="beginDrag(index, $event)"
                @mousedown.stop.prevent="beginDrag(index, $event)"
              >
                <text class="drag-handle-icon">↕</text>
                <text>拖拽换位</text>
              </view>
              <view
                v-if="isDesktopInteraction"
                class="report-delete-btn"
                @click.stop="deleteReport(report)"
              >删除</view>
            </view>
          </view>
          <image :src="report.image_url" class="report-image" mode="aspectFill" />
          <text class="report-hint">点击查看大图</text>
        </view>
      </view>
    </view>
  </SideLayout>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, ref } from 'vue'
import { onLoad, onShow } from '@dcloudio/uni-app'
import SideLayout from '@/components/SideLayout.vue'
import { getPet } from '@/api/pet'
import { uploadFile } from '@/api/upload'
import { createPetBathReport, deletePetBathReport, getPetBathReports, reorderPetBathReports, updatePetBathReport, type PetBathReport } from '@/api/pet-bath-report'
import { useDesktopInteraction } from '@/utils/interaction'

const petId = ref(0)
const petName = ref('')
const loading = ref(false)
const reports = ref<PetBathReport[]>([])
const uploading = ref(false)
const savingOrder = ref(false)
const draggingReportId = ref<number | null>(null)
let dragMoved = false
let dragSnapshot: PetBathReport[] = []
let longPressTimer: ReturnType<typeof setTimeout> | null = null
let longPressTriggered = false
const { isDesktopInteraction } = useDesktopInteraction()

const canDragSort = computed(() => reports.value.length > 1 && !savingOrder.value)
const previewUrls = computed(() => reports.value.map(item => item.image_url))

onLoad((query) => {
  if (query?.id) {
    petId.value = parseInt(query.id)
  }
})

onShow(() => {
  if (!petId.value) return
  void Promise.all([loadPet(), loadReports()])
})

onBeforeUnmount(() => {
  removeDragListeners()
  clearCardLongPress()
})

async function loadPet() {
  if (!petId.value) return
  try {
    const res = await getPet(petId.value)
    petName.value = res.data?.name || ''
  } catch {}
}

async function loadReports() {
  if (!petId.value) return
  loading.value = true
  try {
    const res = await getPetBathReports(petId.value)
    reports.value = Array.isArray(res.data) ? res.data : []
  } finally {
    loading.value = false
  }
}

function formatBathDate(report: PetBathReport) {
  return formatDateText(report.bath_date || report.CreatedAt)
}

function getEditableBathDate(report: PetBathReport) {
  return normalizeDateText(report.bath_date || report.CreatedAt)
}

function formatDateText(value?: string) {
  if (!value) return '-'
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return value.slice(0, 10)
  return `${date.getFullYear()}-${date.getMonth() + 1}-${date.getDate()}`
}

function normalizeDateText(value?: string) {
  if (!value) return new Date().toISOString().slice(0, 10)
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return value.slice(0, 10)
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  return `${date.getFullYear()}-${month}-${day}`
}

function previewReport(index: number) {
  if (!previewUrls.value.length) return
  uni.previewImage({
    urls: previewUrls.value,
    current: previewUrls.value[index],
  })
}

function onReportClick(index: number) {
  if (longPressTriggered) {
    longPressTriggered = false
    return
  }
  if (draggingReportId.value != null) return
  previewReport(index)
}

async function onBathDateChange(report: PetBathReport, e: any) {
  const bathDate = e?.detail?.value
  if (!bathDate) return
  try {
    await updatePetBathReport(petId.value, report.ID, bathDate)
    report.bath_date = bathDate
    uni.showToast({ title: '已更新洗浴日期', icon: 'success' })
  } catch {
    uni.showToast({ title: '更新失败', icon: 'none' })
  }
}

function openReportActions(report: PetBathReport) {
  longPressTriggered = true
  uni.showModal({
    title: '删除洗浴报告',
    content: `确认删除 ${formatBathDate(report)} 的报告？`,
    confirmColor: '#DC2626',
    success: async (res) => {
      if (!res.confirm) return
      await deleteReport(report)
    },
  })
}

function startCardLongPress(report: PetBathReport) {
  if (draggingReportId.value != null) return
  clearCardLongPress()
  longPressTriggered = false
  longPressTimer = setTimeout(() => {
    longPressTimer = null
    openReportActions(report)
  }, 500)
}

function clearCardLongPress() {
  if (!longPressTimer) return
  clearTimeout(longPressTimer)
  longPressTimer = null
}

async function deleteReport(report: PetBathReport) {
  try {
    await deletePetBathReport(petId.value, report.ID)
    uni.showToast({ title: '已删除', icon: 'success' })
    await loadReports()
  } catch {
    uni.showToast({ title: '删除失败', icon: 'none' })
  }
}

async function chooseReportImage() {
  if (!petId.value || uploading.value) return
  uni.chooseImage({
    count: 1,
    sizeType: ['compressed'],
    sourceType: ['album', 'camera'],
    success: async (res) => {
      const rawPath = res.tempFilePaths?.[0]
      if (!rawPath) return

      uploading.value = true
      uni.showLoading({ title: '上传中...' })
      try {
        const uploadPath = await compressImageIfPossible(rawPath)
        const url = await uploadFile(uploadPath)
        await createPetBathReport(petId.value, url, new Date().toISOString().slice(0, 10))
        uni.showToast({ title: '上传成功', icon: 'success' })
        await loadReports()
      } catch {
        uni.showToast({ title: '上传失败', icon: 'none' })
      } finally {
        uploading.value = false
        uni.hideLoading()
      }
    },
  })
}

async function compressImageIfPossible(filePath: string) {
  try {
    const result = await uni.compressImage({
      src: filePath,
      quality: 72,
      compressedWidth: 1400,
    })
    return result.tempFilePath || filePath
  } catch {
    return filePath
  }
}

function swapReports(list: PetBathReport[], fromIndex: number, toIndex: number) {
  const next = [...list]
  ;[next[fromIndex], next[toIndex]] = [next[toIndex], next[fromIndex]]
  return next
}

function getEventPoint(event: any) {
  const touch = event?.touches?.[0] || event?.changedTouches?.[0]
  if (touch) {
    return { x: touch.clientX, y: touch.clientY }
  }
  if (typeof event?.clientX === 'number' && typeof event?.clientY === 'number') {
    return { x: event.clientX, y: event.clientY }
  }
  return null
}

function removeDragListeners() {
  if (typeof window === 'undefined') return
  window.removeEventListener('touchmove', handleDragMove as EventListener)
  window.removeEventListener('touchend', handleDragEnd as EventListener)
  window.removeEventListener('touchcancel', handleDragEnd as EventListener)
  window.removeEventListener('mousemove', handleDragMove as EventListener)
  window.removeEventListener('mouseup', handleDragEnd as EventListener)
  document.body.style.userSelect = ''
}

async function saveReportOrder(list: PetBathReport[]) {
  if (!petId.value) return
  savingOrder.value = true
  try {
    await reorderPetBathReports(
      petId.value,
      list.map(item => item.ID),
    )
    list.forEach((item, index) => {
      item.sort_order = list.length - index
    })
    uni.showToast({ title: '顺序已更新', icon: 'success' })
  } finally {
    savingOrder.value = false
  }
}

function beginDrag(index: number, event: any) {
  if (typeof window === 'undefined' || savingOrder.value || !reports.value[index] || reports.value.length < 2) return
  draggingReportId.value = reports.value[index].ID
  dragSnapshot = [...reports.value]
  dragMoved = false
  document.body.style.userSelect = 'none'
  window.addEventListener('touchmove', handleDragMove as EventListener, { passive: false })
  window.addEventListener('touchend', handleDragEnd as EventListener)
  window.addEventListener('touchcancel', handleDragEnd as EventListener)
  window.addEventListener('mousemove', handleDragMove as EventListener)
  window.addEventListener('mouseup', handleDragEnd as EventListener)
  handleDragMove(event)
}

function handleDragMove(event: Event) {
  if (draggingReportId.value == null || typeof document === 'undefined') return
  const point = getEventPoint(event)
  if (!point) return
  if ('preventDefault' in event) {
    event.preventDefault()
  }
  const element = document.elementFromPoint(point.x, point.y) as HTMLElement | null
  const card = element?.closest('.report-card') as HTMLElement | null
  const targetId = Number(card?.dataset?.reportId || 0)
  if (!targetId || targetId === draggingReportId.value) return
  const fromIndex = reports.value.findIndex(item => item.ID === draggingReportId.value)
  const targetIndex = reports.value.findIndex(item => item.ID === targetId)
  if (fromIndex < 0 || targetIndex < 0 || fromIndex === targetIndex) return
  reports.value = swapReports(reports.value, fromIndex, targetIndex)
  dragMoved = true
}

async function handleDragEnd() {
  const activeId = draggingReportId.value
  removeDragListeners()
  draggingReportId.value = null
  if (!activeId || !dragMoved) return
  try {
    await saveReportOrder(reports.value)
  } catch {
    reports.value = dragSnapshot
    uni.showToast({ title: '保存顺序失败', icon: 'none' })
  } finally {
    dragSnapshot = []
    dragMoved = false
  }
}
</script>

<style scoped>
.page {
  min-height: 100vh;
  padding: 24rpx;
  background: linear-gradient(180deg, #f7f6ff 0%, #f5f6fa 28%, #f5f6fa 100%);
}

.hero {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 20rpx;
  margin-bottom: 24rpx;
  padding: 28rpx;
  background: #fff;
  border-radius: 20rpx;
  box-shadow: 0 14rpx 36rpx rgba(79, 70, 229, 0.08);
}

.title {
  display: block;
  font-size: 40rpx;
  font-weight: 700;
  color: #111827;
}

.subtitle {
  display: block;
  margin-top: 10rpx;
  font-size: 24rpx;
  color: #6b7280;
}

.upload-button {
  flex-shrink: 0;
  padding: 18rpx 26rpx;
  border-radius: 999rpx;
  background: linear-gradient(135deg, #5b55ff 0%, #7c73ff 100%);
  color: #fff;
  font-size: 26rpx;
  font-weight: 600;
}

.report-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 20rpx;
}

.report-card {
  background: #fff;
  border-radius: 20rpx;
  padding: 18rpx;
  box-shadow: 0 12rpx 30rpx rgba(15, 23, 42, 0.06);
  transition: transform 0.18s ease, box-shadow 0.18s ease;
}

.report-card-active {
  transform: scale(1.02);
  box-shadow: 0 20rpx 44rpx rgba(79, 70, 229, 0.18);
}

.report-toolbar {
  display: flex;
  flex-direction: column;
  gap: 10rpx;
}
.report-tools {
  display: flex;
  align-items: center;
  gap: 10rpx;
  flex-wrap: wrap;
}

.report-date {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 10rpx;
  margin-bottom: 12rpx;
  font-weight: 700;
  color: #111827;
}

.editable-date {
  min-width: 0;
  padding: 12rpx 14rpx;
  border-radius: 16rpx;
  background: #f8f8ff;
  border: 1rpx solid #ececff;
}

.report-date-copy {
  min-width: 0;
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 4rpx;
}

.report-date-label {
  font-size: 20rpx;
  font-weight: 600;
  color: #6b7280;
}

.report-date-main {
  font-size: 24rpx;
  line-height: 1.3;
  white-space: nowrap;
}

.report-date-arrow {
  flex-shrink: 0;
  font-size: 21rpx;
  color: #4f46e5;
  font-weight: 600;
}

.drag-handle {
  display: inline-flex;
  align-items: center;
  gap: 6rpx;
  flex-shrink: 0;
  align-self: flex-start;
  padding: 8rpx 12rpx;
  border-radius: 999rpx;
  background: #eef2ff;
  color: #4f46e5;
  font-size: 20rpx;
  font-weight: 600;
  cursor: grab;
  user-select: none;
  touch-action: none;
}

.drag-handle:active {
  cursor: grabbing;
}

.drag-handle-icon {
  font-size: 24rpx;
}

.report-delete-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  padding: 8rpx 16rpx;
  border-radius: 999rpx;
  background: #FEF2F2;
  color: #DC2626;
  font-size: 20rpx;
  font-weight: 700;
}

.report-image {
  width: 100%;
  height: 280rpx;
  border-radius: 16rpx;
  background: #eef2ff;
}

.report-hint {
  display: block;
  margin-top: 10rpx;
  font-size: 20rpx;
  color: #6b7280;
}

.empty-state {
  display: flex;
  min-height: 480rpx;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 48rpx;
  background: #fff;
  border-radius: 24rpx;
  color: #6b7280;
  text-align: center;
}

.empty-title {
  font-size: 34rpx;
  font-weight: 700;
  color: #111827;
}

.empty-desc {
  margin-top: 14rpx;
  font-size: 24rpx;
}

@media (max-width: 380px) {
  .page {
    padding: 18rpx;
  }

  .report-grid {
    gap: 14rpx;
  }

  .report-card {
    padding: 14rpx;
  }

  .report-date-main {
    font-size: 22rpx;
  }

  .report-image {
    height: 240rpx;
  }
}

</style>
