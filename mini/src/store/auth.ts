import { defineStore } from 'pinia'
import { ref, computed } from 'vue'

export const useAuthStore = defineStore('auth', () => {
  const token = ref('')
  const customerInfo = ref<any>(null)
  const isLoggedIn = computed(() => !!token.value)

  function loadFromStorage() {
    token.value = uni.getStorageSync('c_token') || ''
    const info = uni.getStorageSync('c_info')
    if (info) customerInfo.value = JSON.parse(info)
  }

  function setAuth(t: string, info: any) {
    token.value = t
    customerInfo.value = info
    uni.setStorageSync('c_token', t)
    uni.setStorageSync('c_info', JSON.stringify(info))
  }

  function logout() {
    token.value = ''
    customerInfo.value = null
    uni.removeStorageSync('c_token')
    uni.removeStorageSync('c_info')
  }

  return { token, customerInfo, isLoggedIn, loadFromStorage, setAuth, logout }
})
