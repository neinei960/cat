export interface AppointmentCalendarSummary {
  total: number
  pendingConfirm: number
  unassigned: number
  waitingCheckout: number
}

export interface AppointmentCalendarSummaryItem {
  ID?: number | null
  status?: number | null
  staff_id?: number | null
}

export function buildAppointmentCalendarSummary(items: AppointmentCalendarSummaryItem[]): AppointmentCalendarSummary {
  return items.reduce<AppointmentCalendarSummary>((summary, item) => {
    const status = typeof item?.status === 'number' ? item.status : -1
    const staffId = typeof item?.staff_id === 'number' ? item.staff_id : 0

    if (status !== 4) {
      summary.total += 1
    }

    if (status === 0) {
      summary.pendingConfirm += 1
    }

    if (status !== 4 && !staffId) {
      summary.unassigned += 1
    }

    if (status === 3) {
      summary.waitingCheckout += 1
    }

    return summary
  }, {
    total: 0,
    pendingConfirm: 0,
    unassigned: 0,
    waitingCheckout: 0,
  })
}
