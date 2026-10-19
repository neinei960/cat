import { readFileSync } from 'fs'
import { resolve } from 'path'
import assert from 'assert'

const repoRoot = resolve(process.cwd(), '..')
const pageSource = readFileSync(resolve(repoRoot, 'web/src/pages/appointment/detail.vue'), 'utf8')
const routerSource = readFileSync(resolve(repoRoot, 'server/internal/router/router.go'), 'utf8')

assert(
  pageSource.includes('editingRecordId'),
  'appointment detail should keep editingRecordId state for service record editing',
)

assert(
  pageSource.includes('function openRecordEdit'),
  'appointment detail should allow opening an existing service record for edit',
)

assert(
  pageSource.includes("method: editingRecordId.value ? 'PUT' : 'POST'"),
  'appointment detail should submit edits with PUT and new records with POST',
)

assert(
  routerSource.includes('b.PUT("/service-records/:id"'),
  'server should expose PUT /service-records/:id for editing service records',
)

console.log('appointment detail service record edit checks passed')
