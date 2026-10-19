import assert from 'node:assert/strict'
import fs from 'node:fs'
import path from 'node:path'

const pagePath = path.resolve(__dirname, '../../src/pages/appointment/detail.vue')
const source = fs.readFileSync(pagePath, 'utf8')

assert(
  source.includes('updateAppointmentNotes'),
  'appointment detail should use the dedicated notes update API',
)

assert(
  source.includes('v-model="notesDraft"') && source.includes('class="notes-textarea"'),
  'appointment detail notes card should render an editable textarea bound to notesDraft',
)

assert(
  source.includes('async function saveNotesEdit()'),
  'appointment detail should have an inline save handler for notes edits',
)

assert(
  source.includes('保存备注'),
  'appointment detail should expose a direct save action for notes',
)

console.log('appointment detail notes edit checks passed')
