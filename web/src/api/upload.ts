const BASE_URL = import.meta.env.VITE_API_BASE_URL || ''

interface UploadOptions {
  preserveOriginal?: boolean
}

export function uploadFile(filePath: string): Promise<string> {
  return new Promise((resolve, reject) => {
    const token = uni.getStorageSync('token')
    uni.uploadFile({
      url: BASE_URL + '/b/upload',
      filePath,
      name: 'file',
      header: token ? { Authorization: `Bearer ${token}` } : {},
      success: (res) => {
        const data = JSON.parse(res.data)
        if (data.code === 0 && data.data?.url) {
          resolve(data.data.url)
        } else {
          uni.showToast({ title: data.msg || '上传失败', icon: 'none' })
          reject(data)
        }
      },
      fail: (err) => {
        uni.showToast({ title: '上传失败', icon: 'none' })
        reject(err)
      },
    })
  })
}

export async function uploadH5File(file: File, options: UploadOptions = {}): Promise<string> {
  if (typeof window === 'undefined' || typeof FormData === 'undefined') {
    throw new Error('当前环境不支持H5上传')
  }

  const token = uni.getStorageSync('token')
  const apiBase = BASE_URL.startsWith('http') ? BASE_URL : window.location.origin + BASE_URL
  const formData = new FormData()
  formData.append('file', file)
  if (options.preserveOriginal) {
    formData.append('preserve_original', '1')
  }

  const res = await fetch(`${apiBase}/b/upload`, {
    method: 'POST',
    headers: token ? { Authorization: `Bearer ${token}` } : {},
    body: formData,
  })
  const data = await res.json().catch(() => null)
  if (!res.ok || data?.code !== 0 || !data?.data?.url) {
    throw new Error(data?.msg || '上传失败')
  }
  return data.data.url
}
