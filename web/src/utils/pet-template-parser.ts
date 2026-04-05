export interface ParsedPetInfo {
  surveyType: '' | 'dental' | 'grooming'
  name: string
  age: string
  birthDate: string
  breed: string
  gender: number
  neutered: boolean
  dailyDiet: string
  furMatted: string
  lastBathTime: string
  vaccination: string
  healthHistory: string
  personality: string
  reactions: string
  reactionsLabel: string
  source: string
}

type ParsedFieldKey =
  | 'name'
  | 'birthDate'
  | 'age'
  | 'breed'
  | 'gender'
  | 'neutered'
  | 'dailyDiet'
  | 'furMatted'
  | 'lastBathTime'
  | 'vaccination'
  | 'healthHistory'
  | 'personality'
  | 'reactions'

const COMMON_TEMPLATE_FIELD_MAP: Record<string, ParsedFieldKey> = {
  大名: 'name',
  名字: 'name',
  姓名: 'name',
  出生日期: 'birthDate',
  生日: 'birthDate',
  年龄: 'age',
  品种: 'breed',
  性别: 'gender',
  是否绝育: 'neutered',
  疫苗是否齐全: 'vaccination',
  疾病史和健康情况: 'healthHistory',
  健康情况: 'healthHistory',
  'i猫e猫': 'personality',
  性格: 'personality',
}

const DENTAL_TEMPLATE_FIELD_MAP: Record<string, ParsedFieldKey> = {
  日常饮食习惯: 'dailyDiet',
  日常饮食: 'dailyDiet',
  对陌生环境反应: 'reactions',
}

const GROOMING_TEMPLATE_FIELD_MAP: Record<string, ParsedFieldKey> = {
  是否毛发打结: 'furMatted',
  毛发打结: 'furMatted',
  上次洗澡时间: 'lastBathTime',
  上次洗澡: 'lastBathTime',
  '对水/声音/陌生环境反应': 'reactions',
  对陌生环境反应: 'reactions',
  对水反应: 'reactions',
  对声音反应: 'reactions',
}

function compactFieldLabel(value: string) {
  return value
    .replace(/[：:\s/|、，,（()）\-_.。]/g, '')
    .trim()
}

function resolveMappedField(
  key: string,
  fieldMap: Record<string, ParsedFieldKey>,
  surveyType: ParsedPetInfo['surveyType']
): ParsedFieldKey | undefined {
  if (fieldMap[key]) return fieldMap[key]

  const compactKey = compactFieldLabel(key)
  if (!compactKey) return undefined

  for (const [label, mappedField] of Object.entries(fieldMap)) {
    if (compactFieldLabel(label) === compactKey) return mappedField
  }

  if (compactKey === '反应' || compactKey.endsWith('反应') || compactKey.includes('环境反应')) {
    return 'reactions'
  }

  if (surveyType === 'grooming' && (compactKey.includes('洗澡时间') || compactKey.includes('上次洗澡'))) {
    return 'lastBathTime'
  }

  return undefined
}

export function createEmptyParsedPet(): ParsedPetInfo {
  return {
    surveyType: '',
    name: '',
    age: '',
    birthDate: '',
    breed: '',
    gender: 0,
    neutered: false,
    dailyDiet: '',
    furMatted: '',
    lastBathTime: '',
    vaccination: '',
    healthHistory: '',
    personality: '',
    reactions: '',
    reactionsLabel: '',
    source: '',
  }
}

export function parsePetTemplate(text: string): ParsedPetInfo {
  const result = createEmptyParsedPet()
  const lines = text.split(/\n/).map(line => line.trim()).filter(Boolean)
  const firstLine = lines[0] || ''

  if (firstLine.includes('|')) {
    const [source, title] = firstLine.split('|')
    result.source = source.trim()
    if ((title || '').includes('洁牙')) result.surveyType = 'dental'
    if ((title || '').includes('洗护')) result.surveyType = 'grooming'
  }

  const sectionHeaders = ['基本信息', '健康情况', '性格小秘密']
  const normalizeFieldLabel = (value: string) => value.replace(/^[^\u4e00-\u9fa5A-Za-z0-9]+/, '').trim()
  const normalizeDateValue = (value: string) => value
    .replace(/[年./]/g, '-')
    .replace(/月/g, '')
    .replace(/-+/g, '-')
    .replace(/^-|-$/g, '')
    .trim()
  const normalizeBooleanValue = (value: string) => value.trim()
  const fieldMap = {
    ...COMMON_TEMPLATE_FIELD_MAP,
    ...(result.surveyType === 'dental' ? DENTAL_TEMPLATE_FIELD_MAP : {}),
    ...(result.surveyType === 'grooming' ? GROOMING_TEMPLATE_FIELD_MAP : {}),
  }

  for (const line of lines) {
    if (
      line.includes('请家长阅读洗浴须知') ||
      line.includes('请家长阅读洗浴需知') ||
      line.includes('洗浴须知') ||
      line.includes('洗浴需知')
    ) {
      continue
    }

    const normalizedLine = normalizeFieldLabel(line)
    if (sectionHeaders.includes(normalizedLine)) continue

    const match = line.match(/^(.+?)[：:](.+)$/)
    if (!match) continue

    const key = normalizeFieldLabel(match[1])
    const value = match[2].trim()
    const mappedField = resolveMappedField(key, fieldMap, result.surveyType)
    if (!mappedField) continue

    switch (mappedField) {
      case 'name':
        result.name = value
        break
      case 'birthDate':
        result.birthDate = normalizeDateValue(value)
        break
      case 'age':
        result.age = value
        break
      case 'breed':
        result.breed = value
        break
      case 'gender':
        if (value.includes('公')) result.gender = 1
        else if (value.includes('母')) result.gender = 2
        else result.gender = 0
        break
      case 'neutered':
        result.neutered = normalizeBooleanValue(value).includes('是') || normalizeBooleanValue(value) === '已绝育'
        break
      case 'dailyDiet':
        result.dailyDiet = value
        break
      case 'furMatted':
        result.furMatted = value
        break
      case 'lastBathTime':
        result.lastBathTime = value
        break
      case 'vaccination':
        result.vaccination = value
        break
      case 'healthHistory':
        result.healthHistory = value
        break
      case 'personality':
        result.personality = value
        break
      case 'reactions':
        result.reactions = value
        result.reactionsLabel = key
        break
    }
  }

  return result
}

export function buildCareNotes(pet: ParsedPetInfo): string {
  const parts: string[] = []
  if (pet.dailyDiet) parts.push(`日常饮食：${pet.dailyDiet}`)
  if (pet.furMatted) parts.push(`毛发打结：${pet.furMatted}`)
  if (pet.lastBathTime) parts.push(`上次洗澡：${pet.lastBathTime}`)
  if (pet.vaccination) parts.push(`疫苗：${pet.vaccination}`)
  if (pet.healthHistory) parts.push(`疾病史：${pet.healthHistory}`)
  return parts.join('\n')
}

export function buildAppointmentRemarkParts(pet: ParsedPetInfo): string[] {
  const parts: string[] = []
  if (pet.dailyDiet) parts.push(`日常饮食：${pet.dailyDiet}`)
  if (pet.furMatted && pet.furMatted !== '否' && pet.furMatted !== '无') parts.push(`毛发打结：${pet.furMatted}`)
  if (pet.lastBathTime) parts.push(`上次洗澡：${pet.lastBathTime}`)
  if (pet.healthHistory) parts.push(`健康情况：${pet.healthHistory}`)
  if (pet.reactions) parts.push(`${getReactionRemarkLabel(pet)}：${pet.reactions}`)
  return parts
}

export function buildAppointmentRemarkPreview(pet: ParsedPetInfo): string {
  return buildAppointmentRemarkParts(pet).join('\n')
}

function getReactionRemarkLabel(pet: ParsedPetInfo): string {
  if (pet.reactionsLabel) return pet.reactionsLabel
  if (pet.surveyType === 'dental') return '对陌生环境反应'
  if (pet.surveyType === 'grooming') return '对水/声音/陌生环境反应'
  return '反应'
}
