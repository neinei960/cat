import fs from 'node:fs'
import path from 'node:path'

const source = fs.readFileSync(path.join(process.cwd(), 'src/pages/appointment/create.vue'), 'utf8')

if (!/async function onSubmit\(\) \{\s*if \(submitting\.value\) return\s*submitting\.value = true/.test(source)) {
  throw new Error('appointment submit must ignore repeated clicks before starting another request')
}

const submitButtons = source.match(/<button[^>]*@click="onSubmit"[^>]*>/g) || []
if (submitButtons.length === 0) {
  throw new Error('appointment page should expose submit buttons')
}
if (submitButtons.some(button => !button.includes(':disabled="submitting"'))) {
  throw new Error('every appointment submit button must be disabled while submitting')
}
