const BASE_URL = 'http://localhost:8080/api/v1'

interface ApiResponse<T = any> {
  code: number
  data: T
  msg: string
}

export function request<T = any>(options: { url: string; method?: 'GET' | 'POST' | 'PUT' | 'DELETE'; data?: any }): Promise<ApiResponse<T>> {
  return new Promise((resolve, reject) => {
    const token = uni.getStorageSync('c_token')
    uni.request({
      url: BASE_URL + options.url,
      method: options.method || 'GET',
      data: options.data,
      header: {
        'Content-Type': 'application/json',
        ...(token ? { Authorization: `Bearer ${token}` } : {}),
      },
      success: (res) => {
        const data = res.data as ApiResponse<T>
        if (data.code === 0) {
          resolve(data)
        } else if (data.code === 401) {
          uni.removeStorageSync('c_token')
          // trigger wx login
          reject(data)
        } else {
          uni.showToast({ title: data.msg || '请求失败', icon: 'none' })
          reject(data)
        }
      },
      fail: (err) => {
        uni.showToast({ title: '网络错误', icon: 'none' })
        reject(err)
      },
    })
  })
}
