import { request } from './request'

export function getBoardingCabinets() {
  return request<BoardingCabinet[]>({ url: '/b/boarding/cabinets' })
}

export function createBoardingCabinet(data: Partial<BoardingCabinet>) {
  return request<BoardingCabinet>({ url: '/b/boarding/cabinets', method: 'POST', data })
}

export function updateBoardingCabinet(id: number, data: Partial<BoardingCabinet>) {
  return request<BoardingCabinet>({ url: `/b/boarding/cabinets/${id}`, method: 'PUT', data })
}

export function getAvailableBoardingCabinets(params: { check_in_at: string; check_out_at: string; pet_count: number; exclude_order_id?: number; exclude_room_id?: number }) {
  return request<BoardingCabinet[]>({ url: '/b/boarding/cabinets/availability', data: params })
}

export function getBoardingHolidays() {
  return request<BoardingHoliday[]>({ url: '/b/boarding/holidays' })
}

export function createBoardingHoliday(data: Partial<BoardingHoliday>) {
  return request<BoardingHoliday>({ url: '/b/boarding/holidays', method: 'POST', data })
}

export function deleteBoardingHoliday(id: number) {
  return request({ url: `/b/boarding/holidays/${id}`, method: 'DELETE' })
}

export function getBoardingPolicies() {
  return request<BoardingDiscountPolicy[]>({ url: '/b/boarding/policies' })
}

export function createBoardingPolicy(data: any) {
  return request<BoardingDiscountPolicy>({ url: '/b/boarding/policies', method: 'POST', data })
}

export function updateBoardingPolicy(id: number, data: any) {
  return request<BoardingDiscountPolicy>({ url: `/b/boarding/policies/${id}`, method: 'PUT', data })
}

export function previewBoardingOrder(data: {
  customer_id?: number
  pet_ids?: number[]
  pet_count?: number
  cabinet_id?: number
  check_in_at?: string
  check_out_at?: string
  policy_ids?: number[]
  room_groups?: Array<{
    pet_ids?: number[]
    pet_count?: number
    cabinet_id: number
    check_in_at: string
    check_out_at: string
  }>
}) {
  return request<BoardingPricePreview>({ url: '/b/boarding/orders/price-preview', method: 'POST', data })
}

export function createBoardingOrder(data: {
  customer_id: number
  pet_ids?: number[]
  cabinet_id?: number
  check_in_at?: string
  check_out_at?: string
  policy_ids?: number[]
  room_groups?: Array<{
    pet_ids?: number[]
    pet_count?: number
    cabinet_id: number
    check_in_at: string
    check_out_at: string
  }>
  has_deworming?: boolean | null
  remark?: string
}) {
  return request<BoardingOrder>({ url: '/b/boarding/orders', method: 'POST', data })
}

export function getBoardingOrders(params?: PageParams & { status?: string }) {
  return request<PageResult<BoardingOrder>>({ url: '/b/boarding/orders', data: params })
}

export function getBoardingOrder(id: number) {
  return request<BoardingOrder>({ url: `/b/boarding/orders/${id}` })
}

export function getBoardingDashboard() {
  return request<BoardingDashboardGroup[]>({ url: '/b/boarding/dashboard' })
}

export function checkInBoardingOrder(id: number, data?: { discount_amount?: number }) {
  return request<BoardingOrder>({ url: `/b/boarding/orders/${id}/check-in`, method: 'PUT', data: data || {} })
}

export function checkInBoardingRoom(id: number, roomId: number, data?: { discount_amount?: number }) {
  return request<BoardingOrder>({ url: `/b/boarding/orders/${id}/rooms/${roomId}/check-in`, method: 'PUT', data: data || {} })
}

export function checkOutBoardingOrder(id: number, actualCheckOutAt: string) {
  return request<BoardingOrder>({ url: `/b/boarding/orders/${id}/check-out`, method: 'PUT', data: { actual_check_out_at: actualCheckOutAt } })
}

export function checkOutBoardingRoom(id: number, roomId: number, actualCheckOutAt: string) {
  return request<BoardingOrder>({ url: `/b/boarding/orders/${id}/rooms/${roomId}/check-out`, method: 'PUT', data: { actual_check_out_at: actualCheckOutAt } })
}

export function extendBoardingOrder(id: number, checkOutAt: string) {
  return request<BoardingOrder>({ url: `/b/boarding/orders/${id}/extend`, method: 'PUT', data: { check_out_at: checkOutAt } })
}

export function extendBoardingRoom(id: number, roomId: number, checkOutAt: string) {
  return request<BoardingOrder>({ url: `/b/boarding/orders/${id}/rooms/${roomId}/extend`, method: 'PUT', data: { check_out_at: checkOutAt } })
}

export function changeBoardingCabinet(id: number, cabinetId: number) {
  return request<BoardingOrder>({ url: `/b/boarding/orders/${id}/change-cabinet`, method: 'PUT', data: { cabinet_id: cabinetId } })
}

export function changeBoardingRoomCabinet(id: number, roomId: number, cabinetId: number) {
  return request<BoardingOrder>({ url: `/b/boarding/orders/${id}/rooms/${roomId}/change-cabinet`, method: 'PUT', data: { cabinet_id: cabinetId } })
}

export function cancelBoardingOrder(id: number) {
  return request<BoardingOrder>({ url: `/b/boarding/orders/${id}/cancel`, method: 'PUT', data: {} })
}

export function cancelBoardingRoom(id: number, roomId: number) {
  return request<BoardingOrder>({ url: `/b/boarding/orders/${id}/rooms/${roomId}/cancel`, method: 'PUT', data: {} })
}
