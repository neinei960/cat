import { request } from './request'

export interface PetListParams extends PageParams {
  keyword?: string
  pet_tag?: string
}

export function getPetList(params?: PetListParams) {
  return request<PageResult<Pet>>({ url: '/b/pets', data: params })
}

export function getPet(id: number) {
  return request<Pet>({ url: `/b/pets/${id}` })
}

export function createPet(data: Partial<Pet>) {
  return request<Pet>({ url: '/b/pets', method: 'POST', data })
}

export function updatePet(id: number, data: Partial<Pet>) {
  return request<Pet>({ url: `/b/pets/${id}`, method: 'PUT', data })
}

export function deletePet(id: number) {
  return request({ url: `/b/pets/${id}`, method: 'DELETE' })
}
