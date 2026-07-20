import { readFileSync } from 'fs'
import { resolve } from 'path'
import assert from 'assert'

const source = readFileSync(resolve(process.cwd(), 'src/pages/order/create.vue'), 'utf8')

assert(
  source.includes('const serviceItems = ref<ServiceCartItem[]>([])'),
  'order create should keep selected services in an array so one order can include multiple services',
)

assert(
  source.includes('function dedupeServicePriceRules'),
  'order create should dedupe service price rules before rendering specs',
)

assert(
  source.includes('for (const item of serviceItems.value)'),
  'order create submit should send every selected service as an order item',
)

assert(
  !source.includes('const serviceSubtotal = computed(() => selectedServiceId.value > 0 ? roundCurrency(servicePrice.value) : 0)'),
  'order create service subtotal should not depend on a single selected service id',
)

console.log('order create multi-service selection checks passed')
