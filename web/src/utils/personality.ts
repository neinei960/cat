export interface PersonalityTagConfig {
  order: string[]
  colors: Record<string, string>
}

const STORAGE_KEY = 'personality_tag_config'

export const defaultPersonalityColors: Record<string, string> = {
  '神仙宝贝': '#10B981',
  '胆大开放': '#14B8A6',
  '胆小敏感': '#F59E0B',
  '过度活跃': '#F97316',
  '笑里藏刀': '#EF4444',
  '绝世凶兽': '#BE123C',
}

export const defaultPersonalityOrder = Object.keys(defaultPersonalityColors)

export const colorOptions = [
  { name: '青绿', value: '#10B981' },
  { name: '湖绿', value: '#14B8A6' },
  { name: '天蓝', value: '#0EA5E9' },
  { name: '靛蓝', value: '#4F46E5' },
  { name: '紫色', value: '#8B5CF6' },
  { name: '粉色', value: '#EC4899' },
  { name: '黄色', value: '#F59E0B' },
  { name: '橙色', value: '#F97316' },
  { name: '红色', value: '#EF4444' },
  { name: '深红', value: '#BE123C' },
  { name: '灰蓝', value: '#64748B' },
  { name: '墨灰', value: '#334155' },
]

function isRecord(value: unknown): value is Record<string, any> {
  return !!value && typeof value === 'object' && !Array.isArray(value)
}

function isHexColor(value: unknown): value is string {
  return typeof value === 'string' && /^#([0-9a-fA-F]{6})$/.test(value)
}

function parseJSON(value: unknown): Record<string, any> {
  if (isRecord(value)) return { ...value }
  if (typeof value === 'string') {
    try {
      const parsed = JSON.parse(value)
      return isRecord(parsed) ? parsed : {}
    } catch {
      return {}
    }
  }
  return {}
}

function normalizeConfig(raw?: Partial<PersonalityTagConfig> | null): PersonalityTagConfig {
  const colors = { ...defaultPersonalityColors }
  const order = [...defaultPersonalityOrder]

  if (raw?.colors && isRecord(raw.colors)) {
    for (const [name, color] of Object.entries(raw.colors)) {
      if (!name || !isHexColor(color)) continue
      colors[name] = color
      if (!order.includes(name)) order.push(name)
    }
  }

  const rawOrder = raw?.order
  if (Array.isArray(rawOrder)) {
    const normalizedOrder = rawOrder.filter((name): name is string => typeof name === 'string' && name.trim().length > 0)
    const deduped = Array.from(new Set(normalizedOrder))
    const missing = Object.keys(colors).filter(name => !deduped.includes(name))
    return {
      order: [...deduped, ...missing],
      colors,
    }
  }

  return { order, colors }
}

function readStorageConfig(): PersonalityTagConfig {
  try {
    const raw = uni.getStorageSync(STORAGE_KEY)
    if (!raw) return normalizeConfig()
    if (typeof raw === 'string') {
      return normalizeConfig(JSON.parse(raw))
    }
    return normalizeConfig(raw)
  } catch {
    return normalizeConfig()
  }
}

export function savePersonalityConfig(config: PersonalityTagConfig) {
  const normalized = normalizeConfig(config)
  uni.setStorageSync(STORAGE_KEY, JSON.stringify(normalized))
  return normalized
}

export function getPersonalityConfig(): PersonalityTagConfig {
  return readStorageConfig()
}

export function getPersonalityNames(): string[] {
  return getPersonalityConfig().order
}

export function getPersonalityColor(name: string): string {
  return getPersonalityConfig().colors[name] || '#6B7280'
}

function hexToRgba(hex: string, alpha: number) {
  const normalized = hex.replace('#', '')
  const r = parseInt(normalized.slice(0, 2), 16)
  const g = parseInt(normalized.slice(2, 4), 16)
  const b = parseInt(normalized.slice(4, 6), 16)
  return `rgba(${r}, ${g}, ${b}, ${alpha})`
}

export function getPersonalityBg(name: string): string {
  return hexToRgba(getPersonalityColor(name), 0.14)
}

export function setPersonalityColor(name: string, color: string) {
  const config = getPersonalityConfig()
  config.colors[name] = isHexColor(color) ? color : '#64748B'
  if (!config.order.includes(name)) config.order.push(name)
  return savePersonalityConfig(config)
}

export function addPersonalityTag(name: string, color: string) {
  const trimmed = name.trim()
  if (!trimmed) return getPersonalityConfig()
  const config = getPersonalityConfig()
  if (!config.order.includes(trimmed)) config.order.push(trimmed)
  config.colors[trimmed] = isHexColor(color) ? color : '#64748B'
  return savePersonalityConfig(config)
}

export function removePersonalityTag(name: string) {
  const config = getPersonalityConfig()
  config.order = config.order.filter(item => item !== name)
  delete config.colors[name]
  return savePersonalityConfig(config)
}

export function reorderPersonalityTags(names: string[]) {
  const config = getPersonalityConfig()
  config.order = Array.from(new Set(names))
  return savePersonalityConfig(config)
}

export function extractPersonalityConfigFromBusinessHours(raw: unknown): PersonalityTagConfig | null {
  const parsed = parseJSON(raw)
  const candidate = parsed.personality_tag_config
  if (!isRecord(candidate)) return null
  return normalizeConfig(candidate as Partial<PersonalityTagConfig>)
}

export function buildBusinessHoursWithPersonality(raw: unknown, config: PersonalityTagConfig) {
  const parsed = parseJSON(raw)
  parsed.personality_tag_config = normalizeConfig(config)
  return parsed
}

export function syncPersonalityConfigFromShop(shop?: Partial<Shop> | null) {
  const config = extractPersonalityConfigFromBusinessHours(shop?.business_hours)
  if (!config) return getPersonalityConfig()
  return savePersonalityConfig(config)
}
