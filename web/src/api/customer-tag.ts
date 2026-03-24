import { request } from './request'

export function getCustomerTags() {
  return request<CustomerTag[]>({ url: '/b/customer-tags' })
}

export function createCustomerTag(data: Partial<CustomerTag>) {
  return request<CustomerTag>({ url: '/b/customer-tags', method: 'POST', data })
}

export function updateCustomerTag(id: number, data: Partial<CustomerTag>) {
  return request<CustomerTag>({ url: `/b/customer-tags/${id}`, method: 'PUT', data })
}

export function deleteCustomerTag(id: number) {
  return request({ url: `/b/customer-tags/${id}`, method: 'DELETE' })
}
