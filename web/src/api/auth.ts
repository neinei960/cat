import { request } from './request'

export function staffLogin(data: LoginRequest) {
  return request<LoginResponse>({
    url: '/auth/staff/login',
    method: 'POST',
    data,
  })
}
