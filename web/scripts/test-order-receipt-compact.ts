import { readFileSync } from 'fs'
import { resolve } from 'path'

function assertIncludes(source: string, expected: string, label: string) {
  if (!source.includes(expected)) {
    throw new Error(`${label}: expected to find "${expected}"`)
  }
}

function main() {
  const file = readFileSync(resolve(process.cwd(), 'src/pages/order/detail.vue'), 'utf8')

  assertIncludes(file, 'gap: 12rpx;', 'receipt.container.gap')
  assertIncludes(file, 'padding: 16rpx;', 'receipt.container.padding')
  assertIncludes(file, 'padding: 18rpx;', 'receipt.card.padding')
  assertIncludes(file, 'width: 84rpx;', 'receipt.logo.size')
  assertIncludes(file, 'gap: 12rpx;\n  margin-bottom: 30rpx;', 'receipt.brand.top.spacing')
  assertIncludes(file, 'gap: 10rpx;', 'receipt.meta.list.gap')
  assertIncludes(file, 'padding: 10rpx 0;', 'receipt.item.padding')
  assertIncludes(file, 'font-size: 22rpx;\n  line-height: 1.28;', 'receipt.item.name.size')
  assertIncludes(file, 'font-size: 52rpx;', 'receipt.total.price.font-size')
  assertIncludes(file, 'font-weight: 300;', 'receipt.total.price.font-weight')
  assertIncludes(file, 'color: #C4A35A;', 'receipt.total.price.color')
  assertIncludes(file, 'letter-spacing: -1rpx;', 'receipt.total.price.letter-spacing')
  assertIncludes(file, 'padding: 12rpx 0 calc(4rpx + env(safe-area-inset-bottom));', 'receipt.actions.padding')
  assertIncludes(file, 'min-height: 76rpx;', 'receipt.actions.button.height')
}

main()
console.log('order receipt compact style checks passed')
