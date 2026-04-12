import { buildAppointmentCalendarSummary } from '../src/utils/appointment-calendar-summary'

function assertEqual(actual: unknown, expected: unknown, label: string) {
  if (actual !== expected) {
    throw new Error(`${label}: expected "${String(expected)}", got "${String(actual)}"`)
  }
}

const summary = buildAppointmentCalendarSummary([
  { ID: 1, status: 0, staff_id: 0 },
  { ID: 2, status: 1, staff_id: 8 },
  { ID: 3, status: 3, staff_id: 0 },
  { ID: 4, status: 4, staff_id: 0 },
])

assertEqual(summary.total, 3, 'summary.total')
assertEqual(summary.pendingConfirm, 1, 'summary.pendingConfirm')
assertEqual(summary.unassigned, 2, 'summary.unassigned')
assertEqual(summary.waitingCheckout, 1, 'summary.waitingCheckout')

console.log('appointment calendar summary tests passed')
