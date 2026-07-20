import assert from 'node:assert/strict'
import fs from 'node:fs'
import path from 'node:path'

const pagePath = path.resolve(process.cwd(), 'src/pages/product/list.vue')
const source = fs.readFileSync(pagePath, 'utf8')

assert(
  source.includes('<text v-if="!keyword" class="search-placeholder">搜索商品名 / 品牌</text>'),
  'product list search should render placeholder only when keyword is empty',
)

assert(
  !source.includes('placeholder="搜索商品名 / 品牌"'),
  'product list search should not use native input placeholder that overlaps on iOS H5',
)

assert(
  source.includes('pointer-events: none'),
  'custom placeholder should not block input focus',
)

console.log('product list search placeholder checks passed')
