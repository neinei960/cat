import { request } from './request'

export function getDashboardOverview(startDate?: string, endDate?: string) {
  let url = '/b/dashboard/overview'
  if (startDate && endDate) {
    url += `?start_date=${startDate}&end_date=${endDate}`
  }
  return request<{
    today_revenue: number
    today_order_count: number
    today_appointment_count: number
    today_service_completed_count: number
    today_pending_settlement_count: number
    today_refunded_order_count: number
    today_new_customers: number
    pending_appointments: number
    total_customers: number
  }>({ url })
}

export function getRevenueChart(startDate: string, endDate: string) {
  return request<any[]>({ url: `/b/dashboard/revenue?start_date=${startDate}&end_date=${endDate}` })
}

export function getServiceRanking(startDate: string, endDate: string) {
  return request<{ service_name: string; count: number; revenue: number }[]>({
    url: `/b/dashboard/services?start_date=${startDate}&end_date=${endDate}`,
  })
}

export function getStaffPerformance(startDate: string, endDate: string) {
  return request<{ staff_name: string; appointment_count: number; revenue: number }[]>({
    url: `/b/dashboard/staff?start_date=${startDate}&end_date=${endDate}`,
  })
}

export function getCategoryStats(startDate: string, endDate: string) {
  return request<{ service_name: string; fur_level: string; count: number; revenue: number }[]>({
    url: `/b/dashboard/category?start_date=${startDate}&end_date=${endDate}`,
  })
}

export function getMemberStats(startDate: string, endDate: string) {
  return request<{
    active_members: number
    frozen_members: number
    total_balance: number
    total_member_spent: number
    range_recharge: number
    range_consumption: number
    template_breakdown: { template_id: number; template_name: string; count: number }[]
  }>({
    url: `/b/dashboard/members?start_date=${startDate}&end_date=${endDate}`,
  })
}
