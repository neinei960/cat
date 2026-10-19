import { calculateMemberPricingPreview } from '../src/utils/member-pricing'

function assertEqual(actual: number, expected: number, message: string) {
  if (actual !== expected) {
    throw new Error(`${message}: expected ${expected}, got ${actual}`)
  }
}

const zeroBalance = calculateMemberPricingPreview({
  serviceTotal: 188,
  productTotal: 0,
  serviceDiscountRate: 0.9,
  productDiscountRate: 1,
  availableBalance: 0,
})
assertEqual(zeroBalance.finalPayAmount, 188, 'zero balance final pay')
assertEqual(zeroBalance.discountAmount, 0, 'zero balance discount')

const partialBalance = calculateMemberPricingPreview({
  serviceTotal: 188,
  productTotal: 0,
  serviceDiscountRate: 0.9,
  productDiscountRate: 1,
  availableBalance: 2,
})
assertEqual(partialBalance.finalPayAmount, 187.78, 'partial balance final pay')
assertEqual(partialBalance.serviceDiscountAmount, 0.22, 'partial balance service discount')

const enoughBalance = calculateMemberPricingPreview({
  serviceTotal: 188,
  productTotal: 0,
  serviceDiscountRate: 0.9,
  productDiscountRate: 1,
  availableBalance: 200,
})
assertEqual(enoughBalance.finalPayAmount, 169.2, 'enough balance final pay')
assertEqual(enoughBalance.serviceDiscountAmount, 18.8, 'enough balance service discount')

console.log('member pricing preview tests passed')
