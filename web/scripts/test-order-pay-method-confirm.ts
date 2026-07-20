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
assert(!source.includes("@click=\"doPay('qrcode')\""), 'qrcode card should not submit directly')
assert(!source.includes("@click=\"doPay('meituan')\""), 'meituan card should not submit directly')

console.log('order pay method confirmation checks passed')
