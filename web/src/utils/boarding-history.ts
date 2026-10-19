export interface BoardingHistoryCustomerLike {
  nickname?: string
  phone?: string
  remark?: string
}

export interface BoardingHistoryCabinetLike {
  cabinet_type?: string
}

export interface BoardingHistoryPetEntityLike {
  name?: string
  breed?: string
  gender?: number
  birth_date?: string
}

export interface BoardingHistoryPetLike {
  pet_name_snapshot?: string
  pet?: BoardingHistoryPetEntityLike
}

export interface BoardingHistoryRoomLike {
  room_index?: number
  cabinet?: BoardingHistoryCabinetLike
  pets?: BoardingHistoryPetLike[]
}

export interface BoardingHistoryOrderLike {
  remark?: string
  customer?: BoardingHistoryCustomerLike
  cabinet?: BoardingHistoryCabinetLike
  rooms?: BoardingHistoryRoomLike[]
  pets?: BoardingHistoryPetLike[]
}

export interface BoardingHistoryPetProfile {
  name: string
  meta: string
}

export function getBoardingHistoryCustomerLabel(order?: BoardingHistoryOrderLike | null): string {
  return order?.customer?.nickname?.trim()
    || order?.customer?.phone?.trim()
    || '-'
}

export function getBoardingHistoryRoomSummary(order?: BoardingHistoryOrderLike | null): string {
  const rooms = order?.rooms || []
  const cabinetTypes = Array.from(
    new Set(
      rooms
        .map((room) => room.cabinet?.cabinet_type?.trim())
        .filter((value): value is string => !!value),
    ),
  )

  if (rooms.length > 1) {
    const roomCopy = `${rooms.length} 个房间`
    return cabinetTypes.length ? `${roomCopy} · ${cabinetTypes.join('、')}` : roomCopy
  }

  if (cabinetTypes.length) return cabinetTypes[0]
  return order?.cabinet?.cabinet_type?.trim() || '未记录房型'
}

export function getBoardingHistoryRemarkSummary(remark?: string | null, maxLength = 22): string {
  const normalized = (remark || '').trim().replace(/\s+/g, ' ')
  if (!normalized) return ''
  if (normalized.length <= maxLength) return normalized
  return `${normalized.slice(0, Math.max(maxLength - 1, 0))}…`
}

export function getBoardingHistoryPetNames(order?: BoardingHistoryOrderLike | null): string {
  const names = buildBoardingHistoryPetProfiles(order).map((item) => item.name)
  return names.length ? names.join('、') : '未记录猫咪'
}

export function buildBoardingHistoryPetProfiles(order?: BoardingHistoryOrderLike | null): BoardingHistoryPetProfile[] {
  const roomPets = (order?.rooms || []).flatMap((room) => room.pets || [])
  const pets = (order?.pets && order.pets.length ? order.pets : roomPets)
  const profiles: BoardingHistoryPetProfile[] = []

  for (const item of pets) {
    const entity = item.pet
    const name = entity?.name?.trim() || item.pet_name_snapshot?.trim() || '未记录猫咪'
    const metaParts = [
      entity?.breed?.trim() || '',
      genderLabel(entity?.gender),
      calcAge(entity?.birth_date),
    ].filter(Boolean)

    profiles.push({
      name,
      meta: metaParts.join(' · '),
    })
  }

  return profiles
}

function genderLabel(gender?: number): string {
  if (gender === 1) return '公'
  if (gender === 2) return '母'
  return ''
}

function calcAge(birthDate?: string): string {
  if (!birthDate) return ''
  const birth = new Date(birthDate)
  if (Number.isNaN(birth.getTime())) return ''

  const now = new Date()
  const months = (now.getFullYear() - birth.getFullYear()) * 12 + (now.getMonth() - birth.getMonth())
  if (months < 1) return '不到1个月'
  if (months < 12) return `${months}个月`

  const years = Math.floor(months / 12)
  const remainder = months % 12
  return remainder > 0 ? `${years}岁${remainder}个月` : `${years}岁`
}
