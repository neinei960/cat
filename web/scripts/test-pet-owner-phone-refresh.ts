import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import { reconcileOwnerPhone } from '../src/utils/pet-owner-phone'

assert.deepEqual(reconcileOwnerPhone('old', 'old', 'new'), { phone: 'new', conflict: false })
assert.deepEqual(reconcileOwnerPhone('old', 'draft', 'old'), { phone: 'draft', conflict: false })
assert.deepEqual(reconcileOwnerPhone('old', 'draft', 'draft'), { phone: 'draft', conflict: false })
assert.deepEqual(reconcileOwnerPhone('old', 'draft', 'new'), { phone: 'draft', conflict: true })
assert.deepEqual(reconcileOwnerPhone('old', '', 'new'), { phone: '', conflict: true })
assert.deepEqual(reconcileOwnerPhone('', '', 'new'), { phone: 'new', conflict: false })
assert.deepEqual(reconcileOwnerPhone('old', ' old ', 'new'), { phone: 'new', conflict: false })

const page = readFileSync('src/pages/pet/edit.vue', 'utf8')
assert.match(page, /onShow\([\s\S]*?refreshOwnerPhone\(/)
assert.match(page, /async function onSubmit\([\s\S]*?await confirmOwnerPhoneBeforeSave\(/)
assert.match(page, /async function persistAvatarAfterUpload\([\s\S]*?await confirmOwnerPhoneBeforeSave\(/)
console.log('pet owner phone refresh tests passed')
