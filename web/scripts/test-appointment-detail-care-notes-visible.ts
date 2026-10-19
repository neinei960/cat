import { readFileSync } from 'fs'
import { resolve } from 'path'

const source = readFileSync(resolve(process.cwd(), 'src/pages/appointment/detail.vue'), 'utf8')

function assertContains(text: string, message: string) {
  if (!source.includes(text)) {
    throw new Error(message)
  }
}

assertContains(
  'v-if="petCareNoteItems.length"',
  'appointment detail should render pet care notes when the pet profile has them'
)
assertContains(
  'pet?.care_notes',
  'appointment detail should read care notes from appointment pet profile data'
)

console.log('appointment detail care notes visibility checks passed')
