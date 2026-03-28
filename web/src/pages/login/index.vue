<template>
  <view class="login-page">
    <view class="login-card">
      <view class="logo-section">
        <text class="logo-icon">🐾</text>
        <text class="logo-title">猫咪洗护</text>
        <text class="logo-subtitle">店铺管理系统</text>
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
            @focus="phoneFocused = true"
            @blur="phoneFocused = false"
            :class="{ 'input-field-focus': phoneFocused }"
          />
        </view>

        <view class="input-group">
          <text class="input-label">密码</text>
          <input
            class="input-field"
            type="password"
            v-model="password"
            placeholder="请输入密码"
            @focus="pwdFocused = true"
            @blur="pwdFocused = false"
            :class="{ 'input-field-focus': pwdFocused }"
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
const phoneFocused = ref(false)
const pwdFocused = ref(false)

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
  background: linear-gradient(160deg, #EEF2FF 0%, #F5F6FA 50%, #F0FDF4 100%);
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 40rpx;
}

.login-card {
  width: 100%;
  max-width: 680rpx;
  background-color: #FFFFFF;
  border-radius: 28rpx;
  padding: 80rpx 60rpx;
  box-shadow: 0 12rpx 48rpx rgba(79, 70, 229, 0.10);
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
  font-size: 44rpx;
  font-weight: 800;
  color: #1F2937;
  margin-bottom: 12rpx;
  letter-spacing: 2rpx;
}

.logo-subtitle {
  font-size: 26rpx;
  color: #9CA3AF;
  letter-spacing: 1rpx;
}

.form-section {
  width: 100%;
}

.input-group {
  margin-bottom: 40rpx;
}

.input-label {
  font-size: 28rpx;
  color: #374151;
  font-weight: 500;
  margin-bottom: 16rpx;
  display: block;
}

.input-field {
  width: 100%;
  height: 96rpx;
  background-color: #FAFAFA;
  border-radius: 16rpx;
  padding: 0 32rpx;
  font-size: 30rpx;
  color: #1F2937;
  box-sizing: border-box;
  border: 1.5rpx solid #E5E7EB;
  transition: border-color 0.2s;
}

.input-field-focus {
  border-color: #6366F1;
  background-color: #FAFBFF;
}

.login-btn {
  width: 100%;
  height: 100rpx;
  background: linear-gradient(135deg, #6366F1, #4F46E5);
  color: #FFFFFF;
  font-size: 32rpx;
  font-weight: 600;
  border-radius: 20rpx;
  margin-top: 24rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  border: none;
  line-height: 100rpx;
  box-shadow: 0 8rpx 24rpx rgba(99, 102, 241, 0.35);
}

.login-btn::after {
  border: none;
}

.login-btn-disabled {
  opacity: 0.6;
  box-shadow: none;
}
</style>
