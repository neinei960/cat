import { isFeedingPlanHistoryByDate } from '../src/utils/feeding'

function assertEqual(actual: boolean, expected: boolean, message: string) {
  if (actual !== expected) {
    throw new Error(`${message}: expected ${expected}, got ${actual}`)
  }
}

assertEqual(
  isFeedingPlanHistoryByDate({ end_date: '2026-04-30', status: 'active' }, '2026-05-01'),
  true,
  'active feeding plan should move to history after the final day has passed',
)

assertEqual(
  isFeedingPlanHistoryByDate({ end_date: '2026-05-01', status: 'completed' }, '2026-05-01'),
  false,
  'completed or paid feeding plan should not move to history before the final day has passed',
)

assertEqual(
  isFeedingPlanHistoryByDate({ end_date: '2026-05-02', status: 'cancelled' }, '2026-05-01'),
  false,
  'future feeding plan should not move to history because of status alone',
)
