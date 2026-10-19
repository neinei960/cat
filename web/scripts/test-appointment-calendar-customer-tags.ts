import fs from 'node:fs'
import path from 'node:path'

const root = process.cwd()
const calendarPath = path.join(root, 'src/pages/appointment/calendar.vue')
const repoPath = path.resolve(root, '../server/internal/repository/appointment_repo.go')

const calendar = fs.readFileSync(calendarPath, 'utf8')
const repo = fs.readFileSync(repoPath, 'utf8')

function assert(condition: unknown, message: string) {
  if (!condition) {
    throw new Error(message)
  }
}

assert(
  repo.includes('Preload("Customer.CustomerTags"'),
  'appointment repository should preload customer tags for calendar cards'
)

assert(
  calendar.includes('getCustomerTagItems(appt)'),
  'calendar cards should derive owner/customer tag items'
)

assert(
  calendar.includes('tag-owner'),
  'calendar should render owner tag chips with a distinct owner class'
)

assert(
  calendar.includes('ownerTags') && calendar.includes('petTags'),
  'calendar should keep owner tags and pet tags in separate data buckets'
)

assert(
  calendar.includes('appt-owner-tag-row'),
  'calendar should render owner tags on their own row'
)

assert(
  calendar.includes('.appt-tag.tag-owner'),
  'calendar should style owner tags separately from pet tags'
)

console.log('appointment calendar customer tag checks passed')
