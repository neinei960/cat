export interface MemberPricingPreviewInput {
  serviceTotal: number
  productTotal: number
  addonTotal?: number
  serviceDiscountRate: number
  productDiscountRate: number
  availableBalance: number
  appointmentDepositDeductionAmount?: number
}

export interface MemberPricingPreview {
  totalAmount: number
  serviceDiscountAmount: number
  productDiscountAmount: number
  discountAmount: number
  finalPayAmount: number
}

function roundCurrency(value: number) {
  return Math.round(Number(value || 0) * 100) / 100
}

function normalizeDiscountRate(rate: number) {
  return rate > 0 && rate < 1 ? rate : 1
}

function consumeDiscountBucket(total: number, rate: number, availableBalance: number) {
  const normalizedTotal = roundCurrency(total)
  const normalizedRate = normalizeDiscountRate(rate)
  const normalizedBalance = roundCurrency(Math.max(availableBalance, 0))
  if (normalizedTotal <= 0 || normalizedBalance <= 0 || normalizedRate >= 1) {
    return { discount: 0, balanceUsed: 0 }
  }

  const fullDiscountedPay = roundCurrency(normalizedTotal * normalizedRate)
  if (normalizedBalance >= fullDiscountedPay) {
    return {
      discount: roundCurrency(normalizedTotal - fullDiscountedPay),
      balanceUsed: fullDiscountedPay,
    }
  }

  const coveredOriginal = roundCurrency(normalizedBalance / normalizedRate)
  return {
    discount: roundCurrency(coveredOriginal * (1 - normalizedRate)),
    balanceUsed: normalizedBalance,
  }
}

export function calculateMemberPricingPreview(input: MemberPricingPreviewInput): MemberPricingPreview {
  const serviceTotal = roundCurrency(input.serviceTotal)
  const productTotal = roundCurrency(input.productTotal)
  const addonTotal = roundCurrency(input.addonTotal || 0)
  const availableBalance = roundCurrency(Math.max(input.availableBalance || 0, 0))

  const service = consumeDiscountBucket(serviceTotal, input.serviceDiscountRate, availableBalance)
  const product = consumeDiscountBucket(
    productTotal,
    input.productDiscountRate,
    roundCurrency(availableBalance - service.balanceUsed),
  )
  const discountAmount = roundCurrency(service.discount + product.discount)
  const totalAmount = roundCurrency(serviceTotal + productTotal + addonTotal)
  const finalPayAmount = roundCurrency(Math.max(
    totalAmount - discountAmount - roundCurrency(input.appointmentDepositDeductionAmount || 0),
    0,
  ))

  return {
    totalAmount,
    serviceDiscountAmount: service.discount,
    productDiscountAmount: product.discount,
    discountAmount,
    finalPayAmount,
  }
}
