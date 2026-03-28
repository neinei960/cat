<template>
  <view v-if="visible" class="cropper-mask">
    <view class="cropper-body" @touchstart.prevent="onTouchStart" @touchmove.prevent="onTouchMove" @touchend="onTouchEnd">
      <!-- 图片层 -->
      <img
        ref="imgEl"
        :src="src"
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
    </view>
    <!-- 底部按钮 -->
    <view class="cropper-actions">
      <view class="cropper-btn cancel" @click="$emit('cancel')">取消</view>
      <view class="cropper-btn confirm" @click="doCrop">确定</view>
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
  const img = new Image()
  img.crossOrigin = 'anonymous'
  img.src = props.src
  await new Promise(resolve => { img.onload = resolve })

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

watch(() => props.visible, (v) => {
  if (v) initLayout()
})
</script>

<style scoped>
.cropper-mask {
  position: fixed; top: 0; left: 0; right: 0; bottom: 0;
  background: #000; z-index: 9999;
  display: flex; flex-direction: column;
}
.cropper-body {
  flex: 1; position: relative; overflow: hidden;
  touch-action: none;
}
.cropper-img {
  position: absolute; top: 0; left: 0;
  user-select: none; pointer-events: none;
}
.overlay {
  position: absolute; background: rgba(0, 0, 0, 0.55);
}
.overlay-top { top: 0; left: 0; right: 0; }
.overlay-bottom { left: 0; right: 0; }
.overlay-left { left: 0; }
.overlay-right { }
.crop-frame {
  position: absolute;
  border: 2px solid #fff;
  box-shadow: 0 0 0 1px rgba(255,255,255,0.3);
  pointer-events: none;
}
.corner {
  position: absolute; width: 20px; height: 20px;
  border-color: #fff; border-style: solid;
}
.corner.tl { top: -2px; left: -2px; border-width: 3px 0 0 3px; }
.corner.tr { top: -2px; right: -2px; border-width: 3px 3px 0 0; }
.corner.bl { bottom: -2px; left: -2px; border-width: 0 0 3px 3px; }
.corner.br { bottom: -2px; right: -2px; border-width: 0 3px 3px 0; }

.cropper-actions {
  display: flex; height: 120px; background: #000;
  align-items: center; justify-content: space-around;
  padding: 0 40px; flex-shrink: 0;
}
.cropper-btn {
  font-size: 17px; padding: 12px 40px; border-radius: 8px;
}
.cropper-btn.cancel { color: #fff; }
.cropper-btn.confirm { background: #4F46E5; color: #fff; }
</style>
