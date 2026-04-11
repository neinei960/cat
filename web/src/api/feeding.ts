import { request } from './request'

export function getFeedingSettings() {
  return request<FeedingSettings>({ url: '/b/feeding/settings' })
}

export function updateFeedingPricing(data: FeedingPricingSetting) {
  return request<FeedingSettings>({ url: '/b/feeding/settings/pricing', method: 'PUT', data })
}

export function updateFeedingItems(data: FeedingItemTemplate[]) {
  return request<FeedingSettings>({ url: '/b/feeding/settings/items', method: 'PUT', data })
}

export function createFeedingPlan(data: {
  customer_id: number
  address_snapshot: FeedingAddressSnapshot
  contact_name: string
  contact_phone: string
  start_date: string
  end_date: string
  selected_dates: string[]
  remark?: string
  pets: Array<{ pet_id: number; remark?: string }>
  rules: Array<{ weekday: number; window_code: string; visit_count: number }>
  item_codes: string[]
}) {
  return request<FeedingPlan>({ url: '/b/feeding/plans', method: 'POST', data })
}

export function updateFeedingPlan(id: number, data: {
  customer_id: number
  address_snapshot: FeedingAddressSnapshot
  contact_name: string
  contact_phone: string
  start_date: string
  end_date: string
  selected_dates: string[]
  remark?: string
  pets: Array<{ pet_id: number; remark?: string }>
  rules: Array<{ weekday: number; window_code: string; visit_count: number }>
  item_codes: string[]
}) {
  return request<FeedingPlan>({ url: `/b/feeding/plans/${id}`, method: 'PUT', data })
}

export function getFeedingPlans(params?: PageParams & { status?: string; customer_id?: number }) {
  return request<PageResult<FeedingPlan>>({ url: '/b/feeding/plans', data: params })
}

export function getFeedingPlan(id: number) {
  return request<FeedingPlan>({ url: `/b/feeding/plans/${id}` })
}

export function updateFeedingDeposit(id: number, deposit: number) {
  return request<FeedingPlan>({ url: `/b/feeding/plans/${id}/deposit`, method: 'PUT', data: { deposit } })
}

export function updateFeedingPlayDates(id: number, playDates: string[]) {
  return request<FeedingPlan>({ url: `/b/feeding/plans/${id}/play-dates`, method: 'PUT', data: { play_dates: playDates } })
}

export function pauseFeedingPlan(id: number) {
  return request<FeedingPlan>({ url: `/b/feeding/plans/${id}/pause`, method: 'PUT', data: {} })
}

export function resumeFeedingPlan(id: number) {
  return request<FeedingPlan>({ url: `/b/feeding/plans/${id}/resume`, method: 'PUT', data: {} })
}

export function cancelFeedingPlan(id: number) {
  return request<FeedingPlan>({ url: `/b/feeding/plans/${id}/cancel`, method: 'PUT', data: {} })
}

export function generateFeedingOrder(id: number) {
  return request<Order>({ url: `/b/feeding/plans/${id}/generate-order`, method: 'POST', data: {} })
}

export function getFeedingDashboard(params?: { date?: string; staff_id?: number; window_code?: string }) {
  return request<FeedingDashboardResponse>({ url: '/b/feeding/dashboard', data: params })
}

export function getFeedingVisits(params?: {
  id?: number
  plan_id?: number
  scheduled_date?: string
  status?: string
  staff_id?: number
  window_code?: string
}) {
  return request<FeedingVisit[]>({ url: '/b/feeding/visits', data: params })
}

export function assignFeedingVisit(id: number, staffId: number) {
  return request<FeedingVisit>({ url: `/b/feeding/visits/${id}/assign`, method: 'PUT', data: { staff_id: staffId } })
}

export function startFeedingVisit(id: number) {
  return request<FeedingVisit>({ url: `/b/feeding/visits/${id}/start`, method: 'PUT', data: {} })
}

export function completeFeedingVisit(id: number, data: {
  item_checks: Array<{ id: number; checked: boolean }>
  customer_note?: string
  internal_note?: string
}) {
  return request<FeedingVisit>({ url: `/b/feeding/visits/${id}/complete`, method: 'PUT', data })
}

export function exceptionFeedingVisit(id: number, data: {
  exception_type: string
  customer_note?: string
  internal_note?: string
}) {
  return request<FeedingVisit>({ url: `/b/feeding/visits/${id}/exception`, method: 'PUT', data })
}

export function updateFeedingVisitNote(id: number, data: { internal_note: string }) {
  return request<FeedingVisit>({ url: `/b/feeding/visits/${id}/note`, method: 'PUT', data })
}

export function addFeedingVisitMedia(id: number, data: { media_type?: string; url: string }) {
  return request<FeedingVisitMedia>({ url: `/b/feeding/visits/${id}/media`, method: 'POST', data })
}
