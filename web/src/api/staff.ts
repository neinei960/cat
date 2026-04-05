import { request } from './request'

export function getStaffList(params?: PageParams) {
  return request<PageResult<Staff>>({ url: '/b/staffs', data: params })
}

export function getStaff(id: number) {
  return request<Staff>({ url: `/b/staffs/${id}` })
}

export function createStaff(data: CreateStaffReq) {
  return request<Staff>({ url: '/b/staffs', method: 'POST', data })
}

export function updateStaff(id: number, data: Partial<Staff>) {
  return request<Staff>({ url: `/b/staffs/${id}`, method: 'PUT', data })
}

export function updateStaffOrder(staffIds: number[]) {
  return request({ url: '/b/staffs/order', method: 'PUT', data: { staff_ids: staffIds } })
}

export function deleteStaff(id: number) {
  return request({ url: `/b/staffs/${id}`, method: 'DELETE' })
}

export function resetStaffPassword(id: number, password: string) {
  return request({ url: `/b/staffs/${id}/password`, method: 'PUT', data: { password } })
}

export function getStaffSchedule(id: number, startDate: string, endDate: string) {
  return request<StaffSchedule[]>({ url: `/b/staffs/${id}/schedule`, data: { start_date: startDate, end_date: endDate } })
}

export function setStaffSchedule(id: number, data: Partial<StaffSchedule>) {
  return request({ url: `/b/staffs/${id}/schedule`, method: 'PUT', data })
}

export function batchSetSchedule(id: number, schedules: Partial<StaffSchedule>[]) {
  return request({ url: `/b/staffs/${id}/schedule/batch`, method: 'PUT', data: { schedules } })
}

export function getStaffServices(id: number) {
  return request<ServiceItem[]>({ url: `/b/staffs/${id}/services` })
}

export function setStaffServices(id: number, serviceIds: number[]) {
  return request({ url: `/b/staffs/${id}/services`, method: 'PUT', data: { service_ids: serviceIds } })
}
