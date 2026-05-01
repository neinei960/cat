import assert from 'node:assert/strict'
import fs from 'node:fs'
import path from 'node:path'

const createFilePath = path.resolve(__dirname, '../../src/pages/boarding/create.vue')
const detailFilePath = path.resolve(__dirname, '../../src/pages/boarding/detail.vue')

const createSource = fs.readFileSync(createFilePath, 'utf8')
const detailSource = fs.readFileSync(detailFilePath, 'utf8')

assert(
  createSource.includes('暂不填写'),
  'boarding create page should expose an explicit optional deworming choice',
)

assert(
  !createSource.includes("title: '请选择是否已驱虫'"),
  'boarding create page should not block submission when deworming is unset',
)

assert(
  createSource.includes('hasDeworming: null as boolean | null'),
  'boarding create page should preserve a nullable deworming state',
)

assert(
  createSource.includes('has_deworming: form.value.hasDeworming'),
  'boarding create submit payload should keep sending nullable deworming data',
)

assert(
  detailSource.includes("return '未填写'"),
  'boarding detail should render a neutral label for unset deworming state',
)

console.log('boarding deworming optional regression test passed')
