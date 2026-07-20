import { readFileSync } from 'fs'
import { resolve } from 'path'

const detailPath = resolve(process.cwd(), 'src/pages/order/detail.vue')
const source = readFileSync(detailPath, 'utf8')

function assertContains(text: string, message: string) {
  if (!source.includes(text)) {
    throw new Error(message)
  }
}

function assertNotContains(text: string, message: string) {
  if (source.includes(text)) {
    throw new Error(message)
  }
}

assertContains('receiptServiceDiscountRate', 'receipt discount label should use member service discount rate')
assertContains('receiptProductDiscountRate', 'receipt discount label should use member product discount rate')
assertContains('getMemberCardDiscountRate', 'order detail should derive labels from member card rules')
assertContains('receiptServiceDiscountRate.value * 10', 'service discount label should display member rule rate')
assertContains('receiptProductDiscountRate.value * 10', 'product discount label should display member rule rate')
assertNotContains('return `${(serviceDiscountRate.value * 10).toFixed(1)}折`', 'service label must not use effective order discount rate')
assertNotContains('return `${(productDiscountRate.value * 10).toFixed(1)}折`', 'product label must not use effective order discount rate')

console.log('order receipt discount label rate checks passed')
