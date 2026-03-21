import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { staffLogin } from '@/api/auth'

export const useAuthStore = defineStore('auth', () => {
  const token = ref<string>('')
  const staffInfo = ref<StaffInfo | null>(null)

  const isLoggedIn = computed(() => !!token.value)

  function loadFromStorage() {
    token.value = uni.getStorageSync('token') || ''
    const info = uni.getStorageSync('staffInfo')
    if (info) {
      staffInfo.value = JSON.parse(info)
    }
  }

  async function login(phone: string, password: string) {
    const res = await staffLogin({ phone, password })
    token.value = res.data.token
    staffInfo.value = res.data.staff
    uni.setStorageSync('token', res.data.token)
    uni.setStorageSync('staffInfo', JSON.stringify(res.data.staff))
  }

  function logout() {
    token.value = ''
    staffInfo.value = null
    uni.removeStorageSync('token')
    uni.removeStorageSync('staffInfo')
    uni.reLaunch({ url: '/pages/login/index' })
  }

  return { token, staffInfo, isLoggedIn, loadFromStorage, login, logout }
})
