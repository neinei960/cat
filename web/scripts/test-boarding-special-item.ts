import assert from 'node:assert/strict'
import fs from 'node:fs'
import path from 'node:path'

const createFilePath = path.resolve(__dirname, '../../src/pages/boarding/create.vue')
const detailFilePath = path.resolve(__dirname, '../../src/pages/boarding/detail.vue')
const apiFilePath = path.resolve(__dirname, '../../src/api/boarding.ts')
const typesFilePath = path.resolve(__dirname, '../../src/types/index.d.ts')
const dashboardFilePath = path.resolve(__dirname, '../../src/pages/boarding/dashboard.vue')

const createSource = fs.readFileSync(createFilePath, 'utf8')
const detailSource = fs.readFileSync(detailFilePath, 'utf8')
const apiSource = fs.readFileSync(apiFilePath, 'utf8')
const typesSource = fs.readFileSync(typesFilePath, 'utf8')
const dashboardSource = fs.readFileSync(dashboardFilePath, 'utf8')

assert(
  createSource.includes('特殊寄养项目'),
  'boarding create page should expose a special boarding item section',
)

assert(
  createSource.includes('special_item_id') && createSource.includes('special_item_daily_price') && createSource.includes('special_item_days'),
  'boarding create page should submit special item fields in room groups payloads',
)

assert(
  detailSource.includes('特殊寄养项目') && detailSource.includes('特殊天数'),
  'boarding detail sheet should expose special item and special day editing fields',
)

assert(
  apiSource.includes('/b/boarding/special-items'),
  'boarding API module should expose special item endpoints',
)

assert(
  typesSource.includes('interface BoardingSpecialItem'),
  'boarding types should include a BoardingSpecialItem interface',
)

assert(
  typesSource.includes('special_item_amount') && typesSource.includes('special_item_days'),
  'boarding types should include special item pricing fields for orders and previews',
)

assert(
  dashboardSource.includes("'/pages/boarding/special-items'"),
  'boarding dashboard should include a special item settings entry',
)

console.log('boarding special item regression test passed')
