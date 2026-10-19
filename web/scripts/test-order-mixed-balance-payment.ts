import { readFileSync } from 'fs'
import { resolve } from 'path'

const detailPath = resolve(process.cwd(), 'src/pages/order/detail.vue')
const apiPath = resolve(process.cwd(), 'src/api/order.ts')
const detail = readFileSync(detailPath, 'utf8')
const api = readFileSync(apiPath, 'utf8')

function assertContains(source: string, text: string, message: string) {
  if (!source.includes(text)) {
    throw new Error(message)
  }
}

function assertNotContains(source: string, text: string, message: string) {
  if (source.includes(text)) {
    throw new Error(message)
  }
}

assertContains(api, 'cashPayMethod', 'payOrder API should accept a cashPayMethod argument')
assertContains(api, 'cash_pay_method', 'payOrder API should send cash_pay_method to backend')
assertContains(detail, 'payWithMixedBalance', 'order detail should have a mixed balance payment flow')
assertContains(detail, "'mixed_balance'", 'order detail should pay with mixed_balance when balance is insufficient')
assertContains(detail, 'cashPayMethod', 'order detail should pass the selected cash pay method')
assertNotContains(detail, '会员余额¥${memberBalance.value.toFixed(2)}，应付¥${order.value.pay_amount.toFixed(2)}，余额不足。', 'insufficient balance should not only show an error modal')

console.log('order mixed balance payment checks passed')
