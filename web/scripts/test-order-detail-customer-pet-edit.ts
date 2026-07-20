import assert from 'node:assert/strict'
import fs from 'node:fs'
import path from 'node:path'

const root = process.cwd()
const pageSource = fs.readFileSync(path.join(root, 'src/pages/order/detail.vue'), 'utf8')
const apiSource = fs.readFileSync(path.join(root, 'src/api/order.ts'), 'utf8')

assert(
  apiSource.includes('updateOrderCustomerPet'),
  'order API should expose a dedicated customer/pet update call',
)

assert(
  apiSource.includes('/customer-pet'),
  'customer/pet update should call the dedicated order endpoint',
)

assert(
  pageSource.includes('showCustomerPetModal'),
  'order detail should track customer/pet edit modal state',
)

assert(
  pageSource.includes('修改客户/猫咪'),
  'order detail should provide a visible customer/pet edit action',
)

assert(
  pageSource.includes('saveCustomerPet'),
  'order detail should save customer/pet edits without opening the full order editor',
)

assert(
  pageSource.includes('updateOrderCustomerPet(order.value.ID'),
  'order detail should persist customer/pet edits through the dedicated API',
)

console.log('order detail customer/pet edit checks passed')
