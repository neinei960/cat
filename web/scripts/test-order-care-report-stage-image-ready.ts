import fs from 'node:fs'
import path from 'node:path'

function main() {
  const root = process.cwd()
  const stageSource = fs.readFileSync(path.join(root, 'src/components/order/OrderCareReportStage.vue'), 'utf8')

  assert(stageSource.includes('await waitForStageImagesReady(element)'), 'stage export should wait for embedded images before html2canvas')
  assert(stageSource.includes('async function waitForStageImagesReady'), 'stage should define an image readiness helper')
  assert(stageSource.includes("querySelectorAll('img')"), 'image readiness helper should inspect rendered img nodes inside the stage')
}

main()

function assert(condition: unknown, message: string) {
  if (!condition) {
    throw new Error(message)
  }
}
