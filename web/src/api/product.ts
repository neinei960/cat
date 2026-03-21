import { request } from './request'

export function getProductList(params?: PageParams & { category_id?: number; keyword?: string }) {
  return request<PageResult<any>>({ url: '/b/products', data: params })
}

export function getProduct(id: number) {
  return request<any>({ url: `/b/products/${id}` })
}

export function createProduct(data: any) {
  return request<any>({ url: '/b/products', method: 'POST', data })
}

export function updateProduct(id: number, data: any) {
  return request<any>({ url: `/b/products/${id}`, method: 'PUT', data })
}

export function deleteProduct(id: number) {
  return request({ url: `/b/products/${id}`, method: 'DELETE' })
}

export function getProductBrands() {
  return request<string[]>({ url: '/b/products/brands' })
}

export function getProductCategories() {
  return request<any[]>({ url: '/b/product-categories' })
}

export function createProductCategory(data: { name: string; sort_order?: number }) {
  return request<any>({ url: '/b/product-categories', method: 'POST', data })
}

export function updateProductCategory(id: number, data: { name?: string; sort_order?: number; status?: number }) {
  return request<any>({ url: `/b/product-categories/${id}`, method: 'PUT', data })
}

export function deleteProductCategory(id: number) {
  return request({ url: `/b/product-categories/${id}`, method: 'DELETE' })
}
