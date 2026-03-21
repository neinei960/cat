import { request } from './request'

// Auth
export function wxLogin(code: string, shopId: number) {
  return request<{ token: string; customer: any; is_new: boolean }>({
    url: '/auth/wx/login', method: 'POST', data: { code, shop_id: shopId }
  })
}

// Services
export function getServices() {
  return request<any[]>({ url: '/c/services' })
}

// Staff
export function getStaffs(serviceId: number) {
  return request<any[]>({ url: `/c/staffs?service_id=${serviceId}` })
}

// Slots
export function getSlots(date: string, serviceId: number) {
  return request<any[]>({ url: `/c/slots?date=${date}&service_id=${serviceId}` })
}

// Appointments
export function createAppointment(data: any) {
  return request<any>({ url: '/c/appointments', method: 'POST', data })
}

export function getAppointments(page = 1, pageSize = 20) {
  return request<{ list: any[]; total: number }>({ url: `/c/appointments?page=${page}&page_size=${pageSize}` })
}

export function getAppointment(id: number) {
  return request<any>({ url: `/c/appointments/${id}` })
}

export function cancelAppointment(id: number, reason?: string) {
  return request({ url: `/c/appointments/${id}/cancel`, method: 'PUT', data: { reason } })
}

// Pets
export function getPets() {
  return request<any[]>({ url: '/c/pets' })
}

export function createPet(data: any) {
  return request<any>({ url: '/c/pets', method: 'POST', data })
}

export function updatePet(id: number, data: any) {
  return request<any>({ url: `/c/pets/${id}`, method: 'PUT', data })
}
