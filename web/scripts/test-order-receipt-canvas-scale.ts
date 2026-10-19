import assert from 'node:assert/strict'
import { getReceiptCanvasScale } from '../src/utils/receipt-image'

function main() {
  assert.equal(getReceiptCanvasScale(365, 823, false), 2, 'short desktop receipts should keep the crisp 2x scale')

  const tallMobileScale = getReceiptCanvasScale(390, 5200, true)
  assert(tallMobileScale < 2, 'tall mobile receipts should reduce scale to avoid iOS canvas limits')
  assert(tallMobileScale >= 1, 'mobile fallback should keep readable output')

  const boundedPixels = 390 * 5200 * tallMobileScale * tallMobileScale
  assert(boundedPixels <= 4_000_000, 'mobile receipt canvas should stay within the conservative iOS pixel budget')
}

main()
console.log('order receipt canvas scale checks passed')
