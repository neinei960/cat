import fs from 'node:fs'
import path from 'node:path'

function main() {
  const root = process.cwd()
  const modalSource = fs.readFileSync(path.join(root, 'src/components/order/OrderCareReportModal.vue'), 'utf8')
  const stageSource = fs.readFileSync(path.join(root, 'src/components/order/OrderCareReportStage.vue'), 'utf8')

  assert(!modalSource.includes('care-report-field-grid'), 'modal should remove the old field grid form')
  assert(!modalSource.includes('care-report-section-card'), 'modal should remove the old section cards')
  assert(modalSource.includes('care-report-editor-dock'), 'modal should render the contextual editor in a fixed dock')
  assert(modalSource.includes('care-report-editor-sheet'), 'modal should render a single contextual editor sheet')
  assert(modalSource.includes('@edit-target="openEditor"'), 'modal should open the editor from stage interactions')
  assert(stageSource.includes('care-report-stage-overlay'), 'stage should expose a non-exported interaction overlay')
  assert(stageSource.includes('care-report-stage-hotspot'), 'stage should define clickable hotspots over the template')
  assert(stageSource.includes("emit('edit-target'"), 'stage should emit edit-target events for writable fields')
  assert(stageSource.includes("emit('toggle-section-check'"), 'stage should emit section toggle events directly from the stage')
}

main()

function assert(condition: unknown, label: string) {
  if (!condition) {
    throw new Error(label)
  }
}
