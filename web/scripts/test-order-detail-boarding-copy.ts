import assert from 'node:assert/strict'
import fs from 'node:fs'
import path from 'node:path'

const filePath = path.resolve(__dirname, '../../src/pages/order/detail.vue')
const source = fs.readFileSync(filePath, 'utf8')

assert(
  source.includes("if (order.value?.order_kind !== 'boarding') {"),
  'boarding header pet list should skip petGroups-derived room names',
)

assert(
  source.includes("pet_name: getReceiptGroupName(group.pet_name, order.value?.order_kind)"),
  'boarding detail groups should reuse normalized group names',
)

assert(
  source.includes("name: getReceiptItemDisplayName(item.name, group.pet_name === '零售商品', retailNamePrefixes.value, order.value?.order_kind)"),
  'boarding detail items should reuse normalized item names with order kind context',
)

console.log('order-detail boarding copy tests passed')
