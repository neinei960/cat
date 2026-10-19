import fs from 'node:fs'
import path from 'node:path'

function main() {
  const root = process.cwd()
  const modalSource = fs.readFileSync(path.join(root, 'src/components/order/OrderCareReportModal.vue'), 'utf8')

  assert(modalSource.includes('const previewImageUrl = computed('), 'modal should normalize the generated preview image url')
  assert(modalSource.includes(':src="previewImageUrl"'), 'preview image should render the normalized preview url')
  assert(modalSource.includes('function resolveAbsoluteUrl(value: string)'), 'modal should share an absolute url helper for generated images')
}

main()

function assert(condition: unknown, message: string) {
  if (!condition) {
    throw new Error(message)
  }
}
