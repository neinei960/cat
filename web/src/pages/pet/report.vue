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
          @click="previewReport(index)"
          @longpress="openReportActions(report)"
        >
          <picker mode="date" :value="getEditableBathDate(report)" @change="onBathDateChange(report, $event)" @click.stop>
            <view class="report-date editable-date">
              <text>洗浴日期：{{ formatBathDate(report) }}</text>
              <text class="report-date-arrow">编辑 ›</text>
            </view>
          </picker>
          <image :src="report.image_url" class="report-image" mode="aspectFill" />
          <text class="report-hint">点击查看大图</text>
        </view>
      </view>
    </view>
  </SideLayout>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { onLoad, onShow } from '@dcloudio/uni-app'
import SideLayout from '@/components/SideLayout.vue'
import { getPet } from '@/api/pet'
import { uploadFile } from '@/api/upload'
import { createPetBathReport, deletePetBathReport, getPetBathReports, updatePetBathReport, type PetBathReport } from '@/api/pet-bath-report'

const petId = ref(0)
const petName = ref('')
const loading = ref(false)
const reports = ref<PetBathReport[]>([])
const uploading = ref(false)

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

async function onBathDateChange(report: PetBathReport, e: any) {
  const bathDate = e?.detail?.value
  if (!bathDate) return
  try {
    await updatePetBathReport(petId.value, report.ID, bathDate)
    report.bath_date = bathDate
    reports.value = [...reports.value].sort((a, b) => getEditableBathDate(b).localeCompare(getEditableBathDate(a)))
    uni.showToast({ title: '已更新洗浴日期', icon: 'success' })
  } catch {
    uni.showToast({ title: '更新失败', icon: 'none' })
  }
}

function openReportActions(report: PetBathReport) {
  uni.showActionSheet({
    itemList: ['删除当前报告'],
    success: async (res) => {
      if (res.tapIndex !== 0) return
      try {
        await deletePetBathReport(petId.value, report.ID)
        uni.showToast({ title: '已删除', icon: 'success' })
        await loadReports()
      } catch {
        uni.showToast({ title: '删除失败', icon: 'none' })
      }
    },
  })
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
  padding: 22rpx;
  box-shadow: 0 12rpx 30rpx rgba(15, 23, 42, 0.06);
}

.report-date {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 16rpx;
  font-size: 26rpx;
  font-weight: 700;
  color: #111827;
}

.editable-date {
  gap: 12rpx;
}

.report-date-arrow {
  flex-shrink: 0;
  font-size: 22rpx;
  color: #4f46e5;
  font-weight: 600;
}

.report-image {
  width: 100%;
  height: 360rpx;
  border-radius: 16rpx;
  background: #eef2ff;
}

.report-hint {
  display: block;
  margin-top: 12rpx;
  font-size: 22rpx;
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
</style>
