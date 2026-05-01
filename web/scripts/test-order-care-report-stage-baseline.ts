import fs from 'node:fs'
import path from 'node:path'

function main() {
  const root = process.cwd()
  const source = fs.readFileSync(path.join(root, 'src/components/order/OrderCareReportStage.vue'), 'utf8')

  assert(source.includes('const UNDERLINE_GAP = 8'), 'stage should define a shared underline gap')
  assert(source.includes("display: 'flex'"), 'stage text styles should use flex anchoring')
  assert(source.includes("alignItems: 'flex-end'"), 'stage text styles should align to the underline from the bottom')
  assert(source.includes('paddingBottom: UNDERLINE_GAP'), 'stage text styles should keep an 8px gap above the underline')
  assert(source.includes('fitCenteredFontSize('), 'stage should shrink long centered text to keep it visually centered on the underline')
  assert(source.includes("overflow: 'hidden'"), 'stage should clip overflow after centered text has been fit to the underline width')
  assert(source.includes('const NOTE_LINE_HEIGHT = 32'), 'stage note styles should reserve a fixed note line height')
}

main()

function assert(condition: unknown, label: string) {
  if (!condition) {
    throw new Error(label)
  }
}
