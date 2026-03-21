interface StaffInfo {
  id: number
  name: string
  phone: string
  role: 'admin' | 'manager' | 'staff'
  shopId: number
  avatar?: string
}

interface LoginRequest {
  phone: string
  password: string
}

interface LoginResponse {
  token: string
  staff: StaffInfo
}
