import { buildStaffCommissionExportHtml } from '../src/utils/staff-commission-export'

function assertContains(source: string, expected: string, label: string) {
  if (!source.includes(expected)) {
    throw new Error(`${label}: missing ${expected}`)
  }
}

function assertOccurrenceCount(source: string, expected: string, count: number, label: string) {
  const actual = source.split(expected).length - 1
  if (actual !== count) {
    throw new Error(`${label}: expected ${count} occurrences of ${expected}, got ${actual}`)
  }
}

const html = buildStaffCommissionExportHtml([
  {
    date: '2026-06-21',
    pay_method_label: '余额+微信',
    pay_amount: 198,
    formula: '¥198.00 × 20% = ¥39.60',
    pet_summary: '',
    items: [
      { pet_name: '奶糖', gender: '母', breed: '英短乳白', project: '深层清洁护理', amount: 108 },
      { pet_name: '饭团', gender: '公', breed: '英短串串', project: '去结去废毛', amount: 90 },
    ],
  },
])

for (const header of ['日期', '猫咪名字', '性别', '品种', '项目', '金额', '总计', '付款渠道', '提成计算']) {
  assertContains(html, `<th>${header}</th>`, `header.${header}`)
}
assertContains(html, '<td>奶糖</td>', 'first.pet')
assertContains(html, '<td>深层清洁护理</td>', 'first.project')
assertContains(html, '<td>108.00</td>', 'first.amount')
assertContains(html, '<td>198.00</td>', 'order.total')
assertContains(html, '<td>余额+微信</td>', 'payment.label')
assertContains(html, '<td>¥198.00 × 20% = ¥39.60</td>', 'commission.formula')
assertContains(html, '<td>饭团</td>', 'second.pet')
assertOccurrenceCount(html, '<td>198.00</td>', 1, 'order.total.once')
assertOccurrenceCount(html, '<td>余额+微信</td>', 1, 'payment.label.once')
assertOccurrenceCount(html, '<td>¥198.00 × 20% = ¥39.60</td>', 1, 'commission.formula.once')

const escaped = buildStaffCommissionExportHtml([
  {
    date: '2026-06-22',
    pay_method_label: '微信',
    pay_amount: 88,
    formula: '¥88.00 × 20% = ¥17.60',
    pet_summary: '咪咪',
    items: [{ pet_name: '咪<咪', gender: '母', breed: '英短 & 布偶', project: '洗护 > 基础', amount: 88 }],
  },
])
assertContains(escaped, '<td>咪&lt;咪</td>', 'escape.pet')
assertContains(escaped, '<td>英短 &amp; 布偶</td>', 'escape.breed')
assertContains(escaped, '<td>洗护 &gt; 基础</td>', 'escape.project')

console.log('staff commission export tests passed')
