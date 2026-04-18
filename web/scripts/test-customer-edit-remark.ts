import assert from 'node:assert/strict'
import fs from 'node:fs'
import path from 'node:path'

const filePath = path.resolve(__dirname, '../../src/pages/customer/edit.vue')
const source = fs.readFileSync(filePath, 'utf8')

assert(
  !source.includes('contenteditable="plaintext-only"'),
  'customer edit remark field should not rely on H5 contenteditable="plaintext-only"; use native textarea for mobile browser editing',
)

assert(
  source.includes('<textarea v-model="form.remark" placeholder="添加文字" class="field-textarea" :auto-height="false" />'),
  'customer edit remark field should render the shared textarea control',
)

console.log('customer edit remark regression test passed')
