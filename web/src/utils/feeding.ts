export const feedingWeekdays = [
  { value: 1, label: '周一' },
  { value: 2, label: '周二' },
  { value: 3, label: '周三' },
  { value: 4, label: '周四' },
  { value: 5, label: '周五' },
  { value: 6, label: '周六' },
  { value: 0, label: '周日' },
]

export const feedingWindows = [
  { value: 'all_day', label: '全天' },
  { value: 'morning', label: '早间' },
  { value: 'afternoon', label: '午后' },
  { value: 'evening', label: '晚间' },
]

export function feedingStatusLabel(status?: string) {
  return {
    draft: '草稿',
    active: '进行中',
    paused: '已暂停',
    completed: '已完成',
    cancelled: '已取消',
    pending: '待上门',
    assigned: '已分配',
    in_progress: '进行中',
    done: '已完成',
    exception: '异常',
  }[status || ''] || status || '-'
}

export function feedingWindowLabel(windowCode?: string) {
  return feedingWindows.find(item => item.value === windowCode)?.label || windowCode || '-'
}

export function parseFeedingAddress(raw?: string): FeedingAddressSnapshot {
  if (!raw) return { address: '', detail: '', door_code: '' }
  try {
    const parsed = JSON.parse(raw)
    return {
      address: parsed.address || '',
      detail: parsed.detail || '',
      door_code: parsed.door_code || '',
    }
  } catch {
    return { address: raw, detail: '', door_code: '' }
  }
}

export function parseFeedingSelectedItems(raw?: string): FeedingItemTemplate[] {
  if (!raw) return []
  try {
    const parsed = JSON.parse(raw)
    return Array.isArray(parsed) ? parsed : []
  } catch {
    return []
  }
}

export function parseFeedingSelectedDates(raw?: string): string[] {
  if (!raw) return []
  try {
    const parsed = JSON.parse(raw)
    return Array.isArray(parsed) ? parsed.filter((item) => typeof item === 'string') : []
  } catch {
    return []
  }
}

export function getFeedingDateOptions(startDate?: string, endDate?: string) {
  if (!startDate || !endDate || startDate > endDate) return [] as Array<{ date: string; label: string; weekday: number }>
  const [startYear, startMonth, startDay] = startDate.split('-').map(Number)
  const [endYear, endMonth, endDay] = endDate.split('-').map(Number)
  const start = new Date(startYear, startMonth - 1, startDay)
  const end = new Date(endYear, endMonth - 1, endDay)
  const result: Array<{ date: string; label: string; weekday: number }> = []
  for (const current = new Date(start); current <= end; current.setDate(current.getDate() + 1)) {
    const year = current.getFullYear()
    const month = `${current.getMonth() + 1}`.padStart(2, '0')
    const day = `${current.getDate()}`.padStart(2, '0')
    const date = `${year}-${month}-${day}`
    const weekday = current.getDay()
    const weekdayLabel = feedingWeekdays.find((item) => item.value === weekday)?.label || ''
    result.push({ date, weekday, label: `${month}-${day} ${weekdayLabel}` })
  }
  return result
}

export function formatFeedingDateRange(startDate?: string, endDate?: string) {
  if (!startDate || !endDate) return '-'
  return `${startDate} 至 ${endDate}`
}
