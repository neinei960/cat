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

console.log('order create product collapse checks passed')
