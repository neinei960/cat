import assert from 'node:assert/strict'
import fs from 'node:fs'
import path from 'node:path'

const filePath = path.resolve(__dirname, '../../src/pages/customer/detail.vue')
const source = fs.readFileSync(filePath, 'utf8')

assert(
  source.includes('records.value.slice(0, 1)'),
  'collapsed customer detail records should keep only the latest record visible',
)

assert(
  !source.includes('records.value.slice(0, 3)'),
  'customer detail records should no longer default to three visible rows',
)

assert(
  source.includes('records-arrow'),
  'customer detail records should render an arrow-style toggle control',
)

console.log('customer detail record collapse regression test passed')
