export interface StaffCommissionExportItem {
  pet_name: string
  gender: string
  breed: string
  project: string
  amount: number
}

export interface StaffCommissionExportDetail {
  date: string
  pay_method_label: string
  pay_amount: number
  formula: string
  pet_summary?: string
  items?: StaffCommissionExportItem[]
}

const headers = ['日期', '猫咪名字', '性别', '品种', '项目', '金额', '总计', '付款渠道', '提成计算']

function escapeCell(value: unknown) {
  return String(value ?? '')
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
}

function money(value: unknown) {
  return Number(value || 0).toFixed(2)
}

export function buildStaffCommissionExportHtml(details: StaffCommissionExportDetail[]) {
  const bodyRows = details.flatMap((detail) => {
    const items = detail.items?.length
      ? detail.items
      : [{
          pet_name: detail.pet_summary || '',
          gender: '',
          breed: '',
          project: '订单',
          amount: detail.pay_amount,
        }]
    return items.map((item, index) => [
      detail.date,
      item.pet_name || '-',
      item.gender || '-',
      item.breed || '-',
      item.project || '-',
      money(item.amount),
      index === 0 ? money(detail.pay_amount) : '',
      index === 0 ? detail.pay_method_label || '-' : '',
      index === 0 ? detail.formula || '-' : '',
    ])
  })

  const head = headers.map((header) => `<th>${header}</th>`).join('')
  const body = bodyRows
    .map((row) => `<tr>${row.map((cell) => `<td>${escapeCell(cell)}</td>`).join('')}</tr>`)
    .join('')

  return `<!DOCTYPE html><html><head><meta charset="UTF-8"><style>
table{border-collapse:collapse;font-family:Arial,"Microsoft YaHei",sans-serif;font-size:12px}
th,td{border:1px solid #b7b7b7;padding:6px 10px;white-space:nowrap}
th{background:#f2f2f2;font-weight:700}
</style></head><body><table><thead><tr>${head}</tr></thead><tbody>${body}</tbody></table></body></html>`
}
