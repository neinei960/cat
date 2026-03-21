<template>
  <view class="login-page">
    <view class="login-card">
      <view class="logo-section">
        <text class="logo-icon">🐾</text>
        <text class="logo-title">宠物洗护管理系统</text>
        <text class="logo-subtitle">员工登录</text>
      </view>

      <view class="form-section">
        <view class="input-group">
          <text class="input-label">手机号</text>
          <input
            class="input-field"
            type="number"
            v-model="phone"
            placeholder="请输入手机号"
            maxlength="11"
          />
        </view>

        <view class="input-group">
          <text class="input-label">密码</text>
          <input
            class="input-field"
            type="password"
            v-model="password"
            placeholder="请输入密码"
          />
        </view>

        <button
          class="login-btn"
          :class="{ 'login-btn-disabled': loading }"
          :disabled="loading"
          @click="handleLogin"
        >
          {{ loading ? '登录中...' : '登 录' }}
        </button>
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useAuthStore } from '@/store/auth'

const phone = ref('')
const password = ref('')
const loading = ref(false)

const authStore = useAuthStore()

function validate(): boolean {
  if (!phone.value) {
    uni.showToast({ title: '请输入手机号', icon: 'none' })
    return false
  }
  if (!/^1\d{10}$/.test(phone.value)) {
    uni.showToast({ title: '请输入正确的手机号', icon: 'none' })
    return false
  }
  if (!password.value) {
    uni.showToast({ title: '请输入密码', icon: 'none' })
    return false
  }
  return true
}

async function handleLogin() {
  if (!validate() || loading.value) return

  loading.value = true
  try {
    await authStore.login(phone.value, password.value)
    uni.reLaunch({ url: '/pages/index/index' })
  } catch {
    // error toast handled by request.ts
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.login-page {
  min-height: 100vh;
  background-color: #F5F6FA;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 40rpx;
}

.login-card {
  width: 100%;
  max-width: 680rpx;
  background-color: #FFFFFF;
  border-radius: 24rpx;
  padding: 80rpx 60rpx;
  box-shadow: 0 8rpx 40rpx rgba(0, 0, 0, 0.06);
}

.logo-section {
  display: flex;
  flex-direction: column;
  align-items: center;
  margin-bottom: 80rpx;
}

.logo-icon {
  font-size: 80rpx;
  margin-bottom: 24rpx;
}

.logo-title {
  font-size: 40rpx;
  font-weight: 700;
  color: #1F2937;
  margin-bottom: 12rpx;
}

.logo-subtitle {
  font-size: 28rpx;
  color: #6B7280;
}

.form-section {
  width: 100%;
}

.input-group {
  margin-bottom: 40rpx;
}

.input-label {
  font-size: 28rpx;
  color: #1F2937;
  font-weight: 500;
  margin-bottom: 16rpx;
  display: block;
}

.input-field {
  width: 100%;
  height: 96rpx;
  background-color: #F5F6FA;
  border-radius: 16rpx;
  padding: 0 32rpx;
  font-size: 30rpx;
  color: #1F2937;
  box-sizing: border-box;
}

.login-btn {
  width: 100%;
  height: 96rpx;
  background-color: #4F46E5;
  color: #FFFFFF;
  font-size: 32rpx;
  font-weight: 600;
  border-radius: 16rpx;
  margin-top: 20rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  border: none;
  line-height: 96rpx;
}

.login-btn::after {
  border: none;
}

.login-btn-disabled {
  opacity: 0.6;
}
</style>
