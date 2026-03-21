// 性格-颜色映射（颜色表示风险等级）
export const personalityColors: Record<string, string> = {
  '神仙宝贝': '#10B981',
  '胆大开放': '#10B981',
  '胆小敏感': '#F59E0B',
  '过度活跃': '#F59E0B',
  '笑里藏刀': '#F97316',
  '绝世凶兽': '#EF4444',
}

// 颜色选项（新增自定义性格时选）
export const colorOptions = [
  { name: '安全(绿)', value: '#10B981' },
  { name: '注意(黄)', value: '#F59E0B' },
  { name: '警惕(橙)', value: '#F97316' },
  { name: '危险(红)', value: '#EF4444' },
]

export function getPersonalityColor(name: string): string {
  return personalityColors[name] || '#6B7280'
}

// 根据颜色生成浅色背景
export function getPersonalityBg(name: string): string {
  const color = getPersonalityColor(name)
  const map: Record<string, string> = {
    '#10B981': '#D1FAE5',
    '#F59E0B': '#FEF3C7',
    '#F97316': '#FFEDD5',
    '#EF4444': '#FEE2E2',
    '#6B7280': '#F3F4F6',
  }
  return map[color] || '#F3F4F6'
}
