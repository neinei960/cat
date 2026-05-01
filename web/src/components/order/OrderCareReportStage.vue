<template>
  <view ref="shellRef" class="care-report-stage-shell" :style="{ height: `${previewHeight}px` }">
    <view ref="stageRef" class="care-report-stage" :style="{ transform: `scale(${previewScale})` }">
      <image class="care-report-stage-base" :src="baseImageHref" mode="scaleToFill" />
      <image v-if="portraitHref" class="care-report-stage-portrait" :src="portraitHref" mode="aspectFill" />

      <view
        v-for="field in primaryFields"
        :key="field.key"
        class="care-report-stage-text"
        :style="field.style"
      >
        {{ field.text }}
      </view>

      <view
        v-for="mark in checkmarks"
        :key="mark.key"
        class="care-report-stage-check"
        :style="mark.style"
      >
        ✓
      </view>

      <view
        v-for="note in notes"
        :key="note.key"
        class="care-report-stage-note"
        :style="note.style"
      >
        {{ note.text }}
      </view>
    </view>

    <view v-if="editable" class="care-report-stage-overlay">
      <view
        v-for="hotspot in hotspots"
        :key="hotspot.key"
        :class="[
          'care-report-stage-hotspot',
          hotspot.circle ? 'circle' : '',
          hotspot.active ? 'active' : '',
          hotspot.selected ? 'selected' : '',
        ]"
        :style="hotspot.style"
        @click.stop="handleHotspotClick(hotspot)"
      >
        <text v-if="hotspot.label" class="care-report-stage-hotspot-label">{{ hotspot.label }}</text>
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref } from 'vue'
import html2canvas from 'html2canvas'
import reportBaseImage from '@/assets/order-care-report-base.jpg'
import type { OrderCareReportDraft, OrderCareReportSectionKey } from '@/utils/order-care-report'

type StyleObject = Record<string, string>
type Point = { x: number; y: number }
type MaybeElementRef = HTMLElement | { $el?: Element | null } | null
type EditableFieldKey = 'petName' | 'breed' | 'gender' | 'age' | 'careContent' | 'careDate' | 'nextCareDate' | 'weight'
type StageEditTarget =
  | { type: 'field'; key: EditableFieldKey }
  | { type: 'note'; sectionKey: OrderCareReportSectionKey }
  | { type: 'portrait' }

type Hotspot =
  | {
      key: string
      style: StyleObject
      circle?: boolean
      active?: boolean
      selected?: boolean
      label?: string
      kind: 'edit'
      target: StageEditTarget
    }
  | {
      key: string
      style: StyleObject
      circle?: boolean
      active?: boolean
      selected?: boolean
      label?: string
      kind: 'body'
      value: string
    }
  | {
      key: string
      style: StyleObject
      circle?: boolean
      active?: boolean
      selected?: boolean
      label?: string
      kind: 'section'
      sectionKey: OrderCareReportSectionKey
      value: string
    }

const props = withDefaults(defineProps<{
  draft: OrderCareReportDraft
  editable?: boolean
  activeEditorKey?: string
}>(), {
  editable: false,
  activeEditorKey: '',
})

const emit = defineEmits<{
  (e: 'edit-target', payload: StageEditTarget): void
  (e: 'toggle-body-shape', value: string): void
  (e: 'toggle-section-check', payload: { sectionKey: OrderCareReportSectionKey; value: string }): void
}>()

const REPORT_WIDTH = 1279
const REPORT_HEIGHT = 1810
const UNDERLINE_GAP = 8
const NOTE_LINE_HEIGHT = 32

const shellRef = ref<MaybeElementRef>(null)
const stageRef = ref<MaybeElementRef>(null)
const previewScale = ref(1)
const previewHeight = computed(() => Math.round(REPORT_HEIGHT * previewScale.value))

let resizeObserver: ResizeObserver | null = null

const fieldRects: Record<EditableFieldKey, { x: number; y: number; width: number; height: number }> = {
  petName: { x: 228, y: 154, width: 492, height: 132 },
  breed: { x: 228, y: 296, width: 492, height: 122 },
  gender: { x: 118, y: 446, width: 286, height: 96 },
  age: { x: 412, y: 446, width: 290, height: 96 },
  careContent: { x: 228, y: 546, width: 494, height: 106 },
  careDate: { x: 228, y: 686, width: 492, height: 96 },
  nextCareDate: { x: 848, y: 680, width: 316, height: 102 },
  weight: { x: 88, y: 792, width: 210, height: 92 },
}

const bodyShapeAnchors: Record<string, Point> = {
  thin: { x: 408, y: 833 },
  skinny: { x: 573, y: 833 },
  standard: { x: 737, y: 833 },
  chubby: { x: 901, y: 833 },
  obese: { x: 1066, y: 833 },
}

const sectionAnchors: Record<OrderCareReportSectionKey, Record<string, Point>> = {
  skin: {
    normal: { x: 408, y: 928 },
    dandruff: { x: 573, y: 928 },
    red: { x: 737, y: 928 },
    greasy: { x: 901, y: 928 },
    scab: { x: 1066, y: 928 },
  },
  hair: {
    shedding: { x: 408, y: 1027 },
    undercoat_many: { x: 573, y: 1027 },
    dry: { x: 737, y: 1027 },
    greasy: { x: 901, y: 1027 },
    matting: { x: 1066, y: 1027 },
  },
  nails: {
    trimmed: { x: 408, y: 1125 },
    dewclaw_abnormal: { x: 573, y: 1125 },
    pads_dry: { x: 737, y: 1125 },
    too_long: { x: 901, y: 1125 },
    wound: { x: 1066, y: 1125 },
  },
  eyesFace: {
    cleaned: { x: 408, y: 1222 },
    tear_many: { x: 573, y: 1222 },
    eye_red: { x: 737, y: 1222 },
    eye_abnormal: { x: 901, y: 1222 },
    wound: { x: 1066, y: 1222 },
  },
  ears: {
    cleaned: { x: 408, y: 1321 },
    touch_sensitive: { x: 573, y: 1321 },
    inflamed: { x: 737, y: 1321 },
    earwax: { x: 901, y: 1321 },
    black_earwax: { x: 1066, y: 1321 },
  },
  oral: {
    normal: { x: 408, y: 1420 },
    tartar: { x: 737, y: 1420 },
    gum_red: { x: 901, y: 1420 },
    gum_swollen: { x: 1066, y: 1420 },
  },
  anus: {
    normal: { x: 408, y: 1605 },
    red: { x: 737, y: 1605 },
    inflamed: { x: 901, y: 1605 },
    anal_gland_swollen: { x: 1066, y: 1605 },
  },
}

const noteSpecs: Array<{ key: OrderCareReportSectionKey; x: number; y: number; width: number; limit: number }> = [
  { key: 'skin', x: 494, y: 968, width: 670, limit: 32 },
  { key: 'hair', x: 494, y: 1068, width: 670, limit: 32 },
  { key: 'nails', x: 494, y: 1166, width: 670, limit: 32 },
  { key: 'eyesFace', x: 494, y: 1262, width: 670, limit: 32 },
  { key: 'ears', x: 494, y: 1362, width: 670, limit: 32 },
  { key: 'oral', x: 494, y: 1460, width: 670, limit: 32 },
  { key: 'anus', x: 494, y: 1646, width: 670, limit: 32 },
]

const baseImageHref = computed(() => resolveAbsoluteUrl(reportBaseImage))
const portraitHref = computed(() => resolveAbsoluteUrl(props.draft.portraitUrl))

const primaryFields = computed(() => [
  createCenteredField('pet_name', props.draft.petName, 279, 228, 387, 72, 54, 700),
  createCenteredField('breed', props.draft.breed, 279, 369, 387, 66, 44, 700),
  createCenteredField('gender', props.draft.gender, 196, 507, 175, 56, 36, 700),
  createCenteredField('age', props.draft.age, 490, 507, 178, 56, 36, 700),
  createCenteredField('care_content', props.draft.careContent, 279, 597, 387, 60, 34, 700, 18),
  createCenteredField('care_date', normalizeDate(props.draft.careDate), 279, 737, 387, 52, 32, 700),
  createCenteredField('next_care_date', normalizeDate(props.draft.nextCareDate), 905, 737, 229, 48, 30, 600),
  createStartField('weight', normalizeWeight(props.draft.weight), 104, 834, 118, 44, 30, 600, 8),
])

const checkmarks = computed(() => {
  const marks: Array<{ key: string; style: StyleObject }> = []
  const bodyAnchor = bodyShapeAnchors[props.draft.bodyShape]
  if (bodyAnchor) {
    marks.push({ key: `body-${props.draft.bodyShape}`, style: createCheckmarkStyle(bodyAnchor.x, bodyAnchor.y) })
  }

  for (const [sectionKey, anchors] of Object.entries(sectionAnchors) as Array<[OrderCareReportSectionKey, Record<string, Point>]>) {
    const selectedValues = new Set(props.draft[sectionKey].checks)
    for (const value of selectedValues) {
      const anchor = anchors[value]
      if (!anchor) continue
      marks.push({ key: `${sectionKey}-${value}`, style: createCheckmarkStyle(anchor.x, anchor.y) })
    }
  }

  return marks
})

const notes = computed(() => {
  return noteSpecs
    .map((spec) => ({
      key: spec.key,
      text: truncateText(props.draft[spec.key].note, spec.limit),
      style: pxStyle({
        left: spec.x,
        top: spec.y,
        width: spec.width,
        height: NOTE_LINE_HEIGHT,
        display: 'flex',
        alignItems: 'flex-end',
        paddingBottom: UNDERLINE_GAP,
      }),
    }))
    .filter((item) => item.text)
})

const hotspots = computed<Hotspot[]>(() => {
  if (!props.editable) return []

  const list: Hotspot[] = [
    {
      key: 'portrait',
      kind: 'edit',
      target: { type: 'portrait' },
      circle: true,
      active: props.activeEditorKey === 'portrait',
      label: props.draft.portraitUrl ? '' : '上传照片',
      style: createOverlayRectStyle(790, 145, 420, 420, true),
    },
  ]

  for (const [fieldKey, rect] of Object.entries(fieldRects) as Array<[EditableFieldKey, { x: number; y: number; width: number; height: number }]>) {
    list.push({
      key: `field-${fieldKey}`,
      kind: 'edit',
      target: { type: 'field', key: fieldKey },
      active: props.activeEditorKey === `field:${fieldKey}`,
      style: createOverlayRectStyle(rect.x, rect.y, rect.width, rect.height),
    })
  }

  for (const spec of noteSpecs) {
    list.push({
      key: `note-${spec.key}`,
      kind: 'edit',
      target: { type: 'note', sectionKey: spec.key },
      active: props.activeEditorKey === `note:${spec.key}`,
      style: createNoteOverlayStyle(spec.x - 26, spec.y - 14, spec.width + 44, 54),
    })
  }

  for (const [value, anchor] of Object.entries(bodyShapeAnchors)) {
    list.push({
      key: `body-${value}`,
      kind: 'body',
      value,
      selected: props.draft.bodyShape === value,
      style: createCenteredOverlayStyle(anchor.x, anchor.y - 8, 112, 72),
    })
  }

  for (const [sectionKey, anchors] of Object.entries(sectionAnchors) as Array<[OrderCareReportSectionKey, Record<string, Point>]>) {
    const selectedValues = new Set(props.draft[sectionKey].checks)
    for (const [value, anchor] of Object.entries(anchors)) {
      list.push({
        key: `${sectionKey}-${value}`,
        kind: 'section',
        sectionKey,
        value,
        selected: selectedValues.has(value),
        style: createCenteredOverlayStyle(anchor.x, anchor.y - 8, 108, 70),
      })
    }
  }

  return list
})

onMounted(() => {
  if (typeof ResizeObserver === 'undefined') return
  resizeObserver = new ResizeObserver(() => {
    updatePreviewScale()
  })
  const shellElement = getDomElement(shellRef.value)
  if (shellElement) {
    resizeObserver.observe(shellElement)
  }
  updatePreviewScale()
})

onBeforeUnmount(() => {
  resizeObserver?.disconnect()
  resizeObserver = null
})

function updatePreviewScale() {
  const width = getDomElement(shellRef.value)?.clientWidth || 0
  if (!width) return
  previewScale.value = width / REPORT_WIDTH
}

function getDomElement(target: MaybeElementRef) {
  if (!target) return null
  if (target instanceof HTMLElement) return target
  if ('$el' in target && target.$el instanceof HTMLElement) return target.$el
  return null
}

function createCenteredField(key: string, text: string, left: number, top: number, width: number, height: number, fontSize: number, fontWeight: number, limit = 14) {
  const compactedText = truncateText(text, limit)
  const fittedFontSize = fitCenteredFontSize(compactedText, width, fontSize)
  return {
    key,
    text: compactedText,
    style: pxStyle({
      left,
      top,
      width,
      height,
      fontSize: fittedFontSize,
      fontWeight,
      display: 'flex',
      alignItems: 'flex-end',
      justifyContent: 'center',
      paddingBottom: UNDERLINE_GAP,
      textAlign: 'center',
      overflow: 'hidden',
    }),
  }
}

function createStartField(key: string, text: string, left: number, top: number, width: number, height: number, fontSize: number, fontWeight: number, limit = 10) {
  return {
    key,
    text: truncateText(text, limit),
    style: pxStyle({
      left,
      top,
      width,
      height,
      fontSize,
      fontWeight,
      display: 'flex',
      alignItems: 'flex-end',
      justifyContent: 'flex-start',
      paddingBottom: UNDERLINE_GAP,
      textAlign: 'left',
    }),
  }
}

function createCheckmarkStyle(x: number, y: number): StyleObject {
  return pxStyle({
    left: x - 16,
    top: y - 30,
    width: 32,
    height: 32,
    fontSize: 34,
    lineHeight: 32,
    textAlign: 'center',
  })
}

function createOverlayRectStyle(left: number, top: number, width: number, height: number, circle = false): StyleObject {
  const scale = previewScale.value
  const scaledWidth = Math.round(width * scale)
  const scaledHeight = Math.round(height * scale)
  return {
    left: `${Math.round(left * scale)}px`,
    top: `${Math.round(top * scale)}px`,
    width: `${scaledWidth}px`,
    height: `${scaledHeight}px`,
    borderRadius: circle ? '999px' : '16px',
  }
}

function createNoteOverlayStyle(left: number, top: number, width: number, height: number): StyleObject {
  const scale = previewScale.value
  const scaledWidth = Math.round(width * scale)
  const scaledHeight = Math.round(height * scale)
  const targetHeight = Math.max(28, scaledHeight)
  const verticalInset = Math.round((targetHeight - scaledHeight) / 2)
  return {
    left: `${Math.round(left * scale)}px`,
    top: `${Math.round(top * scale) - verticalInset}px`,
    width: `${scaledWidth}px`,
    height: `${targetHeight}px`,
    borderRadius: '16px',
    zIndex: '4',
  }
}

function createCenteredOverlayStyle(centerX: number, centerY: number, width: number, height: number): StyleObject {
  const scale = previewScale.value
  const scaledWidth = Math.max(34, Math.round(width * scale))
  const scaledHeight = Math.max(22, Math.round(height * scale))
  return {
    left: `${Math.round(centerX * scale - scaledWidth / 2)}px`,
    top: `${Math.round(centerY * scale - scaledHeight / 2)}px`,
    width: `${scaledWidth}px`,
    height: `${scaledHeight}px`,
    borderRadius: '12px',
  }
}

function handleHotspotClick(hotspot: Hotspot) {
  if (hotspot.kind === 'edit') {
    emit('edit-target', hotspot.target)
    return
  }
  if (hotspot.kind === 'body') {
    emit('toggle-body-shape', hotspot.value)
    return
  }
  emit('toggle-section-check', {
    sectionKey: hotspot.sectionKey,
    value: hotspot.value,
  })
}

function pxStyle(input: Record<string, string | number>) {
  const style: StyleObject = {}
  Object.entries(input).forEach(([key, value]) => {
    if (typeof value === 'number') {
      style[key] = `${value}px`
      return
    }
    style[key] = value
  })
  return style
}

function truncateText(value: string | undefined | null, limit: number) {
  const text = String(value || '').trim()
  if (!text) return ''
  const compacted = text.replace(/\s+/g, ' ')
  const runes = Array.from(compacted)
  if (runes.length <= limit) return compacted
  return `${runes.slice(0, limit).join('')}…`
}

function fitCenteredFontSize(text: string, width: number, baseFontSize: number, minFontSize = 22) {
  const compacted = String(text || '').trim()
  if (!compacted) return baseFontSize

  const estimatedWidth = estimateTextUnits(compacted) * baseFontSize
  if (estimatedWidth <= width) return baseFontSize

  const scaledFontSize = Math.floor((width / estimatedWidth) * baseFontSize)
  return Math.max(minFontSize, Math.min(baseFontSize, scaledFontSize))
}

function estimateTextUnits(text: string) {
  return Array.from(text).reduce((sum, char) => sum + estimateCharacterUnit(char), 0)
}

function estimateCharacterUnit(char: string) {
  if (/\s/.test(char)) return 0.28
  if (/[（()）]/.test(char)) return 0.38
  if (/[、，,。.:\-]/.test(char)) return 0.3
  if (/[A-Z]/.test(char)) return 0.7
  if (/[a-z0-9]/.test(char)) return 0.58
  if (/[\u4E00-\u9FFF\u3400-\u4DBF]/.test(char)) return 1
  return 0.62
}

function normalizeDate(value: string | undefined | null) {
  return String(value || '').trim().replace(/\./g, '-')
}

function normalizeWeight(value: string | undefined | null) {
  return String(value || '')
    .replace(/kg|KG|Kg|kG|公斤/g, '')
    .trim()
}

function resolveAbsoluteUrl(value: string) {
  if (!value) return ''
  if (/^(https?:)?\/\//i.test(value) || value.startsWith('data:')) return value
  if (typeof window === 'undefined') return value
  return new URL(value, window.location.origin).toString()
}

async function exportPngBlob() {
  await nextTick()
  const shellElement = getDomElement(shellRef.value)
  const element = (shellElement?.querySelector('.care-report-stage') as HTMLElement | null) || getDomElement(stageRef.value)
  if (!element) {
    throw new Error('找不到护理报告预览')
  }
  await waitForStageImagesReady(element)
  const rect = element.getBoundingClientRect()
  const exportScale = rect.width ? REPORT_WIDTH / rect.width : 1
  const canvas = await html2canvas(element, {
    backgroundColor: '#FFFFFF',
    scale: exportScale,
    useCORS: true,
    allowTaint: true,
    logging: false,
    width: rect.width,
    height: rect.height,
    scrollX: 0,
    scrollY: 0,
  })
  return await new Promise<Blob>((resolve, reject) => {
    canvas.toBlob((blob) => {
      if (!blob) {
        reject(new Error('护理报告导出失败'))
        return
      }
      resolve(blob)
    }, 'image/png')
  })
}

async function waitForStageImagesReady(element: HTMLElement) {
  const images = Array.from(element.querySelectorAll('img')) as HTMLImageElement[]
  if (!images.length) return

  await Promise.all(
    images.map((image) => {
      if (image.complete && image.naturalWidth > 0) {
        return Promise.resolve()
      }
      return new Promise<void>((resolve) => {
        const done = () => {
          image.removeEventListener('load', done)
          image.removeEventListener('error', done)
          resolve()
        }
        image.addEventListener('load', done, { once: true })
        image.addEventListener('error', done, { once: true })
      })
    })
  )
}

defineExpose({
  exportPngBlob,
})
</script>

<style scoped>
.care-report-stage-shell {
  position: relative;
  width: 100%;
  overflow: hidden;
  border-radius: 24rpx;
  background: #FFFFFF;
}

.care-report-stage {
  position: absolute;
  top: 0;
  left: 0;
  width: 1279px;
  height: 1810px;
  transform-origin: top left;
  background: #FFFFFF;
  overflow: hidden;
}

.care-report-stage-base {
  position: absolute;
  inset: 0;
  width: 1279px;
  height: 1810px;
}

.care-report-stage-portrait {
  position: absolute;
  left: 790px;
  top: 145px;
  width: 420px;
  height: 420px;
  border-radius: 999px;
}

.care-report-stage-text,
.care-report-stage-note,
.care-report-stage-check {
  position: absolute;
  box-sizing: border-box;
  color: #111111;
  font-family: "PingFang SC", "Hiragino Sans GB", "Microsoft YaHei", sans-serif;
}

.care-report-stage-text {
  white-space: nowrap;
}

.care-report-stage-note {
  font-size: 22px;
  font-weight: 400;
  line-height: 1;
  white-space: nowrap;
}

.care-report-stage-check {
  font-size: 34px;
  font-weight: 700;
}

.care-report-stage-overlay {
  position: absolute;
  inset: 0;
  z-index: 3;
}

.care-report-stage-hotspot {
  position: absolute;
  box-sizing: border-box;
  border: 2px solid transparent;
  background: rgba(255, 255, 255, 0.01);
}

.care-report-stage-hotspot.circle {
  border-radius: 999px;
}

.care-report-stage-hotspot.active {
  border-color: rgba(215, 168, 67, 0.9);
  box-shadow: 0 0 0 2px rgba(255, 240, 196, 0.65) inset;
}

.care-report-stage-hotspot.selected {
  border-color: rgba(215, 168, 67, 0.72);
  background: rgba(244, 212, 121, 0.18);
}

.care-report-stage-hotspot-label {
  position: absolute;
  left: 50%;
  top: 50%;
  transform: translate(-50%, -50%);
  padding: 10rpx 18rpx;
  border-radius: 999rpx;
  background: rgba(255, 249, 240, 0.88);
  color: #8A6B2F;
  font-size: 22rpx;
  font-weight: 600;
  white-space: nowrap;
}
</style>
