import { request } from './request'

// Templates
export function getCardTemplates() {
  return request<MemberCardTemplate[]>({ url: '/b/member-card-templates' })
}

export function createCardTemplate(data: Partial<MemberCardTemplate>) {
  return request<MemberCardTemplate>({ url: '/b/member-card-templates', method: 'POST', data })
}

export function updateCardTemplate(id: number, data: Partial<MemberCardTemplate>) {
  return request<MemberCardTemplate>({ url: `/b/member-card-templates/${id}`, method: 'PUT', data })
}

export function deleteCardTemplate(id: number) {
  return request({ url: `/b/member-card-templates/${id}`, method: 'DELETE' })
}

export function setTemplateDiscounts(id: number, discounts: { category_id: number; category_name: string; discount_rate: number }[]) {
  return request<MemberCardTemplate>({ url: `/b/member-card-templates/${id}/discounts`, method: 'PUT', data: { discounts } })
}

// Customer member card operations
export function getCustomerCard(customerId: number) {
  return request<MemberCard | null>({ url: `/b/customers/${customerId}/member-card` })
}

export function openCard(customerId: number, data: { template_id: number; recharge_amount: number; remark?: string }) {
  return request<MemberCard>({ url: `/b/customers/${customerId}/member-card`, method: 'POST', data })
}

export function recharge(customerId: number, data: { amount: number; remark?: string }) {
  return request<MemberCard>({ url: `/b/customers/${customerId}/recharge`, method: 'POST', data })
}

export function adjustBalance(customerId: number, data: { amount: number; remark: string }) {
  return request<MemberCard>({ url: `/b/customers/${customerId}/adjust-balance`, method: 'PUT', data })
}

export function getRechargeRecords(customerId: number) {
  return request<RechargeRecord[]>({ url: `/b/customers/${customerId}/recharge-records` })
}

export function updateRechargeRecord(id: number, data: { amount?: number; remark?: string }) {
  return request<RechargeRecord>({ url: `/b/recharge-records/${id}`, method: 'PUT', data })
}

export function deleteRechargeRecord(id: number) {
  return request({ url: `/b/recharge-records/${id}`, method: 'DELETE' })
}
