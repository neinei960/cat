import assert from 'node:assert/strict'
import fs from 'node:fs'
import path from 'node:path'

const filePath = path.resolve(__dirname, '../../src/pages/customer/detail.vue')
const source = fs.readFileSync(filePath, 'utf8')

assert(
  source.includes('const openCardAmount = ref(\'\')'),
  'customer detail should use a dedicated amount state for opening a member card',
)

assert(
  source.includes('function openRechargeModal()'),
  'customer detail should expose an explicit recharge modal opener',
)

assert(
  source.includes('function closeRechargeModal()'),
  'customer detail should expose an explicit recharge modal closer',
)

assert(
  source.includes('openCardAmount.value') && !source.includes('parseFloat(rechargeAmount.value) < tpl.min_recharge'),
  'member card threshold validation should only read the open-card amount state',
)

assert(
  source.includes('rechargeAmount.value = \'\''),
  'recharge modal flow should clear stale recharge amount state',
)

console.log('customer member recharge state regression test passed')
