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

export function updateOrder(id: number, data: CreateOrderReq) {
  return request<any>({ url: `/b/orders/${id}`, method: 'PUT', data })
}

export function createOrderFromAppointment(appointmentId: number) {
  return request<any>({ url: '/b/orders/from-appointment', method: 'POST', data: { appointment_id: appointmentId } })
}

export function createBatchOrdersFromAppointment(appointmentId: number, extra?: any) {
  return request<Order>({ url: '/b/orders/from-appointment/batch', method: 'POST', data: { appointment_id: appointmentId, ...extra } })
}

export function getOrderList(params?: PageParams & { status?: number; keyword?: string; date_from?: string; date_to?: string; product_keyword?: string; customer_id?: number; order_kind?: string }) {
  return request<PageResult<any>>({ url: '/b/orders', data: params })
}

export function getDeletedOrders(params?: PageParams) {
  return request<PageResult<any>>({ url: '/b/orders/trash', data: params })
}

export function getOrder(id: number, includeDeleted = false) {
  return request<any>({ url: `/b/orders/${id}`, data: includeDeleted ? { include_deleted: 1 } : undefined })
}

export function payOrder(id: number, payMethod: string, transactionId?: string, remark?: string) {
  return request({ url: `/b/orders/${id}/pay`, method: 'PUT', data: { pay_method: payMethod, transaction_id: transactionId, remark } })
}

export function updateOrderRemark(id: number, remark: string) {
  return request({ url: `/b/orders/${id}/remark`, method: 'PUT', data: { remark } })
}

export function refundOrder(id: number, remark?: string) {
  return request({ url: `/b/orders/${id}/refund`, method: 'PUT', data: { remark } })
}

export function cancelOrder(id: number) {
  return request({ url: `/b/orders/${id}/cancel`, method: 'PUT' })
}

export function deleteOrder(id: number) {
  return request({ url: `/b/orders/${id}`, method: 'DELETE' })
}

export function restoreOrder(id: number) {
  return request({ url: `/b/orders/${id}/restore`, method: 'POST' })
}
