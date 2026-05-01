import fs from 'node:fs'
import path from 'node:path'

function main() {
  const root = process.cwd()
  const source = fs.readFileSync(path.join(root, 'src/components/order/OrderCareReportStage.vue'), 'utf8')

  assert(source.includes('createNoteOverlayStyle('), 'stage should define a dedicated note hotspot helper')
  assert(source.includes("style: createNoteOverlayStyle(spec.x - 26, spec.y - 14, spec.width + 44, 54)"), 'note hotspots should use the dedicated helper instead of the generic field overlay')
  assert(source.includes('Math.max(28,'), 'note hotspots should enforce a mobile-sized minimum tap height')
  assert(source.includes("zIndex: '4'"), 'note hotspots should stay above nearby checkbox hotspots')
}

main()

function assert(condition: unknown, label: string) {
  if (!condition) {
    throw new Error(label)
  }
}
