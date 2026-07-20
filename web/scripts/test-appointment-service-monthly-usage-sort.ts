import { readFileSync } from 'fs'
import { resolve } from 'path'
import assert from 'assert'

const source = readFileSync(resolve(process.cwd(), 'src/pages/appointment/create.vue'), 'utf8')

assert(
  source.includes('function compareServicesByMonthlyUsage'),
  'appointment service picker should use a named monthly usage comparator',
)
assert(
  source.includes('monthly_usage_count'),
  'appointment service picker should sort services by monthly_usage_count from /b/services?order_by=monthly_usage',
)
assert(
  !source.includes('serviceRankingMap.value[b.name]'),
  'appointment service picker should not sort by service name ranking, which can mix duplicate service names',
)

console.log('appointment service monthly usage sort checks passed')
