import { request } from './request'

interface StaffSlot {
  staff: Staff
  slots: { start_time: string; end_time: string }[]
}

interface AppointmentPetInput {
  pet_id: number
  service_ids: number[]
}

interface CreateApptReq {
  customer_id: number
  pet_id?: number
  pets?: AppointmentPetInput[]
  staff_id?: number
  date: string
  start_time: string
  end_time?: string
  service_ids?: number[]
  source?: number
  notes?: string
}

export function getAvailableSlots(date: string, options: { service_ids: number[]; duration: number; exclude_id?: number }) {
  const params = new URLSearchParams({ date, duration: String(options.duration || 0) })
  if (options.service_ids.length > 0) {
    params.set('service_ids', options.service_ids.join(','))
  }
  if (options.exclude_id) {
    params.set('exclude_id', String(options.exclude_id))
  }
  return request<StaffSlot[]>({ url: `/b/appointments/slots?${params.toString()}` })
}

export function getAppointmentCalendar(startDate: string, endDate: string) {
  return request<any[]>({ url: `/b/appointments/calendar?start_date=${startDate}&end_date=${endDate}` })
}

export function createAppointment(data: CreateApptReq) {
  return request<any>({ url: '/b/appointments', method: 'POST', data })
}

export function updateAppointment(id: number, data: CreateApptReq) {
  return request<any>({ url: `/b/appointments/${id}`, method: 'PUT', data })
}

export function getAppointmentList(params?: PageParams & { status?: number; date_from?: string; date_to?: string }) {
  return request<PageResult<any>>({ url: '/b/appointments', data: params })
}

export function getAppointment(id: number) {
  return request<any>({ url: `/b/appointments/${id}` })
}

export function updateAppointmentStatus(id: number, data: { status: number; staff_notes?: string; cancel_reason?: string; cancelled_by?: string }) {
  return request({ url: `/b/appointments/${id}/status`, method: 'PUT', data })
}

export function assignStaff(id: number, staffId: number) {
  return request({ url: `/b/appointments/${id}/assign`, method: 'PUT', data: { staff_id: staffId } })
}

export function rescheduleAppointment(id: number, date: string, startTime: string) {
  return request({ url: `/b/appointments/${id}/reschedule`, method: 'PUT', data: { date, start_time: startTime } })
}
