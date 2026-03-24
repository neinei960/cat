import { request } from './request'

export function getCustomerList(params?: PageParams & { keyword?: string; member_card_template_id?: number; customer_tag_id?: number }) {
  return request<PageResult<Customer>>({ url: '/b/customers', data: params })
}

export function getCustomer(id: number) {
  return request<Customer>({ url: `/b/customers/${id}` })
}

export function createCustomer(data: Partial<Customer>) {
  return request<Customer>({ url: '/b/customers', method: 'POST', data })
}

export function updateCustomer(id: number, data: Partial<Customer>) {
  return request<Customer>({ url: `/b/customers/${id}`, method: 'PUT', data })
}

export function deleteCustomer(id: number) {
  return request({ url: `/b/customers/${id}`, method: 'DELETE' })
}

export function getCustomerPets(id: number) {
  return request<Pet[]>({ url: `/b/customers/${id}/pets` })
}

export function getDeletedCustomers(params?: PageParams) {
  return request<PageResult<Customer>>({ url: '/b/customers/trash', data: params })
}

export function restoreCustomer(id: number) {
  return request({ url: `/b/customers/${id}/restore`, method: 'POST' })
}
