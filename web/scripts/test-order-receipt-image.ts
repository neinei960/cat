import assert from 'node:assert/strict'
import fs from 'node:fs'
import path from 'node:path'

const filePath = path.resolve(__dirname, '../../src/pages/order/detail.vue')
const source = fs.readFileSync(filePath, 'utf8')

assert(
  source.includes('<view class="receipt-wrap" v-if="!receiptImageUrl">'),
  'receipt modal should hide the original receipt DOM after image generation to avoid duplicate preview content',
)

assert(
  source.includes('async function downloadReceiptImage()'),
  'receipt save flow should be asynchronous so it can use share/open fallbacks on mobile Safari',
)

assert(
  source.includes('navigator.share') || source.includes('window.open('),
  'receipt save flow should provide a mobile Safari fallback instead of relying only on a.download',
)

assert(
  source.includes('<view class="modal-mask" v-if="showReceipt" @click="closeReceipt">'),
  'receipt modal should clear generated image state when dismissed from the mask',
)

console.log('order receipt image regression test passed')
