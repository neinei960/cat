import { readFileSync } from 'fs'
import { resolve } from 'path'

const detailPath = resolve(process.cwd(), 'src/pages/order/detail.vue')
const source = readFileSync(detailPath, 'utf8')

const maskRule = source.match(/\.modal-mask\s*\{[\s\S]*?\n\}/)
if (!maskRule) {
  throw new Error('order detail should define modal-mask style')
}

const zIndexMatch = maskRule[0].match(/z-index:\s*(\d+)/)
if (!zIndexMatch) {
  throw new Error('modal-mask should define an explicit z-index')
}

const zIndex = Number(zIndexMatch[1])
if (zIndex >= 999) {
  throw new Error(`modal-mask z-index ${zIndex} should stay below uni-modal z-index 999`)
}

console.log('order pay modal z-index checks passed')
