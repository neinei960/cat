import { request } from './request'

export interface DashboardPaymentBreakdownItem {
  key: string
  label: string
  amount: number
}

export interface DashboardOverview {
  today_revenue: number
  month_revenue: number
  month_recharge: number
  month_collection: number
  today_order_count: number
  today_appointment_count: number
  today_service_completed_count: number
  today_pending_settlement_count: number
  today_refunded_order_count: number
  today_new_customers: number
  regular_customer_count: number
  pending_appointments: number
  total_customers: number
  avg_order_value: number
  no_show_rate: number
  no_show_count: number
  total_appointments: number
  payment_breakdown: DashboardPaymentBreakdownItem[]
}

export function getDashboardOverview(startDate?: string, endDate?: string) {
  let url = '/b/dashboard/overview'
  if (startDate && endDate) {
    url += `?start_date=${startDate}&end_date=${endDate}`
  }
  return request<DashboardOverview>({ url })
}

export function getRevenueChart(startDate: string, endDate: string) {
  return request<any[]>({ url: `/b/dashboard/revenue?start_date=${startDate}&end_date=${endDate}` })
}

export function getServiceRanking(startDate: string, endDate: string) {
  return request<{ service_name: string; count: number; revenue: number }[]>({
    url: `/b/dashboard/services?start_date=${startDate}&end_date=${endDate}`,
  })
}

export interface ProjectRevenueNode {
  key: string
  name: string
  kind: string
  count: number
  revenue: number
  children?: ProjectRevenueNode[]
}

export interface StaffPerformanceItem {
  staff_id: number
  staff_name: string
  appointment_count: number
  revenue: number
  product_revenue: number
  commission_rate: number
  product_commission_rate: number
  commission: number
}

export interface StaffCommissionDetail {
  order_id: number
  order_no: string
  date: string
  pay_method: string
  pay_method_label: string
  pay_amount: number
  service_amount: number
  product_amount: number
  commission_rate: number
  commission: number
  formula: string
  customer_name: string
  pet_summary: string
  remark: string
}

export function getProjectRevenueTree(startDate: string, endDate: string) {
  return request<ProjectRevenueNode[]>({
    url: `/b/dashboard/project-revenue?start_date=${startDate}&end_date=${endDate}`,
  })
}

export function getStaffPerformance(startDate: string, endDate: string) {
  return request<StaffPerformanceItem[]>({
    url: `/b/dashboard/staff?start_date=${startDate}&end_date=${endDate}`,
  })
}

export function getStaffCommissionDetails(staffId: number, startDate: string, endDate: string) {
  return request<StaffCommissionDetail[]>({
    url: `/b/dashboard/staff/${staffId}/commission-details?start_date=${startDate}&end_date=${endDate}`,
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
