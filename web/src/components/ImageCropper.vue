<template>
  <view v-if="visible" class="cropper-mask">
    <view class="cropper-topbar">
      <view class="cropper-topbar-text">
        <text class="cropper-title">调整头像</text>
        <text class="cropper-subtitle">拖动定位，双指缩放，必要时可旋转图片</text>
      </view>
    </view>
    <view class="cropper-body" @touchstart.prevent="onTouchStart" @touchmove.prevent="onTouchMove" @touchend="onTouchEnd">
      <!-- 图片层 -->
      <img
        ref="imgEl"
        :src="displaySrc"
        class="cropper-img"
        :style="imgStyle"
        @load="onImgLoad"
        draggable="false"
      />
      <!-- 遮罩层（裁剪框外的半透明黑色） -->
      <view class="overlay overlay-top" :style="{ height: `${cropTop}px` }"></view>
      <view class="overlay overlay-bottom" :style="{ top: `${cropTop + cropSize}px`, bottom: '0' }"></view>
      <view class="overlay overlay-left" :style="{ top: `${cropTop}px`, height: `${cropSize}px`, width: `${cropLeft}px` }"></view>
      <view class="overlay overlay-right" :style="{ top: `${cropTop}px`, height: `${cropSize}px`, left: `${cropLeft + cropSize}px`, right: '0' }"></view>
      <!-- 裁剪框边框 -->
      <view class="crop-frame" :style="{ top: `${cropTop}px`, left: `${cropLeft}px`, width: `${cropSize}px`, height: `${cropSize}px` }">
        <view class="corner tl"></view><view class="corner tr"></view>
        <view class="corner bl"></view><view class="corner br"></view>
      </view>
      <view class="crop-center-tip" :style="{ top: `${cropTop + cropSize + 18}px`, left: `${cropLeft}px`, width: `${cropSize}px` }">
        头像将按圆形区域展示
      </view>
    </view>
    <!-- 底部按钮 -->
    <view class="cropper-actions">
      <view class="cropper-actions-inner">
        <view class="cropper-btn cancel" @click="$emit('cancel')">取消</view>
        <view class="cropper-btn rotate" @click="rotateClockwise">旋转</view>
        <view class="cropper-btn confirm" @click="doCrop">确定</view>
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'

const props = defineProps<{
  src: string
  visible: boolean
}>()

const emit = defineEmits<{
  (e: 'confirm', blob: Blob): void
  (e: 'cancel'): void
}>()

const displaySrc = ref('')
const generatedSrc = ref('')

// Image natural dimensions
const natW = ref(0)
const natH = ref(0)

// Image transform state
const scale = ref(1)
const offsetX = ref(0)
const offsetY = ref(0)

// Crop frame (fixed center)
const cropSize = ref(280)
const cropLeft = ref(0)
const cropTop = ref(0)

// Viewport
const vpW = ref(375)
const vpH = ref(500)

function initLayout() {
  if (typeof window !== 'undefined') {
    vpW.value = window.innerWidth
    vpH.value = window.innerHeight - 120 // leave room for buttons
  }
  cropSize.value = Math.floor(Math.min(vpW.value, vpH.value) * 0.75)
  cropLeft.value = Math.floor((vpW.value - cropSize.value) / 2)
  cropTop.value = Math.floor((vpH.value - cropSize.value) / 2)
}

function onImgLoad(e: any) {
  const img = e.target || e.currentTarget
  natW.value = img.naturalWidth
  natH.value = img.naturalHeight

  initLayout()

  // Scale image to fit: fill the crop frame
  const scaleW = cropSize.value / natW.value
  const scaleH = cropSize.value / natH.value
  scale.value = Math.max(scaleW, scaleH)

  // Center image on crop frame
  const dispW = natW.value * scale.value
  const dispH = natH.value * scale.value
  offsetX.value = cropLeft.value + (cropSize.value - dispW) / 2
  offsetY.value = cropTop.value + (cropSize.value - dispH) / 2
}

function resetImageState() {
  natW.value = 0
  natH.value = 0
  scale.value = 1
  offsetX.value = 0
  offsetY.value = 0
  lastTouchDist = 0
  lastTouchX = 0
  lastTouchY = 0
  touching = false
}

function revokeGeneratedSrc() {
  if (generatedSrc.value) {
    URL.revokeObjectURL(generatedSrc.value)
    generatedSrc.value = ''
  }
}

function loadImage(src: string): Promise<HTMLImageElement> {
  return new Promise((resolve, reject) => {
    const img = new Image()
    img.crossOrigin = 'anonymous'
    img.onload = () => resolve(img)
    img.onerror = reject
    img.src = src
  })
}

async function rotateClockwise() {
  if (!displaySrc.value) return
  try {
    const img = await loadImage(displaySrc.value)
    const canvas = document.createElement('canvas')
    canvas.width = img.naturalHeight
    canvas.height = img.naturalWidth
    const ctx = canvas.getContext('2d')
    if (!ctx) return

    ctx.translate(canvas.width / 2, canvas.height / 2)
    ctx.rotate(Math.PI / 2)
    ctx.drawImage(img, -img.naturalWidth / 2, -img.naturalHeight / 2)

    const blob = await new Promise<Blob | null>((resolve) => {
      canvas.toBlob(resolve, 'image/png')
    })
    if (!blob) return

    const nextSrc = URL.createObjectURL(blob)
    revokeGeneratedSrc()
    generatedSrc.value = nextSrc
    displaySrc.value = nextSrc
    resetImageState()
  } catch (error) {
    console.error('rotate image failed', error)
  }
}

const imgStyle = computed(() => ({
  width: `${natW.value * scale.value}px`,
  height: `${natH.value * scale.value}px`,
  transform: `translate(${offsetX.value}px, ${offsetY.value}px)`,
}))

// 锁边：图片不允许移出裁剪框
function clampOffset() {
  const dispW = natW.value * scale.value
  const dispH = natH.value * scale.value
  // 图片右边不能超过裁剪框左边（即图片左边最多到 cropLeft - dispW + cropSize）
  // 图片左边不能超过裁剪框右边
  const minX = cropLeft.value + cropSize.value - dispW
  const maxX = cropLeft.value
  const minY = cropTop.value + cropSize.value - dispH
  const maxY = cropTop.value
  offsetX.value = Math.min(maxX, Math.max(minX, offsetX.value))
  offsetY.value = Math.min(maxY, Math.max(minY, offsetY.value))
}

// 锁边：缩放不能小于裁剪框
function clampScale() {
  const minScaleW = cropSize.value / natW.value
  const minScaleH = cropSize.value / natH.value
  const minScale = Math.max(minScaleW, minScaleH)
  if (scale.value < minScale) scale.value = minScale
}

// Touch handling
let lastTouchDist = 0
let lastTouchX = 0
let lastTouchY = 0
let touching = false

function onTouchStart(e: TouchEvent) {
  if (e.touches.length === 1) {
    lastTouchX = e.touches[0].clientX
    lastTouchY = e.touches[0].clientY
    touching = true
  } else if (e.touches.length === 2) {
    lastTouchDist = Math.hypot(
      e.touches[1].clientX - e.touches[0].clientX,
      e.touches[1].clientY - e.touches[0].clientY
    )
  }
}

function onTouchMove(e: TouchEvent) {
  if (e.touches.length === 1 && touching) {
    const dx = e.touches[0].clientX - lastTouchX
    const dy = e.touches[0].clientY - lastTouchY
    offsetX.value += dx
    offsetY.value += dy
    lastTouchX = e.touches[0].clientX
    lastTouchY = e.touches[0].clientY
    clampOffset()
  } else if (e.touches.length === 2) {
    const dist = Math.hypot(
      e.touches[1].clientX - e.touches[0].clientX,
      e.touches[1].clientY - e.touches[0].clientY
    )
    if (lastTouchDist > 0) {
      const ratio = dist / lastTouchDist
      const centerX = (e.touches[0].clientX + e.touches[1].clientX) / 2
      const centerY = (e.touches[0].clientY + e.touches[1].clientY) / 2

      const newScale = Math.max(0.2, Math.min(5, scale.value * ratio))
      offsetX.value = centerX - (centerX - offsetX.value) * (newScale / scale.value)
      offsetY.value = centerY - (centerY - offsetY.value) * (newScale / scale.value)
      scale.value = newScale
      clampScale()
      clampOffset()
    }
    lastTouchDist = dist
  }
}

function onTouchEnd() {
  touching = false
  lastTouchDist = 0
}

// Crop and output
async function doCrop() {
  const canvas = document.createElement('canvas')
  const outputSize = 800 // output 800x800
  canvas.width = outputSize
  canvas.height = outputSize
  const ctx = canvas.getContext('2d')!

  // White background
  ctx.fillStyle = '#FFFFFF'
  ctx.fillRect(0, 0, outputSize, outputSize)

  // Load image
  const img = await loadImage(displaySrc.value || props.src)

  // Calculate which part of the image falls in the crop frame
  // Crop frame is at (cropLeft, cropTop) with size cropSize in screen coords
  // Image is at (offsetX, offsetY) with displayed size (natW*scale, natH*scale)
  const sx = (cropLeft.value - offsetX.value) / scale.value
  const sy = (cropTop.value - offsetY.value) / scale.value
  const sSize = cropSize.value / scale.value

  ctx.drawImage(img, sx, sy, sSize, sSize, 0, 0, outputSize, outputSize)

  canvas.toBlob((blob) => {
    if (blob) emit('confirm', blob)
  }, 'image/jpeg', 0.85)
}

watch(() => props.src, (nextSrc) => {
  revokeGeneratedSrc()
  displaySrc.value = nextSrc || ''
  resetImageState()
})

watch(() => props.visible, (v) => {
  if (v) {
    displaySrc.value = props.src || ''
    resetImageState()
    initLayout()
    return
  }
  revokeGeneratedSrc()
  displaySrc.value = props.src || ''
  resetImageState()
})
</script>

<style scoped>
.cropper-mask {
  position: fixed; top: 0; left: 0; right: 0; bottom: 0;
  background:
    radial-gradient(circle at top, rgba(99, 102, 241, 0.12), transparent 34%),
    linear-gradient(180deg, #07111f 0%, #04070d 100%);
  z-index: 9999;
  display: flex; flex-direction: column;
}
.cropper-topbar {
  padding: 32px 24px 10px;
  color: #fff;
  flex-shrink: 0;
}
.cropper-topbar-text {
  display: flex;
  flex-direction: column;
  gap: 6px;
}
.cropper-title {
  font-size: 22px;
  font-weight: 700;
  letter-spacing: 0.02em;
}
.cropper-subtitle {
  font-size: 13px;
  color: rgba(226, 232, 240, 0.82);
  line-height: 1.5;
}
.cropper-body {
  flex: 1; position: relative; overflow: hidden;
  touch-action: none;
}
.cropper-img {
  position: absolute; top: 0; left: 0;
  user-select: none; pointer-events: none;
  filter: drop-shadow(0 18px 42px rgba(0, 0, 0, 0.42));
}
.overlay {
  position: absolute; background: rgba(1, 6, 16, 0.64);
}
.overlay-top { top: 0; left: 0; right: 0; }
.overlay-bottom { left: 0; right: 0; }
.overlay-left { left: 0; }
.overlay-right { }
.crop-frame {
  position: absolute;
  border: 2px solid rgba(255,255,255,0.96);
  border-radius: 28px;
  box-shadow:
    0 0 0 9999px rgba(0,0,0,0.02),
    0 18px 36px rgba(15, 23, 42, 0.28),
    inset 0 0 0 1px rgba(255,255,255,0.24);
  pointer-events: none;
}
.corner {
  position: absolute; width: 22px; height: 22px;
  border-color: #fff; border-style: solid;
}
.corner.tl { top: 10px; left: 10px; border-width: 4px 0 0 4px; border-top-left-radius: 10px; }
.corner.tr { top: 10px; right: 10px; border-width: 4px 4px 0 0; border-top-right-radius: 10px; }
.corner.bl { bottom: 10px; left: 10px; border-width: 0 0 4px 4px; border-bottom-left-radius: 10px; }
.corner.br { bottom: 10px; right: 10px; border-width: 0 4px 4px 0; border-bottom-right-radius: 10px; }

.crop-center-tip {
  position: absolute;
  text-align: center;
  font-size: 12px;
  color: rgba(226, 232, 240, 0.9);
  line-height: 1.4;
  pointer-events: none;
}

.cropper-actions {
  padding: 16px 18px 28px;
  flex-shrink: 0;
}
.cropper-actions-inner {
  display: grid;
  grid-template-columns: 1fr 1fr 1.2fr;
  gap: 12px;
  padding: 12px;
  border-radius: 28px;
  background: rgba(15, 23, 42, 0.72);
  border: 1px solid rgba(148, 163, 184, 0.18);
  backdrop-filter: blur(18px);
  box-shadow: 0 16px 36px rgba(0, 0, 0, 0.34);
}
.cropper-btn {
  min-height: 52px;
  font-size: 16px;
  padding: 0 16px;
  border-radius: 18px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: 700;
  letter-spacing: 0.01em;
}
.cropper-btn.cancel {
  color: rgba(241, 245, 249, 0.92);
  background: rgba(30, 41, 59, 0.86);
  border: 1px solid rgba(148, 163, 184, 0.18);
}
.cropper-btn.rotate {
  color: #E2E8F0;
  background: rgba(37, 99, 235, 0.14);
  border: 1px solid rgba(96, 165, 250, 0.26);
}
.cropper-btn.confirm {
  background: linear-gradient(135deg, #4F46E5, #6366F1);
  color: #fff;
  box-shadow: 0 10px 24px rgba(79, 70, 229, 0.32);
}
</style>
