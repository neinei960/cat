import { request } from './request'

export function getCategoryTree() {
  return request<ServiceCategory[]>({ url: '/b/service-categories' })
}

export function createCategory(data: { name: string; parent_id?: number; sort_order?: number }) {
  return request<ServiceCategory>({ url: '/b/service-categories', method: 'POST', data })
}

export function updateCategory(id: number, data: { name: string; sort_order?: number }) {
  return request<ServiceCategory>({ url: `/b/service-categories/${id}`, method: 'PUT', data })
}

export function deleteCategory(id: number) {
  return request({ url: `/b/service-categories/${id}`, method: 'DELETE' })
}
