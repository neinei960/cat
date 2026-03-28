import { request } from './request'

interface OrderItem {
  item_type: number
  item_id?: number
  name: string
  quantity: number
  unit_price: number
}

interface AddonItem {
  name: string
  amount: number
}

interface CreateOrderReq {
  pet_id?: number
  customer_id?: number
  staff_id?: number
  service_id?: number
  discount_amount?: number
  remark?: string
  items?: OrderItem[]
  addons?: AddonItem[]
}

export function createOrder(data: CreateOrderReq) {
  return request<any>({ url: '/b/orders', method: 'POST', data })
}

export function createOrderFromAppointment(appointmentId: number) {
  return request<any>({ url: '/b/orders/from-appointment', method: 'POST', data: { appointment_id: appointmentId } })
}

export function createBatchOrdersFromAppointment(appointmentId: number, extra?: any) {
  return request<any[]>({ url: '/b/orders/from-appointment/batch', method: 'POST', data: { appointment_id: appointmentId, ...extra } })
}

export function getOrderList(params?: PageParams & { status?: number; keyword?: string; date_from?: string; date_to?: string }) {
  return request<PageResult<any>>({ url: '/b/orders', data: params })
}

export function getOrder(id: number) {
  return request<any>({ url: `/b/orders/${id}` })
}

export function payOrder(id: number, payMethod: string, transactionId?: string) {
  return request({ url: `/b/orders/${id}/pay`, method: 'PUT', data: { pay_method: payMethod, transaction_id: transactionId } })
}

export function refundOrder(id: number, remark?: string) {
  return request({ url: `/b/orders/${id}/refund`, method: 'PUT', data: { remark } })
}

export function cancelOrder(id: number) {
  return request({ url: `/b/orders/${id}/cancel`, method: 'PUT' })
}
