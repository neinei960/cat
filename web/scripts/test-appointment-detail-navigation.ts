import { readFileSync } from 'fs'
import { resolve } from 'path'

const file = readFileSync(resolve(process.cwd(), 'src/pages/appointment/detail.vue'), 'utf8')

function assertContains(value: string, message: string) {
  if (!file.includes(value)) {
    throw new Error(message)
  }
}

assertContains('goCustomerDetail', 'appointment detail should expose customer detail navigation')
assertContains('/pages/customer/detail?id=', 'customer navigation should target customer detail')
assertContains('goPetDetail', 'appointment detail should expose pet detail navigation')
assertContains('/pages/pet/edit?id=', 'pet navigation should target pet edit/detail page')
assertContains('link-value', 'linked appointment fields should use the shared clickable style')

console.log('appointment detail navigation checks passed')
