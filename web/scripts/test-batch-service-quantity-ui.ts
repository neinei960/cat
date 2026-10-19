import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'

const page = readFileSync('src/pages/order/batch-create.vue', 'utf8')
assert.match(page, /class="qty-ctrl service-quantity-controls"/, 'service and product counters must share the compact container')
assert.match(page, /class="qty-btn service-quantity-button"/, 'service buttons must use the product counter dimensions')
assert.match(page, /class="qty-val"[^>]*>\{\{ svc.quantity \}\}/)
assert.match(page, /:disabled="svc.quantity <= 1"/)
assert.match(page, /:disabled="!canRepeatService\(svc.service_name\) \|\| svc.quantity >= 999"/)
assert.match(page, /class="service-line-total"/, 'line subtotal must remain visible outside the counter')
assert.doesNotMatch(page, /\.service-quantity-button\s*\{[^}]*min-height: 44px/)
console.log('compact service quantity UI checks passed')
