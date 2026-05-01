import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'

const filePath = resolve(new URL('.', import.meta.url).pathname, '../src/components/order/OrderCareReportModal.vue')
const source = readFileSync(filePath, 'utf8')

assertNotContains(source, '导出图片会以这张预览为准', 'stage preview tip should be removed')
assertMatches(source, /\.care-report-input\s*\{[\s\S]*?display:\s*flex;[\s\S]*?min-height:\s*\d+rpx;/, 'input shell should define a real flex height')
assertContains(source, '.care-report-input :deep(.uni-input-wrapper)', 'missing deep wrapper selector')
assertContains(source, '.care-report-input :deep(.uni-input-input)', 'missing deep input selector')
assertContains(source, '.care-report-input :deep(.uni-input-placeholder)', 'missing deep placeholder selector')
assertMatches(source, /\.care-report-input\s*:deep\(\.uni-input-wrapper\)[\s\S]*?(min-height:\s*100%|height:\s*100%)/, 'wrapper selector should fill the input shell height')
assertMatches(source, /\.care-report-input\s*:deep\(\.uni-input-input\)[\s\S]*?(min-height|height):\s*\d+rpx;/, 'input selector should define height')
assertMatches(source, /\.care-report-input\s*:deep\(\.uni-input-input\)[\s\S]*?line-height:\s*\d+rpx;/, 'input selector should define line-height')

function assertContains(content: string, token: string, message: string) {
  if (!content.includes(token)) {
    throw new Error(message)
  }
}

function assertNotContains(content: string, token: string, message: string) {
  if (content.includes(token)) {
    throw new Error(message)
  }
}

function assertMatches(content: string, pattern: RegExp, message: string) {
  if (!pattern.test(content)) {
    throw new Error(message)
  }
}
