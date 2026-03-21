import { request } from './request'

interface StaffSlot {
  staff: Staff
  slots: { start_time: string; end_time: string }[]
}

interface CreateApptReq {
  customer_id: number
  pet_id: number
  staff_id?: number
  date: string
  start_time: string
  service_ids: number[]
  source?: number
  notes?: string
}

export function getAvailableSlots(date: string, serviceId: number) {
  return request<StaffSlot[]>({ url: `/b/appointments/slots?date=${date}&service_id=${serviceId}` })
}

export function getAppointmentCalendar(startDate: string, endDate: string) {
  return request<any[]>({ url: `/b/appointments/calendar?start_date=${startDate}&end_date=${endDate}` })
}

export function createAppointment(data: CreateApptReq) {
  return request<any>({ url: '/b/appointments', method: 'POST', data })
}

export function getAppointmentList(params?: PageParams & { status?: number }) {
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
