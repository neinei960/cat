<template>
  <view class="login-page">
    <view class="login-shell">
      <view class="brand-panel">
        <view class="brand-chip">CAT SPA MANAGER</view>
        <text class="brand-title">猫咪洗护门店工作台</text>
        <text class="brand-subtitle">让预约、开单、会员和客户管理都在一个地方顺畅完成。</text>

        <view class="brand-metrics">
          <view class="metric-card">
            <text class="metric-label">门店状态</text>
            <text class="metric-value">实时经营</text>
          </view>
          <view class="metric-card">
            <text class="metric-label">工作重点</text>
            <text class="metric-value">预约到店与结算</text>
          </view>
          <view class="metric-card">
            <text class="metric-label">管理范围</text>
            <text class="metric-value">客户 / 猫咪 / 会员卡</text>
          </view>
        </view>
      </view>

      <view class="login-card">
        <view class="logo-section">
          <view class="logo-badge">🐾</view>
          <text class="logo-title">账号登录</text>
          <text class="logo-subtitle">进入今天的门店工作流</text>
        </view>

        <view class="status-banner" v-if="errorMessage">
          <text class="status-banner-title">登录未完成</text>
          <text class="status-banner-text">{{ errorMessage }}</text>
        </view>

        <form class="form-section" @submit.prevent="handleLogin">
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
              confirm-type="next"
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
              confirm-type="done"
              @confirm="handleLogin"
            />
          </view>

          <view class="tips-row">
            <text class="tips-text">如遇登录异常，请先确认服务状态或联系管理员处理。</text>
          </view>

          <button
            form-type="submit"
            class="login-btn"
            :class="{ 'login-btn-disabled': loading }"
            :disabled="loading"
            @click="handleLogin"
          >
            {{ loading ? '登录中...' : '进入工作台' }}
          </button>
        </form>
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
const errorMessage = ref('')

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
  errorMessage.value = ''
  try {
    await authStore.login(phone.value, password.value)
    uni.reLaunch({ url: '/pages/index/index' })
  } catch (error: any) {
    errorMessage.value = error?.message || '当前登录服务响应异常，请稍后再试。'
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.login-page {
  min-height: 100vh;
  background:
    radial-gradient(circle at top left, rgba(79, 70, 229, 0.08), transparent 28%),
    radial-gradient(circle at bottom right, rgba(238, 242, 255, 0.6), transparent 32%),
    linear-gradient(160deg, #F8FAFC 0%, #F5F6FA 60%, #EEF2FF 100%);
  padding: 44rpx;
  box-sizing: border-box;
}

.login-shell {
  min-height: calc(100vh - 88rpx);
  display: grid;
  grid-template-columns: 1.1fr 0.9fr;
  gap: 28rpx;
  align-items: stretch;
}

.brand-panel,
.login-card {
  border-radius: 36rpx;
  overflow: hidden;
}

.brand-panel {
  background:
    linear-gradient(155deg, rgba(79, 70, 229, 0.96), rgba(67, 56, 202, 0.92)),
    linear-gradient(135deg, #4F46E5, #374151);
  color: #FFFFFF;
  padding: 72rpx 56rpx;
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  box-shadow: 0 24rpx 72rpx rgba(17, 24, 39, 0.22);
}

.brand-chip {
  align-self: flex-start;
  padding: 10rpx 18rpx;
  border-radius: 999rpx;
  background: rgba(255, 255, 255, 0.1);
  color: #FDE68A;
  font-size: 22rpx;
  letter-spacing: 2rpx;
}

.brand-title {
  display: block;
  margin-top: 36rpx;
  font-size: 64rpx;
  line-height: 1.08;
  font-weight: 800;
}

.brand-subtitle {
  display: block;
  margin-top: 22rpx;
  color: rgba(255, 255, 255, 0.72);
  font-size: 28rpx;
  line-height: 1.7;
  max-width: 540rpx;
}

.brand-metrics {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 18rpx;
  margin-top: 56rpx;
}

.metric-card {
  padding: 22rpx;
  border-radius: 24rpx;
  background: rgba(255, 255, 255, 0.08);
  border: 1rpx solid rgba(255, 255, 255, 0.08);
  transition: transform 0.2s ease, box-shadow 0.2s ease;
}

.metric-card:hover {
  transform: translateY(-4rpx);
  box-shadow: 0 8rpx 20rpx rgba(0, 0, 0, 0.12);
}

.metric-label {
  display: block;
  color: rgba(255, 255, 255, 0.56);
  font-size: 22rpx;
}

.metric-value {
  display: block;
  margin-top: 12rpx;
  font-size: 28rpx;
  font-weight: 700;
}

.login-card {
  background: rgba(255, 255, 255, 0.92);
  backdrop-filter: blur(20px);
  padding: 72rpx 56rpx;
  box-shadow: 0 18rpx 60rpx rgba(99, 102, 241, 0.12);
}

.logo-section {
  margin-bottom: 40rpx;
}

.logo-badge {
  width: 96rpx;
  height: 96rpx;
  border-radius: 28rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #FB7185, #F59E0B);
  box-shadow: 0 10rpx 30rpx rgba(251, 113, 133, 0.24);
  font-size: 42rpx;
}

.logo-title {
  display: block;
  margin-top: 28rpx;
  font-size: 44rpx;
  font-weight: 800;
  color: #111827;
}

.logo-subtitle {
  display: block;
  margin-top: 10rpx;
  color: #6B7280;
  font-size: 26rpx;
}

.status-banner {
  margin-bottom: 28rpx;
  padding: 22rpx 24rpx;
  border-radius: 20rpx;
  background: #FFF7ED;
  border: 1rpx solid #FDBA74;
}

.status-banner-title {
  display: block;
  color: #C2410C;
  font-size: 24rpx;
  font-weight: 700;
}

.status-banner-text {
  display: block;
  margin-top: 8rpx;
  color: #9A3412;
  font-size: 24rpx;
  line-height: 1.6;
}

.form-section {
  width: 100%;
}

.input-group {
  margin-bottom: 32rpx;
}

.input-label {
  font-size: 26rpx;
  color: #374151;
  font-weight: 600;
  margin-bottom: 14rpx;
  display: block;
}

.input-field {
  width: 100%;
  height: 98rpx;
  background-color: #FFFFFF;
  border-radius: 20rpx;
  padding: 0 30rpx;
  font-size: 30rpx;
  color: #111827;
  box-sizing: border-box;
  border: 2rpx solid #E5E7EB;
  box-shadow: inset 0 1rpx 0 rgba(255, 255, 255, 0.8);
}

.input-field-focus {
  border-color: #F97316;
  background-color: #FFFDF8;
  box-shadow: 0 0 0 6rpx rgba(249, 115, 22, 0.08);
}

.tips-row {
  margin-top: -4rpx;
  margin-bottom: 26rpx;
}

.tips-text {
  font-size: 22rpx;
  line-height: 1.6;
  color: #9CA3AF;
}

.login-btn {
  width: 100%;
  height: 104rpx;
  background: linear-gradient(135deg, #F97316, #EA580C);
  color: #FFFFFF;
  font-size: 32rpx;
  font-weight: 700;
  border-radius: 22rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  border: none;
  line-height: 104rpx;
  box-shadow: 0 12rpx 28rpx rgba(234, 88, 12, 0.28);
}

.login-btn::after {
  border: none;
}

.login-btn-disabled {
  opacity: 0.65;
  box-shadow: none;
}

@media (max-width: 900px) {
  .login-shell {
    grid-template-columns: 1fr;
  }

  .brand-panel {
    padding: 44rpx 36rpx;
  }

  .brand-title {
    font-size: 48rpx;
  }

  .brand-metrics {
    grid-template-columns: 1fr;
  }

  .login-card {
    padding: 48rpx 36rpx;
  }
}
</style>
