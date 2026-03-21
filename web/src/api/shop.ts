import { request } from './request'

export function getShop() {
  return request<Shop>({ url: '/b/shop' })
}

export function updateShop(data: Partial<Shop>) {
  return request<Shop>({ url: '/b/shop', method: 'PUT', data })
}
