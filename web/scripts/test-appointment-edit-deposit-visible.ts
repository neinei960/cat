import { readFileSync } from 'fs'
import { resolve } from 'path'
import assert from 'assert'

const source = readFileSync(resolve(process.cwd(), 'src/pages/appointment/create.vue'), 'utf8')

const stepsIndex = source.indexOf('<!-- Step indicator -->')
const editDepositIndex = source.indexOf('class="card edit-deposit-card"')
const stepOneIndex = source.indexOf('<!-- Step 1: Customer & Pet -->')

assert(editDepositIndex > stepsIndex, 'edit deposit card should render after the step indicator')
assert(editDepositIndex < stepOneIndex, 'edit deposit card should render before step content so it is visible while editing')
assert(source.includes('添加/修改预约金'), 'edit deposit card should clearly say it can add or modify the deposit')
assert(source.includes('deposit: form.value.deposit'), 'appointment update payload should include deposit')

console.log('appointment edit deposit visibility checks passed')
