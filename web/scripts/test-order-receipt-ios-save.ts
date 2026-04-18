import { readFileSync } from 'fs'
import { resolve } from 'path'

function assertIncludes(source: string, expected: string, label: string) {
  if (!source.includes(expected)) {
    throw new Error(`${label}: expected to find "${expected}"`)
  }
}

function main() {
  const file = readFileSync(resolve(process.cwd(), 'src/pages/order/detail.vue'), 'utf8')

  assertIncludes(file, 'const isAppleSafariBrowser = computed(() => {', 'receipt.ios.detect')
  assertIncludes(file, 'const receiptPreviewSrc = computed(() => {', 'receipt.preview.src')
  assertIncludes(file, 'return isAppleSafariBrowser.value ? receiptImageUrl.value : receiptBlobUrl.value || receiptImageUrl.value', 'receipt.preview.prefers.data-url')
  assertIncludes(file, 'const receiptImageHint = computed(() => {', 'receipt.image.hint')
  assertIncludes(file, '如未出现菜单，点击「保存图片」后在新页面长按', 'receipt.image.hint.ios')
  assertIncludes(file, '<img', 'receipt.image.native-tag')
  assertIncludes(file, 'class="receipt-image receipt-image-native"', 'receipt.image.native-class')
  assertIncludes(file, 'function openReceiptImagePreview(src: string)', 'receipt.preview.window.helper')
  assertIncludes(file, "uni.showToast({ title: '新页面已打开，请长按图片保存', icon: 'none' })", 'receipt.preview.window.toast')
  assertIncludes(file, '-webkit-touch-callout: default;', 'receipt.image.native.touch-callout')
}

main()
console.log('order receipt ios save checks passed')
