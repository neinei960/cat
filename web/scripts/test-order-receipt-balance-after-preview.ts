import { readFileSync } from 'fs'
import { resolve } from 'path'

const detailPath = resolve(process.cwd(), 'src/pages/order/detail.vue')
const source = readFileSync(detailPath, 'utf8')

function assertContains(text: string, message: string) {
  if (!source.includes(text)) {
    throw new Error(message)
  }
}

assertContains('isMixedBalancePreview', 'receipt should detect unpaid mixed balance preview state')
assertContains('return 0', 'receipt should show zero balance after when balance will be fully used')
assertContains('memberBalance.value < Number(order.value?.pay_amount || 0)', 'mixed balance preview should require insufficient balance')
assertContains('receiptBalanceUsedAmount', 'receipt should show the balance amount that will be used')
assertContains('receiptCashPayAmount', 'receipt should show the customer cash supplement amount')
assertContains('用户需补', 'receipt should label the cash supplement amount clearly')

console.log('order receipt balance after preview checks passed')
