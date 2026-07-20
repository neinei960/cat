import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'

function assert(condition: unknown, message: string) {
  if (!condition) {
    throw new Error(message)
  }
}

const pagePath = resolve(process.cwd(), 'src/pages/order/batch-create.vue')
const source = readFileSync(pagePath, 'utf8')

assert(source.includes("import { getStaffList } from '@/api/staff'"), 'batch order edit should load staff options')
assert(source.includes('const selectedStaffIdx = ref(0)'), 'batch order edit should track selected staff index')
assert(source.includes('const selectedStaff = computed'), 'batch order edit should expose selected staff')
assert(source.includes('<picker :range="staffNames"'), 'batch order edit should render a staff picker')
assert(source.includes('staff_id: selectedStaff.value?.ID'), 'batch order save should submit the currently selected staff')
assert(!source.includes('staff_id: existingOrder.value?.staff_id || appt.value?.staff_id'), 'batch order save must not keep the old staff_id when user changes staff')

console.log('order batch staff edit static checks passed')
