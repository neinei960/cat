import { request } from './request'

export interface FurCategory {
  ID: number
  name: string
  sort_order: number
  status: number
}

export function getFurCategoryList() {
  return request<FurCategory[]>({ url: '/b/fur-categories' })
}

export function createFurCategory(data: Partial<FurCategory>) {
  return request<FurCategory>({ url: '/b/fur-categories', method: 'POST', data })
}

export function updateFurCategory(id: number, data: Partial<FurCategory>) {
  return request<FurCategory>({ url: `/b/fur-categories/${id}`, method: 'PUT', data })
}

export function deleteFurCategory(id: number) {
  return request({ url: `/b/fur-categories/${id}`, method: 'DELETE' })
}
