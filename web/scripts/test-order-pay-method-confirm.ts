import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'

function assert(condition: unknown, message: string) {
  if (!condition) {
    throw new Error(message)
  }
}

const pagePath = resolve(process.cwd(), 'src/pages/order/detail.vue')
const source = readFileSync(pagePath, 'utf8')

assert(source.includes('const paying = ref(false)'), 'order detail should lock payment submission while paying')
assert(source.includes('const payMethodConfirmMap'), 'order detail should define explicit payment confirmation labels')
assert(source.includes('function confirmPayMethod'), 'order detail should confirm the chosen pay method before submit')
assert(source.includes("@click=\"confirmPayMethod('qrcode')\""), 'qrcode card should open confirmation for qrcode')
assert(source.includes("@click=\"confirmPayMethod('meituan')\""), 'meituan card should open confirmation for meituan')
assert(source.includes('await doPay(method, true)'), 'confirmed method should be the method submitted to payOrder')
assert(source.includes('const balancePaymentPreviewAmount = computed'), 'balance payment should calculate the current member-card price before confirmation')
assert(source.includes('const excludesMemberServiceDiscount = computed'), 'order detail should centralize member service discount exclusions')
assert(source.includes("order.value?.order_kind === 'feeding'"), 'feeding orders should be excluded from member service discounts')
assert(source.includes('Number(order.value?.feeding_plan_id || 0) > 0'), 'feeding-linked orders should be excluded even when order_kind is unavailable')
assert(source.includes('if (excludesMemberServiceDiscount.value) return 1'), 'receipt and balance preview should show no service discount for excluded orders')
assert(source.includes('memberBalance.value < balancePaymentPreviewAmount.value'), 'balance sufficiency check should use the discounted preview amount')
assert(source.includes('从会员余额中扣除¥${balancePaymentPreviewAmount.value.toFixed(2)}'), 'balance confirmation should show the discounted preview amount')
assert(!source.includes("@click=\"doPay('qrcode')\""), 'qrcode card should not submit directly')
assert(!source.includes("@click=\"doPay('meituan')\""), 'meituan card should not submit directly')
assert(source.includes('onShow(async () =>'), 'returning from customer detail should refresh order pricing')
assert(source.includes('async function refreshCustomerCard()'), 'card state refresh should be shared by page reload and payment')
assert(/async function reload\(\)[\s\S]*?await refreshCustomerCard\(\)/.test(source), 'order reload must refresh disabled card state too')

console.log('order pay method confirmation checks passed')
