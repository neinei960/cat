import assert from 'node:assert/strict'
import { canRepeatService, serviceQuantity, serviceLineAmount, serviceSelectionKey } from '../src/utils/service-quantity'

assert.equal(serviceQuantity(undefined), 1)
assert.equal(serviceLineAmount(20, 2), 40)
assert.equal(serviceLineAmount(19.9, 3), 59.7)
assert.equal(canRepeatService('刷牙 · 全年口腔护理'), false)
assert.equal(canRepeatService('刷牙 · 已购口腔护理年卡'), true)
assert.notEqual(serviceSelectionKey(1, 1, '自带牙刷', 20), serviceSelectionKey(1, 2, '非自带', 30))
assert.notEqual(serviceSelectionKey(1, 0, '自定义', 20), serviceSelectionKey(1, 0, '自定义', 30))
console.log('service quantity checks passed')
