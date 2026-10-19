import assert from 'node:assert/strict'
import fs from 'node:fs'
import path from 'node:path'

const root = process.cwd()
const pageSource = fs.readFileSync(path.join(root, 'src/pages/appointment/detail.vue'), 'utf8')
const modelSource = fs.readFileSync(path.resolve(root, '../server/internal/model/service_record.go'), 'utf8')
const routerSource = fs.readFileSync(path.resolve(root, '../server/internal/router/router.go'), 'utf8')
const repoSource = fs.readFileSync(path.resolve(root, '../server/internal/repository/service_record_repo.go'), 'utf8')

assert(
  !pageSource.includes('使用了什么浴液、剃了哪个部位、发现什么问题'),
  'care record form should remove the free-form service record textarea',
)

assert(
  pageSource.includes('牙齿状况') && pageSource.includes('recordForm.dental_condition'),
  'care record form should include a dental condition field',
)

assert(
  pageSource.includes('其他问题') && pageSource.includes('recordForm.other_issues') && pageSource.includes('form-input-xs'),
  'care record form should include a smaller other issues field',
)

assert(
  pageSource.includes('order_item_summary') && pageSource.includes('消费项目'),
  'care record list should display prior order consumption items',
)

assert(
  modelSource.includes('DentalCondition') && modelSource.includes('OtherIssues') && modelSource.includes('OrderItemSummary'),
  'service record model should expose dental, other issues, and order item summary fields',
)

assert(
  routerSource.includes('DentalCondition') && routerSource.includes('OtherIssues'),
  'service record routes should bind dental and other issue fields',
)

assert(
  repoSource.includes('attachOrderItemSummaries') && repoSource.includes('ItemType == 1') && repoSource.includes('ItemType == 3'),
  'service record repository should attach service/add-on order item summaries for appointment records',
)

console.log('appointment service record care field checks passed')
