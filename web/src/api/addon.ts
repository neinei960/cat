import { request } from './request'

export interface ServiceAddon {
  ID: number
  name: string
  default_price: number
  is_variable: boolean
  sort_order: number
  status: number
}

export function getAddonList() {
  return request<ServiceAddon[]>({ url: '/b/addons' })
}

export function createAddon(data: Partial<ServiceAddon>) {
  return request<ServiceAddon>({ url: '/b/addons', method: 'POST', data })
}

export function updateAddon(id: number, data: Partial<ServiceAddon>) {
  return request<ServiceAddon>({ url: `/b/addons/${id}`, method: 'PUT', data })
}

export function deleteAddon(id: number) {
  return request({ url: `/b/addons/${id}`, method: 'DELETE' })
}

export function priceLookup(serviceId: number, furLevel: string) {
  return request<{ price: number; service_name: string; fur_level: string }>({
    url: `/b/orders/price-lookup?service_id=${serviceId}&fur_level=${furLevel}`,
  })
}
