import fs from 'node:fs'
import path from 'node:path'

function main() {
  const root = process.cwd()
  const modalSource = fs.readFileSync(path.join(root, 'src/components/order/OrderCareReportModal.vue'), 'utf8')
  const stageSource = fs.readFileSync(path.join(root, 'src/components/order/OrderCareReportStage.vue'), 'utf8')
  const utilSource = fs.readFileSync(path.join(root, 'src/utils/order-care-report.ts'), 'utf8')
  const apiSource = fs.readFileSync(path.join(root, 'src/api/order-care-report.ts'), 'utf8')

  assert(modalSource.includes("from '@/api/order-care-report'"), 'modal should import backend order care report api')
  assert(modalSource.includes('createOrderCareReport('), 'modal should call backend order care report api')
  assert(!modalSource.includes("from '@/api/pet-bath-report'"), 'modal should not import the pet bath report api')
  assert(!modalSource.includes('createPetBathReport('), 'modal should let the backend persist the pet bath report')
  assert(utilSource.includes('buildOrderCareReportPayload'), 'frontend should map the editor draft to the backend payload')
  assert(apiSource.includes("method: 'POST'"), 'order care report api should use POST')
  assert(apiSource.includes('/care-report'), 'order care report api should target the care report endpoint')
  assert(!stageSource.includes("from 'html2canvas'"), 'preview stage should not depend on browser screenshot rendering')
  assert(!stageSource.includes('exportPngBlob'), 'preview stage should not expose a browser screenshot exporter')
  assert(stageSource.includes("standard: { x: 732, y: 833 }"), 'body preview checkbox should use the real template center')
  assert(stageSource.includes("normal: { x: 406, y: 929 }"), 'skin preview checkbox should use the real template center')
  assert(stageSource.includes("dry: { x: 732, y: 1025 }"), 'hair preview checkbox should use the real template center')
  assert(stageSource.includes("cleaned: { x: 406, y: 1217 }"), 'eyes preview checkbox should use the real template center')
  assert(stageSource.includes("black_earwax: { x: 1058, y: 1313 }"), 'ears preview checkbox should use the real template center')
  assert(stageSource.includes("wound: { x: 406, y: 1361 }"), 'ears preview should include the second-row wound checkbox')
  assert(stageSource.includes("touch_sensitive: { x: 569, y: 1410 }"), 'oral preview should include the touch-sensitive checkbox')
  assert(stageSource.includes("tartar: { x: 732, y: 1410 }"), 'oral preview checkbox should use the real template center')
  assert(stageSource.includes("oral_ulcer: { x: 406, y: 1458 }"), 'oral preview should include the oral-ulcer checkbox')
  assert(stageSource.includes("bad_breath: { x: 569, y: 1458 }"), 'oral preview should include the bad-breath checkbox')
  assert(stageSource.includes("dental_abnormal: { x: 732, y: 1458 }"), 'oral preview should include the dental-abnormal checkbox')
  assert(stageSource.includes("normal: { x: 406, y: 1553 }"), 'anus preview checkbox should use the real template center')
  assert(stageSource.includes("prolapse: { x: 569, y: 1553 }"), 'anus preview should include the prolapse checkbox')
  assert(stageSource.includes("inflamed: { x: 895, y: 1553 }"), 'anus preview swelling should use the fourth printed checkbox')
  assert(!stageSource.includes("anal_gland_swollen: { x: 1058, y: 1553 }"), 'anus preview must not draw into a nonexistent fifth checkbox')
  assert(stageSource.includes('care-report-stage-label-override'), 'preview should replace the old last-care label')
  assert(stageSource.includes('护理内容') && stageSource.includes('Content of care'), 'preview should show the real care-content labels')
  assert(stageSource.includes('formatDisplayDate'), 'preview should use the real dotted date format')
  assert(stageSource.includes('fontWeight: 700'), 'preview primary fields should use bold text')
  assert(stageSource.includes('left: x - 11') && stageSource.includes('top: y - 15'), 'preview checkmark should use the authentic larger bounds')
  assert(stageSource.includes('border-right: 5px solid #111111'), 'preview checkmark should use a bold stroke')
  assert(stageSource.includes("whiteSpace: 'normal'") && stageSource.includes("lineHeight: '24px'"), 'preview notes should support two centered lines')
  assert(stageSource.includes("createCenteredField('pet_name', props.draft.petName, 279, 196"), 'pet name preview should sit above its underline')
  assert(stageSource.includes("createCenteredField('care_date', formatDisplayDate(props.draft.careDate), 279, 686"), 'care date preview should sit above its underline')
  assert(stageSource.includes("createCenteredField('next_care_date', formatDisplayDate(props.draft.nextCareDate), 905, 690, 229, 48, 34, 400)"), 'next care date preview should use the lighter real-template weight')
  assert(stageSource.includes("{ key: 'skin', x: 648, y: 941"), 'skin preview note should stay centered around its underline')
  assert(stageSource.includes("{ key: 'ears', x: 648, y: 1325"), 'ears preview note should stay centered around its underline')
  assert(stageSource.includes("{ key: 'oral', x: 494, y: 1469"), 'oral preview note should stay centered around its underline')
  assert(stageSource.includes("{ key: 'anus', x: 494, y: 1566"), 'anus preview note should stay centered around its underline')
  assert(stageSource.includes("height: NOTE_LINE_HEIGHT,\n        display: 'flex',\n        alignItems: 'center',\n        justifyContent: 'center',\n        textAlign: 'center'"), 'section note previews should be centered on one or two lines')
}

main()

function assert(condition: unknown, label: string) {
  if (!condition) {
    throw new Error(label)
  }
}
