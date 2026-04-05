import { request } from './request'

export interface PetBathReport {
  ID: number
  pet_id: number
  shop_id: number
  image_url: string
  bath_date?: string
  sort_order?: number
  CreatedAt: string
}

export function getPetBathReports(petId: number) {
  return request<PetBathReport[]>({ url: `/b/pets/${petId}/bath-reports` })
}

export function createPetBathReport(petId: number, imageUrl: string, bathDate?: string) {
  return request<PetBathReport>({
    url: `/b/pets/${petId}/bath-reports`,
    method: 'POST',
    data: { image_url: imageUrl, bath_date: bathDate },
  })
}

export function updatePetBathReport(petId: number, reportId: number, bathDate: string) {
  return request({
    url: `/b/pets/${petId}/bath-reports/${reportId}`,
    method: 'PUT',
    data: { bath_date: bathDate },
  })
}

export function deletePetBathReport(petId: number, reportId: number) {
  return request({
    url: `/b/pets/${petId}/bath-reports/${reportId}`,
    method: 'DELETE',
  })
}

export function reorderPetBathReports(petId: number, reportIds: number[]) {
  return request({
    url: `/b/pets/${petId}/bath-reports/reorder`,
    method: 'PUT',
    data: { report_ids: reportIds },
  })
}
