import { request } from './request'

export function getServiceList(params?: (PageParams & { order_by?: string })) {
  return request<PageResult<ServiceItem>>({ url: '/b/services', data: params })
}

export function getService(id: number) {
  return request<ServiceItem>({ url: `/b/services/${id}` })
}

export function createService(data: Partial<ServiceItem>) {
  return request<ServiceItem>({ url: '/b/services', method: 'POST', data })
}

export function updateService(id: number, data: Partial<ServiceItem>) {
  return request<ServiceItem>({ url: `/b/services/${id}`, method: 'PUT', data })
}

export function deleteService(id: number) {
  return request({ url: `/b/services/${id}`, method: 'DELETE' })
}

export function getPriceRules(serviceId: number) {
  return request<ServicePriceRule[]>({ url: `/b/services/${serviceId}/prices` })
}

export function createPriceRule(serviceId: number, data: Partial<ServicePriceRule>) {
  return request<ServicePriceRule>({ url: `/b/services/${serviceId}/prices`, method: 'POST', data })
}

export function deletePriceRule(serviceId: number, ruleId: number) {
  return request({ url: `/b/services/${serviceId}/prices/${ruleId}`, method: 'DELETE' })
}

export function getDiscounts(serviceId: number) {
  return request<ServiceDiscount[]>({ url: `/b/services/${serviceId}/discounts` })
}

export function createDiscount(serviceId: number, data: Partial<ServiceDiscount>) {
  return request<ServiceDiscount>({ url: `/b/services/${serviceId}/discounts`, method: 'POST', data })
}

export function deleteDiscount(serviceId: number, discountId: number) {
  return request({ url: `/b/services/${serviceId}/discounts/${discountId}`, method: 'DELETE' })
}
