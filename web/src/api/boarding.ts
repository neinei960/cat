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

export type BoardingHolidayCreatePayload = {
  holiday_date?: string
  start_date?: string
  end_date?: string
  name?: string
}

export function createBoardingHoliday(data: BoardingHolidayCreatePayload) {
  return request<BoardingHoliday[]>({ url: '/b/boarding/holidays', method: 'POST', data })
}

export function deleteBoardingHoliday(id: number) {
  return request({ url: `/b/boarding/holidays/${id}`, method: 'DELETE' })
}

export function getBoardingPolicies() {
  return request<BoardingDiscountPolicy[]>({ url: '/b/boarding/policies' })
}

export function getBoardingSpecialItems(params?: { active_only?: 0 | 1 }) {
  return request<BoardingSpecialItem[]>({ url: '/b/boarding/special-items', data: params })
}

export function createBoardingSpecialItem(data: Partial<BoardingSpecialItem>) {
  return request<BoardingSpecialItem>({ url: '/b/boarding/special-items', method: 'POST', data })
}

export function updateBoardingSpecialItem(id: number, data: Partial<BoardingSpecialItem>) {
  return request<BoardingSpecialItem>({ url: `/b/boarding/special-items/${id}`, method: 'PUT', data })
}

export function deleteBoardingSpecialItem(id: number) {
  return request({ url: `/b/boarding/special-items/${id}`, method: 'DELETE' })
}

export function createBoardingPolicy(data: any) {
  return request<BoardingDiscountPolicy>({ url: '/b/boarding/policies', method: 'POST', data })
}

export function updateBoardingPolicy(id: number, data: any) {
  return request<BoardingDiscountPolicy>({ url: `/b/boarding/policies/${id}`, method: 'PUT', data })
}

export type BoardingSpecialItemSelectionPayload = {
  id: number
  daily_price: number
  days: number
}

export function previewBoardingOrder(data: {
  customer_id?: number
  pet_ids?: number[]
  pet_count?: number
  cabinet_id?: number
  check_in_at?: string
  check_out_at?: string
  deposit_enabled?: boolean
  special_item_id?: number
  special_item_daily_price?: number
  special_item_days?: number
  special_items?: BoardingSpecialItemSelectionPayload[]
  policy_ids?: number[]
  room_groups?: Array<{
    pet_ids?: number[]
    pet_count?: number
    cabinet_id: number
    check_in_at: string
    check_out_at: string
    special_item_id?: number
    special_item_daily_price?: number
    special_item_days?: number
    special_items?: BoardingSpecialItemSelectionPayload[]
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
  deposit_enabled?: boolean
  special_item_id?: number
  special_item_daily_price?: number
  special_item_days?: number
  special_items?: BoardingSpecialItemSelectionPayload[]
  policy_ids?: number[]
  room_groups?: Array<{
    pet_ids?: number[]
    pet_count?: number
    cabinet_id: number
    check_in_at: string
    check_out_at: string
    special_item_id?: number
    special_item_daily_price?: number
    special_item_days?: number
    special_items?: BoardingSpecialItemSelectionPayload[]
  }>
  has_deworming?: boolean | null
  remark?: string
}) {
  return request<BoardingOrder>({ url: '/b/boarding/orders', method: 'POST', data })
}

export function getBoardingOrders(params?: PageParams & {
  status?: string
  date_from?: string
  date_to?: string
  cabinet_id?: number
}) {
  return request<PageResult<BoardingOrder>>({ url: '/b/boarding/orders', data: params })
}

export function getBoardingOrder(id: number) {
  return request<BoardingOrder>({ url: `/b/boarding/orders/${id}` })
}

export function updateBoardingOrderDeworming(id: number, hasDeworming: boolean | null) {
  return request<BoardingOrder>({ url: `/b/boarding/orders/${id}/deworming`, method: 'PUT', data: { has_deworming: hasDeworming } })
}

export function getBoardingDashboard() {
  return request<BoardingDashboardGroup[]>({ url: '/b/boarding/dashboard' })
}

export type BoardingPriceAdjustPayload = { discount_amount?: number; special_item_id?: number; special_item_daily_price?: number; special_item_days?: number; special_items?: BoardingSpecialItemSelectionPayload[] }

export function checkInBoardingOrder(id: number, data?: BoardingPriceAdjustPayload) {
  return request<BoardingOrder>({ url: `/b/boarding/orders/${id}/check-in`, method: 'PUT', data: data || {} })
}

export function checkInBoardingRoom(id: number, roomId: number, data?: BoardingPriceAdjustPayload) {
  return request<BoardingOrder>({ url: `/b/boarding/orders/${id}/rooms/${roomId}/check-in`, method: 'PUT', data: data || {} })
}

export function adjustBoardingOrderPrice(id: number, data?: BoardingPriceAdjustPayload) {
  return request<BoardingOrder>({ url: `/b/boarding/orders/${id}/adjust-price`, method: 'PUT', data: data || {} })
}

export function adjustBoardingRoomPrice(id: number, roomId: number, data?: BoardingPriceAdjustPayload) {
  return request<BoardingOrder>({ url: `/b/boarding/orders/${id}/rooms/${roomId}/adjust-price`, method: 'PUT', data: data || {} })
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
