import fs from 'node:fs'
import path from 'node:path'

function main() {
  const root = process.cwd()
  const modalSource = fs.readFileSync(path.join(root, 'src/components/order/OrderCareReportModal.vue'), 'utf8')

  assert(modalSource.includes('care-report-basic-section'), 'modal should render a standalone basic-information form')
  assert(modalSource.includes('orderCareReportBodyShapeOptions'), 'modal should reuse shared body-shape options')
  assert(modalSource.includes('v-for="section in sectionDefinitions"'), 'modal should render all inspection sections from shared definitions')
  assert(modalSource.includes('expandedSectionKey'), 'inspection sections should use a single accordion state')
  assert(modalSource.includes('care-report-section-note'), 'each inspection section should expose its own note input')
  assert(modalSource.includes('const previewExpanded = ref(false)'), 'draft preview should be closed by default')
  assert(modalSource.includes('v-if="previewExpanded"'), 'draft preview should render only on demand')
  assert(modalSource.includes('scrollIntoView'), 'validation should bring the first missing field into view')
  assert(modalSource.includes("replace(/\\./g, '-')"), 'picker dates should only replace dot separators')
  assert(modalSource.includes('<OrderCareReportStage :draft="draft" />'), 'draft preview should render the current form data')
  assert(
    modalSource.includes('<view class="care-report-actions care-report-draft-preview-actions">'),
    'draft preview should expose a dedicated bottom action bar'
  )
  assert(
    modalSource.includes('<view class="care-report-btn primary" @click="submit">{{ submitting ? \'生成中...\' : \'生成报告\' }}</view>'),
    'draft preview should allow users to generate the confirmed report directly'
  )
  assert(!modalSource.includes('@edit-target="openEditor"'), 'draft preview should not open image hotspot editors')
  assert(!modalSource.includes('care-report-editor-dock'), 'contextual image editor should be removed')
  assert(!modalSource.includes(':active-editor-key='), 'modal should not pass image editing state')
  assert(!modalSource.includes('<OrderCareReportStage\n              :draft="draft"\n              editable'), 'draft preview should not enable stage editing')
}

main()

function assert(condition: unknown, label: string) {
  if (!condition) {
    throw new Error(label)
  }
}
