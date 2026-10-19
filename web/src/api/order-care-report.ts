import { request } from './request'

export interface OrderCareReportSectionPayload {
  checks: string[]
  note: string
}

export interface CreateOrderCareReportPayload {
  pet_id: number
  pet_name: string
  breed: string
  gender: string
  age: string
  portrait_url: string
  weight: string
  care_date: string
  next_care_date: string
  care_content: string
  body_shape: string
  skin: OrderCareReportSectionPayload
  hair: OrderCareReportSectionPayload
  nails: OrderCareReportSectionPayload
  eyes_face: OrderCareReportSectionPayload
  ears: OrderCareReportSectionPayload
  oral: OrderCareReportSectionPayload
  anus: OrderCareReportSectionPayload
}

export interface OrderCareReportResult {
  image_url: string
  report_id: number
  bath_date: string
}

export function createOrderCareReport(orderId: number, payload: CreateOrderCareReportPayload) {
  return request<OrderCareReportResult>({
    url: `/b/orders/${orderId}/care-report`,
    method: 'POST',
    data: payload,
  })
}
