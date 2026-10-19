import assert from 'node:assert/strict'
import fs from 'node:fs'
import path from 'node:path'

const filePath = path.resolve(__dirname, '../../src/pages/order/detail.vue')
const source = fs.readFileSync(filePath, 'utf8')
const saveUtilPath = path.resolve(__dirname, '../../src/utils/web-image-save.ts')
const saveUtilSource = fs.readFileSync(saveUtilPath, 'utf8')

assert(
  source.includes('<view class="receipt-wrap" v-if="!receiptImageUrl">'),
  'receipt modal should hide the original receipt DOM after image generation to avoid duplicate preview content',
)

assert(
  source.includes('async function downloadReceiptImage()'),
  'receipt save flow should be asynchronous so it can use share/open fallbacks on mobile Safari',
)

assert(
  saveUtilSource.includes('navigator.share') || saveUtilSource.includes('window.open('),
  'receipt save flow should provide a mobile Safari fallback instead of relying only on a.download',
)

assert(
  source.includes("import { uploadH5File } from '@/api/upload'"),
  'receipt generation should upload the PNG so iPhone Safari receives a normal image URL',
)

assert(
  source.includes('const receiptRemoteUrl = ref'),
  'receipt preview should retain the uploaded image URL for saving',
)

assert(
  source.includes('uploadH5File(receiptFile, { preserveOriginal: true })'),
  'receipt upload should preserve the original PNG',
)

assert(
  source.includes('blob: receiptImageBlob.value || undefined'),
  'receipt save should pass the already generated blob to preserve the user gesture for native sharing',
)

assert(
  !source.includes('@touchstart.stop'),
  'native receipt images must not intercept the iPhone long-press gesture',
)

assert(
  source.includes('.receipt-image-wrap {') && source.includes('overflow-y: auto;'),
  'generated receipt previews should remain scrollable on a phone viewport',
)

const shareIndex = saveUtilSource.indexOf('navigator.share')
const safariPreviewIndex = saveUtilSource.indexOf('isAppleSafariBrowser() && openImagePreviewWindow')
assert(
  shareIndex >= 0 && safariPreviewIndex >= 0 && shareIndex < safariPreviewIndex,
  'native file sharing should be attempted before the Safari preview-window fallback',
)

assert(
  source.includes('<view class="modal-mask" v-if="showReceipt" @click="closeReceipt">'),
  'receipt modal should clear generated image state when dismissed from the mask',
)

console.log('order receipt image regression test passed')
