import { readFileSync } from 'fs'
import { resolve } from 'path'

const pagePath = resolve(process.cwd(), 'src/pages/order/detail.vue')
const source = readFileSync(pagePath, 'utf8')

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

assertContains('member_balance_before', 'order receipt should read member balance before snapshot from order data')
assertContains('member_balance_after', 'order receipt should read member balance after snapshot from order data')
assertContains('receiptBalanceBeforePay', 'order receipt should use snapshot-aware before-balance computed value')
assertContains('receiptBalanceAfterPay', 'order receipt should use snapshot-aware after-balance computed value')
assertNotContains('¥{{ balanceBeforePay.toFixed(2) }}', 'receipt header must not render live-card before-balance computed value')
assertNotContains('¥{{ balanceAfterPay.toFixed(2) }}', 'receipt summary must not render live-card after-balance computed value')

console.log('order receipt member balance snapshot checks passed')
