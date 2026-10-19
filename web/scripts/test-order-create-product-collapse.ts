import assert from 'node:assert/strict'
import fs from 'node:fs'
import path from 'node:path'

const pagePath = path.resolve(__dirname, '../../src/pages/order/create.vue')
const source = fs.readFileSync(pagePath, 'utf8')

assert(
  source.includes('const PRODUCT_COLLAPSED_LIMIT = 5'),
  'order create should define a five-product collapsed limit',
)

assert(
  source.includes('const productListExpanded = ref(false)'),
  'order create should track whether the product list is expanded',
)

assert(
  source.includes('const visibleProductCards = computed'),
  'order create should derive the displayed product cards separately from all filtered products',
)

assert(
  source.includes('v-for="product in visibleProductCards"'),
  'order create should render the collapsed product card list instead of all filtered products',
)

assert(
  source.includes('展开全部') && source.includes('收起商品'),
  'order create should expose expand and collapse controls for long product lists',
)

assert(
  source.includes('function resetManualPayAfterProductChange()'),
  'product mutations should share one path that clears a stale manual total',
)

for (const functionName of ['addProductToCart', 'changeCartQuantity', 'removeCartItem']) {
  const functionStart = source.indexOf(`function ${functionName}`)
  const nextFunction = source.indexOf('\nfunction ', functionStart + 1)
  const functionSource = source.slice(functionStart, nextFunction > functionStart ? nextFunction : undefined)
  assert(
    functionSource.includes('resetManualPayAfterProductChange()'),
    `${functionName} should restore automatic pricing when the cart changes`,
  )
}

assert(
  source.includes('const persistedManualAdjustment ='),
  'order edit should detect an actually persisted manual adjustment',
)

assert(
  !source.includes("} else {\n      manualPayTouched.value = true\n      manualPayInput.value = Number(order.pay_amount || 0).toFixed(2)\n    }"),
  'ordinary edited orders must not enter manual-total mode unconditionally',
)

console.log('order create product collapse checks passed')
