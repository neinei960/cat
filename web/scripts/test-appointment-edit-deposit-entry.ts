import { readFileSync } from 'fs'
import { resolve } from 'path'

function assert(condition: unknown, message: string) {
  if (!condition) throw new Error(message)
}

const page = readFileSync(resolve(process.cwd(), 'src/pages/appointment/create.vue'), 'utf8')

assert(
  page.includes('v-if="isEditMode && step < 4"'),
  'edit appointment flow should expose a deposit editor before the final confirm step',
)

assert(
  page.includes('编辑预约金') && page.includes('edit-deposit-card'),
  'edit deposit editor should have a dedicated card title and class',
)

const depositInputCount = (page.match(/@input="onDepositInput"/g) || []).length
assert(
  depositInputCount >= 2,
  'edit deposit editor should reuse the existing deposit input handler outside the final confirm card',
)

const payloadDepositIndex = page.indexOf('deposit: form.value.deposit')
const updateAppointmentIndex = page.indexOf('await updateAppointment(editAppointmentId.value, payload)')
assert(payloadDepositIndex >= 0, 'submit payload should include deposit')
assert(
  payloadDepositIndex < updateAppointmentIndex,
  'deposit must be included before updateAppointment submits the edit payload',
)

console.log('appointment edit deposit entry assertions passed')
