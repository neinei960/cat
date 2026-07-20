import assert from 'node:assert/strict'
import fs from 'node:fs'
import path from 'node:path'

const pagePath = path.resolve(__dirname, '../../src/pages/order/create.vue')
const apiPath = path.resolve(__dirname, '../../src/api/order.ts')
const pageSource = fs.readFileSync(pagePath, 'utf8')
const apiSource = fs.readFileSync(apiPath, 'utf8')

assert(
  apiSource.includes('customer_phone?: string'),
  'order create API payload should accept customer_phone for guest retail orders',
)

assert(
  pageSource.includes('function normalizePhoneInput'),
  'order create should normalize typed phone input before matching',
)

assert(
  pageSource.includes('async function resolveCustomerByPhoneBeforeSubmit'),
  'order create should resolve an existing customer by phone before submit',
)

assert(
  pageSource.includes('await resolveCustomerByPhoneBeforeSubmit(typedCustomerPhone)'),
  'order create submit should await phone auto-binding before building payload',
)

assert(
  pageSource.includes('customer_phone: customerId ? undefined : typedCustomerPhone || undefined'),
  'order create payload should send customer_phone only when no customer is selected',
)

console.log('order create customer phone auto-bind checks passed')
